package services

import (
	"context"
	"testing"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
)

func TestSetupServiceWorkflow(t *testing.T) {
	db := repositoriesTestDB(t)
	service := NewSetupService(db)
	ctx := context.Background()

	complete, err := service.IsSetupComplete(ctx)
	if err != nil {
		t.Fatalf("is setup complete: %v", err)
	}
	if complete {
		t.Fatal("expected setup to be incomplete")
	}

	if err := service.EnsureDefaultSettings(ctx); err != nil {
		t.Fatalf("ensure default settings with no admin: %v", err)
	}

	adminRepo := repositories.NewAdminRepository(db)
	if _, err := adminRepo.Create(ctx, models.Admin{
		Name:         "Existing Admin",
		Email:        "existing@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Unix(1_700_000_200, 0).UTC(),
	}); err != nil {
		t.Fatalf("create admin: %v", err)
	}

	if err := service.EnsureDefaultSettings(ctx); err != nil {
		t.Fatalf("ensure default settings with admin: %v", err)
	}

	settingsRepo := repositories.NewSettingRepository(db)
	current, err := settingsRepo.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current settings: %v", err)
	}
	if current.AppName != "GoAppMon" {
		t.Fatalf("unexpected default app name: %s", current.AppName)
	}

	// A separate database validates the full initial setup flow without
	// interference from the direct admin insert above.
	db2 := repositoriesTestDB(t)
	service2 := NewSetupService(db2)
	if err := service2.CreateInitialSetup(ctx, "First Admin", "first@example.com", "secret123", "Control Center"); err != nil {
		t.Fatalf("create initial setup: %v", err)
	}

	complete, err = service2.IsSetupComplete(ctx)
	if err != nil {
		t.Fatalf("is setup complete after setup: %v", err)
	}
	if !complete {
		t.Fatal("expected setup to be complete")
	}

	adminRepo2 := repositories.NewAdminRepository(db2)
	admin, err := adminRepo2.GetByEmail(ctx, "first@example.com")
	if err != nil {
		t.Fatalf("get first admin: %v", err)
	}
	if admin.Name != "First Admin" {
		t.Fatalf("unexpected first admin: %+v", admin)
	}

	settingsRepo2 := repositories.NewSettingRepository(db2)
	current2, err := settingsRepo2.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get setup settings: %v", err)
	}
	if current2.AppName != "Control Center" {
		t.Fatalf("unexpected setup app name: %s", current2.AppName)
	}

	if err := service2.CreateInitialSetup(ctx, "Second", "second@example.com", "secret123", "Other"); err != ErrSetupAlreadyComplete {
		t.Fatalf("expected ErrSetupAlreadyComplete, got %v", err)
	}
}
