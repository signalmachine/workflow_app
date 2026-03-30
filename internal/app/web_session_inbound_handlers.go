package app

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/reporting"
	"workflow_app/internal/workflow"
)

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
		} else {
			sortInboundRequestStatusSummaries(data.InboundSummary)
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

func (h *AgentAPIHandler) handleWebLogin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webLoginPath {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if _, err := h.sessionContextFromRequest(r); err == nil {
			http.Redirect(w, r, webAppPath, http.StatusSeeOther)
			return
		}
		h.renderWebPage(w, webPageData{
			Title:      "workflow_app",
			Notice:     strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:      strings.TrimSpace(r.URL.Query().Get("error")),
			ShowLogin:  true,
			LoginPath:  webLoginPath,
			ActivePath: webLoginPath,
		})
		return
	case http.MethodPost:
	default:
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
		Password:    r.FormValue("password"),
		DeviceLabel: deviceLabel,
		ExpiresAt:   time.Now().UTC().Add(browserSessionDuration),
	})
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("invalid session credentials"), http.StatusSeeOther)
		return
	}

	setSessionCookies(w, sessionCookiesShouldBeSecure(r), session.Session.ID, session.RefreshToken, session.Session.ExpiresAt)
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
	clearSessionCookies(w, sessionCookiesShouldBeSecure(r))
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

	files, err := parseMultipartAttachments(r.MultipartForm)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("failed to read attachment"), http.StatusSeeOther)
		return
	}

	intent := strings.TrimSpace(r.FormValue("intent"))
	requestID := strings.TrimSpace(r.FormValue("request_id"))
	messageID := strings.TrimSpace(r.FormValue("message_id"))
	returnTo := sanitizeWebReturnPath(r.FormValue("return_to"))
	if returnTo == "" {
		if requestID != "" {
			returnTo = webInboundDetailPrefix + url.PathEscape(requestID)
		} else {
			returnTo = webAppPath
		}
	}

	switch intent {
	case "save_draft":
		result, err := h.submissionService.SaveInboundDraft(r.Context(), SaveInboundDraftInput{
			RequestID:   requestID,
			MessageID:   messageID,
			OriginType:  intake.OriginHuman,
			Channel:     "browser",
			Metadata:    map[string]any{"submitter_label": strings.TrimSpace(r.FormValue("submitter_label"))},
			MessageRole: intake.MessageRoleRequest,
			MessageText: strings.TrimSpace(r.FormValue("message_text")),
			Attachments: files,
			Actor:       actor,
		})
		if err != nil {
			http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to save inbound draft"), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(result.Request.RequestReference)+"?notice="+url.QueryEscape("Draft saved."), http.StatusSeeOther)
		return
	default:
		if requestID != "" {
			saved, err := h.submissionService.SaveInboundDraft(r.Context(), SaveInboundDraftInput{
				RequestID:   requestID,
				MessageID:   messageID,
				OriginType:  intake.OriginHuman,
				Channel:     "browser",
				Metadata:    map[string]any{"submitter_label": strings.TrimSpace(r.FormValue("submitter_label"))},
				MessageRole: intake.MessageRoleRequest,
				MessageText: strings.TrimSpace(r.FormValue("message_text")),
				Attachments: files,
				Actor:       actor,
			})
			if err != nil {
				http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to save inbound draft"), http.StatusSeeOther)
				return
			}
			queued, err := h.submissionService.QueueInboundRequest(r.Context(), QueueInboundRequestInput{
				RequestID: saved.Request.ID,
				Actor:     actor,
			})
			if err != nil {
				http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to queue inbound request"), http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(queued.RequestReference)+"?notice="+url.QueryEscape("Inbound request queued."), http.StatusSeeOther)
			return
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
			http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to submit inbound request"), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(result.Request.RequestReference)+"?notice="+url.QueryEscape("Inbound request submitted."), http.StatusSeeOther)
		return
	}
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
		if result.Request.RequestReference != "" {
			http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(result.Request.RequestReference)+"?error="+url.QueryEscape("failed to process queued inbound request"), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("failed to process queued inbound request"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(result.Request.RequestReference)+"?notice="+url.QueryEscape("Queued inbound request processed."), http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebInboundRequestDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		requestID, action, ok := parseChildActionPath(webInboundDetailPrefix, r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}
		h.handleWebInboundRequestAction(w, r, requestID, action)
		return
	}
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
	populateInboundRequestDetailLookup(&input, lookup)

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
			Session:                sessionContext,
			Notice:                 strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:                  strings.TrimSpace(r.URL.Query().Get("error")),
			Detail:                 detail,
			EditableMessageID:      editableInboundMessageID(detail),
			EditableMessageText:    editableInboundMessageText(detail),
			EditableSubmitterLabel: inboundRequestMetadataString(detail.Request.Metadata, "submitter_label"),
		},
	})
}

func (h *AgentAPIHandler) handleWebInboundRequestAction(w http.ResponseWriter, r *http.Request, requestID, action string) {
	if h.submissionService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("submission service unavailable"), http.StatusSeeOther)
		return
	}

	actor, err := h.actorFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("unauthorized"), http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("invalid inbound request form"), http.StatusSeeOther)
		return
	}

	returnTo := sanitizeWebReturnPath(r.FormValue("return_to"))
	if returnTo == "" {
		returnTo = webInboundDetailPrefix + url.PathEscape(requestID)
	}

	switch action {
	case "queue":
		request, err := h.submissionService.QueueInboundRequest(r.Context(), QueueInboundRequestInput{
			RequestID: requestID,
			Actor:     actor,
		})
		if err != nil {
			http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to queue inbound request"), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(request.RequestReference)+"?notice="+url.QueryEscape("Inbound request queued."), http.StatusSeeOther)
		return
	case "cancel":
		request, err := h.submissionService.CancelInboundRequest(r.Context(), CancelInboundRequestInput{
			RequestID: requestID,
			Reason:    strings.TrimSpace(r.FormValue("reason")),
			Actor:     actor,
		})
		if err != nil {
			http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to cancel inbound request"), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(request.RequestReference)+"?notice="+url.QueryEscape("Inbound request cancelled."), http.StatusSeeOther)
		return
	case "amend":
		request, err := h.submissionService.AmendInboundRequest(r.Context(), AmendInboundRequestInput{
			RequestID: requestID,
			Actor:     actor,
		})
		if err != nil {
			http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to return request to draft"), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, webInboundDetailPrefix+url.PathEscape(request.RequestReference)+"?notice="+url.QueryEscape("Inbound request returned to draft."), http.StatusSeeOther)
		return
	case "delete":
		if err := h.submissionService.DeleteInboundDraft(r.Context(), DeleteInboundDraftInput{
			RequestID: requestID,
			Actor:     actor,
		}); err != nil {
			http.Redirect(w, r, appendWebMessage(returnTo, "error", "failed to delete inbound draft"), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, appendWebMessage(webInboundRequestsPath, "notice", "Draft deleted."), http.StatusSeeOther)
		return
	default:
		http.NotFound(w, r)
		return
	}
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
