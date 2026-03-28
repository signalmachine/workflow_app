package app

import (
	"time"

	"github.com/shopspring/decimal"
)

// CreateOrderRequest is the input for creating a new sales order.
type CreateOrderRequest struct {
	CompanyCode  string
	CustomerCode string
	Currency     string
	OrderDate    string
	Notes        string
	ExchangeRate decimal.Decimal
	Lines        []OrderLineInput
}

// OrderLineInput is a single line within a CreateOrderRequest.
type OrderLineInput struct {
	ProductCode string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal // zero means "use product default"
}

// CreateVendorRequest is the input for creating a new vendor.
type CreateVendorRequest struct {
	CompanyCode               string
	Code                      string
	Name                      string
	ContactPerson             string
	Email                     string
	Phone                     string
	Address                   string
	PaymentTermsDays          int
	APAccountCode             string
	DefaultExpenseAccountCode string
}

// CreatePurchaseOrderRequest is the input for creating a new purchase order.
type CreatePurchaseOrderRequest struct {
	CompanyCode string
	VendorCode  string
	PODate      string // YYYY-MM-DD
	Notes       string
	Lines       []POLineInput
}

// POLineInput is a single line within a CreatePurchaseOrderRequest.
type POLineInput struct {
	ProductCode        string
	Description        string
	Quantity           decimal.Decimal
	UnitCost           decimal.Decimal
	ExpenseAccountCode string
}

// ReceiveStockRequest is the input for recording a goods receipt into a warehouse.
type ReceiveStockRequest struct {
	CompanyCode       string
	ProductCode       string
	WarehouseCode     string
	CreditAccountCode string
	MovementDate      string
	Qty               decimal.Decimal
	UnitCost          decimal.Decimal
}

// ReceivePORequest is the input for recording goods/services received against a PO.
type ReceivePORequest struct {
	CompanyCode   string
	POID          int
	WarehouseCode string // optional; defaults to the company's default warehouse
	Lines         []ReceivedLineInput
}

// ReceivedLineInput is a single line in a ReceivePORequest.
type ReceivedLineInput struct {
	POLineID    int
	QtyReceived decimal.Decimal
}

// VendorInvoiceRequest is the input for recording a vendor invoice against a RECEIVED PO.
type VendorInvoiceRequest struct {
	CompanyCode   string
	POID          int
	InvoiceNumber string
	InvoiceDate   time.Time
	InvoiceAmount decimal.Decimal
}

// PayVendorRequest is the input for recording payment against an INVOICED PO.
type PayVendorRequest struct {
	CompanyCode     string
	POID            int
	BankAccountCode string
	PaymentDate     time.Time
}

// DirectVendorInvoiceLineInput is one line in a direct vendor invoice request.
type DirectVendorInvoiceLineInput struct {
	Description        string
	ExpenseAccountCode string
	Amount             decimal.Decimal
}

// DirectVendorInvoiceRequest records a vendor invoice without strict PO matching.
type DirectVendorInvoiceRequest struct {
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
	POID            *int
	Source          string
	ClosePO         bool
	CloseReason     string
	ClosedByUserID  *int
	CreatedByUserID *int
	Lines           []DirectVendorInvoiceLineInput
}

// PayVendorInvoiceRequest records a payment against a direct/bypass vendor invoice.
type PayVendorInvoiceRequest struct {
	CompanyCode     string
	VendorInvoiceID int
	BankAccountCode string
	PaymentDate     time.Time
	Amount          decimal.Decimal
	IdempotencyKey  string
}

// ClosePurchaseOrderRequest closes an open PO with a reason.
type ClosePurchaseOrderRequest struct {
	CompanyCode    string
	POID           int
	CloseReason    string
	ClosedByUserID *int
}

// CreateUserRequest is the input for creating a new user.
type CreateUserRequest struct {
	CompanyCode string
	Username    string
	Email       string
	Password    string // plain-text; hashed by the service
	Role        string // ACCOUNTANT | FINANCE_MANAGER | ADMIN
}

// RegisterCompanyRequest is the input for self-service tenant registration.
// Creates a new company and its first ADMIN user in a single atomic transaction.
type RegisterCompanyRequest struct {
	CompanyName string
	Username    string
	Email       string
	Password    string // plain-text; hashed by the service
}

// UpdateUserRoleRequest changes the role of a user within a company.
type UpdateUserRoleRequest struct {
	CompanyCode string
	UserID      int
	Role        string // ACCOUNTANT | FINANCE_MANAGER | ADMIN
}

// ManualJEControlAccountWarning is one detected control-account line in a manual JE request.
type ManualJEControlAccountWarning struct {
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
	ControlType string `json:"control_type"`
}

// ManualJEControlAccountAttemptRequest is the audit payload for a manual JE attempt.
type ManualJEControlAccountAttemptRequest struct {
	CompanyCode             string
	UserID                  *int
	Username                string
	Action                  string // validate | post
	PostingDate             string
	Narration               string
	AccountCodes            []string
	WarningDetails          []ManualJEControlAccountWarning
	EnforcementMode         string // off | warn | enforce
	OverrideControlAccounts bool
	OverrideReason          string
	IsBlocked               bool
}
