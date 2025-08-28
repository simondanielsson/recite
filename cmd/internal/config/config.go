package config

import (
	"log"
	"time"

	"github.com/simondanielsson/recite/pkg/env"
)

// Load loads a config.
func Load(getenv func(string) string, logger *log.Logger) (Config, error) {
	if err := env.Load(); err != nil {
		logger.Fatal("failed loading .env")
	}
	// TODO: read these values from yaml config
	cfg := Config{
		Server: ServerConfig{
			AppEnv:       "local",
			Addr:         "8999",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		DB: DBConfig{
			Name:     getenv("POSTGRES_DBNAME"),
			User:     getenv("POSTGRES_USER"),
			Password: getenv("POSTGRES_PASSWORD"),
			Host:     getenv("POSTGRES_HOST"),
			Port:     getenv("POSTGRES_PORT"),
			Driver:   getenv("DRIVER_NAME"),
		},
		JWT: JWTConfig{
			Secret: "123",
		},
	}
	return cfg, nil
}
