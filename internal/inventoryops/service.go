package inventoryops

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
	ErrItemNotFound        = errors.New("inventory item not found")
	ErrLocationNotFound    = errors.New("inventory location not found")
	ErrMovementNotFound    = errors.New("inventory movement not found")
	ErrInvalidItem         = errors.New("invalid inventory item")
	ErrInvalidLocation     = errors.New("invalid inventory location")
	ErrInvalidMovement     = errors.New("invalid inventory movement")
	ErrInvalidInventoryDoc = errors.New("invalid inventory document")
	ErrInventoryDocExists  = errors.New("inventory document payload already exists")
	ErrInsufficientStock   = errors.New("insufficient stock")
)

const (
	ItemRoleResale                  = "resale"
	ItemRoleServiceMaterial         = "service_material"
	ItemRoleTraceableEquipment      = "traceable_equipment"
	ItemRoleDirectExpenseConsumable = "direct_expense_consumable"

	TrackingModeNone   = "none"
	TrackingModeSerial = "serial"
	TrackingModeLot    = "lot"

	LocationRoleWarehouse  = "warehouse"
	LocationRoleVan        = "van"
	LocationRoleSite       = "site"
	LocationRoleVendor     = "vendor"
	LocationRoleCustomer   = "customer"
	LocationRoleAdjustment = "adjustment"
	LocationRoleInstalled  = "installed"

	MovementTypeReceipt    = "receipt"
	MovementTypeIssue      = "issue"
	MovementTypeAdjustment = "adjustment"

	MovementPurposeResale             = "resale"
	MovementPurposeServiceConsumption = "service_consumption"
	MovementPurposeInstalledEquipment = "installed_equipment"
	MovementPurposeDirectExpense      = "direct_expense"
	MovementPurposeStockAdjustment    = "stock_adjustment"

	UsageNotApplicable = "not_applicable"
	UsageBillable      = "billable"
	UsageNonBillable   = "non_billable"

	AccountingHandoffStatusPending = "pending"
	AccountingHandoffStatusPosted  = "posted"

	ExecutionContextWorkOrder = "work_order"
	ExecutionContextProject   = "project"

	ExecutionLinkStatusPending = "pending"
	ExecutionLinkStatusLinked  = "linked"
)

type Item struct {
	ID              string
	OrgID           string
	SKU             string
	Name            string
	ItemRole        string
	TrackingMode    string
	Status          string
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Location struct {
	ID              string
	OrgID           string
	Code            string
	Name            string
	LocationRole    string
	Status          string
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Movement struct {
	ID                    string
	OrgID                 string
	MovementNumber        int64
	DocumentID            sql.NullString
	ItemID                string
	MovementType          string
	MovementPurpose       string
	UsageClassification   string
	SourceLocationID      sql.NullString
	DestinationLocationID sql.NullString
	QuantityMilli         int64
	ReferenceNote         string
	CreatedByUserID       string
	CreatedAt             time.Time
}

type StockBalance struct {
	ItemID      string
	LocationID  string
	OnHandMilli int64
}

type Document struct {
	DocumentID      string
	OrgID           string
	MovementType    string
	ReferenceNote   string
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DocumentLine struct {
	ID                    string
	DocumentID            string
	OrgID                 string
	LineNumber            int
	MovementID            string
	ItemID                string
	MovementPurpose       string
	UsageClassification   string
	SourceLocationID      sql.NullString
	DestinationLocationID sql.NullString
	QuantityMilli         int64
	ReferenceNote         string
	CreatedAt             time.Time
}

type AccountingHandoff struct {
	ID              string
	OrgID           string
	DocumentID      string
	DocumentLineID  string
	JournalEntryID  sql.NullString
	HandoffStatus   string
	CreatedByUserID string
	CreatedAt       time.Time
}

type ExecutionLink struct {
	ID                   string
	OrgID                string
	DocumentID           string
	DocumentLineID       string
	ExecutionContextType string
	ExecutionContextID   string
	LinkageStatus        string
	CreatedByUserID      string
	CreatedAt            time.Time
}

type CaptureDocumentResult struct {
	Document           Document
	Lines              []DocumentLine
	Movements          []Movement
	AccountingHandoffs []AccountingHandoff
	ExecutionLinks     []ExecutionLink
}

type CreateItemInput struct {
	SKU          string
	Name         string
	ItemRole     string
	TrackingMode string
	Actor        identityaccess.Actor
}

type CreateLocationInput struct {
	Code         string
	Name         string
	LocationRole string
	Actor        identityaccess.Actor
}

type RecordMovementInput struct {
	DocumentID            string
	ItemID                string
	MovementType          string
	MovementPurpose       string
	UsageClassification   string
	SourceLocationID      string
	DestinationLocationID string
	QuantityMilli         int64
	ReferenceNote         string
	Actor                 identityaccess.Actor
}

type CaptureDocumentInput struct {
	DocumentID    string
	ReferenceNote string
	Lines         []CaptureDocumentLineInput
	Actor         identityaccess.Actor
}

type CaptureDocumentLineInput struct {
	ItemID                string
	MovementPurpose       string
	UsageClassification   string
	SourceLocationID      string
	DestinationLocationID string
	QuantityMilli         int64
	ReferenceNote         string
	AccountingHandoff     bool
	ExecutionContextType  string
	ExecutionContextID    string
}

type ListStockInput struct {
	ItemID      string
	LocationID  string
	IncludeZero bool
	Actor       identityaccess.Actor
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateItem(ctx context.Context, input CreateItemInput) (Item, error) {
	if !isValidItemRole(input.ItemRole) {
		return Item{}, ErrInvalidItem
	}
	if input.TrackingMode == "" {
		input.TrackingMode = TrackingModeNone
	}
	if !isValidTrackingMode(input.TrackingMode) {
		return Item{}, ErrInvalidItem
	}
	if strings.TrimSpace(input.SKU) == "" || strings.TrimSpace(input.Name) == "" {
		return Item{}, ErrInvalidItem
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Item{}, fmt.Errorf("begin create item: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Item{}, err
	}

	item, err := scanItem(tx.QueryRowContext(ctx, `
INSERT INTO inventory_ops.items (
	org_id,
	sku,
	name,
	item_role,
	tracking_mode,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
	id,
	org_id,
	sku,
	name,
	item_role,
	tracking_mode,
	status,
	created_by_user_id,
	created_at,
	updated_at;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.SKU),
		strings.TrimSpace(input.Name),
		input.ItemRole,
		input.TrackingMode,
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return Item{}, fmt.Errorf("insert inventory item: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "inventory_ops.item_created",
		EntityType:  "inventory_ops.item",
		EntityID:    item.ID,
		Payload: map[string]any{
			"sku":           item.SKU,
			"item_role":     item.ItemRole,
			"tracking_mode": item.TrackingMode,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Item{}, err
	}

	if err := tx.Commit(); err != nil {
		return Item{}, fmt.Errorf("commit create item: %w", err)
	}

	return item, nil
}

func (s *Service) CreateLocation(ctx context.Context, input CreateLocationInput) (Location, error) {
	if !isValidLocationRole(input.LocationRole) {
		return Location{}, ErrInvalidLocation
	}
	if strings.TrimSpace(input.Code) == "" || strings.TrimSpace(input.Name) == "" {
		return Location{}, ErrInvalidLocation
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Location{}, fmt.Errorf("begin create location: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Location{}, err
	}

	location, err := scanLocation(tx.QueryRowContext(ctx, `
INSERT INTO inventory_ops.locations (
	org_id,
	code,
	name,
	location_role,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING
	id,
	org_id,
	code,
	name,
	location_role,
	status,
	created_by_user_id,
	created_at,
	updated_at;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.Code),
		strings.TrimSpace(input.Name),
		input.LocationRole,
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return Location{}, fmt.Errorf("insert inventory location: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "inventory_ops.location_created",
		EntityType:  "inventory_ops.location",
		EntityID:    location.ID,
		Payload: map[string]any{
			"code":          location.Code,
			"location_role": location.LocationRole,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Location{}, err
	}

	if err := tx.Commit(); err != nil {
		return Location{}, fmt.Errorf("commit create location: %w", err)
	}

	return location, nil
}

func (s *Service) RecordMovement(ctx context.Context, input RecordMovementInput) (Movement, error) {
	if err := validateMovementInput(input); err != nil {
		return Movement{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Movement{}, fmt.Errorf("begin record movement: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Movement{}, err
	}

	movement, err := recordMovementTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Movement{}, err
	}

	if err := tx.Commit(); err != nil {
		return Movement{}, fmt.Errorf("commit record movement: %w", err)
	}

	return movement, nil
}

func (s *Service) CaptureDocument(ctx context.Context, input CaptureDocumentInput) (CaptureDocumentResult, error) {
	if strings.TrimSpace(input.DocumentID) == "" || len(input.Lines) == 0 {
		return CaptureDocumentResult{}, ErrInvalidInventoryDoc
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CaptureDocumentResult{}, fmt.Errorf("begin capture inventory document: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return CaptureDocumentResult{}, err
	}

	movementType, err := loadInventoryDocumentMovementType(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return CaptureDocumentResult{}, err
	}

	exists, err := inventoryDocumentExists(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return CaptureDocumentResult{}, err
	}
	if exists {
		_ = tx.Rollback()
		return CaptureDocumentResult{}, ErrInventoryDocExists
	}

	document, err := scanInventoryDocument(tx.QueryRowContext(ctx, `
INSERT INTO inventory_ops.documents (
	document_id,
	org_id,
	movement_type,
	reference_note,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING
	document_id,
	org_id,
	movement_type,
	reference_note,
	created_by_user_id,
	created_at,
	updated_at;`,
		input.DocumentID,
		input.Actor.OrgID,
		movementType,
		strings.TrimSpace(input.ReferenceNote),
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return CaptureDocumentResult{}, fmt.Errorf("insert inventory document payload: %w", err)
	}

	result := CaptureDocumentResult{
		Document: document,
	}

	for idx, line := range input.Lines {
		if err := validateCaptureDocumentLine(movementType, line); err != nil {
			_ = tx.Rollback()
			return CaptureDocumentResult{}, err
		}

		movement, err := recordMovementTx(ctx, tx, RecordMovementInput{
			DocumentID:            input.DocumentID,
			ItemID:                line.ItemID,
			MovementType:          movementType,
			MovementPurpose:       line.MovementPurpose,
			UsageClassification:   line.UsageClassification,
			SourceLocationID:      line.SourceLocationID,
			DestinationLocationID: line.DestinationLocationID,
			QuantityMilli:         line.QuantityMilli,
			ReferenceNote:         line.ReferenceNote,
			Actor:                 input.Actor,
		})
		if err != nil {
			_ = tx.Rollback()
			return CaptureDocumentResult{}, err
		}

		documentLine, err := scanDocumentLine(tx.QueryRowContext(ctx, `
INSERT INTO inventory_ops.document_lines (
	document_id,
	org_id,
	line_number,
	movement_id,
	item_id,
	movement_purpose,
	usage_classification,
	source_location_id,
	destination_location_id,
	quantity_milli,
	reference_note
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING
	id,
	document_id,
	org_id,
	line_number,
	movement_id,
	item_id,
	movement_purpose,
	usage_classification,
	source_location_id,
	destination_location_id,
	quantity_milli,
	reference_note,
	created_at;`,
			input.DocumentID,
			input.Actor.OrgID,
			idx+1,
			movement.ID,
			line.ItemID,
			line.MovementPurpose,
			line.UsageClassification,
			nullIfEmpty(line.SourceLocationID),
			nullIfEmpty(line.DestinationLocationID),
			line.QuantityMilli,
			strings.TrimSpace(line.ReferenceNote),
		))
		if err != nil {
			_ = tx.Rollback()
			return CaptureDocumentResult{}, fmt.Errorf("insert inventory document line: %w", err)
		}

		result.Lines = append(result.Lines, documentLine)
		result.Movements = append(result.Movements, movement)

		if line.AccountingHandoff {
			handoff, err := scanAccountingHandoff(tx.QueryRowContext(ctx, `
INSERT INTO inventory_ops.accounting_handoffs (
	org_id,
	document_id,
	document_line_id,
	created_by_user_id
) VALUES ($1, $2, $3, $4)
RETURNING
	id,
	org_id,
	document_id,
	document_line_id,
	journal_entry_id,
	handoff_status,
	created_by_user_id,
	created_at;`,
				input.Actor.OrgID,
				input.DocumentID,
				documentLine.ID,
				input.Actor.UserID,
			))
			if err != nil {
				_ = tx.Rollback()
				return CaptureDocumentResult{}, fmt.Errorf("insert accounting handoff: %w", err)
			}
			result.AccountingHandoffs = append(result.AccountingHandoffs, handoff)
		}

		if line.ExecutionContextType != "" {
			link, err := scanExecutionLink(tx.QueryRowContext(ctx, `
INSERT INTO inventory_ops.execution_links (
	org_id,
	document_id,
	document_line_id,
	execution_context_type,
	execution_context_id,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
	id,
	org_id,
	document_id,
	document_line_id,
	execution_context_type,
	execution_context_id,
	linkage_status,
	created_by_user_id,
	created_at;`,
				input.Actor.OrgID,
				input.DocumentID,
				documentLine.ID,
				line.ExecutionContextType,
				strings.TrimSpace(line.ExecutionContextID),
				input.Actor.UserID,
			))
			if err != nil {
				_ = tx.Rollback()
				return CaptureDocumentResult{}, fmt.Errorf("insert execution link: %w", err)
			}
			result.ExecutionLinks = append(result.ExecutionLinks, link)
		}
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "inventory_ops.document_captured",
		EntityType:  "inventory_ops.document",
		EntityID:    input.DocumentID,
		Payload: map[string]any{
			"movement_type":          movementType,
			"line_count":             len(result.Lines),
			"accounting_handoff_cnt": len(result.AccountingHandoffs),
			"execution_link_cnt":     len(result.ExecutionLinks),
		},
	}); err != nil {
		_ = tx.Rollback()
		return CaptureDocumentResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return CaptureDocumentResult{}, fmt.Errorf("commit capture inventory document: %w", err)
	}

	return result, nil
}

func (s *Service) ListStock(ctx context.Context, input ListStockInput) ([]StockBalance, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list stock: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator, identityaccess.RoleApprover); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
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
) balances
WHERE ($2 = '' OR item_id = $2::uuid)
  AND ($3 = '' OR location_id = $3::uuid)
GROUP BY item_id, location_id
HAVING $4 OR SUM(on_hand_milli) <> 0
ORDER BY item_id, location_id;`,
		input.Actor.OrgID,
		input.ItemID,
		input.LocationID,
		input.IncludeZero,
	)
	if err != nil {
		return nil, fmt.Errorf("query inventory stock: %w", err)
	}
	defer rows.Close()

	var balances []StockBalance
	for rows.Next() {
		var balance StockBalance
		if err := rows.Scan(&balance.ItemID, &balance.LocationID, &balance.OnHandMilli); err != nil {
			return nil, fmt.Errorf("scan inventory stock: %w", err)
		}
		balances = append(balances, balance)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory stock: %w", err)
	}

	return balances, nil
}

func validateMovementInput(input RecordMovementInput) error {
	if input.QuantityMilli <= 0 {
		return ErrInvalidMovement
	}
	if input.ItemID == "" {
		return ErrInvalidMovement
	}
	if !isValidMovementType(input.MovementType) || !isValidMovementPurpose(input.MovementPurpose) || !isValidUsageClassification(input.UsageClassification) {
		return ErrInvalidMovement
	}
	if input.SourceLocationID == input.DestinationLocationID && input.SourceLocationID != "" {
		return ErrInvalidMovement
	}

	switch input.MovementType {
	case MovementTypeReceipt:
		if input.SourceLocationID != "" || input.DestinationLocationID == "" {
			return ErrInvalidMovement
		}
	case MovementTypeIssue:
		if input.SourceLocationID == "" || input.DestinationLocationID != "" {
			return ErrInvalidMovement
		}
	case MovementTypeAdjustment:
		if (input.SourceLocationID == "" && input.DestinationLocationID == "") || (input.SourceLocationID != "" && input.DestinationLocationID != "") {
			return ErrInvalidMovement
		}
	}

	if input.UsageClassification == UsageNotApplicable {
		if input.MovementPurpose == MovementPurposeServiceConsumption || input.MovementPurpose == MovementPurposeDirectExpense {
			return ErrInvalidMovement
		}
		return nil
	}

	if input.MovementPurpose != MovementPurposeServiceConsumption && input.MovementPurpose != MovementPurposeDirectExpense {
		return ErrInvalidMovement
	}

	return nil
}

func validatePurposeAgainstItem(itemRole, movementPurpose string) error {
	switch itemRole {
	case ItemRoleResale:
		if movementPurpose != MovementPurposeResale && movementPurpose != MovementPurposeStockAdjustment {
			return ErrInvalidMovement
		}
	case ItemRoleServiceMaterial:
		if movementPurpose != MovementPurposeServiceConsumption && movementPurpose != MovementPurposeStockAdjustment {
			return ErrInvalidMovement
		}
	case ItemRoleTraceableEquipment:
		if movementPurpose != MovementPurposeInstalledEquipment && movementPurpose != MovementPurposeStockAdjustment {
			return ErrInvalidMovement
		}
	case ItemRoleDirectExpenseConsumable:
		if movementPurpose != MovementPurposeDirectExpense && movementPurpose != MovementPurposeStockAdjustment {
			return ErrInvalidMovement
		}
	default:
		return ErrInvalidItem
	}

	return nil
}

func validateCaptureDocumentLine(movementType string, input CaptureDocumentLineInput) error {
	if err := validateMovementInput(RecordMovementInput{
		ItemID:                input.ItemID,
		MovementType:          movementType,
		MovementPurpose:       input.MovementPurpose,
		UsageClassification:   input.UsageClassification,
		SourceLocationID:      input.SourceLocationID,
		DestinationLocationID: input.DestinationLocationID,
		QuantityMilli:         input.QuantityMilli,
		ReferenceNote:         input.ReferenceNote,
	}); err != nil {
		return err
	}

	if input.ExecutionContextType == "" && strings.TrimSpace(input.ExecutionContextID) == "" {
		return nil
	}
	if !isValidExecutionContextType(input.ExecutionContextType) || strings.TrimSpace(input.ExecutionContextID) == "" {
		return ErrInvalidInventoryDoc
	}
	if movementType == MovementTypeReceipt {
		return ErrInvalidInventoryDoc
	}
	switch input.MovementPurpose {
	case MovementPurposeServiceConsumption, MovementPurposeInstalledEquipment, MovementPurposeDirectExpense:
		return nil
	default:
		return ErrInvalidInventoryDoc
	}
}

func validateInventoryDocument(ctx context.Context, tx *sql.Tx, orgID, documentID, movementType string) error {
	loadedMovementType, err := loadInventoryDocumentMovementType(ctx, tx, orgID, documentID)
	if err != nil {
		return err
	}
	if loadedMovementType != movementType {
		return ErrInvalidInventoryDoc
	}
	return nil
}

func loadInventoryDocumentMovementType(ctx context.Context, tx *sql.Tx, orgID, documentID string) (string, error) {
	var typeCode string
	var status string
	err := tx.QueryRowContext(ctx, `
SELECT type_code, status
FROM documents.documents
WHERE org_id = $1
  AND id = $2;`,
		orgID,
		documentID,
	).Scan(&typeCode, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidInventoryDoc
		}
		return "", fmt.Errorf("load inventory document: %w", err)
	}

	movementType, ok := inventoryDocumentMovementType(typeCode)
	if !ok {
		return "", ErrInvalidInventoryDoc
	}
	if status != "approved" && status != "posted" {
		return "", ErrInvalidInventoryDoc
	}

	return movementType, nil
}

func inventoryDocumentExists(ctx context.Context, tx *sql.Tx, orgID, documentID string) (bool, error) {
	var exists bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM inventory_ops.documents
	WHERE org_id = $1
	  AND document_id = $2
);`,
		orgID,
		documentID,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check inventory document payload: %w", err)
	}
	return exists, nil
}

func inventoryDocumentMovementType(typeCode string) (string, bool) {
	switch typeCode {
	case "inventory_receipt":
		return MovementTypeReceipt, true
	case "inventory_issue":
		return MovementTypeIssue, true
	case "inventory_adjustment":
		return MovementTypeAdjustment, true
	default:
		return "", false
	}
}

func recordMovementTx(ctx context.Context, tx *sql.Tx, input RecordMovementInput) (Movement, error) {
	item, err := loadItemForOrg(ctx, tx, input.Actor.OrgID, input.ItemID)
	if err != nil {
		return Movement{}, err
	}
	if item.Status != "active" {
		return Movement{}, ErrInvalidItem
	}

	if err := validatePurposeAgainstItem(item.ItemRole, input.MovementPurpose); err != nil {
		return Movement{}, err
	}

	if input.SourceLocationID != "" {
		location, err := loadLocationForOrg(ctx, tx, input.Actor.OrgID, input.SourceLocationID)
		if err != nil {
			return Movement{}, err
		}
		if location.Status != "active" {
			return Movement{}, ErrInvalidLocation
		}
	}

	if input.DestinationLocationID != "" {
		location, err := loadLocationForOrg(ctx, tx, input.Actor.OrgID, input.DestinationLocationID)
		if err != nil {
			return Movement{}, err
		}
		if location.Status != "active" {
			return Movement{}, ErrInvalidLocation
		}
	}

	if input.DocumentID != "" {
		if err := validateInventoryDocument(ctx, tx, input.Actor.OrgID, input.DocumentID, input.MovementType); err != nil {
			return Movement{}, err
		}
	}

	if input.SourceLocationID != "" {
		if err := lockStockKey(ctx, tx, input.Actor.OrgID, input.ItemID, input.SourceLocationID); err != nil {
			return Movement{}, fmt.Errorf("lock source stock: %w", err)
		}

		onHand, err := currentStockMilli(ctx, tx, input.Actor.OrgID, input.ItemID, input.SourceLocationID)
		if err != nil {
			return Movement{}, err
		}
		if onHand < input.QuantityMilli {
			return Movement{}, ErrInsufficientStock
		}
	}

	movementNumber, err := nextMovementNumber(ctx, tx, input.Actor.OrgID)
	if err != nil {
		return Movement{}, err
	}

	movement, err := scanMovement(tx.QueryRowContext(ctx, `
INSERT INTO inventory_ops.movements (
	org_id,
	movement_number,
	document_id,
	item_id,
	movement_type,
	movement_purpose,
	usage_classification,
	source_location_id,
	destination_location_id,
	quantity_milli,
	reference_note,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING
	id,
	org_id,
	movement_number,
	document_id,
	item_id,
	movement_type,
	movement_purpose,
	usage_classification,
	source_location_id,
	destination_location_id,
	quantity_milli,
	reference_note,
	created_by_user_id,
	created_at;`,
		input.Actor.OrgID,
		movementNumber,
		nullIfEmpty(input.DocumentID),
		input.ItemID,
		input.MovementType,
		input.MovementPurpose,
		input.UsageClassification,
		nullIfEmpty(input.SourceLocationID),
		nullIfEmpty(input.DestinationLocationID),
		input.QuantityMilli,
		strings.TrimSpace(input.ReferenceNote),
		input.Actor.UserID,
	))
	if err != nil {
		return Movement{}, fmt.Errorf("insert inventory movement: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "inventory_ops.movement_recorded",
		EntityType:  "inventory_ops.movement",
		EntityID:    movement.ID,
		Payload: map[string]any{
			"movement_number":         movement.MovementNumber,
			"movement_type":           movement.MovementType,
			"movement_purpose":        movement.MovementPurpose,
			"usage_classification":    movement.UsageClassification,
			"item_id":                 movement.ItemID,
			"source_location_id":      nullableString(movement.SourceLocationID),
			"destination_location_id": nullableString(movement.DestinationLocationID),
			"quantity_milli":          movement.QuantityMilli,
			"document_id":             nullableString(movement.DocumentID),
		},
	}); err != nil {
		return Movement{}, err
	}

	return movement, nil
}

func nextMovementNumber(ctx context.Context, tx *sql.Tx, orgID string) (int64, error) {
	const statement = `
INSERT INTO inventory_ops.movement_numbering_series (org_id, next_number)
VALUES ($1, 2)
ON CONFLICT (org_id)
DO UPDATE SET
	next_number = inventory_ops.movement_numbering_series.next_number + 1,
	updated_at = NOW()
RETURNING next_number - 1;`

	var number int64
	if err := tx.QueryRowContext(ctx, statement, orgID).Scan(&number); err != nil {
		return 0, fmt.Errorf("allocate movement number: %w", err)
	}
	return number, nil
}

func lockStockKey(ctx context.Context, tx *sql.Tx, orgID, itemID, locationID string) error {
	const statement = `SELECT pg_advisory_xact_lock(hashtextextended($1, 0));`
	key := orgID + ":" + itemID + ":" + locationID
	_, err := tx.ExecContext(ctx, statement, key)
	return err
}

func currentStockMilli(ctx context.Context, tx *sql.Tx, orgID, itemID, locationID string) (int64, error) {
	const query = `
SELECT COALESCE(SUM(
	CASE
		WHEN destination_location_id = $3::uuid THEN quantity_milli
		WHEN source_location_id = $3::uuid THEN -quantity_milli
		ELSE 0
	END
), 0)
FROM inventory_ops.movements
WHERE org_id = $1
  AND item_id = $2
  AND (source_location_id = $3::uuid OR destination_location_id = $3::uuid);`

	var onHand int64
	if err := tx.QueryRowContext(ctx, query, orgID, itemID, locationID).Scan(&onHand); err != nil {
		return 0, fmt.Errorf("load current stock: %w", err)
	}
	return onHand, nil
}

func loadItemForOrg(ctx context.Context, tx *sql.Tx, orgID, itemID string) (Item, error) {
	item, err := scanItem(tx.QueryRowContext(ctx, `
SELECT
	id,
	org_id,
	sku,
	name,
	item_role,
	tracking_mode,
	status,
	created_by_user_id,
	created_at,
	updated_at
FROM inventory_ops.items
WHERE org_id = $1
  AND id = $2;`,
		orgID,
		itemID,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Item{}, ErrItemNotFound
		}
		return Item{}, fmt.Errorf("load inventory item: %w", err)
	}
	return item, nil
}

func loadLocationForOrg(ctx context.Context, tx *sql.Tx, orgID, locationID string) (Location, error) {
	location, err := scanLocation(tx.QueryRowContext(ctx, `
SELECT
	id,
	org_id,
	code,
	name,
	location_role,
	status,
	created_by_user_id,
	created_at,
	updated_at
FROM inventory_ops.locations
WHERE org_id = $1
  AND id = $2;`,
		orgID,
		locationID,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Location{}, ErrLocationNotFound
		}
		return Location{}, fmt.Errorf("load inventory location: %w", err)
	}
	return location, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanItem(row rowScanner) (Item, error) {
	var item Item
	if err := row.Scan(
		&item.ID,
		&item.OrgID,
		&item.SKU,
		&item.Name,
		&item.ItemRole,
		&item.TrackingMode,
		&item.Status,
		&item.CreatedByUserID,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return Item{}, err
	}
	return item, nil
}

func scanLocation(row rowScanner) (Location, error) {
	var location Location
	if err := row.Scan(
		&location.ID,
		&location.OrgID,
		&location.Code,
		&location.Name,
		&location.LocationRole,
		&location.Status,
		&location.CreatedByUserID,
		&location.CreatedAt,
		&location.UpdatedAt,
	); err != nil {
		return Location{}, err
	}
	return location, nil
}

func scanMovement(row rowScanner) (Movement, error) {
	var movement Movement
	if err := row.Scan(
		&movement.ID,
		&movement.OrgID,
		&movement.MovementNumber,
		&movement.DocumentID,
		&movement.ItemID,
		&movement.MovementType,
		&movement.MovementPurpose,
		&movement.UsageClassification,
		&movement.SourceLocationID,
		&movement.DestinationLocationID,
		&movement.QuantityMilli,
		&movement.ReferenceNote,
		&movement.CreatedByUserID,
		&movement.CreatedAt,
	); err != nil {
		return Movement{}, err
	}
	return movement, nil
}

func scanInventoryDocument(row rowScanner) (Document, error) {
	var document Document
	if err := row.Scan(
		&document.DocumentID,
		&document.OrgID,
		&document.MovementType,
		&document.ReferenceNote,
		&document.CreatedByUserID,
		&document.CreatedAt,
		&document.UpdatedAt,
	); err != nil {
		return Document{}, err
	}
	return document, nil
}

func scanDocumentLine(row rowScanner) (DocumentLine, error) {
	var line DocumentLine
	if err := row.Scan(
		&line.ID,
		&line.DocumentID,
		&line.OrgID,
		&line.LineNumber,
		&line.MovementID,
		&line.ItemID,
		&line.MovementPurpose,
		&line.UsageClassification,
		&line.SourceLocationID,
		&line.DestinationLocationID,
		&line.QuantityMilli,
		&line.ReferenceNote,
		&line.CreatedAt,
	); err != nil {
		return DocumentLine{}, err
	}
	return line, nil
}

func scanAccountingHandoff(row rowScanner) (AccountingHandoff, error) {
	var handoff AccountingHandoff
	if err := row.Scan(
		&handoff.ID,
		&handoff.OrgID,
		&handoff.DocumentID,
		&handoff.DocumentLineID,
		&handoff.JournalEntryID,
		&handoff.HandoffStatus,
		&handoff.CreatedByUserID,
		&handoff.CreatedAt,
	); err != nil {
		return AccountingHandoff{}, err
	}
	return handoff, nil
}

func scanExecutionLink(row rowScanner) (ExecutionLink, error) {
	var link ExecutionLink
	if err := row.Scan(
		&link.ID,
		&link.OrgID,
		&link.DocumentID,
		&link.DocumentLineID,
		&link.ExecutionContextType,
		&link.ExecutionContextID,
		&link.LinkageStatus,
		&link.CreatedByUserID,
		&link.CreatedAt,
	); err != nil {
		return ExecutionLink{}, err
	}
	return link, nil
}

func isValidItemRole(value string) bool {
	switch value {
	case ItemRoleResale, ItemRoleServiceMaterial, ItemRoleTraceableEquipment, ItemRoleDirectExpenseConsumable:
		return true
	default:
		return false
	}
}

func isValidTrackingMode(value string) bool {
	switch value {
	case TrackingModeNone, TrackingModeSerial, TrackingModeLot:
		return true
	default:
		return false
	}
}

func isValidLocationRole(value string) bool {
	switch value {
	case LocationRoleWarehouse, LocationRoleVan, LocationRoleSite, LocationRoleVendor, LocationRoleCustomer, LocationRoleAdjustment, LocationRoleInstalled:
		return true
	default:
		return false
	}
}

func isValidMovementType(value string) bool {
	switch value {
	case MovementTypeReceipt, MovementTypeIssue, MovementTypeAdjustment:
		return true
	default:
		return false
	}
}

func isValidMovementPurpose(value string) bool {
	switch value {
	case MovementPurposeResale, MovementPurposeServiceConsumption, MovementPurposeInstalledEquipment, MovementPurposeDirectExpense, MovementPurposeStockAdjustment:
		return true
	default:
		return false
	}
}

func isValidUsageClassification(value string) bool {
	switch value {
	case UsageNotApplicable, UsageBillable, UsageNonBillable:
		return true
	default:
		return false
	}
}

func isValidExecutionContextType(value string) bool {
	switch value {
	case ExecutionContextWorkOrder, ExecutionContextProject:
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

func nullableString(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return value.String
}
