package redigosrv

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

//RedigoCollector struct to access metrics
type RedigoCollector struct {
	subscriptionsActive prometheus.Gauge
	publishTrafficSize  prometheus.Counter
	subscribeSuccesses  prometheus.Counter
	subscribeFailures   prometheus.Counter
	commandCalls        *prometheus.CounterVec
	methodDuration      *prometheus.CounterVec
}

// RedigoCollectorOptions struct to add custom options in metrics
type RedigoCollectorOptions struct {
	Prefix string
}

const (
	//PublishMetricMethodName has the name of the specific method
	PublishMetricMethodName string = "Publish"
	//SubscribeMetricMethodName has the name of the specific method
	SubscribeMetricMethodName string = "Subscribe"
)

var redigoMetricsLabels = []string{"method", "command"}

//RedigoCollectorDefaultOptions will return the instance of RedigoCollectorDefaultOptions with values default
func RedigoCollectorDefaultOptions() RedigoCollectorOptions {
	return RedigoCollectorOptions{
		Prefix: "",
	}
}

// NewRedigoCollector will return new instance of RedigoCollector with all metrics started
func NewRedigoCollector(opts RedigoCollectorOptions) *RedigoCollector {

	prefix := opts.Prefix
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	return &RedigoCollector{
		subscriptionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("redigo_%ssubscriptions_active", prefix),
			Help: "Current total of subscriptions",
		}),
		publishTrafficSize: prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("redigo_%spublish_traffic_size", prefix),
			Help: "Total of data trafficked",
		}),
		subscribeSuccesses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("redigo_%ssubscribe_success", prefix),
			Help: "Total of success when call Subscribed",
		}),
		subscribeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("redigo_%ssubscribe_failures", prefix),
			Help: "Total of failed when call Subscribed",
		}),
		commandCalls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: fmt.Sprintf("redigo_%scommand_calls", prefix),
			Help: "Total of command calls (Success or failures)",
		}, redigoMetricsLabels),
		methodDuration: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: fmt.Sprintf("redigo_%smethod_duration", prefix),
			Help: "Total of duration from method",
		}, redigoMetricsLabels),
	}

}

// Describe returns the description of metrics colllected by this collector
func (collector *RedigoCollector) Describe(desc chan<- *prometheus.Desc) {
	collector.commandCalls.Describe(desc)
	collector.subscriptionsActive.Describe(desc)
	collector.subscribeSuccesses.Describe(desc)
	collector.subscribeFailures.Describe(desc)
	collector.publishTrafficSize.Describe(desc)
}

// Collect provides metrics to prometheus
func (collector *RedigoCollector) Collect(metrics chan<- prometheus.Metric) {
	collector.commandCalls.Collect(metrics)
	collector.subscriptionsActive.Collect(metrics)
	collector.subscribeSuccesses.Collect(metrics)
	collector.subscribeFailures.Collect(metrics)
	collector.publishTrafficSize.Collect(metrics)
}
