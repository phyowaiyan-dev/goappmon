package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

var ErrSettingsNotFound = errors.New("settings not found")

type SettingRepository struct {
	db DBTX
}

func NewSettingRepository(db DBTX) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM settings`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SettingRepository) Create(ctx context.Context, setting models.Setting) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO settings (
			app_name, android_enabled, android_latest_version, android_min_version, android_force_update,
			ios_enabled, ios_latest_version, ios_min_version, ios_force_update,
			maintenance_mode, maintenance_message, banner_enabled, banner_message, api_url,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		setting.AppName,
		boolToInt(setting.AndroidEnabled),
		setting.AndroidLatestVersion,
		setting.AndroidMinVersion,
		boolToInt(setting.AndroidForceUpdate),
		boolToInt(setting.IOSEnabled),
		setting.IOSLatestVersion,
		setting.IOSMinVersion,
		boolToInt(setting.IOSForceUpdate),
		boolToInt(setting.MaintenanceMode),
		setting.MaintenanceMessage,
		boolToInt(setting.BannerEnabled),
		setting.BannerMessage,
		setting.APIURL,
		setting.CreatedAt.Unix(),
		setting.UpdatedAt.Unix(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SettingRepository) GetCurrent(ctx context.Context) (*models.Setting, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, app_name, android_enabled, android_latest_version, android_min_version, android_force_update,
		       ios_enabled, ios_latest_version, ios_min_version, ios_force_update, maintenance_mode,
		       maintenance_message, banner_enabled, banner_message, api_url, created_at, updated_at
		FROM settings
		ORDER BY id ASC
		LIMIT 1
	`)
	return scanSetting(row)
}

func (r *SettingRepository) UpdateApplication(ctx context.Context, appName, apiURL string) error {
	return r.update(ctx, `
		UPDATE settings
		SET app_name = ?, api_url = ?, updated_at = ?
		WHERE id = (SELECT id FROM settings ORDER BY id ASC LIMIT 1)
	`, appName, apiURL)
}

func (r *SettingRepository) UpdateVersion(ctx context.Context, androidLatest, androidMin string, androidForce bool, iosLatest, iosMin string, iosForce bool) error {
	return r.update(ctx, `
		UPDATE settings
		SET android_latest_version = ?,
		    android_min_version = ?,
		    android_force_update = ?,
		    ios_latest_version = ?,
		    ios_min_version = ?,
		    ios_force_update = ?,
		    updated_at = ?
		WHERE id = (SELECT id FROM settings ORDER BY id ASC LIMIT 1)
	`, androidLatest, androidMin, boolToInt(androidForce), iosLatest, iosMin, boolToInt(iosForce))
}

func (r *SettingRepository) UpdatePlatformVersion(ctx context.Context, platform, latest, minimum string, force bool) error {
	switch platform {
	case "android":
		return r.update(ctx, `
			UPDATE settings
			SET android_latest_version = ?, android_min_version = ?, android_force_update = ?, updated_at = ?
			WHERE id = (SELECT id FROM settings ORDER BY id ASC LIMIT 1)
		`, latest, minimum, boolToInt(force))
	case "ios":
		return r.update(ctx, `
			UPDATE settings
			SET ios_latest_version = ?, ios_min_version = ?, ios_force_update = ?, updated_at = ?
			WHERE id = (SELECT id FROM settings ORDER BY id ASC LIMIT 1)
		`, latest, minimum, boolToInt(force))
	default:
		return ErrSettingsNotFound
	}
}

func (r *SettingRepository) UpdatePlatforms(ctx context.Context, androidEnabled, iosEnabled bool) error {
	return r.update(ctx, `
		UPDATE settings
		SET android_enabled = ?, ios_enabled = ?, updated_at = ?
		WHERE id = (SELECT id FROM settings ORDER BY id ASC LIMIT 1)
	`, boolToInt(androidEnabled), boolToInt(iosEnabled))
}

func (r *SettingRepository) UpdateMaintenance(ctx context.Context, enabled bool, message string) error {
	return r.update(ctx, `
		UPDATE settings
		SET maintenance_mode = ?, maintenance_message = ?, updated_at = ?
		WHERE id = (SELECT id FROM settings ORDER BY id ASC LIMIT 1)
	`, boolToInt(enabled), message)
}

func (r *SettingRepository) UpdateBanner(ctx context.Context, enabled bool, message string) error {
	return r.update(ctx, `
		UPDATE settings
		SET banner_enabled = ?, banner_message = ?, updated_at = ?
		WHERE id = (SELECT id FROM settings ORDER BY id ASC LIMIT 1)
	`, boolToInt(enabled), message)
}

func (r *SettingRepository) update(ctx context.Context, query string, args ...any) error {
	args = append(args, time.Now().UTC().Unix())
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrSettingsNotFound
	}
	return nil
}

func scanSetting(row *sql.Row) (*models.Setting, error) {
	var setting models.Setting
	var androidEnabled, androidForce, iosEnabled, iosForce, maintenanceMode, bannerEnabled int
	var createdAt, updatedAt int64
	if err := row.Scan(
		&setting.ID,
		&setting.AppName,
		&androidEnabled,
		&setting.AndroidLatestVersion,
		&setting.AndroidMinVersion,
		&androidForce,
		&iosEnabled,
		&setting.IOSLatestVersion,
		&setting.IOSMinVersion,
		&iosForce,
		&maintenanceMode,
		&setting.MaintenanceMessage,
		&bannerEnabled,
		&setting.BannerMessage,
		&setting.APIURL,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSettingsNotFound
		}
		return nil, err
	}
	setting.AndroidEnabled = androidEnabled != 0
	setting.AndroidForceUpdate = androidForce != 0
	setting.IOSEnabled = iosEnabled != 0
	setting.IOSForceUpdate = iosForce != 0
	setting.MaintenanceMode = maintenanceMode != 0
	setting.BannerEnabled = bannerEnabled != 0
	setting.CreatedAt = time.Unix(createdAt, 0).UTC()
	setting.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &setting, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
