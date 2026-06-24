package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

type VersionReleaseRepository struct {
	db DBTX
}

func NewVersionReleaseRepository(db DBTX) *VersionReleaseRepository {
	return &VersionReleaseRepository{db: db}
}

func (r *VersionReleaseRepository) Create(ctx context.Context, release models.VersionRelease) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO version_releases (
			platform, latest_version, minimum_version, force_update, release_notes, created_by_admin_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		release.Platform,
		release.LatestVersion,
		release.MinimumVersion,
		boolToInt(release.ForceUpdate),
		release.ReleaseNotes,
		release.CreatedByAdminID,
		release.CreatedAt.Unix(),
		release.UpdatedAt.Unix(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *VersionReleaseRepository) DeleteByID(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM version_releases WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *VersionReleaseRepository) ListByPlatform(ctx context.Context, platform string, limit int) ([]models.VersionRelease, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, platform, latest_version, minimum_version, force_update, release_notes, created_by_admin_id, created_at, updated_at
		FROM version_releases
		WHERE platform = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, platform, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	releases := make([]models.VersionRelease, 0)
	for rows.Next() {
		entry, err := scanVersionRelease(rows)
		if err != nil {
			return nil, err
		}
		releases = append(releases, *entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return releases, nil
}

func (r *VersionReleaseRepository) LatestByPlatform(ctx context.Context, platform string) (*models.VersionRelease, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, platform, latest_version, minimum_version, force_update, release_notes, created_by_admin_id, created_at, updated_at
		FROM version_releases
		WHERE platform = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, platform)
	return scanVersionReleaseRow(row)
}

func scanVersionRelease(rows *sql.Rows) (*models.VersionRelease, error) {
	var entry models.VersionRelease
	var force int
	var createdAt, updatedAt int64
	if err := rows.Scan(
		&entry.ID,
		&entry.Platform,
		&entry.LatestVersion,
		&entry.MinimumVersion,
		&force,
		&entry.ReleaseNotes,
		&entry.CreatedByAdminID,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}
	entry.ForceUpdate = force != 0
	entry.CreatedAt = time.Unix(createdAt, 0).UTC()
	entry.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &entry, nil
}

func scanVersionReleaseRow(row *sql.Row) (*models.VersionRelease, error) {
	var entry models.VersionRelease
	var force int
	var createdAt, updatedAt int64
	if err := row.Scan(
		&entry.ID,
		&entry.Platform,
		&entry.LatestVersion,
		&entry.MinimumVersion,
		&force,
		&entry.ReleaseNotes,
		&entry.CreatedByAdminID,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	entry.ForceUpdate = force != 0
	entry.CreatedAt = time.Unix(createdAt, 0).UTC()
	entry.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &entry, nil
}
