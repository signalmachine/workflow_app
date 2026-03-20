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

const testDatabaseLockKey int64 = 20260319

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

	lockConn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("open test database lock connection: %v", err)
	}

	if _, err := lockConn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, testDatabaseLockKey); err != nil {
		_ = lockConn.Close()
		t.Fatalf("acquire test database lock: %v", err)
	}

	t.Cleanup(func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = lockConn.ExecContext(unlockCtx, `SELECT pg_advisory_unlock($1)`, testDatabaseLockKey)
		_ = lockConn.Close()
	})

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
	ai.agent_delegations,
	ai.agent_recommendations,
	ai.agent_artifacts,
	ai.agent_tool_policies,
	ai.agent_run_steps,
	ai.agent_runs,
	ai.agent_tools,
	workforce.labor_accounting_handoffs,
	workforce.labor_entries,
	workflow.tasks,
	workforce.workers,
	work_orders.material_usages,
	work_orders.status_history,
	work_orders.work_orders,
	inventory_ops.execution_links,
	inventory_ops.accounting_handoffs,
	inventory_ops.document_lines,
	inventory_ops.documents,
	inventory_ops.movements,
	inventory_ops.movement_numbering_series,
	inventory_ops.locations,
	inventory_ops.items,
	accounting.journal_lines,
	accounting.journal_entries,
	accounting.journal_numbering_series,
	accounting.tax_codes,
	accounting.periods,
	accounting.ledger_accounts,
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
