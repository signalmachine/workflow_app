package app

import (
	"errors"
	"net/http"
	"strings"

	"workflow_app/internal/ai"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/workflow"
)

func (h *AgentAPIHandler) handleProcessedProposalAction(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == reviewProposalListPath {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodGet {
		h.handleGetProcessedProposalDetail(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.proposalApproval == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "proposal approval service unavailable"})
		return
	}

	recommendationID, action, ok := parseChildActionPath(reviewProposalListPath, r.URL.Path)
	if !ok || action != "request-approval" {
		http.NotFound(w, r)
		return
	}
	if !uuidPattern.MatchString(recommendationID) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid review filter"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	defer r.Body.Close()

	var req requestProcessedProposalApprovalRequest
	if err := decodeJSONBody(r, &req, false); err != nil {
		writeJSONBodyError(w, err)
		return
	}

	approval, proposal, err := h.proposalApproval.RequestProcessedProposalApproval(r.Context(), requestProcessedProposalApprovalInput{
		RecommendationID: recommendationID,
		QueueCode:        strings.TrimSpace(req.QueueCode),
		Reason:           strings.TrimSpace(req.Reason),
		Actor:            actor,
	})
	if err != nil {
		handleReviewError(w, err, "failed to request proposal approval")
		return
	}

	response := processedProposalApprovalResponse{
		RecommendationID:     proposal.RecommendationID,
		RecommendationStatus: ai.RecommendationStatusApprovalRequested,
		ApprovalID:           approval.ID,
		ApprovalStatus:       approval.Status,
		ApprovalQueueCode:    approval.QueueCode,
		DocumentID:           approval.DocumentID,
		DocumentStatus:       stringPtr(proposal.DocumentStatus),
		RequestedAt:          approval.RequestedAt,
	}
	writeJSON(w, http.StatusCreated, response)
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
	if !uuidPattern.MatchString(approvalID) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid approval"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	var req decideApprovalRequest
	defer r.Body.Close()
	if err := decodeJSONBody(r, &req, false); err != nil {
		writeJSONBodyError(w, err)
		return
	}
	req.Decision = strings.TrimSpace(req.Decision)
	if req.DecisionNote != "" && strings.TrimSpace(req.DecisionNote) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid approval decision"})
		return
	}
	req.DecisionNote = strings.TrimSpace(req.DecisionNote)
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
		case errors.Is(err, workflow.ErrInvalidApproval):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid approval"})
		case errors.Is(err, workflow.ErrInvalidApprovalInput):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid approval decision"})
		case errors.Is(err, workflow.ErrApprovalNotFound):
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "approval not found"})
		case errors.Is(err, workflow.ErrApprovalState), errors.Is(err, documents.ErrInvalidDocumentState):
			writeJSON(w, http.StatusConflict, approvalDecisionResponse{
				Error:           "approval cannot be decided in the current state",
				ApprovalID:      approval.ID,
				Status:          approval.Status,
				QueueCode:       approval.QueueCode,
				DocumentID:      approval.DocumentID,
				DocumentStatus:  string(document.Status),
				DecisionNote:    stringPtr(approval.DecisionNote),
				DecidedByUserID: stringPtr(approval.DecidedByUserID),
				DecidedAt:       timePtr(approval.DecidedAt),
			})
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
