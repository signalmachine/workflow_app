package identityaccess_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/testsupport/dbtest"
)

func TestServiceProvisionOrgUserListAndUpdateRole(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, _, _, adminUserID := seedAuthUser(t, ctx, db)
	service := identityaccess.NewService(db)
	adminActor := issueAdminActor(t, ctx, service, orgID, adminUserID)

	created, err := service.ProvisionOrgUser(ctx, identityaccess.ProvisionOrgUserInput{
		Email:       "operator@example.com",
		DisplayName: "Operator One",
		RoleCode:    identityaccess.RoleOperator,
		Password:    "operator-password-123",
		Actor:       adminActor,
	})
	if err != nil {
		t.Fatalf("provision org user: %v", err)
	}
	if created.UserEmail != "operator@example.com" || created.RoleCode != identityaccess.RoleOperator {
		t.Fatalf("unexpected provisioned user: %+v", created)
	}

	listed, err := service.ListOrgUsers(ctx, identityaccess.ListOrgUsersInput{Actor: adminActor})
	if err != nil {
		t.Fatalf("list org users: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected 2 org users, got %d", len(listed))
	}

	updated, err := service.UpdateMembershipRole(ctx, identityaccess.UpdateMembershipRoleInput{
		MembershipID: created.MembershipID,
		RoleCode:     identityaccess.RoleApprover,
		Actor:        adminActor,
	})
	if err != nil {
		t.Fatalf("update membership role: %v", err)
	}
	if updated.RoleCode != identityaccess.RoleApprover {
		t.Fatalf("updated role = %q, want %q", updated.RoleCode, identityaccess.RoleApprover)
	}

	var auditCount int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
FROM platform.audit_events
WHERE org_id = $1
  AND event_type IN (
	'identityaccess.user_created',
	'identityaccess.membership_provisioned',
	'identityaccess.membership_role_changed'
  )`,
		orgID,
	).Scan(&auditCount); err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	if auditCount < 3 {
		t.Fatalf("expected audit events for provision and role change, got %d", auditCount)
	}
}

func TestServiceProvisionOrgUserAttachesExistingUserWithoutPasswordChange(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, _, _, adminUserID := seedAuthUser(t, ctx, db)
	service := identityaccess.NewService(db)
	adminActor := issueAdminActor(t, ctx, service, orgID, adminUserID)

	var existingUserID string
	if err := db.QueryRowContext(ctx, `
INSERT INTO identityaccess.users (email, display_name, password_hash, password_updated_at, status)
VALUES ($1, $2, $3, NOW(), 'active')
RETURNING id`,
		"shared-user@example.com",
		"Shared User",
		mustHashPassword(t, "shared-user-password-123"),
	).Scan(&existingUserID); err != nil {
		t.Fatalf("insert shared user: %v", err)
	}

	provisioned, err := service.ProvisionOrgUser(ctx, identityaccess.ProvisionOrgUserInput{
		Email:       "shared-user@example.com",
		DisplayName: "Ignored Name",
		RoleCode:    identityaccess.RoleOperator,
		Actor:       adminActor,
	})
	if err != nil {
		t.Fatalf("attach existing user membership: %v", err)
	}
	if provisioned.UserID != existingUserID {
		t.Fatalf("attached user ID = %s, want %s", provisioned.UserID, existingUserID)
	}

	_, err = service.ProvisionOrgUser(ctx, identityaccess.ProvisionOrgUserInput{
		Email:       "shared-user@example.com",
		DisplayName: "Shared User",
		RoleCode:    identityaccess.RoleOperator,
		Password:    "should-not-be-accepted",
		Actor:       adminActor,
	})
	if !errors.Is(err, identityaccess.ErrInvalidUser) {
		t.Fatalf("expected invalid user when password supplied for existing user, got %v", err)
	}
}

func TestServiceUpdateMembershipRoleProtectsCurrentAdmin(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, _, _, adminUserID := seedAuthUser(t, ctx, db)
	service := identityaccess.NewService(db)
	adminActor := issueAdminActor(t, ctx, service, orgID, adminUserID)

	var membershipID string
	if err := db.QueryRowContext(ctx, `
SELECT id
FROM identityaccess.memberships
WHERE org_id = $1
  AND user_id = $2`,
		orgID,
		adminUserID,
	).Scan(&membershipID); err != nil {
		t.Fatalf("load admin membership: %v", err)
	}

	_, err := service.UpdateMembershipRole(ctx, identityaccess.UpdateMembershipRoleInput{
		MembershipID: membershipID,
		RoleCode:     identityaccess.RoleOperator,
		Actor:        adminActor,
	})
	if !errors.Is(err, identityaccess.ErrProtectedMembership) {
		t.Fatalf("expected protected membership error, got %v", err)
	}
}

func issueAdminActor(t *testing.T, ctx context.Context, service *identityaccess.Service, orgID, userID string) identityaccess.Actor {
	t.Helper()

	session, err := service.StartSession(ctx, identityaccess.StartSessionInput{
		OrgID:            orgID,
		UserID:           userID,
		DeviceLabel:      "admin-maintenance-test",
		RefreshTokenHash: "refresh-token-hash",
		ExpiresAt:        time.Now().UTC().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("start admin session: %v", err)
	}
	return identityaccess.Actor{
		OrgID:     orgID,
		UserID:    userID,
		SessionID: session.ID,
	}
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := identityaccess.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return hash
}

func seedAuthUser(t *testing.T, ctx context.Context, db *sql.DB) (orgID, orgSlug, userEmail, userID string) {
	t.Helper()

	suffix := time.Now().UTC().Format("150405.000000000")
	orgSlug = "admin-org-" + suffix
	userEmail = "admin-" + suffix + "@example.com"

	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.orgs (slug, name) VALUES ($1, 'North Harbor') RETURNING id`,
		orgSlug,
	).Scan(&orgID); err != nil {
		t.Fatalf("insert org: %v", err)
	}

	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name, password_hash, password_updated_at, status)
VALUES ($1, 'Admin User', $2, NOW(), 'active')
RETURNING id`,
		userEmail,
		mustHashPassword(t, "north-harbor-password-123"),
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO identityaccess.memberships (org_id, user_id, role_code) VALUES ($1, $2, $3)`,
		orgID,
		userID,
		identityaccess.RoleAdmin,
	); err != nil {
		t.Fatalf("insert membership: %v", err)
	}

	return orgID, orgSlug, userEmail, userID
}
