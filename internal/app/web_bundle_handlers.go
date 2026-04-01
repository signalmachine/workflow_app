package app

import (
	"net/http"
	"net/url"
	"sort"
	"strings"
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

	if snapshot, snapshotErr := h.reviewService.GetOperationsLandingSnapshot(r.Context(), sessionContext.Actor, 10, 10); snapshotErr != nil {
		data.Error = "failed to load operations landing"
	} else {
		sortInboundRequestStatusSummaries(snapshot.Navigation.InboundSummary)
		data.QueuedRequestCount = countQueuedRequests(snapshot.Navigation.InboundSummary)
		data.PendingApprovalCount = len(snapshot.Navigation.PendingApprovals)
		data.ProposalReviewCount = len(snapshot.Feed.Proposals)
		data.RecentFeed = append(data.RecentFeed, buildOperationsFeedFromRequests(snapshot.Feed.Requests)...)
		data.RecentFeed = append(data.RecentFeed, buildOperationsFeedFromProposals(snapshot.Feed.Proposals)...)
		data.RecentFeed = append(data.RecentFeed, buildOperationsFeedFromApprovals(snapshot.Navigation.PendingApprovals)...)
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

	if snapshot, snapshotErr := h.reviewService.GetInventoryLandingSnapshot(r.Context(), sessionContext.Actor, 8); snapshotErr != nil {
		data.Error = "failed to load inventory landing"
	} else {
		data.Stock = snapshot.Stock
		data.Movements = snapshot.Movements
		data.Reconciliation = snapshot.Reconciliation
		data.PendingExecutionCount = countPendingReconciliation(snapshot.Reconciliation, "execution")
		data.PendingAccountingCount = countPendingReconciliation(snapshot.Reconciliation, "accounting")
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
