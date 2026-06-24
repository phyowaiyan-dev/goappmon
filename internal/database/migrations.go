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
	if err := ensureSettingsColumns(ctx, db); err != nil {
		return err
	}
	return nil
}

func ensureSettingsColumns(ctx context.Context, db *sql.DB) error {
	columns, err := tableColumns(ctx, db, "settings")
	if err != nil {
		return err
	}
	required := map[string]string{
		"android_enabled": "INTEGER NOT NULL DEFAULT 1",
		"ios_enabled":     "INTEGER NOT NULL DEFAULT 1",
	}
	for name, ddl := range required {
		if _, ok := columns[name]; ok {
			continue
		}
		if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE settings ADD COLUMN %s %s", name, ddl)); err != nil {
			return fmt.Errorf("add settings column %s: %w", name, err)
		}
	}
	return nil
}

func tableColumns(ctx context.Context, db *sql.DB, table string) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]struct{})
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
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
