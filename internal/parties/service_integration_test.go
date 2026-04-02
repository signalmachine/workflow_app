package parties_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/parties"
	"workflow_app/internal/testsupport/dbtest"
)

func TestCreatePartyAndPrimaryContactIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	service := parties.NewService(db)

	party, err := service.CreateParty(ctx, parties.CreatePartyInput{
		PartyCode:   "CUST-001",
		DisplayName: "Northwind Service Customer",
		LegalName:   "Northwind Service Customer Pvt Ltd",
		PartyKind:   parties.PartyKindCustomer,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("create party: %v", err)
	}
	if party.Status != parties.StatusActive {
		t.Fatalf("unexpected party status: %s", party.Status)
	}

	firstContact, err := service.CreateContact(ctx, parties.CreateContactInput{
		PartyID:   party.ID,
		FullName:  "Asha Nair",
		RoleTitle: "Accounts",
		Email:     "asha@example.com",
		IsPrimary: true,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("create first contact: %v", err)
	}
	if !firstContact.IsPrimary {
		t.Fatalf("expected first contact to be primary")
	}

	secondContact, err := service.CreateContact(ctx, parties.CreateContactInput{
		PartyID:   party.ID,
		FullName:  "Rahul Mehta",
		Phone:     "+1-555-2000",
		IsPrimary: true,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("create second contact: %v", err)
	}
	if !secondContact.IsPrimary {
		t.Fatalf("expected second contact to be primary")
	}

	contacts, err := service.ListContacts(ctx, parties.ListContactsInput{
		PartyID: party.ID,
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("list contacts: %v", err)
	}
	if len(contacts) != 2 {
		t.Fatalf("unexpected contact count: %d", len(contacts))
	}
	if contacts[0].ID != secondContact.ID || !contacts[0].IsPrimary {
		t.Fatalf("expected latest primary contact first, got %+v", contacts[0])
	}

	var firstStillPrimary bool
	if err := db.QueryRowContext(ctx, `SELECT is_primary FROM parties.contacts WHERE id = $1`, firstContact.ID).Scan(&firstStillPrimary); err != nil {
		t.Fatalf("load first contact primary flag: %v", err)
	}
	if firstStillPrimary {
		t.Fatalf("expected previous primary contact to be cleared")
	}
}

func TestListPartiesFiltersByKindIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	service := parties.NewService(db)

	for _, input := range []parties.CreatePartyInput{
		{PartyCode: "CUST-100", DisplayName: "Customer One", PartyKind: parties.PartyKindCustomer, Actor: operator},
		{PartyCode: "VEND-100", DisplayName: "Vendor One", PartyKind: parties.PartyKindVendor, Actor: operator},
		{PartyCode: "BOTH-100", DisplayName: "Hybrid Counterparty", PartyKind: parties.PartyKindCustomerVendor, Actor: operator},
	} {
		if _, err := service.CreateParty(ctx, input); err != nil {
			t.Fatalf("create party %s: %v", input.PartyCode, err)
		}
	}

	customers, err := service.ListParties(ctx, parties.ListPartiesInput{
		PartyKind: parties.PartyKindCustomer,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("list customer parties: %v", err)
	}
	if len(customers) != 1 {
		t.Fatalf("unexpected customer party count: %d", len(customers))
	}
	if customers[0].PartyCode != "CUST-100" {
		t.Fatalf("unexpected filtered party code: %s", customers[0].PartyCode)
	}
}

func TestGetPartyReturnsExactTenantScopedPartyIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	service := parties.NewService(db)

	created, err := service.CreateParty(ctx, parties.CreatePartyInput{
		PartyCode:   "CUST-200",
		DisplayName: "Exact Customer",
		LegalName:   "Exact Customer Pvt Ltd",
		PartyKind:   parties.PartyKindCustomerVendor,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("create party: %v", err)
	}

	loaded, err := service.GetParty(ctx, parties.GetPartyInput{
		PartyID: created.ID,
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("get party: %v", err)
	}
	if loaded.ID != created.ID || loaded.PartyCode != "CUST-200" || loaded.PartyKind != parties.PartyKindCustomerVendor {
		t.Fatalf("unexpected party detail: %+v", loaded)
	}
}

func TestCreateContactRejectsMissingOrCrossTenantPartyIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	operatorSession := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: operatorSession.ID}

	otherOrgID, otherUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	otherSession := startSession(t, ctx, db, otherOrgID, otherUserID)
	otherActor := identityaccess.Actor{OrgID: otherOrgID, UserID: otherUserID, SessionID: otherSession.ID}

	service := parties.NewService(db)

	otherParty, err := service.CreateParty(ctx, parties.CreatePartyInput{
		PartyCode:   "VEND-900",
		DisplayName: "Other Tenant Vendor",
		PartyKind:   parties.PartyKindVendor,
		Actor:       otherActor,
	})
	if err != nil {
		t.Fatalf("create other tenant party: %v", err)
	}

	_, err = service.CreateContact(ctx, parties.CreateContactInput{
		PartyID:  otherParty.ID,
		FullName: "Blocked Contact",
		Email:    "blocked@example.com",
		Actor:    operator,
	})
	if !errors.Is(err, parties.ErrPartyNotFound) {
		t.Fatalf("unexpected cross-tenant error: got %v want %v", err, parties.ErrPartyNotFound)
	}

	_, err = service.CreateContact(ctx, parties.CreateContactInput{
		PartyID:  "",
		FullName: "No Party",
		Email:    "no-party@example.com",
		Actor:    operator,
	})
	if !errors.Is(err, parties.ErrInvalidContact) {
		t.Fatalf("unexpected invalid-contact error: got %v want %v", err, parties.ErrInvalidContact)
	}
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
