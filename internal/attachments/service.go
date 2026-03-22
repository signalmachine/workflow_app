package attachments

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
)

var (
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrInvalidAttachment  = errors.New("invalid attachment")
	ErrInvalidLink        = errors.New("invalid attachment link")
	ErrInvalidDerivative  = errors.New("invalid attachment derivative")
)

const (
	StorageBackendPostgres  = "postgres"
	LinkRoleSource          = "source"
	LinkRoleEvidence        = "evidence"
	DerivativeTranscription = "transcription"
	MaxAttachmentBytes      = 10 << 20
)

type Attachment struct {
	ID               string
	OrgID            string
	StorageBackend   string
	StorageLocator   string
	OriginalFileName string
	MediaType        string
	SizeBytes        int64
	ChecksumSHA256   string
	UploadedByUserID sql.NullString
	CreatedAt        time.Time
}

type RequestMessageLink struct {
	ID               string
	OrgID            string
	RequestMessageID string
	AttachmentID     string
	LinkRole         string
	CreatedAt        time.Time
}

type DerivedText struct {
	ID                 string
	OrgID              string
	SourceAttachmentID string
	RequestMessageID   sql.NullString
	CreatedByRunID     sql.NullString
	DerivativeType     string
	ContentText        string
	CreatedAt          time.Time
}

type AttachmentContent struct {
	Attachment
	Content []byte
}

type CreateAttachmentInput struct {
	OriginalFileName string
	MediaType        string
	Content          []byte
	Actor            identityaccess.Actor
}

type LinkRequestMessageInput struct {
	RequestMessageID string
	AttachmentID     string
	LinkRole         string
	Actor            identityaccess.Actor
}

type RecordDerivedTextInput struct {
	SourceAttachmentID string
	RequestMessageID   string
	CreatedByRunID     string
	DerivativeType     string
	ContentText        string
	Actor              identityaccess.Actor
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateAttachment(ctx context.Context, input CreateAttachmentInput) (Attachment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Attachment{}, fmt.Errorf("begin create attachment: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Attachment{}, err
	}

	attachment, err := createAttachmentTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Attachment{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "attachments.attachment_created",
		EntityType:  "attachments.attachment",
		EntityID:    attachment.ID,
		Payload: map[string]any{
			"media_type":         attachment.MediaType,
			"original_file_name": attachment.OriginalFileName,
			"size_bytes":         attachment.SizeBytes,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Attachment{}, err
	}

	if err := tx.Commit(); err != nil {
		return Attachment{}, fmt.Errorf("commit create attachment: %w", err)
	}

	return attachment, nil
}

func (s *Service) LinkRequestMessage(ctx context.Context, input LinkRequestMessageInput) (RequestMessageLink, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RequestMessageLink{}, fmt.Errorf("begin link request message attachment: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return RequestMessageLink{}, err
	}

	link, err := linkRequestMessageTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return RequestMessageLink{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "attachments.request_message_linked",
		EntityType:  "attachments.request_message_link",
		EntityID:    link.ID,
		Payload: map[string]any{
			"request_message_id": link.RequestMessageID,
			"attachment_id":      link.AttachmentID,
			"link_role":          link.LinkRole,
		},
	}); err != nil {
		_ = tx.Rollback()
		return RequestMessageLink{}, err
	}

	if err := tx.Commit(); err != nil {
		return RequestMessageLink{}, fmt.Errorf("commit link request message attachment: %w", err)
	}

	return link, nil
}

func (s *Service) RecordDerivedText(ctx context.Context, input RecordDerivedTextInput) (DerivedText, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DerivedText{}, fmt.Errorf("begin record derived text: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return DerivedText{}, err
	}

	derived, err := recordDerivedTextTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return DerivedText{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "attachments.derived_text_recorded",
		EntityType:  "attachments.derived_text",
		EntityID:    derived.ID,
		Payload: map[string]any{
			"source_attachment_id": derived.SourceAttachmentID,
			"derivative_type":      derived.DerivativeType,
		},
	}); err != nil {
		_ = tx.Rollback()
		return DerivedText{}, err
	}

	if err := tx.Commit(); err != nil {
		return DerivedText{}, fmt.Errorf("commit record derived text: %w", err)
	}

	return derived, nil
}

func (s *Service) GetAttachmentContent(ctx context.Context, attachmentID string, actor identityaccess.Actor) (AttachmentContent, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AttachmentContent{}, fmt.Errorf("begin get attachment content: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return AttachmentContent{}, err
	}

	attachment, err := getAttachmentContentTx(ctx, tx, actor.OrgID, attachmentID)
	if err != nil {
		_ = tx.Rollback()
		return AttachmentContent{}, err
	}

	if err := tx.Commit(); err != nil {
		return AttachmentContent{}, fmt.Errorf("commit get attachment content: %w", err)
	}

	return attachment, nil
}

func createAttachmentTx(ctx context.Context, tx *sql.Tx, input CreateAttachmentInput) (Attachment, error) {
	name := strings.TrimSpace(input.OriginalFileName)
	mediaType := strings.TrimSpace(input.MediaType)
	if name == "" || mediaType == "" || len(input.Content) > MaxAttachmentBytes {
		return Attachment{}, ErrInvalidAttachment
	}

	checksum := sha256.Sum256(input.Content)
	const statement = `
INSERT INTO attachments.attachments (
	org_id,
	storage_backend,
	storage_locator,
	original_file_name,
	media_type,
	size_bytes,
	checksum_sha256,
	content,
	uploaded_by_user_id
) VALUES ($1, $2, gen_random_uuid()::text, $3, $4, $5, $6, $7, $8)
RETURNING
	id,
	org_id,
	storage_backend,
	storage_locator,
	original_file_name,
	media_type,
	size_bytes,
	checksum_sha256,
	uploaded_by_user_id,
	created_at;`

	attachment, err := scanAttachment(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		StorageBackendPostgres,
		name,
		mediaType,
		len(input.Content),
		hex.EncodeToString(checksum[:]),
		input.Content,
		nullIfEmpty(input.Actor.UserID),
	))
	if err != nil {
		return Attachment{}, fmt.Errorf("insert attachment: %w", err)
	}
	return attachment, nil
}

func linkRequestMessageTx(ctx context.Context, tx *sql.Tx, input LinkRequestMessageInput) (RequestMessageLink, error) {
	if strings.TrimSpace(input.RequestMessageID) == "" || strings.TrimSpace(input.AttachmentID) == "" {
		return RequestMessageLink{}, ErrInvalidLink
	}
	linkRole := strings.TrimSpace(input.LinkRole)
	if linkRole == "" {
		linkRole = LinkRoleSource
	}
	if linkRole != LinkRoleSource && linkRole != LinkRoleEvidence {
		return RequestMessageLink{}, ErrInvalidLink
	}

	const statement = `
INSERT INTO attachments.request_message_links (
	org_id,
	request_message_id,
	attachment_id,
	link_role
) VALUES ($1, $2, $3, $4)
RETURNING
	id,
	org_id,
	request_message_id,
	attachment_id,
	link_role,
	created_at;`

	link, err := scanRequestMessageLink(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		input.RequestMessageID,
		input.AttachmentID,
		linkRole,
	))
	if err != nil {
		if isForeignKeyViolation(err) {
			return RequestMessageLink{}, ErrInvalidLink
		}
		return RequestMessageLink{}, fmt.Errorf("insert request message attachment link: %w", err)
	}
	return link, nil
}

func recordDerivedTextTx(ctx context.Context, tx *sql.Tx, input RecordDerivedTextInput) (DerivedText, error) {
	if strings.TrimSpace(input.SourceAttachmentID) == "" || strings.TrimSpace(input.ContentText) == "" {
		return DerivedText{}, ErrInvalidDerivative
	}
	derivativeType := strings.TrimSpace(input.DerivativeType)
	if derivativeType == "" {
		derivativeType = DerivativeTranscription
	}
	if derivativeType != DerivativeTranscription {
		return DerivedText{}, ErrInvalidDerivative
	}

	const statement = `
INSERT INTO attachments.derived_texts (
	org_id,
	source_attachment_id,
	request_message_id,
	created_by_run_id,
	derivative_type,
	content_text
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
	id,
	org_id,
	source_attachment_id,
	request_message_id,
	created_by_run_id,
	derivative_type,
	content_text,
	created_at;`

	derived, err := scanDerivedText(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		input.SourceAttachmentID,
		nullIfEmpty(input.RequestMessageID),
		nullIfEmpty(input.CreatedByRunID),
		derivativeType,
		strings.TrimSpace(input.ContentText),
	))
	if err != nil {
		if isForeignKeyViolation(err) {
			return DerivedText{}, ErrInvalidDerivative
		}
		return DerivedText{}, fmt.Errorf("insert derived text: %w", err)
	}
	return derived, nil
}

func getAttachmentContentTx(ctx context.Context, tx *sql.Tx, orgID, attachmentID string) (AttachmentContent, error) {
	if strings.TrimSpace(attachmentID) == "" {
		return AttachmentContent{}, ErrAttachmentNotFound
	}

	const statement = `
SELECT
	id,
	org_id,
	storage_backend,
	storage_locator,
	original_file_name,
	media_type,
	size_bytes,
	checksum_sha256,
	uploaded_by_user_id,
	created_at,
	content
FROM attachments.attachments
WHERE org_id = $1
  AND id = $2;`

	var attachment AttachmentContent
	err := tx.QueryRowContext(ctx, statement, orgID, strings.TrimSpace(attachmentID)).Scan(
		&attachment.ID,
		&attachment.OrgID,
		&attachment.StorageBackend,
		&attachment.StorageLocator,
		&attachment.OriginalFileName,
		&attachment.MediaType,
		&attachment.SizeBytes,
		&attachment.ChecksumSHA256,
		&attachment.UploadedByUserID,
		&attachment.CreatedAt,
		&attachment.Content,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AttachmentContent{}, ErrAttachmentNotFound
		}
		return AttachmentContent{}, fmt.Errorf("select attachment content: %w", err)
	}

	return attachment, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAttachment(row rowScanner) (Attachment, error) {
	var attachment Attachment
	err := row.Scan(
		&attachment.ID,
		&attachment.OrgID,
		&attachment.StorageBackend,
		&attachment.StorageLocator,
		&attachment.OriginalFileName,
		&attachment.MediaType,
		&attachment.SizeBytes,
		&attachment.ChecksumSHA256,
		&attachment.UploadedByUserID,
		&attachment.CreatedAt,
	)
	if err != nil {
		return Attachment{}, err
	}
	return attachment, nil
}

func scanRequestMessageLink(row rowScanner) (RequestMessageLink, error) {
	var link RequestMessageLink
	err := row.Scan(
		&link.ID,
		&link.OrgID,
		&link.RequestMessageID,
		&link.AttachmentID,
		&link.LinkRole,
		&link.CreatedAt,
	)
	if err != nil {
		return RequestMessageLink{}, err
	}
	return link, nil
}

func scanDerivedText(row rowScanner) (DerivedText, error) {
	var derived DerivedText
	err := row.Scan(
		&derived.ID,
		&derived.OrgID,
		&derived.SourceAttachmentID,
		&derived.RequestMessageID,
		&derived.CreatedByRunID,
		&derived.DerivativeType,
		&derived.ContentText,
		&derived.CreatedAt,
	)
	if err != nil {
		return DerivedText{}, err
	}
	return derived, nil
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func isForeignKeyViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "foreign key")
}
