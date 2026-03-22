package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"workflow_app/internal/ai"
	"workflow_app/internal/app"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/testsupport/dbtest"
)

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
