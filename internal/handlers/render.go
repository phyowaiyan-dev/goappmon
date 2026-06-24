package handlers

import (
	"net/http"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

type Renderer interface {
	RenderPage(w http.ResponseWriter, page string, data any) error
	RenderFragment(w http.ResponseWriter, fragment string, data any) error
}

type PageData struct {
	Title              string
	Error              string
	Notice             string
	Authenticated      bool
	CurrentAdmin       *models.Admin
	Settings           *models.Setting
	Flags              []models.FeatureFlag
	AndroidReleases    []models.VersionRelease
	IOSReleases        []models.VersionRelease
	MaintenanceHistory []models.StateChange
	BannerHistory      []models.StateChange
	AuditLogs          []models.AuditLog
	SystemHealth       any
	AdminName          string
	AdminEmail         string
	AppName            string
	LoginEmail         string
	SetupComplete      bool
}

type AuditLogPageData struct {
	Title         string
	Error         string
	Notice        string
	Authenticated bool
	CurrentAdmin  *models.Admin
	Logs          []models.AuditLog
	Query         string
	Action        string
	EntityType    string
	EntityID      string
	ActorID       string
	From          string
	To            string
	Page          int
	PageSize      int
	Total         int64
	TotalPages    int
	HasPrev       bool
	HasNext       bool
	PrevPage      int
	NextPage      int
	PrevURL       string
	NextURL       string
}
