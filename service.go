package redigosrv

import (
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/lab259/go-rscsrv"
)

// PubSubConfiguration is the configuration for PubSub
// connections and subscriptions.
type PubSubConfiguration struct {
	ReadTimeout         time.Duration `yaml:"read_timeout"`
	WriteTimeout        time.Duration `yaml:"write_timeout"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
}

// Configuration is the configuration for the `RedigoService`.
type Configuration struct {
	Address     string              `yaml:"address"`
	MaxIdle     int                 `yaml:"max_idle"`
	IdleTimeout time.Duration       `yaml:"idle_timeout"`
	PubSub      PubSubConfiguration `yaml:"pubsub"`
}

// RedigoService is the service which manages a Redis connection using the
// `redigo` library.
type RedigoService struct {
	redis.Args
	serviceState
	pool          *redis.Pool
	Configuration Configuration
}

// ConnHandler handler redis connection with timeout
type ConnHandler func(conn redis.ConnWithTimeout) error

// LoadConfiguration loading configuration from a repository.
func (service *RedigoService) LoadConfiguration() (interface{}, error) {
	return nil, errors.New("not implemented")
}

// ApplyConfiguration applies a given configuration to the service.
func (service *RedigoService) ApplyConfiguration(configuration interface{}) error {
	switch c := configuration.(type) {
	case Configuration:
		service.Configuration = c
	case *Configuration:
		service.Configuration = *c
	default:
		return rscsrv.ErrWrongConfigurationInformed
	}

	// set defaults for pubsub if not present
	if service.Configuration.PubSub.HealthCheckInterval == 0 {
		service.Configuration.PubSub.HealthCheckInterval = time.Minute
	}
	if service.Configuration.PubSub.ReadTimeout == 0 {
		service.Configuration.PubSub.ReadTimeout = 10*time.Second + service.Configuration.PubSub.HealthCheckInterval
	}
	if service.Configuration.PubSub.WriteTimeout == 0 {
		service.Configuration.PubSub.WriteTimeout = 10 * time.Second
	}

	return nil
}

// Restart stops and then starts the service again.
func (service *RedigoService) Restart() error {
	if service.isRunning() {
		err := service.Stop()
		if err != nil {
			return err
		}
	}
	return service.Start()
}

// Start starts the redis pool.
func (service *RedigoService) Start() error {
	if !service.isRunning() {
		service.pool = &redis.Pool{
			MaxIdle:      service.Configuration.MaxIdle,
			IdleTimeout:  service.Configuration.IdleTimeout,
			Dial:         service.newConn,
			TestOnBorrow: service.testOnBorrow,
		}
		conn, err := service.pool.Dial()
		if err != nil {
			return err
		}
		defer conn.Close()
		_, err = conn.Do("PING")
		if err != nil {
			return err
		}
		service.setRunning(true)
	}
	return nil
}

// newConn is used inside of the connection pool definition to create new
// connections.
func (service *RedigoService) newConn() (redis.Conn, error) {
	return redis.Dial("tcp", service.Configuration.Address)
}

// testOnBorrow is used inside of the connection pool definition for testing
// connection before they be acquired.
func (service *RedigoService) testOnBorrow(conn redis.Conn, lastUsage time.Time) error {
	if time.Since(lastUsage) < 1*time.Minute {
		return nil
	}
	_, err := conn.Do("PING")
	return err
}

// Stop closes the connection pool.
func (service *RedigoService) Stop() error {
	if service.isRunning() {
		err := service.pool.Close()
		if err != nil {
			return err
		}
		service.setRunning(false)
	}
	return nil
}

// RunWithConn acquires the connection from a pool ensuring it will be put back
// after the handler is done.
func (service *RedigoService) RunWithConn(handler ConnHandler) error {
	if service.isRunning() {
		conn := service.pool.Get()
		if conn.Err() != nil {
			return conn.Err()
		}
		defer conn.Close()
		return handler(conn.(redis.ConnWithTimeout))
	}

	return rscsrv.ErrServiceNotRunning
}

// GetConn gets a connection from the pool.
func (service *RedigoService) GetConn() (redis.Conn, error) {
	if service.isRunning() {
		conn := service.pool.Get()
		if conn.Err() != nil {
			return nil, conn.Err()
		}
		return conn, nil
	}

	return nil, rscsrv.ErrServiceNotRunning
}
