package identityaccess

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type BootstrapAdminInput struct {
	OrgName         string
	OrgSlug         string
	UserEmail       string
	UserDisplayName string
	Password        string
	UpdatedAt       time.Time
}

type BootstrapAdminResult struct {
	OrgID        string
	UserID       string
	MembershipID string
}

func (s *Service) BootstrapAdmin(ctx context.Context, input BootstrapAdminInput) (BootstrapAdminResult, error) {
	orgName := strings.TrimSpace(input.OrgName)
	orgSlug := strings.TrimSpace(input.OrgSlug)
	userEmail := strings.TrimSpace(input.UserEmail)
	userDisplayName := strings.TrimSpace(input.UserDisplayName)

	if orgName == "" {
		return BootstrapAdminResult{}, fmt.Errorf("org name is required")
	}
	if orgSlug == "" {
		return BootstrapAdminResult{}, fmt.Errorf("org slug is required")
	}
	if userEmail == "" {
		return BootstrapAdminResult{}, fmt.Errorf("user email is required")
	}
	if userDisplayName == "" {
		return BootstrapAdminResult{}, fmt.Errorf("user display name is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("begin bootstrap admin: %w", err)
	}

	result, err := bootstrapAdminTx(ctx, tx, BootstrapAdminInput{
		OrgName:         orgName,
		OrgSlug:         orgSlug,
		UserEmail:       userEmail,
		UserDisplayName: userDisplayName,
		Password:        input.Password,
		UpdatedAt:       input.UpdatedAt,
	})
	if err != nil {
		_ = tx.Rollback()
		return BootstrapAdminResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("commit bootstrap admin: %w", err)
	}

	return result, nil
}

func bootstrapAdminTx(ctx context.Context, tx *sql.Tx, input BootstrapAdminInput) (BootstrapAdminResult, error) {
	var orgID string
	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.orgs (slug, name, status)
VALUES ($1, $2, 'active')
ON CONFLICT ((lower(slug))) DO UPDATE
SET name = EXCLUDED.name,
	status = 'active'
RETURNING id`,
		input.OrgSlug,
		input.OrgName,
	).Scan(&orgID); err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("upsert org: %w", err)
	}

	var userID string
	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name, status)
VALUES ($1, $2, 'active')
ON CONFLICT ((lower(email))) DO UPDATE
SET display_name = EXCLUDED.display_name,
	status = 'active'
RETURNING id`,
		input.UserEmail,
		input.UserDisplayName,
	).Scan(&userID); err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("upsert user: %w", err)
	}

	var membershipID string
	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.memberships (org_id, user_id, role_code, status)
VALUES ($1, $2, $3, 'active')
ON CONFLICT (org_id, user_id) DO UPDATE
SET role_code = EXCLUDED.role_code,
	status = 'active'
RETURNING id`,
		orgID,
		userID,
		RoleAdmin,
	).Scan(&membershipID); err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("upsert membership: %w", err)
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return BootstrapAdminResult{}, err
	}

	updatedAt := input.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE identityaccess.users
SET password_hash = $2,
	password_updated_at = $3
WHERE id = $1`,
		userID,
		passwordHash,
		updatedAt,
	); err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("set bootstrap admin password: %w", err)
	}

	return BootstrapAdminResult{
		OrgID:        orgID,
		UserID:       userID,
		MembershipID: membershipID,
	}, nil
}
