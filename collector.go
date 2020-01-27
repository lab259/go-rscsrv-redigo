package redigosrv

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

//RedigoCollector struct to access metrics
type RedigoCollector struct {
	publishTrafficSize prometheus.Counter
	methodCalls        *prometheus.CounterVec
	subscribeAmount    prometheus.Gauge
	subscribeSuccesses prometheus.Counter
	subscribeFailures  prometheus.Counter
}

// RedigoCollectorOptions struct to add custom name in metrics
type RedigoCollectorOptions struct {
	Separator string
	Prefix    []string
}

const (
	//PublishMetricMethodName has the name of the specific method
	PublishMetricMethodName string = "Publish"
	//SubscribeMetricMethodName has the name of the specific method
	SubscribeMetricMethodName string = "Subscribe"
)

var redigoMetricLabels = []string{"method"}

//RedigoCollectorDefaultOptions will return the instance of RedigoCollectorDefaultOptions with values default
func RedigoCollectorDefaultOptions() RedigoCollectorOptions {
	return RedigoCollectorOptions{
		Separator: "_",
		Prefix:    []string{"redigo"},
	}
}

/*
	Output with default options:

	Ex: fmt.Sprintf("%spublishTrafficSize", prefix)
	Exit: redigo_publish_traffic_size
*/

// NewRedigoCollector will return new instance of RedigoCollector with all metrics started
func NewRedigoCollector(opts RedigoCollectorOptions) *RedigoCollector {

	prefix := strings.Join(opts.Prefix[:], opts.Separator)
	prefix += opts.Separator

	return &RedigoCollector{
		publishTrafficSize: prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%spublish_traffic_size", prefix),
			Help: "Total of data trafficked",
		}),
		methodCalls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: fmt.Sprintf("%smethod_calls", prefix),
			Help: "Total of calls of method Subscribe (Success or failures)",
		}, redigoMetricLabels),
		subscribeAmount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("%ssubscribe_amount", prefix),
			Help: "Current total of subscriptions",
		}),
		subscribeSuccesses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%ssubscribe_success", prefix),
			Help: "Total of success when call Subscribed",
		}),
		subscribeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%ssubscribe_failures", prefix),
			Help: "Total of failed when call Subscribed",
		}),
	}

}

func (collector *RedigoCollector) Describe(desc chan<- *prometheus.Desc) {
	collector.methodCalls.Describe(desc)
	collector.subscribeAmount.Describe(desc)
	collector.subscribeSuccesses.Describe(desc)
	collector.subscribeFailures.Describe(desc)
	collector.publishTrafficSize.Describe(desc)
}

func (collector *RedigoCollector) Collect(metrics chan<- prometheus.Metric) {
	collector.methodCalls.Collect(metrics)
	collector.subscribeAmount.Collect(metrics)
	collector.subscribeSuccesses.Collect(metrics)
	collector.subscribeFailures.Collect(metrics)
	collector.publishTrafficSize.Collect(metrics)
}
