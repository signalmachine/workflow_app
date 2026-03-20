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
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
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

	doc := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)

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

	doc := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)

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

	docWithoutTaxCode := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)
	docWithUnknownTaxCode := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)
	docWithWrongTaxType := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)

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

	doc := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)

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

	docForPosting := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)
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

	docBlockedByClosedPeriod := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)
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

	docOne := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)
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

	docTwo := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)
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

	doc := prepareApprovedDocument(t, ctx, documentService, workflowService, operator, approver)
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

func prepareApprovedDocument(t *testing.T, ctx context.Context, documentService *documents.Service, workflowService *workflow.Service, operator, approver identityaccess.Actor) documents.Document {
	t.Helper()

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Approved invoice",
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
		Reason:     "ready for posting review",
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}

	_, doc, err = workflowService.DecideApproval(ctx, workflow.DecideApprovalInput{
		ApprovalID:   approval.ID,
		Decision:     "approved",
		DecisionNote: "approved for posting",
		Actor:        approver,
	})
	if err != nil {
		t.Fatalf("decide approval: %v", err)
	}

	return doc
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
