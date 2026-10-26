package config

import (
	"errors"
	"net"
	"net/url"
	"os"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func Load() (Config, error) {
	return FromEnv(os.Getenv)
}

func FromEnv(getenv func(string) string) (Config, error) {
	cfg := Config{HTTPAddr: getenv("HTTP_ADDR"), DatabaseURL: getenv("DATABASE_URL")}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	if _, _, err := net.SplitHostPort(cfg.HTTPAddr); err != nil {
		return Config{}, errors.New("HTTP_ADDR must contain a host and port")
	}
	u, err := url.Parse(cfg.DatabaseURL)
	if err != nil || u == nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" || u.Path == "" || u.Path == "/" {
		// Configuration errors must never echo a connection string or password.
		return Config{}, errors.New("DATABASE_URL must be a PostgreSQL URL with host and database")
	}
	return cfg, nil
}
