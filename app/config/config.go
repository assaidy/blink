package config

import (
	"fmt"
	"os"
	"sync"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Environment  string
	Port         string
	Secret       string
	PostgresUrl  string
	ValkeyAddr   string
	EmailFrom    string
	PapercutHost string
}

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

var (
	config *Config
	once   sync.Once
)

func Load() *Config {
	once.Do(func() {
		config = &Config{}
		config.Environment = getEnvOrDefault("ENVIRONMENT", EnvDevelopment)
		config.Port = getEnvOrPanic("PORT")
		config.Secret = getEnvOrPanic("SECRET")
		config.EmailFrom = getEnvOrPanic("EMAIL_FROM")

		if config.Environment == EnvProduction {
			config.PostgresUrl = getEnvOrPanic("POSTGRES_URL")
			config.ValkeyAddr = getEnvOrPanic("VALKEY_ADDR")
			config.PapercutHost = getEnvOrPanic("PAPERCUT_SMTP_HOST")
		} else {
			config.PostgresUrl = getEnvOrPanic("POSTGRES_URL_LOCAL")
			config.ValkeyAddr = getEnvOrPanic("VALKEY_ADDR_LOCAL")
			config.PapercutHost = getEnvOrPanic("PAPERCUT_SMTP_HOST_LOCAL")
		}
	})

	return config
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getEnvOrPanic(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic(fmt.Sprintf("missing required env var: `%s`", key))
}
