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

	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
)

const (
	agentProcessNextQueuedPath = "/api/agent/process-next-queued-inbound-request"
	submitInboundRequestPath   = "/api/inbound-requests"
	attachmentContentPrefix    = "/api/attachments/"
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

type AgentAPIHandler struct {
	loadProcessor     queuedInboundRequestProcessorLoader
	submissionService inboundRequestSubmitter
}

func NewAgentAPIHandler(db *sql.DB) http.Handler {
	return NewAgentAPIHandlerWithServices(func() (ProcessNextQueuedInboundRequester, error) {
		return NewOpenAIAgentProcessorFromEnv(db)
	}, NewSubmissionService(db))
}

func NewAgentAPIHandlerWithProcessorLoader(loader queuedInboundRequestProcessorLoader) http.Handler {
	return NewAgentAPIHandlerWithServices(loader, nil)
}

func NewAgentAPIHandlerWithServices(loader queuedInboundRequestProcessorLoader, submissionService inboundRequestSubmitter) http.Handler {
	handler := &AgentAPIHandler{
		loadProcessor:     loader,
		submissionService: submissionService,
	}
	mux := http.NewServeMux()
	mux.HandleFunc(agentProcessNextQueuedPath, handler.handleProcessNextQueuedInboundRequest)
	mux.HandleFunc(submitInboundRequestPath, handler.handleSubmitInboundRequest)
	mux.HandleFunc(attachmentContentPrefix, handler.handleDownloadAttachment)
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

func contentDisposition(fileName string) string {
	encoded := url.PathEscape(fileName)
	return fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", fileName, encoded)
}
