package workforce

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
	"workflow_app/internal/workflow"
)

var (
	ErrWorkerNotFound        = errors.New("worker not found")
	ErrInvalidWorker         = errors.New("invalid worker")
	ErrInvalidLaborEntry     = errors.New("invalid labor entry")
	ErrTaskOwnershipMismatch = errors.New("task accountable worker mismatch")
)

const (
	WorkerStatusActive   = "active"
	WorkerStatusInactive = "inactive"
)

type Worker struct {
	ID                     string
	OrgID                  string
	WorkerCode             string
	DisplayName            string
	LinkedUserID           sql.NullString
	Status                 string
	DefaultHourlyCostMinor int64
	CostCurrencyCode       string
	CreatedByUserID        string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type LaborEntry struct {
	ID               string
	OrgID            string
	WorkerID         string
	WorkOrderID      string
	TaskID           sql.NullString
	StartedAt        time.Time
	EndedAt          time.Time
	DurationMinutes  int
	HourlyCostMinor  int64
	CostMinor        int64
	CostCurrencyCode string
	Note             string
	CapturedByUserID string
	CreatedAt        time.Time
}

type laborAccountingHandoff struct {
	ID              string
	OrgID           string
	LaborEntryID    string
	WorkOrderID     string
	TaskID          sql.NullString
	JournalEntryID  sql.NullString
	HandoffStatus   string
	CreatedByUserID string
	CreatedAt       time.Time
	PostedAt        sql.NullTime
}

type CreateWorkerInput struct {
	WorkerCode             string
	DisplayName            string
	LinkedUserID           string
	DefaultHourlyCostMinor int64
	CostCurrencyCode       string
	Actor                  identityaccess.Actor
}

type RecordLaborInput struct {
	WorkerID    string
	WorkOrderID string
	TaskID      string
	StartedAt   time.Time
	EndedAt     time.Time
	Note        string
	Actor       identityaccess.Actor
}

type ListLaborEntriesInput struct {
	WorkOrderID string
	Actor       identityaccess.Actor
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateWorker(ctx context.Context, input CreateWorkerInput) (Worker, error) {
	currencyCode := strings.ToUpper(strings.TrimSpace(input.CostCurrencyCode))
	if strings.TrimSpace(input.WorkerCode) == "" || strings.TrimSpace(input.DisplayName) == "" || !isValidCurrencyCode(currencyCode) || input.DefaultHourlyCostMinor < 0 {
		return Worker{}, ErrInvalidWorker
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Worker{}, fmt.Errorf("begin create worker: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Worker{}, err
	}

	worker, err := scanWorker(tx.QueryRowContext(ctx, `
INSERT INTO workforce.workers (
	org_id,
	worker_code,
	display_name,
	linked_user_id,
	default_hourly_cost_minor,
	cost_currency_code,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
	id,
	org_id,
	worker_code,
	display_name,
	linked_user_id,
	status,
	default_hourly_cost_minor,
	cost_currency_code,
	created_by_user_id,
	created_at,
	updated_at;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.WorkerCode),
		strings.TrimSpace(input.DisplayName),
		nullIfEmpty(strings.TrimSpace(input.LinkedUserID)),
		input.DefaultHourlyCostMinor,
		currencyCode,
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return Worker{}, fmt.Errorf("insert worker: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "workforce.worker_created",
		EntityType:  "workforce.worker",
		EntityID:    worker.ID,
		Payload: map[string]any{
			"worker_code":               worker.WorkerCode,
			"default_hourly_cost_minor": worker.DefaultHourlyCostMinor,
			"cost_currency_code":        worker.CostCurrencyCode,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Worker{}, err
	}

	if err := tx.Commit(); err != nil {
		return Worker{}, fmt.Errorf("commit create worker: %w", err)
	}

	return worker, nil
}

func (s *Service) RecordLabor(ctx context.Context, input RecordLaborInput) (LaborEntry, error) {
	if strings.TrimSpace(input.WorkerID) == "" || strings.TrimSpace(input.WorkOrderID) == "" || !input.EndedAt.After(input.StartedAt) {
		return LaborEntry{}, ErrInvalidLaborEntry
	}
	if input.Note != "" && strings.TrimSpace(input.Note) == "" {
		return LaborEntry{}, ErrInvalidLaborEntry
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return LaborEntry{}, fmt.Errorf("begin record labor: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return LaborEntry{}, err
	}

	worker, err := getWorkerForUpdate(ctx, tx, input.Actor.OrgID, input.WorkerID)
	if err != nil {
		_ = tx.Rollback()
		return LaborEntry{}, err
	}
	if worker.Status != WorkerStatusActive {
		_ = tx.Rollback()
		return LaborEntry{}, ErrInvalidLaborEntry
	}

	if err := ensureWorkOrderExists(ctx, tx, input.Actor.OrgID, input.WorkOrderID); err != nil {
		_ = tx.Rollback()
		return LaborEntry{}, err
	}

	var taskID sql.NullString
	if trimmedTaskID := strings.TrimSpace(input.TaskID); trimmedTaskID != "" {
		task, err := getTaskForLabor(ctx, tx, input.Actor.OrgID, trimmedTaskID)
		if err != nil {
			_ = tx.Rollback()
			return LaborEntry{}, err
		}
		if task.ContextType != "work_order" || task.ContextID != input.WorkOrderID {
			_ = tx.Rollback()
			return LaborEntry{}, ErrInvalidLaborEntry
		}
		if task.AccountableWorkerID != worker.ID {
			_ = tx.Rollback()
			return LaborEntry{}, ErrTaskOwnershipMismatch
		}
		taskID = sql.NullString{String: task.ID, Valid: true}
	}

	durationMinutes := int(input.EndedAt.Sub(input.StartedAt).Minutes())
	if durationMinutes <= 0 {
		_ = tx.Rollback()
		return LaborEntry{}, ErrInvalidLaborEntry
	}

	entry, err := scanLaborEntry(tx.QueryRowContext(ctx, `
INSERT INTO workforce.labor_entries (
	org_id,
	worker_id,
	work_order_id,
	task_id,
	started_at,
	ended_at,
	duration_minutes,
	hourly_cost_minor,
	cost_minor,
	cost_currency_code,
	note,
	captured_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING
	id,
	org_id,
	worker_id,
	work_order_id,
	task_id,
	started_at,
	ended_at,
	duration_minutes,
	hourly_cost_minor,
	cost_minor,
	cost_currency_code,
	note,
	captured_by_user_id,
	created_at;`,
		input.Actor.OrgID,
		worker.ID,
		input.WorkOrderID,
		nullStringParam(taskID),
		input.StartedAt.UTC(),
		input.EndedAt.UTC(),
		durationMinutes,
		worker.DefaultHourlyCostMinor,
		roundedCostMinor(worker.DefaultHourlyCostMinor, durationMinutes),
		worker.CostCurrencyCode,
		strings.TrimSpace(input.Note),
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return LaborEntry{}, fmt.Errorf("insert labor entry: %w", err)
	}

	if _, err := insertLaborAccountingHandoffTx(ctx, tx, entry); err != nil {
		_ = tx.Rollback()
		return LaborEntry{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "workforce.labor_recorded",
		EntityType:  "workforce.labor_entry",
		EntityID:    entry.ID,
		Payload: map[string]any{
			"worker_id":         entry.WorkerID,
			"work_order_id":     entry.WorkOrderID,
			"task_id":           nullStringPayload(entry.TaskID),
			"duration_minutes":  entry.DurationMinutes,
			"hourly_cost_minor": entry.HourlyCostMinor,
			"cost_minor":        entry.CostMinor,
		},
	}); err != nil {
		_ = tx.Rollback()
		return LaborEntry{}, err
	}

	if err := tx.Commit(); err != nil {
		return LaborEntry{}, fmt.Errorf("commit record labor: %w", err)
	}

	return entry, nil
}

func (s *Service) ListLaborEntries(ctx context.Context, input ListLaborEntriesInput) ([]LaborEntry, error) {
	if strings.TrimSpace(input.WorkOrderID) == "" {
		return nil, ErrInvalidLaborEntry
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list labor entries: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator, identityaccess.RoleApprover); err != nil {
		return nil, err
	}

	if err := ensureWorkOrderExists(ctx, tx, input.Actor.OrgID, input.WorkOrderID); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
	id,
	org_id,
	worker_id,
	work_order_id,
	task_id,
	started_at,
	ended_at,
	duration_minutes,
	hourly_cost_minor,
	cost_minor,
	cost_currency_code,
	note,
	captured_by_user_id,
	created_at
FROM workforce.labor_entries
WHERE org_id = $1
  AND work_order_id = $2
ORDER BY started_at DESC, id DESC;`,
		input.Actor.OrgID,
		input.WorkOrderID,
	)
	if err != nil {
		return nil, fmt.Errorf("query labor entries: %w", err)
	}
	defer rows.Close()

	var entries []LaborEntry
	for rows.Next() {
		entry, err := scanLaborEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate labor entries: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list labor entries: %w", err)
	}

	return entries, nil
}

func getWorkerForUpdate(ctx context.Context, tx *sql.Tx, orgID, workerID string) (Worker, error) {
	const query = `
SELECT
	id,
	org_id,
	worker_code,
	display_name,
	linked_user_id,
	status,
	default_hourly_cost_minor,
	cost_currency_code,
	created_by_user_id,
	created_at,
	updated_at
FROM workforce.workers
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	worker, err := scanWorker(tx.QueryRowContext(ctx, query, orgID, workerID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Worker{}, ErrWorkerNotFound
		}
		return Worker{}, fmt.Errorf("get worker: %w", err)
	}
	return worker, nil
}

func ensureWorkOrderExists(ctx context.Context, tx *sql.Tx, orgID, workOrderID string) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM work_orders.work_orders
	WHERE org_id = $1
	  AND id = $2
);`, orgID, workOrderID).Scan(&exists); err != nil {
		return fmt.Errorf("check work order: %w", err)
	}
	if !exists {
		return ErrInvalidLaborEntry
	}
	return nil
}

func getTaskForLabor(ctx context.Context, tx *sql.Tx, orgID, taskID string) (workflow.Task, error) {
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
			return workflow.Task{}, workflow.ErrTaskNotFound
		}
		return workflow.Task{}, fmt.Errorf("get task for labor: %w", err)
	}
	return task, nil
}

func insertLaborAccountingHandoffTx(ctx context.Context, tx *sql.Tx, entry LaborEntry) (laborAccountingHandoff, error) {
	const statement = `
INSERT INTO workforce.labor_accounting_handoffs (
	org_id,
	labor_entry_id,
	work_order_id,
	task_id,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING
	id,
	org_id,
	labor_entry_id,
	work_order_id,
	task_id,
	journal_entry_id,
	handoff_status,
	created_by_user_id,
	created_at,
	posted_at;`

	handoff, err := scanLaborAccountingHandoff(tx.QueryRowContext(
		ctx,
		statement,
		entry.OrgID,
		entry.ID,
		entry.WorkOrderID,
		nullStringParam(entry.TaskID),
		entry.CapturedByUserID,
	))
	if err != nil {
		return laborAccountingHandoff{}, fmt.Errorf("insert labor accounting handoff: %w", err)
	}

	return handoff, nil
}

func roundedCostMinor(hourlyCostMinor int64, durationMinutes int) int64 {
	return (hourlyCostMinor*int64(durationMinutes) + 30) / 60
}

func isValidCurrencyCode(code string) bool {
	return len(code) == 3
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullStringParam(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func nullStringPayload(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func scanWorker(row rowScanner) (Worker, error) {
	var worker Worker
	if err := row.Scan(
		&worker.ID,
		&worker.OrgID,
		&worker.WorkerCode,
		&worker.DisplayName,
		&worker.LinkedUserID,
		&worker.Status,
		&worker.DefaultHourlyCostMinor,
		&worker.CostCurrencyCode,
		&worker.CreatedByUserID,
		&worker.CreatedAt,
		&worker.UpdatedAt,
	); err != nil {
		return Worker{}, err
	}
	return worker, nil
}

func scanLaborEntry(row rowScanner) (LaborEntry, error) {
	var entry LaborEntry
	if err := row.Scan(
		&entry.ID,
		&entry.OrgID,
		&entry.WorkerID,
		&entry.WorkOrderID,
		&entry.TaskID,
		&entry.StartedAt,
		&entry.EndedAt,
		&entry.DurationMinutes,
		&entry.HourlyCostMinor,
		&entry.CostMinor,
		&entry.CostCurrencyCode,
		&entry.Note,
		&entry.CapturedByUserID,
		&entry.CreatedAt,
	); err != nil {
		return LaborEntry{}, err
	}
	return entry, nil
}

func scanLaborAccountingHandoff(row rowScanner) (laborAccountingHandoff, error) {
	var handoff laborAccountingHandoff
	if err := row.Scan(
		&handoff.ID,
		&handoff.OrgID,
		&handoff.LaborEntryID,
		&handoff.WorkOrderID,
		&handoff.TaskID,
		&handoff.JournalEntryID,
		&handoff.HandoffStatus,
		&handoff.CreatedByUserID,
		&handoff.CreatedAt,
		&handoff.PostedAt,
	); err != nil {
		return laborAccountingHandoff{}, err
	}
	return handoff, nil
}

func scanTask(row rowScanner) (workflow.Task, error) {
	var task workflow.Task
	if err := row.Scan(
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
	); err != nil {
		return workflow.Task{}, err
	}
	return task, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}
