package config

// ServerConfig represents the configuration for the server.
type ServerConfig struct {
	BindAddr string `mapstructure:"bind_address"`
	AppEnv   string `mapstructure:"app_env"`
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
