package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phyowaiyan-dev/goappmon/internal/services"
	"github.com/phyowaiyan-dev/goappmon/internal/utils"
)

type AuthHandler struct {
	renderer    Renderer
	authService *services.AuthService
	cookieName  string
	sessionTTL  time.Duration
}

func NewAuthHandler(renderer Renderer, authService *services.AuthService, cookieName string, sessionTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		renderer:    renderer,
		authService: authService,
		cookieName:  cookieName,
		sessionTTL:  sessionTTL,
	}
}

func (h *AuthHandler) LoginPage(c *gin.Context) {
	if h.isAuthenticated(c) {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	_ = h.renderer.RenderPage(c.Writer, "login.html", PageData{Title: "Admin Login", Authenticated: false})
}

func (h *AuthHandler) Login(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		_ = h.renderer.RenderPage(c.Writer, "login.html", PageData{Title: "Admin Login", Error: "Invalid login form", Authenticated: false})
		return
	}

	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	if err := utils.ValidateEmail(email); err != nil {
		_ = h.renderer.RenderPage(c.Writer, "login.html", PageData{Title: "Admin Login", Error: err.Error(), LoginEmail: email, Authenticated: false})
		return
	}
	admin, err := h.authService.Authenticate(c.Request.Context(), email, password)
	if err != nil {
		_ = h.renderer.RenderPage(c.Writer, "login.html", PageData{Title: "Admin Login", Error: "Invalid email or password", LoginEmail: email, Authenticated: false})
		return
	}

	token, err := h.authService.SignSession(admin.ID)
	if err != nil {
		_ = h.renderer.RenderPage(c.Writer, "login.html", PageData{Title: "Admin Login", Error: "Failed to create session", Authenticated: false})
		return
	}

	h.setSessionCookie(c, token)
	c.Redirect(http.StatusFound, "/admin")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearSessionCookie(c)
	c.Redirect(http.StatusFound, "/admin/login")
}

func (h *AuthHandler) isAuthenticated(c *gin.Context) bool {
	cookie, err := c.Cookie(h.cookieName)
	if err != nil || strings.TrimSpace(cookie) == "" {
		return false
	}
	_, err = h.authService.VerifySession(cookie)
	return err == nil
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, token string) {
	secure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
