package config

import "time"

// ServerConfig represents the configuration for the server.
type ServerConfig struct {
	AppEnv       string        `mapstructure:"app_env"`
	Addr         string        `mapstructure:"address"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DBConfig struct {
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Driver   string `mapstructure:"driver"`
}

// JWTConfig represents the configuration of the JWT secret
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// Config represents a yaml configuration file.
type Config struct {
	Server ServerConfig
	DB     DBConfig
	JWT    JWTConfig
}
