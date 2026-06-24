package repositories

import (
	"context"
	"errors"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

var ErrFeatureFlagNotFound = errors.New("feature flag not found")

type FeatureFlagRepository struct {
	db DBTX
}

func NewFeatureFlagRepository(db DBTX) *FeatureFlagRepository {
	return &FeatureFlagRepository{db: db}
}

func (r *FeatureFlagRepository) List(ctx context.Context) ([]models.FeatureFlag, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, key, enabled FROM feature_flags ORDER BY key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flags := make([]models.FeatureFlag, 0)
	for rows.Next() {
		var flag models.FeatureFlag
		var enabled int
		if err := rows.Scan(&flag.ID, &flag.Key, &enabled); err != nil {
			return nil, err
		}
		flag.Enabled = enabled != 0
		flags = append(flags, flag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return flags, nil
}

func (r *FeatureFlagRepository) Create(ctx context.Context, key string, enabled bool) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO feature_flags (key, enabled)
		VALUES (?, ?)
	`, key, boolToInt(enabled))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *FeatureFlagRepository) Update(ctx context.Context, id int64, key string, enabled bool) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE feature_flags
		SET key = ?, enabled = ?
		WHERE id = ?
	`, key, boolToInt(enabled), id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrFeatureFlagNotFound
	}
	return nil
}

func (r *FeatureFlagRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM feature_flags WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrFeatureFlagNotFound
	}
	return nil
}

func (r *FeatureFlagRepository) AsMap(ctx context.Context) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, enabled FROM feature_flags ORDER BY key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flags := make(map[string]bool)
	for rows.Next() {
		var key string
		var enabled int
		if err := rows.Scan(&key, &enabled); err != nil {
			return nil, err
		}
		flags[key] = enabled != 0
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return flags, nil
}
