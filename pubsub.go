package redigosrv

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
)

// SubscriptionHandler is called for each new message.
type SubscriptionHandler func(channel string, data []byte) error

// SubscribedHandler it called when all channels are subscribed.
type SubscribedHandler func() error

// Publish sends a data payload to a specific channel.
func (service *RedigoService) Publish(ctx context.Context, channel string, data interface{}) error {

	counter := service.Collector.publishTrafficSize

	// Total of calls of method
	service.Collector.methodCalls.With(prometheus.Labels{
		"method": SubscribeMetricMethodName,
	}).Inc()

	return service.RunWithConn(func(conn redis.ConnWithTimeout) error {
		var message []byte
		switch t := data.(type) {
		case []byte:
			// Increment publishTrafficSize with the size of message
			counter.Add(float64(len(t)))
			message = t
		default:
			m, err := json.Marshal(t)
			if err != nil {
				return err
			}
			// Increment publishTrafficSize with the size of message
			counter.Add(float64(len(m)))
			message = m
		}

		_, err := conn.Do("PUBLISH", channel, message)
		return err
	})
}

// Subscribe listens for messages on Redis pubsub channels. The
// subscribed function is called after the channels are subscribed. The subscription
// function is called for each message.
func (service *RedigoService) Subscribe(ctx context.Context, subscribed SubscribedHandler, subscription SubscriptionHandler, channels ...string) error {

	// Total of calls of method
	service.Collector.methodCalls.With(prometheus.Labels{
		"method": SubscribeMetricMethodName,
	}).Inc()

	c, err := redis.Dial("tcp", service.Configuration.Address,
		// Read timeout on server should be greater than ping period.
		redis.DialReadTimeout(service.Configuration.PubSub.ReadTimeout),
		redis.DialWriteTimeout(service.Configuration.PubSub.WriteTimeout),
	)
	if err != nil {
		return err
	}
	defer c.Close()

	psc := redis.PubSubConn{Conn: c}
	if err := psc.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
		return err
	}

	done := make(chan error, 1)

	// Start a goroutine to receive notifications from the server.
	go func() {
		for {
			switch n := psc.Receive().(type) {
			case error:
				// Increment to count failures
				service.Collector.subscribeFailures.Inc()

				done <- n
				return
			case redis.Message:
				if err := subscription(n.Channel, n.Data); err != nil {

					// Increment to count failures
					service.Collector.subscribeFailures.Inc()

					done <- err
					return
				}

				// Increment to count success
				service.Collector.subscribeSuccesses.Inc()

			case redis.Subscription:
				switch n.Count {
				case len(channels):

					// Increment 1 in subscribeAmount
					service.Collector.subscribeAmount.Inc()

					// Notify application when all channels are subscribed.
					if err := subscribed(); err != nil {

						// Increment to count failures
						service.Collector.subscribeFailures.Inc()

						done <- err
						return
					}

					// Increment to count success
					service.Collector.subscribeSuccesses.Inc()

				case 0:
					// Return from the goroutine when all channels are unsubscribed.
					done <- nil
					return
				}
			}
		}
	}()

	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	ticker := time.NewTicker(service.Configuration.PubSub.HealthCheckInterval)
	defer ticker.Stop()

loop:
	for err == nil {
		select {
		case <-ticker.C:
			// Send ping to test health of connection and server. If
			// corresponding pong is not received, then receive on the
			// connection will timeout and the receive goroutine will exit.
			if err = psc.Ping(""); err != nil {

				// Increment to count failures
				service.Collector.subscribeFailures.Inc()

				break loop
			}
		case <-ctx.Done():
			break loop
		case err := <-done:
			// Return error from the receive goroutine.
			return err
		}
	}

	// Decrement 1 in subscribeAmount
	service.Collector.subscribeAmount.Dec()

	// Signal the receiving goroutine to exit by unsubscribing from all channels.
	psc.Unsubscribe()

	// Wait for goroutine to complete.
	return <-done
}
