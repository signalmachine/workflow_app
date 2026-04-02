package app

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/inventoryops"
	"workflow_app/internal/parties"
)

func writeAdminActorError(w http.ResponseWriter, err error) {
	if errors.Is(err, identityaccess.ErrUnauthorized) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
}

func (h *AgentAPIHandler) handleAdminLedgerAccounts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != adminLedgerAccountsPath {
		http.NotFound(w, r)
		return
	}
	if h.accountingAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "accounting admin service unavailable"})
		return
	}

	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
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

	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
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

	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
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

	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
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

func (h *AgentAPIHandler) handleAdminParties(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != adminPartiesPath {
		http.NotFound(w, r)
		return
	}
	if h.partiesAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "party admin service unavailable"})
		return
	}

	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.partiesAdmin.ListParties(r.Context(), parties.ListPartiesInput{
			PartyKind: strings.TrimSpace(r.URL.Query().Get("party_kind")),
			Actor:     actor,
		})
		if err != nil {
			handlePartyAdminError(w, err, "failed to list parties")
			return
		}
		response := struct {
			Items []partyResponse `json:"items"`
		}{Items: make([]partyResponse, 0, len(items))}
		for _, item := range items {
			response.Items = append(response.Items, mapParty(item))
		}
		writeJSON(w, http.StatusOK, response)
	case http.MethodPost:
		var req createPartyRequest
		if err := decodeJSONBody(r, &req, false); err != nil {
			writeJSONBodyError(w, err)
			return
		}
		party, err := h.partiesAdmin.CreateParty(r.Context(), parties.CreatePartyInput{
			PartyCode:   strings.TrimSpace(req.PartyCode),
			DisplayName: strings.TrimSpace(req.DisplayName),
			LegalName:   strings.TrimSpace(req.LegalName),
			PartyKind:   strings.TrimSpace(req.PartyKind),
			Actor:       actor,
		})
		if err != nil {
			handlePartyAdminError(w, err, "failed to create party")
			return
		}
		writeJSON(w, http.StatusCreated, mapParty(party))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func handleInventoryAdminError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, identityaccess.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	case errors.Is(err, inventoryops.ErrInvalidItem):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid inventory item"})
	case errors.Is(err, inventoryops.ErrInvalidLocation):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid inventory location"})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: fallback})
	}
}

func (h *AgentAPIHandler) handleAdminInventoryItems(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != adminInventoryItemsPath {
		http.NotFound(w, r)
		return
	}
	if h.inventoryAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "inventory admin service unavailable"})
		return
	}

	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.inventoryAdmin.ListItems(r.Context(), inventoryops.ListItemsInput{
			ItemRole: strings.TrimSpace(r.URL.Query().Get("item_role")),
			Actor:    actor,
		})
		if err != nil {
			handleInventoryAdminError(w, err, "failed to list inventory items")
			return
		}
		response := struct {
			Items []inventoryItemResponse `json:"items"`
		}{Items: make([]inventoryItemResponse, 0, len(items))}
		for _, item := range items {
			response.Items = append(response.Items, mapInventoryItem(item))
		}
		writeJSON(w, http.StatusOK, response)
	case http.MethodPost:
		var req createInventoryItemRequest
		if err := decodeJSONBody(r, &req, false); err != nil {
			writeJSONBodyError(w, err)
			return
		}
		item, err := h.inventoryAdmin.CreateItem(r.Context(), inventoryops.CreateItemInput{
			SKU:          strings.TrimSpace(req.SKU),
			Name:         strings.TrimSpace(req.Name),
			ItemRole:     strings.TrimSpace(req.ItemRole),
			TrackingMode: strings.TrimSpace(req.TrackingMode),
			Actor:        actor,
		})
		if err != nil {
			handleInventoryAdminError(w, err, "failed to create inventory item")
			return
		}
		writeJSON(w, http.StatusCreated, mapInventoryItem(item))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (h *AgentAPIHandler) handleAdminInventoryLocations(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != adminInventoryLocsPath {
		http.NotFound(w, r)
		return
	}
	if h.inventoryAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "inventory admin service unavailable"})
		return
	}

	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.inventoryAdmin.ListLocations(r.Context(), inventoryops.ListLocationsInput{
			LocationRole: strings.TrimSpace(r.URL.Query().Get("location_role")),
			Actor:        actor,
		})
		if err != nil {
			handleInventoryAdminError(w, err, "failed to list inventory locations")
			return
		}
		response := struct {
			Items []inventoryLocationResponse `json:"items"`
		}{Items: make([]inventoryLocationResponse, 0, len(items))}
		for _, item := range items {
			response.Items = append(response.Items, mapInventoryLocation(item))
		}
		writeJSON(w, http.StatusOK, response)
	case http.MethodPost:
		var req createInventoryLocationRequest
		if err := decodeJSONBody(r, &req, false); err != nil {
			writeJSONBodyError(w, err)
			return
		}
		location, err := h.inventoryAdmin.CreateLocation(r.Context(), inventoryops.CreateLocationInput{
			Code:         strings.TrimSpace(req.Code),
			Name:         strings.TrimSpace(req.Name),
			LocationRole: strings.TrimSpace(req.LocationRole),
			Actor:        actor,
		})
		if err != nil {
			handleInventoryAdminError(w, err, "failed to create inventory location")
			return
		}
		writeJSON(w, http.StatusCreated, mapInventoryLocation(location))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (h *AgentAPIHandler) handleAdminPartyDetail(w http.ResponseWriter, r *http.Request) {
	if h.partiesAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "party admin service unavailable"})
		return
	}
	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		partyID, ok := parseChildPath(adminPartiesPath, r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		party, err := h.partiesAdmin.GetParty(r.Context(), parties.GetPartyInput{
			PartyID: partyID,
			Actor:   actor,
		})
		if err != nil {
			handlePartyAdminError(w, err, "failed to load party")
			return
		}

		contacts, err := h.partiesAdmin.ListContacts(r.Context(), parties.ListContactsInput{
			PartyID: partyID,
			Actor:   actor,
		})
		if err != nil {
			handlePartyAdminError(w, err, "failed to list party contacts")
			return
		}

		response := struct {
			Party    partyResponse     `json:"party"`
			Contacts []contactResponse `json:"contacts"`
		}{
			Party:    mapParty(party),
			Contacts: make([]contactResponse, 0, len(contacts)),
		}
		for _, contact := range contacts {
			response.Contacts = append(response.Contacts, mapContact(contact))
		}
		writeJSON(w, http.StatusOK, response)
	case http.MethodPost:
		partyID, action, ok := parseChildActionPath(adminPartyContactsPath, r.URL.Path)
		if !ok || !strings.EqualFold(strings.TrimSpace(action), "contacts") {
			http.NotFound(w, r)
			return
		}

		var req createContactRequest
		if err := decodeJSONBody(r, &req, false); err != nil {
			writeJSONBodyError(w, err)
			return
		}

		contact, err := h.partiesAdmin.CreateContact(r.Context(), parties.CreateContactInput{
			PartyID:   partyID,
			FullName:  strings.TrimSpace(req.FullName),
			RoleTitle: strings.TrimSpace(req.RoleTitle),
			Email:     strings.TrimSpace(req.Email),
			Phone:     strings.TrimSpace(req.Phone),
			IsPrimary: req.IsPrimary,
			Actor:     actor,
		})
		if err != nil {
			handlePartyAdminError(w, err, "failed to create party contact")
			return
		}
		writeJSON(w, http.StatusCreated, mapContact(contact))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}
