package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
)

const (
	agentProcessNextQueuedPath = "/api/agent/process-next-queued-inbound-request"
	headerOrgID                = "X-Workflow-Org-ID"
	headerUserID               = "X-Workflow-User-ID"
	headerSessionID            = "X-Workflow-Session-ID"
)

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type ProcessNextQueuedInboundRequester interface {
	ProcessNextQueuedInboundRequest(ctx context.Context, input ProcessNextQueuedInboundRequestInput) (ProcessNextQueuedInboundRequestResult, error)
}

type queuedInboundRequestProcessorLoader func() (ProcessNextQueuedInboundRequester, error)

type processNextQueuedRequest struct {
	Channel string `json:"channel"`
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

type errorResponse struct {
	Error string `json:"error"`
}

type AgentAPIHandler struct {
	loadProcessor queuedInboundRequestProcessorLoader
}

func NewAgentAPIHandler(db *sql.DB) http.Handler {
	return NewAgentAPIHandlerWithProcessorLoader(func() (ProcessNextQueuedInboundRequester, error) {
		return NewOpenAIAgentProcessorFromEnv(db)
	})
}

func NewAgentAPIHandlerWithProcessorLoader(loader queuedInboundRequestProcessorLoader) http.Handler {
	handler := &AgentAPIHandler{loadProcessor: loader}
	mux := http.NewServeMux()
	mux.HandleFunc(agentProcessNextQueuedPath, handler.handleProcessNextQueuedInboundRequest)
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
