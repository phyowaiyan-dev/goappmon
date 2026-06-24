package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

func TestSettingRepositoryLifecycle(t *testing.T) {
	db := newTestDB(t)
	repo := NewSettingRepository(db)
	ctx := context.Background()

	if count, err := repo.Count(ctx); err != nil || count != 0 {
		t.Fatalf("expected empty settings count, got %d, %v", count, err)
	}

	now := time.Unix(1_700_000_100, 0).UTC()
	id, err := repo.Create(ctx, models.Setting{
		AppName:              "GoAppMon",
		AndroidLatestVersion: "1.0.0",
		AndroidMinVersion:    "1.0.0",
		IOSLatestVersion:     "1.0.0",
		IOSMinVersion:        "1.0.0",
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		t.Fatalf("create settings: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	current, err := repo.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current: %v", err)
	}
	if current.AppName != "GoAppMon" || !current.CreatedAt.Equal(now) || !current.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected settings: %+v", current)
	}

	if err := repo.UpdateApplication(ctx, "App Two", "https://api.example.com"); err != nil {
		t.Fatalf("update application: %v", err)
	}
	if err := repo.UpdateVersion(ctx, "2.0.0", "1.5.0", true, "2.1.0", "1.5.1", false); err != nil {
		t.Fatalf("update version: %v", err)
	}
	if err := repo.UpdateMaintenance(ctx, true, "maintenance"); err != nil {
		t.Fatalf("update maintenance: %v", err)
	}
	if err := repo.UpdateBanner(ctx, true, "banner"); err != nil {
		t.Fatalf("update banner: %v", err)
	}

	updated, err := repo.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get updated current: %v", err)
	}
	if updated.AppName != "App Two" || updated.APIURL != "https://api.example.com" {
		t.Fatalf("unexpected updated app settings: %+v", updated)
	}
	if updated.AndroidLatestVersion != "2.0.0" || updated.AndroidMinVersion != "1.5.0" || !updated.AndroidForceUpdate {
		t.Fatalf("unexpected android settings: %+v", updated)
	}
	if updated.IOSLatestVersion != "2.1.0" || updated.IOSMinVersion != "1.5.1" || updated.IOSForceUpdate {
		t.Fatalf("unexpected ios settings: %+v", updated)
	}
	if !updated.MaintenanceMode || updated.MaintenanceMessage != "maintenance" {
		t.Fatalf("unexpected maintenance settings: %+v", updated)
	}
	if !updated.BannerEnabled || updated.BannerMessage != "banner" {
		t.Fatalf("unexpected banner settings: %+v", updated)
	}
	if updated.UpdatedAt.Equal(now) {
		t.Fatalf("expected updated_at to change")
	}
}

func TestSettingRepositoryUpdateMissing(t *testing.T) {
	db := newTestDB(t)
	repo := NewSettingRepository(db)
	ctx := context.Background()

	if err := repo.UpdateApplication(ctx, "App", ""); err != ErrSettingsNotFound {
		t.Fatalf("expected ErrSettingsNotFound, got %v", err)
	}
}
