package identityaccess

import (
	"context"
	"testing"
	"time"

	"workflow_app/internal/testsupport/dbtest"
)

func TestBootstrapAdminCreatesAndUpdatesFriendlyLoginIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := NewService(db)
	first, err := service.BootstrapAdmin(ctx, BootstrapAdminInput{
		OrgName:         "North Harbor Works",
		OrgSlug:         "north-harbor",
		UserEmail:       "admin@northharbor.local",
		UserDisplayName: "North Harbor Admin",
		Password:        "NorthHarbor2026",
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("BootstrapAdmin() first error = %v", err)
	}

	second, err := service.BootstrapAdmin(ctx, BootstrapAdminInput{
		OrgName:         "North Harbor Works HQ",
		OrgSlug:         "north-harbor",
		UserEmail:       "admin@northharbor.local",
		UserDisplayName: "Operations Admin",
		Password:        "NorthHarbor2027",
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("BootstrapAdmin() second error = %v", err)
	}

	if first.OrgID != second.OrgID {
		t.Fatalf("org IDs differ: %s vs %s", first.OrgID, second.OrgID)
	}
	if first.UserID != second.UserID {
		t.Fatalf("user IDs differ: %s vs %s", first.UserID, second.UserID)
	}
	if first.MembershipID != second.MembershipID {
		t.Fatalf("membership IDs differ: %s vs %s", first.MembershipID, second.MembershipID)
	}

	var orgName, roleCode, status, displayName string
	if err := db.QueryRowContext(
		ctx,
		`SELECT o.name, m.role_code, u.status, u.display_name
FROM identityaccess.orgs o
JOIN identityaccess.memberships m ON m.org_id = o.id
JOIN identityaccess.users u ON u.id = m.user_id
WHERE o.id = $1 AND u.id = $2`,
		second.OrgID,
		second.UserID,
	).Scan(&orgName, &roleCode, &status, &displayName); err != nil {
		t.Fatalf("load bootstrap admin: %v", err)
	}

	if orgName != "North Harbor Works HQ" {
		t.Fatalf("org name = %q, want updated org name", orgName)
	}
	if roleCode != RoleAdmin {
		t.Fatalf("role code = %q, want %q", roleCode, RoleAdmin)
	}
	if status != "active" {
		t.Fatalf("user status = %q, want active", status)
	}
	if displayName != "Operations Admin" {
		t.Fatalf("display name = %q, want updated display name", displayName)
	}

	session, err := service.StartBrowserSession(ctx, StartBrowserSessionInput{
		OrgSlug:     "north-harbor",
		Email:       "admin@northharbor.local",
		Password:    "NorthHarbor2027",
		DeviceLabel: "browser-test",
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("StartBrowserSession() error = %v", err)
	}
	if session.Session.UserID != second.UserID {
		t.Fatalf("session user ID = %s, want %s", session.Session.UserID, second.UserID)
	}
}
