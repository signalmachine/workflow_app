package workorders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
)

var (
	ErrWorkOrderNotFound       = errors.New("work order not found")
	ErrInvalidWorkOrder        = errors.New("invalid work order")
	ErrInvalidWorkOrderStatus  = errors.New("invalid work order status")
	ErrInvalidStatusTransition = errors.New("invalid work order status transition")
)

const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

type WorkOrder struct {
	ID              string
	OrgID           string
	WorkOrderCode   string
	Title           string
	Summary         string
	Status          string
	CreatedByUserID string
	ClosedAt        sql.NullTime
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StatusHistory struct {
	ID              string
	OrgID           string
	WorkOrderID     string
	FromStatus      sql.NullString
	ToStatus        string
	Note            string
	ChangedByUserID string
	ChangedAt       time.Time
}

type MaterialUsage struct {
	ID                       string
	OrgID                    string
	WorkOrderID              string
	InventoryExecutionLinkID string
	InventoryDocumentID      string
	InventoryDocumentLineID  string
	InventoryMovementID      string
	ItemID                   string
	MovementPurpose          string
	UsageClassification      string
	QuantityMilli            int64
	LinkedByUserID           string
	LinkedAt                 time.Time
}

type CreateWorkOrderInput struct {
	WorkOrderCode string
	Title         string
	Summary       string
	Actor         identityaccess.Actor
}

type CreateWorkOrderResult struct {
	WorkOrder      WorkOrder
	InitialHistory StatusHistory
	MaterialUsages []MaterialUsage
}

type UpdateStatusInput struct {
	WorkOrderID string
	Status      string
	Note        string
	Actor       identityaccess.Actor
}

type SyncInventoryUsageInput struct {
	WorkOrderID string
	Actor       identityaccess.Actor
}

type ListMaterialUsagesInput struct {
	WorkOrderID string
	Actor       identityaccess.Actor
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateWorkOrder(ctx context.Context, input CreateWorkOrderInput) (CreateWorkOrderResult, error) {
	if strings.TrimSpace(input.WorkOrderCode) == "" || strings.TrimSpace(input.Title) == "" {
		return CreateWorkOrderResult{}, ErrInvalidWorkOrder
	}
	if input.Summary != "" && strings.TrimSpace(input.Summary) == "" {
		return CreateWorkOrderResult{}, ErrInvalidWorkOrder
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CreateWorkOrderResult{}, fmt.Errorf("begin create work order: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return CreateWorkOrderResult{}, err
	}

	workOrder, err := scanWorkOrder(tx.QueryRowContext(ctx, `
INSERT INTO work_orders.work_orders (
	org_id,
	work_order_code,
	title,
	summary,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING
	id,
	org_id,
	work_order_code,
	title,
	summary,
	status,
	created_by_user_id,
	closed_at,
	created_at,
	updated_at;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.WorkOrderCode),
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Summary),
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return CreateWorkOrderResult{}, fmt.Errorf("insert work order: %w", err)
	}

	history, err := insertStatusHistoryTx(ctx, tx, input.Actor.OrgID, workOrder.ID, sql.NullString{}, StatusOpen, "", input.Actor.UserID)
	if err != nil {
		_ = tx.Rollback()
		return CreateWorkOrderResult{}, err
	}

	materialUsages, err := consumePendingInventoryUsageTx(ctx, tx, workOrder, input.Actor.UserID)
	if err != nil {
		_ = tx.Rollback()
		return CreateWorkOrderResult{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "work_orders.work_order_created",
		EntityType:  "work_orders.work_order",
		EntityID:    workOrder.ID,
		Payload: map[string]any{
			"work_order_code":    workOrder.WorkOrderCode,
			"status":             workOrder.Status,
			"material_usage_cnt": len(materialUsages),
		},
	}); err != nil {
		_ = tx.Rollback()
		return CreateWorkOrderResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return CreateWorkOrderResult{}, fmt.Errorf("commit create work order: %w", err)
	}

	return CreateWorkOrderResult{
		WorkOrder:      workOrder,
		InitialHistory: history,
		MaterialUsages: materialUsages,
	}, nil
}

func (s *Service) UpdateStatus(ctx context.Context, input UpdateStatusInput) (WorkOrder, StatusHistory, error) {
	if !isValidStatus(input.Status) {
		return WorkOrder{}, StatusHistory{}, ErrInvalidWorkOrderStatus
	}
	if input.Note != "" && strings.TrimSpace(input.Note) == "" {
		return WorkOrder{}, StatusHistory{}, ErrInvalidWorkOrder
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WorkOrder{}, StatusHistory{}, fmt.Errorf("begin update work order status: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return WorkOrder{}, StatusHistory{}, err
	}

	workOrder, err := getWorkOrderForUpdate(ctx, tx, input.Actor.OrgID, input.WorkOrderID)
	if err != nil {
		_ = tx.Rollback()
		return WorkOrder{}, StatusHistory{}, err
	}
	previousStatus := workOrder.Status
	if !isAllowedTransition(workOrder.Status, input.Status) {
		_ = tx.Rollback()
		return WorkOrder{}, StatusHistory{}, ErrInvalidStatusTransition
	}

	workOrder, err = scanWorkOrder(tx.QueryRowContext(ctx, `
UPDATE work_orders.work_orders
SET status = $3,
	closed_at = CASE WHEN $3 IN ('completed', 'cancelled') THEN NOW() ELSE NULL END,
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	work_order_code,
	title,
	summary,
	status,
	created_by_user_id,
	closed_at,
	created_at,
	updated_at;`,
		input.Actor.OrgID,
		input.WorkOrderID,
		input.Status,
	))
	if err != nil {
		_ = tx.Rollback()
		return WorkOrder{}, StatusHistory{}, fmt.Errorf("update work order status: %w", err)
	}

	history, err := insertStatusHistoryTx(ctx, tx, input.Actor.OrgID, workOrder.ID, sql.NullString{String: previousStatus, Valid: true}, input.Status, strings.TrimSpace(input.Note), input.Actor.UserID)
	if err != nil {
		_ = tx.Rollback()
		return WorkOrder{}, StatusHistory{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "work_orders.work_order_status_updated",
		EntityType:  "work_orders.work_order",
		EntityID:    workOrder.ID,
		Payload: map[string]any{
			"status": workOrder.Status,
			"note":   strings.TrimSpace(input.Note),
		},
	}); err != nil {
		_ = tx.Rollback()
		return WorkOrder{}, StatusHistory{}, err
	}

	if err := tx.Commit(); err != nil {
		return WorkOrder{}, StatusHistory{}, fmt.Errorf("commit update work order status: %w", err)
	}

	return workOrder, history, nil
}

func (s *Service) SyncInventoryUsage(ctx context.Context, input SyncInventoryUsageInput) ([]MaterialUsage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin sync work order inventory usage: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	workOrder, err := getWorkOrderForUpdate(ctx, tx, input.Actor.OrgID, input.WorkOrderID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	materialUsages, err := consumePendingInventoryUsageTx(ctx, tx, workOrder, input.Actor.UserID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "work_orders.inventory_usage_synced",
		EntityType:  "work_orders.work_order",
		EntityID:    workOrder.ID,
		Payload: map[string]any{
			"work_order_code":    workOrder.WorkOrderCode,
			"material_usage_cnt": len(materialUsages),
		},
	}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit sync work order inventory usage: %w", err)
	}

	return materialUsages, nil
}

func (s *Service) ListMaterialUsages(ctx context.Context, input ListMaterialUsagesInput) ([]MaterialUsage, error) {
	if strings.TrimSpace(input.WorkOrderID) == "" {
		return nil, ErrInvalidWorkOrder
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list work order material usage: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator, identityaccess.RoleApprover); err != nil {
		return nil, err
	}

	if _, err := getWorkOrder(ctx, tx, input.Actor.OrgID, input.WorkOrderID); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
	id,
	org_id,
	work_order_id,
	inventory_execution_link_id,
	inventory_document_id,
	inventory_document_line_id,
	inventory_movement_id,
	item_id,
	movement_purpose,
	usage_classification,
	quantity_milli,
	linked_by_user_id,
	linked_at
FROM work_orders.material_usages
WHERE org_id = $1
  AND work_order_id = $2
ORDER BY linked_at, id;`,
		input.Actor.OrgID,
		input.WorkOrderID,
	)
	if err != nil {
		return nil, fmt.Errorf("query work order material usage: %w", err)
	}
	defer rows.Close()

	var usages []MaterialUsage
	for rows.Next() {
		usage, err := scanMaterialUsage(rows)
		if err != nil {
			return nil, err
		}
		usages = append(usages, usage)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate work order material usage: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list work order material usage: %w", err)
	}

	return usages, nil
}

func getWorkOrderForUpdate(ctx context.Context, tx *sql.Tx, orgID, workOrderID string) (WorkOrder, error) {
	return getWorkOrderWithLockClause(ctx, tx, orgID, workOrderID, "FOR UPDATE")
}

func getWorkOrder(ctx context.Context, tx *sql.Tx, orgID, workOrderID string) (WorkOrder, error) {
	return getWorkOrderWithLockClause(ctx, tx, orgID, workOrderID, "")
}

func getWorkOrderWithLockClause(ctx context.Context, tx *sql.Tx, orgID, workOrderID, lockClause string) (WorkOrder, error) {
	query := `
SELECT
	id,
	org_id,
	work_order_code,
	title,
	summary,
	status,
	created_by_user_id,
	closed_at,
	created_at,
	updated_at
FROM work_orders.work_orders
WHERE org_id = $1
  AND id = $2`
	if lockClause != "" {
		query += "\n" + lockClause
	}
	query += ";"

	workOrder, err := scanWorkOrder(tx.QueryRowContext(ctx, query, orgID, workOrderID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkOrder{}, ErrWorkOrderNotFound
		}
		return WorkOrder{}, fmt.Errorf("get work order: %w", err)
	}
	return workOrder, nil
}

func insertStatusHistoryTx(ctx context.Context, tx *sql.Tx, orgID, workOrderID string, fromStatus sql.NullString, toStatus, note, actorUserID string) (StatusHistory, error) {
	history, err := scanStatusHistory(tx.QueryRowContext(ctx, `
INSERT INTO work_orders.status_history (
	org_id,
	work_order_id,
	from_status,
	to_status,
	note,
	changed_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
	id,
	org_id,
	work_order_id,
	from_status,
	to_status,
	note,
	changed_by_user_id,
	changed_at;`,
		orgID,
		workOrderID,
		nullStringParam(fromStatus),
		toStatus,
		note,
		actorUserID,
	))
	if err != nil {
		return StatusHistory{}, fmt.Errorf("insert work order status history: %w", err)
	}
	return history, nil
}

func consumePendingInventoryUsageTx(ctx context.Context, tx *sql.Tx, workOrder WorkOrder, actorUserID string) ([]MaterialUsage, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT
	el.id,
	el.document_id,
	el.document_line_id,
	dl.movement_id,
	dl.item_id,
	dl.movement_purpose,
	dl.usage_classification,
	dl.quantity_milli
FROM inventory_ops.execution_links el
JOIN inventory_ops.document_lines dl
	ON dl.id = el.document_line_id
WHERE el.org_id = $1
  AND el.execution_context_type = 'work_order'
  AND el.execution_context_id = $2
  AND el.linkage_status = 'pending'
ORDER BY el.created_at, el.id
FOR UPDATE OF el;`,
		workOrder.OrgID,
		workOrder.WorkOrderCode,
	)
	if err != nil {
		return nil, fmt.Errorf("query pending inventory execution links: %w", err)
	}
	defer rows.Close()

	type pendingUsage struct {
		executionLinkID     string
		documentID          string
		documentLineID      string
		movementID          string
		itemID              string
		movementPurpose     string
		usageClassification string
		quantityMilli       int64
	}

	var pending []pendingUsage
	for rows.Next() {
		var usage pendingUsage
		if err := rows.Scan(
			&usage.executionLinkID,
			&usage.documentID,
			&usage.documentLineID,
			&usage.movementID,
			&usage.itemID,
			&usage.movementPurpose,
			&usage.usageClassification,
			&usage.quantityMilli,
		); err != nil {
			return nil, fmt.Errorf("scan pending inventory execution link: %w", err)
		}
		pending = append(pending, usage)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending inventory execution links: %w", err)
	}

	materialUsages := make([]MaterialUsage, 0, len(pending))
	for _, usage := range pending {
		materialUsage, err := scanMaterialUsage(tx.QueryRowContext(ctx, `
INSERT INTO work_orders.material_usages (
	org_id,
	work_order_id,
	inventory_execution_link_id,
	inventory_document_id,
	inventory_document_line_id,
	inventory_movement_id,
	item_id,
	movement_purpose,
	usage_classification,
	quantity_milli,
	linked_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING
	id,
	org_id,
	work_order_id,
	inventory_execution_link_id,
	inventory_document_id,
	inventory_document_line_id,
	inventory_movement_id,
	item_id,
	movement_purpose,
	usage_classification,
	quantity_milli,
	linked_by_user_id,
	linked_at;`,
			workOrder.OrgID,
			workOrder.ID,
			usage.executionLinkID,
			usage.documentID,
			usage.documentLineID,
			usage.movementID,
			usage.itemID,
			usage.movementPurpose,
			usage.usageClassification,
			usage.quantityMilli,
			actorUserID,
		))
		if err != nil {
			return nil, fmt.Errorf("insert work order material usage: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
UPDATE inventory_ops.execution_links
SET linkage_status = 'linked'
WHERE org_id = $1
  AND id = $2;`,
			workOrder.OrgID,
			usage.executionLinkID,
		); err != nil {
			return nil, fmt.Errorf("update inventory execution link: %w", err)
		}

		materialUsages = append(materialUsages, materialUsage)
	}

	return materialUsages, nil
}

func scanWorkOrder(row rowScanner) (WorkOrder, error) {
	var workOrder WorkOrder
	if err := row.Scan(
		&workOrder.ID,
		&workOrder.OrgID,
		&workOrder.WorkOrderCode,
		&workOrder.Title,
		&workOrder.Summary,
		&workOrder.Status,
		&workOrder.CreatedByUserID,
		&workOrder.ClosedAt,
		&workOrder.CreatedAt,
		&workOrder.UpdatedAt,
	); err != nil {
		return WorkOrder{}, err
	}
	return workOrder, nil
}

func scanStatusHistory(row rowScanner) (StatusHistory, error) {
	var history StatusHistory
	if err := row.Scan(
		&history.ID,
		&history.OrgID,
		&history.WorkOrderID,
		&history.FromStatus,
		&history.ToStatus,
		&history.Note,
		&history.ChangedByUserID,
		&history.ChangedAt,
	); err != nil {
		return StatusHistory{}, err
	}
	return history, nil
}

func scanMaterialUsage(row rowScanner) (MaterialUsage, error) {
	var usage MaterialUsage
	if err := row.Scan(
		&usage.ID,
		&usage.OrgID,
		&usage.WorkOrderID,
		&usage.InventoryExecutionLinkID,
		&usage.InventoryDocumentID,
		&usage.InventoryDocumentLineID,
		&usage.InventoryMovementID,
		&usage.ItemID,
		&usage.MovementPurpose,
		&usage.UsageClassification,
		&usage.QuantityMilli,
		&usage.LinkedByUserID,
		&usage.LinkedAt,
	); err != nil {
		return MaterialUsage{}, err
	}
	return usage, nil
}

func isValidStatus(status string) bool {
	switch status {
	case StatusOpen, StatusInProgress, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}

func isAllowedTransition(fromStatus, toStatus string) bool {
	switch fromStatus {
	case StatusOpen:
		return toStatus == StatusInProgress || toStatus == StatusCompleted || toStatus == StatusCancelled
	case StatusInProgress:
		return toStatus == StatusCompleted || toStatus == StatusCancelled
	default:
		return false
	}
}

func nullStringParam(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}
