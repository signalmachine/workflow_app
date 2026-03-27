package app_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/ai"
	"workflow_app/internal/app"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/inventoryops"
	"workflow_app/internal/reporting"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
	"workflow_app/internal/workforce"
	"workflow_app/internal/workorders"
)

func TestAgentAPISessionLoginCurrentSessionAndLogoutIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandler(db)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{
		"org_slug":"`+orgSlug+`",
		"email":"`+userEmail+`",
		"device_label":"browser-integration"
	}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)

	if loginRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	var loginResponse struct {
		SessionID       string `json:"session_id"`
		OrgID           string `json:"org_id"`
		OrgSlug         string `json:"org_slug"`
		UserID          string `json:"user_id"`
		UserEmail       string `json:"user_email"`
		UserDisplayName string `json:"user_display_name"`
		RoleCode        string `json:"role_code"`
		DeviceLabel     string `json:"device_label"`
	}
	if err := json.Unmarshal(loginRecorder.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResponse.OrgID != orgID || loginResponse.UserID != operatorUserID {
		t.Fatalf("unexpected login identity: %+v", loginResponse)
	}
	if loginResponse.OrgSlug != orgSlug || loginResponse.UserEmail != userEmail {
		t.Fatalf("unexpected login profile: %+v", loginResponse)
	}
	if loginResponse.RoleCode != identityaccess.RoleOperator || loginResponse.DeviceLabel != "browser-integration" {
		t.Fatalf("unexpected login session metadata: %+v", loginResponse)
	}

	currentReq := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	applyResponseCookies(currentReq, loginRecorder.Result().Cookies())
	currentRecorder := httptest.NewRecorder()
	handler.ServeHTTP(currentRecorder, currentReq)

	if currentRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected current-session status: got %d body=%s", currentRecorder.Code, currentRecorder.Body.String())
	}

	var currentResponse struct {
		SessionID string `json:"session_id"`
		OrgID     string `json:"org_id"`
		UserID    string `json:"user_id"`
	}
	if err := json.Unmarshal(currentRecorder.Body.Bytes(), &currentResponse); err != nil {
		t.Fatalf("decode current-session response: %v", err)
	}
	if currentResponse.SessionID != loginResponse.SessionID || currentResponse.OrgID != orgID || currentResponse.UserID != operatorUserID {
		t.Fatalf("unexpected current session payload: %+v", currentResponse)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/session/logout", nil)
	applyResponseCookies(logoutReq, loginRecorder.Result().Cookies())
	logoutRecorder := httptest.NewRecorder()
	handler.ServeHTTP(logoutRecorder, logoutReq)

	if logoutRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected logout status: got %d body=%s", logoutRecorder.Code, logoutRecorder.Body.String())
	}

	postLogoutReq := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	applyResponseCookies(postLogoutReq, loginRecorder.Result().Cookies())
	postLogoutRecorder := httptest.NewRecorder()
	handler.ServeHTTP(postLogoutRecorder, postLogoutReq)
	if postLogoutRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected post-logout current-session status: got %d body=%s", postLogoutRecorder.Code, postLogoutRecorder.Body.String())
	}
}

func TestAgentAPISubmitInboundRequestWithBrowserSessionCookies(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandler(db)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{
		"org_slug":"`+orgSlug+`",
		"email":"`+userEmail+`"
	}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	body := bytes.NewBufferString(`{
		"origin_type":"human",
		"channel":"browser",
		"metadata":{"submitter_label":"front desk"},
		"message":{"message_role":"request","text_content":"Submit this request with cookie auth."}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", body)
	req.Header.Set("Content-Type", "application/json")
	applyResponseCookies(req, loginRecorder.Result().Cookies())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}
	if response.RequestID == "" || response.Status != "queued" {
		t.Fatalf("unexpected submit response: %+v", response)
	}

	var storedSessionID sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT session_id FROM ai.inbound_requests WHERE id = $1`, response.RequestID).Scan(&storedSessionID); err != nil {
		t.Fatalf("load submitted request session: %v", err)
	}
	if !storedSessionID.Valid {
		t.Fatal("expected persisted browser session linkage")
	}
}

func TestAgentBrowserAppFlowIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, adminUserID)

	processor, err := app.NewAgentProcessor(db, fakeCoordinatorProvider{
		output: ai.CoordinatorProviderOutput{
			ProviderName:       "openai",
			ProviderResponseID: "resp_browser_flow_123",
			Model:              "gpt-5.2",
			Summary:            "Operator review is required for the urgent pump issue.",
			Priority:           "urgent",
			ArtifactTitle:      "Inbound request review brief",
			ArtifactBody:       "The request describes an urgent equipment problem that should be reviewed immediately.",
			Rationale: []string{
				"The request describes a time-sensitive equipment failure.",
			},
			NextActions: []string{
				"Confirm the site details and route controlled follow-up.",
			},
			SpecialistDelegation: &ai.CoordinatorSpecialistDelegation{
				CapabilityCode: "inbound_request.approval_triage",
				Reason:         "The request needs narrower approval-focused review framing before action.",
			},
		},
	})
	if err != nil {
		t.Fatalf("new agent processor: %v", err)
	}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	handler := app.NewAgentAPIHandlerWithDependencies(
		func() (app.ProcessNextQueuedInboundRequester, error) { return processor, nil },
		app.NewSubmissionService(db),
		reporting.NewService(db),
		workflowService,
		identityaccess.NewService(db),
	)

	adminSession := startSession(t, ctx, db, orgID, adminUserID)
	createPendingApproval(t, ctx, documentService, workflowService, identityaccess.Actor{
		OrgID:     orgID,
		UserID:    adminUserID,
		SessionID: adminSession.ID,
	})

	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected dashboard status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "Start browser session") {
		t.Fatalf("expected login page, body=%s", recorder.Body.String())
	}

	loginReq := httptest.NewRequest(
		http.MethodPost,
		"/app/login",
		strings.NewReader("org_slug="+orgSlug+"&email="+userEmail+"&device_label=browser-ui"),
	)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected web login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	submitReq := newMultipartRequest(t, http.MethodPost, "/app/inbound-requests", map[string]string{
		"submitter_label": "front desk",
		"message_text":    "The warehouse pump failed and needs review.",
	}, map[string]multipartUpload{
		"attachments": {
			FileName:    "pump-note.txt",
			ContentType: "text/plain",
			Content:     []byte("urgent pump failure details"),
		},
	})
	applyResponseCookies(submitReq, loginRecorder.Result().Cookies())
	submitRecorder := httptest.NewRecorder()
	handler.ServeHTTP(submitRecorder, submitReq)
	if submitRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected web submit status: got %d body=%s", submitRecorder.Code, submitRecorder.Body.String())
	}

	detailPath := submitRecorder.Result().Header.Get("Location")
	if !strings.HasPrefix(detailPath, "/app/inbound-requests/REQ-") {
		t.Fatalf("unexpected detail redirect: %s", detailPath)
	}

	detailReq := httptest.NewRequest(http.MethodGet, detailPath, nil)
	applyResponseCookies(detailReq, loginRecorder.Result().Cookies())
	detailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(detailRecorder, detailReq)
	if detailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected detail status before processing: got %d body=%s", detailRecorder.Code, detailRecorder.Body.String())
	}
	requireContains(t, detailRecorder.Body.String(), "pump-note.txt")
	requireContains(t, detailRecorder.Body.String(), "The warehouse pump failed and needs review.")

	processReq := httptest.NewRequest(http.MethodPost, "/app/agent/process-next-queued-inbound-request", nil)
	applyResponseCookies(processReq, loginRecorder.Result().Cookies())
	processRecorder := httptest.NewRecorder()
	handler.ServeHTTP(processRecorder, processReq)
	if processRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected web process status: got %d body=%s", processRecorder.Code, processRecorder.Body.String())
	}

	processedDetailPath := processRecorder.Result().Header.Get("Location")
	processedDetailReq := httptest.NewRequest(http.MethodGet, processedDetailPath, nil)
	applyResponseCookies(processedDetailReq, loginRecorder.Result().Cookies())
	processedDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(processedDetailRecorder, processedDetailReq)
	if processedDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected detail status after processing: got %d body=%s", processedDetailRecorder.Code, processedDetailRecorder.Body.String())
	}
	processedDetailBody := processedDetailRecorder.Body.String()
	requireContains(t, processedDetailBody, "Operator review is required for the urgent pump issue.")
	requireContains(t, processedDetailBody, "Inbound request review brief")
	requireContains(t, processedDetailBody, "AI steps")
	requireContains(t, processedDetailBody, "Execute provider-backed coordinator review")
	requireContains(t, processedDetailBody, "Delegations")
	requireContains(t, processedDetailBody, "inbound_request.approval_triage")
	requireContains(t, processedDetailBody, "The request needs narrower approval-focused review framing before action.")
	requireContains(t, processedDetailBody, "/app/review/inbound-requests?request_reference=")
	requireContains(t, processedDetailBody, "/app/review/audit?entity_type=ai.inbound_request&amp;entity_id=")
	requireContains(t, processedDetailBody, "/app/review/proposals/")
	requireContains(t, processedDetailBody, "/app/review/proposals?recommendation_id=")
	requireContains(t, processedDetailBody, "/app/review/audit?entity_type=ai.agent_recommendation&amp;entity_id=")

	dashboardReq := httptest.NewRequest(http.MethodGet, "/app", nil)
	applyResponseCookies(dashboardReq, loginRecorder.Result().Cookies())
	dashboardRecorder := httptest.NewRecorder()
	handler.ServeHTTP(dashboardRecorder, dashboardReq)
	if dashboardRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected dashboard status after processing: got %d body=%s", dashboardRecorder.Code, dashboardRecorder.Body.String())
	}
	dashboardBody := dashboardRecorder.Body.String()
	requireContains(t, dashboardBody, "/app/review/approvals?status=pending")
	requireContains(t, dashboardBody, "/app/review/approvals/")
	requireContains(t, dashboardBody, "/app/review/audit?entity_type=documents.document&amp;entity_id=")

	approvalQueueReq := httptest.NewRequest(http.MethodGet, "/api/review/approval-queue?status=pending", nil)
	applyResponseCookies(approvalQueueReq, loginRecorder.Result().Cookies())
	approvalQueueRecorder := httptest.NewRecorder()
	handler.ServeHTTP(approvalQueueRecorder, approvalQueueReq)
	if approvalQueueRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected approval queue status with cookies: got %d body=%s", approvalQueueRecorder.Code, approvalQueueRecorder.Body.String())
	}

	var approvalQueueResponse struct {
		Items []struct {
			ApprovalID string `json:"approval_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(approvalQueueRecorder.Body.Bytes(), &approvalQueueResponse); err != nil {
		t.Fatalf("decode approval queue response: %v", err)
	}
	if len(approvalQueueResponse.Items) != 1 {
		t.Fatalf("expected one pending approval, got %d", len(approvalQueueResponse.Items))
	}

	approvalReq := httptest.NewRequest(
		http.MethodPost,
		"/app/approvals/"+approvalQueueResponse.Items[0].ApprovalID+"/decision",
		strings.NewReader("decision=approved&decision_note=Looks+correct.&return_to="+detailPath),
	)
	approvalReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	applyResponseCookies(approvalReq, loginRecorder.Result().Cookies())
	approvalRecorder := httptest.NewRecorder()
	handler.ServeHTTP(approvalRecorder, approvalReq)
	if approvalRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected web approval decision status: got %d body=%s", approvalRecorder.Code, approvalRecorder.Body.String())
	}

	var approvalStatus string
	if err := db.QueryRowContext(ctx, `SELECT status FROM workflow.approvals WHERE id = $1`, approvalQueueResponse.Items[0].ApprovalID).Scan(&approvalStatus); err != nil {
		t.Fatalf("load approval status: %v", err)
	}
	if approvalStatus != "approved" {
		t.Fatalf("unexpected approval status: %s", approvalStatus)
	}
}

func TestAgentBrowserDraftLifecycleIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandler(db)

	loginReq := httptest.NewRequest(
		http.MethodPost,
		"/app/login",
		strings.NewReader("org_slug="+orgSlug+"&email="+userEmail+"&device_label=browser-draft"),
	)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected web login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	saveDraftReq := newMultipartRequest(t, http.MethodPost, "/app/inbound-requests", map[string]string{
		"submitter_label": "front desk",
		"message_text":    "Initial draft request from browser flow.",
		"intent":          "save_draft",
	}, map[string]multipartUpload{
		"attachments": {
			FileName:    "draft-note.txt",
			ContentType: "text/plain",
			Content:     []byte("draft attachment body"),
		},
	})
	applyResponseCookies(saveDraftReq, loginRecorder.Result().Cookies())
	saveDraftRecorder := httptest.NewRecorder()
	handler.ServeHTTP(saveDraftRecorder, saveDraftReq)
	if saveDraftRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected save-draft status: got %d body=%s", saveDraftRecorder.Code, saveDraftRecorder.Body.String())
	}

	draftDetailPath := saveDraftRecorder.Result().Header.Get("Location")
	requireContains(t, draftDetailPath, "/app/inbound-requests/REQ-")
	requireContains(t, draftDetailPath, "notice=Draft+saved.")

	draftDetailReq := httptest.NewRequest(http.MethodGet, draftDetailPath, nil)
	applyResponseCookies(draftDetailReq, loginRecorder.Result().Cookies())
	draftDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(draftDetailRecorder, draftDetailReq)
	if draftDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected draft detail status: got %d body=%s", draftDetailRecorder.Code, draftDetailRecorder.Body.String())
	}

	draftDetailBody := draftDetailRecorder.Body.String()
	requireContains(t, draftDetailBody, "Edit draft")
	requireContains(t, draftDetailBody, "draft-note.txt")
	requireContains(t, draftDetailBody, "Delete draft")
	requireContains(t, draftDetailBody, "Initial draft request from browser flow.")

	requestID := requireHiddenInputValue(t, draftDetailBody, "request_id")
	messageID := requireHiddenInputValue(t, draftDetailBody, "message_id")
	requestReference := requireRequestReferenceFromPath(t, draftDetailPath)

	queueReq := newMultipartRequest(t, http.MethodPost, "/app/inbound-requests", map[string]string{
		"request_id":      requestID,
		"message_id":      messageID,
		"submitter_label": "front desk",
		"message_text":    "Updated and queued from draft.",
		"intent":          "queue",
		"return_to":       "/app/inbound-requests/" + requestReference,
	}, nil)
	applyResponseCookies(queueReq, loginRecorder.Result().Cookies())
	queueRecorder := httptest.NewRecorder()
	handler.ServeHTTP(queueRecorder, queueReq)
	if queueRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected queue status: got %d body=%s", queueRecorder.Code, queueRecorder.Body.String())
	}

	queuedDetailPath := queueRecorder.Result().Header.Get("Location")
	requireContains(t, queuedDetailPath, "/app/inbound-requests/"+requestReference)
	requireContains(t, queuedDetailPath, "notice=Inbound+request+queued.")

	queuedDetailReq := httptest.NewRequest(http.MethodGet, queuedDetailPath, nil)
	applyResponseCookies(queuedDetailReq, loginRecorder.Result().Cookies())
	queuedDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(queuedDetailRecorder, queuedDetailReq)
	if queuedDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected queued detail status: got %d body=%s", queuedDetailRecorder.Code, queuedDetailRecorder.Body.String())
	}

	queuedDetailBody := queuedDetailRecorder.Body.String()
	requireContains(t, queuedDetailBody, "Queued request actions")
	requireContains(t, queuedDetailBody, "Cancel request")
	requireContains(t, queuedDetailBody, "Return to draft")
	requireContains(t, queuedDetailBody, "Updated and queued from draft.")

	cancelBody := strings.NewReader("reason=operator+paused+request&return_to=" + url.QueryEscape("/app/inbound-requests/"+requestReference))
	cancelReq := httptest.NewRequest(http.MethodPost, "/app/inbound-requests/"+requestID+"/cancel", cancelBody)
	cancelReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	applyResponseCookies(cancelReq, loginRecorder.Result().Cookies())
	cancelRecorder := httptest.NewRecorder()
	handler.ServeHTTP(cancelRecorder, cancelReq)
	if cancelRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected cancel status: got %d body=%s", cancelRecorder.Code, cancelRecorder.Body.String())
	}

	cancelledDetailPath := cancelRecorder.Result().Header.Get("Location")
	requireContains(t, cancelledDetailPath, "/app/inbound-requests/"+requestReference)
	requireContains(t, cancelledDetailPath, "notice=Inbound+request+cancelled.")

	cancelledDetailReq := httptest.NewRequest(http.MethodGet, cancelledDetailPath, nil)
	applyResponseCookies(cancelledDetailReq, loginRecorder.Result().Cookies())
	cancelledDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(cancelledDetailRecorder, cancelledDetailReq)
	if cancelledDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected cancelled detail status: got %d body=%s", cancelledDetailRecorder.Code, cancelledDetailRecorder.Body.String())
	}

	cancelledDetailBody := cancelledDetailRecorder.Body.String()
	requireContains(t, cancelledDetailBody, "Cancelled request actions")
	requireContains(t, cancelledDetailBody, "Amend back to draft")
	requireContains(t, cancelledDetailBody, "operator paused request")

	amendReq := httptest.NewRequest(http.MethodPost, "/app/inbound-requests/"+requestID+"/amend", strings.NewReader("return_to="+url.QueryEscape("/app/inbound-requests/"+requestReference)))
	amendReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	applyResponseCookies(amendReq, loginRecorder.Result().Cookies())
	amendRecorder := httptest.NewRecorder()
	handler.ServeHTTP(amendRecorder, amendReq)
	if amendRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected amend status: got %d body=%s", amendRecorder.Code, amendRecorder.Body.String())
	}

	amendedDetailPath := amendRecorder.Result().Header.Get("Location")
	requireContains(t, amendedDetailPath, "/app/inbound-requests/"+requestReference)
	requireContains(t, amendedDetailPath, "notice=Inbound+request+returned+to+draft.")

	amendedDetailReq := httptest.NewRequest(http.MethodGet, amendedDetailPath, nil)
	applyResponseCookies(amendedDetailReq, loginRecorder.Result().Cookies())
	amendedDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(amendedDetailRecorder, amendedDetailReq)
	if amendedDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected amended detail status: got %d body=%s", amendedDetailRecorder.Code, amendedDetailRecorder.Body.String())
	}

	amendedDetailBody := amendedDetailRecorder.Body.String()
	requireContains(t, amendedDetailBody, "Edit draft")
	requireContains(t, amendedDetailBody, "Updated and queued from draft.")

	deleteReq := httptest.NewRequest(http.MethodPost, "/app/inbound-requests/"+requestID+"/delete", strings.NewReader("return_to="+url.QueryEscape("/app/inbound-requests/"+requestReference)))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	applyResponseCookies(deleteReq, loginRecorder.Result().Cookies())
	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteReq)
	if deleteRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected delete status: got %d body=%s", deleteRecorder.Code, deleteRecorder.Body.String())
	}

	inboundListPath := deleteRecorder.Result().Header.Get("Location")
	requireContains(t, inboundListPath, "/app/review/inbound-requests")
	requireContains(t, inboundListPath, "notice=Draft+deleted.")

	inboundListReq := httptest.NewRequest(http.MethodGet, inboundListPath, nil)
	applyResponseCookies(inboundListReq, loginRecorder.Result().Cookies())
	inboundListRecorder := httptest.NewRecorder()
	handler.ServeHTTP(inboundListRecorder, inboundListReq)
	if inboundListRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound list status: got %d body=%s", inboundListRecorder.Code, inboundListRecorder.Body.String())
	}

	inboundListBody := inboundListRecorder.Body.String()
	requireContains(t, inboundListBody, "Draft deleted.")
	requireNotContains(t, inboundListBody, requestReference)

	var remaining int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.inbound_requests WHERE id = $1`, requestID).Scan(&remaining); err != nil {
		t.Fatalf("count remaining requests: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected deleted draft to be removed, found %d", remaining)
	}
}

func TestAgentBrowserDashboardStatusCoverageIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, operatorUserID)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}
	intakeService := intake.NewService(db)

	draft, err := intakeService.CreateDraft(ctx, intake.CreateDraftInput{
		OriginType: intake.OriginHuman,
		Channel:    "browser",
		Metadata:   map[string]any{"submitter_label": "draft-test"},
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("create draft request: %v", err)
	}
	if _, err := intakeService.AddMessage(ctx, intake.AddMessageInput{
		RequestID:   draft.ID,
		MessageRole: intake.MessageRoleRequest,
		TextContent: "Draft request for browser status coverage.",
		Actor:       operator,
	}); err != nil {
		t.Fatalf("add draft message: %v", err)
	}

	processing := createQueuedRequest(t, ctx, db, operator, "Processing request for browser status coverage.")
	processing, err = intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("claim processing request: %v", err)
	}

	failed := createQueuedRequest(t, ctx, db, operator, "Failed request for browser status coverage.")
	failed, err = intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("claim failed request: %v", err)
	}
	failed, err = intakeService.AdvanceRequest(ctx, intake.AdvanceRequestInput{
		RequestID:     failed.ID,
		Status:        intake.StatusFailed,
		FailureReason: "provider timeout during browser review",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("advance failed request: %v", err)
	}

	cancelled := createQueuedRequest(t, ctx, db, operator, "Cancelled request for browser status coverage.")
	cancelled, err = intakeService.CancelRequest(ctx, intake.CancelRequestInput{
		RequestID: cancelled.ID,
		Reason:    "operator withdrew before processing",
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("cancel request: %v", err)
	}

	processed := createQueuedRequest(t, ctx, db, operator, "Processed request for browser status coverage.")
	processed, err = intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("claim processed request: %v", err)
	}
	processed, err = intakeService.AdvanceRequest(ctx, intake.AdvanceRequestInput{
		RequestID: processed.ID,
		Status:    intake.StatusProcessed,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("advance processed request: %v", err)
	}

	completed := createQueuedRequest(t, ctx, db, operator, "Completed request for browser status coverage.")
	completed, err = intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("claim completed request: %v", err)
	}
	completed, err = intakeService.AdvanceRequest(ctx, intake.AdvanceRequestInput{
		RequestID: completed.ID,
		Status:    intake.StatusCompleted,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("advance completed request: %v", err)
	}

	queued := createQueuedRequest(t, ctx, db, operator, "Queued request for browser status coverage.")

	handler := app.NewAgentAPIHandler(db)
	loginReq := httptest.NewRequest(
		http.MethodPost,
		"/app/login",
		strings.NewReader("org_slug="+orgSlug+"&email="+userEmail+"&device_label=browser-status"),
	)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected web login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	dashboardReq := httptest.NewRequest(http.MethodGet, "/app", nil)
	applyResponseCookies(dashboardReq, loginRecorder.Result().Cookies())
	dashboardRecorder := httptest.NewRecorder()
	handler.ServeHTTP(dashboardRecorder, dashboardReq)
	if dashboardRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected dashboard status: got %d body=%s", dashboardRecorder.Code, dashboardRecorder.Body.String())
	}

	dashboardBody := dashboardRecorder.Body.String()
	requireContains(t, dashboardBody, "/app/review/inbound-requests?status=draft")
	requireContains(t, dashboardBody, "/app/review/inbound-requests?status=queued")
	requireContains(t, dashboardBody, "/app/review/inbound-requests?status=processing")
	requireContains(t, dashboardBody, "/app/review/inbound-requests?status=failed")
	requireContains(t, dashboardBody, "/app/review/inbound-requests?status=cancelled")
	requireContains(t, dashboardBody, "/app/review/inbound-requests?status=processed")
	requireContains(t, dashboardBody, "/app/review/inbound-requests?status=completed")
	requireContains(t, dashboardBody, "provider timeout during browser review")
	requireContains(t, dashboardBody, "operator withdrew before processing")
	requireContains(t, dashboardBody, "Continue drafts")
	requireContains(t, dashboardBody, "Open queued requests")
	requireContains(t, dashboardBody, "Watch in-flight requests")
	requireContains(t, dashboardBody, "Review failures")
	requireContains(t, dashboardBody, "Recover cancellations")
	requireContains(t, dashboardBody, "Review outcomes")

	completedReq := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests?status=completed", nil)
	applyResponseCookies(completedReq, loginRecorder.Result().Cookies())
	completedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(completedRecorder, completedReq)
	if completedRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected completed review status: got %d body=%s", completedRecorder.Code, completedRecorder.Body.String())
	}

	completedBody := completedRecorder.Body.String()
	requireContains(t, completedBody, "Inbound-request review")
	requireContains(t, completedBody, completed.RequestReference)
	requireContains(t, completedBody, "completed")

	failedReq := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests?status=failed", nil)
	applyResponseCookies(failedReq, loginRecorder.Result().Cookies())
	failedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(failedRecorder, failedReq)
	if failedRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected failed review status: got %d body=%s", failedRecorder.Code, failedRecorder.Body.String())
	}

	failedBody := failedRecorder.Body.String()
	requireContains(t, failedBody, failed.RequestReference)
	requireContains(t, failedBody, "provider timeout during browser review")

	cancelledReq := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests?status=cancelled", nil)
	applyResponseCookies(cancelledReq, loginRecorder.Result().Cookies())
	cancelledRecorder := httptest.NewRecorder()
	handler.ServeHTTP(cancelledRecorder, cancelledReq)
	if cancelledRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected cancelled review status: got %d body=%s", cancelledRecorder.Code, cancelledRecorder.Body.String())
	}

	cancelledBody := cancelledRecorder.Body.String()
	requireContains(t, cancelledBody, cancelled.RequestReference)
	requireContains(t, cancelledBody, "Manage lifecycle")

	processingReq := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests?status=processing", nil)
	applyResponseCookies(processingReq, loginRecorder.Result().Cookies())
	processingRecorder := httptest.NewRecorder()
	handler.ServeHTTP(processingRecorder, processingReq)
	if processingRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected processing review status: got %d body=%s", processingRecorder.Code, processingRecorder.Body.String())
	}

	processingBody := processingRecorder.Body.String()
	requireContains(t, processingBody, processing.RequestReference)
	requireContains(t, processingBody, "processing")

	queuedReq := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests?status=queued", nil)
	applyResponseCookies(queuedReq, loginRecorder.Result().Cookies())
	queuedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(queuedRecorder, queuedReq)
	if queuedRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected queued review status: got %d body=%s", queuedRecorder.Code, queuedRecorder.Body.String())
	}

	queuedBody := queuedRecorder.Body.String()
	requireContains(t, queuedBody, queued.RequestReference)
	requireContains(t, queuedBody, "Manage lifecycle")

	processedReq := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests?status=processed", nil)
	applyResponseCookies(processedReq, loginRecorder.Result().Cookies())
	processedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(processedRecorder, processedReq)
	if processedRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected processed review status: got %d body=%s", processedRecorder.Code, processedRecorder.Body.String())
	}

	processedBody := processedRecorder.Body.String()
	requireContains(t, processedBody, processed.RequestReference)
	requireContains(t, processedBody, "processed")

	draftReq := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests?status=draft", nil)
	applyResponseCookies(draftReq, loginRecorder.Result().Cookies())
	draftRecorder := httptest.NewRecorder()
	handler.ServeHTTP(draftRecorder, draftReq)
	if draftRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected draft review status: got %d body=%s", draftRecorder.Code, draftRecorder.Body.String())
	}

	draftBody := draftRecorder.Body.String()
	requireContains(t, draftBody, draft.RequestReference)
	requireContains(t, draftBody, "Continue draft")
}

func TestAgentBrowserReportingIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, adminUserID)
	_, approverUserID := seedOrgAndUserInOrg(t, ctx, db, identityaccess.RoleApprover, orgID)

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	accountingService := accounting.NewService(db, documentService)
	inventoryService := inventoryops.NewService(db)
	workOrderService := workorders.NewService(db, documentService)
	workforceService := workforce.NewService(db)
	adminSession := startSession(t, ctx, db, orgID, adminUserID)
	approverSession := startSession(t, ctx, db, orgID, approverUserID)
	adminActor := identityaccess.Actor{OrgID: orgID, UserID: adminUserID, SessionID: adminSession.ID}
	approverActor := identityaccess.Actor{OrgID: orgID, UserID: approverUserID, SessionID: approverSession.ID}

	postApprovedGSTInvoice(t, ctx, accountingService, documentService, workflowService, adminActor, approverActor)
	workOrder := seedBrowserReviewData(t, ctx, documentService, workflowService, accountingService, inventoryService, workOrderService, workforceService, adminActor, approverActor)

	var gstInvoiceDocumentID string
	if err := db.QueryRowContext(ctx, `SELECT id FROM documents.documents WHERE org_id = $1 AND title = $2`, orgID, "Posted GST invoice").Scan(&gstInvoiceDocumentID); err != nil {
		t.Fatalf("load gst invoice document id: %v", err)
	}
	var gstInvoiceJournalEntryID string
	if err := db.QueryRowContext(ctx, `SELECT id FROM accounting.journal_entries WHERE org_id = $1 AND source_document_id = $2`, orgID, gstInvoiceDocumentID).Scan(&gstInvoiceJournalEntryID); err != nil {
		t.Fatalf("load gst invoice journal entry id: %v", err)
	}
	var issueMovementID string
	if err := db.QueryRowContext(ctx, `
SELECT m.id
FROM inventory_ops.movements m
JOIN documents.documents d
	ON d.id = m.document_id
   AND d.org_id = m.org_id
WHERE m.org_id = $1
  AND d.title = $2
ORDER BY m.created_at DESC
LIMIT 1`,
		orgID,
		"Inventory issue",
	).Scan(&issueMovementID); err != nil {
		t.Fatalf("load inventory issue movement id: %v", err)
	}

	handler := app.NewAgentAPIHandler(db)

	loginReq := httptest.NewRequest(
		http.MethodPost,
		"/app/login",
		strings.NewReader("org_slug="+orgSlug+"&email="+userEmail+"&device_label=browser-ui"),
	)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected web login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	documentsReq := httptest.NewRequest(http.MethodGet, "/app/review/documents", nil)
	applyResponseCookies(documentsReq, loginRecorder.Result().Cookies())
	documentsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(documentsRecorder, documentsReq)
	if documentsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected documents page status: got %d body=%s", documentsRecorder.Code, documentsRecorder.Body.String())
	}
	requireContains(t, documentsRecorder.Body.String(), "Document review")
	requireContains(t, documentsRecorder.Body.String(), "Posted GST invoice")
	requireContains(t, documentsRecorder.Body.String(), "/app/review/documents/"+gstInvoiceDocumentID)
	requireContains(t, documentsRecorder.Body.String(), "/app/review/approvals")
	requireContains(t, documentsRecorder.Body.String(), "/app/review/audit?entity_type=documents.document&amp;entity_id="+gstInvoiceDocumentID)
	requireContains(t, documentsRecorder.Body.String(), "/app/review/work-orders?document_id="+workOrder.DocumentID)
	requireContains(t, documentsRecorder.Body.String(), "/app/review/accounting/"+gstInvoiceJournalEntryID)
	requireContains(t, documentsRecorder.Body.String(), "/app/review/approvals/")

	exactDocumentsReq := httptest.NewRequest(http.MethodGet, "/app/review/documents/"+gstInvoiceDocumentID, nil)
	applyResponseCookies(exactDocumentsReq, loginRecorder.Result().Cookies())
	exactDocumentsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(exactDocumentsRecorder, exactDocumentsReq)
	if exactDocumentsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected document detail page status: got %d body=%s", exactDocumentsRecorder.Code, exactDocumentsRecorder.Body.String())
	}
	requireContains(t, exactDocumentsRecorder.Body.String(), "Control chain")
	requireContains(t, exactDocumentsRecorder.Body.String(), "Posted GST invoice")
	requireContains(t, exactDocumentsRecorder.Body.String(), gstInvoiceDocumentID)
	requireContains(t, exactDocumentsRecorder.Body.String(), "/app/review/documents?document_id="+gstInvoiceDocumentID)
	requireContains(t, exactDocumentsRecorder.Body.String(), "/app/review/accounting/"+gstInvoiceJournalEntryID)
	requireContains(t, exactDocumentsRecorder.Body.String(), "/app/review/approvals/")

	accountingReq := httptest.NewRequest(http.MethodGet, "/app/review/accounting", nil)
	applyResponseCookies(accountingReq, loginRecorder.Result().Cookies())
	accountingRecorder := httptest.NewRecorder()
	handler.ServeHTTP(accountingRecorder, accountingReq)
	if accountingRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected accounting page status: got %d body=%s", accountingRecorder.Code, accountingRecorder.Body.String())
	}
	requireContains(t, accountingRecorder.Body.String(), "Accounting review")
	requireContains(t, accountingRecorder.Body.String(), "Post approved invoice with GST")
	requireContains(t, accountingRecorder.Body.String(), "GST18")
	requireContains(t, accountingRecorder.Body.String(), "name=\"tax_type\"")
	requireContains(t, accountingRecorder.Body.String(), "name=\"tax_code\"")
	requireContains(t, accountingRecorder.Body.String(), "name=\"control_type\"")
	requireContains(t, accountingRecorder.Body.String(), "name=\"account_id\"")
	requireContains(t, accountingRecorder.Body.String(), "/app/review/documents/"+gstInvoiceDocumentID)
	requireContains(t, accountingRecorder.Body.String(), "/app/review/accounting/"+gstInvoiceJournalEntryID)
	requireContains(t, accountingRecorder.Body.String(), "/app/review/accounting/control-accounts/")
	requireContains(t, accountingRecorder.Body.String(), "/app/review/accounting/tax-summaries/GST18")
	requireContains(t, accountingRecorder.Body.String(), "/app/review/audit?entity_type=documents.document&amp;entity_id="+gstInvoiceDocumentID)
	requireContains(t, accountingRecorder.Body.String(), "/app/review/audit?entity_type=accounting.journal_entry&amp;entity_id="+gstInvoiceJournalEntryID)
	requireContains(t, accountingRecorder.Body.String(), "/app/review/accounting/control-accounts/")
	requireContains(t, accountingRecorder.Body.String(), "/app/review/accounting?control_type=")
	requireContains(t, accountingRecorder.Body.String(), "/app/review/accounting?tax_code=GST18&amp;tax_type=gst#tax-summaries")

	controlAccountMatch := regexp.MustCompile(`/app/review/accounting/control-accounts/([^"?&]+)">2101</a>`).FindStringSubmatch(accountingRecorder.Body.String())
	if len(controlAccountMatch) != 2 {
		t.Fatalf("expected control-account detail link in accounting page body=%s", accountingRecorder.Body.String())
	}
	gstOutputAccountID := controlAccountMatch[1]

	exactAccountingReq := httptest.NewRequest(http.MethodGet, "/app/review/accounting?document_id="+gstInvoiceDocumentID, nil)
	applyResponseCookies(exactAccountingReq, loginRecorder.Result().Cookies())
	exactAccountingRecorder := httptest.NewRecorder()
	handler.ServeHTTP(exactAccountingRecorder, exactAccountingReq)
	if exactAccountingRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact accounting page status: got %d body=%s", exactAccountingRecorder.Code, exactAccountingRecorder.Body.String())
	}
	requireContains(t, exactAccountingRecorder.Body.String(), "Post approved invoice with GST")
	requireNotContains(t, exactAccountingRecorder.Body.String(), "Issue inventory to work order")

	exactAccountingDetailReq := httptest.NewRequest(http.MethodGet, "/app/review/accounting/"+gstInvoiceJournalEntryID, nil)
	applyResponseCookies(exactAccountingDetailReq, loginRecorder.Result().Cookies())
	exactAccountingDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(exactAccountingDetailRecorder, exactAccountingDetailReq)
	if exactAccountingDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected accounting detail page status: got %d body=%s", exactAccountingDetailRecorder.Code, exactAccountingDetailRecorder.Body.String())
	}
	requireContains(t, exactAccountingDetailRecorder.Body.String(), "Journal entry #")
	requireContains(t, exactAccountingDetailRecorder.Body.String(), gstInvoiceJournalEntryID)
	requireContains(t, exactAccountingDetailRecorder.Body.String(), "/app/review/accounting?entry_id="+gstInvoiceJournalEntryID)
	requireContains(t, exactAccountingDetailRecorder.Body.String(), "/app/review/audit?entity_type=accounting.journal_entry&amp;entity_id="+gstInvoiceJournalEntryID)
	requireContains(t, exactAccountingDetailRecorder.Body.String(), "/app/review/documents/"+gstInvoiceDocumentID)

	controlAccountDetailReq := httptest.NewRequest(http.MethodGet, "/app/review/accounting/control-accounts/"+gstOutputAccountID, nil)
	applyResponseCookies(controlAccountDetailReq, loginRecorder.Result().Cookies())
	controlAccountDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(controlAccountDetailRecorder, controlAccountDetailReq)
	if controlAccountDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected control-account detail page status: got %d body=%s", controlAccountDetailRecorder.Code, controlAccountDetailRecorder.Body.String())
	}
	requireContains(t, controlAccountDetailRecorder.Body.String(), "Control account 2101")
	requireContains(t, controlAccountDetailRecorder.Body.String(), "/app/review/accounting?account_id="+gstOutputAccountID)
	requireContains(t, controlAccountDetailRecorder.Body.String(), "/app/review/accounting/tax-summaries/GST18")

	taxSummaryDetailReq := httptest.NewRequest(http.MethodGet, "/app/review/accounting/tax-summaries/GST18", nil)
	applyResponseCookies(taxSummaryDetailReq, loginRecorder.Result().Cookies())
	taxSummaryDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(taxSummaryDetailRecorder, taxSummaryDetailReq)
	if taxSummaryDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected tax-summary detail page status: got %d body=%s", taxSummaryDetailRecorder.Code, taxSummaryDetailRecorder.Body.String())
	}
	requireContains(t, taxSummaryDetailRecorder.Body.String(), "Tax summary GST18")
	requireContains(t, taxSummaryDetailRecorder.Body.String(), "/app/review/accounting?tax_code=GST18&amp;tax_type=gst#tax-summaries")
	requireContains(t, taxSummaryDetailRecorder.Body.String(), "/app/review/accounting/control-accounts/"+gstOutputAccountID)

	inboundRequestsReq := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests", nil)
	applyResponseCookies(inboundRequestsReq, loginRecorder.Result().Cookies())
	inboundRequestsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(inboundRequestsRecorder, inboundRequestsReq)
	if inboundRequestsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound requests page status: got %d body=%s", inboundRequestsRecorder.Code, inboundRequestsRecorder.Body.String())
	}
	requireContains(t, inboundRequestsRecorder.Body.String(), "Inbound-request review")
	requireContains(t, inboundRequestsRecorder.Body.String(), "Request status summary")
	requireContains(t, inboundRequestsRecorder.Body.String(), "No inbound requests")

	inventoryReq := httptest.NewRequest(http.MethodGet, "/app/review/inventory", nil)
	applyResponseCookies(inventoryReq, loginRecorder.Result().Cookies())
	inventoryRecorder := httptest.NewRecorder()
	handler.ServeHTTP(inventoryRecorder, inventoryReq)
	if inventoryRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected inventory page status: got %d body=%s", inventoryRecorder.Code, inventoryRecorder.Body.String())
	}
	requireContains(t, inventoryRecorder.Body.String(), "Inventory review")
	requireContains(t, inventoryRecorder.Body.String(), "RPT-MAT-1")
	requireContains(t, inventoryRecorder.Body.String(), "Inventory issue")
	requireContains(t, inventoryRecorder.Body.String(), "/app/review/inventory/"+issueMovementID)
	requireContains(t, inventoryRecorder.Body.String(), "/app/review/work-orders/"+workOrder.ID)
	requireContains(t, inventoryRecorder.Body.String(), "/app/review/documents/")
	requireContains(t, inventoryRecorder.Body.String(), "/app/review/audit?entity_type=inventory_ops.movement&amp;entity_id=")
	requireContains(t, inventoryRecorder.Body.String(), "/app/review/accounting/")

	exactInventoryReq := httptest.NewRequest(http.MethodGet, "/app/review/inventory?movement_id="+issueMovementID, nil)
	applyResponseCookies(exactInventoryReq, loginRecorder.Result().Cookies())
	exactInventoryRecorder := httptest.NewRecorder()
	handler.ServeHTTP(exactInventoryRecorder, exactInventoryReq)
	if exactInventoryRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact inventory page status: got %d body=%s", exactInventoryRecorder.Code, exactInventoryRecorder.Body.String())
	}
	requireContains(t, exactInventoryRecorder.Body.String(), issueMovementID)
	requireContains(t, exactInventoryRecorder.Body.String(), "Inventory issue")
	requireNotContains(t, exactInventoryRecorder.Body.String(), "Inventory receipt")

	inventoryDetailReq := httptest.NewRequest(http.MethodGet, "/app/review/inventory/"+issueMovementID, nil)
	applyResponseCookies(inventoryDetailReq, loginRecorder.Result().Cookies())
	inventoryDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(inventoryDetailRecorder, inventoryDetailReq)
	if inventoryDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected inventory detail page status: got %d body=%s", inventoryDetailRecorder.Code, inventoryDetailRecorder.Body.String())
	}
	requireContains(t, inventoryDetailRecorder.Body.String(), "Inventory movement #")
	requireContains(t, inventoryDetailRecorder.Body.String(), "Filtered inventory view")
	requireContains(t, inventoryDetailRecorder.Body.String(), "/app/review/audit?entity_type=inventory_ops.movement&amp;entity_id="+issueMovementID)
	requireContains(t, inventoryDetailRecorder.Body.String(), "/app/review/documents/")
	requireContains(t, inventoryDetailRecorder.Body.String(), "/app/review/accounting/")

	workOrdersReq := httptest.NewRequest(http.MethodGet, "/app/review/work-orders", nil)
	applyResponseCookies(workOrdersReq, loginRecorder.Result().Cookies())
	workOrdersRecorder := httptest.NewRecorder()
	handler.ServeHTTP(workOrdersRecorder, workOrdersReq)
	if workOrdersRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected work orders page status: got %d body=%s", workOrdersRecorder.Code, workOrdersRecorder.Body.String())
	}
	requireContains(t, workOrdersRecorder.Body.String(), "Work-order review")
	requireContains(t, workOrdersRecorder.Body.String(), "WO-RPT-1001")
	requireContains(t, workOrdersRecorder.Body.String(), "/app/review/documents/"+workOrder.DocumentID)

	exactWorkOrdersReq := httptest.NewRequest(http.MethodGet, "/app/review/work-orders?document_id="+workOrder.DocumentID, nil)
	applyResponseCookies(exactWorkOrdersReq, loginRecorder.Result().Cookies())
	exactWorkOrdersRecorder := httptest.NewRecorder()
	handler.ServeHTTP(exactWorkOrdersRecorder, exactWorkOrdersReq)
	if exactWorkOrdersRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact work orders page status: got %d body=%s", exactWorkOrdersRecorder.Code, exactWorkOrdersRecorder.Body.String())
	}
	requireContains(t, exactWorkOrdersRecorder.Body.String(), "WO-RPT-1001")
	requireContains(t, exactWorkOrdersRecorder.Body.String(), `name="work_order_id"`)

	exactWorkOrderIDReq := httptest.NewRequest(http.MethodGet, "/app/review/work-orders?work_order_id="+workOrder.ID, nil)
	applyResponseCookies(exactWorkOrderIDReq, loginRecorder.Result().Cookies())
	exactWorkOrderIDRecorder := httptest.NewRecorder()
	handler.ServeHTTP(exactWorkOrderIDRecorder, exactWorkOrderIDReq)
	if exactWorkOrderIDRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact work-order-id page status: got %d body=%s", exactWorkOrderIDRecorder.Code, exactWorkOrderIDRecorder.Body.String())
	}
	requireContains(t, exactWorkOrderIDRecorder.Body.String(), "WO-RPT-1001")

	workOrderDetailReq := httptest.NewRequest(http.MethodGet, "/app/review/work-orders/"+workOrder.ID, nil)
	applyResponseCookies(workOrderDetailReq, loginRecorder.Result().Cookies())
	workOrderDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(workOrderDetailRecorder, workOrderDetailReq)
	if workOrderDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected work-order detail page status: got %d body=%s", workOrderDetailRecorder.Code, workOrderDetailRecorder.Body.String())
	}
	requireContains(t, workOrderDetailRecorder.Body.String(), "Work order WO-RPT-1001")
	requireContains(t, workOrderDetailRecorder.Body.String(), "Review execution chain")
	requireContains(t, workOrderDetailRecorder.Body.String(), "/app/review/work-orders?work_order_id="+workOrder.ID)
	requireContains(t, workOrderDetailRecorder.Body.String(), "/app/review/documents/")
	requireContains(t, workOrderDetailRecorder.Body.String(), "/app/review/audit?entity_type=work_orders.work_order&amp;entity_id="+workOrder.ID)
	requireContains(t, workOrderDetailRecorder.Body.String(), "/app/review/accounting?document_id="+workOrder.DocumentID)

	auditReq := httptest.NewRequest(http.MethodGet, "/app/review/audit?entity_type=work_orders.work_order&entity_id="+workOrder.ID, nil)
	applyResponseCookies(auditReq, loginRecorder.Result().Cookies())
	auditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(auditRecorder, auditReq)
	if auditRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected audit page status: got %d body=%s", auditRecorder.Code, auditRecorder.Body.String())
	}
	requireContains(t, auditRecorder.Body.String(), "Audit lookup")
	requireContains(t, auditRecorder.Body.String(), "work_orders.work_order_created")
	requireContains(t, auditRecorder.Body.String(), "/app/review/work-orders/"+workOrder.ID)
	requireContains(t, auditRecorder.Body.String(), "/app/review/audit/")

	movementAuditReq := httptest.NewRequest(http.MethodGet, "/app/review/audit?entity_type=inventory_ops.movement&entity_id="+issueMovementID, nil)
	applyResponseCookies(movementAuditReq, loginRecorder.Result().Cookies())
	movementAuditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(movementAuditRecorder, movementAuditReq)
	if movementAuditRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected movement audit page status: got %d body=%s", movementAuditRecorder.Code, movementAuditRecorder.Body.String())
	}
	requireContains(t, movementAuditRecorder.Body.String(), "/app/review/inventory/"+issueMovementID)

	journalAuditReq := httptest.NewRequest(http.MethodGet, "/app/review/audit?entity_type=accounting.journal_entry&entity_id="+gstInvoiceJournalEntryID, nil)
	applyResponseCookies(journalAuditReq, loginRecorder.Result().Cookies())
	journalAuditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(journalAuditRecorder, journalAuditReq)
	if journalAuditRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected journal audit page status: got %d body=%s", journalAuditRecorder.Code, journalAuditRecorder.Body.String())
	}
	requireContains(t, journalAuditRecorder.Body.String(), "/app/review/accounting/"+gstInvoiceJournalEntryID)

	listAuditAPIReq := httptest.NewRequest(http.MethodGet, "/api/review/audit-events?entity_type=work_orders.work_order&entity_id="+workOrder.ID, nil)
	applyResponseCookies(listAuditAPIReq, loginRecorder.Result().Cookies())
	listAuditAPIRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listAuditAPIRecorder, listAuditAPIReq)
	if listAuditAPIRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected audit api status: got %d body=%s", listAuditAPIRecorder.Code, listAuditAPIRecorder.Body.String())
	}
	var auditListResponse struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listAuditAPIRecorder.Body.Bytes(), &auditListResponse); err != nil {
		t.Fatalf("unmarshal audit api response: %v", err)
	}
	if len(auditListResponse.Items) == 0 {
		t.Fatal("expected audit api items")
	}

	exactAuditReq := httptest.NewRequest(http.MethodGet, "/app/review/audit/"+auditListResponse.Items[0].ID, nil)
	applyResponseCookies(exactAuditReq, loginRecorder.Result().Cookies())
	exactAuditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(exactAuditRecorder, exactAuditReq)
	if exactAuditRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact audit page status: got %d body=%s", exactAuditRecorder.Code, exactAuditRecorder.Body.String())
	}
	requireContains(t, exactAuditRecorder.Body.String(), "Audit event "+auditListResponse.Items[0].ID)
	requireContains(t, exactAuditRecorder.Body.String(), "Filtered audit view")
	requireContains(t, exactAuditRecorder.Body.String(), "/app/review/work-orders/"+workOrder.ID)

	apiDocumentsReq := httptest.NewRequest(http.MethodGet, "/api/review/documents", nil)
	applyResponseCookies(apiDocumentsReq, loginRecorder.Result().Cookies())
	apiDocumentsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiDocumentsRecorder, apiDocumentsReq)
	if apiDocumentsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected documents api status: got %d body=%s", apiDocumentsRecorder.Code, apiDocumentsRecorder.Body.String())
	}
	requireContains(t, apiDocumentsRecorder.Body.String(), "Posted GST invoice")

	apiExactDocumentsReq := httptest.NewRequest(http.MethodGet, "/api/review/documents?document_id="+gstInvoiceDocumentID, nil)
	applyResponseCookies(apiExactDocumentsReq, loginRecorder.Result().Cookies())
	apiExactDocumentsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactDocumentsRecorder, apiExactDocumentsReq)
	if apiExactDocumentsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact documents api status: got %d body=%s", apiExactDocumentsRecorder.Code, apiExactDocumentsRecorder.Body.String())
	}
	requireContains(t, apiExactDocumentsRecorder.Body.String(), "\"document_id\":\""+gstInvoiceDocumentID+"\"")

	apiJournalReq := httptest.NewRequest(http.MethodGet, "/api/review/accounting/journal-entries", nil)
	applyResponseCookies(apiJournalReq, loginRecorder.Result().Cookies())
	apiJournalRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiJournalRecorder, apiJournalReq)
	if apiJournalRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected journal api status: got %d body=%s", apiJournalRecorder.Code, apiJournalRecorder.Body.String())
	}
	requireContains(t, apiJournalRecorder.Body.String(), "Post approved invoice with GST")
	requireContains(t, apiJournalRecorder.Body.String(), "\"source_document_id\":\""+gstInvoiceDocumentID+"\"")

	apiExactJournalReq := httptest.NewRequest(http.MethodGet, "/api/review/accounting/journal-entries?document_id="+gstInvoiceDocumentID, nil)
	applyResponseCookies(apiExactJournalReq, loginRecorder.Result().Cookies())
	apiExactJournalRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactJournalRecorder, apiExactJournalReq)
	if apiExactJournalRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact journal api status: got %d body=%s", apiExactJournalRecorder.Code, apiExactJournalRecorder.Body.String())
	}
	requireContains(t, apiExactJournalRecorder.Body.String(), "\"source_document_id\":\""+gstInvoiceDocumentID+"\"")
	requireNotContains(t, apiExactJournalRecorder.Body.String(), "Issue inventory to work order")

	apiExactJournalEntryReq := httptest.NewRequest(http.MethodGet, "/api/review/accounting/journal-entries?entry_id="+gstInvoiceJournalEntryID, nil)
	applyResponseCookies(apiExactJournalEntryReq, loginRecorder.Result().Cookies())
	apiExactJournalEntryRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactJournalEntryRecorder, apiExactJournalEntryReq)
	if apiExactJournalEntryRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact journal-entry api status: got %d body=%s", apiExactJournalEntryRecorder.Code, apiExactJournalEntryRecorder.Body.String())
	}
	requireContains(t, apiExactJournalEntryRecorder.Body.String(), "\"entry_id\":\""+gstInvoiceJournalEntryID+"\"")
	requireNotContains(t, apiExactJournalEntryRecorder.Body.String(), "Issue inventory to work order")

	apiBalanceReq := httptest.NewRequest(http.MethodGet, "/api/review/accounting/control-account-balances", nil)
	applyResponseCookies(apiBalanceReq, loginRecorder.Result().Cookies())
	apiBalanceRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiBalanceRecorder, apiBalanceReq)
	if apiBalanceRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected control balance api status: got %d body=%s", apiBalanceRecorder.Code, apiBalanceRecorder.Body.String())
	}
	requireContains(t, apiBalanceRecorder.Body.String(), "\"account_code\":\"2101\"")

	apiExactBalanceReq := httptest.NewRequest(http.MethodGet, "/api/review/accounting/control-account-balances?control_type=gst_output", nil)
	applyResponseCookies(apiExactBalanceReq, loginRecorder.Result().Cookies())
	apiExactBalanceRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactBalanceRecorder, apiExactBalanceReq)
	if apiExactBalanceRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected filtered control balance api status: got %d body=%s", apiExactBalanceRecorder.Code, apiExactBalanceRecorder.Body.String())
	}
	requireContains(t, apiExactBalanceRecorder.Body.String(), "\"control_type\":\"gst_output\"")
	requireNotContains(t, apiExactBalanceRecorder.Body.String(), "\"control_type\":\"tds_payable\"")

	apiTaxReq := httptest.NewRequest(http.MethodGet, "/api/review/accounting/tax-summaries", nil)
	applyResponseCookies(apiTaxReq, loginRecorder.Result().Cookies())
	apiTaxRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiTaxRecorder, apiTaxReq)
	if apiTaxRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected tax summary api status: got %d body=%s", apiTaxRecorder.Code, apiTaxRecorder.Body.String())
	}
	requireContains(t, apiTaxRecorder.Body.String(), "\"tax_code\":\"GST18\"")

	apiExactTaxReq := httptest.NewRequest(http.MethodGet, "/api/review/accounting/tax-summaries?tax_type=gst", nil)
	applyResponseCookies(apiExactTaxReq, loginRecorder.Result().Cookies())
	apiExactTaxRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactTaxRecorder, apiExactTaxReq)
	if apiExactTaxRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected filtered tax summary api status: got %d body=%s", apiExactTaxRecorder.Code, apiExactTaxRecorder.Body.String())
	}
	requireContains(t, apiExactTaxRecorder.Body.String(), "\"tax_type\":\"gst\"")
	requireNotContains(t, apiExactTaxRecorder.Body.String(), "\"tax_type\":\"tds\"")

	apiExactTaxCodeReq := httptest.NewRequest(http.MethodGet, "/api/review/accounting/tax-summaries?tax_code=GST18", nil)
	applyResponseCookies(apiExactTaxCodeReq, loginRecorder.Result().Cookies())
	apiExactTaxCodeRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactTaxCodeRecorder, apiExactTaxCodeReq)
	if apiExactTaxCodeRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact tax-code api status: got %d body=%s", apiExactTaxCodeRecorder.Code, apiExactTaxCodeRecorder.Body.String())
	}
	requireContains(t, apiExactTaxCodeRecorder.Body.String(), "\"tax_code\":\"GST18\"")
	requireNotContains(t, apiExactTaxCodeRecorder.Body.String(), "\"tax_code\":\"TDS1\"")

	apiInventoryStockReq := httptest.NewRequest(http.MethodGet, "/api/review/inventory/stock", nil)
	applyResponseCookies(apiInventoryStockReq, loginRecorder.Result().Cookies())
	apiInventoryStockRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiInventoryStockRecorder, apiInventoryStockReq)
	if apiInventoryStockRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected inventory stock api status: got %d body=%s", apiInventoryStockRecorder.Code, apiInventoryStockRecorder.Body.String())
	}
	requireContains(t, apiInventoryStockRecorder.Body.String(), "\"item_sku\":\"RPT-MAT-1\"")

	apiInventoryMovesReq := httptest.NewRequest(http.MethodGet, "/api/review/inventory/movements", nil)
	applyResponseCookies(apiInventoryMovesReq, loginRecorder.Result().Cookies())
	apiInventoryMovesRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiInventoryMovesRecorder, apiInventoryMovesReq)
	if apiInventoryMovesRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected inventory movement api status: got %d body=%s", apiInventoryMovesRecorder.Code, apiInventoryMovesRecorder.Body.String())
	}
	requireContains(t, apiInventoryMovesRecorder.Body.String(), "\"movement_type\":\"issue\"")

	apiExactInventoryMovesReq := httptest.NewRequest(http.MethodGet, "/api/review/inventory/movements?movement_id="+issueMovementID, nil)
	applyResponseCookies(apiExactInventoryMovesReq, loginRecorder.Result().Cookies())
	apiExactInventoryMovesRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactInventoryMovesRecorder, apiExactInventoryMovesReq)
	if apiExactInventoryMovesRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact inventory movement api status: got %d body=%s", apiExactInventoryMovesRecorder.Code, apiExactInventoryMovesRecorder.Body.String())
	}
	requireContains(t, apiExactInventoryMovesRecorder.Body.String(), "\"movement_id\":\""+issueMovementID+"\"")
	requireContains(t, apiExactInventoryMovesRecorder.Body.String(), "\"movement_type\":\"issue\"")
	requireNotContains(t, apiExactInventoryMovesRecorder.Body.String(), "\"movement_type\":\"receipt\"")

	apiInventoryReconReq := httptest.NewRequest(http.MethodGet, "/api/review/inventory/reconciliation", nil)
	applyResponseCookies(apiInventoryReconReq, loginRecorder.Result().Cookies())
	apiInventoryReconRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiInventoryReconRecorder, apiInventoryReconReq)
	if apiInventoryReconRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected inventory reconciliation api status: got %d body=%s", apiInventoryReconRecorder.Code, apiInventoryReconRecorder.Body.String())
	}
	requireContains(t, apiInventoryReconRecorder.Body.String(), "\"work_order_code\":\"WO-RPT-1001\"")

	apiWorkOrdersReq := httptest.NewRequest(http.MethodGet, "/api/review/work-orders", nil)
	applyResponseCookies(apiWorkOrdersReq, loginRecorder.Result().Cookies())
	apiWorkOrdersRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiWorkOrdersRecorder, apiWorkOrdersReq)
	if apiWorkOrdersRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected work orders api status: got %d body=%s", apiWorkOrdersRecorder.Code, apiWorkOrdersRecorder.Body.String())
	}
	requireContains(t, apiWorkOrdersRecorder.Body.String(), "\"work_order_code\":\"WO-RPT-1001\"")

	apiExactWorkOrdersReq := httptest.NewRequest(http.MethodGet, "/api/review/work-orders?document_id="+workOrder.DocumentID, nil)
	applyResponseCookies(apiExactWorkOrdersReq, loginRecorder.Result().Cookies())
	apiExactWorkOrdersRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactWorkOrdersRecorder, apiExactWorkOrdersReq)
	if apiExactWorkOrdersRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact work orders api status: got %d body=%s", apiExactWorkOrdersRecorder.Code, apiExactWorkOrdersRecorder.Body.String())
	}
	requireContains(t, apiExactWorkOrdersRecorder.Body.String(), "\"document_id\":\""+workOrder.DocumentID+"\"")

	apiExactWorkOrderIDReq := httptest.NewRequest(http.MethodGet, "/api/review/work-orders?work_order_id="+workOrder.ID, nil)
	applyResponseCookies(apiExactWorkOrderIDReq, loginRecorder.Result().Cookies())
	apiExactWorkOrderIDRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactWorkOrderIDRecorder, apiExactWorkOrderIDReq)
	if apiExactWorkOrderIDRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact work-order-id api status: got %d body=%s", apiExactWorkOrderIDRecorder.Code, apiExactWorkOrderIDRecorder.Body.String())
	}
	requireContains(t, apiExactWorkOrderIDRecorder.Body.String(), "\"work_order_id\":\""+workOrder.ID+"\"")

	apiWorkOrderDetailReq := httptest.NewRequest(http.MethodGet, "/api/review/work-orders/"+workOrder.ID, nil)
	applyResponseCookies(apiWorkOrderDetailReq, loginRecorder.Result().Cookies())
	apiWorkOrderDetailRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiWorkOrderDetailRecorder, apiWorkOrderDetailReq)
	if apiWorkOrderDetailRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected work order detail api status: got %d body=%s", apiWorkOrderDetailRecorder.Code, apiWorkOrderDetailRecorder.Body.String())
	}
	requireContains(t, apiWorkOrderDetailRecorder.Body.String(), "\"work_order_id\":\""+workOrder.ID+"\"")

	apiAuditReq := httptest.NewRequest(http.MethodGet, "/api/review/audit-events?entity_type=work_orders.work_order&entity_id="+workOrder.ID, nil)
	applyResponseCookies(apiAuditReq, loginRecorder.Result().Cookies())
	apiAuditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiAuditRecorder, apiAuditReq)
	if apiAuditRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected audit api status: got %d body=%s", apiAuditRecorder.Code, apiAuditRecorder.Body.String())
	}
	requireContains(t, apiAuditRecorder.Body.String(), "\"event_type\":\"work_orders.work_order_created\"")

	var apiAuditListResponse struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(apiAuditRecorder.Body.Bytes(), &apiAuditListResponse); err != nil {
		t.Fatalf("unmarshal exact audit seed response: %v", err)
	}
	if len(apiAuditListResponse.Items) == 0 {
		t.Fatal("expected audit api items for exact filter")
	}

	apiExactAuditReq := httptest.NewRequest(http.MethodGet, "/api/review/audit-events?event_id="+apiAuditListResponse.Items[0].ID, nil)
	applyResponseCookies(apiExactAuditReq, loginRecorder.Result().Cookies())
	apiExactAuditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(apiExactAuditRecorder, apiExactAuditReq)
	if apiExactAuditRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact audit api status: got %d body=%s", apiExactAuditRecorder.Code, apiExactAuditRecorder.Body.String())
	}
	requireContains(t, apiExactAuditRecorder.Body.String(), "\"id\":\""+apiAuditListResponse.Items[0].ID+"\"")
}

func TestAgentAPISessionLoginRejectsUnknownMembership(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	handler := app.NewAgentAPIHandler(db)

	req := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewBufferString(`{
		"org_slug":"missing-org",
		"email":"missing@example.com"
	}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected login failure status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAgentAPIProcessNextQueuedInboundRequestIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	request := createQueuedRequest(t, ctx, db, operator, "Urgent pump issue reported from the warehouse.")
	processor, err := app.NewAgentProcessor(db, fakeCoordinatorProvider{
		output: ai.CoordinatorProviderOutput{
			ProviderName:       "openai",
			ProviderResponseID: "resp_api_test_123",
			Model:              "gpt-5.2",
			Summary:            "Operator review is required for the urgent pump issue.",
			Priority:           "urgent",
			ArtifactTitle:      "Inbound request review brief",
			ArtifactBody:       "The request describes an urgent equipment problem that should be reviewed immediately.",
			Rationale: []string{
				"The request describes a time-sensitive equipment failure.",
			},
			NextActions: []string{
				"Confirm the site details and route controlled follow-up.",
			},
		},
	})
	if err != nil {
		t.Fatalf("new agent processor: %v", err)
	}

	handler := app.NewAgentAPIHandlerWithProcessorLoader(func() (app.ProcessNextQueuedInboundRequester, error) {
		return processor, nil
	})

	body := bytes.NewBufferString(`{"channel":"browser"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/process-next-queued-inbound-request", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Processed             bool   `json:"processed"`
		RequestReference      string `json:"request_reference"`
		RequestStatus         string `json:"request_status"`
		RunID                 string `json:"run_id"`
		RunStatus             string `json:"run_status"`
		ArtifactID            string `json:"artifact_id"`
		RecommendationID      string `json:"recommendation_id"`
		RecommendationSummary string `json:"recommendation_summary"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !response.Processed {
		t.Fatal("expected processed response")
	}
	if response.RequestReference != request.RequestReference {
		t.Fatalf("unexpected request reference: got %s want %s", response.RequestReference, request.RequestReference)
	}
	if response.RequestStatus != "processed" {
		t.Fatalf("unexpected request status: %s", response.RequestStatus)
	}
	if response.RunID == "" || response.ArtifactID == "" || response.RecommendationID == "" {
		t.Fatalf("expected run, artifact, and recommendation identifiers in response: %+v", response)
	}
	if response.RecommendationSummary == "" {
		t.Fatal("expected recommendation summary")
	}
}

func TestAgentAPIProcessNextQueuedInboundRequestReturnsNotProcessedWhenQueueEmpty(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	processor, err := app.NewAgentProcessor(db, fakeCoordinatorProvider{})
	if err != nil {
		t.Fatalf("new agent processor: %v", err)
	}

	handler := app.NewAgentAPIHandlerWithProcessorLoader(func() (app.ProcessNextQueuedInboundRequester, error) {
		return processor, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/agent/process-next-queued-inbound-request", bytes.NewBufferString(`{"channel":"browser"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Processed bool `json:"processed"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Processed {
		t.Fatal("expected queue-empty response")
	}
}

func TestAgentAPIProcessNextQueuedInboundRequestUnauthorized(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}
	createQueuedRequest(t, ctx, db, operator, "Urgent pump issue reported from the warehouse.")

	processor, err := app.NewAgentProcessor(db, fakeCoordinatorProvider{})
	if err != nil {
		t.Fatalf("new agent processor: %v", err)
	}

	handler := app.NewAgentAPIHandlerWithProcessorLoader(func() (app.ProcessNextQueuedInboundRequester, error) {
		return processor, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/agent/process-next-queued-inbound-request", bytes.NewBufferString(`{"channel":"browser"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", "00000000-0000-4000-8000-000000000001")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAgentAPIProcessNextQueuedInboundRequestReturnsProviderConfigurationError(t *testing.T) {
	handler := app.NewAgentAPIHandlerWithProcessorLoader(func() (app.ProcessNextQueuedInboundRequester, error) {
		return nil, app.ErrAgentProviderNotConfigured
	})

	req := httptest.NewRequest(http.MethodPost, "/api/agent/process-next-queued-inbound-request", bytes.NewBufferString(`{"channel":"browser"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", "00000000-0000-4000-8000-000000000001")
	req.Header.Set("X-Workflow-User-ID", "00000000-0000-4000-8000-000000000002")
	req.Header.Set("X-Workflow-Session-ID", "00000000-0000-4000-8000-000000000003")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAgentAPIProcessNextQueuedInboundRequestRequiresHeaders(t *testing.T) {
	handler := app.NewAgentAPIHandlerWithProcessorLoader(func() (app.ProcessNextQueuedInboundRequester, error) {
		return nil, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/agent/process-next-queued-inbound-request", bytes.NewBufferString(`{"channel":"browser"}`))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAgentAPIProcessNextQueuedInboundRequestRejectsMalformedHeaders(t *testing.T) {
	handler := app.NewAgentAPIHandlerWithProcessorLoader(func() (app.ProcessNextQueuedInboundRequester, error) {
		return nil, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/agent/process-next-queued-inbound-request", bytes.NewBufferString(`{"channel":"browser"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", "not-a-uuid")
	req.Header.Set("X-Workflow-User-ID", "also-not-a-uuid")
	req.Header.Set("X-Workflow-Session-ID", "still-not-a-uuid")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAgentAPISubmitInboundRequestIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandlerWithServices(nil, app.NewSubmissionService(db))

	body := bytes.NewBufferString(`{
		"origin_type":"human",
		"channel":"browser",
		"metadata":{"submitter_label":"front desk"},
		"message":{"message_role":"request","text_content":"The warehouse pump has failed and needs review."},
		"attachments":[
			{
				"original_file_name":"pump-note.txt",
				"media_type":"text/plain",
				"content_base64":"` + base64.StdEncoding.EncodeToString([]byte("urgent pump failure details")) + `",
				"link_role":"evidence"
			}
		]
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		RequestID        string     `json:"request_id"`
		RequestReference string     `json:"request_reference"`
		Status           string     `json:"status"`
		MessageID        string     `json:"message_id"`
		AttachmentIDs    []string   `json:"attachment_ids"`
		ReceivedAt       time.Time  `json:"received_at"`
		QueuedAt         *time.Time `json:"queued_at"`
		UpdatedAt        time.Time  `json:"updated_at"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.RequestID == "" || response.RequestReference == "" || response.MessageID == "" {
		t.Fatalf("expected request identifiers in response: %+v", response)
	}
	if response.Status != "queued" {
		t.Fatalf("unexpected request status: %s", response.Status)
	}
	if len(response.AttachmentIDs) != 1 {
		t.Fatalf("unexpected attachment ids: %+v", response.AttachmentIDs)
	}
	if response.ReceivedAt.IsZero() || response.UpdatedAt.IsZero() || response.QueuedAt == nil || response.QueuedAt.IsZero() {
		t.Fatalf("expected lifecycle timestamps in submit response: %+v", response)
	}

	var requestStatus string
	if err := db.QueryRowContext(ctx, `SELECT status FROM ai.inbound_requests WHERE id = $1`, response.RequestID).Scan(&requestStatus); err != nil {
		t.Fatalf("load queued request: %v", err)
	}
	if requestStatus != "queued" {
		t.Fatalf("unexpected persisted request status: %s", requestStatus)
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/api/attachments/"+response.AttachmentIDs[0]+"/content", nil)
	downloadReq.Header.Set("X-Workflow-Org-ID", orgID)
	downloadReq.Header.Set("X-Workflow-User-ID", operatorUserID)
	downloadReq.Header.Set("X-Workflow-Session-ID", session.ID)

	downloadRecorder := httptest.NewRecorder()
	handler.ServeHTTP(downloadRecorder, downloadReq)

	if downloadRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected download status: got %d body=%s", downloadRecorder.Code, downloadRecorder.Body.String())
	}
	if got := downloadRecorder.Header().Get("Content-Type"); got != "text/plain" {
		t.Fatalf("unexpected content type: %s", got)
	}
	if got := downloadRecorder.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("unexpected cache control: %s", got)
	}
	if got := downloadRecorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("unexpected nosniff header: %s", got)
	}
	if got := downloadRecorder.Header().Get("Content-Disposition"); !strings.Contains(got, `attachment; filename="pump-note.txt"`) {
		t.Fatalf("unexpected content disposition: %s", got)
	}
	if got := downloadRecorder.Body.String(); got != "urgent pump failure details" {
		t.Fatalf("unexpected attachment payload: %q", got)
	}
}

func TestAgentAPISubmitInboundRequestRejectsInvalidAttachmentMediaType(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandlerWithServices(nil, app.NewSubmissionService(db))

	req := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", bytes.NewBufferString(`{
		"channel":"browser",
		"message":{"text_content":"Attachment upload should fail."},
		"attachments":[{"original_file_name":"broken.txt","media_type":"not a media type","content_base64":"`+base64.StdEncoding.EncodeToString([]byte("broken"))+`"}]
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var requestCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.inbound_requests WHERE org_id = $1`, orgID).Scan(&requestCount); err != nil {
		t.Fatalf("count inbound requests: %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected failed submission cleanup, found %d requests", requestCount)
	}
}

func TestAgentAPIDownloadAttachmentRejectsMalformedAttachmentID(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandlerWithServices(nil, app.NewSubmissionService(db))

	req := httptest.NewRequest(http.MethodGet, "/api/attachments/not-a-uuid/content", nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	requireContains(t, recorder.Body.String(), `"error":"invalid attachment"`)
}

func TestAgentAPISubmitInboundRequestRejectsInvalidAttachmentContent(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandlerWithServices(nil, app.NewSubmissionService(db))

	req := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", bytes.NewBufferString(`{
		"channel":"browser",
		"message":{"text_content":"Attachment upload should fail."},
		"attachments":[{"original_file_name":"broken.txt","media_type":"text/plain","content_base64":"not-base64%%%"}]
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var requestCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.inbound_requests WHERE org_id = $1`, orgID).Scan(&requestCount); err != nil {
		t.Fatalf("count inbound requests: %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected failed submission cleanup, found %d requests", requestCount)
	}
}

func TestAgentAPIDraftLifecycleIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandlerWithServices(nil, app.NewSubmissionService(db))

	saveDraftReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", bytes.NewBufferString(`{
		"origin_type":"human",
		"channel":"browser",
		"metadata":{"submitter_label":"front desk"},
		"message":{"message_role":"request","text_content":"Draft request from API."},
		"queue_for_review":false
	}`))
	saveDraftReq.Header.Set("Content-Type", "application/json")
	saveDraftReq.Header.Set("X-Workflow-Org-ID", orgID)
	saveDraftReq.Header.Set("X-Workflow-User-ID", operatorUserID)
	saveDraftReq.Header.Set("X-Workflow-Session-ID", session.ID)

	saveDraftRecorder := httptest.NewRecorder()
	handler.ServeHTTP(saveDraftRecorder, saveDraftReq)
	if saveDraftRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected save-draft status: got %d body=%s", saveDraftRecorder.Code, saveDraftRecorder.Body.String())
	}

	var draftResponse struct {
		RequestID        string    `json:"request_id"`
		RequestReference string    `json:"request_reference"`
		MessageID        string    `json:"message_id"`
		Status           string    `json:"status"`
		ReceivedAt       time.Time `json:"received_at"`
		CreatedAt        time.Time `json:"created_at"`
		UpdatedAt        time.Time `json:"updated_at"`
	}
	if err := json.Unmarshal(saveDraftRecorder.Body.Bytes(), &draftResponse); err != nil {
		t.Fatalf("decode save-draft response: %v", err)
	}
	if draftResponse.Status != intake.StatusDraft {
		t.Fatalf("unexpected draft status: %+v", draftResponse)
	}
	if draftResponse.RequestReference == "" || draftResponse.ReceivedAt.IsZero() || draftResponse.CreatedAt.IsZero() || draftResponse.UpdatedAt.IsZero() {
		t.Fatalf("expected draft lifecycle metadata: %+v", draftResponse)
	}

	updateReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draftResponse.RequestID+"/draft", bytes.NewBufferString(`{
		"message_id":"`+draftResponse.MessageID+`",
		"origin_type":"human",
		"channel":"browser",
		"message":{"message_role":"request","text_content":"Updated API draft."}
	}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("X-Workflow-Org-ID", orgID)
	updateReq.Header.Set("X-Workflow-User-ID", operatorUserID)
	updateReq.Header.Set("X-Workflow-Session-ID", session.ID)
	updateRecorder := httptest.NewRecorder()
	handler.ServeHTTP(updateRecorder, updateReq)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected draft-update status: got %d body=%s", updateRecorder.Code, updateRecorder.Body.String())
	}
	var updateResponse struct {
		RequestID        string    `json:"request_id"`
		RequestReference string    `json:"request_reference"`
		Status           string    `json:"status"`
		MessageID        string    `json:"message_id"`
		ReceivedAt       time.Time `json:"received_at"`
		UpdatedAt        time.Time `json:"updated_at"`
	}
	if err := json.Unmarshal(updateRecorder.Body.Bytes(), &updateResponse); err != nil {
		t.Fatalf("decode draft-update response: %v", err)
	}
	if updateResponse.RequestID != draftResponse.RequestID || updateResponse.RequestReference != draftResponse.RequestReference || updateResponse.Status != intake.StatusDraft {
		t.Fatalf("unexpected draft-update response: %+v", updateResponse)
	}
	if updateResponse.MessageID != draftResponse.MessageID || updateResponse.ReceivedAt.IsZero() || updateResponse.UpdatedAt.IsZero() {
		t.Fatalf("expected draft-update lifecycle metadata: %+v", updateResponse)
	}

	queueReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draftResponse.RequestID+"/queue", nil)
	queueReq.Header.Set("X-Workflow-Org-ID", orgID)
	queueReq.Header.Set("X-Workflow-User-ID", operatorUserID)
	queueReq.Header.Set("X-Workflow-Session-ID", session.ID)
	queueRecorder := httptest.NewRecorder()
	handler.ServeHTTP(queueRecorder, queueReq)
	if queueRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected queue status: got %d body=%s", queueRecorder.Code, queueRecorder.Body.String())
	}
	var queueResponse struct {
		RequestID        string     `json:"request_id"`
		RequestReference string     `json:"request_reference"`
		Status           string     `json:"status"`
		QueuedAt         *time.Time `json:"queued_at"`
	}
	if err := json.Unmarshal(queueRecorder.Body.Bytes(), &queueResponse); err != nil {
		t.Fatalf("decode queue response: %v", err)
	}
	if queueResponse.Status != intake.StatusQueued || queueResponse.RequestReference != draftResponse.RequestReference || queueResponse.QueuedAt == nil || queueResponse.QueuedAt.IsZero() {
		t.Fatalf("unexpected queue response: %+v", queueResponse)
	}

	cancelReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draftResponse.RequestID+"/cancel", bytes.NewBufferString(`{"reason":"operator paused request"}`))
	cancelReq.Header.Set("Content-Type", "application/json")
	cancelReq.Header.Set("X-Workflow-Org-ID", orgID)
	cancelReq.Header.Set("X-Workflow-User-ID", operatorUserID)
	cancelReq.Header.Set("X-Workflow-Session-ID", session.ID)
	cancelRecorder := httptest.NewRecorder()
	handler.ServeHTTP(cancelRecorder, cancelReq)
	if cancelRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected cancel status: got %d body=%s", cancelRecorder.Code, cancelRecorder.Body.String())
	}
	var cancelResponse struct {
		RequestID          string     `json:"request_id"`
		RequestReference   string     `json:"request_reference"`
		Status             string     `json:"status"`
		CancellationReason string     `json:"cancellation_reason"`
		CancelledAt        *time.Time `json:"cancelled_at"`
	}
	if err := json.Unmarshal(cancelRecorder.Body.Bytes(), &cancelResponse); err != nil {
		t.Fatalf("decode cancel response: %v", err)
	}
	if cancelResponse.Status != intake.StatusCancelled || cancelResponse.RequestReference != draftResponse.RequestReference || cancelResponse.CancellationReason != "operator paused request" || cancelResponse.CancelledAt == nil || cancelResponse.CancelledAt.IsZero() {
		t.Fatalf("unexpected cancel response: %+v", cancelResponse)
	}

	amendReq := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draftResponse.RequestID+"/amend", nil)
	amendReq.Header.Set("X-Workflow-Org-ID", orgID)
	amendReq.Header.Set("X-Workflow-User-ID", operatorUserID)
	amendReq.Header.Set("X-Workflow-Session-ID", session.ID)
	amendRecorder := httptest.NewRecorder()
	handler.ServeHTTP(amendRecorder, amendReq)
	if amendRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected amend status: got %d body=%s", amendRecorder.Code, amendRecorder.Body.String())
	}
	var amendResponse struct {
		RequestID        string     `json:"request_id"`
		RequestReference string     `json:"request_reference"`
		Status           string     `json:"status"`
		QueuedAt         *time.Time `json:"queued_at"`
		CancelledAt      *time.Time `json:"cancelled_at"`
	}
	if err := json.Unmarshal(amendRecorder.Body.Bytes(), &amendResponse); err != nil {
		t.Fatalf("decode amend response: %v", err)
	}
	if amendResponse.Status != intake.StatusDraft || amendResponse.RequestReference != draftResponse.RequestReference || amendResponse.QueuedAt != nil || amendResponse.CancelledAt != nil {
		t.Fatalf("unexpected amend response: %+v", amendResponse)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/inbound-requests/"+draftResponse.RequestID+"/delete", nil)
	deleteReq.Header.Set("X-Workflow-Org-ID", orgID)
	deleteReq.Header.Set("X-Workflow-User-ID", operatorUserID)
	deleteReq.Header.Set("X-Workflow-Session-ID", session.ID)
	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteReq)
	if deleteRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected delete status: got %d body=%s", deleteRecorder.Code, deleteRecorder.Body.String())
	}

	var remaining int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.inbound_requests WHERE id = $1`, draftResponse.RequestID).Scan(&remaining); err != nil {
		t.Fatalf("count remaining requests: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected deleted draft to be removed, found %d", remaining)
	}
}

func TestAgentAPIInboundRequestCancelRejectsInvalidJSONIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandlerWithServices(nil, app.NewSubmissionService(db))

	service := app.NewSubmissionService(db)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}
	draft, err := service.SaveInboundDraft(ctx, app.SaveInboundDraftInput{
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		MessageRole: intake.MessageRoleRequest,
		MessageText: "Draft request that will be queued.",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("save draft: %v", err)
	}
	if _, err := service.QueueInboundRequest(ctx, app.QueueInboundRequestInput{
		RequestID: draft.Request.ID,
		Actor:     operator,
	}); err != nil {
		t.Fatalf("queue draft: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/inbound-requests/"+draft.Request.ID+"/cancel", bytes.NewBufferString(`{"reason":"bad"`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected invalid-cancel status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM ai.inbound_requests WHERE id = $1`, draft.Request.ID).Scan(&status); err != nil {
		t.Fatalf("load request status: %v", err)
	}
	if status != intake.StatusQueued {
		t.Fatalf("expected invalid cancel payload to preserve queued status, got %s", status)
	}
}

func TestAgentAPISubmitInboundRequestRejectsUnknownJSONFieldsIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandlerWithServices(nil, app.NewSubmissionService(db))

	req := httptest.NewRequest(http.MethodPost, "/api/inbound-requests", bytes.NewBufferString(`{
		"channel":"browser",
		"message":{"text_content":"Unknown field should fail."},
		"unexpected":"value"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", session.ID)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected unknown-field status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var requestCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.inbound_requests WHERE org_id = $1`, orgID).Scan(&requestCount); err != nil {
		t.Fatalf("count inbound requests: %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected unknown-field request rejection before persistence, found %d requests", requestCount)
	}
}

func TestAgentAPIReviewSurfacesIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	orgSlug, userEmail := loadOrgSlugAndUserEmail(t, ctx, db, orgID, operatorUserID)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	request := createQueuedRequest(t, ctx, db, operator, "Urgent pump issue reported from the warehouse.")

	processor, err := app.NewAgentProcessor(db, fakeCoordinatorProvider{
		output: ai.CoordinatorProviderOutput{
			ProviderName:       "openai",
			ProviderResponseID: "resp_review_api_test_123",
			Model:              "gpt-5.2",
			Summary:            "Operator review is required for the urgent pump issue.",
			Priority:           "urgent",
			ArtifactTitle:      "Inbound request review brief",
			ArtifactBody:       "The request describes an urgent equipment problem that should be reviewed immediately.",
			Rationale: []string{
				"The request describes a time-sensitive equipment failure.",
			},
			NextActions: []string{
				"Confirm the site details and route controlled follow-up.",
			},
		},
	})
	if err != nil {
		t.Fatalf("new agent processor: %v", err)
	}

	processResult, err := processor.ProcessNextQueuedInboundRequest(ctx, app.ProcessNextQueuedInboundRequestInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("process next queued inbound request: %v", err)
	}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	aiService := ai.NewService(db)

	approval, _ := createPendingApproval(t, ctx, documentService, workflowService, operator)
	if _, err := aiService.LinkRecommendationApproval(ctx, ai.LinkRecommendationApprovalInput{
		RecommendationID: processResult.Recommendation.ID,
		ApprovalID:       approval.ID,
		Actor:            operator,
	}); err != nil {
		t.Fatalf("link recommendation approval: %v", err)
	}

	handler := app.NewAgentAPIHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/api/review/inbound-requests?status=processed", nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", operatorSession.ID)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound request list status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var listResponse struct {
		Items []struct {
			RequestReference         string `json:"request_reference"`
			Status                   string `json:"status"`
			LastRecommendationStatus string `json:"last_recommendation_status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &listResponse); err != nil {
		t.Fatalf("decode inbound request list: %v", err)
	}
	if len(listResponse.Items) != 1 {
		t.Fatalf("unexpected inbound request item count: %d", len(listResponse.Items))
	}
	if listResponse.Items[0].RequestReference != request.RequestReference || listResponse.Items[0].Status != intake.StatusProcessed {
		t.Fatalf("unexpected inbound request list item: %+v", listResponse.Items[0])
	}
	if listResponse.Items[0].LastRecommendationStatus != ai.RecommendationStatusApprovalRequested {
		t.Fatalf("unexpected recommendation status: %+v", listResponse.Items[0])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/inbound-request-status-summary", nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", operatorSession.ID)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound request summary status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var summaryResponse struct {
		Items []struct {
			Status       string `json:"status"`
			RequestCount int    `json:"request_count"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &summaryResponse); err != nil {
		t.Fatalf("decode inbound request summary: %v", err)
	}
	if len(summaryResponse.Items) == 0 || summaryResponse.Items[0].Status != intake.StatusProcessed {
		t.Fatalf("unexpected inbound request summary items: %+v", summaryResponse.Items)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/inbound-requests/"+request.RequestReference, nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", operatorSession.ID)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound request detail status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var detailResponse struct {
		Request struct {
			RequestReference string `json:"request_reference"`
		} `json:"request"`
		Runs            []struct{} `json:"runs"`
		Recommendations []struct {
			ApprovalID *string `json:"approval_id"`
		} `json:"recommendations"`
		Proposals []struct {
			ApprovalID *string `json:"approval_id"`
		} `json:"proposals"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &detailResponse); err != nil {
		t.Fatalf("decode inbound request detail: %v", err)
	}
	if detailResponse.Request.RequestReference != request.RequestReference {
		t.Fatalf("unexpected inbound request detail reference: %+v", detailResponse.Request)
	}
	if len(detailResponse.Runs) == 0 || len(detailResponse.Recommendations) == 0 || len(detailResponse.Proposals) == 0 {
		t.Fatalf("expected review detail slices, got %+v", detailResponse)
	}
	if detailResponse.Recommendations[0].ApprovalID == nil || *detailResponse.Recommendations[0].ApprovalID != approval.ID {
		t.Fatalf("unexpected recommendation approval linkage: %+v", detailResponse.Recommendations[0])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/processed-proposals?request_reference="+request.RequestReference, nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", operatorSession.ID)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected processed proposal list status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var proposalListResponse struct {
		Items []struct {
			RecommendationID     string  `json:"recommendation_id"`
			RecommendationStatus string  `json:"recommendation_status"`
			ApprovalID           *string `json:"approval_id"`
			ApprovalStatus       *string `json:"approval_status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &proposalListResponse); err != nil {
		t.Fatalf("decode processed proposal list: %v", err)
	}
	if len(proposalListResponse.Items) != 1 {
		t.Fatalf("unexpected processed proposal count: %d", len(proposalListResponse.Items))
	}
	if proposalListResponse.Items[0].ApprovalID == nil || *proposalListResponse.Items[0].ApprovalID != approval.ID {
		t.Fatalf("unexpected processed proposal approval linkage: %+v", proposalListResponse.Items[0])
	}
	if proposalListResponse.Items[0].RecommendationStatus != ai.RecommendationStatusApprovalRequested {
		t.Fatalf("unexpected processed proposal status: %+v", proposalListResponse.Items[0])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/processed-proposals?recommendation_id="+proposalListResponse.Items[0].RecommendationID, nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", operatorSession.ID)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact processed proposal list status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &proposalListResponse); err != nil {
		t.Fatalf("decode exact processed proposal list: %v", err)
	}
	if len(proposalListResponse.Items) != 1 || proposalListResponse.Items[0].RecommendationID == "" {
		t.Fatalf("unexpected exact processed proposal items: %+v", proposalListResponse.Items)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/processed-proposal-status-summary", nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", operatorSession.ID)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected processed proposal summary status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var proposalSummaryResponse struct {
		Items []struct {
			RecommendationStatus string `json:"recommendation_status"`
			ProposalCount        int    `json:"proposal_count"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &proposalSummaryResponse); err != nil {
		t.Fatalf("decode processed proposal summary: %v", err)
	}
	if len(proposalSummaryResponse.Items) == 0 || proposalSummaryResponse.Items[0].RecommendationStatus != ai.RecommendationStatusApprovalRequested {
		t.Fatalf("unexpected processed proposal summary items: %+v", proposalSummaryResponse.Items)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/approval-queue?status=pending", nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", operatorSession.ID)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected approval queue status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var queueResponse struct {
		Items []struct {
			ApprovalID     string `json:"approval_id"`
			ApprovalStatus string `json:"approval_status"`
			QueueCode      string `json:"queue_code"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &queueResponse); err != nil {
		t.Fatalf("decode approval queue: %v", err)
	}
	if len(queueResponse.Items) != 1 || queueResponse.Items[0].ApprovalID != approval.ID || queueResponse.Items[0].ApprovalStatus != "pending" {
		t.Fatalf("unexpected approval queue items: %+v", queueResponse.Items)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/approval-queue?approval_id="+approval.ID, nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", operatorUserID)
	req.Header.Set("X-Workflow-Session-ID", operatorSession.ID)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact approval queue status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &queueResponse); err != nil {
		t.Fatalf("decode exact approval queue: %v", err)
	}
	if len(queueResponse.Items) != 1 || queueResponse.Items[0].ApprovalID != approval.ID {
		t.Fatalf("unexpected exact approval queue items: %+v", queueResponse.Items)
	}

	loginReq := httptest.NewRequest(
		http.MethodPost,
		"/app/login",
		strings.NewReader("org_slug="+orgSlug+"&email="+userEmail+"&device_label=browser-review"),
	)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusSeeOther {
		t.Fatalf("unexpected browser login status: got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	proposalsReq := httptest.NewRequest(http.MethodGet, "/app/review/proposals?request_reference="+request.RequestReference, nil)
	applyResponseCookies(proposalsReq, loginRecorder.Result().Cookies())
	proposalsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(proposalsRecorder, proposalsReq)
	if proposalsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected proposals page status: got %d body=%s", proposalsRecorder.Code, proposalsRecorder.Body.String())
	}
	requireContains(t, proposalsRecorder.Body.String(), "Proposal review")
	requireContains(t, proposalsRecorder.Body.String(), "Proposal status summary")
	requireContains(t, proposalsRecorder.Body.String(), request.RequestReference)
	requireContains(t, proposalsRecorder.Body.String(), ai.RecommendationStatusApprovalRequested)
	requireContains(t, proposalsRecorder.Body.String(), "/app/inbound-requests/"+request.RequestReference)
	requireContains(t, proposalsRecorder.Body.String(), "/app/review/approvals?queue_code="+queueResponse.Items[0].QueueCode+"&amp;status=pending")
	requireContains(t, proposalsRecorder.Body.String(), "/app/review/approvals/"+approval.ID)
	requireContains(t, proposalsRecorder.Body.String(), "/app/review/proposals/"+proposalListResponse.Items[0].RecommendationID)

	inboundRequestsReq := httptest.NewRequest(http.MethodGet, "/app/review/inbound-requests?request_reference="+request.RequestReference, nil)
	applyResponseCookies(inboundRequestsReq, loginRecorder.Result().Cookies())
	inboundRequestsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(inboundRequestsRecorder, inboundRequestsReq)
	if inboundRequestsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound requests review status: got %d body=%s", inboundRequestsRecorder.Code, inboundRequestsRecorder.Body.String())
	}
	requireContains(t, inboundRequestsRecorder.Body.String(), "Inbound-request review")
	requireContains(t, inboundRequestsRecorder.Body.String(), request.RequestReference)
	requireContains(t, inboundRequestsRecorder.Body.String(), ai.RunStatusCompleted)
	requireContains(t, inboundRequestsRecorder.Body.String(), ai.RecommendationStatusApprovalRequested)
	requireContains(t, inboundRequestsRecorder.Body.String(), "/app/inbound-requests/run:"+processResult.Run.ID+"#run-"+processResult.Run.ID)
	requireContains(t, inboundRequestsRecorder.Body.String(), "/app/review/proposals/"+proposalListResponse.Items[0].RecommendationID)

	approvalsReq := httptest.NewRequest(http.MethodGet, "/app/review/approvals?queue_code="+queueResponse.Items[0].QueueCode+"&status=pending", nil)
	applyResponseCookies(approvalsReq, loginRecorder.Result().Cookies())
	approvalsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(approvalsRecorder, approvalsReq)
	if approvalsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected approvals page status: got %d body=%s", approvalsRecorder.Code, approvalsRecorder.Body.String())
	}
	requireContains(t, approvalsRecorder.Body.String(), "Approval review")
	requireContains(t, approvalsRecorder.Body.String(), queueResponse.Items[0].QueueCode)
	requireContains(t, approvalsRecorder.Body.String(), "/app/review/documents/")

	exactApprovalsReq := httptest.NewRequest(http.MethodGet, "/app/review/approvals/"+approval.ID, nil)
	applyResponseCookies(exactApprovalsReq, loginRecorder.Result().Cookies())
	exactApprovalsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(exactApprovalsRecorder, exactApprovalsReq)
	if exactApprovalsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact approvals page status: got %d body=%s", exactApprovalsRecorder.Code, exactApprovalsRecorder.Body.String())
	}
	requireContains(t, exactApprovalsRecorder.Body.String(), "Approval "+approval.ID)
	requireContains(t, exactApprovalsRecorder.Body.String(), approval.ID)
	requireContains(t, exactApprovalsRecorder.Body.String(), "/app/review/audit?entity_type=workflow.approval&amp;entity_id="+approval.ID)

	exactProposalsReq := httptest.NewRequest(http.MethodGet, "/app/review/proposals/"+proposalListResponse.Items[0].RecommendationID, nil)
	applyResponseCookies(exactProposalsReq, loginRecorder.Result().Cookies())
	exactProposalsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(exactProposalsRecorder, exactProposalsReq)
	if exactProposalsRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected exact proposals page status: got %d body=%s", exactProposalsRecorder.Code, exactProposalsRecorder.Body.String())
	}
	requireContains(t, exactProposalsRecorder.Body.String(), "Proposal "+proposalListResponse.Items[0].RecommendationID)
	requireContains(t, exactProposalsRecorder.Body.String(), proposalListResponse.Items[0].RecommendationID)
	requireContains(t, exactProposalsRecorder.Body.String(), "/app/review/audit?entity_type=ai.agent_recommendation&amp;entity_id="+proposalListResponse.Items[0].RecommendationID)

	inboundAuditReq := httptest.NewRequest(http.MethodGet, "/app/review/audit?entity_type=ai.inbound_request&entity_id="+request.ID, nil)
	applyResponseCookies(inboundAuditReq, loginRecorder.Result().Cookies())
	inboundAuditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(inboundAuditRecorder, inboundAuditReq)
	if inboundAuditRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected inbound audit page status: got %d body=%s", inboundAuditRecorder.Code, inboundAuditRecorder.Body.String())
	}
	requireContains(t, inboundAuditRecorder.Body.String(), "/app/inbound-requests/"+request.ID)

	approvalAuditReq := httptest.NewRequest(http.MethodGet, "/app/review/audit?entity_type=workflow.approval&entity_id="+approval.ID, nil)
	applyResponseCookies(approvalAuditReq, loginRecorder.Result().Cookies())
	approvalAuditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(approvalAuditRecorder, approvalAuditReq)
	if approvalAuditRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected approval audit page status: got %d body=%s", approvalAuditRecorder.Code, approvalAuditRecorder.Body.String())
	}
	requireContains(t, approvalAuditRecorder.Body.String(), "/app/review/approvals/"+approval.ID)

	recommendationAuditReq := httptest.NewRequest(http.MethodGet, "/app/review/audit?entity_type=ai.agent_recommendation&entity_id="+proposalListResponse.Items[0].RecommendationID, nil)
	applyResponseCookies(recommendationAuditReq, loginRecorder.Result().Cookies())
	recommendationAuditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(recommendationAuditRecorder, recommendationAuditReq)
	if recommendationAuditRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected recommendation audit page status: got %d body=%s", recommendationAuditRecorder.Code, recommendationAuditRecorder.Body.String())
	}
	requireContains(t, recommendationAuditRecorder.Body.String(), "/app/review/proposals/"+proposalListResponse.Items[0].RecommendationID)
}

func TestAgentAPIReviewSurfacesRejectInvalidExactIDFiltersIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)

	handler := app.NewAgentAPIHandler(db)

	testCases := []struct {
		name string
		path string
	}{
		{name: "approval queue", path: "/api/review/approval-queue?approval_id=not-a-uuid"},
		{name: "documents", path: "/api/review/documents?document_id=not-a-uuid"},
		{name: "journal entries", path: "/api/review/accounting/journal-entries?entry_id=not-a-uuid"},
		{name: "control balances", path: "/api/review/accounting/control-account-balances?account_id=not-a-uuid"},
		{name: "inventory stock", path: "/api/review/inventory/stock?item_id=not-a-uuid"},
		{name: "inventory movements", path: "/api/review/inventory/movements?movement_id=not-a-uuid"},
		{name: "inventory reconciliation", path: "/api/review/inventory/reconciliation?document_id=not-a-uuid"},
		{name: "work order list", path: "/api/review/work-orders?work_order_id=not-a-uuid"},
		{name: "work order detail", path: "/api/review/work-orders/not-a-uuid"},
		{name: "processed proposals", path: "/api/review/processed-proposals?recommendation_id=not-a-uuid"},
		{name: "processed proposals request", path: "/api/review/processed-proposals?request_id=not-a-uuid"},
		{name: "inbound request detail", path: "/api/review/inbound-requests/not-a-uuid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("X-Workflow-Org-ID", orgID)
			req.Header.Set("X-Workflow-User-ID", operatorUserID)
			req.Header.Set("X-Workflow-Session-ID", session.ID)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("unexpected status for %s: got %d body=%s", tc.path, recorder.Code, recorder.Body.String())
			}
			requireContains(t, recorder.Body.String(), `"error":"invalid review filter"`)
		})
	}
}

func TestAgentAPIDecideApprovalIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	_, approverUserID := seedOrgAndUserInOrg(t, ctx, db, identityaccess.RoleApprover, orgID)
	approverSession := startSession(t, ctx, db, orgID, approverUserID)
	approver := identityaccess.Actor{OrgID: orgID, UserID: approverUserID, SessionID: approverSession.ID}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	approval, doc := createPendingApproval(t, ctx, documentService, workflowService, operator)

	handler := app.NewAgentAPIHandler(db)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals/"+approval.ID+"/decision", bytes.NewBufferString(`{"decision":"approved","decision_note":"Looks correct."}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", approverUserID)
	req.Header.Set("X-Workflow-Session-ID", approverSession.ID)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected approval decision status: got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		ApprovalID     string  `json:"approval_id"`
		Status         string  `json:"status"`
		DocumentID     string  `json:"document_id"`
		DocumentStatus string  `json:"document_status"`
		DecisionNote   *string `json:"decision_note"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode approval decision response: %v", err)
	}
	if response.ApprovalID != approval.ID || response.DocumentID != doc.ID {
		t.Fatalf("unexpected approval decision response: %+v", response)
	}
	if response.Status != "approved" || response.DocumentStatus != string(documents.StatusApproved) {
		t.Fatalf("unexpected approval decision states: %+v", response)
	}
	if response.DecisionNote == nil || *response.DecisionNote != "Looks correct." {
		t.Fatalf("unexpected decision note: %+v", response)
	}

	var documentStatus string
	if err := db.QueryRowContext(ctx, `SELECT status FROM documents.documents WHERE id = $1`, doc.ID).Scan(&documentStatus); err != nil {
		t.Fatalf("load document status: %v", err)
	}
	if documentStatus != string(documents.StatusApproved) {
		t.Fatalf("unexpected persisted document status: %s", documentStatus)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/review/approval-queue?status=closed", nil)
	req.Header.Set("X-Workflow-Org-ID", orgID)
	req.Header.Set("X-Workflow-User-ID", approver.UserID)
	req.Header.Set("X-Workflow-Session-ID", approver.SessionID)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected closed approval queue status: got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func seedOrgAndUserInOrg(t *testing.T, ctx context.Context, db *sql.DB, roleCode, orgID string) (string, string) {
	t.Helper()

	var userID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name) VALUES ($1, 'Example User') RETURNING id`,
		"user-"+time.Now().UTC().Format("150405.000000000")+"@example.com",
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO identityaccess.memberships (org_id, user_id, role_code) VALUES ($1, $2, $3)`,
		orgID,
		userID,
		roleCode,
	); err != nil {
		t.Fatalf("insert membership: %v", err)
	}

	return orgID, userID
}

func loadOrgSlugAndUserEmail(t *testing.T, ctx context.Context, db *sql.DB, orgID, userID string) (string, string) {
	t.Helper()

	var orgSlug string
	if err := db.QueryRowContext(ctx, `SELECT slug FROM identityaccess.orgs WHERE id = $1`, orgID).Scan(&orgSlug); err != nil {
		t.Fatalf("load org slug: %v", err)
	}

	var userEmail string
	if err := db.QueryRowContext(ctx, `SELECT email FROM identityaccess.users WHERE id = $1`, userID).Scan(&userEmail); err != nil {
		t.Fatalf("load user email: %v", err)
	}

	return orgSlug, userEmail
}

func applyResponseCookies(req *http.Request, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
}

type multipartUpload struct {
	FileName    string
	ContentType string
	Content     []byte
}

func newMultipartRequest(t *testing.T, method, target string, fields map[string]string, files map[string]multipartUpload) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write multipart field %s: %v", key, err)
		}
	}
	for fieldName, upload := range files {
		part, err := writer.CreateFormFile(fieldName, upload.FileName)
		if err != nil {
			t.Fatalf("create multipart file %s: %v", fieldName, err)
		}
		if _, err := io.Copy(part, bytes.NewReader(upload.Content)); err != nil {
			t.Fatalf("write multipart content %s: %v", fieldName, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(method, target, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func requireContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("expected response to contain %q, body=%s", want, body)
	}
}

func requireNotContains(t *testing.T, body, unwanted string) {
	t.Helper()
	if strings.Contains(body, unwanted) {
		t.Fatalf("expected response not to contain %q, body=%s", unwanted, body)
	}
}

func requireHiddenInputValue(t *testing.T, body, name string) string {
	t.Helper()

	pattern := regexp.MustCompile(`name="` + regexp.QuoteMeta(name) + `" value="([^"]+)"`)
	matches := pattern.FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatalf("expected hidden input %q in body=%s", name, body)
	}
	return matches[1]
}

func requireRequestReferenceFromPath(t *testing.T, path string) string {
	t.Helper()

	pattern := regexp.MustCompile(`/app/inbound-requests/([^?]+)`)
	matches := pattern.FindStringSubmatch(path)
	if len(matches) != 2 {
		t.Fatalf("expected inbound-request detail path, got %q", path)
	}
	return matches[1]
}

func createPendingApproval(t *testing.T, ctx context.Context, documentService *documents.Service, workflowService *workflow.Service, actor identityaccess.Actor) (workflow.Approval, documents.Document) {
	t.Helper()

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Approval-backed invoice",
		Actor:    actor,
	})
	if err != nil {
		t.Fatalf("create draft document: %v", err)
	}

	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("submit document: %v", err)
	}

	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance",
		Reason:     "needs approval",
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}

	return approval, doc
}

func postApprovedGSTInvoice(t *testing.T, ctx context.Context, accountingService *accounting.Service, documentService *documents.Service, workflowService *workflow.Service, operator, approver identityaccess.Actor) {
	t.Helper()

	doc, _, err := accountingService.CreateInvoice(ctx, accounting.CreateInvoiceInput{
		Title:          "Posted GST invoice",
		InvoiceRole:    accounting.InvoiceRoleSales,
		CurrencyCode:   "INR",
		ReferenceValue: "INV-TEST-1001",
		Summary:        "Browser accounting review test invoice",
		Actor:          operator,
	})
	if err != nil {
		t.Fatalf("create invoice draft: %v", err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit invoice draft: %v", err)
	}
	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance",
		Reason:     "post invoice",
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}
	if _, _, err := workflowService.DecideApproval(ctx, workflow.DecideApprovalInput{
		ApprovalID: approval.ID,
		Decision:   "approved",
		Actor:      approver,
	}); err != nil {
		t.Fatalf("approve invoice: %v", err)
	}

	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               operator,
	})
	gstOutput := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "2101",
		Name:                "GST Output",
		AccountClass:        accounting.AccountClassLiability,
		ControlType:         accounting.ControlTypeGSTOutput,
		AllowsDirectPosting: false,
		Actor:               operator,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4000",
		Name:         "Service Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        operator,
	})
	gst18 := createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "GST18",
		Name:             "GST Output 18%",
		TaxType:          accounting.TaxTypeGST,
		RateBasisPoints:  1800,
		PayableAccountID: gstOutput.ID,
		Actor:            operator,
	})

	if _, _, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   doc.ID,
		Summary:      "Post approved invoice with GST",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		EffectiveOn:  time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC),
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 177000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 150000},
			{AccountID: gstOutput.ID, Description: "GST payable", CreditMinor: 27000, TaxCode: gst18.Code},
		},
		Actor: operator,
	}); err != nil {
		t.Fatalf("post invoice: %v", err)
	}
}

func seedBrowserReviewData(t *testing.T, ctx context.Context, documentService *documents.Service, workflowService *workflow.Service, accountingService *accounting.Service, inventoryService *inventoryops.Service, workOrderService *workorders.Service, workforceService *workforce.Service, operator, approver identityaccess.Actor) workorders.WorkOrder {
	t.Helper()

	workOrderResult, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-RPT-1001",
		Title:         "Review execution chain",
		Summary:       "Browser reporting coverage",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}

	worker := createWorker(t, ctx, workforceService, workforce.CreateWorkerInput{
		WorkerCode:             "TECH-RPT-1",
		DisplayName:            "Reporting Technician",
		DefaultHourlyCostMinor: 3600,
		CostCurrencyCode:       "INR",
		Actor:                  operator,
	})
	task, err := workflowService.CreateTask(ctx, workflow.CreateTaskInput{
		ContextType:         "work_order",
		ContextID:           workOrderResult.WorkOrder.ID,
		Title:               "Inspect and post",
		QueueCode:           "dispatch",
		AccountableWorkerID: worker.ID,
		Actor:               operator,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if _, err := workflowService.UpdateTaskStatus(ctx, workflow.UpdateTaskStatusInput{
		TaskID: task.ID,
		Status: "completed",
		Actor:  operator,
	}); err != nil {
		t.Fatalf("complete task: %v", err)
	}

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "RPT-MAT-1",
		Name:         "Reporting Material",
		ItemRole:     inventoryops.ItemRoleServiceMaterial,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "RPT-WH-1",
		Name:         "Reporting Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	receiptDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "inventory_receipt", "Inventory receipt")
	if _, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: receiptDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:                item.ID,
			MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:   inventoryops.UsageBillable,
			DestinationLocationID: warehouse.ID,
			QuantityMilli:         5000,
		}},
		Actor: operator,
	}); err != nil {
		t.Fatalf("capture receipt: %v", err)
	}

	issueDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "inventory_issue", "Inventory issue")
	if _, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: issueDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:               item.ID,
			MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:  inventoryops.UsageBillable,
			SourceLocationID:     warehouse.ID,
			QuantityMilli:        2000,
			CostMinor:            5400,
			CostCurrencyCode:     "INR",
			AccountingHandoff:    true,
			ExecutionContextType: inventoryops.ExecutionContextWorkOrder,
			ExecutionContextID:   workOrderResult.WorkOrder.WorkOrderCode,
		}},
		Actor: operator,
	}); err != nil {
		t.Fatalf("capture issue: %v", err)
	}

	if _, err := workOrderService.SyncInventoryUsage(ctx, workorders.SyncInventoryUsageInput{
		WorkOrderID: workOrderResult.WorkOrder.ID,
		Actor:       operator,
	}); err != nil {
		t.Fatalf("sync inventory usage: %v", err)
	}

	startedAt := time.Date(2026, 3, 21, 9, 0, 0, 0, time.UTC)
	laborEntry, err := workforceService.RecordLabor(ctx, workforce.RecordLaborInput{
		WorkerID:    worker.ID,
		WorkOrderID: workOrderResult.WorkOrder.ID,
		TaskID:      task.ID,
		StartedAt:   startedAt,
		EndedAt:     startedAt.Add(2 * time.Hour),
		Note:        "Execution review labor",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("record labor: %v", err)
	}

	materialExpense := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "5101",
		Name:         "Material Expense",
		AccountClass: accounting.AccountClassExpense,
		Actor:        operator,
	})
	laborExpense := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "5102",
		Name:         "Labor Expense",
		AccountClass: accounting.AccountClassExpense,
		Actor:        operator,
	})
	accruedOffset := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "2201",
		Name:         "Accrued Costs",
		AccountClass: accounting.AccountClassLiability,
		Actor:        operator,
	})

	laborJournalDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "journal", "Labor posting")
	if _, err := accountingService.PostWorkOrderLabor(ctx, accounting.PostWorkOrderLaborInput{
		DocumentID:       laborJournalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: laborExpense.ID,
		OffsetAccountID:  accruedOffset.ID,
		Summary:          "Post labor review costs",
		EffectiveOn:      startedAt,
		Actor:            operator,
	}); err != nil {
		t.Fatalf("post work order labor: %v", err)
	}

	materialJournalDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "journal", "Material posting")
	if _, err := accountingService.PostWorkOrderInventory(ctx, accounting.PostWorkOrderInventoryInput{
		DocumentID:       materialJournalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: materialExpense.ID,
		OffsetAccountID:  accruedOffset.ID,
		Summary:          "Post material review costs",
		EffectiveOn:      startedAt,
		Actor:            operator,
	}); err != nil {
		t.Fatalf("post work order inventory: %v", err)
	}

	if laborEntry.ID == "" {
		t.Fatal("expected labor entry id")
	}
	return workOrderResult.WorkOrder
}

func prepareApprovedDocumentOfType(t *testing.T, ctx context.Context, documentService *documents.Service, workflowService *workflow.Service, operator, approver identityaccess.Actor, typeCode, title string) documents.Document {
	t.Helper()

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: typeCode,
		Title:    title,
		Actor:    operator,
	})
	if err != nil {
		t.Fatalf("create draft %s document: %v", typeCode, err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit %s document: %v", typeCode, err)
	}
	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "operations",
		Reason:     "prepare review data",
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("request %s approval: %v", typeCode, err)
	}
	_, doc, err = workflowService.DecideApproval(ctx, workflow.DecideApprovalInput{
		ApprovalID: approval.ID,
		Decision:   "approved",
		Actor:      approver,
	})
	if err != nil {
		t.Fatalf("approve %s document: %v", typeCode, err)
	}
	return doc
}

func createItem(t *testing.T, ctx context.Context, service *inventoryops.Service, input inventoryops.CreateItemInput) inventoryops.Item {
	t.Helper()

	item, err := service.CreateItem(ctx, input)
	if err != nil {
		t.Fatalf("create item %s: %v", input.SKU, err)
	}
	return item
}

func createLocation(t *testing.T, ctx context.Context, service *inventoryops.Service, input inventoryops.CreateLocationInput) inventoryops.Location {
	t.Helper()

	location, err := service.CreateLocation(ctx, input)
	if err != nil {
		t.Fatalf("create location %s: %v", input.Code, err)
	}
	return location
}

func createWorker(t *testing.T, ctx context.Context, service *workforce.Service, input workforce.CreateWorkerInput) workforce.Worker {
	t.Helper()

	worker, err := service.CreateWorker(ctx, input)
	if err != nil {
		t.Fatalf("create worker %s: %v", input.WorkerCode, err)
	}
	return worker
}

func createLedgerAccount(t *testing.T, ctx context.Context, service *accounting.Service, input accounting.CreateLedgerAccountInput) accounting.LedgerAccount {
	t.Helper()

	account, err := service.CreateLedgerAccount(ctx, input)
	if err != nil {
		t.Fatalf("create ledger account %s: %v", input.Code, err)
	}
	return account
}

func createTaxCode(t *testing.T, ctx context.Context, service *accounting.Service, input accounting.CreateTaxCodeInput) accounting.TaxCode {
	t.Helper()

	code, err := service.CreateTaxCode(ctx, input)
	if err != nil {
		t.Fatalf("create tax code %s: %v", input.Code, err)
	}
	return code
}
