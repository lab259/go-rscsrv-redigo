package redigosrv

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Recovery", func() {
	It("should call recovery when to happen panic in method subscription", func(done Done) {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		defer service.Stop()

		ctx := context.Background()

		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-01", []byte("hello"))).To(Succeed())
			return nil
		}

		err := service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			panic("Panic when subscription is called")
		}, "test-01")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Panic when subscription is called"))

		close(done)
	})

	It("should call recovery when to happen panic in method subscribed", func(done Done) {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-01", []byte("hello"))).To(Succeed())
			panic("Panic when subscribed is called")
		}

		err := service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			cancel()
			return nil
		}, "test-01")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Panic when subscribed is called"))

		close(done)
	})
})
