package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFullHTTPFlow(t *testing.T) {
	app := newTestApp(t)

	// Setup wizard page and validation.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Create the first admin account") {
		t.Fatalf("unexpected setup page response: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader("admin_name=&admin_email=&password=&app_name="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "All fields are required") {
		t.Fatalf("expected setup validation error, got %d %q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	form := url.Values{}
	form.Set("admin_name", "Admin")
	form.Set("admin_email", "admin@example.com")
	form.Set("password", "secret123")
	form.Set("app_name", "GoAppMon")
	req = httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin/login" {
		t.Fatalf("expected setup redirect, got %d %q", rec.Code, rec.Header().Get("Location"))
	}

	// Login page and authentication.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Sign in to the dashboard") {
		t.Fatalf("unexpected login page response: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	form = url.Values{}
	form.Set("email", "admin@example.com")
	form.Set("password", "wrong")
	req = httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Invalid email or password") {
		t.Fatalf("expected login error, got %d %q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	form.Set("password", "secret123")
	req = httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin" {
		t.Fatalf("expected login redirect, got %d %q", rec.Code, rec.Header().Get("Location"))
	}

	sessionCookie := rec.Result().Cookies()[0]

	// Authenticated dashboard.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(sessionCookie)
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Admin Dashboard") {
		t.Fatalf("unexpected dashboard response: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/postman-collection", nil)
	req.AddCookie(sessionCookie)
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected postman download response, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, "GoAppMon.postman_collection.json") {
		t.Fatalf("expected postman attachment header, got %q", got)
	}
	if !strings.Contains(rec.Body.String(), "\"Health\"") || !strings.Contains(rec.Body.String(), "\"Admin Login\"") {
		t.Fatalf("expected postman collection payload, got %q", rec.Body.String())
	}

	// Login page should redirect when already authenticated.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	req.AddCookie(sessionCookie)
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin" {
		t.Fatalf("expected login redirect for authenticated user, got %d %q", rec.Code, rec.Header().Get("Location"))
	}

	// Admin updates and feature flag CRUD.
	post := func(path string, values url.Values) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(values.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(sessionCookie)
		app.router.ServeHTTP(rec, req)
		return rec
	}

	if rec := post("/admin/settings/application", url.Values{"app_name": {"GoAppMon Plus"}, "api_url": {"https://api.example.com"}}); rec.Code != http.StatusFound {
		t.Fatalf("expected application update redirect, got %d", rec.Code)
	}
	if rec := post("/admin/settings/version", url.Values{
		"android_latest_version": {"2.0.0"},
		"android_min_version":    {"1.5.0"},
		"android_force_update":   {"true"},
		"ios_latest_version":     {"2.1.0"},
		"ios_min_version":        {"1.6.0"},
		"ios_force_update":       {"false"},
	}); rec.Code != http.StatusFound {
		t.Fatalf("expected version update redirect, got %d", rec.Code)
	}
	if rec := post("/admin/settings/maintenance", url.Values{"maintenance_mode": {"true"}, "maintenance_message": {"maintenance"}}); rec.Code != http.StatusFound {
		t.Fatalf("expected maintenance redirect, got %d", rec.Code)
	}
	if rec := post("/admin/settings/banner", url.Values{"banner_enabled": {"true"}, "banner_message": {"banner"}}); rec.Code != http.StatusFound {
		t.Fatalf("expected banner redirect, got %d", rec.Code)
	}
	if rec := post("/admin/feature-flags", url.Values{"key": {"chat"}, "enabled": {"true"}}); rec.Code != http.StatusFound {
		t.Fatalf("expected create flag redirect, got %d", rec.Code)
	}
	if rec := post("/admin/feature-flags", url.Values{"key": {"payment"}, "enabled": {"false"}}); rec.Code != http.StatusFound {
		t.Fatalf("expected create second flag redirect, got %d", rec.Code)
	}

	// Verify dashboard shows success notice and updated content.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin?success=application_updated", nil)
	req.AddCookie(sessionCookie)
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Application settings updated.") {
		t.Fatalf("expected dashboard notice, got %d %q", rec.Code, rec.Body.String())
	}

	// Feature flag update/delete.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/feature-flags/1", strings.NewReader(url.Values{"key": {"chat-v2"}, "enabled": {"false"}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected update flag redirect, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/feature-flags/1/delete", nil)
	req.AddCookie(sessionCookie)
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected delete flag redirect, got %d", rec.Code)
	}

	// Logout clears session.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	req.AddCookie(sessionCookie)
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin/login" {
		t.Fatalf("expected logout redirect, got %d %q", rec.Code, rec.Header().Get("Location"))
	}
}
