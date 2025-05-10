package config

import "time"

// Load loads a config.
func Load(getenv func(string) string) (Config, error) {
	// TODO: load environment variables into the config using viper
	return Config{
		Server: ServerConfig{
			AppEnv:       "local",
			Addr:         "8999",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		JWT: JWTConfig{
			Secret: "123",
		},
	}, nil
}
