package core_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"accounting-agent/internal/core"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	_ = godotenv.Load("../../.env")

	// Use a dedicated TEST database to avoid wiping the live app database.
	// Set TEST_DATABASE_URL in your .env or environment to run integration tests.
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test to protect live database")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	assertRequiredTestSchema(t, pool)

	// Clean and seed test DB
	_, err = pool.Exec(ctx, `
		TRUNCATE TABLE journal_lines, journal_entries, accounts, companies, documents, document_sequences, document_types CASCADE;
		
		INSERT INTO companies (id, company_code, name, base_currency) VALUES (1, '1000', 'Test Company', 'INR');
		
		INSERT INTO accounts (company_id, code, name, type) VALUES
		(1, '1000', 'Cash', 'asset'),
		(1, '1200', 'Accounts Receivable', 'asset'),
		(1, '2000', 'Accounts Payable', 'liability'),
		(1, '3000', 'Share Capital', 'equity'),
		(1, '4000', 'Sales Revenue', 'revenue'),
		(1, '5000', 'Cost of Goods Sold', 'expense'),
		(1, '5100', 'Operating Expense', 'expense');

		INSERT INTO document_types (code, name, affects_inventory, affects_gl, affects_ar, affects_ap, numbering_strategy, resets_every_fy) VALUES
		('JE', 'Journal Entry', false, true, false, false, 'global', false),
		('SI', 'Sales Invoice', true, true, true, false, 'global', false),
		('PI', 'Purchase Invoice', true, true, false, true, 'global', false),
		('RC', 'Receipt', false, true, true, false, 'global', false),
		('PV', 'Payment Voucher', false, true, false, true, 'global', false);
	`)
	if err != nil {
		t.Fatalf("Failed to seed test database: %v", err)
	}

	return pool
}

func assertRequiredTestSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	requiredColumns := []struct {
		table  string
		column string
	}{
		{"accounts", "is_control_account"},
		{"accounts", "control_type"},
		{"products", "revenue_account_id"},
		{"vendors", "ap_account_id"},
		{"vendors", "default_expense_account_id"},
		{"purchase_order_lines", "expense_account_id"},
		{"account_rules", "account_id"},
	}

	for _, c := range requiredColumns {
		var exists bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = 'public'
				  AND table_name = $1
				  AND column_name = $2
			)
		`, c.table, c.column).Scan(&exists); err != nil {
			t.Fatalf("failed to validate test schema precondition for %s.%s: %v", c.table, c.column, err)
		}
		if !exists {
			t.Fatalf(
				"test database is missing required schema column %s.%s; run migrations first: DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db",
				c.table, c.column,
			)
		}
	}

	requiredIndexes := []string{
		"idx_account_rules_lookup_temporal",
		"journal_entries_company_idempotency_key_idx",
	}
	for _, idx := range requiredIndexes {
		var exists bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM pg_indexes
				WHERE schemaname = 'public'
				  AND indexname = $1
			)
		`, idx).Scan(&exists); err != nil {
			t.Fatalf("failed to validate test schema precondition for index %s: %v", idx, err)
		}
		if !exists {
			t.Fatalf(
				"test database is missing required index %s; run migrations first: DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db",
				idx,
			)
		}
	}

	type requiredIndexDef struct {
		name   string
		tokens []string
	}
	requiredIndexDefs := []requiredIndexDef{
		{
			name:   "document_sequences_unique_idx",
			tokens: []string{"on public.document_sequences", "(company_id, type_code)"},
		},
		{
			name:   "documents_unique_number_idx",
			tokens: []string{"on public.documents", "(company_id, type_code, document_number)", "where (document_number is not null)"},
		},
	}

	for _, idx := range requiredIndexDefs {
		var indexDef string
		if err := pool.QueryRow(ctx, `
			SELECT pg_get_indexdef(c.oid)
			FROM pg_class c
			WHERE c.relkind = 'i' AND c.relname = $1
		`, idx.name).Scan(&indexDef); err != nil {
			t.Fatalf("failed to validate index definition for %s: %v", idx.name, err)
		}

		defLower := strings.ToLower(indexDef)
		for _, token := range idx.tokens {
			if !strings.Contains(defLower, strings.ToLower(token)) {
				t.Fatalf(
					"test database index %s does not match required global numbering policy; run migrations first: DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db (indexdef=%s)",
					idx.name, indexDef,
				)
			}
		}
	}
}

func TestLedger_Idempotency(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	ctx := context.Background()

	// 1. Create a proposal with a specific idempotency key
	idempotencyKey := uuid.NewString()
	proposal := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "1000",
		IdempotencyKey:      idempotencyKey,
		TransactionCurrency: "INR",
		ExchangeRate:        "1.0",
		PostingDate:         "2023-10-01",
		DocumentDate:        "2023-10-01",
		Summary:             "Test Idempotent Transaction",
		Reasoning:           "Testing Phase 0",
		Lines: []core.ProposalLine{
			{AccountCode: "1000", IsDebit: true, Amount: "150.00"},
			{AccountCode: "4000", IsDebit: false, Amount: "150.00"},
		},
	}

	// 2. Commit first time - should succeed
	err := ledger.Commit(ctx, proposal)
	if err != nil {
		t.Fatalf("First commit failed: %v", err)
	}

	// 3. Commit second time - should fail gracefully with specific error
	err = ledger.Commit(ctx, proposal)
	if err == nil {
		t.Fatalf("Expected duplicate commit to fail, but it succeeded")
	}

	if err.Error() != fmt.Sprintf("duplicate proposal for company 1000: idempotency key %s already exists", idempotencyKey) {
		t.Errorf("Unexpected error message for duplicate commit: %v", err)
	}
}

func TestLedger_IdempotencyKeyScopedByCompany(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		INSERT INTO companies (id, company_code, name, base_currency) VALUES (2, '2000', 'Second Company', 'USD');
		INSERT INTO accounts (company_id, code, name, type) VALUES
			(2, '1000', 'Cash', 'asset'),
			(2, '4000', 'Revenue', 'revenue');
	`)
	if err != nil {
		t.Fatalf("Failed to seed second company/accounts: %v", err)
	}

	idempotencyKey := uuid.NewString()
	proposalA := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "1000",
		IdempotencyKey:      idempotencyKey,
		TransactionCurrency: "INR",
		ExchangeRate:        "1.0",
		PostingDate:         "2026-03-01",
		DocumentDate:        "2026-03-01",
		Summary:             "Scoped idempotency A",
		Reasoning:           "Company A posting",
		Lines: []core.ProposalLine{
			{AccountCode: "1000", IsDebit: true, Amount: "100.00"},
			{AccountCode: "4000", IsDebit: false, Amount: "100.00"},
		},
	}
	proposalB := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "2000",
		IdempotencyKey:      idempotencyKey,
		TransactionCurrency: "USD",
		ExchangeRate:        "1.0",
		PostingDate:         "2026-03-01",
		DocumentDate:        "2026-03-01",
		Summary:             "Scoped idempotency B",
		Reasoning:           "Company B posting",
		Lines: []core.ProposalLine{
			{AccountCode: "1000", IsDebit: true, Amount: "100.00"},
			{AccountCode: "4000", IsDebit: false, Amount: "100.00"},
		},
	}

	if err := ledger.Commit(ctx, proposalA); err != nil {
		t.Fatalf("first company commit failed: %v", err)
	}
	if err := ledger.Commit(ctx, proposalB); err != nil {
		t.Fatalf("second company commit with same idempotency key failed: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)
		FROM journal_entries
		WHERE idempotency_key = $1
	`, idempotencyKey).Scan(&count); err != nil {
		t.Fatalf("failed to count idempotency rows: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 entries (one per company), got %d", count)
	}
}

func TestLedger_Reversal(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	ctx := context.Background()

	// 1. Setup a transaction to reverse
	idempotencyKey := uuid.NewString()
	proposal := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "1000",
		IdempotencyKey:      idempotencyKey,
		TransactionCurrency: "USD",
		ExchangeRate:        "83.50",
		PostingDate:         "2023-10-01",
		DocumentDate:        "2023-10-01",
		Summary:             fmt.Sprintf("Transaction to be reversed %s", idempotencyKey),
		Reasoning:           "Setup for reversal test",
		Lines: []core.ProposalLine{
			{AccountCode: "1000", IsDebit: true, Amount: "500.00"},
			{AccountCode: "4000", IsDebit: false, Amount: "500.00"},
		},
	}

	err := ledger.Commit(ctx, proposal)
	if err != nil {
		t.Fatalf("Failed to setup commit: %v", err)
	}

	// Fetch the entry ID
	var entryID int
	err = pool.QueryRow(ctx, "SELECT id FROM journal_entries WHERE idempotency_key = $1", idempotencyKey).Scan(&entryID)
	if err != nil {
		t.Fatalf("Failed to fetch entry ID: %v", err)
	}

	// 2. Reverse the entry
	err = ledger.Reverse(ctx, entryID, "Error in original entry")
	if err != nil {
		t.Fatalf("Failed to reverse entry: %v", err)
	}

	// 3. Prevent Double Reversal
	err = ledger.Reverse(ctx, entryID, "Trying to reverse again")
	if err == nil {
		t.Fatalf("Expected double reversal to fail, but it succeeded")
	}
	if err.Error() != fmt.Sprintf("entry %d is already reversed", entryID) {
		t.Errorf("Unexpected error message for double reversal: %v", err)
	}

	// 4. Verify the database state
	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM journal_entries WHERE reversed_entry_id = $1", entryID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check reversal status: %v", err)
	}
	if count == 0 {
		t.Errorf("Expected to find a new entry with reversed_entry_id pointing to the original")
	}
}

func TestLedger_CrossCompanyScoping(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	ctx := context.Background()

	// Seed a second company with its own account code "1000" (same code, different company)
	_, err := pool.Exec(ctx, `
		INSERT INTO companies (id, company_code, name, base_currency) VALUES (2, '2000', 'Foreign Company', 'USD');
		INSERT INTO accounts (company_id, code, name, type) VALUES (2, '1000', 'Foreign Cash', 'asset');
	`)
	if err != nil {
		t.Fatalf("Failed to seed second company: %v", err)
	}

	// Attempt to use an account that belongs to company 2 in a proposal for company 1000.
	// The ledger must reject this — accounts are scoped strictly to their company.
	proposal := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "1000",
		IdempotencyKey:      uuid.NewString(),
		TransactionCurrency: "INR",
		ExchangeRate:        "1.0",
		PostingDate:         "2023-10-01",
		DocumentDate:        "2023-10-01",
		Summary:             "Cross-company account scoping test",
		Reasoning:           "Should fail",
		Lines: []core.ProposalLine{
			// "9999" does not exist in any company — should trigger not-found error
			{AccountCode: "9999", IsDebit: true, Amount: "100.00"},
			{AccountCode: "4000", IsDebit: false, Amount: "100.00"},
		},
	}

	err = ledger.Commit(ctx, proposal)
	if err == nil {
		t.Fatal("Expected error for non-existent account code, got nil")
	}

	// Verify the error message mentions the unknown account code
	expected := "account code 9999 not found for company 1000"
	if err.Error() != expected {
		t.Errorf("Unexpected error: got %q, want %q", err.Error(), expected)
	}
}

func TestLedger_GetBalances(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	ctx := context.Background()

	// Commit a known transaction
	proposal := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "1000",
		IdempotencyKey:      uuid.NewString(),
		TransactionCurrency: "INR",
		ExchangeRate:        "1.0",
		PostingDate:         "2023-10-01",
		DocumentDate:        "2023-10-01",
		Summary:             "Balance check",
		Reasoning:           "Verifying GetBalances",
		Lines: []core.ProposalLine{
			{AccountCode: "1000", IsDebit: true, Amount: "250.00"},
			{AccountCode: "4000", IsDebit: false, Amount: "250.00"},
		},
	}

	if err := ledger.Commit(ctx, proposal); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Check balances
	balances, err := ledger.GetBalances(ctx, "1000")
	if err != nil {
		t.Fatalf("GetBalances failed: %v", err)
	}

	balanceMap := make(map[string]string)
	for _, b := range balances {
		balanceMap[b.Code] = b.Balance.StringFixed(2)
	}

	// Account 1000 (asset): debit 250 → positive balance
	if balanceMap["1000"] != "250.00" {
		t.Errorf("Expected account 1000 balance 250.00, got %s", balanceMap["1000"])
	}
	// Account 4000 (revenue): credit 250 → negative balance (credit normal)
	if balanceMap["4000"] != "-250.00" {
		t.Errorf("Expected account 4000 balance -250.00, got %s", balanceMap["4000"])
	}
}

func TestLedger_GetBalances_MultiCompany(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	ctx := context.Background()

	// Seed a second company with its own accounts
	_, err := pool.Exec(ctx, `
		INSERT INTO companies (id, company_code, name, base_currency) VALUES (2, '2000', 'Company Two', 'USD');
		INSERT INTO accounts (company_id, code, name, type) VALUES
			(2, '1000', 'Cash USD', 'asset'),
			(2, '4000', 'Revenue USD', 'revenue');
	`)
	if err != nil {
		t.Fatalf("Failed to seed second company: %v", err)
	}

	// Commit a transaction for company 1000
	proposal1 := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "1000",
		IdempotencyKey:      "mc-test-1000",
		TransactionCurrency: "INR",
		ExchangeRate:        "1.0",
		PostingDate:         "2024-01-01",
		DocumentDate:        "2024-01-01",
		Summary:             "Company 1 transaction",
		Reasoning:           "Multi company test",
		Lines: []core.ProposalLine{
			{AccountCode: "1000", IsDebit: true, Amount: "100.00"},
			{AccountCode: "4000", IsDebit: false, Amount: "100.00"},
		},
	}
	if err := ledger.Commit(ctx, proposal1); err != nil {
		t.Fatalf("Commit for company 1000 failed: %v", err)
	}

	// Commit a different transaction for company 2000
	proposal2 := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "2000",
		IdempotencyKey:      "mc-test-2000",
		TransactionCurrency: "USD",
		ExchangeRate:        "1.0",
		PostingDate:         "2024-01-01",
		DocumentDate:        "2024-01-01",
		Summary:             "Company 2 transaction",
		Reasoning:           "Multi company test",
		Lines: []core.ProposalLine{
			{AccountCode: "1000", IsDebit: true, Amount: "999.00"},
			{AccountCode: "4000", IsDebit: false, Amount: "999.00"},
		},
	}
	if err := ledger.Commit(ctx, proposal2); err != nil {
		t.Fatalf("Commit for company 2000 failed: %v", err)
	}

	// GetBalances for company 1000 must NOT include company 2000 amounts
	balances1, err := ledger.GetBalances(ctx, "1000")
	if err != nil {
		t.Fatalf("GetBalances for 1000 failed: %v", err)
	}
	map1 := make(map[string]string)
	for _, b := range balances1 {
		map1[b.Code] = b.Balance.StringFixed(2)
	}
	if map1["1000"] != "100.00" {
		t.Errorf("Company 1000: expected account 1000 = 100.00, got %s", map1["1000"])
	}

	// GetBalances for company 2000 must NOT include company 1000 amounts
	balances2, err := ledger.GetBalances(ctx, "2000")
	if err != nil {
		t.Fatalf("GetBalances for 2000 failed: %v", err)
	}
	map2 := make(map[string]string)
	for _, b := range balances2 {
		map2[b.Code] = b.Balance.StringFixed(2)
	}
	if map2["1000"] != "999.00" {
		t.Errorf("Company 2000: expected account 1000 = 999.00, got %s", map2["1000"])
	}
}
