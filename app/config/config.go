package config

import (
	"fmt"
	"os"
	"sync"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Port         string
	Secret       string
	DBUrl        string
	ValkeyAddr   string
	EmailFrom    string
	PapercutHost string
	Environment  string
}

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

func Load() *Config {
	return sync.OnceValue(func() *Config {
		var config Config

		config.Environment = getEnvOrDefault("ENVIRONMENT", EnvDevelopment)
		config.Port = getEnvOrPanic("PORT")
		config.Secret = getEnvOrPanic("SECRET")
		config.EmailFrom = getEnvOrPanic("EMAIL_FROM")

		if config.Environment == EnvProduction {
			config.DBUrl = getEnvOrPanic("DB_URL")
			config.ValkeyAddr = getEnvOrPanic("VALKEY_ADDR")
			config.PapercutHost = getEnvOrPanic("PAPERCUT_SMTP_HOST")
		} else {
			config.DBUrl = getEnvOrPanic("DB_URL_LOCAL")
			config.ValkeyAddr = getEnvOrPanic("VALKEY_ADDR_LOCAL")
			config.PapercutHost = getEnvOrPanic("PAPERCUT_SMTP_HOST_LOCAL")
		}

		return &config
	})()
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
