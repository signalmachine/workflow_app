package app

import (
	"net/http"
	"strings"
	"time"

	"workflow_app/internal/accounting"
)

func (h *AgentAPIHandler) handleAdminLedgerAccounts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != adminLedgerAccountsPath {
		http.NotFound(w, r)
		return
	}
	if h.accountingAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "accounting admin service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.accountingAdmin.ListLedgerAccounts(r.Context(), accounting.ListLedgerAccountsInput{Actor: actor})
		if err != nil {
			handleAccountingAdminError(w, err, "failed to list ledger accounts")
			return
		}
		response := struct {
			Items []ledgerAccountResponse `json:"items"`
		}{Items: make([]ledgerAccountResponse, 0, len(items))}
		for _, item := range items {
			response.Items = append(response.Items, mapLedgerAccount(item))
		}
		writeJSON(w, http.StatusOK, response)
	case http.MethodPost:
		var req createLedgerAccountRequest
		if err := decodeJSONBody(r, &req, false); err != nil {
			writeJSONBodyError(w, err)
			return
		}
		account, err := h.accountingAdmin.CreateLedgerAccount(r.Context(), accounting.CreateLedgerAccountInput{
			Code:                strings.TrimSpace(req.Code),
			Name:                strings.TrimSpace(req.Name),
			AccountClass:        strings.TrimSpace(req.AccountClass),
			ControlType:         strings.TrimSpace(req.ControlType),
			AllowsDirectPosting: req.AllowsDirectPosting,
			TaxCategoryCode:     strings.TrimSpace(req.TaxCategoryCode),
			Actor:               actor,
		})
		if err != nil {
			handleAccountingAdminError(w, err, "failed to create ledger account")
			return
		}
		writeJSON(w, http.StatusCreated, mapLedgerAccount(account))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (h *AgentAPIHandler) handleAdminTaxCodes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != adminTaxCodesPath {
		http.NotFound(w, r)
		return
	}
	if h.accountingAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "accounting admin service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.accountingAdmin.ListTaxCodes(r.Context(), accounting.ListTaxCodesInput{Actor: actor})
		if err != nil {
			handleAccountingAdminError(w, err, "failed to list tax codes")
			return
		}
		response := struct {
			Items []taxCodeResponse `json:"items"`
		}{Items: make([]taxCodeResponse, 0, len(items))}
		for _, item := range items {
			response.Items = append(response.Items, mapTaxCode(item))
		}
		writeJSON(w, http.StatusOK, response)
	case http.MethodPost:
		var req createTaxCodeRequest
		if err := decodeJSONBody(r, &req, false); err != nil {
			writeJSONBodyError(w, err)
			return
		}
		code, err := h.accountingAdmin.CreateTaxCode(r.Context(), accounting.CreateTaxCodeInput{
			Code:                strings.TrimSpace(req.Code),
			Name:                strings.TrimSpace(req.Name),
			TaxType:             strings.TrimSpace(req.TaxType),
			RateBasisPoints:     req.RateBasisPoints,
			ReceivableAccountID: strings.TrimSpace(req.ReceivableAccountID),
			PayableAccountID:    strings.TrimSpace(req.PayableAccountID),
			Actor:               actor,
		})
		if err != nil {
			handleAccountingAdminError(w, err, "failed to create tax code")
			return
		}
		writeJSON(w, http.StatusCreated, mapTaxCode(code))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (h *AgentAPIHandler) handleAdminAccountingPeriods(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != adminPeriodsPath {
		http.NotFound(w, r)
		return
	}
	if h.accountingAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "accounting admin service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.accountingAdmin.ListAccountingPeriods(r.Context(), accounting.ListAccountingPeriodsInput{Actor: actor})
		if err != nil {
			handleAccountingAdminError(w, err, "failed to list accounting periods")
			return
		}
		response := struct {
			Items []accountingPeriodResponse `json:"items"`
		}{Items: make([]accountingPeriodResponse, 0, len(items))}
		for _, item := range items {
			response.Items = append(response.Items, mapAccountingPeriod(item))
		}
		writeJSON(w, http.StatusOK, response)
	case http.MethodPost:
		var req createAccountingPeriodRequest
		if err := decodeJSONBody(r, &req, false); err != nil {
			writeJSONBodyError(w, err)
			return
		}
		startOn, err := time.Parse(time.DateOnly, strings.TrimSpace(req.StartOn))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid accounting period"})
			return
		}
		endOn, err := time.Parse(time.DateOnly, strings.TrimSpace(req.EndOn))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid accounting period"})
			return
		}
		period, err := h.accountingAdmin.CreateAccountingPeriod(r.Context(), accounting.CreateAccountingPeriodInput{
			PeriodCode: strings.TrimSpace(req.PeriodCode),
			StartOn:    startOn,
			EndOn:      endOn,
			Actor:      actor,
		})
		if err != nil {
			handleAccountingAdminError(w, err, "failed to create accounting period")
			return
		}
		writeJSON(w, http.StatusCreated, mapAccountingPeriod(period))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (h *AgentAPIHandler) handleAdminAccountingPeriodAction(w http.ResponseWriter, r *http.Request) {
	if h.accountingAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "accounting admin service unavailable"})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	periodID, action, ok := parseChildActionPath(adminPeriodsPath, r.URL.Path)
	if !ok || !strings.EqualFold(strings.TrimSpace(action), "close") {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	period, err := h.accountingAdmin.CloseAccountingPeriod(r.Context(), accounting.CloseAccountingPeriodInput{
		PeriodID: periodID,
		Actor:    actor,
	})
	if err != nil {
		handleAccountingAdminError(w, err, "failed to close accounting period")
		return
	}
	writeJSON(w, http.StatusOK, mapAccountingPeriod(period))
}
