package app

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"accounting-agent/internal/core"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupAppTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test to protect live database")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}

	assertAppTestSchema(t, pool)

	if _, err := pool.Exec(ctx, `
		TRUNCATE TABLE purchase_orders, vendors, companies CASCADE;

		INSERT INTO companies (id, company_code, name, base_currency)
		VALUES (1, '1000', 'Test Company', 'INR');

		INSERT INTO vendors (id, company_id, code, name)
		VALUES
		  (1, 1, 'V001', 'Vendor One'),
		  (2, 1, 'V002', 'Vendor Two');
	`); err != nil {
		t.Fatalf("seed app integration test data: %v", err)
	}

	return pool
}

func assertAppTestSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	requiredColumns := []struct {
		table  string
		column string
	}{
		{"accounts", "is_control_account"},
		{"accounts", "control_type"},
		{"purchase_orders", "invoice_amount"},
		{"purchase_orders", "paid_at"},
		{"purchase_orders", "currency"},
		{"purchase_orders", "exchange_rate"},
		{"manual_je_control_account_audits", "warning_details"},
		{"manual_je_control_account_audits", "enforcement_mode"},
		{"manual_je_control_account_audits", "override_control_accounts"},
		{"manual_je_control_account_audits", "override_reason"},
		{"manual_je_control_account_audits", "is_blocked"},
		{"document_type_policy_violation_audits", "policy_mode"},
		{"document_type_policy_violation_audits", "is_enforced"},
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
			t.Fatalf("validate app test schema %s.%s: %v", c.table, c.column, err)
		}
		if !exists {
			t.Fatalf(
				"test database is missing required schema column %s.%s; run migrations first: DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db",
				c.table, c.column,
			)
		}
	}
}

func TestAppService_GetOutstandingVendorInvoicesJSON_MultiCurrency(t *testing.T) {
	pool := setupAppTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	if _, err := pool.Exec(ctx, `
		INSERT INTO purchase_orders (
			id, company_id, vendor_id, status, po_date, currency, exchange_rate,
			total_transaction, total_base, invoice_amount
		) VALUES
		  (101, 1, 1, 'INVOICED', CURRENT_DATE, 'INR', 1, 5000.00, 5000.00, 5000.00),
		  (102, 1, 1, 'INVOICED', CURRENT_DATE, 'USD', 80, 200.00, 16000.00, 250.00),
		  (103, 1, 2, 'PAID',     CURRENT_DATE, 'USD', 80, 100.00, 8000.00, 100.00);
	`); err != nil {
		t.Fatalf("seed purchase_orders: %v", err)
	}

	svc := &appService{pool: pool}
	payload, err := svc.getOutstandingVendorInvoicesJSON(ctx, "1000", "")
	if err != nil {
		t.Fatalf("getOutstandingVendorInvoicesJSON: %v", err)
	}

	type currencyBucket struct {
		Currency          string `json:"currency"`
		AmountTransaction string `json:"amount_transaction"`
		InvoiceCount      int    `json:"invoice_count"`
	}
	type out struct {
		BaseCurrency                string           `json:"base_currency"`
		OutstandingInvoiceTotal     string           `json:"outstanding_invoice_total"`
		OutstandingInvoiceTotalBase string           `json:"outstanding_invoice_total_base"`
		OutstandingInvoiceCount     int              `json:"outstanding_invoice_count"`
		OutstandingByCurrency       []currencyBucket `json:"outstanding_by_currency"`
	}
	var got out
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatalf("unmarshal outstanding JSON: %v; payload=%s", err, payload)
	}

	if got.BaseCurrency != "INR" {
		t.Fatalf("expected base_currency INR, got %s", got.BaseCurrency)
	}
	if got.OutstandingInvoiceCount != 2 {
		t.Fatalf("expected outstanding_invoice_count 2, got %d", got.OutstandingInvoiceCount)
	}
	if got.OutstandingInvoiceTotalBase != "25000.00" {
		t.Fatalf("expected outstanding_invoice_total_base 25000.00, got %s", got.OutstandingInvoiceTotalBase)
	}
	if got.OutstandingInvoiceTotal != "25000.00" {
		t.Fatalf("expected outstanding_invoice_total alias 25000.00, got %s", got.OutstandingInvoiceTotal)
	}

	byCurrency := map[string]currencyBucket{}
	for _, b := range got.OutstandingByCurrency {
		byCurrency[b.Currency] = b
	}
	if byCurrency["INR"].AmountTransaction != "5000.00" || byCurrency["INR"].InvoiceCount != 1 {
		t.Fatalf("unexpected INR bucket: %+v", byCurrency["INR"])
	}
	if byCurrency["USD"].AmountTransaction != "250.00" || byCurrency["USD"].InvoiceCount != 1 {
		t.Fatalf("unexpected USD bucket: %+v", byCurrency["USD"])
	}
}

func TestAppService_GetVendorPaymentHistoryJSON_CurrencySemantics(t *testing.T) {
	pool := setupAppTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	paidAtRecent := time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC)
	paidAtOlder := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)
	if _, err := pool.Exec(ctx, `
		INSERT INTO purchase_orders (
			id, company_id, vendor_id, po_number, status, po_date, currency, exchange_rate,
			total_transaction, total_base, invoice_number, invoice_date, invoice_amount,
			paid_at, pi_document_number
		) VALUES
		  (201, 1, 1, 'PO-201', 'PAID', CURRENT_DATE, 'USD', 82, 120.00, 9840.00, 'INV-201', CURRENT_DATE, 100.00, $1, 'PI-201'),
		  (202, 1, 1, 'PO-202', 'PAID', CURRENT_DATE, 'INR', 1, 3000.00, 3000.00, 'INV-202', CURRENT_DATE, NULL,   $2, 'PI-202'),
		  (203, 1, 2, 'PO-203', 'PAID', CURRENT_DATE, 'USD', 80, 100.00, 8000.00, 'INV-203', CURRENT_DATE, 100.00, $1, 'PI-203');
	`, paidAtRecent, paidAtOlder); err != nil {
		t.Fatalf("seed paid purchase_orders: %v", err)
	}

	svc := &appService{pool: pool}
	payload, err := svc.getVendorPaymentHistoryJSON(ctx, "1000", "V001")
	if err != nil {
		t.Fatalf("getVendorPaymentHistoryJSON: %v", err)
	}

	type paymentRecord struct {
		POID                     int     `json:"po_id"`
		Currency                 string  `json:"currency"`
		ExchangeRate             string  `json:"exchange_rate"`
		InvoiceAmountTransaction *string `json:"invoice_amount_transaction"`
		InvoiceAmountBase        *string `json:"invoice_amount_base"`
		POTotalTransaction       string  `json:"po_total_transaction"`
		POTotalBase              string  `json:"po_total_base"`
	}
	type out struct {
		BaseCurrency   string          `json:"base_currency"`
		VendorCode     string          `json:"vendor_code"`
		PaymentCount   int             `json:"payment_count"`
		PaymentHistory []paymentRecord `json:"payment_history"`
	}
	var got out
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatalf("unmarshal payment history JSON: %v; payload=%s", err, payload)
	}

	if got.BaseCurrency != "INR" {
		t.Fatalf("expected base_currency INR, got %s", got.BaseCurrency)
	}
	if got.VendorCode != "V001" {
		t.Fatalf("expected vendor_code V001, got %s", got.VendorCode)
	}
	if got.PaymentCount != 2 {
		t.Fatalf("expected payment_count 2, got %d", got.PaymentCount)
	}

	var usdRecord, inrRecord *paymentRecord
	for i := range got.PaymentHistory {
		r := &got.PaymentHistory[i]
		switch r.POID {
		case 201:
			usdRecord = r
		case 202:
			inrRecord = r
		}
	}
	if usdRecord == nil || inrRecord == nil {
		t.Fatalf("expected payment records for POIDs 201 and 202, got %+v", got.PaymentHistory)
	}

	if usdRecord.Currency != "USD" || usdRecord.ExchangeRate != "82" {
		t.Fatalf("unexpected USD record currency/rate: %+v", usdRecord)
	}
	if usdRecord.InvoiceAmountTransaction == nil || *usdRecord.InvoiceAmountTransaction != "100.00" {
		t.Fatalf("unexpected USD invoice_amount_transaction: %+v", usdRecord.InvoiceAmountTransaction)
	}
	if usdRecord.InvoiceAmountBase == nil || *usdRecord.InvoiceAmountBase != "8200.00" {
		t.Fatalf("unexpected USD invoice_amount_base: %+v", usdRecord.InvoiceAmountBase)
	}

	if inrRecord.Currency != "INR" || inrRecord.ExchangeRate != "1" {
		t.Fatalf("unexpected INR record currency/rate: %+v", inrRecord)
	}
	if inrRecord.InvoiceAmountTransaction != nil || inrRecord.InvoiceAmountBase != nil {
		t.Fatalf("expected nil invoice amounts when invoice_amount is NULL, got %+v", inrRecord)
	}
}

func TestAppService_ManualJEControlAccountWarningsAndAudit(t *testing.T) {
	pool := setupAppTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	if _, err := pool.Exec(ctx, `
		INSERT INTO accounts (company_id, code, name, type, is_control_account, control_type)
		VALUES
		  (1, '1200', 'Accounts Receivable', 'asset', true, 'AR'),
		  (1, '4000', 'Sales Revenue', 'revenue', false, NULL);
	`); err != nil {
		t.Fatalf("seed accounts: %v", err)
	}

	svc := &appService{pool: pool}
	warnings, err := svc.GetManualJEControlAccountWarnings(ctx, "1000", []string{"1200", "4000"})
	if err != nil {
		t.Fatalf("GetManualJEControlAccountWarnings: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].AccountCode != "1200" || warnings[0].ControlType != "AR" {
		t.Fatalf("unexpected warning payload: %+v", warnings[0])
	}

	if err := svc.RecordManualJEControlAccountAttempt(ctx, ManualJEControlAccountAttemptRequest{
		CompanyCode:    "1000",
		UserID:         nil,
		Username:       "test.user",
		Action:         "validate",
		PostingDate:    "2026-03-05",
		Narration:      "manual AR adjustment",
		AccountCodes:   []string{"1200", "4000"},
		WarningDetails: warnings,
	}); err != nil {
		t.Fatalf("RecordManualJEControlAccountAttempt: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM manual_je_control_account_audits WHERE company_id = 1`).Scan(&count); err != nil {
		t.Fatalf("count audit rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 audit row, got %d", count)
	}
}

func TestAppService_SharedControlAccountPolicy_LowRiskCompatibility(t *testing.T) {
	pool := setupAppTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	if _, err := pool.Exec(ctx, `
		INSERT INTO accounts (company_id, code, name, type, is_control_account, control_type)
		VALUES
		  (1, '1200', 'Accounts Receivable', 'asset', true, 'AR'),
		  (1, '4000', 'Sales Revenue', 'revenue', false, NULL)
	`); err != nil {
		t.Fatalf("seed accounts: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO document_types (code, name, affects_inventory, affects_gl, affects_ar, affects_ap, numbering_strategy, resets_every_fy)
		VALUES ('JE', 'Journal Entry', false, true, false, false, 'global', false)
		ON CONFLICT (code) DO NOTHING
	`); err != nil {
		t.Fatalf("seed JE document type: %v", err)
	}

	ledger := core.NewLedger(pool, core.NewDocumentService(pool))
	svc := &appService{pool: pool, ledger: ledger}

	proposal := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "1000",
		IdempotencyKey:      "test-shared-policy-1",
		TransactionCurrency: "INR",
		ExchangeRate:        "1.0",
		PostingDate:         "2026-03-05",
		DocumentDate:        "2026-03-05",
		Summary:             "test JE",
		Reasoning:           "test",
		Lines: []core.ProposalLine{
			{AccountCode: "1200", IsDebit: true, Amount: "100.00"},
			{AccountCode: "4000", IsDebit: false, Amount: "100.00"},
		},
	}

	prevMode := os.Getenv("CONTROL_ACCOUNT_ENFORCEMENT_MODE")
	if err := os.Setenv("CONTROL_ACCOUNT_ENFORCEMENT_MODE", "enforce"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer func() {
		if prevMode == "" {
			_ = os.Unsetenv("CONTROL_ACCOUNT_ENFORCEMENT_MODE")
		} else {
			_ = os.Setenv("CONTROL_ACCOUNT_ENFORCEMENT_MODE", prevMode)
		}
	}()

	// Explicit manual_web source should enforce and fail without override.
	err := svc.ValidateProposal(WithProposalSource(ctx, ProposalSourceManualWeb), proposal)
	if err == nil || !strings.Contains(err.Error(), "CONTROL_ACCOUNT_ENFORCED") {
		t.Fatalf("expected CONTROL_ACCOUNT_ENFORCED error, got: %v", err)
	}
	lowercaseDocProposal := proposal
	lowercaseDocProposal.DocumentTypeCode = "je"
	lowercaseDocProposal.IdempotencyKey = "test-shared-policy-1-lowercase-doc"
	err = svc.ValidateProposal(WithProposalSource(ctx, ProposalSourceManualWeb), lowercaseDocProposal)
	if err == nil || !strings.Contains(err.Error(), "CONTROL_ACCOUNT_ENFORCED") {
		t.Fatalf("expected CONTROL_ACCOUNT_ENFORCED for lowercase JE, got: %v", err)
	}

	// AI source should remain non-blocking in low-risk mode.
	aiCtx := WithProposalSource(ctx, ProposalSourceAIAgent)
	if err := svc.ValidateProposal(aiCtx, proposal); err != nil {
		t.Fatalf("expected AI source validation to pass, got: %v", err)
	}

	// Manual source with admin override should pass.
	overrideCtx := WithControlAccountOverride(
		WithProposalSource(ctx, ProposalSourceManualWeb),
		true, "emergency correction", "ADMIN",
	)
	if err := svc.ValidateProposal(overrideCtx, proposal); err != nil {
		t.Fatalf("expected manual admin override validation to pass, got: %v", err)
	}
}

func TestAppService_DocumentTypePolicy_ModeBehavior(t *testing.T) {
	pool := setupAppTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	if _, err := pool.Exec(ctx, `
		INSERT INTO accounts (company_id, code, name, type)
		VALUES
		  (1, '1100', 'Bank Account', 'asset'),
		  (1, '1200', 'Accounts Receivable', 'asset')
		ON CONFLICT (company_id, code) DO NOTHING
	`); err != nil {
		t.Fatalf("seed accounts: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO document_types (code, name, affects_inventory, affects_gl, affects_ar, affects_ap, numbering_strategy, resets_every_fy)
		VALUES
		  ('JE', 'Journal Entry', false, true, false, false, 'global', false),
		  ('RC', 'Receipt', false, true, true, false, 'global', false)
		ON CONFLICT (code) DO NOTHING
	`); err != nil {
		t.Fatalf("seed document types: %v", err)
	}

	svc := &appService{pool: pool, ledger: core.NewLedger(pool, core.NewDocumentService(pool))}
	proposal := core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         "1000",
		IdempotencyKey:      "payment-order-101",
		TransactionCurrency: "INR",
		ExchangeRate:        "1.0",
		PostingDate:         "2026-03-05",
		DocumentDate:        "2026-03-05",
		Summary:             "customer payment received",
		Reasoning:           "test",
		Lines: []core.ProposalLine{
			{AccountCode: "1100", IsDebit: true, Amount: "100.00"},
			{AccountCode: "1200", IsDebit: false, Amount: "100.00"},
		},
	}

	prevMode := os.Getenv("DOCUMENT_TYPE_POLICY_MODE")
	defer func() {
		if prevMode == "" {
			_ = os.Unsetenv("DOCUMENT_TYPE_POLICY_MODE")
		} else {
			_ = os.Setenv("DOCUMENT_TYPE_POLICY_MODE", prevMode)
		}
	}()

	if err := os.Setenv("DOCUMENT_TYPE_POLICY_MODE", "warn"); err != nil {
		t.Fatalf("set warn mode: %v", err)
	}
	if err := svc.ValidateProposal(WithProposalSource(ctx, ProposalSourceManualWeb), proposal); err != nil {
		t.Fatalf("expected warn mode to pass, got: %v", err)
	}
	var warnAuditCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM document_type_policy_violation_audits
		WHERE company_id = 1
		  AND policy_mode = 'warn'
		  AND is_enforced = false
		  AND idempotency_key = $1
	`, proposal.IdempotencyKey).Scan(&warnAuditCount); err != nil {
		t.Fatalf("query warn audit rows: %v", err)
	}
	if warnAuditCount < 1 {
		t.Fatalf("expected at least one warn audit row, found %d", warnAuditCount)
	}

	if err := os.Setenv("DOCUMENT_TYPE_POLICY_MODE", "enforce"); err != nil {
		t.Fatalf("set enforce mode: %v", err)
	}
	err := svc.ValidateProposal(WithProposalSource(ctx, ProposalSourceManualWeb), proposal)
	if err == nil || !strings.Contains(err.Error(), "DOCUMENT_TYPE_POLICY_ENFORCED") {
		t.Fatalf("expected DOCUMENT_TYPE_POLICY_ENFORCED error, got: %v", err)
	}
	err = svc.ValidateProposal(WithProposalSource(ctx, ProposalSourceAIAgent), proposal)
	if err == nil || !strings.Contains(err.Error(), "DOCUMENT_TYPE_POLICY_ENFORCED") {
		t.Fatalf("expected AI path DOCUMENT_TYPE_POLICY_ENFORCED error, got: %v", err)
	}
	var enforceAuditCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM document_type_policy_violation_audits
		WHERE company_id = 1
		  AND policy_mode = 'enforce'
		  AND is_enforced = true
		  AND idempotency_key = $1
	`, proposal.IdempotencyKey).Scan(&enforceAuditCount); err != nil {
		t.Fatalf("query enforce audit rows: %v", err)
	}
	if enforceAuditCount < 2 {
		t.Fatalf("expected at least two enforce audit rows (manual + ai), found %d", enforceAuditCount)
	}
}
