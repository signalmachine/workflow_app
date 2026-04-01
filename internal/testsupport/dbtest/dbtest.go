package dbtest

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"workflow_app/internal/platform/migrations"
	"workflow_app/internal/testsupport/testdb"
)

var (
	migrateOnce   sync.Once
	migrateErr    error
	runMigrations = migrations.Up
)

func Open(t *testing.T) *sql.DB {
	t.Helper()

	db := testdb.OpenFromEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), testdb.DefaultSetupTimeout)
	defer cancel()

	// Hold the shared advisory lock only during destructive setup work.
	// Keeping the lock for the full test lifetime makes interrupted runs leave
	// behind stale holders that block the whole suite on the next attempt.
	if err := withSetupLock(ctx, db, func() error {
		// The schema is stable for the lifetime of a test process, and each test
		// already performs a full data reset. Running migrations once keeps the
		// DB-backed suite isolated without paying repeated no-op migration cost.
		return ensureMigrated(ctx, db)
	}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return db
}

func ensureMigrated(ctx context.Context, db *sql.DB) error {
	migrateOnce.Do(func() {
		_, migrateErr = runMigrations(ctx, db)
	})
	return migrateErr
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
	attachments.derived_texts,
	attachments.request_message_links,
	attachments.attachments,
	ai.inbound_request_messages,
	ai.inbound_requests,
	ai.inbound_request_numbering_series,
	ai.agent_tools,
	workforce.labor_accounting_handoffs,
	workforce.labor_entries,
	workflow.tasks,
	workforce.workers,
	work_orders.material_usages,
	work_orders.status_history,
	work_orders.documents,
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
	accounting.payment_receipt_documents,
	accounting.invoice_documents,
	accounting.tax_codes,
	accounting.periods,
	accounting.ledger_accounts,
	workflow.approval_decisions,
	workflow.approval_queue_entries,
	workflow.approvals,
	parties.contacts,
	parties.parties,
	documents.documents,
	documents.numbering_series,
	identityaccess.sessions,
	platform.audit_events,
	platform.idempotency_keys,
	identityaccess.memberships,
	identityaccess.users,
	identityaccess.orgs
RESTART IDENTITY CASCADE;`

	if err := withSetupLock(ctx, db, func() error {
		_, err := db.ExecContext(ctx, statement)
		return err
	}); err != nil {
		t.Fatalf("reset test database: %v", err)
	}
}

func withSetupLock(ctx context.Context, db *sql.DB, fn func() error) error {
	lockConn, err := testdb.AcquireAdvisoryLock(ctx, db, testdb.DefaultLockKey)
	if err != nil {
		return err
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = lockConn.ExecContext(releaseCtx, `SELECT pg_advisory_unlock($1)`, testdb.DefaultLockKey)
		_ = lockConn.Close()
	}()

	return fn()
}
