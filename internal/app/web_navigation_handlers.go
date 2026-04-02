package app

import (
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/reporting"
)

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
		{Title: "Submit inbound request", Href: webSubmitInboundPagePath, Category: "Operations", Summary: "Dedicated persisted intake route for new browser-originated requests.", Keywords: "intake submit inbound request create new"},
		{Title: "Operations landing", Href: webOperationsPath, Category: "Operations", Summary: "Bundle landing for queue movement, durable feed review, and agent-chat continuity.", Keywords: "operations queue feed agent chat landing"},
		{Title: "Operations feed", Href: webOperationsFeedPath, Category: "Operations", Summary: "Durable feed of recent request, proposal, and approval movement.", Keywords: "operations feed recent movement queue"},
		{Title: "Agent chat", Href: webAgentChatPath, Category: "Operations", Summary: "Request-centered chat-style intake and continuity surface on the same persisted backend.", Keywords: "agent chat request intake guidance"},
		{Title: "Review landing", Href: webReviewPath, Category: "Review", Summary: "Grouped review taxonomy for inbound requests, proposals, approvals, and downstream review families.", Keywords: "review landing taxonomy approvals proposals documents"},
		{Title: "Inbound requests review", Href: webInboundRequestsPath, Category: "Review", Summary: "Exact request review for draft, queued, processing, failed, and completed lifecycle states.", Keywords: "requests review drafts queued failed processed"},
		{Title: "Proposal review", Href: webProposalsPath, Category: "Review", Summary: "Processed coordinator proposals with continuity into approvals and downstream documents.", Keywords: "proposals processed recommendations approval requested"},
		{Title: "Approval review", Href: webApprovalsPath, Category: "Review", Summary: "Approval queue review with exact decision continuity and upstream request provenance.", Keywords: "approvals queue pending decision approver"},
		{Title: "Document review", Href: webDocumentsPath, Category: "Review", Summary: "Document review after proposal and approval work has crossed into document truth.", Keywords: "documents review downstream"},
		{Title: "Accounting review", Href: webAccountingPath, Category: "Review", Summary: "Journal-entry, control-account, and tax-summary review on centralized posting truth.", Keywords: "accounting journal control account tax"},
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
			scored = append(scored, scoredRouteCatalogEntry{
				entry: entry,
				score: score,
			})
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

func (h *AgentAPIHandler) handleWebRouteCatalog(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webRouteCatalogPath {
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

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	data := webRouteCatalogData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
		Query:   query,
		Results: filterRouteCatalogEntries(sessionContext, query),
	}

	h.renderWebPage(w, webPageData{
		Title:        "workflow_app",
		ActivePath:   webRouteCatalogPath,
		Notice:       data.Notice,
		Error:        data.Error,
		Session:      &sessionContext,
		RouteCatalog: &data,
	})
}

func (h *AgentAPIHandler) handleWebSettings(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webSettingsPath {
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

	data := webSettingsData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
		SettingsPrinciples: []string{
			"Settings stays user-scoped: session context, personal continuity, and safe workflow shortcuts belong here.",
			"Org-scoped maintenance, access-sensitive setup, and governed controls belong under Admin for authorized actors.",
			"Workflow pages remain the primary working surfaces; utility pages should route back into exact operational or review paths.",
		},
		PersonalUtilityLinks: []webHomeAction{
			{Title: "Open route catalog", Summary: "Search grouped route families when the next workflow surface is not obvious from the shell.", Href: webRouteCatalogPath},
			{Title: "Open home", Summary: "Return to the role-aware home surface for workload-prioritized entry points.", Href: webAppPath},
		},
	}

	if h.reviewService != nil {
		snapshot, snapshotErr := h.reviewService.GetWorkflowNavigationSnapshot(r.Context(), sessionContext.Actor, 10)
		if snapshotErr != nil {
			data.Error = "failed to load settings"
		} else {
			sortInboundRequestStatusSummaries(snapshot.InboundSummary)
			data.PrimaryActions, _ = buildHomeActions(sessionContext, snapshot.InboundSummary, snapshot.ProposalSummary, snapshot.PendingApprovals)
		}
	}
	if strings.EqualFold(strings.TrimSpace(sessionContext.RoleCode), identityaccess.RoleAdmin) {
		data.AdminContinuation = &webHomeAction{
			Title:   "Open admin maintenance hub",
			Summary: "Use the separate admin surface for org-scoped setup families, governance review, and later privileged controls.",
			Href:    webAdminPath,
		}
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webSettingsPath,
		Notice:     data.Notice,
		Error:      data.Error,
		Session:    &sessionContext,
		Settings:   &data,
	})
}

func (h *AgentAPIHandler) handleWebAdmin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAdminPath {
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
	if !strings.EqualFold(strings.TrimSpace(sessionContext.RoleCode), identityaccess.RoleAdmin) {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("admin surface requires admin role"), http.StatusSeeOther)
		return
	}

	data := webAdminData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
		AdminPrinciples: []string{
			"Admin owns org-scoped maintenance, access-sensitive setup, and governed operational controls.",
			"Privileged maintenance should stay bounded to foundational setup and exception handling on shared domain services.",
			"Review pages remain read-first workflow surfaces; admin maintenance should not dissolve them into broad spreadsheet editing.",
		},
		MaintenanceFamilies: []webAdminFamily{
			{
				Title:        "Accounting setup",
				Summary:      "Foundational ledger-account, tax-code, and accounting-period maintenance belongs here instead of under ordinary review routes.",
				CurrentHref:  webAdminAccountingPath,
				CurrentLabel: "Open accounting setup",
				NextSlice:    "Current slice: admin-only browser and API maintenance now expose bounded list, create, and period-close controls on the shared accounting service seam.",
			},
			{
				Title:        "Party setup",
				Summary:      "Customer and vendor support records should reuse the shared party model rather than reopening CRM-first product gravity.",
				CurrentHref:  webDocumentsPath,
				CurrentLabel: "Open document review",
				NextSlice:    "Next slice: expose bounded admin-only customer and party maintenance while keeping workflow ownership in the shared backend.",
			},
			{
				Title:        "Access and governance",
				Summary:      "Approval, audit, and later access-management controls stay grouped under one privileged maintenance posture.",
				CurrentHref:  webAuditPath,
				CurrentLabel: "Open audit review",
				NextSlice:    "Follow-on slice: add user, role, and later policy controls only after the foundational maintenance seams above are stable.",
			},
		},
		AdminActions: []webHomeAction{
			{Title: "Open accounting setup", Summary: "Create ledger accounts, tax codes, and accounting periods from the bounded admin maintenance surface.", Href: webAdminAccountingPath},
			{Title: "Open approval queue", Summary: "Keep explicit approval decisions ahead of downstream document or posting review.", Href: webApprovalsPath + "?status=pending"},
			{Title: "Open accounting review", Summary: "Use centralized accounting review for posted truth, control accounts, and tax summaries.", Href: webAccountingPath},
			{Title: "Open audit review", Summary: "Trace actor, timestamp, and workflow causation without leaving the shared browser seam.", Href: webAuditPath},
			{Title: "Open route catalog", Summary: "Search the grouped shell when the next exact route is outside the primary navigation bands.", Href: webRouteCatalogPath},
		},
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webAdminPath,
		Notice:     data.Notice,
		Error:      data.Error,
		Session:    &sessionContext,
		Admin:      &data,
	})
}
