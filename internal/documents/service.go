package documents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
)

var (
	ErrDocumentNotFound        = errors.New("document not found")
	ErrInvalidDocumentType     = errors.New("invalid document type")
	ErrInvalidDocumentState    = errors.New("invalid document state")
	ErrDocumentAlreadyNumbered = errors.New("document already numbered")
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusSubmitted Status = "submitted"
	StatusApproved  Status = "approved"
	StatusRejected  Status = "rejected"
	StatusPosted    Status = "posted"
	StatusReversed  Status = "reversed"
	StatusVoided    Status = "voided"
)

type Document struct {
	ID                string
	OrgID             string
	TypeCode          string
	Status            Status
	Title             string
	NumberSeriesID    sql.NullString
	NumberValue       sql.NullString
	SourceDocumentID  sql.NullString
	CreatedByUserID   string
	SubmittedByUserID sql.NullString
	SubmittedAt       sql.NullTime
	ApprovedAt        sql.NullTime
	RejectedAt        sql.NullTime
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CreateDraftInput struct {
	TypeCode         string
	Title            string
	SourceDocumentID string
	Actor            identityaccess.Actor
}

type SubmitInput struct {
	DocumentID string
	Actor      identityaccess.Actor
}

type ApprovalOutcomeInput struct {
	DocumentID string
	Decision   string
	Actor      identityaccess.Actor
}

type PostingOutcomeInput struct {
	DocumentID string
	Action     string
	Actor      identityaccess.Actor
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateDraft(ctx context.Context, input CreateDraftInput) (Document, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Document{}, fmt.Errorf("begin create draft: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Document{}, err
	}

	doc, err := createDraftTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Document{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "documents.document_created",
		EntityType:  "documents.document",
		EntityID:    doc.ID,
		Payload: map[string]any{
			"type_code": input.TypeCode,
			"status":    doc.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Document{}, err
	}

	if err := tx.Commit(); err != nil {
		return Document{}, fmt.Errorf("commit create draft: %w", err)
	}

	return doc, nil
}

func (s *Service) Submit(ctx context.Context, input SubmitInput) (Document, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Document{}, fmt.Errorf("begin submit document: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Document{}, err
	}

	doc, err := submitTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Document{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "documents.document_submitted",
		EntityType:  "documents.document",
		EntityID:    doc.ID,
		Payload: map[string]any{
			"status": doc.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Document{}, err
	}

	if err := tx.Commit(); err != nil {
		return Document{}, fmt.Errorf("commit submit document: %w", err)
	}

	return doc, nil
}

func (s *Service) ApplyApprovalOutcome(ctx context.Context, tx *sql.Tx, input ApprovalOutcomeInput) (Document, error) {
	return applyApprovalOutcomeTx(ctx, tx, input)
}

func (s *Service) ApplyPostingOutcome(ctx context.Context, tx *sql.Tx, input PostingOutcomeInput) (Document, error) {
	return applyPostingOutcomeTx(ctx, tx, input)
}

func createDraftTx(ctx context.Context, tx *sql.Tx, input CreateDraftInput) (Document, error) {
	if !isSupportedType(input.TypeCode) {
		return Document{}, ErrInvalidDocumentType
	}

	const statement = `
INSERT INTO documents.documents (
	org_id,
	type_code,
	status,
	title,
	source_document_id,
	created_by_user_id
) VALUES ($1, $2, 'draft', $3, $4, $5)
RETURNING
	id,
	org_id,
	type_code,
	status,
	title,
	number_series_id,
	number_value,
	source_document_id,
	created_by_user_id,
	submitted_by_user_id,
	submitted_at,
	approved_at,
	rejected_at,
	created_at,
	updated_at;`

	row := tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		input.TypeCode,
		input.Title,
		nullIfEmpty(input.SourceDocumentID),
		input.Actor.UserID,
	)

	return scanDocument(row)
}

func submitTx(ctx context.Context, tx *sql.Tx, input SubmitInput) (Document, error) {
	doc, err := getDocumentForUpdate(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil {
		return Document{}, err
	}
	if doc.Status != StatusDraft {
		return Document{}, ErrInvalidDocumentState
	}

	const statement = `
UPDATE documents.documents
SET status = 'submitted',
	submitted_by_user_id = $3,
	submitted_at = NOW(),
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	type_code,
	status,
	title,
	number_series_id,
	number_value,
	source_document_id,
	created_by_user_id,
	submitted_by_user_id,
	submitted_at,
	approved_at,
	rejected_at,
	created_at,
	updated_at;`

	return scanDocument(tx.QueryRowContext(ctx, statement, input.Actor.OrgID, input.DocumentID, input.Actor.UserID))
}

func applyApprovalOutcomeTx(ctx context.Context, tx *sql.Tx, input ApprovalOutcomeInput) (Document, error) {
	doc, err := getDocumentForUpdate(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil {
		return Document{}, err
	}
	if doc.Status != StatusSubmitted {
		return Document{}, ErrInvalidDocumentState
	}

	var statement string
	switch input.Decision {
	case "approved":
		statement = `
UPDATE documents.documents
SET status = 'approved',
	approved_at = NOW(),
	rejected_at = NULL,
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	type_code,
	status,
	title,
	number_series_id,
	number_value,
	source_document_id,
	created_by_user_id,
	submitted_by_user_id,
	submitted_at,
	approved_at,
	rejected_at,
	created_at,
	updated_at;`
	case "rejected":
		statement = `
UPDATE documents.documents
SET status = 'rejected',
	rejected_at = NOW(),
	approved_at = NULL,
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	type_code,
	status,
	title,
	number_series_id,
	number_value,
	source_document_id,
	created_by_user_id,
	submitted_by_user_id,
	submitted_at,
	approved_at,
	rejected_at,
	created_at,
	updated_at;`
	default:
		return Document{}, ErrInvalidDocumentState
	}

	return scanDocument(tx.QueryRowContext(ctx, statement, input.Actor.OrgID, input.DocumentID))
}

func applyPostingOutcomeTx(ctx context.Context, tx *sql.Tx, input PostingOutcomeInput) (Document, error) {
	doc, err := getDocumentForUpdate(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil {
		return Document{}, err
	}

	var statement string
	switch input.Action {
	case "posted":
		if doc.Status != StatusApproved {
			return Document{}, ErrInvalidDocumentState
		}
		statement = `
UPDATE documents.documents
SET status = 'posted',
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	type_code,
	status,
	title,
	number_series_id,
	number_value,
	source_document_id,
	created_by_user_id,
	submitted_by_user_id,
	submitted_at,
	approved_at,
	rejected_at,
	created_at,
	updated_at;`
	case "reversed":
		if doc.Status != StatusPosted {
			return Document{}, ErrInvalidDocumentState
		}
		statement = `
UPDATE documents.documents
SET status = 'reversed',
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	type_code,
	status,
	title,
	number_series_id,
	number_value,
	source_document_id,
	created_by_user_id,
	submitted_by_user_id,
	submitted_at,
	approved_at,
	rejected_at,
	created_at,
	updated_at;`
	default:
		return Document{}, ErrInvalidDocumentState
	}

	return scanDocument(tx.QueryRowContext(ctx, statement, input.Actor.OrgID, input.DocumentID))
}

func getDocumentForUpdate(ctx context.Context, tx *sql.Tx, orgID, documentID string) (Document, error) {
	const query = `
SELECT
	id,
	org_id,
	type_code,
	status,
	title,
	number_series_id,
	number_value,
	source_document_id,
	created_by_user_id,
	submitted_by_user_id,
	submitted_at,
	approved_at,
	rejected_at,
	created_at,
	updated_at
FROM documents.documents
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	doc, err := scanDocument(tx.QueryRowContext(ctx, query, orgID, documentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Document{}, ErrDocumentNotFound
		}
		return Document{}, err
	}

	return doc, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDocument(row rowScanner) (Document, error) {
	var doc Document
	err := row.Scan(
		&doc.ID,
		&doc.OrgID,
		&doc.TypeCode,
		&doc.Status,
		&doc.Title,
		&doc.NumberSeriesID,
		&doc.NumberValue,
		&doc.SourceDocumentID,
		&doc.CreatedByUserID,
		&doc.SubmittedByUserID,
		&doc.SubmittedAt,
		&doc.ApprovedAt,
		&doc.RejectedAt,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return Document{}, err
	}
	return doc, nil
}

func isSupportedType(typeCode string) bool {
	switch typeCode {
	case "work_order", "invoice", "payment_receipt", "inventory_receipt", "inventory_issue", "inventory_adjustment", "journal", "ai_draft":
		return true
	default:
		return false
	}
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
