package dbtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"workflow_app/internal/platform/migrations"
	"workflow_app/internal/testsupport/testdb"
)

func Open(t *testing.T) *sql.DB {
	t.Helper()

	db := testdb.OpenFromEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), testdb.DefaultSetupTimeout)
	defer cancel()

	testdb.MustAcquireAdvisoryLock(t, ctx, db, testdb.DefaultLockKey)

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

	if _, err := db.ExecContext(ctx, statement); err != nil {
		t.Fatalf("reset test database: %v", err)
	}
}
