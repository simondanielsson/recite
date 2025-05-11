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
	cfg := Config{
		Server: ServerConfig{
			AppEnv:       "local",
			Addr:         "8999",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		JWT: JWTConfig{
			Secret: "123",
		},
	}
	return cfg, nil
}
