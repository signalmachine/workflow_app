package migrations

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestWorkOrderDocumentOwnershipMigrationBackfillsAndRollsBackIntegration(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	resetTestDB(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, downSQL(t, "sql/000015_milestone5_work_order_document_ownership.down.sql")); err != nil {
		t.Fatalf("apply down migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM platform.schema_migrations WHERE version = '000015_milestone5_work_order_document_ownership'`); err != nil {
		t.Fatalf("delete migration record: %v", err)
	}

	orgID, userID := insertLegacyOrgAndUser(t, ctx, db)

	var workOrderID string
	if err := db.QueryRowContext(ctx, `
INSERT INTO work_orders.work_orders (
	org_id,
	work_order_code,
	title,
	summary,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING id;`,
		orgID,
		"WO-LEGACY-1001",
		"Legacy work order",
		"Created before document ownership backfill",
		userID,
	).Scan(&workOrderID); err != nil {
		t.Fatalf("insert legacy work order: %v", err)
	}

	migrations, err := loadUpMigrations()
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}

	var target migration
	found := false
	for _, m := range migrations {
		if m.version == "000015_milestone5_work_order_document_ownership" {
			target = m
			found = true
			break
		}
	}
	if !found {
		t.Fatal("did not find 000015 migration")
	}

	if err := applyMigration(ctx, db, target); err != nil {
		t.Fatalf("apply 000015 migration: %v", err)
	}

	var (
		documentID     string
		documentType   string
		documentTitle  string
		documentStatus string
	)
	if err := db.QueryRowContext(ctx, `
SELECT wd.document_id, d.type_code, d.title, d.status
FROM work_orders.documents wd
JOIN documents.documents d
	ON d.id = wd.document_id
WHERE wd.work_order_id = $1;`, workOrderID).Scan(&documentID, &documentType, &documentTitle, &documentStatus); err != nil {
		t.Fatalf("load backfilled work order document: %v", err)
	}
	if documentID != workOrderID {
		t.Fatalf("unexpected backfilled document id: got %s want %s", documentID, workOrderID)
	}
	if documentType != "work_order" || documentTitle != "Legacy work order" || documentStatus != "draft" {
		t.Fatalf("unexpected backfilled document values: type=%s title=%q status=%s", documentType, documentTitle, documentStatus)
	}

	if _, err := db.ExecContext(ctx, downSQL(t, "sql/000015_milestone5_work_order_document_ownership.down.sql")); err != nil {
		t.Fatalf("reapply down migration: %v", err)
	}

	var payloadCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'work_orders' AND table_name = 'documents'`).Scan(&payloadCount); err != nil {
		t.Fatalf("check payload table dropped: %v", err)
	}
	if payloadCount != 0 {
		t.Fatalf("expected work_orders.documents to be dropped, count=%d", payloadCount)
	}

	var documentCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents.documents WHERE id = $1`, workOrderID).Scan(&documentCount); err != nil {
		t.Fatalf("count rolled back document rows: %v", err)
	}
	if documentCount != 0 {
		t.Fatalf("expected rolled back work-order document row to be removed, count=%d", documentCount)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM platform.schema_migrations WHERE version = '000015_milestone5_work_order_document_ownership'`); err != nil {
		t.Fatalf("delete migration record before restore: %v", err)
	}
	if err := applyMigration(ctx, db, target); err != nil {
		t.Fatalf("restore 000015 migration after rollback verification: %v", err)
	}
}

func TestAccountingDocumentOwnershipMigrationBackfillsAndRollsBackIntegration(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	resetTestDB(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, downSQL(t, "sql/000016_milestone5_accounting_document_ownership.down.sql")); err != nil {
		t.Fatalf("apply down migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM platform.schema_migrations WHERE version = '000016_milestone5_accounting_document_ownership'`); err != nil {
		t.Fatalf("delete migration record: %v", err)
	}

	orgID, userID := insertLegacyOrgAndUser(t, ctx, db)

	var invoiceID string
	if err := db.QueryRowContext(ctx, `
INSERT INTO documents.documents (
	org_id,
	type_code,
	status,
	title,
	created_by_user_id
) VALUES ($1, 'invoice', 'draft', 'Legacy invoice', $2)
RETURNING id;`,
		orgID,
		userID,
	).Scan(&invoiceID); err != nil {
		t.Fatalf("insert legacy invoice document: %v", err)
	}

	var paymentReceiptID string
	if err := db.QueryRowContext(ctx, `
INSERT INTO documents.documents (
	org_id,
	type_code,
	status,
	title,
	created_by_user_id
) VALUES ($1, 'payment_receipt', 'draft', 'Legacy receipt', $2)
RETURNING id;`,
		orgID,
		userID,
	).Scan(&paymentReceiptID); err != nil {
		t.Fatalf("insert legacy payment receipt document: %v", err)
	}

	migrations, err := loadUpMigrations()
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}

	var target migration
	found := false
	for _, m := range migrations {
		if m.version == "000016_milestone5_accounting_document_ownership" {
			target = m
			found = true
			break
		}
	}
	if !found {
		t.Fatal("did not find 000016 migration")
	}

	if err := applyMigration(ctx, db, target); err != nil {
		t.Fatalf("apply 000016 migration: %v", err)
	}

	var invoicePayloadCount int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM accounting.invoice_documents
WHERE org_id = $1
  AND document_id = $2;`,
		orgID,
		invoiceID,
	).Scan(&invoicePayloadCount); err != nil {
		t.Fatalf("count backfilled invoice payload rows: %v", err)
	}
	if invoicePayloadCount != 1 {
		t.Fatalf("unexpected backfilled invoice payload count: %d", invoicePayloadCount)
	}

	var paymentPayloadCount int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM accounting.payment_receipt_documents
WHERE org_id = $1
  AND document_id = $2;`,
		orgID,
		paymentReceiptID,
	).Scan(&paymentPayloadCount); err != nil {
		t.Fatalf("count backfilled payment receipt payload rows: %v", err)
	}
	if paymentPayloadCount != 1 {
		t.Fatalf("unexpected backfilled payment receipt payload count: %d", paymentPayloadCount)
	}

	if _, err := db.ExecContext(ctx, downSQL(t, "sql/000016_milestone5_accounting_document_ownership.down.sql")); err != nil {
		t.Fatalf("reapply down migration: %v", err)
	}

	var invoiceTableCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'accounting' AND table_name = 'invoice_documents'`).Scan(&invoiceTableCount); err != nil {
		t.Fatalf("check invoice payload table dropped: %v", err)
	}
	if invoiceTableCount != 0 {
		t.Fatalf("expected accounting.invoice_documents to be dropped, count=%d", invoiceTableCount)
	}

	var paymentTableCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'accounting' AND table_name = 'payment_receipt_documents'`).Scan(&paymentTableCount); err != nil {
		t.Fatalf("check payment receipt payload table dropped: %v", err)
	}
	if paymentTableCount != 0 {
		t.Fatalf("expected accounting.payment_receipt_documents to be dropped, count=%d", paymentTableCount)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM platform.schema_migrations WHERE version = '000016_milestone5_accounting_document_ownership'`); err != nil {
		t.Fatalf("delete migration record before restore: %v", err)
	}
	if err := applyMigration(ctx, db, target); err != nil {
		t.Fatalf("restore 000016 migration after rollback verification: %v", err)
	}
}

func downSQL(t *testing.T, path string) string {
	t.Helper()

	body, err := migrationFiles.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

func insertLegacyOrgAndUser(t *testing.T, ctx context.Context, db *sql.DB) (string, string) {
	t.Helper()

	var orgID string
	if err := db.QueryRowContext(ctx, `
INSERT INTO identityaccess.orgs (slug, name)
VALUES ('legacy-acme', 'Legacy Acme')
RETURNING id;`).Scan(&orgID); err != nil {
		t.Fatalf("insert org: %v", err)
	}

	var userID string
	if err := db.QueryRowContext(ctx, `
INSERT INTO identityaccess.users (email, display_name)
VALUES ('legacy-user@example.com', 'Legacy User')
RETURNING id;`).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	return orgID, userID
}

const testDatabaseLockKey int64 = 20260319

func openTestDB(t *testing.T) *sql.DB {
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

	if _, err := Up(ctx, db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return db
}

func resetTestDB(t *testing.T, db *sql.DB) {
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
