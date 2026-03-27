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
	ErrDocumentNotFound    = errors.New("document not found")
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
	RequestID            sql.NullString
	RequestReference     sql.NullString
	RecommendationID     sql.NullString
	RecommendationStatus sql.NullString
	RunID                sql.NullString
	JournalEntryID       sql.NullString
	JournalEntryNumber   sql.NullInt64
	JournalEntryPostedAt sql.NullTime
}

type ListApprovalQueueInput struct {
	ApprovalID string
	QueueCode  string
	Status     string
	Limit      int
	Actor      identityaccess.Actor
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
	RequestID            sql.NullString
	RequestReference     sql.NullString
	RecommendationID     sql.NullString
	RecommendationStatus sql.NullString
	RunID                sql.NullString
	JournalEntryID       sql.NullString
	JournalEntryNumber   sql.NullInt64
	JournalEntryPostedAt sql.NullTime
}

type ListDocumentsInput struct {
	DocumentID string
	TypeCode   string
	Status     string
	Limit      int
	Actor      identityaccess.Actor
}

type GetDocumentReviewInput struct {
	DocumentID string
	Actor      identityaccess.Actor
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

type InventoryMovementReview struct {
	MovementID              string
	MovementNumber          int64
	DocumentID              sql.NullString
	DocumentTypeCode        sql.NullString
	DocumentTitle           sql.NullString
	DocumentNumber          sql.NullString
	DocumentStatus          sql.NullString
	ItemID                  string
	ItemSKU                 string
	ItemName                string
	ItemRole                string
	MovementType            string
	MovementPurpose         string
	UsageClassification     string
	SourceLocationID        sql.NullString
	SourceLocationCode      sql.NullString
	SourceLocationName      sql.NullString
	SourceLocationRole      sql.NullString
	DestinationLocationID   sql.NullString
	DestinationLocationCode sql.NullString
	DestinationLocationName sql.NullString
	DestinationLocationRole sql.NullString
	QuantityMilli           int64
	ReferenceNote           string
	CreatedByUserID         string
	CreatedAt               time.Time
}

type ListInventoryMovementsInput struct {
	MovementID   string
	ItemID       string
	LocationID   string
	DocumentID   string
	MovementType string
	Limit        int
	Actor        identityaccess.Actor
}

type InventoryReconciliationItem struct {
	DocumentID              string
	DocumentTypeCode        string
	DocumentTitle           string
	DocumentNumber          sql.NullString
	DocumentStatus          string
	DocumentLineID          string
	LineNumber              int
	MovementID              string
	MovementNumber          int64
	MovementType            string
	MovementPurpose         string
	UsageClassification     string
	ItemID                  string
	ItemSKU                 string
	ItemName                string
	ItemRole                string
	SourceLocationID        sql.NullString
	SourceLocationCode      sql.NullString
	SourceLocationName      sql.NullString
	DestinationLocationID   sql.NullString
	DestinationLocationCode sql.NullString
	DestinationLocationName sql.NullString
	QuantityMilli           int64
	ExecutionLinkID         sql.NullString
	ExecutionContextType    sql.NullString
	ExecutionContextID      sql.NullString
	ExecutionLinkStatus     sql.NullString
	WorkOrderID             sql.NullString
	WorkOrderCode           sql.NullString
	WorkOrderStatus         sql.NullString
	AccountingHandoffID     sql.NullString
	AccountingHandoffStatus sql.NullString
	CostMinor               sql.NullInt64
	CostCurrencyCode        sql.NullString
	JournalEntryID          sql.NullString
	JournalEntryNumber      sql.NullInt64
	AccountingPostedAt      sql.NullTime
	MovementCreatedAt       time.Time
}

type ListInventoryReconciliationInput struct {
	MovementID            string
	ItemID                string
	DocumentID            string
	OnlyPendingAccounting bool
	OnlyPendingExecution  bool
	Limit                 int
	Actor                 identityaccess.Actor
}

type WorkOrderReview struct {
	WorkOrderID              string
	DocumentID               string
	DocumentStatus           string
	DocumentNumber           sql.NullString
	ApprovalID               sql.NullString
	ApprovalStatus           sql.NullString
	ApprovalQueueCode        sql.NullString
	RequestID                sql.NullString
	RequestReference         sql.NullString
	RecommendationID         sql.NullString
	RecommendationStatus     sql.NullString
	RunID                    sql.NullString
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

type ListWorkOrdersInput struct {
	WorkOrderID string
	Status      string
	DocumentID  string
	Limit       int
	Actor       identityaccess.Actor
}

type JournalEntryReview struct {
	EntryID           string
	EntryNumber       int64
	EntryKind         string
	SourceDocumentID  sql.NullString
	ReversalOfEntryID sql.NullString
	CurrencyCode      string
	TaxScopeCode      string
	Summary           string
	ReversalReason    sql.NullString
	PostedByUserID    string
	EffectiveOn       time.Time
	PostedAt          time.Time
	CreatedAt         time.Time
	DocumentTypeCode  sql.NullString
	DocumentNumber    sql.NullString
	DocumentStatus    sql.NullString
	LineCount         int
	TotalDebitMinor   int64
	TotalCreditMinor  int64
	HasReversal       bool
}

type ListJournalEntriesInput struct {
	StartOn    time.Time
	EndOn      time.Time
	EntryID    string
	DocumentID string
	Limit      int
	Actor      identityaccess.Actor
}

type ControlAccountBalance struct {
	AccountID        string
	AccountCode      string
	AccountName      string
	AccountClass     string
	ControlType      string
	TotalDebitMinor  int64
	TotalCreditMinor int64
	NetMinor         int64
	LastEffectiveOn  sql.NullTime
}

type ListControlAccountBalancesInput struct {
	AsOf        time.Time
	AccountID   string
	ControlType string
	Actor       identityaccess.Actor
}

type TaxSummary struct {
	TaxType               string
	TaxCode               string
	TaxName               string
	RateBasisPoints       int
	EntryCount            int
	DocumentCount         int
	TotalDebitMinor       int64
	TotalCreditMinor      int64
	NetMinor              int64
	ReceivableAccountID   sql.NullString
	ReceivableAccountCode sql.NullString
	ReceivableAccountName sql.NullString
	PayableAccountID      sql.NullString
	PayableAccountCode    sql.NullString
	PayableAccountName    sql.NullString
	LastEffectiveOn       sql.NullTime
}

type ListTaxSummariesInput struct {
	StartOn time.Time
	EndOn   time.Time
	TaxType string
	TaxCode string
	Limit   int
	Actor   identityaccess.Actor
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
	EventID    string
	EntityType string
	EntityID   string
	Limit      int
	Actor      identityaccess.Actor
}

type InboundRequestReview struct {
	RequestID                string
	RequestReference         string
	SessionID                sql.NullString
	ActorUserID              sql.NullString
	OriginType               string
	Channel                  string
	Status                   string
	Metadata                 json.RawMessage
	CancellationReason       string
	FailureReason            string
	ReceivedAt               time.Time
	QueuedAt                 sql.NullTime
	ProcessingStartedAt      sql.NullTime
	ProcessedAt              sql.NullTime
	ActedOnAt                sql.NullTime
	CompletedAt              sql.NullTime
	FailedAt                 sql.NullTime
	CancelledAt              sql.NullTime
	CreatedAt                time.Time
	UpdatedAt                time.Time
	MessageCount             int
	AttachmentCount          int
	LastRunID                sql.NullString
	LastRunStatus            sql.NullString
	LastRecommendationID     sql.NullString
	LastRecommendationStatus sql.NullString
}

type ListInboundRequestsInput struct {
	Status           string
	RequestReference string
	Limit            int
	Actor            identityaccess.Actor
}

type InboundRequestStatusSummary struct {
	Status           string
	RequestCount     int
	MessageCount     int
	AttachmentCount  int
	LatestReceivedAt sql.NullTime
	LatestQueuedAt   sql.NullTime
	LatestUpdatedAt  time.Time
}

type InboundRequestMessageReview struct {
	MessageID       string
	MessageIndex    int
	MessageRole     string
	TextContent     string
	CreatedByUserID sql.NullString
	AttachmentCount int
	CreatedAt       time.Time
}

type RequestAttachmentReview struct {
	AttachmentID         string
	RequestMessageID     string
	LinkRole             string
	OriginalFileName     string
	MediaType            string
	SizeBytes            int64
	UploadedByUserID     sql.NullString
	LatestDerivedText    sql.NullString
	LatestDerivedByRunID sql.NullString
	DerivedTextCount     int
	CreatedAt            time.Time
}

type AIRunReview struct {
	RunID          string
	AgentRole      string
	CapabilityCode string
	Status         string
	Summary        string
	StartedAt      time.Time
	CompletedAt    sql.NullTime
}

type AIStepReview struct {
	StepID        string
	RunID         string
	StepIndex     int
	StepType      string
	StepTitle     string
	Status        string
	InputPayload  json.RawMessage
	OutputPayload json.RawMessage
	CreatedAt     time.Time
}

type AIDelegationReview struct {
	DelegationID        string
	ParentRunID         string
	ChildRunID          string
	RequestedByStepID   sql.NullString
	CapabilityCode      string
	Reason              string
	ChildAgentRole      string
	ChildCapabilityCode string
	ChildRunStatus      string
	CreatedAt           time.Time
}

type AIArtifactReview struct {
	ArtifactID      string
	RunID           string
	StepID          sql.NullString
	ArtifactType    string
	Title           string
	Payload         json.RawMessage
	CreatedByUserID string
	CreatedAt       time.Time
}

type AIRecommendationReview struct {
	RecommendationID   string
	RunID              string
	ArtifactID         sql.NullString
	ApprovalID         sql.NullString
	RecommendationType string
	Status             string
	Summary            string
	Payload            json.RawMessage
	CreatedByUserID    string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ProcessedProposalReview struct {
	RequestID            string
	RequestReference     string
	RequestStatus        string
	RecommendationID     string
	RunID                string
	RecommendationType   string
	RecommendationStatus string
	Summary              string
	ApprovalID           sql.NullString
	ApprovalStatus       sql.NullString
	ApprovalQueueCode    sql.NullString
	DocumentID           sql.NullString
	DocumentTypeCode     sql.NullString
	DocumentTitle        sql.NullString
	DocumentNumber       sql.NullString
	DocumentStatus       sql.NullString
	CreatedAt            time.Time
}

type ListProcessedProposalsInput struct {
	RecommendationID string
	Status           string
	RequestID        string
	RequestReference string
	Limit            int
	Actor            identityaccess.Actor
}

type ProcessedProposalStatusSummary struct {
	RecommendationStatus string
	ProposalCount        int
	RequestCount         int
	DocumentCount        int
	LatestCreatedAt      time.Time
}

type GetInboundRequestDetailInput struct {
	RequestID        string
	RequestReference string
	RunID            string
	DelegationID     string
	StepID           string
	Actor            identityaccess.Actor
}

type InboundRequestDetail struct {
	Request         InboundRequestReview
	Messages        []InboundRequestMessageReview
	Attachments     []RequestAttachmentReview
	Runs            []AIRunReview
	Steps           []AIStepReview
	Delegations     []AIDelegationReview
	Artifacts       []AIArtifactReview
	Recommendations []AIRecommendationReview
	Proposals       []ProcessedProposalReview
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
	rec.request_id,
	rec.request_reference,
	rec.recommendation_id,
	rec.recommendation_status,
	rec.run_id,
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
LEFT JOIN LATERAL (
	SELECT
		r.id AS request_id,
		r.request_reference,
		rec.id AS recommendation_id,
		rec.status AS recommendation_status,
		rec.run_id
	FROM ai.agent_recommendations rec
	JOIN ai.agent_runs ar
		ON ar.id = rec.run_id
	   AND ar.org_id = aqe.org_id
	JOIN ai.inbound_requests r
		ON r.id = ar.inbound_request_id
	   AND r.org_id = aqe.org_id
	WHERE rec.org_id = aqe.org_id
	  AND rec.approval_id = a.id
	ORDER BY rec.created_at DESC, rec.id DESC
	LIMIT 1
) rec ON TRUE
LEFT JOIN accounting.journal_entries je
	ON je.org_id = aqe.org_id
   AND je.source_document_id = d.id
   AND je.entry_kind = 'posting'
WHERE aqe.org_id = $1
  AND ($2 = '' OR aqe.approval_id = $2::uuid)
  AND ($3 = '' OR aqe.queue_code = $3)
  AND ($4 = '' OR aqe.status = $4)
ORDER BY aqe.enqueued_at DESC, aqe.id DESC
LIMIT $5;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.ApprovalID),
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
			&entry.RequestID,
			&entry.RequestReference,
			&entry.RecommendationID,
			&entry.RecommendationStatus,
			&entry.RunID,
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
	rec.request_id,
	rec.request_reference,
	rec.recommendation_id,
	rec.recommendation_status,
	rec.run_id,
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
LEFT JOIN LATERAL (
	SELECT
		r.id AS request_id,
		r.request_reference,
		rec.id AS recommendation_id,
		rec.status AS recommendation_status,
		rec.run_id
	FROM ai.agent_recommendations rec
	JOIN ai.agent_runs ar
		ON ar.id = rec.run_id
	   AND ar.org_id = d.org_id
	JOIN ai.inbound_requests r
		ON r.id = ar.inbound_request_id
	   AND r.org_id = d.org_id
	WHERE rec.org_id = d.org_id
	  AND rec.approval_id = a.id
	ORDER BY rec.created_at DESC, rec.id DESC
	LIMIT 1
) rec ON TRUE
LEFT JOIN accounting.journal_entries je
	ON je.org_id = d.org_id
   AND je.source_document_id = d.id
   AND je.entry_kind = 'posting'
WHERE d.org_id = $1
  AND ($2 = '' OR d.id::text = $2)
  AND ($3 = '' OR d.type_code = $3)
  AND ($4 = '' OR d.status = $4)
ORDER BY d.created_at DESC, d.id DESC
LIMIT $5;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.DocumentID),
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
			&review.RequestID,
			&review.RequestReference,
			&review.RecommendationID,
			&review.RecommendationStatus,
			&review.RunID,
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

func (s *Service) GetDocumentReview(ctx context.Context, input GetDocumentReviewInput) (DocumentReview, error) {
	documentID := strings.TrimSpace(input.DocumentID)
	if documentID == "" {
		return DocumentReview{}, ErrInvalidReviewFilter
	}

	reviews, err := s.ListDocuments(ctx, ListDocumentsInput{
		DocumentID: documentID,
		Limit:      1,
		Actor:      input.Actor,
	})
	if err != nil {
		return DocumentReview{}, err
	}
	if len(reviews) == 0 {
		return DocumentReview{}, ErrDocumentNotFound
	}
	return reviews[0], nil
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

func (s *Service) ListInventoryMovements(ctx context.Context, input ListInventoryMovementsInput) ([]InventoryMovementReview, error) {
	if input.MovementType != "" && input.MovementType != "receipt" && input.MovementType != "issue" && input.MovementType != "adjustment" {
		return nil, ErrInvalidReviewFilter
	}

	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
SELECT
	m.id,
	m.movement_number,
	m.document_id,
	d.type_code,
	d.title,
	d.number_value,
	d.status,
	i.id,
	i.sku,
	i.name,
	i.item_role,
	m.movement_type,
	m.movement_purpose,
	m.usage_classification,
	sl.id,
	sl.code,
	sl.name,
	sl.location_role,
	dl.id,
	dl.code,
	dl.name,
	dl.location_role,
	m.quantity_milli,
	m.reference_note,
	m.created_by_user_id,
	m.created_at
FROM inventory_ops.movements m
JOIN inventory_ops.items i
	ON i.id = m.item_id
   AND i.org_id = m.org_id
LEFT JOIN documents.documents d
	ON d.id = m.document_id
   AND d.org_id = m.org_id
LEFT JOIN inventory_ops.locations sl
	ON sl.id = m.source_location_id
   AND sl.org_id = m.org_id
LEFT JOIN inventory_ops.locations dl
	ON dl.id = m.destination_location_id
   AND dl.org_id = m.org_id
WHERE m.org_id = $1
  AND ($2 = '' OR m.id = $2::uuid)
  AND ($3 = '' OR m.item_id = $3::uuid)
  AND ($4 = '' OR m.document_id = $4::uuid)
  AND (
	$5 = ''
	OR m.movement_type = $5
  )
  AND (
	$6 = ''
	OR m.source_location_id = $6::uuid
	OR m.destination_location_id = $6::uuid
  )
ORDER BY m.created_at DESC, m.movement_number DESC
LIMIT $7;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.MovementID),
		strings.TrimSpace(input.ItemID),
		strings.TrimSpace(input.DocumentID),
		strings.TrimSpace(input.MovementType),
		strings.TrimSpace(input.LocationID),
		normalizeLimit(input.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("query inventory movement review: %w", err)
	}
	defer rows.Close()

	var reviews []InventoryMovementReview
	for rows.Next() {
		var review InventoryMovementReview
		if err := rows.Scan(
			&review.MovementID,
			&review.MovementNumber,
			&review.DocumentID,
			&review.DocumentTypeCode,
			&review.DocumentTitle,
			&review.DocumentNumber,
			&review.DocumentStatus,
			&review.ItemID,
			&review.ItemSKU,
			&review.ItemName,
			&review.ItemRole,
			&review.MovementType,
			&review.MovementPurpose,
			&review.UsageClassification,
			&review.SourceLocationID,
			&review.SourceLocationCode,
			&review.SourceLocationName,
			&review.SourceLocationRole,
			&review.DestinationLocationID,
			&review.DestinationLocationCode,
			&review.DestinationLocationName,
			&review.DestinationLocationRole,
			&review.QuantityMilli,
			&review.ReferenceNote,
			&review.CreatedByUserID,
			&review.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan inventory movement review: %w", err)
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory movement reviews: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit inventory movement review read: %w", err)
	}

	return reviews, nil
}

func (s *Service) ListInventoryReconciliation(ctx context.Context, input ListInventoryReconciliationInput) ([]InventoryReconciliationItem, error) {
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
	dl.id,
	dl.line_number,
	m.id,
	m.movement_number,
	m.movement_type,
	m.movement_purpose,
	m.usage_classification,
	i.id,
	i.sku,
	i.name,
	i.item_role,
	sl.id,
	sl.code,
	sl.name,
	dst.id,
	dst.code,
	dst.name,
	m.quantity_milli,
	el.id,
	el.execution_context_type,
	el.execution_context_id,
	el.linkage_status,
	wo.id,
	wo.work_order_code,
	wo.status,
	ah.id,
	ah.handoff_status,
	ah.cost_minor,
	ah.cost_currency_code,
	ah.journal_entry_id,
	je.entry_number,
	ah.posted_at,
	m.created_at
FROM inventory_ops.document_lines dl
JOIN inventory_ops.movements m
	ON m.id = dl.movement_id
   AND m.org_id = dl.org_id
JOIN documents.documents d
	ON d.id = dl.document_id
   AND d.org_id = dl.org_id
JOIN inventory_ops.items i
	ON i.id = dl.item_id
   AND i.org_id = dl.org_id
LEFT JOIN inventory_ops.locations sl
	ON sl.id = dl.source_location_id
   AND sl.org_id = dl.org_id
LEFT JOIN inventory_ops.locations dst
	ON dst.id = dl.destination_location_id
   AND dst.org_id = dl.org_id
LEFT JOIN inventory_ops.execution_links el
	ON el.document_line_id = dl.id
   AND el.org_id = dl.org_id
LEFT JOIN work_orders.material_usages mu
	ON mu.inventory_execution_link_id = el.id
   AND mu.org_id = dl.org_id
LEFT JOIN work_orders.work_orders wo
	ON wo.id = mu.work_order_id
   AND wo.org_id = dl.org_id
LEFT JOIN inventory_ops.accounting_handoffs ah
	ON ah.document_line_id = dl.id
   AND ah.org_id = dl.org_id
LEFT JOIN accounting.journal_entries je
	ON je.id = ah.journal_entry_id
   AND je.org_id = dl.org_id
WHERE dl.org_id = $1
  AND ($2 = '' OR m.id = $2::uuid)
  AND ($3 = '' OR dl.item_id = $3::uuid)
  AND ($4 = '' OR dl.document_id = $4::uuid)
  AND (NOT $5 OR ah.handoff_status = 'pending')
  AND (NOT $6 OR el.linkage_status = 'pending')
ORDER BY m.created_at DESC, m.movement_number DESC, dl.line_number ASC
LIMIT $7;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.MovementID),
		strings.TrimSpace(input.ItemID),
		strings.TrimSpace(input.DocumentID),
		input.OnlyPendingAccounting,
		input.OnlyPendingExecution,
		normalizeLimit(input.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("query inventory reconciliation review: %w", err)
	}
	defer rows.Close()

	var items []InventoryReconciliationItem
	for rows.Next() {
		var item InventoryReconciliationItem
		if err := rows.Scan(
			&item.DocumentID,
			&item.DocumentTypeCode,
			&item.DocumentTitle,
			&item.DocumentNumber,
			&item.DocumentStatus,
			&item.DocumentLineID,
			&item.LineNumber,
			&item.MovementID,
			&item.MovementNumber,
			&item.MovementType,
			&item.MovementPurpose,
			&item.UsageClassification,
			&item.ItemID,
			&item.ItemSKU,
			&item.ItemName,
			&item.ItemRole,
			&item.SourceLocationID,
			&item.SourceLocationCode,
			&item.SourceLocationName,
			&item.DestinationLocationID,
			&item.DestinationLocationCode,
			&item.DestinationLocationName,
			&item.QuantityMilli,
			&item.ExecutionLinkID,
			&item.ExecutionContextType,
			&item.ExecutionContextID,
			&item.ExecutionLinkStatus,
			&item.WorkOrderID,
			&item.WorkOrderCode,
			&item.WorkOrderStatus,
			&item.AccountingHandoffID,
			&item.AccountingHandoffStatus,
			&item.CostMinor,
			&item.CostCurrencyCode,
			&item.JournalEntryID,
			&item.JournalEntryNumber,
			&item.AccountingPostedAt,
			&item.MovementCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan inventory reconciliation review: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory reconciliation review: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit inventory reconciliation review read: %w", err)
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
	d.id,
	d.status,
	d.number_value,
	a.id,
	a.status,
	a.queue_code,
	rec.request_id,
	rec.request_reference,
	rec.recommendation_id,
	rec.recommendation_status,
	rec.run_id,
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
JOIN work_orders.documents wd
	ON wd.work_order_id = wo.id
   AND wd.org_id = wo.org_id
JOIN documents.documents d
	ON d.id = wd.document_id
   AND d.org_id = wd.org_id
LEFT JOIN LATERAL (
	SELECT
		id,
		status,
		queue_code
	FROM workflow.approvals
	WHERE org_id = wo.org_id
	  AND document_id = d.id
	ORDER BY requested_at DESC, id DESC
	LIMIT 1
) a ON TRUE
LEFT JOIN LATERAL (
	SELECT
		r.id AS request_id,
		r.request_reference,
		rec.id AS recommendation_id,
		rec.status AS recommendation_status,
		rec.run_id
	FROM ai.agent_recommendations rec
	JOIN ai.agent_runs ar
		ON ar.id = rec.run_id
	   AND ar.org_id = wo.org_id
	JOIN ai.inbound_requests r
		ON r.id = ar.inbound_request_id
	   AND r.org_id = wo.org_id
	WHERE rec.org_id = wo.org_id
	  AND rec.approval_id = a.id
	ORDER BY rec.created_at DESC, rec.id DESC
	LIMIT 1
) rec ON TRUE
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
		&review.DocumentID,
		&review.DocumentStatus,
		&review.DocumentNumber,
		&review.ApprovalID,
		&review.ApprovalStatus,
		&review.ApprovalQueueCode,
		&review.RequestID,
		&review.RequestReference,
		&review.RecommendationID,
		&review.RecommendationStatus,
		&review.RunID,
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

func (s *Service) ListWorkOrders(ctx context.Context, input ListWorkOrdersInput) ([]WorkOrderReview, error) {
	if input.Status != "" && input.Status != "open" && input.Status != "in_progress" && input.Status != "completed" && input.Status != "cancelled" {
		return nil, ErrInvalidReviewFilter
	}

	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
SELECT
	wo.id,
	d.id,
	d.status,
	d.number_value,
	a.id,
	a.status,
	a.queue_code,
	rec.request_id,
	rec.request_reference,
	rec.recommendation_id,
	rec.recommendation_status,
	rec.run_id,
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
JOIN work_orders.documents wd
	ON wd.work_order_id = wo.id
   AND wd.org_id = wo.org_id
JOIN documents.documents d
	ON d.id = wd.document_id
   AND d.org_id = wd.org_id
LEFT JOIN LATERAL (
	SELECT
		id,
		status,
		queue_code
	FROM workflow.approvals
	WHERE org_id = wo.org_id
	  AND document_id = d.id
	ORDER BY requested_at DESC, id DESC
	LIMIT 1
) a ON TRUE
LEFT JOIN LATERAL (
	SELECT
		r.id AS request_id,
		r.request_reference,
		rec.id AS recommendation_id,
		rec.status AS recommendation_status,
		rec.run_id
	FROM ai.agent_recommendations rec
	JOIN ai.agent_runs ar
		ON ar.id = rec.run_id
	   AND ar.org_id = wo.org_id
	JOIN ai.inbound_requests r
		ON r.id = ar.inbound_request_id
	   AND r.org_id = wo.org_id
	WHERE rec.org_id = wo.org_id
	  AND rec.approval_id = a.id
	ORDER BY rec.created_at DESC, rec.id DESC
	LIMIT 1
) rec ON TRUE
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
  AND ($2 = '' OR wo.id::text = $2)
  AND ($3 = '' OR wo.status = $3)
  AND ($4 = '' OR d.id::text = $4)
ORDER BY wo.updated_at DESC, wo.work_order_code ASC
LIMIT $5;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.WorkOrderID),
		strings.TrimSpace(input.Status),
		strings.TrimSpace(input.DocumentID),
		normalizeLimit(input.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("query work order reviews: %w", err)
	}
	defer rows.Close()

	var reviews []WorkOrderReview
	for rows.Next() {
		var review WorkOrderReview
		if err := rows.Scan(
			&review.WorkOrderID,
			&review.DocumentID,
			&review.DocumentStatus,
			&review.DocumentNumber,
			&review.ApprovalID,
			&review.ApprovalStatus,
			&review.ApprovalQueueCode,
			&review.RequestID,
			&review.RequestReference,
			&review.RecommendationID,
			&review.RecommendationStatus,
			&review.RunID,
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
		); err != nil {
			return nil, fmt.Errorf("scan work order review: %w", err)
		}
		if !review.LastAccountingPostedAt.Valid || review.LastAccountingPostedAt.Time.Equal(time.Unix(0, 0)) {
			review.LastAccountingPostedAt = sql.NullTime{}
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate work order reviews: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit work order review read: %w", err)
	}

	return reviews, nil
}

func (s *Service) ListJournalEntries(ctx context.Context, input ListJournalEntriesInput) ([]JournalEntryReview, error) {
	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	startOn, startSet := normalizeOptionalDate(input.StartOn)
	endOn, endSet := normalizeOptionalDate(input.EndOn)

	rows, err := tx.QueryContext(ctx, `
SELECT
	e.id,
	e.entry_number,
	e.entry_kind,
	e.source_document_id,
	e.reversal_of_entry_id,
	e.currency_code,
	e.tax_scope_code,
	e.summary,
	e.reversal_reason,
	e.posted_by_user_id,
	e.effective_on,
	e.posted_at,
	e.created_at,
	d.type_code,
	d.number_value,
	d.status,
	COUNT(l.id) AS line_count,
	COALESCE(SUM(l.debit_minor), 0) AS total_debit_minor,
	COALESCE(SUM(l.credit_minor), 0) AS total_credit_minor,
	EXISTS (
		SELECT 1
		FROM accounting.journal_entries reversals
		WHERE reversals.org_id = e.org_id
		  AND reversals.reversal_of_entry_id = e.id
	) AS has_reversal
FROM accounting.journal_entries e
JOIN accounting.journal_lines l
	ON l.entry_id = e.id
LEFT JOIN documents.documents d
	ON d.org_id = e.org_id
   AND d.id = e.source_document_id
WHERE e.org_id = $1
  AND ($2::date IS NULL OR e.effective_on >= $2::date)
  AND ($3::date IS NULL OR e.effective_on <= $3::date)
  AND ($4 = '' OR e.id = $4::uuid)
  AND ($5 = '' OR e.source_document_id = $5::uuid)
GROUP BY
	e.id,
	e.entry_number,
	e.entry_kind,
	e.source_document_id,
	e.reversal_of_entry_id,
	e.currency_code,
	e.tax_scope_code,
	e.summary,
	e.reversal_reason,
	e.posted_by_user_id,
	e.effective_on,
	e.posted_at,
	e.created_at,
	d.type_code,
	d.number_value,
	d.status
ORDER BY e.effective_on DESC, e.entry_number DESC
LIMIT $6;`,
		input.Actor.OrgID,
		nullableDate(startOn, startSet),
		nullableDate(endOn, endSet),
		strings.TrimSpace(input.EntryID),
		strings.TrimSpace(input.DocumentID),
		normalizeLimit(input.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("query journal entry review: %w", err)
	}
	defer rows.Close()

	var reviews []JournalEntryReview
	for rows.Next() {
		var review JournalEntryReview
		if err := rows.Scan(
			&review.EntryID,
			&review.EntryNumber,
			&review.EntryKind,
			&review.SourceDocumentID,
			&review.ReversalOfEntryID,
			&review.CurrencyCode,
			&review.TaxScopeCode,
			&review.Summary,
			&review.ReversalReason,
			&review.PostedByUserID,
			&review.EffectiveOn,
			&review.PostedAt,
			&review.CreatedAt,
			&review.DocumentTypeCode,
			&review.DocumentNumber,
			&review.DocumentStatus,
			&review.LineCount,
			&review.TotalDebitMinor,
			&review.TotalCreditMinor,
			&review.HasReversal,
		); err != nil {
			return nil, fmt.Errorf("scan journal entry review: %w", err)
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate journal entry reviews: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit journal entry review read: %w", err)
	}

	return reviews, nil
}

func (s *Service) ListControlAccountBalances(ctx context.Context, input ListControlAccountBalancesInput) ([]ControlAccountBalance, error) {
	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	asOf, asOfSet := normalizeOptionalDate(input.AsOf)

	rows, err := tx.QueryContext(ctx, `
SELECT
	a.id,
	a.code,
	a.name,
	a.account_class,
	a.control_type,
	COALESCE(SUM(l.debit_minor) FILTER (WHERE $2::date IS NULL OR e.effective_on <= $2::date), 0) AS total_debit_minor,
	COALESCE(SUM(l.credit_minor) FILTER (WHERE $2::date IS NULL OR e.effective_on <= $2::date), 0) AS total_credit_minor,
	MAX(e.effective_on) FILTER (WHERE $2::date IS NULL OR e.effective_on <= $2::date) AS last_effective_on
FROM accounting.ledger_accounts a
LEFT JOIN accounting.journal_lines l
	ON l.account_id = a.id
   AND l.org_id = a.org_id
LEFT JOIN accounting.journal_entries e
	ON e.id = l.entry_id
   AND e.org_id = a.org_id
WHERE a.org_id = $1
  AND a.status = 'active'
  AND a.control_type <> 'none'
  AND ($3 = '' OR a.id = $3::uuid)
  AND ($4 = '' OR a.control_type = $4)
GROUP BY a.id, a.code, a.name, a.account_class, a.control_type
ORDER BY a.code ASC;`,
		input.Actor.OrgID,
		nullableDate(asOf, asOfSet),
		strings.TrimSpace(input.AccountID),
		strings.TrimSpace(input.ControlType),
	)
	if err != nil {
		return nil, fmt.Errorf("query control account balances: %w", err)
	}
	defer rows.Close()

	var balances []ControlAccountBalance
	for rows.Next() {
		var balance ControlAccountBalance
		if err := rows.Scan(
			&balance.AccountID,
			&balance.AccountCode,
			&balance.AccountName,
			&balance.AccountClass,
			&balance.ControlType,
			&balance.TotalDebitMinor,
			&balance.TotalCreditMinor,
			&balance.LastEffectiveOn,
		); err != nil {
			return nil, fmt.Errorf("scan control account balance: %w", err)
		}
		balance.NetMinor = balance.TotalDebitMinor - balance.TotalCreditMinor
		balances = append(balances, balance)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate control account balances: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit control account balance read: %w", err)
	}

	return balances, nil
}

func (s *Service) ListTaxSummaries(ctx context.Context, input ListTaxSummariesInput) ([]TaxSummary, error) {
	if input.TaxType != "" && input.TaxType != "gst" && input.TaxType != "tds" {
		return nil, ErrInvalidReviewFilter
	}

	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	startOn, startSet := normalizeOptionalDate(input.StartOn)
	endOn, endSet := normalizeOptionalDate(input.EndOn)

	rows, err := tx.QueryContext(ctx, `
SELECT
	tc.tax_type,
	tc.code,
	tc.name,
	tc.rate_basis_points,
	COUNT(DISTINCT e.id) AS entry_count,
	COUNT(DISTINCT e.source_document_id) FILTER (WHERE e.source_document_id IS NOT NULL) AS document_count,
	COALESCE(SUM(l.debit_minor), 0) AS total_debit_minor,
	COALESCE(SUM(l.credit_minor), 0) AS total_credit_minor,
	ra.id,
	ra.code,
	ra.name,
	pa.id,
	pa.code,
	pa.name,
	MAX(e.effective_on) AS last_effective_on
FROM accounting.tax_codes tc
LEFT JOIN accounting.ledger_accounts ra
	ON ra.id = tc.receivable_account_id
   AND ra.org_id = tc.org_id
LEFT JOIN accounting.ledger_accounts pa
	ON pa.id = tc.payable_account_id
   AND pa.org_id = tc.org_id
LEFT JOIN accounting.journal_lines l
	ON l.org_id = tc.org_id
   AND l.tax_code = tc.code
LEFT JOIN accounting.journal_entries e
	ON e.id = l.entry_id
   AND e.org_id = tc.org_id
   AND ($2::date IS NULL OR e.effective_on >= $2::date)
   AND ($3::date IS NULL OR e.effective_on <= $3::date)
WHERE tc.org_id = $1
  AND tc.status = 'active'
  AND ($4 = '' OR tc.tax_type = $4)
  AND ($5 = '' OR tc.code = $5)
GROUP BY
	tc.tax_type,
	tc.code,
	tc.name,
	tc.rate_basis_points,
	ra.id,
	ra.code,
	ra.name,
	pa.id,
	pa.code,
	pa.name
ORDER BY tc.tax_type ASC, tc.code ASC
LIMIT $6;`,
		input.Actor.OrgID,
		nullableDate(startOn, startSet),
		nullableDate(endOn, endSet),
		input.TaxType,
		strings.TrimSpace(input.TaxCode),
		normalizeLimit(input.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("query tax summaries: %w", err)
	}
	defer rows.Close()

	var summaries []TaxSummary
	for rows.Next() {
		var summary TaxSummary
		if err := rows.Scan(
			&summary.TaxType,
			&summary.TaxCode,
			&summary.TaxName,
			&summary.RateBasisPoints,
			&summary.EntryCount,
			&summary.DocumentCount,
			&summary.TotalDebitMinor,
			&summary.TotalCreditMinor,
			&summary.ReceivableAccountID,
			&summary.ReceivableAccountCode,
			&summary.ReceivableAccountName,
			&summary.PayableAccountID,
			&summary.PayableAccountCode,
			&summary.PayableAccountName,
			&summary.LastEffectiveOn,
		); err != nil {
			return nil, fmt.Errorf("scan tax summary: %w", err)
		}
		summary.NetMinor = summary.TotalDebitMinor - summary.TotalCreditMinor
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tax summaries: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tax summary read: %w", err)
	}

	return summaries, nil
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
  AND ($2 = '' OR id::text = $2)
  AND ($3 = '' OR entity_type = $3)
  AND ($4 = '' OR entity_id = $4)
ORDER BY occurred_at DESC, id DESC
LIMIT $5;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.EventID),
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

func (s *Service) ListInboundRequests(ctx context.Context, input ListInboundRequestsInput) ([]InboundRequestReview, error) {
	if input.Status != "" && !isValidInboundRequestStatus(input.Status) {
		return nil, ErrInvalidReviewFilter
	}

	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}

	const query = `
SELECT
	r.id,
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
	r.updated_at,
	COALESCE(msg.message_count, 0),
	COALESCE(att.attachment_count, 0),
	lr.run_id,
	lr.run_status,
	lrec.recommendation_id,
	lrec.recommendation_status
FROM ai.inbound_requests r
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS message_count
	FROM ai.inbound_request_messages m
	WHERE m.org_id = r.org_id
	  AND m.request_id = r.id
) msg ON TRUE
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS attachment_count
	FROM ai.inbound_request_messages m
	JOIN attachments.request_message_links l
	  ON l.org_id = m.org_id
	 AND l.request_message_id = m.id
	WHERE m.org_id = r.org_id
	  AND m.request_id = r.id
) att ON TRUE
LEFT JOIN LATERAL (
	SELECT ar.id AS run_id, ar.status AS run_status
	FROM ai.agent_runs ar
	WHERE ar.org_id = r.org_id
	  AND ar.inbound_request_id = r.id
	ORDER BY ar.started_at DESC, ar.id DESC
	LIMIT 1
) lr ON TRUE
LEFT JOIN LATERAL (
	SELECT rec.id AS recommendation_id, rec.status AS recommendation_status
	FROM ai.agent_runs ar
	JOIN ai.agent_recommendations rec
	  ON rec.run_id = ar.id
	WHERE ar.org_id = r.org_id
	  AND ar.inbound_request_id = r.id
	ORDER BY rec.created_at DESC, rec.id DESC
	LIMIT 1
) lrec ON TRUE
WHERE r.org_id = $1
  AND ($2 = '' OR r.status = $2)
  AND ($3 = '' OR r.request_reference = $3)
ORDER BY COALESCE(r.queued_at, r.received_at) DESC, r.id DESC
LIMIT $4;`

	rows, err := tx.QueryContext(ctx, query, input.Actor.OrgID, strings.TrimSpace(input.Status), strings.TrimSpace(input.RequestReference), normalizeLimit(input.Limit))
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("list inbound requests: %w", err)
	}
	defer rows.Close()

	var reviews []InboundRequestReview
	for rows.Next() {
		var (
			review   InboundRequestReview
			metadata []byte
		)
		if err := rows.Scan(
			&review.RequestID,
			&review.RequestReference,
			&review.SessionID,
			&review.ActorUserID,
			&review.OriginType,
			&review.Channel,
			&review.Status,
			&metadata,
			&review.CancellationReason,
			&review.FailureReason,
			&review.ReceivedAt,
			&review.QueuedAt,
			&review.ProcessingStartedAt,
			&review.ProcessedAt,
			&review.ActedOnAt,
			&review.CompletedAt,
			&review.FailedAt,
			&review.CancelledAt,
			&review.CreatedAt,
			&review.UpdatedAt,
			&review.MessageCount,
			&review.AttachmentCount,
			&review.LastRunID,
			&review.LastRunStatus,
			&review.LastRecommendationID,
			&review.LastRecommendationStatus,
		); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("scan inbound request review: %w", err)
		}
		review.Metadata = append(review.Metadata[:0], metadata...)
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("iterate inbound request reviews: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit inbound request review read: %w", err)
	}

	return reviews, nil
}

func (s *Service) GetInboundRequestDetail(ctx context.Context, input GetInboundRequestDetailInput) (InboundRequestDetail, error) {
	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return InboundRequestDetail{}, err
	}

	requestID, err := normalizeInboundRequestLookupTx(
		ctx,
		tx,
		input.Actor.OrgID,
		input.RequestID,
		input.RequestReference,
		input.RunID,
		input.DelegationID,
		input.StepID,
	)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, err
	}
	if requestID == "" {
		_ = tx.Rollback()
		return InboundRequestDetail{}, ErrInvalidReviewFilter
	}

	requests, err := s.listInboundRequestsTx(ctx, tx, input.Actor.OrgID, requestID)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, err
	}
	if len(requests) == 0 {
		_ = tx.Rollback()
		return InboundRequestDetail{}, sql.ErrNoRows
	}

	detail := InboundRequestDetail{Request: requests[0]}

	const messagesQuery = `
SELECT
	m.id,
	m.message_index,
	m.message_role,
	m.text_content,
	m.created_by_user_id,
	COUNT(l.id) AS attachment_count,
	m.created_at
FROM ai.inbound_request_messages m
LEFT JOIN attachments.request_message_links l
	ON l.org_id = m.org_id
 AND l.request_message_id = m.id
WHERE m.org_id = $1
  AND m.request_id = $2
GROUP BY m.id, m.message_index, m.message_role, m.text_content, m.created_at
ORDER BY m.message_index ASC;`

	messageRows, err := tx.QueryContext(ctx, messagesQuery, input.Actor.OrgID, requestID)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("query inbound request messages: %w", err)
	}
	for messageRows.Next() {
		var message InboundRequestMessageReview
		if err := messageRows.Scan(
			&message.MessageID,
			&message.MessageIndex,
			&message.MessageRole,
			&message.TextContent,
			&message.CreatedByUserID,
			&message.AttachmentCount,
			&message.CreatedAt,
		); err != nil {
			messageRows.Close()
			_ = tx.Rollback()
			return InboundRequestDetail{}, fmt.Errorf("scan inbound request message review: %w", err)
		}
		detail.Messages = append(detail.Messages, message)
	}
	if err := messageRows.Err(); err != nil {
		messageRows.Close()
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("iterate inbound request messages: %w", err)
	}
	messageRows.Close()

	const attachmentsQuery = `
SELECT
	a.id,
	l.request_message_id,
	l.link_role,
	a.original_file_name,
	a.media_type,
	a.size_bytes,
	a.uploaded_by_user_id,
	dt.latest_text,
	dt.latest_created_by_run_id,
	COALESCE(dt.derived_count, 0),
	a.created_at
FROM ai.inbound_request_messages m
JOIN attachments.request_message_links l
	ON l.org_id = m.org_id
 AND l.request_message_id = m.id
JOIN attachments.attachments a
	ON a.org_id = l.org_id
 AND a.id = l.attachment_id
LEFT JOIN LATERAL (
	SELECT
		COUNT(*) AS derived_count,
		(
			SELECT content_text
			FROM attachments.derived_texts dt2
			WHERE dt2.org_id = a.org_id
			  AND dt2.source_attachment_id = a.id
			ORDER BY dt2.created_at DESC, dt2.id DESC
			LIMIT 1
		) AS latest_text,
		(
			SELECT created_by_run_id
			FROM attachments.derived_texts dt2
			WHERE dt2.org_id = a.org_id
			  AND dt2.source_attachment_id = a.id
			ORDER BY dt2.created_at DESC, dt2.id DESC
			LIMIT 1
		) AS latest_created_by_run_id
	FROM attachments.derived_texts dt
	WHERE dt.org_id = a.org_id
	  AND dt.source_attachment_id = a.id
) dt ON TRUE
WHERE m.org_id = $1
  AND m.request_id = $2
ORDER BY a.created_at ASC, a.id ASC;`

	attachmentRows, err := tx.QueryContext(ctx, attachmentsQuery, input.Actor.OrgID, requestID)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("query inbound request attachments: %w", err)
	}
	for attachmentRows.Next() {
		var attachment RequestAttachmentReview
		if err := attachmentRows.Scan(
			&attachment.AttachmentID,
			&attachment.RequestMessageID,
			&attachment.LinkRole,
			&attachment.OriginalFileName,
			&attachment.MediaType,
			&attachment.SizeBytes,
			&attachment.UploadedByUserID,
			&attachment.LatestDerivedText,
			&attachment.LatestDerivedByRunID,
			&attachment.DerivedTextCount,
			&attachment.CreatedAt,
		); err != nil {
			attachmentRows.Close()
			_ = tx.Rollback()
			return InboundRequestDetail{}, fmt.Errorf("scan inbound request attachment review: %w", err)
		}
		detail.Attachments = append(detail.Attachments, attachment)
	}
	if err := attachmentRows.Err(); err != nil {
		attachmentRows.Close()
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("iterate inbound request attachments: %w", err)
	}
	attachmentRows.Close()

	const runsQuery = `
SELECT
	id,
	agent_role,
	capability_code,
	status,
	summary,
	started_at,
	completed_at
FROM ai.agent_runs
WHERE org_id = $1
  AND inbound_request_id = $2
ORDER BY started_at ASC, id ASC;`

	runRows, err := tx.QueryContext(ctx, runsQuery, input.Actor.OrgID, requestID)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("query inbound request runs: %w", err)
	}
	for runRows.Next() {
		var run AIRunReview
		if err := runRows.Scan(
			&run.RunID,
			&run.AgentRole,
			&run.CapabilityCode,
			&run.Status,
			&run.Summary,
			&run.StartedAt,
			&run.CompletedAt,
		); err != nil {
			runRows.Close()
			_ = tx.Rollback()
			return InboundRequestDetail{}, fmt.Errorf("scan inbound request run review: %w", err)
		}
		detail.Runs = append(detail.Runs, run)
	}
	if err := runRows.Err(); err != nil {
		runRows.Close()
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("iterate inbound request runs: %w", err)
	}
	runRows.Close()

	const stepsQuery = `
SELECT
	st.id,
	st.run_id,
	st.step_index,
	st.step_type,
	st.step_title,
	st.status,
	st.input_payload,
	st.output_payload,
	st.created_at
FROM ai.agent_runs ar
JOIN ai.agent_run_steps st
	ON st.run_id = ar.id
WHERE ar.org_id = $1
  AND ar.inbound_request_id = $2
ORDER BY st.created_at ASC, st.run_id ASC, st.step_index ASC;`

	stepRows, err := tx.QueryContext(ctx, stepsQuery, input.Actor.OrgID, requestID)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("query inbound request steps: %w", err)
	}
	for stepRows.Next() {
		var (
			step          AIStepReview
			inputPayload  []byte
			outputPayload []byte
		)
		if err := stepRows.Scan(
			&step.StepID,
			&step.RunID,
			&step.StepIndex,
			&step.StepType,
			&step.StepTitle,
			&step.Status,
			&inputPayload,
			&outputPayload,
			&step.CreatedAt,
		); err != nil {
			stepRows.Close()
			_ = tx.Rollback()
			return InboundRequestDetail{}, fmt.Errorf("scan inbound request step review: %w", err)
		}
		step.InputPayload = append(step.InputPayload[:0], inputPayload...)
		step.OutputPayload = append(step.OutputPayload[:0], outputPayload...)
		detail.Steps = append(detail.Steps, step)
	}
	if err := stepRows.Err(); err != nil {
		stepRows.Close()
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("iterate inbound request steps: %w", err)
	}
	stepRows.Close()

	const delegationsQuery = `
SELECT
	d.id,
	d.parent_run_id,
	d.child_run_id,
	d.requested_by_step_id,
	d.capability_code,
	d.reason,
	child.agent_role,
	child.capability_code,
	child.status,
	d.created_at
FROM ai.agent_runs parent
JOIN ai.agent_delegations d
	ON d.parent_run_id = parent.id
JOIN ai.agent_runs child
	ON child.id = d.child_run_id
WHERE parent.org_id = $1
  AND parent.inbound_request_id = $2
ORDER BY d.created_at ASC, d.id ASC;`

	delegationRows, err := tx.QueryContext(ctx, delegationsQuery, input.Actor.OrgID, requestID)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("query inbound request delegations: %w", err)
	}
	for delegationRows.Next() {
		var delegation AIDelegationReview
		if err := delegationRows.Scan(
			&delegation.DelegationID,
			&delegation.ParentRunID,
			&delegation.ChildRunID,
			&delegation.RequestedByStepID,
			&delegation.CapabilityCode,
			&delegation.Reason,
			&delegation.ChildAgentRole,
			&delegation.ChildCapabilityCode,
			&delegation.ChildRunStatus,
			&delegation.CreatedAt,
		); err != nil {
			delegationRows.Close()
			_ = tx.Rollback()
			return InboundRequestDetail{}, fmt.Errorf("scan inbound request delegation review: %w", err)
		}
		detail.Delegations = append(detail.Delegations, delegation)
	}
	if err := delegationRows.Err(); err != nil {
		delegationRows.Close()
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("iterate inbound request delegations: %w", err)
	}
	delegationRows.Close()

	const artifactsQuery = `
SELECT
	art.id,
	art.run_id,
	art.step_id,
	art.artifact_type,
	art.title,
	art.payload,
	art.created_by_user_id,
	art.created_at
FROM ai.agent_runs ar
JOIN ai.agent_artifacts art
	ON art.run_id = ar.id
WHERE ar.org_id = $1
  AND ar.inbound_request_id = $2
ORDER BY art.created_at ASC, art.id ASC;`

	artifactRows, err := tx.QueryContext(ctx, artifactsQuery, input.Actor.OrgID, requestID)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("query inbound request artifacts: %w", err)
	}
	for artifactRows.Next() {
		var (
			artifact AIArtifactReview
			payload  []byte
		)
		if err := artifactRows.Scan(
			&artifact.ArtifactID,
			&artifact.RunID,
			&artifact.StepID,
			&artifact.ArtifactType,
			&artifact.Title,
			&payload,
			&artifact.CreatedByUserID,
			&artifact.CreatedAt,
		); err != nil {
			artifactRows.Close()
			_ = tx.Rollback()
			return InboundRequestDetail{}, fmt.Errorf("scan inbound request artifact review: %w", err)
		}
		artifact.Payload = append(artifact.Payload[:0], payload...)
		detail.Artifacts = append(detail.Artifacts, artifact)
	}
	if err := artifactRows.Err(); err != nil {
		artifactRows.Close()
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("iterate inbound request artifacts: %w", err)
	}
	artifactRows.Close()

	const recommendationsQuery = `
SELECT
	rec.id,
	rec.run_id,
	rec.artifact_id,
	rec.approval_id,
	rec.recommendation_type,
	rec.status,
	rec.summary,
	rec.payload,
	rec.created_by_user_id,
	rec.created_at,
	rec.updated_at
FROM ai.agent_runs ar
JOIN ai.agent_recommendations rec
	ON rec.run_id = ar.id
WHERE ar.org_id = $1
  AND ar.inbound_request_id = $2
ORDER BY rec.created_at ASC, rec.id ASC;`

	recommendationRows, err := tx.QueryContext(ctx, recommendationsQuery, input.Actor.OrgID, requestID)
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("query inbound request recommendations: %w", err)
	}
	for recommendationRows.Next() {
		var (
			recommendation AIRecommendationReview
			payload        []byte
		)
		if err := recommendationRows.Scan(
			&recommendation.RecommendationID,
			&recommendation.RunID,
			&recommendation.ArtifactID,
			&recommendation.ApprovalID,
			&recommendation.RecommendationType,
			&recommendation.Status,
			&recommendation.Summary,
			&payload,
			&recommendation.CreatedByUserID,
			&recommendation.CreatedAt,
			&recommendation.UpdatedAt,
		); err != nil {
			recommendationRows.Close()
			_ = tx.Rollback()
			return InboundRequestDetail{}, fmt.Errorf("scan inbound request recommendation review: %w", err)
		}
		recommendation.Payload = append(recommendation.Payload[:0], payload...)
		detail.Recommendations = append(detail.Recommendations, recommendation)
	}
	if err := recommendationRows.Err(); err != nil {
		recommendationRows.Close()
		_ = tx.Rollback()
		return InboundRequestDetail{}, fmt.Errorf("iterate inbound request recommendations: %w", err)
	}
	recommendationRows.Close()

	proposals, err := s.listProcessedProposalsTx(ctx, tx, input.Actor.OrgID, requestID, "", "")
	if err != nil {
		_ = tx.Rollback()
		return InboundRequestDetail{}, err
	}
	detail.Proposals = proposals

	if err := tx.Commit(); err != nil {
		return InboundRequestDetail{}, fmt.Errorf("commit inbound request detail read: %w", err)
	}

	return detail, nil
}

func (s *Service) ListInboundRequestStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]InboundRequestStatusSummary, error) {
	tx, err := s.beginAuthorizedRead(ctx, actor)
	if err != nil {
		return nil, err
	}

	const query = `
SELECT
	r.status,
	COUNT(*) AS request_count,
	COALESCE(SUM(msg.message_count), 0) AS message_count,
	COALESCE(SUM(att.attachment_count), 0) AS attachment_count,
	MAX(r.received_at) AS latest_received_at,
	MAX(r.queued_at) AS latest_queued_at,
	MAX(r.updated_at) AS latest_updated_at
FROM ai.inbound_requests r
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS message_count
	FROM ai.inbound_request_messages m
	WHERE m.org_id = r.org_id
	  AND m.request_id = r.id
) msg ON TRUE
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS attachment_count
	FROM ai.inbound_request_messages m
	JOIN attachments.request_message_links l
	  ON l.org_id = m.org_id
	 AND l.request_message_id = m.id
	WHERE m.org_id = r.org_id
	  AND m.request_id = r.id
) att ON TRUE
WHERE r.org_id = $1
GROUP BY r.status
ORDER BY MAX(COALESCE(r.queued_at, r.received_at)) DESC, r.status ASC;`

	rows, err := tx.QueryContext(ctx, query, actor.OrgID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("list inbound request status summary: %w", err)
	}
	defer rows.Close()

	var summaries []InboundRequestStatusSummary
	for rows.Next() {
		var summary InboundRequestStatusSummary
		if err := rows.Scan(
			&summary.Status,
			&summary.RequestCount,
			&summary.MessageCount,
			&summary.AttachmentCount,
			&summary.LatestReceivedAt,
			&summary.LatestQueuedAt,
			&summary.LatestUpdatedAt,
		); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("scan inbound request status summary: %w", err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("iterate inbound request status summary: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit inbound request status summary read: %w", err)
	}

	return summaries, nil
}

func (s *Service) ListProcessedProposals(ctx context.Context, input ListProcessedProposalsInput) ([]ProcessedProposalReview, error) {
	if input.Status != "" && !isValidRecommendationStatus(input.Status) {
		return nil, ErrInvalidReviewFilter
	}

	tx, err := s.beginAuthorizedRead(ctx, input.Actor)
	if err != nil {
		return nil, err
	}

	requestID, err := normalizeInboundRequestLookupTx(ctx, tx, input.Actor.OrgID, input.RequestID, input.RequestReference, "", "", "")
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	proposals, err := s.listProcessedProposalsTx(
		ctx,
		tx,
		input.Actor.OrgID,
		requestID,
		strings.TrimSpace(input.RecommendationID),
		strings.TrimSpace(input.Status),
	)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if len(proposals) > normalizeLimit(input.Limit) {
		proposals = proposals[:normalizeLimit(input.Limit)]
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit processed proposal review read: %w", err)
	}

	return proposals, nil
}

func (s *Service) ListProcessedProposalStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]ProcessedProposalStatusSummary, error) {
	tx, err := s.beginAuthorizedRead(ctx, actor)
	if err != nil {
		return nil, err
	}

	const query = `
SELECT
	rec.status,
	COUNT(*) AS proposal_count,
	COUNT(DISTINCT r.id) AS request_count,
	COUNT(DISTINCT ap.document_id) AS document_count,
	MAX(rec.created_at) AS latest_created_at
FROM ai.inbound_requests r
JOIN ai.agent_runs ar
	ON ar.org_id = r.org_id
 AND ar.inbound_request_id = r.id
JOIN ai.agent_recommendations rec
	ON rec.run_id = ar.id
LEFT JOIN workflow.approvals ap
	ON ap.id = rec.approval_id
WHERE r.org_id = $1
GROUP BY rec.status
ORDER BY MAX(rec.created_at) DESC, rec.status ASC;`

	rows, err := tx.QueryContext(ctx, query, actor.OrgID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("list processed proposal status summary: %w", err)
	}
	defer rows.Close()

	var summaries []ProcessedProposalStatusSummary
	for rows.Next() {
		var summary ProcessedProposalStatusSummary
		if err := rows.Scan(
			&summary.RecommendationStatus,
			&summary.ProposalCount,
			&summary.RequestCount,
			&summary.DocumentCount,
			&summary.LatestCreatedAt,
		); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("scan processed proposal status summary: %w", err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("iterate processed proposal status summary: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit processed proposal status summary read: %w", err)
	}

	return summaries, nil
}

func (s *Service) listInboundRequestsTx(ctx context.Context, tx *sql.Tx, orgID, requestID string) ([]InboundRequestReview, error) {
	const query = `
SELECT
	r.id,
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
	r.updated_at,
	COALESCE(msg.message_count, 0),
	COALESCE(att.attachment_count, 0),
	lr.run_id,
	lr.run_status,
	lrec.recommendation_id,
	lrec.recommendation_status
FROM ai.inbound_requests r
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS message_count
	FROM ai.inbound_request_messages m
	WHERE m.org_id = r.org_id
	  AND m.request_id = r.id
) msg ON TRUE
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS attachment_count
	FROM ai.inbound_request_messages m
	JOIN attachments.request_message_links l
	  ON l.org_id = m.org_id
	 AND l.request_message_id = m.id
	WHERE m.org_id = r.org_id
	  AND m.request_id = r.id
) att ON TRUE
LEFT JOIN LATERAL (
	SELECT ar.id AS run_id, ar.status AS run_status
	FROM ai.agent_runs ar
	WHERE ar.org_id = r.org_id
	  AND ar.inbound_request_id = r.id
	ORDER BY ar.started_at DESC, ar.id DESC
	LIMIT 1
) lr ON TRUE
LEFT JOIN LATERAL (
	SELECT rec.id AS recommendation_id, rec.status AS recommendation_status
	FROM ai.agent_runs ar
	JOIN ai.agent_recommendations rec
	  ON rec.run_id = ar.id
	WHERE ar.org_id = r.org_id
	  AND ar.inbound_request_id = r.id
	ORDER BY rec.created_at DESC, rec.id DESC
	LIMIT 1
) lrec ON TRUE
WHERE r.org_id = $1
  AND r.id = $2;`

	rows, err := tx.QueryContext(ctx, query, orgID, requestID)
	if err != nil {
		return nil, fmt.Errorf("get inbound request detail header: %w", err)
	}
	defer rows.Close()

	var reviews []InboundRequestReview
	for rows.Next() {
		var (
			review   InboundRequestReview
			metadata []byte
		)
		if err := rows.Scan(
			&review.RequestID,
			&review.RequestReference,
			&review.SessionID,
			&review.ActorUserID,
			&review.OriginType,
			&review.Channel,
			&review.Status,
			&metadata,
			&review.CancellationReason,
			&review.FailureReason,
			&review.ReceivedAt,
			&review.QueuedAt,
			&review.ProcessingStartedAt,
			&review.ProcessedAt,
			&review.ActedOnAt,
			&review.CompletedAt,
			&review.FailedAt,
			&review.CancelledAt,
			&review.CreatedAt,
			&review.UpdatedAt,
			&review.MessageCount,
			&review.AttachmentCount,
			&review.LastRunID,
			&review.LastRunStatus,
			&review.LastRecommendationID,
			&review.LastRecommendationStatus,
		); err != nil {
			return nil, fmt.Errorf("scan inbound request detail header: %w", err)
		}
		review.Metadata = append(review.Metadata[:0], metadata...)
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inbound request detail header: %w", err)
	}
	return reviews, nil
}

func (s *Service) listProcessedProposalsTx(ctx context.Context, tx *sql.Tx, orgID, requestID, recommendationID, status string) ([]ProcessedProposalReview, error) {
	const query = `
SELECT
	r.id,
	r.request_reference,
	r.status,
	rec.id,
	rec.run_id,
	rec.recommendation_type,
	rec.status,
	rec.summary,
	rec.approval_id,
	ap.status,
	ap.queue_code,
	d.id,
	d.type_code,
	d.title,
	d.number_value,
	d.status,
	rec.created_at
FROM ai.inbound_requests r
JOIN ai.agent_runs ar
	ON ar.org_id = r.org_id
 AND ar.inbound_request_id = r.id
JOIN ai.agent_recommendations rec
	ON rec.run_id = ar.id
LEFT JOIN workflow.approvals ap
	ON ap.id = rec.approval_id
LEFT JOIN documents.documents d
	ON d.id = ap.document_id
WHERE r.org_id = $1
  AND ($2 = '' OR r.id = NULLIF($2, '')::uuid)
  AND ($3 = '' OR rec.id = NULLIF($3, '')::uuid)
  AND ($4 = '' OR rec.status = $4)
ORDER BY rec.created_at DESC, rec.id DESC
LIMIT 200;`

	rows, err := tx.QueryContext(ctx, query, orgID, requestID, recommendationID, status)
	if err != nil {
		return nil, fmt.Errorf("list processed proposals: %w", err)
	}
	defer rows.Close()

	var proposals []ProcessedProposalReview
	for rows.Next() {
		var proposal ProcessedProposalReview
		if err := rows.Scan(
			&proposal.RequestID,
			&proposal.RequestReference,
			&proposal.RequestStatus,
			&proposal.RecommendationID,
			&proposal.RunID,
			&proposal.RecommendationType,
			&proposal.RecommendationStatus,
			&proposal.Summary,
			&proposal.ApprovalID,
			&proposal.ApprovalStatus,
			&proposal.ApprovalQueueCode,
			&proposal.DocumentID,
			&proposal.DocumentTypeCode,
			&proposal.DocumentTitle,
			&proposal.DocumentNumber,
			&proposal.DocumentStatus,
			&proposal.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan processed proposal review: %w", err)
		}
		proposals = append(proposals, proposal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate processed proposal reviews: %w", err)
	}

	return proposals, nil
}

func normalizeInboundRequestLookupTx(ctx context.Context, tx *sql.Tx, orgID, requestID, requestReference, runID, delegationID, stepID string) (string, error) {
	trimmedID := strings.TrimSpace(requestID)
	trimmedReference := strings.TrimSpace(requestReference)
	trimmedRunID := strings.TrimSpace(runID)
	trimmedDelegationID := strings.TrimSpace(delegationID)
	trimmedStepID := strings.TrimSpace(stepID)
	provided := 0
	for _, value := range []string{trimmedID, trimmedReference, trimmedRunID, trimmedDelegationID, trimmedStepID} {
		if value != "" {
			provided++
		}
	}
	switch {
	case provided > 1:
		return "", ErrInvalidReviewFilter
	case trimmedID != "":
		return trimmedID, nil
	case trimmedReference != "":
		var resolvedID string
		err := tx.QueryRowContext(ctx, `
SELECT id
FROM ai.inbound_requests
WHERE org_id = $1
  AND request_reference = $2;`, orgID, trimmedReference).Scan(&resolvedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", sql.ErrNoRows
			}
			return "", fmt.Errorf("resolve inbound request reference: %w", err)
		}
		return resolvedID, nil
	case trimmedRunID != "":
		var resolvedID string
		err := tx.QueryRowContext(ctx, `
SELECT inbound_request_id
FROM ai.agent_runs
WHERE org_id = $1
  AND id = $2;`, orgID, trimmedRunID).Scan(&resolvedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", sql.ErrNoRows
			}
			return "", fmt.Errorf("resolve inbound request run: %w", err)
		}
		return resolvedID, nil
	case trimmedDelegationID != "":
		var resolvedID string
		err := tx.QueryRowContext(ctx, `
SELECT parent.inbound_request_id
FROM ai.agent_delegations d
JOIN ai.agent_runs parent
  ON parent.id = d.parent_run_id
WHERE parent.org_id = $1
  AND d.id = $2;`, orgID, trimmedDelegationID).Scan(&resolvedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", sql.ErrNoRows
			}
			return "", fmt.Errorf("resolve inbound request delegation: %w", err)
		}
		return resolvedID, nil
	case trimmedStepID != "":
		var resolvedID string
		err := tx.QueryRowContext(ctx, `
SELECT ar.inbound_request_id
FROM ai.agent_run_steps st
JOIN ai.agent_runs ar
  ON ar.id = st.run_id
WHERE ar.org_id = $1
  AND st.id = $2;`, orgID, trimmedStepID).Scan(&resolvedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", sql.ErrNoRows
			}
			return "", fmt.Errorf("resolve inbound request step: %w", err)
		}
		return resolvedID, nil
	default:
		return "", nil
	}
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

func normalizeOptionalDate(value time.Time) (time.Time, bool) {
	if value.IsZero() {
		return time.Time{}, false
	}
	return value.UTC().Truncate(24 * time.Hour), true
}

func nullableDate(value time.Time, set bool) any {
	if !set {
		return nil
	}
	return value.Format(time.DateOnly)
}

func isValidDocumentStatus(status string) bool {
	switch status {
	case "draft", "submitted", "approved", "rejected", "posted", "reversed", "voided":
		return true
	default:
		return false
	}
}

func isValidInboundRequestStatus(status string) bool {
	switch status {
	case "draft", "queued", "processing", "processed", "acted_on", "completed", "failed", "cancelled":
		return true
	default:
		return false
	}
}

func isValidRecommendationStatus(status string) bool {
	switch status {
	case "proposed", "approval_requested", "accepted", "rejected":
		return true
	default:
		return false
	}
}
