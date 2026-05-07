package config

import (
	"errors"
	"fmt"
	"os"
)

type PGConfig interface {
	DSN() string
}

type pgConfig struct {
	DbName string
	Host   string
	User   string
	Pass   string
	Port   string
}

func NewPGConfig() (PGConfig, error) {
	dbname := os.Getenv("PG_DATABASE")
	if len(dbname) == 0 {
		return nil, errors.New("PG_DATABASE environment variable not set")
	}

	host := os.Getenv("PG_HOST")
	if len(host) == 0 {
		return nil, errors.New("PG_HOST environment variable not set")
	}

	port := os.Getenv("PG_PORT")
	if len(port) == 0 {
		return nil, errors.New("PG_PORT environment variable not set")
	}

	user := os.Getenv("PG_USER")
	if len(user) == 0 {
		return nil, errors.New("PG_USER environment variable not set")
	}

	pass := os.Getenv("PG_PASSWORD")
	if len(pass) == 0 {
		return nil, errors.New("PG_PASS environment variable not set")
	}

	return &pgConfig{
		DbName: dbname,
		Host:   host,
		User:   user,
		Pass:   pass,
		Port:   port,
	}, nil
}

func (c *pgConfig) DSN() string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v", c.User, c.Pass, c.Host, c.Port, c.DbName)
}
