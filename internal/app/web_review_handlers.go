package app

import (
	"net/http"
	"net/url"
	"strings"

	"workflow_app/internal/reporting"
)

func (h *AgentAPIHandler) handleWebDocuments(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webDocumentsPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	data := webDocumentsData{
		Session:    sessionContext,
		Notice:     strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:      strings.TrimSpace(r.URL.Query().Get("error")),
		DocumentID: strings.TrimSpace(r.URL.Query().Get("document_id")),
		TypeCode:   strings.TrimSpace(r.URL.Query().Get("type_code")),
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
	}
	data.Documents, err = h.reviewService.ListDocuments(r.Context(), reporting.ListDocumentsInput{
		DocumentID: data.DocumentID,
		TypeCode:   data.TypeCode,
		Status:     data.Status,
		Limit:      50,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load documents"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webDocumentsPath,
		Notice:     data.Notice,
		Error:      data.Error,
		Session:    &sessionContext,
		Documents:  &data,
	})
}

func (h *AgentAPIHandler) handleWebDocumentDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	documentID, ok := parseChildPath(webDocumentsPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	review, err := h.reviewService.GetDocumentReview(r.Context(), reporting.GetDocumentReviewInput{
		DocumentID: documentID,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, webDocumentsPath+"?error="+url.QueryEscape("failed to load document"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webDocumentsPath,
		Session:    &sessionContext,
		DocumentDetail: &webDocumentDetailData{
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			Review:  review,
		},
	})
}

func (h *AgentAPIHandler) handleWebAccounting(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAccountingPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	startOn := parseOptionalDate(r.URL.Query().Get("start_on"))
	endOn := parseOptionalDate(r.URL.Query().Get("end_on"))
	asOf := parseOptionalDate(r.URL.Query().Get("as_of"))
	data := webAccountingData{
		Session:     sessionContext,
		Notice:      strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:       strings.TrimSpace(r.URL.Query().Get("error")),
		StartOn:     formatDateInput(startOn),
		EndOn:       formatDateInput(endOn),
		AsOf:        formatDateInput(asOf),
		EntryID:     strings.TrimSpace(r.URL.Query().Get("entry_id")),
		DocumentID:  strings.TrimSpace(r.URL.Query().Get("document_id")),
		TaxType:     strings.TrimSpace(r.URL.Query().Get("tax_type")),
		TaxCode:     strings.TrimSpace(r.URL.Query().Get("tax_code")),
		ControlType: strings.TrimSpace(r.URL.Query().Get("control_type")),
		AccountID:   strings.TrimSpace(r.URL.Query().Get("account_id")),
	}

	data.JournalEntries, err = h.reviewService.ListJournalEntries(r.Context(), reporting.ListJournalEntriesInput{
		StartOn:    startOn,
		EndOn:      endOn,
		EntryID:    data.EntryID,
		DocumentID: data.DocumentID,
		Limit:      50,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load journal entries"
	}
	if data.ControlBalances, err = h.reviewService.ListControlAccountBalances(r.Context(), reporting.ListControlAccountBalancesInput{
		AsOf:        asOf,
		AccountID:   data.AccountID,
		ControlType: data.ControlType,
		Actor:       sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load control account balances"
	}
	if data.TaxSummaries, err = h.reviewService.ListTaxSummaries(r.Context(), reporting.ListTaxSummariesInput{
		StartOn: startOn,
		EndOn:   endOn,
		TaxType: data.TaxType,
		TaxCode: data.TaxCode,
		Limit:   50,
		Actor:   sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load tax summaries"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webAccountingPath,
		Notice:     data.Notice,
		Error:      data.Error,
		Session:    &sessionContext,
		Accounting: &data,
	})
}

func (h *AgentAPIHandler) handleWebControlAccountDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	accountID, ok := parseChildPath(webAccountingControlsPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	startOn := parseOptionalDate(r.URL.Query().Get("start_on"))
	endOn := parseOptionalDate(r.URL.Query().Get("end_on"))
	asOf := parseOptionalDate(r.URL.Query().Get("as_of"))

	balances, err := h.reviewService.ListControlAccountBalances(r.Context(), reporting.ListControlAccountBalancesInput{
		AsOf:      asOf,
		AccountID: accountID,
		Actor:     sessionContext.Actor,
	})
	if err != nil || len(balances) == 0 {
		http.Redirect(w, r, webAccountingPath+"?error="+url.QueryEscape("failed to load control account"), http.StatusSeeOther)
		return
	}

	summaries, err := h.reviewService.ListTaxSummaries(r.Context(), reporting.ListTaxSummariesInput{
		StartOn: startOn,
		EndOn:   endOn,
		Limit:   50,
		Actor:   sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, webAccountingPath+"?error="+url.QueryEscape("failed to load related tax summaries"), http.StatusSeeOther)
		return
	}

	related := make([]reporting.TaxSummary, 0, len(summaries))
	for _, summary := range summaries {
		if (summary.ReceivableAccountID.Valid && summary.ReceivableAccountID.String == accountID) ||
			(summary.PayableAccountID.Valid && summary.PayableAccountID.String == accountID) {
			related = append(related, summary)
		}
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webAccountingPath,
		Session:    &sessionContext,
		ControlAccountDetail: &webControlAccountDetailData{
			Session:          sessionContext,
			Notice:           strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:            strings.TrimSpace(r.URL.Query().Get("error")),
			StartOn:          formatDateInput(startOn),
			EndOn:            formatDateInput(endOn),
			AsOf:             formatDateInput(asOf),
			Balance:          balances[0],
			RelatedSummaries: related,
		},
	})
}

func (h *AgentAPIHandler) handleWebTaxSummaryDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	taxCode, ok := parseChildPath(webAccountingTaxesPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	startOn := parseOptionalDate(r.URL.Query().Get("start_on"))
	endOn := parseOptionalDate(r.URL.Query().Get("end_on"))

	summaries, err := h.reviewService.ListTaxSummaries(r.Context(), reporting.ListTaxSummariesInput{
		StartOn: startOn,
		EndOn:   endOn,
		TaxType: strings.TrimSpace(r.URL.Query().Get("tax_type")),
		TaxCode: taxCode,
		Limit:   2,
		Actor:   sessionContext.Actor,
	})
	if err != nil || len(summaries) == 0 {
		http.Redirect(w, r, webAccountingPath+"?error="+url.QueryEscape("failed to load tax summary"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webAccountingPath,
		Session:    &sessionContext,
		TaxSummaryDetail: &webTaxSummaryDetailData{
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			StartOn: formatDateInput(startOn),
			EndOn:   formatDateInput(endOn),
			Summary: summaries[0],
		},
	})
}

func (h *AgentAPIHandler) handleWebAccountingDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	entryID, ok := parseChildPath(webAccountingPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	entries, err := h.reviewService.ListJournalEntries(r.Context(), reporting.ListJournalEntriesInput{
		EntryID: entryID,
		Limit:   2,
		Actor:   sessionContext.Actor,
	})
	if err != nil || len(entries) == 0 {
		http.Redirect(w, r, webAccountingPath+"?error="+url.QueryEscape("failed to load journal entry"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webAccountingPath,
		Session:    &sessionContext,
		AccountingDetail: &webAccountingDetailData{
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			Review:  entries[0],
		},
	})
}

func (h *AgentAPIHandler) handleWebProposals(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webProposalsPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	data := webProposalsData{
		Session:          sessionContext,
		Notice:           strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:            strings.TrimSpace(r.URL.Query().Get("error")),
		RecommendationID: strings.TrimSpace(r.URL.Query().Get("recommendation_id")),
		Status:           strings.TrimSpace(r.URL.Query().Get("status")),
		RequestReference: strings.TrimSpace(r.URL.Query().Get("request_reference")),
	}

	data.StatusSummary, err = h.reviewService.ListProcessedProposalStatusSummary(r.Context(), sessionContext.Actor)
	if err != nil {
		data.Error = "failed to load proposal summary"
	}
	if data.ProcessedProposals, err = h.reviewService.ListProcessedProposals(r.Context(), reporting.ListProcessedProposalsInput{
		RecommendationID: data.RecommendationID,
		Status:           data.Status,
		RequestReference: data.RequestReference,
		Limit:            50,
		Actor:            sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load processed proposals"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webProposalsPath,
		Notice:     data.Notice,
		Error:      data.Error,
		Session:    &sessionContext,
		Proposals:  &data,
	})
}

func (h *AgentAPIHandler) handleWebProposalDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		recommendationID, action, ok := parseChildActionPath(webProposalsPath, r.URL.Path)
		if !ok || action != "request-approval" {
			http.NotFound(w, r)
			return
		}
		h.handleWebProposalAction(w, r, recommendationID, action)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	recommendationID, ok := parseChildPath(webProposalsPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	proposals, err := h.reviewService.ListProcessedProposals(r.Context(), reporting.ListProcessedProposalsInput{
		RecommendationID: recommendationID,
		Limit:            2,
		Actor:            sessionContext.Actor,
	})
	if err != nil || len(proposals) == 0 {
		http.Redirect(w, r, webProposalsPath+"?error="+url.QueryEscape("failed to load proposal"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webProposalsPath,
		Session:    &sessionContext,
		ProposalDetail: &webProposalDetailData{
			Session:                sessionContext,
			Notice:                 strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:                  strings.TrimSpace(r.URL.Query().Get("error")),
			Review:                 proposals[0],
			ApprovalQueueCodeDraft: proposalQueueCode(proposals[0]),
		},
	})
}

func (h *AgentAPIHandler) handleWebProposalAction(w http.ResponseWriter, r *http.Request, recommendationID, action string) {
	if h.proposalApproval == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("proposal approval service unavailable"), http.StatusSeeOther)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("unauthorized"), http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, webProposalsPath+"?error="+url.QueryEscape("invalid proposal form"), http.StatusSeeOther)
		return
	}

	returnTo := sanitizeWebReturnPath(r.FormValue("return_to"))
	if returnTo == "" {
		returnTo = webProposalDetailPrefix + url.PathEscape(recommendationID)
	}

	switch action {
	case "request-approval":
		if _, _, err := h.proposalApproval.RequestProcessedProposalApproval(r.Context(), requestProcessedProposalApprovalInput{
			RecommendationID: recommendationID,
			QueueCode:        strings.TrimSpace(r.FormValue("queue_code")),
			Reason:           strings.TrimSpace(r.FormValue("reason")),
			Actor:            actor,
		}); err != nil {
			http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to request approval"), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, appendWebMessage(returnTo, "notice", "Approval requested."), http.StatusSeeOther)
		return
	default:
		http.NotFound(w, r)
		return
	}
}
