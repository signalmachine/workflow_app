package app

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/reporting"
)

type webHomeAction struct {
	Title   string
	Summary string
	Href    string
	Badge   string
}

type webRouteCatalogEntry struct {
	Title        string
	Href         string
	Category     string
	Summary      string
	Keywords     string
	RequiresRole string
}

type webOperationsFeedItem struct {
	OccurredAt     time.Time
	Kind           string
	Title          string
	Summary        string
	Status         string
	PrimaryLabel   string
	PrimaryHref    string
	SecondaryLabel string
	SecondaryHref  string
}

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

func roleAwareHomeIntro(session identityaccess.SessionContext) (string, string) {
	name := strings.TrimSpace(session.UserDisplayName)
	if name == "" {
		name = strings.TrimSpace(session.UserEmail)
	}

	switch strings.TrimSpace(session.RoleCode) {
	case identityaccess.RoleAdmin:
		return "Admin control surface", name + ", start from the workflow queue that is blocking other operators, then use utility surfaces for audit, accounting, and access-sensitive review."
	case identityaccess.RoleApprover:
		return "Approval-focused home", name + ", keep decision work ahead of broad browsing so pending approvals and approval-ready proposals stay close to the first click."
	default:
		return "Operator home", name + ", keep request intake, queue movement, and exact workflow continuity ahead of broad downstream review."
	}
}

func countInboundRequestsByStatus(rows []reporting.InboundRequestStatusSummary, status string) int {
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.Status), strings.TrimSpace(status)) {
			return row.RequestCount
		}
	}
	return 0
}

func countProposalsByStatus(rows []reporting.ProcessedProposalStatusSummary, status string) int {
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.RecommendationStatus), strings.TrimSpace(status)) {
			return row.ProposalCount
		}
	}
	return 0
}

func buildHomeActions(session identityaccess.SessionContext, inboundSummary []reporting.InboundRequestStatusSummary, proposalSummary []reporting.ProcessedProposalStatusSummary, approvals []reporting.ApprovalQueueEntry) ([]webHomeAction, []webHomeAction) {
	var primary []webHomeAction
	var secondary []webHomeAction

	appendAction := func(items []webHomeAction, title, summary, href string, count int) []webHomeAction {
		action := webHomeAction{
			Title:   title,
			Summary: summary,
			Href:    href,
		}
		if count > 0 {
			action.Badge = strconv.Itoa(count)
		}
		return append(items, action)
	}

	draftCount := countInboundRequestsByStatus(inboundSummary, "draft")
	queuedCount := countInboundRequestsByStatus(inboundSummary, "queued")
	failedCount := countInboundRequestsByStatus(inboundSummary, "failed")
	pendingApprovalCount := len(approvals)
	approvalReadyProposalCount := countProposalsByStatus(proposalSummary, "approval_requested")

	switch strings.TrimSpace(session.RoleCode) {
	case identityaccess.RoleAdmin:
		if pendingApprovalCount > 0 {
			primary = appendAction(primary, "Review pending approvals", "Explicit decision seams remain the highest-leverage unblocker for downstream workflow movement.", webApprovalsPath+"?status=pending", pendingApprovalCount)
		}
		if queuedCount > 0 {
			primary = appendAction(primary, "Open queued requests", "Use the grouped review queue before processing so operators can confirm the next exact request path.", webInboundRequestsPath+"?status=queued", queuedCount)
		}
		primary = appendAction(primary, "Open admin maintenance hub", "Use the admin hub for privileged setup families, governed review paths, and later maintenance flows.", webAdminPath, 0)
		secondary = appendAction(secondary, "Review failures", "Failure visibility stays close to the operator home so broken coordinator paths do not hide behind downstream review.", webInboundRequestsPath+"?status=failed", failedCount)
		secondary = appendAction(secondary, "Open accounting review", "Use centralized accounting review for posted truth, control accounts, and tax-summary continuity.", webAccountingPath, 0)
		secondary = appendAction(secondary, "Open audit review", "Use audit lookup when the question is actor, causation, or exact state-transition provenance.", webAuditPath, 0)
	case identityaccess.RoleApprover:
		if pendingApprovalCount > 0 {
			primary = appendAction(primary, "Review pending approvals", "Decision work should stay on the shortest path from home to exact approval detail.", webApprovalsPath+"?status=pending", pendingApprovalCount)
		}
		if approvalReadyProposalCount > 0 {
			primary = appendAction(primary, "Open approval-ready proposals", "Processed proposals that already point toward an approval queue stay close to approval work.", webProposalsPath+"?status=approval_requested", approvalReadyProposalCount)
		}
		primary = appendAction(primary, "Open review landing", "Use the grouped review taxonomy before dropping into downstream accounting, document, or audit paths.", webReviewPath, 0)
		secondary = appendAction(secondary, "Continue drafts", "Draft requests still matter because incomplete intake blocks the future approval queue.", webInboundRequestsPath+"?status=draft", draftCount)
		secondary = appendAction(secondary, "Open route catalog", "Search by workflow term, route family, or operator intent when the next surface is not obvious.", webRouteCatalogPath, 0)
	default:
		if draftCount > 0 {
			primary = appendAction(primary, "Continue drafts", "Resume parked drafts before they enter the queue and turn into downstream review work.", webInboundRequestsPath+"?status=draft", draftCount)
		}
		if queuedCount > 0 {
			primary = appendAction(primary, "Open queued requests", "Queued requests are ready for coordinator pickup and exact lifecycle review.", webInboundRequestsPath+"?status=queued", queuedCount)
		}
		primary = appendAction(primary, "Start a new request", "Use the dedicated intake route instead of flattening request creation into the home surface.", webSubmitInboundPagePath, 0)
		secondary = appendAction(secondary, "Review failures", "When coordinator work breaks, the home surface should keep recovery one click away.", webInboundRequestsPath+"?status=failed", failedCount)
		secondary = appendAction(secondary, "Open operations landing", "Use the operations bundle for feed monitoring, agent chat, and queue-moving actions.", webOperationsPath, 0)
		secondary = appendAction(secondary, "Open route catalog", "Search by route title, domain term, or operator intent when the grouped shell still leaves ambiguity.", webRouteCatalogPath, 0)
	}

	if len(primary) == 0 {
		primary = appendAction(primary, "Open review landing", "The grouped review landing is the safest default when workload-specific signals are absent.", webReviewPath, 0)
	}
	if len(secondary) == 0 {
		secondary = appendAction(secondary, "Open route catalog", "Use the searchable catalog when the home surface has no stronger live recommendation.", webRouteCatalogPath, 0)
	}

	return primary, secondary
}

func routeCatalogEntries() []webRouteCatalogEntry {
	return []webRouteCatalogEntry{
		{Title: "Home", Href: webAppPath, Category: "Core shell", Summary: "Role-aware operator start surface with workload-oriented shortcuts and continuity into active workflow families.", Keywords: "dashboard home quick links workload operator"},
		{Title: "Route catalog", Href: webRouteCatalogPath, Category: "Core shell", Summary: "Searchable navigation surface for route titles, workflow terms, and common operator intent.", Keywords: "catalog routes search command palette navigation"},
		{Title: "Settings", Href: webSettingsPath, Category: "Utility", Summary: "User-scoped utility surface for session context, personal continuity, and safe handoff back into workflow routes.", Keywords: "settings session preferences home shortcuts utility personal"},
		{Title: "Admin", Href: webAdminPath, Category: "Utility", Summary: "Privileged maintenance hub for governed setup families, review controls, and admin-only route continuity.", Keywords: "admin governance accounting setup maintenance privileged utility", RequiresRole: identityaccess.RoleAdmin},
		{Title: "Admin accounting setup", Href: webAdminAccountingPath, Category: "Utility", Summary: "Admin-only setup surface for ledger accounts, tax codes, and accounting periods on the shared backend.", Keywords: "admin accounting setup ledger accounts tax codes accounting periods maintenance", RequiresRole: identityaccess.RoleAdmin},
		{Title: "Admin party setup", Href: webAdminPartiesPath, Category: "Utility", Summary: "Admin-only customer and party maintenance on the shared support-record model with exact detail continuity.", Keywords: "admin party setup customer vendor counterparty contacts maintenance", RequiresRole: identityaccess.RoleAdmin},
		{Title: "Admin access controls", Href: webAdminAccessPath, Category: "Utility", Summary: "Admin-only user membership and role maintenance on the shared identity and session foundation.", Keywords: "admin access users roles memberships identity governance maintenance", RequiresRole: identityaccess.RoleAdmin},
		{Title: "Admin inventory setup", Href: webAdminInventoryPath, Category: "Utility", Summary: "Admin-only item and location setup on the shared inventory foundation before downstream stock and movement work begins.", Keywords: "admin inventory setup items locations warehouse stock maintenance", RequiresRole: identityaccess.RoleAdmin},
		{Title: "Submit inbound request", Href: webSubmitInboundPagePath, Category: "Operations", Summary: "Dedicated persisted intake route for new browser-originated requests.", Keywords: "intake submit inbound request create new"},
		{Title: "Operations landing", Href: webOperationsPath, Category: "Operations", Summary: "Bundle landing for queue movement, durable feed review, and agent-chat continuity.", Keywords: "operations queue feed agent chat landing"},
		{Title: "Operations feed", Href: webOperationsFeedPath, Category: "Operations", Summary: "Durable feed of recent request, proposal, and approval movement.", Keywords: "operations feed recent movement queue"},
		{Title: "Agent chat", Href: webAgentChatPath, Category: "Operations", Summary: "Request-centered chat-style intake and continuity surface on the same persisted backend.", Keywords: "agent chat request intake guidance"},
		{Title: "Review landing", Href: webReviewPath, Category: "Review", Summary: "Grouped review taxonomy for inbound requests, proposals, approvals, and downstream review families.", Keywords: "review landing taxonomy approvals proposals documents"},
		{Title: "Inbound requests review", Href: webInboundRequestsPath, Category: "Review", Summary: "Exact request review for draft, queued, processing, failed, and completed lifecycle states.", Keywords: "requests review drafts queued failed processed"},
		{Title: "Proposal review", Href: webProposalsPath, Category: "Review", Summary: "Processed coordinator proposals with continuity into approvals and downstream documents.", Keywords: "proposals processed recommendations approval requested"},
		{Title: "Approval review", Href: webApprovalsPath, Category: "Review", Summary: "Approval queue review with exact decision continuity and upstream request provenance.", Keywords: "approvals queue pending decision approver"},
		{Title: "Document review", Href: webDocumentsPath, Category: "Review", Summary: "Document review after proposal and approval work has crossed into document truth.", Keywords: "documents review downstream"},
		{Title: "Accounting review", Href: webAccountingPath, Category: "Review", Summary: "Journal-entry, control-account, tax-summary, and financial-statement review on centralized posting truth.", Keywords: "accounting journal control account tax trial balance sheet income statement reports"},
		{Title: "Trial balance", Href: webAccountingTrialPath, Category: "Review", Summary: "Debit and credit balance report across active ledger accounts.", Keywords: "accounting trial balance debit credit imbalance report"},
		{Title: "Balance sheet", Href: webAccountingBalancePath, Category: "Review", Summary: "Financial position report for assets, liabilities, equity, and current earnings.", Keywords: "accounting balance sheet assets liabilities equity current earnings report"},
		{Title: "Income statement", Href: webAccountingIncomePath, Category: "Review", Summary: "Revenue, expense, and net income report by effective-date range.", Keywords: "accounting income statement profit loss revenue expense net income report"},
		{Title: "Inventory landing", Href: webInventoryHubPath, Category: "Inventory", Summary: "Domain landing for stock position, movement history, and pending handoff exceptions.", Keywords: "inventory landing stock movement reconciliation"},
		{Title: "Inventory review", Href: webInventoryPath, Category: "Inventory", Summary: "Review stock, movements, and reconciliation on the shared inventory reporting seam.", Keywords: "inventory review stock movement reconciliation"},
		{Title: "Work-order review", Href: webWorkOrdersPath, Category: "Execution", Summary: "Execution review for work orders, tasks, labor, material usage, and posted costs.", Keywords: "work orders execution labor material review"},
		{Title: "Audit review", Href: webAuditPath, Category: "Trace", Summary: "Lookup audit events when the question is actor, timestamp, or causal chain.", Keywords: "audit trace actor causation history"},
	}
}

func routeCatalogSearchTerms(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}
	return strings.FieldsFunc(query, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == ';' || r == '/' || r == '-'
	})
}

func routeCatalogSearchScore(entry webRouteCatalogEntry, query string) int {
	searchable := strings.ToLower(strings.Join([]string{
		entry.Title,
		entry.Category,
		entry.Summary,
		entry.Href,
		entry.Keywords,
	}, " "))
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return 1
	}

	score := 0
	if strings.Contains(searchable, query) {
		score += 100
	}

	terms := routeCatalogSearchTerms(query)
	if len(terms) == 0 {
		return score
	}
	for _, term := range terms {
		if !strings.Contains(searchable, term) {
			return 0
		}
		score += 10
		if strings.Contains(strings.ToLower(entry.Title), term) {
			score += 5
		}
		if strings.Contains(strings.ToLower(entry.Keywords), term) {
			score += 3
		}
	}
	return score
}

func filterRouteCatalogEntries(session identityaccess.SessionContext, query string) []webRouteCatalogEntry {
	query = strings.ToLower(strings.TrimSpace(query))
	type scoredRouteCatalogEntry struct {
		entry webRouteCatalogEntry
		score int
	}

	scored := make([]scoredRouteCatalogEntry, 0, len(routeCatalogEntries()))
	for _, entry := range routeCatalogEntries() {
		if entry.RequiresRole != "" && !strings.EqualFold(entry.RequiresRole, strings.TrimSpace(session.RoleCode)) {
			continue
		}
		score := routeCatalogSearchScore(entry, query)
		if score > 0 {
			scored = append(scored, scoredRouteCatalogEntry{entry: entry, score: score})
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if scored[i].entry.Category != scored[j].entry.Category {
			return scored[i].entry.Category < scored[j].entry.Category
		}
		return scored[i].entry.Title < scored[j].entry.Title
	})

	results := make([]webRouteCatalogEntry, 0, len(scored))
	for _, item := range scored {
		results = append(results, item.entry)
	}
	return results
}

func countQueuedRequests(rows []reporting.InboundRequestStatusSummary) int {
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.Status), "queued") {
			return row.RequestCount
		}
	}
	return 0
}

func sumInboundRequestCount(rows []reporting.InboundRequestStatusSummary) int {
	total := 0
	for _, row := range rows {
		total += row.RequestCount
	}
	return total
}

func sumProposalCount(rows []reporting.ProcessedProposalStatusSummary) int {
	total := 0
	for _, row := range rows {
		total += row.ProposalCount
	}
	return total
}

func sortInboundRequestStatusSummaries(rows []reporting.InboundRequestStatusSummary) {
	statusOrder := map[string]int{
		"draft":      0,
		"queued":     1,
		"processing": 2,
		"failed":     3,
		"cancelled":  4,
		"processed":  5,
		"acted_on":   6,
		"completed":  7,
	}
	sort.SliceStable(rows, func(i, j int) bool {
		left := strings.ToLower(strings.TrimSpace(rows[i].Status))
		right := strings.ToLower(strings.TrimSpace(rows[j].Status))
		leftRank, leftOK := statusOrder[left]
		rightRank, rightOK := statusOrder[right]
		switch {
		case leftOK && rightOK && leftRank != rightRank:
			return leftRank < rightRank
		case leftOK != rightOK:
			return leftOK
		case rows[i].LatestUpdatedAt != rows[j].LatestUpdatedAt:
			return rows[i].LatestUpdatedAt.After(rows[j].LatestUpdatedAt)
		default:
			return left < right
		}
	})
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
