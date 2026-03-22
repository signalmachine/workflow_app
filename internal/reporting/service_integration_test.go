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
	workOrderService := workorders.NewService(db, documentService)
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
	gstOutput := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "2105",
		Name:                "GST Output",
		AccountClass:        accounting.AccountClassLiability,
		ControlType:         accounting.ControlTypeGSTOutput,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	tdsPayable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "2206",
		Name:                "TDS Payable",
		AccountClass:        accounting.AccountClassLiability,
		ControlType:         accounting.ControlTypeTDSPayable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1105",
		Name:                "Tax Review Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4105",
		Name:         "Tax Review Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        admin,
	})
	gst18 := createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "GST18-RPT",
		Name:             "GST Output 18%",
		TaxType:          accounting.TaxTypeGST,
		RateBasisPoints:  1800,
		PayableAccountID: gstOutput.ID,
		Actor:            admin,
	})
	tds2 := createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "TDS2-RPT",
		Name:             "TDS 194C 2%",
		TaxType:          accounting.TaxTypeTDS,
		RateBasisPoints:  200,
		PayableAccountID: tdsPayable.ID,
		Actor:            admin,
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

	gstInvoiceDoc := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver, "GST invoice")
	gstPostedAt := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)
	gstEntry, _, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   gstInvoiceDoc.ID,
		Summary:      "Post GST invoice",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		EffectiveOn:  gstPostedAt,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 59000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 50000},
			{AccountID: gstOutput.ID, Description: "GST payable", CreditMinor: 9000, TaxCode: gst18.Code},
		},
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("post GST document: %v", err)
	}

	tdsBillDoc := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver, "TDS bill")
	tdsPostedAt := gstPostedAt.Add(24 * time.Hour)
	tdsEntry, _, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   tdsBillDoc.ID,
		Summary:      "Post TDS bill",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeTDS,
		EffectiveOn:  tdsPostedAt,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 49000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 50000},
			{AccountID: tdsPayable.ID, Description: "TDS payable", DebitMinor: 1000, TaxCode: tds2.Code},
		},
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("post TDS document: %v", err)
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

	exactDocumentReview, err := reportingService.ListDocuments(ctx, reporting.ListDocumentsInput{
		DocumentID: laborJournalDoc.ID,
		Limit:      20,
		Actor:      approver,
	})
	if err != nil {
		t.Fatalf("list documents by exact id: %v", err)
	}
	if len(exactDocumentReview) != 1 || exactDocumentReview[0].DocumentID != laborJournalDoc.ID {
		t.Fatalf("unexpected exact document review result: %+v", exactDocumentReview)
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

	movements, err := reportingService.ListInventoryMovements(ctx, reporting.ListInventoryMovementsInput{
		ItemID: item.ID,
		Limit:  20,
		Actor:  operator,
	})
	if err != nil {
		t.Fatalf("list inventory movements: %v", err)
	}
	if len(movements) != 2 {
		t.Fatalf("unexpected movement review count: %d", len(movements))
	}
	if movements[0].DocumentID.String != issueDoc.ID || movements[0].MovementType != inventoryops.MovementTypeIssue {
		t.Fatalf("unexpected latest movement review: %+v", movements[0])
	}
	if !movements[0].SourceLocationCode.Valid || movements[0].SourceLocationCode.String != warehouse.Code {
		t.Fatalf("expected source location context in movement review: %+v", movements[0])
	}
	if movements[1].DocumentID.String != receiptDoc.ID || movements[1].MovementType != inventoryops.MovementTypeReceipt {
		t.Fatalf("unexpected receipt movement review: %+v", movements[1])
	}
	if !movements[1].DestinationLocationCode.Valid || movements[1].DestinationLocationCode.String != warehouse.Code {
		t.Fatalf("expected destination location context in movement review: %+v", movements[1])
	}

	reconciliation, err := reportingService.ListInventoryReconciliation(ctx, reporting.ListInventoryReconciliationInput{
		ItemID: item.ID,
		Limit:  20,
		Actor:  admin,
	})
	if err != nil {
		t.Fatalf("list inventory reconciliation: %v", err)
	}
	if len(reconciliation) != 2 {
		t.Fatalf("unexpected reconciliation row count: %d", len(reconciliation))
	}
	issueReconciliation := findInventoryReconciliationByDocument(t, reconciliation, issueDoc.ID)
	if !issueReconciliation.ExecutionLinkStatus.Valid || issueReconciliation.ExecutionLinkStatus.String != inventoryops.ExecutionLinkStatusLinked {
		t.Fatalf("expected linked execution status: %+v", issueReconciliation)
	}
	if !issueReconciliation.AccountingHandoffStatus.Valid || issueReconciliation.AccountingHandoffStatus.String != inventoryops.AccountingHandoffStatusPosted {
		t.Fatalf("expected posted accounting status: %+v", issueReconciliation)
	}
	if !issueReconciliation.WorkOrderCode.Valid || issueReconciliation.WorkOrderCode.String != workOrderResult.WorkOrder.WorkOrderCode {
		t.Fatalf("expected work-order linkage in reconciliation review: %+v", issueReconciliation)
	}
	if !issueReconciliation.JournalEntryID.Valid {
		t.Fatalf("expected posted journal linkage in reconciliation review: %+v", issueReconciliation)
	}

	pendingAdjustmentDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "inventory_adjustment", "Pending adjustment")
	pendingAdjustment, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: pendingAdjustmentDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{{
			ItemID:                item.ID,
			MovementPurpose:       inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:   inventoryops.UsageNonBillable,
			DestinationLocationID: warehouse.ID,
			QuantityMilli:         500,
			ReferenceNote:         "counted return pending review",
		}, {
			ItemID:               item.ID,
			MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
			UsageClassification:  inventoryops.UsageNonBillable,
			SourceLocationID:     warehouse.ID,
			QuantityMilli:        250,
			ReferenceNote:        "pending service consumption issue",
			AccountingHandoff:    true,
			CostMinor:            700,
			CostCurrencyCode:     "INR",
			ExecutionContextType: inventoryops.ExecutionContextProject,
			ExecutionContextID:   "PROJECT-RPT-1",
		}},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("capture pending adjustment: %v", err)
	}
	if len(pendingAdjustment.ExecutionLinks) != 1 || len(pendingAdjustment.AccountingHandoffs) != 1 {
		t.Fatalf("unexpected pending adjustment bridge counts: %+v", pendingAdjustment)
	}

	pendingAccounting, err := reportingService.ListInventoryReconciliation(ctx, reporting.ListInventoryReconciliationInput{
		DocumentID:            pendingAdjustmentDoc.ID,
		OnlyPendingAccounting: true,
		Limit:                 20,
		Actor:                 approver,
	})
	if err != nil {
		t.Fatalf("list pending accounting reconciliation: %v", err)
	}
	if len(pendingAccounting) != 1 {
		t.Fatalf("unexpected pending accounting reconciliation count: %d", len(pendingAccounting))
	}
	if !pendingAccounting[0].AccountingHandoffStatus.Valid || pendingAccounting[0].AccountingHandoffStatus.String != inventoryops.AccountingHandoffStatusPending {
		t.Fatalf("expected pending accounting handoff: %+v", pendingAccounting[0])
	}
	if pendingAccounting[0].JournalEntryID.Valid {
		t.Fatalf("did not expect posted journal on pending accounting row: %+v", pendingAccounting[0])
	}

	pendingExecution, err := reportingService.ListInventoryReconciliation(ctx, reporting.ListInventoryReconciliationInput{
		DocumentID:           pendingAdjustmentDoc.ID,
		OnlyPendingExecution: true,
		Limit:                20,
		Actor:                operator,
	})
	if err != nil {
		t.Fatalf("list pending execution reconciliation: %v", err)
	}
	if len(pendingExecution) != 1 {
		t.Fatalf("unexpected pending execution reconciliation count: %d", len(pendingExecution))
	}
	if !pendingExecution[0].ExecutionLinkStatus.Valid || pendingExecution[0].ExecutionLinkStatus.String != inventoryops.ExecutionLinkStatusPending {
		t.Fatalf("expected pending execution link: %+v", pendingExecution[0])
	}
	if pendingExecution[0].WorkOrderID.Valid {
		t.Fatalf("did not expect work-order linkage on pending project row: %+v", pendingExecution[0])
	}

	workOrderReview, err := reportingService.GetWorkOrderReview(ctx, reporting.GetWorkOrderReviewInput{
		WorkOrderID: workOrderResult.WorkOrder.ID,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("get work order review: %v", err)
	}
	if workOrderReview.DocumentID != workOrderResult.WorkOrder.DocumentID || workOrderReview.DocumentStatus != "draft" {
		t.Fatalf("unexpected work-order document review linkage: %+v", workOrderReview)
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

	workOrderList, err := reportingService.ListWorkOrders(ctx, reporting.ListWorkOrdersInput{
		Status: workorders.StatusOpen,
		Limit:  20,
		Actor:  operator,
	})
	if err != nil {
		t.Fatalf("list work order reviews: %v", err)
	}
	if len(workOrderList) == 0 || workOrderList[0].WorkOrderID != workOrderResult.WorkOrder.ID {
		t.Fatalf("unexpected work order list: %+v", workOrderList)
	}

	journalReviews, err := reportingService.ListJournalEntries(ctx, reporting.ListJournalEntriesInput{
		StartOn: startedAt,
		EndOn:   tdsPostedAt,
		Limit:   20,
		Actor:   admin,
	})
	if err != nil {
		t.Fatalf("list journal reviews: %v", err)
	}
	if len(journalReviews) < 4 {
		t.Fatalf("expected journal reviews, got %d", len(journalReviews))
	}
	if journalReviews[0].EntryID != tdsEntry.ID || journalReviews[0].TaxScopeCode != accounting.TaxScopeTDS {
		t.Fatalf("unexpected latest journal review: %+v", journalReviews[0])
	}
	if journalReviews[1].EntryID != gstEntry.ID || journalReviews[1].TaxScopeCode != accounting.TaxScopeGST {
		t.Fatalf("unexpected GST journal review: %+v", journalReviews[1])
	}
	if !journalReviews[0].SourceDocumentID.Valid || journalReviews[0].DocumentTypeCode.String != "invoice" || journalReviews[0].DocumentStatus.String != "posted" {
		t.Fatalf("expected document linkage in journal review: %+v", journalReviews[0])
	}

	controlBalances, err := reportingService.ListControlAccountBalances(ctx, reporting.ListControlAccountBalancesInput{
		AsOf:  tdsPostedAt,
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("list control account balances: %v", err)
	}
	if got := findControlAccountBalance(t, controlBalances, receivable.Code).NetMinor; got != 108000 {
		t.Fatalf("unexpected receivable net balance: %d", got)
	}
	if got := findControlAccountBalance(t, controlBalances, gstOutput.Code).NetMinor; got != -9000 {
		t.Fatalf("unexpected GST control balance: %d", got)
	}
	if got := findControlAccountBalance(t, controlBalances, tdsPayable.Code).NetMinor; got != 1000 {
		t.Fatalf("unexpected TDS control balance: %d", got)
	}

	taxSummaries, err := reportingService.ListTaxSummaries(ctx, reporting.ListTaxSummariesInput{
		StartOn: startedAt,
		EndOn:   tdsPostedAt,
		Limit:   20,
		Actor:   admin,
	})
	if err != nil {
		t.Fatalf("list tax summaries: %v", err)
	}
	if len(taxSummaries) != 2 {
		t.Fatalf("unexpected tax summary count: %d", len(taxSummaries))
	}
	gstSummary := findTaxSummary(t, taxSummaries, gst18.Code)
	if gstSummary.TaxType != accounting.TaxTypeGST || gstSummary.TotalCreditMinor != 9000 || gstSummary.NetMinor != -9000 {
		t.Fatalf("unexpected GST summary: %+v", gstSummary)
	}
	if !gstSummary.PayableAccountCode.Valid || gstSummary.PayableAccountCode.String != gstOutput.Code {
		t.Fatalf("unexpected GST payable account linkage: %+v", gstSummary)
	}
	tdsSummary := findTaxSummary(t, taxSummaries, tds2.Code)
	if tdsSummary.TaxType != accounting.TaxTypeTDS || tdsSummary.TotalDebitMinor != 1000 || tdsSummary.NetMinor != 1000 {
		t.Fatalf("unexpected TDS summary: %+v", tdsSummary)
	}
	if tdsSummary.DocumentCount != 1 || !tdsSummary.LastEffectiveOn.Valid || tdsSummary.LastEffectiveOn.Time.Format(time.DateOnly) != tdsPostedAt.Format(time.DateOnly) {
		t.Fatalf("unexpected TDS summary timing: %+v", tdsSummary)
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

func createTaxCode(t *testing.T, ctx context.Context, service *accounting.Service, input accounting.CreateTaxCodeInput) accounting.TaxCode {
	t.Helper()
	taxCode, err := service.CreateTaxCode(ctx, input)
	if err != nil {
		t.Fatalf("create tax code: %v", err)
	}
	return taxCode
}

func findControlAccountBalance(t *testing.T, balances []reporting.ControlAccountBalance, code string) reporting.ControlAccountBalance {
	t.Helper()
	for _, balance := range balances {
		if balance.AccountCode == code {
			return balance
		}
	}
	t.Fatalf("control account balance not found for code %s", code)
	return reporting.ControlAccountBalance{}
}

func findTaxSummary(t *testing.T, summaries []reporting.TaxSummary, taxCode string) reporting.TaxSummary {
	t.Helper()
	for _, summary := range summaries {
		if summary.TaxCode == taxCode {
			return summary
		}
	}
	t.Fatalf("tax summary not found for tax code %s", taxCode)
	return reporting.TaxSummary{}
}

func findInventoryReconciliationByDocument(t *testing.T, rows []reporting.InventoryReconciliationItem, documentID string) reporting.InventoryReconciliationItem {
	t.Helper()
	for _, row := range rows {
		if row.DocumentID == documentID {
			return row
		}
	}
	t.Fatalf("inventory reconciliation row not found for document %s", documentID)
	return reporting.InventoryReconciliationItem{}
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

func prepareApprovedInvoiceDocument(t *testing.T, ctx context.Context, accountingService *accounting.Service, documentService *documents.Service, workflowService *workflow.Service, operator, approver identityaccess.Actor, title string) documents.Document {
	t.Helper()

	doc, _, err := accountingService.CreateInvoice(ctx, accounting.CreateInvoiceInput{
		Title:        title,
		InvoiceRole:  accounting.InvoiceRoleSales,
		CurrencyCode: "INR",
		Summary:      title,
		Actor:        operator,
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}

	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit invoice: %v", err)
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
