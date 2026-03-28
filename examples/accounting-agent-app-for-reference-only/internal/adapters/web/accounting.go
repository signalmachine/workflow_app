package web

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"accounting-agent/internal/app"
	"accounting-agent/internal/core"
	"accounting-agent/web/templates/pages"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ── Browser page handlers ─────────────────────────────────────────────────────

// trialBalancePage handles GET /reports/trial-balance.
func (h *Handler) trialBalancePage(w http.ResponseWriter, r *http.Request) {
	d := h.buildAppLayoutData(r, "Trial Balance", "trial-balance")

	if d.CompanyCode == "" {
		http.Error(w, "Company not resolved — please log in again", http.StatusUnauthorized)
		return
	}

	result, err := h.svc.GetTrialBalance(r.Context(), d.CompanyCode)
	if err != nil {
		d.FlashMsg = "Failed to load trial balance: " + err.Error()
		d.FlashKind = "error"
		result = &app.TrialBalanceResult{CompanyCode: d.CompanyCode}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.TrialBalance(d, result).Render(r.Context(), w)
}

// plReportPage handles GET /reports/pl.
func (h *Handler) plReportPage(w http.ResponseWriter, r *http.Request) {
	d := h.buildAppLayoutData(r, "P&L Report", "pl")

	if d.CompanyCode == "" {
		http.Error(w, "Company not resolved — please log in again", http.StatusUnauthorized)
		return
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil {
			month = parsed
		}
	}

	report, err := h.svc.GetProfitAndLoss(r.Context(), d.CompanyCode, year, month)
	if err != nil {
		d.FlashMsg = "Failed to load P&L: " + err.Error()
		d.FlashKind = "error"
		report = &core.PLReport{CompanyCode: d.CompanyCode, Year: year, Month: month}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.PLReport(d, report, year, month).Render(r.Context(), w)
}

// balanceSheetPage handles GET /reports/balance-sheet.
func (h *Handler) balanceSheetPage(w http.ResponseWriter, r *http.Request) {
	d := h.buildAppLayoutData(r, "Balance Sheet", "balance-sheet")

	if d.CompanyCode == "" {
		http.Error(w, "Company not resolved — please log in again", http.StatusUnauthorized)
		return
	}

	asOfDate := r.URL.Query().Get("date")
	if asOfDate == "" {
		asOfDate = time.Now().Format("2006-01-02")
	}

	report, err := h.svc.GetBalanceSheet(r.Context(), d.CompanyCode, asOfDate)
	if err != nil {
		d.FlashMsg = "Failed to load balance sheet: " + err.Error()
		d.FlashKind = "error"
		report = &core.BSReport{CompanyCode: d.CompanyCode, AsOfDate: asOfDate}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.BalanceSheet(d, report, asOfDate).Render(r.Context(), w)
}

// controlAccountReconciliationPage handles GET /reports/control-account-reconciliation.
func (h *Handler) controlAccountReconciliationPage(w http.ResponseWriter, r *http.Request) {
	d := h.buildAppLayoutData(r, "Control Account Reconciliation", "control-account-reconciliation")

	if d.CompanyCode == "" {
		http.Error(w, "Company not resolved — please log in again", http.StatusUnauthorized)
		return
	}

	asOfDate := r.URL.Query().Get("date")
	if asOfDate == "" {
		asOfDate = time.Now().Format("2006-01-02")
	}

	report, err := h.svc.GetControlAccountReconciliation(r.Context(), d.CompanyCode, asOfDate)
	if err != nil {
		d.FlashMsg = "Failed to load reconciliation: " + err.Error()
		d.FlashKind = "error"
		report = &core.ControlAccountReconciliationReport{CompanyCode: d.CompanyCode, AsOfDate: asOfDate}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.ControlAccountReconciliation(d, report, asOfDate).Render(r.Context(), w)
}

// documentTypeGovernancePage handles GET /reports/document-type-governance.
func (h *Handler) documentTypeGovernancePage(w http.ResponseWriter, r *http.Request) {
	d := h.buildAppLayoutData(r, "Document Type Governance", "document-type-governance")

	if d.CompanyCode == "" {
		http.Error(w, "Company not resolved — please log in again", http.StatusUnauthorized)
		return
	}

	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")
	if toDate == "" {
		toDate = time.Now().Format("2006-01-02")
	}
	if fromDate == "" {
		fromDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}

	report, err := h.svc.GetDocumentTypeGovernance(r.Context(), d.CompanyCode, fromDate, toDate)
	if err != nil {
		d.FlashMsg = "Failed to load document type governance report: " + err.Error()
		d.FlashKind = "error"
		report = &core.DocumentTypeGovernanceReport{
			CompanyCode: d.CompanyCode,
			FromDate:    fromDate,
			ToDate:      toDate,
			Counts:      []core.DocumentTypeCount{},
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.DocumentTypeGovernance(d, report, fromDate, toDate).Render(r.Context(), w)
}

// accountStatementPage handles GET /reports/statement.
// When format=csv, streams CSV instead of HTML.
func (h *Handler) accountStatementPage(w http.ResponseWriter, r *http.Request) {
	accountCode := r.URL.Query().Get("account")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	format := r.URL.Query().Get("format")

	d := h.buildAppLayoutData(r, "Account Statement", "statement")

	if d.CompanyCode == "" {
		http.Error(w, "Company not resolved — please log in again", http.StatusUnauthorized)
		return
	}

	if accountCode == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AccountStatement(d, nil, "", from, to).Render(r.Context(), w)
		return
	}

	stmtResult, err := h.svc.GetAccountStatement(r.Context(), d.CompanyCode, accountCode, from, to)
	if err != nil {
		d.FlashMsg = "Failed to load statement: " + err.Error()
		d.FlashKind = "error"
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.AccountStatement(d, nil, accountCode, from, to).Render(r.Context(), w)
		return
	}

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="statement-`+accountCode+`.csv"`)
		cw := csv.NewWriter(w)
		_ = cw.Write([]string{"Date", "Narration", "Reference", "Debit", "Credit", "Balance"})
		for _, line := range stmtResult.Lines {
			_ = cw.Write([]string{
				line.PostingDate,
				csvSafe(line.Narration),
				csvSafe(line.Reference),
				line.Debit.StringFixed(2),
				line.Credit.StringFixed(2),
				line.RunningBalance.StringFixed(2),
			})
		}
		cw.Flush()
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.AccountStatement(d, stmtResult, accountCode, from, to).Render(r.Context(), w)
}

// journalEntryPage handles GET /accounting/journal-entry.
func (h *Handler) journalEntryPage(w http.ResponseWriter, r *http.Request) {
	d := h.buildAppLayoutData(r, "Journal Entry", "reports")

	if d.CompanyCode == "" {
		http.Error(w, "Company not resolved — please log in again", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.JournalEntry(d, d.CompanyCode).Render(r.Context(), w)
}

// csvSafe prevents CSV formula injection by prefixing cells that begin with a
// formula-triggering character with a single quote.
func csvSafe(s string) string {
	if len(s) == 0 {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + s
	}
	return s
}

// ── API handlers ──────────────────────────────────────────────────────────────

// apiTrialBalance handles GET /api/companies/{code}/trial-balance.
func (h *Handler) apiTrialBalance(w http.ResponseWriter, r *http.Request) {
	code := companyCode(r)
	if !h.requireCompanyAccess(w, r, code) {
		return
	}
	result, err := h.svc.GetTrialBalance(r.Context(), code)
	if err != nil {
		writeError(w, r, err.Error(), "INTERNAL", http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

// apiAccountStatement handles GET /api/companies/{code}/accounts/{accountCode}/statement.
func (h *Handler) apiAccountStatement(w http.ResponseWriter, r *http.Request) {
	code := companyCode(r)
	if !h.requireCompanyAccess(w, r, code) {
		return
	}
	result, err := h.svc.GetAccountStatement(r.Context(),
		code,
		chi.URLParam(r, "accountCode"),
		r.URL.Query().Get("from"),
		r.URL.Query().Get("to"),
	)
	if err != nil {
		writeError(w, r, err.Error(), "INTERNAL", http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

// apiProfitAndLoss handles GET /api/companies/{code}/reports/pl.
func (h *Handler) apiProfitAndLoss(w http.ResponseWriter, r *http.Request) {
	code := companyCode(r)
	if !h.requireCompanyAccess(w, r, code) {
		return
	}
	now := time.Now()
	year, month := now.Year(), int(now.Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil {
			month = parsed
		}
	}

	result, err := h.svc.GetProfitAndLoss(r.Context(), code, year, month)
	if err != nil {
		writeError(w, r, err.Error(), "INTERNAL", http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

// apiBalanceSheet handles GET /api/companies/{code}/reports/balance-sheet.
func (h *Handler) apiBalanceSheet(w http.ResponseWriter, r *http.Request) {
	code := companyCode(r)
	if !h.requireCompanyAccess(w, r, code) {
		return
	}
	result, err := h.svc.GetBalanceSheet(r.Context(), code, r.URL.Query().Get("date"))
	if err != nil {
		writeError(w, r, err.Error(), "INTERNAL", http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

// apiControlAccountReconciliation handles GET /api/companies/{code}/reports/control-account-reconciliation.
func (h *Handler) apiControlAccountReconciliation(w http.ResponseWriter, r *http.Request) {
	code := companyCode(r)
	if !h.requireCompanyAccess(w, r, code) {
		return
	}
	result, err := h.svc.GetControlAccountReconciliation(r.Context(), code, r.URL.Query().Get("date"))
	if err != nil {
		writeError(w, r, err.Error(), "INTERNAL", http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

// apiDocumentTypeGovernance handles GET /api/companies/{code}/reports/document-type-governance.
func (h *Handler) apiDocumentTypeGovernance(w http.ResponseWriter, r *http.Request) {
	code := companyCode(r)
	if !h.requireCompanyAccess(w, r, code) {
		return
	}
	result, err := h.svc.GetDocumentTypeGovernance(
		r.Context(),
		code,
		r.URL.Query().Get("from"),
		r.URL.Query().Get("to"),
	)
	if err != nil {
		writeError(w, r, err.Error(), "INTERNAL", http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

// apiManualJEControlAccountHits handles GET /api/companies/{code}/reports/control-account-journal-entries.
func (h *Handler) apiManualJEControlAccountHits(w http.ResponseWriter, r *http.Request) {
	code := companyCode(r)
	if !h.requireCompanyAccess(w, r, code) {
		return
	}
	result, err := h.svc.GetManualJEControlAccountHits(
		r.Context(),
		code,
		r.URL.Query().Get("from"),
		r.URL.Query().Get("to"),
	)
	if err != nil {
		writeError(w, r, err.Error(), "INTERNAL", http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

// ── Journal Entry API ─────────────────────────────────────────────────────────

// journalEntryRequest is the JSON body for manual journal entry creation/validation.
type journalEntryRequest struct {
	Narration               string `json:"narration"`
	PostingDate             string `json:"posting_date"`
	DocumentDate            string `json:"document_date"`
	Currency                string `json:"currency"`
	ExchangeRate            string `json:"exchange_rate"`
	OverrideControlAccounts bool   `json:"override_control_accounts"`
	OverrideReason          string `json:"override_reason"`
	Lines                   []struct {
		AccountCode string `json:"account_code"`
		Debit       string `json:"debit"`
		Credit      string `json:"credit"`
	} `json:"lines"`
}

type journalEntryResponse struct {
	Status         string   `json:"status"`
	IdempotencyKey string   `json:"idempotency_key,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
}

// apiPostJournalEntry handles POST /api/companies/{code}/journal-entries.
func (h *Handler) apiPostJournalEntry(w http.ResponseWriter, r *http.Request) {
	code := companyCode(r)
	if !h.requireCompanyAccess(w, r, code) {
		return
	}

	var req journalEntryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Currency == "" {
		company, err := h.svc.GetCompanyByCode(r.Context(), code)
		if err != nil {
			writeError(w, r, "failed to resolve company currency: "+err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
			return
		}
		req.Currency = company.BaseCurrency
	}

	proposal, err := buildProposal(code, req)
	if err != nil {
		writeError(w, r, err.Error(), "BAD_REQUEST", http.StatusBadRequest)
		return
	}

	warnings, warningDetails, err := h.manualJEControlWarnings(r, code, proposal)
	if err != nil {
		writeError(w, r, err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}
	mode := controlAccountEnforcementMode()
	if blocked, status, errCode, msg := h.manualJEEnforcementBlock(r, req, warningDetails, mode); blocked {
		h.auditManualJEControlWarningAttempt(r, code, req, proposal, "post", warningDetails, mode, true)
		writeError(w, r, msg, errCode, status)
		return
	}
	h.auditManualJEControlWarningAttempt(r, code, req, proposal, "post", warningDetails, mode, false)

	commitCtx := app.WithProposalSource(r.Context(), app.ProposalSourceManualWeb)
	role := ""
	if claims := authFromContext(r.Context()); claims != nil {
		role = claims.Role
	}
	commitCtx = app.WithControlAccountOverride(commitCtx, req.OverrideControlAccounts, req.OverrideReason, role)
	if err := h.svc.CommitProposal(commitCtx, proposal); err != nil {
		writeError(w, r, err.Error(), "COMMIT_FAILED", http.StatusUnprocessableEntity)
		return
	}

	if claims := authFromContext(r.Context()); claims != nil {
		if err := h.svc.SetJournalEntryCreatedBy(r.Context(), code, proposal.IdempotencyKey, claims.UserID); err != nil {
			log.Printf("set created_by_user_id for manual JE failed: %v", err)
		}
	}

	writeJSON(w, journalEntryResponse{Status: "posted", IdempotencyKey: proposal.IdempotencyKey, Warnings: warnings})
}

// apiValidateJournalEntry handles POST /api/companies/{code}/journal-entries/validate.
func (h *Handler) apiValidateJournalEntry(w http.ResponseWriter, r *http.Request) {
	code := companyCode(r)
	if !h.requireCompanyAccess(w, r, code) {
		return
	}

	var req journalEntryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Currency == "" {
		company, err := h.svc.GetCompanyByCode(r.Context(), code)
		if err != nil {
			writeError(w, r, "failed to resolve company currency: "+err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
			return
		}
		req.Currency = company.BaseCurrency
	}

	proposal, err := buildProposal(code, req)
	if err != nil {
		writeError(w, r, err.Error(), "BAD_REQUEST", http.StatusBadRequest)
		return
	}

	warnings, warningDetails, err := h.manualJEControlWarnings(r, code, proposal)
	if err != nil {
		writeError(w, r, err.Error(), "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}
	mode := controlAccountEnforcementMode()
	if blocked, status, errCode, msg := h.manualJEEnforcementBlock(r, req, warningDetails, mode); blocked {
		h.auditManualJEControlWarningAttempt(r, code, req, proposal, "validate", warningDetails, mode, true)
		writeError(w, r, msg, errCode, status)
		return
	}
	h.auditManualJEControlWarningAttempt(r, code, req, proposal, "validate", warningDetails, mode, false)

	validateCtx := app.WithProposalSource(r.Context(), app.ProposalSourceManualWeb)
	role := ""
	if claims := authFromContext(r.Context()); claims != nil {
		role = claims.Role
	}
	validateCtx = app.WithControlAccountOverride(validateCtx, req.OverrideControlAccounts, req.OverrideReason, role)
	if err := h.svc.ValidateProposal(validateCtx, proposal); err != nil {
		writeError(w, r, err.Error(), "VALIDATION_FAILED", http.StatusUnprocessableEntity)
		return
	}

	writeJSON(w, journalEntryResponse{Status: "valid", Warnings: warnings})
}

func (h *Handler) manualJEControlWarnings(r *http.Request, companyCode string, proposal core.Proposal) ([]string, []app.ManualJEControlAccountWarning, error) {
	mode := controlAccountEnforcementMode()
	if mode == "off" {
		return nil, nil, nil
	}

	details, err := h.svc.GetManualJEControlAccountWarnings(r.Context(), companyCode, uniqueAccountCodesFromProposal(proposal))
	if err != nil {
		return nil, nil, err
	}
	warnings := make([]string, 0, len(details))
	for _, d := range details {
		controlType := d.ControlType
		if controlType == "" {
			controlType = "CONTROL"
		}
		warnings = append(warnings, fmt.Sprintf("Account %s (%s) is a %s control account. Prefer sales/purchase/inventory flow.", d.AccountCode, d.AccountName, controlType))
	}
	return warnings, details, nil
}

func (h *Handler) manualJEEnforcementBlock(r *http.Request, req journalEntryRequest, details []app.ManualJEControlAccountWarning, mode string) (bool, int, string, string) {
	if mode != "enforce" || len(details) == 0 {
		return false, 0, "", ""
	}

	if !req.OverrideControlAccounts {
		return true, http.StatusUnprocessableEntity, "CONTROL_ACCOUNT_ENFORCED", "Control account detected. This posting is blocked in enforce mode unless an admin provides override_reason."
	}
	if strings.TrimSpace(req.OverrideReason) == "" {
		return true, http.StatusUnprocessableEntity, "CONTROL_ACCOUNT_OVERRIDE_REASON_REQUIRED", "override_reason is required when override_control_accounts is true."
	}

	claims := authFromContext(r.Context())
	if claims == nil || claims.Role != "ADMIN" {
		return true, http.StatusForbidden, "FORBIDDEN", "Only ADMIN can override control-account enforcement."
	}

	return false, 0, "", ""
}

func (h *Handler) auditManualJEControlWarningAttempt(r *http.Request, companyCode string, req journalEntryRequest, proposal core.Proposal, action string, details []app.ManualJEControlAccountWarning, mode string, isBlocked bool) {
	if len(details) == 0 {
		return
	}

	var userID *int
	username := ""
	if claims := authFromContext(r.Context()); claims != nil {
		userID = &claims.UserID
		username = claims.Username
	}

	auditReq := app.ManualJEControlAccountAttemptRequest{
		CompanyCode:             companyCode,
		UserID:                  userID,
		Username:                username,
		Action:                  action,
		PostingDate:             req.PostingDate,
		Narration:               proposal.Summary,
		AccountCodes:            uniqueAccountCodesFromProposal(proposal),
		WarningDetails:          details,
		EnforcementMode:         mode,
		OverrideControlAccounts: req.OverrideControlAccounts,
		OverrideReason:          req.OverrideReason,
		IsBlocked:               isBlocked,
	}
	if err := h.svc.RecordManualJEControlAccountAttempt(r.Context(), auditReq); err != nil {
		log.Printf("record manual JE control-account audit failed: %v", err)
	}
}

func controlAccountEnforcementMode() string {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("CONTROL_ACCOUNT_ENFORCEMENT_MODE")))
	switch mode {
	case "off", "warn", "enforce":
		return mode
	default:
		return "warn"
	}
}

func uniqueAccountCodesFromProposal(proposal core.Proposal) []string {
	seen := make(map[string]struct{}, len(proposal.Lines))
	codes := make([]string, 0, len(proposal.Lines))
	for _, l := range proposal.Lines {
		code := strings.TrimSpace(l.AccountCode)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	return codes
}

// buildProposal converts a journalEntryRequest into a core.Proposal.
func buildProposal(code string, req journalEntryRequest) (core.Proposal, error) {
	if req.Narration == "" {
		return core.Proposal{}, fmt.Errorf("narration is required")
	}
	if req.PostingDate == "" {
		return core.Proposal{}, fmt.Errorf("posting_date is required")
	}

	currency := req.Currency
	if currency == "" {
		return core.Proposal{}, fmt.Errorf("currency is required")
	}
	exchangeRate := req.ExchangeRate
	if exchangeRate == "" {
		exchangeRate = "1.0"
	}
	docDate := req.DocumentDate
	if docDate == "" {
		docDate = req.PostingDate
	}

	var lines []core.ProposalLine
	for _, l := range req.Lines {
		if l.AccountCode == "" {
			continue
		}

		debitAmt := decimal.Zero
		creditAmt := decimal.Zero

		if l.Debit != "" && l.Debit != "0" {
			d, err := decimal.NewFromString(l.Debit)
			if err != nil {
				return core.Proposal{}, fmt.Errorf("invalid debit amount %q: %w", l.Debit, err)
			}
			debitAmt = d
		}
		if l.Credit != "" && l.Credit != "0" {
			c, err := decimal.NewFromString(l.Credit)
			if err != nil {
				return core.Proposal{}, fmt.Errorf("invalid credit amount %q: %w", l.Credit, err)
			}
			creditAmt = c
		}

		if debitAmt.IsPositive() {
			lines = append(lines, core.ProposalLine{
				AccountCode: l.AccountCode,
				IsDebit:     true,
				Amount:      debitAmt.StringFixed(2),
			})
		}
		if creditAmt.IsPositive() {
			lines = append(lines, core.ProposalLine{
				AccountCode: l.AccountCode,
				IsDebit:     false,
				Amount:      creditAmt.StringFixed(2),
			})
		}
	}

	if len(lines) < 2 {
		return core.Proposal{}, fmt.Errorf("at least two non-zero journal lines are required")
	}

	return core.Proposal{
		DocumentTypeCode:    "JE",
		CompanyCode:         code,
		IdempotencyKey:      uuid.New().String(),
		TransactionCurrency: currency,
		ExchangeRate:        exchangeRate,
		Summary:             req.Narration,
		PostingDate:         req.PostingDate,
		DocumentDate:        docDate,
		Confidence:          1.0,
		Reasoning:           "Manual journal entry submitted via web UI",
		Lines:               lines,
	}, nil
}
