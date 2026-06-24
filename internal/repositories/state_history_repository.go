package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

type StateHistoryRepository struct {
	db DBTX
}

func NewStateHistoryRepository(db DBTX) *StateHistoryRepository {
	return &StateHistoryRepository{db: db}
}

func (r *StateHistoryRepository) Create(ctx context.Context, change models.StateChange) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO state_changes (
			kind, enabled, message, created_by_admin_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		change.Kind,
		boolToInt(change.Enabled),
		change.Message,
		change.CreatedByAdminID,
		change.CreatedAt.Unix(),
		change.UpdatedAt.Unix(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *StateHistoryRepository) ListByKind(ctx context.Context, kind string, limit int) ([]models.StateChange, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, kind, enabled, message, created_by_admin_id, created_at, updated_at
		FROM state_changes
		WHERE kind = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, kind, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	changes := make([]models.StateChange, 0)
	for rows.Next() {
		change, err := scanStateChange(rows)
		if err != nil {
			return nil, err
		}
		changes = append(changes, *change)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return changes, nil
}

func (r *StateHistoryRepository) LatestByKind(ctx context.Context, kind string) (*models.StateChange, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, kind, enabled, message, created_by_admin_id, created_at, updated_at
		FROM state_changes
		WHERE kind = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, kind)
	return scanStateChangeRow(row)
}

func scanStateChange(rows *sql.Rows) (*models.StateChange, error) {
	var change models.StateChange
	var enabled int
	var createdAt, updatedAt int64
	if err := rows.Scan(
		&change.ID,
		&change.Kind,
		&enabled,
		&change.Message,
		&change.CreatedByAdminID,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}
	change.Enabled = enabled != 0
	change.CreatedAt = time.Unix(createdAt, 0).UTC()
	change.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &change, nil
}

func scanStateChangeRow(row *sql.Row) (*models.StateChange, error) {
	var change models.StateChange
	var enabled int
	var createdAt, updatedAt int64
	if err := row.Scan(
		&change.ID,
		&change.Kind,
		&enabled,
		&change.Message,
		&change.CreatedByAdminID,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	change.Enabled = enabled != 0
	change.CreatedAt = time.Unix(createdAt, 0).UTC()
	change.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	return &change, nil
}
