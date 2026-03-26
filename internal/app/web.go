package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/reporting"
	"workflow_app/internal/workflow"
)

var webAppTemplate = template.Must(template.New("app").Funcs(template.FuncMap{
	"formatTime":         formatTemplateTime,
	"prettyJSON":         prettyTemplateJSON,
	"statusClass":        templateStatusClass,
	"documentReviewHref": templateDocumentReviewHref,
	"approvalReviewHref": templateApprovalReviewHref,
	"approvalQueueHref":  templateApprovalQueueHref,
	"proposalDetailHref": templateProposalDetailHref,
	"proposalReviewHref": templateProposalReviewHref,
	"auditEntityHref":    templateAuditEntityHref,
	"auditEntityLabel":   templateAuditEntityLabel,
}).Parse(webAppHTML))

type webAppDashboardData struct {
	Session         identityaccess.SessionContext
	Notice          string
	Error           string
	InboundSummary  []reporting.InboundRequestStatusSummary
	InboundRequests []reporting.InboundRequestReview
	Proposals       []reporting.ProcessedProposalReview
	Approvals       []reporting.ApprovalQueueEntry
}

type webInboundDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Detail  reporting.InboundRequestDetail
}

type webInboundRequestsData struct {
	Session          identityaccess.SessionContext
	Notice           string
	Error            string
	Status           string
	RequestReference string
	StatusSummary    []reporting.InboundRequestStatusSummary
	Requests         []reporting.InboundRequestReview
}

type webDocumentsData struct {
	Session    identityaccess.SessionContext
	Notice     string
	Error      string
	DocumentID string
	TypeCode   string
	Status     string
	Documents  []reporting.DocumentReview
}

type webDocumentDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Review  reporting.DocumentReview
}

type webAccountingData struct {
	Session         identityaccess.SessionContext
	Notice          string
	Error           string
	StartOn         string
	EndOn           string
	AsOf            string
	DocumentID      string
	JournalEntries  []reporting.JournalEntryReview
	ControlBalances []reporting.ControlAccountBalance
	TaxSummaries    []reporting.TaxSummary
}

type webProposalsData struct {
	Session            identityaccess.SessionContext
	Notice             string
	Error              string
	RecommendationID   string
	Status             string
	RequestReference   string
	StatusSummary      []reporting.ProcessedProposalStatusSummary
	ProcessedProposals []reporting.ProcessedProposalReview
}

type webProposalDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Review  reporting.ProcessedProposalReview
}

type webApprovalsData struct {
	Session    identityaccess.SessionContext
	Notice     string
	Error      string
	ApprovalID string
	Status     string
	QueueCode  string
	Approvals  []reporting.ApprovalQueueEntry
}

type webApprovalDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Entry   reporting.ApprovalQueueEntry
}

type webInventoryData struct {
	Session               identityaccess.SessionContext
	Notice                string
	Error                 string
	MovementID            string
	ItemID                string
	LocationID            string
	DocumentID            string
	MovementType          string
	OnlyPendingAccounting bool
	OnlyPendingExecution  bool
	Stock                 []reporting.InventoryStockItem
	Movements             []reporting.InventoryMovementReview
	Reconciliation        []reporting.InventoryReconciliationItem
}

type webWorkOrdersData struct {
	Session    identityaccess.SessionContext
	Notice     string
	Error      string
	Status     string
	DocumentID string
	WorkOrders []reporting.WorkOrderReview
}

type webWorkOrderDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Review  reporting.WorkOrderReview
}

type webAuditData struct {
	Session    identityaccess.SessionContext
	Notice     string
	Error      string
	EntityType string
	EntityID   string
	Events     []reporting.AuditEvent
}

func (h *AgentAPIHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, webAppPath, http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebAppDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAppPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		if errors.Is(err, identityaccess.ErrUnauthorized) {
			h.renderWebPage(w, webPageData{
				Title:      "workflow_app",
				Notice:     strings.TrimSpace(r.URL.Query().Get("notice")),
				Error:      strings.TrimSpace(r.URL.Query().Get("error")),
				ShowLogin:  true,
				LoginPath:  webLoginPath,
				ActivePath: webAppPath,
			})
			return
		}
		http.Error(w, "failed to load session", http.StatusInternalServerError)
		return
	}

	actor := sessionContext.Actor
	data := webAppDashboardData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
	}
	if h.reviewService != nil {
		if data.InboundSummary, err = h.reviewService.ListInboundRequestStatusSummary(r.Context(), actor); err != nil {
			data.Error = "failed to load inbound request summary"
		}
		if data.InboundRequests, err = h.reviewService.ListInboundRequests(r.Context(), reporting.ListInboundRequestsInput{
			Limit: 20,
			Actor: actor,
		}); err != nil {
			data.Error = "failed to load inbound requests"
		}
		if data.Proposals, err = h.reviewService.ListProcessedProposals(r.Context(), reporting.ListProcessedProposalsInput{
			Limit: 10,
			Actor: actor,
		}); err != nil {
			data.Error = "failed to load processed proposals"
		}
		if data.Approvals, err = h.reviewService.ListApprovalQueue(r.Context(), reporting.ListApprovalQueueInput{
			Status: "pending",
			Limit:  10,
			Actor:  actor,
		}); err != nil {
			data.Error = "failed to load approval queue"
		}
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webAppPath,
		Session:    &sessionContext,
		Dashboard:  &data,
	})
}

func (h *AgentAPIHandler) handleWebInboundRequests(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webInboundRequestsPath {
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

	data := webInboundRequestsData{
		Session:          sessionContext,
		Notice:           strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:            strings.TrimSpace(r.URL.Query().Get("error")),
		Status:           strings.TrimSpace(r.URL.Query().Get("status")),
		RequestReference: strings.TrimSpace(r.URL.Query().Get("request_reference")),
	}
	data.StatusSummary, err = h.reviewService.ListInboundRequestStatusSummary(r.Context(), sessionContext.Actor)
	if err != nil {
		data.Error = "failed to load inbound request summary"
	}
	if data.Requests, err = h.reviewService.ListInboundRequests(r.Context(), reporting.ListInboundRequestsInput{
		Status:           data.Status,
		RequestReference: data.RequestReference,
		Limit:            50,
		Actor:            sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load inbound requests"
	}

	h.renderWebPage(w, webPageData{
		Title:           "workflow_app",
		ActivePath:      webInboundRequestsPath,
		Session:         &sessionContext,
		InboundRequests: &data,
	})
}

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
		Session:    sessionContext,
		Notice:     strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:      strings.TrimSpace(r.URL.Query().Get("error")),
		StartOn:    formatDateInput(startOn),
		EndOn:      formatDateInput(endOn),
		AsOf:       formatDateInput(asOf),
		DocumentID: strings.TrimSpace(r.URL.Query().Get("document_id")),
	}

	data.JournalEntries, err = h.reviewService.ListJournalEntries(r.Context(), reporting.ListJournalEntriesInput{
		StartOn:    startOn,
		EndOn:      endOn,
		DocumentID: data.DocumentID,
		Limit:      50,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load journal entries"
	}
	if data.ControlBalances, err = h.reviewService.ListControlAccountBalances(r.Context(), reporting.ListControlAccountBalancesInput{
		AsOf:  asOf,
		Actor: sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load control account balances"
	}
	if data.TaxSummaries, err = h.reviewService.ListTaxSummaries(r.Context(), reporting.ListTaxSummariesInput{
		StartOn: startOn,
		EndOn:   endOn,
		Limit:   50,
		Actor:   sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load tax summaries"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webAccountingPath,
		Session:    &sessionContext,
		Accounting: &data,
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
		Session:    &sessionContext,
		Proposals:  &data,
	})
}

func (h *AgentAPIHandler) handleWebProposalDetail(w http.ResponseWriter, r *http.Request) {
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
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			Review:  proposals[0],
		},
	})
}

func (h *AgentAPIHandler) handleWebApprovals(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webApprovalsPath {
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

	data := webApprovalsData{
		Session:    sessionContext,
		Notice:     strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:      strings.TrimSpace(r.URL.Query().Get("error")),
		ApprovalID: strings.TrimSpace(r.URL.Query().Get("approval_id")),
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		QueueCode:  strings.TrimSpace(r.URL.Query().Get("queue_code")),
	}
	data.Approvals, err = h.reviewService.ListApprovalQueue(r.Context(), reporting.ListApprovalQueueInput{
		ApprovalID: data.ApprovalID,
		Status:     data.Status,
		QueueCode:  data.QueueCode,
		Limit:      50,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load approval queue"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webApprovalsPath,
		Session:    &sessionContext,
		Approvals:  &data,
	})
}

func (h *AgentAPIHandler) handleWebApprovalDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	approvalID, ok := parseChildPath(webApprovalsPath, r.URL.Path)
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

	entries, err := h.reviewService.ListApprovalQueue(r.Context(), reporting.ListApprovalQueueInput{
		ApprovalID: approvalID,
		Limit:      2,
		Actor:      sessionContext.Actor,
	})
	if err != nil || len(entries) == 0 {
		http.Redirect(w, r, webApprovalsPath+"?error="+url.QueryEscape("failed to load approval"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webApprovalsPath,
		Session:    &sessionContext,
		ApprovalDetail: &webApprovalDetailData{
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			Entry:   entries[0],
		},
	})
}

func (h *AgentAPIHandler) handleWebInventory(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webInventoryPath {
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

	data := webInventoryData{
		Session:               sessionContext,
		Notice:                strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:                 strings.TrimSpace(r.URL.Query().Get("error")),
		MovementID:            strings.TrimSpace(r.URL.Query().Get("movement_id")),
		ItemID:                strings.TrimSpace(r.URL.Query().Get("item_id")),
		LocationID:            strings.TrimSpace(r.URL.Query().Get("location_id")),
		DocumentID:            strings.TrimSpace(r.URL.Query().Get("document_id")),
		MovementType:          strings.TrimSpace(r.URL.Query().Get("movement_type")),
		OnlyPendingAccounting: strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("only_pending_accounting")), "true"),
		OnlyPendingExecution:  strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("only_pending_execution")), "true"),
	}

	data.Stock, err = h.reviewService.ListInventoryStock(r.Context(), reporting.ListInventoryStockInput{
		ItemID:     data.ItemID,
		LocationID: data.LocationID,
		Limit:      50,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load inventory stock"
	}
	if data.Movements, err = h.reviewService.ListInventoryMovements(r.Context(), reporting.ListInventoryMovementsInput{
		MovementID:   data.MovementID,
		ItemID:       data.ItemID,
		LocationID:   data.LocationID,
		DocumentID:   data.DocumentID,
		MovementType: data.MovementType,
		Limit:        50,
		Actor:        sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load inventory movements"
	}
	if data.Reconciliation, err = h.reviewService.ListInventoryReconciliation(r.Context(), reporting.ListInventoryReconciliationInput{
		MovementID:            data.MovementID,
		ItemID:                data.ItemID,
		DocumentID:            data.DocumentID,
		OnlyPendingAccounting: data.OnlyPendingAccounting,
		OnlyPendingExecution:  data.OnlyPendingExecution,
		Limit:                 50,
		Actor:                 sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load inventory reconciliation"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webInventoryPath,
		Session:    &sessionContext,
		Inventory:  &data,
	})
}

func (h *AgentAPIHandler) handleWebWorkOrders(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webWorkOrdersPath {
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

	data := webWorkOrdersData{
		Session:    sessionContext,
		Notice:     strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:      strings.TrimSpace(r.URL.Query().Get("error")),
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		DocumentID: strings.TrimSpace(r.URL.Query().Get("document_id")),
	}
	data.WorkOrders, err = h.reviewService.ListWorkOrders(r.Context(), reporting.ListWorkOrdersInput{
		Status:     data.Status,
		DocumentID: data.DocumentID,
		Limit:      50,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load work orders"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webWorkOrdersPath,
		Session:    &sessionContext,
		WorkOrders: &data,
	})
}

func (h *AgentAPIHandler) handleWebWorkOrderDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	workOrderID, ok := parseChildPath(webWorkOrdersPath, r.URL.Path)
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

	review, err := h.reviewService.GetWorkOrderReview(r.Context(), reporting.GetWorkOrderReviewInput{
		WorkOrderID: workOrderID,
		Actor:       sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, webWorkOrdersPath+"?error="+url.QueryEscape("failed to load work order"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webWorkOrdersPath,
		Session:    &sessionContext,
		WorkOrderDetail: &webWorkOrderDetailData{
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			Review:  review,
		},
	})
}

func (h *AgentAPIHandler) handleWebAudit(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAuditPath {
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

	data := webAuditData{
		Session:    sessionContext,
		Notice:     strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:      strings.TrimSpace(r.URL.Query().Get("error")),
		EntityType: strings.TrimSpace(r.URL.Query().Get("entity_type")),
		EntityID:   strings.TrimSpace(r.URL.Query().Get("entity_id")),
	}
	data.Events, err = h.reviewService.LookupAuditEvents(r.Context(), reporting.LookupAuditEventsInput{
		EntityType: data.EntityType,
		EntityID:   data.EntityID,
		Limit:      100,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load audit events"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webAuditPath,
		Session:    &sessionContext,
		Audit:      &data,
	})
}

func (h *AgentAPIHandler) handleWebLogin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webLoginPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if h.authService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("auth service unavailable"), http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("invalid login form"), http.StatusSeeOther)
		return
	}

	deviceLabel := strings.TrimSpace(r.FormValue("device_label"))
	if deviceLabel == "" {
		deviceLabel = "browser"
	}

	session, err := h.authService.StartBrowserSession(r.Context(), identityaccess.StartBrowserSessionInput{
		OrgSlug:     strings.TrimSpace(r.FormValue("org_slug")),
		Email:       strings.TrimSpace(r.FormValue("email")),
		DeviceLabel: deviceLabel,
		ExpiresAt:   time.Now().UTC().Add(browserSessionDuration),
	})
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("invalid session credentials"), http.StatusSeeOther)
		return
	}

	setSessionCookies(w, session.Session.ID, session.RefreshToken, session.Session.ExpiresAt)
	http.Redirect(w, r, webAppPath+"?notice="+url.QueryEscape("Signed in."), http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebLogout(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webLogoutPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	sessionID, refreshToken, ok := sessionCookiesFromRequest(r)
	if ok && h.authService != nil {
		_ = h.authService.RevokeAuthenticatedSession(r.Context(), sessionID, refreshToken)
	}
	clearSessionCookies(w)
	http.Redirect(w, r, webAppPath+"?notice="+url.QueryEscape("Signed out."), http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebSubmitInboundRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webSubmitInboundPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if h.submissionService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("submission service unavailable"), http.StatusSeeOther)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("unauthorized"), http.StatusSeeOther)
		return
	}

	if err := r.ParseMultipartForm(16 << 20); err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("invalid submission form"), http.StatusSeeOther)
		return
	}

	var files []SubmitInboundRequestAttachmentInput
	if r.MultipartForm != nil {
		for _, fileHeader := range r.MultipartForm.File["attachments"] {
			file, openErr := fileHeader.Open()
			if openErr != nil {
				http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("failed to read attachment"), http.StatusSeeOther)
				return
			}

			content, readErr := io.ReadAll(file)
			_ = file.Close()
			if readErr != nil {
				http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("failed to read attachment"), http.StatusSeeOther)
				return
			}
			if len(content) == 0 {
				continue
			}

			mediaType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
			if mediaType == "" {
				mediaType = "application/octet-stream"
			}

			files = append(files, SubmitInboundRequestAttachmentInput{
				OriginalFileName: fileHeader.Filename,
				MediaType:        mediaType,
				ContentBase64:    base64.StdEncoding.EncodeToString(content),
				LinkRole:         attachments.LinkRoleEvidence,
			})
		}
	}

	result, err := h.submissionService.SubmitInboundRequest(r.Context(), SubmitInboundRequestInput{
		OriginType:     intake.OriginHuman,
		Channel:        "browser",
		Metadata:       map[string]any{"submitter_label": strings.TrimSpace(r.FormValue("submitter_label"))},
		MessageRole:    intake.MessageRoleRequest,
		MessageText:    strings.TrimSpace(r.FormValue("message_text")),
		Attachments:    files,
		QueueForReview: true,
		Actor:          actor,
	})
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("failed to submit inbound request"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(result.Request.RequestReference)+"?notice="+url.QueryEscape("Inbound request submitted."), http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebProcessNextQueuedInboundRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webProcessNextQueuedPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("unauthorized"), http.StatusSeeOther)
		return
	}

	processor, err := h.loadProcessor()
	if err != nil {
		if errors.Is(err, ErrAgentProviderNotConfigured) {
			http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("AI provider not configured."), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("failed to initialize agent processor"), http.StatusSeeOther)
		return
	}

	result, err := processor.ProcessNextQueuedInboundRequest(r.Context(), ProcessNextQueuedInboundRequestInput{
		Channel: "browser",
		Actor:   actor,
	})
	if err != nil {
		if errors.Is(err, intake.ErrNoQueuedInboundRequest) {
			http.Redirect(w, r, webAppPath+"?notice="+url.QueryEscape("No queued inbound requests."), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("failed to process queued inbound request"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(result.Request.RequestReference)+"?notice="+url.QueryEscape("Queued inbound request processed."), http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebInboundRequestDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	lookup, ok := parseChildPath(webSubmitInboundPath, r.URL.Path)
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

	input := reporting.GetInboundRequestDetailInput{Actor: sessionContext.Actor}
	if strings.HasPrefix(strings.ToUpper(lookup), "REQ-") {
		input.RequestReference = lookup
	} else {
		input.RequestID = lookup
	}

	detail, err := h.reviewService.GetInboundRequestDetail(r.Context(), input)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("failed to load inbound request detail"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webInboundRequestsPath,
		Session:    &sessionContext,
		Detail: &webInboundDetailData{
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			Detail:  detail,
		},
	})
}

func (h *AgentAPIHandler) handleWebApprovalDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if h.approvalService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("approval service unavailable"), http.StatusSeeOther)
		return
	}

	approvalID, ok := parseChildPath(strings.TrimSuffix(webApprovalDecisionPrefix, "/"), strings.TrimSuffix(r.URL.Path, "/decision"))
	if !ok || !strings.HasSuffix(r.URL.Path, "/decision") {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("unauthorized"), http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("invalid approval form"), http.StatusSeeOther)
		return
	}

	decision := strings.TrimSpace(r.FormValue("decision"))
	if decision != "approved" && decision != "rejected" {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("invalid approval decision"), http.StatusSeeOther)
		return
	}

	returnTo := sanitizeWebReturnPath(r.FormValue("return_to"))
	if returnTo == "" {
		returnTo = webAppPath
	}

	if _, _, err := h.approvalService.DecideApproval(r.Context(), workflow.DecideApprovalInput{
		ApprovalID:   approvalID,
		Decision:     decision,
		DecisionNote: strings.TrimSpace(r.FormValue("decision_note")),
		Actor:        actor,
	}); err != nil {
		http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to decide approval"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, appendWebMessage(returnTo, "notice", "Approval updated."), http.StatusSeeOther)
}

type webPageData struct {
	Title           string
	ActivePath      string
	Notice          string
	Error           string
	ShowLogin       bool
	LoginPath       string
	Session         *identityaccess.SessionContext
	Dashboard       *webAppDashboardData
	InboundRequests *webInboundRequestsData
	Detail          *webInboundDetailData
	Documents       *webDocumentsData
	DocumentDetail  *webDocumentDetailData
	Accounting      *webAccountingData
	Approvals       *webApprovalsData
	ApprovalDetail  *webApprovalDetailData
	Proposals       *webProposalsData
	ProposalDetail  *webProposalDetailData
	Inventory       *webInventoryData
	WorkOrders      *webWorkOrdersData
	WorkOrderDetail *webWorkOrderDetailData
	Audit           *webAuditData
}

func (h *AgentAPIHandler) renderWebPage(w http.ResponseWriter, data webPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = webAppTemplate.Execute(w, data)
}

func sanitizeWebReturnPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, webAppPath) {
		return ""
	}
	if strings.Contains(raw, "://") {
		return ""
	}
	return raw
}

func appendWebMessage(target, key, message string) string {
	separator := "?"
	if strings.Contains(target, "?") {
		separator = "&"
	}
	return target + separator + key + "=" + url.QueryEscape(message)
}

func formatTemplateTime(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.UTC().Format("2006-01-02 15:04:05 UTC")
}

func formatDateInput(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.DateOnly)
}

func prettyTemplateJSON(raw any) string {
	switch v := raw.(type) {
	case nil:
		return "{}"
	case []byte:
		if len(v) == 0 {
			return "{}"
		}
		var out bytes.Buffer
		if err := json.Indent(&out, v, "", "  "); err == nil {
			return out.String()
		}
		return string(v)
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(b)
	}
}

func templateStatusClass(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "processed", "completed", "approved":
		return "status-good"
	case "failed", "rejected", "cancelled":
		return "status-bad"
	default:
		return "status-neutral"
	}
}

func templateDocumentReviewHref(documentID string) string {
	documentID = strings.TrimSpace(documentID)
	if documentID == "" {
		return webDocumentsPath
	}
	return webDocumentDetailPrefix + url.PathEscape(documentID)
}

func templateApprovalReviewHref(approvalID string) string {
	approvalID = strings.TrimSpace(approvalID)
	if approvalID == "" {
		return webApprovalsPath
	}
	return webApprovalDetailPrefix + url.PathEscape(approvalID)
}

func templateApprovalQueueHref(queueCode, status string) string {
	values := url.Values{}
	if strings.TrimSpace(queueCode) != "" {
		values.Set("queue_code", strings.TrimSpace(queueCode))
	}
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "pending", "closed":
		values.Set("status", status)
	case "approved", "rejected":
		values.Set("status", "closed")
	}
	if encoded := values.Encode(); encoded != "" {
		return webApprovalsPath + "?" + encoded
	}
	return webApprovalsPath
}

func templateProposalReviewHref(recommendationID, status, requestReference string) string {
	values := url.Values{}
	if strings.TrimSpace(recommendationID) != "" {
		values.Set("recommendation_id", strings.TrimSpace(recommendationID))
	}
	if strings.TrimSpace(status) != "" {
		values.Set("status", strings.TrimSpace(status))
	}
	if strings.TrimSpace(requestReference) != "" {
		values.Set("request_reference", strings.TrimSpace(requestReference))
	}
	if encoded := values.Encode(); encoded != "" {
		return webProposalsPath + "?" + encoded
	}
	return webProposalsPath
}

func templateProposalDetailHref(recommendationID string) string {
	recommendationID = strings.TrimSpace(recommendationID)
	if recommendationID == "" {
		return webProposalsPath
	}
	return webProposalDetailPrefix + url.PathEscape(recommendationID)
}

func templateAuditEntityHref(entityType, entityID string) string {
	entityType = strings.TrimSpace(entityType)
	entityID = strings.TrimSpace(entityID)
	if entityType == "" || entityID == "" {
		return ""
	}

	switch entityType {
	case "documents.document":
		return templateDocumentReviewHref(entityID)
	case "ai.inbound_request":
		return webInboundDetailPrefix + url.PathEscape(entityID)
	case "workflow.approval":
		return templateApprovalReviewHref(entityID)
	case "ai.agent_recommendation":
		return templateProposalDetailHref(entityID)
	case "work_orders.work_order":
		return webWorkOrdersPath + "/" + url.PathEscape(entityID)
	case "inventory_ops.item":
		return webInventoryPath + "?item_id=" + url.QueryEscape(entityID)
	case "inventory_ops.location":
		return webInventoryPath + "?location_id=" + url.QueryEscape(entityID)
	case "inventory_ops.movement":
		return webInventoryPath + "?movement_id=" + url.QueryEscape(entityID)
	default:
		return ""
	}
}

func templateAuditEntityLabel(entityType string) string {
	switch strings.TrimSpace(entityType) {
	case "documents.document":
		return "Open document"
	case "ai.inbound_request":
		return "Open inbound request"
	case "workflow.approval":
		return "Open approval review"
	case "ai.agent_recommendation":
		return "Open proposal review"
	case "work_orders.work_order":
		return "Open work order"
	case "inventory_ops.item":
		return "Filter inventory by item"
	case "inventory_ops.location":
		return "Filter inventory by location"
	case "inventory_ops.movement":
		return "Open movement review"
	default:
		return ""
	}
}

const webAppHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    :root {
      --bg: #f5efe3;
      --panel: rgba(255,255,255,0.88);
      --ink: #1f1f1f;
      --muted: #5d5d5d;
      --line: #d8cdb8;
      --accent: #0f766e;
      --accent-soft: #dff3f1;
      --warn: #9a3412;
      --warn-soft: #fde8d8;
      --bad: #991b1b;
      --bad-soft: #fee2e2;
      --good: #166534;
      --good-soft: #dcfce7;
      --shadow: 0 24px 60px rgba(60, 41, 12, 0.12);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Georgia, "Times New Roman", serif;
      color: var(--ink);
      background:
        radial-gradient(circle at top left, rgba(15,118,110,0.18), transparent 28%),
        radial-gradient(circle at top right, rgba(154,52,18,0.16), transparent 30%),
        linear-gradient(180deg, #f7f2e8 0%, #efe6d6 100%);
    }
    a { color: #0f5e58; }
    .shell {
      width: min(1200px, calc(100% - 32px));
      margin: 24px auto 48px;
    }
    .masthead, .panel {
      background: var(--panel);
      border: 1px solid var(--line);
      box-shadow: var(--shadow);
      backdrop-filter: blur(10px);
      border-radius: 18px;
    }
    .masthead {
      padding: 24px;
      margin-bottom: 18px;
    }
    .masthead h1 {
      margin: 0 0 8px;
      font-size: clamp(2rem, 4vw, 3.4rem);
      line-height: 1;
      letter-spacing: -0.04em;
    }
    .masthead p, .meta {
      margin: 0;
      color: var(--muted);
    }
    .nav {
      margin-top: 16px;
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      align-items: center;
      justify-content: space-between;
    }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
      gap: 18px;
      align-items: start;
    }
    .panel { padding: 18px; }
    .panel h2, .panel h3 {
      margin-top: 0;
      margin-bottom: 12px;
      font-size: 1.15rem;
    }
    .notice, .error {
      padding: 12px 14px;
      border-radius: 12px;
      margin-bottom: 16px;
      border: 1px solid transparent;
    }
    .notice {
      background: var(--accent-soft);
      border-color: rgba(15,118,110,0.16);
    }
    .error {
      background: var(--warn-soft);
      border-color: rgba(154,52,18,0.22);
      color: var(--warn);
    }
    form { display: grid; gap: 12px; }
    label { font-weight: 600; }
    input, textarea, select, button {
      width: 100%;
      font: inherit;
      padding: 10px 12px;
      border-radius: 12px;
      border: 1px solid var(--line);
      background: rgba(255,255,255,0.9);
      color: var(--ink);
    }
    textarea { min-height: 132px; resize: vertical; }
    button {
      background: var(--accent);
      color: #fff;
      border: none;
      cursor: pointer;
      font-weight: 700;
    }
    button.secondary {
      background: #6b5c43;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      font-size: 0.96rem;
    }
    th, td {
      text-align: left;
      padding: 10px 8px;
      border-top: 1px solid var(--line);
      vertical-align: top;
    }
    th { color: var(--muted); font-size: 0.82rem; text-transform: uppercase; letter-spacing: 0.08em; }
    .status-pill {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 4px 10px;
      border-radius: 999px;
      font-size: 0.82rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .status-good { background: var(--good-soft); color: var(--good); }
    .status-bad { background: var(--bad-soft); color: var(--bad); }
    .status-neutral { background: #ece8df; color: #5f513d; }
    pre {
      margin: 0;
      white-space: pre-wrap;
      word-break: break-word;
      background: #f4efe7;
      border: 1px solid var(--line);
      border-radius: 12px;
      padding: 12px;
      overflow-x: auto;
    }
    .split {
      display: grid;
      grid-template-columns: 1.2fr 0.8fr;
      gap: 18px;
    }
    .summary-list {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
      gap: 12px;
    }
    .summary-card {
      padding: 14px;
      border: 1px solid var(--line);
      border-radius: 14px;
      background: rgba(255,255,255,0.68);
    }
    .summary-card strong {
      display: block;
      font-size: 1.4rem;
      margin-bottom: 6px;
    }
    .inline-form {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      align-items: center;
    }
    .inline-form input[type="text"] { min-width: 220px; }
    .stack { display: grid; gap: 18px; }
    .detail-block + .detail-block { margin-top: 16px; }
    @media (max-width: 880px) {
      .split { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <div class="shell">
    <section class="masthead">
      <h1>workflow_app</h1>
      <p>AI-agent-first intake, review, approvals, and operator control on one browser surface.</p>
      {{if .Session}}
      <div class="nav">
        <div>
          <div class="meta">Signed in as {{.Session.UserEmail}} in {{.Session.OrgName}} ({{.Session.RoleCode}})</div>
          <div class="meta" style="margin-top:8px;">
            <a href="/app">Operations</a> |
            <a href="/app/review/inbound-requests">Inbound requests</a> |
            <a href="/app/review/documents">Documents</a> |
            <a href="/app/review/accounting">Accounting</a> |
            <a href="/app/review/approvals">Approvals</a> |
            <a href="/app/review/proposals">Proposals</a> |
            <a href="/app/review/inventory">Inventory</a> |
            <a href="/app/review/work-orders">Work orders</a> |
            <a href="/app/review/audit">Audit</a>
          </div>
        </div>
        <form method="post" action="/app/logout" style="display:inline-grid;">
          <button type="submit" class="secondary">Sign out</button>
        </form>
      </div>
      {{end}}
    </section>

    {{if .ShowLogin}}
    <section class="panel" style="max-width: 560px;">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <h2>Sign in</h2>
      <form method="post" action="{{.LoginPath}}">
        <label>Org slug
          <input type="text" name="org_slug" autocomplete="organization" required>
        </label>
        <label>User email
          <input type="email" name="email" autocomplete="email" required>
        </label>
        <label>Device label
          <input type="text" name="device_label" value="browser">
        </label>
        <button type="submit">Start browser session</button>
      </form>
    </section>
    {{end}}

    {{with .Dashboard}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}

      <section class="panel">
        <div class="split">
          <div>
            <h2>Submit inbound request</h2>
            <form method="post" action="/app/inbound-requests" enctype="multipart/form-data">
              <label>Submitter label
                <input type="text" name="submitter_label" placeholder="front desk">
              </label>
              <label>Request message
                <textarea name="message_text" required placeholder="Describe the request, evidence, and expected follow-up."></textarea>
              </label>
              <label>Attachments
                <input type="file" name="attachments" multiple>
              </label>
              <button type="submit">Queue inbound request</button>
            </form>
          </div>
          <div>
            <h2>Agent queue</h2>
            <p class="meta">Process the next queued request through the provider-backed coordinator on the same backend seam used by the API.</p>
            <form method="post" action="/app/agent/process-next-queued-inbound-request">
              <button type="submit">Process next queued request</button>
            </form>
          </div>
        </div>
      </section>

      <section class="panel">
        <h2>Inbound request status summary</h2>
        <div class="summary-list">
          {{range .InboundSummary}}
          <div class="summary-card">
            <strong>{{.RequestCount}}</strong>
            <span class="status-pill {{statusClass .Status}}">{{.Status}}</span>
            <div class="meta">Messages: {{.MessageCount}} | Attachments: {{.AttachmentCount}}</div>
            <div class="meta">Updated: {{formatTime .LatestUpdatedAt}}</div>
          </div>
          {{else}}
          <div class="summary-card">No inbound requests yet.</div>
          {{end}}
        </div>
      </section>

      <div class="grid">
        <section class="panel">
          <h2>Recent inbound requests</h2>
          <p class="meta"><a href="/app/review/inbound-requests">Open full inbound-request review</a></p>
          <table>
            <thead>
              <tr>
                <th>Reference</th>
                <th>Status</th>
                <th>Channel</th>
                <th>Messages</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              {{range .InboundRequests}}
              <tr>
                <td><a href="/app/inbound-requests/{{.RequestReference}}">{{.RequestReference}}</a></td>
                <td><span class="status-pill {{statusClass .Status}}">{{.Status}}</span></td>
                <td>{{.Channel}}</td>
                <td>{{.MessageCount}} messages / {{.AttachmentCount}} files</td>
                <td>{{formatTime .UpdatedAt}}</td>
              </tr>
              {{else}}
              <tr><td colspan="5">No inbound requests available.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>

        <section class="panel">
          <h2>Pending approvals</h2>
          <p class="meta"><a href="/app/review/approvals?status=pending">Open full approval review</a></p>
          <table>
            <thead>
              <tr>
                <th>Queue</th>
                <th>Document</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {{range .Approvals}}
              <tr>
                <td>{{.QueueCode}}</td>
                <td><a href="{{documentReviewHref .DocumentID}}">{{.DocumentTitle}}</a></td>
                <td>
                  <div class="status-pill {{statusClass .ApprovalStatus}}">{{.ApprovalStatus}}</div>
                  <form method="post" action="/app/approvals/{{.ApprovalID}}/decision" style="margin-top:8px;">
                    <input type="hidden" name="return_to" value="/app">
                    <input type="text" name="decision_note" placeholder="Decision note">
                    <div class="inline-form">
                      <button type="submit" name="decision" value="approved">Approve</button>
                      <button type="submit" name="decision" value="rejected" class="secondary">Reject</button>
                    </div>
                  </form>
                </td>
              </tr>
              {{else}}
              <tr><td colspan="3">No pending approvals.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
      </div>

      <section class="panel">
        <h2>Processed proposals</h2>
        <p class="meta"><a href="/app/review/proposals">Open full proposal review</a></p>
        <table>
          <thead>
            <tr>
              <th>Request</th>
              <th>Recommendation</th>
              <th>Approval</th>
              <th>Document</th>
            </tr>
          </thead>
          <tbody>
            {{range .Proposals}}
            <tr>
              <td><a href="/app/inbound-requests/{{.RequestReference}}">{{.RequestReference}}</a></td>
              <td>
                <span class="status-pill {{statusClass .RecommendationStatus}}">{{.RecommendationStatus}}</span>
                <div>{{.Summary}}</div>
              </td>
              <td>{{.ApprovalStatus.String}}</td>
              <td>{{if .DocumentID.Valid}}<a href="{{documentReviewHref .DocumentID.String}}">{{.DocumentTitle.String}}</a>{{else}}-{{end}}</td>
            </tr>
            {{else}}
            <tr><td colspan="4">No processed proposals available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .InboundRequests}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}

      <section class="panel">
        <h2>Inbound-request review</h2>
        <form method="get" action="/app/review/inbound-requests" class="inline-form">
          <input type="text" name="status" value="{{.Status}}" placeholder="status">
          <input type="text" name="request_reference" value="{{.RequestReference}}" placeholder="REQ-... reference">
          <button type="submit">Filter requests</button>
        </form>
      </section>

      <section class="panel">
        <h2>Request status summary</h2>
        <div class="summary-list">
          {{range .StatusSummary}}
          <div class="summary-card">
            <strong>{{.RequestCount}}</strong>
            <span class="status-pill {{statusClass .Status}}">{{.Status}}</span>
            <div class="meta">Messages: {{.MessageCount}} | Attachments: {{.AttachmentCount}}</div>
            <div class="meta">Updated: {{formatTime .LatestUpdatedAt}}</div>
            <div class="meta"><a href="/app/review/inbound-requests?status={{.Status}}">Open {{.Status}}</a></div>
          </div>
          {{else}}
          <div class="summary-card">No inbound requests yet.</div>
          {{end}}
        </div>
      </section>

      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Reference</th>
              <th>Status</th>
              <th>Channel</th>
              <th>Messages</th>
              <th>AI</th>
              <th>Updated</th>
            </tr>
          </thead>
          <tbody>
            {{range .Requests}}
            <tr>
              <td>
                <a href="/app/inbound-requests/{{.RequestReference}}">{{.RequestReference}}</a>
                <div class="meta">{{.RequestID}}</div>
              </td>
              <td>
                <span class="status-pill {{statusClass .Status}}">{{.Status}}</span>
                {{if .CancelledAt.Valid}}<div class="meta">Cancelled: {{formatTime .CancelledAt.Time}}</div>{{end}}
                {{if .FailedAt.Valid}}<div class="meta">Failed: {{formatTime .FailedAt.Time}}</div>{{end}}
              </td>
              <td>{{.Channel}}<div class="meta">{{.OriginType}}</div></td>
              <td>{{.MessageCount}} messages / {{.AttachmentCount}} files</td>
              <td>
                {{if .LastRunID.Valid}}
                <div><span class="status-pill {{statusClass .LastRunStatus.String}}">{{.LastRunStatus.String}}</span></div>
                {{else}}
                -
                {{end}}
                {{if .LastRecommendationStatus.Valid}}<div class="meta">{{.LastRecommendationStatus.String}}</div>{{end}}
              </td>
              <td>{{formatTime .UpdatedAt}}</td>
            </tr>
            {{else}}
            <tr><td colspan="6">No inbound requests available for the selected filters.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .Approvals}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Approval review</h2>
        <form method="get" action="/app/review/approvals" class="inline-form">
          <input type="text" name="approval_id" value="{{.ApprovalID}}" placeholder="approval id">
          <input type="text" name="status" value="{{.Status}}" placeholder="pending or closed">
          <input type="text" name="queue_code" value="{{.QueueCode}}" placeholder="queue code">
          <button type="submit">Filter approvals</button>
        </form>
      </section>
      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Queue</th>
              <th>Document</th>
              <th>Approval</th>
              <th>Posting</th>
            </tr>
          </thead>
          <tbody>
            {{range .Approvals}}
            <tr>
              <td>
                <a href="{{approvalQueueHref .QueueCode .QueueStatus}}">{{.QueueCode}}</a>
                <div class="meta">Enqueued: {{formatTime .EnqueuedAt}}</div>
                <div class="meta"><span class="status-pill {{statusClass .QueueStatus}}">{{.QueueStatus}}</span></div>
              </td>
              <td>
                <a href="{{documentReviewHref .DocumentID}}">{{.DocumentTitle}}</a>
                <div class="meta">{{.DocumentTypeCode}} | <span class="status-pill {{statusClass .DocumentStatus}}">{{.DocumentStatus}}</span></div>
                <div class="meta"><a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.DocumentID}}">Audit trail</a></div>
              </td>
              <td>
                <div class="status-pill {{statusClass .ApprovalStatus}}">{{.ApprovalStatus}}</div>
                <div class="meta">Requested: {{formatTime .RequestedAt}}</div>
                {{if eq .QueueStatus "pending"}}
                <form method="post" action="/app/approvals/{{.ApprovalID}}/decision" style="margin-top:8px;">
                  <input type="hidden" name="return_to" value="{{approvalQueueHref $.Approvals.QueueCode $.Approvals.Status}}">
                  <input type="text" name="decision_note" placeholder="Decision note">
                  <div class="inline-form">
                    <button type="submit" name="decision" value="approved">Approve</button>
                    <button type="submit" name="decision" value="rejected" class="secondary">Reject</button>
                  </div>
                </form>
                {{else}}
                <div class="meta">Closed: {{if .ClosedAt.Valid}}{{formatTime .ClosedAt.Time}}{{else}}-{{end}}</div>
                {{end}}
              </td>
              <td>
                {{if .JournalEntryNumber.Valid}}
                <a href="/app/review/accounting?document_id={{.DocumentID}}">Entry #{{.JournalEntryNumber.Int64}}</a>
                <div class="meta">{{if .JournalEntryPostedAt.Valid}}{{formatTime .JournalEntryPostedAt.Time}}{{else}}Not posted{{end}}</div>
                {{else}}
                -
                {{end}}
              </td>
            </tr>
            {{else}}
            <tr><td colspan="4">No approval queue rows available for the selected filters.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .ApprovalDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Approval {{.Entry.ApprovalID}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Entry.ApprovalStatus}}">{{.Entry.ApprovalStatus}}</span>
          <p><strong>{{.Entry.DocumentTitle}}</strong></p>
          <p class="meta">{{.Entry.QueueCode}} | queue {{.Entry.QueueStatus}} | requested {{formatTime .Entry.RequestedAt}}</p>
          <p class="meta">
            <a href="{{approvalQueueHref .Entry.QueueCode .Entry.QueueStatus}}">Filtered queue view</a> |
            <a href="{{documentReviewHref .Entry.DocumentID}}">Open document</a> |
            <a href="/app/review/audit?entity_type=workflow.approval&amp;entity_id={{.Entry.ApprovalID}}">Audit trail</a>
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Decision</h2>
          <table>
            <tbody>
              <tr><th>Queue status</th><td><span class="status-pill {{statusClass .Entry.QueueStatus}}">{{.Entry.QueueStatus}}</span></td></tr>
              <tr><th>Requested by</th><td>{{.Entry.RequestedByUserID}}</td></tr>
              <tr><th>Requested at</th><td>{{formatTime .Entry.RequestedAt}}</td></tr>
              <tr><th>Decided by</th><td>{{if .Entry.DecidedByUserID.Valid}}{{.Entry.DecidedByUserID.String}}{{else}}-{{end}}</td></tr>
              <tr><th>Decided at</th><td>{{if .Entry.DecidedAt.Valid}}{{formatTime .Entry.DecidedAt.Time}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
          {{if eq .Entry.QueueStatus "pending"}}
          <form method="post" action="/app/approvals/{{.Entry.ApprovalID}}/decision" style="margin-top:12px;">
            <input type="hidden" name="return_to" value="{{approvalReviewHref .Entry.ApprovalID}}">
            <input type="text" name="decision_note" placeholder="Decision note">
            <div class="inline-form">
              <button type="submit" name="decision" value="approved">Approve</button>
              <button type="submit" name="decision" value="rejected" class="secondary">Reject</button>
            </div>
          </form>
          {{end}}
        </section>
        <section class="panel">
          <h2>Linked record</h2>
          <table>
            <tbody>
              <tr><th>Document</th><td><a href="{{documentReviewHref .Entry.DocumentID}}">{{.Entry.DocumentTitle}}</a></td></tr>
              <tr><th>Type</th><td>{{.Entry.DocumentTypeCode}}</td></tr>
              <tr><th>Status</th><td><span class="status-pill {{statusClass .Entry.DocumentStatus}}">{{.Entry.DocumentStatus}}</span></td></tr>
              <tr><th>Document number</th><td>{{if .Entry.DocumentNumber.Valid}}{{.Entry.DocumentNumber.String}}{{else}}-{{end}}</td></tr>
              <tr><th>Posting</th><td>{{if .Entry.JournalEntryNumber.Valid}}<a href="/app/review/accounting?document_id={{.Entry.DocumentID}}">Entry #{{.Entry.JournalEntryNumber.Int64}}</a>{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
      </div>
    </div>
    {{end}}

    {{with .Proposals}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Proposal review</h2>
        <form method="get" action="/app/review/proposals" class="inline-form">
          <input type="text" name="recommendation_id" value="{{.RecommendationID}}" placeholder="recommendation id">
          <input type="text" name="status" value="{{.Status}}" placeholder="recommendation status">
          <input type="text" name="request_reference" value="{{.RequestReference}}" placeholder="REQ-... reference">
          <button type="submit">Filter proposals</button>
        </form>
      </section>
      <section class="panel">
        <h2>Proposal status summary</h2>
        <div class="summary-list">
          {{range .StatusSummary}}
          <div class="summary-card">
            <strong>{{.ProposalCount}}</strong>
            <span class="status-pill {{statusClass .RecommendationStatus}}">{{.RecommendationStatus}}</span>
            <div class="meta">Requests: {{.RequestCount}} | Documents: {{.DocumentCount}}</div>
            <div class="meta">Updated: {{formatTime .LatestCreatedAt}}</div>
          </div>
          {{else}}
          <div class="summary-card">No processed proposals yet.</div>
          {{end}}
        </div>
      </section>
      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Request</th>
              <th>Recommendation</th>
              <th>Approval</th>
              <th>Document</th>
            </tr>
          </thead>
          <tbody>
            {{range .ProcessedProposals}}
            <tr>
              <td>
                <a href="/app/inbound-requests/{{.RequestReference}}">{{.RequestReference}}</a>
                <div class="meta"><span class="status-pill {{statusClass .RequestStatus}}">{{.RequestStatus}}</span></div>
              </td>
              <td>
                <span class="status-pill {{statusClass .RecommendationStatus}}">{{.RecommendationStatus}}</span>
                <div>{{.Summary}}</div>
                <div class="meta">Created: {{formatTime .CreatedAt}}</div>
                <div class="meta"><a href="{{proposalDetailHref .RecommendationID}}">Open exact proposal</a></div>
              </td>
              <td>
                {{if .ApprovalID.Valid}}
                <div><a href="{{approvalQueueHref .ApprovalQueueCode.String .ApprovalStatus.String}}">{{.ApprovalQueueCode.String}}</a></div>
                <div class="status-pill {{statusClass .ApprovalStatus.String}}">{{.ApprovalStatus.String}}</div>
                <div class="meta"><a href="{{approvalReviewHref .ApprovalID.String}}">Open exact approval</a></div>
                {{else}}
                -
                {{end}}
              </td>
              <td>
                {{if .DocumentID.Valid}}
                <a href="{{documentReviewHref .DocumentID.String}}">{{.DocumentTitle.String}}</a>
                <div class="meta">{{.DocumentTypeCode.String}} | <span class="status-pill {{statusClass .DocumentStatus.String}}">{{.DocumentStatus.String}}</span></div>
                {{else}}
                -
                {{end}}
              </td>
            </tr>
            {{else}}
            <tr><td colspan="4">No processed proposals available for the selected filters.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .ProposalDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Proposal {{.Review.RecommendationID}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Review.RecommendationStatus}}">{{.Review.RecommendationStatus}}</span>
          <p><strong>{{.Review.Summary}}</strong></p>
          <p class="meta">Request <a href="/app/inbound-requests/{{.Review.RequestReference}}">{{.Review.RequestReference}}</a> | run {{.Review.RunID}} | created {{formatTime .Review.CreatedAt}}</p>
          <p class="meta">
            <a href="{{proposalReviewHref .Review.RecommendationID .Review.RecommendationStatus .Review.RequestReference}}">Filtered proposal view</a> |
            <a href="/app/review/audit?entity_type=ai.agent_recommendation&amp;entity_id={{.Review.RecommendationID}}">Audit trail</a>
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Control chain</h2>
          <table>
            <tbody>
              <tr><th>Request</th><td><a href="/app/inbound-requests/{{.Review.RequestReference}}">{{.Review.RequestReference}}</a> | <span class="status-pill {{statusClass .Review.RequestStatus}}">{{.Review.RequestStatus}}</span></td></tr>
              <tr><th>Approval</th><td>{{if .Review.ApprovalID.Valid}}<a href="{{approvalReviewHref .Review.ApprovalID.String}}">{{if .Review.ApprovalQueueCode.Valid}}{{.Review.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{if .Review.ApprovalStatus.Valid}} | <span class="status-pill {{statusClass .Review.ApprovalStatus.String}}">{{.Review.ApprovalStatus.String}}</span>{{end}}{{else}}-{{end}}</td></tr>
              <tr><th>Document</th><td>{{if .Review.DocumentID.Valid}}<a href="{{documentReviewHref .Review.DocumentID.String}}">{{.Review.DocumentTitle.String}}</a>{{if .Review.DocumentStatus.Valid}} | <span class="status-pill {{statusClass .Review.DocumentStatus.String}}">{{.Review.DocumentStatus.String}}</span>{{end}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Identifiers</h2>
          <table>
            <tbody>
              <tr><th>Recommendation</th><td>{{.Review.RecommendationID}}</td></tr>
              <tr><th>Run</th><td>{{.Review.RunID}}</td></tr>
              <tr><th>Type</th><td>{{.Review.RecommendationType}}</td></tr>
              <tr><th>Document number</th><td>{{if .Review.DocumentNumber.Valid}}{{.Review.DocumentNumber.String}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
      </div>
    </div>
    {{end}}

    {{with .Documents}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Document review</h2>
        <form method="get" action="/app/review/documents" class="inline-form">
          <input type="text" name="document_id" value="{{.DocumentID}}" placeholder="document id">
          <input type="text" name="type_code" value="{{.TypeCode}}" placeholder="type code">
          <input type="text" name="status" value="{{.Status}}" placeholder="status">
          <button type="submit">Filter documents</button>
        </form>
      </section>
      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Type</th>
              <th>Title</th>
              <th>Status</th>
              <th>Approval</th>
              <th>Posting</th>
            </tr>
          </thead>
          <tbody>
            {{range .Documents}}
            <tr>
              <td>{{.TypeCode}}</td>
              <td>
                <strong><a href="{{documentReviewHref .DocumentID}}">{{.Title}}</a></strong>
                <div class="meta">{{.DocumentID}}</div>
                <div class="meta">
                  <a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.DocumentID}}">Audit trail</a>
                  {{if eq .TypeCode "work_order"}} | <a href="/app/review/work-orders?document_id={{.DocumentID}}">Execution review</a>{{end}}
                  {{if or (eq .TypeCode "inventory_receipt") (eq .TypeCode "inventory_issue") (eq .TypeCode "inventory_adjustment")}} | <a href="/app/review/inventory?document_id={{.DocumentID}}">Inventory review</a>{{end}}
                </div>
              </td>
              <td><span class="status-pill {{statusClass .Status}}">{{.Status}}</span></td>
              <td>{{if .ApprovalQueueCode.Valid}}<a href="{{approvalQueueHref .ApprovalQueueCode.String .ApprovalStatus.String}}">{{.ApprovalStatus.String}}</a>{{else}}{{.ApprovalStatus.String}}{{end}}</td>
              <td>{{if .JournalEntryNumber.Valid}}<a href="/app/review/accounting?document_id={{.DocumentID}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else}}-{{end}}</td>
            </tr>
            {{else}}
            <tr><td colspan="5">No documents available for the selected filters.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .DocumentDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Document {{if .Review.NumberValue.Valid}}{{.Review.NumberValue.String}}{{else}}{{.Review.TypeCode}}{{end}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Review.Status}}">{{.Review.Status}}</span>
          <p><strong>{{.Review.Title}}</strong></p>
          <p class="meta">{{.Review.DocumentID}}</p>
          <p class="meta">Type: {{.Review.TypeCode}} | Created: {{formatTime .Review.CreatedAt}} | Updated: {{formatTime .Review.UpdatedAt}}</p>
          <p class="meta">
            <a href="/app/review/documents?document_id={{.Review.DocumentID}}">Filtered list view</a> |
            <a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.Review.DocumentID}}">Audit trail</a>
            {{if .Review.SourceDocumentID.Valid}} | <a href="{{documentReviewHref .Review.SourceDocumentID.String}}">Source document</a>{{end}}
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Control chain</h2>
          <table>
            <tbody>
              <tr>
                <th>Approval</th>
                <td>
                  {{if .Review.ApprovalID.Valid}}
                  <a href="{{approvalReviewHref .Review.ApprovalID.String}}">{{if .Review.ApprovalQueueCode.Valid}}{{.Review.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>
                  {{if .Review.ApprovalStatus.Valid}} | <span class="status-pill {{statusClass .Review.ApprovalStatus.String}}">{{.Review.ApprovalStatus.String}}</span>{{end}}
                  {{else}}
                  -
                  {{end}}
                </td>
              </tr>
              <tr>
                <th>Accounting</th>
                <td>
                  {{if .Review.JournalEntryNumber.Valid}}
                  <a href="/app/review/accounting?document_id={{.Review.DocumentID}}">Entry #{{.Review.JournalEntryNumber.Int64}}</a>
                  {{if .Review.JournalEntryPostedAt.Valid}} | {{formatTime .Review.JournalEntryPostedAt.Time}}{{end}}
                  {{else}}
                  -
                  {{end}}
                </td>
              </tr>
              <tr>
                <th>Execution</th>
                <td>{{if eq .Review.TypeCode "work_order"}}<a href="/app/review/work-orders?document_id={{.Review.DocumentID}}">Work-order review</a>{{else}}-{{end}}</td>
              </tr>
              <tr>
                <th>Inventory</th>
                <td>{{if or (eq .Review.TypeCode "inventory_receipt") (eq .Review.TypeCode "inventory_issue") (eq .Review.TypeCode "inventory_adjustment")}}<a href="/app/review/inventory?document_id={{.Review.DocumentID}}">Inventory review</a>{{else}}-{{end}}</td>
              </tr>
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Lifecycle</h2>
          <table>
            <tbody>
              <tr><th>Created by</th><td>{{.Review.CreatedByUserID}}</td></tr>
              <tr><th>Submitted by</th><td>{{if .Review.SubmittedByUserID.Valid}}{{.Review.SubmittedByUserID.String}}{{else}}-{{end}}</td></tr>
              <tr><th>Submitted at</th><td>{{if .Review.SubmittedAt.Valid}}{{formatTime .Review.SubmittedAt.Time}}{{else}}-{{end}}</td></tr>
              <tr><th>Approved at</th><td>{{if .Review.ApprovedAt.Valid}}{{formatTime .Review.ApprovedAt.Time}}{{else}}-{{end}}</td></tr>
              <tr><th>Rejected at</th><td>{{if .Review.RejectedAt.Valid}}{{formatTime .Review.RejectedAt.Time}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
      </div>
    </div>
    {{end}}

    {{with .Accounting}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Accounting review</h2>
        <form method="get" action="/app/review/accounting" class="inline-form">
          <input type="date" name="start_on" value="{{.StartOn}}">
          <input type="date" name="end_on" value="{{.EndOn}}">
          <input type="date" name="as_of" value="{{.AsOf}}">
          <input type="text" name="document_id" value="{{.DocumentID}}" placeholder="source document id">
          <button type="submit">Apply filters</button>
        </form>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Journal entries</h2>
          <table>
            <thead>
              <tr>
                <th>Entry</th>
                <th>Scope</th>
                <th>Summary</th>
                <th>Totals</th>
              </tr>
            </thead>
            <tbody>
              {{range .JournalEntries}}
              <tr>
                <td>
                  #{{.EntryNumber}}
                  <div class="meta">{{.EntryKind}} | {{formatTime .PostedAt}}</div>
                </td>
                <td>{{.TaxScopeCode}}</td>
                <td>
                  {{.Summary}}
                  {{if .SourceDocumentID.Valid}}
                  <div class="meta">
                    <a href="{{documentReviewHref .SourceDocumentID.String}}">Source document</a>
                    {{if .DocumentStatus.Valid}} | <span class="status-pill {{statusClass .DocumentStatus.String}}">{{.DocumentStatus.String}}</span>{{end}}
                  </div>
                  <div class="meta"><a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.SourceDocumentID.String}}">Document audit</a></div>
                  {{end}}
                </td>
                <td>Dr {{.TotalDebitMinor}} / Cr {{.TotalCreditMinor}}</td>
              </tr>
              {{else}}
              <tr><td colspan="4">No journal entries available.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Control accounts</h2>
          <table>
            <thead>
              <tr>
                <th>Code</th>
                <th>Type</th>
                <th>Net</th>
              </tr>
            </thead>
            <tbody>
              {{range .ControlBalances}}
              <tr>
                <td>{{.AccountCode}}</td>
                <td>{{.ControlType}}</td>
                <td>{{.NetMinor}}</td>
              </tr>
              {{else}}
              <tr><td colspan="3">No control accounts available.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
      </div>
      <section class="panel">
        <h2>Tax summaries</h2>
        <table>
          <thead>
            <tr>
              <th>Tax code</th>
              <th>Type</th>
              <th>Entries</th>
              <th>Net</th>
            </tr>
          </thead>
          <tbody>
            {{range .TaxSummaries}}
            <tr>
              <td>{{.TaxCode}}</td>
              <td>{{.TaxType}}</td>
              <td>{{.EntryCount}}</td>
              <td>{{.NetMinor}}</td>
            </tr>
            {{else}}
            <tr><td colspan="4">No tax summaries available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .Inventory}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Inventory review</h2>
        <form method="get" action="/app/review/inventory" class="inline-form">
          <input type="text" name="movement_id" value="{{.MovementID}}" placeholder="movement id">
          <input type="text" name="item_id" value="{{.ItemID}}" placeholder="item id">
          <input type="text" name="location_id" value="{{.LocationID}}" placeholder="location id">
          <input type="text" name="document_id" value="{{.DocumentID}}" placeholder="document id">
          <input type="text" name="movement_type" value="{{.MovementType}}" placeholder="movement type">
          <label><input type="checkbox" name="only_pending_accounting" value="true" {{if .OnlyPendingAccounting}}checked{{end}}> pending accounting</label>
          <label><input type="checkbox" name="only_pending_execution" value="true" {{if .OnlyPendingExecution}}checked{{end}}> pending execution</label>
          <button type="submit">Apply filters</button>
        </form>
      </section>
      <section class="panel">
        <h2>Stock balances</h2>
        <table>
          <thead>
            <tr>
              <th>Item</th>
              <th>Role</th>
              <th>Location</th>
              <th>On hand</th>
            </tr>
          </thead>
          <tbody>
            {{range .Stock}}
            <tr>
              <td>{{.ItemSKU}} | {{.ItemName}}</td>
              <td>{{.ItemRole}}</td>
              <td>{{.LocationCode}} | {{.LocationName}}</td>
              <td>{{.OnHandMilli}}</td>
            </tr>
            {{else}}
            <tr><td colspan="4">No stock balances available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
      <section class="panel">
        <h2>Movement history</h2>
        <table>
          <thead>
            <tr>
              <th>Movement</th>
              <th>Item</th>
              <th>Route</th>
              <th>Quantity</th>
            </tr>
          </thead>
          <tbody>
            {{range .Movements}}
            <tr>
              <td>
                #{{.MovementNumber}} | {{.MovementType}}
                <div class="meta"><a href="/app/review/audit?entity_type=inventory_ops.movement&amp;entity_id={{.MovementID}}">Audit trail</a></div>
              </td>
              <td>
                {{.ItemSKU}} | {{.ItemName}}
                <div class="meta"><a href="/app/review/inventory?item_id={{.ItemID}}">Filter by item</a></div>
              </td>
              <td>{{if .SourceLocationCode.Valid}}{{.SourceLocationCode.String}}{{else}}-{{end}} -> {{if .DestinationLocationCode.Valid}}{{.DestinationLocationCode.String}}{{else}}-{{end}}</td>
              <td>
                {{.QuantityMilli}}
                {{if .DocumentID.Valid}}<div class="meta"><a href="{{documentReviewHref .DocumentID.String}}">{{.DocumentTitle.String}}</a></div>{{end}}
              </td>
            </tr>
            {{else}}
            <tr><td colspan="4">No inventory movements available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
      <section class="panel">
        <h2>Reconciliation</h2>
        <table>
          <thead>
            <tr>
              <th>Document</th>
              <th>Item</th>
              <th>Execution</th>
              <th>Accounting</th>
            </tr>
          </thead>
          <tbody>
            {{range .Reconciliation}}
            <tr>
              <td>
                <a href="{{documentReviewHref .DocumentID}}">{{.DocumentTitle}}</a> line {{.LineNumber}}
                <div class="meta"><a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.DocumentID}}">Audit trail</a></div>
              </td>
              <td>{{.ItemSKU}} | {{.ItemName}}</td>
              <td>{{if .WorkOrderID.Valid}}<a href="/app/review/work-orders/{{.WorkOrderID.String}}">{{.WorkOrderCode.String}}</a>{{else}}-{{end}} / {{if .ExecutionLinkStatus.Valid}}{{.ExecutionLinkStatus.String}}{{else}}-{{end}}</td>
              <td>{{if .JournalEntryNumber.Valid}}<a href="/app/review/accounting?document_id={{.DocumentID}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else}}-{{end}} / {{if .AccountingHandoffStatus.Valid}}{{.AccountingHandoffStatus.String}}{{else}}-{{end}}</td>
            </tr>
            {{else}}
            <tr><td colspan="4">No reconciliation rows available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .WorkOrders}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Work-order review</h2>
        <form method="get" action="/app/review/work-orders" class="inline-form">
          <input type="text" name="status" value="{{.Status}}" placeholder="status">
          <input type="text" name="document_id" value="{{.DocumentID}}" placeholder="document id">
          <button type="submit">Filter work orders</button>
        </form>
      </section>
      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Code</th>
              <th>Status</th>
              <th>Tasks</th>
              <th>Labor</th>
              <th>Material</th>
            </tr>
          </thead>
          <tbody>
            {{range .WorkOrders}}
            <tr>
              <td>
                <a href="/app/review/work-orders/{{.WorkOrderID}}">{{.WorkOrderCode}}</a>
                <div>{{.Title}}</div>
                <div class="meta">
                  <a href="{{documentReviewHref .DocumentID}}">Source document</a> |
                  <a href="/app/review/audit?entity_type=work_orders.work_order&amp;entity_id={{.WorkOrderID}}">Audit trail</a>
                </div>
              </td>
              <td><span class="status-pill {{statusClass .Status}}">{{.Status}}</span></td>
              <td>{{.OpenTaskCount}} open / {{.CompletedTaskCount}} done</td>
              <td>{{.LaborEntryCount}} entries / {{.TotalLaborMinutes}} min</td>
              <td>{{.MaterialUsageCount}} usages / {{.MaterialQuantityMilli}}</td>
            </tr>
            {{else}}
            <tr><td colspan="5">No work orders available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .WorkOrderDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Work order {{.Review.WorkOrderCode}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Review.Status}}">{{.Review.Status}}</span>
          <p>{{.Review.Title}}</p>
          <p class="meta">{{.Review.Summary}}</p>
          <p class="meta">
            <a href="{{documentReviewHref .Review.DocumentID}}">Source document</a> |
            <a href="/app/review/audit?entity_type=work_orders.work_order&amp;entity_id={{.Review.WorkOrderID}}">Audit trail</a>
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Execution rollup</h2>
          <div class="detail-block">Tasks: {{.Review.OpenTaskCount}} open / {{.Review.CompletedTaskCount}} completed</div>
          <div class="detail-block">Labor: {{.Review.LaborEntryCount}} entries / {{.Review.TotalLaborMinutes}} minutes / {{.Review.TotalLaborCostMinor}} minor</div>
          <div class="detail-block">Material: {{.Review.MaterialUsageCount}} usages / {{.Review.MaterialQuantityMilli}} milli / {{.Review.PostedMaterialCostMinor}} posted cost</div>
        </section>
        <section class="panel">
          <h2>Accounting linkage</h2>
          <div class="detail-block">Document status: {{.Review.DocumentStatus}}</div>
          <div class="detail-block">Posted labor entries: {{.Review.PostedLaborEntryCount}} / {{.Review.PostedLaborCostMinor}}</div>
          <div class="detail-block">Posted material usages: {{.Review.PostedMaterialUsageCount}}</div>
          <div class="detail-block">Last accounting post: {{if .Review.LastAccountingPostedAt.Valid}}<a href="/app/review/accounting?document_id={{.Review.DocumentID}}">{{formatTime .Review.LastAccountingPostedAt.Time}}</a>{{else}}-{{end}}</div>
        </section>
      </div>
    </div>
    {{end}}

    {{with .Audit}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Audit lookup</h2>
        <form method="get" action="/app/review/audit" class="inline-form">
          <input type="text" name="entity_type" value="{{.EntityType}}" placeholder="entity type">
          <input type="text" name="entity_id" value="{{.EntityID}}" placeholder="entity id">
          <button type="submit">Search audit</button>
        </form>
      </section>
      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Occurred</th>
              <th>Event</th>
              <th>Entity</th>
              <th>Payload</th>
            </tr>
          </thead>
          <tbody>
            {{range .Events}}
            <tr>
              <td>{{formatTime .OccurredAt}}</td>
              <td>{{.EventType}}</td>
              <td>
                {{.EntityType}} / {{.EntityID}}
                {{if auditEntityHref .EntityType .EntityID}}
                <div class="meta"><a href="{{auditEntityHref .EntityType .EntityID}}">{{auditEntityLabel .EntityType}}</a></div>
                {{end}}
              </td>
              <td><pre>{{prettyJSON .Payload}}</pre></td>
            </tr>
            {{else}}
            <tr><td colspan="4">No audit events available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .Detail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}

      <section class="panel">
        <h2>Inbound request {{.Detail.Request.RequestReference}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Detail.Request.Status}}">{{.Detail.Request.Status}}</span>
          <p class="meta">Channel: {{.Detail.Request.Channel}} | Origin: {{.Detail.Request.OriginType}} | Received: {{formatTime .Detail.Request.ReceivedAt}}</p>
        </div>
        <div class="detail-block">
          <h3>Metadata</h3>
          <pre>{{prettyJSON .Detail.Request.Metadata}}</pre>
        </div>
      </section>

      <div class="grid">
        <section class="panel">
          <h2>Messages</h2>
          {{range .Detail.Messages}}
          <div class="detail-block">
            <strong>#{{.MessageIndex}} {{.MessageRole}}</strong>
            <p>{{.TextContent}}</p>
            <div class="meta">{{formatTime .CreatedAt}}</div>
          </div>
          {{else}}
          <p>No messages.</p>
          {{end}}
        </section>

        <section class="panel">
          <h2>Attachments</h2>
          {{range .Detail.Attachments}}
          <div class="detail-block">
            <div><a href="/api/attachments/{{.AttachmentID}}/content">{{.OriginalFileName}}</a></div>
            <div class="meta">{{.MediaType}} | {{.SizeBytes}} bytes | {{.LinkRole}}</div>
            {{if .LatestDerivedText.Valid}}
            <pre>{{.LatestDerivedText.String}}</pre>
            {{end}}
          </div>
          {{else}}
          <p>No attachments.</p>
          {{end}}
        </section>
      </div>

      <div class="grid">
        <section class="panel">
          <h2>AI runs</h2>
          {{range .Detail.Runs}}
          <div class="detail-block">
            <div><strong>{{.AgentRole}}</strong> / {{.CapabilityCode}}</div>
            <div class="status-pill {{statusClass .Status}}">{{.Status}}</div>
            <p>{{.Summary}}</p>
          </div>
          {{else}}
          <p>No AI runs yet.</p>
          {{end}}
        </section>

        <section class="panel">
          <h2>AI steps</h2>
          {{range .Detail.Steps}}
          <div class="detail-block">
            <strong>#{{.StepIndex}} {{.StepTitle}}</strong>
            <div class="meta">{{.StepType}} | Run {{.RunID}}</div>
            <div class="status-pill {{statusClass .Status}}">{{.Status}}</div>
            <div class="meta">Created: {{formatTime .CreatedAt}}</div>
            <details style="margin-top:10px;">
              <summary>Step payloads</summary>
              <div class="detail-block">
                <div class="meta">Input</div>
                <pre>{{prettyJSON .InputPayload}}</pre>
              </div>
              <div class="detail-block">
                <div class="meta">Output</div>
                <pre>{{prettyJSON .OutputPayload}}</pre>
              </div>
            </details>
          </div>
          {{else}}
          <p>No AI steps yet.</p>
          {{end}}
        </section>

        <section class="panel">
          <h2>Artifacts</h2>
          {{range .Detail.Artifacts}}
          <div class="detail-block">
            <strong>{{.Title}}</strong>
            <div class="meta">{{.ArtifactType}} | {{formatTime .CreatedAt}}</div>
            <pre>{{prettyJSON .Payload}}</pre>
          </div>
          {{else}}
          <p>No artifacts yet.</p>
          {{end}}
        </section>
      </div>

      <div class="grid">
        <section class="panel">
          <h2>Delegations</h2>
          {{range .Detail.Delegations}}
          <div class="detail-block">
            <strong>{{.CapabilityCode}}</strong>
            <div class="meta">Parent run: {{.ParentRunID}}</div>
            <div class="meta">Child run: {{.ChildRunID}} | {{.ChildAgentRole}} / {{.ChildCapabilityCode}}</div>
            {{if .RequestedByStepID.Valid}}<div class="meta">Requested by step: {{.RequestedByStepID.String}}</div>{{end}}
            <div class="status-pill {{statusClass .ChildRunStatus}}">{{.ChildRunStatus}}</div>
            <p>{{.Reason}}</p>
            <div class="meta">Created: {{formatTime .CreatedAt}}</div>
          </div>
          {{else}}
          <p>No delegations yet.</p>
          {{end}}
        </section>

        <section class="panel">
          <h2>Recommendations</h2>
          {{range .Detail.Recommendations}}
          <div class="detail-block">
            <strong>{{.Summary}}</strong>
            <div class="status-pill {{statusClass .Status}}">{{.Status}}</div>
            <pre>{{prettyJSON .Payload}}</pre>
          </div>
          {{else}}
          <p>No recommendations yet.</p>
          {{end}}
        </section>

        <section class="panel">
          <h2>Proposals</h2>
          {{range .Detail.Proposals}}
          <div class="detail-block">
            <strong>{{.Summary}}</strong>
            <div class="meta">Recommendation: {{.RecommendationStatus}} | Approval: {{.ApprovalStatus.String}}</div>
            <div class="meta">Document: {{if .DocumentID.Valid}}<a href="{{documentReviewHref .DocumentID.String}}">{{.DocumentTitle.String}}</a>{{else}}{{.DocumentTitle.String}}{{end}}</div>
            {{if .ApprovalID.Valid}}
            <form method="post" action="/app/approvals/{{.ApprovalID.String}}/decision">
              <input type="hidden" name="return_to" value="/app/inbound-requests/{{$.Detail.Request.RequestReference}}">
              <input type="text" name="decision_note" placeholder="Decision note">
              <div class="inline-form">
                <button type="submit" name="decision" value="approved">Approve</button>
                <button type="submit" name="decision" value="rejected" class="secondary">Reject</button>
              </div>
            </form>
            {{end}}
          </div>
          {{else}}
          <p>No downstream proposals yet.</p>
          {{end}}
        </section>
      </div>
    </div>
    {{end}}
  </div>
</body>
</html>`
