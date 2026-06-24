package repositories

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/database"
	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
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
	return db
}

func TestAdminRepositoryCRUD(t *testing.T) {
	db := newTestDB(t)
	repo := NewAdminRepository(db)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("count empty: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected empty count, got %d", count)
	}

	createdAt := time.Unix(1_700_000_000, 0).UTC()
	id, err := repo.Create(ctx, models.Admin{
		Name:         "Admin",
		Email:        "admin@example.com",
		PasswordHash: "hash",
		CreatedAt:    createdAt,
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("count after create: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	admin, err := repo.GetByEmail(ctx, "admin@example.com")
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if admin.Name != "Admin" || admin.Email != "admin@example.com" || admin.PasswordHash != "hash" {
		t.Fatalf("unexpected admin: %+v", admin)
	}
	if !admin.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected created at: %v", admin.CreatedAt)
	}

	byID, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if byID.ID != id {
		t.Fatalf("unexpected id: %d", byID.ID)
	}

	if _, err := repo.GetByEmail(ctx, "missing@example.com"); err != ErrAdminNotFound {
		t.Fatalf("expected ErrAdminNotFound, got %v", err)
	}
}
