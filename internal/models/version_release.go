package models

import "time"

type VersionRelease struct {
	ID               int64     `json:"id"`
	Platform         string    `json:"platform"`
	LatestVersion    string    `json:"latest_version"`
	MinimumVersion   string    `json:"minimum_version"`
	ForceUpdate      bool      `json:"force_update"`
	ReleaseNotes     string    `json:"release_notes"`
	CreatedByAdminID int64     `json:"created_by_admin_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
