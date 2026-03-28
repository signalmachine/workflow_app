package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// PurchaseOrder represents a purchase order header.
type PurchaseOrder struct {
	ID                   int
	CompanyID            int
	VendorID             int
	VendorCode           string
	VendorName           string
	PONumber             *string
	Status               string
	PODate               string // YYYY-MM-DD
	ExpectedDeliveryDate *string
	Currency             string
	ExchangeRate         decimal.Decimal
	TotalTransaction     decimal.Decimal
	TotalBase            decimal.Decimal
	Notes                *string
	ApprovedAt           *time.Time
	ReceivedAt           *time.Time
	// Invoice fields (set by RecordVendorInvoice)
	InvoiceNumber    *string
	InvoiceDate      *string // YYYY-MM-DD
	InvoiceAmount    *decimal.Decimal
	PIDocumentNumber *string
	InvoicedAt       *time.Time
	// Payment fields (set by PayVendor)
	PaidAt *time.Time
	// Close fields (set by ClosePO)
	ClosedAt       *time.Time
	CloseReason    *string
	ClosedByUserID *int
	CreatedAt      time.Time
	Lines          []PurchaseOrderLine
}

// PurchaseOrderLine represents a single line on a purchase order.
type PurchaseOrderLine struct {
	ID                   int
	OrderID              int
	LineNumber           int
	ProductID            *int
	ProductCode          *string
	ProductName          *string
	Description          string
	Quantity             decimal.Decimal
	UnitCost             decimal.Decimal
	LineTotalTransaction decimal.Decimal
	LineTotalBase        decimal.Decimal
	ExpenseAccountCode   *string
}

// PurchaseOrderLineInput holds the fields required to create a purchase order line.
type PurchaseOrderLineInput struct {
	ProductCode        string
	Description        string
	Quantity           decimal.Decimal
	UnitCost           decimal.Decimal
	ExpenseAccountCode string
}

// ReceivedLine represents one PO line being received.
type ReceivedLine struct {
	POLineID    int             // references purchase_order_lines.id
	QtyReceived decimal.Decimal // quantity being received on this call
}

// DirectVendorInvoiceLineInput is one allocation line for a direct vendor invoice.
type DirectVendorInvoiceLineInput struct {
	Description        string
	ExpenseAccountCode string
	Amount             decimal.Decimal
}

// DirectVendorInvoiceInput captures direct/bypass purchase invoice posting input.
type DirectVendorInvoiceInput struct {
	CompanyID       int
	CompanyCode     string
	VendorID        int
	InvoiceNumber   string
	InvoiceDate     time.Time
	PostingDate     time.Time
	DocumentDate    time.Time
	Currency        string
	ExchangeRate    decimal.Decimal
	InvoiceAmount   decimal.Decimal
	IdempotencyKey  string
	Source          string // direct | po_strict | po_bypass
	POID            *int
	ClosePO         bool
	CloseReason     string
	ClosedByUserID  *int
	CreatedByUserID *int
	Lines           []DirectVendorInvoiceLineInput
}

// VendorInvoicePaymentInput captures payment posting input against vendor_invoices.
type VendorInvoicePaymentInput struct {
	CompanyID       int
	CompanyCode     string
	VendorInvoiceID int
	BankAccountCode string
	PaymentDate     time.Time
	Amount          decimal.Decimal
	IdempotencyKey  string
}

// VendorInvoice is a direct/bypass invoice header plus lines/payments.
type VendorInvoice struct {
	ID               int
	CompanyID        int
	VendorID         int
	VendorCode       string
	VendorName       string
	POID             *int
	Source           string
	Status           string
	InvoiceNumber    string
	InvoiceDate      string // YYYY-MM-DD
	Currency         string
	ExchangeRate     decimal.Decimal
	InvoiceAmount    decimal.Decimal
	AmountPaid       decimal.Decimal
	LastPaidAt       *time.Time
	IdempotencyKey   string
	PIDocumentNumber *string
	JournalEntryID   *int
	CreatedByUserID  *int
	CreatedAt        time.Time
	Lines            []VendorInvoiceLine
	Payments         []VendorInvoicePayment
}

// VendorInvoiceLine is a persisted allocation line under vendor_invoices.
type VendorInvoiceLine struct {
	ID                 int
	VendorInvoiceID    int
	LineNumber         int
	Description        string
	ExpenseAccountCode string
	Amount             decimal.Decimal
	CreatedAt          time.Time
}

// VendorInvoicePayment is one settlement posting against a vendor invoice.
type VendorInvoicePayment struct {
	ID                    int
	VendorInvoiceID       int
	PaymentDocumentNumber string
	PaymentAmount         decimal.Decimal
	PaymentDate           string // YYYY-MM-DD
	JournalEntryID        *int
	CreatedAt             time.Time
}

// PurchaseOrderService provides purchase order lifecycle operations.
type PurchaseOrderService interface {
	// CreatePO creates a new DRAFT purchase order with computed line totals.
	CreatePO(ctx context.Context, companyID, vendorID int, poDate time.Time, lines []PurchaseOrderLineInput, notes string) (*PurchaseOrder, error)

	// ApprovePO transitions a DRAFT PO to APPROVED, assigning a gapless PO number.
	// companyID must match the PO's company; returns an error if they differ.
	// It is idempotent: approving an already-APPROVED PO is a no-op.
	ApprovePO(ctx context.Context, companyID, poID int, docService DocumentService) error

	// ReceivePO records goods and/or services received against an APPROVED purchase order.
	// For physical-goods lines (product_id set): updates inventory via InventoryService.ReceiveStock
	// and links the movement to the PO line.
	// For service/expense lines (expense_account_code set, no product): posts DR expense / CR AP.
	// Transitions PO status to RECEIVED only when all PO lines are fully received.
	ReceivePO(ctx context.Context, poID int, warehouseCode, companyCode string,
		receivedLines []ReceivedLine, apAccountCode string,
		ledger *Ledger, docService DocumentService, inv InventoryService) error

	// RecordVendorInvoice records the vendor's invoice against a RECEIVED purchase order.
	// companyID must match the PO's company; returns an error if they differ.
	// Creates and posts a PI document (gapless number). Enforces strict match:
	// invoiceAmount must equal PO total_transaction. Transitions status to INVOICED.
	// Returns warning="" on success.
	RecordVendorInvoice(ctx context.Context, companyID, poID int, invoiceNumber string, invoiceDate time.Time,
		invoiceAmount decimal.Decimal, docService DocumentService) (warning string, err error)

	// PayVendor records payment against an INVOICED purchase order.
	// Posts DR AP / CR Bank and transitions status to PAID.
	PayVendor(ctx context.Context, poID int, bankAccountCode string, paymentDate time.Time,
		companyCode string, ledger *Ledger) error

	// ClosePO closes an open purchase order when invoicing is handled outside strict PO matching.
	ClosePO(ctx context.Context, companyID, poID int, reason string, closedByUserID *int) error

	// RecordDirectVendorInvoice posts a direct/bypass purchase invoice (PI) and persists vendor_invoices data.
	RecordDirectVendorInvoice(ctx context.Context, req DirectVendorInvoiceInput, ledger *Ledger) (*VendorInvoice, error)

	// PayVendorInvoice posts settlement (PV) against a vendor invoice and updates invoice payment status.
	PayVendorInvoice(ctx context.Context, req VendorInvoicePaymentInput, ledger *Ledger) (*VendorInvoice, error)

	// GetVendorInvoice returns one vendor invoice by ID for a company, including lines and payments.
	GetVendorInvoice(ctx context.Context, companyID, vendorInvoiceID int) (*VendorInvoice, error)

	// GetPO returns a purchase order by its internal ID for a company, including all lines.
	GetPO(ctx context.Context, companyID, poID int) (*PurchaseOrder, error)

	// GetPOs returns purchase orders for a company, optionally filtered by status.
	// An empty status string returns all orders.
	GetPOs(ctx context.Context, companyID int, status string) ([]PurchaseOrder, error)
}
