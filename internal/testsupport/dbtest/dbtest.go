package dbtest

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"workflow_app/internal/platform/migrations"
)

func Open(t *testing.T) *sql.DB {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	if _, err := migrations.Up(ctx, db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return db
}

func Reset(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const statement = `
TRUNCATE TABLE
	workflow.approval_decisions,
	workflow.approval_queue_entries,
	workflow.approvals,
	documents.documents,
	documents.numbering_series,
	identityaccess.sessions,
	platform.audit_events,
	platform.idempotency_keys,
	identityaccess.memberships,
	identityaccess.users,
	identityaccess.orgs
RESTART IDENTITY CASCADE;`

	if _, err := db.ExecContext(ctx, statement); err != nil {
		t.Fatalf("reset test database: %v", err)
	}
}
