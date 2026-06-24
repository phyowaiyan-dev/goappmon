package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phyowaiyan-dev/goappmon/internal/services"
	"github.com/phyowaiyan-dev/goappmon/internal/utils"
)

type SetupHandler struct {
	renderer     Renderer
	setupService *services.SetupService
}

func NewSetupHandler(renderer Renderer, setupService *services.SetupService) *SetupHandler {
	return &SetupHandler{renderer: renderer, setupService: setupService}
}

func (h *SetupHandler) Page(c *gin.Context) {
	complete, err := h.setupService.IsSetupComplete(c.Request.Context())
	if err == nil && complete {
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}
	_ = h.renderer.RenderPage(c.Writer, "setup.html", PageData{Title: "Initial Setup", Authenticated: false})
}

func (h *SetupHandler) Submit(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		_ = h.renderer.RenderPage(c.Writer, "setup.html", PageData{Title: "Initial Setup", Error: "Invalid setup form", Authenticated: false})
		return
	}

	adminName := strings.TrimSpace(c.PostForm("admin_name"))
	adminEmail := strings.TrimSpace(c.PostForm("admin_email"))
	password := c.PostForm("password")
	appName := strings.TrimSpace(c.PostForm("app_name"))

	if adminName == "" || adminEmail == "" || password == "" || appName == "" {
		_ = h.renderer.RenderPage(c.Writer, "setup.html", PageData{
			Title:         "Initial Setup",
			Error:         "All fields are required",
			AdminName:     adminName,
			AdminEmail:    adminEmail,
			AppName:       appName,
			Authenticated: false,
		})
		return
	}

	if err := utils.ValidateAdminName(adminName); err != nil {
		_ = h.renderer.RenderPage(c.Writer, "setup.html", PageData{
			Title:         "Initial Setup",
			Error:         err.Error(),
			AdminName:     adminName,
			AdminEmail:    adminEmail,
			AppName:       appName,
			Authenticated: false,
		})
		return
	}
	if err := utils.ValidateEmail(adminEmail); err != nil {
		_ = h.renderer.RenderPage(c.Writer, "setup.html", PageData{
			Title:         "Initial Setup",
			Error:         err.Error(),
			AdminName:     adminName,
			AdminEmail:    adminEmail,
			AppName:       appName,
			Authenticated: false,
		})
		return
	}

	if err := h.setupService.CreateInitialSetup(c.Request.Context(), adminName, adminEmail, password, appName); err != nil {
		_ = h.renderer.RenderPage(c.Writer, "setup.html", PageData{
			Title:         "Initial Setup",
			Error:         "Setup could not be completed",
			AdminName:     adminName,
			AdminEmail:    adminEmail,
			AppName:       appName,
			Authenticated: false,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/login")
}
