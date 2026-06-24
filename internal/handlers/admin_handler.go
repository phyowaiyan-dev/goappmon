package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phyowaiyan-dev/goappmon/internal/middleware"
	"github.com/phyowaiyan-dev/goappmon/internal/services"
	"github.com/phyowaiyan-dev/goappmon/web"
)

type AdminHandler struct {
	renderer     Renderer
	adminService *services.AdminService
}

func NewAdminHandler(renderer Renderer, adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{renderer: renderer, adminService: adminService}
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	dashboard, err := h.adminService.Dashboard(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to load dashboard")
		return
	}
	admin, _ := middleware.CurrentAdmin(c)
	_ = h.renderer.RenderPage(c.Writer, "dashboard.html", PageData{
		Title:        "Admin Dashboard",
		CurrentAdmin: admin,
		Settings:     dashboard.Settings,
		Flags:        dashboard.Flags,
		SystemHealth: dashboard.SystemHealth,
		Notice:       friendlyNotice(c.Query("success")),
		Error:        friendlyError(c.Query("error")),
	})
}

func (h *AdminHandler) DownloadPostmanCollection(c *gin.Context) {
	data, err := web.PostmanFS.ReadFile("postman/GoAppMon.postman_collection.json")
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to load postman collection")
		return
	}
	c.Header("Content-Disposition", `attachment; filename="GoAppMon.postman_collection.json"`)
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

func (h *AdminHandler) UpdateApplication(c *gin.Context) {
	appName := strings.TrimSpace(c.PostForm("app_name"))
	apiURL := strings.TrimSpace(c.PostForm("api_url"))
	if appName == "" {
		c.Redirect(http.StatusFound, "/admin?error=app_name_required")
		return
	}
	if err := h.adminService.UpdateApplication(c.Request.Context(), appName, apiURL); err != nil {
		c.String(http.StatusInternalServerError, "failed to update application settings")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=application_updated")
}

func (h *AdminHandler) UpdateVersion(c *gin.Context) {
	androidLatest := strings.TrimSpace(c.PostForm("android_latest_version"))
	androidMin := strings.TrimSpace(c.PostForm("android_min_version"))
	iosLatest := strings.TrimSpace(c.PostForm("ios_latest_version"))
	iosMin := strings.TrimSpace(c.PostForm("ios_min_version"))
	androidForce := parseBoolForm(c.PostForm("android_force_update"))
	iosForce := parseBoolForm(c.PostForm("ios_force_update"))
	if err := h.adminService.UpdateVersion(c.Request.Context(), androidLatest, androidMin, androidForce, iosLatest, iosMin, iosForce); err != nil {
		c.String(http.StatusInternalServerError, "failed to update versions")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=version_updated")
}

func (h *AdminHandler) UpdateMaintenance(c *gin.Context) {
	enabled := parseBoolForm(c.PostForm("maintenance_mode"))
	message := strings.TrimSpace(c.PostForm("maintenance_message"))
	if err := h.adminService.UpdateMaintenance(c.Request.Context(), enabled, message); err != nil {
		c.String(http.StatusInternalServerError, "failed to update maintenance")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=maintenance_updated")
}

func (h *AdminHandler) UpdateBanner(c *gin.Context) {
	enabled := parseBoolForm(c.PostForm("banner_enabled"))
	message := strings.TrimSpace(c.PostForm("banner_message"))
	if err := h.adminService.UpdateBanner(c.Request.Context(), enabled, message); err != nil {
		c.String(http.StatusInternalServerError, "failed to update banner")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=banner_updated")
}

func (h *AdminHandler) CreateFlag(c *gin.Context) {
	key := strings.TrimSpace(c.PostForm("key"))
	enabled := parseBoolForm(c.PostForm("enabled"))
	if key == "" {
		c.Redirect(http.StatusFound, "/admin?error=flag_key_required")
		return
	}
	if err := h.adminService.CreateFlag(c.Request.Context(), key, enabled); err != nil {
		c.String(http.StatusInternalServerError, "failed to create feature flag")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=flag_created")
}

func (h *AdminHandler) UpdateFlag(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid flag id")
		return
	}
	key := strings.TrimSpace(c.PostForm("key"))
	enabled := parseBoolForm(c.PostForm("enabled"))
	if key == "" {
		c.Redirect(http.StatusFound, "/admin?error=flag_key_required")
		return
	}
	if err := h.adminService.UpdateFlag(c.Request.Context(), id, key, enabled); err != nil {
		c.String(http.StatusInternalServerError, "failed to update feature flag")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=flag_updated")
}

func (h *AdminHandler) DeleteFlag(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid flag id")
		return
	}
	if err := h.adminService.DeleteFlag(c.Request.Context(), id); err != nil {
		c.String(http.StatusInternalServerError, "failed to delete feature flag")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=flag_deleted")
}

func parseBoolForm(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "on", "yes", "enabled":
		return true
	default:
		return false
	}
}

func friendlyNotice(code string) string {
	switch code {
	case "application_updated":
		return "Application settings updated."
	case "version_updated":
		return "Version settings updated."
	case "maintenance_updated":
		return "Maintenance settings updated."
	case "banner_updated":
		return "Banner settings updated."
	case "flag_created":
		return "Feature flag created."
	case "flag_updated":
		return "Feature flag updated."
	case "flag_deleted":
		return "Feature flag deleted."
	default:
		return ""
	}
}

func friendlyError(code string) string {
	switch code {
	case "app_name_required":
		return "App name is required."
	case "flag_key_required":
		return "Feature flag key is required."
	default:
		return ""
	}
}
