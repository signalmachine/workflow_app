package workflow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
)

var (
	ErrApprovalNotFound      = errors.New("approval not found")
	ErrInvalidApproval       = errors.New("invalid approval")
	ErrInvalidApprovalInput  = errors.New("invalid approval input")
	ErrApprovalState         = errors.New("invalid approval state")
	ErrApprovalQueueRequired = errors.New("approval queue is required")
	ErrTaskNotFound          = errors.New("task not found")
	ErrInvalidTask           = errors.New("invalid task")
	ErrInvalidTaskState      = errors.New("invalid task state")
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

type Task struct {
	ID                  string
	OrgID               string
	ContextType         string
	ContextID           string
	Title               string
	Instructions        string
	QueueCode           sql.NullString
	Status              string
	AccountableWorkerID string
	CreatedByUserID     string
	CompletedByUserID   sql.NullString
	CompletedAt         sql.NullTime
	CreatedAt           time.Time
	UpdatedAt           time.Time
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

type CreateTaskInput struct {
	ContextType         string
	ContextID           string
	Title               string
	Instructions        string
	QueueCode           string
	AccountableWorkerID string
	Actor               identityaccess.Actor
}

type UpdateTaskStatusInput struct {
	TaskID string
	Status string
	Actor  identityaccess.Actor
}

type ListTasksInput struct {
	ContextType string
	ContextID   string
	Actor       identityaccess.Actor
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
	input.ApprovalID = strings.TrimSpace(input.ApprovalID)
	input.Decision = strings.TrimSpace(input.Decision)
	if input.ApprovalID == "" {
		return Approval{}, documents.Document{}, ErrInvalidApproval
	}
	if input.DecisionNote != "" && strings.TrimSpace(input.DecisionNote) == "" {
		return Approval{}, documents.Document{}, ErrInvalidApprovalInput
	}
	input.DecisionNote = strings.TrimSpace(input.DecisionNote)
	if input.Decision != "approved" && input.Decision != "rejected" {
		return Approval{}, documents.Document{}, ErrInvalidApprovalInput
	}

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
	documentState, err := loadDocumentState(ctx, tx, input.Actor.OrgID, approval.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return Approval{}, documents.Document{}, err
	}
	currentDocument := documents.Document{
		ID:     approval.DocumentID,
		Status: documents.Status(documentState),
	}
	if approval.Status != "pending" {
		_ = tx.Rollback()
		return approval, currentDocument, ErrApprovalState
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

func (s *Service) CreateTask(ctx context.Context, input CreateTaskInput) (Task, error) {
	if !isValidTaskContextType(input.ContextType) || strings.TrimSpace(input.ContextID) == "" || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.AccountableWorkerID) == "" {
		return Task{}, ErrInvalidTask
	}
	if input.Instructions != "" && strings.TrimSpace(input.Instructions) == "" {
		return Task{}, ErrInvalidTask
	}
	if input.QueueCode != "" && strings.TrimSpace(input.QueueCode) == "" {
		return Task{}, ErrInvalidTask
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, fmt.Errorf("begin create task: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Task{}, err
	}

	if err := ensureTaskContextExists(ctx, tx, input.Actor.OrgID, input.ContextType, input.ContextID); err != nil {
		_ = tx.Rollback()
		return Task{}, err
	}
	if err := ensureAccountableWorkerExists(ctx, tx, input.Actor.OrgID, input.AccountableWorkerID); err != nil {
		_ = tx.Rollback()
		return Task{}, err
	}

	task, err := scanTask(tx.QueryRowContext(ctx, `
INSERT INTO workflow.tasks (
	org_id,
	context_type,
	context_id,
	title,
	instructions,
	queue_code,
	accountable_worker_id,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
	id,
	org_id,
	context_type,
	context_id,
	title,
	instructions,
	queue_code,
	status,
	accountable_worker_id,
	created_by_user_id,
	completed_by_user_id,
	completed_at,
	created_at,
	updated_at;`,
		input.Actor.OrgID,
		input.ContextType,
		input.ContextID,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Instructions),
		nullIfEmpty(strings.TrimSpace(input.QueueCode)),
		input.AccountableWorkerID,
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return Task{}, fmt.Errorf("insert task: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "workflow.task_created",
		EntityType:  "workflow.task",
		EntityID:    task.ID,
		Payload: map[string]any{
			"context_type":          task.ContextType,
			"context_id":            task.ContextID,
			"accountable_worker_id": task.AccountableWorkerID,
			"status":                task.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return Task{}, fmt.Errorf("commit create task: %w", err)
	}

	return task, nil
}

func (s *Service) UpdateTaskStatus(ctx context.Context, input UpdateTaskStatusInput) (Task, error) {
	if !isValidTaskStatus(input.Status) {
		return Task{}, ErrInvalidTaskState
	}
	if strings.TrimSpace(input.TaskID) == "" {
		return Task{}, ErrInvalidTask
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, fmt.Errorf("begin update task status: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Task{}, err
	}

	task, err := getTaskForUpdate(ctx, tx, input.Actor.OrgID, input.TaskID)
	if err != nil {
		_ = tx.Rollback()
		return Task{}, err
	}
	if !isAllowedTaskTransition(task.Status, input.Status) {
		_ = tx.Rollback()
		return Task{}, ErrInvalidTaskState
	}

	task, err = scanTask(tx.QueryRowContext(ctx, `
UPDATE workflow.tasks
SET status = $3,
	completed_by_user_id = CASE WHEN $3 IN ('completed', 'cancelled') THEN $4::uuid ELSE NULL END,
	completed_at = CASE WHEN $3 IN ('completed', 'cancelled') THEN NOW() ELSE NULL END,
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	context_type,
	context_id,
	title,
	instructions,
	queue_code,
	status,
	accountable_worker_id,
	created_by_user_id,
	completed_by_user_id,
	completed_at,
	created_at,
	updated_at;`,
		input.Actor.OrgID,
		input.TaskID,
		input.Status,
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return Task{}, fmt.Errorf("update task status: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "workflow.task_status_updated",
		EntityType:  "workflow.task",
		EntityID:    task.ID,
		Payload: map[string]any{
			"status":       task.Status,
			"context_type": task.ContextType,
			"context_id":   task.ContextID,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return Task{}, fmt.Errorf("commit update task status: %w", err)
	}

	return task, nil
}

func (s *Service) ListTasks(ctx context.Context, input ListTasksInput) ([]Task, error) {
	if !isValidTaskContextType(input.ContextType) || strings.TrimSpace(input.ContextID) == "" {
		return nil, ErrInvalidTask
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list tasks: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator, identityaccess.RoleApprover); err != nil {
		return nil, err
	}

	if err := ensureTaskContextExists(ctx, tx, input.Actor.OrgID, input.ContextType, input.ContextID); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
	id,
	org_id,
	context_type,
	context_id,
	title,
	instructions,
	queue_code,
	status,
	accountable_worker_id,
	created_by_user_id,
	completed_by_user_id,
	completed_at,
	created_at,
	updated_at
FROM workflow.tasks
WHERE org_id = $1
  AND context_type = $2
  AND context_id = $3
ORDER BY created_at, id;`,
		input.Actor.OrgID,
		input.ContextType,
		input.ContextID,
	)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list tasks: %w", err)
	}

	return tasks, nil
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

func getTaskForUpdate(ctx context.Context, tx *sql.Tx, orgID, taskID string) (Task, error) {
	const query = `
SELECT
	id,
	org_id,
	context_type,
	context_id,
	title,
	instructions,
	queue_code,
	status,
	accountable_worker_id,
	created_by_user_id,
	completed_by_user_id,
	completed_at,
	created_at,
	updated_at
FROM workflow.tasks
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	task, err := scanTask(tx.QueryRowContext(ctx, query, orgID, taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrTaskNotFound
		}
		return Task{}, fmt.Errorf("get task: %w", err)
	}
	return task, nil
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

func ensureTaskContextExists(ctx context.Context, tx *sql.Tx, orgID, contextType, contextID string) error {
	switch contextType {
	case "work_order":
		var exists bool
		if err := tx.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM work_orders.work_orders
	WHERE org_id = $1
	  AND id = $2
);`, orgID, contextID).Scan(&exists); err != nil {
			return fmt.Errorf("check work order task context: %w", err)
		}
		if !exists {
			return ErrInvalidTask
		}
		return nil
	default:
		return ErrInvalidTask
	}
}

func ensureAccountableWorkerExists(ctx context.Context, tx *sql.Tx, orgID, workerID string) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM workforce.workers
	WHERE org_id = $1
	  AND id = $2
	  AND status = 'active'
);`, orgID, workerID).Scan(&exists); err != nil {
		return fmt.Errorf("check accountable worker: %w", err)
	}
	if !exists {
		return ErrInvalidTask
	}
	return nil
}

func isValidTaskContextType(contextType string) bool {
	return contextType == "work_order"
}

func isValidTaskStatus(status string) bool {
	switch status {
	case "open", "in_progress", "completed", "cancelled":
		return true
	default:
		return false
	}
}

func isAllowedTaskTransition(fromStatus, toStatus string) bool {
	switch fromStatus {
	case "open":
		return toStatus == "in_progress" || toStatus == "completed" || toStatus == "cancelled"
	case "in_progress":
		return toStatus == "completed" || toStatus == "cancelled"
	default:
		return false
	}
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

func scanTask(row rowScanner) (Task, error) {
	var task Task
	err := row.Scan(
		&task.ID,
		&task.OrgID,
		&task.ContextType,
		&task.ContextID,
		&task.Title,
		&task.Instructions,
		&task.QueueCode,
		&task.Status,
		&task.AccountableWorkerID,
		&task.CreatedByUserID,
		&task.CompletedByUserID,
		&task.CompletedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
