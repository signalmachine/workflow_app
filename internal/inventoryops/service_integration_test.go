package inventoryops_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/inventoryops"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
)

func TestInventoryMovementFlowIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	inventoryService := inventoryops.NewService(db)

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "WIRE-001",
		Name:         "Copper Wiring",
		ItemRole:     inventoryops.ItemRoleServiceMaterial,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "WH-A",
		Name:         "Main Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})
	adjustmentBin := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "ADJ-A",
		Name:         "Adjustment Bin",
		LocationRole: inventoryops.LocationRoleAdjustment,
		Actor:        operator,
	})

	receiptDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_receipt")
	issueDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_issue")

	receipt := recordMovement(t, ctx, inventoryService, inventoryops.RecordMovementInput{
		DocumentID:            receiptDoc.ID,
		ItemID:                item.ID,
		MovementType:          inventoryops.MovementTypeReceipt,
		MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
		UsageClassification:   inventoryops.UsageBillable,
		DestinationLocationID: warehouse.ID,
		QuantityMilli:         10000,
		ReferenceNote:         "supplier delivery",
		Actor:                 operator,
	})
	if receipt.MovementNumber != 1 {
		t.Fatalf("unexpected receipt movement number: %d", receipt.MovementNumber)
	}

	issue := recordMovement(t, ctx, inventoryService, inventoryops.RecordMovementInput{
		DocumentID:          issueDoc.ID,
		ItemID:              item.ID,
		MovementType:        inventoryops.MovementTypeIssue,
		MovementPurpose:     inventoryops.MovementPurposeServiceConsumption,
		UsageClassification: inventoryops.UsageBillable,
		SourceLocationID:    warehouse.ID,
		QuantityMilli:       2500,
		ReferenceNote:       "billable work order usage",
		Actor:               operator,
	})
	if issue.MovementNumber != 2 {
		t.Fatalf("unexpected issue movement number: %d", issue.MovementNumber)
	}

	adjustment := recordMovement(t, ctx, inventoryService, inventoryops.RecordMovementInput{
		ItemID:                item.ID,
		MovementType:          inventoryops.MovementTypeAdjustment,
		MovementPurpose:       inventoryops.MovementPurposeStockAdjustment,
		UsageClassification:   inventoryops.UsageNotApplicable,
		DestinationLocationID: adjustmentBin.ID,
		QuantityMilli:         500,
		ReferenceNote:         "count correction",
		Actor:                 operator,
	})
	if adjustment.MovementNumber != 3 {
		t.Fatalf("unexpected adjustment movement number: %d", adjustment.MovementNumber)
	}

	stock, err := inventoryService.ListStock(ctx, inventoryops.ListStockInput{
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("list stock: %v", err)
	}

	gotWarehouse := stockAt(stock, item.ID, warehouse.ID)
	if gotWarehouse != 7500 {
		t.Fatalf("unexpected warehouse stock: got %d want %d", gotWarehouse, 7500)
	}

	gotAdjustment := stockAt(stock, item.ID, adjustmentBin.ID)
	if gotAdjustment != 500 {
		t.Fatalf("unexpected adjustment stock: got %d want %d", gotAdjustment, 500)
	}

	var auditCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM platform.audit_events WHERE org_id = $1`, orgID).Scan(&auditCount); err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	if auditCount != 16 {
		t.Fatalf("unexpected audit event count: got %d want 16", auditCount)
	}
}

func TestInventoryMovementRejectsOverIssueAndMutationIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	inventoryService := inventoryops.NewService(db)

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "BOLT-001",
		Name:         "Anchor Bolt",
		ItemRole:     inventoryops.ItemRoleResale,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "WH-B",
		Name:         "Secondary Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	recordMovement(t, ctx, inventoryService, inventoryops.RecordMovementInput{
		ItemID:                item.ID,
		MovementType:          inventoryops.MovementTypeReceipt,
		MovementPurpose:       inventoryops.MovementPurposeResale,
		UsageClassification:   inventoryops.UsageNotApplicable,
		DestinationLocationID: warehouse.ID,
		QuantityMilli:         2000,
		Actor:                 operator,
	})

	_, err := inventoryService.RecordMovement(ctx, inventoryops.RecordMovementInput{
		ItemID:              item.ID,
		MovementType:        inventoryops.MovementTypeIssue,
		MovementPurpose:     inventoryops.MovementPurposeResale,
		UsageClassification: inventoryops.UsageNotApplicable,
		SourceLocationID:    warehouse.ID,
		QuantityMilli:       2500,
		Actor:               operator,
	})
	if !errors.Is(err, inventoryops.ErrInsufficientStock) {
		t.Fatalf("unexpected over-issue error: got %v want %v", err, inventoryops.ErrInsufficientStock)
	}

	var movementID string
	if err := db.QueryRowContext(ctx, `SELECT id FROM inventory_ops.movements WHERE org_id = $1 LIMIT 1`, orgID).Scan(&movementID); err != nil {
		t.Fatalf("load movement id: %v", err)
	}

	_, err = db.ExecContext(ctx, `UPDATE inventory_ops.movements SET reference_note = 'mutated' WHERE id = $1`, movementID)
	if err == nil || !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("unexpected mutation error: %v", err)
	}
}

func TestInventoryMovementRejectsMismatchedDocumentAndPurposeIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	inventoryService := inventoryops.NewService(db)

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "KIT-001",
		Name:         "Install Kit",
		ItemRole:     inventoryops.ItemRoleTraceableEquipment,
		TrackingMode: inventoryops.TrackingModeSerial,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "WH-C",
		Name:         "Equipment Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	wrongDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_receipt")
	_, err := inventoryService.RecordMovement(ctx, inventoryops.RecordMovementInput{
		DocumentID:          wrongDoc.ID,
		ItemID:              item.ID,
		MovementType:        inventoryops.MovementTypeIssue,
		MovementPurpose:     inventoryops.MovementPurposeInstalledEquipment,
		UsageClassification: inventoryops.UsageNotApplicable,
		SourceLocationID:    warehouse.ID,
		QuantityMilli:       1000,
		Actor:               operator,
	})
	if !errors.Is(err, inventoryops.ErrInvalidInventoryDoc) {
		t.Fatalf("unexpected mismatched document error: got %v want %v", err, inventoryops.ErrInvalidInventoryDoc)
	}

	_, err = inventoryService.RecordMovement(ctx, inventoryops.RecordMovementInput{
		ItemID:                item.ID,
		MovementType:          inventoryops.MovementTypeReceipt,
		MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
		UsageClassification:   inventoryops.UsageBillable,
		DestinationLocationID: warehouse.ID,
		QuantityMilli:         1000,
		Actor:                 operator,
	})
	if !errors.Is(err, inventoryops.ErrInvalidMovement) {
		t.Fatalf("unexpected invalid purpose error: got %v want %v", err, inventoryops.ErrInvalidMovement)
	}
}

func TestInventoryListItemsAndLocationsIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin, "")
	adminSession := startSession(t, ctx, db, orgID, adminUserID)
	admin := identityaccess.Actor{OrgID: orgID, UserID: adminUserID, SessionID: adminSession.ID}

	inventoryService := inventoryops.NewService(db)

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "PUMP-100",
		Name:         "Warehouse Pump",
		ItemRole:     inventoryops.ItemRoleTraceableEquipment,
		TrackingMode: inventoryops.TrackingModeSerial,
		Actor:        admin,
	})
	location := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "WH-Z",
		Name:         "North Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        admin,
	})

	items, err := inventoryService.ListItems(ctx, inventoryops.ListItemsInput{
		ItemRole: inventoryops.ItemRoleTraceableEquipment,
		Actor:    admin,
	})
	if err != nil {
		t.Fatalf("list inventory items: %v", err)
	}
	if len(items) != 1 || items[0].ID != item.ID || items[0].TrackingMode != inventoryops.TrackingModeSerial {
		t.Fatalf("unexpected inventory items: %+v", items)
	}

	locations, err := inventoryService.ListLocations(ctx, inventoryops.ListLocationsInput{
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        admin,
	})
	if err != nil {
		t.Fatalf("list inventory locations: %v", err)
	}
	if len(locations) != 1 || locations[0].ID != location.ID || locations[0].Code != "WH-Z" {
		t.Fatalf("unexpected inventory locations: %+v", locations)
	}
}

func TestCaptureInventoryDocumentCreatesPayloadMovementsAndHandoffsIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	inventoryService := inventoryops.NewService(db)

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "CABLE-001",
		Name:         "Service Cable",
		ItemRole:     inventoryops.ItemRoleServiceMaterial,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "WH-D",
		Name:         "Dispatch Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	receiptDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_receipt")
	recordMovement(t, ctx, inventoryService, inventoryops.RecordMovementInput{
		DocumentID:            receiptDoc.ID,
		ItemID:                item.ID,
		MovementType:          inventoryops.MovementTypeReceipt,
		MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
		UsageClassification:   inventoryops.UsageBillable,
		DestinationLocationID: warehouse.ID,
		QuantityMilli:         12000,
		ReferenceNote:         "seed stock",
		Actor:                 operator,
	})

	issueDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_issue")
	result, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID:    issueDoc.ID,
		ReferenceNote: "captured service issue",
		Lines: []inventoryops.CaptureDocumentLineInput{
			{
				ItemID:               item.ID,
				MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
				UsageClassification:  inventoryops.UsageBillable,
				SourceLocationID:     warehouse.ID,
				QuantityMilli:        3500,
				ReferenceNote:        "billable work",
				AccountingHandoff:    true,
				CostMinor:            8750,
				CostCurrencyCode:     "INR",
				ExecutionContextType: inventoryops.ExecutionContextWorkOrder,
				ExecutionContextID:   "WO-1001",
			},
			{
				ItemID:               item.ID,
				MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
				UsageClassification:  inventoryops.UsageNonBillable,
				SourceLocationID:     warehouse.ID,
				QuantityMilli:        500,
				ReferenceNote:        "warranty usage",
				AccountingHandoff:    true,
				CostMinor:            1250,
				CostCurrencyCode:     "INR",
				ExecutionContextType: inventoryops.ExecutionContextProject,
				ExecutionContextID:   "PROJ-9",
			},
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("capture inventory document: %v", err)
	}

	if result.Document.DocumentID != issueDoc.ID {
		t.Fatalf("unexpected captured document id: got %s want %s", result.Document.DocumentID, issueDoc.ID)
	}
	if result.Document.MovementType != inventoryops.MovementTypeIssue {
		t.Fatalf("unexpected movement type: %s", result.Document.MovementType)
	}
	if len(result.Lines) != 2 || len(result.Movements) != 2 {
		t.Fatalf("unexpected capture counts: lines=%d movements=%d", len(result.Lines), len(result.Movements))
	}
	if len(result.AccountingHandoffs) != 2 {
		t.Fatalf("unexpected accounting handoff count: %d", len(result.AccountingHandoffs))
	}
	if !result.AccountingHandoffs[0].CostMinor.Valid || result.AccountingHandoffs[0].CostMinor.Int64 != 8750 {
		t.Fatalf("unexpected first accounting handoff cost: %+v", result.AccountingHandoffs[0].CostMinor)
	}
	if !result.AccountingHandoffs[0].CostCurrencyCode.Valid || result.AccountingHandoffs[0].CostCurrencyCode.String != "INR" {
		t.Fatalf("unexpected first accounting handoff currency: %+v", result.AccountingHandoffs[0].CostCurrencyCode)
	}
	if len(result.ExecutionLinks) != 2 {
		t.Fatalf("unexpected execution link count: %d", len(result.ExecutionLinks))
	}
	if result.Lines[0].MovementID != result.Movements[0].ID {
		t.Fatalf("expected first line to reference first movement")
	}
	if result.ExecutionLinks[0].ExecutionContextType != inventoryops.ExecutionContextWorkOrder {
		t.Fatalf("unexpected first execution context type: %s", result.ExecutionLinks[0].ExecutionContextType)
	}

	var payloadCount, lineCount, accountingCount, executionCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_ops.documents WHERE org_id = $1`, orgID).Scan(&payloadCount); err != nil {
		t.Fatalf("count inventory payloads: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_ops.document_lines WHERE org_id = $1`, orgID).Scan(&lineCount); err != nil {
		t.Fatalf("count inventory lines: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_ops.accounting_handoffs WHERE org_id = $1`, orgID).Scan(&accountingCount); err != nil {
		t.Fatalf("count accounting handoffs: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_ops.execution_links WHERE org_id = $1`, orgID).Scan(&executionCount); err != nil {
		t.Fatalf("count execution links: %v", err)
	}
	if payloadCount != 1 || lineCount != 2 || accountingCount != 2 || executionCount != 2 {
		t.Fatalf("unexpected persisted counts: payloads=%d lines=%d accounting=%d execution=%d", payloadCount, lineCount, accountingCount, executionCount)
	}

	stock, err := inventoryService.ListStock(ctx, inventoryops.ListStockInput{Actor: operator})
	if err != nil {
		t.Fatalf("list stock after capture: %v", err)
	}
	if got := stockAt(stock, item.ID, warehouse.ID); got != 8000 {
		t.Fatalf("unexpected warehouse stock after capture: got %d want %d", got, 8000)
	}
}

func TestCaptureInventoryDocumentRejectsDuplicateAndInvalidExecutionContextIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	inventoryService := inventoryops.NewService(db)

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "PIPE-001",
		Name:         "Installation Pipe",
		ItemRole:     inventoryops.ItemRoleServiceMaterial,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "WH-E",
		Name:         "Field Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	receiptDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_receipt")
	recordMovement(t, ctx, inventoryService, inventoryops.RecordMovementInput{
		DocumentID:            receiptDoc.ID,
		ItemID:                item.ID,
		MovementType:          inventoryops.MovementTypeReceipt,
		MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
		UsageClassification:   inventoryops.UsageBillable,
		DestinationLocationID: warehouse.ID,
		QuantityMilli:         3000,
		Actor:                 operator,
	})

	issueDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_issue")
	_, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: issueDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{
			{
				ItemID:              item.ID,
				MovementPurpose:     inventoryops.MovementPurposeServiceConsumption,
				UsageClassification: inventoryops.UsageBillable,
				SourceLocationID:    warehouse.ID,
				QuantityMilli:       1000,
				AccountingHandoff:   true,
				CostMinor:           2400,
				CostCurrencyCode:    "INR",
			},
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("capture initial inventory document: %v", err)
	}

	_, err = inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: issueDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{
			{
				ItemID:              item.ID,
				MovementPurpose:     inventoryops.MovementPurposeServiceConsumption,
				UsageClassification: inventoryops.UsageBillable,
				SourceLocationID:    warehouse.ID,
				QuantityMilli:       500,
			},
		},
		Actor: operator,
	})
	if !errors.Is(err, inventoryops.ErrInventoryDocExists) {
		t.Fatalf("unexpected duplicate capture error: got %v want %v", err, inventoryops.ErrInventoryDocExists)
	}

	secondIssueDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_issue")
	_, err = inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: secondIssueDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{
			{
				ItemID:               item.ID,
				MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
				UsageClassification:  inventoryops.UsageBillable,
				SourceLocationID:     warehouse.ID,
				QuantityMilli:        250,
				ExecutionContextType: inventoryops.ExecutionContextWorkOrder,
			},
		},
		Actor: operator,
	})
	if !errors.Is(err, inventoryops.ErrInvalidInventoryDoc) {
		t.Fatalf("unexpected invalid execution context error: got %v want %v", err, inventoryops.ErrInvalidInventoryDoc)
	}
}

func createItem(t *testing.T, ctx context.Context, service *inventoryops.Service, input inventoryops.CreateItemInput) inventoryops.Item {
	t.Helper()

	item, err := service.CreateItem(ctx, input)
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	return item
}

func createLocation(t *testing.T, ctx context.Context, service *inventoryops.Service, input inventoryops.CreateLocationInput) inventoryops.Location {
	t.Helper()

	location, err := service.CreateLocation(ctx, input)
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	return location
}

func recordMovement(t *testing.T, ctx context.Context, service *inventoryops.Service, input inventoryops.RecordMovementInput) inventoryops.Movement {
	t.Helper()

	movement, err := service.RecordMovement(ctx, input)
	if err != nil {
		t.Fatalf("record movement: %v", err)
	}
	return movement
}

func prepareApprovedDocument(t *testing.T, ctx context.Context, db *sql.DB, documentService *documents.Service, actor identityaccess.Actor, typeCode string) documents.Document {
	t.Helper()

	orgID, approverUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleApprover, actor.OrgID)
	if orgID != actor.OrgID {
		t.Fatalf("unexpected org mismatch")
	}
	approverSession := startSession(t, ctx, db, actor.OrgID, approverUserID)
	approver := identityaccess.Actor{OrgID: actor.OrgID, UserID: approverUserID, SessionID: approverSession.ID}

	workflowService := workflow.NewService(db, documentService)

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: typeCode,
		Title:    typeCode + " draft",
		Actor:    actor,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("submit document: %v", err)
	}

	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "inventory-review",
		Actor:      actor,
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}

	_, doc, err = workflowService.DecideApproval(ctx, workflow.DecideApprovalInput{
		ApprovalID: approval.ID,
		Decision:   "approved",
		Actor:      approver,
	})
	if err != nil {
		t.Fatalf("decide approval: %v", err)
	}

	return doc
}

func seedOrgAndUser(t *testing.T, ctx context.Context, db *sql.DB, roleCode, existingOrgID string) (string, string) {
	t.Helper()

	orgID := existingOrgID
	if orgID == "" {
		if err := db.QueryRowContext(
			ctx,
			`INSERT INTO identityaccess.orgs (slug, name) VALUES ($1, $2) RETURNING id`,
			uniqueSlug("acme"),
			"Acme",
		).Scan(&orgID); err != nil {
			t.Fatalf("insert org: %v", err)
		}
	}

	var userID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name) VALUES ($1, 'Example User') RETURNING id`,
		uniqueEmail(),
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO identityaccess.memberships (org_id, user_id, role_code) VALUES ($1, $2, $3)`,
		orgID,
		userID,
		roleCode,
	); err != nil {
		t.Fatalf("insert membership: %v", err)
	}

	return orgID, userID
}

func startSession(t *testing.T, ctx context.Context, db *sql.DB, orgID, userID string) identityaccess.Session {
	t.Helper()

	service := identityaccess.NewService(db)
	session, err := service.StartSession(ctx, identityaccess.StartSessionInput{
		OrgID:            orgID,
		UserID:           userID,
		DeviceLabel:      "test-device",
		RefreshTokenHash: uniqueTokenHash(),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	return session
}

func stockAt(stock []inventoryops.StockBalance, itemID, locationID string) int64 {
	for _, balance := range stock {
		if balance.ItemID == itemID && balance.LocationID == locationID {
			return balance.OnHandMilli
		}
	}
	return 0
}

func uniqueSlug(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("150405.000000000")
}

func uniqueEmail() string {
	return "user-" + time.Now().UTC().Format("150405.000000000") + "@example.com"
}

func uniqueTokenHash() string {
	return "token-" + time.Now().UTC().Format("150405.000000000")
}
