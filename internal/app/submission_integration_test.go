package app_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
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
	if updated.Request.ID != draft.Request.ID || updated.Request.RequestReference != draft.Request.RequestReference || updated.Request.Status != intake.StatusDraft {
		t.Fatalf("unexpected updated draft request: %+v", updated.Request)
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

func TestSubmissionServiceSaveInboundDraftPersistsUpdatedMetadataIntegration(t *testing.T) {
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
		t.Fatalf("save initial draft: %v", err)
	}

	updated, err := service.SaveInboundDraft(ctx, app.SaveInboundDraftInput{
		RequestID:   draft.Request.ID,
		MessageID:   draft.Message.ID,
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		Metadata:    map[string]any{"submitter_label": "dispatch desk"},
		MessageRole: intake.MessageRoleRequest,
		MessageText: "Updated draft details",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("update draft metadata: %v", err)
	}

	var metadataJSON []byte
	if err := db.QueryRowContext(ctx, `SELECT metadata FROM ai.inbound_requests WHERE id = $1`, draft.Request.ID).Scan(&metadataJSON); err != nil {
		t.Fatalf("load draft metadata: %v", err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		t.Fatalf("decode draft metadata: %v", err)
	}
	if metadata["submitter_label"] != "dispatch desk" {
		t.Fatalf("unexpected persisted metadata: %#v", metadata)
	}
	if !strings.Contains(string(updated.Request.Metadata), "dispatch desk") {
		t.Fatalf("expected updated metadata in response: %s", string(updated.Request.Metadata))
	}
}

func TestSubmissionServiceSaveInboundDraftRejectsMismatchedRequestMessageIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator)
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	service := app.NewSubmissionService(db)
	firstDraft, err := service.SaveInboundDraft(ctx, app.SaveInboundDraftInput{
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		Metadata:    map[string]any{"submitter_label": "first"},
		MessageRole: intake.MessageRoleRequest,
		MessageText: "First draft details",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("save first draft: %v", err)
	}
	secondDraft, err := service.SaveInboundDraft(ctx, app.SaveInboundDraftInput{
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		Metadata:    map[string]any{"submitter_label": "second"},
		MessageRole: intake.MessageRoleRequest,
		MessageText: "Second draft details",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("save second draft: %v", err)
	}

	_, err = service.SaveInboundDraft(ctx, app.SaveInboundDraftInput{
		RequestID:   firstDraft.Request.ID,
		MessageID:   secondDraft.Message.ID,
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		Metadata:    map[string]any{"submitter_label": "bad update"},
		MessageRole: intake.MessageRoleRequest,
		MessageText: "Should fail",
		Actor:       operator,
	})
	if !errors.Is(err, intake.ErrInvalidInboundRequest) {
		t.Fatalf("expected invalid inbound request, got %v", err)
	}
}

func TestSubmissionServiceSaveInboundDraftRollsBackOnAttachmentDecodeFailureIntegration(t *testing.T) {
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
		t.Fatalf("save initial draft: %v", err)
	}

	_, err = service.SaveInboundDraft(ctx, app.SaveInboundDraftInput{
		RequestID:   draft.Request.ID,
		MessageID:   draft.Message.ID,
		OriginType:  intake.OriginHuman,
		Channel:     "browser",
		Metadata:    map[string]any{"submitter_label": "dispatch desk"},
		MessageRole: intake.MessageRoleRequest,
		MessageText: "Attempted update",
		Attachments: []app.SubmitInboundRequestAttachmentInput{
			{
				OriginalFileName: "broken.txt",
				MediaType:        "text/plain",
				ContentBase64:    "%%%not-base64%%%",
				LinkRole:         attachments.LinkRoleEvidence,
			},
		},
		Actor: operator,
	})
	if !errors.Is(err, app.ErrAttachmentContentEncoding) {
		t.Fatalf("expected invalid attachment encoding error, got %v", err)
	}

	var (
		messageText   string
		metadataJSON  []byte
		attachmentCnt int
	)
	if err := db.QueryRowContext(ctx, `SELECT text_content FROM ai.inbound_request_messages WHERE id = $1`, draft.Message.ID).Scan(&messageText); err != nil {
		t.Fatalf("load draft message: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT metadata FROM ai.inbound_requests WHERE id = $1`, draft.Request.ID).Scan(&metadataJSON); err != nil {
		t.Fatalf("load draft request metadata: %v", err)
	}
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM attachments.request_message_links
WHERE request_message_id = $1`, draft.Message.ID).Scan(&attachmentCnt); err != nil {
		t.Fatalf("count request attachments: %v", err)
	}

	if messageText != "Initial draft details" {
		t.Fatalf("expected original message text after rollback, got %q", messageText)
	}
	if !strings.Contains(string(metadataJSON), "front desk") {
		t.Fatalf("expected original metadata after rollback, got %s", string(metadataJSON))
	}
	if attachmentCnt != 0 {
		t.Fatalf("expected no attachments after rollback, found %d", attachmentCnt)
	}
}
