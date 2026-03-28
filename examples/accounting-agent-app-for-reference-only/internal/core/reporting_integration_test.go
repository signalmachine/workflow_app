package core_test

import (
	"context"
	"testing"

	"accounting-agent/internal/core"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestReporting_GetProfitAndLoss(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	reporting := core.NewReportingService(pool)
	ctx := context.Background()

	// Post Jan 2026 revenue and expense entries.
	// Entry 1: DR Cash 1000 / CR Sales Revenue 1000  → Jan revenue = 1000
	// Entry 2: DR Operating Expenses 300 / CR Cash 300 → Jan expenses = 300
	// Entry 3 (Feb): DR Cash 500 / CR Sales Revenue 500 → NOT in Jan
	proposals := []core.Proposal{
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-01-10", DocumentDate: "2026-01-10",
			Summary: "Jan sale", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "1000.00"},
				{AccountCode: "4000", IsDebit: false, Amount: "1000.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-01-20", DocumentDate: "2026-01-20",
			Summary: "Jan expense", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "5100", IsDebit: true, Amount: "300.00"},
				{AccountCode: "1000", IsDebit: false, Amount: "300.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-02-05", DocumentDate: "2026-02-05",
			Summary: "Feb sale", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "500.00"},
				{AccountCode: "4000", IsDebit: false, Amount: "500.00"},
			},
		},
	}
	for _, p := range proposals {
		if err := ledger.Commit(ctx, p); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}
	}

	t.Run("Jan P&L revenue and expenses", func(t *testing.T) {
		report, err := reporting.GetProfitAndLoss(ctx, "1000", 2026, 1)
		if err != nil {
			t.Fatalf("GetProfitAndLoss failed: %v", err)
		}
		if report.Year != 2026 || report.Month != 1 {
			t.Errorf("Period mismatch: got %d/%d", report.Year, report.Month)
		}

		// Revenue: credit 1000 - debit 0 = 1000
		var revTotal decimal.Decimal
		for _, r := range report.Revenue {
			revTotal = revTotal.Add(r.Balance)
		}
		if !revTotal.Equal(decimal.NewFromInt(1000)) {
			t.Errorf("Jan revenue: want 1000, got %s", revTotal)
		}

		// Expenses: debit 300 - credit 0 = 300
		var expTotal decimal.Decimal
		for _, e := range report.Expenses {
			expTotal = expTotal.Add(e.Balance)
		}
		if !expTotal.Equal(decimal.NewFromInt(300)) {
			t.Errorf("Jan expenses: want 300, got %s", expTotal)
		}

		// Net income = 700
		if !report.NetIncome.Equal(decimal.NewFromInt(700)) {
			t.Errorf("Jan net income: want 700, got %s", report.NetIncome)
		}
	})

	t.Run("Feb P&L excludes Jan entries", func(t *testing.T) {
		report, err := reporting.GetProfitAndLoss(ctx, "1000", 2026, 2)
		if err != nil {
			t.Fatalf("GetProfitAndLoss (Feb) failed: %v", err)
		}
		// Feb revenue = 500, expenses = 0, net income = 500
		var revTotal decimal.Decimal
		for _, r := range report.Revenue {
			revTotal = revTotal.Add(r.Balance)
		}
		if !revTotal.Equal(decimal.NewFromInt(500)) {
			t.Errorf("Feb revenue: want 500, got %s", revTotal)
		}
		if !report.NetIncome.Equal(decimal.NewFromInt(500)) {
			t.Errorf("Feb net income: want 500, got %s", report.NetIncome)
		}
	})
}

func TestReporting_GetBalanceSheet(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	reporting := core.NewReportingService(pool)
	ctx := context.Background()

	// Post entries only to balance-sheet accounts so IsBalanced holds.
	// Entry 1 (Jan 5): DR Cash 10000 / CR Share Capital 10000
	//   → Assets=10000, Equity=10000, Liabilities=0 → balanced
	// Entry 2 (Jan 15): DR Cash 5000 / CR Accounts Payable 5000
	//   → Assets=15000, Equity=10000, Liabilities=5000 → balanced
	// Entry 3 (Feb 1): DR AR 2000 / CR Cash 2000
	//   → internal asset swap; still balanced
	proposals := []core.Proposal{
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-01-05", DocumentDate: "2026-01-05",
			Summary: "Share capital injection", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "10000.00"},
				{AccountCode: "3000", IsDebit: false, Amount: "10000.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-01-15", DocumentDate: "2026-01-15",
			Summary: "Loan received", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "5000.00"},
				{AccountCode: "2000", IsDebit: false, Amount: "5000.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-02-01", DocumentDate: "2026-02-01",
			Summary: "Cash to AR", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1200", IsDebit: true, Amount: "2000.00"},
				{AccountCode: "1000", IsDebit: false, Amount: "2000.00"},
			},
		},
	}
	for _, p := range proposals {
		if err := ledger.Commit(ctx, p); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}
	}

	t.Run("balance sheet as of Jan 31 excludes Feb entry", func(t *testing.T) {
		report, err := reporting.GetBalanceSheet(ctx, "1000", "2026-01-31")
		if err != nil {
			t.Fatalf("GetBalanceSheet failed: %v", err)
		}

		// Cash (1000): DR 15000 (10000+5000), net = 15000 asset
		var cashBalance decimal.Decimal
		for _, a := range report.Assets {
			if a.Code == "1000" {
				cashBalance = a.Balance
			}
		}
		if !cashBalance.Equal(decimal.NewFromInt(15000)) {
			t.Errorf("Cash balance as of Jan 31: want 15000, got %s", cashBalance)
		}

		// AR (1200): 0 (Feb entry excluded)
		var arBalance decimal.Decimal
		for _, a := range report.Assets {
			if a.Code == "1200" {
				arBalance = a.Balance
			}
		}
		if !arBalance.IsZero() {
			t.Errorf("AR balance as of Jan 31: want 0, got %s", arBalance)
		}

		// Total assets = 15000, total liabilities = 5000, total equity = 10000
		if !report.TotalAssets.Equal(decimal.NewFromInt(15000)) {
			t.Errorf("TotalAssets: want 15000, got %s", report.TotalAssets)
		}
		if !report.TotalLiabilities.Equal(decimal.NewFromInt(5000)) {
			t.Errorf("TotalLiabilities: want 5000, got %s", report.TotalLiabilities)
		}
		if !report.TotalEquity.Equal(decimal.NewFromInt(10000)) {
			t.Errorf("TotalEquity: want 10000, got %s", report.TotalEquity)
		}
		if !report.IsBalanced {
			t.Error("Expected IsBalanced=true")
		}
	})

	t.Run("balance sheet includes all entries with empty date", func(t *testing.T) {
		report, err := reporting.GetBalanceSheet(ctx, "1000", "2026-12-31")
		if err != nil {
			t.Fatalf("GetBalanceSheet (all) failed: %v", err)
		}
		// Cash = 15000 - 2000 = 13000; AR = 2000; total assets still 15000
		if !report.TotalAssets.Equal(decimal.NewFromInt(15000)) {
			t.Errorf("TotalAssets (all): want 15000, got %s", report.TotalAssets)
		}
		if !report.IsBalanced {
			t.Error("Expected IsBalanced=true for all entries")
		}
	})
}

func TestReporting_GetAccountStatement(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	reporting := core.NewReportingService(pool)
	ctx := context.Background()

	// Post 3 journal entries touching account 1000 (Test Asset):
	//   Entry 1 (2026-01-01): DR 1000 100, CR 4000 100  → running +100
	//   Entry 2 (2026-01-15): DR 1000 200, CR 4000 200  → running +300
	//   Entry 3 (2026-02-01): DR 4000  50, CR 1000  50  → running +250
	proposals := []core.Proposal{
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-01-01", DocumentDate: "2026-01-01",
			Summary: "Entry 1", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "100.00"},
				{AccountCode: "4000", IsDebit: false, Amount: "100.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-01-15", DocumentDate: "2026-01-15",
			Summary: "Entry 2", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "200.00"},
				{AccountCode: "4000", IsDebit: false, Amount: "200.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-02-01", DocumentDate: "2026-02-01",
			Summary: "Entry 3", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "4000", IsDebit: true, Amount: "50.00"},
				{AccountCode: "1000", IsDebit: false, Amount: "50.00"},
			},
		},
	}
	for _, p := range proposals {
		if err := ledger.Commit(ctx, p); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}
	}

	t.Run("full statement no date filter", func(t *testing.T) {
		lines, err := reporting.GetAccountStatement(ctx, "1000", "1000", "", "")
		if err != nil {
			t.Fatalf("GetAccountStatement failed: %v", err)
		}
		if len(lines) != 3 {
			t.Fatalf("Expected 3 lines, got %d", len(lines))
		}

		// Line 1: debit 100, running 100
		if !lines[0].Debit.Equal(decimal.NewFromInt(100)) {
			t.Errorf("Line 1 debit: want 100, got %s", lines[0].Debit)
		}
		if !lines[0].Credit.IsZero() {
			t.Errorf("Line 1 credit: want 0, got %s", lines[0].Credit)
		}
		if !lines[0].RunningBalance.Equal(decimal.NewFromInt(100)) {
			t.Errorf("Line 1 running: want 100, got %s", lines[0].RunningBalance)
		}

		// Line 2: debit 200, running 300
		if !lines[1].Debit.Equal(decimal.NewFromInt(200)) {
			t.Errorf("Line 2 debit: want 200, got %s", lines[1].Debit)
		}
		if !lines[1].RunningBalance.Equal(decimal.NewFromInt(300)) {
			t.Errorf("Line 2 running: want 300, got %s", lines[1].RunningBalance)
		}

		// Line 3: credit 50, running 250
		if !lines[2].Debit.IsZero() {
			t.Errorf("Line 3 debit: want 0, got %s", lines[2].Debit)
		}
		if !lines[2].Credit.Equal(decimal.NewFromInt(50)) {
			t.Errorf("Line 3 credit: want 50, got %s", lines[2].Credit)
		}
		if !lines[2].RunningBalance.Equal(decimal.NewFromInt(250)) {
			t.Errorf("Line 3 running: want 250, got %s", lines[2].RunningBalance)
		}
	})

	t.Run("date range Jan only", func(t *testing.T) {
		lines, err := reporting.GetAccountStatement(ctx, "1000", "1000", "2026-01-01", "2026-01-31")
		if err != nil {
			t.Fatalf("GetAccountStatement (Jan) failed: %v", err)
		}
		if len(lines) != 2 {
			t.Fatalf("Expected 2 Jan lines, got %d", len(lines))
		}
		// Running balance after two Jan debits = 300
		if !lines[1].RunningBalance.Equal(decimal.NewFromInt(300)) {
			t.Errorf("Jan closing running: want 300, got %s", lines[1].RunningBalance)
		}
	})

	t.Run("empty result for unknown account", func(t *testing.T) {
		lines, err := reporting.GetAccountStatement(ctx, "1000", "9999", "", "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(lines) != 0 {
			t.Errorf("Expected 0 lines for non-existent account, got %d", len(lines))
		}
	})
}

func TestReporting_GetDocumentTypeGovernance(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	reporting := core.NewReportingService(pool)
	ctx := context.Background()

	proposals := []core.Proposal{
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: "manual-adjustment-1", TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-03-01", DocumentDate: "2026-03-01",
			Summary: "Manual adjustment", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "100.00"},
				{AccountCode: "4000", IsDebit: false, Amount: "100.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: "payment-order-42", TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-03-02", DocumentDate: "2026-03-02",
			Summary: "Operational-like JE", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "50.00"},
				{AccountCode: "1200", IsDebit: false, Amount: "50.00"},
			},
		},
		{
			DocumentTypeCode: "RC", CompanyCode: "1000",
			IdempotencyKey: "payment-order-43", TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-03-03", DocumentDate: "2026-03-03",
			Summary: "Customer receipt", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "200.00"},
				{AccountCode: "1200", IsDebit: false, Amount: "200.00"},
			},
		},
		{
			DocumentTypeCode: "PV", CompanyCode: "1000",
			IdempotencyKey: "pay-vendor-po-9", TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-03-04", DocumentDate: "2026-03-04",
			Summary: "Vendor payment", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "2000", IsDebit: true, Amount: "75.00"},
				{AccountCode: "1000", IsDebit: false, Amount: "75.00"},
			},
		},
	}
	for _, p := range proposals {
		if err := ledger.Commit(ctx, p); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}
	}

	report, err := reporting.GetDocumentTypeGovernance(ctx, "1000", "2026-03-01", "2026-03-31")
	if err != nil {
		t.Fatalf("GetDocumentTypeGovernance failed: %v", err)
	}

	if report.TotalPostings != 4 {
		t.Fatalf("expected total postings 4, got %d", report.TotalPostings)
	}
	if report.JEPostings != 2 {
		t.Fatalf("expected JE postings 2, got %d", report.JEPostings)
	}
	if report.OperationalLikeJEPostings != 1 {
		t.Fatalf("expected operational-like JE postings 1, got %d", report.OperationalLikeJEPostings)
	}
	if report.ManualJEPostings != 1 {
		t.Fatalf("expected manual JE postings 1, got %d", report.ManualJEPostings)
	}
	if report.JESharePct.StringFixed(2) != "50.00" {
		t.Fatalf("expected JE share 50.00, got %s", report.JESharePct.StringFixed(2))
	}

	countMap := make(map[string]int64, len(report.Counts))
	for _, c := range report.Counts {
		countMap[c.DocumentTypeCode] = c.PostingCount
	}
	if countMap["JE"] != 2 || countMap["RC"] != 1 || countMap["PV"] != 1 {
		t.Fatalf("unexpected document type counts: %+v", countMap)
	}
}

func TestReporting_GetManualJEControlAccountHits(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	reporting := core.NewReportingService(pool)
	ctx := context.Background()

	// Mark AR/AP as control accounts for the reporting query.
	if _, err := pool.Exec(ctx, `
		UPDATE accounts
		SET is_control_account = true,
		    control_type = CASE code
		        WHEN '1200' THEN 'AR'
		        WHEN '2000' THEN 'AP'
		        ELSE control_type
		    END
		WHERE company_id = 1
		  AND code IN ('1200', '2000')
	`); err != nil {
		t.Fatalf("failed to flag control accounts: %v", err)
	}

	// JE #1 hits AR (control); JE #2 only non-control accounts.
	proposals := []core.Proposal{
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-03-01", DocumentDate: "2026-03-01",
			Summary: "Manual AR adjustment", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1200", IsDebit: true, Amount: "250.00"},
				{AccountCode: "4000", IsDebit: false, Amount: "250.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-03-02", DocumentDate: "2026-03-02",
			Summary: "Cash vs revenue", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "100.00"},
				{AccountCode: "4000", IsDebit: false, Amount: "100.00"},
			},
		},
	}
	for _, p := range proposals {
		if err := ledger.Commit(ctx, p); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}
	}

	hits, err := reporting.GetManualJEControlAccountHits(ctx, "1000", "", "")
	if err != nil {
		t.Fatalf("GetManualJEControlAccountHits failed: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 control-account hit, got %d", len(hits))
	}
	if hits[0].AccountCode != "1200" {
		t.Errorf("expected account 1200, got %s", hits[0].AccountCode)
	}
	if hits[0].ControlType != "AR" {
		t.Errorf("expected control type AR, got %s", hits[0].ControlType)
	}
	if hits[0].Direction != "DEBIT" {
		t.Errorf("expected direction DEBIT, got %s", hits[0].Direction)
	}
	if !hits[0].Amount.Equal(decimal.NewFromInt(250)) {
		t.Errorf("expected amount 250, got %s", hits[0].Amount)
	}
}

func TestReporting_GetControlAccountReconciliation(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ledger := core.NewLedger(pool, docService)
	reporting := core.NewReportingService(pool)
	ctx := context.Background()

	// Flag control accounts for AR/AP/INVENTORY.
	if _, err := pool.Exec(ctx, `
		UPDATE accounts
		SET is_control_account = true,
		    control_type = CASE code
		        WHEN '1200' THEN 'AR'
		        WHEN '2000' THEN 'AP'
		        WHEN '1000' THEN 'INVENTORY'
		        ELSE control_type
		    END
		WHERE company_id = 1
		  AND code IN ('1000', '1200', '2000')
	`); err != nil {
		t.Fatalf("failed to flag control accounts: %v", err)
	}

	// Seed operational rows for AR/AP/Inventory.
	if _, err := pool.Exec(ctx, `
		INSERT INTO customers (company_id, code, name, email, phone, address, credit_limit, payment_terms_days)
		VALUES (1, 'C001', 'Customer One', '', '', '', 0, 30);

		INSERT INTO vendors (company_id, code, name)
		VALUES (1, 'V001', 'Vendor One');

		INSERT INTO products (company_id, code, name, description, unit_price, unit, revenue_account_code, is_active)
		VALUES (1, 'P001', 'Widget', '', 10.00, 'unit', '4000', true);

		INSERT INTO warehouses (company_id, code, name, is_active)
		VALUES (1, 'WH1', 'Main Warehouse', true);
	`); err != nil {
		t.Fatalf("seed reconciliation dimensions: %v", err)
	}

	var customerID, vendorID, productID, warehouseID int
	if err := pool.QueryRow(ctx, `SELECT id FROM customers WHERE company_id = 1 AND code = 'C001'`).Scan(&customerID); err != nil {
		t.Fatalf("load customer id: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM vendors WHERE company_id = 1 AND code = 'V001'`).Scan(&vendorID); err != nil {
		t.Fatalf("load vendor id: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM products WHERE company_id = 1 AND code = 'P001'`).Scan(&productID); err != nil {
		t.Fatalf("load product id: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM warehouses WHERE company_id = 1 AND code = 'WH1'`).Scan(&warehouseID); err != nil {
		t.Fatalf("load warehouse id: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO sales_orders (company_id, customer_id, status, order_date, currency, exchange_rate, total_transaction, total_base, notes)
		VALUES (1, $1, 'INVOICED', '2026-03-10', 'INR', 1.0, 900.00, 900.00, '')
	`, customerID); err != nil {
		t.Fatalf("seed AR operational balance: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO purchase_orders (company_id, vendor_id, status, po_date, currency, exchange_rate, total_transaction, total_base, invoice_amount, notes)
		VALUES (1, $1, 'INVOICED', '2026-03-10', 'INR', 1.0, 500.00, 500.00, 500.00, '')
	`, vendorID); err != nil {
		t.Fatalf("seed AP operational balance: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO inventory_items (company_id, product_id, warehouse_id, qty_on_hand, qty_reserved, unit_cost)
		VALUES (1, $1, $2, 10.0000, 0.0000, 70.000000)
	`, productID, warehouseID); err != nil {
		t.Fatalf("seed operational balances: %v", err)
	}

	// GL balances: AR=1000, AP=600 (credit), INVENTORY=800.
	proposals := []core.Proposal{
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-03-10", DocumentDate: "2026-03-10",
			Summary: "AR posting", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1200", IsDebit: true, Amount: "1000.00"},
				{AccountCode: "4000", IsDebit: false, Amount: "1000.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-03-10", DocumentDate: "2026-03-10",
			Summary: "AP posting", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "5100", IsDebit: true, Amount: "600.00"},
				{AccountCode: "2000", IsDebit: false, Amount: "600.00"},
			},
		},
		{
			DocumentTypeCode: "JE", CompanyCode: "1000",
			IdempotencyKey: uuid.NewString(), TransactionCurrency: "INR", ExchangeRate: "1.0",
			PostingDate: "2026-03-10", DocumentDate: "2026-03-10",
			Summary: "Inventory posting", Reasoning: "test",
			Lines: []core.ProposalLine{
				{AccountCode: "1000", IsDebit: true, Amount: "800.00"},
				{AccountCode: "3000", IsDebit: false, Amount: "800.00"},
			},
		},
	}
	for _, p := range proposals {
		if err := ledger.Commit(ctx, p); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}
	}

	report, err := reporting.GetControlAccountReconciliation(ctx, "1000", "2026-03-31")
	if err != nil {
		t.Fatalf("GetControlAccountReconciliation failed: %v", err)
	}
	if len(report.Lines) != 3 {
		t.Fatalf("expected 3 reconciliation lines, got %d", len(report.Lines))
	}

	byType := make(map[string]core.ControlAccountReconciliationLine, 3)
	for _, l := range report.Lines {
		byType[l.ControlType] = l
	}

	if !byType["AR"].Variance.Equal(decimal.NewFromInt(100)) {
		t.Errorf("AR variance: want 100, got %s", byType["AR"].Variance)
	}
	if !byType["AP"].Variance.Equal(decimal.NewFromInt(100)) {
		t.Errorf("AP variance: want 100, got %s", byType["AP"].Variance)
	}
	if !byType["INVENTORY"].Variance.Equal(decimal.NewFromInt(100)) {
		t.Errorf("INVENTORY variance: want 100, got %s", byType["INVENTORY"].Variance)
	}
}
