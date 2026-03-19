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

	postedEntry, postedLines, _, err := accountingService.PostDocument(ctx, accounting.PostDocumentInput{
		DocumentID:   doc.ID,
		Summary:      "Post approved invoice",
		CurrencyCode: "INR",
		TaxScopeCode: accounting.TaxScopeGST,
		Lines: []accounting.PostingLineInput{
			{AccountID: receivable.ID, Description: "Customer receivable", DebitMinor: 150000},
			{AccountID: revenue.ID, Description: "Recognized revenue", CreditMinor: 150000, TaxCode: "GST18"},
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
