package handlers

import (
	"errors"
	"net/http"
	"net/url"
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
		Title:              "Admin Dashboard",
		Authenticated:      true,
		CurrentAdmin:       admin,
		Settings:           dashboard.Settings,
		Flags:              dashboard.Flags,
		AndroidReleases:    dashboard.AndroidReleases,
		IOSReleases:        dashboard.IOSReleases,
		MaintenanceHistory: dashboard.MaintenanceHistory,
		BannerHistory:      dashboard.BannerHistory,
		AuditLogs:          dashboard.AuditLogs,
		SystemHealth:       dashboard.SystemHealth,
		Notice:             friendlyNotice(c.Query("success")),
		Error:              friendlyError(c.Query("error")),
	})
}

func (h *AdminHandler) SystemHealthPanel(c *gin.Context) {
	pageData := PageData{
		SystemHealth: h.adminService.SystemHealth(),
	}
	if isHXRequest(c) {
		_ = h.renderer.RenderFragment(c.Writer, "system-health-panel", pageData)
		return
	}
	_ = h.renderer.RenderFragment(c.Writer, "system-health-panel", pageData)
}

func (h *AdminHandler) AuditLogsPage(c *gin.Context) {
	params, err := parseAuditLogParams(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.adminService.SearchAuditLogs(c.Request.Context(), params)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to load audit logs")
		return
	}
	admin, _ := middleware.CurrentAdmin(c)
	prevURL, nextURL := buildAuditLogPageURLs(params, result.Page, result.TotalPages)
	pageData := AuditLogPageData{
		Title:         "Audit Log",
		Authenticated: true,
		CurrentAdmin:  admin,
		Logs:          result.Logs,
		Query:         params.Query,
		Page:          result.Page,
		PageSize:      result.PageSize,
		Total:         result.Total,
		TotalPages:    result.TotalPages,
		HasPrev:       result.Page > 1,
		HasNext:       result.Page < result.TotalPages,
		PrevPage:      result.Page - 1,
		NextPage:      result.Page + 1,
		PrevURL:       prevURL,
		NextURL:       nextURL,
		Notice:        friendlyNotice(c.Query("success")),
		Error:         friendlyError(c.Query("error")),
	}
	if isHXRequest(c) {
		_ = h.renderer.RenderFragment(c.Writer, "audit-logs-table", pageData)
		return
	}
	_ = h.renderer.RenderPage(c.Writer, "audit_logs.html", pageData)
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
	if err := h.adminService.UpdateApplication(c.Request.Context(), adminMeta(c), appName, apiURL); err != nil {
		c.String(http.StatusInternalServerError, "failed to update application settings")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=application_updated")
}

func (h *AdminHandler) UpdatePlatforms(c *gin.Context) {
	if err := h.adminService.UpdatePlatforms(c.Request.Context(), adminMeta(c), parseBoolForm(c.PostForm("android_enabled")), parseBoolForm(c.PostForm("ios_enabled"))); err != nil {
		c.String(http.StatusInternalServerError, "failed to update platform settings")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=platforms_updated")
}

func (h *AdminHandler) UpdateVersion(c *gin.Context) {
	androidLatest := strings.TrimSpace(c.PostForm("android_latest_version"))
	androidMin := strings.TrimSpace(c.PostForm("android_min_version"))
	androidForce := parseBoolForm(c.PostForm("android_force_update"))
	androidNotes := strings.TrimSpace(c.PostForm("android_release_notes"))

	iosLatest := strings.TrimSpace(c.PostForm("ios_latest_version"))
	iosMin := strings.TrimSpace(c.PostForm("ios_min_version"))
	iosForce := parseBoolForm(c.PostForm("ios_force_update"))
	iosNotes := strings.TrimSpace(c.PostForm("ios_release_notes"))

	if androidLatest != "" || androidMin != "" {
		if err := h.adminService.PublishVersion(c.Request.Context(), adminMeta(c), "android", androidLatest, androidMin, androidForce, androidNotes); err != nil {
			if code := versionErrorCode(err); code != "" {
				c.Redirect(http.StatusFound, "/admin?error="+code)
				return
			}
			c.String(http.StatusInternalServerError, "failed to update android version")
			return
		}
	}
	if iosLatest != "" || iosMin != "" {
		if err := h.adminService.PublishVersion(c.Request.Context(), adminMeta(c), "ios", iosLatest, iosMin, iosForce, iosNotes); err != nil {
			if code := versionErrorCode(err); code != "" {
				c.Redirect(http.StatusFound, "/admin?error="+code)
				return
			}
			c.String(http.StatusInternalServerError, "failed to update ios version")
			return
		}
	}
	c.Redirect(http.StatusFound, "/admin?success=version_updated")
}

func (h *AdminHandler) PublishVersion(c *gin.Context) {
	platform := strings.ToLower(strings.TrimSpace(c.Param("platform")))
	latest := strings.TrimSpace(c.PostForm("latest_version"))
	minimum := strings.TrimSpace(c.PostForm("minimum_version"))
	force := parseBoolForm(c.PostForm("force_update"))
	notes := strings.TrimSpace(c.PostForm("release_notes"))
	if platform != "android" && platform != "ios" {
		c.String(http.StatusBadRequest, "invalid platform")
		return
	}
	if latest == "" || minimum == "" {
		c.Redirect(http.StatusFound, "/admin?error=version_required")
		return
	}
	if err := h.adminService.PublishVersion(c.Request.Context(), adminMeta(c), platform, latest, minimum, force, notes); err != nil {
		if errors.Is(err, services.ErrPlatformDisabled) {
			c.String(http.StatusBadRequest, "platform is disabled")
			return
		}
		if code := versionErrorCode(err); code != "" {
			c.Redirect(http.StatusFound, "/admin?error="+code)
			return
		}
		c.String(http.StatusInternalServerError, "failed to publish version")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=version_published")
}

func (h *AdminHandler) DeleteCurrentVersion(c *gin.Context) {
	platform := strings.ToLower(strings.TrimSpace(c.Param("platform")))
	if platform != "android" && platform != "ios" {
		c.String(http.StatusBadRequest, "invalid platform")
		return
	}
	if err := h.adminService.DeleteCurrentVersion(c.Request.Context(), adminMeta(c), platform); err != nil {
		if code := versionErrorCode(err); code != "" {
			c.Redirect(http.StatusFound, "/admin?error="+code)
			return
		}
		c.String(http.StatusInternalServerError, "failed to delete version")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=version_deleted")
}

func (h *AdminHandler) UpdateMaintenance(c *gin.Context) {
	enabled := parseBoolForm(c.PostForm("maintenance_mode"))
	message := strings.TrimSpace(c.PostForm("maintenance_message"))
	if err := h.adminService.UpdateMaintenance(c.Request.Context(), adminMeta(c), enabled, message); err != nil {
		c.String(http.StatusInternalServerError, "failed to update maintenance")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=maintenance_updated")
}

func (h *AdminHandler) UpdateBanner(c *gin.Context) {
	enabled := parseBoolForm(c.PostForm("banner_enabled"))
	message := strings.TrimSpace(c.PostForm("banner_message"))
	if err := h.adminService.UpdateBanner(c.Request.Context(), adminMeta(c), enabled, message); err != nil {
		c.String(http.StatusInternalServerError, "failed to update banner")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=banner_updated")
}

func (h *AdminHandler) CreateFlag(c *gin.Context) {
	key := strings.TrimSpace(c.PostForm("key"))
	if key == "" {
		c.Redirect(http.StatusFound, "/admin?error=flag_key_required")
		return
	}
	if err := h.adminService.CreateFlag(c.Request.Context(), adminMeta(c), key, false); err != nil {
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
	if err := h.adminService.UpdateFlag(c.Request.Context(), adminMeta(c), id, key, enabled); err != nil {
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
	if err := h.adminService.DeleteFlag(c.Request.Context(), adminMeta(c), id); err != nil {
		c.String(http.StatusInternalServerError, "failed to delete feature flag")
		return
	}
	c.Redirect(http.StatusFound, "/admin?success=flag_deleted")
}

func adminMeta(c *gin.Context) services.ActionMeta {
	admin, _ := middleware.CurrentAdmin(c)
	meta := services.ActionMeta{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	if admin != nil {
		meta.ActorID = admin.ID
	}
	return meta
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
	case "platforms_updated":
		return "Platform settings updated."
	case "version_updated":
		return "Version history saved."
	case "version_published":
		return "Version release created."
	case "version_deleted":
		return "Version release deleted."
	case "maintenance_updated":
		return "Maintenance history saved."
	case "banner_updated":
		return "Banner history saved."
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
	case "version_required":
		return "Version details are required."
	case "version_format_invalid":
		return "Version must follow major.minor.patch format like 1.0.1."
	case "version_not_increasing":
		return "Latest version must be greater than the current version."
	case "minimum_version_invalid":
		return "Minimum version must be less than or equal to latest version."
	case "version_delete_last":
		return "At least one version must remain."
	default:
		return ""
	}
}

func versionErrorCode(err error) string {
	switch {
	case errors.Is(err, services.ErrInvalidVersionFormat):
		return "version_format_invalid"
	case errors.Is(err, services.ErrLatestVersionNotGreater):
		return "version_not_increasing"
	case errors.Is(err, services.ErrMinimumVersionGreaterThanLatest):
		return "minimum_version_invalid"
	case errors.Is(err, services.ErrCannotDeleteLastVersion):
		return "version_delete_last"
	default:
		return ""
	}
}

func parseAuditLogParams(c *gin.Context) (services.AuditLogSearchParams, error) {
	params := services.AuditLogSearchParams{
		Query:    strings.TrimSpace(c.Query("q")),
		PageSize: 20,
	}
	if page, err := strconv.Atoi(strings.TrimSpace(c.Query("page"))); err == nil && page > 0 {
		params.Page = page
	} else {
		params.Page = 1
	}
	if pageSize, err := strconv.Atoi(strings.TrimSpace(c.Query("page_size"))); err == nil && pageSize > 0 && pageSize <= 100 {
		params.PageSize = pageSize
	}
	return params, nil
}

func buildAuditLogPageURLs(params services.AuditLogSearchParams, page, totalPages int) (string, string) {
	base := "/admin/audit-logs"
	build := func(targetPage int) string {
		values := url.Values{}
		if params.Query != "" {
			values.Set("q", params.Query)
		}
		if params.PageSize > 0 {
			values.Set("page_size", strconv.Itoa(params.PageSize))
		}
		values.Set("page", strconv.Itoa(targetPage))
		encoded := values.Encode()
		if encoded == "" {
			return base
		}
		return base + "?" + encoded
	}

	var prevURL, nextURL string
	if page > 1 {
		prevURL = build(page - 1)
	}
	if page < totalPages {
		nextURL = build(page + 1)
	}
	return prevURL, nextURL
}

func isHXRequest(c *gin.Context) bool {
	return strings.EqualFold(strings.TrimSpace(c.GetHeader("HX-Request")), "true")
}
