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
