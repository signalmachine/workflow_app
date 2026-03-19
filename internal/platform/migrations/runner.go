package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed sql/*.sql
var migrationFiles embed.FS

type migration struct {
	version string
	sql     string
}

func Up(ctx context.Context, db *sql.DB) (int, error) {
	if err := ensureMetadata(ctx, db); err != nil {
		return 0, err
	}

	migrations, err := loadUpMigrations()
	if err != nil {
		return 0, err
	}

	applied := 0
	for _, m := range migrations {
		done, err := alreadyApplied(ctx, db, m.version)
		if err != nil {
			return applied, err
		}
		if done {
			continue
		}

		if err := applyMigration(ctx, db, m); err != nil {
			return applied, err
		}
		applied++
	}

	return applied, nil
}

func ensureMetadata(ctx context.Context, db *sql.DB) error {
	const statement = `
CREATE SCHEMA IF NOT EXISTS platform;

CREATE TABLE IF NOT EXISTS platform.schema_migrations (
	version TEXT PRIMARY KEY,
	checksum TEXT NOT NULL,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`

	_, err := db.ExecContext(ctx, statement)
	return err
}

func loadUpMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "sql")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		path := filepath.Join("sql", entry.Name())
		body, err := fs.ReadFile(migrationFiles, path)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, migration{
			version: strings.TrimSuffix(entry.Name(), ".up.sql"),
			sql:     string(body),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func alreadyApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	const query = `
SELECT EXISTS (
	SELECT 1
	FROM platform.schema_migrations
	WHERE version = $1
);`

	var exists bool
	if err := db.QueryRowContext(ctx, query, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}

	return exists, nil
}

func applyMigration(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", m.version, err)
	}

	if _, err := tx.ExecContext(ctx, m.sql); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("execute migration %s: %w", m.version, err)
	}

	const insert = `
INSERT INTO platform.schema_migrations (version, checksum)
VALUES ($1, $2);`

	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(m.sql)))
	if _, err := tx.ExecContext(ctx, insert, m.version, checksum); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s: %w", m.version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", m.version, err)
	}

	return nil
}
