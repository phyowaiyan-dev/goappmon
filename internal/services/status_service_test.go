package services

import (
	"context"
	"testing"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
)

func TestStatusServicePublicViews(t *testing.T) {
	db := repositoriesTestDB(t)
	ctx := context.Background()
	settingsRepo := repositories.NewSettingRepository(db)
	flagsRepo := repositories.NewFeatureFlagRepository(db)

	now := time.Unix(1_700_000_300, 0).UTC()
	if _, err := settingsRepo.Create(ctx, models.Setting{
		AndroidEnabled:       true,
		AppName:              "GoAppMon",
		AndroidLatestVersion: "1.2.0",
		AndroidMinVersion:    "1.0.0",
		AndroidForceUpdate:   true,
		IOSEnabled:           true,
		IOSLatestVersion:     "1.1.0",
		IOSMinVersion:        "1.0.0",
		IOSForceUpdate:       false,
		MaintenanceMode:      true,
		MaintenanceMessage:   "maintenance",
		BannerEnabled:        true,
		BannerMessage:        "banner",
		APIURL:               "https://api.example.com",
		CreatedAt:            now,
		UpdatedAt:            now,
	}); err != nil {
		t.Fatalf("create settings: %v", err)
	}
	if _, err := flagsRepo.Create(ctx, "chat", true); err != nil {
		t.Fatalf("create feature flag: %v", err)
	}
	if _, err := flagsRepo.Create(ctx, "payment", false); err != nil {
		t.Fatalf("create feature flag: %v", err)
	}

	service := NewStatusService(settingsRepo, flagsRepo)
	status, err := service.PublicStatus(ctx)
	if err != nil {
		t.Fatalf("public status: %v", err)
	}
	if !status.MaintenanceMode || !status.BannerEnabled {
		t.Fatalf("unexpected status: %+v", status)
	}

	version, err := service.PublicVersion(ctx)
	if err != nil {
		t.Fatalf("public version: %v", err)
	}
	if !version.Android.ForceUpdate || version.Android.LatestVersion != "1.2.0" {
		t.Fatalf("unexpected version: %+v", version)
	}

	cfg, err := service.PublicConfig(ctx)
	if err != nil {
		t.Fatalf("public config: %v", err)
	}
	if cfg.AppName != "GoAppMon" || cfg.APIURL != "https://api.example.com" {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	flagMap, err := service.PublicFeatureFlags(ctx)
	if err != nil {
		t.Fatalf("public feature flags: %v", err)
	}
	if !flagMap["chat"] || flagMap["payment"] {
		t.Fatalf("unexpected feature map: %#v", flagMap)
	}
}
