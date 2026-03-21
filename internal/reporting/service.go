package reporting

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/identityaccess"
)

var (
	ErrInvalidReviewFilter = errors.New("invalid review filter")
	ErrWorkOrderNotFound   = errors.New("work order not found")
)

type ApprovalQueueEntry struct {
	QueueEntryID         string
	ApprovalID           string
	QueueCode            string
	QueueStatus          string
	EnqueuedAt           time.Time
	ClosedAt             sql.NullTime
	ApprovalStatus       string
	RequestedAt          time.Time
	RequestedByUserID    string
	DecidedAt            sql.NullTime
	DecidedByUserID      sql.NullString
	DocumentID           string
	DocumentTypeCode     string
	DocumentTitle        string
	DocumentNumber       sql.NullString
	DocumentStatus       string
	JournalEntryID       sql.NullString
	JournalEntryNumber   sql.NullInt64
	JournalEntryPostedAt sql.NullTime
}

type ListApprovalQueueInput struct {
	QueueCode string
	Status    string
	Limit     int
	Actor     identityaccess.Actor
}

type DocumentReview struct {
	DocumentID           string
	TypeCode             string
	Title                string
	NumberValue          sql.NullString
	Status               string
	SourceDocumentID     sql.NullString
	CreatedByUserID      string
	SubmittedByUserID    sql.NullString
	SubmittedAt          sql.NullTime
	ApprovedAt           sql.NullTime
	RejectedAt           sql.NullTime
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ApprovalID           sql.NullString
	ApprovalStatus       sql.NullString
	ApprovalQueueCode    sql.NullString
	ApprovalRequestedAt  sql.NullTime
	ApprovalDecidedAt    sql.NullTime
	JournalEntryID       sql.NullString
	JournalEntryNumber   sql.NullInt64
	JournalEntryPostedAt sql.NullTime
}

type ListDocumentsInput struct {
	TypeCode string
	Status   string
	Limit    int
	Actor    identityaccess.Actor
}

type InventoryStockItem struct {
	ItemID       string
	ItemSKU      string
	ItemName     string
	ItemRole     string
	LocationID   string
	LocationCode string
	LocationName string
	LocationRole string
	OnHandMilli  int64
}

type ListInventoryStockInput struct {
	ItemID      string
	LocationID  string
	IncludeZero bool
	Limit       int
	Actor       identityaccess.Actor
}

type WorkOrderReview struct {
	WorkOrderID              string
	WorkOrderCode            string
	Title                    string
	Summary                  string
	Status                   string
	ClosedAt                 sql.NullTime
	CreatedAt                time.Time
	UpdatedAt                time.Time
	LastStatusChangedAt      time.Time
	OpenTaskCount            int
	CompletedTaskCount       int
	LaborEntryCount          int
	TotalLaborMinutes        int
	TotalLaborCostMinor      int64
	PostedLaborEntryCount    int
	PostedLaborCostMinor     int64
	MaterialUsageCount       int
	MaterialQuantityMilli    int64
	PostedMaterialUsageCount int
	PostedMaterialCostMinor  int64
	LastAccountingPostedAt   sql.NullTime
}

type GetWorkOrderReviewInput struct {
	WorkOrderID string
	Actor       identityaccess.Actor
}

type AuditEvent struct {
	ID          string
	OrgID       sql.NullString
	ActorUserID sql.NullString
	EventType   string
	EntityType  string
	EntityID    string
	Payload     json.RawMessage
	OccurredAt  time.Time
}

type LookupAuditEventsInput struct {
	EntityType string
	EntityID   string
	Limit      int
	Actor      identityaccess.Actor
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListApprovalQueue(ctx context.Context, input ListApprovalQueueInput) ([]ApprovalQueueEntry, error) {
	if input.Status != "" && input.Status != "pending" && input.Status != "closed" {
		return nil, ErrInvalidReviewFilter
	}

	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
SELECT
	aqe.id,
	aqe.approval_id,
	aqe.queue_code,
	aqe.status,
	aqe.enqueued_at,
	aqe.closed_at,
	a.status,
	a.requested_at,
	a.requested_by_user_id,
	a.decided_at,
	a.decided_by_user_id,
	d.id,
	d.type_code,
	d.title,
	d.number_value,
	d.status,
	je.id,
	je.entry_number,
	je.posted_at
FROM workflow.approval_queue_entries aqe
JOIN workflow.approvals a
	ON a.id = aqe.approval_id
   AND a.org_id = aqe.org_id
JOIN documents.documents d
	ON d.id = a.document_id
   AND d.org_id = aqe.org_id
LEFT JOIN accounting.journal_entries je
	ON je.org_id = aqe.org_id
   AND je.source_document_id = d.id
   AND je.entry_kind = 'posting'
WHERE aqe.org_id = $1
  AND ($2 = '' OR aqe.queue_code = $2)
  AND ($3 = '' OR aqe.status = $3)
ORDER BY aqe.enqueued_at DESC, aqe.id DESC
LIMIT $4;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.QueueCode),
		input.Status,
		normalizeLimit(input.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("query approval queue: %w", err)
	}
	defer rows.Close()

	var entries []ApprovalQueueEntry
	for rows.Next() {
		var entry ApprovalQueueEntry
		if err := rows.Scan(
			&entry.QueueEntryID,
			&entry.ApprovalID,
			&entry.QueueCode,
			&entry.QueueStatus,
			&entry.EnqueuedAt,
			&entry.ClosedAt,
			&entry.ApprovalStatus,
			&entry.RequestedAt,
			&entry.RequestedByUserID,
			&entry.DecidedAt,
			&entry.DecidedByUserID,
			&entry.DocumentID,
			&entry.DocumentTypeCode,
			&entry.DocumentTitle,
			&entry.DocumentNumber,
			&entry.DocumentStatus,
			&entry.JournalEntryID,
			&entry.JournalEntryNumber,
			&entry.JournalEntryPostedAt,
		); err != nil {
			return nil, fmt.Errorf("scan approval queue entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate approval queue entries: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit approval queue read: %w", err)
	}

	return entries, nil
}

func (s *Service) ListDocuments(ctx context.Context, input ListDocumentsInput) ([]DocumentReview, error) {
	if input.Status != "" && !isValidDocumentStatus(input.Status) {
		return nil, ErrInvalidReviewFilter
	}

	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
SELECT
	d.id,
	d.type_code,
	d.title,
	d.number_value,
	d.status,
	d.source_document_id,
	d.created_by_user_id,
	d.submitted_by_user_id,
	d.submitted_at,
	d.approved_at,
	d.rejected_at,
	d.created_at,
	d.updated_at,
	a.id,
	a.status,
	a.queue_code,
	a.requested_at,
	a.decided_at,
	je.id,
	je.entry_number,
	je.posted_at
FROM documents.documents d
LEFT JOIN LATERAL (
	SELECT
		id,
		status,
		queue_code,
		requested_at,
		decided_at
	FROM workflow.approvals
	WHERE org_id = d.org_id
	  AND document_id = d.id
	ORDER BY requested_at DESC, id DESC
	LIMIT 1
) a ON TRUE
LEFT JOIN accounting.journal_entries je
	ON je.org_id = d.org_id
   AND je.source_document_id = d.id
   AND je.entry_kind = 'posting'
WHERE d.org_id = $1
  AND ($2 = '' OR d.type_code = $2)
  AND ($3 = '' OR d.status = $3)
ORDER BY d.created_at DESC, d.id DESC
LIMIT $4;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.TypeCode),
		input.Status,
		normalizeLimit(input.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("query documents: %w", err)
	}
	defer rows.Close()

	var reviews []DocumentReview
	for rows.Next() {
		var review DocumentReview
		if err := rows.Scan(
			&review.DocumentID,
			&review.TypeCode,
			&review.Title,
			&review.NumberValue,
			&review.Status,
			&review.SourceDocumentID,
			&review.CreatedByUserID,
			&review.SubmittedByUserID,
			&review.SubmittedAt,
			&review.ApprovedAt,
			&review.RejectedAt,
			&review.CreatedAt,
			&review.UpdatedAt,
			&review.ApprovalID,
			&review.ApprovalStatus,
			&review.ApprovalQueueCode,
			&review.ApprovalRequestedAt,
			&review.ApprovalDecidedAt,
			&review.JournalEntryID,
			&review.JournalEntryNumber,
			&review.JournalEntryPostedAt,
		); err != nil {
			return nil, fmt.Errorf("scan document review: %w", err)
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document reviews: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit document review read: %w", err)
	}

	return reviews, nil
}

func (s *Service) ListInventoryStock(ctx context.Context, input ListInventoryStockInput) ([]InventoryStockItem, error) {
	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
SELECT
	i.id,
	i.sku,
	i.name,
	i.item_role,
	l.id,
	l.code,
	l.name,
	l.location_role,
	b.on_hand_milli
FROM (
	SELECT
		item_id,
		location_id,
		SUM(on_hand_milli) AS on_hand_milli
	FROM (
		SELECT
			item_id,
			destination_location_id AS location_id,
			SUM(quantity_milli) AS on_hand_milli
		FROM inventory_ops.movements
		WHERE org_id = $1
		  AND destination_location_id IS NOT NULL
		GROUP BY item_id, destination_location_id

		UNION ALL

		SELECT
			item_id,
			source_location_id AS location_id,
			-SUM(quantity_milli) AS on_hand_milli
		FROM inventory_ops.movements
		WHERE org_id = $1
		  AND source_location_id IS NOT NULL
		GROUP BY item_id, source_location_id
	) raw_balances
	WHERE ($2 = '' OR item_id = $2::uuid)
	  AND ($3 = '' OR location_id = $3::uuid)
	GROUP BY item_id, location_id
	HAVING $4 OR SUM(on_hand_milli) <> 0
) b
JOIN inventory_ops.items i
	ON i.id = b.item_id
   AND i.org_id = $1
JOIN inventory_ops.locations l
	ON l.id = b.location_id
   AND l.org_id = $1
ORDER BY i.sku, l.code
LIMIT $5;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.ItemID),
		strings.TrimSpace(input.LocationID),
		input.IncludeZero,
		normalizeLimit(input.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("query inventory stock review: %w", err)
	}
	defer rows.Close()

	var items []InventoryStockItem
	for rows.Next() {
		var item InventoryStockItem
		if err := rows.Scan(
			&item.ItemID,
			&item.ItemSKU,
			&item.ItemName,
			&item.ItemRole,
			&item.LocationID,
			&item.LocationCode,
			&item.LocationName,
			&item.LocationRole,
			&item.OnHandMilli,
		); err != nil {
			return nil, fmt.Errorf("scan inventory stock review: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory stock review: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit inventory stock review read: %w", err)
	}

	return items, nil
}

func (s *Service) GetWorkOrderReview(ctx context.Context, input GetWorkOrderReviewInput) (WorkOrderReview, error) {
	if strings.TrimSpace(input.WorkOrderID) == "" {
		return WorkOrderReview{}, ErrInvalidReviewFilter
	}

	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return WorkOrderReview{}, err
	}
	defer tx.Rollback()

	var review WorkOrderReview
	err = tx.QueryRowContext(ctx, `
SELECT
	wo.id,
	wo.work_order_code,
	wo.title,
	wo.summary,
	wo.status,
	wo.closed_at,
	wo.created_at,
	wo.updated_at,
	last_status.changed_at,
	COALESCE(task_counts.open_count, 0),
	COALESCE(task_counts.completed_count, 0),
	COALESCE(labor_totals.entry_count, 0),
	COALESCE(labor_totals.total_minutes, 0),
	COALESCE(labor_totals.total_cost_minor, 0),
	COALESCE(labor_posted.posted_count, 0),
	COALESCE(labor_posted.posted_cost_minor, 0),
	COALESCE(material_totals.usage_count, 0),
	COALESCE(material_totals.quantity_milli, 0),
	COALESCE(material_posted.posted_count, 0),
	COALESCE(material_posted.posted_cost_minor, 0),
	GREATEST(COALESCE(labor_posted.last_posted_at, '-infinity'::timestamptz), COALESCE(material_posted.last_posted_at, '-infinity'::timestamptz))
FROM work_orders.work_orders wo
JOIN LATERAL (
	SELECT changed_at
	FROM work_orders.status_history
	WHERE org_id = wo.org_id
	  AND work_order_id = wo.id
	ORDER BY changed_at DESC, id DESC
	LIMIT 1
) last_status ON TRUE
LEFT JOIN LATERAL (
	SELECT
		COUNT(*) FILTER (WHERE status IN ('open', 'in_progress')) AS open_count,
		COUNT(*) FILTER (WHERE status = 'completed') AS completed_count
	FROM workflow.tasks
	WHERE org_id = wo.org_id
	  AND context_type = 'work_order'
	  AND context_id = wo.id
) task_counts ON TRUE
LEFT JOIN LATERAL (
	SELECT
		COUNT(*) AS entry_count,
		COALESCE(SUM(duration_minutes), 0) AS total_minutes,
		COALESCE(SUM(cost_minor), 0) AS total_cost_minor
	FROM workforce.labor_entries
	WHERE org_id = wo.org_id
	  AND work_order_id = wo.id
) labor_totals ON TRUE
LEFT JOIN LATERAL (
	SELECT
		COUNT(*) AS posted_count,
		COALESCE(SUM(le.cost_minor), 0) AS posted_cost_minor,
		MAX(lah.posted_at) AS last_posted_at
	FROM workforce.labor_accounting_handoffs lah
	JOIN workforce.labor_entries le
		ON le.id = lah.labor_entry_id
	   AND le.org_id = lah.org_id
	WHERE lah.org_id = wo.org_id
	  AND lah.work_order_id = wo.id
	  AND lah.handoff_status = 'posted'
) labor_posted ON TRUE
LEFT JOIN LATERAL (
	SELECT
		COUNT(*) AS usage_count,
		COALESCE(SUM(quantity_milli), 0) AS quantity_milli
	FROM work_orders.material_usages
	WHERE org_id = wo.org_id
	  AND work_order_id = wo.id
) material_totals ON TRUE
LEFT JOIN LATERAL (
	SELECT
		COUNT(*) AS posted_count,
		COALESCE(SUM(ah.cost_minor), 0) AS posted_cost_minor,
		MAX(ah.posted_at) AS last_posted_at
	FROM work_orders.material_usages mu
	JOIN inventory_ops.accounting_handoffs ah
		ON ah.document_line_id = mu.inventory_document_line_id
	   AND ah.org_id = mu.org_id
	WHERE mu.org_id = wo.org_id
	  AND mu.work_order_id = wo.id
	  AND ah.handoff_status = 'posted'
) material_posted ON TRUE
WHERE wo.org_id = $1
  AND wo.id = $2;`,
		input.Actor.OrgID,
		input.WorkOrderID,
	).Scan(
		&review.WorkOrderID,
		&review.WorkOrderCode,
		&review.Title,
		&review.Summary,
		&review.Status,
		&review.ClosedAt,
		&review.CreatedAt,
		&review.UpdatedAt,
		&review.LastStatusChangedAt,
		&review.OpenTaskCount,
		&review.CompletedTaskCount,
		&review.LaborEntryCount,
		&review.TotalLaborMinutes,
		&review.TotalLaborCostMinor,
		&review.PostedLaborEntryCount,
		&review.PostedLaborCostMinor,
		&review.MaterialUsageCount,
		&review.MaterialQuantityMilli,
		&review.PostedMaterialUsageCount,
		&review.PostedMaterialCostMinor,
		&review.LastAccountingPostedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkOrderReview{}, ErrWorkOrderNotFound
		}
		return WorkOrderReview{}, fmt.Errorf("query work order review: %w", err)
	}

	if !review.LastAccountingPostedAt.Valid || review.LastAccountingPostedAt.Time.Equal(time.Unix(0, 0)) {
		review.LastAccountingPostedAt = sql.NullTime{}
	}

	if err := tx.Commit(); err != nil {
		return WorkOrderReview{}, fmt.Errorf("commit work order review read: %w", err)
	}

	return review, nil
}

func (s *Service) LookupAuditEvents(ctx context.Context, input LookupAuditEventsInput) ([]AuditEvent, error) {
	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
SELECT
	id,
	org_id,
	actor_user_id,
	event_type,
	entity_type,
	entity_id,
	payload,
	occurred_at
FROM platform.audit_events
WHERE org_id = $1
  AND ($2 = '' OR entity_type = $2)
  AND ($3 = '' OR entity_id = $3)
ORDER BY occurred_at DESC, id DESC
LIMIT $4;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.EntityType),
		strings.TrimSpace(input.EntityID),
		normalizeLimit(input.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var (
			event   AuditEvent
			payload []byte
		)
		if err := rows.Scan(
			&event.ID,
			&event.OrgID,
			&event.ActorUserID,
			&event.EventType,
			&event.EntityType,
			&event.EntityID,
			&payload,
			&event.OccurredAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		event.Payload = append(event.Payload[:0], payload...)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit events: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit audit lookup read: %w", err)
	}

	return events, nil
}

func (s *Service) beginAuthorizedRead(ctx context.Context, actor identityaccess.Actor) (*sql.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin reporting read: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, actor, identityaccess.RoleAdmin, identityaccess.RoleOperator, identityaccess.RoleApprover); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return tx, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 50
	}
	return limit
}

func isValidDocumentStatus(status string) bool {
	switch status {
	case "draft", "submitted", "approved", "rejected", "posted", "reversed", "voided":
		return true
	default:
		return false
	}
}
