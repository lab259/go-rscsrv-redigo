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
	subscribeSuccess   prometheus.Counter
	subscribeFailures  prometheus.Counter
}

// RedigoCollectorOptions struct to add custom name in metrics
type RedigoCollectorOptions struct {
	SplitBy string
	Prefix  []string
	Suffix  []string
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
		SplitBy: "_",
		Prefix:  []string{"redigo"},
		Suffix:  []string{"collector"},
	}
}

/*
	Output with default options:

	Ex: fmt.Sprintf("%spublishTrafficSize%s", prefix, suffix)
	Exit: redigo_publishTrafficSize_collector
*/

// NewRendigoCollector will return new instance of RedigoCollector with all metrics started
func NewRendigoCollector(opts RedigoCollectorOptions) *RedigoCollector {

	prefix := strings.Join(opts.Prefix[:], opts.SplitBy)
	suffix := strings.Join(opts.Suffix[:], opts.SplitBy)

	if len(opts.Prefix) == 1 {
		prefix += opts.SplitBy
	}

	if len(opts.Suffix) == 1 {
		suffix += opts.SplitBy
	}

	return &RedigoCollector{
		publishTrafficSize: prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%spublishTrafficSize%s", prefix, suffix),
			Help: "Total of data trafficked",
		}),
		methodCalls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: fmt.Sprintf("%ssubscribeCalls%s", prefix, suffix),
			Help: "Total of calls of method Subscribe (Success or failures)",
		}, redigoMetricLabels),
		subscribeAmount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("%ssubscribeAmount%s", prefix, suffix),
			Help: "Current total of subscriptions",
		}),
		subscribeSuccess: prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%ssubscribeSuccess%s", prefix, suffix),
			Help: "Total of success when call Subscribed",
		}),
		subscribeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%ssubscribeFailures%s", prefix, suffix),
			Help: "Total of failed when call Subscribed",
		}),
	}

}

func (collector *RedigoCollector) Describe(desc chan<- *prometheus.Desc) {
	collector.methodCalls.Describe(desc)
	collector.subscribeAmount.Describe(desc)
	collector.subscribeSuccess.Describe(desc)
	collector.subscribeFailures.Describe(desc)
	collector.publishTrafficSize.Describe(desc)
}

func (collector *RedigoCollector) Collect(metrics chan<- prometheus.Metric) {
	collector.methodCalls.Collect(metrics)
	collector.subscribeAmount.Collect(metrics)
	collector.subscribeSuccess.Collect(metrics)
	collector.subscribeFailures.Collect(metrics)
	collector.publishTrafficSize.Collect(metrics)
}
