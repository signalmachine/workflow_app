package app_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"workflow_app/internal/ai"
	"workflow_app/internal/app"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/testsupport/dbtest"
)

func TestNewAgentProcessorRequiresProvider(t *testing.T) {
	db := &sql.DB{}
	processor, err := app.NewAgentProcessor(db, nil)
	if !errors.Is(err, app.ErrAgentProviderNotConfigured) {
		t.Fatalf("expected missing provider error, got processor=%v err=%v", processor, err)
	}
}

func TestNewOpenAIAgentProcessorFromEnvRequiresConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_MODEL", "")

	processor, err := app.NewOpenAIAgentProcessorFromEnv(&sql.DB{})
	if !errors.Is(err, app.ErrAgentProviderNotConfigured) {
		t.Fatalf("expected missing provider config error, got processor=%v err=%v", processor, err)
	}
}

func TestAgentProcessorProcessNextQueuedInboundRequestIntegration(t *testing.T) {
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
			ProviderResponseID: "resp_app_test_123",
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

	result, err := processor.ProcessNextQueuedInboundRequest(ctx, app.ProcessNextQueuedInboundRequestInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("process next queued inbound request: %v", err)
	}

	if result.Request.RequestReference != request.RequestReference {
		t.Fatalf("unexpected request reference: got %s want %s", result.Request.RequestReference, request.RequestReference)
	}
	if result.Request.Status != intake.StatusProcessed {
		t.Fatalf("unexpected request status: %s", result.Request.Status)
	}
	if result.Run.Status != ai.RunStatusCompleted {
		t.Fatalf("unexpected run status: %s", result.Run.Status)
	}
	if result.Artifact.ID == "" {
		t.Fatal("expected artifact to be created")
	}
	if result.Recommendation.ID == "" {
		t.Fatal("expected recommendation to be created")
	}
}

func TestAgentProcessorRejectsGenericTransientStatusOnlyBriefIntegration(t *testing.T) {
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
			ProviderResponseID: "resp_app_test_stale_123",
			Model:              "gpt-5.2",
			Summary:            "The request is currently processing and still needs review.",
			Priority:           "high",
			ArtifactTitle:      "Inbound request review brief",
			ArtifactBody:       "Queue status shows the request is in processing, so the operator should wait.",
			Rationale: []string{
				"The queue indicates active processing.",
			},
			NextActions: []string{
				"Monitor the queue.",
			},
		},
	})
	if err != nil {
		t.Fatalf("new agent processor: %v", err)
	}

	_, err = processor.ProcessNextQueuedInboundRequest(ctx, app.ProcessNextQueuedInboundRequestInput{
		Channel: "browser",
		Actor:   operator,
	})
	if !errors.Is(err, ai.ErrInvalidCoordinatorOutput) {
		t.Fatalf("expected invalid coordinator output error, got %v", err)
	}

	detailService := intake.NewService(db)
	detail, err := detailService.GetRequest(ctx, intake.GetRequestInput{
		RequestID: request.ID,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("get request after failed processing: %v", err)
	}
	if detail.Status != intake.StatusFailed {
		t.Fatalf("unexpected request status after failed processing: %s", detail.Status)
	}
}

type fakeCoordinatorProvider struct {
	output ai.CoordinatorProviderOutput
	err    error
}

func (f fakeCoordinatorProvider) ExecuteInboundRequest(context.Context, ai.CoordinatorProviderInput) (ai.CoordinatorProviderOutput, error) {
	if f.err != nil {
		return ai.CoordinatorProviderOutput{}, f.err
	}
	return f.output, nil
}

func seedOrgAndUser(t *testing.T, ctx context.Context, db *sql.DB, roleCode string) (string, string) {
	t.Helper()

	var orgID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.orgs (slug, name) VALUES ($1, $2) RETURNING id`,
		"acme-"+time.Now().UTC().Format("150405.000000000"),
		"Acme",
	).Scan(&orgID); err != nil {
		t.Fatalf("insert org: %v", err)
	}

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

func startSession(t *testing.T, ctx context.Context, db *sql.DB, orgID, userID string) identityaccess.Session {
	t.Helper()

	service := identityaccess.NewService(db)
	session, err := service.StartSession(ctx, identityaccess.StartSessionInput{
		OrgID:            orgID,
		UserID:           userID,
		DeviceLabel:      "test-device",
		RefreshTokenHash: "token-" + time.Now().UTC().Format("150405.000000000"),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	return session
}

func createQueuedRequest(t *testing.T, ctx context.Context, db *sql.DB, actor identityaccess.Actor, text string) intake.InboundRequest {
	t.Helper()

	service := intake.NewService(db)
	request, err := service.CreateDraft(ctx, intake.CreateDraftInput{
		OriginType: intake.OriginHuman,
		Channel:    "browser",
		Metadata: map[string]any{
			"submitter_label": "integration-test",
		},
		Actor: actor,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	if _, err := service.AddMessage(ctx, intake.AddMessageInput{
		RequestID:   request.ID,
		MessageRole: intake.MessageRoleRequest,
		TextContent: text,
		Actor:       actor,
	}); err != nil {
		t.Fatalf("add request message: %v", err)
	}

	request, err = service.QueueRequest(ctx, intake.QueueRequestInput{
		RequestID: request.ID,
		Actor:     actor,
	})
	if err != nil {
		t.Fatalf("queue request: %v", err)
	}

	return request
}
