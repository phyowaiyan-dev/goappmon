package models

import "time"

type StateChange struct {
	ID               int64     `json:"id"`
	Kind             string    `json:"kind"`
	Enabled          bool      `json:"enabled"`
	Message          string    `json:"message"`
	CreatedByAdminID int64     `json:"created_by_admin_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
