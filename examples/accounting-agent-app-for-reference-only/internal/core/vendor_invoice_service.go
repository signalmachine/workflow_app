package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

// ClosePO closes an open purchase order when invoice handling is done outside the strict PO flow.
func (s *purchaseOrderService) ClosePO(ctx context.Context, companyID, poID int, reason string, closedByUserID *int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.closePOInTx(ctx, tx, companyID, poID, reason, closedByUserID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit close PO: %w", err)
	}
	return nil
}

func (s *purchaseOrderService) closePOInTx(ctx context.Context, tx pgx.Tx, companyID, poID int, reason string, closedByUserID *int) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("close_reason is required")
	}

	var poCompanyID int
	var status string
	if err := tx.QueryRow(ctx,
		"SELECT company_id, status FROM purchase_orders WHERE id = $1 FOR UPDATE",
		poID,
	).Scan(&poCompanyID, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("purchase order %d not found", poID)
		}
		return fmt.Errorf("fetch purchase order %d: %w", poID, err)
	}
	if poCompanyID != companyID {
		return fmt.Errorf("purchase order %d does not belong to company %d", poID, companyID)
	}

	switch status {
	case "CLOSED":
		return nil
	case "PAID":
		return fmt.Errorf("purchase order %d cannot be closed: status is PAID", poID)
	case "APPROVED", "RECEIVED", "INVOICED":
		// Allowed.
	default:
		return fmt.Errorf("purchase order %d cannot be closed: status is %s", poID, status)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE purchase_orders
		SET status = 'CLOSED',
		    closed_at = NOW(),
		    close_reason = $1,
		    closed_by_user_id = $2
		WHERE id = $3`,
		reason, closedByUserID, poID,
	); err != nil {
		return fmt.Errorf("close purchase order %d: %w", poID, err)
	}

	return nil
}

// RecordDirectVendorInvoice posts a PI and persists a vendor_invoices record with line allocations.
func (s *purchaseOrderService) RecordDirectVendorInvoice(ctx context.Context, req DirectVendorInvoiceInput, ledger *Ledger) (*VendorInvoice, error) {
	if req.CompanyID <= 0 {
		return nil, fmt.Errorf("company_id is required")
	}
	if strings.TrimSpace(req.CompanyCode) == "" {
		return nil, fmt.Errorf("company_code is required")
	}
	if req.VendorID <= 0 {
		return nil, fmt.Errorf("vendor_id is required")
	}
	if strings.TrimSpace(req.InvoiceNumber) == "" {
		return nil, fmt.Errorf("invoice_number is required")
	}
	if req.InvoiceDate.IsZero() {
		return nil, fmt.Errorf("invoice_date is required")
	}
	if req.InvoiceAmount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("invoice_amount must be > 0")
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		return nil, fmt.Errorf("idempotency_key is required")
	}
	if len(req.Lines) == 0 {
		return nil, fmt.Errorf("at least one invoice line is required")
	}
	if req.ClosePO && req.POID == nil {
		return nil, fmt.Errorf("po_id is required when close_po is true")
	}

	source := strings.ToLower(strings.TrimSpace(req.Source))
	if source == "" {
		if req.POID != nil {
			source = "po_bypass"
		} else {
			source = "direct"
		}
	}
	if source != "direct" && source != "po_strict" && source != "po_bypass" {
		return nil, fmt.Errorf("invalid source %q", req.Source)
	}

	postingDate := req.PostingDate
	if postingDate.IsZero() {
		postingDate = req.InvoiceDate
	}
	documentDate := req.DocumentDate
	if documentDate.IsZero() {
		documentDate = req.InvoiceDate
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		if err := tx.QueryRow(ctx, "SELECT base_currency FROM companies WHERE id = $1", req.CompanyID).Scan(&currency); err != nil {
			return nil, fmt.Errorf("resolve company base currency: %w", err)
		}
	}
	rate := req.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}

	// Resolve vendor AP account and verify ownership.
	var vendorCompanyID int
	var apAccountCode string
	if err := tx.QueryRow(ctx, `
		SELECT v.company_id, COALESCE(a.code, v.ap_account_code, '2000')
		FROM vendors v
		LEFT JOIN accounts a ON a.id = v.ap_account_id AND a.company_id = v.company_id
		WHERE v.id = $1
	`, req.VendorID).Scan(&vendorCompanyID, &apAccountCode); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("vendor %d not found", req.VendorID)
		}
		return nil, fmt.Errorf("resolve vendor AP account: %w", err)
	}
	if vendorCompanyID != req.CompanyID {
		return nil, fmt.Errorf("vendor %d does not belong to company %d", req.VendorID, req.CompanyID)
	}

	// Validate optional PO link ownership and status.
	if req.POID != nil {
		var poCompanyID, poVendorID int
		var poStatus string
		if err := tx.QueryRow(ctx, `
			SELECT company_id, vendor_id, status
			FROM purchase_orders
			WHERE id = $1
			FOR UPDATE
		`, *req.POID).Scan(&poCompanyID, &poVendorID, &poStatus); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("purchase order %d not found", *req.POID)
			}
			return nil, fmt.Errorf("fetch purchase order %d: %w", *req.POID, err)
		}
		if poCompanyID != req.CompanyID {
			return nil, fmt.Errorf("purchase order %d does not belong to company %d", *req.POID, req.CompanyID)
		}
		if poVendorID != req.VendorID {
			return nil, fmt.Errorf("purchase order %d vendor does not match vendor_id %d", *req.POID, req.VendorID)
		}
		if poStatus == "PAID" || poStatus == "CLOSED" {
			return nil, fmt.Errorf("purchase order %d cannot be linked in status %s", *req.POID, poStatus)
		}
	}

	// Validate lines and build PI proposal.
	var lineTotal decimal.Decimal
	proposalLines := make([]ProposalLine, 0, len(req.Lines)+1)
	for i, l := range req.Lines {
		accountCode := strings.TrimSpace(l.ExpenseAccountCode)
		if accountCode == "" {
			return nil, fmt.Errorf("line %d: expense_account_code is required", i+1)
		}
		if l.Amount.LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("line %d: amount must be > 0", i+1)
		}
		var accountExists bool
		if err := tx.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM accounts WHERE company_id = $1 AND code = $2)",
			req.CompanyID, accountCode,
		).Scan(&accountExists); err != nil {
			return nil, fmt.Errorf("line %d: validate expense account: %w", i+1, err)
		}
		if !accountExists {
			return nil, fmt.Errorf("line %d: expense account %q not found", i+1, accountCode)
		}

		lineTotal = lineTotal.Add(l.Amount)
		proposalLines = append(proposalLines, ProposalLine{
			AccountCode: accountCode,
			IsDebit:     true,
			Amount:      l.Amount.StringFixed(2),
		})
	}
	if !lineTotal.Round(2).Equal(req.InvoiceAmount.Round(2)) {
		return nil, fmt.Errorf("invoice_amount mismatch: expected sum(lines) %s, got %s",
			lineTotal.Round(2).StringFixed(2), req.InvoiceAmount.Round(2).StringFixed(2))
	}
	proposalLines = append(proposalLines, ProposalLine{
		AccountCode: apAccountCode,
		IsDebit:     false,
		Amount:      req.InvoiceAmount.Round(2).StringFixed(2),
	})

	proposal := Proposal{
		DocumentTypeCode:    "PI",
		CompanyCode:         req.CompanyCode,
		IdempotencyKey:      req.IdempotencyKey,
		TransactionCurrency: currency,
		ExchangeRate:        rate.String(),
		Summary:             fmt.Sprintf("Vendor invoice %s", req.InvoiceNumber),
		PostingDate:         postingDate.Format("2006-01-02"),
		DocumentDate:        documentDate.Format("2006-01-02"),
		Confidence:          1.0,
		Reasoning:           fmt.Sprintf("Direct vendor invoice %s.", req.InvoiceNumber),
		Lines:               proposalLines,
	}

	if err := ledger.CommitInTx(ctx, tx, proposal); err != nil {
		return nil, fmt.Errorf("post direct vendor invoice journal entry: %w", err)
	}

	var journalEntryID int
	var piDocumentNumber string
	if err := tx.QueryRow(ctx, `
		SELECT je.id, COALESCE(je.reference_id, '')
		FROM journal_entries je
		WHERE je.company_id = $1
		  AND je.idempotency_key = $2
		ORDER BY je.id DESC
		LIMIT 1
	`, req.CompanyID, req.IdempotencyKey).Scan(&journalEntryID, &piDocumentNumber); err != nil {
		return nil, fmt.Errorf("load posted PI journal entry: %w", err)
	}
	var piDocumentNumberPtr *string
	if strings.TrimSpace(piDocumentNumber) != "" {
		piDocumentNumberPtr = &piDocumentNumber
	}

	var vendorInvoiceID int
	if err := tx.QueryRow(ctx, `
		INSERT INTO vendor_invoices
		            (company_id, vendor_id, po_id, source, status, invoice_number, invoice_date,
		             currency, exchange_rate, invoice_amount, amount_paid, last_paid_at,
		             idempotency_key, pi_document_number, journal_entry_id, created_by_user_id)
		VALUES      ($1, $2, $3, $4, 'OPEN', $5, $6, $7, $8, $9, 0, NULL, $10, $11, $12, $13)
		RETURNING id
	`, req.CompanyID, req.VendorID, req.POID, source, req.InvoiceNumber, req.InvoiceDate.Format("2006-01-02"),
		currency, rate, req.InvoiceAmount.Round(2), req.IdempotencyKey, piDocumentNumberPtr, journalEntryID, req.CreatedByUserID,
	).Scan(&vendorInvoiceID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "idempotency") {
				return nil, fmt.Errorf("duplicate direct vendor invoice request: idempotency key %s already exists", req.IdempotencyKey)
			}
			if strings.Contains(pgErr.ConstraintName, "invoice_norm") {
				return nil, fmt.Errorf("duplicate vendor invoice number %q for this vendor", req.InvoiceNumber)
			}
		}
		return nil, fmt.Errorf("insert vendor invoice: %w", err)
	}

	for i, l := range req.Lines {
		if _, err := tx.Exec(ctx, `
			INSERT INTO vendor_invoice_lines
			            (vendor_invoice_id, line_number, description, expense_account_code, amount)
			VALUES      ($1, $2, $3, $4, $5)
		`, vendorInvoiceID, i+1, strings.TrimSpace(l.Description), strings.TrimSpace(l.ExpenseAccountCode), l.Amount.Round(2)); err != nil {
			return nil, fmt.Errorf("insert vendor invoice line %d: %w", i+1, err)
		}
	}

	if req.ClosePO && req.POID != nil {
		if err := s.closePOInTx(ctx, tx, req.CompanyID, *req.POID, req.CloseReason, req.ClosedByUserID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit direct vendor invoice: %w", err)
	}

	return s.GetVendorInvoice(ctx, req.CompanyID, vendorInvoiceID)
}

// PayVendorInvoice posts a PV settlement against vendor_invoices and updates payment status.
func (s *purchaseOrderService) PayVendorInvoice(ctx context.Context, req VendorInvoicePaymentInput, ledger *Ledger) (*VendorInvoice, error) {
	if req.CompanyID <= 0 {
		return nil, fmt.Errorf("company_id is required")
	}
	if strings.TrimSpace(req.CompanyCode) == "" {
		return nil, fmt.Errorf("company_code is required")
	}
	if req.VendorInvoiceID <= 0 {
		return nil, fmt.Errorf("vendor_invoice_id is required")
	}
	if strings.TrimSpace(req.BankAccountCode) == "" {
		return nil, fmt.Errorf("bank_account_code is required")
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amount must be > 0")
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		return nil, fmt.Errorf("idempotency_key is required")
	}
	if req.PaymentDate.IsZero() {
		req.PaymentDate = time.Now()
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var companyID int
	var status string
	var currency string
	var exchangeRate decimal.Decimal
	var invoiceAmount decimal.Decimal
	var amountPaid decimal.Decimal
	var apAccountCode string
	if err := tx.QueryRow(ctx, `
		SELECT vi.company_id, vi.status, vi.currency, vi.exchange_rate, vi.invoice_amount, vi.amount_paid,
		       COALESCE(a.code, v.ap_account_code, '2000')
		FROM vendor_invoices vi
		JOIN vendors v ON v.id = vi.vendor_id
		LEFT JOIN accounts a ON a.id = v.ap_account_id AND a.company_id = vi.company_id
		WHERE vi.id = $1
		FOR UPDATE OF vi
	`, req.VendorInvoiceID).Scan(
		&companyID, &status, &currency, &exchangeRate, &invoiceAmount, &amountPaid, &apAccountCode,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("vendor invoice %d not found", req.VendorInvoiceID)
		}
		return nil, fmt.Errorf("fetch vendor invoice %d: %w", req.VendorInvoiceID, err)
	}
	if companyID != req.CompanyID {
		return nil, fmt.Errorf("vendor invoice %d does not belong to company %d", req.VendorInvoiceID, req.CompanyID)
	}
	if status == "VOID" {
		return nil, fmt.Errorf("vendor invoice %d is VOID", req.VendorInvoiceID)
	}
	if status == "PAID" {
		return nil, fmt.Errorf("vendor invoice %d is already PAID", req.VendorInvoiceID)
	}

	remaining := invoiceAmount.Sub(amountPaid).Round(2)
	amount := req.Amount.Round(2)
	if amount.GreaterThan(remaining) {
		return nil, fmt.Errorf("payment amount %s exceeds remaining %s", amount.StringFixed(2), remaining.StringFixed(2))
	}

	proposal := Proposal{
		DocumentTypeCode:    "PV",
		CompanyCode:         req.CompanyCode,
		IdempotencyKey:      req.IdempotencyKey,
		TransactionCurrency: currency,
		ExchangeRate:        exchangeRate.String(),
		Summary:             fmt.Sprintf("Vendor payment for invoice %d", req.VendorInvoiceID),
		PostingDate:         req.PaymentDate.Format("2006-01-02"),
		DocumentDate:        req.PaymentDate.Format("2006-01-02"),
		Confidence:          1.0,
		Reasoning:           fmt.Sprintf("Payment against vendor invoice %d.", req.VendorInvoiceID),
		Lines: []ProposalLine{
			{AccountCode: apAccountCode, IsDebit: true, Amount: amount.StringFixed(2)},
			{AccountCode: strings.TrimSpace(req.BankAccountCode), IsDebit: false, Amount: amount.StringFixed(2)},
		},
	}
	if err := ledger.CommitInTx(ctx, tx, proposal); err != nil {
		return nil, fmt.Errorf("post vendor invoice payment journal entry: %w", err)
	}

	var journalEntryID int
	var paymentDocumentNumber string
	if err := tx.QueryRow(ctx, `
		SELECT je.id, COALESCE(je.reference_id, '')
		FROM journal_entries je
		WHERE je.company_id = $1
		  AND je.idempotency_key = $2
		ORDER BY je.id DESC
		LIMIT 1
	`, req.CompanyID, req.IdempotencyKey).Scan(&journalEntryID, &paymentDocumentNumber); err != nil {
		return nil, fmt.Errorf("load posted PV journal entry: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO vendor_invoice_payments
		            (vendor_invoice_id, payment_document_number, payment_amount, payment_date, journal_entry_id)
		VALUES      ($1, $2, $3, $4, $5)
	`, req.VendorInvoiceID, paymentDocumentNumber, amount, req.PaymentDate.Format("2006-01-02"), journalEntryID); err != nil {
		return nil, fmt.Errorf("insert vendor invoice payment: %w", err)
	}

	newPaid := amountPaid.Add(amount).Round(2)
	newStatus := "PARTIALLY_PAID"
	if newPaid.Equal(invoiceAmount.Round(2)) {
		newStatus = "PAID"
	}
	if newPaid.Equal(decimal.Zero) {
		newStatus = "OPEN"
	}

	if _, err := tx.Exec(ctx, `
		UPDATE vendor_invoices
		SET amount_paid = $1,
		    status = $2,
		    last_paid_at = NOW()
		WHERE id = $3
	`, newPaid, newStatus, req.VendorInvoiceID); err != nil {
		return nil, fmt.Errorf("update vendor invoice payment state: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit vendor invoice payment: %w", err)
	}
	return s.GetVendorInvoice(ctx, req.CompanyID, req.VendorInvoiceID)
}

// GetVendorInvoice returns one vendor invoice by ID, including lines and payments.
func (s *purchaseOrderService) GetVendorInvoice(ctx context.Context, companyID, vendorInvoiceID int) (*VendorInvoice, error) {
	var vi VendorInvoice
	if err := s.pool.QueryRow(ctx, `
		SELECT vi.id, vi.company_id, vi.vendor_id, v.code, v.name,
		       vi.po_id, vi.source, vi.status, vi.invoice_number, vi.invoice_date::text,
		       vi.currency, vi.exchange_rate, vi.invoice_amount, vi.amount_paid, vi.last_paid_at,
		       vi.idempotency_key, vi.pi_document_number, vi.journal_entry_id, vi.created_by_user_id,
		       vi.created_at
		FROM vendor_invoices vi
		JOIN vendors v ON v.id = vi.vendor_id
		WHERE vi.id = $1 AND vi.company_id = $2
	`, vendorInvoiceID, companyID).Scan(
		&vi.ID, &vi.CompanyID, &vi.VendorID, &vi.VendorCode, &vi.VendorName,
		&vi.POID, &vi.Source, &vi.Status, &vi.InvoiceNumber, &vi.InvoiceDate,
		&vi.Currency, &vi.ExchangeRate, &vi.InvoiceAmount, &vi.AmountPaid, &vi.LastPaidAt,
		&vi.IdempotencyKey, &vi.PIDocumentNumber, &vi.JournalEntryID, &vi.CreatedByUserID,
		&vi.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("vendor invoice %d not found", vendorInvoiceID)
		}
		return nil, fmt.Errorf("get vendor invoice %d: %w", vendorInvoiceID, err)
	}

	lineRows, err := s.pool.Query(ctx, `
		SELECT id, vendor_invoice_id, line_number, description, expense_account_code, amount, created_at
		FROM vendor_invoice_lines
		WHERE vendor_invoice_id = $1
		ORDER BY line_number
	`, vendorInvoiceID)
	if err != nil {
		return nil, fmt.Errorf("list vendor invoice lines: %w", err)
	}
	defer lineRows.Close()
	for lineRows.Next() {
		var l VendorInvoiceLine
		if err := lineRows.Scan(&l.ID, &l.VendorInvoiceID, &l.LineNumber, &l.Description, &l.ExpenseAccountCode, &l.Amount, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan vendor invoice line: %w", err)
		}
		vi.Lines = append(vi.Lines, l)
	}

	paymentRows, err := s.pool.Query(ctx, `
		SELECT id, vendor_invoice_id, payment_document_number, payment_amount, payment_date::text, journal_entry_id, created_at
		FROM vendor_invoice_payments
		WHERE vendor_invoice_id = $1
		ORDER BY id
	`, vendorInvoiceID)
	if err != nil {
		return nil, fmt.Errorf("list vendor invoice payments: %w", err)
	}
	defer paymentRows.Close()
	for paymentRows.Next() {
		var p VendorInvoicePayment
		if err := paymentRows.Scan(&p.ID, &p.VendorInvoiceID, &p.PaymentDocumentNumber, &p.PaymentAmount, &p.PaymentDate, &p.JournalEntryID, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan vendor invoice payment: %w", err)
		}
		vi.Payments = append(vi.Payments, p)
	}

	return &vi, nil
}
