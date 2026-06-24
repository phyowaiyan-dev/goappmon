package models

import "time"

type AuditLog struct {
	ID           int64     `json:"id"`
	ActorAdminID int64     `json:"actor_admin_id"`
	ActorName    string    `json:"actor_name"`
	Action       string    `json:"action"`
	EntityType   string    `json:"entity_type"`
	EntityID     string    `json:"entity_id"`
	BeforeJSON   string    `json:"before_json"`
	AfterJSON    string    `json:"after_json"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
}
