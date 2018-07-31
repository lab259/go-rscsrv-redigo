package http_redigo_service

import (
	"github.com/gomodule/redigo/redis"
	"errors"
	"github.com/jamillosantos/http"
	"time"
)

type RedigoServiceConfiguration struct {
	Address     string `yaml:"address"`
	MaxIdle     int    `yaml:"max_idle"`
	IdleTimeout int    `yaml:"idle_timeout_ms"`
}

type RedigoService struct {
	redis.Args
	running       bool
	pool          *redis.Pool
	Configuration RedigoServiceConfiguration
}

type RedigoServiceConnHandler func(conn redis.Conn) error

func (service *RedigoService) LoadConfiguration() (interface{}, error) {
	return nil, errors.New("not implemented")
}

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

func (service *RedigoService) Restart() error {
	if service.running {
		err := service.Stop()
		if err != nil {
			return err
		}
	}
	return service.Start()
}

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

func (service *RedigoService) newConn() (redis.Conn, error) {
	return redis.Dial("tcp", service.Configuration.Address)
}

func (service *RedigoService) testOnBorrow(conn redis.Conn, lastUsage time.Time) error {
	if time.Since(lastUsage) < 1*time.Minute {
		return nil
	}
	_, err := conn.Do("PING")
	return err
}

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

func (service *RedigoService) RunWithConn(handler RedigoServiceConnHandler) error {
	if !service.running {
		return http.ErrServiceNotRunning
	}
	conn := service.pool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	return handler(conn)
}
