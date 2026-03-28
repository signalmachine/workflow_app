package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// ── Report types ──────────────────────────────────────────────────────────────

// StatementLine represents a single journal line in an account statement.
// RunningBalance is the cumulative net-debit position after this line
// (positive = net debit, negative = net credit).
type StatementLine struct {
	PostingDate    string
	DocumentDate   string
	Narration      string
	Reference      string
	Debit          decimal.Decimal
	Credit         decimal.Decimal
	RunningBalance decimal.Decimal
}

// AccountLine is a single account entry in a P&L or Balance Sheet report.
// Balance is expressed in the sign convention for that section:
//   - P&L Revenue:  positive = income received
//   - P&L Expenses: positive = cost incurred
//   - BS Assets:    positive = net debit (normal asset balance)
//   - BS Liabilities/Equity: positive = net credit (normal balance)
type AccountLine struct {
	Code    string
	Name    string
	Balance decimal.Decimal
}

// ManualControlAccountHit is one manual JE line posted to a control account.
type ManualControlAccountHit struct {
	EntryID     int
	PostingDate string
	Username    string
	AccountCode string
	ControlType string
	Direction   string
	Amount      decimal.Decimal
}

// ControlAccountReconciliationLine compares GL and operational balances for one control type.
type ControlAccountReconciliationLine struct {
	ControlType        string
	GLBalance          decimal.Decimal
	OperationalBalance decimal.Decimal
	Variance           decimal.Decimal
	HasVariance        bool
	DetailLinks        []string
}

// ControlAccountReconciliationReport is a lightweight drift diagnostics report.
type ControlAccountReconciliationReport struct {
	CompanyCode string
	AsOfDate    string
	Lines       []ControlAccountReconciliationLine
}

// DocumentTypeCount is one aggregate count bucket by document type.
type DocumentTypeCount struct {
	DocumentTypeCode string
	PostingCount     int64
}

// DocumentTypeGovernanceReport summarizes posting mix and operational-like JE usage.
type DocumentTypeGovernanceReport struct {
	CompanyCode               string
	FromDate                  string
	ToDate                    string
	TotalPostings             int64
	JEPostings                int64
	OperationalLikeJEPostings int64
	ManualJEPostings          int64
	JESharePct                decimal.Decimal
	Counts                    []DocumentTypeCount
}

// PLReport is the Profit & Loss report for one calendar period.
type PLReport struct {
	CompanyCode string
	Year        int
	Month       int
	Revenue     []AccountLine   // credit-dominant accounts (type = 'revenue')
	Expenses    []AccountLine   // debit-dominant accounts  (type = 'expense')
	NetIncome   decimal.Decimal // Revenue total - Expenses total
}

// BSReport is the Balance Sheet report as of a given date.
// IsBalanced is true when TotalAssets == TotalLiabilities + TotalEquity,
// which holds for any correctly posted double-entry ledger when all
// income/expense has been closed to retained earnings.
type BSReport struct {
	CompanyCode      string
	AsOfDate         string
	Assets           []AccountLine
	Liabilities      []AccountLine
	Equity           []AccountLine
	TotalAssets      decimal.Decimal
	TotalLiabilities decimal.Decimal
	TotalEquity      decimal.Decimal
	IsBalanced       bool
}

// ── Interface ─────────────────────────────────────────────────────────────────

// ReportingService provides read-only reporting queries over the ledger.
type ReportingService interface {
	// GetAccountStatement returns all journal lines for an account within the
	// given date range, ordered by posting_date ASC then entry id ASC.
	// fromDate and toDate are optional — pass empty string for no bound.
	// RunningBalance on each line is the cumulative (debit_base − credit_base).
	GetAccountStatement(ctx context.Context, companyCode, accountCode, fromDate, toDate string) ([]StatementLine, error)

	// GetProfitAndLoss returns the P&L report for the given year and month.
	// Revenue balances are expressed as positive credit-minus-debit amounts.
	// Expense balances are expressed as positive debit-minus-credit amounts.
	GetProfitAndLoss(ctx context.Context, companyCode string, year, month int) (*PLReport, error)

	// GetBalanceSheet returns the Balance Sheet as of the given date.
	// If asOfDate is empty, today's date is used.
	GetBalanceSheet(ctx context.Context, companyCode, asOfDate string) (*BSReport, error)

	// GetManualJEControlAccountHits returns manual JE lines posted to control accounts.
	// fromDate and toDate are optional — pass empty string for no bound.
	GetManualJEControlAccountHits(ctx context.Context, companyCode, fromDate, toDate string) ([]ManualControlAccountHit, error)

	// GetControlAccountReconciliation returns AR/AP/INVENTORY GL-vs-operational variance diagnostics.
	GetControlAccountReconciliation(ctx context.Context, companyCode, asOfDate string) (*ControlAccountReconciliationReport, error)

	// GetDocumentTypeGovernance returns posting-mix diagnostics including JE share
	// and likely operational-like JE counts derived from idempotency key patterns.
	GetDocumentTypeGovernance(ctx context.Context, companyCode, fromDate, toDate string) (*DocumentTypeGovernanceReport, error)
}

// ── Implementation ────────────────────────────────────────────────────────────

type reportingService struct {
	pool *pgxpool.Pool
}

// NewReportingService constructs a ReportingService backed by the given pool.
func NewReportingService(pool *pgxpool.Pool) ReportingService {
	return &reportingService{pool: pool}
}

// resolveCompanyID looks up the integer primary key for a company code.
func (s *reportingService) resolveCompanyID(ctx context.Context, companyCode string) (int, error) {
	var id int
	if err := s.pool.QueryRow(ctx,
		"SELECT id FROM companies WHERE company_code = $1", companyCode,
	).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("company %s not found", companyCode)
		}
		return 0, fmt.Errorf("failed to resolve company: %w", err)
	}
	return id, nil
}

// ── GetAccountStatement ───────────────────────────────────────────────────────

func (s *reportingService) GetAccountStatement(ctx context.Context, companyCode, accountCode, fromDate, toDate string) ([]StatementLine, error) {
	companyID, err := s.resolveCompanyID(ctx, companyCode)
	if err != nil {
		return nil, err
	}

	q := `
		SELECT je.posting_date::text,
		       je.document_date::text,
		       je.narration,
		       COALESCE(je.reference_id, ''),
		       jl.debit_base,
		       jl.credit_base
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.entry_id
		JOIN accounts a         ON a.id  = jl.account_id
		WHERE je.company_id = $1
		  AND a.code = $2`

	args := []any{companyID, accountCode}
	if fromDate != "" {
		args = append(args, fromDate)
		q += fmt.Sprintf(" AND je.posting_date >= $%d::date", len(args))
	}
	if toDate != "" {
		args = append(args, toDate)
		q += fmt.Sprintf(" AND je.posting_date <= $%d::date", len(args))
	}
	q += " ORDER BY je.posting_date ASC, je.id ASC"

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query account statement: %w", err)
	}
	defer rows.Close()

	var lines []StatementLine
	running := decimal.Zero
	for rows.Next() {
		var sl StatementLine
		if err := rows.Scan(
			&sl.PostingDate, &sl.DocumentDate, &sl.Narration, &sl.Reference,
			&sl.Debit, &sl.Credit,
		); err != nil {
			return nil, fmt.Errorf("failed to scan statement line: %w", err)
		}
		running = running.Add(sl.Debit).Sub(sl.Credit)
		sl.RunningBalance = running
		lines = append(lines, sl)
	}
	return lines, nil
}

// ── GetProfitAndLoss ──────────────────────────────────────────────────────────

// GetProfitAndLoss returns the P&L for the given year/month by querying
// journal_lines directly so the result is always current (not dependent on
// a materialized view refresh cycle).
func (s *reportingService) GetProfitAndLoss(ctx context.Context, companyCode string, year, month int) (*PLReport, error) {
	companyID, err := s.resolveCompanyID(ctx, companyCode)
	if err != nil {
		return nil, err
	}

	// Subquery aggregates only lines whose entry falls in the target period.
	const q = `
		SELECT a.code, a.name, a.type,
		       COALESCE(s.debit_total,  0) AS debit_total,
		       COALESCE(s.credit_total, 0) AS credit_total
		FROM accounts a
		JOIN companies c ON c.id = a.company_id
		LEFT JOIN (
		    SELECT jl.account_id,
		           SUM(jl.debit_base)  AS debit_total,
		           SUM(jl.credit_base) AS credit_total
		    FROM journal_lines jl
		    JOIN journal_entries je ON je.id = jl.entry_id
		    WHERE je.company_id = $1
		      AND EXTRACT(YEAR  FROM je.posting_date)::int = $2
		      AND EXTRACT(MONTH FROM je.posting_date)::int = $3
		    GROUP BY jl.account_id
		) s ON s.account_id = a.id
		WHERE c.id = $1
		  AND a.type IN ('revenue', 'expense')
		ORDER BY a.type, a.code`

	rows, err := s.pool.Query(ctx, q, companyID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to query P&L: %w", err)
	}
	defer rows.Close()

	report := &PLReport{CompanyCode: companyCode, Year: year, Month: month}
	var totalRevenue, totalExpenses decimal.Decimal

	for rows.Next() {
		var code, name, accType string
		var debit, credit decimal.Decimal
		if err := rows.Scan(&code, &name, &accType, &debit, &credit); err != nil {
			return nil, fmt.Errorf("failed to scan P&L row: %w", err)
		}

		switch accType {
		case "revenue":
			// Positive balance = credit > debit (income received).
			bal := credit.Sub(debit)
			report.Revenue = append(report.Revenue, AccountLine{Code: code, Name: name, Balance: bal})
			totalRevenue = totalRevenue.Add(bal)
		case "expense":
			// Positive balance = debit > credit (cost incurred).
			bal := debit.Sub(credit)
			report.Expenses = append(report.Expenses, AccountLine{Code: code, Name: name, Balance: bal})
			totalExpenses = totalExpenses.Add(bal)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("P&L row iteration error: %w", err)
	}

	report.NetIncome = totalRevenue.Sub(totalExpenses)
	return report, nil
}

// ── GetBalanceSheet ───────────────────────────────────────────────────────────

// GetBalanceSheet returns the Balance Sheet as of the given date by querying
// journal_lines directly with a date ceiling filter.
func (s *reportingService) GetBalanceSheet(ctx context.Context, companyCode, asOfDate string) (*BSReport, error) {
	companyID, err := s.resolveCompanyID(ctx, companyCode)
	if err != nil {
		return nil, err
	}

	if asOfDate == "" {
		asOfDate = time.Now().Format("2006-01-02")
	}

	// Subquery aggregates lines whose entry was posted on or before asOfDate.
	const q = `
		SELECT a.code, a.name, a.type,
		       COALESCE(s.total_debit,  0) - COALESCE(s.total_credit, 0) AS net_balance
		FROM accounts a
		JOIN companies c ON c.id = a.company_id
		LEFT JOIN (
		    SELECT jl.account_id,
		           SUM(jl.debit_base)  AS total_debit,
		           SUM(jl.credit_base) AS total_credit
		    FROM journal_lines jl
		    JOIN journal_entries je ON je.id = jl.entry_id
		    WHERE je.company_id = $1
		      AND je.posting_date <= $2::date
		    GROUP BY jl.account_id
		) s ON s.account_id = a.id
		WHERE c.id = $1
		  AND a.type IN ('asset', 'liability', 'equity')
		ORDER BY a.type, a.code`

	rows, err := s.pool.Query(ctx, q, companyID, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query balance sheet: %w", err)
	}
	defer rows.Close()

	report := &BSReport{CompanyCode: companyCode, AsOfDate: asOfDate}

	for rows.Next() {
		var code, name, accType string
		var netBalance decimal.Decimal // debit - credit
		if err := rows.Scan(&code, &name, &accType, &netBalance); err != nil {
			return nil, fmt.Errorf("failed to scan balance sheet row: %w", err)
		}

		switch accType {
		case "asset":
			// Assets carry debit balances: positive net = normal.
			report.Assets = append(report.Assets, AccountLine{Code: code, Name: name, Balance: netBalance})
			report.TotalAssets = report.TotalAssets.Add(netBalance)
		case "liability":
			// Liabilities carry credit balances: negate so positive = normal liability.
			bal := netBalance.Neg()
			report.Liabilities = append(report.Liabilities, AccountLine{Code: code, Name: name, Balance: bal})
			report.TotalLiabilities = report.TotalLiabilities.Add(bal)
		case "equity":
			// Equity carries credit balances: same convention as liabilities.
			bal := netBalance.Neg()
			report.Equity = append(report.Equity, AccountLine{Code: code, Name: name, Balance: bal})
			report.TotalEquity = report.TotalEquity.Add(bal)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("balance sheet row iteration error: %w", err)
	}

	report.IsBalanced = report.TotalAssets.Equal(report.TotalLiabilities.Add(report.TotalEquity))
	return report, nil
}

// GetManualJEControlAccountHits lists control-account hits on manual JE postings.
func (s *reportingService) GetManualJEControlAccountHits(ctx context.Context, companyCode, fromDate, toDate string) ([]ManualControlAccountHit, error) {
	companyID, err := s.resolveCompanyID(ctx, companyCode)
	if err != nil {
		return nil, err
	}

	q := `
		SELECT je.id,
		       je.posting_date::text,
		       COALESCE(u.username, ''),
		       a.code,
		       COALESCE(a.control_type, ''),
		       CASE WHEN jl.debit_base > 0 THEN 'DEBIT' ELSE 'CREDIT' END AS direction,
		       CASE WHEN jl.debit_base > 0 THEN jl.debit_base ELSE jl.credit_base END AS amount
		FROM journal_entries je
		JOIN journal_lines jl ON jl.entry_id = je.id
		JOIN accounts a ON a.id = jl.account_id
		LEFT JOIN users u ON u.id = je.created_by_user_id
		WHERE je.company_id = $1
		  AND a.is_control_account = true
		  AND je.reference_type = 'DOCUMENT'
		  AND EXISTS (
		      SELECT 1
		      FROM documents d
		      WHERE d.company_id = je.company_id
		        AND d.type_code = 'JE'
		        AND d.document_number = je.reference_id
		  )`

	args := []any{companyID}
	if fromDate != "" {
		args = append(args, fromDate)
		q += fmt.Sprintf(" AND je.posting_date >= $%d::date", len(args))
	}
	if toDate != "" {
		args = append(args, toDate)
		q += fmt.Sprintf(" AND je.posting_date <= $%d::date", len(args))
	}
	q += " ORDER BY je.posting_date DESC, je.id DESC, a.code ASC"

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query manual JE control-account hits: %w", err)
	}
	defer rows.Close()

	var hits []ManualControlAccountHit
	for rows.Next() {
		var h ManualControlAccountHit
		if err := rows.Scan(&h.EntryID, &h.PostingDate, &h.Username, &h.AccountCode, &h.ControlType, &h.Direction, &h.Amount); err != nil {
			return nil, fmt.Errorf("failed to scan manual JE control-account hit: %w", err)
		}
		hits = append(hits, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("manual JE control-account rows iteration error: %w", err)
	}

	return hits, nil
}

// GetControlAccountReconciliation returns lightweight drift diagnostics for control accounts.
func (s *reportingService) GetControlAccountReconciliation(ctx context.Context, companyCode, asOfDate string) (*ControlAccountReconciliationReport, error) {
	companyID, err := s.resolveCompanyID(ctx, companyCode)
	if err != nil {
		return nil, err
	}

	if asOfDate == "" {
		asOfDate = time.Now().Format("2006-01-02")
	}

	glByType := map[string]decimal.Decimal{
		"AR":        decimal.Zero,
		"AP":        decimal.Zero,
		"INVENTORY": decimal.Zero,
	}
	rows, err := s.pool.Query(ctx, `
		SELECT a.control_type,
		       COALESCE(SUM(jl.debit_base), 0) - COALESCE(SUM(jl.credit_base), 0) AS net_debit
		FROM accounts a
		LEFT JOIN journal_lines jl ON jl.account_id = a.id
		LEFT JOIN journal_entries je
		       ON je.id = jl.entry_id
		      AND je.company_id = a.company_id
		      AND je.posting_date <= $2::date
		WHERE a.company_id = $1
		  AND a.is_control_account = true
		  AND a.control_type IN ('AR', 'AP', 'INVENTORY')
		GROUP BY a.control_type
	`, companyID, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("query GL control balances: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var controlType string
		var netDebit decimal.Decimal
		if err := rows.Scan(&controlType, &netDebit); err != nil {
			return nil, fmt.Errorf("scan GL control balances: %w", err)
		}
		if controlType == "AP" {
			glByType[controlType] = netDebit.Neg()
		} else {
			glByType[controlType] = netDebit
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate GL control balances: %w", err)
	}

	var arOperational decimal.Decimal
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_base), 0)
		FROM sales_orders
		WHERE company_id = $1
		  AND status = 'INVOICED'
	`, companyID).Scan(&arOperational); err != nil {
		return nil, fmt.Errorf("query AR operational balance: %w", err)
	}

	var apOperational decimal.Decimal
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(COALESCE(invoice_amount, total_base)), 0)
		FROM purchase_orders
		WHERE company_id = $1
		  AND status = 'INVOICED'
	`, companyID).Scan(&apOperational); err != nil {
		return nil, fmt.Errorf("query AP operational balance: %w", err)
	}

	var inventoryOperational decimal.Decimal
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(ii.qty_on_hand * ii.unit_cost), 0)
		FROM inventory_items ii
		WHERE ii.company_id = $1
	`, companyID).Scan(&inventoryOperational); err != nil {
		return nil, fmt.Errorf("query inventory operational valuation: %w", err)
	}

	operationalByType := map[string]decimal.Decimal{
		"AR":        arOperational,
		"AP":        apOperational,
		"INVENTORY": inventoryOperational,
	}
	linksByType := map[string][]string{
		"AR": {
			"/reports/statement?account=1200",
			"/sales/orders",
		},
		"AP": {
			"/reports/statement?account=2000",
			"/purchases/orders",
		},
		"INVENTORY": {
			"/reports/statement?account=1400",
			"/inventory/stock",
		},
	}

	lines := make([]ControlAccountReconciliationLine, 0, 3)
	for _, typ := range []string{"AR", "AP", "INVENTORY"} {
		gl := glByType[typ]
		op := operationalByType[typ]
		variance := gl.Sub(op)
		lines = append(lines, ControlAccountReconciliationLine{
			ControlType:        typ,
			GLBalance:          gl,
			OperationalBalance: op,
			Variance:           variance,
			HasVariance:        !variance.Equal(decimal.Zero),
			DetailLinks:        linksByType[typ],
		})
	}

	return &ControlAccountReconciliationReport{
		CompanyCode: companyCode,
		AsOfDate:    asOfDate,
		Lines:       lines,
	}, nil
}

// GetDocumentTypeGovernance returns posting mix and JE diagnostics for governance monitoring.
func (s *reportingService) GetDocumentTypeGovernance(ctx context.Context, companyCode, fromDate, toDate string) (*DocumentTypeGovernanceReport, error) {
	companyID, err := s.resolveCompanyID(ctx, companyCode)
	if err != nil {
		return nil, err
	}

	where := `
		WHERE je.company_id = $1
		  AND je.reference_type = 'DOCUMENT'
		  AND d.company_id = je.company_id
		  AND d.document_number = je.reference_id`
	args := []any{companyID}
	if fromDate != "" {
		args = append(args, fromDate)
		where += fmt.Sprintf(" AND je.posting_date >= $%d::date", len(args))
	}
	if toDate != "" {
		args = append(args, toDate)
		where += fmt.Sprintf(" AND je.posting_date <= $%d::date", len(args))
	}

	rows, err := s.pool.Query(ctx, `
		SELECT d.type_code, COUNT(*)::bigint AS posting_count
		FROM journal_entries je
		JOIN documents d
		  ON d.company_id = je.company_id
		 AND d.document_number = je.reference_id
	`+where+`
		GROUP BY d.type_code
		ORDER BY d.type_code
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("query document type counts: %w", err)
	}
	defer rows.Close()

	report := &DocumentTypeGovernanceReport{
		CompanyCode: companyCode,
		FromDate:    fromDate,
		ToDate:      toDate,
		Counts:      []DocumentTypeCount{},
	}

	for rows.Next() {
		var c DocumentTypeCount
		if err := rows.Scan(&c.DocumentTypeCode, &c.PostingCount); err != nil {
			return nil, fmt.Errorf("scan document type counts: %w", err)
		}
		report.Counts = append(report.Counts, c)
		report.TotalPostings += c.PostingCount
		if c.DocumentTypeCode == "JE" {
			report.JEPostings = c.PostingCount
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document type counts: %w", err)
	}

	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::bigint
		FROM journal_entries je
		JOIN documents d
		  ON d.company_id = je.company_id
		 AND d.document_number = je.reference_id
	`+where+`
		  AND d.type_code = 'JE'
		  AND (
		      je.idempotency_key LIKE 'payment-order-%'
		   OR je.idempotency_key LIKE 'pay-vendor-po-%'
		   OR je.idempotency_key LIKE 'invoice-order-%'
		   OR je.idempotency_key LIKE 'goods-receipt-%'
		   OR je.idempotency_key LIKE 'goods-issue-order-%'
		   OR je.idempotency_key LIKE '%-service-receipt'
		  )
	`, args...).Scan(&report.OperationalLikeJEPostings); err != nil {
		return nil, fmt.Errorf("query operational-like JE count: %w", err)
	}

	report.ManualJEPostings = report.JEPostings - report.OperationalLikeJEPostings
	if report.ManualJEPostings < 0 {
		report.ManualJEPostings = 0
	}

	if report.TotalPostings > 0 {
		report.JESharePct = decimal.NewFromInt(report.JEPostings).
			Mul(decimal.NewFromInt(100)).
			Div(decimal.NewFromInt(report.TotalPostings)).
			Round(2)
	}

	return report, nil
}
