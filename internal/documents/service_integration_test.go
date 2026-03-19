package documents_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
)

func TestDocumentApprovalFlowIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	_, approverUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleApprover, orgID)
	approverSession := startSession(t, ctx, db, orgID, approverUserID)
	approver := identityaccess.Actor{OrgID: orgID, UserID: approverUserID, SessionID: approverSession.ID}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Invoice draft for approval",
		Actor:    operator,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	if doc.Status != documents.StatusDraft {
		t.Fatalf("unexpected draft status: %s", doc.Status)
	}

	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit document: %v", err)
	}
	if doc.Status != documents.StatusSubmitted {
		t.Fatalf("unexpected submitted status: %s", doc.Status)
	}

	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance-review",
		Reason:     "invoice requires human approval",
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}
	if approval.Status != "pending" {
		t.Fatalf("unexpected approval status: %s", approval.Status)
	}

	approval, doc, err = workflowService.DecideApproval(ctx, workflow.DecideApprovalInput{
		ApprovalID:   approval.ID,
		Decision:     "approved",
		DecisionNote: "looks correct",
		Actor:        approver,
	})
	if err != nil {
		t.Fatalf("decide approval: %v", err)
	}
	if approval.Status != "approved" {
		t.Fatalf("unexpected final approval status: %s", approval.Status)
	}
	if doc.Status != documents.StatusApproved {
		t.Fatalf("unexpected final document status: %s", doc.Status)
	}

	var auditCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM platform.audit_events WHERE org_id = $1`, orgID).Scan(&auditCount); err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	if auditCount != 5 {
		t.Fatalf("unexpected audit event count: got %d want 5", auditCount)
	}

	var queueStatus string
	if err := db.QueryRowContext(ctx, `SELECT status FROM workflow.approval_queue_entries WHERE approval_id = $1`, approval.ID).Scan(&queueStatus); err != nil {
		t.Fatalf("load queue status: %v", err)
	}
	if queueStatus != "closed" {
		t.Fatalf("unexpected queue status: %s", queueStatus)
	}
}

func TestRequestApprovalRequiresSubmittedDocument(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, userID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	session := startSession(t, ctx, db, orgID, userID)
	actor := identityaccess.Actor{OrgID: orgID, UserID: userID, SessionID: session.ID}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Draft only",
		Actor:    actor,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	_, err = workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance-review",
		Actor:      actor,
	})
	if err != documents.ErrInvalidDocumentState {
		t.Fatalf("unexpected error: got %v want %v", err, documents.ErrInvalidDocumentState)
	}
}

func TestCreateDraftRequiresActiveSession(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, userID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	session := startSession(t, ctx, db, orgID, userID)

	identityService := identityaccess.NewService(db)
	if err := identityService.RevokeSession(ctx, identityaccess.Actor{
		OrgID:     orgID,
		UserID:    userID,
		SessionID: session.ID,
	}, session.ID); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	documentService := documents.NewService(db)
	_, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Should fail",
		Actor: identityaccess.Actor{
			OrgID:     orgID,
			UserID:    userID,
			SessionID: session.ID,
		},
	})
	if err != identityaccess.ErrUnauthorized {
		t.Fatalf("unexpected error: got %v want %v", err, identityaccess.ErrUnauthorized)
	}
}

func seedOrgAndUser(t *testing.T, ctx context.Context, db *sql.DB, roleCode, existingOrgID string) (string, string) {
	t.Helper()

	orgID := existingOrgID
	if orgID == "" {
		if err := db.QueryRowContext(
			ctx,
			`INSERT INTO identityaccess.orgs (slug, name) VALUES ($1, $2) RETURNING id`,
			uniqueSlug("acme"),
			"Acme",
		).Scan(&orgID); err != nil {
			t.Fatalf("insert org: %v", err)
		}
	}

	var userID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name) VALUES ($1, 'Example User') RETURNING id`,
		uniqueEmail(),
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
		RefreshTokenHash: uniqueTokenHash(),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	return session
}

func uniqueSlug(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("150405.000000000")
}

func uniqueEmail() string {
	return "user-" + time.Now().UTC().Format("150405.000000000") + "@example.com"
}

func uniqueTokenHash() string {
	return "token-" + time.Now().UTC().Format("150405.000000000")
}
