package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
)

//go:embed migrations.sql
var migrationsFS embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	content, err := migrationsFS.ReadFile("migrations.sql")
	if err != nil {
		return err
	}

	for _, statement := range splitStatements(string(content)) {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}
	return nil
}

func splitStatements(content string) []string {
	parts := strings.Split(content, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		statement := strings.TrimSpace(part)
		if statement == "" {
			continue
		}
		statements = append(statements, statement)
	}
	return statements
}
