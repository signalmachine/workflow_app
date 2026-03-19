package accounting

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
)

var (
	ErrLedgerAccountNotFound = errors.New("ledger account not found")
	ErrJournalEntryNotFound  = errors.New("journal entry not found")
	ErrTaxCodeNotFound       = errors.New("tax code not found")
	ErrPostingAlreadyExists  = errors.New("posting already exists for document")
	ErrAlreadyReversed       = errors.New("journal entry already reversed")
	ErrInvalidAccount        = errors.New("invalid ledger account")
	ErrInvalidTaxCode        = errors.New("invalid tax code")
	ErrInvalidCurrencyCode   = errors.New("invalid currency code")
	ErrInvalidTaxScope       = errors.New("invalid tax scope")
	ErrInvalidJournalLine    = errors.New("invalid journal line")
	ErrInvalidReversal       = errors.New("invalid reversal")
	ErrUnbalancedJournal     = errors.New("journal entry is unbalanced")
)

const (
	AccountClassAsset     = "asset"
	AccountClassLiability = "liability"
	AccountClassEquity    = "equity"
	AccountClassRevenue   = "revenue"
	AccountClassExpense   = "expense"

	ControlTypeNone          = "none"
	ControlTypeReceivable    = "receivable"
	ControlTypePayable       = "payable"
	ControlTypeGSTInput      = "gst_input"
	ControlTypeGSTOutput     = "gst_output"
	ControlTypeTDSReceivable = "tds_receivable"
	ControlTypeTDSPayable    = "tds_payable"

	EntryKindPosting  = "posting"
	EntryKindReversal = "reversal"

	TaxScopeNone   = "none"
	TaxScopeGST    = "gst"
	TaxScopeTDS    = "tds"
	TaxScopeGSTTDS = "gst_tds"

	TaxTypeGST = "gst"
	TaxTypeTDS = "tds"
)

type LedgerAccount struct {
	ID                  string
	OrgID               string
	Code                string
	Name                string
	AccountClass        string
	ControlType         string
	AllowsDirectPosting bool
	Status              string
	TaxCategoryCode     sql.NullString
	CreatedByUserID     string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type JournalEntry struct {
	ID                 string
	OrgID              string
	EntryNumber        int64
	EntryKind          string
	SourceDocumentID   sql.NullString
	ReversalOfEntryID  sql.NullString
	PostingFingerprint string
	CurrencyCode       string
	TaxScopeCode       string
	Summary            string
	ReversalReason     sql.NullString
	PostedByUserID     string
	PostedAt           time.Time
	CreatedAt          time.Time
}

type JournalLine struct {
	ID          string
	OrgID       string
	EntryID     string
	LineNumber  int
	AccountID   string
	Description string
	DebitMinor  int64
	CreditMinor int64
	TaxCode     sql.NullString
	CreatedAt   time.Time
}

type TaxCode struct {
	ID                  string
	OrgID               string
	Code                string
	Name                string
	TaxType             string
	RateBasisPoints     int
	ReceivableAccountID sql.NullString
	PayableAccountID    sql.NullString
	Status              string
	CreatedByUserID     string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CreateLedgerAccountInput struct {
	Code                string
	Name                string
	AccountClass        string
	ControlType         string
	AllowsDirectPosting bool
	TaxCategoryCode     string
	Actor               identityaccess.Actor
}

type CreateTaxCodeInput struct {
	Code                string
	Name                string
	TaxType             string
	RateBasisPoints     int
	ReceivableAccountID string
	PayableAccountID    string
	Actor               identityaccess.Actor
}

type PostingLineInput struct {
	AccountID   string
	Description string
	DebitMinor  int64
	CreditMinor int64
	TaxCode     string
}

type PostDocumentInput struct {
	DocumentID   string
	Summary      string
	CurrencyCode string
	TaxScopeCode string
	Lines        []PostingLineInput
	Actor        identityaccess.Actor
}

type ReverseDocumentInput struct {
	DocumentID string
	Reason     string
	Actor      identityaccess.Actor
}

type Service struct {
	db        *sql.DB
	documents *documents.Service
}

func NewService(db *sql.DB, documentService *documents.Service) *Service {
	return &Service{db: db, documents: documentService}
}

func (s *Service) CreateLedgerAccount(ctx context.Context, input CreateLedgerAccountInput) (LedgerAccount, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return LedgerAccount{}, fmt.Errorf("begin create ledger account: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin); err != nil {
		_ = tx.Rollback()
		return LedgerAccount{}, err
	}

	account, err := createLedgerAccountTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return LedgerAccount{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.ledger_account_created",
		EntityType:  "accounting.ledger_account",
		EntityID:    account.ID,
		Payload: map[string]any{
			"code":          account.Code,
			"account_class": account.AccountClass,
			"control_type":  account.ControlType,
		},
	}); err != nil {
		_ = tx.Rollback()
		return LedgerAccount{}, err
	}

	if err := tx.Commit(); err != nil {
		return LedgerAccount{}, fmt.Errorf("commit create ledger account: %w", err)
	}

	return account, nil
}

func (s *Service) CreateTaxCode(ctx context.Context, input CreateTaxCodeInput) (TaxCode, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TaxCode{}, fmt.Errorf("begin create tax code: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin); err != nil {
		_ = tx.Rollback()
		return TaxCode{}, err
	}

	taxCode, err := createTaxCodeTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return TaxCode{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.tax_code_created",
		EntityType:  "accounting.tax_code",
		EntityID:    taxCode.ID,
		Payload: map[string]any{
			"code":               taxCode.Code,
			"tax_type":           taxCode.TaxType,
			"rate_basis_points":  taxCode.RateBasisPoints,
			"receivable_account": taxCode.ReceivableAccountID.String,
			"payable_account":    taxCode.PayableAccountID.String,
		},
	}); err != nil {
		_ = tx.Rollback()
		return TaxCode{}, err
	}

	if err := tx.Commit(); err != nil {
		return TaxCode{}, fmt.Errorf("commit create tax code: %w", err)
	}

	return taxCode, nil
}

func (s *Service) PostDocument(ctx context.Context, input PostDocumentInput) (JournalEntry, []JournalLine, documents.Document, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JournalEntry{}, nil, documents.Document{}, fmt.Errorf("begin post document: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleApprover); err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	fingerprint, err := postingFingerprint(input)
	if err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	entry, lines, existing, err := s.postDocumentTx(ctx, tx, input, fingerprint)
	if err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}
	if existing {
		if err := tx.Commit(); err != nil {
			return JournalEntry{}, nil, documents.Document{}, fmt.Errorf("commit idempotent post document: %w", err)
		}
		doc, err := s.loadDocument(ctx, input.Actor.OrgID, input.DocumentID)
		return entry, lines, doc, err
	}

	doc, err := s.documents.ApplyPostingOutcome(ctx, tx, documents.PostingOutcomeInput{
		DocumentID: input.DocumentID,
		Action:     "posted",
		Actor:      input.Actor,
	})
	if err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.document_posted",
		EntityType:  "accounting.journal_entry",
		EntityID:    entry.ID,
		Payload: map[string]any{
			"document_id":   input.DocumentID,
			"entry_number":  entry.EntryNumber,
			"currency_code": entry.CurrencyCode,
			"tax_scope":     entry.TaxScopeCode,
		},
	}); err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "documents.document_posted",
		EntityType:  "documents.document",
		EntityID:    doc.ID,
		Payload: map[string]any{
			"journal_entry_id": entry.ID,
			"status":           doc.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	if err := tx.Commit(); err != nil {
		return JournalEntry{}, nil, documents.Document{}, fmt.Errorf("commit post document: %w", err)
	}

	return entry, lines, doc, nil
}

func (s *Service) ReverseDocument(ctx context.Context, input ReverseDocumentInput) (JournalEntry, []JournalLine, documents.Document, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JournalEntry{}, nil, documents.Document{}, fmt.Errorf("begin reverse document: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleApprover); err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	original, originalLines, _, err := getPostingEntryByDocumentForUpdate(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	existingReversal, reversalLines, found, err := getReversalEntryForUpdate(ctx, tx, input.Actor.OrgID, original.ID)
	if err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}
	if found {
		if !existingReversal.ReversalReason.Valid || existingReversal.ReversalReason.String != strings.TrimSpace(input.Reason) {
			_ = tx.Rollback()
			return JournalEntry{}, nil, documents.Document{}, ErrAlreadyReversed
		}
		if err := tx.Commit(); err != nil {
			return JournalEntry{}, nil, documents.Document{}, fmt.Errorf("commit idempotent reverse document: %w", err)
		}
		doc, err := s.loadDocument(ctx, input.Actor.OrgID, input.DocumentID)
		return existingReversal, reversalLines, doc, err
	}

	reversal, lines, err := createReversalTx(ctx, tx, original, originalLines, strings.TrimSpace(input.Reason), input.Actor)
	if err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	doc, err := s.documents.ApplyPostingOutcome(ctx, tx, documents.PostingOutcomeInput{
		DocumentID: input.DocumentID,
		Action:     "reversed",
		Actor:      input.Actor,
	})
	if err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.document_reversed",
		EntityType:  "accounting.journal_entry",
		EntityID:    reversal.ID,
		Payload: map[string]any{
			"document_id":        input.DocumentID,
			"reversal_of_entry":  original.ID,
			"reversal_entry_num": reversal.EntryNumber,
			"reason":             input.Reason,
		},
	}); err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "documents.document_reversed",
		EntityType:  "documents.document",
		EntityID:    doc.ID,
		Payload: map[string]any{
			"journal_entry_id": reversal.ID,
			"status":           doc.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return JournalEntry{}, nil, documents.Document{}, err
	}

	if err := tx.Commit(); err != nil {
		return JournalEntry{}, nil, documents.Document{}, fmt.Errorf("commit reverse document: %w", err)
	}

	return reversal, lines, doc, nil
}

func (s *Service) postDocumentTx(ctx context.Context, tx *sql.Tx, input PostDocumentInput, fingerprint string) (JournalEntry, []JournalLine, bool, error) {
	if strings.TrimSpace(input.DocumentID) == "" || strings.TrimSpace(input.Summary) == "" {
		return JournalEntry{}, nil, false, ErrInvalidJournalLine
	}
	if !isValidCurrencyCode(input.CurrencyCode) {
		return JournalEntry{}, nil, false, ErrInvalidCurrencyCode
	}
	if !isValidTaxScope(input.TaxScopeCode) {
		return JournalEntry{}, nil, false, ErrInvalidTaxScope
	}
	if err := validatePostingLines(input.Lines); err != nil {
		return JournalEntry{}, nil, false, err
	}
	if err := validateAccountsTx(ctx, tx, input.Actor.OrgID, input.Lines); err != nil {
		return JournalEntry{}, nil, false, err
	}
	if err := validatePostingTaxCodesTx(ctx, tx, input.Actor.OrgID, input.TaxScopeCode, input.Lines); err != nil {
		return JournalEntry{}, nil, false, err
	}

	existing, lines, found, err := getPostingEntryByDocumentForUpdate(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil && !errors.Is(err, ErrJournalEntryNotFound) {
		return JournalEntry{}, nil, false, err
	}
	if found {
		if existing.PostingFingerprint != fingerprint {
			return JournalEntry{}, nil, false, ErrPostingAlreadyExists
		}
		return existing, lines, true, nil
	}

	entryNumber, err := reserveEntryNumberTx(ctx, tx, input.Actor.OrgID)
	if err != nil {
		return JournalEntry{}, nil, false, err
	}

	entry, err := insertJournalEntryTx(ctx, tx, insertJournalEntryParams{
		OrgID:              input.Actor.OrgID,
		EntryNumber:        entryNumber,
		EntryKind:          EntryKindPosting,
		SourceDocumentID:   input.DocumentID,
		PostingFingerprint: fingerprint,
		CurrencyCode:       strings.ToUpper(input.CurrencyCode),
		TaxScopeCode:       input.TaxScopeCode,
		Summary:            strings.TrimSpace(input.Summary),
		PostedByUserID:     input.Actor.UserID,
	})
	if err != nil {
		return JournalEntry{}, nil, false, err
	}

	lines, err = insertJournalLinesTx(ctx, tx, input.Actor.OrgID, entry.ID, input.Lines)
	if err != nil {
		return JournalEntry{}, nil, false, err
	}

	return entry, lines, false, nil
}

func createLedgerAccountTx(ctx context.Context, tx *sql.Tx, input CreateLedgerAccountInput) (LedgerAccount, error) {
	accountClass := strings.TrimSpace(input.AccountClass)
	controlType := strings.TrimSpace(input.ControlType)
	if controlType == "" {
		controlType = ControlTypeNone
	}
	if !isValidAccountClass(accountClass) || !isValidControlType(controlType) {
		return LedgerAccount{}, ErrInvalidAccount
	}

	const statement = `
INSERT INTO accounting.ledger_accounts (
	org_id,
	code,
	name,
	account_class,
	control_type,
	allows_direct_posting,
	tax_category_code,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
	id,
	org_id,
	code,
	name,
	account_class,
	control_type,
	allows_direct_posting,
	status,
	tax_category_code,
	created_by_user_id,
	created_at,
	updated_at;`

	return scanLedgerAccount(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		strings.TrimSpace(input.Code),
		strings.TrimSpace(input.Name),
		accountClass,
		controlType,
		input.AllowsDirectPosting,
		nullIfEmpty(strings.TrimSpace(input.TaxCategoryCode)),
		input.Actor.UserID,
	))
}

func createTaxCodeTx(ctx context.Context, tx *sql.Tx, input CreateTaxCodeInput) (TaxCode, error) {
	taxType := strings.TrimSpace(input.TaxType)
	if !isValidTaxType(taxType) {
		return TaxCode{}, ErrInvalidTaxCode
	}
	if strings.TrimSpace(input.Code) == "" || strings.TrimSpace(input.Name) == "" {
		return TaxCode{}, ErrInvalidTaxCode
	}
	if input.RateBasisPoints < 0 || input.RateBasisPoints > 10000 {
		return TaxCode{}, ErrInvalidTaxCode
	}
	if strings.TrimSpace(input.ReceivableAccountID) == "" && strings.TrimSpace(input.PayableAccountID) == "" {
		return TaxCode{}, ErrInvalidTaxCode
	}
	if err := validateTaxControlAccountsTx(ctx, tx, input.Actor.OrgID, taxType, input.ReceivableAccountID, input.PayableAccountID); err != nil {
		return TaxCode{}, err
	}

	const statement = `
INSERT INTO accounting.tax_codes (
	org_id,
	code,
	name,
	tax_type,
	rate_basis_points,
	receivable_account_id,
	payable_account_id,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
	id,
	org_id,
	code,
	name,
	tax_type,
	rate_basis_points,
	receivable_account_id,
	payable_account_id,
	status,
	created_by_user_id,
	created_at,
	updated_at;`

	return scanTaxCode(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		strings.TrimSpace(input.Code),
		strings.TrimSpace(input.Name),
		taxType,
		input.RateBasisPoints,
		nullIfEmpty(strings.TrimSpace(input.ReceivableAccountID)),
		nullIfEmpty(strings.TrimSpace(input.PayableAccountID)),
		input.Actor.UserID,
	))
}

func validatePostingLines(lines []PostingLineInput) error {
	if len(lines) < 2 {
		return ErrInvalidJournalLine
	}

	var debitTotal int64
	var creditTotal int64
	for _, line := range lines {
		if strings.TrimSpace(line.AccountID) == "" {
			return ErrInvalidJournalLine
		}
		if (line.DebitMinor > 0 && line.CreditMinor > 0) || (line.DebitMinor == 0 && line.CreditMinor == 0) {
			return ErrInvalidJournalLine
		}
		if line.DebitMinor < 0 || line.CreditMinor < 0 {
			return ErrInvalidJournalLine
		}
		debitTotal += line.DebitMinor
		creditTotal += line.CreditMinor
	}
	if debitTotal != creditTotal {
		return ErrUnbalancedJournal
	}
	return nil
}

func validateAccountsTx(ctx context.Context, tx *sql.Tx, orgID string, lines []PostingLineInput) error {
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		accountID := strings.TrimSpace(line.AccountID)
		if _, ok := seen[accountID]; ok {
			continue
		}
		seen[accountID] = struct{}{}

		const query = `
SELECT 1
FROM accounting.ledger_accounts
WHERE org_id = $1
  AND id = $2
  AND status = 'active';`

		var exists int
		if err := tx.QueryRowContext(ctx, query, orgID, accountID).Scan(&exists); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrLedgerAccountNotFound
			}
			return fmt.Errorf("validate ledger account: %w", err)
		}
	}
	return nil
}

func validatePostingTaxCodesTx(ctx context.Context, tx *sql.Tx, orgID, taxScopeCode string, lines []PostingLineInput) error {
	scope := strings.TrimSpace(taxScopeCode)
	seen := make(map[string]struct{}, len(lines))
	hasTaxCode := false

	for _, line := range lines {
		code := strings.TrimSpace(line.TaxCode)
		if code == "" {
			continue
		}
		hasTaxCode = true
		if scope == TaxScopeNone {
			return ErrInvalidTaxScope
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}

		taxCode, err := getActiveTaxCodeTx(ctx, tx, orgID, code)
		if err != nil {
			return err
		}
		if !taxScopeAllowsType(scope, taxCode.TaxType) {
			return ErrInvalidTaxScope
		}
	}

	if scope != TaxScopeNone && !hasTaxCode {
		return ErrInvalidTaxScope
	}

	return nil
}

func validateTaxControlAccountsTx(ctx context.Context, tx *sql.Tx, orgID, taxType, receivableAccountID, payableAccountID string) error {
	if accountID := strings.TrimSpace(receivableAccountID); accountID != "" {
		expected := ControlTypeGSTInput
		if taxType == TaxTypeTDS {
			expected = ControlTypeTDSReceivable
		}
		if err := validateControlAccountTypeTx(ctx, tx, orgID, accountID, expected); err != nil {
			return err
		}
	}

	if accountID := strings.TrimSpace(payableAccountID); accountID != "" {
		expected := ControlTypeGSTOutput
		if taxType == TaxTypeTDS {
			expected = ControlTypeTDSPayable
		}
		if err := validateControlAccountTypeTx(ctx, tx, orgID, accountID, expected); err != nil {
			return err
		}
	}

	return nil
}

func validateControlAccountTypeTx(ctx context.Context, tx *sql.Tx, orgID, accountID, expectedControlType string) error {
	const query = `
SELECT control_type
FROM accounting.ledger_accounts
WHERE org_id = $1
  AND id = $2
  AND status = 'active';`

	var controlType string
	if err := tx.QueryRowContext(ctx, query, orgID, accountID).Scan(&controlType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLedgerAccountNotFound
		}
		return fmt.Errorf("validate control account: %w", err)
	}
	if controlType != expectedControlType {
		return ErrInvalidTaxCode
	}
	return nil
}

type insertJournalEntryParams struct {
	OrgID              string
	EntryNumber        int64
	EntryKind          string
	SourceDocumentID   string
	ReversalOfEntryID  string
	PostingFingerprint string
	CurrencyCode       string
	TaxScopeCode       string
	Summary            string
	ReversalReason     string
	PostedByUserID     string
}

func insertJournalEntryTx(ctx context.Context, tx *sql.Tx, params insertJournalEntryParams) (JournalEntry, error) {
	const statement = `
INSERT INTO accounting.journal_entries (
	org_id,
	entry_number,
	entry_kind,
	source_document_id,
	reversal_of_entry_id,
	posting_fingerprint,
	currency_code,
	tax_scope_code,
	summary,
	reversal_reason,
	posted_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING
	id,
	org_id,
	entry_number,
	entry_kind,
	source_document_id,
	reversal_of_entry_id,
	posting_fingerprint,
	currency_code,
	tax_scope_code,
	summary,
	reversal_reason,
	posted_by_user_id,
	posted_at,
	created_at;`

	return scanJournalEntry(tx.QueryRowContext(
		ctx,
		statement,
		params.OrgID,
		params.EntryNumber,
		params.EntryKind,
		nullIfEmpty(params.SourceDocumentID),
		nullIfEmpty(params.ReversalOfEntryID),
		params.PostingFingerprint,
		params.CurrencyCode,
		params.TaxScopeCode,
		params.Summary,
		nullIfEmpty(params.ReversalReason),
		params.PostedByUserID,
	))
}

func insertJournalLinesTx(ctx context.Context, tx *sql.Tx, orgID, entryID string, inputs []PostingLineInput) ([]JournalLine, error) {
	const statement = `
INSERT INTO accounting.journal_lines (
	org_id,
	entry_id,
	line_number,
	account_id,
	description,
	debit_minor,
	credit_minor,
	tax_code
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
	id,
	org_id,
	entry_id,
	line_number,
	account_id,
	description,
	debit_minor,
	credit_minor,
	tax_code,
	created_at;`

	lines := make([]JournalLine, 0, len(inputs))
	for idx, line := range inputs {
		created, err := scanJournalLine(tx.QueryRowContext(
			ctx,
			statement,
			orgID,
			entryID,
			idx+1,
			line.AccountID,
			strings.TrimSpace(line.Description),
			line.DebitMinor,
			line.CreditMinor,
			nullIfEmpty(strings.TrimSpace(line.TaxCode)),
		))
		if err != nil {
			return nil, fmt.Errorf("insert journal line: %w", err)
		}
		lines = append(lines, created)
	}
	return lines, nil
}

func reserveEntryNumberTx(ctx context.Context, tx *sql.Tx, orgID string) (int64, error) {
	const upsert = `
INSERT INTO accounting.journal_numbering_series (org_id, next_number)
VALUES ($1, 1)
ON CONFLICT (org_id) DO NOTHING;`

	if _, err := tx.ExecContext(ctx, upsert, orgID); err != nil {
		return 0, fmt.Errorf("initialize journal numbering series: %w", err)
	}

	const statement = `
UPDATE accounting.journal_numbering_series
SET next_number = next_number + 1,
	updated_at = NOW()
WHERE org_id = $1
RETURNING next_number - 1;`

	var entryNumber int64
	if err := tx.QueryRowContext(ctx, statement, orgID).Scan(&entryNumber); err != nil {
		return 0, fmt.Errorf("reserve journal entry number: %w", err)
	}
	return entryNumber, nil
}

func getPostingEntryByDocumentForUpdate(ctx context.Context, tx *sql.Tx, orgID, documentID string) (JournalEntry, []JournalLine, bool, error) {
	const query = `
SELECT
	id,
	org_id,
	entry_number,
	entry_kind,
	source_document_id,
	reversal_of_entry_id,
	posting_fingerprint,
	currency_code,
	tax_scope_code,
	summary,
	reversal_reason,
	posted_by_user_id,
	posted_at,
	created_at
FROM accounting.journal_entries
WHERE org_id = $1
  AND source_document_id = $2
  AND entry_kind = 'posting'
FOR UPDATE;`

	entry, err := scanJournalEntry(tx.QueryRowContext(ctx, query, orgID, documentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JournalEntry{}, nil, false, ErrJournalEntryNotFound
		}
		return JournalEntry{}, nil, false, fmt.Errorf("load posting journal entry: %w", err)
	}

	lines, err := listJournalLinesTx(ctx, tx, entry.ID)
	if err != nil {
		return JournalEntry{}, nil, false, err
	}

	return entry, lines, true, nil
}

func getReversalEntryForUpdate(ctx context.Context, tx *sql.Tx, orgID, entryID string) (JournalEntry, []JournalLine, bool, error) {
	const query = `
SELECT
	id,
	org_id,
	entry_number,
	entry_kind,
	source_document_id,
	reversal_of_entry_id,
	posting_fingerprint,
	currency_code,
	tax_scope_code,
	summary,
	reversal_reason,
	posted_by_user_id,
	posted_at,
	created_at
FROM accounting.journal_entries
WHERE org_id = $1
  AND reversal_of_entry_id = $2
FOR UPDATE;`

	entry, err := scanJournalEntry(tx.QueryRowContext(ctx, query, orgID, entryID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JournalEntry{}, nil, false, nil
		}
		return JournalEntry{}, nil, false, fmt.Errorf("load reversal journal entry: %w", err)
	}

	lines, err := listJournalLinesTx(ctx, tx, entry.ID)
	if err != nil {
		return JournalEntry{}, nil, false, err
	}

	return entry, lines, true, nil
}

func createReversalTx(ctx context.Context, tx *sql.Tx, original JournalEntry, originalLines []JournalLine, reason string, actor identityaccess.Actor) (JournalEntry, []JournalLine, error) {
	if strings.TrimSpace(reason) == "" {
		return JournalEntry{}, nil, ErrInvalidReversal
	}

	entryNumber, err := reserveEntryNumberTx(ctx, tx, actor.OrgID)
	if err != nil {
		return JournalEntry{}, nil, err
	}

	fingerprint, err := reversalFingerprint(original.ID, reason)
	if err != nil {
		return JournalEntry{}, nil, err
	}

	reversal, err := insertJournalEntryTx(ctx, tx, insertJournalEntryParams{
		OrgID:              actor.OrgID,
		EntryNumber:        entryNumber,
		EntryKind:          EntryKindReversal,
		ReversalOfEntryID:  original.ID,
		PostingFingerprint: fingerprint,
		CurrencyCode:       original.CurrencyCode,
		TaxScopeCode:       original.TaxScopeCode,
		Summary:            "Reversal of entry " + fmt.Sprint(original.EntryNumber),
		ReversalReason:     reason,
		PostedByUserID:     actor.UserID,
	})
	if err != nil {
		return JournalEntry{}, nil, err
	}

	reversalInputs := make([]PostingLineInput, 0, len(originalLines))
	for _, line := range originalLines {
		reversalInputs = append(reversalInputs, PostingLineInput{
			AccountID:   line.AccountID,
			Description: line.Description,
			DebitMinor:  line.CreditMinor,
			CreditMinor: line.DebitMinor,
			TaxCode:     line.TaxCode.String,
		})
	}

	lines, err := insertJournalLinesTx(ctx, tx, actor.OrgID, reversal.ID, reversalInputs)
	if err != nil {
		return JournalEntry{}, nil, err
	}

	return reversal, lines, nil
}

func listJournalLinesTx(ctx context.Context, tx *sql.Tx, entryID string) ([]JournalLine, error) {
	const query = `
SELECT
	id,
	org_id,
	entry_id,
	line_number,
	account_id,
	description,
	debit_minor,
	credit_minor,
	tax_code,
	created_at
FROM accounting.journal_lines
WHERE entry_id = $1
ORDER BY line_number ASC;`

	rows, err := tx.QueryContext(ctx, query, entryID)
	if err != nil {
		return nil, fmt.Errorf("list journal lines: %w", err)
	}
	defer rows.Close()

	var lines []JournalLine
	for rows.Next() {
		line, err := scanJournalLine(rows)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate journal lines: %w", err)
	}
	return lines, nil
}

func getActiveTaxCodeTx(ctx context.Context, tx *sql.Tx, orgID, code string) (TaxCode, error) {
	const query = `
SELECT
	id,
	org_id,
	code,
	name,
	tax_type,
	rate_basis_points,
	receivable_account_id,
	payable_account_id,
	status,
	created_by_user_id,
	created_at,
	updated_at
FROM accounting.tax_codes
WHERE org_id = $1
  AND lower(code) = lower($2)
  AND status = 'active';`

	taxCode, err := scanTaxCode(tx.QueryRowContext(ctx, query, orgID, strings.TrimSpace(code)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TaxCode{}, ErrTaxCodeNotFound
		}
		return TaxCode{}, fmt.Errorf("load tax code: %w", err)
	}
	return taxCode, nil
}

func (s *Service) loadDocument(ctx context.Context, orgID, documentID string) (documents.Document, error) {
	const query = `
SELECT
	id,
	org_id,
	type_code,
	status,
	title,
	number_series_id,
	number_value,
	source_document_id,
	created_by_user_id,
	submitted_by_user_id,
	submitted_at,
	approved_at,
	rejected_at,
	created_at,
	updated_at
FROM documents.documents
WHERE org_id = $1
  AND id = $2;`

	doc, err := scanDocument(txRow{s.db.QueryRowContext(ctx, query, orgID, documentID)})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return documents.Document{}, documents.ErrDocumentNotFound
		}
		return documents.Document{}, fmt.Errorf("load document: %w", err)
	}
	return doc, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

type txRow struct {
	row *sql.Row
}

func (r txRow) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

func scanLedgerAccount(row rowScanner) (LedgerAccount, error) {
	var account LedgerAccount
	err := row.Scan(
		&account.ID,
		&account.OrgID,
		&account.Code,
		&account.Name,
		&account.AccountClass,
		&account.ControlType,
		&account.AllowsDirectPosting,
		&account.Status,
		&account.TaxCategoryCode,
		&account.CreatedByUserID,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return LedgerAccount{}, err
	}
	return account, nil
}

func scanJournalEntry(row rowScanner) (JournalEntry, error) {
	var entry JournalEntry
	err := row.Scan(
		&entry.ID,
		&entry.OrgID,
		&entry.EntryNumber,
		&entry.EntryKind,
		&entry.SourceDocumentID,
		&entry.ReversalOfEntryID,
		&entry.PostingFingerprint,
		&entry.CurrencyCode,
		&entry.TaxScopeCode,
		&entry.Summary,
		&entry.ReversalReason,
		&entry.PostedByUserID,
		&entry.PostedAt,
		&entry.CreatedAt,
	)
	if err != nil {
		return JournalEntry{}, err
	}
	return entry, nil
}

func scanJournalLine(row rowScanner) (JournalLine, error) {
	var line JournalLine
	err := row.Scan(
		&line.ID,
		&line.OrgID,
		&line.EntryID,
		&line.LineNumber,
		&line.AccountID,
		&line.Description,
		&line.DebitMinor,
		&line.CreditMinor,
		&line.TaxCode,
		&line.CreatedAt,
	)
	if err != nil {
		return JournalLine{}, err
	}
	return line, nil
}

func scanTaxCode(row rowScanner) (TaxCode, error) {
	var taxCode TaxCode
	err := row.Scan(
		&taxCode.ID,
		&taxCode.OrgID,
		&taxCode.Code,
		&taxCode.Name,
		&taxCode.TaxType,
		&taxCode.RateBasisPoints,
		&taxCode.ReceivableAccountID,
		&taxCode.PayableAccountID,
		&taxCode.Status,
		&taxCode.CreatedByUserID,
		&taxCode.CreatedAt,
		&taxCode.UpdatedAt,
	)
	if err != nil {
		return TaxCode{}, err
	}
	return taxCode, nil
}

func scanDocument(row rowScanner) (documents.Document, error) {
	var doc documents.Document
	err := row.Scan(
		&doc.ID,
		&doc.OrgID,
		&doc.TypeCode,
		&doc.Status,
		&doc.Title,
		&doc.NumberSeriesID,
		&doc.NumberValue,
		&doc.SourceDocumentID,
		&doc.CreatedByUserID,
		&doc.SubmittedByUserID,
		&doc.SubmittedAt,
		&doc.ApprovedAt,
		&doc.RejectedAt,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return documents.Document{}, err
	}
	return doc, nil
}

func postingFingerprint(input PostDocumentInput) (string, error) {
	payload := struct {
		DocumentID   string             `json:"document_id"`
		Summary      string             `json:"summary"`
		CurrencyCode string             `json:"currency_code"`
		TaxScopeCode string             `json:"tax_scope_code"`
		Lines        []PostingLineInput `json:"lines"`
	}{
		DocumentID:   strings.TrimSpace(input.DocumentID),
		Summary:      strings.TrimSpace(input.Summary),
		CurrencyCode: strings.ToUpper(input.CurrencyCode),
		TaxScopeCode: strings.TrimSpace(input.TaxScopeCode),
		Lines:        input.Lines,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal posting fingerprint: %w", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(body)), nil
}

func reversalFingerprint(entryID, reason string) (string, error) {
	payload := struct {
		EntryID string `json:"entry_id"`
		Reason  string `json:"reason"`
	}{
		EntryID: entryID,
		Reason:  strings.TrimSpace(reason),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal reversal fingerprint: %w", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(body)), nil
}

func isValidAccountClass(value string) bool {
	switch value {
	case AccountClassAsset, AccountClassLiability, AccountClassEquity, AccountClassRevenue, AccountClassExpense:
		return true
	default:
		return false
	}
}

func isValidControlType(value string) bool {
	switch value {
	case ControlTypeNone, ControlTypeReceivable, ControlTypePayable, ControlTypeGSTInput, ControlTypeGSTOutput, ControlTypeTDSReceivable, ControlTypeTDSPayable:
		return true
	default:
		return false
	}
}

func isValidCurrencyCode(value string) bool {
	value = strings.ToUpper(strings.TrimSpace(value))
	return len(value) == 3 && value == strings.ToUpper(value)
}

func isValidTaxScope(value string) bool {
	switch strings.TrimSpace(value) {
	case TaxScopeNone, TaxScopeGST, TaxScopeTDS, TaxScopeGSTTDS:
		return true
	default:
		return false
	}
}

func isValidTaxType(value string) bool {
	switch strings.TrimSpace(value) {
	case TaxTypeGST, TaxTypeTDS:
		return true
	default:
		return false
	}
}

func taxScopeAllowsType(scope, taxType string) bool {
	switch strings.TrimSpace(scope) {
	case TaxScopeGST:
		return taxType == TaxTypeGST
	case TaxScopeTDS:
		return taxType == TaxTypeTDS
	case TaxScopeGSTTDS:
		return taxType == TaxTypeGST || taxType == TaxTypeTDS
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
