package services

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/phyowaiyan-dev/goappmon/internal/database"
	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

func repositoriesTestDB(t *testing.T) *sql.DB {
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

func adminRow(name, email, hash string) models.Admin {
	return models.Admin{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
	}
}
