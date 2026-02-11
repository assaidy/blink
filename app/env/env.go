package env

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

var (
	Environment = getEnvOrAlt("ENVIRONMENT", EnvDevelopment)
	Port        = getEnvOrPanic("PORT")
	Secret      = getEnvOrPanic("SECRET")

	DBUrl      = getEnvOrPanic(ifElse(Environment == EnvProduction, "DB_URL", "DB_URL_LOCAL"))
	ValkeyAddr = getEnvOrPanic(ifElse(Environment == EnvProduction, "VALKEY_ADDR", "VALKEY_ADDR_LOCAL"))

	EmailFrom       = getEnvOrPanic("EMAIL_FROM")
	PapercutSmtHost = getEnvOrPanic(ifElse(Environment == EnvProduction, "PAPERCUT_SMTP_HOST", "PAPERCUT_SMTP_HOST_LOCAL"))
)

func getEnvOrAlt(key, alt string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return alt
}

func getEnvOrPanic(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic(fmt.Sprintf("missing env var: `%s`", key))
}

func ifElse[T any](condition bool, ifTrue, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}
