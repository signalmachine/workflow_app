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
	ErrLedgerAccountNotFound    = errors.New("ledger account not found")
	ErrJournalEntryNotFound     = errors.New("journal entry not found")
	ErrTaxCodeNotFound          = errors.New("tax code not found")
	ErrAccountingPeriodNotFound = errors.New("accounting period not found")
	ErrInvoiceDocumentNotFound  = errors.New("invoice document not found")
	ErrPaymentReceiptNotFound   = errors.New("payment or receipt document not found")
	ErrPostingAlreadyExists     = errors.New("posting already exists for document")
	ErrAlreadyReversed          = errors.New("journal entry already reversed")
	ErrInvalidAccount           = errors.New("invalid ledger account")
	ErrInvalidTaxCode           = errors.New("invalid tax code")
	ErrInvalidCurrencyCode      = errors.New("invalid currency code")
	ErrInvalidTaxScope          = errors.New("invalid tax scope")
	ErrInvalidJournalLine       = errors.New("invalid journal line")
	ErrInvalidReversal          = errors.New("invalid reversal")
	ErrInvalidAccountingPeriod  = errors.New("invalid accounting period")
	ErrAccountingPeriodOverlap  = errors.New("accounting period overlaps an existing period")
	ErrAccountingPeriodNotOpen  = errors.New("accounting period is not open")
	ErrUnbalancedJournal        = errors.New("journal entry is unbalanced")
	ErrInvalidInvoiceDocument   = errors.New("invalid invoice document")
	ErrInvalidPaymentReceipt    = errors.New("invalid payment or receipt document")
	ErrLaborHandoffNotFound     = errors.New("labor accounting handoff not found")
	ErrInvalidLaborHandoff      = errors.New("invalid labor accounting handoff")
	ErrInventoryHandoffNotFound = errors.New("inventory accounting handoff not found")
	ErrInvalidInventoryHandoff  = errors.New("invalid inventory accounting handoff")
)

const (
	AccountClassAsset     = "asset"
	AccountClassLiability = "liability"
	AccountClassEquity    = "equity"
	AccountClassRevenue   = "revenue"
	AccountClassExpense   = "expense"

	StatusActive   = "active"
	StatusInactive = "inactive"

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

	InvoiceRoleSales    = "sales"
	InvoiceRolePurchase = "purchase"

	PaymentReceiptDirectionPayment = "payment"
	PaymentReceiptDirectionReceipt = "receipt"
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

type AccountingPeriod struct {
	ID              string
	OrgID           string
	PeriodCode      string
	StartOn         time.Time
	EndOn           time.Time
	Status          string
	ClosedByUserID  sql.NullString
	ClosedAt        sql.NullTime
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
	EffectiveOn        time.Time
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

type InvoiceDocument struct {
	DocumentID       string
	OrgID            string
	InvoiceRole      sql.NullString
	BilledPartyID    sql.NullString
	BillingContactID sql.NullString
	CurrencyCode     sql.NullString
	ReferenceValue   string
	Summary          string
	CreatedByUserID  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PaymentReceiptDocument struct {
	DocumentID            string
	OrgID                 string
	Direction             sql.NullString
	CounterpartyID        sql.NullString
	CounterpartyContactID sql.NullString
	CurrencyCode          sql.NullString
	ReferenceValue        string
	Summary               string
	CreatedByUserID       string
	CreatedAt             time.Time
	UpdatedAt             time.Time
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

type CreateAccountingPeriodInput struct {
	PeriodCode string
	StartOn    time.Time
	EndOn      time.Time
	Actor      identityaccess.Actor
}

type CloseAccountingPeriodInput struct {
	PeriodID string
	Actor    identityaccess.Actor
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

type UpdateLedgerAccountStatusInput struct {
	AccountID string
	Status    string
	Actor     identityaccess.Actor
}

type UpdateTaxCodeStatusInput struct {
	TaxCodeID string
	Status    string
	Actor     identityaccess.Actor
}

type ListLedgerAccountsInput struct {
	Actor identityaccess.Actor
}

type ListTaxCodesInput struct {
	Actor identityaccess.Actor
}

type ListAccountingPeriodsInput struct {
	Actor identityaccess.Actor
}

type CreateInvoiceInput struct {
	Title            string
	InvoiceRole      string
	BilledPartyID    string
	BillingContactID string
	CurrencyCode     string
	ReferenceValue   string
	Summary          string
	Actor            identityaccess.Actor
}

type CreatePaymentReceiptInput struct {
	Title                 string
	Direction             string
	CounterpartyID        string
	CounterpartyContactID string
	CurrencyCode          string
	ReferenceValue        string
	Summary               string
	Actor                 identityaccess.Actor
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
	EffectiveOn  time.Time
	Lines        []PostingLineInput
	Actor        identityaccess.Actor
}

type ReverseDocumentInput struct {
	DocumentID  string
	Reason      string
	EffectiveOn time.Time
	Actor       identityaccess.Actor
}

type ListJournalEntriesInput struct {
	StartOn time.Time
	EndOn   time.Time
	Limit   int
	Actor   identityaccess.Actor
}

type JournalEntryReview struct {
	Entry            JournalEntry
	DocumentTypeCode sql.NullString
	DocumentNumber   sql.NullString
	DocumentStatus   sql.NullString
	LineCount        int
	TotalDebitMinor  int64
	TotalCreditMinor int64
	HasReversal      bool
}

type ListControlAccountBalancesInput struct {
	AsOf  time.Time
	Actor identityaccess.Actor
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

type PostWorkOrderLaborInput struct {
	DocumentID       string
	WorkOrderID      string
	ExpenseAccountID string
	OffsetAccountID  string
	Summary          string
	EffectiveOn      time.Time
	Actor            identityaccess.Actor
}

type WorkOrderLaborPostingResult struct {
	Entry           JournalEntry
	Lines           []JournalLine
	Document        documents.Document
	LaborEntryCount int
	TotalCostMinor  int64
	CurrencyCode    string
}

type PostWorkOrderInventoryInput struct {
	DocumentID       string
	WorkOrderID      string
	ExpenseAccountID string
	OffsetAccountID  string
	Summary          string
	EffectiveOn      time.Time
	Actor            identityaccess.Actor
}

type WorkOrderInventoryPostingResult struct {
	Entry              JournalEntry
	Lines              []JournalLine
	Document           documents.Document
	InventoryLineCount int
	TotalCostMinor     int64
	CurrencyCode       string
}

type laborAccountingHandoff struct {
	ID               string
	LaborEntryID     string
	WorkOrderID      string
	TaskID           sql.NullString
	CostMinor        int64
	CostCurrencyCode string
}

type inventoryAccountingHandoff struct {
	ID               string
	DocumentLineID   string
	WorkOrderID      string
	CostMinor        int64
	CostCurrencyCode string
}

type Service struct {
	db        *sql.DB
	documents *documents.Service
}

func NewService(db *sql.DB, documentService *documents.Service) *Service {
	return &Service{db: db, documents: documentService}
}

func (s *Service) CreateInvoice(ctx context.Context, input CreateInvoiceInput) (documents.Document, InvoiceDocument, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return documents.Document{}, InvoiceDocument{}, fmt.Errorf("begin create invoice: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return documents.Document{}, InvoiceDocument{}, err
	}

	document, payload, err := s.createInvoiceTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return documents.Document{}, InvoiceDocument{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "documents.document_created",
		EntityType:  "documents.document",
		EntityID:    document.ID,
		Payload: map[string]any{
			"type_code": document.TypeCode,
			"status":    document.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return documents.Document{}, InvoiceDocument{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.invoice_document_created",
		EntityType:  "accounting.invoice_document",
		EntityID:    payload.DocumentID,
		Payload: map[string]any{
			"invoice_role":    nullableString(payload.InvoiceRole),
			"billed_party_id": nullableString(payload.BilledPartyID),
			"currency_code":   nullableString(payload.CurrencyCode),
		},
	}); err != nil {
		_ = tx.Rollback()
		return documents.Document{}, InvoiceDocument{}, err
	}

	if err := tx.Commit(); err != nil {
		return documents.Document{}, InvoiceDocument{}, fmt.Errorf("commit create invoice: %w", err)
	}

	return document, payload, nil
}

func (s *Service) CreatePaymentReceipt(ctx context.Context, input CreatePaymentReceiptInput) (documents.Document, PaymentReceiptDocument, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return documents.Document{}, PaymentReceiptDocument{}, fmt.Errorf("begin create payment receipt: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return documents.Document{}, PaymentReceiptDocument{}, err
	}

	document, payload, err := s.createPaymentReceiptTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return documents.Document{}, PaymentReceiptDocument{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "documents.document_created",
		EntityType:  "documents.document",
		EntityID:    document.ID,
		Payload: map[string]any{
			"type_code": document.TypeCode,
			"status":    document.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return documents.Document{}, PaymentReceiptDocument{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.payment_receipt_document_created",
		EntityType:  "accounting.payment_receipt_document",
		EntityID:    payload.DocumentID,
		Payload: map[string]any{
			"direction":       nullableString(payload.Direction),
			"counterparty_id": nullableString(payload.CounterpartyID),
			"currency_code":   nullableString(payload.CurrencyCode),
		},
	}); err != nil {
		_ = tx.Rollback()
		return documents.Document{}, PaymentReceiptDocument{}, err
	}

	if err := tx.Commit(); err != nil {
		return documents.Document{}, PaymentReceiptDocument{}, fmt.Errorf("commit create payment receipt: %w", err)
	}

	return document, payload, nil
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

func (s *Service) CreateAccountingPeriod(ctx context.Context, input CreateAccountingPeriodInput) (AccountingPeriod, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AccountingPeriod{}, fmt.Errorf("begin create accounting period: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin); err != nil {
		_ = tx.Rollback()
		return AccountingPeriod{}, err
	}

	period, err := createAccountingPeriodTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return AccountingPeriod{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.period_created",
		EntityType:  "accounting.period",
		EntityID:    period.ID,
		Payload: map[string]any{
			"period_code": period.PeriodCode,
			"start_on":    period.StartOn.Format(time.DateOnly),
			"end_on":      period.EndOn.Format(time.DateOnly),
			"status":      period.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return AccountingPeriod{}, err
	}

	if err := tx.Commit(); err != nil {
		return AccountingPeriod{}, fmt.Errorf("commit create accounting period: %w", err)
	}

	return period, nil
}

func (s *Service) CloseAccountingPeriod(ctx context.Context, input CloseAccountingPeriodInput) (AccountingPeriod, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AccountingPeriod{}, fmt.Errorf("begin close accounting period: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin); err != nil {
		_ = tx.Rollback()
		return AccountingPeriod{}, err
	}

	period, err := closeAccountingPeriodTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return AccountingPeriod{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.period_closed",
		EntityType:  "accounting.period",
		EntityID:    period.ID,
		Payload: map[string]any{
			"period_code": period.PeriodCode,
			"start_on":    period.StartOn.Format(time.DateOnly),
			"end_on":      period.EndOn.Format(time.DateOnly),
			"status":      period.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return AccountingPeriod{}, err
	}

	if err := tx.Commit(); err != nil {
		return AccountingPeriod{}, fmt.Errorf("commit close accounting period: %w", err)
	}

	return period, nil
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

func (s *Service) UpdateLedgerAccountStatus(ctx context.Context, input UpdateLedgerAccountStatusInput) (LedgerAccount, error) {
	status := strings.TrimSpace(input.Status)
	if strings.TrimSpace(input.AccountID) == "" || !isValidSetupStatus(status) {
		return LedgerAccount{}, ErrInvalidAccount
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return LedgerAccount{}, fmt.Errorf("begin update ledger account status: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin); err != nil {
		_ = tx.Rollback()
		return LedgerAccount{}, err
	}

	account, err := scanLedgerAccount(tx.QueryRowContext(ctx, `
UPDATE accounting.ledger_accounts
SET status = $3,
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
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
	updated_at;`,
		input.Actor.OrgID,
		input.AccountID,
		status,
	))
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrLedgerAccountNotFound) {
			return LedgerAccount{}, ErrLedgerAccountNotFound
		}
		return LedgerAccount{}, fmt.Errorf("update ledger account status: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.ledger_account_status_updated",
		EntityType:  "accounting.ledger_account",
		EntityID:    account.ID,
		Payload: map[string]any{
			"code":   account.Code,
			"status": account.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return LedgerAccount{}, err
	}

	if err := tx.Commit(); err != nil {
		return LedgerAccount{}, fmt.Errorf("commit update ledger account status: %w", err)
	}

	return account, nil
}

func (s *Service) UpdateTaxCodeStatus(ctx context.Context, input UpdateTaxCodeStatusInput) (TaxCode, error) {
	status := strings.TrimSpace(input.Status)
	if strings.TrimSpace(input.TaxCodeID) == "" || !isValidSetupStatus(status) {
		return TaxCode{}, ErrInvalidTaxCode
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TaxCode{}, fmt.Errorf("begin update tax code status: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin); err != nil {
		_ = tx.Rollback()
		return TaxCode{}, err
	}

	taxCode, err := scanTaxCode(tx.QueryRowContext(ctx, `
UPDATE accounting.tax_codes
SET status = $3,
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
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
	updated_at;`,
		input.Actor.OrgID,
		input.TaxCodeID,
		status,
	))
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrTaxCodeNotFound) {
			return TaxCode{}, ErrTaxCodeNotFound
		}
		return TaxCode{}, fmt.Errorf("update tax code status: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.tax_code_status_updated",
		EntityType:  "accounting.tax_code",
		EntityID:    taxCode.ID,
		Payload: map[string]any{
			"code":   taxCode.Code,
			"status": taxCode.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return TaxCode{}, err
	}

	if err := tx.Commit(); err != nil {
		return TaxCode{}, fmt.Errorf("commit update tax code status: %w", err)
	}

	return taxCode, nil
}

func (s *Service) ListLedgerAccounts(ctx context.Context, input ListLedgerAccountsInput) ([]LedgerAccount, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list ledger accounts: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin); err != nil {
		return nil, err
	}

	items, err := listLedgerAccountsTx(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list ledger accounts: %w", err)
	}

	return items, nil
}

func (s *Service) ListTaxCodes(ctx context.Context, input ListTaxCodesInput) ([]TaxCode, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list tax codes: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin); err != nil {
		return nil, err
	}

	items, err := listTaxCodesTx(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list tax codes: %w", err)
	}

	return items, nil
}

func (s *Service) ListAccountingPeriods(ctx context.Context, input ListAccountingPeriodsInput) ([]AccountingPeriod, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list accounting periods: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin); err != nil {
		return nil, err
	}

	items, err := listAccountingPeriodsTx(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list accounting periods: %w", err)
	}

	return items, nil
}

func (s *Service) ListJournalEntries(ctx context.Context, input ListJournalEntriesInput) ([]JournalEntryReview, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list journal entries: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleApprover); err != nil {
		return nil, err
	}

	reviews, err := listJournalEntriesTx(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list journal entries: %w", err)
	}

	return reviews, nil
}

func (s *Service) ListControlAccountBalances(ctx context.Context, input ListControlAccountBalancesInput) ([]ControlAccountBalance, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list control account balances: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleApprover); err != nil {
		return nil, err
	}

	balances, err := listControlAccountBalancesTx(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list control account balances: %w", err)
	}

	return balances, nil
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

	reversal, lines, err := createReversalTx(ctx, tx, original, originalLines, strings.TrimSpace(input.Reason), normalizeEffectiveOn(input.EffectiveOn), input.Actor)
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

func (s *Service) PostWorkOrderLabor(ctx context.Context, input PostWorkOrderLaborInput) (WorkOrderLaborPostingResult, error) {
	if strings.TrimSpace(input.DocumentID) == "" ||
		strings.TrimSpace(input.WorkOrderID) == "" ||
		strings.TrimSpace(input.ExpenseAccountID) == "" ||
		strings.TrimSpace(input.OffsetAccountID) == "" ||
		strings.TrimSpace(input.Summary) == "" {
		return WorkOrderLaborPostingResult{}, ErrInvalidLaborHandoff
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WorkOrderLaborPostingResult{}, fmt.Errorf("begin post work order labor: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleApprover); err != nil {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, err
	}

	entry, lines, found, err := getPostingEntryByDocumentForUpdate(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil && !errors.Is(err, ErrJournalEntryNotFound) {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, err
	}
	if found {
		count, total, currency, summaryFound, err := getPostedLaborSummaryTx(ctx, tx, input.Actor.OrgID, input.WorkOrderID, entry.ID)
		if err != nil {
			_ = tx.Rollback()
			return WorkOrderLaborPostingResult{}, err
		}
		if !summaryFound {
			_ = tx.Rollback()
			return WorkOrderLaborPostingResult{}, ErrPostingAlreadyExists
		}
		doc, err := s.loadDocument(ctx, input.Actor.OrgID, input.DocumentID)
		if err != nil {
			_ = tx.Rollback()
			return WorkOrderLaborPostingResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return WorkOrderLaborPostingResult{}, fmt.Errorf("commit idempotent post work order labor: %w", err)
		}
		return WorkOrderLaborPostingResult{
			Entry:           entry,
			Lines:           lines,
			Document:        doc,
			LaborEntryCount: count,
			TotalCostMinor:  total,
			CurrencyCode:    currency,
		}, nil
	}

	handoffs, totalCostMinor, currencyCode, err := listPendingLaborHandoffsForWorkOrderTx(ctx, tx, input.Actor.OrgID, input.WorkOrderID)
	if err != nil {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, err
	}
	if len(handoffs) == 0 {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, ErrLaborHandoffNotFound
	}
	if totalCostMinor <= 0 || !isValidCurrencyCode(currencyCode) {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, ErrInvalidLaborHandoff
	}

	postInput := PostDocumentInput{
		DocumentID:   input.DocumentID,
		Summary:      strings.TrimSpace(input.Summary),
		CurrencyCode: currencyCode,
		TaxScopeCode: TaxScopeNone,
		EffectiveOn:  input.EffectiveOn,
		Lines: []PostingLineInput{
			{
				AccountID:   strings.TrimSpace(input.ExpenseAccountID),
				Description: "Work-order labor cost",
				DebitMinor:  totalCostMinor,
			},
			{
				AccountID:   strings.TrimSpace(input.OffsetAccountID),
				Description: "Labor cost clearing",
				CreditMinor: totalCostMinor,
			},
		},
		Actor: input.Actor,
	}

	fingerprint, err := postingFingerprint(postInput)
	if err != nil {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, err
	}

	entry, lines, existing, err := s.postDocumentTx(ctx, tx, postInput, fingerprint)
	if err != nil {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, err
	}
	if existing {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, ErrPostingAlreadyExists
	}

	doc, err := s.documents.ApplyPostingOutcome(ctx, tx, documents.PostingOutcomeInput{
		DocumentID: input.DocumentID,
		Action:     "posted",
		Actor:      input.Actor,
	})
	if err != nil {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, err
	}

	if err := markLaborHandoffsPostedTx(ctx, tx, input.Actor.OrgID, entry.ID, handoffs); err != nil {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.work_order_labor_posted",
		EntityType:  "accounting.journal_entry",
		EntityID:    entry.ID,
		Payload: map[string]any{
			"document_id":        input.DocumentID,
			"work_order_id":      input.WorkOrderID,
			"labor_entry_count":  len(handoffs),
			"total_cost_minor":   totalCostMinor,
			"cost_currency_code": currencyCode,
		},
	}); err != nil {
		_ = tx.Rollback()
		return WorkOrderLaborPostingResult{}, err
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
		return WorkOrderLaborPostingResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return WorkOrderLaborPostingResult{}, fmt.Errorf("commit post work order labor: %w", err)
	}

	return WorkOrderLaborPostingResult{
		Entry:           entry,
		Lines:           lines,
		Document:        doc,
		LaborEntryCount: len(handoffs),
		TotalCostMinor:  totalCostMinor,
		CurrencyCode:    currencyCode,
	}, nil
}

func (s *Service) PostWorkOrderInventory(ctx context.Context, input PostWorkOrderInventoryInput) (WorkOrderInventoryPostingResult, error) {
	if strings.TrimSpace(input.DocumentID) == "" ||
		strings.TrimSpace(input.WorkOrderID) == "" ||
		strings.TrimSpace(input.ExpenseAccountID) == "" ||
		strings.TrimSpace(input.OffsetAccountID) == "" ||
		strings.TrimSpace(input.Summary) == "" {
		return WorkOrderInventoryPostingResult{}, ErrInvalidInventoryHandoff
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WorkOrderInventoryPostingResult{}, fmt.Errorf("begin post work order inventory: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleApprover); err != nil {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, err
	}

	entry, lines, found, err := getPostingEntryByDocumentForUpdate(ctx, tx, input.Actor.OrgID, input.DocumentID)
	if err != nil && !errors.Is(err, ErrJournalEntryNotFound) {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, err
	}
	if found {
		count, total, currency, summaryFound, err := getPostedInventorySummaryTx(ctx, tx, input.Actor.OrgID, input.WorkOrderID, entry.ID)
		if err != nil {
			_ = tx.Rollback()
			return WorkOrderInventoryPostingResult{}, err
		}
		if !summaryFound {
			_ = tx.Rollback()
			return WorkOrderInventoryPostingResult{}, ErrPostingAlreadyExists
		}
		doc, err := s.loadDocument(ctx, input.Actor.OrgID, input.DocumentID)
		if err != nil {
			_ = tx.Rollback()
			return WorkOrderInventoryPostingResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return WorkOrderInventoryPostingResult{}, fmt.Errorf("commit idempotent post work order inventory: %w", err)
		}
		return WorkOrderInventoryPostingResult{
			Entry:              entry,
			Lines:              lines,
			Document:           doc,
			InventoryLineCount: count,
			TotalCostMinor:     total,
			CurrencyCode:       currency,
		}, nil
	}

	handoffs, totalCostMinor, currencyCode, err := listPendingInventoryHandoffsForWorkOrderTx(ctx, tx, input.Actor.OrgID, input.WorkOrderID)
	if err != nil {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, err
	}
	if len(handoffs) == 0 {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, ErrInventoryHandoffNotFound
	}
	if totalCostMinor <= 0 || !isValidCurrencyCode(currencyCode) {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, ErrInvalidInventoryHandoff
	}

	postInput := PostDocumentInput{
		DocumentID:   input.DocumentID,
		Summary:      strings.TrimSpace(input.Summary),
		CurrencyCode: currencyCode,
		TaxScopeCode: TaxScopeNone,
		EffectiveOn:  input.EffectiveOn,
		Lines: []PostingLineInput{
			{
				AccountID:   strings.TrimSpace(input.ExpenseAccountID),
				Description: "Work-order material cost",
				DebitMinor:  totalCostMinor,
			},
			{
				AccountID:   strings.TrimSpace(input.OffsetAccountID),
				Description: "Inventory issue clearing",
				CreditMinor: totalCostMinor,
			},
		},
		Actor: input.Actor,
	}

	fingerprint, err := postingFingerprint(postInput)
	if err != nil {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, err
	}

	entry, lines, existing, err := s.postDocumentTx(ctx, tx, postInput, fingerprint)
	if err != nil {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, err
	}
	if existing {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, ErrPostingAlreadyExists
	}

	doc, err := s.documents.ApplyPostingOutcome(ctx, tx, documents.PostingOutcomeInput{
		DocumentID: input.DocumentID,
		Action:     "posted",
		Actor:      input.Actor,
	})
	if err != nil {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, err
	}

	if err := markInventoryHandoffsPostedTx(ctx, tx, input.Actor.OrgID, entry.ID, handoffs); err != nil {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "accounting.work_order_inventory_posted",
		EntityType:  "accounting.journal_entry",
		EntityID:    entry.ID,
		Payload: map[string]any{
			"document_id":          input.DocumentID,
			"work_order_id":        input.WorkOrderID,
			"inventory_line_count": len(handoffs),
			"total_cost_minor":     totalCostMinor,
			"cost_currency_code":   currencyCode,
		},
	}); err != nil {
		_ = tx.Rollback()
		return WorkOrderInventoryPostingResult{}, err
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
		return WorkOrderInventoryPostingResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return WorkOrderInventoryPostingResult{}, fmt.Errorf("commit post work order inventory: %w", err)
	}

	return WorkOrderInventoryPostingResult{
		Entry:              entry,
		Lines:              lines,
		Document:           doc,
		InventoryLineCount: len(handoffs),
		TotalCostMinor:     totalCostMinor,
		CurrencyCode:       currencyCode,
	}, nil
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
	if err := validateAccountingDocumentOwnershipTx(ctx, tx, input.Actor.OrgID, input.DocumentID); err != nil {
		return JournalEntry{}, nil, false, err
	}

	effectiveOn := normalizeEffectiveOn(input.EffectiveOn)
	if err := validateAccountingPeriodOpenTx(ctx, tx, input.Actor.OrgID, effectiveOn); err != nil {
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
		EffectiveOn:        effectiveOn,
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

func (s *Service) createInvoiceTx(ctx context.Context, tx *sql.Tx, input CreateInvoiceInput) (documents.Document, InvoiceDocument, error) {
	role := normalizeInvoiceRole(input.InvoiceRole)
	currencyCode := strings.ToUpper(strings.TrimSpace(input.CurrencyCode))
	if strings.TrimSpace(input.Title) == "" || !isValidInvoiceRole(role) || !isValidCurrencyCode(currencyCode) {
		return documents.Document{}, InvoiceDocument{}, ErrInvalidInvoiceDocument
	}

	if err := validatePartyContactPairTx(ctx, tx, input.Actor.OrgID, input.BilledPartyID, input.BillingContactID, ErrInvalidInvoiceDocument); err != nil {
		return documents.Document{}, InvoiceDocument{}, err
	}

	document, err := s.documents.CreateDraftTx(ctx, tx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    strings.TrimSpace(input.Title),
		Actor:    input.Actor,
	})
	if err != nil {
		return documents.Document{}, InvoiceDocument{}, fmt.Errorf("create invoice document: %w", err)
	}

	payload, err := scanInvoiceDocument(tx.QueryRowContext(ctx, `
INSERT INTO accounting.invoice_documents (
	document_id,
	org_id,
	invoice_role,
	billed_party_id,
	billing_contact_id,
	currency_code,
	reference_value,
	summary,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING
	document_id,
	org_id,
	invoice_role,
	billed_party_id,
	billing_contact_id,
	currency_code,
	reference_value,
	summary,
	created_by_user_id,
	created_at,
	updated_at;`,
		document.ID,
		input.Actor.OrgID,
		role,
		nullIfEmpty(strings.TrimSpace(input.BilledPartyID)),
		nullIfEmpty(strings.TrimSpace(input.BillingContactID)),
		currencyCode,
		strings.TrimSpace(input.ReferenceValue),
		strings.TrimSpace(input.Summary),
		input.Actor.UserID,
	))
	if err != nil {
		return documents.Document{}, InvoiceDocument{}, fmt.Errorf("insert invoice payload: %w", err)
	}

	return document, payload, nil
}

func (s *Service) createPaymentReceiptTx(ctx context.Context, tx *sql.Tx, input CreatePaymentReceiptInput) (documents.Document, PaymentReceiptDocument, error) {
	direction := normalizePaymentReceiptDirection(input.Direction)
	currencyCode := strings.ToUpper(strings.TrimSpace(input.CurrencyCode))
	if strings.TrimSpace(input.Title) == "" || !isValidPaymentReceiptDirection(direction) || !isValidCurrencyCode(currencyCode) {
		return documents.Document{}, PaymentReceiptDocument{}, ErrInvalidPaymentReceipt
	}

	if err := validatePartyContactPairTx(ctx, tx, input.Actor.OrgID, input.CounterpartyID, input.CounterpartyContactID, ErrInvalidPaymentReceipt); err != nil {
		return documents.Document{}, PaymentReceiptDocument{}, err
	}

	document, err := s.documents.CreateDraftTx(ctx, tx, documents.CreateDraftInput{
		TypeCode: "payment_receipt",
		Title:    strings.TrimSpace(input.Title),
		Actor:    input.Actor,
	})
	if err != nil {
		return documents.Document{}, PaymentReceiptDocument{}, fmt.Errorf("create payment receipt document: %w", err)
	}

	payload, err := scanPaymentReceiptDocument(tx.QueryRowContext(ctx, `
INSERT INTO accounting.payment_receipt_documents (
	document_id,
	org_id,
	direction,
	counterparty_id,
	counterparty_contact_id,
	currency_code,
	reference_value,
	summary,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING
	document_id,
	org_id,
	direction,
	counterparty_id,
	counterparty_contact_id,
	currency_code,
	reference_value,
	summary,
	created_by_user_id,
	created_at,
	updated_at;`,
		document.ID,
		input.Actor.OrgID,
		direction,
		nullIfEmpty(strings.TrimSpace(input.CounterpartyID)),
		nullIfEmpty(strings.TrimSpace(input.CounterpartyContactID)),
		currencyCode,
		strings.TrimSpace(input.ReferenceValue),
		strings.TrimSpace(input.Summary),
		input.Actor.UserID,
	))
	if err != nil {
		return documents.Document{}, PaymentReceiptDocument{}, fmt.Errorf("insert payment receipt payload: %w", err)
	}

	return document, payload, nil
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

func createAccountingPeriodTx(ctx context.Context, tx *sql.Tx, input CreateAccountingPeriodInput) (AccountingPeriod, error) {
	periodCode := strings.TrimSpace(input.PeriodCode)
	startOn, endOn, err := normalizePeriodRange(input.StartOn, input.EndOn)
	if err != nil || periodCode == "" {
		return AccountingPeriod{}, ErrInvalidAccountingPeriod
	}
	if err := ensureNoPeriodOverlapTx(ctx, tx, input.Actor.OrgID, "", startOn, endOn); err != nil {
		return AccountingPeriod{}, err
	}

	const statement = `
INSERT INTO accounting.periods (
	org_id,
	period_code,
	start_on,
	end_on,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING
	id,
	org_id,
	period_code,
	start_on,
	end_on,
	status,
	closed_by_user_id,
	closed_at,
	created_by_user_id,
	created_at,
	updated_at;`

	return scanAccountingPeriod(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		periodCode,
		startOn,
		endOn,
		input.Actor.UserID,
	))
}

func closeAccountingPeriodTx(ctx context.Context, tx *sql.Tx, input CloseAccountingPeriodInput) (AccountingPeriod, error) {
	const query = `
SELECT
	id,
	org_id,
	period_code,
	start_on,
	end_on,
	status,
	closed_by_user_id,
	closed_at,
	created_by_user_id,
	created_at,
	updated_at
FROM accounting.periods
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	period, err := scanAccountingPeriod(tx.QueryRowContext(ctx, query, input.Actor.OrgID, strings.TrimSpace(input.PeriodID)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccountingPeriod{}, ErrAccountingPeriodNotFound
		}
		return AccountingPeriod{}, fmt.Errorf("load accounting period: %w", err)
	}
	if period.Status != "open" {
		return AccountingPeriod{}, ErrAccountingPeriodNotOpen
	}

	const statement = `
UPDATE accounting.periods
SET status = 'closed',
	closed_by_user_id = $3,
	closed_at = NOW(),
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	period_code,
	start_on,
	end_on,
	status,
	closed_by_user_id,
	closed_at,
	created_by_user_id,
	created_at,
	updated_at;`

	return scanAccountingPeriod(tx.QueryRowContext(ctx, statement, input.Actor.OrgID, input.PeriodID, input.Actor.UserID))
}

func listLedgerAccountsTx(ctx context.Context, tx *sql.Tx, input ListLedgerAccountsInput) ([]LedgerAccount, error) {
	const query = `
SELECT
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
	updated_at
FROM accounting.ledger_accounts
WHERE org_id = $1
ORDER BY code ASC, created_at ASC;`

	rows, err := tx.QueryContext(ctx, query, input.Actor.OrgID)
	if err != nil {
		return nil, fmt.Errorf("query ledger accounts: %w", err)
	}
	defer rows.Close()

	var items []LedgerAccount
	for rows.Next() {
		item, scanErr := scanLedgerAccount(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan ledger account: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ledger accounts: %w", err)
	}
	return items, nil
}

func listTaxCodesTx(ctx context.Context, tx *sql.Tx, input ListTaxCodesInput) ([]TaxCode, error) {
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
ORDER BY tax_type ASC, code ASC, created_at ASC;`

	rows, err := tx.QueryContext(ctx, query, input.Actor.OrgID)
	if err != nil {
		return nil, fmt.Errorf("query tax codes: %w", err)
	}
	defer rows.Close()

	var items []TaxCode
	for rows.Next() {
		item, scanErr := scanTaxCode(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan tax code: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tax codes: %w", err)
	}
	return items, nil
}

func listAccountingPeriodsTx(ctx context.Context, tx *sql.Tx, input ListAccountingPeriodsInput) ([]AccountingPeriod, error) {
	const query = `
SELECT
	id,
	org_id,
	period_code,
	start_on,
	end_on,
	status,
	closed_by_user_id,
	closed_at,
	created_by_user_id,
	created_at,
	updated_at
FROM accounting.periods
WHERE org_id = $1
ORDER BY start_on DESC, period_code DESC, created_at DESC;`

	rows, err := tx.QueryContext(ctx, query, input.Actor.OrgID)
	if err != nil {
		return nil, fmt.Errorf("query accounting periods: %w", err)
	}
	defer rows.Close()

	var items []AccountingPeriod
	for rows.Next() {
		item, scanErr := scanAccountingPeriod(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan accounting period: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounting periods: %w", err)
	}
	return items, nil
}

func listJournalEntriesTx(ctx context.Context, tx *sql.Tx, input ListJournalEntriesInput) ([]JournalEntryReview, error) {
	limit := input.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	startOn, startSet := normalizeOptionalDate(input.StartOn)
	endOn, endSet := normalizeOptionalDate(input.EndOn)

	const query = `
SELECT
	e.id,
	e.org_id,
	e.entry_number,
	e.entry_kind,
	e.source_document_id,
	e.reversal_of_entry_id,
	e.posting_fingerprint,
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
GROUP BY
	e.id,
	e.org_id,
	e.entry_number,
	e.entry_kind,
	e.source_document_id,
	e.reversal_of_entry_id,
	e.posting_fingerprint,
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
LIMIT $4;`

	rows, err := tx.QueryContext(
		ctx,
		query,
		input.Actor.OrgID,
		nullableDate(startOn, startSet),
		nullableDate(endOn, endSet),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list journal entries: %w", err)
	}
	defer rows.Close()

	var reviews []JournalEntryReview
	for rows.Next() {
		var review JournalEntryReview
		if err := rows.Scan(
			&review.Entry.ID,
			&review.Entry.OrgID,
			&review.Entry.EntryNumber,
			&review.Entry.EntryKind,
			&review.Entry.SourceDocumentID,
			&review.Entry.ReversalOfEntryID,
			&review.Entry.PostingFingerprint,
			&review.Entry.CurrencyCode,
			&review.Entry.TaxScopeCode,
			&review.Entry.Summary,
			&review.Entry.ReversalReason,
			&review.Entry.PostedByUserID,
			&review.Entry.EffectiveOn,
			&review.Entry.PostedAt,
			&review.Entry.CreatedAt,
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
		return nil, fmt.Errorf("iterate journal entries: %w", err)
	}
	return reviews, nil
}

func listControlAccountBalancesTx(ctx context.Context, tx *sql.Tx, input ListControlAccountBalancesInput) ([]ControlAccountBalance, error) {
	asOf, asOfSet := normalizeOptionalDate(input.AsOf)

	const query = `
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
GROUP BY a.id, a.code, a.name, a.account_class, a.control_type
ORDER BY a.code ASC;`

	rows, err := tx.QueryContext(ctx, query, input.Actor.OrgID, nullableDate(asOf, asOfSet))
	if err != nil {
		return nil, fmt.Errorf("list control account balances: %w", err)
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
	return balances, nil
}

func listPendingLaborHandoffsForWorkOrderTx(ctx context.Context, tx *sql.Tx, orgID, workOrderID string) ([]laborAccountingHandoff, int64, string, error) {
	const query = `
SELECT
	h.id,
	h.labor_entry_id,
	h.work_order_id,
	h.task_id,
	le.cost_minor,
	le.cost_currency_code
FROM workforce.labor_accounting_handoffs h
JOIN workforce.labor_entries le
	ON le.id = h.labor_entry_id
   AND le.org_id = h.org_id
WHERE h.org_id = $1
  AND h.work_order_id = $2
  AND h.handoff_status = 'pending'
ORDER BY le.started_at ASC, le.id ASC
FOR UPDATE OF h, le;`

	rows, err := tx.QueryContext(ctx, query, orgID, workOrderID)
	if err != nil {
		return nil, 0, "", fmt.Errorf("list pending labor handoffs: %w", err)
	}
	defer rows.Close()

	var (
		handoffs     []laborAccountingHandoff
		totalCost    int64
		currencyCode string
	)
	for rows.Next() {
		var handoff laborAccountingHandoff
		if err := rows.Scan(
			&handoff.ID,
			&handoff.LaborEntryID,
			&handoff.WorkOrderID,
			&handoff.TaskID,
			&handoff.CostMinor,
			&handoff.CostCurrencyCode,
		); err != nil {
			return nil, 0, "", fmt.Errorf("scan pending labor handoff: %w", err)
		}
		if handoff.CostMinor < 0 || !isValidCurrencyCode(handoff.CostCurrencyCode) {
			return nil, 0, "", ErrInvalidLaborHandoff
		}
		if currencyCode == "" {
			currencyCode = handoff.CostCurrencyCode
		} else if currencyCode != handoff.CostCurrencyCode {
			return nil, 0, "", ErrInvalidLaborHandoff
		}
		totalCost += handoff.CostMinor
		handoffs = append(handoffs, handoff)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, "", fmt.Errorf("iterate pending labor handoffs: %w", err)
	}

	return handoffs, totalCost, currencyCode, nil
}

func getPostedLaborSummaryTx(ctx context.Context, tx *sql.Tx, orgID, workOrderID, journalEntryID string) (int, int64, string, bool, error) {
	const query = `
SELECT
	COUNT(*),
	COALESCE(SUM(le.cost_minor), 0),
	MIN(le.cost_currency_code),
	MAX(le.cost_currency_code)
FROM workforce.labor_accounting_handoffs h
JOIN workforce.labor_entries le
	ON le.id = h.labor_entry_id
   AND le.org_id = h.org_id
WHERE h.org_id = $1
  AND h.work_order_id = $2
  AND h.journal_entry_id = $3
  AND h.handoff_status = 'posted';`

	var (
		count       int
		total       int64
		minCurrency sql.NullString
		maxCurrency sql.NullString
	)
	if err := tx.QueryRowContext(ctx, query, orgID, workOrderID, journalEntryID).Scan(&count, &total, &minCurrency, &maxCurrency); err != nil {
		return 0, 0, "", false, fmt.Errorf("load posted labor summary: %w", err)
	}
	if count == 0 {
		return 0, 0, "", false, nil
	}
	if !minCurrency.Valid || !maxCurrency.Valid || minCurrency.String != maxCurrency.String {
		return 0, 0, "", false, ErrInvalidLaborHandoff
	}
	return count, total, minCurrency.String, true, nil
}

func listPendingInventoryHandoffsForWorkOrderTx(ctx context.Context, tx *sql.Tx, orgID, workOrderID string) ([]inventoryAccountingHandoff, int64, string, error) {
	const query = `
SELECT
	h.id,
	h.document_line_id,
	mu.work_order_id,
	h.cost_minor,
	h.cost_currency_code
FROM inventory_ops.accounting_handoffs h
JOIN work_orders.material_usages mu
	ON mu.inventory_document_line_id = h.document_line_id
   AND mu.org_id = h.org_id
WHERE h.org_id = $1
  AND mu.work_order_id = $2
  AND h.handoff_status = 'pending'
ORDER BY mu.linked_at ASC, mu.id ASC
FOR UPDATE OF h, mu;`

	rows, err := tx.QueryContext(ctx, query, orgID, workOrderID)
	if err != nil {
		return nil, 0, "", fmt.Errorf("list pending inventory handoffs: %w", err)
	}
	defer rows.Close()

	var (
		handoffs     []inventoryAccountingHandoff
		totalCost    int64
		currencyCode string
	)
	for rows.Next() {
		var (
			handoff   inventoryAccountingHandoff
			costMinor sql.NullInt64
			currency  sql.NullString
		)
		if err := rows.Scan(
			&handoff.ID,
			&handoff.DocumentLineID,
			&handoff.WorkOrderID,
			&costMinor,
			&currency,
		); err != nil {
			return nil, 0, "", fmt.Errorf("scan pending inventory handoff: %w", err)
		}
		if !costMinor.Valid || costMinor.Int64 <= 0 || !currency.Valid || !isValidCurrencyCode(currency.String) {
			return nil, 0, "", ErrInvalidInventoryHandoff
		}
		handoff.CostMinor = costMinor.Int64
		handoff.CostCurrencyCode = strings.ToUpper(currency.String)
		if currencyCode == "" {
			currencyCode = handoff.CostCurrencyCode
		} else if currencyCode != handoff.CostCurrencyCode {
			return nil, 0, "", ErrInvalidInventoryHandoff
		}
		totalCost += handoff.CostMinor
		handoffs = append(handoffs, handoff)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, "", fmt.Errorf("iterate pending inventory handoffs: %w", err)
	}

	return handoffs, totalCost, currencyCode, nil
}

func getPostedInventorySummaryTx(ctx context.Context, tx *sql.Tx, orgID, workOrderID, journalEntryID string) (int, int64, string, bool, error) {
	const query = `
SELECT
	COUNT(*),
	COALESCE(SUM(h.cost_minor), 0),
	MIN(h.cost_currency_code),
	MAX(h.cost_currency_code)
FROM inventory_ops.accounting_handoffs h
JOIN work_orders.material_usages mu
	ON mu.inventory_document_line_id = h.document_line_id
   AND mu.org_id = h.org_id
WHERE h.org_id = $1
  AND mu.work_order_id = $2
  AND h.journal_entry_id = $3
  AND h.handoff_status = 'posted';`

	var (
		count       int
		total       int64
		minCurrency sql.NullString
		maxCurrency sql.NullString
	)
	if err := tx.QueryRowContext(ctx, query, orgID, workOrderID, journalEntryID).Scan(&count, &total, &minCurrency, &maxCurrency); err != nil {
		return 0, 0, "", false, fmt.Errorf("load posted inventory summary: %w", err)
	}
	if count == 0 {
		return 0, 0, "", false, nil
	}
	if !minCurrency.Valid || !maxCurrency.Valid || minCurrency.String != maxCurrency.String {
		return 0, 0, "", false, ErrInvalidInventoryHandoff
	}
	return count, total, strings.ToUpper(minCurrency.String), true, nil
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

func markLaborHandoffsPostedTx(ctx context.Context, tx *sql.Tx, orgID, journalEntryID string, handoffs []laborAccountingHandoff) error {
	const statement = `
UPDATE workforce.labor_accounting_handoffs
SET journal_entry_id = $3,
	handoff_status = 'posted',
	posted_at = NOW()
WHERE org_id = $1
  AND id = $2
  AND handoff_status = 'pending';`

	for _, handoff := range handoffs {
		result, err := tx.ExecContext(ctx, statement, orgID, handoff.ID, journalEntryID)
		if err != nil {
			return fmt.Errorf("mark labor handoff posted: %w", err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("count marked labor handoff rows: %w", err)
		}
		if rowsAffected != 1 {
			return ErrLaborHandoffNotFound
		}
	}
	return nil
}

func markInventoryHandoffsPostedTx(ctx context.Context, tx *sql.Tx, orgID, journalEntryID string, handoffs []inventoryAccountingHandoff) error {
	const statement = `
UPDATE inventory_ops.accounting_handoffs
SET journal_entry_id = $3,
	handoff_status = 'posted',
	posted_at = NOW()
WHERE org_id = $1
  AND id = $2
  AND handoff_status = 'pending';`

	for _, handoff := range handoffs {
		result, err := tx.ExecContext(ctx, statement, orgID, handoff.ID, journalEntryID)
		if err != nil {
			return fmt.Errorf("mark inventory handoff posted: %w", err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("count marked inventory handoff rows: %w", err)
		}
		if rowsAffected != 1 {
			return ErrInventoryHandoffNotFound
		}
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

func validateAccountingPeriodOpenTx(ctx context.Context, tx *sql.Tx, orgID string, effectiveOn time.Time) error {
	const countQuery = `
SELECT COUNT(*)
FROM accounting.periods
WHERE org_id = $1;`

	var count int
	if err := tx.QueryRowContext(ctx, countQuery, orgID).Scan(&count); err != nil {
		return fmt.Errorf("count accounting periods: %w", err)
	}
	if count == 0 {
		return nil
	}

	const query = `
SELECT status
FROM accounting.periods
WHERE org_id = $1
  AND $2 BETWEEN start_on AND end_on
ORDER BY start_on DESC
LIMIT 1
FOR UPDATE;`

	var status string
	if err := tx.QueryRowContext(ctx, query, orgID, effectiveOn).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAccountingPeriodNotOpen
		}
		return fmt.Errorf("load accounting period: %w", err)
	}
	if status != "open" {
		return ErrAccountingPeriodNotOpen
	}
	return nil
}

func validateAccountingDocumentOwnershipTx(ctx context.Context, tx *sql.Tx, orgID, documentID string) error {
	const query = `
SELECT type_code
FROM documents.documents
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	var typeCode string
	if err := tx.QueryRowContext(ctx, query, orgID, documentID).Scan(&typeCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return documents.ErrDocumentNotFound
		}
		return fmt.Errorf("load posting document: %w", err)
	}

	switch typeCode {
	case "invoice":
		var exists bool
		if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM accounting.invoice_documents
	WHERE org_id = $1
	  AND document_id = $2
);`, orgID, documentID).Scan(&exists); err != nil {
			return fmt.Errorf("check invoice payload: %w", err)
		}
		if !exists {
			return ErrInvoiceDocumentNotFound
		}
	case "payment_receipt":
		var exists bool
		if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM accounting.payment_receipt_documents
	WHERE org_id = $1
	  AND document_id = $2
);`, orgID, documentID).Scan(&exists); err != nil {
			return fmt.Errorf("check payment receipt payload: %w", err)
		}
		if !exists {
			return ErrPaymentReceiptNotFound
		}
	}

	return nil
}

func validatePartyContactPairTx(ctx context.Context, tx *sql.Tx, orgID, partyID, contactID string, invalidErr error) error {
	trimmedPartyID := strings.TrimSpace(partyID)
	trimmedContactID := strings.TrimSpace(contactID)
	if trimmedContactID != "" && trimmedPartyID == "" {
		return invalidErr
	}

	if trimmedPartyID == "" {
		return nil
	}

	var partyExists bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM parties.parties
	WHERE org_id = $1
	  AND id = $2
	  AND status = 'active'
);`, orgID, trimmedPartyID).Scan(&partyExists); err != nil {
		return fmt.Errorf("check party: %w", err)
	}
	if !partyExists {
		return invalidErr
	}

	if trimmedContactID == "" {
		return nil
	}

	var contactExists bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM parties.contacts
	WHERE org_id = $1
	  AND id = $2
	  AND party_id = $3
	  AND status = 'active'
);`, orgID, trimmedContactID, trimmedPartyID).Scan(&contactExists); err != nil {
		return fmt.Errorf("check contact: %w", err)
	}
	if !contactExists {
		return invalidErr
	}

	return nil
}

func isValidInvoiceRole(value string) bool {
	switch value {
	case InvoiceRoleSales, InvoiceRolePurchase:
		return true
	default:
		return false
	}
}

func normalizeInvoiceRole(value string) string {
	role := strings.TrimSpace(value)
	if role == "" {
		return InvoiceRoleSales
	}
	return role
}

func isValidPaymentReceiptDirection(value string) bool {
	switch value {
	case PaymentReceiptDirectionPayment, PaymentReceiptDirectionReceipt:
		return true
	default:
		return false
	}
}

func normalizePaymentReceiptDirection(value string) string {
	return strings.TrimSpace(value)
}

func ensureNoPeriodOverlapTx(ctx context.Context, tx *sql.Tx, orgID, periodID string, startOn, endOn time.Time) error {
	const query = `
SELECT 1
FROM accounting.periods
WHERE org_id = $1
  AND ($2::uuid IS NULL OR id <> $2::uuid)
  AND start_on <= $4
  AND end_on >= $3
LIMIT 1;`

	var exists int
	if err := tx.QueryRowContext(ctx, query, orgID, nullIfEmpty(strings.TrimSpace(periodID)), startOn, endOn).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("check accounting period overlap: %w", err)
	}
	return ErrAccountingPeriodOverlap
}

func normalizeEffectiveOn(value time.Time) time.Time {
	if value.IsZero() {
		value = time.Now().UTC()
	}
	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func normalizeOptionalDate(value time.Time) (time.Time, bool) {
	if value.IsZero() {
		return time.Time{}, false
	}
	return normalizeEffectiveOn(value), true
}

func isValidSetupStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case StatusActive, StatusInactive:
		return true
	default:
		return false
	}
}

func normalizePeriodRange(startOn, endOn time.Time) (time.Time, time.Time, error) {
	if startOn.IsZero() || endOn.IsZero() {
		return time.Time{}, time.Time{}, ErrInvalidAccountingPeriod
	}
	normalizedStart := normalizeEffectiveOn(startOn)
	normalizedEnd := normalizeEffectiveOn(endOn)
	if normalizedEnd.Before(normalizedStart) {
		return time.Time{}, time.Time{}, ErrInvalidAccountingPeriod
	}
	return normalizedStart, normalizedEnd, nil
}

func nullableDate(value time.Time, ok bool) any {
	if !ok {
		return nil
	}
	return value
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
	EffectiveOn        time.Time
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
	posted_by_user_id,
	effective_on
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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
	effective_on,
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
		params.EffectiveOn,
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
	effective_on,
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
	effective_on,
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

func createReversalTx(ctx context.Context, tx *sql.Tx, original JournalEntry, originalLines []JournalLine, reason string, effectiveOn time.Time, actor identityaccess.Actor) (JournalEntry, []JournalLine, error) {
	if strings.TrimSpace(reason) == "" {
		return JournalEntry{}, nil, ErrInvalidReversal
	}
	if err := validateAccountingPeriodOpenTx(ctx, tx, actor.OrgID, effectiveOn); err != nil {
		return JournalEntry{}, nil, err
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
		EffectiveOn:        effectiveOn,
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

func nullableString(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return value.String
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

func scanInvoiceDocument(row rowScanner) (InvoiceDocument, error) {
	var doc InvoiceDocument
	err := row.Scan(
		&doc.DocumentID,
		&doc.OrgID,
		&doc.InvoiceRole,
		&doc.BilledPartyID,
		&doc.BillingContactID,
		&doc.CurrencyCode,
		&doc.ReferenceValue,
		&doc.Summary,
		&doc.CreatedByUserID,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return InvoiceDocument{}, err
	}
	return doc, nil
}

func scanPaymentReceiptDocument(row rowScanner) (PaymentReceiptDocument, error) {
	var doc PaymentReceiptDocument
	err := row.Scan(
		&doc.DocumentID,
		&doc.OrgID,
		&doc.Direction,
		&doc.CounterpartyID,
		&doc.CounterpartyContactID,
		&doc.CurrencyCode,
		&doc.ReferenceValue,
		&doc.Summary,
		&doc.CreatedByUserID,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return PaymentReceiptDocument{}, err
	}
	return doc, nil
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

func scanAccountingPeriod(row rowScanner) (AccountingPeriod, error) {
	var period AccountingPeriod
	err := row.Scan(
		&period.ID,
		&period.OrgID,
		&period.PeriodCode,
		&period.StartOn,
		&period.EndOn,
		&period.Status,
		&period.ClosedByUserID,
		&period.ClosedAt,
		&period.CreatedByUserID,
		&period.CreatedAt,
		&period.UpdatedAt,
	)
	if err != nil {
		return AccountingPeriod{}, err
	}
	return period, nil
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
		&entry.EffectiveOn,
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
