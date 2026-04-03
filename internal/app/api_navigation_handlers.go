package app

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

type navigationHomeActionResponse struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Href    string `json:"href"`
	Badge   string `json:"badge,omitempty"`
}

type navigationRouteCatalogEntryResponse struct {
	Title    string `json:"title"`
	Href     string `json:"href"`
	Category string `json:"category"`
	Summary  string `json:"summary"`
}

type navigationOperationsFeedItemResponse struct {
	OccurredAt     time.Time `json:"occurred_at"`
	Kind           string    `json:"kind"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary"`
	Status         string    `json:"status"`
	PrimaryLabel   string    `json:"primary_label"`
	PrimaryHref    string    `json:"primary_href"`
	SecondaryLabel string    `json:"secondary_label,omitempty"`
	SecondaryHref  string    `json:"secondary_href,omitempty"`
}

type navigationDashboardResponse struct {
	RoleHeadline     string                                   `json:"role_headline"`
	RoleBody         string                                   `json:"role_body"`
	PrimaryActions   []navigationHomeActionResponse           `json:"primary_actions"`
	SecondaryActions []navigationHomeActionResponse           `json:"secondary_actions"`
	InboundSummary   []inboundRequestStatusSummaryResponse    `json:"inbound_summary"`
	ProposalSummary  []processedProposalStatusSummaryResponse `json:"proposal_summary"`
	InboundRequests  []inboundRequestReviewResponse           `json:"inbound_requests"`
	Proposals        []processedProposalReviewResponse        `json:"proposals"`
	Approvals        []approvalQueueEntryResponse             `json:"approvals"`
}

type navigationOperationsResponse struct {
	QueuedRequestCount   int                                    `json:"queued_request_count"`
	PendingApprovalCount int                                    `json:"pending_approval_count"`
	ProposalReviewCount  int                                    `json:"proposal_review_count"`
	RecentFeed           []navigationOperationsFeedItemResponse `json:"recent_feed"`
}

type navigationOperationsFeedResponse struct {
	Items []navigationOperationsFeedItemResponse `json:"items"`
}

type navigationReviewResponse struct {
	InboundSummary      []inboundRequestStatusSummaryResponse    `json:"inbound_summary"`
	ProposalSummary     []processedProposalStatusSummaryResponse `json:"proposal_summary"`
	PendingApprovals    []approvalQueueEntryResponse             `json:"pending_approvals"`
	InboundRequestCount int                                      `json:"inbound_request_count"`
	ProposalCount       int                                      `json:"proposal_count"`
}

type navigationAgentChatResponse struct {
	RequestReference string                            `json:"request_reference,omitempty"`
	RequestStatus    string                            `json:"request_status,omitempty"`
	RecentRequests   []inboundRequestReviewResponse    `json:"recent_requests"`
	RecentProposals  []processedProposalReviewResponse `json:"recent_proposals"`
}

type navigationRouteCatalogResponse struct {
	Query string                                `json:"query"`
	Items []navigationRouteCatalogEntryResponse `json:"items"`
}

func mapNavigationHomeActions(items []webHomeAction) []navigationHomeActionResponse {
	response := make([]navigationHomeActionResponse, 0, len(items))
	for _, item := range items {
		response = append(response, navigationHomeActionResponse{
			Title:   item.Title,
			Summary: item.Summary,
			Href:    item.Href,
			Badge:   item.Badge,
		})
	}
	return response
}

func mapNavigationRouteCatalogEntries(items []webRouteCatalogEntry) []navigationRouteCatalogEntryResponse {
	response := make([]navigationRouteCatalogEntryResponse, 0, len(items))
	for _, item := range items {
		response = append(response, navigationRouteCatalogEntryResponse{
			Title:    item.Title,
			Href:     item.Href,
			Category: item.Category,
			Summary:  item.Summary,
		})
	}
	return response
}

func mapNavigationOperationsFeedItems(items []webOperationsFeedItem) []navigationOperationsFeedItemResponse {
	response := make([]navigationOperationsFeedItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, navigationOperationsFeedItemResponse{
			OccurredAt:     item.OccurredAt,
			Kind:           item.Kind,
			Title:          item.Title,
			Summary:        item.Summary,
			Status:         item.Status,
			PrimaryLabel:   item.PrimaryLabel,
			PrimaryHref:    item.PrimaryHref,
			SecondaryLabel: item.SecondaryLabel,
			SecondaryHref:  item.SecondaryHref,
		})
	}
	return response
}

func buildSvelteOperationsFeedFromRequests(items []inboundRequestReviewResponse) []navigationOperationsFeedItemResponse {
	feed := make([]navigationOperationsFeedItemResponse, 0, len(items))
	for _, item := range items {
		summary := item.RequestReference + " via " + item.Channel
		if item.FailureReason != "" {
			summary = item.FailureReason
		} else if item.CancellationReason != "" {
			summary = item.CancellationReason
		}

		secondaryLabel := ""
		secondaryHref := ""
		if item.LastRecommendationID != nil {
			secondaryLabel = "Open proposal"
			secondaryHref = webProposalsPath + "?request_reference=" + item.RequestReference
		} else if item.LastRunID != nil {
			secondaryLabel = "Filter request"
			secondaryHref = webInboundRequestsPath + "?request_reference=" + item.RequestReference
		}

		feed = append(feed, navigationOperationsFeedItemResponse{
			OccurredAt:     item.UpdatedAt,
			Kind:           "Request status",
			Title:          item.RequestReference + " moved through " + item.Status,
			Summary:        summary,
			Status:         item.Status,
			PrimaryLabel:   "Open request",
			PrimaryHref:    webInboundRequestsPath + "?request_reference=" + item.RequestReference,
			SecondaryLabel: secondaryLabel,
			SecondaryHref:  secondaryHref,
		})
	}
	return feed
}

func buildSvelteOperationsFeedFromProposals(items []processedProposalReviewResponse) []navigationOperationsFeedItemResponse {
	feed := make([]navigationOperationsFeedItemResponse, 0, len(items))
	for _, item := range items {
		secondaryLabel := "Open request"
		secondaryHref := webInboundRequestsPath + "?request_reference=" + item.RequestReference
		if item.DocumentID != nil {
			secondaryLabel = "Open document"
			secondaryHref = webDocumentsPath + "?document_id=" + *item.DocumentID
		} else if item.ApprovalID != nil {
			secondaryLabel = "Open approval"
			secondaryHref = webApprovalsPath + "?approval_id=" + *item.ApprovalID
		}

		feed = append(feed, navigationOperationsFeedItemResponse{
			OccurredAt:     item.CreatedAt,
			Kind:           "Coordinator proposal",
			Title:          item.RecommendationStatus + " proposal for " + item.RequestReference,
			Summary:        item.Summary,
			Status:         item.RecommendationStatus,
			PrimaryLabel:   "Open proposal",
			PrimaryHref:    webProposalsPath + "?request_reference=" + item.RequestReference,
			SecondaryLabel: secondaryLabel,
			SecondaryHref:  secondaryHref,
		})
	}
	return feed
}

func buildSvelteOperationsFeedFromApprovals(items []approvalQueueEntryResponse) []navigationOperationsFeedItemResponse {
	feed := make([]navigationOperationsFeedItemResponse, 0, len(items))
	for _, item := range items {
		occurredAt := item.RequestedAt
		if item.ClosedAt != nil {
			occurredAt = *item.ClosedAt
		}
		feed = append(feed, navigationOperationsFeedItemResponse{
			OccurredAt:     occurredAt,
			Kind:           "Approval queue",
			Title:          item.DocumentTitle + " approval is " + item.ApprovalStatus,
			Summary:        item.QueueCode + " on " + item.DocumentTitle,
			Status:         item.ApprovalStatus,
			PrimaryLabel:   "Open approval",
			PrimaryHref:    webApprovalsPath + "?approval_id=" + item.ApprovalID,
			SecondaryLabel: "Open document",
			SecondaryHref:  webDocumentsPath + "?document_id=" + item.DocumentID,
		})
	}
	return feed
}

func sortNavigationFeed(items []navigationOperationsFeedItemResponse, limit int) []navigationOperationsFeedItemResponse {
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].OccurredAt.Equal(items[j].OccurredAt) {
			return items[i].OccurredAt.After(items[j].OccurredAt)
		}
		return items[i].Title < items[j].Title
	})
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

func (h *AgentAPIHandler) handleGetNavigationDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != navigationDashboardPath {
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

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	snapshot, err := h.reviewService.GetDashboardSnapshot(r.Context(), sessionContext.Actor, 10, 20, 10)
	if err != nil {
		handleReviewError(w, err, "failed to load dashboard")
		return
	}

	sortInboundRequestStatusSummaries(snapshot.Navigation.InboundSummary)

	inboundSummary := make([]inboundRequestStatusSummaryResponse, 0, len(snapshot.Navigation.InboundSummary))
	for _, item := range snapshot.Navigation.InboundSummary {
		inboundSummary = append(inboundSummary, inboundRequestStatusSummaryResponse{
			Status:           item.Status,
			RequestCount:     item.RequestCount,
			MessageCount:     item.MessageCount,
			AttachmentCount:  item.AttachmentCount,
			LatestReceivedAt: timePtr(item.LatestReceivedAt),
			LatestQueuedAt:   timePtr(item.LatestQueuedAt),
			LatestUpdatedAt:  item.LatestUpdatedAt,
		})
	}

	proposalSummary := make([]processedProposalStatusSummaryResponse, 0, len(snapshot.Navigation.ProposalSummary))
	for _, item := range snapshot.Navigation.ProposalSummary {
		proposalSummary = append(proposalSummary, processedProposalStatusSummaryResponse{
			RecommendationStatus: item.RecommendationStatus,
			ProposalCount:        item.ProposalCount,
			RequestCount:         item.RequestCount,
			DocumentCount:        item.DocumentCount,
			LatestCreatedAt:      item.LatestCreatedAt,
		})
	}

	primaryActions, secondaryActions := buildHomeActions(sessionContext, snapshot.Navigation.InboundSummary, snapshot.Navigation.ProposalSummary, snapshot.Navigation.PendingApprovals)
	roleHeadline, roleBody := roleAwareHomeIntro(sessionContext)

	response := navigationDashboardResponse{
		RoleHeadline:     roleHeadline,
		RoleBody:         roleBody,
		PrimaryActions:   mapNavigationHomeActions(primaryActions),
		SecondaryActions: mapNavigationHomeActions(secondaryActions),
		InboundSummary:   inboundSummary,
		ProposalSummary:  proposalSummary,
		InboundRequests:  make([]inboundRequestReviewResponse, 0, len(snapshot.InboundRequests)),
		Proposals:        make([]processedProposalReviewResponse, 0, len(snapshot.Proposals)),
		Approvals:        make([]approvalQueueEntryResponse, 0, len(snapshot.Navigation.PendingApprovals)),
	}
	for _, item := range snapshot.InboundRequests {
		response.InboundRequests = append(response.InboundRequests, mapInboundRequestReview(item))
	}
	for _, item := range snapshot.Proposals {
		response.Proposals = append(response.Proposals, mapProcessedProposalReview(item))
	}
	for _, item := range snapshot.Navigation.PendingApprovals {
		response.Approvals = append(response.Approvals, mapApprovalQueueEntry(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetNavigationOperations(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != navigationOperationsPath {
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
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	snapshot, err := h.reviewService.GetOperationsLandingSnapshot(r.Context(), actor, 10, 10)
	if err != nil {
		handleReviewError(w, err, "failed to load operations landing")
		return
	}

	sortInboundRequestStatusSummaries(snapshot.Navigation.InboundSummary)
	requests := make([]inboundRequestReviewResponse, 0, len(snapshot.Feed.Requests))
	for _, item := range snapshot.Feed.Requests {
		requests = append(requests, mapInboundRequestReview(item))
	}
	proposals := make([]processedProposalReviewResponse, 0, len(snapshot.Feed.Proposals))
	for _, item := range snapshot.Feed.Proposals {
		proposals = append(proposals, mapProcessedProposalReview(item))
	}
	approvals := make([]approvalQueueEntryResponse, 0, len(snapshot.Navigation.PendingApprovals))
	for _, item := range snapshot.Navigation.PendingApprovals {
		approvals = append(approvals, mapApprovalQueueEntry(item))
	}

	feed := append(buildSvelteOperationsFeedFromRequests(requests), buildSvelteOperationsFeedFromProposals(proposals)...)
	feed = append(feed, buildSvelteOperationsFeedFromApprovals(approvals)...)

	writeJSON(w, http.StatusOK, navigationOperationsResponse{
		QueuedRequestCount:   countQueuedRequests(snapshot.Navigation.InboundSummary),
		PendingApprovalCount: len(snapshot.Navigation.PendingApprovals),
		ProposalReviewCount:  len(snapshot.Feed.Proposals),
		RecentFeed:           sortNavigationFeed(feed, 8),
	})
}

func (h *AgentAPIHandler) handleGetNavigationOperationsFeed(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != navigationOperationsFeedPath {
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
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	snapshot, err := h.reviewService.GetOperationsFeedSnapshot(r.Context(), actor, 30)
	if err != nil {
		handleReviewError(w, err, "failed to load operations feed")
		return
	}

	requests := make([]inboundRequestReviewResponse, 0, len(snapshot.Requests))
	for _, item := range snapshot.Requests {
		requests = append(requests, mapInboundRequestReview(item))
	}
	proposals := make([]processedProposalReviewResponse, 0, len(snapshot.Proposals))
	for _, item := range snapshot.Proposals {
		proposals = append(proposals, mapProcessedProposalReview(item))
	}
	approvals := make([]approvalQueueEntryResponse, 0, len(snapshot.Approvals))
	for _, item := range snapshot.Approvals {
		approvals = append(approvals, mapApprovalQueueEntry(item))
	}

	feed := append(buildSvelteOperationsFeedFromRequests(requests), buildSvelteOperationsFeedFromProposals(proposals)...)
	feed = append(feed, buildSvelteOperationsFeedFromApprovals(approvals)...)
	writeJSON(w, http.StatusOK, navigationOperationsFeedResponse{
		Items: sortNavigationFeed(feed, 30),
	})
}

func (h *AgentAPIHandler) handleGetNavigationReview(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != navigationReviewPath {
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
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	snapshot, err := h.reviewService.GetWorkflowNavigationSnapshot(r.Context(), actor, 8)
	if err != nil {
		handleReviewError(w, err, "failed to load review landing")
		return
	}

	sortInboundRequestStatusSummaries(snapshot.InboundSummary)
	response := navigationReviewResponse{
		InboundSummary:      make([]inboundRequestStatusSummaryResponse, 0, len(snapshot.InboundSummary)),
		ProposalSummary:     make([]processedProposalStatusSummaryResponse, 0, len(snapshot.ProposalSummary)),
		PendingApprovals:    make([]approvalQueueEntryResponse, 0, len(snapshot.PendingApprovals)),
		InboundRequestCount: sumInboundRequestCount(snapshot.InboundSummary),
		ProposalCount:       sumProposalCount(snapshot.ProposalSummary),
	}
	for _, item := range snapshot.InboundSummary {
		response.InboundSummary = append(response.InboundSummary, inboundRequestStatusSummaryResponse{
			Status:           item.Status,
			RequestCount:     item.RequestCount,
			MessageCount:     item.MessageCount,
			AttachmentCount:  item.AttachmentCount,
			LatestReceivedAt: timePtr(item.LatestReceivedAt),
			LatestQueuedAt:   timePtr(item.LatestQueuedAt),
			LatestUpdatedAt:  item.LatestUpdatedAt,
		})
	}
	for _, item := range snapshot.ProposalSummary {
		response.ProposalSummary = append(response.ProposalSummary, processedProposalStatusSummaryResponse{
			RecommendationStatus: item.RecommendationStatus,
			ProposalCount:        item.ProposalCount,
			RequestCount:         item.RequestCount,
			DocumentCount:        item.DocumentCount,
			LatestCreatedAt:      item.LatestCreatedAt,
		})
	}
	for _, item := range snapshot.PendingApprovals {
		response.PendingApprovals = append(response.PendingApprovals, mapApprovalQueueEntry(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetNavigationAgentChat(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != navigationAgentChatPath {
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
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	snapshot, err := h.reviewService.GetAgentChatSnapshot(r.Context(), actor, 40, 40)
	if err != nil {
		handleReviewError(w, err, "failed to load recent coordinator conversations")
		return
	}

	response := navigationAgentChatResponse{
		RequestReference: strings.TrimSpace(r.URL.Query().Get("request_reference")),
		RequestStatus:    strings.TrimSpace(r.URL.Query().Get("request_status")),
		RecentRequests:   make([]inboundRequestReviewResponse, 0, len(snapshot.RecentRequests)),
		RecentProposals:  make([]processedProposalReviewResponse, 0, len(snapshot.RecentProposals)),
	}
	for _, item := range snapshot.RecentRequests {
		response.RecentRequests = append(response.RecentRequests, mapInboundRequestReview(item))
	}
	for _, item := range snapshot.RecentProposals {
		response.RecentProposals = append(response.RecentProposals, mapProcessedProposalReview(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AgentAPIHandler) handleGetNavigationRouteCatalog(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != navigationRouteCatalogPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	writeJSON(w, http.StatusOK, navigationRouteCatalogResponse{
		Query: query,
		Items: mapNavigationRouteCatalogEntries(filterRouteCatalogEntries(sessionContext, query)),
	})
}
