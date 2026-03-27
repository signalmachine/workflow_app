package intake

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
)

var (
	ErrInboundRequestNotFound        = errors.New("inbound request not found")
	ErrInboundRequestState           = errors.New("invalid inbound request state")
	ErrInvalidInboundRequest         = errors.New("invalid inbound request")
	ErrNoQueuedInboundRequest        = errors.New("no queued inbound request")
	ErrInboundRequestMessageNotFound = errors.New("inbound request message not found")
)

const (
	StatusDraft      = "draft"
	StatusQueued     = "queued"
	StatusProcessing = "processing"
	StatusProcessed  = "processed"
	StatusActedOn    = "acted_on"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled"

	OriginHuman  = "human"
	OriginSystem = "system"

	MessageRoleRequest       = "request"
	MessageRoleSystem        = "system"
	MessageRoleAssistant     = "assistant"
	MessageRoleTranscription = "transcription"
	requestReferencePrefix   = "REQ-"
)

type InboundRequest struct {
	ID                  string
	OrgID               string
	RequestNumber       int64
	RequestReference    string
	SessionID           sql.NullString
	ActorUserID         sql.NullString
	OriginType          string
	Channel             string
	Status              string
	Metadata            json.RawMessage
	CancellationReason  string
	FailureReason       string
	ReceivedAt          time.Time
	QueuedAt            sql.NullTime
	ProcessingStartedAt sql.NullTime
	ProcessedAt         sql.NullTime
	ActedOnAt           sql.NullTime
	CompletedAt         sql.NullTime
	FailedAt            sql.NullTime
	CancelledAt         sql.NullTime
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Message struct {
	ID              string
	OrgID           string
	RequestID       string
	MessageIndex    int
	MessageRole     string
	TextContent     string
	CreatedByUserID sql.NullString
	CreatedAt       time.Time
}

type CreateDraftInput struct {
	OriginType string
	Channel    string
	Metadata   any
	Actor      identityaccess.Actor
}

type AddMessageInput struct {
	RequestID   string
	MessageRole string
	TextContent string
	Actor       identityaccess.Actor
}

type UpdateMessageInput struct {
	MessageID   string
	TextContent string
	MessageRole string
	Actor       identityaccess.Actor
}

type QueueRequestInput struct {
	RequestID string
	Actor     identityaccess.Actor
}

type CancelRequestInput struct {
	RequestID string
	Reason    string
	Actor     identityaccess.Actor
}

type ClaimNextQueuedInput struct {
	Channel string
	Actor   identityaccess.Actor
}

type AdvanceRequestInput struct {
	RequestID     string
	Status        string
	FailureReason string
	Actor         identityaccess.Actor
}

type AmendRequestInput struct {
	RequestID string
	Actor     identityaccess.Actor
}

type DeleteDraftInput struct {
	RequestID string
	Actor     identityaccess.Actor
}

type GetRequestInput struct {
	RequestID string
	Actor     identityaccess.Actor
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateDraft(ctx context.Context, input CreateDraftInput) (InboundRequest, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return InboundRequest{}, fmt.Errorf("begin create inbound request draft: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	request, err := createDraftTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.inbound_request_created",
		EntityType:  "ai.inbound_request",
		EntityID:    request.ID,
		Payload: map[string]any{
			"origin_type": request.OriginType,
			"channel":     request.Channel,
			"status":      request.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	if err := tx.Commit(); err != nil {
		return InboundRequest{}, fmt.Errorf("commit create inbound request draft: %w", err)
	}

	return request, nil
}

func (s *Service) AddMessage(ctx context.Context, input AddMessageInput) (Message, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Message{}, fmt.Errorf("begin add inbound request message: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}

	message, err := addMessageTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.inbound_request_message_added",
		EntityType:  "ai.inbound_request_message",
		EntityID:    message.ID,
		Payload: map[string]any{
			"request_id":    message.RequestID,
			"message_role":  message.MessageRole,
			"message_index": message.MessageIndex,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}

	if err := tx.Commit(); err != nil {
		return Message{}, fmt.Errorf("commit add inbound request message: %w", err)
	}

	return message, nil
}

func (s *Service) UpdateMessage(ctx context.Context, input UpdateMessageInput) (Message, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Message{}, fmt.Errorf("begin update inbound request message: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}

	message, err := updateMessageTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.inbound_request_message_updated",
		EntityType:  "ai.inbound_request_message",
		EntityID:    message.ID,
		Payload: map[string]any{
			"request_id":   message.RequestID,
			"message_role": message.MessageRole,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}

	if err := tx.Commit(); err != nil {
		return Message{}, fmt.Errorf("commit update inbound request message: %w", err)
	}

	return message, nil
}

func (s *Service) QueueRequest(ctx context.Context, input QueueRequestInput) (InboundRequest, error) {
	return s.transitionRequest(ctx, input.Actor, input.RequestID, "ai.inbound_request_queued", func(ctx context.Context, tx *sql.Tx) (InboundRequest, error) {
		return queueRequestTx(ctx, tx, input)
	})
}

func (s *Service) CancelRequest(ctx context.Context, input CancelRequestInput) (InboundRequest, error) {
	return s.transitionRequest(ctx, input.Actor, input.RequestID, "ai.inbound_request_cancelled", func(ctx context.Context, tx *sql.Tx) (InboundRequest, error) {
		return cancelRequestTx(ctx, tx, input)
	})
}

func (s *Service) ClaimNextQueued(ctx context.Context, input ClaimNextQueuedInput) (InboundRequest, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return InboundRequest{}, fmt.Errorf("begin claim inbound request: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	request, err := claimNextQueuedTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.inbound_request_claimed",
		EntityType:  "ai.inbound_request",
		EntityID:    request.ID,
		Payload: map[string]any{
			"channel": request.Channel,
			"status":  request.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	if err := tx.Commit(); err != nil {
		return InboundRequest{}, fmt.Errorf("commit claim inbound request: %w", err)
	}

	return request, nil
}

func (s *Service) AdvanceRequest(ctx context.Context, input AdvanceRequestInput) (InboundRequest, error) {
	return s.transitionRequest(ctx, input.Actor, input.RequestID, "ai.inbound_request_status_advanced", func(ctx context.Context, tx *sql.Tx) (InboundRequest, error) {
		return advanceRequestTx(ctx, tx, input)
	})
}

func (s *Service) AmendRequest(ctx context.Context, input AmendRequestInput) (InboundRequest, error) {
	return s.transitionRequest(ctx, input.Actor, input.RequestID, "ai.inbound_request_amended", func(ctx context.Context, tx *sql.Tx) (InboundRequest, error) {
		return amendRequestTx(ctx, tx, input)
	})
}

func (s *Service) DeleteDraft(ctx context.Context, input DeleteDraftInput) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete inbound request draft: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return err
	}

	request, err := deleteDraftTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.inbound_request_deleted",
		EntityType:  "ai.inbound_request",
		EntityID:    request.ID,
		Payload: map[string]any{
			"request_reference": request.RequestReference,
			"status":            request.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete inbound request draft: %w", err)
	}

	return nil
}

func (s *Service) GetRequest(ctx context.Context, input GetRequestInput) (InboundRequest, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return InboundRequest{}, fmt.Errorf("begin load inbound request: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	request, err := getInboundRequest(ctx, tx, input.Actor.OrgID, input.RequestID)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	if err := tx.Commit(); err != nil {
		return InboundRequest{}, fmt.Errorf("commit load inbound request: %w", err)
	}

	return request, nil
}

func (s *Service) transitionRequest(
	ctx context.Context,
	actor identityaccess.Actor,
	requestID string,
	eventType string,
	fn func(context.Context, *sql.Tx) (InboundRequest, error),
) (InboundRequest, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return InboundRequest{}, fmt.Errorf("begin inbound request transition: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	request, err := fn(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       actor.OrgID,
		ActorUserID: actor.UserID,
		EventType:   eventType,
		EntityType:  "ai.inbound_request",
		EntityID:    requestID,
		Payload: map[string]any{
			"status":              request.Status,
			"cancellation_reason": request.CancellationReason,
			"failure_reason":      request.FailureReason,
		},
	}); err != nil {
		_ = tx.Rollback()
		return InboundRequest{}, err
	}

	if err := tx.Commit(); err != nil {
		return InboundRequest{}, fmt.Errorf("commit inbound request transition: %w", err)
	}

	return request, nil
}

func createDraftTx(ctx context.Context, tx *sql.Tx, input CreateDraftInput) (InboundRequest, error) {
	originType := strings.TrimSpace(input.OriginType)
	if originType == "" {
		originType = OriginHuman
	}
	if originType != OriginHuman && originType != OriginSystem {
		return InboundRequest{}, ErrInvalidInboundRequest
	}
	channel := strings.TrimSpace(input.Channel)
	if channel == "" {
		channel = "browser"
	}

	metadata, err := marshalJSON(input.Metadata)
	if err != nil {
		return InboundRequest{}, err
	}

	const statement = `
INSERT INTO ai.inbound_requests (
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
RETURNING
	id,
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata,
	cancellation_reason,
	failure_reason,
	received_at,
	queued_at,
	processing_started_at,
	processed_at,
	acted_on_at,
	completed_at,
	failed_at,
	cancelled_at,
	created_at,
	updated_at;`

	requestNumber, requestReference, err := nextRequestReference(ctx, tx, input.Actor.OrgID)
	if err != nil {
		return InboundRequest{}, err
	}

	return scanInboundRequest(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		requestNumber,
		requestReference,
		nullIfEmpty(input.Actor.SessionID),
		nullIfEmpty(input.Actor.UserID),
		originType,
		channel,
		StatusDraft,
		string(metadata),
	))
}

func addMessageTx(ctx context.Context, tx *sql.Tx, input AddMessageInput) (Message, error) {
	role := strings.TrimSpace(input.MessageRole)
	if role == "" {
		role = MessageRoleRequest
	}
	if !isValidMessageRole(role) || strings.TrimSpace(input.RequestID) == "" {
		return Message{}, ErrInvalidInboundRequest
	}

	request, err := getInboundRequestForUpdate(ctx, tx, input.Actor.OrgID, input.RequestID)
	if err != nil {
		return Message{}, err
	}
	if request.Status != StatusDraft {
		return Message{}, ErrInboundRequestState
	}

	const statement = `
WITH next_message AS (
	SELECT COALESCE(MAX(message_index), 0) + 1 AS message_index
	FROM ai.inbound_request_messages
	WHERE request_id = $2
)
INSERT INTO ai.inbound_request_messages (
	org_id,
	request_id,
	message_index,
	message_role,
	text_content,
	created_by_user_id
)
SELECT $1, $2, message_index, $3, $4, $5
FROM next_message
RETURNING
	id,
	org_id,
	request_id,
	message_index,
	message_role,
	text_content,
	created_by_user_id,
	created_at;`

	return scanMessage(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		input.RequestID,
		role,
		strings.TrimSpace(input.TextContent),
		nullIfEmpty(input.Actor.UserID),
	))
}

func updateMessageTx(ctx context.Context, tx *sql.Tx, input UpdateMessageInput) (Message, error) {
	messageID := strings.TrimSpace(input.MessageID)
	if messageID == "" {
		return Message{}, ErrInvalidInboundRequest
	}

	message, request, err := getMessageForUpdate(ctx, tx, input.Actor.OrgID, messageID)
	if err != nil {
		return Message{}, err
	}
	if request.Status != StatusDraft {
		return Message{}, ErrInboundRequestState
	}

	role := strings.TrimSpace(input.MessageRole)
	if role == "" {
		role = message.MessageRole
	}
	if !isValidMessageRole(role) {
		return Message{}, ErrInvalidInboundRequest
	}

	const statement = `
UPDATE ai.inbound_request_messages
SET message_role = $3,
	text_content = $4
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	request_id,
	message_index,
	message_role,
	text_content,
	created_by_user_id,
	created_at;`

	return scanMessage(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		messageID,
		role,
		strings.TrimSpace(input.TextContent),
	))
}

func queueRequestTx(ctx context.Context, tx *sql.Tx, input QueueRequestInput) (InboundRequest, error) {
	request, err := getInboundRequestForUpdate(ctx, tx, input.Actor.OrgID, input.RequestID)
	if err != nil {
		return InboundRequest{}, err
	}
	if request.Status != StatusDraft {
		return InboundRequest{}, ErrInboundRequestState
	}

	var messageCount int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM ai.inbound_request_messages
WHERE org_id = $1
  AND request_id = $2;`, input.Actor.OrgID, input.RequestID).Scan(&messageCount); err != nil {
		return InboundRequest{}, fmt.Errorf("count inbound request messages: %w", err)
	}
	if messageCount == 0 {
		return InboundRequest{}, ErrInvalidInboundRequest
	}

	const statement = `
UPDATE ai.inbound_requests
SET status = $3,
	queued_at = NOW(),
	updated_at = NOW(),
	cancellation_reason = '',
	failure_reason = ''
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata,
	cancellation_reason,
	failure_reason,
	received_at,
	queued_at,
	processing_started_at,
	processed_at,
	acted_on_at,
	completed_at,
	failed_at,
	cancelled_at,
	created_at,
	updated_at;`

	return scanInboundRequest(tx.QueryRowContext(ctx, statement, input.Actor.OrgID, input.RequestID, StatusQueued))
}

func cancelRequestTx(ctx context.Context, tx *sql.Tx, input CancelRequestInput) (InboundRequest, error) {
	request, err := getInboundRequestForUpdate(ctx, tx, input.Actor.OrgID, input.RequestID)
	if err != nil {
		return InboundRequest{}, err
	}
	if request.Status != StatusDraft && request.Status != StatusQueued {
		return InboundRequest{}, ErrInboundRequestState
	}

	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "cancelled by user"
	}

	const statement = `
UPDATE ai.inbound_requests
SET status = $3,
	cancellation_reason = $4,
	cancelled_at = NOW(),
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata,
	cancellation_reason,
	failure_reason,
	received_at,
	queued_at,
	processing_started_at,
	processed_at,
	acted_on_at,
	completed_at,
	failed_at,
	cancelled_at,
	created_at,
	updated_at;`

	return scanInboundRequest(tx.QueryRowContext(ctx, statement, input.Actor.OrgID, input.RequestID, StatusCancelled, reason))
}

func amendRequestTx(ctx context.Context, tx *sql.Tx, input AmendRequestInput) (InboundRequest, error) {
	request, err := getInboundRequestForUpdate(ctx, tx, input.Actor.OrgID, input.RequestID)
	if err != nil {
		return InboundRequest{}, err
	}
	if request.Status != StatusQueued && request.Status != StatusCancelled {
		return InboundRequest{}, ErrInboundRequestState
	}
	if request.ProcessingStartedAt.Valid {
		return InboundRequest{}, ErrInboundRequestState
	}

	const statement = `
UPDATE ai.inbound_requests
SET status = $3,
	queued_at = NULL,
	cancelled_at = NULL,
	cancellation_reason = '',
	failure_reason = '',
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata,
	cancellation_reason,
	failure_reason,
	received_at,
	queued_at,
	processing_started_at,
	processed_at,
	acted_on_at,
	completed_at,
	failed_at,
	cancelled_at,
	created_at,
	updated_at;`

	return scanInboundRequest(tx.QueryRowContext(ctx, statement, input.Actor.OrgID, input.RequestID, StatusDraft))
}

func deleteDraftTx(ctx context.Context, tx *sql.Tx, input DeleteDraftInput) (InboundRequest, error) {
	request, err := getInboundRequestForUpdate(ctx, tx, input.Actor.OrgID, input.RequestID)
	if err != nil {
		return InboundRequest{}, err
	}
	if request.Status != StatusDraft {
		return InboundRequest{}, ErrInboundRequestState
	}

	attachmentIDs, err := loadDraftAttachmentIDs(ctx, tx, input.Actor.OrgID, input.RequestID)
	if err != nil {
		return InboundRequest{}, err
	}

	if _, err := tx.ExecContext(ctx, `
DELETE FROM ai.inbound_requests
WHERE org_id = $1
  AND id = $2;`, input.Actor.OrgID, input.RequestID); err != nil {
		return InboundRequest{}, fmt.Errorf("delete inbound request draft: %w", err)
	}

	for _, attachmentID := range attachmentIDs {
		if _, err := tx.ExecContext(ctx, `
DELETE FROM attachments.attachments a
WHERE a.org_id = $1
  AND a.id = $2
  AND NOT EXISTS (
		SELECT 1
		FROM attachments.request_message_links rml
		WHERE rml.org_id = a.org_id
		  AND rml.attachment_id = a.id
	);`, input.Actor.OrgID, attachmentID); err != nil {
			return InboundRequest{}, fmt.Errorf("delete inbound request draft attachment %s: %w", attachmentID, err)
		}
	}

	return request, nil
}

func claimNextQueuedTx(ctx context.Context, tx *sql.Tx, input ClaimNextQueuedInput) (InboundRequest, error) {
	channel := strings.TrimSpace(input.Channel)
	args := []any{input.Actor.OrgID}
	filter := ""
	if channel != "" {
		filter = " AND channel = $2"
		args = append(args, channel)
	}

	query := `
SELECT id
FROM ai.inbound_requests
WHERE org_id = $1
  AND status = 'queued'` + filter + `
ORDER BY queued_at ASC NULLS LAST, received_at ASC, id ASC
FOR UPDATE SKIP LOCKED
LIMIT 1;`

	var requestID string
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&requestID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InboundRequest{}, ErrNoQueuedInboundRequest
		}
		return InboundRequest{}, fmt.Errorf("select queued inbound request: %w", err)
	}

	const statement = `
UPDATE ai.inbound_requests
SET status = 'processing',
	processing_started_at = NOW(),
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata,
	cancellation_reason,
	failure_reason,
	received_at,
	queued_at,
	processing_started_at,
	processed_at,
	acted_on_at,
	completed_at,
	failed_at,
	cancelled_at,
	created_at,
	updated_at;`

	return scanInboundRequest(tx.QueryRowContext(ctx, statement, input.Actor.OrgID, requestID))
}

func advanceRequestTx(ctx context.Context, tx *sql.Tx, input AdvanceRequestInput) (InboundRequest, error) {
	request, err := getInboundRequestForUpdate(ctx, tx, input.Actor.OrgID, input.RequestID)
	if err != nil {
		return InboundRequest{}, err
	}

	status := strings.TrimSpace(input.Status)
	if !isValidAdvanceStatus(status) || !isAllowedAdvance(request.Status, status) {
		return InboundRequest{}, ErrInboundRequestState
	}

	failureReason := strings.TrimSpace(input.FailureReason)
	if status == StatusFailed && failureReason == "" {
		failureReason = "processing failed"
	}

	const statement = `
UPDATE ai.inbound_requests
SET status = $3,
	failure_reason = CASE WHEN $3 = 'failed' THEN $4 ELSE '' END,
	processed_at = CASE WHEN $3 IN ('processed', 'acted_on', 'completed') AND processed_at IS NULL THEN NOW() ELSE processed_at END,
	acted_on_at = CASE WHEN $3 IN ('acted_on', 'completed') AND acted_on_at IS NULL THEN NOW() ELSE acted_on_at END,
	completed_at = CASE WHEN $3 = 'completed' THEN NOW() ELSE completed_at END,
	failed_at = CASE WHEN $3 = 'failed' THEN NOW() ELSE failed_at END,
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata,
	cancellation_reason,
	failure_reason,
	received_at,
	queued_at,
	processing_started_at,
	processed_at,
	acted_on_at,
	completed_at,
	failed_at,
	cancelled_at,
	created_at,
	updated_at;`

	return scanInboundRequest(tx.QueryRowContext(ctx, statement, input.Actor.OrgID, input.RequestID, status, failureReason))
}

func getInboundRequestForUpdate(ctx context.Context, tx *sql.Tx, orgID, requestID string) (InboundRequest, error) {
	const query = `
SELECT
	id,
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata,
	cancellation_reason,
	failure_reason,
	received_at,
	queued_at,
	processing_started_at,
	processed_at,
	acted_on_at,
	completed_at,
	failed_at,
	cancelled_at,
	created_at,
	updated_at
FROM ai.inbound_requests
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	request, err := scanInboundRequest(tx.QueryRowContext(ctx, query, orgID, requestID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InboundRequest{}, ErrInboundRequestNotFound
		}
		return InboundRequest{}, fmt.Errorf("load inbound request: %w", err)
	}
	return request, nil
}

func getInboundRequest(ctx context.Context, tx *sql.Tx, orgID, requestID string) (InboundRequest, error) {
	const query = `
SELECT
	id,
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata,
	cancellation_reason,
	failure_reason,
	received_at,
	queued_at,
	processing_started_at,
	processed_at,
	acted_on_at,
	completed_at,
	failed_at,
	cancelled_at,
	created_at,
	updated_at
FROM ai.inbound_requests
WHERE org_id = $1
  AND id = $2;`

	request, err := scanInboundRequest(tx.QueryRowContext(ctx, query, orgID, requestID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InboundRequest{}, ErrInboundRequestNotFound
		}
		return InboundRequest{}, fmt.Errorf("load inbound request: %w", err)
	}
	return request, nil
}

func getMessageForUpdate(ctx context.Context, tx *sql.Tx, orgID, messageID string) (Message, InboundRequest, error) {
	const query = `
SELECT
	m.id,
	m.org_id,
	m.request_id,
	m.message_index,
	m.message_role,
	m.text_content,
	m.created_by_user_id,
	m.created_at,
	r.id,
	r.org_id,
	r.request_number,
	r.request_reference,
	r.session_id,
	r.actor_user_id,
	r.origin_type,
	r.channel,
	r.status,
	r.metadata,
	r.cancellation_reason,
	r.failure_reason,
	r.received_at,
	r.queued_at,
	r.processing_started_at,
	r.processed_at,
	r.acted_on_at,
	r.completed_at,
	r.failed_at,
	r.cancelled_at,
	r.created_at,
	r.updated_at
FROM ai.inbound_request_messages m
JOIN ai.inbound_requests r
  ON r.org_id = m.org_id
 AND r.id = m.request_id
WHERE m.org_id = $1
  AND m.id = $2
FOR UPDATE OF m, r;`

	var (
		message  Message
		request  InboundRequest
		metadata []byte
	)
	err := tx.QueryRowContext(ctx, query, orgID, messageID).Scan(
		&message.ID,
		&message.OrgID,
		&message.RequestID,
		&message.MessageIndex,
		&message.MessageRole,
		&message.TextContent,
		&message.CreatedByUserID,
		&message.CreatedAt,
		&request.ID,
		&request.OrgID,
		&request.RequestNumber,
		&request.RequestReference,
		&request.SessionID,
		&request.ActorUserID,
		&request.OriginType,
		&request.Channel,
		&request.Status,
		&metadata,
		&request.CancellationReason,
		&request.FailureReason,
		&request.ReceivedAt,
		&request.QueuedAt,
		&request.ProcessingStartedAt,
		&request.ProcessedAt,
		&request.ActedOnAt,
		&request.CompletedAt,
		&request.FailedAt,
		&request.CancelledAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Message{}, InboundRequest{}, ErrInboundRequestMessageNotFound
		}
		return Message{}, InboundRequest{}, fmt.Errorf("load inbound request message: %w", err)
	}
	request.Metadata = append(request.Metadata[:0], metadata...)
	return message, request, nil
}

func loadDraftAttachmentIDs(ctx context.Context, tx *sql.Tx, orgID, requestID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT DISTINCT rml.attachment_id
FROM attachments.request_message_links rml
JOIN ai.inbound_request_messages m
  ON m.org_id = rml.org_id
 AND m.id = rml.request_message_id
WHERE rml.org_id = $1
  AND m.request_id = $2;`, orgID, requestID)
	if err != nil {
		return nil, fmt.Errorf("load inbound request draft attachments: %w", err)
	}
	defer rows.Close()

	var attachmentIDs []string
	for rows.Next() {
		var attachmentID string
		if err := rows.Scan(&attachmentID); err != nil {
			return nil, fmt.Errorf("scan inbound request draft attachment: %w", err)
		}
		attachmentIDs = append(attachmentIDs, attachmentID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inbound request draft attachments: %w", err)
	}

	return attachmentIDs, nil
}

func marshalJSON(value any) ([]byte, error) {
	if value == nil {
		return []byte(`{}`), nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal inbound request metadata: %w", err)
	}
	if len(payload) == 0 {
		return []byte(`{}`), nil
	}
	return payload, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanInboundRequest(row rowScanner) (InboundRequest, error) {
	var request InboundRequest
	var metadata []byte
	err := row.Scan(
		&request.ID,
		&request.OrgID,
		&request.RequestNumber,
		&request.RequestReference,
		&request.SessionID,
		&request.ActorUserID,
		&request.OriginType,
		&request.Channel,
		&request.Status,
		&metadata,
		&request.CancellationReason,
		&request.FailureReason,
		&request.ReceivedAt,
		&request.QueuedAt,
		&request.ProcessingStartedAt,
		&request.ProcessedAt,
		&request.ActedOnAt,
		&request.CompletedAt,
		&request.FailedAt,
		&request.CancelledAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	if err != nil {
		return InboundRequest{}, err
	}
	request.Metadata = append(request.Metadata[:0], metadata...)
	return request, nil
}

func nextRequestReference(ctx context.Context, tx *sql.Tx, orgID string) (int64, string, error) {
	const statement = `
INSERT INTO ai.inbound_request_numbering_series (org_id, next_number)
VALUES ($1, 2)
ON CONFLICT (org_id)
DO UPDATE SET
	next_number = ai.inbound_request_numbering_series.next_number + 1,
	updated_at = NOW()
RETURNING next_number - 1;`

	var requestNumber int64
	if err := tx.QueryRowContext(ctx, statement, orgID).Scan(&requestNumber); err != nil {
		return 0, "", fmt.Errorf("allocate inbound request number: %w", err)
	}

	return requestNumber, fmt.Sprintf("%s%06d", requestReferencePrefix, requestNumber), nil
}

func scanMessage(row rowScanner) (Message, error) {
	var message Message
	err := row.Scan(
		&message.ID,
		&message.OrgID,
		&message.RequestID,
		&message.MessageIndex,
		&message.MessageRole,
		&message.TextContent,
		&message.CreatedByUserID,
		&message.CreatedAt,
	)
	if err != nil {
		return Message{}, err
	}
	return message, nil
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func isValidMessageRole(role string) bool {
	switch role {
	case MessageRoleRequest, MessageRoleSystem, MessageRoleAssistant, MessageRoleTranscription:
		return true
	default:
		return false
	}
}

func isValidAdvanceStatus(status string) bool {
	switch status {
	case StatusProcessed, StatusActedOn, StatusCompleted, StatusFailed:
		return true
	default:
		return false
	}
}

func isAllowedAdvance(current, next string) bool {
	switch current {
	case StatusProcessing:
		return next == StatusProcessed || next == StatusActedOn || next == StatusCompleted || next == StatusFailed
	case StatusProcessed:
		return next == StatusActedOn || next == StatusCompleted
	case StatusActedOn:
		return next == StatusCompleted
	default:
		return false
	}
}
