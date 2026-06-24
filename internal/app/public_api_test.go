package app

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/phyowaiyan-dev/goappmon/internal/config"
)

func newTestApp(t *testing.T) *App {
	t.Helper()

	dir := t.TempDir()
	cfg := config.Config{
		Address:         ":0",
		DatabasePath:    filepath.Join(dir, "goappmon.sqlite"),
		SessionKeyPath:  filepath.Join(dir, "session.key"),
		LogLevel:        slog.LevelError,
		Environment:     "test",
		CookieName:      "goappmon_session",
		SessionDuration: 60 * 60,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	app, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() {
		_ = app.db.Close()
	})
	return app
}

func TestPublicRoutesRedirectBeforeSetup(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	app.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rr.Code)
	}
	if location := rr.Header().Get("Location"); location != "/setup" {
		t.Fatalf("expected redirect to /setup, got %q", location)
	}
}

func TestPublicAPIsAfterSetup(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.setupService.CreateInitialSetup(ctx, "Admin", "admin@example.com", "secret123", "GoAppMon"); err != nil {
		t.Fatalf("create setup: %v", err)
	}
	if err := app.adminService.UpdateApplication(ctx, "GoAppMon Plus", "https://api.example.com"); err != nil {
		t.Fatalf("update application: %v", err)
	}
	if err := app.adminService.UpdateVersion(ctx, "2.0.0", "1.5.0", true, "2.1.0", "1.6.0", false); err != nil {
		t.Fatalf("update version: %v", err)
	}
	if err := app.adminService.UpdateMaintenance(ctx, true, "maintenance"); err != nil {
		t.Fatalf("update maintenance: %v", err)
	}
	if err := app.adminService.UpdateBanner(ctx, true, "banner"); err != nil {
		t.Fatalf("update banner: %v", err)
	}
	if err := app.adminService.CreateFlag(ctx, "chat", true); err != nil {
		t.Fatalf("create flag: %v", err)
	}
	if err := app.adminService.CreateFlag(ctx, "payment", false); err != nil {
		t.Fatalf("create flag: %v", err)
	}

	tests := []struct {
		name string
		path string
		want map[string]any
	}{
		{name: "health", path: "/health", want: map[string]any{"status": "ok"}},
		{name: "status", path: "/api/status", want: map[string]any{"maintenance_mode": true, "maintenance_message": "maintenance", "banner_enabled": true, "banner_message": "banner"}},
		{name: "version", path: "/api/version", want: map[string]any{"android": map[string]any{"latest_version": "2.0.0", "minimum_version": "1.5.0", "force_update": true}, "ios": map[string]any{"latest_version": "2.1.0", "minimum_version": "1.6.0", "force_update": false}}},
		{name: "config", path: "/api/config", want: map[string]any{"app_name": "GoAppMon Plus", "api_url": "https://api.example.com"}},
		{name: "flags", path: "/api/feature-flags", want: map[string]any{"chat": true, "payment": false}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()
			app.router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rr.Code)
			}

			var got map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}

			assertJSONSubset(t, got, tc.want)
		})
	}
}

func assertJSONSubset(t *testing.T, got, want map[string]any) {
	t.Helper()
	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Fatalf("missing key %q in %#v", key, got)
		}
		switch wantTyped := wantValue.(type) {
		case map[string]any:
			gotTyped, ok := gotValue.(map[string]any)
			if !ok {
				t.Fatalf("key %q expected object, got %#v", key, gotValue)
			}
			assertJSONSubset(t, gotTyped, wantTyped)
		case bool:
			gotBool, ok := gotValue.(bool)
			if !ok || gotBool != wantTyped {
				t.Fatalf("key %q expected %v, got %#v", key, wantTyped, gotValue)
			}
		case string:
			gotString, ok := gotValue.(string)
			if !ok || gotString != wantTyped {
				t.Fatalf("key %q expected %q, got %#v", key, wantTyped, gotValue)
			}
		default:
			t.Fatalf("unsupported type %T", wantValue)
		}
	}
}
