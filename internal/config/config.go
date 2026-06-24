package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Address         string
	DatabasePath    string
	SessionKeyPath  string
	LogLevel        slog.Level
	Environment     string
	CookieName      string
	SessionDuration int64
}

func Load() Config {
	return Config{
		Address:         getEnv("GOAPPMON_ADDR", ":8080"),
		DatabasePath:    getEnv("GOAPPMON_DB_PATH", "storage/goappmon.sqlite"),
		SessionKeyPath:  getEnv("GOAPPMON_SESSION_KEY_PATH", "storage/session.key"),
		LogLevel:        parseLogLevel(getEnv("GOAPPMON_LOG_LEVEL", "info")),
		Environment:     getEnv("GOAPPMON_ENV", "development"),
		CookieName:      getEnv("GOAPPMON_COOKIE_NAME", "goappmon_session"),
		SessionDuration: 60 * 60 * 24 * 7,
	}
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
