package app_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"workflow_app/internal/app"
	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/testsupport/dbtest"
)

func TestSubmissionServiceSubmitInboundRequestIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	service := app.NewSubmissionService(db)
	result, err := service.SubmitInboundRequest(ctx, app.SubmitInboundRequestInput{
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		Metadata:    map[string]any{"submitter_label": "front desk"},
		MessageRole: intake.MessageRoleRequest,
		MessageText: "The warehouse pump has failed and needs review.",
		Attachments: []app.SubmitInboundRequestAttachmentInput{
			{
				OriginalFileName: "pump-note.txt",
				MediaType:        "text/plain",
				ContentBase64:    base64.StdEncoding.EncodeToString([]byte("urgent pump failure details")),
				LinkRole:         attachments.LinkRoleEvidence,
			},
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("submit inbound request: %v", err)
	}

	if result.Request.Status != intake.StatusQueued {
		t.Fatalf("unexpected request status: %s", result.Request.Status)
	}
	if result.Request.RequestReference == "" {
		t.Fatal("expected request reference")
	}
	if result.Message.ID == "" {
		t.Fatal("expected message")
	}
	if len(result.Attachments) != 1 {
		t.Fatalf("unexpected attachment count: %d", len(result.Attachments))
	}

	downloaded, err := service.DownloadAttachment(ctx, app.DownloadAttachmentInput{
		AttachmentID: result.Attachments[0].ID,
		Actor:        operator,
	})
	if err != nil {
		t.Fatalf("download attachment: %v", err)
	}
	if string(downloaded.Content) != "urgent pump failure details" {
		t.Fatalf("unexpected attachment content: %q", string(downloaded.Content))
	}
}

func TestSubmissionServiceSubmitInboundRequestCleansUpDraftOnInvalidAttachmentEncoding(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	service := app.NewSubmissionService(db)
	_, err := service.SubmitInboundRequest(ctx, app.SubmitInboundRequestInput{
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		MessageRole: intake.MessageRoleRequest,
		MessageText: "Attachment upload should fail.",
		Attachments: []app.SubmitInboundRequestAttachmentInput{
			{
				OriginalFileName: "broken.txt",
				MediaType:        "text/plain",
				ContentBase64:    "%%%not-base64%%%",
			},
		},
		Actor: operator,
	})
	if !errors.Is(err, app.ErrAttachmentContentEncoding) {
		t.Fatalf("expected invalid attachment encoding error, got %v", err)
	}

	var requestCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.inbound_requests WHERE org_id = $1`, orgID).Scan(&requestCount); err != nil {
		t.Fatalf("count inbound requests: %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("expected draft cleanup after failed submission, found %d requests", requestCount)
	}
}

func TestSubmissionServiceDraftLifecycleIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	service := app.NewSubmissionService(db)
	draft, err := service.SaveInboundDraft(ctx, app.SaveInboundDraftInput{
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		Metadata:    map[string]any{"submitter_label": "front desk"},
		MessageRole: intake.MessageRoleRequest,
		MessageText: "Initial draft details",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("save draft: %v", err)
	}
	if draft.Request.ID == "" || draft.Request.Status != intake.StatusDraft {
		t.Fatalf("unexpected draft request: %+v", draft.Request)
	}
	if draft.Message.ID == "" {
		t.Fatal("expected draft message")
	}

	updated, err := service.SaveInboundDraft(ctx, app.SaveInboundDraftInput{
		RequestID:   draft.Request.ID,
		MessageID:   draft.Message.ID,
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		MessageRole: intake.MessageRoleRequest,
		MessageText: "Updated draft details",
		Attachments: []app.SubmitInboundRequestAttachmentInput{
			{
				OriginalFileName: "draft-note.txt",
				MediaType:        "text/plain",
				ContentBase64:    base64.StdEncoding.EncodeToString([]byte("draft attachment text")),
				LinkRole:         attachments.LinkRoleEvidence,
			},
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("update draft: %v", err)
	}
	if updated.Message.TextContent != "Updated draft details" {
		t.Fatalf("unexpected updated message text: %q", updated.Message.TextContent)
	}
	if len(updated.Attachments) != 1 {
		t.Fatalf("unexpected added attachment count: %d", len(updated.Attachments))
	}

	queued, err := service.QueueInboundRequest(ctx, app.QueueInboundRequestInput{
		RequestID: draft.Request.ID,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("queue draft: %v", err)
	}
	if queued.Status != intake.StatusQueued {
		t.Fatalf("unexpected queued status: %s", queued.Status)
	}

	cancelled, err := service.CancelInboundRequest(ctx, app.CancelInboundRequestInput{
		RequestID: draft.Request.ID,
		Reason:    "operator paused request",
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("cancel queued request: %v", err)
	}
	if cancelled.Status != intake.StatusCancelled {
		t.Fatalf("unexpected cancelled status: %s", cancelled.Status)
	}

	amended, err := service.AmendInboundRequest(ctx, app.AmendInboundRequestInput{
		RequestID: draft.Request.ID,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("amend request: %v", err)
	}
	if amended.Status != intake.StatusDraft {
		t.Fatalf("unexpected amended status: %s", amended.Status)
	}

	if err := service.DeleteInboundDraft(ctx, app.DeleteInboundDraftInput{
		RequestID: draft.Request.ID,
		Actor:     operator,
	}); err != nil {
		t.Fatalf("delete draft: %v", err)
	}

	var remaining int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.inbound_requests WHERE id = $1`, draft.Request.ID).Scan(&remaining); err != nil {
		t.Fatalf("count remaining requests: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected deleted draft to be removed, found %d", remaining)
	}
}
