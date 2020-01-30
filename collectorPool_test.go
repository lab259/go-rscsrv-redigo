package redigosrv

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type poolStatsFake struct{}

func (psf *poolStatsFake) Stats() redis.PoolStats {
	return redis.PoolStats{
		ActiveCount: 10,
		IdleCount:   7,
	}
}

var _ = Describe("PoolStatsCollector", func() {
	It("should generate description name", func() {
		collector := NewPoolStatsCollector(nil, RedigoCollectorPoolOptions{
			Prefix: "",
		})
		ch := make(chan *prometheus.Desc)
		go func() {
			collector.Describe(ch)
			close(ch)
		}()
		Expect((<-ch).String()).To(ContainSubstring("redigo_active_count"))
		Expect((<-ch).String()).To(ContainSubstring("redigo_idle_count"))
	})

	It("should generate custom description name", func() {
		customName := "mabel"
		collector := NewPoolStatsCollector(nil, RedigoCollectorPoolOptions{
			Prefix: customName,
		})
		ch := make(chan *prometheus.Desc)
		go func() {
			collector.Describe(ch)
			close(ch)
		}()
		Expect((<-ch).String()).To(ContainSubstring(fmt.Sprintf("redigo_%s_active_count", customName)))
		Expect((<-ch).String()).To(ContainSubstring(fmt.Sprintf("redigo_%s_idle_count", customName)))
	})

	It("should test default values", func() {
		fakePool := poolStatsFake{}
		collector := NewPoolStatsCollector(&fakePool, RedigoCollectorPoolOptions{
			Prefix: "",
		})
		ch := make(chan prometheus.Metric)

		go func() {
			collector.Collect(ch)
			close(ch)
		}()
		var metric dto.Metric

		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetGauge().GetValue()).To(Equal(float64(10)))
		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetGauge().GetValue()).To(Equal(float64(7)))
	})
})
