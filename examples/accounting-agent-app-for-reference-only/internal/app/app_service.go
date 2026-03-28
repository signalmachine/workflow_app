package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"accounting-agent/internal/ai"
	"accounting-agent/internal/core"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

type appService struct {
	pool                 *pgxpool.Pool
	ledger               *core.Ledger
	docService           core.DocumentService
	orderService         core.OrderService
	inventoryService     core.InventoryService
	reportingService     core.ReportingService
	userService          core.UserService
	vendorService        core.VendorService
	purchaseOrderService core.PurchaseOrderService
	agent                *ai.Agent
}

// NewAppService constructs an appService that satisfies ApplicationService.
func NewAppService(
	pool *pgxpool.Pool,
	ledger *core.Ledger,
	docService core.DocumentService,
	orderService core.OrderService,
	inventoryService core.InventoryService,
	reportingService core.ReportingService,
	userService core.UserService,
	vendorService core.VendorService,
	purchaseOrderService core.PurchaseOrderService,
	agent *ai.Agent,
) ApplicationService {
	return &appService{
		pool:                 pool,
		ledger:               ledger,
		docService:           docService,
		orderService:         orderService,
		inventoryService:     inventoryService,
		reportingService:     reportingService,
		userService:          userService,
		vendorService:        vendorService,
		purchaseOrderService: purchaseOrderService,
		agent:                agent,
	}
}

// HealthCheck verifies critical dependencies used by the application.
func (s *appService) HealthCheck(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// GetTrialBalance returns the trial balance for the given company.
// Reads journal_lines directly so results are always current.
func (s *appService) GetTrialBalance(ctx context.Context, companyCode string) (*TrialBalanceResult, error) {
	var companyID int
	var companyName, currency string
	if err := s.pool.QueryRow(ctx,
		"SELECT id, name, base_currency FROM companies WHERE company_code = $1", companyCode,
	).Scan(&companyID, &companyName, &currency); err != nil {
		return nil, fmt.Errorf("company %s not found: %w", companyCode, err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT a.code, a.name,
		       COALESCE(s.total_debit, 0) - COALESCE(s.total_credit, 0) AS net_balance
		FROM accounts a
		JOIN companies c ON c.id = a.company_id
		LEFT JOIN (
		    SELECT jl.account_id,
		           SUM(jl.debit_base)  AS total_debit,
		           SUM(jl.credit_base) AS total_credit
		    FROM journal_lines jl
		    JOIN journal_entries je ON je.id = jl.entry_id
		    WHERE je.company_id = $1
		    GROUP BY jl.account_id
		) s ON s.account_id = a.id
		WHERE c.id = $1
		  AND COALESCE(s.total_debit, 0) - COALESCE(s.total_credit, 0) <> 0
		ORDER BY a.code
	`, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query trial balance: %w", err)
	}
	defer rows.Close()

	var accounts []core.AccountBalance
	for rows.Next() {
		var b core.AccountBalance
		if err := rows.Scan(&b.Code, &b.Name, &b.Balance); err != nil {
			return nil, fmt.Errorf("failed to scan trial balance: %w", err)
		}
		accounts = append(accounts, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("trial balance row iteration: %w", err)
	}

	return &TrialBalanceResult{
		CompanyCode: companyCode,
		CompanyName: companyName,
		Currency:    currency,
		Accounts:    accounts,
	}, nil
}

// ListCustomers returns all active customers for a company.
func (s *appService) ListCustomers(ctx context.Context, companyCode string) (*CustomerListResult, error) {
	customers, err := s.orderService.GetCustomers(ctx, companyCode)
	if err != nil {
		return nil, err
	}
	return &CustomerListResult{Customers: customers}, nil
}

// ListProducts returns all active products for a company.
func (s *appService) ListProducts(ctx context.Context, companyCode string) (*ProductListResult, error) {
	products, err := s.orderService.GetProducts(ctx, companyCode)
	if err != nil {
		return nil, err
	}
	return &ProductListResult{Products: products}, nil
}

// ListOrders returns sales orders for a company, optionally filtered by status.
func (s *appService) ListOrders(ctx context.Context, companyCode string, status *string) (*OrderListResult, error) {
	orders, err := s.orderService.GetOrders(ctx, companyCode, status)
	if err != nil {
		return nil, err
	}
	return &OrderListResult{Orders: orders, CompanyCode: companyCode}, nil
}

// GetOrder returns a single sales order by numeric ID or order number string.
func (s *appService) GetOrder(ctx context.Context, ref, companyCode string) (*OrderResult, error) {
	order, err := s.resolveOrder(ctx, ref, companyCode)
	if err != nil {
		return nil, err
	}
	return &OrderResult{Order: order}, nil
}

// CreateOrder creates a new DRAFT sales order.
func (s *appService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResult, error) {
	lines := make([]core.OrderLineInput, len(req.Lines))
	for i, l := range req.Lines {
		lines[i] = core.OrderLineInput{
			ProductCode: l.ProductCode,
			Quantity:    l.Quantity,
			UnitPrice:   l.UnitPrice,
		}
	}

	exchangeRate := req.ExchangeRate
	if exchangeRate.IsZero() {
		exchangeRate = decimal.NewFromFloat(1.0)
	}

	orderDate := req.OrderDate
	if orderDate == "" {
		orderDate = time.Now().Format("2006-01-02")
	}

	order, err := s.orderService.CreateOrder(ctx, req.CompanyCode, req.CustomerCode, req.Currency,
		exchangeRate, orderDate, lines, req.Notes)
	if err != nil {
		return nil, err
	}
	return &OrderResult{Order: order}, nil
}

// ConfirmOrder transitions a DRAFT order to CONFIRMED, assigning an order number and reserving stock.
func (s *appService) ConfirmOrder(ctx context.Context, ref, companyCode string) (*OrderResult, error) {
	order, err := s.resolveOrder(ctx, ref, companyCode)
	if err != nil {
		return nil, err
	}
	order, err = s.orderService.ConfirmOrder(ctx, order.ID, s.docService, s.inventoryService)
	if err != nil {
		return nil, err
	}
	return &OrderResult{Order: order}, nil
}

// ShipOrder transitions a CONFIRMED order to SHIPPED, deducting inventory and booking COGS.
func (s *appService) ShipOrder(ctx context.Context, ref, companyCode string) (*OrderResult, error) {
	order, err := s.resolveOrder(ctx, ref, companyCode)
	if err != nil {
		return nil, err
	}
	order, err = s.orderService.ShipOrder(ctx, order.ID, s.inventoryService, s.ledger, s.docService)
	if err != nil {
		return nil, err
	}
	return &OrderResult{Order: order}, nil
}

// InvoiceOrder transitions a SHIPPED order to INVOICED, posting the sales invoice journal entry.
func (s *appService) InvoiceOrder(ctx context.Context, ref, companyCode string) (*OrderResult, error) {
	order, err := s.resolveOrder(ctx, ref, companyCode)
	if err != nil {
		return nil, err
	}
	order, err = s.orderService.InvoiceOrder(ctx, order.ID, s.ledger, s.docService)
	if err != nil {
		return nil, err
	}
	return &OrderResult{Order: order}, nil
}

// RecordPayment transitions an INVOICED order to PAID, posting the cash receipt journal entry.
func (s *appService) RecordPayment(ctx context.Context, ref, bankCode, companyCode string) (*OrderResult, error) {
	order, err := s.resolveOrder(ctx, ref, companyCode)
	if err != nil {
		return nil, err
	}
	if err := s.orderService.RecordPayment(ctx, order.ID, bankCode, "", s.ledger); err != nil {
		return nil, err
	}
	// Re-fetch to return the updated order with PAID status.
	order, err = s.orderService.GetOrder(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	return &OrderResult{Order: order}, nil
}

// ListWarehouses returns all active warehouses for a company.
func (s *appService) ListWarehouses(ctx context.Context, companyCode string) (*WarehouseListResult, error) {
	warehouses, err := s.inventoryService.GetWarehouses(ctx, companyCode)
	if err != nil {
		return nil, err
	}
	return &WarehouseListResult{Warehouses: warehouses}, nil
}

// GetStockLevels returns current stock levels for all inventory items in a company.
func (s *appService) GetStockLevels(ctx context.Context, companyCode string) (*StockResult, error) {
	levels, err := s.inventoryService.GetStockLevels(ctx, companyCode)
	if err != nil {
		return nil, err
	}
	return &StockResult{Levels: levels, CompanyCode: companyCode}, nil
}

// ReceiveStock records a goods receipt: increases qty_on_hand and books DR Inventory / CR creditAccount.
func (s *appService) ReceiveStock(ctx context.Context, req ReceiveStockRequest) error {
	warehouseCode := req.WarehouseCode
	if warehouseCode == "" {
		wh, err := s.inventoryService.GetDefaultWarehouse(ctx, req.CompanyCode)
		if err != nil {
			return fmt.Errorf("no active warehouse found: %w", err)
		}
		warehouseCode = wh.Code
	}

	creditAccount := req.CreditAccountCode
	if creditAccount == "" {
		creditAccount = "2000"
	}

	movementDate := req.MovementDate
	if movementDate == "" {
		movementDate = time.Now().Format("2006-01-02")
	}

	return s.inventoryService.ReceiveStock(ctx, req.CompanyCode, warehouseCode, req.ProductCode,
		req.Qty, req.UnitCost, movementDate, creditAccount, nil, s.ledger, s.docService)
}

// GetAccountStatement returns a chronological account statement with running balance.
func (s *appService) GetAccountStatement(ctx context.Context, companyCode, accountCode, fromDate, toDate string) (*AccountStatementResult, error) {
	var currency string
	if err := s.pool.QueryRow(ctx,
		"SELECT base_currency FROM companies WHERE company_code = $1", companyCode,
	).Scan(&currency); err != nil {
		return nil, fmt.Errorf("company %s not found: %w", companyCode, err)
	}

	lines, err := s.reportingService.GetAccountStatement(ctx, companyCode, accountCode, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	return &AccountStatementResult{
		CompanyCode: companyCode,
		AccountCode: accountCode,
		Currency:    currency,
		Lines:       lines,
	}, nil
}

// GetProfitAndLoss returns the P&L report for the given year and month.
func (s *appService) GetProfitAndLoss(ctx context.Context, companyCode string, year, month int) (*core.PLReport, error) {
	return s.reportingService.GetProfitAndLoss(ctx, companyCode, year, month)
}

// GetBalanceSheet returns the Balance Sheet as of the given date.
func (s *appService) GetBalanceSheet(ctx context.Context, companyCode, asOfDate string) (*core.BSReport, error) {
	return s.reportingService.GetBalanceSheet(ctx, companyCode, asOfDate)
}

// GetControlAccountReconciliation returns AR/AP/INVENTORY variance diagnostics.
func (s *appService) GetControlAccountReconciliation(ctx context.Context, companyCode, asOfDate string) (*core.ControlAccountReconciliationReport, error) {
	return s.reportingService.GetControlAccountReconciliation(ctx, companyCode, asOfDate)
}

// GetDocumentTypeGovernance returns posting mix diagnostics for JE governance.
func (s *appService) GetDocumentTypeGovernance(ctx context.Context, companyCode, fromDate, toDate string) (*core.DocumentTypeGovernanceReport, error) {
	return s.reportingService.GetDocumentTypeGovernance(ctx, companyCode, fromDate, toDate)
}

// GetManualJEControlAccountHits returns manual JE lines posted to control accounts.
func (s *appService) GetManualJEControlAccountHits(ctx context.Context, companyCode, fromDate, toDate string) (*ManualControlAccountHitsResult, error) {
	hits, err := s.reportingService.GetManualJEControlAccountHits(ctx, companyCode, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	return &ManualControlAccountHitsResult{
		CompanyCode: companyCode,
		FromDate:    fromDate,
		ToDate:      toDate,
		Hits:        hits,
	}, nil
}

// GetManualJEControlAccountWarnings detects control-account usage for manual JE lines.
func (s *appService) GetManualJEControlAccountWarnings(ctx context.Context, companyCode string, accountCodes []string) ([]ManualJEControlAccountWarning, error) {
	if len(accountCodes) == 0 {
		return nil, nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT a.code, a.name, COALESCE(a.control_type, '')
		FROM accounts a
		JOIN companies c ON c.id = a.company_id
		WHERE c.company_code = $1
		  AND a.is_control_account = true
		  AND a.code = ANY($2)
		ORDER BY a.code
	`, companyCode, accountCodes)
	if err != nil {
		return nil, fmt.Errorf("lookup control accounts: %w", err)
	}
	defer rows.Close()

	warnings := make([]ManualJEControlAccountWarning, 0)
	for rows.Next() {
		var w ManualJEControlAccountWarning
		if err := rows.Scan(&w.AccountCode, &w.AccountName, &w.ControlType); err != nil {
			return nil, fmt.Errorf("scan control account warning: %w", err)
		}
		warnings = append(warnings, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate control account warnings: %w", err)
	}

	return warnings, nil
}

// RecordManualJEControlAccountAttempt writes a non-blocking audit record for manual JE attempts.
func (s *appService) RecordManualJEControlAccountAttempt(ctx context.Context, req ManualJEControlAccountAttemptRequest) error {
	if len(req.WarningDetails) == 0 {
		return nil
	}

	var companyID int
	if err := s.pool.QueryRow(ctx, "SELECT id FROM companies WHERE company_code = $1", req.CompanyCode).Scan(&companyID); err != nil {
		return fmt.Errorf("resolve company for manual JE control-account audit: %w", err)
	}

	warningJSON, err := json.Marshal(req.WarningDetails)
	if err != nil {
		return fmt.Errorf("marshal warning details: %w", err)
	}
	mode := req.EnforcementMode
	if mode == "" {
		mode = "warn"
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO manual_je_control_account_audits
		    (company_id, user_id, username, action, posting_date, narration, account_codes, warning_details, enforcement_mode, override_control_accounts, override_reason, is_blocked)
		VALUES
		    ($1, $2, $3, $4, NULLIF($5, '')::date, $6, $7, $8::jsonb, NULLIF($9, ''), $10, NULLIF($11, ''), $12)
	`, companyID, req.UserID, req.Username, req.Action, req.PostingDate, req.Narration, req.AccountCodes, string(warningJSON), mode, req.OverrideControlAccounts, req.OverrideReason, req.IsBlocked)
	if err != nil {
		return fmt.Errorf("insert manual JE control-account audit: %w", err)
	}

	return nil
}

// SetJournalEntryCreatedBy attaches the authenticated user to a posted JE.
func (s *appService) SetJournalEntryCreatedBy(ctx context.Context, companyCode, idempotencyKey string, userID int) error {
	if idempotencyKey == "" {
		return nil
	}

	_, err := s.pool.Exec(ctx, `
		UPDATE journal_entries je
		SET created_by_user_id = $1
		FROM companies c
		WHERE c.id = je.company_id
		  AND c.company_code = $2
		  AND je.idempotency_key = $3
	`, userID, companyCode, idempotencyKey)
	if err != nil {
		return fmt.Errorf("set created_by_user_id for JE: %w", err)
	}

	return nil
}

// InterpretEvent sends a natural language event description to the AI agent and returns
// either a Proposal or a clarification request.
func (s *appService) InterpretEvent(ctx context.Context, text, companyCode string) (*AIResult, error) {
	coa, err := s.fetchCoA(ctx, companyCode)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chart of accounts: %w", err)
	}

	documentTypes, err := s.fetchDocumentTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch document types: %w", err)
	}

	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch company: %w", err)
	}

	response, err := s.agent.InterpretEvent(ctx, text, coa, documentTypes, company)
	if err != nil {
		return nil, err
	}

	if response.IsClarificationRequest {
		return &AIResult{
			IsClarification:      true,
			ClarificationMessage: response.Clarification.Message,
		}, nil
	}

	return &AIResult{
		IsClarification: false,
		Proposal:        response.Proposal,
	}, nil
}

// InterpretDomainAction routes a natural language input through the agentic tool loop.
// It builds a ToolRegistry with read tools for the current domain, delegates to the agent,
// and translates the AgentDomainResult to a DomainActionResult for the adapter layer.
// attachments is variadic — REPL/CLI callers omit it.
func (s *appService) InterpretDomainAction(ctx context.Context, text, companyCode string, attachments ...Attachment) (*DomainActionResult, error) {
	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch company: %w", err)
	}

	registry := s.buildToolRegistry(ctx, companyCode)

	// Convert app.Attachment → ai.Attachment for the agent layer.
	aiAtts := make([]ai.Attachment, len(attachments))
	for i, a := range attachments {
		aiAtts[i] = ai.Attachment{MimeType: a.MimeType, Data: a.Data}
	}

	result, err := s.agent.InterpretDomainAction(ctx, text, company, registry, aiAtts)
	if err != nil {
		return nil, err
	}

	return &DomainActionResult{
		Kind:             DomainActionKind(result.Kind),
		Answer:           result.Answer,
		Question:         result.Question,
		Context:          result.Context,
		ToolName:         result.ToolName,
		ToolArgs:         result.ToolArgs,
		EventDescription: result.EventDescription,
	}, nil
}

// ExecuteWriteTool executes a previously proposed write tool action after human confirmation.
// Args are parsed from the map stored at proposal time; returned string is a JSON-encoded summary.
func (s *appService) ExecuteWriteTool(ctx context.Context, companyCode, toolName string, args map[string]any) (string, error) {
	intArg := func(key string) (int, error) {
		raw, ok := args[key]
		if !ok {
			return 0, fmt.Errorf("missing required argument %q", key)
		}
		switch v := raw.(type) {
		case float64:
			if v != math.Trunc(v) {
				return 0, fmt.Errorf("argument %q must be an integer, got %v", key, v)
			}
			return int(v), nil
		case int:
			return v, nil
		case int64:
			return int(v), nil
		case string:
			i, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				return 0, fmt.Errorf("argument %q must be an integer, got %q", key, v)
			}
			return i, nil
		default:
			return 0, fmt.Errorf("argument %q must be an integer", key)
		}
	}
	strArg := func(key string, required bool) (string, error) {
		raw, ok := args[key]
		if !ok {
			if required {
				return "", fmt.Errorf("missing required argument %q", key)
			}
			return "", nil
		}
		s, ok := raw.(string)
		if !ok {
			return "", fmt.Errorf("argument %q must be a string", key)
		}
		s = strings.TrimSpace(s)
		if required && s == "" {
			return "", fmt.Errorf("argument %q cannot be empty", key)
		}
		return s, nil
	}
	decArg := func(key string, required bool) (decimal.Decimal, error) {
		raw, ok := args[key]
		if !ok {
			if required {
				return decimal.Zero, fmt.Errorf("missing required argument %q", key)
			}
			return decimal.Zero, nil
		}
		switch v := raw.(type) {
		case string:
			d, err := decimal.NewFromString(strings.TrimSpace(v))
			if err != nil {
				return decimal.Zero, fmt.Errorf("argument %q must be a valid decimal", key)
			}
			return d, nil
		case float64:
			return decimal.NewFromFloat(v), nil
		case int:
			return decimal.NewFromInt(int64(v)), nil
		case int64:
			return decimal.NewFromInt(v), nil
		default:
			return decimal.Zero, fmt.Errorf("argument %q must be a decimal number", key)
		}
	}
	decArgFromAny := func(raw any) (decimal.Decimal, error) {
		switch v := raw.(type) {
		case nil:
			return decimal.Zero, fmt.Errorf("missing value")
		case string:
			d, err := decimal.NewFromString(strings.TrimSpace(v))
			if err != nil {
				return decimal.Zero, err
			}
			return d, nil
		case float64:
			return decimal.NewFromFloat(v), nil
		case int:
			return decimal.NewFromInt(int64(v)), nil
		case int64:
			return decimal.NewFromInt(v), nil
		default:
			return decimal.Zero, fmt.Errorf("unsupported decimal value type")
		}
	}
	switch toolName {

	case "approve_po":
		poID, err := intArg("po_id")
		if err != nil {
			return "", err
		}
		if poID <= 0 {
			return "", fmt.Errorf("argument %q must be > 0", "po_id")
		}
		result, err := s.ApprovePurchaseOrder(ctx, companyCode, poID)
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]any{
			"message":   "Purchase order approved.",
			"po_number": result.PurchaseOrder.PONumber,
			"status":    result.PurchaseOrder.Status,
		})
		return string(b), nil

	case "create_vendor":
		code, err := strArg("code", true)
		if err != nil {
			return "", err
		}
		name, err := strArg("name", true)
		if err != nil {
			return "", err
		}
		contactPerson, err := strArg("contact_person", false)
		if err != nil {
			return "", err
		}
		email, err := strArg("email", false)
		if err != nil {
			return "", err
		}
		phone, err := strArg("phone", false)
		if err != nil {
			return "", err
		}
		address, err := strArg("address", false)
		if err != nil {
			return "", err
		}
		apAccountCode, err := strArg("ap_account_code", false)
		if err != nil {
			return "", err
		}
		defaultExpenseAccountCode, err := strArg("default_expense_account_code", false)
		if err != nil {
			return "", err
		}
		req := CreateVendorRequest{
			CompanyCode:               companyCode,
			Code:                      code,
			Name:                      name,
			ContactPerson:             contactPerson,
			Email:                     email,
			Phone:                     phone,
			Address:                   address,
			APAccountCode:             apAccountCode,
			DefaultExpenseAccountCode: defaultExpenseAccountCode,
		}
		if rawPT, ok := args["payment_terms_days"]; ok {
			switch pt := rawPT.(type) {
			case float64:
				if pt != math.Trunc(pt) {
					return "", fmt.Errorf("argument %q must be an integer", "payment_terms_days")
				}
				req.PaymentTermsDays = int(pt)
			case int:
				req.PaymentTermsDays = pt
			default:
				return "", fmt.Errorf("argument %q must be an integer", "payment_terms_days")
			}
		}
		result, err := s.CreateVendor(ctx, req)
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]any{
			"message": "Vendor created.",
			"code":    result.Vendor.Code,
			"name":    result.Vendor.Name,
		})
		return string(b), nil

	case "create_purchase_order":
		// Parse nested lines via JSON round-trip.
		type lineIn struct {
			ProductCode        string `json:"product_code"`
			Description        string `json:"description"`
			Quantity           string `json:"quantity"`
			UnitCost           string `json:"unit_cost"`
			ExpenseAccountCode string `json:"expense_account_code"`
		}
		type poIn struct {
			VendorCode string   `json:"vendor_code"`
			PODate     string   `json:"po_date"`
			Notes      string   `json:"notes"`
			Lines      []lineIn `json:"lines"`
		}
		raw, _ := json.Marshal(args)
		var inp poIn
		if err := json.Unmarshal(raw, &inp); err != nil {
			return "", fmt.Errorf("invalid create_purchase_order args: %w", err)
		}
		if strings.TrimSpace(inp.VendorCode) == "" {
			return "", fmt.Errorf("argument %q cannot be empty", "vendor_code")
		}
		if strings.TrimSpace(inp.PODate) == "" {
			return "", fmt.Errorf("argument %q cannot be empty", "po_date")
		}
		if len(inp.Lines) == 0 {
			return "", fmt.Errorf("argument %q must include at least one line", "lines")
		}
		lines := make([]POLineInput, len(inp.Lines))
		for i, l := range inp.Lines {
			qty, err := decimal.NewFromString(l.Quantity)
			if err != nil {
				return "", fmt.Errorf("invalid create_purchase_order line %d quantity: %w", i+1, err)
			}
			unitCost, err := decimal.NewFromString(l.UnitCost)
			if err != nil {
				return "", fmt.Errorf("invalid create_purchase_order line %d unit_cost: %w", i+1, err)
			}
			lines[i] = POLineInput{
				ProductCode:        l.ProductCode,
				Description:        l.Description,
				Quantity:           qty,
				UnitCost:           unitCost,
				ExpenseAccountCode: l.ExpenseAccountCode,
			}
		}
		result, err := s.CreatePurchaseOrder(ctx, CreatePurchaseOrderRequest{
			CompanyCode: companyCode,
			VendorCode:  inp.VendorCode,
			PODate:      inp.PODate,
			Notes:       inp.Notes,
			Lines:       lines,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]any{
			"message": "Purchase order created as DRAFT.",
			"po_id":   result.PurchaseOrder.ID,
			"status":  result.PurchaseOrder.Status,
		})
		return string(b), nil

	case "receive_po":
		type lineIn struct {
			POLineID    int     `json:"po_line_id"`
			QtyReceived float64 `json:"qty_received"`
		}
		type receiveIn struct {
			POID          int      `json:"po_id"`
			WarehouseCode string   `json:"warehouse_code"`
			Lines         []lineIn `json:"lines"`
		}
		raw, _ := json.Marshal(args)
		var inp receiveIn
		if err := json.Unmarshal(raw, &inp); err != nil {
			return "", fmt.Errorf("invalid receive_po args: %w", err)
		}
		if inp.POID <= 0 {
			return "", fmt.Errorf("argument %q must be > 0", "po_id")
		}
		if len(inp.Lines) == 0 {
			return "", fmt.Errorf("argument %q must include at least one line", "lines")
		}
		lines := make([]ReceivedLineInput, len(inp.Lines))
		for i, l := range inp.Lines {
			if l.POLineID <= 0 {
				return "", fmt.Errorf("receive_po line %d: po_line_id must be > 0", i+1)
			}
			if l.QtyReceived <= 0 {
				return "", fmt.Errorf("receive_po line %d: qty_received must be > 0", i+1)
			}
			lines[i] = ReceivedLineInput{
				POLineID:    l.POLineID,
				QtyReceived: decimal.NewFromFloat(l.QtyReceived),
			}
		}
		result, err := s.ReceivePurchaseOrder(ctx, ReceivePORequest{
			CompanyCode:   companyCode,
			POID:          inp.POID,
			WarehouseCode: inp.WarehouseCode,
			Lines:         lines,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]any{
			"message":        "Goods received against PO.",
			"lines_received": result.LinesReceived,
			"status":         result.PurchaseOrder.Status,
		})
		return string(b), nil

	case "record_vendor_invoice":
		poID, err := intArg("po_id")
		if err != nil {
			return "", err
		}
		if poID <= 0 {
			return "", fmt.Errorf("argument %q must be > 0", "po_id")
		}
		invoiceNumber, err := strArg("invoice_number", true)
		if err != nil {
			return "", err
		}
		invoiceDateStr, err := strArg("invoice_date", true)
		if err != nil {
			return "", err
		}
		rawAmt, ok := args["invoice_amount"]
		if !ok {
			return "", fmt.Errorf("missing required argument %q", "invoice_amount")
		}
		var amt decimal.Decimal
		switch v := rawAmt.(type) {
		case string:
			amt, err = decimal.NewFromString(strings.TrimSpace(v))
			if err != nil {
				return "", fmt.Errorf("invalid invoice_amount %q", v)
			}
		case float64:
			amt = decimal.NewFromFloat(v)
		default:
			return "", fmt.Errorf("argument %q must be a string or number", "invoice_amount")
		}
		invoiceDate, err := time.Parse("2006-01-02", invoiceDateStr)
		if err != nil {
			return "", fmt.Errorf("invalid invoice_date: %w", err)
		}
		result, err := s.RecordVendorInvoice(ctx, VendorInvoiceRequest{
			CompanyCode:   companyCode,
			POID:          poID,
			InvoiceNumber: invoiceNumber,
			InvoiceDate:   invoiceDate,
			InvoiceAmount: amt,
		})
		if err != nil {
			return "", err
		}
		msg := "Vendor invoice recorded. PI document: " + result.PIDocumentNumber
		if result.Warning != "" {
			msg += " Warning: " + result.Warning
		}
		b, _ := json.Marshal(map[string]any{"message": msg, "status": result.PurchaseOrder.Status})
		return string(b), nil

	case "pay_vendor":
		poID, err := intArg("po_id")
		if err != nil {
			return "", err
		}
		if poID <= 0 {
			return "", fmt.Errorf("argument %q must be > 0", "po_id")
		}
		bankAccountCode, err := strArg("bank_account_code", true)
		if err != nil {
			return "", err
		}
		paymentDateStr, err := strArg("payment_date", true)
		if err != nil {
			return "", err
		}
		paymentDate, err := time.Parse("2006-01-02", paymentDateStr)
		if err != nil {
			return "", fmt.Errorf("invalid payment_date: %w", err)
		}
		result, err := s.PayVendor(ctx, PayVendorRequest{
			CompanyCode:     companyCode,
			POID:            poID,
			BankAccountCode: bankAccountCode,
			PaymentDate:     paymentDate,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]any{
			"message": "Vendor payment posted.",
			"status":  result.PurchaseOrder.Status,
		})
		return string(b), nil

	case "record_direct_vendor_invoice":
		type lineIn struct {
			Description        string `json:"description"`
			ExpenseAccountCode string `json:"expense_account_code"`
			Amount             any    `json:"amount"`
		}
		type invoiceIn struct {
			VendorID       int      `json:"vendor_id"`
			InvoiceNumber  string   `json:"invoice_number"`
			InvoiceDate    string   `json:"invoice_date"`
			PostingDate    string   `json:"posting_date"`
			DocumentDate   string   `json:"document_date"`
			Currency       string   `json:"currency"`
			ExchangeRate   any      `json:"exchange_rate"`
			InvoiceAmount  any      `json:"invoice_amount"`
			IdempotencyKey string   `json:"idempotency_key"`
			Source         string   `json:"source"`
			POID           *int     `json:"po_id"`
			ClosePO        bool     `json:"close_po"`
			CloseReason    string   `json:"close_reason"`
			Lines          []lineIn `json:"lines"`
		}
		raw, _ := json.Marshal(args)
		var inp invoiceIn
		if err := json.Unmarshal(raw, &inp); err != nil {
			return "", fmt.Errorf("invalid record_direct_vendor_invoice args: %w", err)
		}
		if inp.VendorID <= 0 {
			return "", fmt.Errorf("argument %q must be > 0", "vendor_id")
		}
		if strings.TrimSpace(inp.InvoiceNumber) == "" {
			return "", fmt.Errorf("argument %q cannot be empty", "invoice_number")
		}
		invoiceDate, err := time.Parse("2006-01-02", strings.TrimSpace(inp.InvoiceDate))
		if err != nil {
			return "", fmt.Errorf("invalid invoice_date: %w", err)
		}
		postingDate := invoiceDate
		if strings.TrimSpace(inp.PostingDate) != "" {
			postingDate, err = time.Parse("2006-01-02", strings.TrimSpace(inp.PostingDate))
			if err != nil {
				return "", fmt.Errorf("invalid posting_date: %w", err)
			}
		}
		documentDate := invoiceDate
		if strings.TrimSpace(inp.DocumentDate) != "" {
			documentDate, err = time.Parse("2006-01-02", strings.TrimSpace(inp.DocumentDate))
			if err != nil {
				return "", fmt.Errorf("invalid document_date: %w", err)
			}
		}
		if len(inp.Lines) == 0 {
			return "", fmt.Errorf("argument %q must include at least one line", "lines")
		}
		exchangeRate := decimal.NewFromInt(1)
		if inp.ExchangeRate != nil {
			switch v := inp.ExchangeRate.(type) {
			case string:
				if strings.TrimSpace(v) != "" {
					exchangeRate, err = decimal.NewFromString(strings.TrimSpace(v))
					if err != nil {
						return "", fmt.Errorf("invalid exchange_rate: %w", err)
					}
				}
			case float64:
				exchangeRate = decimal.NewFromFloat(v)
			default:
				return "", fmt.Errorf("invalid exchange_rate type")
			}
		}
		invoiceAmount, err := decArgFromAny(inp.InvoiceAmount)
		if err != nil {
			return "", fmt.Errorf("invalid invoice_amount: %w", err)
		}
		lines := make([]DirectVendorInvoiceLineInput, len(inp.Lines))
		for i, l := range inp.Lines {
			amt, err := decArgFromAny(l.Amount)
			if err != nil {
				return "", fmt.Errorf("line %d invalid amount: %w", i+1, err)
			}
			lines[i] = DirectVendorInvoiceLineInput{
				Description:        strings.TrimSpace(l.Description),
				ExpenseAccountCode: strings.TrimSpace(l.ExpenseAccountCode),
				Amount:             amt,
			}
		}
		idempotencyKey := strings.TrimSpace(inp.IdempotencyKey)
		if idempotencyKey == "" {
			idempotencyKey = fmt.Sprintf("vendor-invoice-%d-%d", inp.VendorID, time.Now().UnixNano())
		}
		result, err := s.RecordDirectVendorInvoice(ctx, DirectVendorInvoiceRequest{
			CompanyCode:    companyCode,
			VendorID:       inp.VendorID,
			InvoiceNumber:  inp.InvoiceNumber,
			InvoiceDate:    invoiceDate,
			PostingDate:    postingDate,
			DocumentDate:   documentDate,
			Currency:       strings.ToUpper(strings.TrimSpace(inp.Currency)),
			ExchangeRate:   exchangeRate,
			InvoiceAmount:  invoiceAmount,
			IdempotencyKey: idempotencyKey,
			POID:           inp.POID,
			Source:         strings.ToLower(strings.TrimSpace(inp.Source)),
			ClosePO:        inp.ClosePO,
			CloseReason:    strings.TrimSpace(inp.CloseReason),
			Lines:          lines,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]any{
			"message":               "Direct vendor invoice recorded.",
			"vendor_invoice_id":     result.VendorInvoice.ID,
			"pi_document_number":    result.VendorInvoice.PIDocumentNumber,
			"vendor_invoice_status": result.VendorInvoice.Status,
		})
		return string(b), nil

	case "pay_vendor_invoice":
		vendorInvoiceID, err := intArg("vendor_invoice_id")
		if err != nil {
			return "", err
		}
		if vendorInvoiceID <= 0 {
			return "", fmt.Errorf("argument %q must be > 0", "vendor_invoice_id")
		}
		bankAccountCode, err := strArg("bank_account_code", true)
		if err != nil {
			return "", err
		}
		amount, err := decArg("amount", true)
		if err != nil {
			return "", err
		}
		paymentDate := time.Now()
		if paymentDateStr, err := strArg("payment_date", false); err != nil {
			return "", err
		} else if paymentDateStr != "" {
			paymentDate, err = time.Parse("2006-01-02", paymentDateStr)
			if err != nil {
				return "", fmt.Errorf("invalid payment_date: %w", err)
			}
		}
		idempotencyKey, err := strArg("idempotency_key", false)
		if err != nil {
			return "", err
		}
		if idempotencyKey == "" {
			idempotencyKey = fmt.Sprintf("pay-vendor-invoice-%d-%d", vendorInvoiceID, time.Now().UnixNano())
		}
		result, err := s.PayVendorInvoice(ctx, PayVendorInvoiceRequest{
			CompanyCode:     companyCode,
			VendorInvoiceID: vendorInvoiceID,
			BankAccountCode: bankAccountCode,
			PaymentDate:     paymentDate,
			Amount:          amount,
			IdempotencyKey:  idempotencyKey,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]any{
			"message":               "Vendor invoice payment posted.",
			"vendor_invoice_id":     result.VendorInvoice.ID,
			"vendor_invoice_status": result.VendorInvoice.Status,
			"amount_paid":           result.VendorInvoice.AmountPaid.StringFixed(2),
		})
		return string(b), nil

	case "close_purchase_order":
		poID, err := intArg("po_id")
		if err != nil {
			return "", err
		}
		if poID <= 0 {
			return "", fmt.Errorf("argument %q must be > 0", "po_id")
		}
		closeReason, err := strArg("close_reason", true)
		if err != nil {
			return "", err
		}
		result, err := s.ClosePurchaseOrder(ctx, ClosePurchaseOrderRequest{
			CompanyCode: companyCode,
			POID:        poID,
			CloseReason: closeReason,
		})
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]any{
			"message": "Purchase order closed.",
			"po_id":   result.PurchaseOrder.ID,
			"status":  result.PurchaseOrder.Status,
		})
		return string(b), nil

	default:
		return "", fmt.Errorf("unknown write tool: %q", toolName)
	}
}

// ListVendors returns all active vendors for a company.
func (s *appService) ListVendors(ctx context.Context, companyCode string) (*VendorsResult, error) {
	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return nil, err
	}
	vendors, err := s.vendorService.GetVendors(ctx, company.ID)
	if err != nil {
		return nil, err
	}
	return &VendorsResult{Vendors: vendors}, nil
}

// CreateVendor creates a new vendor record for the given company.
func (s *appService) CreateVendor(ctx context.Context, req CreateVendorRequest) (*VendorResult, error) {
	company, err := s.fetchCompany(ctx, req.CompanyCode)
	if err != nil {
		return nil, err
	}
	vendor, err := s.vendorService.CreateVendor(ctx, company.ID, core.VendorInput{
		Code:                      req.Code,
		Name:                      req.Name,
		ContactPerson:             req.ContactPerson,
		Email:                     req.Email,
		Phone:                     req.Phone,
		Address:                   req.Address,
		PaymentTermsDays:          req.PaymentTermsDays,
		APAccountCode:             req.APAccountCode,
		DefaultExpenseAccountCode: req.DefaultExpenseAccountCode,
	})
	if err != nil {
		return nil, err
	}
	return &VendorResult{Vendor: vendor}, nil
}

// GetPurchaseOrder returns a single purchase order by its internal ID, validating company ownership.
func (s *appService) GetPurchaseOrder(ctx context.Context, companyCode string, poID int) (*PurchaseOrderResult, error) {
	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return nil, err
	}
	po, err := s.purchaseOrderService.GetPO(ctx, company.ID, poID)
	if err != nil {
		return nil, err
	}
	return &PurchaseOrderResult{PurchaseOrder: po}, nil
}

// ListPurchaseOrders returns purchase orders for a company, optionally filtered by status.
func (s *appService) ListPurchaseOrders(ctx context.Context, companyCode, status string) (*PurchaseOrdersResult, error) {
	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return nil, err
	}
	orders, err := s.purchaseOrderService.GetPOs(ctx, company.ID, status)
	if err != nil {
		return nil, err
	}
	return &PurchaseOrdersResult{Orders: orders}, nil
}

// CreatePurchaseOrder creates a new DRAFT purchase order.
func (s *appService) CreatePurchaseOrder(ctx context.Context, req CreatePurchaseOrderRequest) (*PurchaseOrderResult, error) {
	company, err := s.fetchCompany(ctx, req.CompanyCode)
	if err != nil {
		return nil, err
	}
	vendor, err := s.vendorService.GetVendorByCode(ctx, company.ID, req.VendorCode)
	if err != nil {
		return nil, fmt.Errorf("vendor %q: %w", req.VendorCode, err)
	}

	poDate, err := time.Parse("2006-01-02", req.PODate)
	if err != nil {
		return nil, fmt.Errorf("invalid po_date %q: %w", req.PODate, err)
	}

	var lines []core.PurchaseOrderLineInput
	for _, l := range req.Lines {
		lines = append(lines, core.PurchaseOrderLineInput{
			ProductCode:        l.ProductCode,
			Description:        l.Description,
			Quantity:           l.Quantity,
			UnitCost:           l.UnitCost,
			ExpenseAccountCode: l.ExpenseAccountCode,
		})
	}

	po, err := s.purchaseOrderService.CreatePO(ctx, company.ID, vendor.ID, poDate, lines, req.Notes)
	if err != nil {
		return nil, err
	}
	return &PurchaseOrderResult{PurchaseOrder: po}, nil
}

// ApprovePurchaseOrder transitions a DRAFT PO to APPROVED, assigning a gapless PO number.
func (s *appService) ApprovePurchaseOrder(ctx context.Context, companyCode string, poID int) (*PurchaseOrderResult, error) {
	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return nil, err
	}
	if err := s.purchaseOrderService.ApprovePO(ctx, company.ID, poID, s.docService); err != nil {
		return nil, err
	}
	po, err := s.purchaseOrderService.GetPO(ctx, company.ID, poID)
	if err != nil {
		return nil, err
	}
	return &PurchaseOrderResult{PurchaseOrder: po}, nil
}

// buildToolRegistry constructs the ToolRegistry for Phase 7.5 with 5 read tools:
// search_accounts, search_customers, search_products, get_stock_levels, get_warehouses.
// Tool handlers are closures that capture the pool and companyCode.
func (s *appService) buildToolRegistry(ctx context.Context, companyCode string) *ai.ToolRegistry {
	registry := ai.NewToolRegistry()

	registry.Register(ai.ToolDefinition{
		Name:        "search_accounts",
		Description: "Search the chart of accounts by name or code. Returns top matching accounts with code, name, type, and current balance.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search text to match against account name or code (e.g. 'accounts receivable', '1200', 'sales revenue').",
				},
			},
			"required": []string{"query"},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			query, _ := params["query"].(string)
			return s.searchAccounts(hctx, companyCode, query)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "search_customers",
		Description: "Search the customer master by name or code. Returns matching customers with code, name, and contact information.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search text to match against customer name or code (e.g. 'Acme', 'C001').",
				},
			},
			"required": []string{"query"},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			query, _ := params["query"].(string)
			return s.searchCustomers(hctx, companyCode, query)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "search_products",
		Description: "Search the product catalogue by name or code. Returns matching products with code, name, and unit price.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search text to match against product name or code (e.g. 'Widget', 'P001').",
				},
			},
			"required": []string{"query"},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			query, _ := params["query"].(string)
			return s.searchProducts(hctx, companyCode, query)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_stock_levels",
		Description: "Get current inventory stock levels. Optionally filter by product code or warehouse code.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"product_code": map[string]any{
					"type":        "string",
					"description": "Filter to a specific product code (optional).",
				},
				"warehouse_code": map[string]any{
					"type":        "string",
					"description": "Filter to a specific warehouse code (optional).",
				},
			},
			"required": []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			productCode, _ := params["product_code"].(string)
			warehouseCode, _ := params["warehouse_code"].(string)
			return s.getStockLevels(hctx, companyCode, productCode, warehouseCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_warehouses",
		Description: "Get all active warehouses for the company.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           map[string]any{},
			"required":             []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			return s.getWarehousesJSON(hctx, companyCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_account_balance",
		Description: "Get the current balance for a specific account code. Returns the net debit position (positive = net debit, negative = net credit).",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"account_code": map[string]any{
					"type":        "string",
					"description": "The account code to query (e.g. '1200', '4000').",
				},
			},
			"required": []string{"account_code"},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			accountCode, _ := params["account_code"].(string)
			return s.getAccountBalanceJSON(hctx, companyCode, accountCode)
		},
	})

	// Phase 11 vendor tools
	registry.Register(ai.ToolDefinition{
		Name:        "get_vendors",
		Description: "List all active vendors for the company. Returns vendor code, name, contact, payment terms, and AP account.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           map[string]any{},
			"required":             []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			return s.getVendorsJSON(hctx, companyCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "search_vendors",
		Description: "Search vendors by name or code using partial match. Returns matching active vendors.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search text to match against vendor name or code (e.g. 'Acme', 'V001').",
				},
			},
			"required": []string{"query"},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			query, _ := params["query"].(string)
			return s.searchVendors(hctx, companyCode, query)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_vendor_info",
		Description: "Get full details for a specific vendor by vendor code, including contact, address, payment terms, and AP account.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"vendor_code": map[string]any{
					"type":        "string",
					"description": "The vendor code (e.g. 'V001').",
				},
			},
			"required": []string{"vendor_code"},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			vendorCode, _ := params["vendor_code"].(string)
			return s.getVendorInfoJSON(hctx, companyCode, vendorCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "create_vendor",
		Description: "Propose creating a new vendor. The user must confirm before the vendor is saved. Requires at least a code and name.",
		IsReadTool:  false, // write tool — requires human confirmation
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"code": map[string]any{
					"type":        "string",
					"description": "Unique vendor code (e.g. 'V004').",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Full vendor name.",
				},
				"contact_person": map[string]any{
					"type":        "string",
					"description": "Name of the primary contact person (optional).",
				},
				"email": map[string]any{
					"type":        "string",
					"description": "Vendor contact email (optional).",
				},
				"phone": map[string]any{
					"type":        "string",
					"description": "Vendor phone number (optional).",
				},
				"address": map[string]any{
					"type":        "string",
					"description": "Vendor mailing address (optional).",
				},
				"payment_terms_days": map[string]any{
					"type":        "integer",
					"description": "Payment due in N days (default 30).",
				},
				"ap_account_code": map[string]any{
					"type":        "string",
					"description": "Accounts Payable account code (default '2000').",
				},
				"default_expense_account_code": map[string]any{
					"type":        "string",
					"description": "Default expense account code for this vendor's invoices (optional).",
				},
			},
			"required": []string{"code", "name"},
		},
		Handler: nil, // write tool — no autonomous execution
	})

	// Phase 12 purchase order tools
	registry.Register(ai.ToolDefinition{
		Name:        "get_purchase_orders",
		Description: "List purchase orders for the company. Optionally filter by status: DRAFT, APPROVED, RECEIVED, INVOICED, PAID, CLOSED. Empty status returns all orders.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"status": map[string]any{
					"type":        "string",
					"description": "Filter by PO status (optional). One of: DRAFT, APPROVED, RECEIVED, INVOICED, PAID, CLOSED.",
				},
			},
			"required": []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			status, _ := params["status"].(string)
			return s.getPurchaseOrdersJSON(hctx, companyCode, status)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_open_pos",
		Description: "List all open (DRAFT or APPROVED) purchase orders for the company — orders not yet received, invoiced, or paid.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           map[string]any{},
			"required":             []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			return s.getOpenPOsJSON(hctx, companyCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "create_purchase_order",
		Description: "Propose creating a new purchase order for a vendor. The user must confirm before the PO is saved. Requires vendor code, PO date, and at least one line item.",
		IsReadTool:  false, // write tool — requires human confirmation
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"vendor_code": map[string]any{
					"type":        "string",
					"description": "Vendor code (e.g. 'V001').",
				},
				"po_date": map[string]any{
					"type":        "string",
					"description": "Purchase order date in YYYY-MM-DD format.",
				},
				"notes": map[string]any{
					"type":        "string",
					"description": "Optional notes or instructions for the PO.",
				},
				"lines": map[string]any{
					"type":        "array",
					"description": "List of PO line items. At least one required.",
					"items": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"product_code": map[string]any{
								"type":        "string",
								"description": "Product code if ordering a catalogued product (optional).",
							},
							"description": map[string]any{
								"type":        "string",
								"description": "Line item description.",
							},
							"quantity": map[string]any{
								"type":        "string",
								"description": "Quantity ordered as a decimal string (e.g. '2.5').",
							},
							"unit_cost": map[string]any{
								"type":        "string",
								"description": "Unit cost in the PO currency as a decimal string.",
							},
							"expense_account_code": map[string]any{
								"type":        "string",
								"description": "Expense account code for non-inventory lines (optional).",
							},
						},
						"required": []string{"description", "quantity", "unit_cost"},
					},
				},
			},
			"required": []string{"vendor_code", "po_date", "lines"},
		},
		Handler: nil, // write tool — no autonomous execution
	})

	registry.Register(ai.ToolDefinition{
		Name:        "approve_po",
		Description: "Propose approving a DRAFT purchase order. Assigns a gapless PO number on approval. The user must confirm before the action is executed.",
		IsReadTool:  false, // write tool — requires human confirmation
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"po_id": map[string]any{
					"type":        "integer",
					"description": "Internal ID of the purchase order to approve.",
				},
			},
			"required": []string{"po_id"},
		},
		Handler: nil, // write tool — no autonomous execution
	})

	// Phase 13 goods receipt tools
	registry.Register(ai.ToolDefinition{
		Name:        "check_stock_availability",
		Description: "Check current inventory stock levels for products, optionally filtered by an APPROVED purchase order. Returns on-hand, reserved, and available quantities per product/warehouse, plus PO line details when a po_id is provided.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"po_id": map[string]any{
					"type":        "integer",
					"description": "Optional: purchase order ID to show stock levels for products in that PO.",
				},
				"product_code": map[string]any{
					"type":        "string",
					"description": "Optional: filter to a specific product code.",
				},
			},
			"required": []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			poID := 0
			if v, ok := params["po_id"].(float64); ok {
				poID = int(v)
			}
			productCode, _ := params["product_code"].(string)
			return s.checkStockAvailabilityJSON(hctx, companyCode, poID, productCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "receive_po",
		Description: "Propose recording goods/services received against an APPROVED purchase order. Updates inventory stock levels and creates the DR Inventory / CR AP accounting entry. The user must confirm before the receipt is posted.",
		IsReadTool:  false, // write tool — requires human confirmation
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"po_id": map[string]any{
					"type":        "integer",
					"description": "Internal ID of the approved purchase order to receive against.",
				},
				"warehouse_code": map[string]any{
					"type":        "string",
					"description": "Warehouse code to receive goods into (optional; defaults to the company's default warehouse).",
				},
				"lines": map[string]any{
					"type":        "array",
					"description": "Lines to receive. At least one required.",
					"items": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"po_line_id": map[string]any{
								"type":        "integer",
								"description": "Internal ID of the purchase order line being received.",
							},
							"qty_received": map[string]any{
								"type":        "number",
								"description": "Quantity received on this line.",
							},
						},
						"required": []string{"po_line_id", "qty_received"},
					},
				},
			},
			"required": []string{"po_id", "lines"},
		},
		Handler: nil, // write tool — no autonomous execution
	})

	// Phase 14 vendor invoice + payment tools
	registry.Register(ai.ToolDefinition{
		Name:        "get_outstanding_vendor_invoices",
		Description: "Get outstanding vendor invoices (INVOICED purchase orders, not yet paid). Returns base-currency total plus grouped transaction-currency totals. This is an operational metric, not the ledger AP balance.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"vendor_code": map[string]any{
					"type":        "string",
					"description": "Optional: filter to a specific vendor code to see AP balance for that vendor only.",
				},
			},
			"required": []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			vendorCode, _ := params["vendor_code"].(string)
			return s.getOutstandingVendorInvoicesJSON(hctx, companyCode, vendorCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_ar_balance",
		Description: "Get the Accounts Receivable balance from the ledger (journal_lines). Returns the net debit balance of the configured AR account. Use this for financial questions about how much customers owe.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           map[string]any{},
			"required":             []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			return s.getARBalanceJSON(hctx, companyCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_ap_balance",
		Description: "Get the Accounts Payable balance from the ledger (journal_lines). Returns the net credit balance of the AP account. Use this for financial questions about how much is owed to vendors.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           map[string]any{},
			"required":             []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			return s.getAPBalanceJSON(hctx, companyCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_control_account_reconciliation",
		Description: "Get lightweight AR/AP/INVENTORY reconciliation diagnostics (GL vs operational balances) with variances.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"as_of_date": map[string]any{
					"type":        "string",
					"description": "Optional report date in YYYY-MM-DD. GL is evaluated up to this date; operational values are current-state snapshot.",
				},
			},
			"required": []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			asOfDate, _ := params["as_of_date"].(string)
			return s.getControlAccountReconciliationJSON(hctx, companyCode, asOfDate)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_manual_je_control_account_hits",
		Description: "List manual journal-entry lines that posted directly to control accounts (AR/AP/INVENTORY).",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"from_date": map[string]any{
					"type":        "string",
					"description": "Optional start date in YYYY-MM-DD format.",
				},
				"to_date": map[string]any{
					"type":        "string",
					"description": "Optional end date in YYYY-MM-DD format.",
				},
			},
			"required": []string{},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			fromDate, _ := params["from_date"].(string)
			toDate, _ := params["to_date"].(string)
			return s.getManualJEControlAccountHitsJSON(hctx, companyCode, fromDate, toDate)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "get_vendor_payment_history",
		Description: "Get payment history for a vendor, including currency, exchange rate, transaction amounts, base amounts, invoice metadata, and payment date.",
		IsReadTool:  true,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"vendor_code": map[string]any{
					"type":        "string",
					"description": "The vendor code to retrieve payment history for (e.g. 'V001').",
				},
			},
			"required": []string{"vendor_code"},
		},
		Handler: func(hctx context.Context, params map[string]any) (string, error) {
			vendorCode, _ := params["vendor_code"].(string)
			return s.getVendorPaymentHistoryJSON(hctx, companyCode, vendorCode)
		},
	})

	registry.Register(ai.ToolDefinition{
		Name:        "record_vendor_invoice",
		Description: "Propose recording a vendor invoice against a RECEIVED purchase order. Creates a PI document number and transitions PO to INVOICED. The user must confirm before the action is executed.",
		IsReadTool:  false, // write tool — requires human confirmation
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"po_id": map[string]any{
					"type":        "integer",
					"description": "Internal ID of the RECEIVED purchase order to invoice.",
				},
				"invoice_number": map[string]any{
					"type":        "string",
					"description": "Vendor's invoice number as printed on the bill.",
				},
				"invoice_date": map[string]any{
					"type":        "string",
					"description": "Invoice date in YYYY-MM-DD format.",
				},
				"invoice_amount": map[string]any{
					"type":        "number",
					"description": "Total invoice amount in base currency. If it differs from the PO total by more than 5%, a warning is produced.",
				},
			},
			"required": []string{"po_id", "invoice_number", "invoice_date", "invoice_amount"},
		},
		Handler: nil, // write tool — no autonomous execution
	})

	registry.Register(ai.ToolDefinition{
		Name:        "pay_vendor",
		Description: "Propose paying a vendor for an INVOICED purchase order. Posts DR AP / CR Bank and transitions PO to PAID. The user must confirm before the payment is posted.",
		IsReadTool:  false, // write tool — requires human confirmation
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"po_id": map[string]any{
					"type":        "integer",
					"description": "Internal ID of the INVOICED purchase order to pay.",
				},
				"bank_account_code": map[string]any{
					"type":        "string",
					"description": "Account code of the bank account to pay from (e.g. '1100').",
				},
				"payment_date": map[string]any{
					"type":        "string",
					"description": "Payment date in YYYY-MM-DD format.",
				},
			},
			"required": []string{"po_id", "bank_account_code", "payment_date"},
		},
		Handler: nil, // write tool — no autonomous execution
	})

	registry.Register(ai.ToolDefinition{
		Name:        "record_direct_vendor_invoice",
		Description: "Propose recording a direct or PO-bypass vendor invoice. Posts a PI entry and stores vendor invoice details. The user must confirm before execution.",
		IsReadTool:  false,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"vendor_id": map[string]any{
					"type":        "integer",
					"description": "Internal ID of the vendor.",
				},
				"invoice_number": map[string]any{
					"type":        "string",
					"description": "Vendor invoice number.",
				},
				"invoice_date": map[string]any{
					"type":        "string",
					"description": "Invoice date in YYYY-MM-DD format.",
				},
				"posting_date": map[string]any{
					"type":        "string",
					"description": "Posting date in YYYY-MM-DD format (optional; defaults to invoice_date).",
				},
				"document_date": map[string]any{
					"type":        "string",
					"description": "Document date in YYYY-MM-DD format (optional; defaults to invoice_date).",
				},
				"currency": map[string]any{
					"type":        "string",
					"description": "Transaction currency (optional; defaults to company base currency).",
				},
				"exchange_rate": map[string]any{
					"type":        "string",
					"description": "Exchange rate string (optional; defaults to 1).",
				},
				"invoice_amount": map[string]any{
					"type":        "string",
					"description": "Total invoice amount.",
				},
				"idempotency_key": map[string]any{
					"type":        "string",
					"description": "Optional idempotency key for replay-safe posting.",
				},
				"source": map[string]any{
					"type":        "string",
					"description": "Invoice source: direct or po_bypass (optional).",
				},
				"po_id": map[string]any{
					"type":        "integer",
					"description": "Optional PO ID for bypass traceability.",
				},
				"close_po": map[string]any{
					"type":        "boolean",
					"description": "Set true to close linked PO after bypass invoice.",
				},
				"close_reason": map[string]any{
					"type":        "string",
					"description": "Required when close_po is true.",
				},
				"lines": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"description": map[string]any{"type": "string"},
							"expense_account_code": map[string]any{
								"type": "string",
							},
							"amount": map[string]any{
								"type": "string",
							},
						},
						"required": []string{"expense_account_code", "amount"},
					},
				},
			},
			"required": []string{"vendor_id", "invoice_number", "invoice_date", "invoice_amount", "lines"},
		},
		Handler: nil,
	})

	registry.Register(ai.ToolDefinition{
		Name:        "pay_vendor_invoice",
		Description: "Propose paying a direct/bypass vendor invoice. Posts PV and updates invoice status. The user must confirm before execution.",
		IsReadTool:  false,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"vendor_invoice_id": map[string]any{
					"type":        "integer",
					"description": "Internal vendor invoice ID.",
				},
				"bank_account_code": map[string]any{
					"type":        "string",
					"description": "Payment account code (e.g. 1100).",
				},
				"payment_date": map[string]any{
					"type":        "string",
					"description": "Payment date in YYYY-MM-DD format (optional; defaults to today).",
				},
				"amount": map[string]any{
					"type":        "string",
					"description": "Payment amount.",
				},
				"idempotency_key": map[string]any{
					"type":        "string",
					"description": "Optional idempotency key.",
				},
			},
			"required": []string{"vendor_invoice_id", "bank_account_code", "amount"},
		},
		Handler: nil,
	})

	registry.Register(ai.ToolDefinition{
		Name:        "close_purchase_order",
		Description: "Propose closing an open purchase order with a reason when invoicing is handled outside strict PO flow.",
		IsReadTool:  false,
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"po_id": map[string]any{
					"type":        "integer",
					"description": "Internal purchase order ID.",
				},
				"close_reason": map[string]any{
					"type":        "string",
					"description": "Reason for closing the PO.",
				},
			},
			"required": []string{"po_id", "close_reason"},
		},
		Handler: nil,
	})

	return registry
}

// RecordVendorInvoice records the vendor's invoice against a RECEIVED PO.
func (s *appService) RecordVendorInvoice(ctx context.Context, req VendorInvoiceRequest) (*VendorInvoiceResult, error) {
	company, err := s.fetchCompany(ctx, req.CompanyCode)
	if err != nil {
		return nil, err
	}
	warning, err := s.purchaseOrderService.RecordVendorInvoice(
		ctx, company.ID, req.POID, req.InvoiceNumber, req.InvoiceDate, req.InvoiceAmount, s.docService,
	)
	if err != nil {
		return nil, err
	}
	po, err := s.purchaseOrderService.GetPO(ctx, company.ID, req.POID)
	if err != nil {
		return nil, err
	}
	piDocNum := ""
	if po.PIDocumentNumber != nil {
		piDocNum = *po.PIDocumentNumber
	}
	return &VendorInvoiceResult{PurchaseOrder: po, PIDocumentNumber: piDocNum, Warning: warning}, nil
}

// PayVendor records payment against an INVOICED PO.
func (s *appService) PayVendor(ctx context.Context, req PayVendorRequest) (*PaymentResult, error) {
	company, err := s.fetchCompany(ctx, req.CompanyCode)
	if err != nil {
		return nil, err
	}
	if err := s.purchaseOrderService.PayVendor(
		ctx, req.POID, req.BankAccountCode, req.PaymentDate, req.CompanyCode, s.ledger,
	); err != nil {
		return nil, err
	}
	po, err := s.purchaseOrderService.GetPO(ctx, company.ID, req.POID)
	if err != nil {
		return nil, err
	}
	return &PaymentResult{PurchaseOrder: po}, nil
}

// RecordDirectVendorInvoice records a direct/bypass vendor invoice and posts PI.
func (s *appService) RecordDirectVendorInvoice(ctx context.Context, req DirectVendorInvoiceRequest) (*DirectVendorInvoiceResult, error) {
	if err := enforceFlexiblePurchaseInvoiceMode(); err != nil {
		return nil, err
	}
	company, err := s.fetchCompany(ctx, req.CompanyCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		req.IdempotencyKey = fmt.Sprintf("vendor-invoice-%d-%d", req.VendorID, time.Now().UnixNano())
	}
	lines := make([]core.DirectVendorInvoiceLineInput, len(req.Lines))
	for i, l := range req.Lines {
		lines[i] = core.DirectVendorInvoiceLineInput{
			Description:        l.Description,
			ExpenseAccountCode: l.ExpenseAccountCode,
			Amount:             l.Amount,
		}
	}
	invoice, err := s.purchaseOrderService.RecordDirectVendorInvoice(ctx, core.DirectVendorInvoiceInput{
		CompanyID:       company.ID,
		CompanyCode:     req.CompanyCode,
		VendorID:        req.VendorID,
		InvoiceNumber:   req.InvoiceNumber,
		InvoiceDate:     req.InvoiceDate,
		PostingDate:     req.PostingDate,
		DocumentDate:    req.DocumentDate,
		Currency:        req.Currency,
		ExchangeRate:    req.ExchangeRate,
		InvoiceAmount:   req.InvoiceAmount,
		IdempotencyKey:  req.IdempotencyKey,
		Source:          req.Source,
		POID:            req.POID,
		ClosePO:         req.ClosePO,
		CloseReason:     req.CloseReason,
		ClosedByUserID:  req.ClosedByUserID,
		CreatedByUserID: req.CreatedByUserID,
		Lines:           lines,
	}, s.ledger)
	if err != nil {
		return nil, err
	}
	return &DirectVendorInvoiceResult{VendorInvoice: invoice}, nil
}

// PayVendorInvoice records settlement against a vendor invoice and posts PV.
func (s *appService) PayVendorInvoice(ctx context.Context, req PayVendorInvoiceRequest) (*VendorInvoicePaymentResult, error) {
	if err := enforceFlexiblePurchaseInvoiceMode(); err != nil {
		return nil, err
	}
	company, err := s.fetchCompany(ctx, req.CompanyCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		req.IdempotencyKey = fmt.Sprintf("pay-vendor-invoice-%d-%d", req.VendorInvoiceID, time.Now().UnixNano())
	}
	invoice, err := s.purchaseOrderService.PayVendorInvoice(ctx, core.VendorInvoicePaymentInput{
		CompanyID:       company.ID,
		CompanyCode:     req.CompanyCode,
		VendorInvoiceID: req.VendorInvoiceID,
		BankAccountCode: req.BankAccountCode,
		PaymentDate:     req.PaymentDate,
		Amount:          req.Amount,
		IdempotencyKey:  req.IdempotencyKey,
	}, s.ledger)
	if err != nil {
		return nil, err
	}
	return &VendorInvoicePaymentResult{VendorInvoice: invoice}, nil
}

// ClosePurchaseOrder closes an open PO with a required reason.
func (s *appService) ClosePurchaseOrder(ctx context.Context, req ClosePurchaseOrderRequest) (*PurchaseOrderResult, error) {
	if err := enforceFlexiblePurchaseInvoiceMode(); err != nil {
		return nil, err
	}
	company, err := s.fetchCompany(ctx, req.CompanyCode)
	if err != nil {
		return nil, err
	}
	if err := s.purchaseOrderService.ClosePO(ctx, company.ID, req.POID, req.CloseReason, req.ClosedByUserID); err != nil {
		return nil, err
	}
	po, err := s.purchaseOrderService.GetPO(ctx, company.ID, req.POID)
	if err != nil {
		return nil, err
	}
	return &PurchaseOrderResult{PurchaseOrder: po}, nil
}

func enforceFlexiblePurchaseInvoiceMode() error {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("PURCHASE_INVOICE_POLICY_MODE")))
	if mode == "" {
		mode = "strict"
	}
	if mode != "flexible" {
		return fmt.Errorf("PURCHASE_INVOICE_POLICY_STRICT: set PURCHASE_INVOICE_POLICY_MODE=flexible to enable direct/bypass purchase invoice flows")
	}
	return nil
}

// ReceivePurchaseOrder records goods and/or services received against an APPROVED PO.
func (s *appService) ReceivePurchaseOrder(ctx context.Context, req ReceivePORequest) (*POReceiptResult, error) {
	company, err := s.fetchCompany(ctx, req.CompanyCode)
	if err != nil {
		return nil, err
	}

	warehouseCode := req.WarehouseCode
	if warehouseCode == "" {
		wh, err := s.inventoryService.GetDefaultWarehouse(ctx, req.CompanyCode)
		if err != nil {
			return nil, fmt.Errorf("no active warehouse found: %w", err)
		}
		warehouseCode = wh.Code
	}

	// Look up the vendor's AP account code via the PO
	var apAccountCode string
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(v.ap_account_code, '2000')
		FROM purchase_orders po
		JOIN vendors v ON v.id = po.vendor_id
		WHERE po.id = $1 AND po.company_id = $2`,
		req.POID, company.ID,
	).Scan(&apAccountCode); err != nil {
		return nil, fmt.Errorf("resolve AP account for PO %d: %w", req.POID, err)
	}

	// Convert request lines to domain lines
	domainLines := make([]core.ReceivedLine, len(req.Lines))
	for i, l := range req.Lines {
		domainLines[i] = core.ReceivedLine{
			POLineID:    l.POLineID,
			QtyReceived: l.QtyReceived,
		}
	}

	if err := s.purchaseOrderService.ReceivePO(ctx, req.POID, warehouseCode, req.CompanyCode,
		domainLines, apAccountCode, s.ledger, s.docService, s.inventoryService); err != nil {
		return nil, err
	}

	po, err := s.purchaseOrderService.GetPO(ctx, company.ID, req.POID)
	if err != nil {
		return nil, err
	}
	return &POReceiptResult{PurchaseOrder: po, LinesReceived: len(req.Lines)}, nil
}

// checkStockAvailabilityJSON returns current stock levels, optionally scoped to a PO's products.
func (s *appService) checkStockAvailabilityJSON(ctx context.Context, companyCode string, poID int, productCode string) (string, error) {
	result := map[string]any{}

	if poID > 0 {
		// Include PO context
		company, err := s.fetchCompany(ctx, companyCode)
		if err != nil {
			return "", err
		}
		po, err := s.purchaseOrderService.GetPO(ctx, company.ID, poID)
		if err != nil {
			return "", fmt.Errorf("PO %d not found: %w", poID, err)
		}
		result["po_id"] = po.ID
		result["po_status"] = po.Status
		if po.PONumber != nil {
			result["po_number"] = *po.PONumber
		}
		lines := make([]map[string]any, 0, len(po.Lines))
		for _, l := range po.Lines {
			m := map[string]any{
				"po_line_id":  l.ID,
				"line_number": l.LineNumber,
				"description": l.Description,
				"quantity":    l.Quantity.String(),
				"unit_cost":   l.UnitCost.String(),
			}
			if l.ProductCode != nil {
				m["product_code"] = *l.ProductCode
			}
			if l.ExpenseAccountCode != nil {
				m["expense_account_code"] = *l.ExpenseAccountCode
			}
			lines = append(lines, m)
		}
		result["po_lines"] = lines
	}

	stockLevels, err := s.inventoryService.GetStockLevels(ctx, companyCode)
	if err != nil {
		return "", err
	}

	var filtered []map[string]any
	for _, sl := range stockLevels {
		if productCode != "" && sl.ProductCode != productCode {
			continue
		}
		filtered = append(filtered, map[string]any{
			"product_code":   sl.ProductCode,
			"product_name":   sl.ProductName,
			"warehouse_code": sl.WarehouseCode,
			"warehouse_name": sl.WarehouseName,
			"on_hand":        sl.OnHand.String(),
			"reserved":       sl.Reserved.String(),
			"available":      sl.Available.String(),
			"unit_cost":      sl.UnitCost.String(),
		})
	}
	if filtered == nil {
		filtered = []map[string]any{}
	}
	result["stock_levels"] = filtered

	data, _ := json.Marshal(result)
	return string(data), nil
}

// getOutstandingVendorInvoicesJSON returns PO workflow outstanding invoices.
// This is an operational metric (INVOICED POs), not the ledger AP balance.
func (s *appService) getOutstandingVendorInvoicesJSON(ctx context.Context, companyCode, vendorCode string) (string, error) {
	var baseCurrency string
	if err := s.pool.QueryRow(ctx,
		"SELECT base_currency FROM companies WHERE company_code = $1",
		companyCode,
	).Scan(&baseCurrency); err != nil {
		return "", fmt.Errorf("resolve company base currency: %w", err)
	}

	query := `
		SELECT COALESCE(SUM(
			COALESCE(po.invoice_amount * po.exchange_rate, po.total_base)
		), 0),
		COUNT(*)
		FROM purchase_orders po
		JOIN vendors v ON v.id = po.vendor_id
		JOIN companies c ON c.id = po.company_id
		WHERE c.company_code = $1 AND po.status = 'INVOICED'`
	args := []any{companyCode}
	if vendorCode != "" {
		query += " AND v.code = $2"
		args = append(args, vendorCode)
	}

	var totalBase decimal.Decimal
	var count int
	if err := s.pool.QueryRow(ctx, query, args...).Scan(&totalBase, &count); err != nil {
		return "", fmt.Errorf("get outstanding vendor invoices: %w", err)
	}

	groupQuery := `
		SELECT po.currency,
		       COALESCE(SUM(COALESCE(po.invoice_amount, po.total_transaction)), 0),
		       COUNT(*)
		FROM purchase_orders po
		JOIN vendors v ON v.id = po.vendor_id
		JOIN companies c ON c.id = po.company_id
		WHERE c.company_code = $1 AND po.status = 'INVOICED'`
	groupArgs := []any{companyCode}
	if vendorCode != "" {
		groupQuery += " AND v.code = $2"
		groupArgs = append(groupArgs, vendorCode)
	}
	groupQuery += " GROUP BY po.currency ORDER BY po.currency"

	rows, err := s.pool.Query(ctx, groupQuery, groupArgs...)
	if err != nil {
		return "", fmt.Errorf("get outstanding vendor invoices by currency: %w", err)
	}
	defer rows.Close()

	outstandingByCurrency := make([]map[string]any, 0)
	for rows.Next() {
		var currency string
		var amountTx decimal.Decimal
		var invoiceCount int
		if err := rows.Scan(&currency, &amountTx, &invoiceCount); err != nil {
			return "", fmt.Errorf("scan outstanding invoices by currency: %w", err)
		}
		outstandingByCurrency = append(outstandingByCurrency, map[string]any{
			"currency":           currency,
			"amount_transaction": amountTx.StringFixed(2),
			"invoice_count":      invoiceCount,
		})
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate outstanding invoices by currency: %w", err)
	}

	result := map[string]any{
		"base_currency":                  baseCurrency,
		"outstanding_invoice_total_base": totalBase.StringFixed(2),
		"outstanding_invoice_count":      count,
		"outstanding_by_currency":        outstandingByCurrency,
		"note":                           "Operational INVOICED purchase orders. Amount fields are explicit by currency context.",
	}
	// Backward-compatible alias; value is always base currency.
	result["outstanding_invoice_total"] = totalBase.StringFixed(2)
	if vendorCode != "" {
		result["vendor_code"] = vendorCode
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

// getARBalanceJSON returns the ledger-sourced Accounts Receivable balance.
func (s *appService) getARBalanceJSON(ctx context.Context, companyCode string) (string, error) {
	var arAccountCode string
	if err := s.pool.QueryRow(ctx, `
		SELECT ar.account_code
		FROM account_rules ar
		JOIN companies c ON c.id = ar.company_id
		WHERE c.company_code = $1 AND ar.rule_type = 'AR'`,
		companyCode,
	).Scan(&arAccountCode); err != nil {
		data, _ := json.Marshal(map[string]any{
			"error": "AR account rule not configured for this company",
		})
		return string(data), nil
	}

	var accountName string
	var netDebit decimal.Decimal
	if err := s.pool.QueryRow(ctx, `
		SELECT a.name,
		       COALESCE(SUM(jl.debit_base), 0) - COALESCE(SUM(jl.credit_base), 0)
		FROM accounts a
		JOIN companies c ON c.id = a.company_id
		LEFT JOIN journal_lines jl ON jl.account_id = a.id
		LEFT JOIN journal_entries je ON je.id = jl.entry_id AND je.company_id = c.id
		WHERE c.company_code = $1 AND a.code = $2
		GROUP BY a.name`,
		companyCode, arAccountCode,
	).Scan(&accountName, &netDebit); err != nil {
		data, _ := json.Marshal(map[string]any{
			"error": fmt.Sprintf("AR account %s not found or has no activity", arAccountCode),
		})
		return string(data), nil
	}

	data, _ := json.Marshal(map[string]any{
		"ar_account_code": arAccountCode,
		"ar_account_name": accountName,
		"ar_balance":      netDebit.StringFixed(2),
		"note":            "Ledger AR balance from journal_lines. Positive = amount owed by customers.",
	})
	return string(data), nil
}

// getAPBalanceJSON returns the ledger-sourced Accounts Payable balance.
func (s *appService) getAPBalanceJSON(ctx context.Context, companyCode string) (string, error) {
	var apAccountCode string
	if err := s.pool.QueryRow(ctx, `
		SELECT ar.account_code
		FROM account_rules ar
		JOIN companies c ON c.id = ar.company_id
		WHERE c.company_code = $1 AND ar.rule_type = 'AP'`,
		companyCode,
	).Scan(&apAccountCode); err != nil {
		data, _ := json.Marshal(map[string]any{
			"error": "AP account rule not configured for this company",
		})
		return string(data), nil
	}

	var accountName string
	var netDebit decimal.Decimal
	if err := s.pool.QueryRow(ctx, `
		SELECT a.name,
		       COALESCE(SUM(jl.debit_base), 0) - COALESCE(SUM(jl.credit_base), 0)
		FROM accounts a
		JOIN companies c ON c.id = a.company_id
		LEFT JOIN journal_lines jl ON jl.account_id = a.id
		LEFT JOIN journal_entries je ON je.id = jl.entry_id AND je.company_id = c.id
		WHERE c.company_code = $1 AND a.code = $2
		GROUP BY a.name`,
		companyCode, apAccountCode,
	).Scan(&accountName, &netDebit); err != nil {
		data, _ := json.Marshal(map[string]any{
			"error": fmt.Sprintf("AP account %s not found or has no activity", apAccountCode),
		})
		return string(data), nil
	}

	apBalance := netDebit.Neg()
	data, _ := json.Marshal(map[string]any{
		"ap_account_code": apAccountCode,
		"ap_account_name": accountName,
		"ap_balance":      apBalance.StringFixed(2),
		"note":            "Ledger AP balance from journal_lines. Positive = amount owed to vendors.",
	})
	return string(data), nil
}

// getManualJEControlAccountHitsJSON returns manual JE control-account hits as JSON.
func (s *appService) getManualJEControlAccountHitsJSON(ctx context.Context, companyCode, fromDate, toDate string) (string, error) {
	hits, err := s.reportingService.GetManualJEControlAccountHits(ctx, companyCode, fromDate, toDate)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(map[string]any{
		"company_code": companyCode,
		"from_date":    fromDate,
		"to_date":      toDate,
		"count":        len(hits),
		"hits":         hits,
	})
	return string(data), nil
}

// getControlAccountReconciliationJSON returns AR/AP/INVENTORY reconciliation diagnostics as JSON.
func (s *appService) getControlAccountReconciliationJSON(ctx context.Context, companyCode, asOfDate string) (string, error) {
	report, err := s.reportingService.GetControlAccountReconciliation(ctx, companyCode, asOfDate)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(report)
	return string(data), nil
}

// getVendorPaymentHistoryJSON returns payment history for a vendor.
func (s *appService) getVendorPaymentHistoryJSON(ctx context.Context, companyCode, vendorCode string) (string, error) {
	var baseCurrency string
	if err := s.pool.QueryRow(ctx,
		"SELECT base_currency FROM companies WHERE company_code = $1",
		companyCode,
	).Scan(&baseCurrency); err != nil {
		return "", fmt.Errorf("resolve company base currency: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT po.id, po.po_number, po.invoice_number, po.invoice_date::text,
		       po.currency, po.exchange_rate,
		       po.invoice_amount, po.total_transaction, po.total_base,
		       po.paid_at, po.pi_document_number
		FROM purchase_orders po
		JOIN vendors v ON v.id = po.vendor_id
		JOIN companies c ON c.id = po.company_id
		WHERE c.company_code = $1 AND v.code = $2 AND po.status = 'PAID'
		ORDER BY po.paid_at DESC`,
		companyCode, vendorCode,
	)
	if err != nil {
		return "", fmt.Errorf("get vendor payment history: %w", err)
	}
	defer rows.Close()

	type paymentRecord struct {
		POID                     int     `json:"po_id"`
		PONumber                 *string `json:"po_number"`
		InvoiceNumber            *string `json:"invoice_number"`
		InvoiceDate              *string `json:"invoice_date"`
		Currency                 string  `json:"currency"`
		ExchangeRate             string  `json:"exchange_rate"`
		InvoiceAmountTransaction *string `json:"invoice_amount_transaction"`
		InvoiceAmountBase        *string `json:"invoice_amount_base"`
		POTotalTransaction       string  `json:"po_total_transaction"`
		POTotalBase              string  `json:"po_total_base"`
		PaidAt                   *string `json:"paid_at"`
		PIDocumentNumber         *string `json:"pi_document_number"`
	}

	var payments []paymentRecord
	for rows.Next() {
		var pr paymentRecord
		var currency string
		var exchangeRate decimal.Decimal
		var totalTransaction decimal.Decimal
		var totalBase decimal.Decimal
		var invoiceAmount *decimal.Decimal
		var paidAt *time.Time
		if err := rows.Scan(
			&pr.POID, &pr.PONumber, &pr.InvoiceNumber, &pr.InvoiceDate,
			&currency, &exchangeRate,
			&invoiceAmount, &totalTransaction, &totalBase,
			&paidAt, &pr.PIDocumentNumber,
		); err != nil {
			return "", fmt.Errorf("scan payment record: %w", err)
		}

		pr.Currency = currency
		pr.ExchangeRate = exchangeRate.String()
		pr.POTotalTransaction = totalTransaction.StringFixed(2)
		pr.POTotalBase = totalBase.StringFixed(2)
		if invoiceAmount != nil {
			txAmount := invoiceAmount.StringFixed(2)
			pr.InvoiceAmountTransaction = &txAmount
			baseAmount := invoiceAmount.Mul(exchangeRate).StringFixed(2)
			pr.InvoiceAmountBase = &baseAmount
		}
		if paidAt != nil {
			s := paidAt.Format("2006-01-02")
			pr.PaidAt = &s
		}
		payments = append(payments, pr)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate payment history: %w", err)
	}
	if payments == nil {
		payments = []paymentRecord{}
	}

	result := map[string]any{
		"vendor_code":     vendorCode,
		"base_currency":   baseCurrency,
		"payment_count":   len(payments),
		"payment_history": payments,
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

// CommitProposal validates and posts an AI-generated proposal to the ledger.
func (s *appService) CommitProposal(ctx context.Context, proposal core.Proposal) error {
	if err := s.enforceDocumentTypePolicy(ctx, proposal); err != nil {
		return err
	}
	if err := s.enforceSharedJEControlAccountPolicy(ctx, proposal); err != nil {
		return err
	}
	return s.ledger.Commit(ctx, proposal)
}

// ValidateProposal validates a proposal without committing it.
func (s *appService) ValidateProposal(ctx context.Context, proposal core.Proposal) error {
	if err := s.enforceDocumentTypePolicy(ctx, proposal); err != nil {
		return err
	}
	if err := s.enforceSharedJEControlAccountPolicy(ctx, proposal); err != nil {
		return err
	}
	return s.ledger.Validate(ctx, proposal)
}

func (s *appService) enforceSharedJEControlAccountPolicy(ctx context.Context, proposal core.Proposal) error {
	docTypeCode := strings.ToUpper(strings.TrimSpace(proposal.DocumentTypeCode))
	if docTypeCode != "JE" {
		return nil
	}

	mode := strings.ToLower(strings.TrimSpace(os.Getenv("CONTROL_ACCOUNT_ENFORCEMENT_MODE")))
	if mode == "" || mode == "off" || mode == "warn" {
		return nil
	}
	if mode != "enforce" {
		return nil
	}

	controlWarnings, err := s.GetManualJEControlAccountWarnings(ctx, proposal.CompanyCode, uniqueProposalAccountCodes(proposal))
	if err != nil {
		return err
	}
	if len(controlWarnings) == 0 {
		return nil
	}

	// Low-risk rollout: enforce in shared path only for explicit manual web calls.
	if proposalSourceFromContext(ctx) != ProposalSourceManualWeb {
		return nil
	}

	override := controlOverrideFromContext(ctx)
	if !override.Enabled {
		return fmt.Errorf("CONTROL_ACCOUNT_ENFORCED: control account detected; manual JE is blocked in enforce mode unless admin override is provided")
	}
	if strings.TrimSpace(override.Reason) == "" {
		return fmt.Errorf("CONTROL_ACCOUNT_OVERRIDE_REASON_REQUIRED: override_reason is required when override_control_accounts is true")
	}
	if strings.TrimSpace(override.Role) != "ADMIN" {
		return fmt.Errorf("FORBIDDEN: only ADMIN can override control-account enforcement")
	}
	return nil
}

func uniqueProposalAccountCodes(proposal core.Proposal) []string {
	seen := make(map[string]struct{}, len(proposal.Lines))
	codes := make([]string, 0, len(proposal.Lines))
	for _, l := range proposal.Lines {
		code := strings.TrimSpace(l.AccountCode)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	return codes
}

// LoadDefaultCompany loads the active company, using COMPANY_CODE env var if set.
func (s *appService) LoadDefaultCompany(ctx context.Context) (*core.Company, error) {
	if code := os.Getenv("COMPANY_CODE"); code != "" {
		c := &core.Company{}
		err := s.pool.QueryRow(ctx,
			"SELECT id, company_code, name, base_currency FROM companies WHERE company_code = $1", code,
		).Scan(&c.ID, &c.CompanyCode, &c.Name, &c.BaseCurrency)
		if err != nil {
			return nil, fmt.Errorf("company %s not found: %w", code, err)
		}
		return c, nil
	}

	var count int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM companies").Scan(&count); err != nil {
		return nil, fmt.Errorf("failed to count companies: %w", err)
	}
	if count > 1 {
		return nil, fmt.Errorf("multiple companies found; set COMPANY_CODE env var (e.g. COMPANY_CODE=1000)")
	}

	c := &core.Company{}
	if err := s.pool.QueryRow(ctx,
		"SELECT id, company_code, name, base_currency FROM companies LIMIT 1",
	).Scan(&c.ID, &c.CompanyCode, &c.Name, &c.BaseCurrency); err != nil {
		return nil, fmt.Errorf("no default company found, have migrations run?: %w", err)
	}
	return c, nil
}

// GetCompanyByCode loads a company by company_code.
func (s *appService) GetCompanyByCode(ctx context.Context, companyCode string) (*core.Company, error) {
	return s.fetchCompany(ctx, companyCode)
}

// ── private helpers ───────────────────────────────────────────────────────────

// resolveOrder looks up a sales order by numeric ID or order number string.
func (s *appService) resolveOrder(ctx context.Context, ref, companyCode string) (*core.SalesOrder, error) {
	if id, err := strconv.Atoi(ref); err == nil {
		return s.orderService.GetOrder(ctx, id)
	}
	return s.orderService.GetOrderByNumber(ctx, companyCode, ref)
}

// fetchCompany retrieves a company record by code.
func (s *appService) fetchCompany(ctx context.Context, companyCode string) (*core.Company, error) {
	c := &core.Company{}
	if err := s.pool.QueryRow(ctx,
		"SELECT id, company_code, name, base_currency FROM companies WHERE company_code = $1", companyCode,
	).Scan(&c.ID, &c.CompanyCode, &c.Name, &c.BaseCurrency); err != nil {
		return nil, fmt.Errorf("company %s not found: %w", companyCode, err)
	}
	return c, nil
}

// fetchCoA returns the chart of accounts for a company as a formatted string for the AI prompt.
func (s *appService) fetchCoA(ctx context.Context, companyCode string) (string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.code, a.name, a.type
		FROM accounts a
		JOIN companies c ON c.id = a.company_id
		WHERE c.company_code = $1
		ORDER BY a.code
	`, companyCode)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var code, name, accType string
		if err := rows.Scan(&code, &name, &accType); err != nil {
			return "", err
		}
		lines = append(lines, fmt.Sprintf("- %s %s (%s)", code, name, accType))
	}
	return strings.Join(lines, "\n"), nil
}

// fetchDocumentTypes returns all document types as a formatted string for the AI prompt.
func (s *appService) fetchDocumentTypes(ctx context.Context) (string, error) {
	rows, err := s.pool.Query(ctx, "SELECT code, name FROM document_types")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			return "", err
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", code, name))
	}
	return strings.Join(lines, "\n"), nil
}

// ── read tool handlers (Phase 7.5) ───────────────────────────────────────────

// searchAccounts queries accounts by name or code using ILIKE similarity and returns JSON.
func (s *appService) searchAccounts(ctx context.Context, companyCode, query string) (string, error) {
	const minTokenLen = 2
	stopWords := map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "balance": {}, "for": {}, "from": {},
		"get": {}, "is": {}, "me": {}, "of": {}, "please": {}, "show": {},
		"the": {}, "what": {},
	}

	tokenRe := regexp.MustCompile(`[a-z0-9]+`)
	rawTokens := tokenRe.FindAllString(strings.ToLower(strings.TrimSpace(query)), -1)
	seen := make(map[string]struct{}, len(rawTokens))
	tokens := make([]string, 0, len(rawTokens))
	for _, tok := range rawTokens {
		if len(tok) <= minTokenLen {
			continue
		}
		if _, isStopWord := stopWords[tok]; isStopWord {
			continue
		}
		if _, exists := seen[tok]; exists {
			continue
		}
		seen[tok] = struct{}{}
		tokens = append(tokens, tok)
	}
	if len(tokens) == 0 && strings.TrimSpace(query) != "" {
		tokens = append(tokens, strings.TrimSpace(query))
	}

	var sb strings.Builder
	sb.WriteString(`
		SELECT a.code, a.name, a.type
		FROM accounts a
		JOIN companies c ON c.id = a.company_id
		WHERE c.company_code = $1
		  AND (
		        a.name ILIKE '%' || $2 || '%'
		     OR a.code ILIKE '%' || $2 || '%'
	`)
	args := []any{companyCode, strings.TrimSpace(query)}
	for _, tok := range tokens {
		args = append(args, tok)
		p := len(args)
		sb.WriteString(fmt.Sprintf(`
		     OR a.name ILIKE '%%' || $%d || '%%'
		     OR a.code ILIKE '%%' || $%d || '%%'
		`, p, p))
	}
	sb.WriteString(`
		  )
		ORDER BY a.code
		LIMIT 10
	`)

	rows, err := s.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	var results []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Code, &r.Name, &r.Type); err != nil {
			return "", err
		}
		results = append(results, r)
	}
	if len(results) == 0 {
		return `{"accounts":[],"note":"No accounts matched the query."}`, nil
	}
	data, _ := json.Marshal(map[string]any{"accounts": results})
	return string(data), nil
}

// searchCustomers queries customers by name or code using ILIKE and returns JSON.
func (s *appService) searchCustomers(ctx context.Context, companyCode, query string) (string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cu.code, cu.name, cu.email
		FROM customers cu
		JOIN companies c ON c.id = cu.company_id
		WHERE c.company_code = $1
		  AND (cu.name ILIKE '%' || $2 || '%' OR cu.code ILIKE '%' || $2 || '%')
		ORDER BY cu.code
		LIMIT 10
	`, companyCode, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		Code  string `json:"code"`
		Name  string `json:"name"`
		Email string `json:"email,omitempty"`
	}
	var results []row
	for rows.Next() {
		var r row
		var email *string
		if err := rows.Scan(&r.Code, &r.Name, &email); err != nil {
			return "", err
		}
		if email != nil {
			r.Email = *email
		}
		results = append(results, r)
	}
	if len(results) == 0 {
		return `{"customers":[],"note":"No customers matched the query."}`, nil
	}
	data, _ := json.Marshal(map[string]any{"customers": results})
	return string(data), nil
}

// searchProducts queries products by name or code using ILIKE and returns JSON.
func (s *appService) searchProducts(ctx context.Context, companyCode, query string) (string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.code, p.name, p.unit_price
		FROM products p
		JOIN companies c ON c.id = p.company_id
		WHERE c.company_code = $1
		  AND (p.name ILIKE '%' || $2 || '%' OR p.code ILIKE '%' || $2 || '%')
		ORDER BY p.code
		LIMIT 10
	`, companyCode, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		Code      string `json:"code"`
		Name      string `json:"name"`
		UnitPrice string `json:"unit_price"`
	}
	var results []row
	for rows.Next() {
		var r row
		var unitPrice decimal.Decimal
		if err := rows.Scan(&r.Code, &r.Name, &unitPrice); err != nil {
			return "", err
		}
		r.UnitPrice = unitPrice.String()
		results = append(results, r)
	}
	if len(results) == 0 {
		return `{"products":[],"note":"No products matched the query."}`, nil
	}
	data, _ := json.Marshal(map[string]any{"products": results})
	return string(data), nil
}

// getStockLevels returns current inventory stock levels, optionally filtered, as JSON.
func (s *appService) getStockLevels(ctx context.Context, companyCode, productCode, warehouseCode string) (string, error) {
	q := `
		SELECT p.code, p.name, w.code AS warehouse_code, w.name AS warehouse_name,
		       ii.qty_on_hand, ii.qty_reserved
		FROM inventory_items ii
		JOIN products p ON p.id = ii.product_id
		JOIN warehouses w ON w.id = ii.warehouse_id
		JOIN companies c ON c.id = p.company_id
		WHERE c.company_code = $1
	`
	args := []any{companyCode}
	if productCode != "" {
		args = append(args, productCode)
		q += fmt.Sprintf(" AND p.code = $%d", len(args))
	}
	if warehouseCode != "" {
		args = append(args, warehouseCode)
		q += fmt.Sprintf(" AND w.code = $%d", len(args))
	}
	q += " ORDER BY p.code, w.code"

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		ProductCode   string `json:"product_code"`
		ProductName   string `json:"product_name"`
		WarehouseCode string `json:"warehouse_code"`
		WarehouseName string `json:"warehouse_name"`
		QtyOnHand     string `json:"qty_on_hand"`
		QtyReserved   string `json:"qty_reserved"`
		QtyAvailable  string `json:"qty_available"`
	}
	var results []row
	for rows.Next() {
		var r row
		var onHand, reserved decimal.Decimal
		if err := rows.Scan(&r.ProductCode, &r.ProductName, &r.WarehouseCode, &r.WarehouseName, &onHand, &reserved); err != nil {
			return "", err
		}
		r.QtyOnHand = onHand.String()
		r.QtyReserved = reserved.String()
		r.QtyAvailable = onHand.Sub(reserved).String()
		results = append(results, r)
	}
	if len(results) == 0 {
		return `{"stock_levels":[],"note":"No inventory records found for the given filters."}`, nil
	}
	data, _ := json.Marshal(map[string]any{"stock_levels": results})
	return string(data), nil
}

// getAccountBalanceJSON returns the current net-debit balance for a single account as JSON.
func (s *appService) getAccountBalanceJSON(ctx context.Context, companyCode, accountCode string) (string, error) {
	var accountName string
	var netDebit decimal.Decimal
	err := s.pool.QueryRow(ctx, `
		SELECT a.name,
		       COALESCE(SUM(jl.debit_base), 0) - COALESCE(SUM(jl.credit_base), 0)
		FROM accounts a
		JOIN companies c ON c.id = a.company_id
		LEFT JOIN journal_lines jl ON jl.account_id = a.id
		WHERE c.company_code = $1 AND a.code = $2
		GROUP BY a.name
	`, companyCode, accountCode).Scan(&accountName, &netDebit)
	if err != nil {
		return fmt.Sprintf(`{"error":"account %s not found or no activity"}`, accountCode), nil
	}
	data, _ := json.Marshal(map[string]any{
		"account_code": accountCode,
		"account_name": accountName,
		"balance":      netDebit.StringFixed(2),
		"note":         "Net debit position: positive = net debit, negative = net credit",
	})
	return string(data), nil
}

// ── vendor tool helpers (Phase 11) ───────────────────────────────────────────

// getVendorsJSON returns all active vendors for the company as JSON.
func (s *appService) getVendorsJSON(ctx context.Context, companyCode string) (string, error) {
	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return "", err
	}
	vendors, err := s.vendorService.GetVendors(ctx, company.ID)
	if err != nil {
		return "", err
	}
	if len(vendors) == 0 {
		return `{"vendors":[],"note":"No active vendors found."}`, nil
	}
	data, _ := json.Marshal(map[string]any{"vendors": vendorsToJSON(vendors)})
	return string(data), nil
}

// searchVendors queries vendors by name (trigram similarity via GIN index from migration 021)
// or by code (ILIKE prefix/contains). Results ordered by name similarity then code.
func (s *appService) searchVendors(ctx context.Context, companyCode, query string) (string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.code, v.name, v.email, v.phone, v.payment_terms_days, v.ap_account_code
		FROM vendors v
		JOIN companies c ON c.id = v.company_id
		WHERE c.company_code = $1
		  AND v.is_active = true
		  AND (v.name % $2 OR v.code ILIKE '%' || $2 || '%')
		ORDER BY similarity(v.name, $2) DESC, v.code
		LIMIT 10
	`, companyCode, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		Code             string  `json:"code"`
		Name             string  `json:"name"`
		Email            *string `json:"email,omitempty"`
		Phone            *string `json:"phone,omitempty"`
		PaymentTermsDays int     `json:"payment_terms_days"`
		APAccountCode    string  `json:"ap_account_code"`
	}
	var results []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Code, &r.Name, &r.Email, &r.Phone, &r.PaymentTermsDays, &r.APAccountCode); err != nil {
			return "", err
		}
		results = append(results, r)
	}
	if len(results) == 0 {
		return `{"vendors":[],"note":"No vendors matched the query."}`, nil
	}
	data, _ := json.Marshal(map[string]any{"vendors": results})
	return string(data), nil
}

// getVendorInfoJSON returns full vendor details by code as JSON.
func (s *appService) getVendorInfoJSON(ctx context.Context, companyCode, vendorCode string) (string, error) {
	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return "", err
	}
	v, err := s.vendorService.GetVendorByCode(ctx, company.ID, vendorCode)
	if err != nil {
		return fmt.Sprintf(`{"error":"vendor %q not found"}`, vendorCode), nil
	}
	type out struct {
		Code                      string  `json:"code"`
		Name                      string  `json:"name"`
		ContactPerson             *string `json:"contact_person,omitempty"`
		Email                     *string `json:"email,omitempty"`
		Phone                     *string `json:"phone,omitempty"`
		Address                   *string `json:"address,omitempty"`
		PaymentTermsDays          int     `json:"payment_terms_days"`
		APAccountCode             string  `json:"ap_account_code"`
		DefaultExpenseAccountCode *string `json:"default_expense_account_code,omitempty"`
		IsActive                  bool    `json:"is_active"`
	}
	data, _ := json.Marshal(out{
		Code:                      v.Code,
		Name:                      v.Name,
		ContactPerson:             v.ContactPerson,
		Email:                     v.Email,
		Phone:                     v.Phone,
		Address:                   v.Address,
		PaymentTermsDays:          v.PaymentTermsDays,
		APAccountCode:             v.APAccountCode,
		DefaultExpenseAccountCode: v.DefaultExpenseAccountCode,
		IsActive:                  v.IsActive,
	})
	return string(data), nil
}

// vendorsToJSON converts a slice of Vendor to a JSON-friendly format.
func vendorsToJSON(vendors []core.Vendor) []map[string]any {
	out := make([]map[string]any, len(vendors))
	for i, v := range vendors {
		m := map[string]any{
			"code":               v.Code,
			"name":               v.Name,
			"payment_terms_days": v.PaymentTermsDays,
			"ap_account_code":    v.APAccountCode,
		}
		if v.Email != nil {
			m["email"] = *v.Email
		}
		if v.Phone != nil {
			m["phone"] = *v.Phone
		}
		if v.ContactPerson != nil {
			m["contact_person"] = *v.ContactPerson
		}
		out[i] = m
	}
	return out
}

// getPurchaseOrdersJSON returns purchase orders for the company as JSON, optionally filtered by status.
func (s *appService) getPurchaseOrdersJSON(ctx context.Context, companyCode, status string) (string, error) {
	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return "", err
	}
	orders, err := s.purchaseOrderService.GetPOs(ctx, company.ID, status)
	if err != nil {
		return "", err
	}
	if len(orders) == 0 {
		return `{"purchase_orders":[],"note":"No purchase orders found."}`, nil
	}
	data, _ := json.Marshal(map[string]any{"purchase_orders": purchaseOrdersToJSON(orders)})
	return string(data), nil
}

// getOpenPOsJSON returns DRAFT and APPROVED purchase orders for the company as JSON.
func (s *appService) getOpenPOsJSON(ctx context.Context, companyCode string) (string, error) {
	company, err := s.fetchCompany(ctx, companyCode)
	if err != nil {
		return "", err
	}

	var allOpen []core.PurchaseOrder
	for _, st := range []string{"DRAFT", "APPROVED"} {
		orders, err := s.purchaseOrderService.GetPOs(ctx, company.ID, st)
		if err != nil {
			return "", err
		}
		allOpen = append(allOpen, orders...)
	}

	if len(allOpen) == 0 {
		return `{"purchase_orders":[],"note":"No open purchase orders found."}`, nil
	}
	data, _ := json.Marshal(map[string]any{"purchase_orders": purchaseOrdersToJSON(allOpen)})
	return string(data), nil
}

// purchaseOrdersToJSON converts a slice of PurchaseOrder to a JSON-friendly format.
func purchaseOrdersToJSON(orders []core.PurchaseOrder) []map[string]any {
	out := make([]map[string]any, len(orders))
	for i, po := range orders {
		m := map[string]any{
			"id":                po.ID,
			"vendor_code":       po.VendorCode,
			"vendor_name":       po.VendorName,
			"status":            po.Status,
			"po_date":           po.PODate,
			"currency":          po.Currency,
			"total_transaction": po.TotalTransaction.String(),
			"total_base":        po.TotalBase.String(),
		}
		if po.PONumber != nil {
			m["po_number"] = *po.PONumber
		}
		if po.ExpectedDeliveryDate != nil {
			m["expected_delivery_date"] = *po.ExpectedDeliveryDate
		}
		if po.Notes != nil {
			m["notes"] = *po.Notes
		}
		out[i] = m
	}
	return out
}

// getWarehousesJSON returns all active warehouses for the company as JSON.
func (s *appService) getWarehousesJSON(ctx context.Context, companyCode string) (string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT w.code, w.name, w.location
		FROM warehouses w
		JOIN companies c ON c.id = w.company_id
		WHERE c.company_code = $1
		  AND w.is_active = true
		ORDER BY w.code
	`, companyCode)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Location *string `json:"location,omitempty"`
	}
	var results []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Code, &r.Name, &r.Location); err != nil {
			return "", err
		}
		results = append(results, r)
	}
	if len(results) == 0 {
		return `{"warehouses":[],"note":"No active warehouses found."}`, nil
	}
	data, _ := json.Marshal(map[string]any{"warehouses": results})
	return string(data), nil
}

// AuthenticateUser verifies credentials and returns a session on success.
func (s *appService) AuthenticateUser(ctx context.Context, username, password string) (*UserSession, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id, u.company_id, u.username, u.password_hash, u.role, u.is_active, c.company_code
		FROM users u
		INNER JOIN companies c ON c.id = u.company_id
		WHERE u.username = $1 AND u.is_active = true
		ORDER BY u.id
	`, username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	defer rows.Close()

	var matches []UserSession
	for rows.Next() {
		var (
			userID       int
			companyID    int
			dbUsername   string
			passwordHash string
			role         string
			isActive     bool
			companyCode  string
		)
		if err := rows.Scan(&userID, &companyID, &dbUsername, &passwordHash, &role, &isActive, &companyCode); err != nil {
			return nil, fmt.Errorf("invalid credentials")
		}

		// Compare against every active user with this username.
		// This avoids tenant ambiguity from SELECT ... LIMIT 1.
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err == nil {
			matches = append(matches, UserSession{
				UserID:      userID,
				Username:    dbUsername,
				Role:        role,
				CompanyCode: companyCode,
				CompanyID:   companyID,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid credentials")
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous credentials: multiple active accounts matched this username")
	}

	return &matches[0], nil
}

// GetUser returns user profile by ID, including company code.
func (s *appService) GetUser(ctx context.Context, userID int) (*UserResult, error) {
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var companyCode string
	if err := s.pool.QueryRow(ctx,
		"SELECT company_code FROM companies WHERE id = $1", user.CompanyID,
	).Scan(&companyCode); err != nil {
		return nil, fmt.Errorf("company not found for user")
	}

	return &UserResult{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CompanyCode: companyCode,
	}, nil
}

func (s *appService) ListUsers(ctx context.Context, companyCode string) (*UsersResult, error) {
	var companyID int
	if err := s.pool.QueryRow(ctx,
		"SELECT id FROM companies WHERE company_code = $1", companyCode,
	).Scan(&companyID); err != nil {
		return nil, fmt.Errorf("company %s not found: %w", companyCode, err)
	}
	users, err := s.userService.ListUsers(ctx, companyID)
	if err != nil {
		return nil, err
	}
	result := &UsersResult{Users: make([]UserResult, 0, len(users))}
	for _, u := range users {
		result.Users = append(result.Users, UserResult{
			UserID:      u.ID,
			Username:    u.Username,
			Email:       u.Email,
			Role:        u.Role,
			IsActive:    u.IsActive,
			CompanyCode: companyCode,
		})
	}
	return result, nil
}

func (s *appService) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResult, error) {
	var companyID int
	if err := s.pool.QueryRow(ctx,
		"SELECT id FROM companies WHERE company_code = $1", req.CompanyCode,
	).Scan(&companyID); err != nil {
		return nil, fmt.Errorf("company %s not found: %w", req.CompanyCode, err)
	}
	user, err := s.userService.CreateUser(ctx, companyID, core.CreateUserParams{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		return nil, err
	}
	return &UserResult{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CompanyCode: req.CompanyCode,
	}, nil
}

// RegisterCompany creates a new tenant company and its first ADMIN user in a single
// atomic transaction. Returns a UserSession ready to be signed into a JWT.
func (s *appService) RegisterCompany(ctx context.Context, req RegisterCompanyRequest) (*UserSession, error) {
	// Validate password policy: 8+ chars, at least one uppercase, one digit.
	if err := validateRegistrationPassword(req.Password); err != nil {
		return nil, err
	}

	// Hash the password before opening the transaction.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Generate a unique 4-char company code derived from the company name.
	code, err := s.generateUniqueCompanyCode(ctx, req.CompanyName)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var companyID int
	err = tx.QueryRow(ctx, `
		INSERT INTO companies (company_code, name, base_currency)
		VALUES ($1, $2, 'INR')
		RETURNING id`,
		code, req.CompanyName,
	).Scan(&companyID)
	if err != nil {
		if strings.Contains(err.Error(), "companies_name_unique") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("company name %q is already registered", req.CompanyName)
		}
		return nil, fmt.Errorf("create company: %w", err)
	}

	// Seed baseline CoA + core account rules for the new company so all runtime
	// flows (orders, inventory, AP/AR reporting, AI read tools) work immediately.
	if err := s.seedDefaultChartOfAccountsTx(ctx, tx, companyID); err != nil {
		return nil, fmt.Errorf("seed chart of accounts: %w", err)
	}
	if err := s.seedDefaultAccountRulesTx(ctx, tx, companyID); err != nil {
		return nil, fmt.Errorf("seed account rules: %w", err)
	}

	var userID int
	err = tx.QueryRow(ctx, `
		INSERT INTO users (company_id, username, email, password_hash, role)
		VALUES ($1, $2, $3, $4, 'ADMIN')
		RETURNING id`,
		companyID, req.Username, req.Email, string(hash),
	).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("username %q is already taken", req.Username)
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit registration: %w", err)
	}

	return &UserSession{
		UserID:      userID,
		Username:    req.Username,
		Role:        "ADMIN",
		CompanyCode: code,
		CompanyID:   companyID,
	}, nil
}

// generateUniqueCompanyCode derives a 4-character company code from the given name
// and ensures it does not collide with an existing company_code.
func (s *appService) generateUniqueCompanyCode(ctx context.Context, name string) (string, error) {
	// Build a 4-char base: first 4 uppercase alphanumeric characters.
	var base []byte
	for _, ch := range strings.ToUpper(name) {
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			base = append(base, byte(ch))
			if len(base) == 4 {
				break
			}
		}
	}
	for len(base) < 4 {
		base = append(base, 'X')
	}

	for i := 0; i <= 20; i++ {
		var code string
		if i == 0 {
			code = string(base)
		} else {
			suffix := strconv.Itoa(i + 1)
			code = string(base[:4-len(suffix)]) + suffix
		}
		var exists bool
		if err := s.pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM companies WHERE company_code = $1)", code,
		).Scan(&exists); err != nil {
			return "", fmt.Errorf("check company code: %w", err)
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("could not generate a unique company code for %q — please contact support", name)
}

func (s *appService) seedDefaultChartOfAccountsTx(ctx context.Context, tx pgx.Tx, companyID int) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO accounts (company_id, code, name, type)
		SELECT $1, a.code, a.name, a.type
		FROM (VALUES
			('1000', 'Cash', 'asset'),
			('1100', 'Bank Account', 'asset'),
			('1200', 'Accounts Receivable', 'asset'),
			('1300', 'Furniture & Fixtures', 'asset'),
			('1350', 'ITC - CGST', 'asset'),
			('1360', 'ITC - SGST', 'asset'),
			('1370', 'ITC - IGST', 'asset'),
			('1380', 'ITC - CESS', 'asset'),
			('1400', 'Inventory', 'asset'),
			('2000', 'Accounts Payable', 'liability'),
			('2100', 'Short-Term Loans', 'liability'),
			('2150', 'GST Payable - CGST', 'liability'),
			('2160', 'GST Payable - SGST', 'liability'),
			('2170', 'GST Payable - IGST', 'liability'),
			('2180', 'GST Payable - CESS', 'liability'),
			('2190', 'TDS Payable', 'liability'),
			('2195', 'TCS Payable', 'liability'),
			('3000', 'Owner Capital', 'equity'),
			('3100', 'Retained Earnings', 'equity'),
			('4000', 'Sales Revenue', 'revenue'),
			('4100', 'Service Revenue', 'revenue'),
			('5000', 'Cost of Goods Sold', 'expense'),
			('5100', 'Rent Expense', 'expense'),
			('5200', 'Salary Expense', 'expense'),
			('5300', 'Utilities Expense', 'expense')
		) AS a(code, name, type)
		ON CONFLICT (company_id, code) DO NOTHING
	`, companyID)
	return err
}

func (s *appService) seedDefaultAccountRulesTx(ctx context.Context, tx pgx.Tx, companyID int) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO account_rules (company_id, rule_type, account_code)
		SELECT $1, r.rule_type, r.account_code
		FROM (VALUES
			('AR', '1200'),
			('AP', '2000'),
			('INVENTORY', '1400'),
			('COGS', '5000'),
			('BANK_DEFAULT', '1100'),
			('RECEIPT_CREDIT', '2000')
		) AS r(rule_type, account_code)
		ON CONFLICT DO NOTHING
	`, companyID)
	return err
}

// validateRegistrationPassword enforces the password policy for self-service registration.
func validateRegistrationPassword(p string) error {
	if len(p) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	hasUpper, hasDigit := false, false
	for _, c := range p {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one number")
	}
	return nil
}

func (s *appService) UpdateUserRole(ctx context.Context, req UpdateUserRoleRequest) error {
	var companyID int
	if err := s.pool.QueryRow(ctx,
		"SELECT id FROM companies WHERE company_code = $1", req.CompanyCode,
	).Scan(&companyID); err != nil {
		return fmt.Errorf("company %s not found: %w", req.CompanyCode, err)
	}
	return s.userService.UpdateUserRole(ctx, companyID, req.UserID, req.Role)
}

func (s *appService) SetUserActive(ctx context.Context, companyCode string, userID int, active bool) error {
	var companyID int
	if err := s.pool.QueryRow(ctx,
		"SELECT id FROM companies WHERE company_code = $1", companyCode,
	).Scan(&companyID); err != nil {
		return fmt.Errorf("company %s not found: %w", companyCode, err)
	}
	return s.userService.SetUserActive(ctx, companyID, userID, active)
}

func (s *appService) GetAccountNamesByCode(ctx context.Context, companyCode string, codes []string) (map[string]string, error) {
	var companyID int
	if err := s.pool.QueryRow(ctx,
		"SELECT id FROM companies WHERE company_code = $1", companyCode,
	).Scan(&companyID); err != nil {
		return nil, fmt.Errorf("company %s not found: %w", companyCode, err)
	}
	rows, err := s.pool.Query(ctx,
		`SELECT code, name FROM accounts WHERE company_id = $1 AND code = ANY($2)`,
		companyID, codes,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]string, len(codes))
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			return nil, err
		}
		result[code] = name
	}
	return result, rows.Err()
}
