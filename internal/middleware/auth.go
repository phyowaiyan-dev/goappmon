package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
	"github.com/phyowaiyan-dev/goappmon/internal/services"
)

const CurrentAdminKey = "current_admin"

func SetupRedirect(setupService *services.SetupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if bypassSetupRedirect(c.Request.URL.Path) {
			c.Next()
			return
		}

		complete, err := setupService.IsSetupComplete(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to check setup status"})
			return
		}
		if !complete {
			c.Redirect(http.StatusFound, "/setup")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAuth(authService *services.AuthService, adminRepo *repositories.AdminRepository, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(cookieName)
		if err != nil || strings.TrimSpace(cookie) == "" {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		adminID, err := authService.VerifySession(cookie)
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		admin, err := adminRepo.GetByID(c.Request.Context(), adminID)
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		c.Set(CurrentAdminKey, admin)
		c.Next()
	}
}

func CurrentAdmin(c *gin.Context) (*models.Admin, bool) {
	value, ok := c.Get(CurrentAdminKey)
	if !ok {
		return nil, false
	}
	admin, ok := value.(*models.Admin)
	return admin, ok
}

func bypassSetupRedirect(path string) bool {
	switch {
	case path == "/setup":
		return true
	case path == "/favicon.ico":
		return true
	case strings.HasPrefix(path, "/assets/"):
		return true
	default:
		return false
	}
}
