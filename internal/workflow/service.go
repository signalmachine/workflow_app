package workflow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
)

var (
	ErrApprovalNotFound      = errors.New("approval not found")
	ErrApprovalState         = errors.New("invalid approval state")
	ErrApprovalQueueRequired = errors.New("approval queue is required")
)

type Approval struct {
	ID                string
	OrgID             string
	DocumentID        string
	Status            string
	QueueCode         string
	RequestedByUserID string
	DecidedByUserID   sql.NullString
	DecisionNote      sql.NullString
	RequestedAt       time.Time
	DecidedAt         sql.NullTime
}

type RequestApprovalInput struct {
	DocumentID string
	QueueCode  string
	Reason     string
	Actor      identityaccess.Actor
}

type DecideApprovalInput struct {
	ApprovalID   string
	Decision     string
	DecisionNote string
	Actor        identityaccess.Actor
}

type Service struct {
	db        *sql.DB
	documents *documents.Service
}

func NewService(db *sql.DB, documentService *documents.Service) *Service {
	return &Service{db: db, documents: documentService}
}

func (s *Service) RequestApproval(ctx context.Context, input RequestApprovalInput) (Approval, error) {
	if input.QueueCode == "" {
		return Approval{}, ErrApprovalQueueRequired
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Approval{}, fmt.Errorf("begin request approval: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Approval{}, err
	}

	documentState, err := loadDocumentState(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return Approval{}, err
	}
	if documentState != string(documents.StatusSubmitted) {
		_ = tx.Rollback()
		return Approval{}, documents.ErrInvalidDocumentState
	}

	approval, err := insertApprovalTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Approval{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "workflow.approval_requested",
		EntityType:  "workflow.approval",
		EntityID:    approval.ID,
		Payload: map[string]any{
			"document_id": input.DocumentID,
			"queue_code":  input.QueueCode,
			"reason":      input.Reason,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Approval{}, err
	}

	if err := tx.Commit(); err != nil {
		return Approval{}, fmt.Errorf("commit request approval: %w", err)
	}

	return approval, nil
}

func (s *Service) DecideApproval(ctx context.Context, input DecideApprovalInput) (Approval, documents.Document, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Approval{}, documents.Document{}, fmt.Errorf("begin decide approval: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleApprover); err != nil {
		_ = tx.Rollback()
		return Approval{}, documents.Document{}, err
	}

	approval, err := getApprovalForUpdate(ctx, tx, input.Actor.OrgID, input.ApprovalID)
	if err != nil {
		_ = tx.Rollback()
		return Approval{}, documents.Document{}, err
	}
	if approval.Status != "pending" {
		_ = tx.Rollback()
		return Approval{}, documents.Document{}, ErrApprovalState
	}
	if input.Decision != "approved" && input.Decision != "rejected" {
		_ = tx.Rollback()
		return Approval{}, documents.Document{}, ErrApprovalState
	}

	approval, err = decideApprovalTx(ctx, tx, input, approval.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return Approval{}, documents.Document{}, err
	}

	document, err := s.documents.ApplyApprovalOutcome(ctx, tx, documents.ApprovalOutcomeInput{
		DocumentID: approval.DocumentID,
		Decision:   input.Decision,
		Actor:      input.Actor,
	})
	if err != nil {
		_ = tx.Rollback()
		return Approval{}, documents.Document{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "workflow.approval_decided",
		EntityType:  "workflow.approval",
		EntityID:    approval.ID,
		Payload: map[string]any{
			"decision":    input.Decision,
			"document_id": approval.DocumentID,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Approval{}, documents.Document{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "documents.document_decision_applied",
		EntityType:  "documents.document",
		EntityID:    document.ID,
		Payload: map[string]any{
			"decision": input.Decision,
			"status":   document.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Approval{}, documents.Document{}, err
	}

	if err := tx.Commit(); err != nil {
		return Approval{}, documents.Document{}, fmt.Errorf("commit decide approval: %w", err)
	}

	return approval, document, nil
}

func loadDocumentState(ctx context.Context, tx *sql.Tx, orgID, documentID string) (string, error) {
	const query = `
SELECT status
FROM documents.documents
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	var status string
	if err := tx.QueryRowContext(ctx, query, orgID, documentID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", documents.ErrDocumentNotFound
		}
		return "", fmt.Errorf("load document state: %w", err)
	}
	return status, nil
}

func insertApprovalTx(ctx context.Context, tx *sql.Tx, input RequestApprovalInput) (Approval, error) {
	const insertApproval = `
INSERT INTO workflow.approvals (
	org_id,
	document_id,
	status,
	queue_code,
	requested_by_user_id
) VALUES ($1, $2, 'pending', $3, $4)
RETURNING
	id,
	org_id,
	document_id,
	status,
	queue_code,
	requested_by_user_id,
	decided_by_user_id,
	decision_note,
	requested_at,
	decided_at;`

	approval, err := scanApproval(tx.QueryRowContext(
		ctx,
		insertApproval,
		input.Actor.OrgID,
		input.DocumentID,
		input.QueueCode,
		input.Actor.UserID,
	))
	if err != nil {
		return Approval{}, fmt.Errorf("insert approval: %w", err)
	}

	const insertQueue = `
INSERT INTO workflow.approval_queue_entries (
	approval_id,
	org_id,
	queue_code,
	status
) VALUES ($1, $2, $3, 'pending');`

	if _, err := tx.ExecContext(ctx, insertQueue, approval.ID, input.Actor.OrgID, input.QueueCode); err != nil {
		return Approval{}, fmt.Errorf("insert approval queue entry: %w", err)
	}

	return approval, nil
}

func getApprovalForUpdate(ctx context.Context, tx *sql.Tx, orgID, approvalID string) (Approval, error) {
	const query = `
SELECT
	id,
	org_id,
	document_id,
	status,
	queue_code,
	requested_by_user_id,
	decided_by_user_id,
	decision_note,
	requested_at,
	decided_at
FROM workflow.approvals
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	approval, err := scanApproval(tx.QueryRowContext(ctx, query, orgID, approvalID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Approval{}, ErrApprovalNotFound
		}
		return Approval{}, err
	}
	return approval, nil
}

func decideApprovalTx(ctx context.Context, tx *sql.Tx, input DecideApprovalInput, documentID string) (Approval, error) {
	const updateApproval = `
UPDATE workflow.approvals
SET status = $3,
	decided_by_user_id = $4,
	decision_note = $5,
	decided_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	document_id,
	status,
	queue_code,
	requested_by_user_id,
	decided_by_user_id,
	decision_note,
	requested_at,
	decided_at;`

	approval, err := scanApproval(tx.QueryRowContext(
		ctx,
		updateApproval,
		input.Actor.OrgID,
		input.ApprovalID,
		input.Decision,
		input.Actor.UserID,
		nullIfEmpty(input.DecisionNote),
	))
	if err != nil {
		return Approval{}, fmt.Errorf("update approval: %w", err)
	}

	const insertDecision = `
INSERT INTO workflow.approval_decisions (
	approval_id,
	org_id,
	document_id,
	decision,
	actor_user_id,
	note
) VALUES ($1, $2, $3, $4, $5, $6);`

	if _, err := tx.ExecContext(
		ctx,
		insertDecision,
		input.ApprovalID,
		input.Actor.OrgID,
		documentID,
		input.Decision,
		input.Actor.UserID,
		nullIfEmpty(input.DecisionNote),
	); err != nil {
		return Approval{}, fmt.Errorf("insert approval decision: %w", err)
	}

	const updateQueue = `
UPDATE workflow.approval_queue_entries
SET status = 'closed',
	closed_at = NOW()
WHERE approval_id = $1;`

	if _, err := tx.ExecContext(ctx, updateQueue, input.ApprovalID); err != nil {
		return Approval{}, fmt.Errorf("update approval queue entry: %w", err)
	}

	return approval, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanApproval(row rowScanner) (Approval, error) {
	var approval Approval
	err := row.Scan(
		&approval.ID,
		&approval.OrgID,
		&approval.DocumentID,
		&approval.Status,
		&approval.QueueCode,
		&approval.RequestedByUserID,
		&approval.DecidedByUserID,
		&approval.DecisionNote,
		&approval.RequestedAt,
		&approval.DecidedAt,
	)
	if err != nil {
		return Approval{}, err
	}
	return approval, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
