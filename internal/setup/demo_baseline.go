package setup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/inventoryops"
	"workflow_app/internal/parties"
	"workflow_app/internal/platform/audit"
)

type DemoBaselineInput struct {
	OrgID       string
	ActorUserID string
}

type DemoBaselineResult struct {
	LedgerAccountsCreated     int
	TaxCodesCreated           int
	AccountingPeriodsCreated  int
	PartiesCreated            int
	ContactsCreated           int
	InventoryItemsCreated     int
	InventoryLocationsCreated int
}

type ledgerAccountSeed struct {
	Code                string
	Name                string
	AccountClass        string
	ControlType         string
	AllowsDirectPosting bool
}

type taxCodeSeed struct {
	Code              string
	Name              string
	TaxType           string
	RateBasisPoints   int
	ReceivableAccount string
	PayableAccount    string
}

type partySeed struct {
	Code        string
	DisplayName string
	LegalName   string
	PartyKind   string
	Contact     contactSeed
}

type contactSeed struct {
	FullName  string
	RoleTitle string
	Email     string
	Phone     string
	IsPrimary bool
}

type inventoryItemSeed struct {
	SKU          string
	Name         string
	ItemRole     string
	TrackingMode string
}

type inventoryLocationSeed struct {
	Code         string
	Name         string
	LocationRole string
}

var northHarborLedgerAccounts = []ledgerAccountSeed{
	{Code: "1000", Name: "Cash and Bank", AccountClass: accounting.AccountClassAsset, ControlType: accounting.ControlTypeNone, AllowsDirectPosting: true},
	{Code: "1100", Name: "Accounts Receivable", AccountClass: accounting.AccountClassAsset, ControlType: accounting.ControlTypeReceivable, AllowsDirectPosting: false},
	{Code: "1200", Name: "Inventory Asset", AccountClass: accounting.AccountClassAsset, ControlType: accounting.ControlTypeNone, AllowsDirectPosting: true},
	{Code: "1300", Name: "GST Input Receivable", AccountClass: accounting.AccountClassAsset, ControlType: accounting.ControlTypeGSTInput, AllowsDirectPosting: false},
	{Code: "2000", Name: "Accounts Payable", AccountClass: accounting.AccountClassLiability, ControlType: accounting.ControlTypePayable, AllowsDirectPosting: false},
	{Code: "2100", Name: "GST Output Payable", AccountClass: accounting.AccountClassLiability, ControlType: accounting.ControlTypeGSTOutput, AllowsDirectPosting: false},
	{Code: "3000", Name: "Owner Equity", AccountClass: accounting.AccountClassEquity, ControlType: accounting.ControlTypeNone, AllowsDirectPosting: true},
	{Code: "4000", Name: "Service Revenue", AccountClass: accounting.AccountClassRevenue, ControlType: accounting.ControlTypeNone, AllowsDirectPosting: true},
	{Code: "4100", Name: "Parts and Materials Revenue", AccountClass: accounting.AccountClassRevenue, ControlType: accounting.ControlTypeNone, AllowsDirectPosting: true},
	{Code: "5000", Name: "Cost of Goods Sold", AccountClass: accounting.AccountClassExpense, ControlType: accounting.ControlTypeNone, AllowsDirectPosting: true},
	{Code: "5100", Name: "Subcontractor Expense", AccountClass: accounting.AccountClassExpense, ControlType: accounting.ControlTypeNone, AllowsDirectPosting: true},
	{Code: "5200", Name: "Inventory Adjustments", AccountClass: accounting.AccountClassExpense, ControlType: accounting.ControlTypeNone, AllowsDirectPosting: true},
	{Code: "6000", Name: "Operating Expense", AccountClass: accounting.AccountClassExpense, ControlType: accounting.ControlTypeNone, AllowsDirectPosting: true},
}

var northHarborTaxCodes = []taxCodeSeed{
	{Code: "GST18-SALES", Name: "GST 18% Sales", TaxType: accounting.TaxTypeGST, RateBasisPoints: 1800, PayableAccount: "2100"},
	{Code: "GST18-PURCH", Name: "GST 18% Purchases", TaxType: accounting.TaxTypeGST, RateBasisPoints: 1800, ReceivableAccount: "1300"},
}

var northHarborParties = []partySeed{
	{Code: "CUST-ACME", DisplayName: "Acme Facilities", LegalName: "Acme Facilities Private Limited", PartyKind: parties.PartyKindCustomer, Contact: contactSeed{FullName: "Asha Rao", RoleTitle: "Facilities Manager", Email: "asha.rao@acme.example", Phone: "+91-80-5550-0101", IsPrimary: true}},
	{Code: "CUST-METRO", DisplayName: "Metro Property Group", LegalName: "Metro Property Group", PartyKind: parties.PartyKindCustomer, Contact: contactSeed{FullName: "Karan Mehta", RoleTitle: "Operations Lead", Email: "karan.mehta@metro.example", Phone: "+91-80-5550-0102", IsPrimary: true}},
	{Code: "VEND-HARBOR", DisplayName: "Harbor Industrial Supply", LegalName: "Harbor Industrial Supply LLP", PartyKind: parties.PartyKindVendor, Contact: contactSeed{FullName: "Neha Iyer", RoleTitle: "Account Manager", Email: "neha.iyer@harbor-supply.example", Phone: "+91-80-5550-0201", IsPrimary: true}},
	{Code: "VEND-POWER", DisplayName: "Powerline Electricals", LegalName: "Powerline Electricals", PartyKind: parties.PartyKindVendor, Contact: contactSeed{FullName: "Vikram Singh", RoleTitle: "Sales Desk", Email: "sales@powerline.example", Phone: "+91-80-5550-0202", IsPrimary: true}},
	{Code: "VEND-SUBCO", DisplayName: "Reliable Field Services", LegalName: "Reliable Field Services", PartyKind: parties.PartyKindVendor, Contact: contactSeed{FullName: "Maya D'Souza", RoleTitle: "Dispatch Coordinator", Email: "dispatch@reliable-field.example", Phone: "+91-80-5550-0203", IsPrimary: true}},
}

var northHarborInventoryItems = []inventoryItemSeed{
	{SKU: "SVC-MAT-FILTER", Name: "Replacement filter kit", ItemRole: inventoryops.ItemRoleServiceMaterial, TrackingMode: inventoryops.TrackingModeNone},
	{SKU: "SVC-MAT-SEAL", Name: "Industrial sealant pack", ItemRole: inventoryops.ItemRoleServiceMaterial, TrackingMode: inventoryops.TrackingModeNone},
	{SKU: "RES-PUMP-100", Name: "Pump assembly", ItemRole: inventoryops.ItemRoleResale, TrackingMode: inventoryops.TrackingModeNone},
	{SKU: "EQ-METER-200", Name: "Field meter", ItemRole: inventoryops.ItemRoleTraceableEquipment, TrackingMode: inventoryops.TrackingModeSerial},
	{SKU: "EXP-CLEANUP", Name: "Shop cleanup consumables", ItemRole: inventoryops.ItemRoleDirectExpenseConsumable, TrackingMode: inventoryops.TrackingModeNone},
}

var northHarborInventoryLocations = []inventoryLocationSeed{
	{Code: "MAIN-WH", Name: "Main warehouse", LocationRole: inventoryops.LocationRoleWarehouse},
	{Code: "FIELD-VAN-1", Name: "Field van 1", LocationRole: inventoryops.LocationRoleVan},
	{Code: "ADJ-BIN", Name: "Inventory adjustment bin", LocationRole: inventoryops.LocationRoleAdjustment},
	{Code: "JOB-SITE", Name: "Active job site", LocationRole: inventoryops.LocationRoleSite},
	{Code: "INSTALLED", Name: "Installed equipment base", LocationRole: inventoryops.LocationRoleInstalled},
}

func EnsureDemoBaseline(ctx context.Context, db *sql.DB, input DemoBaselineInput) (DemoBaselineResult, error) {
	if strings.TrimSpace(input.OrgID) == "" || strings.TrimSpace(input.ActorUserID) == "" {
		return DemoBaselineResult{}, fmt.Errorf("org id and actor user id are required")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return DemoBaselineResult{}, fmt.Errorf("begin demo baseline seed: %w", err)
	}

	result, err := ensureDemoBaselineTx(ctx, tx, DemoBaselineInput{
		OrgID:       strings.TrimSpace(input.OrgID),
		ActorUserID: strings.TrimSpace(input.ActorUserID),
	})
	if err != nil {
		_ = tx.Rollback()
		return DemoBaselineResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return DemoBaselineResult{}, fmt.Errorf("commit demo baseline seed: %w", err)
	}

	return result, nil
}

func ensureDemoBaselineTx(ctx context.Context, tx *sql.Tx, input DemoBaselineInput) (DemoBaselineResult, error) {
	var result DemoBaselineResult
	accountIDs := make(map[string]string, len(northHarborLedgerAccounts))

	for _, seed := range northHarborLedgerAccounts {
		id, created, err := ensureLedgerAccount(ctx, tx, input, seed)
		if err != nil {
			return DemoBaselineResult{}, err
		}
		accountIDs[seed.Code] = id
		if created {
			result.LedgerAccountsCreated++
		}
	}

	for _, seed := range northHarborTaxCodes {
		created, err := ensureTaxCode(ctx, tx, input, seed, accountIDs)
		if err != nil {
			return DemoBaselineResult{}, err
		}
		if created {
			result.TaxCodesCreated++
		}
	}

	createdPeriod, err := ensureAccountingPeriod(ctx, tx, input)
	if err != nil {
		return DemoBaselineResult{}, err
	}
	if createdPeriod {
		result.AccountingPeriodsCreated++
	}

	for _, seed := range northHarborParties {
		partyID, created, err := ensureParty(ctx, tx, input, seed)
		if err != nil {
			return DemoBaselineResult{}, err
		}
		if created {
			result.PartiesCreated++
		}
		contactCreated, err := ensureContact(ctx, tx, input, partyID, seed.Contact)
		if err != nil {
			return DemoBaselineResult{}, err
		}
		if contactCreated {
			result.ContactsCreated++
		}
	}

	for _, seed := range northHarborInventoryItems {
		created, err := ensureInventoryItem(ctx, tx, input, seed)
		if err != nil {
			return DemoBaselineResult{}, err
		}
		if created {
			result.InventoryItemsCreated++
		}
	}

	for _, seed := range northHarborInventoryLocations {
		created, err := ensureInventoryLocation(ctx, tx, input, seed)
		if err != nil {
			return DemoBaselineResult{}, err
		}
		if created {
			result.InventoryLocationsCreated++
		}
	}

	return result, nil
}

func ensureLedgerAccount(ctx context.Context, tx *sql.Tx, input DemoBaselineInput, seed ledgerAccountSeed) (string, bool, error) {
	const statement = `
WITH inserted AS (
	INSERT INTO accounting.ledger_accounts (
		org_id,
		code,
		name,
		account_class,
		control_type,
		allows_direct_posting,
		created_by_user_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT id, TRUE FROM inserted
UNION ALL
SELECT id, FALSE
FROM accounting.ledger_accounts
WHERE org_id = $1
  AND lower(code) = lower($2)
  AND NOT EXISTS (SELECT 1 FROM inserted);`

	var id string
	var created bool
	if err := tx.QueryRowContext(ctx, statement, input.OrgID, seed.Code, seed.Name, seed.AccountClass, seed.ControlType, seed.AllowsDirectPosting, input.ActorUserID).Scan(&id, &created); err != nil {
		return "", false, fmt.Errorf("ensure ledger account %s: %w", seed.Code, err)
	}
	if created {
		if err := writeSeedAudit(ctx, tx, input, "accounting.ledger_account_seeded", "accounting.ledger_account", id, map[string]any{"code": seed.Code, "account_class": seed.AccountClass, "control_type": seed.ControlType}); err != nil {
			return "", false, err
		}
	}
	return id, created, nil
}

func ensureTaxCode(ctx context.Context, tx *sql.Tx, input DemoBaselineInput, seed taxCodeSeed, accountIDs map[string]string) (bool, error) {
	receivableID := accountIDs[seed.ReceivableAccount]
	payableID := accountIDs[seed.PayableAccount]
	const statement = `
WITH inserted AS (
	INSERT INTO accounting.tax_codes (
		org_id,
		code,
		name,
		tax_type,
		rate_basis_points,
		receivable_account_id,
		payable_account_id,
		created_by_user_id
	) VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::uuid, NULLIF($7, '')::uuid, $8)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT id, TRUE FROM inserted
UNION ALL
SELECT id, FALSE
FROM accounting.tax_codes
WHERE org_id = $1
  AND lower(code) = lower($2)
  AND NOT EXISTS (SELECT 1 FROM inserted);`

	var id string
	var created bool
	if err := tx.QueryRowContext(ctx, statement, input.OrgID, seed.Code, seed.Name, seed.TaxType, seed.RateBasisPoints, receivableID, payableID, input.ActorUserID).Scan(&id, &created); err != nil {
		return false, fmt.Errorf("ensure tax code %s: %w", seed.Code, err)
	}
	if created {
		if err := writeSeedAudit(ctx, tx, input, "accounting.tax_code_seeded", "accounting.tax_code", id, map[string]any{"code": seed.Code, "tax_type": seed.TaxType, "rate_basis_points": seed.RateBasisPoints}); err != nil {
			return false, err
		}
	}
	return created, nil
}

func ensureAccountingPeriod(ctx context.Context, tx *sql.Tx, input DemoBaselineInput) (bool, error) {
	startOn := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	endOn := time.Date(2027, time.March, 31, 0, 0, 0, 0, time.UTC)
	const statement = `
WITH inserted AS (
	INSERT INTO accounting.periods (
		org_id,
		period_code,
		start_on,
		end_on,
		created_by_user_id
	) VALUES ($1, 'FY2026-27', $2, $3, $4)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT id, TRUE FROM inserted
UNION ALL
SELECT id, FALSE
FROM accounting.periods
WHERE org_id = $1
  AND lower(period_code) = lower('FY2026-27')
  AND NOT EXISTS (SELECT 1 FROM inserted);`

	var id string
	var created bool
	if err := tx.QueryRowContext(ctx, statement, input.OrgID, startOn, endOn, input.ActorUserID).Scan(&id, &created); err != nil {
		return false, fmt.Errorf("ensure accounting period FY2026-27: %w", err)
	}
	if created {
		if err := writeSeedAudit(ctx, tx, input, "accounting.period_seeded", "accounting.period", id, map[string]any{"period_code": "FY2026-27", "start_on": "2026-04-01", "end_on": "2027-03-31"}); err != nil {
			return false, err
		}
	}
	return created, nil
}

func ensureParty(ctx context.Context, tx *sql.Tx, input DemoBaselineInput, seed partySeed) (string, bool, error) {
	const statement = `
WITH inserted AS (
	INSERT INTO parties.parties (
		org_id,
		party_code,
		display_name,
		legal_name,
		party_kind,
		created_by_user_id
	) VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT id, TRUE FROM inserted
UNION ALL
SELECT id, FALSE
FROM parties.parties
WHERE org_id = $1
  AND lower(party_code) = lower($2)
  AND NOT EXISTS (SELECT 1 FROM inserted);`

	var id string
	var created bool
	if err := tx.QueryRowContext(ctx, statement, input.OrgID, seed.Code, seed.DisplayName, seed.LegalName, seed.PartyKind, input.ActorUserID).Scan(&id, &created); err != nil {
		return "", false, fmt.Errorf("ensure party %s: %w", seed.Code, err)
	}
	if created {
		if err := writeSeedAudit(ctx, tx, input, "parties.party_seeded", "parties.party", id, map[string]any{"party_code": seed.Code, "party_kind": seed.PartyKind}); err != nil {
			return "", false, err
		}
	}
	return id, created, nil
}

func ensureContact(ctx context.Context, tx *sql.Tx, input DemoBaselineInput, partyID string, seed contactSeed) (bool, error) {
	var existingID string
	err := tx.QueryRowContext(ctx, `
SELECT id
FROM parties.contacts
WHERE org_id = $1
  AND party_id = $2
  AND (lower(email) = lower($3) OR ($4 AND is_primary))
LIMIT 1;`, input.OrgID, partyID, seed.Email, seed.IsPrimary).Scan(&existingID)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("check contact %s: %w", seed.Email, err)
	}

	var id string
	if err := tx.QueryRowContext(ctx, `
INSERT INTO parties.contacts (
	org_id,
	party_id,
	full_name,
	role_title,
	email,
	phone,
	is_primary,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;`, input.OrgID, partyID, seed.FullName, seed.RoleTitle, seed.Email, seed.Phone, seed.IsPrimary, input.ActorUserID).Scan(&id); err != nil {
		return false, fmt.Errorf("ensure contact %s: %w", seed.Email, err)
	}
	if err := writeSeedAudit(ctx, tx, input, "parties.contact_seeded", "parties.contact", id, map[string]any{"party_id": partyID, "email": seed.Email}); err != nil {
		return false, err
	}
	return true, nil
}

func ensureInventoryItem(ctx context.Context, tx *sql.Tx, input DemoBaselineInput, seed inventoryItemSeed) (bool, error) {
	const statement = `
WITH inserted AS (
	INSERT INTO inventory_ops.items (
		org_id,
		sku,
		name,
		item_role,
		tracking_mode,
		created_by_user_id
	) VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT id, TRUE FROM inserted
UNION ALL
SELECT id, FALSE
FROM inventory_ops.items
WHERE org_id = $1
  AND lower(sku) = lower($2)
  AND NOT EXISTS (SELECT 1 FROM inserted);`

	var id string
	var created bool
	if err := tx.QueryRowContext(ctx, statement, input.OrgID, seed.SKU, seed.Name, seed.ItemRole, seed.TrackingMode, input.ActorUserID).Scan(&id, &created); err != nil {
		return false, fmt.Errorf("ensure inventory item %s: %w", seed.SKU, err)
	}
	if created {
		if err := writeSeedAudit(ctx, tx, input, "inventory_ops.item_seeded", "inventory_ops.item", id, map[string]any{"sku": seed.SKU, "item_role": seed.ItemRole}); err != nil {
			return false, err
		}
	}
	return created, nil
}

func ensureInventoryLocation(ctx context.Context, tx *sql.Tx, input DemoBaselineInput, seed inventoryLocationSeed) (bool, error) {
	const statement = `
WITH inserted AS (
	INSERT INTO inventory_ops.locations (
		org_id,
		code,
		name,
		location_role,
		created_by_user_id
	) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT id, TRUE FROM inserted
UNION ALL
SELECT id, FALSE
FROM inventory_ops.locations
WHERE org_id = $1
  AND lower(code) = lower($2)
  AND NOT EXISTS (SELECT 1 FROM inserted);`

	var id string
	var created bool
	if err := tx.QueryRowContext(ctx, statement, input.OrgID, seed.Code, seed.Name, seed.LocationRole, input.ActorUserID).Scan(&id, &created); err != nil {
		return false, fmt.Errorf("ensure inventory location %s: %w", seed.Code, err)
	}
	if created {
		if err := writeSeedAudit(ctx, tx, input, "inventory_ops.location_seeded", "inventory_ops.location", id, map[string]any{"code": seed.Code, "location_role": seed.LocationRole}); err != nil {
			return false, err
		}
	}
	return created, nil
}

func writeSeedAudit(ctx context.Context, tx *sql.Tx, input DemoBaselineInput, eventType, entityType, entityID string, payload map[string]any) error {
	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.OrgID,
		ActorUserID: input.ActorUserID,
		EventType:   eventType,
		EntityType:  entityType,
		EntityID:    entityID,
		Payload:     payload,
	}); err != nil {
		return fmt.Errorf("write seed audit %s %s: %w", entityType, entityID, err)
	}
	return nil
}
