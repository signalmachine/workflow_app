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
	agentProcessNextQueuedPath = "/api/agent/process-next-queued-inbound-request"
	submitInboundRequestPath   = "/api/inbound-requests"
	attachmentContentPrefix    = "/api/attachments/"
	reviewInboundRequestsPath  = "/api/review/inbound-requests"
	reviewInboundSummaryPath   = "/api/review/inbound-request-status-summary"
	reviewProposalListPath     = "/api/review/processed-proposals"
	reviewProposalSummaryPath  = "/api/review/processed-proposal-status-summary"
	reviewApprovalQueuePath    = "/api/review/approval-queue"
	approvalDecisionPrefix     = "/api/approvals/"
	headerOrgID                = "X-Workflow-Org-ID"
	headerUserID               = "X-Workflow-User-ID"
	headerSessionID            = "X-Workflow-Session-ID"
)

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

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
	ListInboundRequests(ctx context.Context, input reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error)
	GetInboundRequestDetail(ctx context.Context, input reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error)
	ListInboundRequestStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error)
	ListProcessedProposals(ctx context.Context, input reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error)
	ListProcessedProposalStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error)
}

type approvalDecisionService interface {
	DecideApproval(ctx context.Context, input workflow.DecideApprovalInput) (workflow.Approval, documents.Document, error)
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

type AgentAPIHandler struct {
	loadProcessor     queuedInboundRequestProcessorLoader
	submissionService inboundRequestSubmitter
	reviewService     operatorReviewReader
	approvalService   approvalDecisionService
}

func NewAgentAPIHandler(db *sql.DB) http.Handler {
	documentService := documents.NewService(db)
	return NewAgentAPIHandlerWithDependencies(func() (ProcessNextQueuedInboundRequester, error) {
		return NewOpenAIAgentProcessorFromEnv(db)
	}, NewSubmissionService(db), reporting.NewService(db), workflow.NewService(db, documentService))
}

func NewAgentAPIHandlerWithProcessorLoader(loader queuedInboundRequestProcessorLoader) http.Handler {
	return NewAgentAPIHandlerWithDependencies(loader, nil, nil, nil)
}

func NewAgentAPIHandlerWithServices(loader queuedInboundRequestProcessorLoader, submissionService inboundRequestSubmitter) http.Handler {
	return NewAgentAPIHandlerWithDependencies(loader, submissionService, nil, nil)
}

func NewAgentAPIHandlerWithDependencies(loader queuedInboundRequestProcessorLoader, submissionService inboundRequestSubmitter, reviewService operatorReviewReader, approvalService approvalDecisionService) http.Handler {
	handler := &AgentAPIHandler{
		loadProcessor:     loader,
		submissionService: submissionService,
		reviewService:     reviewService,
		approvalService:   approvalService,
	}
	mux := http.NewServeMux()
	mux.HandleFunc(agentProcessNextQueuedPath, handler.handleProcessNextQueuedInboundRequest)
	mux.HandleFunc(submitInboundRequestPath, handler.handleSubmitInboundRequest)
	mux.HandleFunc(attachmentContentPrefix, handler.handleDownloadAttachment)
	mux.HandleFunc(reviewInboundRequestsPath, handler.handleListInboundRequests)
	mux.HandleFunc(reviewInboundRequestsPath+"/", handler.handleGetInboundRequestDetail)
	mux.HandleFunc(reviewInboundSummaryPath, handler.handleListInboundRequestStatusSummary)
	mux.HandleFunc(reviewProposalListPath, handler.handleListProcessedProposals)
	mux.HandleFunc(reviewProposalSummaryPath, handler.handleListProcessedProposalStatusSummary)
	mux.HandleFunc(reviewApprovalQueuePath, handler.handleListApprovalQueue)
	mux.HandleFunc(approvalDecisionPrefix, handler.handleDecideApproval)
	return mux
}

func (h *AgentAPIHandler) handleProcessNextQueuedInboundRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	actor, err := actorFromHeaders(r)
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

	actor, err := actorFromHeaders(r)
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

	actor, err := actorFromHeaders(r)
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

	actor, err := actorFromHeaders(r)
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

	actor, err := actorFromHeaders(r)
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

	actor, err := actorFromHeaders(r)
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

	actor, err := actorFromHeaders(r)
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

	actor, err := actorFromHeaders(r)
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

	actor, err := actorFromHeaders(r)
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

	actor, err := actorFromHeaders(r)
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
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
	case errors.Is(err, sql.ErrNoRows):
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
