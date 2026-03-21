package workorders_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/inventoryops"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
	"workflow_app/internal/workforce"
	"workflow_app/internal/workorders"
)

func TestCreateWorkOrderConsumesPendingInventoryLinksIntegration(t *testing.T) {
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
	workOrderService := workorders.NewService(db, documentService)

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "CABLE-001",
		Name:         "Service Cable",
		ItemRole:     inventoryops.ItemRoleServiceMaterial,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "WH-WO-1",
		Name:         "Execution Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	receiptDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_receipt")
	issueDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_issue")

	_, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: receiptDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:                item.ID,
			MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:   inventoryops.UsageBillable,
			DestinationLocationID: warehouse.ID,
			QuantityMilli:         5000,
		}},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("capture receipt document: %v", err)
	}

	captured, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: issueDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:               item.ID,
			MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:  inventoryops.UsageBillable,
			SourceLocationID:     warehouse.ID,
			QuantityMilli:        2000,
			ExecutionContextType: inventoryops.ExecutionContextWorkOrder,
			ExecutionContextID:   "WO-1001",
		}},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("capture issue document: %v", err)
	}

	result, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-1001",
		Title:         "Install service cable",
		Summary:       "Thin-v1 work order foundation",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}
	if result.WorkOrder.Status != workorders.StatusOpen {
		t.Fatalf("unexpected work order status: %s", result.WorkOrder.Status)
	}
	if result.InitialHistory.ToStatus != workorders.StatusOpen {
		t.Fatalf("unexpected initial history status: %s", result.InitialHistory.ToStatus)
	}
	if len(result.MaterialUsages) != 1 {
		t.Fatalf("unexpected material usage count: %d", len(result.MaterialUsages))
	}
	if result.WorkOrder.DocumentID == "" {
		t.Fatal("expected work order document linkage")
	}
	if result.MaterialUsages[0].InventoryExecutionLinkID != captured.ExecutionLinks[0].ID {
		t.Fatalf("unexpected execution link linkage: %s", result.MaterialUsages[0].InventoryExecutionLinkID)
	}

	var (
		payloadDocumentID string
		documentTypeCode  string
		documentStatus    string
	)
	if err := db.QueryRowContext(ctx, `
SELECT wd.document_id, d.type_code, d.status
FROM work_orders.documents wd
JOIN documents.documents d
	ON d.id = wd.document_id
WHERE wd.work_order_id = $1;`, result.WorkOrder.ID).Scan(&payloadDocumentID, &documentTypeCode, &documentStatus); err != nil {
		t.Fatalf("load work order document payload: %v", err)
	}
	if payloadDocumentID != result.WorkOrder.DocumentID {
		t.Fatalf("unexpected work order document id: %s", payloadDocumentID)
	}
	if documentTypeCode != "work_order" || documentStatus != "draft" {
		t.Fatalf("unexpected work order document state: %s/%s", documentTypeCode, documentStatus)
	}

	var linkageStatus string
	if err := db.QueryRowContext(ctx, `SELECT linkage_status FROM inventory_ops.execution_links WHERE id = $1`, captured.ExecutionLinks[0].ID).Scan(&linkageStatus); err != nil {
		t.Fatalf("load execution link status: %v", err)
	}
	if linkageStatus != inventoryops.ExecutionLinkStatusLinked {
		t.Fatalf("unexpected execution link status: %s", linkageStatus)
	}
}

func TestSyncInventoryUsageConsumesLinksCreatedAfterWorkOrderIntegration(t *testing.T) {
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
	workOrderService := workorders.NewService(db, documentService)

	workOrderResult, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-2001",
		Title:         "Replace valve set",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "VALVE-001",
		Name:         "Valve",
		ItemRole:     inventoryops.ItemRoleServiceMaterial,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "WH-WO-2",
		Name:         "Valve Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	receiptDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_receipt")
	issueDoc := prepareApprovedDocument(t, ctx, db, documentService, operator, "inventory_issue")

	_, err = inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: receiptDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:                item.ID,
			MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:   inventoryops.UsageBillable,
			DestinationLocationID: warehouse.ID,
			QuantityMilli:         3000,
		}},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("capture receipt document: %v", err)
	}

	captured, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: issueDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:               item.ID,
			MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:  inventoryops.UsageBillable,
			SourceLocationID:     warehouse.ID,
			QuantityMilli:        1500,
			ExecutionContextType: inventoryops.ExecutionContextWorkOrder,
			ExecutionContextID:   "WO-2001",
		}},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("capture issue document: %v", err)
	}

	materialUsages, err := workOrderService.SyncInventoryUsage(ctx, workorders.SyncInventoryUsageInput{
		WorkOrderID: workOrderResult.WorkOrder.ID,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("sync work order inventory usage: %v", err)
	}
	if len(materialUsages) != 1 {
		t.Fatalf("unexpected synced material usage count: %d", len(materialUsages))
	}

	usages, err := workOrderService.ListMaterialUsages(ctx, workorders.ListMaterialUsagesInput{
		WorkOrderID: workOrderResult.WorkOrder.ID,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("list work order material usage: %v", err)
	}
	if len(usages) != 1 {
		t.Fatalf("unexpected stored material usage count: %d", len(usages))
	}
	if usages[0].InventoryExecutionLinkID != captured.ExecutionLinks[0].ID {
		t.Fatalf("unexpected stored execution link id: %s", usages[0].InventoryExecutionLinkID)
	}
}

func TestUpdateWorkOrderStatusRecordsHistoryAndRejectsInvalidTransitionIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	workOrderService := workorders.NewService(db, documentService)

	result, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-3001",
		Title:         "Inspect pump station",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}

	workOrder, _, err := workOrderService.UpdateStatus(ctx, workorders.UpdateStatusInput{
		WorkOrderID: result.WorkOrder.ID,
		Status:      workorders.StatusInProgress,
		Note:        "technician dispatched",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("update status to in_progress: %v", err)
	}
	if workOrder.Status != workorders.StatusInProgress {
		t.Fatalf("unexpected in_progress status: %s", workOrder.Status)
	}

	workOrder, history, err := workOrderService.UpdateStatus(ctx, workorders.UpdateStatusInput{
		WorkOrderID: result.WorkOrder.ID,
		Status:      workorders.StatusCompleted,
		Note:        "field work completed",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("update status to completed: %v", err)
	}
	if workOrder.Status != workorders.StatusCompleted {
		t.Fatalf("unexpected completed status: %s", workOrder.Status)
	}
	if !workOrder.ClosedAt.Valid {
		t.Fatal("expected closed_at to be set")
	}
	if history.FromStatus.String != workorders.StatusInProgress || history.ToStatus != workorders.StatusCompleted {
		t.Fatalf("unexpected history transition: %s -> %s", history.FromStatus.String, history.ToStatus)
	}

	_, _, err = workOrderService.UpdateStatus(ctx, workorders.UpdateStatusInput{
		WorkOrderID: result.WorkOrder.ID,
		Status:      workorders.StatusOpen,
		Actor:       operator,
	})
	if !errors.Is(err, workorders.ErrInvalidStatusTransition) {
		t.Fatalf("unexpected invalid transition error: %v", err)
	}

	var historyCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM work_orders.status_history WHERE work_order_id = $1`, result.WorkOrder.ID).Scan(&historyCount); err != nil {
		t.Fatalf("count work order history: %v", err)
	}
	if historyCount != 3 {
		t.Fatalf("unexpected history count: %d", historyCount)
	}
}

func TestWorkOrderTasksAndLaborCaptureIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	workforceService := workforce.NewService(db)
	workOrderService := workorders.NewService(db, documentService)

	result, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-4001",
		Title:         "Service rooftop unit",
		Summary:       "Execution slice with accountable task and labor capture",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}

	worker, err := workforceService.CreateWorker(ctx, workforce.CreateWorkerInput{
		WorkerCode:             "TECH-001",
		DisplayName:            "Field Technician",
		DefaultHourlyCostMinor: 3600,
		CostCurrencyCode:       "INR",
		Actor:                  operator,
	})
	if err != nil {
		t.Fatalf("create worker: %v", err)
	}

	task, err := workflowService.CreateTask(ctx, workflow.CreateTaskInput{
		ContextType:         "work_order",
		ContextID:           result.WorkOrder.ID,
		Title:               "Install replacement control board",
		Instructions:        "Bring calibrated tools and record labor against the task",
		QueueCode:           "dispatch",
		AccountableWorkerID: worker.ID,
		Actor:               operator,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.Status != "open" {
		t.Fatalf("unexpected task status: %s", task.Status)
	}

	task, err = workflowService.UpdateTaskStatus(ctx, workflow.UpdateTaskStatusInput{
		TaskID: task.ID,
		Status: "in_progress",
		Actor:  operator,
	})
	if err != nil {
		t.Fatalf("update task status: %v", err)
	}
	if task.Status != "in_progress" {
		t.Fatalf("unexpected updated task status: %s", task.Status)
	}

	startedAt := time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC)
	entry, err := workforceService.RecordLabor(ctx, workforce.RecordLaborInput{
		WorkerID:    worker.ID,
		WorkOrderID: result.WorkOrder.ID,
		TaskID:      task.ID,
		StartedAt:   startedAt,
		EndedAt:     startedAt.Add(90 * time.Minute),
		Note:        "On-site installation and calibration",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("record labor: %v", err)
	}
	if entry.DurationMinutes != 90 {
		t.Fatalf("unexpected labor duration: %d", entry.DurationMinutes)
	}
	if entry.CostMinor != 5400 {
		t.Fatalf("unexpected labor cost: %d", entry.CostMinor)
	}

	tasks, err := workflowService.ListTasks(ctx, workflow.ListTasksInput{
		ContextType: "work_order",
		ContextID:   result.WorkOrder.ID,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("unexpected task count: %d", len(tasks))
	}
	if tasks[0].AccountableWorkerID != worker.ID {
		t.Fatalf("unexpected accountable worker id: %s", tasks[0].AccountableWorkerID)
	}

	laborEntries, err := workforceService.ListLaborEntries(ctx, workforce.ListLaborEntriesInput{
		WorkOrderID: result.WorkOrder.ID,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("list labor entries: %v", err)
	}
	if len(laborEntries) != 1 {
		t.Fatalf("unexpected labor entry count: %d", len(laborEntries))
	}
	if laborEntries[0].TaskID.String != task.ID {
		t.Fatalf("unexpected labor task linkage: %s", laborEntries[0].TaskID.String)
	}

	var (
		handoffStatus string
		handoffCount  int
	)
	if err := db.QueryRowContext(ctx, `
SELECT handoff_status
FROM workforce.labor_accounting_handoffs
WHERE labor_entry_id = $1;`, entry.ID).Scan(&handoffStatus); err != nil {
		t.Fatalf("load labor accounting handoff: %v", err)
	}
	if handoffStatus != "pending" {
		t.Fatalf("unexpected labor handoff status: %s", handoffStatus)
	}
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM workforce.labor_accounting_handoffs
WHERE work_order_id = $1;`, result.WorkOrder.ID).Scan(&handoffCount); err != nil {
		t.Fatalf("count labor accounting handoffs: %v", err)
	}
	if handoffCount != 1 {
		t.Fatalf("unexpected labor accounting handoff count: %d", handoffCount)
	}
}

func TestRecordLaborRejectsTaskOwnershipMismatchIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	workforceService := workforce.NewService(db)
	workOrderService := workorders.NewService(db, documentService)

	result, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-4002",
		Title:         "Inspect generator controls",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}

	assignedWorker, err := workforceService.CreateWorker(ctx, workforce.CreateWorkerInput{
		WorkerCode:             "TECH-002",
		DisplayName:            "Assigned Technician",
		DefaultHourlyCostMinor: 3000,
		CostCurrencyCode:       "INR",
		Actor:                  operator,
	})
	if err != nil {
		t.Fatalf("create assigned worker: %v", err)
	}

	otherWorker, err := workforceService.CreateWorker(ctx, workforce.CreateWorkerInput{
		WorkerCode:             "TECH-003",
		DisplayName:            "Other Technician",
		DefaultHourlyCostMinor: 3200,
		CostCurrencyCode:       "INR",
		Actor:                  operator,
	})
	if err != nil {
		t.Fatalf("create other worker: %v", err)
	}

	task, err := workflowService.CreateTask(ctx, workflow.CreateTaskInput{
		ContextType:         "work_order",
		ContextID:           result.WorkOrder.ID,
		Title:               "Run diagnostic sequence",
		AccountableWorkerID: assignedWorker.ID,
		Actor:               operator,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	startedAt := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	_, err = workforceService.RecordLabor(ctx, workforce.RecordLaborInput{
		WorkerID:    otherWorker.ID,
		WorkOrderID: result.WorkOrder.ID,
		TaskID:      task.ID,
		StartedAt:   startedAt,
		EndedAt:     startedAt.Add(30 * time.Minute),
		Actor:       operator,
	})
	if !errors.Is(err, workforce.ErrTaskOwnershipMismatch) {
		t.Fatalf("unexpected ownership mismatch error: %v", err)
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
		QueueCode:  "work-order-review",
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

func uniqueSlug(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("150405.000000000")
}

func uniqueEmail() string {
	return "user-" + time.Now().UTC().Format("150405.000000000") + "@example.com"
}

func uniqueTokenHash() string {
	return "token-" + time.Now().UTC().Format("150405.000000000")
}
