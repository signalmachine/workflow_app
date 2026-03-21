package reporting_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/inventoryops"
	"workflow_app/internal/reporting"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
	"workflow_app/internal/workforce"
	"workflow_app/internal/workorders"
)

func TestReportingReviewSurfacesIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	_, approverUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleApprover, orgID)
	approverSession := startSession(t, ctx, db, orgID, approverUserID)
	approver := identityaccess.Actor{OrgID: orgID, UserID: approverUserID, SessionID: approverSession.ID}

	_, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin, orgID)
	adminSession := startSession(t, ctx, db, orgID, adminUserID)
	admin := identityaccess.Actor{OrgID: orgID, UserID: adminUserID, SessionID: adminSession.ID}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	accountingService := accounting.NewService(db, documentService)
	inventoryService := inventoryops.NewService(db)
	workOrderService := workorders.NewService(db)
	workforceService := workforce.NewService(db)
	reportingService := reporting.NewService(db)

	workOrderResult, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-RPT-1001",
		Title:         "Review execution chain",
		Summary:       "First reporting slice",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}

	worker := createWorker(t, ctx, workforceService, workforce.CreateWorkerInput{
		WorkerCode:             "TECH-RPT-1",
		DisplayName:            "Reporting Technician",
		DefaultHourlyCostMinor: 3600,
		CostCurrencyCode:       "INR",
		Actor:                  operator,
	})
	task, err := workflowService.CreateTask(ctx, workflow.CreateTaskInput{
		ContextType:         "work_order",
		ContextID:           workOrderResult.WorkOrder.ID,
		Title:               "Inspect and post",
		QueueCode:           "dispatch",
		AccountableWorkerID: worker.ID,
		Actor:               operator,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if _, err := workflowService.UpdateTaskStatus(ctx, workflow.UpdateTaskStatusInput{
		TaskID: task.ID,
		Status: "completed",
		Actor:  operator,
	}); err != nil {
		t.Fatalf("complete task: %v", err)
	}

	item := createItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "RPT-MAT-1",
		Name:         "Reporting Material",
		ItemRole:     inventoryops.ItemRoleServiceMaterial,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "RPT-WH-1",
		Name:         "Reporting Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	receiptDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "inventory_receipt", "Inventory receipt")
	if _, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: receiptDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:                item.ID,
			MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:   inventoryops.UsageBillable,
			DestinationLocationID: warehouse.ID,
			QuantityMilli:         5000,
		}},
		Actor: operator,
	}); err != nil {
		t.Fatalf("capture receipt: %v", err)
	}

	issueDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "inventory_issue", "Inventory issue")
	capturedIssue, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: issueDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:               item.ID,
			MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:  inventoryops.UsageBillable,
			SourceLocationID:     warehouse.ID,
			QuantityMilli:        2000,
			CostMinor:            5400,
			CostCurrencyCode:     "INR",
			AccountingHandoff:    true,
			ExecutionContextType: inventoryops.ExecutionContextWorkOrder,
			ExecutionContextID:   workOrderResult.WorkOrder.WorkOrderCode,
		}},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("capture issue: %v", err)
	}
	if len(capturedIssue.AccountingHandoffs) != 1 {
		t.Fatalf("unexpected issue handoff count: %d", len(capturedIssue.AccountingHandoffs))
	}

	materialUsages, err := workOrderService.SyncInventoryUsage(ctx, workorders.SyncInventoryUsageInput{
		WorkOrderID: workOrderResult.WorkOrder.ID,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("sync inventory usage: %v", err)
	}
	if len(materialUsages) != 1 {
		t.Fatalf("unexpected material usage count: %d", len(materialUsages))
	}

	startedAt := time.Date(2026, 3, 21, 9, 0, 0, 0, time.UTC)
	laborEntry, err := workforceService.RecordLabor(ctx, workforce.RecordLaborInput{
		WorkerID:    worker.ID,
		WorkOrderID: workOrderResult.WorkOrder.ID,
		TaskID:      task.ID,
		StartedAt:   startedAt,
		EndedAt:     startedAt.Add(2 * time.Hour),
		Note:        "Execution review labor",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("record labor: %v", err)
	}

	materialExpense := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "5101",
		Name:         "Material Expense",
		AccountClass: accounting.AccountClassExpense,
		Actor:        admin,
	})
	laborExpense := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "5102",
		Name:         "Labor Expense",
		AccountClass: accounting.AccountClassExpense,
		Actor:        admin,
	})
	accruedOffset := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "2201",
		Name:         "Accrued Costs",
		AccountClass: accounting.AccountClassLiability,
		Actor:        admin,
	})

	laborJournalDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "journal", "Labor posting")
	if _, err := accountingService.PostWorkOrderLabor(ctx, accounting.PostWorkOrderLaborInput{
		DocumentID:       laborJournalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: laborExpense.ID,
		OffsetAccountID:  accruedOffset.ID,
		Summary:          "Post labor costs",
		EffectiveOn:      startedAt,
		Actor:            admin,
	}); err != nil {
		t.Fatalf("post labor costs: %v", err)
	}

	materialJournalDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "journal", "Material posting")
	if _, err := accountingService.PostWorkOrderInventory(ctx, accounting.PostWorkOrderInventoryInput{
		DocumentID:       materialJournalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: materialExpense.ID,
		OffsetAccountID:  accruedOffset.ID,
		Summary:          "Post material costs",
		EffectiveOn:      startedAt,
		Actor:            admin,
	}); err != nil {
		t.Fatalf("post material costs: %v", err)
	}

	queueEntries, err := reportingService.ListApprovalQueue(ctx, reporting.ListApprovalQueueInput{
		Status: "closed",
		Limit:  20,
		Actor:  admin,
	})
	if err != nil {
		t.Fatalf("list approval queue: %v", err)
	}
	if len(queueEntries) < 4 {
		t.Fatalf("expected closed approvals, got %d", len(queueEntries))
	}
	foundMaterialPosting := false
	for _, entry := range queueEntries {
		if entry.DocumentID == materialJournalDoc.ID {
			foundMaterialPosting = true
			if entry.QueueCode != "finance-review" {
				t.Fatalf("unexpected queue code: %s", entry.QueueCode)
			}
			if !entry.JournalEntryID.Valid {
				t.Fatal("expected material posting journal entry in queue review")
			}
		}
	}
	if !foundMaterialPosting {
		t.Fatal("expected material posting document in queue review")
	}

	documentReviews, err := reportingService.ListDocuments(ctx, reporting.ListDocumentsInput{
		Status: "posted",
		Limit:  20,
		Actor:  approver,
	})
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	foundLaborPosting := false
	for _, review := range documentReviews {
		if review.DocumentID == laborJournalDoc.ID {
			foundLaborPosting = true
			if !review.JournalEntryID.Valid || !review.ApprovalID.Valid {
				t.Fatalf("expected approval and posting linkage in document review: %+v", review)
			}
		}
	}
	if !foundLaborPosting {
		t.Fatal("expected labor journal document in document review")
	}

	stock, err := reportingService.ListInventoryStock(ctx, reporting.ListInventoryStockInput{
		IncludeZero: false,
		Limit:       20,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("list inventory stock: %v", err)
	}
	if len(stock) != 1 {
		t.Fatalf("unexpected stock row count: %d", len(stock))
	}
	if stock[0].OnHandMilli != 3000 {
		t.Fatalf("unexpected on-hand stock: %d", stock[0].OnHandMilli)
	}

	workOrderReview, err := reportingService.GetWorkOrderReview(ctx, reporting.GetWorkOrderReviewInput{
		WorkOrderID: workOrderResult.WorkOrder.ID,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("get work order review: %v", err)
	}
	if workOrderReview.CompletedTaskCount != 1 || workOrderReview.OpenTaskCount != 0 {
		t.Fatalf("unexpected task counts: %+v", workOrderReview)
	}
	if workOrderReview.LaborEntryCount != 1 || workOrderReview.TotalLaborMinutes != 120 || workOrderReview.TotalLaborCostMinor != laborEntry.CostMinor {
		t.Fatalf("unexpected labor rollup: %+v", workOrderReview)
	}
	if workOrderReview.MaterialUsageCount != 1 || workOrderReview.MaterialQuantityMilli != 2000 {
		t.Fatalf("unexpected material rollup: %+v", workOrderReview)
	}
	if workOrderReview.PostedLaborEntryCount != 1 || workOrderReview.PostedMaterialUsageCount != 1 {
		t.Fatalf("unexpected posted counts: %+v", workOrderReview)
	}
	if workOrderReview.PostedMaterialCostMinor != 5400 {
		t.Fatalf("unexpected posted material cost: %d", workOrderReview.PostedMaterialCostMinor)
	}
	if !workOrderReview.LastAccountingPostedAt.Valid {
		t.Fatal("expected last accounting posted timestamp")
	}

	auditEvents, err := reportingService.LookupAuditEvents(ctx, reporting.LookupAuditEventsInput{
		EntityType: "work_orders.work_order",
		EntityID:   workOrderResult.WorkOrder.ID,
		Limit:      20,
		Actor:      admin,
	})
	if err != nil {
		t.Fatalf("lookup audit events: %v", err)
	}
	if len(auditEvents) == 0 {
		t.Fatal("expected work order audit events")
	}
}

func createLedgerAccount(t *testing.T, ctx context.Context, service *accounting.Service, input accounting.CreateLedgerAccountInput) accounting.LedgerAccount {
	t.Helper()
	account, err := service.CreateLedgerAccount(ctx, input)
	if err != nil {
		t.Fatalf("create ledger account: %v", err)
	}
	return account
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

func createWorker(t *testing.T, ctx context.Context, service *workforce.Service, input workforce.CreateWorkerInput) workforce.Worker {
	t.Helper()
	worker, err := service.CreateWorker(ctx, input)
	if err != nil {
		t.Fatalf("create worker: %v", err)
	}
	return worker
}

func prepareApprovedDocumentOfType(t *testing.T, ctx context.Context, documentService *documents.Service, workflowService *workflow.Service, operator, approver identityaccess.Actor, typeCode, title string) documents.Document {
	t.Helper()

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: typeCode,
		Title:    title,
		Actor:    operator,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit document: %v", err)
	}

	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance-review",
		Reason:     "ready for review",
		Actor:      operator,
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
