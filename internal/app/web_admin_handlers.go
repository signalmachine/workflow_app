package app

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/parties"
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
	return sessionContext, true
}

func (h *AgentAPIHandler) requireWebAdminSessionWithService(w http.ResponseWriter, r *http.Request, serviceAvailable bool, unavailableMessage string) (identityaccess.SessionContext, bool) {
	sessionContext, ok := h.requireWebAdminSession(w, r)
	if !ok {
		return identityaccess.SessionContext{}, false
	}
	if !serviceAvailable {
		http.Redirect(w, r, webAdminPath+"?error="+url.QueryEscape(unavailableMessage), http.StatusSeeOther)
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

func partyAdminWebErrorMessage(err error, fallback string) string {
	switch {
	case err == nil:
		return ""
	case err == identityaccess.ErrUnauthorized:
		return "unauthorized"
	case err == parties.ErrInvalidParty:
		return "invalid party"
	case err == parties.ErrPartyNotFound:
		return "party not found"
	case err == parties.ErrInvalidContact:
		return "invalid contact"
	case err == parties.ErrContactNotFound:
		return "contact not found"
	default:
		return fallback
	}
}

func adminPartiesPathWithMessage(key, message string) string {
	if strings.TrimSpace(message) == "" {
		return webAdminPartiesPath
	}
	return appendWebMessage(webAdminPartiesPath, key, message)
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

	sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.accountingAdmin != nil, "accounting admin service unavailable")
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

func (h *AgentAPIHandler) handleWebAdminParties(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAdminPartiesPath {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.partiesAdmin != nil, "party admin service unavailable")
		if !ok {
			return
		}

		data := webAdminPartiesData{
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			PartyKindOptions: []string{
				parties.PartyKindCustomer,
				parties.PartyKindVendor,
				parties.PartyKindCustomerVendor,
				parties.PartyKindOther,
			},
		}

		items, err := h.partiesAdmin.ListParties(r.Context(), parties.ListPartiesInput{Actor: sessionContext.Actor})
		if err != nil {
			data.Error = partyAdminWebErrorMessage(err, "failed to load parties")
		} else {
			data.Parties = items
		}

		h.renderWebPage(w, webPageData{
			Title:        "workflow_app",
			ActivePath:   webAdminPath,
			Notice:       data.Notice,
			Error:        data.Error,
			Session:      &sessionContext,
			AdminParties: &data,
		})
	case http.MethodPost:
		sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.partiesAdmin != nil, "party admin service unavailable")
		if !ok {
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, adminPartiesPathWithMessage("error", "invalid party form"), http.StatusSeeOther)
			return
		}

		_, err := h.partiesAdmin.CreateParty(r.Context(), parties.CreatePartyInput{
			PartyCode:   strings.TrimSpace(r.FormValue("party_code")),
			DisplayName: strings.TrimSpace(r.FormValue("display_name")),
			LegalName:   strings.TrimSpace(r.FormValue("legal_name")),
			PartyKind:   strings.TrimSpace(r.FormValue("party_kind")),
			Actor:       sessionContext.Actor,
		})
		if err != nil {
			http.Redirect(w, r, adminPartiesPathWithMessage("error", partyAdminWebErrorMessage(err, "failed to create party")), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, adminPartiesPathWithMessage("notice", "Party created."), http.StatusSeeOther)
	default:
		http.NotFound(w, r)
	}
}

func (h *AgentAPIHandler) handleWebAdminPartyDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	partyID, ok := parseChildPath(strings.TrimSuffix(webAdminPartyDetailPrefix, "/"), r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.partiesAdmin != nil, "party admin service unavailable")
	if !ok {
		return
	}

	data := webAdminPartiesData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
		PartyKindOptions: []string{
			parties.PartyKindCustomer,
			parties.PartyKindVendor,
			parties.PartyKindCustomerVendor,
			parties.PartyKindOther,
		},
	}

	items, err := h.partiesAdmin.ListParties(r.Context(), parties.ListPartiesInput{Actor: sessionContext.Actor})
	if err != nil {
		data.Error = partyAdminWebErrorMessage(err, "failed to load parties")
	} else {
		data.Parties = items
	}

	party, err := h.partiesAdmin.GetParty(r.Context(), parties.GetPartyInput{
		PartyID: partyID,
		Actor:   sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, adminPartiesPathWithMessage("error", partyAdminWebErrorMessage(err, "failed to load party")), http.StatusSeeOther)
		return
	}

	contacts, err := h.partiesAdmin.ListContacts(r.Context(), parties.ListContactsInput{
		PartyID: partyID,
		Actor:   sessionContext.Actor,
	})
	if err != nil {
		data.Error = partyAdminWebErrorMessage(err, "failed to load party contacts")
	}
	data.Detail = &webAdminPartyDetailData{
		Party:    party,
		Contacts: contacts,
	}

	h.renderWebPage(w, webPageData{
		Title:        "workflow_app",
		ActivePath:   webAdminPath,
		Notice:       data.Notice,
		Error:        data.Error,
		Session:      &sessionContext,
		AdminParties: &data,
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

	sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.accountingAdmin != nil, "accounting admin service unavailable")
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

	sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.accountingAdmin != nil, "accounting admin service unavailable")
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

	sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.accountingAdmin != nil, "accounting admin service unavailable")
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

	sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.accountingAdmin != nil, "accounting admin service unavailable")
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
