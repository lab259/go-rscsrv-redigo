package redigosrv

import (
	"errors"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jamillosantos/macchiato"
	"github.com/lab259/go-rscsrv"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
	log.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	description := "Redigo Test Suite"
	if os.Getenv("CI") == "" {
		macchiato.RunSpecs(t, description)
	} else {
		reporterOutputDir := "./test-results/go-rscsrv-redigo"
		os.MkdirAll(reporterOutputDir, os.ModePerm)
		junitReporter := reporters.NewJUnitReporter(path.Join(reporterOutputDir, "results.xml"))
		macchiatoReporter := macchiato.NewReporter()
		RunSpecsWithCustomReporters(t, description, []ginkgo.Reporter{macchiatoReporter, junitReporter})
	}
}

func pingConnection(conn redis.ConnWithTimeout) error {
	Expect(conn.Err()).ToNot(HaveOccurred())
	_, err := conn.Do("PING")
	Expect(err).ToNot(HaveOccurred())
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
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not implemented"))
		Expect(configuration).To(BeNil())
	})

	It("should fail applying configuration", func() {
		var service RedigoService
		err := service.ApplyConfiguration(map[string]interface{}{
			"address": "localhost",
		})
		Expect(err).To(Equal(rscsrv.ErrWrongConfigurationInformed))
	})

	It("should apply the configuration using a pointer", func() {
		var service RedigoService
		err := service.ApplyConfiguration(&Configuration{
			MaxIdle:     1,
			IdleTimeout: 2 * time.Second,
			Address:     "3",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(service.Configuration.Address).To(Equal("3"))
		Expect(service.Configuration.IdleTimeout).To(Equal(2 * time.Second))
		Expect(service.Configuration.MaxIdle).To(Equal(1))
	})

	It("should apply the configuration using a copy", func() {
		var service RedigoService
		err := service.ApplyConfiguration(Configuration{
			MaxIdle:     1,
			IdleTimeout: 2 * time.Second,
			Address:     "3",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(service.Configuration.Address).To(Equal("3"))
		Expect(service.Configuration.IdleTimeout).To(Equal(2 * time.Second))
		Expect(service.Configuration.MaxIdle).To(Equal(1))
	})

	It("should start the service", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(Succeed())
		Expect(service.Start()).To(Succeed())
		defer service.Stop()
		Expect(service.RunWithConn(pingConnection)).To(Succeed())
	})

	It("should stop the service", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(Succeed())
		Expect(service.Start()).To(Succeed())
		Expect(service.Stop()).To(Succeed())
		Expect(service.RunWithConn(func(conn redis.ConnWithTimeout) error {
			return nil
		})).To(Equal(rscsrv.ErrServiceNotRunning))
	})

	It("should restart the service", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(Succeed())
		Expect(service.Start()).To(Succeed())
		Expect(service.Restart()).To(Succeed())
		Expect(service.RunWithConn(pingConnection)).To(Succeed())
	})

	It("should skip the test due to time valid", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(Succeed())
		Expect(service.Start()).To(Succeed())
		Expect(service.testOnBorrow(errorConn{
			err: errors.New("this error should not show up"),
		}, time.Now().Add(-time.Minute+time.Second))).To(Succeed())
	})

	It("should test on borrow", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(Succeed())
		Expect(service.Start()).To(Succeed())
		Expect(service.RunWithConn(func(conn redis.ConnWithTimeout) error {
			Expect(service.testOnBorrow(conn, time.Now().Add(-time.Minute-time.Second))).To(Succeed())
			return nil
		})).To(Succeed())
	})

	It("should error when borrowing", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(Succeed())
		Expect(service.Start()).To(Succeed())
		err := service.testOnBorrow(errorConn{
			err: errors.New("this error should show up"),
		}, time.Now().Add(-time.Minute))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("this error should show up"))
	})

	It("should get a connection from the pool", func() {
		var service RedigoService
		Expect(service.ApplyConfiguration(Configuration{
			Address: "localhost:6379",
		})).To(Succeed())
		Expect(service.Start()).To(Succeed())
		conn, err := service.GetConn()
		Expect(err).ToNot(HaveOccurred())
		_, err = conn.Do("PING")
		Expect(err).ToNot(HaveOccurred())
	})
})
