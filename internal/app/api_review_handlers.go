package app

import (
	"net/http"
	"strings"

	"workflow_app/internal/reporting"
)

func (h *AgentAPIHandler) handleListInboundRequests(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewInboundRequestsPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListInboundRequests(r.Context(), reporting.ListInboundRequestsInput{
		Status:           strings.TrimSpace(r.URL.Query().Get("status")),
		RequestReference: strings.TrimSpace(r.URL.Query().Get("request_reference")),
		Limit:            parseLimit(r.URL.Query().Get("limit")),
		Actor:            actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list inbound requests")
		return
	}

	response := struct {
		Items []inboundRequestReviewResponse `json:"items"`
	}{Items: make([]inboundRequestReviewResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapInboundRequestReview(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetInboundRequestDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	lookup, ok := parseChildPath(reviewInboundRequestsPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	input := reporting.GetInboundRequestDetailInput{Actor: actor}
	populateInboundRequestDetailLookup(&input, lookup)

	detail, err := h.reviewService.GetInboundRequestDetail(r.Context(), input)
	if err != nil {
		handleReviewError(w, err, "failed to load inbound request detail")
		return
	}

	writeJSON(w, http.StatusOK, mapInboundRequestDetail(detail))
}

func (h *AgentAPIHandler) handleListInboundRequestStatusSummary(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewInboundSummaryPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListInboundRequestStatusSummary(r.Context(), actor)
	if err != nil {
		handleReviewError(w, err, "failed to load inbound request status summary")
		return
	}

	response := struct {
		Items []inboundRequestStatusSummaryResponse `json:"items"`
	}{Items: make([]inboundRequestStatusSummaryResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, inboundRequestStatusSummaryResponse{
			Status:           item.Status,
			RequestCount:     item.RequestCount,
			MessageCount:     item.MessageCount,
			AttachmentCount:  item.AttachmentCount,
			LatestReceivedAt: timePtr(item.LatestReceivedAt),
			LatestQueuedAt:   timePtr(item.LatestQueuedAt),
			LatestUpdatedAt:  item.LatestUpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleListProcessedProposals(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewProposalListPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListProcessedProposals(r.Context(), reporting.ListProcessedProposalsInput{
		RecommendationID: strings.TrimSpace(r.URL.Query().Get("recommendation_id")),
		Status:           strings.TrimSpace(r.URL.Query().Get("status")),
		RequestID:        strings.TrimSpace(r.URL.Query().Get("request_id")),
		RequestReference: strings.TrimSpace(r.URL.Query().Get("request_reference")),
		Limit:            parseLimit(r.URL.Query().Get("limit")),
		Actor:            actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list processed proposals")
		return
	}

	response := struct {
		Items []processedProposalReviewResponse `json:"items"`
	}{Items: make([]processedProposalReviewResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapProcessedProposalReview(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetProcessedProposalDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	recommendationID, ok := parseChildPath(reviewProposalListPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListProcessedProposals(r.Context(), reporting.ListProcessedProposalsInput{
		RecommendationID: recommendationID,
		Limit:            2,
		Actor:            actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load processed proposal")
		return
	}
	if len(items) == 0 {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "processed proposal not found"})
		return
	}

	writeJSON(w, http.StatusOK, mapProcessedProposalReview(items[0]))
}

func (h *AgentAPIHandler) handleListProcessedProposalStatusSummary(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewProposalSummaryPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListProcessedProposalStatusSummary(r.Context(), actor)
	if err != nil {
		handleReviewError(w, err, "failed to load processed proposal status summary")
		return
	}

	response := struct {
		Items []processedProposalStatusSummaryResponse `json:"items"`
	}{Items: make([]processedProposalStatusSummaryResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, processedProposalStatusSummaryResponse{
			RecommendationStatus: item.RecommendationStatus,
			ProposalCount:        item.ProposalCount,
			RequestCount:         item.RequestCount,
			DocumentCount:        item.DocumentCount,
			LatestCreatedAt:      item.LatestCreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleListApprovalQueue(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewApprovalQueuePath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListApprovalQueue(r.Context(), reporting.ListApprovalQueueInput{
		ApprovalID: strings.TrimSpace(r.URL.Query().Get("approval_id")),
		QueueCode:  strings.TrimSpace(r.URL.Query().Get("queue_code")),
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		Limit:      parseLimit(r.URL.Query().Get("limit")),
		Actor:      actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list approval queue")
		return
	}

	response := struct {
		Items []approvalQueueEntryResponse `json:"items"`
	}{Items: make([]approvalQueueEntryResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapApprovalQueueEntry(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetApprovalQueueDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	approvalID, ok := parseChildPath(reviewApprovalQueuePath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListApprovalQueue(r.Context(), reporting.ListApprovalQueueInput{
		ApprovalID: approvalID,
		Limit:      2,
		Actor:      actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load approval")
		return
	}
	if len(items) == 0 {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "approval not found"})
		return
	}

	writeJSON(w, http.StatusOK, mapApprovalQueueEntry(items[0]))
}

func (h *AgentAPIHandler) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewDocumentsPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListDocuments(r.Context(), reporting.ListDocumentsInput{
		DocumentID: strings.TrimSpace(r.URL.Query().Get("document_id")),
		TypeCode:   strings.TrimSpace(r.URL.Query().Get("type_code")),
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		Limit:      parseLimit(r.URL.Query().Get("limit")),
		Actor:      actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list documents")
		return
	}

	response := struct {
		Items []documentReviewResponse `json:"items"`
	}{Items: make([]documentReviewResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapDocumentReview(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetDocumentReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	documentID, ok := parseChildPath(reviewDocumentsPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	item, err := h.reviewService.GetDocumentReview(r.Context(), reporting.GetDocumentReviewInput{
		DocumentID: documentID,
		Actor:      actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load document review")
		return
	}

	writeJSON(w, http.StatusOK, mapDocumentReview(item))
}

func (h *AgentAPIHandler) handleListJournalEntries(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewJournalEntriesPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListJournalEntries(r.Context(), reporting.ListJournalEntriesInput{
		StartOn:    parseOptionalDate(r.URL.Query().Get("start_on")),
		EndOn:      parseOptionalDate(r.URL.Query().Get("end_on")),
		EntryID:    strings.TrimSpace(r.URL.Query().Get("entry_id")),
		DocumentID: strings.TrimSpace(r.URL.Query().Get("document_id")),
		Limit:      parseLimit(r.URL.Query().Get("limit")),
		Actor:      actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list journal entries")
		return
	}

	response := struct {
		Items []journalEntryReviewResponse `json:"items"`
	}{Items: make([]journalEntryReviewResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapJournalEntryReview(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetJournalEntryDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	entryID, ok := parseChildPath(reviewJournalEntriesPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListJournalEntries(r.Context(), reporting.ListJournalEntriesInput{
		EntryID: entryID,
		Limit:   2,
		Actor:   actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load journal entry")
		return
	}
	if len(items) == 0 {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "journal entry not found"})
		return
	}

	writeJSON(w, http.StatusOK, mapJournalEntryReview(items[0]))
}

func (h *AgentAPIHandler) handleListControlAccountBalances(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewControlBalancesPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListControlAccountBalances(r.Context(), reporting.ListControlAccountBalancesInput{
		AsOf:        parseOptionalDate(r.URL.Query().Get("as_of")),
		AccountID:   strings.TrimSpace(r.URL.Query().Get("account_id")),
		ControlType: strings.TrimSpace(r.URL.Query().Get("control_type")),
		Actor:       actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list control account balances")
		return
	}

	response := struct {
		Items []controlAccountBalanceResponse `json:"items"`
	}{Items: make([]controlAccountBalanceResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapControlAccountBalance(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleListTaxSummaries(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewTaxSummariesPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListTaxSummaries(r.Context(), reporting.ListTaxSummariesInput{
		StartOn: parseOptionalDate(r.URL.Query().Get("start_on")),
		EndOn:   parseOptionalDate(r.URL.Query().Get("end_on")),
		TaxType: strings.TrimSpace(r.URL.Query().Get("tax_type")),
		TaxCode: strings.TrimSpace(r.URL.Query().Get("tax_code")),
		Limit:   parseLimit(r.URL.Query().Get("limit")),
		Actor:   actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list tax summaries")
		return
	}

	response := struct {
		Items []taxSummaryResponse `json:"items"`
	}{Items: make([]taxSummaryResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapTaxSummary(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetTrialBalance(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewTrialBalancePath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	report, err := h.reviewService.GetTrialBalance(r.Context(), reporting.GetTrialBalanceInput{
		AsOf:  parseOptionalDate(r.URL.Query().Get("as_of")),
		Actor: actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load trial balance")
		return
	}
	writeJSON(w, http.StatusOK, mapTrialBalanceReport(report))
}

func (h *AgentAPIHandler) handleGetBalanceSheet(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewBalanceSheetPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	report, err := h.reviewService.GetBalanceSheet(r.Context(), reporting.GetBalanceSheetInput{
		AsOf:  parseOptionalDate(r.URL.Query().Get("as_of")),
		Actor: actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load balance sheet")
		return
	}
	writeJSON(w, http.StatusOK, mapBalanceSheetReport(report))
}

func (h *AgentAPIHandler) handleGetIncomeStatement(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewIncomeStatementPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	report, err := h.reviewService.GetIncomeStatement(r.Context(), reporting.GetIncomeStatementInput{
		StartOn: parseOptionalDate(r.URL.Query().Get("start_on")),
		EndOn:   parseOptionalDate(r.URL.Query().Get("end_on")),
		Actor:   actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load income statement")
		return
	}
	writeJSON(w, http.StatusOK, mapIncomeStatementReport(report))
}

func (h *AgentAPIHandler) handleListInventoryStock(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewInventoryStockPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListInventoryStock(r.Context(), reporting.ListInventoryStockInput{
		ItemID:      strings.TrimSpace(r.URL.Query().Get("item_id")),
		LocationID:  strings.TrimSpace(r.URL.Query().Get("location_id")),
		IncludeZero: strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_zero")), "true"),
		Limit:       parseLimit(r.URL.Query().Get("limit")),
		Actor:       actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list inventory stock")
		return
	}

	response := struct {
		Items []inventoryStockResponse `json:"items"`
	}{Items: make([]inventoryStockResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapInventoryStock(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleListInventoryMovements(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewInventoryMovesPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListInventoryMovements(r.Context(), reporting.ListInventoryMovementsInput{
		MovementID:   strings.TrimSpace(r.URL.Query().Get("movement_id")),
		ItemID:       strings.TrimSpace(r.URL.Query().Get("item_id")),
		LocationID:   strings.TrimSpace(r.URL.Query().Get("location_id")),
		DocumentID:   strings.TrimSpace(r.URL.Query().Get("document_id")),
		MovementType: strings.TrimSpace(r.URL.Query().Get("movement_type")),
		Limit:        parseLimit(r.URL.Query().Get("limit")),
		Actor:        actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list inventory movements")
		return
	}

	response := struct {
		Items []inventoryMovementResponse `json:"items"`
	}{Items: make([]inventoryMovementResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapInventoryMovement(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetInventoryMovementDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	movementID, ok := parseChildPath(reviewInventoryMovesPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListInventoryMovements(r.Context(), reporting.ListInventoryMovementsInput{
		MovementID: movementID,
		Limit:      2,
		Actor:      actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load inventory movement")
		return
	}
	if len(items) == 0 {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "inventory movement not found"})
		return
	}

	reconciliation, err := h.reviewService.ListInventoryReconciliation(r.Context(), reporting.ListInventoryReconciliationInput{
		MovementID: movementID,
		Limit:      50,
		Actor:      actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load inventory reconciliation")
		return
	}

	response := inventoryMovementDetailResponse{
		Review:         mapInventoryMovement(items[0]),
		Reconciliation: make([]inventoryReconciliationResponse, 0, len(reconciliation)),
	}
	for _, item := range reconciliation {
		response.Reconciliation = append(response.Reconciliation, mapInventoryReconciliation(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleListInventoryReconciliation(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewInventoryReconPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListInventoryReconciliation(r.Context(), reporting.ListInventoryReconciliationInput{
		MovementID:            strings.TrimSpace(r.URL.Query().Get("movement_id")),
		ItemID:                strings.TrimSpace(r.URL.Query().Get("item_id")),
		DocumentID:            strings.TrimSpace(r.URL.Query().Get("document_id")),
		OnlyPendingAccounting: strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("only_pending_accounting")), "true"),
		OnlyPendingExecution:  strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("only_pending_execution")), "true"),
		Limit:                 parseLimit(r.URL.Query().Get("limit")),
		Actor:                 actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list inventory reconciliation")
		return
	}

	response := struct {
		Items []inventoryReconciliationResponse `json:"items"`
	}{Items: make([]inventoryReconciliationResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapInventoryReconciliation(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleListWorkOrders(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewWorkOrdersPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.ListWorkOrders(r.Context(), reporting.ListWorkOrdersInput{
		WorkOrderID: strings.TrimSpace(r.URL.Query().Get("work_order_id")),
		Status:      strings.TrimSpace(r.URL.Query().Get("status")),
		DocumentID:  strings.TrimSpace(r.URL.Query().Get("document_id")),
		Limit:       parseLimit(r.URL.Query().Get("limit")),
		Actor:       actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to list work orders")
		return
	}

	response := struct {
		Items []workOrderReviewResponse `json:"items"`
	}{Items: make([]workOrderReviewResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapWorkOrderReview(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetWorkOrderReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	workOrderID, ok := parseChildPath(reviewWorkOrdersPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	item, err := h.reviewService.GetWorkOrderReview(r.Context(), reporting.GetWorkOrderReviewInput{
		WorkOrderID: workOrderID,
		Actor:       actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load work order review")
		return
	}

	writeJSON(w, http.StatusOK, mapWorkOrderReview(item))
}

func (h *AgentAPIHandler) handleLookupAuditEvents(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != reviewAuditEventsPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.LookupAuditEvents(r.Context(), reporting.LookupAuditEventsInput{
		EventID:    strings.TrimSpace(r.URL.Query().Get("event_id")),
		EntityType: strings.TrimSpace(r.URL.Query().Get("entity_type")),
		EntityID:   strings.TrimSpace(r.URL.Query().Get("entity_id")),
		Limit:      parseLimit(r.URL.Query().Get("limit")),
		Actor:      actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to look up audit events")
		return
	}

	response := struct {
		Items []auditEventResponse `json:"items"`
	}{Items: make([]auditEventResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapAuditEvent(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetAuditEventDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if h.reviewService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "review service unavailable"})
		return
	}

	eventID, ok := parseChildPath(reviewAuditEventsPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	items, err := h.reviewService.LookupAuditEvents(r.Context(), reporting.LookupAuditEventsInput{
		EventID: eventID,
		Limit:   1,
		Actor:   actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to load audit event")
		return
	}
	if len(items) == 0 {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "audit event not found"})
		return
	}

	writeJSON(w, http.StatusOK, mapAuditEvent(items[0]))
}
