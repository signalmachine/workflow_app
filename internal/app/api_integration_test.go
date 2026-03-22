package app_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"workflow_app/internal/ai"
	"workflow_app/internal/app"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
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
		RequestID        string   `json:"request_id"`
		RequestReference string   `json:"request_reference"`
		Status           string   `json:"status"`
		MessageID        string   `json:"message_id"`
		AttachmentIDs    []string `json:"attachment_ids"`
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
	if got := downloadRecorder.Body.String(); got != "urgent pump failure details" {
		t.Fatalf("unexpected attachment payload: %q", got)
	}
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

func TestAgentAPIReviewSurfacesIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
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
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &queueResponse); err != nil {
		t.Fatalf("decode approval queue: %v", err)
	}
	if len(queueResponse.Items) != 1 || queueResponse.Items[0].ApprovalID != approval.ID || queueResponse.Items[0].ApprovalStatus != "pending" {
		t.Fatalf("unexpected approval queue items: %+v", queueResponse.Items)
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
