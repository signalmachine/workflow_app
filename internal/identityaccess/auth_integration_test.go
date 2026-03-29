package identityaccess

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"workflow_app/internal/testsupport/dbtest"
)

func TestHashPasswordRejectsShortPassword(t *testing.T) {
	if _, err := HashPassword("too-short"); !errors.Is(err, ErrPasswordInvalid) {
		t.Fatalf("expected short password rejection, got %v", err)
	}
}

func TestStartBrowserSessionRequiresPasswordIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := NewService(db)
	orgID, orgSlug, userEmail, userID := seedAuthUser(t, ctx, db)
	_ = orgID

	if err := service.SetUserPassword(ctx, SetUserPasswordInput{
		UserID:    userID,
		Password:  "workflow-auth-password",
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("set password: %v", err)
	}

	if _, err := service.StartBrowserSession(ctx, StartBrowserSessionInput{
		OrgSlug:     orgSlug,
		Email:       userEmail,
		Password:    "wrong-password",
		DeviceLabel: "browser-test",
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
	}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized for wrong password, got %v", err)
	}

	session, err := service.StartBrowserSession(ctx, StartBrowserSessionInput{
		OrgSlug:     orgSlug,
		Email:       userEmail,
		Password:    "workflow-auth-password",
		DeviceLabel: "browser-test",
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("start browser session: %v", err)
	}
	if session.Session.UserID != userID || session.UserEmail != userEmail || session.RefreshToken == "" {
		t.Fatalf("unexpected browser session payload: %+v", session)
	}
}

func seedAuthUser(t *testing.T, ctx context.Context, db *sql.DB) (orgID, orgSlug, userEmail, userID string) {
	t.Helper()

	orgSlug = "auth-org-" + time.Now().UTC().Format("150405.000000000")
	userEmail = "auth-user-" + time.Now().UTC().Format("150405.000000000") + "@example.com"

	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.orgs (slug, name) VALUES ($1, 'Auth Org') RETURNING id`,
		orgSlug,
	).Scan(&orgID); err != nil {
		t.Fatalf("insert org: %v", err)
	}

	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name) VALUES ($1, 'Auth User') RETURNING id`,
		userEmail,
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO identityaccess.memberships (org_id, user_id, role_code) VALUES ($1, $2, $3)`,
		orgID,
		userID,
		RoleOperator,
	); err != nil {
		t.Fatalf("insert membership: %v", err)
	}

	return orgID, orgSlug, userEmail, userID
}
