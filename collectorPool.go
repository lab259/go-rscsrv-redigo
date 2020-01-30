package redigosrv

import (
	"fmt"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
)

//PoolStatsCollector is a struct of statistics/metrics
type PoolStatsCollector struct {
	pool        PoolStats
	ActiveCount *prometheus.Desc
	IdleCount   *prometheus.Desc
}

//PoolStats interface of Stats
type PoolStats interface {
	Stats() redis.PoolStats
}

//RedigoCollectorPoolOptions struct for options of PoolCollector
type RedigoCollectorPoolOptions struct {
	Prefix string
}

//NewPoolStatsCollector will return new instance of PoolStatsCollector
func NewPoolStatsCollector(pool PoolStats, opts RedigoCollectorPoolOptions) *PoolStatsCollector {

	prefix := opts.Prefix
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	return &PoolStatsCollector{
		pool:        pool,
		ActiveCount: prometheus.NewDesc(fmt.Sprintf("redigo_%sactive_count", prefix), "The number of connections actived in pool (used or not).", nil, nil),
		IdleCount:   prometheus.NewDesc(fmt.Sprintf("redigo_%sidle_count", prefix), "The number of idle connections in the pool.", nil, nil),
	}
}

// Describe returns the description of metrics colllected by this collector
func (collector *PoolStatsCollector) Describe(desc chan<- *prometheus.Desc) {
	desc <- collector.ActiveCount
	desc <- collector.IdleCount
}

// Collect gets the pool stats information and provides it to the prometheus
func (collector *PoolStatsCollector) Collect(metrics chan<- prometheus.Metric) {
	stats := collector.pool.Stats()
	metrics <- prometheus.MustNewConstMetric(collector.ActiveCount, prometheus.GaugeValue, float64(stats.ActiveCount))
	metrics <- prometheus.MustNewConstMetric(collector.IdleCount, prometheus.GaugeValue, float64(stats.IdleCount))
}
