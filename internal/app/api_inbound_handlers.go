package app

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
)

func (h *AgentAPIHandler) handleProcessNextQueuedInboundRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		if errors.Is(err, identityaccess.ErrUnauthorized) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	var req processNextQueuedRequest
	if r.Body != nil {
		defer r.Body.Close()
		if err := decodeJSONBody(r, &req, true); err != nil {
			writeJSONBodyError(w, err)
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
	defer r.Body.Close()
	if err := decodeJSONBody(r, &req, false); err != nil {
		writeJSONBodyError(w, err)
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

	queueForReview := true
	if req.QueueForReview != nil {
		queueForReview = *req.QueueForReview
	}

	if !queueForReview {
		result, err := h.submissionService.SaveInboundDraft(r.Context(), SaveInboundDraftInput{
			OriginType:  req.OriginType,
			Channel:     req.Channel,
			Metadata:    req.Metadata,
			MessageRole: req.Message.MessageRole,
			MessageText: req.Message.TextContent,
			Attachments: attachmentsInput,
			Actor:       actor,
		})
		if err != nil {
			switch {
			case errors.Is(err, intake.ErrInvalidInboundRequest), errors.Is(err, attachments.ErrInvalidAttachment), errors.Is(err, attachments.ErrInvalidLink), errors.Is(err, ErrAttachmentContentEncoding):
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid inbound request"})
			case errors.Is(err, identityaccess.ErrUnauthorized):
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to save inbound draft"})
			}
			return
		}
		response := mapInboundRequestMutationResponse(result.Request)
		response.MessageID = result.Message.ID
		for _, attachment := range result.Attachments {
			response.AttachmentIDs = append(response.AttachmentIDs, attachment.ID)
		}
		writeJSON(w, http.StatusCreated, response)
		return
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

	response := mapInboundRequestMutationResponse(result.Request)
	response.MessageID = result.Message.ID
	for _, attachment := range result.Attachments {
		response.AttachmentIDs = append(response.AttachmentIDs, attachment.ID)
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *AgentAPIHandler) handleInboundRequestAction(w http.ResponseWriter, r *http.Request) {
	if h.submissionService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "submission service unavailable"})
		return
	}

	requestID, action, ok := parseChildActionPath(inboundRequestActionPrefix, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	switch action {
	case "draft":
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}
		defer r.Body.Close()

		var req saveInboundDraftRequest
		if err := decodeJSONBody(r, &req, false); err != nil {
			writeJSONBodyError(w, err)
			return
		}
		req.RequestID = requestID

		attachmentsInput := make([]SubmitInboundRequestAttachmentInput, 0, len(req.Attachments))
		for _, attachment := range req.Attachments {
			attachmentsInput = append(attachmentsInput, SubmitInboundRequestAttachmentInput{
				OriginalFileName: attachment.OriginalFileName,
				MediaType:        attachment.MediaType,
				ContentBase64:    attachment.ContentBase64,
				LinkRole:         attachment.LinkRole,
			})
		}

		result, err := h.submissionService.SaveInboundDraft(r.Context(), SaveInboundDraftInput{
			RequestID:   req.RequestID,
			MessageID:   req.MessageID,
			OriginType:  req.OriginType,
			Channel:     req.Channel,
			Metadata:    req.Metadata,
			MessageRole: req.Message.MessageRole,
			MessageText: req.Message.TextContent,
			Attachments: attachmentsInput,
			Actor:       actor,
		})
		if err != nil {
			switch {
			case errors.Is(err, intake.ErrInvalidInboundRequest), errors.Is(err, intake.ErrInboundRequestState), errors.Is(err, intake.ErrInboundRequestNotFound), errors.Is(err, attachments.ErrInvalidAttachment), errors.Is(err, attachments.ErrInvalidLink), errors.Is(err, ErrAttachmentContentEncoding):
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid inbound request"})
			case errors.Is(err, identityaccess.ErrUnauthorized):
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to save inbound draft"})
			}
			return
		}
		response := mapInboundRequestMutationResponse(result.Request)
		response.MessageID = result.Message.ID
		writeJSON(w, http.StatusOK, response)
		return
	case "queue":
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}
		request, err := h.submissionService.QueueInboundRequest(r.Context(), QueueInboundRequestInput{
			RequestID: requestID,
			Actor:     actor,
		})
		if err != nil {
			switch {
			case errors.Is(err, intake.ErrInvalidInboundRequest), errors.Is(err, intake.ErrInboundRequestState), errors.Is(err, intake.ErrInboundRequestNotFound):
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid inbound request"})
			case errors.Is(err, identityaccess.ErrUnauthorized):
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to queue inbound request"})
			}
			return
		}
		writeJSON(w, http.StatusOK, mapInboundRequestMutationResponse(request))
		return
	case "cancel":
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}
		var req inboundRequestActionRequest
		if r.Body != nil {
			defer r.Body.Close()
			if err := decodeJSONBody(r, &req, true); err != nil {
				writeJSONBodyError(w, err)
				return
			}
		}
		request, err := h.submissionService.CancelInboundRequest(r.Context(), CancelInboundRequestInput{
			RequestID: requestID,
			Reason:    req.Reason,
			Actor:     actor,
		})
		if err != nil {
			switch {
			case errors.Is(err, intake.ErrInboundRequestState), errors.Is(err, intake.ErrInboundRequestNotFound):
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid inbound request"})
			case errors.Is(err, identityaccess.ErrUnauthorized):
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to cancel inbound request"})
			}
			return
		}
		writeJSON(w, http.StatusOK, mapInboundRequestMutationResponse(request))
		return
	case "amend":
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}
		request, err := h.submissionService.AmendInboundRequest(r.Context(), AmendInboundRequestInput{
			RequestID: requestID,
			Actor:     actor,
		})
		if err != nil {
			switch {
			case errors.Is(err, intake.ErrInboundRequestState), errors.Is(err, intake.ErrInboundRequestNotFound):
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid inbound request"})
			case errors.Is(err, identityaccess.ErrUnauthorized):
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to amend inbound request"})
			}
			return
		}
		writeJSON(w, http.StatusOK, mapInboundRequestMutationResponse(request))
		return
	case "delete":
		if r.Method != http.MethodDelete {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}
		err := h.submissionService.DeleteInboundDraft(r.Context(), DeleteInboundDraftInput{
			RequestID: requestID,
			Actor:     actor,
		})
		if err != nil {
			switch {
			case errors.Is(err, intake.ErrInboundRequestState), errors.Is(err, intake.ErrInboundRequestNotFound):
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid inbound request"})
			case errors.Is(err, identityaccess.ErrUnauthorized):
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to delete inbound draft"})
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
		return
	default:
		http.NotFound(w, r)
		return
	}
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
		case errors.Is(err, attachments.ErrInvalidAttachment):
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid attachment"})
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
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(attachment.Content)
}
