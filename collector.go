package redigosrv

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

//RedigoCollector //TODO
type RedigoCollector struct {
	publishTrafficSize *prometheus.CounterVec
	subscribeCalls     *prometheus.CounterVec
	subscribeAmount    *prometheus.GaugeVec
	subscribeSuccess   *prometheus.CounterVec
	subscribeFailures  *prometheus.CounterVec
}

// RedigoCollectorOptions //TODO
type RedigoCollectorOptions struct {
	SplitBy string
	Prefix  []string
	Suffix  []string
}

const (
	//PublishMetricMethod has the name of the specific method
	PublishMetricMethod string = "Publish"
	//SubscribeMetricMethod has the name of the specific method
	SubscribeMetricMethod string = "Subscribe"
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
		publishTrafficSize: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: fmt.Sprintf("%spublishTrafficSize%s", prefix, suffix),
			Help: "Total of data trafficked",
		},
			redigoMetricLabels),
		subscribeCalls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: fmt.Sprintf("%ssubscribeCalls%s", prefix, suffix),
			Help: "Total of calls of method Subscribe (Success and failures)",
		}, redigoMetricLabels),
		subscribeAmount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: fmt.Sprintf("%ssubscribeAmount%s", prefix, suffix),
			Help: "Current total of subscriptions",
		}, redigoMetricLabels),
		subscribeSuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: fmt.Sprintf("%ssubscribeSuccess%s", prefix, suffix),
			Help: "Total of success when call Subscribe",
		},
			redigoMetricLabels),
		subscribeFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: fmt.Sprintf("%ssubscribeFailures%s", prefix, suffix),
			Help: "Total of failed when call Subscribe",
		},
			redigoMetricLabels),
	}

}

func (collector *RedigoCollector) Describe(desc chan<- *prometheus.Desc) {
	collector.publishTrafficSize.Describe(desc)
	collector.subscribeCalls.Describe(desc)
	collector.subscribeAmount.Describe(desc)
	collector.subscribeSuccess.Describe(desc)
	collector.subscribeFailures.Describe(desc)
}

func (collector *RedigoCollector) Collect(metrics chan<- prometheus.Metric) {
	collector.publishTrafficSize.Collect(metrics)
	collector.subscribeCalls.Collect(metrics)
	collector.subscribeAmount.Collect(metrics)
	collector.subscribeSuccess.Collect(metrics)
	collector.subscribeFailures.Collect(metrics)
}
