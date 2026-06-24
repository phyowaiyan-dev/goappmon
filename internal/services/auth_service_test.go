package services

import (
	"context"
	"testing"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
	"github.com/phyowaiyan-dev/goappmon/internal/utils"
)

func TestAuthServiceAuthenticateAndSession(t *testing.T) {
	db := repositoriesTestDB(t)
	adminRepo := repositories.NewAdminRepository(db)

	hash, err := utils.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	_, err = adminRepo.Create(context.Background(), models.Admin{
		Name:         "Admin",
		Email:        "admin@example.com",
		PasswordHash: hash,
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}

	service := NewAuthService(adminRepo, []byte("01234567890123456789012345678901"), time.Hour)
	admin, err := service.Authenticate(context.Background(), "admin@example.com", "secret123")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if admin.Email != "admin@example.com" {
		t.Fatalf("unexpected admin: %+v", admin)
	}

	if _, err := service.Authenticate(context.Background(), "admin@example.com", "bad"); err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if _, err := service.Authenticate(context.Background(), "missing@example.com", "secret123"); err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials for missing admin, got %v", err)
	}

	token, err := service.SignSession(admin.ID)
	if err != nil {
		t.Fatalf("sign session: %v", err)
	}
	adminID, err := service.VerifySession(token)
	if err != nil {
		t.Fatalf("verify session: %v", err)
	}
	if adminID != admin.ID {
		t.Fatalf("unexpected admin id: %d", adminID)
	}

	if _, err := service.VerifySession(token + "tamper"); err == nil {
		t.Fatal("expected tampered session to fail")
	}
}
