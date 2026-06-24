package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phyowaiyan-dev/goappmon/internal/services"
	"github.com/phyowaiyan-dev/goappmon/internal/utils"
)

type PublicHandler struct {
	statusService *services.StatusService
}

func NewPublicHandler(statusService *services.StatusService) *PublicHandler {
	return &PublicHandler{statusService: statusService}
}

func (h *PublicHandler) Health(c *gin.Context) {
	utils.JSON(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *PublicHandler) Status(c *gin.Context) {
	status, err := h.statusService.PublicStatus(c.Request.Context())
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "failed to load status")
		return
	}
	utils.JSON(c, http.StatusOK, status)
}

func (h *PublicHandler) Version(c *gin.Context) {
	version, err := h.statusService.PublicVersion(c.Request.Context())
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "failed to load version")
		return
	}
	utils.JSON(c, http.StatusOK, version)
}

func (h *PublicHandler) Config(c *gin.Context) {
	cfg, err := h.statusService.PublicConfig(c.Request.Context())
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "failed to load config")
		return
	}
	utils.JSON(c, http.StatusOK, cfg)
}

func (h *PublicHandler) FeatureFlags(c *gin.Context) {
	flags, err := h.statusService.PublicFeatureFlags(c.Request.Context())
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "failed to load feature flags")
		return
	}
	utils.JSON(c, http.StatusOK, flags)
}
