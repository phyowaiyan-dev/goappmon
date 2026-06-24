package app

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phyowaiyan-dev/goappmon/internal/config"
	"github.com/phyowaiyan-dev/goappmon/internal/database"
	"github.com/phyowaiyan-dev/goappmon/internal/handlers"
	"github.com/phyowaiyan-dev/goappmon/internal/middleware"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
	"github.com/phyowaiyan-dev/goappmon/internal/services"
	"github.com/phyowaiyan-dev/goappmon/web"
)

type App struct {
	cfg       config.Config
	logger    *slog.Logger
	startedAt time.Time
	db        *sql.DB
	router    *gin.Engine
	server    *http.Server

	renderer *templateRenderer

	setupService  *services.SetupService
	authService   *services.AuthService
	statusService *services.StatusService
	adminService  *services.AdminService
	adminRepo     *repositories.AdminRepository
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.SessionKeyPath), 0o755); err != nil {
		return nil, err
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := database.Migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	sessionSecret, err := loadSessionSecret(cfg.SessionKeyPath)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	adminRepo := repositories.NewAdminRepository(db)
	settingRepo := repositories.NewSettingRepository(db)
	flagRepo := repositories.NewFeatureFlagRepository(db)
	setupService := services.NewSetupService(db)
	if err := setupService.EnsureDefaultSettings(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	authService := services.NewAuthService(adminRepo, sessionSecret, time.Duration(cfg.SessionDuration)*time.Second)
	statusService := services.NewStatusService(settingRepo, flagRepo)
	startedAt := time.Now().UTC()
	adminService := services.NewAdminService(db, settingRepo, flagRepo, cfg.DatabasePath, startedAt)
	renderer, err := newTemplateRenderer()
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(logger))
	router.Use(middleware.SetupRedirect(setupService))

	app := &App{
		cfg:           cfg,
		logger:        logger,
		startedAt:     startedAt,
		db:            db,
		router:        router,
		renderer:      renderer,
		setupService:  setupService,
		authService:   authService,
		statusService: statusService,
		adminService:  adminService,
		adminRepo:     adminRepo,
	}
	app.registerRoutes()
	app.server = &http.Server{
		Addr:              cfg.Address,
		Handler:           app.router,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("starting server", "addr", a.cfg.Address)
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return a.db.Close()
	case err := <-errCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return a.db.Close()
		}
		_ = a.db.Close()
		return err
	}
}

func (a *App) RenderPage(w http.ResponseWriter, page string, data any) error {
	return a.renderer.RenderPage(w, page, data)
}

func (a *App) registerRoutes() {
	publicHandler := handlers.NewPublicHandler(a.statusService)
	setupHandler := handlers.NewSetupHandler(a.renderer, a.setupService)
	authHandler := handlers.NewAuthHandler(a.renderer, a.authService, a.cfg.CookieName, time.Duration(a.cfg.SessionDuration)*time.Second)
	adminHandler := handlers.NewAdminHandler(a.renderer, a.adminService)

	a.router.GET("/", func(c *gin.Context) {
		if cookie, err := c.Cookie(a.cfg.CookieName); err == nil && strings.TrimSpace(cookie) != "" {
			c.Redirect(http.StatusFound, "/admin")
			return
		}
		c.Redirect(http.StatusFound, "/admin/login")
	})

	a.router.GET("/health", publicHandler.Health)
	a.router.GET("/api/status", publicHandler.Status)
	a.router.GET("/api/version", publicHandler.Version)
	a.router.GET("/api/config", publicHandler.Config)
	a.router.GET("/api/feature-flags", publicHandler.FeatureFlags)

	a.router.GET("/setup", setupHandler.Page)
	a.router.POST("/setup", setupHandler.Submit)

	a.router.GET("/admin/login", authHandler.LoginPage)
	a.router.POST("/admin/login", authHandler.Login)
	a.router.POST("/admin/logout", authHandler.Logout)

	adminGroup := a.router.Group("/admin")
	adminGroup.Use(middleware.RequireAuth(a.authService, a.adminRepo, a.cfg.CookieName))
	adminGroup.GET("", adminHandler.Dashboard)
	adminGroup.GET("/system-health", adminHandler.SystemHealthPanel)
	adminGroup.GET("/postman-collection", adminHandler.DownloadPostmanCollection)
	adminGroup.GET("/audit-logs", adminHandler.AuditLogsPage)
	adminGroup.POST("/platforms", adminHandler.UpdatePlatforms)
	adminGroup.POST("/settings/application", adminHandler.UpdateApplication)
	adminGroup.POST("/settings/version", adminHandler.UpdateVersion)
	adminGroup.POST("/version/:platform", adminHandler.PublishVersion)
	adminGroup.POST("/version/:platform/delete", adminHandler.DeleteCurrentVersion)
	adminGroup.POST("/settings/maintenance", adminHandler.UpdateMaintenance)
	adminGroup.POST("/settings/banner", adminHandler.UpdateBanner)
	adminGroup.POST("/feature-flags", adminHandler.CreateFlag)
	adminGroup.POST("/feature-flags/:id", adminHandler.UpdateFlag)
	adminGroup.POST("/feature-flags/:id/delete", adminHandler.DeleteFlag)
}

func loadSessionSecret(path string) ([]byte, error) {
	if data, err := os.ReadFile(path); err == nil {
		decoded, decodeErr := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
		if decodeErr == nil && len(decoded) >= 32 {
			return decoded, nil
		}
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(secret)), 0o600); err != nil {
		return nil, err
	}
	return secret, nil
}

func loggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
			"ip", c.ClientIP(),
		)
	}
}

type templateRenderer struct {
	baseTemplates     *template.Template
	fragmentTemplates *template.Template
}

func newTemplateRenderer() (*templateRenderer, error) {
	funcMap := templateFuncs()
	baseTemplates, err := template.New("base").Funcs(funcMap).ParseFS(web.TemplatesFS,
		"templates/layouts/*.html",
		"templates/components/*.html",
	)
	if err != nil {
		return nil, err
	}
	fragmentTemplates, err := template.New("fragments").Funcs(funcMap).ParseFS(web.TemplatesFS,
		"templates/components/*.html",
	)
	if err != nil {
		return nil, err
	}
	return &templateRenderer{
		baseTemplates:     baseTemplates,
		fragmentTemplates: fragmentTemplates,
	}, nil
}

func (r *templateRenderer) RenderPage(w http.ResponseWriter, page string, data any) error {
	tpl, err := r.baseTemplates.Clone()
	if err != nil {
		return fmt.Errorf("clone base templates: %w", err)
	}
	if _, err := tpl.ParseFS(web.TemplatesFS, "templates/pages/"+page); err != nil {
		return fmt.Errorf("parse template %s: %w", page, err)
	}
	if err := tpl.ExecuteTemplate(w, "base", data); err != nil {
		return err
	}
	return nil
}

func (r *templateRenderer) RenderFragment(w http.ResponseWriter, fragment string, data any) error {
	tpl, err := r.fragmentTemplates.Clone()
	if err != nil {
		return fmt.Errorf("clone fragment templates: %w", err)
	}
	if err := tpl.ExecuteTemplate(w, fragment, data); err != nil {
		return err
	}
	return nil
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"humanTime": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			now := time.Now()
			localTime := t.In(now.Location())
			now = now.In(now.Location())

			if localTime.After(now) {
				return localTime.Format("Jan 2, 2006 3:04 PM")
			}

			diff := now.Sub(localTime)
			switch {
			case diff < time.Minute:
				return "just now"
			case diff < time.Hour:
				minutes := int(diff.Minutes())
				if minutes == 1 {
					return "1 minute ago"
				}
				return fmt.Sprintf("%d minutes ago", minutes)
			case diff < 24*time.Hour && sameDay(now, localTime):
				hours := int(diff.Hours())
				if hours == 1 {
					return "1 hour ago"
				}
				return fmt.Sprintf("%d hours ago", hours)
			case diff < 48*time.Hour:
				return "Yesterday at " + localTime.Format("3:04 PM")
			case diff < 7*24*time.Hour:
				return localTime.Format("Mon at 3:04 PM")
			default:
				return localTime.Format("Jan 2, 2006 3:04 PM")
			}
		},
		"appName": func() string {
			return config.AppName
		},
		"appVersion": func() string {
			return config.AppVersion
		},
		"yearNow": func() int {
			return time.Now().Year()
		},
		"boolLabel": func(v bool) string {
			if v {
				return "Enabled"
			}
			return "Disabled"
		},
		"usageTone": func(v float64) string {
			switch {
			case v >= 85:
				return "error"
			case v >= 70:
				return "warning"
			default:
				return "success"
			}
		},
	}
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
