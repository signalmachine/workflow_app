package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"workflow_app/internal/attachments"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/reporting"
	"workflow_app/internal/workflow"
)

const (
	sessionLoginPath           = "/api/session/login"
	sessionCurrentPath         = "/api/session"
	sessionLogoutPath          = "/api/session/logout"
	webAppPath                 = "/app"
	webLoginPath               = "/app/login"
	webLogoutPath              = "/app/logout"
	webSubmitInboundPath       = "/app/inbound-requests"
	webProcessNextQueuedPath   = "/app/agent/process-next-queued-inbound-request"
	webInboundDetailPrefix     = "/app/inbound-requests/"
	webApprovalDecisionPrefix  = "/app/approvals/"
	webDocumentsPath           = "/app/review/documents"
	webAccountingPath          = "/app/review/accounting"
	webInventoryPath           = "/app/review/inventory"
	webWorkOrdersPath          = "/app/review/work-orders"
	webAuditPath               = "/app/review/audit"
	agentProcessNextQueuedPath = "/api/agent/process-next-queued-inbound-request"
	submitInboundRequestPath   = "/api/inbound-requests"
	attachmentContentPrefix    = "/api/attachments/"
	reviewInboundRequestsPath  = "/api/review/inbound-requests"
	reviewInboundSummaryPath   = "/api/review/inbound-request-status-summary"
	reviewProposalListPath     = "/api/review/processed-proposals"
	reviewProposalSummaryPath  = "/api/review/processed-proposal-status-summary"
	reviewApprovalQueuePath    = "/api/review/approval-queue"
	reviewDocumentsPath        = "/api/review/documents"
	reviewJournalEntriesPath   = "/api/review/accounting/journal-entries"
	reviewControlBalancesPath  = "/api/review/accounting/control-account-balances"
	reviewTaxSummariesPath     = "/api/review/accounting/tax-summaries"
	reviewInventoryStockPath   = "/api/review/inventory/stock"
	reviewInventoryMovesPath   = "/api/review/inventory/movements"
	reviewInventoryReconPath   = "/api/review/inventory/reconciliation"
	reviewWorkOrdersPath       = "/api/review/work-orders"
	reviewAuditEventsPath      = "/api/review/audit-events"
	approvalDecisionPrefix     = "/api/approvals/"
	headerOrgID                = "X-Workflow-Org-ID"
	headerUserID               = "X-Workflow-User-ID"
	headerSessionID            = "X-Workflow-Session-ID"
	sessionIDCookieName        = "workflow_session_id"
	refreshTokenCookieName     = "workflow_refresh_token"
)

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

const browserSessionDuration = 24 * time.Hour

type ProcessNextQueuedInboundRequester interface {
	ProcessNextQueuedInboundRequest(ctx context.Context, input ProcessNextQueuedInboundRequestInput) (ProcessNextQueuedInboundRequestResult, error)
}

type queuedInboundRequestProcessorLoader func() (ProcessNextQueuedInboundRequester, error)

type inboundRequestSubmitter interface {
	SubmitInboundRequest(ctx context.Context, input SubmitInboundRequestInput) (SubmitInboundRequestResult, error)
	DownloadAttachment(ctx context.Context, input DownloadAttachmentInput) (attachments.AttachmentContent, error)
}

type operatorReviewReader interface {
	ListApprovalQueue(ctx context.Context, input reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error)
	ListDocuments(ctx context.Context, input reporting.ListDocumentsInput) ([]reporting.DocumentReview, error)
	ListJournalEntries(ctx context.Context, input reporting.ListJournalEntriesInput) ([]reporting.JournalEntryReview, error)
	ListControlAccountBalances(ctx context.Context, input reporting.ListControlAccountBalancesInput) ([]reporting.ControlAccountBalance, error)
	ListTaxSummaries(ctx context.Context, input reporting.ListTaxSummariesInput) ([]reporting.TaxSummary, error)
	ListInventoryStock(ctx context.Context, input reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error)
	ListInventoryMovements(ctx context.Context, input reporting.ListInventoryMovementsInput) ([]reporting.InventoryMovementReview, error)
	ListInventoryReconciliation(ctx context.Context, input reporting.ListInventoryReconciliationInput) ([]reporting.InventoryReconciliationItem, error)
	ListWorkOrders(ctx context.Context, input reporting.ListWorkOrdersInput) ([]reporting.WorkOrderReview, error)
	GetWorkOrderReview(ctx context.Context, input reporting.GetWorkOrderReviewInput) (reporting.WorkOrderReview, error)
	LookupAuditEvents(ctx context.Context, input reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error)
	ListInboundRequests(ctx context.Context, input reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error)
	GetInboundRequestDetail(ctx context.Context, input reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error)
	ListInboundRequestStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error)
	ListProcessedProposals(ctx context.Context, input reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error)
	ListProcessedProposalStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error)
}

type approvalDecisionService interface {
	DecideApproval(ctx context.Context, input workflow.DecideApprovalInput) (workflow.Approval, documents.Document, error)
}

type browserSessionService interface {
	StartBrowserSession(ctx context.Context, input identityaccess.StartBrowserSessionInput) (identityaccess.BrowserSession, error)
	AuthenticateSession(ctx context.Context, sessionID, refreshToken string) (identityaccess.SessionContext, error)
	RevokeAuthenticatedSession(ctx context.Context, sessionID, refreshToken string) error
}

type processNextQueuedRequest struct {
	Channel string `json:"channel"`
}

type submitInboundRequestRequest struct {
	OriginType  string                              `json:"origin_type"`
	Channel     string                              `json:"channel"`
	Metadata    map[string]any                      `json:"metadata"`
	Message     submitInboundRequestMessageRequest  `json:"message"`
	Attachments []submitInboundRequestAttachmentDTO `json:"attachments"`
}

type submitInboundRequestMessageRequest struct {
	MessageRole string `json:"message_role"`
	TextContent string `json:"text_content"`
}

type submitInboundRequestAttachmentDTO struct {
	OriginalFileName string `json:"original_file_name"`
	MediaType        string `json:"media_type"`
	ContentBase64    string `json:"content_base64"`
	LinkRole         string `json:"link_role"`
}

type processNextQueuedResponse struct {
	Processed             bool   `json:"processed"`
	RequestReference      string `json:"request_reference,omitempty"`
	RequestStatus         string `json:"request_status,omitempty"`
	RunID                 string `json:"run_id,omitempty"`
	RunStatus             string `json:"run_status,omitempty"`
	ArtifactID            string `json:"artifact_id,omitempty"`
	RecommendationID      string `json:"recommendation_id,omitempty"`
	RecommendationSummary string `json:"recommendation_summary,omitempty"`
}

type submitInboundRequestResponse struct {
	RequestID        string   `json:"request_id"`
	RequestReference string   `json:"request_reference"`
	Status           string   `json:"status"`
	MessageID        string   `json:"message_id"`
	AttachmentIDs    []string `json:"attachment_ids,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type decideApprovalRequest struct {
	Decision     string `json:"decision"`
	DecisionNote string `json:"decision_note"`
}

type sessionLoginRequest struct {
	OrgSlug     string `json:"org_slug"`
	Email       string `json:"email"`
	DeviceLabel string `json:"device_label"`
}

type AgentAPIHandler struct {
	loadProcessor     queuedInboundRequestProcessorLoader
	submissionService inboundRequestSubmitter
	reviewService     operatorReviewReader
	approvalService   approvalDecisionService
	authService       browserSessionService
}

func NewAgentAPIHandler(db *sql.DB) http.Handler {
	documentService := documents.NewService(db)
	authService := identityaccess.NewService(db)
	return NewAgentAPIHandlerWithDependencies(func() (ProcessNextQueuedInboundRequester, error) {
		return NewOpenAIAgentProcessorFromEnv(db)
	}, NewSubmissionService(db), reporting.NewService(db), workflow.NewService(db, documentService), authService)
}

func NewAgentAPIHandlerWithProcessorLoader(loader queuedInboundRequestProcessorLoader) http.Handler {
	return NewAgentAPIHandlerWithDependencies(loader, nil, nil, nil, nil)
}

func NewAgentAPIHandlerWithServices(loader queuedInboundRequestProcessorLoader, submissionService inboundRequestSubmitter) http.Handler {
	return NewAgentAPIHandlerWithDependencies(loader, submissionService, nil, nil, nil)
}

func NewAgentAPIHandlerWithDependencies(loader queuedInboundRequestProcessorLoader, submissionService inboundRequestSubmitter, reviewService operatorReviewReader, approvalService approvalDecisionService, authService browserSessionService) http.Handler {
	handler := &AgentAPIHandler{
		loadProcessor:     loader,
		submissionService: submissionService,
		reviewService:     reviewService,
		approvalService:   approvalService,
		authService:       authService,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.handleRoot)
	mux.HandleFunc(webAppPath, handler.handleWebAppDashboard)
	mux.HandleFunc(webLoginPath, handler.handleWebLogin)
	mux.HandleFunc(webLogoutPath, handler.handleWebLogout)
	mux.HandleFunc(webSubmitInboundPath, handler.handleWebSubmitInboundRequest)
	mux.HandleFunc(webProcessNextQueuedPath, handler.handleWebProcessNextQueuedInboundRequest)
	mux.HandleFunc(webInboundDetailPrefix, handler.handleWebInboundRequestDetail)
	mux.HandleFunc(webApprovalDecisionPrefix, handler.handleWebApprovalDecision)
	mux.HandleFunc(webDocumentsPath, handler.handleWebDocuments)
	mux.HandleFunc(webAccountingPath, handler.handleWebAccounting)
	mux.HandleFunc(webInventoryPath, handler.handleWebInventory)
	mux.HandleFunc(webWorkOrdersPath, handler.handleWebWorkOrders)
	mux.HandleFunc(webWorkOrdersPath+"/", handler.handleWebWorkOrderDetail)
	mux.HandleFunc(webAuditPath, handler.handleWebAudit)
	mux.HandleFunc(sessionLoginPath, handler.handleSessionLogin)
	mux.HandleFunc(sessionCurrentPath, handler.handleCurrentSession)
	mux.HandleFunc(sessionLogoutPath, handler.handleSessionLogout)
	mux.HandleFunc(agentProcessNextQueuedPath, handler.handleProcessNextQueuedInboundRequest)
	mux.HandleFunc(submitInboundRequestPath, handler.handleSubmitInboundRequest)
	mux.HandleFunc(attachmentContentPrefix, handler.handleDownloadAttachment)
	mux.HandleFunc(reviewInboundRequestsPath, handler.handleListInboundRequests)
	mux.HandleFunc(reviewInboundRequestsPath+"/", handler.handleGetInboundRequestDetail)
	mux.HandleFunc(reviewInboundSummaryPath, handler.handleListInboundRequestStatusSummary)
	mux.HandleFunc(reviewProposalListPath, handler.handleListProcessedProposals)
	mux.HandleFunc(reviewProposalSummaryPath, handler.handleListProcessedProposalStatusSummary)
	mux.HandleFunc(reviewApprovalQueuePath, handler.handleListApprovalQueue)
	mux.HandleFunc(reviewDocumentsPath, handler.handleListDocuments)
	mux.HandleFunc(reviewJournalEntriesPath, handler.handleListJournalEntries)
	mux.HandleFunc(reviewControlBalancesPath, handler.handleListControlAccountBalances)
	mux.HandleFunc(reviewTaxSummariesPath, handler.handleListTaxSummaries)
	mux.HandleFunc(reviewInventoryStockPath, handler.handleListInventoryStock)
	mux.HandleFunc(reviewInventoryMovesPath, handler.handleListInventoryMovements)
	mux.HandleFunc(reviewInventoryReconPath, handler.handleListInventoryReconciliation)
	mux.HandleFunc(reviewWorkOrdersPath, handler.handleListWorkOrders)
	mux.HandleFunc(reviewWorkOrdersPath+"/", handler.handleGetWorkOrderReview)
	mux.HandleFunc(reviewAuditEventsPath, handler.handleLookupAuditEvents)
	mux.HandleFunc(approvalDecisionPrefix, handler.handleDecideApproval)
	return mux
}

func (h *AgentAPIHandler) handleSessionLogin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != sessionLoginPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.authService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "auth service unavailable"})
		return
	}
	if r.Body == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "request body is required"})
		return
	}
	defer r.Body.Close()

	var req sessionLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON request body"})
		return
	}

	deviceLabel := strings.TrimSpace(req.DeviceLabel)
	if deviceLabel == "" {
		deviceLabel = "browser"
	}

	session, err := h.authService.StartBrowserSession(r.Context(), identityaccess.StartBrowserSessionInput{
		OrgSlug:     req.OrgSlug,
		Email:       req.Email,
		DeviceLabel: deviceLabel,
		ExpiresAt:   time.Now().UTC().Add(browserSessionDuration),
	})
	if err != nil {
		switch {
		case errors.Is(err, identityaccess.ErrUnauthorized), errors.Is(err, identityaccess.ErrMembershipMissing):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid session credentials"})
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to start session"})
		}
		return
	}

	setSessionCookies(w, session.Session.ID, session.RefreshToken, session.Session.ExpiresAt)
	writeJSON(w, http.StatusCreated, mapSessionContext(identityaccess.SessionContext{
		Actor:           identityaccess.Actor{OrgID: session.Session.OrgID, UserID: session.Session.UserID, SessionID: session.Session.ID},
		Session:         session.Session,
		RoleCode:        session.RoleCode,
		OrgSlug:         session.OrgSlug,
		OrgName:         session.OrgName,
		UserEmail:       session.UserEmail,
		UserDisplayName: session.UserDisplayName,
	}))
}

func (h *AgentAPIHandler) handleCurrentSession(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != sessionCurrentPath {
		http.NotFound(w, r)
		return
	}
	if h.authService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "auth service unavailable"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		context, err := h.sessionContextFromRequest(r)
		if err != nil {
			if errors.Is(err, identityaccess.ErrUnauthorized) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		if refreshToken := cookieValue(r, refreshTokenCookieName); refreshToken != "" {
			setSessionCookies(w, context.Session.ID, refreshToken, context.Session.ExpiresAt)
		}
		writeJSON(w, http.StatusOK, mapSessionContext(context))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (h *AgentAPIHandler) handleSessionLogout(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != sessionLogoutPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.authService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "auth service unavailable"})
		return
	}

	sessionID, refreshToken, ok := sessionCookiesFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	if err := h.authService.RevokeAuthenticatedSession(r.Context(), sessionID, refreshToken); err != nil {
		if errors.Is(err, identityaccess.ErrUnauthorized) || errors.Is(err, identityaccess.ErrSessionNotActive) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to revoke session"})
		return
	}

	clearSessionCookies(w)
	writeJSON(w, http.StatusOK, struct {
		Revoked bool `json:"revoked"`
	}{Revoked: true})
}

func (h *AgentAPIHandler) handleProcessNextQueuedInboundRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	var req processNextQueuedRequest
	if r.Body != nil {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON request body"})
			return
		}
	}

	processor, err := h.loadProcessor()
	if err != nil {
		if errors.Is(err, ErrAgentProviderNotConfigured) {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "ai provider not configured"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to initialize agent processor"})
		return
	}

	result, err := processor.ProcessNextQueuedInboundRequest(r.Context(), ProcessNextQueuedInboundRequestInput{
		Channel: strings.TrimSpace(req.Channel),
		Actor:   actor,
	})
	if err != nil {
		switch {
		case errors.Is(err, intake.ErrNoQueuedInboundRequest):
			writeJSON(w, http.StatusOK, processNextQueuedResponse{Processed: false})
		case errors.Is(err, identityaccess.ErrUnauthorized):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to process queued inbound request"})
		}
		return
	}

	writeJSON(w, http.StatusOK, processNextQueuedResponse{
		Processed:             true,
		RequestReference:      result.Request.RequestReference,
		RequestStatus:         result.Request.Status,
		RunID:                 result.Run.ID,
		RunStatus:             result.Run.Status,
		ArtifactID:            result.Artifact.ID,
		RecommendationID:      result.Recommendation.ID,
		RecommendationSummary: result.Recommendation.Summary,
	})
}

func (h *AgentAPIHandler) handleSubmitInboundRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != submitInboundRequestPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.submissionService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "submission service unavailable"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	var req submitInboundRequestRequest
	if r.Body == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "request body is required"})
		return
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON request body"})
		return
	}

	attachmentsInput := make([]SubmitInboundRequestAttachmentInput, 0, len(req.Attachments))
	for _, attachment := range req.Attachments {
		attachmentsInput = append(attachmentsInput, SubmitInboundRequestAttachmentInput{
			OriginalFileName: attachment.OriginalFileName,
			MediaType:        attachment.MediaType,
			ContentBase64:    attachment.ContentBase64,
			LinkRole:         attachment.LinkRole,
		})
	}

	result, err := h.submissionService.SubmitInboundRequest(r.Context(), SubmitInboundRequestInput{
		OriginType:     req.OriginType,
		Channel:        req.Channel,
		Metadata:       req.Metadata,
		MessageRole:    req.Message.MessageRole,
		MessageText:    req.Message.TextContent,
		Attachments:    attachmentsInput,
		QueueForReview: true,
		Actor:          actor,
	})
	if err != nil {
		switch {
		case errors.Is(err, intake.ErrInvalidInboundRequest), errors.Is(err, attachments.ErrInvalidAttachment), errors.Is(err, attachments.ErrInvalidLink), errors.Is(err, ErrAttachmentContentEncoding):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid inbound request"})
		case errors.Is(err, identityaccess.ErrUnauthorized):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to submit inbound request"})
		}
		return
	}

	response := submitInboundRequestResponse{
		RequestID:        result.Request.ID,
		RequestReference: result.Request.RequestReference,
		Status:           result.Request.Status,
		MessageID:        result.Message.ID,
	}
	for _, attachment := range result.Attachments {
		response.AttachmentIDs = append(response.AttachmentIDs, attachment.ID)
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *AgentAPIHandler) handleDownloadAttachment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if h.submissionService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "submission service unavailable"})
		return
	}

	attachmentID, ok := parseAttachmentContentPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	attachment, err := h.submissionService.DownloadAttachment(r.Context(), DownloadAttachmentInput{
		AttachmentID: attachmentID,
		Actor:        actor,
	})
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrAttachmentNotFound):
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "attachment not found"})
		case errors.Is(err, identityaccess.ErrUnauthorized):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to download attachment"})
		}
		return
	}

	fileName := attachment.OriginalFileName
	if fileName == "" {
		fileName = attachment.ID
	}
	w.Header().Set("Content-Type", attachment.MediaType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(attachment.Content)))
	w.Header().Set("Content-Disposition", contentDisposition(fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(attachment.Content)
}

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
	if strings.HasPrefix(strings.ToUpper(lookup), "REQ-") {
		input.RequestReference = lookup
	} else {
		input.RequestID = lookup
	}

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
		QueueCode: strings.TrimSpace(r.URL.Query().Get("queue_code")),
		Status:    strings.TrimSpace(r.URL.Query().Get("status")),
		Limit:     parseLimit(r.URL.Query().Get("limit")),
		Actor:     actor,
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
		StartOn: parseOptionalDate(r.URL.Query().Get("start_on")),
		EndOn:   parseOptionalDate(r.URL.Query().Get("end_on")),
		Limit:   parseLimit(r.URL.Query().Get("limit")),
		Actor:   actor,
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
		AsOf:  parseOptionalDate(r.URL.Query().Get("as_of")),
		Actor: actor,
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
		Status: strings.TrimSpace(r.URL.Query().Get("status")),
		Limit:  parseLimit(r.URL.Query().Get("limit")),
		Actor:  actor,
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

func (h *AgentAPIHandler) handleDecideApproval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if h.approvalService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "approval service unavailable"})
		return
	}

	approvalID, ok := parseApprovalDecisionPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	var req decideApprovalRequest
	if r.Body == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "request body is required"})
		return
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON request body"})
		return
	}
	req.Decision = strings.TrimSpace(req.Decision)
	if req.Decision != "approved" && req.Decision != "rejected" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid approval decision"})
		return
	}

	approval, document, err := h.approvalService.DecideApproval(r.Context(), workflow.DecideApprovalInput{
		ApprovalID:   approvalID,
		Decision:     req.Decision,
		DecisionNote: req.DecisionNote,
		Actor:        actor,
	})
	if err != nil {
		switch {
		case errors.Is(err, identityaccess.ErrUnauthorized):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		case errors.Is(err, workflow.ErrApprovalNotFound):
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "approval not found"})
		case errors.Is(err, workflow.ErrApprovalState), errors.Is(err, documents.ErrInvalidDocumentState):
			writeJSON(w, http.StatusConflict, errorResponse{Error: "approval cannot be decided in the current state"})
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to decide approval"})
		}
		return
	}

	writeJSON(w, http.StatusOK, approvalDecisionResponse{
		ApprovalID:      approval.ID,
		Status:          approval.Status,
		QueueCode:       approval.QueueCode,
		DocumentID:      approval.DocumentID,
		DocumentStatus:  string(document.Status),
		DecisionNote:    stringPtr(approval.DecisionNote),
		DecidedByUserID: stringPtr(approval.DecidedByUserID),
		DecidedAt:       timePtr(approval.DecidedAt),
	})
}

func (h *AgentAPIHandler) actorFromRequest(r *http.Request) (identityaccess.Actor, error) {
	if actor, err := actorFromHeaders(r); err == nil {
		return actor, nil
	}
	if h.authService == nil {
		return identityaccess.Actor{}, fmt.Errorf("missing required authentication headers")
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		return identityaccess.Actor{}, err
	}
	return sessionContext.Actor, nil
}

func actorFromHeaders(r *http.Request) (identityaccess.Actor, error) {
	orgID := strings.TrimSpace(r.Header.Get(headerOrgID))
	userID := strings.TrimSpace(r.Header.Get(headerUserID))
	sessionID := strings.TrimSpace(r.Header.Get(headerSessionID))
	if orgID == "" || userID == "" || sessionID == "" {
		return identityaccess.Actor{}, fmt.Errorf("missing required authentication headers")
	}
	if !uuidPattern.MatchString(orgID) || !uuidPattern.MatchString(userID) || !uuidPattern.MatchString(sessionID) {
		return identityaccess.Actor{}, fmt.Errorf("authentication headers must be UUIDs")
	}

	return identityaccess.Actor{
		OrgID:     orgID,
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}

func (h *AgentAPIHandler) sessionContextFromRequest(r *http.Request) (identityaccess.SessionContext, error) {
	sessionID, refreshToken, ok := sessionCookiesFromRequest(r)
	if !ok {
		if _, err := actorFromHeaders(r); err != nil {
			return identityaccess.SessionContext{}, identityaccess.ErrUnauthorized
		}
		return identityaccess.SessionContext{}, identityaccess.ErrUnauthorized
	}
	return h.authService.AuthenticateSession(r.Context(), sessionID, refreshToken)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func setSessionCookies(w http.ResponseWriter, sessionID, refreshToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionIDCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func clearSessionCookies(w http.ResponseWriter) {
	for _, name := range []string{sessionIDCookieName, refreshTokenCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
			Expires:  time.Unix(0, 0).UTC(),
		})
	}
}

func sessionCookiesFromRequest(r *http.Request) (string, string, bool) {
	sessionIDCookie, err := r.Cookie(sessionIDCookieName)
	if err != nil {
		return "", "", false
	}
	refreshTokenCookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		return "", "", false
	}

	sessionID := strings.TrimSpace(sessionIDCookie.Value)
	refreshToken := strings.TrimSpace(refreshTokenCookie.Value)
	if !uuidPattern.MatchString(sessionID) || refreshToken == "" {
		return "", "", false
	}
	return sessionID, refreshToken, true
}

func cookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func parseAttachmentContentPath(path string) (string, bool) {
	if !strings.HasPrefix(path, attachmentContentPrefix) || !strings.HasSuffix(path, "/content") {
		return "", false
	}
	attachmentID := strings.TrimSuffix(strings.TrimPrefix(path, attachmentContentPrefix), "/content")
	attachmentID = strings.Trim(attachmentID, "/")
	if attachmentID == "" {
		return "", false
	}
	return attachmentID, true
}

func parseApprovalDecisionPath(path string) (string, bool) {
	if !strings.HasPrefix(path, approvalDecisionPrefix) || !strings.HasSuffix(path, "/decision") {
		return "", false
	}
	approvalID := strings.TrimSuffix(strings.TrimPrefix(path, approvalDecisionPrefix), "/decision")
	approvalID = strings.Trim(approvalID, "/")
	if approvalID == "" {
		return "", false
	}
	return approvalID, true
}

func parseChildPath(prefix, path string) (string, bool) {
	if !strings.HasPrefix(path, prefix+"/") {
		return "", false
	}
	segment := strings.TrimPrefix(path, prefix+"/")
	segment = strings.Trim(segment, "/")
	if segment == "" || strings.Contains(segment, "/") {
		return "", false
	}
	unescaped, err := url.PathUnescape(segment)
	if err != nil || strings.TrimSpace(unescaped) == "" {
		return "", false
	}
	return strings.TrimSpace(unescaped), true
}

func parseLimit(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	var limit int
	if _, err := fmt.Sscanf(raw, "%d", &limit); err != nil {
		return -1
	}
	return limit
}

func parseOptionalDate(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func contentDisposition(fileName string) string {
	encoded := url.PathEscape(fileName)
	return fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", fileName, encoded)
}

type approvalQueueEntryResponse struct {
	QueueEntryID         string     `json:"queue_entry_id"`
	ApprovalID           string     `json:"approval_id"`
	QueueCode            string     `json:"queue_code"`
	QueueStatus          string     `json:"queue_status"`
	EnqueuedAt           time.Time  `json:"enqueued_at"`
	ClosedAt             *time.Time `json:"closed_at,omitempty"`
	ApprovalStatus       string     `json:"approval_status"`
	RequestedAt          time.Time  `json:"requested_at"`
	RequestedByUserID    string     `json:"requested_by_user_id"`
	DecidedAt            *time.Time `json:"decided_at,omitempty"`
	DecidedByUserID      *string    `json:"decided_by_user_id,omitempty"`
	DocumentID           string     `json:"document_id"`
	DocumentTypeCode     string     `json:"document_type_code"`
	DocumentTitle        string     `json:"document_title"`
	DocumentNumber       *string    `json:"document_number,omitempty"`
	DocumentStatus       string     `json:"document_status"`
	JournalEntryID       *string    `json:"journal_entry_id,omitempty"`
	JournalEntryNumber   *int64     `json:"journal_entry_number,omitempty"`
	JournalEntryPostedAt *time.Time `json:"journal_entry_posted_at,omitempty"`
}

type documentReviewResponse struct {
	DocumentID           string     `json:"document_id"`
	TypeCode             string     `json:"type_code"`
	Title                string     `json:"title"`
	NumberValue          *string    `json:"number_value,omitempty"`
	Status               string     `json:"status"`
	SourceDocumentID     *string    `json:"source_document_id,omitempty"`
	CreatedByUserID      string     `json:"created_by_user_id"`
	SubmittedByUserID    *string    `json:"submitted_by_user_id,omitempty"`
	SubmittedAt          *time.Time `json:"submitted_at,omitempty"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`
	RejectedAt           *time.Time `json:"rejected_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	ApprovalID           *string    `json:"approval_id,omitempty"`
	ApprovalStatus       *string    `json:"approval_status,omitempty"`
	ApprovalQueueCode    *string    `json:"approval_queue_code,omitempty"`
	ApprovalRequestedAt  *time.Time `json:"approval_requested_at,omitempty"`
	ApprovalDecidedAt    *time.Time `json:"approval_decided_at,omitempty"`
	JournalEntryID       *string    `json:"journal_entry_id,omitempty"`
	JournalEntryNumber   *int64     `json:"journal_entry_number,omitempty"`
	JournalEntryPostedAt *time.Time `json:"journal_entry_posted_at,omitempty"`
}

type journalEntryReviewResponse struct {
	EntryID           string    `json:"entry_id"`
	EntryNumber       int64     `json:"entry_number"`
	EntryKind         string    `json:"entry_kind"`
	SourceDocumentID  *string   `json:"source_document_id,omitempty"`
	ReversalOfEntryID *string   `json:"reversal_of_entry_id,omitempty"`
	CurrencyCode      string    `json:"currency_code"`
	TaxScopeCode      string    `json:"tax_scope_code"`
	Summary           string    `json:"summary"`
	ReversalReason    *string   `json:"reversal_reason,omitempty"`
	PostedByUserID    string    `json:"posted_by_user_id"`
	EffectiveOn       time.Time `json:"effective_on"`
	PostedAt          time.Time `json:"posted_at"`
	CreatedAt         time.Time `json:"created_at"`
	DocumentTypeCode  *string   `json:"document_type_code,omitempty"`
	DocumentNumber    *string   `json:"document_number,omitempty"`
	DocumentStatus    *string   `json:"document_status,omitempty"`
	LineCount         int       `json:"line_count"`
	TotalDebitMinor   int64     `json:"total_debit_minor"`
	TotalCreditMinor  int64     `json:"total_credit_minor"`
	HasReversal       bool      `json:"has_reversal"`
}

type controlAccountBalanceResponse struct {
	AccountID        string     `json:"account_id"`
	AccountCode      string     `json:"account_code"`
	AccountName      string     `json:"account_name"`
	AccountClass     string     `json:"account_class"`
	ControlType      string     `json:"control_type"`
	TotalDebitMinor  int64      `json:"total_debit_minor"`
	TotalCreditMinor int64      `json:"total_credit_minor"`
	NetMinor         int64      `json:"net_minor"`
	LastEffectiveOn  *time.Time `json:"last_effective_on,omitempty"`
}

type taxSummaryResponse struct {
	TaxType               string     `json:"tax_type"`
	TaxCode               string     `json:"tax_code"`
	TaxName               string     `json:"tax_name"`
	RateBasisPoints       int        `json:"rate_basis_points"`
	EntryCount            int        `json:"entry_count"`
	DocumentCount         int        `json:"document_count"`
	TotalDebitMinor       int64      `json:"total_debit_minor"`
	TotalCreditMinor      int64      `json:"total_credit_minor"`
	NetMinor              int64      `json:"net_minor"`
	ReceivableAccountID   *string    `json:"receivable_account_id,omitempty"`
	ReceivableAccountCode *string    `json:"receivable_account_code,omitempty"`
	ReceivableAccountName *string    `json:"receivable_account_name,omitempty"`
	PayableAccountID      *string    `json:"payable_account_id,omitempty"`
	PayableAccountCode    *string    `json:"payable_account_code,omitempty"`
	PayableAccountName    *string    `json:"payable_account_name,omitempty"`
	LastEffectiveOn       *time.Time `json:"last_effective_on,omitempty"`
}

type inventoryStockResponse struct {
	ItemID       string `json:"item_id"`
	ItemSKU      string `json:"item_sku"`
	ItemName     string `json:"item_name"`
	ItemRole     string `json:"item_role"`
	LocationID   string `json:"location_id"`
	LocationCode string `json:"location_code"`
	LocationName string `json:"location_name"`
	LocationRole string `json:"location_role"`
	OnHandMilli  int64  `json:"on_hand_milli"`
}

type inventoryMovementResponse struct {
	MovementID              string    `json:"movement_id"`
	MovementNumber          int64     `json:"movement_number"`
	DocumentID              *string   `json:"document_id,omitempty"`
	DocumentTypeCode        *string   `json:"document_type_code,omitempty"`
	DocumentTitle           *string   `json:"document_title,omitempty"`
	DocumentNumber          *string   `json:"document_number,omitempty"`
	DocumentStatus          *string   `json:"document_status,omitempty"`
	ItemID                  string    `json:"item_id"`
	ItemSKU                 string    `json:"item_sku"`
	ItemName                string    `json:"item_name"`
	ItemRole                string    `json:"item_role"`
	MovementType            string    `json:"movement_type"`
	MovementPurpose         string    `json:"movement_purpose"`
	UsageClassification     string    `json:"usage_classification"`
	SourceLocationID        *string   `json:"source_location_id,omitempty"`
	SourceLocationCode      *string   `json:"source_location_code,omitempty"`
	SourceLocationName      *string   `json:"source_location_name,omitempty"`
	SourceLocationRole      *string   `json:"source_location_role,omitempty"`
	DestinationLocationID   *string   `json:"destination_location_id,omitempty"`
	DestinationLocationCode *string   `json:"destination_location_code,omitempty"`
	DestinationLocationName *string   `json:"destination_location_name,omitempty"`
	DestinationLocationRole *string   `json:"destination_location_role,omitempty"`
	QuantityMilli           int64     `json:"quantity_milli"`
	ReferenceNote           string    `json:"reference_note"`
	CreatedByUserID         string    `json:"created_by_user_id"`
	CreatedAt               time.Time `json:"created_at"`
}

type inventoryReconciliationResponse struct {
	DocumentID              string     `json:"document_id"`
	DocumentTypeCode        string     `json:"document_type_code"`
	DocumentTitle           string     `json:"document_title"`
	DocumentNumber          *string    `json:"document_number,omitempty"`
	DocumentStatus          string     `json:"document_status"`
	DocumentLineID          string     `json:"document_line_id"`
	LineNumber              int        `json:"line_number"`
	MovementID              string     `json:"movement_id"`
	MovementNumber          int64      `json:"movement_number"`
	MovementType            string     `json:"movement_type"`
	MovementPurpose         string     `json:"movement_purpose"`
	UsageClassification     string     `json:"usage_classification"`
	ItemID                  string     `json:"item_id"`
	ItemSKU                 string     `json:"item_sku"`
	ItemName                string     `json:"item_name"`
	ItemRole                string     `json:"item_role"`
	SourceLocationID        *string    `json:"source_location_id,omitempty"`
	SourceLocationCode      *string    `json:"source_location_code,omitempty"`
	SourceLocationName      *string    `json:"source_location_name,omitempty"`
	DestinationLocationID   *string    `json:"destination_location_id,omitempty"`
	DestinationLocationCode *string    `json:"destination_location_code,omitempty"`
	DestinationLocationName *string    `json:"destination_location_name,omitempty"`
	QuantityMilli           int64      `json:"quantity_milli"`
	ExecutionLinkID         *string    `json:"execution_link_id,omitempty"`
	ExecutionContextType    *string    `json:"execution_context_type,omitempty"`
	ExecutionContextID      *string    `json:"execution_context_id,omitempty"`
	ExecutionLinkStatus     *string    `json:"execution_link_status,omitempty"`
	WorkOrderID             *string    `json:"work_order_id,omitempty"`
	WorkOrderCode           *string    `json:"work_order_code,omitempty"`
	WorkOrderStatus         *string    `json:"work_order_status,omitempty"`
	AccountingHandoffID     *string    `json:"accounting_handoff_id,omitempty"`
	AccountingHandoffStatus *string    `json:"accounting_handoff_status,omitempty"`
	CostMinor               *int64     `json:"cost_minor,omitempty"`
	CostCurrencyCode        *string    `json:"cost_currency_code,omitempty"`
	JournalEntryID          *string    `json:"journal_entry_id,omitempty"`
	JournalEntryNumber      *int64     `json:"journal_entry_number,omitempty"`
	AccountingPostedAt      *time.Time `json:"accounting_posted_at,omitempty"`
	MovementCreatedAt       time.Time  `json:"movement_created_at"`
}

type workOrderReviewResponse struct {
	WorkOrderID              string     `json:"work_order_id"`
	DocumentID               string     `json:"document_id"`
	DocumentStatus           string     `json:"document_status"`
	DocumentNumber           *string    `json:"document_number,omitempty"`
	WorkOrderCode            string     `json:"work_order_code"`
	Title                    string     `json:"title"`
	Summary                  string     `json:"summary"`
	Status                   string     `json:"status"`
	ClosedAt                 *time.Time `json:"closed_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	LastStatusChangedAt      time.Time  `json:"last_status_changed_at"`
	OpenTaskCount            int        `json:"open_task_count"`
	CompletedTaskCount       int        `json:"completed_task_count"`
	LaborEntryCount          int        `json:"labor_entry_count"`
	TotalLaborMinutes        int        `json:"total_labor_minutes"`
	TotalLaborCostMinor      int64      `json:"total_labor_cost_minor"`
	PostedLaborEntryCount    int        `json:"posted_labor_entry_count"`
	PostedLaborCostMinor     int64      `json:"posted_labor_cost_minor"`
	MaterialUsageCount       int        `json:"material_usage_count"`
	MaterialQuantityMilli    int64      `json:"material_quantity_milli"`
	PostedMaterialUsageCount int        `json:"posted_material_usage_count"`
	PostedMaterialCostMinor  int64      `json:"posted_material_cost_minor"`
	LastAccountingPostedAt   *time.Time `json:"last_accounting_posted_at,omitempty"`
}

type auditEventResponse struct {
	ID          string          `json:"id"`
	OrgID       *string         `json:"org_id,omitempty"`
	ActorUserID *string         `json:"actor_user_id,omitempty"`
	EventType   string          `json:"event_type"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	Payload     json.RawMessage `json:"payload"`
	OccurredAt  time.Time       `json:"occurred_at"`
}

type sessionContextResponse struct {
	SessionID       string    `json:"session_id"`
	OrgID           string    `json:"org_id"`
	OrgSlug         string    `json:"org_slug"`
	OrgName         string    `json:"org_name"`
	UserID          string    `json:"user_id"`
	UserEmail       string    `json:"user_email"`
	UserDisplayName string    `json:"user_display_name"`
	RoleCode        string    `json:"role_code"`
	DeviceLabel     string    `json:"device_label"`
	ExpiresAt       time.Time `json:"expires_at"`
	IssuedAt        time.Time `json:"issued_at"`
	LastSeenAt      time.Time `json:"last_seen_at"`
}

func mapSessionContext(context identityaccess.SessionContext) sessionContextResponse {
	return sessionContextResponse{
		SessionID:       context.Session.ID,
		OrgID:           context.Actor.OrgID,
		OrgSlug:         context.OrgSlug,
		OrgName:         context.OrgName,
		UserID:          context.Actor.UserID,
		UserEmail:       context.UserEmail,
		UserDisplayName: context.UserDisplayName,
		RoleCode:        context.RoleCode,
		DeviceLabel:     context.Session.DeviceLabel,
		ExpiresAt:       context.Session.ExpiresAt,
		IssuedAt:        context.Session.IssuedAt,
		LastSeenAt:      context.Session.LastSeenAt,
	}
}

type inboundRequestReviewResponse struct {
	RequestID                string          `json:"request_id"`
	RequestReference         string          `json:"request_reference"`
	SessionID                *string         `json:"session_id,omitempty"`
	ActorUserID              *string         `json:"actor_user_id,omitempty"`
	OriginType               string          `json:"origin_type"`
	Channel                  string          `json:"channel"`
	Status                   string          `json:"status"`
	Metadata                 json.RawMessage `json:"metadata"`
	CancellationReason       string          `json:"cancellation_reason,omitempty"`
	FailureReason            string          `json:"failure_reason,omitempty"`
	ReceivedAt               time.Time       `json:"received_at"`
	QueuedAt                 *time.Time      `json:"queued_at,omitempty"`
	ProcessingStartedAt      *time.Time      `json:"processing_started_at,omitempty"`
	ProcessedAt              *time.Time      `json:"processed_at,omitempty"`
	ActedOnAt                *time.Time      `json:"acted_on_at,omitempty"`
	CompletedAt              *time.Time      `json:"completed_at,omitempty"`
	FailedAt                 *time.Time      `json:"failed_at,omitempty"`
	CancelledAt              *time.Time      `json:"cancelled_at,omitempty"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
	MessageCount             int             `json:"message_count"`
	AttachmentCount          int             `json:"attachment_count"`
	LastRunID                *string         `json:"last_run_id,omitempty"`
	LastRunStatus            *string         `json:"last_run_status,omitempty"`
	LastRecommendationID     *string         `json:"last_recommendation_id,omitempty"`
	LastRecommendationStatus *string         `json:"last_recommendation_status,omitempty"`
}

type inboundRequestStatusSummaryResponse struct {
	Status           string     `json:"status"`
	RequestCount     int        `json:"request_count"`
	MessageCount     int        `json:"message_count"`
	AttachmentCount  int        `json:"attachment_count"`
	LatestReceivedAt *time.Time `json:"latest_received_at,omitempty"`
	LatestQueuedAt   *time.Time `json:"latest_queued_at,omitempty"`
	LatestUpdatedAt  time.Time  `json:"latest_updated_at"`
}

type inboundRequestMessageResponse struct {
	MessageID       string    `json:"message_id"`
	MessageIndex    int       `json:"message_index"`
	MessageRole     string    `json:"message_role"`
	TextContent     string    `json:"text_content"`
	CreatedByUserID *string   `json:"created_by_user_id,omitempty"`
	AttachmentCount int       `json:"attachment_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type requestAttachmentResponse struct {
	AttachmentID         string    `json:"attachment_id"`
	RequestMessageID     string    `json:"request_message_id"`
	LinkRole             string    `json:"link_role"`
	OriginalFileName     string    `json:"original_file_name"`
	MediaType            string    `json:"media_type"`
	SizeBytes            int64     `json:"size_bytes"`
	UploadedByUserID     *string   `json:"uploaded_by_user_id,omitempty"`
	LatestDerivedText    *string   `json:"latest_derived_text,omitempty"`
	LatestDerivedByRunID *string   `json:"latest_derived_by_run_id,omitempty"`
	DerivedTextCount     int       `json:"derived_text_count"`
	CreatedAt            time.Time `json:"created_at"`
}

type aiRunResponse struct {
	RunID          string     `json:"run_id"`
	AgentRole      string     `json:"agent_role"`
	CapabilityCode string     `json:"capability_code"`
	Status         string     `json:"status"`
	Summary        string     `json:"summary"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type aiStepResponse struct {
	StepID        string          `json:"step_id"`
	RunID         string          `json:"run_id"`
	StepIndex     int             `json:"step_index"`
	StepType      string          `json:"step_type"`
	StepTitle     string          `json:"step_title"`
	Status        string          `json:"status"`
	InputPayload  json.RawMessage `json:"input_payload"`
	OutputPayload json.RawMessage `json:"output_payload"`
	CreatedAt     time.Time       `json:"created_at"`
}

type aiDelegationResponse struct {
	DelegationID        string    `json:"delegation_id"`
	ParentRunID         string    `json:"parent_run_id"`
	ChildRunID          string    `json:"child_run_id"`
	RequestedByStepID   *string   `json:"requested_by_step_id,omitempty"`
	CapabilityCode      string    `json:"capability_code"`
	Reason              string    `json:"reason"`
	ChildAgentRole      string    `json:"child_agent_role"`
	ChildCapabilityCode string    `json:"child_capability_code"`
	ChildRunStatus      string    `json:"child_run_status"`
	CreatedAt           time.Time `json:"created_at"`
}

type aiArtifactResponse struct {
	ArtifactID      string          `json:"artifact_id"`
	RunID           string          `json:"run_id"`
	StepID          *string         `json:"step_id,omitempty"`
	ArtifactType    string          `json:"artifact_type"`
	Title           string          `json:"title"`
	Payload         json.RawMessage `json:"payload"`
	CreatedByUserID string          `json:"created_by_user_id"`
	CreatedAt       time.Time       `json:"created_at"`
}

type aiRecommendationResponse struct {
	RecommendationID   string          `json:"recommendation_id"`
	RunID              string          `json:"run_id"`
	ArtifactID         *string         `json:"artifact_id,omitempty"`
	ApprovalID         *string         `json:"approval_id,omitempty"`
	RecommendationType string          `json:"recommendation_type"`
	Status             string          `json:"status"`
	Summary            string          `json:"summary"`
	Payload            json.RawMessage `json:"payload"`
	CreatedByUserID    string          `json:"created_by_user_id"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type processedProposalReviewResponse struct {
	RequestID            string    `json:"request_id"`
	RequestReference     string    `json:"request_reference"`
	RequestStatus        string    `json:"request_status"`
	RecommendationID     string    `json:"recommendation_id"`
	RunID                string    `json:"run_id"`
	RecommendationType   string    `json:"recommendation_type"`
	RecommendationStatus string    `json:"recommendation_status"`
	Summary              string    `json:"summary"`
	ApprovalID           *string   `json:"approval_id,omitempty"`
	ApprovalStatus       *string   `json:"approval_status,omitempty"`
	ApprovalQueueCode    *string   `json:"approval_queue_code,omitempty"`
	DocumentID           *string   `json:"document_id,omitempty"`
	DocumentTypeCode     *string   `json:"document_type_code,omitempty"`
	DocumentTitle        *string   `json:"document_title,omitempty"`
	DocumentNumber       *string   `json:"document_number,omitempty"`
	DocumentStatus       *string   `json:"document_status,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

type processedProposalStatusSummaryResponse struct {
	RecommendationStatus string    `json:"recommendation_status"`
	ProposalCount        int       `json:"proposal_count"`
	RequestCount         int       `json:"request_count"`
	DocumentCount        int       `json:"document_count"`
	LatestCreatedAt      time.Time `json:"latest_created_at"`
}

type inboundRequestDetailResponse struct {
	Request         inboundRequestReviewResponse      `json:"request"`
	Messages        []inboundRequestMessageResponse   `json:"messages"`
	Attachments     []requestAttachmentResponse       `json:"attachments"`
	Runs            []aiRunResponse                   `json:"runs"`
	Steps           []aiStepResponse                  `json:"steps"`
	Delegations     []aiDelegationResponse            `json:"delegations"`
	Artifacts       []aiArtifactResponse              `json:"artifacts"`
	Recommendations []aiRecommendationResponse        `json:"recommendations"`
	Proposals       []processedProposalReviewResponse `json:"proposals"`
}

type approvalDecisionResponse struct {
	ApprovalID      string     `json:"approval_id"`
	Status          string     `json:"status"`
	QueueCode       string     `json:"queue_code"`
	DocumentID      string     `json:"document_id"`
	DocumentStatus  string     `json:"document_status"`
	DecisionNote    *string    `json:"decision_note,omitempty"`
	DecidedByUserID *string    `json:"decided_by_user_id,omitempty"`
	DecidedAt       *time.Time `json:"decided_at,omitempty"`
}

func handleReviewError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, identityaccess.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	case errors.Is(err, reporting.ErrInvalidReviewFilter):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid review filter"})
	case errors.Is(err, sql.ErrNoRows), errors.Is(err, reporting.ErrWorkOrderNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "record not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: fallback})
	}
}

func mapApprovalQueueEntry(entry reporting.ApprovalQueueEntry) approvalQueueEntryResponse {
	return approvalQueueEntryResponse{
		QueueEntryID:         entry.QueueEntryID,
		ApprovalID:           entry.ApprovalID,
		QueueCode:            entry.QueueCode,
		QueueStatus:          entry.QueueStatus,
		EnqueuedAt:           entry.EnqueuedAt,
		ClosedAt:             timePtr(entry.ClosedAt),
		ApprovalStatus:       entry.ApprovalStatus,
		RequestedAt:          entry.RequestedAt,
		RequestedByUserID:    entry.RequestedByUserID,
		DecidedAt:            timePtr(entry.DecidedAt),
		DecidedByUserID:      stringPtr(entry.DecidedByUserID),
		DocumentID:           entry.DocumentID,
		DocumentTypeCode:     entry.DocumentTypeCode,
		DocumentTitle:        entry.DocumentTitle,
		DocumentNumber:       stringPtr(entry.DocumentNumber),
		DocumentStatus:       entry.DocumentStatus,
		JournalEntryID:       stringPtr(entry.JournalEntryID),
		JournalEntryNumber:   int64Ptr(entry.JournalEntryNumber),
		JournalEntryPostedAt: timePtr(entry.JournalEntryPostedAt),
	}
}

func mapDocumentReview(review reporting.DocumentReview) documentReviewResponse {
	return documentReviewResponse{
		DocumentID:           review.DocumentID,
		TypeCode:             review.TypeCode,
		Title:                review.Title,
		NumberValue:          stringPtr(review.NumberValue),
		Status:               review.Status,
		SourceDocumentID:     stringPtr(review.SourceDocumentID),
		CreatedByUserID:      review.CreatedByUserID,
		SubmittedByUserID:    stringPtr(review.SubmittedByUserID),
		SubmittedAt:          timePtr(review.SubmittedAt),
		ApprovedAt:           timePtr(review.ApprovedAt),
		RejectedAt:           timePtr(review.RejectedAt),
		CreatedAt:            review.CreatedAt,
		UpdatedAt:            review.UpdatedAt,
		ApprovalID:           stringPtr(review.ApprovalID),
		ApprovalStatus:       stringPtr(review.ApprovalStatus),
		ApprovalQueueCode:    stringPtr(review.ApprovalQueueCode),
		ApprovalRequestedAt:  timePtr(review.ApprovalRequestedAt),
		ApprovalDecidedAt:    timePtr(review.ApprovalDecidedAt),
		JournalEntryID:       stringPtr(review.JournalEntryID),
		JournalEntryNumber:   int64Ptr(review.JournalEntryNumber),
		JournalEntryPostedAt: timePtr(review.JournalEntryPostedAt),
	}
}

func mapJournalEntryReview(review reporting.JournalEntryReview) journalEntryReviewResponse {
	return journalEntryReviewResponse{
		EntryID:           review.EntryID,
		EntryNumber:       review.EntryNumber,
		EntryKind:         review.EntryKind,
		SourceDocumentID:  stringPtr(review.SourceDocumentID),
		ReversalOfEntryID: stringPtr(review.ReversalOfEntryID),
		CurrencyCode:      review.CurrencyCode,
		TaxScopeCode:      review.TaxScopeCode,
		Summary:           review.Summary,
		ReversalReason:    stringPtr(review.ReversalReason),
		PostedByUserID:    review.PostedByUserID,
		EffectiveOn:       review.EffectiveOn,
		PostedAt:          review.PostedAt,
		CreatedAt:         review.CreatedAt,
		DocumentTypeCode:  stringPtr(review.DocumentTypeCode),
		DocumentNumber:    stringPtr(review.DocumentNumber),
		DocumentStatus:    stringPtr(review.DocumentStatus),
		LineCount:         review.LineCount,
		TotalDebitMinor:   review.TotalDebitMinor,
		TotalCreditMinor:  review.TotalCreditMinor,
		HasReversal:       review.HasReversal,
	}
}

func mapControlAccountBalance(balance reporting.ControlAccountBalance) controlAccountBalanceResponse {
	return controlAccountBalanceResponse{
		AccountID:        balance.AccountID,
		AccountCode:      balance.AccountCode,
		AccountName:      balance.AccountName,
		AccountClass:     balance.AccountClass,
		ControlType:      balance.ControlType,
		TotalDebitMinor:  balance.TotalDebitMinor,
		TotalCreditMinor: balance.TotalCreditMinor,
		NetMinor:         balance.NetMinor,
		LastEffectiveOn:  timePtr(balance.LastEffectiveOn),
	}
}

func mapTaxSummary(summary reporting.TaxSummary) taxSummaryResponse {
	return taxSummaryResponse{
		TaxType:               summary.TaxType,
		TaxCode:               summary.TaxCode,
		TaxName:               summary.TaxName,
		RateBasisPoints:       summary.RateBasisPoints,
		EntryCount:            summary.EntryCount,
		DocumentCount:         summary.DocumentCount,
		TotalDebitMinor:       summary.TotalDebitMinor,
		TotalCreditMinor:      summary.TotalCreditMinor,
		NetMinor:              summary.NetMinor,
		ReceivableAccountID:   stringPtr(summary.ReceivableAccountID),
		ReceivableAccountCode: stringPtr(summary.ReceivableAccountCode),
		ReceivableAccountName: stringPtr(summary.ReceivableAccountName),
		PayableAccountID:      stringPtr(summary.PayableAccountID),
		PayableAccountCode:    stringPtr(summary.PayableAccountCode),
		PayableAccountName:    stringPtr(summary.PayableAccountName),
		LastEffectiveOn:       timePtr(summary.LastEffectiveOn),
	}
}

func mapInventoryStock(item reporting.InventoryStockItem) inventoryStockResponse {
	return inventoryStockResponse{
		ItemID:       item.ItemID,
		ItemSKU:      item.ItemSKU,
		ItemName:     item.ItemName,
		ItemRole:     item.ItemRole,
		LocationID:   item.LocationID,
		LocationCode: item.LocationCode,
		LocationName: item.LocationName,
		LocationRole: item.LocationRole,
		OnHandMilli:  item.OnHandMilli,
	}
}

func mapInventoryMovement(review reporting.InventoryMovementReview) inventoryMovementResponse {
	return inventoryMovementResponse{
		MovementID:              review.MovementID,
		MovementNumber:          review.MovementNumber,
		DocumentID:              stringPtr(review.DocumentID),
		DocumentTypeCode:        stringPtr(review.DocumentTypeCode),
		DocumentTitle:           stringPtr(review.DocumentTitle),
		DocumentNumber:          stringPtr(review.DocumentNumber),
		DocumentStatus:          stringPtr(review.DocumentStatus),
		ItemID:                  review.ItemID,
		ItemSKU:                 review.ItemSKU,
		ItemName:                review.ItemName,
		ItemRole:                review.ItemRole,
		MovementType:            review.MovementType,
		MovementPurpose:         review.MovementPurpose,
		UsageClassification:     review.UsageClassification,
		SourceLocationID:        stringPtr(review.SourceLocationID),
		SourceLocationCode:      stringPtr(review.SourceLocationCode),
		SourceLocationName:      stringPtr(review.SourceLocationName),
		SourceLocationRole:      stringPtr(review.SourceLocationRole),
		DestinationLocationID:   stringPtr(review.DestinationLocationID),
		DestinationLocationCode: stringPtr(review.DestinationLocationCode),
		DestinationLocationName: stringPtr(review.DestinationLocationName),
		DestinationLocationRole: stringPtr(review.DestinationLocationRole),
		QuantityMilli:           review.QuantityMilli,
		ReferenceNote:           review.ReferenceNote,
		CreatedByUserID:         review.CreatedByUserID,
		CreatedAt:               review.CreatedAt,
	}
}

func mapInventoryReconciliation(item reporting.InventoryReconciliationItem) inventoryReconciliationResponse {
	return inventoryReconciliationResponse{
		DocumentID:              item.DocumentID,
		DocumentTypeCode:        item.DocumentTypeCode,
		DocumentTitle:           item.DocumentTitle,
		DocumentNumber:          stringPtr(item.DocumentNumber),
		DocumentStatus:          item.DocumentStatus,
		DocumentLineID:          item.DocumentLineID,
		LineNumber:              item.LineNumber,
		MovementID:              item.MovementID,
		MovementNumber:          item.MovementNumber,
		MovementType:            item.MovementType,
		MovementPurpose:         item.MovementPurpose,
		UsageClassification:     item.UsageClassification,
		ItemID:                  item.ItemID,
		ItemSKU:                 item.ItemSKU,
		ItemName:                item.ItemName,
		ItemRole:                item.ItemRole,
		SourceLocationID:        stringPtr(item.SourceLocationID),
		SourceLocationCode:      stringPtr(item.SourceLocationCode),
		SourceLocationName:      stringPtr(item.SourceLocationName),
		DestinationLocationID:   stringPtr(item.DestinationLocationID),
		DestinationLocationCode: stringPtr(item.DestinationLocationCode),
		DestinationLocationName: stringPtr(item.DestinationLocationName),
		QuantityMilli:           item.QuantityMilli,
		ExecutionLinkID:         stringPtr(item.ExecutionLinkID),
		ExecutionContextType:    stringPtr(item.ExecutionContextType),
		ExecutionContextID:      stringPtr(item.ExecutionContextID),
		ExecutionLinkStatus:     stringPtr(item.ExecutionLinkStatus),
		WorkOrderID:             stringPtr(item.WorkOrderID),
		WorkOrderCode:           stringPtr(item.WorkOrderCode),
		WorkOrderStatus:         stringPtr(item.WorkOrderStatus),
		AccountingHandoffID:     stringPtr(item.AccountingHandoffID),
		AccountingHandoffStatus: stringPtr(item.AccountingHandoffStatus),
		CostMinor:               int64Ptr(item.CostMinor),
		CostCurrencyCode:        stringPtr(item.CostCurrencyCode),
		JournalEntryID:          stringPtr(item.JournalEntryID),
		JournalEntryNumber:      int64Ptr(item.JournalEntryNumber),
		AccountingPostedAt:      timePtr(item.AccountingPostedAt),
		MovementCreatedAt:       item.MovementCreatedAt,
	}
}

func mapWorkOrderReview(review reporting.WorkOrderReview) workOrderReviewResponse {
	return workOrderReviewResponse{
		WorkOrderID:              review.WorkOrderID,
		DocumentID:               review.DocumentID,
		DocumentStatus:           review.DocumentStatus,
		DocumentNumber:           stringPtr(review.DocumentNumber),
		WorkOrderCode:            review.WorkOrderCode,
		Title:                    review.Title,
		Summary:                  review.Summary,
		Status:                   review.Status,
		ClosedAt:                 timePtr(review.ClosedAt),
		CreatedAt:                review.CreatedAt,
		UpdatedAt:                review.UpdatedAt,
		LastStatusChangedAt:      review.LastStatusChangedAt,
		OpenTaskCount:            review.OpenTaskCount,
		CompletedTaskCount:       review.CompletedTaskCount,
		LaborEntryCount:          review.LaborEntryCount,
		TotalLaborMinutes:        review.TotalLaborMinutes,
		TotalLaborCostMinor:      review.TotalLaborCostMinor,
		PostedLaborEntryCount:    review.PostedLaborEntryCount,
		PostedLaborCostMinor:     review.PostedLaborCostMinor,
		MaterialUsageCount:       review.MaterialUsageCount,
		MaterialQuantityMilli:    review.MaterialQuantityMilli,
		PostedMaterialUsageCount: review.PostedMaterialUsageCount,
		PostedMaterialCostMinor:  review.PostedMaterialCostMinor,
		LastAccountingPostedAt:   timePtr(review.LastAccountingPostedAt),
	}
}

func mapAuditEvent(event reporting.AuditEvent) auditEventResponse {
	return auditEventResponse{
		ID:          event.ID,
		OrgID:       stringPtr(event.OrgID),
		ActorUserID: stringPtr(event.ActorUserID),
		EventType:   event.EventType,
		EntityType:  event.EntityType,
		EntityID:    event.EntityID,
		Payload:     event.Payload,
		OccurredAt:  event.OccurredAt,
	}
}

func mapInboundRequestReview(review reporting.InboundRequestReview) inboundRequestReviewResponse {
	return inboundRequestReviewResponse{
		RequestID:                review.RequestID,
		RequestReference:         review.RequestReference,
		SessionID:                stringPtr(review.SessionID),
		ActorUserID:              stringPtr(review.ActorUserID),
		OriginType:               review.OriginType,
		Channel:                  review.Channel,
		Status:                   review.Status,
		Metadata:                 review.Metadata,
		CancellationReason:       review.CancellationReason,
		FailureReason:            review.FailureReason,
		ReceivedAt:               review.ReceivedAt,
		QueuedAt:                 timePtr(review.QueuedAt),
		ProcessingStartedAt:      timePtr(review.ProcessingStartedAt),
		ProcessedAt:              timePtr(review.ProcessedAt),
		ActedOnAt:                timePtr(review.ActedOnAt),
		CompletedAt:              timePtr(review.CompletedAt),
		FailedAt:                 timePtr(review.FailedAt),
		CancelledAt:              timePtr(review.CancelledAt),
		CreatedAt:                review.CreatedAt,
		UpdatedAt:                review.UpdatedAt,
		MessageCount:             review.MessageCount,
		AttachmentCount:          review.AttachmentCount,
		LastRunID:                stringPtr(review.LastRunID),
		LastRunStatus:            stringPtr(review.LastRunStatus),
		LastRecommendationID:     stringPtr(review.LastRecommendationID),
		LastRecommendationStatus: stringPtr(review.LastRecommendationStatus),
	}
}

func mapInboundRequestDetail(detail reporting.InboundRequestDetail) inboundRequestDetailResponse {
	response := inboundRequestDetailResponse{
		Request:         mapInboundRequestReview(detail.Request),
		Messages:        make([]inboundRequestMessageResponse, 0, len(detail.Messages)),
		Attachments:     make([]requestAttachmentResponse, 0, len(detail.Attachments)),
		Runs:            make([]aiRunResponse, 0, len(detail.Runs)),
		Steps:           make([]aiStepResponse, 0, len(detail.Steps)),
		Delegations:     make([]aiDelegationResponse, 0, len(detail.Delegations)),
		Artifacts:       make([]aiArtifactResponse, 0, len(detail.Artifacts)),
		Recommendations: make([]aiRecommendationResponse, 0, len(detail.Recommendations)),
		Proposals:       make([]processedProposalReviewResponse, 0, len(detail.Proposals)),
	}
	for _, item := range detail.Messages {
		response.Messages = append(response.Messages, inboundRequestMessageResponse{
			MessageID:       item.MessageID,
			MessageIndex:    item.MessageIndex,
			MessageRole:     item.MessageRole,
			TextContent:     item.TextContent,
			CreatedByUserID: stringPtr(item.CreatedByUserID),
			AttachmentCount: item.AttachmentCount,
			CreatedAt:       item.CreatedAt,
		})
	}
	for _, item := range detail.Attachments {
		response.Attachments = append(response.Attachments, requestAttachmentResponse{
			AttachmentID:         item.AttachmentID,
			RequestMessageID:     item.RequestMessageID,
			LinkRole:             item.LinkRole,
			OriginalFileName:     item.OriginalFileName,
			MediaType:            item.MediaType,
			SizeBytes:            item.SizeBytes,
			UploadedByUserID:     stringPtr(item.UploadedByUserID),
			LatestDerivedText:    stringPtr(item.LatestDerivedText),
			LatestDerivedByRunID: stringPtr(item.LatestDerivedByRunID),
			DerivedTextCount:     item.DerivedTextCount,
			CreatedAt:            item.CreatedAt,
		})
	}
	for _, item := range detail.Runs {
		response.Runs = append(response.Runs, aiRunResponse{
			RunID:          item.RunID,
			AgentRole:      item.AgentRole,
			CapabilityCode: item.CapabilityCode,
			Status:         item.Status,
			Summary:        item.Summary,
			StartedAt:      item.StartedAt,
			CompletedAt:    timePtr(item.CompletedAt),
		})
	}
	for _, item := range detail.Steps {
		response.Steps = append(response.Steps, aiStepResponse{
			StepID:        item.StepID,
			RunID:         item.RunID,
			StepIndex:     item.StepIndex,
			StepType:      item.StepType,
			StepTitle:     item.StepTitle,
			Status:        item.Status,
			InputPayload:  item.InputPayload,
			OutputPayload: item.OutputPayload,
			CreatedAt:     item.CreatedAt,
		})
	}
	for _, item := range detail.Delegations {
		response.Delegations = append(response.Delegations, aiDelegationResponse{
			DelegationID:        item.DelegationID,
			ParentRunID:         item.ParentRunID,
			ChildRunID:          item.ChildRunID,
			RequestedByStepID:   stringPtr(item.RequestedByStepID),
			CapabilityCode:      item.CapabilityCode,
			Reason:              item.Reason,
			ChildAgentRole:      item.ChildAgentRole,
			ChildCapabilityCode: item.ChildCapabilityCode,
			ChildRunStatus:      item.ChildRunStatus,
			CreatedAt:           item.CreatedAt,
		})
	}
	for _, item := range detail.Artifacts {
		response.Artifacts = append(response.Artifacts, aiArtifactResponse{
			ArtifactID:      item.ArtifactID,
			RunID:           item.RunID,
			StepID:          stringPtr(item.StepID),
			ArtifactType:    item.ArtifactType,
			Title:           item.Title,
			Payload:         item.Payload,
			CreatedByUserID: item.CreatedByUserID,
			CreatedAt:       item.CreatedAt,
		})
	}
	for _, item := range detail.Recommendations {
		response.Recommendations = append(response.Recommendations, aiRecommendationResponse{
			RecommendationID:   item.RecommendationID,
			RunID:              item.RunID,
			ArtifactID:         stringPtr(item.ArtifactID),
			ApprovalID:         stringPtr(item.ApprovalID),
			RecommendationType: item.RecommendationType,
			Status:             item.Status,
			Summary:            item.Summary,
			Payload:            item.Payload,
			CreatedByUserID:    item.CreatedByUserID,
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
		})
	}
	for _, item := range detail.Proposals {
		response.Proposals = append(response.Proposals, mapProcessedProposalReview(item))
	}
	return response
}

func mapProcessedProposalReview(item reporting.ProcessedProposalReview) processedProposalReviewResponse {
	return processedProposalReviewResponse{
		RequestID:            item.RequestID,
		RequestReference:     item.RequestReference,
		RequestStatus:        item.RequestStatus,
		RecommendationID:     item.RecommendationID,
		RunID:                item.RunID,
		RecommendationType:   item.RecommendationType,
		RecommendationStatus: item.RecommendationStatus,
		Summary:              item.Summary,
		ApprovalID:           stringPtr(item.ApprovalID),
		ApprovalStatus:       stringPtr(item.ApprovalStatus),
		ApprovalQueueCode:    stringPtr(item.ApprovalQueueCode),
		DocumentID:           stringPtr(item.DocumentID),
		DocumentTypeCode:     stringPtr(item.DocumentTypeCode),
		DocumentTitle:        stringPtr(item.DocumentTitle),
		DocumentNumber:       stringPtr(item.DocumentNumber),
		DocumentStatus:       stringPtr(item.DocumentStatus),
		CreatedAt:            item.CreatedAt,
	}
}

func stringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func timePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}

func int64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}
