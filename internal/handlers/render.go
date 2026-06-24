package handlers

import (
	"net/http"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

type Renderer interface {
	RenderPage(w http.ResponseWriter, page string, data any) error
}

type PageData struct {
	Title         string
	Error         string
	Notice        string
	CurrentAdmin  *models.Admin
	Settings      *models.Setting
	Flags         []models.FeatureFlag
	SystemHealth  any
	AdminName     string
	AdminEmail    string
	AppName       string
	LoginEmail    string
	SetupComplete bool
}
