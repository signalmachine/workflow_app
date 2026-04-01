package app

import (
	"net/http"
	"net/url"
	"sort"
	"strings"

	"workflow_app/internal/reporting"
)

func (h *AgentAPIHandler) handleWebOperationsLanding(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webOperationsPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webLoginPath+"?notice="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	data := webOperationsLandingData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
	}

	if snapshot, snapshotErr := h.reviewService.GetWorkflowNavigationSnapshot(r.Context(), sessionContext.Actor, 10); snapshotErr != nil {
		data.Error = "failed to load operations landing"
	} else {
		sortInboundRequestStatusSummaries(snapshot.InboundSummary)
		data.QueuedRequestCount = countQueuedRequests(snapshot.InboundSummary)
		data.PendingApprovalCount = len(snapshot.PendingApprovals)

		requests, requestErr := h.reviewService.ListInboundRequests(r.Context(), reporting.ListInboundRequestsInput{
			Limit: 10,
			Actor: sessionContext.Actor,
		})
		if requestErr != nil {
			data.Error = "failed to load operations landing"
		} else {
			data.RecentFeed = append(data.RecentFeed, buildOperationsFeedFromRequests(requests)...)
		}

		proposals, proposalErr := h.reviewService.ListProcessedProposals(r.Context(), reporting.ListProcessedProposalsInput{
			Limit: 10,
			Actor: sessionContext.Actor,
		})
		if proposalErr != nil {
			data.Error = "failed to load operations landing"
		} else {
			data.RecentFeed = append(data.RecentFeed, buildOperationsFeedFromProposals(proposals)...)
			data.ProposalReviewCount = len(proposals)
		}

		data.RecentFeed = append(data.RecentFeed, buildOperationsFeedFromApprovals(snapshot.PendingApprovals)...)
	}

	sort.SliceStable(data.RecentFeed, func(i, j int) bool {
		if !data.RecentFeed[i].OccurredAt.Equal(data.RecentFeed[j].OccurredAt) {
			return data.RecentFeed[i].OccurredAt.After(data.RecentFeed[j].OccurredAt)
		}
		return data.RecentFeed[i].Title < data.RecentFeed[j].Title
	})
	if len(data.RecentFeed) > 8 {
		data.RecentFeed = data.RecentFeed[:8]
	}

	h.renderWebPage(w, webPageData{
		Title:             "workflow_app",
		ActivePath:        webOperationsPath,
		Notice:            data.Notice,
		Error:             data.Error,
		Session:           &sessionContext,
		OperationsLanding: &data,
	})
}

func (h *AgentAPIHandler) handleWebReviewLanding(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webReviewPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webLoginPath+"?notice="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	data := webReviewLandingData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
	}

	if snapshot, snapshotErr := h.reviewService.GetWorkflowNavigationSnapshot(r.Context(), sessionContext.Actor, 8); snapshotErr != nil {
		data.Error = "failed to load review landing"
	} else {
		sortInboundRequestStatusSummaries(snapshot.InboundSummary)
		data.InboundSummary = snapshot.InboundSummary
		data.InboundRequestCount = sumInboundRequestCount(snapshot.InboundSummary)
		data.ProposalSummary = snapshot.ProposalSummary
		data.ProposalCount = sumProposalCount(snapshot.ProposalSummary)
		data.PendingApprovals = snapshot.PendingApprovals
	}

	h.renderWebPage(w, webPageData{
		Title:         "workflow_app",
		ActivePath:    webReviewPath,
		Notice:        data.Notice,
		Error:         data.Error,
		Session:       &sessionContext,
		ReviewLanding: &data,
	})
}

func (h *AgentAPIHandler) handleWebInventoryLanding(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webInventoryHubPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webLoginPath+"?notice="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	data := webInventoryLandingData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
	}

	if stock, stockErr := h.reviewService.ListInventoryStock(r.Context(), reporting.ListInventoryStockInput{
		Limit: 8,
		Actor: sessionContext.Actor,
	}); stockErr != nil {
		data.Error = "failed to load inventory landing"
	} else {
		data.Stock = stock
	}

	if moves, moveErr := h.reviewService.ListInventoryMovements(r.Context(), reporting.ListInventoryMovementsInput{
		Limit: 8,
		Actor: sessionContext.Actor,
	}); moveErr != nil {
		if data.Error == "" {
			data.Error = "failed to load inventory landing"
		}
	} else {
		data.Movements = moves
	}

	if reconciliation, reconErr := h.reviewService.ListInventoryReconciliation(r.Context(), reporting.ListInventoryReconciliationInput{
		Limit: 8,
		Actor: sessionContext.Actor,
	}); reconErr != nil {
		if data.Error == "" {
			data.Error = "failed to load inventory landing"
		}
	} else {
		data.Reconciliation = reconciliation
		data.PendingExecutionCount = countPendingReconciliation(reconciliation, "execution")
		data.PendingAccountingCount = countPendingReconciliation(reconciliation, "accounting")
	}

	h.renderWebPage(w, webPageData{
		Title:            "workflow_app",
		ActivePath:       webInventoryHubPath,
		Notice:           data.Notice,
		Error:            data.Error,
		Session:          &sessionContext,
		InventoryLanding: &data,
	})
}
