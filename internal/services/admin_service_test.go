package services

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/database"
	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

func TestAdminServiceHistoryAndAuditTrail(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "goappmon.sqlite")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	ctx := context.Background()
	settingsRepo := repositories.NewSettingRepository(db)
	flagsRepo := repositories.NewFeatureFlagRepository(db)
	now := time.Unix(1_700_000_500, 0).UTC()
	if _, err := settingsRepo.Create(ctx, models.Setting{
		AppName:              "GoAppMon",
		AndroidEnabled:       true,
		AndroidLatestVersion: "1.0.0",
		AndroidMinVersion:    "1.0.0",
		IOSEnabled:           true,
		IOSLatestVersion:     "1.0.0",
		IOSMinVersion:        "1.0.0",
		CreatedAt:            now,
		UpdatedAt:            now,
	}); err != nil {
		t.Fatalf("create settings: %v", err)
	}

	service := NewAdminService(db, settingsRepo, flagsRepo, dbPath, now)
	meta := ActionMeta{ActorID: 42, IP: "127.0.0.1", UserAgent: "codex-test"}

	if err := service.PublishVersion(ctx, meta, "android", "2.0.0", "1.5.0", true, "android release"); err != nil {
		t.Fatalf("publish android version: %v", err)
	}
	if err := service.UpdateMaintenance(ctx, meta, true, "maintenance"); err != nil {
		t.Fatalf("update maintenance: %v", err)
	}
	if err := service.UpdateBanner(ctx, meta, true, "banner"); err != nil {
		t.Fatalf("update banner: %v", err)
	}
	if err := service.CreateFlag(ctx, meta, "chat", true); err != nil {
		t.Fatalf("create flag: %v", err)
	}

	dashboard, err := service.Dashboard(ctx)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.AndroidReleases) != 1 {
		t.Fatalf("expected 1 android release, got %d", len(dashboard.AndroidReleases))
	}
	if len(dashboard.MaintenanceHistory) != 1 || len(dashboard.BannerHistory) != 1 {
		t.Fatalf("expected history entries, got maintenance=%d banner=%d", len(dashboard.MaintenanceHistory), len(dashboard.BannerHistory))
	}
	if len(dashboard.AuditLogs) == 0 {
		t.Fatal("expected audit log entries")
	}
	if dashboard.SystemHealth.Score < 36 || dashboard.SystemHealth.Score > 96 {
		t.Fatalf("expected bounded health score, got %d", dashboard.SystemHealth.Score)
	}
	if dashboard.SystemHealth.MemoryTotalMB < 0 || dashboard.SystemHealth.DiskTotalGB < 0 {
		t.Fatalf("expected non-negative machine metrics, got %+v", dashboard.SystemHealth)
	}
}

func TestAdminServiceVersionValidationAndDeleteLatest(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "goappmon.sqlite")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	ctx := context.Background()
	settingsRepo := repositories.NewSettingRepository(db)
	flagsRepo := repositories.NewFeatureFlagRepository(db)
	now := time.Unix(1_700_000_500, 0).UTC()
	if _, err := settingsRepo.Create(ctx, models.Setting{
		AppName:              "GoAppMon",
		AndroidEnabled:       true,
		AndroidLatestVersion: "1.0.0",
		AndroidMinVersion:    "1.0.0",
		IOSEnabled:           true,
		IOSLatestVersion:     "1.0.0",
		IOSMinVersion:        "1.0.0",
		CreatedAt:            now,
		UpdatedAt:            now,
	}); err != nil {
		t.Fatalf("create settings: %v", err)
	}

	service := NewAdminService(db, settingsRepo, flagsRepo, dbPath, now)
	meta := ActionMeta{ActorID: 42, IP: "127.0.0.1", UserAgent: "codex-test"}

	if err := service.PublishVersion(ctx, meta, "android", "1.0.0", "1.0.0", false, "seed"); err != nil {
		t.Fatalf("publish android seed version: %v", err)
	}
	if err := service.PublishVersion(ctx, meta, "android", "2.0.0", "1.5.0", true, "android release"); err != nil {
		t.Fatalf("publish android latest version: %v", err)
	}
	if err := service.PublishVersion(ctx, meta, "ios", "1.0.0", "1.0.0", false, "seed"); err != nil {
		t.Fatalf("publish ios seed version: %v", err)
	}
	if err := service.PublishVersion(ctx, meta, "ios", "2.0.0", "1.5.0", true, "ios release"); err != nil {
		t.Fatalf("publish ios latest version: %v", err)
	}

	if err := service.PublishVersion(ctx, meta, "android", "2.0", "1.5.0", false, "bad"); !errors.Is(err, ErrInvalidVersionFormat) {
		t.Fatalf("expected invalid version format, got %v", err)
	}
	if err := service.PublishVersion(ctx, meta, "android", "2.0.0", "2.1.0", false, "bad"); !errors.Is(err, ErrMinimumVersionGreaterThanLatest) {
		t.Fatalf("expected minimum greater than latest, got %v", err)
	}
	if err := service.PublishVersion(ctx, meta, "android", "2.0.0", "1.5.0", false, "bad"); !errors.Is(err, ErrLatestVersionNotGreater) {
		t.Fatalf("expected latest not increasing, got %v", err)
	}

	if err := service.DeleteCurrentVersion(ctx, meta, "android"); err != nil {
		t.Fatalf("delete android latest version: %v", err)
	}
	if err := service.DeleteCurrentVersion(ctx, meta, "ios"); err != nil {
		t.Fatalf("delete ios latest version: %v", err)
	}

	dashboard, err := service.Dashboard(ctx)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dashboard.AndroidReleases) != 1 {
		t.Fatalf("expected 1 android release after delete, got %d", len(dashboard.AndroidReleases))
	}
	if dashboard.Settings.AndroidLatestVersion != "1.0.0" {
		t.Fatalf("expected android latest version to roll back, got %s", dashboard.Settings.AndroidLatestVersion)
	}
	if len(dashboard.IOSReleases) != 1 {
		t.Fatalf("expected 1 ios release after delete, got %d", len(dashboard.IOSReleases))
	}
	if dashboard.Settings.IOSLatestVersion != "1.0.0" {
		t.Fatalf("expected ios latest version to roll back, got %s", dashboard.Settings.IOSLatestVersion)
	}

	if err := service.DeleteCurrentVersion(ctx, meta, "android"); !errors.Is(err, ErrCannotDeleteLastVersion) {
		t.Fatalf("expected cannot delete last version, got %v", err)
	}
}

func TestCalculateHealthScorePenaltyProfile(t *testing.T) {
	score, status, notes := calculateHealthScore(
		73.1,
		&mem.VirtualMemoryStat{UsedPercent: 85.1, Available: 1221 * 1024 * 1024},
		nil,
		&disk.UsageStat{UsedPercent: 97.3, Free: 6 * 1024 * 1024 * 1024},
		nil,
		7,
		56*1024*1024,
		4*1024*1024,
	)

	if score >= 70 {
		t.Fatalf("expected heavy pressure score to be low, got %d", score)
	}
	if status != "Needs attention" {
		t.Fatalf("expected needs attention status, got %q", status)
	}
	if len(notes) == 0 {
		t.Fatal("expected warning notes for heavy pressure")
	}
}
