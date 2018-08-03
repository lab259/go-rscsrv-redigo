package redigosrv

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jamillosantos/macchiato"
	"github.com/lab259/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
	log.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	macchiato.RunSpecs(t, "Redigo Test Suite")
}

func pingConnection(conn redis.ConnWithTimeout) error {
	Expect(conn.Err()).To(BeNil())
	_, err := conn.Do("PING")
	Expect(err).To(BeNil())
	return nil
}

// errConn was copied form the redigo.
type errorConn struct{ err error }

func (ec errorConn) Do(string, ...interface{}) (interface{}, error) { return nil, ec.err }
func (ec errorConn) DoWithTimeout(time.Duration, string, ...interface{}) (interface{}, error) {
	return nil, ec.err
}
func (ec errorConn) Send(string, ...interface{}) error                     { return ec.err }
func (ec errorConn) Err() error                                            { return ec.err }
func (ec errorConn) Close() error                                          { return nil }
func (ec errorConn) Flush() error                                          { return ec.err }
func (ec errorConn) Receive() (interface{}, error)                         { return nil, ec.err }
func (ec errorConn) ReceiveWithTimeout(time.Duration) (interface{}, error) { return nil, ec.err }

var _ = Describe("RedigoService", func() {
	It("should fail loading a configuration", func() {
		var service RedigoService
		configuration, err := service.LoadConfiguration()
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("not implemented"))
		Expect(configuration).To(BeNil())
	})

	It("should fail applying configuration", func() {
		var service RedigoService
		err := service.ApplyConfiguration(map[string]interface{}{
			"address": "localhost",
		})
		Expect(err).To(Equal(http.ErrWrongConfigurationInformed))
	})

	It("should apply the configuration using a pointer", func() {
		var service RedigoService
		err := service.ApplyConfiguration(&RedigoServiceConfiguration{
			MaxIdle:     1,
			IdleTimeout: 2,
			Address:     "3",
		})
		Expect(err).To(BeNil())
		Expect(service.Configuration.Address).To(Equal("3"))
		Expect(service.Configuration.IdleTimeout).To(Equal(2))
		Expect(service.Configuration.MaxIdle).To(Equal(1))
	})

	It("should apply the configuration using a copy", func() {
		var service RedigoService
		err := service.ApplyConfiguration(RedigoServiceConfiguration{
			MaxIdle:     1,
			IdleTimeout: 2,
			Address:     "3",
		})
		Expect(err).To(BeNil())
		Expect(service.Configuration.Address).To(Equal("3"))
		Expect(service.Configuration.IdleTimeout).To(Equal(2))
		Expect(service.Configuration.MaxIdle).To(Equal(1))
	})

	It("should start the service", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(RedigoServiceConfiguration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		defer service.Stop()
		Expect(service.RunWithConn(pingConnection)).To(BeNil())
	})

	It("should stop the service", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(RedigoServiceConfiguration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		Expect(service.Stop()).To(BeNil())
		Expect(service.RunWithConn(func(conn redis.ConnWithTimeout) error {
			return nil
		})).To(Equal(http.ErrServiceNotRunning))
	})

	It("should restart the service", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(RedigoServiceConfiguration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		Expect(service.Restart()).To(BeNil())
		Expect(service.RunWithConn(pingConnection)).To(BeNil())
	})

	It("should skip the test due to time valid", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(RedigoServiceConfiguration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		Expect(service.testOnBorrow(errorConn{
			err: errors.New("this error should not show up"),
		}, time.Now().Add(-time.Minute+time.Second))).To(BeNil())
	})

	It("should test on borrow", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(RedigoServiceConfiguration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		Expect(service.RunWithConn(func(conn redis.ConnWithTimeout) error {
			Expect(service.testOnBorrow(conn, time.Now().Add(-time.Minute-time.Second))).To(BeNil())
			return nil
		})).To(BeNil())
	})

	It("should error when borrowing", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(RedigoServiceConfiguration{
			Address: "localhost:6379",
		})).To(BeNil())
		Expect(service.Start()).To(BeNil())
		err := service.testOnBorrow(errorConn{
			err: errors.New("this error should show up"),
		}, time.Now().Add(-time.Minute))
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("this error should show up"))
	})
})
