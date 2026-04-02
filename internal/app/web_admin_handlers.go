package app

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/identityaccess"
)

func (h *AgentAPIHandler) requireWebAdminSession(w http.ResponseWriter, r *http.Request) (identityaccess.SessionContext, bool) {
	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webLoginPath+"?notice="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return identityaccess.SessionContext{}, false
	}
	if !strings.EqualFold(strings.TrimSpace(sessionContext.RoleCode), identityaccess.RoleAdmin) {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("admin surface requires admin role"), http.StatusSeeOther)
		return identityaccess.SessionContext{}, false
	}
	if h.accountingAdmin == nil {
		http.Redirect(w, r, webAdminPath+"?error="+url.QueryEscape("accounting admin service unavailable"), http.StatusSeeOther)
		return identityaccess.SessionContext{}, false
	}
	return sessionContext, true
}

func accountingAdminWebErrorMessage(err error, fallback string) string {
	switch {
	case err == nil:
		return ""
	case err == identityaccess.ErrUnauthorized:
		return "unauthorized"
	case err == accounting.ErrInvalidAccount:
		return "invalid ledger account"
	case err == accounting.ErrInvalidTaxCode:
		return "invalid tax code"
	case err == accounting.ErrInvalidAccountingPeriod:
		return "invalid accounting period"
	case err == accounting.ErrLedgerAccountNotFound:
		return "ledger account not found"
	case err == accounting.ErrTaxCodeNotFound:
		return "tax code not found"
	case err == accounting.ErrAccountingPeriodNotFound:
		return "accounting period not found"
	case err == accounting.ErrAccountingPeriodOverlap || err == accounting.ErrAccountingPeriodNotOpen:
		return err.Error()
	default:
		return fallback
	}
}

func adminAccountingPathWithMessage(key, message string) string {
	if strings.TrimSpace(message) == "" {
		return webAdminAccountingPath
	}
	return appendWebMessage(webAdminAccountingPath, key, message)
}

func (h *AgentAPIHandler) handleWebAdminAccounting(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAdminAccountingPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, ok := h.requireWebAdminSession(w, r)
	if !ok {
		return
	}

	data := webAdminAccountingData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
		AccountClassOptions: []string{
			accounting.AccountClassAsset,
			accounting.AccountClassLiability,
			accounting.AccountClassEquity,
			accounting.AccountClassRevenue,
			accounting.AccountClassExpense,
		},
		ControlTypeOptions: []string{
			accounting.ControlTypeNone,
			accounting.ControlTypeReceivable,
			accounting.ControlTypePayable,
			accounting.ControlTypeGSTInput,
			accounting.ControlTypeGSTOutput,
			accounting.ControlTypeTDSReceivable,
			accounting.ControlTypeTDSPayable,
		},
		TaxTypeOptions: []string{
			accounting.TaxTypeGST,
			accounting.TaxTypeTDS,
		},
	}

	ledgerAccounts, err := h.accountingAdmin.ListLedgerAccounts(r.Context(), accounting.ListLedgerAccountsInput{Actor: sessionContext.Actor})
	if err != nil {
		data.Error = accountingAdminWebErrorMessage(err, "failed to load ledger accounts")
	} else {
		data.LedgerAccounts = ledgerAccounts
	}
	taxCodes, err := h.accountingAdmin.ListTaxCodes(r.Context(), accounting.ListTaxCodesInput{Actor: sessionContext.Actor})
	if err != nil && data.Error == "" {
		data.Error = accountingAdminWebErrorMessage(err, "failed to load tax codes")
	} else if err == nil {
		data.TaxCodes = taxCodes
	}
	periods, err := h.accountingAdmin.ListAccountingPeriods(r.Context(), accounting.ListAccountingPeriodsInput{Actor: sessionContext.Actor})
	if err != nil && data.Error == "" {
		data.Error = accountingAdminWebErrorMessage(err, "failed to load accounting periods")
	} else if err == nil {
		data.AccountingPeriods = periods
	}

	h.renderWebPage(w, webPageData{
		Title:           "workflow_app",
		ActivePath:      webAdminPath,
		Notice:          data.Notice,
		Error:           data.Error,
		Session:         &sessionContext,
		AdminAccounting: &data,
	})
}

func (h *AgentAPIHandler) handleWebCreateLedgerAccount(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAdminLedgerAccountsPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	sessionContext, ok := h.requireWebAdminSession(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", "invalid ledger account form"), http.StatusSeeOther)
		return
	}

	_, err := h.accountingAdmin.CreateLedgerAccount(r.Context(), accounting.CreateLedgerAccountInput{
		Code:                strings.TrimSpace(r.FormValue("code")),
		Name:                strings.TrimSpace(r.FormValue("name")),
		AccountClass:        strings.TrimSpace(r.FormValue("account_class")),
		ControlType:         strings.TrimSpace(r.FormValue("control_type")),
		AllowsDirectPosting: strings.TrimSpace(r.FormValue("allows_direct_posting")) != "",
		TaxCategoryCode:     strings.TrimSpace(r.FormValue("tax_category_code")),
		Actor:               sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", accountingAdminWebErrorMessage(err, "failed to create ledger account")), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, adminAccountingPathWithMessage("notice", "Ledger account created."), http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebCreateTaxCode(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAdminTaxCodesPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	sessionContext, ok := h.requireWebAdminSession(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", "invalid tax code form"), http.StatusSeeOther)
		return
	}

	rateBasisPoints, err := strconv.Atoi(strings.TrimSpace(r.FormValue("rate_basis_points")))
	if err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", "invalid tax code"), http.StatusSeeOther)
		return
	}

	_, err = h.accountingAdmin.CreateTaxCode(r.Context(), accounting.CreateTaxCodeInput{
		Code:                strings.TrimSpace(r.FormValue("code")),
		Name:                strings.TrimSpace(r.FormValue("name")),
		TaxType:             strings.TrimSpace(r.FormValue("tax_type")),
		RateBasisPoints:     rateBasisPoints,
		ReceivableAccountID: strings.TrimSpace(r.FormValue("receivable_account_id")),
		PayableAccountID:    strings.TrimSpace(r.FormValue("payable_account_id")),
		Actor:               sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", accountingAdminWebErrorMessage(err, "failed to create tax code")), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, adminAccountingPathWithMessage("notice", "Tax code created."), http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebAccountingPeriods(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAdminPeriodsPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	sessionContext, ok := h.requireWebAdminSession(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", "invalid accounting period form"), http.StatusSeeOther)
		return
	}

	startOn, err := time.Parse(time.DateOnly, strings.TrimSpace(r.FormValue("start_on")))
	if err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", "invalid accounting period"), http.StatusSeeOther)
		return
	}
	endOn, err := time.Parse(time.DateOnly, strings.TrimSpace(r.FormValue("end_on")))
	if err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", "invalid accounting period"), http.StatusSeeOther)
		return
	}

	_, err = h.accountingAdmin.CreateAccountingPeriod(r.Context(), accounting.CreateAccountingPeriodInput{
		PeriodCode: strings.TrimSpace(r.FormValue("period_code")),
		StartOn:    startOn,
		EndOn:      endOn,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", accountingAdminWebErrorMessage(err, "failed to create accounting period")), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, adminAccountingPathWithMessage("notice", "Accounting period created."), http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebAccountingPeriodAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	periodID, action, ok := parseChildActionPath(webAdminPeriodsPath, r.URL.Path)
	if !ok || !strings.EqualFold(strings.TrimSpace(action), "close") {
		http.NotFound(w, r)
		return
	}

	sessionContext, ok := h.requireWebAdminSession(w, r)
	if !ok {
		return
	}

	if _, err := h.accountingAdmin.CloseAccountingPeriod(r.Context(), accounting.CloseAccountingPeriodInput{
		PeriodID: periodID,
		Actor:    sessionContext.Actor,
	}); err != nil {
		http.Redirect(w, r, adminAccountingPathWithMessage("error", accountingAdminWebErrorMessage(err, "failed to close accounting period")), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, adminAccountingPathWithMessage("notice", "Accounting period closed."), http.StatusSeeOther)
}
