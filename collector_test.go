package redigosrv

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var _ = Describe("RedigoCollector", func() {
	It("should test publishTrafficSize", func(done Done) {
		var service RedigoService
		var metric dto.Metric

		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())

		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		Expect(service.Publish(ctx, "test-01", "hello from subscription")).To(Succeed())

		Expect(service.Collector.publishTrafficSize.With(prometheus.Labels{
			"method": "publish",
		}).Write(&metric)).To(Succeed())

		Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))

		cancel()
		close(done)
	})

	It("should test subscribeCalls", func(done Done) {
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

			service.Collector.subscribeCalls.With(prometheus.Labels{
				"method": "publish",
			}).Inc()

			service.Collector.subscribeCalls.With(prometheus.Labels{
				"method": "publish",
			}).Inc()

			cancel()
			return nil
		}, "test-02")

		Expect(service.Collector.subscribeCalls.With(prometheus.Labels{
			"method": "publish",
		}).Write(&metric)).To(BeNil())

		Expect(metric.GetCounter().GetValue()).To(Equal(float64(2)))
		close(done)
	})

	It("should test subscribeAmount", func(done Done) {
		var service RedigoService
		var metric dto.Metric

		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())

		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-03", []byte("second test"))).To(Succeed())
			return nil
		}

		service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			Expect(data).To(Equal([]byte("second test")))

			service.Collector.subscribeAmount.With(prometheus.Labels{
				"method": "publish",
			}).Add(2)

			service.Collector.subscribeAmount.With(prometheus.Labels{
				"method": "publish",
			}).Dec()

			cancel()
			return nil
		}, "test-03")

		Expect(service.Collector.subscribeAmount.With(prometheus.Labels{
			"method": "publish",
		}).Write(&metric)).To(BeNil())

		Expect(metric.GetGauge().GetValue()).To(Equal(float64(1)))
		close(done)
	})

})
