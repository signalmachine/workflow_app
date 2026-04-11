package setup

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/testsupport/dbtest"
)

func TestEnsureDemoBaselineSeedsNorthHarborMinimumDataIdempotently(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bootstrap, err := identityaccess.NewService(db).BootstrapAdmin(ctx, identityaccess.BootstrapAdminInput{
		OrgName:         "North Harbor Works",
		OrgSlug:         "north-harbor",
		UserEmail:       "admin@northharbor.local",
		UserDisplayName: "North Harbor Admin",
		Password:        "NorthHarbor2026",
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("bootstrap admin: %v", err)
	}

	first, err := EnsureDemoBaseline(ctx, db, DemoBaselineInput{OrgID: bootstrap.OrgID, ActorUserID: bootstrap.UserID})
	if err != nil {
		t.Fatalf("first demo baseline seed: %v", err)
	}
	if first.LedgerAccountsCreated != len(northHarborLedgerAccounts) {
		t.Fatalf("ledger accounts created = %d, want %d", first.LedgerAccountsCreated, len(northHarborLedgerAccounts))
	}
	if first.TaxCodesCreated != len(northHarborTaxCodes) {
		t.Fatalf("tax codes created = %d, want %d", first.TaxCodesCreated, len(northHarborTaxCodes))
	}
	if first.AccountingPeriodsCreated != 1 {
		t.Fatalf("accounting periods created = %d, want 1", first.AccountingPeriodsCreated)
	}
	if first.PartiesCreated != len(northHarborParties) || first.ContactsCreated != len(northHarborParties) {
		t.Fatalf("party/contact created counts = %d/%d, want %d/%d", first.PartiesCreated, first.ContactsCreated, len(northHarborParties), len(northHarborParties))
	}
	if first.InventoryItemsCreated != len(northHarborInventoryItems) || first.InventoryLocationsCreated != len(northHarborInventoryLocations) {
		t.Fatalf("inventory created counts = %d/%d, want %d/%d", first.InventoryItemsCreated, first.InventoryLocationsCreated, len(northHarborInventoryItems), len(northHarborInventoryLocations))
	}

	second, err := EnsureDemoBaseline(ctx, db, DemoBaselineInput{OrgID: bootstrap.OrgID, ActorUserID: bootstrap.UserID})
	if err != nil {
		t.Fatalf("second demo baseline seed: %v", err)
	}
	if second != (DemoBaselineResult{}) {
		t.Fatalf("second seed created records: %+v", second)
	}

	assertCount(t, ctx, db, `SELECT COUNT(*) FROM accounting.ledger_accounts WHERE org_id = $1`, bootstrap.OrgID, len(northHarborLedgerAccounts))
	assertCount(t, ctx, db, `SELECT COUNT(*) FROM accounting.tax_codes WHERE org_id = $1`, bootstrap.OrgID, len(northHarborTaxCodes))
	assertCount(t, ctx, db, `SELECT COUNT(*) FROM accounting.periods WHERE org_id = $1 AND period_code = 'FY2026-27'`, bootstrap.OrgID, 1)
	assertCount(t, ctx, db, `SELECT COUNT(*) FROM parties.parties WHERE org_id = $1`, bootstrap.OrgID, len(northHarborParties))
	assertCount(t, ctx, db, `SELECT COUNT(*) FROM parties.contacts WHERE org_id = $1`, bootstrap.OrgID, len(northHarborParties))
	assertCount(t, ctx, db, `SELECT COUNT(*) FROM inventory_ops.items WHERE org_id = $1`, bootstrap.OrgID, len(northHarborInventoryItems))
	assertCount(t, ctx, db, `SELECT COUNT(*) FROM inventory_ops.locations WHERE org_id = $1`, bootstrap.OrgID, len(northHarborInventoryLocations))
}

func assertCount(t *testing.T, ctx context.Context, db *sql.DB, query string, orgID string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRowContext(ctx, query, orgID).Scan(&got); err != nil {
		t.Fatalf("count query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("count query %q = %d, want %d", query, got, want)
	}
}
