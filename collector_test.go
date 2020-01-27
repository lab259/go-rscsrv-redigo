package redigosrv

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var _ = Describe("RedigoCollector", func() {
	It("should count total of data send using method Publish", func(done Done) {
		var service RedigoService
		var metric dto.Metric

		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())

		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		Expect(service.Publish(ctx, "test-01", "hello from subscription")).To(Succeed())

		Expect(service.Collector.publishTrafficSize.Write(&metric)).To(Succeed())

		Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))

		cancel()
		close(done)
	})

	It("should count all calls in method subscribe", func(done Done) {
		var service RedigoService
		var metric dto.Metric

		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())

		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-02", []byte("second test"))).To(Succeed())
			return nil
		}

		service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			Expect(data).To(Equal([]byte("second test")))
			cancel()
			return nil
		}, "test-02")

		service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			Expect(data).To(Equal([]byte("second test")))
			cancel()
			return nil
		}, "test-02")

		Expect(service.Collector.methodCalls.With(prometheus.Labels{
			"method": SubscribeMetricMethodName,
		}).Write(&metric)).To(BeNil())

		Expect(metric.GetCounter().GetValue()).To(Equal(float64(2)))
		close(done)
	})

	It("should increment when subscribed is called", func(done Done) {
		var service RedigoService
		var metric dto.Metric

		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())

		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-03", []byte("third test"))).To(Succeed())
			return nil
		}

		service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			Expect(data).To(Equal([]byte("third test")))
			Expect(service.Collector.subscribeAmount.Write(&metric)).To(BeNil())
			Expect(metric.GetGauge().GetValue()).To(Equal(float64(1)))

			cancel()
			return nil
		}, "test-03")

		close(done)
	}, 1)

	It("should decrement when subscribed is finished", func(done Done) {
		var service RedigoService
		var metric dto.Metric

		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())

		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-03", []byte("third test"))).To(Succeed())
			return nil
		}

		service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			Expect(data).To(Equal([]byte("third test")))
			cancel()
			return nil
		}, "test-03")

		ctx, cancel = context.WithCancel(context.Background())
		onSubscribed = func() error {
			Expect(service.Publish(ctx, "test-03", []byte("third test"))).To(Succeed())
			return nil
		}

		service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			Expect(data).To(Equal([]byte("third test")))
			cancel()
			return nil
		}, "test-03")

		Expect(service.Collector.subscribeAmount.Write(&metric)).To(BeNil())

		Expect(metric.GetGauge().GetValue()).To(Equal(float64(0)))
		close(done)
	})

	It("should count failures when any error is found in subscribe", func(done Done) {
		var service RedigoService
		var metric dto.Metric

		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())

		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-04", []byte("four test"))).To(Succeed())
			return nil
		}

		service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			cancel()
			return errors.New("Error")
		}, "test-04")

		ctx, cancel = context.WithCancel(context.Background())
		onSubscribed = func() error {
			Expect(service.Publish(ctx, "test-04", []byte("four test"))).To(Succeed())
			return errors.New("Error")
		}

		service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			cancel()
			return nil
		}, "test-04")

		Expect(service.Collector.subscribeFailures.Write(&metric)).To(BeNil())

		Expect(metric.GetCounter().GetValue()).To(Equal(float64(2)))
		close(done)
	})

	It("should count successes when not found errors in subscribe", func(done Done) {
		var service RedigoService
		var metric dto.Metric

		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())

		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-05", []byte("five test"))).To(Succeed())
			return nil
		}

		service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			cancel()
			return nil
		}, "test-05")

		Expect(service.Collector.subscribeSuccess.Write(&metric)).To(BeNil())

		Expect(metric.GetCounter().GetValue()).To(Equal(float64(2)))
		close(done)
	})

})
