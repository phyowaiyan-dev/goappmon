package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

var ErrAdminNotFound = errors.New("admin not found")

type AdminRepository struct {
	db DBTX
}

func NewAdminRepository(db DBTX) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admins`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *AdminRepository) Create(ctx context.Context, admin models.Admin) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO admins (name, email, password_hash, created_at)
		VALUES (?, ?, ?, ?)
	`, admin.Name, admin.Email, admin.PasswordHash, admin.CreatedAt.Unix())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *AdminRepository) GetByEmail(ctx context.Context, email string) (*models.Admin, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, created_at
		FROM admins
		WHERE email = ?
	`, email)
	return scanAdmin(row)
}

func (r *AdminRepository) GetByID(ctx context.Context, id int64) (*models.Admin, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, created_at
		FROM admins
		WHERE id = ?
	`, id)
	return scanAdmin(row)
}

func scanAdmin(row *sql.Row) (*models.Admin, error) {
	var admin models.Admin
	var createdAt int64
	if err := row.Scan(&admin.ID, &admin.Name, &admin.Email, &admin.PasswordHash, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminNotFound
		}
		return nil, err
	}
	admin.CreatedAt = time.Unix(createdAt, 0).UTC()
	return &admin, nil
}
