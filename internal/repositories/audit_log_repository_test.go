package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

func TestAuditLogRepositorySearch(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	adminRepo := NewAdminRepository(db)
	auditRepo := NewAuditLogRepository(db)

	adminID, err := adminRepo.Create(ctx, models.Admin{
		Name:         "Admin",
		Email:        "admin@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Unix(1_700_000_000, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}

	now := time.Unix(1_700_000_100, 0).UTC()
	if _, err := auditRepo.Create(ctx, models.AuditLog{
		ActorAdminID: adminID,
		Action:       "version.created",
		EntityType:   "version_release",
		EntityID:     "android",
		IP:           "127.0.0.1",
		UserAgent:    "test",
		CreatedAt:    now,
	}); err != nil {
		t.Fatalf("create audit: %v", err)
	}
	if _, err := auditRepo.Create(ctx, models.AuditLog{
		ActorAdminID: adminID,
		Action:       "flag.deleted",
		EntityType:   "feature_flag",
		EntityID:     "1",
		IP:           "127.0.0.1",
		UserAgent:    "test",
		CreatedAt:    now.Add(time.Minute),
	}); err != nil {
		t.Fatalf("create audit: %v", err)
	}

	page, err := auditRepo.Search(ctx, AuditLogFilter{Query: "deleted", Limit: 10})
	if err != nil {
		t.Fatalf("search audit: %v", err)
	}
	if page.Total != 1 || len(page.Logs) != 1 {
		t.Fatalf("unexpected page result: %+v", page)
	}
	if page.Logs[0].ActorName != "Admin" {
		t.Fatalf("expected actor name, got %+v", page.Logs[0])
	}
}
