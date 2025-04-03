package services

import "os"

var (
	PostgresUrl = resolveEnvironment("DATABASE_URL",
		"postgres://postgres:postgres@postgres-db:5432/unknown?sslmode=disable&search_path=unknown")
)

func resolveEnvironment(key, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return v
}
