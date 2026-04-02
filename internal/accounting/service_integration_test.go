package accounting_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/inventoryops"
	"workflow_app/internal/parties"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
	"workflow_app/internal/workforce"
	"workflow_app/internal/workorders"
)

func TestPostDocumentIntegration(t *testing.T) {
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

	doc := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)

	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4000",
		Name:         "Service Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        admin,
	})

	postInput := accounting.PostDocumentInput{
		DocumentID:   doc.ID,
		Summary:      "Post approved invoice",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeNone,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 150000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 150000},
		},
		Actor: admin,
	}

	entry, lines, postedDoc, err := accountingService.PostDocument(ctx, postInput)
	if err != nil {
		t.Fatalf("post document: %v", err)
	}
	if entry.EntryKind != accounting.EntryKindPosting {
		t.Fatalf("unexpected entry kind: %s", entry.EntryKind)
	}
	if entry.EntryNumber != 1 {
		t.Fatalf("unexpected entry number: %d", entry.EntryNumber)
	}
	if postedDoc.Status != documents.StatusPosted {
		t.Fatalf("unexpected document status: %s", postedDoc.Status)
	}
	if len(lines) != 2 {
		t.Fatalf("unexpected line count: %d", len(lines))
	}

	idempotentEntry, idempotentLines, idempotentDoc, err := accountingService.PostDocument(ctx, postInput)
	if err != nil {
		t.Fatalf("idempotent post document: %v", err)
	}
	if idempotentEntry.ID != entry.ID {
		t.Fatalf("unexpected idempotent entry id: got %s want %s", idempotentEntry.ID, entry.ID)
	}
	if len(idempotentLines) != len(lines) {
		t.Fatalf("unexpected idempotent line count: %d", len(idempotentLines))
	}
	if idempotentDoc.Status != documents.StatusPosted {
		t.Fatalf("unexpected idempotent document status: %s", idempotentDoc.Status)
	}

	_, _, _, err = accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   doc.ID,
		Summary:      "Post approved invoice with different payload",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeNone,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 200000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 200000},
		},
		Actor: admin,
	})
	if !errors.Is(err, accounting.ErrPostingAlreadyExists) {
		t.Fatalf("unexpected duplicate posting error: got %v want %v", err, accounting.ErrPostingAlreadyExists)
	}

	var journalCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM accounting.journal_entries WHERE org_id = $1`, orgID).Scan(&journalCount); err != nil {
		t.Fatalf("count journal entries: %v", err)
	}
	if journalCount != 1 {
		t.Fatalf("unexpected journal entry count: %d", journalCount)
	}
}

func TestCreateTaxCodeAndUseItInPostingIntegration(t *testing.T) {
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

	doc := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)

	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	gstOutput := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "2101",
		Name:                "GST Output",
		AccountClass:        accounting.AccountClassLiability,
		ControlType:         accounting.ControlTypeGSTOutput,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4000",
		Name:         "Service Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        admin,
	})

	gst18 := createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "GST18",
		Name:             "GST Output 18%",
		TaxType:          accounting.TaxTypeGST,
		RateBasisPoints:  1800,
		PayableAccountID: gstOutput.ID,
		Actor:            admin,
	})

	entry, lines, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   doc.ID,
		Summary:      "Post approved invoice with GST",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 177000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 150000},
			{AccountID: gstOutput.ID, Description: "GST payable", CreditMinor: 27000, TaxCode: gst18.Code},
		},
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("post document with GST tax code: %v", err)
	}
	if entry.TaxScopeCode != accounting.TaxScopeGST {
		t.Fatalf("unexpected tax scope: %s", entry.TaxScopeCode)
	}
	if len(lines) != 3 {
		t.Fatalf("unexpected journal line count: %d", len(lines))
	}
	if !lines[2].TaxCode.Valid || lines[2].TaxCode.String != gst18.Code {
		t.Fatalf("unexpected tax code on journal line: %+v", lines[2].TaxCode)
	}
}

func TestUpdateSetupStatusesIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin, "")
	adminSession := startSession(t, ctx, db, orgID, adminUserID)
	admin := identityaccess.Actor{OrgID: orgID, UserID: adminUserID, SessionID: adminSession.ID}

	documentService := documents.NewService(db)
	accountingService := accounting.NewService(db, documentService)

	ledger := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "5000",
		Name:         "Field Expenses",
		AccountClass: accounting.AccountClassExpense,
		Actor:        admin,
	})
	updatedLedger, err := accountingService.UpdateLedgerAccountStatus(ctx, accounting.UpdateLedgerAccountStatusInput{
		AccountID: ledger.ID,
		Status:    accounting.StatusInactive,
		Actor:     admin,
	})
	if err != nil {
		t.Fatalf("update ledger account status: %v", err)
	}
	if updatedLedger.Status != accounting.StatusInactive {
		t.Fatalf("unexpected ledger status: %s", updatedLedger.Status)
	}

	control := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "2101",
		Name:         "GST Output",
		AccountClass: accounting.AccountClassLiability,
		ControlType:  accounting.ControlTypeGSTOutput,
		Actor:        admin,
	})
	taxCode := createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "GST18",
		Name:             "GST 18%",
		TaxType:          accounting.TaxTypeGST,
		RateBasisPoints:  1800,
		PayableAccountID: control.ID,
		Actor:            admin,
	})
	updatedTaxCode, err := accountingService.UpdateTaxCodeStatus(ctx, accounting.UpdateTaxCodeStatusInput{
		TaxCodeID: taxCode.ID,
		Status:    accounting.StatusInactive,
		Actor:     admin,
	})
	if err != nil {
		t.Fatalf("update tax code status: %v", err)
	}
	if updatedTaxCode.Status != accounting.StatusInactive {
		t.Fatalf("unexpected tax code status: %s", updatedTaxCode.Status)
	}
}

func TestCreateAdoptedAccountingDocumentsIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	accountingService := accounting.NewService(db, documentService)
	partiesService := parties.NewService(db)

	customer := createParty(t, ctx, partiesService, parties.CreatePartyInput{
		PartyCode:   "CUST-1001",
		DisplayName: "Acme Customer",
		LegalName:   "Acme Customer Pvt Ltd",
		PartyKind:   parties.PartyKindCustomerVendor,
		Actor:       operator,
	})
	contact := createContact(t, ctx, partiesService, parties.CreateContactInput{
		PartyID:   customer.ID,
		FullName:  "Taylor Example",
		RoleTitle: "Finance",
		Email:     "taylor@example.com",
		IsPrimary: true,
		Actor:     operator,
	})

	invoiceDoc, invoicePayload, err := accountingService.CreateInvoice(ctx, accounting.CreateInvoiceInput{
		Title:            "Customer invoice",
		InvoiceRole:      accounting.InvoiceRoleSales,
		BilledPartyID:    customer.ID,
		BillingContactID: contact.ID,
		CurrencyCode:     "INR",
		ReferenceValue:   "INV-REF-1001",
		Summary:          "Monthly services",
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if invoiceDoc.TypeCode != "invoice" || invoicePayload.DocumentID != invoiceDoc.ID {
		t.Fatalf("unexpected invoice ownership result: doc=%+v payload=%+v", invoiceDoc, invoicePayload)
	}
	if !invoicePayload.BilledPartyID.Valid || invoicePayload.BilledPartyID.String != customer.ID {
		t.Fatalf("unexpected invoice billed party: %+v", invoicePayload.BilledPartyID)
	}
	if !invoicePayload.BillingContactID.Valid || invoicePayload.BillingContactID.String != contact.ID {
		t.Fatalf("unexpected invoice billing contact: %+v", invoicePayload.BillingContactID)
	}

	paymentDoc, paymentPayload, err := accountingService.CreatePaymentReceipt(ctx, accounting.CreatePaymentReceiptInput{
		Title:                 "Customer receipt",
		Direction:             accounting.PaymentReceiptDirectionReceipt,
		CounterpartyID:        customer.ID,
		CounterpartyContactID: contact.ID,
		CurrencyCode:          "INR",
		ReferenceValue:        "RCPT-1001",
		Summary:               "Receipt against invoice",
		Actor:                 operator,
	})
	if err != nil {
		t.Fatalf("create payment receipt: %v", err)
	}
	if paymentDoc.TypeCode != "payment_receipt" || paymentPayload.DocumentID != paymentDoc.ID {
		t.Fatalf("unexpected payment receipt ownership result: doc=%+v payload=%+v", paymentDoc, paymentPayload)
	}
	if !paymentPayload.Direction.Valid || paymentPayload.Direction.String != accounting.PaymentReceiptDirectionReceipt {
		t.Fatalf("unexpected payment receipt direction: %+v", paymentPayload.Direction)
	}

	var invoicePayloadCount int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM accounting.invoice_documents
WHERE org_id = $1
  AND document_id = $2;`,
		orgID,
		invoiceDoc.ID,
	).Scan(&invoicePayloadCount); err != nil {
		t.Fatalf("count invoice payload rows: %v", err)
	}
	if invoicePayloadCount != 1 {
		t.Fatalf("unexpected invoice payload count: %d", invoicePayloadCount)
	}

	var paymentPayloadCount int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM accounting.payment_receipt_documents
WHERE org_id = $1
  AND document_id = $2;`,
		orgID,
		paymentDoc.ID,
	).Scan(&paymentPayloadCount); err != nil {
		t.Fatalf("count payment receipt payload rows: %v", err)
	}
	if paymentPayloadCount != 1 {
		t.Fatalf("unexpected payment receipt payload count: %d", paymentPayloadCount)
	}
}

func TestPostDocumentRejectsBareInvoiceWithoutPayloadIntegration(t *testing.T) {
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

	doc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "invoice", "Bare invoice")

	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4000",
		Name:         "Service Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        admin,
	})

	_, _, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   doc.ID,
		Summary:      "Post bare invoice",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeNone,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 100000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 100000},
		},
		Actor: admin,
	})
	if !errors.Is(err, accounting.ErrInvoiceDocumentNotFound) {
		t.Fatalf("unexpected bare invoice posting error: got %v want %v", err, accounting.ErrInvoiceDocumentNotFound)
	}
}

func TestPostWorkOrderLaborIntegration(t *testing.T) {
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
	workOrderService := workorders.NewService(db, documentService)
	workforceService := workforce.NewService(db)
	accountingService := accounting.NewService(db, documentService)

	workOrderResult, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-5001",
		Title:         "Commission air handling unit",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}

	worker, err := workforceService.CreateWorker(ctx, workforce.CreateWorkerInput{
		WorkerCode:             "TECH-5001",
		DisplayName:            "Commissioning Technician",
		DefaultHourlyCostMinor: 3600,
		CostCurrencyCode:       "INR",
		Actor:                  operator,
	})
	if err != nil {
		t.Fatalf("create worker: %v", err)
	}

	task, err := workflowService.CreateTask(ctx, workflow.CreateTaskInput{
		ContextType:         "work_order",
		ContextID:           workOrderResult.WorkOrder.ID,
		Title:               "Calibrate controls",
		AccountableWorkerID: worker.ID,
		Actor:               operator,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	startedAt := time.Date(2026, 3, 20, 8, 0, 0, 0, time.UTC)
	laborEntry, err := workforceService.RecordLabor(ctx, workforce.RecordLaborInput{
		WorkerID:    worker.ID,
		WorkOrderID: workOrderResult.WorkOrder.ID,
		TaskID:      task.ID,
		StartedAt:   startedAt,
		EndedAt:     startedAt.Add(2 * time.Hour),
		Note:        "Commissioning and validation",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("record labor: %v", err)
	}
	if laborEntry.CostMinor != 7200 {
		t.Fatalf("unexpected labor cost: %d", laborEntry.CostMinor)
	}

	journalDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "journal", "Labor posting")
	laborExpense := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "5100",
		Name:         "Direct Labor Expense",
		AccountClass: accounting.AccountClassExpense,
		Actor:        admin,
	})
	accruedLabor := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "2205",
		Name:         "Accrued Labor",
		AccountClass: accounting.AccountClassLiability,
		Actor:        admin,
	})

	result, err := accountingService.PostWorkOrderLabor(ctx, accounting.PostWorkOrderLaborInput{
		DocumentID:       journalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: laborExpense.ID,
		OffsetAccountID:  accruedLabor.ID,
		Summary:          "Post work-order labor costs",
		EffectiveOn:      startedAt,
		Actor:            admin,
	})
	if err != nil {
		t.Fatalf("post work-order labor: %v", err)
	}
	if result.Document.Status != documents.StatusPosted {
		t.Fatalf("unexpected posted document status: %s", result.Document.Status)
	}
	if result.LaborEntryCount != 1 {
		t.Fatalf("unexpected labor entry count: %d", result.LaborEntryCount)
	}
	if result.TotalCostMinor != 7200 {
		t.Fatalf("unexpected total labor cost: %d", result.TotalCostMinor)
	}
	if result.CurrencyCode != "INR" {
		t.Fatalf("unexpected labor posting currency: %s", result.CurrencyCode)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("unexpected journal line count: %d", len(result.Lines))
	}
	if result.Lines[0].DebitMinor != 7200 || result.Lines[1].CreditMinor != 7200 {
		t.Fatalf("unexpected journal amounts: %+v", result.Lines)
	}

	idempotent, err := accountingService.PostWorkOrderLabor(ctx, accounting.PostWorkOrderLaborInput{
		DocumentID:       journalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: laborExpense.ID,
		OffsetAccountID:  accruedLabor.ID,
		Summary:          "Post work-order labor costs",
		EffectiveOn:      startedAt,
		Actor:            admin,
	})
	if err != nil {
		t.Fatalf("idempotent post work-order labor: %v", err)
	}
	if idempotent.Entry.ID != result.Entry.ID {
		t.Fatalf("unexpected idempotent journal entry id: %s", idempotent.Entry.ID)
	}

	var (
		handoffStatus  string
		journalEntryID sql.NullString
		postedAt       sql.NullTime
	)
	if err := db.QueryRowContext(ctx, `
SELECT handoff_status, journal_entry_id, posted_at
FROM workforce.labor_accounting_handoffs
WHERE labor_entry_id = $1;`, laborEntry.ID).Scan(&handoffStatus, &journalEntryID, &postedAt); err != nil {
		t.Fatalf("load labor accounting handoff: %v", err)
	}
	if handoffStatus != "posted" {
		t.Fatalf("unexpected labor handoff status: %s", handoffStatus)
	}
	if !journalEntryID.Valid || journalEntryID.String != result.Entry.ID {
		t.Fatalf("unexpected labor handoff journal entry: %+v", journalEntryID)
	}
	if !postedAt.Valid {
		t.Fatal("expected labor handoff posted_at")
	}
}

func TestPostWorkOrderInventoryIntegration(t *testing.T) {
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

	workOrderResult, err := workOrderService.CreateWorkOrder(ctx, workorders.CreateWorkOrderInput{
		WorkOrderCode: "WO-INV-1001",
		Title:         "Install service materials",
		Actor:         operator,
	})
	if err != nil {
		t.Fatalf("create work order: %v", err)
	}

	item := createInventoryItem(t, ctx, inventoryService, inventoryops.CreateItemInput{
		SKU:          "MAT-1001",
		Name:         "Service Material",
		ItemRole:     inventoryops.ItemRoleServiceMaterial,
		TrackingMode: inventoryops.TrackingModeNone,
		Actor:        operator,
	})
	warehouse := createInventoryLocation(t, ctx, inventoryService, inventoryops.CreateLocationInput{
		Code:         "WH-INV-1",
		Name:         "Inventory Warehouse",
		LocationRole: inventoryops.LocationRoleWarehouse,
		Actor:        operator,
	})

	receiptDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "inventory_receipt", "Inventory receipt")
	_, err = inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
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

	issueDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "inventory_issue", "Inventory issue")
	captured, err := inventoryService.CaptureDocument(ctx, inventoryops.CaptureDocumentInput{
		DocumentID: issueDoc.ID,
		Lines: []inventoryops.CaptureDocumentLineInput{
			{
				ItemID:               item.ID,
				MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
				UsageClassification:  inventoryops.UsageBillable,
				SourceLocationID:     warehouse.ID,
				QuantityMilli:        2000,
				AccountingHandoff:    true,
				CostMinor:            8600,
				CostCurrencyCode:     "INR",
				ExecutionContextType: inventoryops.ExecutionContextWorkOrder,
				ExecutionContextID:   "WO-INV-1001",
			},
			{
				ItemID:               item.ID,
				MovementPurpose:      inventoryops.MovementPurposeServiceConsumption,
				UsageClassification:  inventoryops.UsageNonBillable,
				SourceLocationID:     warehouse.ID,
				QuantityMilli:        500,
				AccountingHandoff:    true,
				CostMinor:            2150,
				CostCurrencyCode:     "INR",
				ExecutionContextType: inventoryops.ExecutionContextWorkOrder,
				ExecutionContextID:   "WO-INV-1001",
			},
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("capture issue document: %v", err)
	}
	if len(captured.AccountingHandoffs) != 2 {
		t.Fatalf("unexpected inventory handoff count: %d", len(captured.AccountingHandoffs))
	}

	materialUsages, err := workOrderService.SyncInventoryUsage(ctx, workorders.SyncInventoryUsageInput{
		WorkOrderID: workOrderResult.WorkOrder.ID,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("sync work order inventory usage: %v", err)
	}
	if len(materialUsages) != 2 {
		t.Fatalf("unexpected material usage count: %d", len(materialUsages))
	}

	journalDoc := prepareApprovedDocumentOfType(t, ctx, documentService, workflowService, operator, approver, "journal", "Inventory posting")
	materialExpense := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "5200",
		Name:         "Material Consumption Expense",
		AccountClass: accounting.AccountClassExpense,
		Actor:        admin,
	})
	inventoryClearing := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "2210",
		Name:         "Inventory Issue Clearing",
		AccountClass: accounting.AccountClassLiability,
		Actor:        admin,
	})

	result, err := accountingService.PostWorkOrderInventory(ctx, accounting.PostWorkOrderInventoryInput{
		DocumentID:       journalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: materialExpense.ID,
		OffsetAccountID:  inventoryClearing.ID,
		Summary:          "Post work-order material costs",
		EffectiveOn:      time.Date(2026, 3, 20, 11, 0, 0, 0, time.UTC),
		Actor:            admin,
	})
	if err != nil {
		t.Fatalf("post work-order inventory: %v", err)
	}
	if result.Document.Status != documents.StatusPosted {
		t.Fatalf("unexpected posted inventory document status: %s", result.Document.Status)
	}
	if result.InventoryLineCount != 2 {
		t.Fatalf("unexpected inventory line count: %d", result.InventoryLineCount)
	}
	if result.TotalCostMinor != 10750 {
		t.Fatalf("unexpected total inventory cost: %d", result.TotalCostMinor)
	}
	if result.CurrencyCode != "INR" {
		t.Fatalf("unexpected inventory posting currency: %s", result.CurrencyCode)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("unexpected inventory journal line count: %d", len(result.Lines))
	}
	if result.Lines[0].DebitMinor != 10750 || result.Lines[1].CreditMinor != 10750 {
		t.Fatalf("unexpected inventory journal amounts: %+v", result.Lines)
	}

	idempotent, err := accountingService.PostWorkOrderInventory(ctx, accounting.PostWorkOrderInventoryInput{
		DocumentID:       journalDoc.ID,
		WorkOrderID:      workOrderResult.WorkOrder.ID,
		ExpenseAccountID: materialExpense.ID,
		OffsetAccountID:  inventoryClearing.ID,
		Summary:          "Post work-order material costs",
		EffectiveOn:      time.Date(2026, 3, 20, 11, 0, 0, 0, time.UTC),
		Actor:            admin,
	})
	if err != nil {
		t.Fatalf("idempotent post work-order inventory: %v", err)
	}
	if idempotent.Entry.ID != result.Entry.ID {
		t.Fatalf("unexpected idempotent inventory journal entry id: %s", idempotent.Entry.ID)
	}

	var (
		handoffStatus  string
		journalEntryID sql.NullString
		postedAt       sql.NullTime
	)
	if err := db.QueryRowContext(ctx, `
SELECT handoff_status, journal_entry_id, posted_at
FROM inventory_ops.accounting_handoffs
WHERE document_line_id = $1;`, materialUsages[0].InventoryDocumentLineID).Scan(&handoffStatus, &journalEntryID, &postedAt); err != nil {
		t.Fatalf("load inventory accounting handoff: %v", err)
	}
	if handoffStatus != "posted" {
		t.Fatalf("unexpected inventory handoff status: %s", handoffStatus)
	}
	if !journalEntryID.Valid || journalEntryID.String != result.Entry.ID {
		t.Fatalf("unexpected inventory handoff journal entry: %+v", journalEntryID)
	}
	if !postedAt.Valid {
		t.Fatal("expected inventory handoff posted_at")
	}
}

func TestPostDocumentRejectsMissingOrMismatchedTaxCodes(t *testing.T) {
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

	docWithoutTaxCode := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)
	docWithUnknownTaxCode := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)
	docWithWrongTaxType := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)

	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	tdsPayable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "2201",
		Name:                "TDS Payable",
		AccountClass:        accounting.AccountClassLiability,
		ControlType:         accounting.ControlTypeTDSPayable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4000",
		Name:         "Service Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        admin,
	})

	tds194c := createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "TDS194C",
		Name:             "TDS 194C",
		TaxType:          accounting.TaxTypeTDS,
		RateBasisPoints:  100,
		PayableAccountID: tdsPayable.ID,
		Actor:            admin,
	})

	_, _, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   docWithoutTaxCode.ID,
		Summary:      "Missing tax code",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 100000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 100000},
		},
		Actor: admin,
	})
	if !errors.Is(err, accounting.ErrInvalidTaxScope) {
		t.Fatalf("unexpected missing tax code error: got %v want %v", err, accounting.ErrInvalidTaxScope)
	}

	_, _, _, err = accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   docWithUnknownTaxCode.ID,
		Summary:      "Unknown tax code",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeTDS,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 99000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 99000, TaxCode: "UNKNOWN"},
		},
		Actor: admin,
	})
	if !errors.Is(err, accounting.ErrTaxCodeNotFound) {
		t.Fatalf("unexpected unknown tax code error: got %v want %v", err, accounting.ErrTaxCodeNotFound)
	}

	_, _, _, err = accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   docWithWrongTaxType.ID,
		Summary:      "Wrong tax type",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 99000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 99000, TaxCode: tds194c.Code},
		},
		Actor: admin,
	})
	if !errors.Is(err, accounting.ErrInvalidTaxScope) {
		t.Fatalf("unexpected wrong tax type error: got %v want %v", err, accounting.ErrInvalidTaxScope)
	}
}

func TestReverseDocumentIntegration(t *testing.T) {
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

	doc := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)

	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4000",
		Name:         "Service Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        admin,
	})
	gstOutput := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "2101",
		Name:                "GST Output",
		AccountClass:        accounting.AccountClassLiability,
		ControlType:         accounting.ControlTypeGSTOutput,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	gst18 := createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "GST18",
		Name:             "GST Output 18%",
		TaxType:          accounting.TaxTypeGST,
		RateBasisPoints:  1800,
		PayableAccountID: gstOutput.ID,
		Actor:            admin,
	})

	postedEntry, postedLines, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   doc.ID,
		Summary:      "Post approved invoice",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 177000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 150000},
			{AccountID: gstOutput.ID, Description: "GST payable", CreditMinor: 27000, TaxCode: gst18.Code},
		},
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("post document: %v", err)
	}

	reversal, reversalLines, reversedDoc, err := accountingService.ReverseDocument(ctx, accounting.ReverseDocumentInput{
		DocumentID: doc.ID,
		Reason:     "customer cancellation",
		Actor:      admin,
	})
	if err != nil {
		t.Fatalf("reverse document: %v", err)
	}
	if reversal.EntryKind != accounting.EntryKindReversal {
		t.Fatalf("unexpected reversal entry kind: %s", reversal.EntryKind)
	}
	if reversal.EntryNumber != 2 {
		t.Fatalf("unexpected reversal entry number: %d", reversal.EntryNumber)
	}
	if !reversal.ReversalOfEntryID.Valid || reversal.ReversalOfEntryID.String != postedEntry.ID {
		t.Fatalf("unexpected reversal_of entry: %+v", reversal.ReversalOfEntryID)
	}
	if reversedDoc.Status != documents.StatusReversed {
		t.Fatalf("unexpected reversed document status: %s", reversedDoc.Status)
	}
	if len(reversalLines) != len(postedLines) {
		t.Fatalf("unexpected reversal line count: %d", len(reversalLines))
	}
	if reversalLines[0].DebitMinor != postedLines[0].CreditMinor || reversalLines[0].CreditMinor != postedLines[0].DebitMinor {
		t.Fatalf("unexpected first reversal line amounts: %+v vs %+v", reversalLines[0], postedLines[0])
	}
	if reversalLines[1].DebitMinor != postedLines[1].CreditMinor || reversalLines[1].CreditMinor != postedLines[1].DebitMinor {
		t.Fatalf("unexpected second reversal line amounts: %+v vs %+v", reversalLines[1], postedLines[1])
	}

	idempotentReversal, _, idempotentDoc, err := accountingService.ReverseDocument(ctx, accounting.ReverseDocumentInput{
		DocumentID: doc.ID,
		Reason:     "customer cancellation",
		Actor:      admin,
	})
	if err != nil {
		t.Fatalf("idempotent reverse document: %v", err)
	}
	if idempotentReversal.ID != reversal.ID {
		t.Fatalf("unexpected idempotent reversal id: got %s want %s", idempotentReversal.ID, reversal.ID)
	}
	if idempotentDoc.Status != documents.StatusReversed {
		t.Fatalf("unexpected idempotent reversed document status: %s", idempotentDoc.Status)
	}

	_, _, _, err = accountingService.ReverseDocument(ctx, accounting.ReverseDocumentInput{
		DocumentID: doc.ID,
		Reason:     "different reason",
		Actor:      admin,
	})
	if !errors.Is(err, accounting.ErrAlreadyReversed) {
		t.Fatalf("unexpected second reversal error: got %v want %v", err, accounting.ErrAlreadyReversed)
	}
}

func TestAccountingPeriodsControlPostingAndReversalIntegration(t *testing.T) {
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

	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4000",
		Name:         "Service Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        admin,
	})

	today := time.Date(2026, 3, 20, 14, 0, 0, 0, time.UTC)
	tomorrow := today.Add(24 * time.Hour)

	period, err := accountingService.CreateAccountingPeriod(ctx, accounting.CreateAccountingPeriodInput{
		PeriodCode: "2026-03-20",
		StartOn:    today,
		EndOn:      today,
		Actor:      admin,
	})
	if err != nil {
		t.Fatalf("create accounting period: %v", err)
	}

	docForPosting := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)
	entry, _, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   docForPosting.ID,
		Summary:      "Post inside open period",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeNone,
		EffectiveOn:  today,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 150000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 150000},
		},
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("post inside open period: %v", err)
	}
	if got := entry.EffectiveOn.Format(time.DateOnly); got != "2026-03-20" {
		t.Fatalf("unexpected effective_on: %s", got)
	}

	period, err = accountingService.CloseAccountingPeriod(ctx, accounting.CloseAccountingPeriodInput{
		PeriodID: period.ID,
		Actor:    admin,
	})
	if err != nil {
		t.Fatalf("close accounting period: %v", err)
	}
	if period.Status != "closed" {
		t.Fatalf("unexpected period status: %s", period.Status)
	}

	docBlockedByClosedPeriod := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)
	_, _, _, err = accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   docBlockedByClosedPeriod.ID,
		Summary:      "Blocked by closed period",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeNone,
		EffectiveOn:  today,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 120000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 120000},
		},
		Actor: admin,
	})
	if !errors.Is(err, accounting.ErrAccountingPeriodNotOpen) {
		t.Fatalf("unexpected closed-period post error: got %v want %v", err, accounting.ErrAccountingPeriodNotOpen)
	}

	_, _, _, err = accountingService.ReverseDocument(ctx, accounting.ReverseDocumentInput{
		DocumentID:  docForPosting.ID,
		Reason:      "blocked in closed period",
		EffectiveOn: today,
		Actor:       admin,
	})
	if !errors.Is(err, accounting.ErrAccountingPeriodNotOpen) {
		t.Fatalf("unexpected closed-period reversal error: got %v want %v", err, accounting.ErrAccountingPeriodNotOpen)
	}

	if _, err := accountingService.CreateAccountingPeriod(ctx, accounting.CreateAccountingPeriodInput{
		PeriodCode: "2026-03-21",
		StartOn:    tomorrow,
		EndOn:      tomorrow,
		Actor:      admin,
	}); err != nil {
		t.Fatalf("create next accounting period: %v", err)
	}

	reversal, _, _, err := accountingService.ReverseDocument(ctx, accounting.ReverseDocumentInput{
		DocumentID:  docForPosting.ID,
		Reason:      "next-day reversal",
		EffectiveOn: tomorrow,
		Actor:       admin,
	})
	if err != nil {
		t.Fatalf("reverse inside next open period: %v", err)
	}
	if got := reversal.EffectiveOn.Format(time.DateOnly); got != "2026-03-21" {
		t.Fatalf("unexpected reversal effective_on: %s", got)
	}
}

func TestAdminAccountingMaintenanceListingIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, adminUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleAdmin, "")
	adminSession := startSession(t, ctx, db, orgID, adminUserID)
	admin := identityaccess.Actor{OrgID: orgID, UserID: adminUserID, SessionID: adminSession.ID}

	_, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, orgID)
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	documentService := documents.NewService(db)
	accountingService := accounting.NewService(db, documentService)

	createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "1100",
		Name:         "Accounts Receivable",
		AccountClass: accounting.AccountClassAsset,
		ControlType:  accounting.ControlTypeReceivable,
		Actor:        admin,
	})
	gstOutput := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "2201",
		Name:         "GST Output",
		AccountClass: accounting.AccountClassLiability,
		ControlType:  accounting.ControlTypeGSTOutput,
		Actor:        admin,
	})
	createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "GST18",
		Name:             "GST 18%",
		TaxType:          accounting.TaxTypeGST,
		RateBasisPoints:  1800,
		PayableAccountID: gstOutput.ID,
		Actor:            admin,
	})
	if _, err := accountingService.CreateAccountingPeriod(ctx, accounting.CreateAccountingPeriodInput{
		PeriodCode: "FY2026-04",
		StartOn:    time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndOn:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		Actor:      admin,
	}); err != nil {
		t.Fatalf("create accounting period: %v", err)
	}

	accounts, err := accountingService.ListLedgerAccounts(ctx, accounting.ListLedgerAccountsInput{Actor: admin})
	if err != nil {
		t.Fatalf("list ledger accounts: %v", err)
	}
	if len(accounts) != 2 || accounts[0].Code != "1100" || accounts[1].Code != "2201" {
		t.Fatalf("unexpected ledger accounts: %+v", accounts)
	}

	taxCodes, err := accountingService.ListTaxCodes(ctx, accounting.ListTaxCodesInput{Actor: admin})
	if err != nil {
		t.Fatalf("list tax codes: %v", err)
	}
	if len(taxCodes) != 1 || taxCodes[0].Code != "GST18" {
		t.Fatalf("unexpected tax codes: %+v", taxCodes)
	}

	periods, err := accountingService.ListAccountingPeriods(ctx, accounting.ListAccountingPeriodsInput{Actor: admin})
	if err != nil {
		t.Fatalf("list accounting periods: %v", err)
	}
	if len(periods) != 1 || periods[0].PeriodCode != "FY2026-04" {
		t.Fatalf("unexpected accounting periods: %+v", periods)
	}

	if _, err := accountingService.ListLedgerAccounts(ctx, accounting.ListLedgerAccountsInput{Actor: operator}); !errors.Is(err, identityaccess.ErrUnauthorized) {
		t.Fatalf("unexpected operator ledger-account list error: got %v want %v", err, identityaccess.ErrUnauthorized)
	}
}

func TestListJournalEntriesAndControlAccountBalancesIntegration(t *testing.T) {
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

	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	gstOutput := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "2101",
		Name:                "GST Output",
		AccountClass:        accounting.AccountClassLiability,
		ControlType:         accounting.ControlTypeGSTOutput,
		AllowsDirectPosting: false,
		Actor:               admin,
	})
	revenue := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:         "4000",
		Name:         "Service Revenue",
		AccountClass: accounting.AccountClassRevenue,
		Actor:        admin,
	})
	gst18 := createTaxCode(t, ctx, accountingService, accounting.CreateTaxCodeInput{
		Code:             "GST18",
		Name:             "GST Output 18%",
		TaxType:          accounting.TaxTypeGST,
		RateBasisPoints:  1800,
		PayableAccountID: gstOutput.ID,
		Actor:            admin,
	})

	dayOne := time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC)
	dayTwo := dayOne.Add(24 * time.Hour)
	dayThree := dayTwo.Add(24 * time.Hour)

	docOne := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)
	postOne, _, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   docOne.ID,
		Summary:      "Invoice one",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		EffectiveOn:  dayOne,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 177000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 150000},
			{AccountID: gstOutput.ID, Description: "GST payable", CreditMinor: 27000, TaxCode: gst18.Code},
		},
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("post first document: %v", err)
	}

	reversal, _, _, err := accountingService.ReverseDocument(ctx, accounting.ReverseDocumentInput{
		DocumentID:  docOne.ID,
		Reason:      "invoice corrected",
		EffectiveOn: dayTwo,
		Actor:       admin,
	})
	if err != nil {
		t.Fatalf("reverse first document: %v", err)
	}

	docTwo := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)
	postTwo, _, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   docTwo.ID,
		Summary:      "Invoice two",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		EffectiveOn:  dayThree,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 118000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 100000},
			{AccountID: gstOutput.ID, Description: "GST payable", CreditMinor: 18000, TaxCode: gst18.Code},
		},
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("post second document: %v", err)
	}

	reviews, err := accountingService.ListJournalEntries(ctx, accounting.ListJournalEntriesInput{
		StartOn: dayOne,
		EndOn:   dayThree,
		Limit:   10,
		Actor:   admin,
	})
	if err != nil {
		t.Fatalf("list journal entries: %v", err)
	}
	if len(reviews) != 3 {
		t.Fatalf("unexpected journal review count: %d", len(reviews))
	}
	if reviews[0].Entry.ID != postTwo.ID || reviews[0].Entry.EffectiveOn.Format(time.DateOnly) != "2026-03-22" {
		t.Fatalf("unexpected latest review entry: %+v", reviews[0].Entry)
	}
	if reviews[1].Entry.ID != reversal.ID || reviews[1].Entry.EntryKind != accounting.EntryKindReversal {
		t.Fatalf("unexpected middle review entry: %+v", reviews[1].Entry)
	}
	if reviews[2].Entry.ID != postOne.ID || !reviews[2].HasReversal {
		t.Fatalf("unexpected original review entry: %+v", reviews[2])
	}
	if reviews[2].DocumentTypeCode.String != "invoice" || reviews[2].DocumentStatus.String != string(documents.StatusReversed) {
		t.Fatalf("unexpected document linkage in review: %+v", reviews[2])
	}

	balancesDayOne, err := accountingService.ListControlAccountBalances(ctx, accounting.ListControlAccountBalancesInput{
		AsOf:  dayOne,
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("list control account balances as of day one: %v", err)
	}
	receivableDayOne := findControlAccountBalance(t, balancesDayOne, receivable.Code)
	if receivableDayOne.NetMinor != 177000 {
		t.Fatalf("unexpected day-one receivable balance: %+v", receivableDayOne)
	}
	gstDayOne := findControlAccountBalance(t, balancesDayOne, gstOutput.Code)
	if gstDayOne.NetMinor != -27000 {
		t.Fatalf("unexpected day-one gst balance: %+v", gstDayOne)
	}

	balancesDayTwo, err := accountingService.ListControlAccountBalances(ctx, accounting.ListControlAccountBalancesInput{
		AsOf:  dayTwo,
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("list control account balances as of day two: %v", err)
	}
	if got := findControlAccountBalance(t, balancesDayTwo, receivable.Code).NetMinor; got != 0 {
		t.Fatalf("unexpected day-two receivable balance: %d", got)
	}
	if got := findControlAccountBalance(t, balancesDayTwo, gstOutput.Code).NetMinor; got != 0 {
		t.Fatalf("unexpected day-two gst balance: %d", got)
	}

	balancesDayThree, err := accountingService.ListControlAccountBalances(ctx, accounting.ListControlAccountBalancesInput{
		AsOf:  dayThree,
		Actor: admin,
	})
	if err != nil {
		t.Fatalf("list control account balances as of day three: %v", err)
	}
	if got := findControlAccountBalance(t, balancesDayThree, receivable.Code).NetMinor; got != 118000 {
		t.Fatalf("unexpected day-three receivable balance: %d", got)
	}
	if got := findControlAccountBalance(t, balancesDayThree, gstOutput.Code).NetMinor; got != -18000 {
		t.Fatalf("unexpected day-three gst balance: %d", got)
	}
}

func TestJournalBalanceConstraintAtDatabaseBoundary(t *testing.T) {
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

	doc := prepareApprovedInvoiceDocument(t, ctx, accountingService, documentService, workflowService, operator, approver)
	receivable := createLedgerAccount(t, ctx, accountingService, accounting.CreateLedgerAccountInput{
		Code:                "1100",
		Name:                "Accounts Receivable",
		AccountClass:        accounting.AccountClassAsset,
		ControlType:         accounting.ControlTypeReceivable,
		AllowsDirectPosting: false,
		Actor:               admin,
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	var entryID string
	if err := tx.QueryRowContext(ctx, `
INSERT INTO accounting.journal_entries (
	org_id,
	entry_number,
	entry_kind,
	source_document_id,
	posting_fingerprint,
	currency_code,
	tax_scope_code,
	summary,
	posted_by_user_id
) VALUES ($1, 99, 'posting', $2, 'boundary-test', 'INR', 'none', 'Boundary test', $3)
RETURNING id;`,
		orgID,
		doc.ID,
		admin.UserID,
	).Scan(&entryID); err != nil {
		t.Fatalf("insert journal entry: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO accounting.journal_lines (
	org_id,
	entry_id,
	line_number,
	account_id,
	description,
	debit_minor,
	credit_minor
) VALUES ($1, $2, 1, $3, 'Only one side', 1000, 0);`,
		orgID,
		entryID,
		receivable.ID,
	); err != nil {
		t.Fatalf("insert journal line: %v", err)
	}

	err = tx.Commit()
	if err == nil {
		t.Fatal("expected commit to fail for unbalanced journal entry")
	}
	if !strings.Contains(err.Error(), "at least two lines") {
		t.Fatalf("unexpected commit error: %v", err)
	}
}

func prepareApprovedInvoiceDocument(t *testing.T, ctx context.Context, accountingService *accounting.Service, documentService *documents.Service, workflowService *workflow.Service, operator, approver identityaccess.Actor) documents.Document {
	t.Helper()

	doc, _, err := accountingService.CreateInvoice(ctx, accounting.CreateInvoiceInput{
		Title:        "Approved invoice",
		InvoiceRole:  accounting.InvoiceRoleSales,
		CurrencyCode: "INR",
		Summary:      "Approved invoice",
		Actor:        operator,
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}

	doc, err = submitAndApproveDocument(ctx, documentService, workflowService, operator, approver, doc)
	if err != nil {
		t.Fatalf("approve invoice: %v", err)
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

	doc, err = submitAndApproveDocument(ctx, documentService, workflowService, operator, approver, doc)
	if err != nil {
		t.Fatalf("approve document: %v", err)
	}

	return doc
}

func submitAndApproveDocument(ctx context.Context, documentService *documents.Service, workflowService *workflow.Service, operator, approver identityaccess.Actor, doc documents.Document) (documents.Document, error) {
	documentServiceDoc, err := documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		return documents.Document{}, err
	}

	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance-review",
		Reason:     "ready for posting review",
		Actor:      operator,
	})
	if err != nil {
		return documents.Document{}, err
	}

	_, documentServiceDoc, err = workflowService.DecideApproval(ctx, workflow.DecideApprovalInput{
		ApprovalID:   approval.ID,
		Decision:     "approved",
		DecisionNote: "approved for posting",
		Actor:        approver,
	})
	if err != nil {
		return documents.Document{}, err
	}

	return documentServiceDoc, nil
}

func createLedgerAccount(t *testing.T, ctx context.Context, service *accounting.Service, input accounting.CreateLedgerAccountInput) accounting.LedgerAccount {
	t.Helper()

	account, err := service.CreateLedgerAccount(ctx, input)
	if err != nil {
		t.Fatalf("create ledger account %s: %v", input.Code, err)
	}
	return account
}

func createTaxCode(t *testing.T, ctx context.Context, service *accounting.Service, input accounting.CreateTaxCodeInput) accounting.TaxCode {
	t.Helper()

	taxCode, err := service.CreateTaxCode(ctx, input)
	if err != nil {
		t.Fatalf("create tax code %s: %v", input.Code, err)
	}
	return taxCode
}

func createParty(t *testing.T, ctx context.Context, service *parties.Service, input parties.CreatePartyInput) parties.Party {
	t.Helper()

	party, err := service.CreateParty(ctx, input)
	if err != nil {
		t.Fatalf("create party %s: %v", input.PartyCode, err)
	}
	return party
}

func createContact(t *testing.T, ctx context.Context, service *parties.Service, input parties.CreateContactInput) parties.Contact {
	t.Helper()

	contact, err := service.CreateContact(ctx, input)
	if err != nil {
		t.Fatalf("create contact %s: %v", input.FullName, err)
	}
	return contact
}

func createInventoryItem(t *testing.T, ctx context.Context, service *inventoryops.Service, input inventoryops.CreateItemInput) inventoryops.Item {
	t.Helper()

	item, err := service.CreateItem(ctx, input)
	if err != nil {
		t.Fatalf("create inventory item %s: %v", input.SKU, err)
	}
	return item
}

func createInventoryLocation(t *testing.T, ctx context.Context, service *inventoryops.Service, input inventoryops.CreateLocationInput) inventoryops.Location {
	t.Helper()

	location, err := service.CreateLocation(ctx, input)
	if err != nil {
		t.Fatalf("create inventory location %s: %v", input.Code, err)
	}
	return location
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

func findControlAccountBalance(t *testing.T, balances []accounting.ControlAccountBalance, code string) accounting.ControlAccountBalance {
	t.Helper()

	for _, balance := range balances {
		if balance.AccountCode == code {
			return balance
		}
	}
	t.Fatalf("control account balance %s not found", code)
	return accounting.ControlAccountBalance{}
}
