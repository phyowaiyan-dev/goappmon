package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phyowaiyan-dev/goappmon/internal/database"
	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
	"github.com/phyowaiyan-dev/goappmon/internal/services"
	"github.com/phyowaiyan-dev/goappmon/internal/utils"
)

func TestSetupRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newMiddlewareDB(t)
	setupService := services.NewSetupService(db)

	r := gin.New()
	r.Use(SetupRedirect(setupService))
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/setup" {
		t.Fatalf("expected redirect to /setup, got %d %q", rec.Code, rec.Header().Get("Location"))
	}

	if err := setupService.CreateInitialSetup(context.Background(), "Admin", "admin@example.com", "secret123", "GoAppMon"); err != nil {
		t.Fatalf("create setup: %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Fatalf("expected ok response, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestRequireAuthAndCurrentAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newMiddlewareDB(t)
	adminRepo := repositories.NewAdminRepository(db)
	hashed, err := utils.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	adminID, err := adminRepo.Create(context.Background(), models.Admin{
		Name:         "Admin",
		Email:        "admin@example.com",
		PasswordHash: hashed,
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}

	authService := services.NewAuthService(adminRepo, []byte("01234567890123456789012345678901"), time.Hour)
	token, err := authService.SignSession(adminID)
	if err != nil {
		t.Fatalf("sign session: %v", err)
	}

	r := gin.New()
	r.Use(RequireAuth(authService, adminRepo, "goappmon_session"))
	r.GET("/admin", func(c *gin.Context) {
		admin, ok := CurrentAdmin(c)
		if !ok {
			c.String(http.StatusInternalServerError, "missing admin")
			return
		}
		c.String(http.StatusOK, admin.Email)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin/login" {
		t.Fatalf("expected redirect for missing cookie, got %d %q", rec.Code, rec.Header().Get("Location"))
	}

	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "goappmon_session", Value: token + "tamper"})
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin/login" {
		t.Fatalf("expected redirect for bad cookie, got %d %q", rec.Code, rec.Header().Get("Location"))
	}

	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "goappmon_session", Value: token})
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "admin@example.com" {
		t.Fatalf("expected authenticated response, got %d %q", rec.Code, rec.Body.String())
	}
}

func newMiddlewareDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}
