package redigosrv

import (
	"github.com/gomodule/redigo/redis"
	"errors"
	"github.com/jamillosantos/http"
	"time"
)

// RedigoServiceConfiguration is the configuration for the `RedigoService`
type RedigoServiceConfiguration struct {
	Address     string `yaml:"address"`
	MaxIdle     int    `yaml:"max_idle"`
	IdleTimeout int    `yaml:"idle_timeout_ms"`
}

// RedigoService is the service which manages a Redis connection using the
// `redigo` library.
type RedigoService struct {
	redis.Args
	running       bool
	pool          *redis.Pool
	Configuration RedigoServiceConfiguration
}

type RedigoServiceConnHandler func(conn redis.ConnWithTimeout) error

func (service *RedigoService) LoadConfiguration() (interface{}, error) {
	return nil, errors.New("not implemented")
}

// ApplyConfiguration applies a given configuration to the service.
func (service *RedigoService) ApplyConfiguration(configuration interface{}) error {
	switch c := configuration.(type) {
	case RedigoServiceConfiguration:
		service.Configuration = c
		return nil
	case *RedigoServiceConfiguration:
		service.Configuration = *c
		return nil
	}
	return http.ErrWrongConfigurationInformed
}

// Restart stops and then starts the service again.
func (service *RedigoService) Restart() error {
	if service.running {
		err := service.Stop()
		if err != nil {
			return err
		}
	}
	return service.Start()
}

// Start starts the redis pool.
func (service *RedigoService) Start() error {
	if !service.running {
		service.pool = &redis.Pool{
			MaxIdle:      service.Configuration.MaxIdle,
			IdleTimeout:  time.Millisecond * time.Duration(service.Configuration.IdleTimeout),
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
		service.running = true
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
	if service.running {
		err := service.pool.Close()
		if err != nil {
			return err
		}
		service.running = false
	}
	return nil
}

// RunWithConn acquires the connection from a pool ensuring it will be put back
// after the handler is done.
func (service *RedigoService) RunWithConn(handler RedigoServiceConnHandler) error {
	if !service.running {
		return http.ErrServiceNotRunning
	}
	conn := service.pool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	return handler(conn.(redis.ConnWithTimeout))
}
