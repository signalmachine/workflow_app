package ai_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"workflow_app/internal/ai"
	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/reporting"
	"workflow_app/internal/testsupport/dbtest"
)

func TestCoordinatorProcessNextQueuedIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	intakeService := intake.NewService(db)
	attachmentService := attachments.NewService(db)
	reportingService := reporting.NewService(db)

	request, err := intakeService.CreateDraft(ctx, intake.CreateDraftInput{
		OriginType: intake.OriginHuman,
		Channel:    "browser",
		Metadata: map[string]any{
			"submitter_label": "front desk",
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	message, err := intakeService.AddMessage(ctx, intake.AddMessageInput{
		RequestID:   request.ID,
		MessageRole: intake.MessageRoleRequest,
		TextContent: "Customer reported a failed pump and attached a voice note.",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("add message: %v", err)
	}

	attachment, err := attachmentService.CreateAttachment(ctx, attachments.CreateAttachmentInput{
		OriginalFileName: "voice-note.m4a",
		MediaType:        "audio/mp4",
		Content:          []byte("placeholder audio"),
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("create attachment: %v", err)
	}

	if _, err := attachmentService.LinkRequestMessage(ctx, attachments.LinkRequestMessageInput{
		RequestMessageID: message.ID,
		AttachmentID:     attachment.ID,
		LinkRole:         attachments.LinkRoleSource,
		Actor:            operator,
	}); err != nil {
		t.Fatalf("link request message: %v", err)
	}

	if _, err := attachmentService.RecordDerivedText(ctx, attachments.RecordDerivedTextInput{
		SourceAttachmentID: attachment.ID,
		RequestMessageID:   message.ID,
		DerivativeType:     attachments.DerivativeTranscription,
		ContentText:        "Pump at the warehouse is failing intermittently and needs urgent inspection.",
		Actor:              operator,
	}); err != nil {
		t.Fatalf("record derived text: %v", err)
	}

	request, err = intakeService.QueueRequest(ctx, intake.QueueRequestInput{
		RequestID: request.ID,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("queue request: %v", err)
	}

	coordinator := ai.NewCoordinator(db, fakeCoordinatorProvider{
		output: ai.CoordinatorProviderOutput{
			ProviderName:       "openai",
			ProviderResponseID: "resp_test_123",
			Model:              "gpt-5.2",
			Summary:            "Operator review needed for an urgent equipment-failure request.",
			Priority:           "urgent",
			ArtifactTitle:      "Inbound request review brief",
			ArtifactBody:       "Customer reports a failing pump at the warehouse. Review and dispatch follow-up.",
			Rationale: []string{
				"Equipment failure can affect active operations.",
				"The attached transcription indicates urgency.",
			},
			NextActions: []string{
				"Review the request details and confirm the affected site.",
				"Create or route a work-order proposal after operator confirmation.",
			},
			InputTokens:  111,
			OutputTokens: 37,
			TotalTokens:  148,
		},
	})

	result, err := coordinator.ProcessNextQueued(ctx, ai.ProcessNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("process next queued request: %v", err)
	}

	if result.Request.Status != intake.StatusProcessed {
		t.Fatalf("unexpected request status: %s", result.Request.Status)
	}
	if result.Run.Status != ai.RunStatusCompleted {
		t.Fatalf("unexpected run status: %s", result.Run.Status)
	}
	if result.Step.Status != ai.StepStatusCompleted {
		t.Fatalf("unexpected step status: %s", result.Step.Status)
	}
	if result.Recommendation.Status != ai.RecommendationStatusProposed {
		t.Fatalf("unexpected recommendation status: %s", result.Recommendation.Status)
	}

	detail, err := reportingService.GetInboundRequestDetail(ctx, reporting.GetInboundRequestDetailInput{
		RequestReference: request.RequestReference,
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("get inbound request detail: %v", err)
	}

	if len(detail.Runs) != 1 {
		t.Fatalf("unexpected run count: %d", len(detail.Runs))
	}
	if len(detail.Artifacts) != 1 {
		t.Fatalf("unexpected artifact count: %d", len(detail.Artifacts))
	}
	if len(detail.Recommendations) != 1 {
		t.Fatalf("unexpected recommendation count: %d", len(detail.Recommendations))
	}

	var artifactPayload map[string]any
	if err := json.Unmarshal(detail.Artifacts[0].Payload, &artifactPayload); err != nil {
		t.Fatalf("unmarshal artifact payload: %v", err)
	}
	if artifactPayload["provider_response_id"] != "resp_test_123" {
		t.Fatalf("unexpected provider response id in artifact: %+v", artifactPayload)
	}
	if artifactPayload["priority"] != "urgent" {
		t.Fatalf("unexpected priority in artifact: %+v", artifactPayload)
	}

	var recommendationPayload map[string]any
	if err := json.Unmarshal(detail.Recommendations[0].Payload, &recommendationPayload); err != nil {
		t.Fatalf("unmarshal recommendation payload: %v", err)
	}
	if recommendationPayload["request_reference"] != request.RequestReference {
		t.Fatalf("unexpected recommendation request reference: %+v", recommendationPayload)
	}
	if recommendationPayload["priority"] != "urgent" {
		t.Fatalf("unexpected recommendation priority: %+v", recommendationPayload)
	}
}

func TestCoordinatorMarksRequestFailedOnProviderErrorIntegration(t *testing.T) {
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

	request := createQueuedRequest(t, ctx, intakeService, operator, "provider should fail for this request")

	coordinator := ai.NewCoordinator(db, fakeCoordinatorProvider{
		err: errors.New("upstream provider timeout"),
	})

	_, err := coordinator.ProcessNextQueued(ctx, ai.ProcessNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err == nil {
		t.Fatal("expected provider-backed coordinator failure")
	}

	detail, err := reportingService.GetInboundRequestDetail(ctx, reporting.GetInboundRequestDetailInput{
		RequestReference: request.RequestReference,
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("get failed request detail: %v", err)
	}

	if detail.Request.Status != intake.StatusFailed {
		t.Fatalf("unexpected failed request status: %s", detail.Request.Status)
	}
	if detail.Request.FailureReason != "upstream provider timeout" {
		t.Fatalf("unexpected failure reason: %q", detail.Request.FailureReason)
	}
	if len(detail.Runs) != 1 {
		t.Fatalf("unexpected run count after failure: %d", len(detail.Runs))
	}
	if detail.Runs[0].Status != ai.RunStatusFailed {
		t.Fatalf("unexpected failed run status: %s", detail.Runs[0].Status)
	}
	if len(detail.Steps) != 1 {
		t.Fatalf("unexpected step count after failure: %d", len(detail.Steps))
	}
	if detail.Steps[0].Status != ai.StepStatusFailed {
		t.Fatalf("unexpected failed step status: %s", detail.Steps[0].Status)
	}
}

type fakeCoordinatorProvider struct {
	output ai.CoordinatorProviderOutput
	err    error
}

func (p fakeCoordinatorProvider) ExecuteInboundRequest(context.Context, ai.CoordinatorProviderInput) (ai.CoordinatorProviderOutput, error) {
	if p.err != nil {
		return ai.CoordinatorProviderOutput{}, p.err
	}
	return p.output, nil
}

func createQueuedRequest(t *testing.T, ctx context.Context, intakeService *intake.Service, actor identityaccess.Actor, messageText string) intake.InboundRequest {
	t.Helper()

	request, err := intakeService.CreateDraft(ctx, intake.CreateDraftInput{
		OriginType: intake.OriginHuman,
		Channel:    "browser",
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("create draft request: %v", err)
	}

	if _, err := intakeService.AddMessage(ctx, intake.AddMessageInput{
		RequestID:   request.ID,
		MessageRole: intake.MessageRoleRequest,
		TextContent: messageText,
		Actor:       actor,
	}); err != nil {
		t.Fatalf("add request message: %v", err)
	}

	request, err = intakeService.QueueRequest(ctx, intake.QueueRequestInput{
		RequestID: request.ID,
		Actor:     actor,
	})
	if err != nil {
		t.Fatalf("queue request: %v", err)
	}

	return request
}
