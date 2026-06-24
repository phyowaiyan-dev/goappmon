package models

import "time"

type Setting struct {
	ID                   int64     `json:"id"`
	AppName              string    `json:"app_name"`
	AndroidEnabled       bool      `json:"android_enabled"`
	AndroidLatestVersion string    `json:"android_latest_version"`
	AndroidMinVersion    string    `json:"android_min_version"`
	AndroidForceUpdate   bool      `json:"android_force_update"`
	IOSEnabled           bool      `json:"ios_enabled"`
	IOSLatestVersion     string    `json:"ios_latest_version"`
	IOSMinVersion        string    `json:"ios_min_version"`
	IOSForceUpdate       bool      `json:"ios_force_update"`
	MaintenanceMode      bool      `json:"maintenance_mode"`
	MaintenanceMessage   string    `json:"maintenance_message"`
	BannerEnabled        bool      `json:"banner_enabled"`
	BannerMessage        string    `json:"banner_message"`
	APIURL               string    `json:"api_url"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
