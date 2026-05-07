package config

import (
	"errors"
	"os"
)

type HTTPConfig interface {
	Address() string
}

type httpConfig struct {
	host string
	port string
}

func NewHTTPConfig() (HTTPConfig, error) {
	host := os.Getenv("HTTP_HOST")
	if len(host) == 0 {
		return nil, errors.New("HTTP environment variable HTTP_HOST is not set")
	}

	port := os.Getenv("HTTP_PORT")
	if len(port) == 0 {
		return nil, errors.New("HTTP environment variable HTTP_PORT is not set")
	}

	return &httpConfig{
		host: host,
		port: port,
	}, nil
}

func (c *httpConfig) Address() string {
	return c.host + ":" + c.port
}
