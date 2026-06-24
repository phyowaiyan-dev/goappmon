package config

import "testing"

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("GOAPPMON_ADDR", "127.0.0.1:9090")
	t.Setenv("GOAPPMON_DB_PATH", "/tmp/goappmon.db")
	t.Setenv("GOAPPMON_SESSION_KEY_PATH", "/tmp/session.key")
	t.Setenv("GOAPPMON_LOG_LEVEL", "warning")
	t.Setenv("GOAPPMON_ENV", "production")
	t.Setenv("GOAPPMON_COOKIE_NAME", "custom_session")

	cfg := Load()
	if cfg.Address != "127.0.0.1:9090" {
		t.Fatalf("unexpected address: %s", cfg.Address)
	}
	if cfg.DatabasePath != "/tmp/goappmon.db" || cfg.SessionKeyPath != "/tmp/session.key" {
		t.Fatalf("unexpected paths: %+v", cfg)
	}
	if cfg.Environment != "production" || cfg.CookieName != "custom_session" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.LogLevel.String() != "WARN" {
		t.Fatalf("unexpected log level: %s", cfg.LogLevel.String())
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("GOAPPMON_ADDR", "")
	t.Setenv("GOAPPMON_DB_PATH", "")
	t.Setenv("GOAPPMON_SESSION_KEY_PATH", "")
	t.Setenv("GOAPPMON_LOG_LEVEL", "")
	t.Setenv("GOAPPMON_ENV", "")
	t.Setenv("GOAPPMON_COOKIE_NAME", "")

	cfg := Load()
	if cfg.Address != ":18180" {
		t.Fatalf("unexpected default address: %s", cfg.Address)
	}
	if cfg.DatabasePath != "storage/goappmon.sqlite" || cfg.SessionKeyPath != "storage/session.key" {
		t.Fatalf("unexpected default paths: %+v", cfg)
	}
	if cfg.Environment != "development" || cfg.CookieName != "goappmon_session" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.LogLevel.String() != "INFO" {
		t.Fatalf("unexpected default log level: %s", cfg.LogLevel.String())
	}
}
