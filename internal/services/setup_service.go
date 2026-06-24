package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
	"github.com/phyowaiyan-dev/goappmon/internal/utils"
)

var ErrSetupAlreadyComplete = errors.New("setup already complete")

type SetupService struct {
	db *sql.DB
}

func NewSetupService(db *sql.DB) *SetupService {
	return &SetupService{db: db}
}

func (s *SetupService) IsSetupComplete(ctx context.Context) (bool, error) {
	count, err := repositories.NewAdminRepository(s.db).Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *SetupService) EnsureDefaultSettings(ctx context.Context) error {
	adminCount, err := repositories.NewAdminRepository(s.db).Count(ctx)
	if err != nil {
		return err
	}
	if adminCount == 0 {
		return nil
	}

	repo := repositories.NewSettingRepository(s.db)
	count, err := repo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	_, err = repo.Create(ctx, models.Setting{
		AppName:              "GoAppMon",
		AndroidEnabled:       true,
		AndroidLatestVersion: "1.0.0",
		AndroidMinVersion:    "1.0.0",
		IOSEnabled:           true,
		IOSLatestVersion:     "1.0.0",
		IOSMinVersion:        "1.0.0",
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	return err
}

func (s *SetupService) CreateInitialSetup(ctx context.Context, adminName, adminEmail, password, appName string) (err error) {
	adminRepo := repositories.NewAdminRepository(s.db)
	count, err := adminRepo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrSetupAlreadyComplete
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	txAdmins := repositories.NewAdminRepository(tx)
	txSettings := repositories.NewSettingRepository(tx)
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if _, err = txAdmins.Create(ctx, models.Admin{
		Name:         adminName,
		Email:        adminEmail,
		PasswordHash: hash,
		CreatedAt:    now,
	}); err != nil {
		return err
	}
	if _, err = txSettings.Create(ctx, models.Setting{
		AppName:              appName,
		AndroidEnabled:       true,
		AndroidLatestVersion: "1.0.0",
		AndroidMinVersion:    "1.0.0",
		IOSEnabled:           true,
		IOSLatestVersion:     "1.0.0",
		IOSMinVersion:        "1.0.0",
		CreatedAt:            now,
		UpdatedAt:            now,
	}); err != nil {
		return err
	}

	return tx.Commit()
}
