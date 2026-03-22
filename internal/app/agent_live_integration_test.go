//go:build integration

package app_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"workflow_app/internal/app"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/testsupport/dbtest"
)

func TestOpenAIAgentProcessorLiveIntegration(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" || os.Getenv("OPENAI_MODEL") == "" {
		t.Skip("OPENAI_API_KEY and OPENAI_MODEL are required for live integration")
	}

	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	request := createQueuedRequest(t, ctx, db, operator, "Customer reports an urgent warehouse pump failure and needs operator follow-up.")

	processor, err := app.NewOpenAIAgentProcessorFromEnv(db)
	if err != nil {
		t.Fatalf("new live OpenAI agent processor: %v", err)
	}

	result, err := processor.ProcessNextQueuedInboundRequest(ctx, app.ProcessNextQueuedInboundRequestInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("process next queued inbound request with live provider: %v", err)
	}

	if result.Request.RequestReference != request.RequestReference {
		t.Fatalf("unexpected request reference: got %s want %s", result.Request.RequestReference, request.RequestReference)
	}
	if result.Request.Status != intake.StatusProcessed {
		t.Fatalf("unexpected request status: %s", result.Request.Status)
	}
	if result.Run.Status != "completed" {
		t.Fatalf("unexpected run status: %s", result.Run.Status)
	}
	if result.Artifact.ID == "" {
		t.Fatal("expected artifact to be created")
	}
	if result.Recommendation.ID == "" {
		t.Fatal("expected recommendation to be created")
	}
	if result.Recommendation.Summary == "" {
		t.Fatal("expected recommendation summary")
	}
}
