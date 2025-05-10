package config

import "time"

// ServerConfig represents the configuration for the server.
type ServerConfig struct {
	AppEnv       string        `mapstructure:"app_env"`
	Addr         string        `mapstructure:"address"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// JWTConfig represents the configuration of the JWT secret
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// Config represents a yaml configuration file.
type Config struct {
	Server ServerConfig
	JWT    JWTConfig
}
