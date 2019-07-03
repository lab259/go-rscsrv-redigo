package redigosrv

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RedigoService (PubSub)", func() {
	It("should subscribe and publish messages", func(done Done) {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())

		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-01", "hello from subscription")).To(Succeed())
			return nil
		}

		Expect(service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			Expect(data).To(Equal([]byte("\"hello from subscription\"")))
			cancel()
			return nil
		}, "test-01")).To(Succeed())

		close(done)
	})

	It("should subscribe and publish raw messages", func(done Done) {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		defer service.Stop()

		ctx, cancel := context.WithCancel(context.Background())

		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-01", []byte("hello from subscription"))).To(Succeed())
			return nil
		}

		Expect(service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			Expect(data).To(Equal([]byte("hello from subscription")))
			cancel()
			return nil
		}, "test-01")).To(Succeed())

		close(done)
	})

	It("should propagate error from subscribed handler", func(done Done) {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		defer service.Stop()

		ctx := context.Background()

		onSubscribed := func() error {
			return errors.New("something bad")
		}

		err := service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			Fail("subscription handler should not be called")
			return nil
		}, "test-01")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("something bad"))

		close(done)
	})

	It("should propagate error from subscription handler", func(done Done) {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		defer service.Stop()

		ctx := context.Background()

		onSubscribed := func() error {
			Expect(service.Publish(ctx, "test-01", []byte("hello from subscription"))).To(Succeed())
			return nil
		}

		err := service.Subscribe(ctx, onSubscribed, func(channel string, data []byte) error {
			return errors.New("something bad")
		}, "test-01")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("something bad"))

		close(done)
	})
})
