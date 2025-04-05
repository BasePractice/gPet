package services

import "os"

var (
	PostgresUrl = resolveEnvironment("DATABASE_URL",
		"postgres://postgres:postgres@local:25432/pet?sslmode=disable&search_path=public")
	RedisUrl = resolveEnvironment("REDIS_URL",
		"redis://localhost:6379/10?protocol=3")
	LogLevel = resolveEnvironment("LOG_LEVEL", "DEBUG")
	LogFile  = resolveEnvironment("LOG_FILE", "")
	LogColor = resolveEnvironment("LOG_COLOR", "true")
)

func resolveEnvironment(key, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return v
}
