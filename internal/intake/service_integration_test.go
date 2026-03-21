package intake_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"workflow_app/internal/ai"
	"workflow_app/internal/attachments"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/reporting"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
)

func TestInboundRequestLifecycleAndReportingIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	aiService := ai.NewService(db)
	intakeService := intake.NewService(db)
	attachmentService := attachments.NewService(db)
	reportingService := reporting.NewService(db)

	request, err := intakeService.CreateDraft(ctx, intake.CreateDraftInput{
		OriginType: intake.OriginHuman,
		Channel:    "browser",
		Metadata: map[string]any{
			"source": "integration-test",
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("create draft request: %v", err)
	}

	message, err := intakeService.AddMessage(ctx, intake.AddMessageInput{
		RequestID:   request.ID,
		MessageRole: intake.MessageRoleRequest,
		TextContent: "Please prepare the invoice approval package from this recording.",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("add draft message: %v", err)
	}

	attachment, err := attachmentService.CreateAttachment(ctx, attachments.CreateAttachmentInput{
		OriginalFileName: "request.m4a",
		MediaType:        "audio/m4a",
		Content:          []byte("fake-audio"),
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("create attachment: %v", err)
	}
	if _, err := attachmentService.LinkRequestMessage(ctx, attachments.LinkRequestMessageInput{
		RequestMessageID: message.ID,
		AttachmentID:     attachment.ID,
		Actor:            operator,
	}); err != nil {
		t.Fatalf("link attachment: %v", err)
	}

	request, err = intakeService.QueueRequest(ctx, intake.QueueRequestInput{
		RequestID: request.ID,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("queue request: %v", err)
	}
	if request.Status != intake.StatusQueued {
		t.Fatalf("unexpected queued status: %s", request.Status)
	}

	if _, err := aiService.StartRun(ctx, ai.StartRunInput{
		AgentRole:        ai.RunRoleCoordinator,
		CapabilityCode:   "workflow.coordination",
		InboundRequestID: request.ID,
		RequestText:      "process inbound request before claim",
		Actor:            operator,
	}); !errors.Is(err, ai.ErrRunNotActive) {
		t.Fatalf("expected queued request to reject direct run start, got %v", err)
	}

	request, err = intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("claim queued request: %v", err)
	}
	if request.Status != intake.StatusProcessing {
		t.Fatalf("unexpected claimed status: %s", request.Status)
	}

	run, err := aiService.StartRun(ctx, ai.StartRunInput{
		AgentRole:        ai.RunRoleCoordinator,
		CapabilityCode:   "workflow.coordination",
		InboundRequestID: request.ID,
		RequestText:      "review inbound request and prepare approval",
		Metadata: map[string]any{
			"request_id": request.ID,
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("start linked run: %v", err)
	}
	if !run.InboundRequestID.Valid || run.InboundRequestID.String != request.ID {
		t.Fatalf("unexpected inbound request link on run: %+v", run.InboundRequestID)
	}

	derivedText, err := attachmentService.RecordDerivedText(ctx, attachments.RecordDerivedTextInput{
		SourceAttachmentID: attachment.ID,
		RequestMessageID:   message.ID,
		CreatedByRunID:     run.ID,
		DerivativeType:     attachments.DerivativeTranscription,
		ContentText:        "Customer asked for invoice approval prep from voice note.",
		Actor:              operator,
	})
	if err != nil {
		t.Fatalf("record derived text: %v", err)
	}
	if derivedText.DerivativeType != attachments.DerivativeTranscription {
		t.Fatalf("unexpected derivative type: %s", derivedText.DerivativeType)
	}

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Inbound request invoice",
		Actor:    operator,
	})
	if err != nil {
		t.Fatalf("create document draft: %v", err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit document: %v", err)
	}

	artifact, err := aiService.CreateArtifact(ctx, ai.CreateArtifactInput{
		RunID:        run.ID,
		ArtifactType: "transcription",
		Title:        "Voice transcription",
		Payload: map[string]any{
			"derived_text_id": derivedText.ID,
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("create artifact: %v", err)
	}

	recommendation, err := aiService.CreateRecommendation(ctx, ai.CreateRecommendationInput{
		RunID:              run.ID,
		ArtifactID:         artifact.ID,
		RecommendationType: "request_approval",
		Summary:            "Request finance approval for the submitted invoice",
		Payload: map[string]any{
			"document_id": doc.ID,
			"queue_code":  "finance-review",
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("create recommendation: %v", err)
	}

	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance-review",
		Reason:     "inbound request processing produced a submitted invoice",
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}

	recommendation, err = aiService.LinkRecommendationApproval(ctx, ai.LinkRecommendationApprovalInput{
		RecommendationID: recommendation.ID,
		ApprovalID:       approval.ID,
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("link recommendation approval: %v", err)
	}

	if _, err := aiService.CompleteRun(ctx, ai.CompleteRunInput{
		RunID:   run.ID,
		Status:  ai.RunStatusCompleted,
		Summary: "prepared approval package from inbound request",
		Metadata: map[string]any{
			"approval_id": approval.ID,
		},
		Actor: operator,
	}); err != nil {
		t.Fatalf("complete run: %v", err)
	}

	request, err = intakeService.AdvanceRequest(ctx, intake.AdvanceRequestInput{
		RequestID: request.ID,
		Status:    intake.StatusProcessed,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("mark request processed: %v", err)
	}
	if request.Status != intake.StatusProcessed {
		t.Fatalf("unexpected processed status: %s", request.Status)
	}

	requests, err := reportingService.ListInboundRequests(ctx, reporting.ListInboundRequestsInput{
		Status: intake.StatusProcessed,
		Limit:  10,
		Actor:  operator,
	})
	if err != nil {
		t.Fatalf("list inbound requests: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("unexpected inbound request review count: %d", len(requests))
	}
	if requests[0].AttachmentCount != 1 || requests[0].MessageCount != 1 {
		t.Fatalf("unexpected request counts: attachments=%d messages=%d", requests[0].AttachmentCount, requests[0].MessageCount)
	}
	if !requests[0].LastRecommendationID.Valid || requests[0].LastRecommendationID.String != recommendation.ID {
		t.Fatalf("unexpected last recommendation: %+v want %s", requests[0].LastRecommendationID, recommendation.ID)
	}

	detail, err := reportingService.GetInboundRequestDetail(ctx, reporting.GetInboundRequestDetailInput{
		RequestID: request.ID,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("get inbound request detail: %v", err)
	}
	if len(detail.Messages) != 1 || len(detail.Attachments) != 1 || len(detail.Runs) != 1 || len(detail.Proposals) != 1 {
		t.Fatalf("unexpected detail sizes: messages=%d attachments=%d runs=%d proposals=%d", len(detail.Messages), len(detail.Attachments), len(detail.Runs), len(detail.Proposals))
	}
	if !detail.Attachments[0].LatestDerivedText.Valid || detail.Attachments[0].LatestDerivedText.String == "" {
		t.Fatal("expected latest derived text in attachment review")
	}
	if detail.Proposals[0].ApprovalID.String != approval.ID || detail.Proposals[0].DocumentID.String != doc.ID {
		t.Fatalf("unexpected proposal linkage: %+v", detail.Proposals[0])
	}

	proposals, err := reportingService.ListProcessedProposals(ctx, reporting.ListProcessedProposalsInput{
		Status: ai.RecommendationStatusApprovalRequested,
		Limit:  10,
		Actor:  operator,
	})
	if err != nil {
		t.Fatalf("list processed proposals: %v", err)
	}
	if len(proposals) != 1 {
		t.Fatalf("unexpected processed proposal count: %d", len(proposals))
	}
	if proposals[0].RequestID != request.ID || proposals[0].DocumentID.String != doc.ID {
		t.Fatalf("unexpected processed proposal row: %+v", proposals[0])
	}
}

func TestCancelQueuedRequestPreventsClaimIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	intakeService := intake.NewService(db)
	reportingService := reporting.NewService(db)

	cancelled := createQueuedRequest(t, ctx, intakeService, operator, "cancel me")
	active := createQueuedRequest(t, ctx, intakeService, operator, "process me")

	cancelled, err := intakeService.CancelRequest(ctx, intake.CancelRequestInput{
		RequestID: cancelled.ID,
		Reason:    "operator withdrew request before processing",
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("cancel queued request: %v", err)
	}
	if cancelled.Status != intake.StatusCancelled {
		t.Fatalf("unexpected cancelled status: %s", cancelled.Status)
	}

	claimed, err := intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("claim next queued request: %v", err)
	}
	if claimed.ID != active.ID {
		t.Fatalf("expected active request to be claimed, got %s want %s", claimed.ID, active.ID)
	}

	cancelledRows, err := reportingService.ListInboundRequests(ctx, reporting.ListInboundRequestsInput{
		Status: intake.StatusCancelled,
		Limit:  10,
		Actor:  operator,
	})
	if err != nil {
		t.Fatalf("list cancelled requests: %v", err)
	}
	if len(cancelledRows) != 1 || cancelledRows[0].RequestID != cancelled.ID {
		t.Fatalf("unexpected cancelled request review rows: %+v", cancelledRows)
	}
}

func createQueuedRequest(t *testing.T, ctx context.Context, service *intake.Service, actor identityaccess.Actor, text string) intake.InboundRequest {
	t.Helper()

	request, err := service.CreateDraft(ctx, intake.CreateDraftInput{
		OriginType: intake.OriginHuman,
		Channel:    "browser",
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("create draft request: %v", err)
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

func seedOrgAndUser(t *testing.T, ctx context.Context, db *sql.DB, roleCode, existingOrgID string) (string, string) {
	t.Helper()

	orgID := existingOrgID
	if orgID == "" {
		if err := db.QueryRowContext(ctx, `
INSERT INTO identityaccess.orgs (slug, name)
VALUES ($1, $2)
RETURNING id;`,
			"org-"+roleCode,
			"Org "+roleCode,
		).Scan(&orgID); err != nil {
			t.Fatalf("insert org: %v", err)
		}
	}

	var userID string
	if err := db.QueryRowContext(ctx, `
INSERT INTO identityaccess.users (email, display_name)
VALUES ($1, $2)
RETURNING id;`,
		roleCode+"@example.com",
		roleCode+" user",
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if _, err := db.ExecContext(ctx, `
INSERT INTO identityaccess.memberships (org_id, user_id, role_code)
VALUES ($1, $2, $3);`,
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
		DeviceLabel:      "integration-test",
		RefreshTokenHash: "refresh-token-hash",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	return session
}
