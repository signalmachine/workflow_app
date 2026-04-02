package identityaccess

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/platform/audit"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrMembershipNotFound  = errors.New("membership not found")
	ErrInvalidUser         = errors.New("invalid user")
	ErrInvalidMembership   = errors.New("invalid membership")
	ErrProtectedMembership = errors.New("protected membership")
)

type OrgUserMembership struct {
	MembershipID     string
	OrgID            string
	UserID           string
	UserEmail        string
	UserDisplayName  string
	UserStatus       string
	RoleCode         string
	MembershipStatus string
	CreatedAt        time.Time
}

type ListOrgUsersInput struct {
	Actor Actor
}

type ProvisionOrgUserInput struct {
	Email       string
	DisplayName string
	RoleCode    string
	Password    string
	Actor       Actor
}

type UpdateMembershipRoleInput struct {
	MembershipID string
	RoleCode     string
	Actor        Actor
}

type identityUser struct {
	ID          string
	Email       string
	DisplayName string
	Status      string
	CreatedAt   time.Time
}

func (s *Service) ListOrgUsers(ctx context.Context, input ListOrgUsersInput) ([]OrgUserMembership, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list org users: %w", err)
	}

	if err := AuthorizeTx(ctx, tx, input.Actor, RoleAdmin); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
	m.id,
	m.org_id,
	m.user_id,
	u.email,
	u.display_name,
	u.status,
	m.role_code,
	m.status,
	m.created_at
FROM identityaccess.memberships m
JOIN identityaccess.users u
  ON u.id = m.user_id
WHERE m.org_id = $1
ORDER BY lower(u.email), m.created_at`,
		input.Actor.OrgID,
	)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("query org users: %w", err)
	}
	defer rows.Close()

	var memberships []OrgUserMembership
	for rows.Next() {
		item, scanErr := scanOrgUserMembership(rows)
		if scanErr != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("scan org user: %w", scanErr)
		}
		memberships = append(memberships, item)
	}
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("iterate org users: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list org users: %w", err)
	}
	return memberships, nil
}

func (s *Service) ProvisionOrgUser(ctx context.Context, input ProvisionOrgUserInput) (OrgUserMembership, error) {
	email := strings.TrimSpace(input.Email)
	displayName := strings.TrimSpace(input.DisplayName)
	roleCode := strings.TrimSpace(input.RoleCode)
	password := strings.TrimSpace(input.Password)

	if email == "" || !isValidRoleCode(roleCode) {
		return OrgUserMembership{}, ErrInvalidUser
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return OrgUserMembership{}, fmt.Errorf("begin provision org user: %w", err)
	}

	if err := AuthorizeTx(ctx, tx, input.Actor, RoleAdmin); err != nil {
		_ = tx.Rollback()
		return OrgUserMembership{}, err
	}

	user, userExists, err := getUserByEmailForUpdate(ctx, tx, email)
	if err != nil {
		_ = tx.Rollback()
		return OrgUserMembership{}, err
	}

	if !userExists {
		if displayName == "" {
			_ = tx.Rollback()
			return OrgUserMembership{}, ErrInvalidUser
		}
		passwordHash, hashErr := HashPassword(password)
		if hashErr != nil {
			_ = tx.Rollback()
			if errors.Is(hashErr, ErrPasswordInvalid) {
				return OrgUserMembership{}, ErrInvalidUser
			}
			return OrgUserMembership{}, hashErr
		}
		user, err = scanIdentityUser(tx.QueryRowContext(ctx, `
INSERT INTO identityaccess.users (
	email,
	display_name,
	password_hash,
	password_updated_at,
	status
) VALUES ($1, $2, $3, NOW(), 'active')
RETURNING
	id,
	email,
	display_name,
	status,
	created_at`,
			email,
			displayName,
			passwordHash,
		))
		if err != nil {
			_ = tx.Rollback()
			return OrgUserMembership{}, fmt.Errorf("insert user: %w", err)
		}
		if err := audit.WriteTx(ctx, tx, audit.Event{
			OrgID:       input.Actor.OrgID,
			ActorUserID: input.Actor.UserID,
			EventType:   "identityaccess.user_created",
			EntityType:  "identityaccess.user",
			EntityID:    user.ID,
			Payload: map[string]any{
				"email": user.Email,
			},
		}); err != nil {
			_ = tx.Rollback()
			return OrgUserMembership{}, err
		}
	} else if password != "" {
		_ = tx.Rollback()
		return OrgUserMembership{}, ErrInvalidUser
	}

	previousMembership, membershipExists, err := getMembershipByOrgUserForUpdate(ctx, tx, input.Actor.OrgID, user.ID)
	if err != nil {
		_ = tx.Rollback()
		return OrgUserMembership{}, err
	}

	membership, err := scanOrgUserMembership(tx.QueryRowContext(ctx, `
INSERT INTO identityaccess.memberships (
	org_id,
	user_id,
	role_code,
	status
) VALUES ($1, $2, $3, 'active')
ON CONFLICT (org_id, user_id) DO UPDATE
SET role_code = EXCLUDED.role_code,
	status = 'active'
RETURNING
	id,
	org_id,
	user_id,
	$4,
	$5,
	$6,
	role_code,
	status,
	created_at`,
		input.Actor.OrgID,
		user.ID,
		roleCode,
		user.Email,
		user.DisplayName,
		user.Status,
	))
	if err != nil {
		_ = tx.Rollback()
		return OrgUserMembership{}, fmt.Errorf("upsert membership: %w", err)
	}

	action := "created"
	if membershipExists {
		action = "updated"
		if previousMembership.MembershipStatus != "active" {
			action = "reactivated"
		}
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "identityaccess.membership_provisioned",
		EntityType:  "identityaccess.membership",
		EntityID:    membership.MembershipID,
		Payload: map[string]any{
			"user_id":       membership.UserID,
			"email":         membership.UserEmail,
			"role_code":     membership.RoleCode,
			"action":        action,
			"prior_role":    previousMembership.RoleCode,
			"prior_status":  previousMembership.MembershipStatus,
			"user_existing": userExists,
		},
	}); err != nil {
		_ = tx.Rollback()
		return OrgUserMembership{}, err
	}

	if err := tx.Commit(); err != nil {
		return OrgUserMembership{}, fmt.Errorf("commit provision org user: %w", err)
	}
	return membership, nil
}

func (s *Service) UpdateMembershipRole(ctx context.Context, input UpdateMembershipRoleInput) (OrgUserMembership, error) {
	membershipID := strings.TrimSpace(input.MembershipID)
	roleCode := strings.TrimSpace(input.RoleCode)
	if membershipID == "" || !isValidRoleCode(roleCode) {
		return OrgUserMembership{}, ErrInvalidMembership
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return OrgUserMembership{}, fmt.Errorf("begin update membership role: %w", err)
	}

	if err := AuthorizeTx(ctx, tx, input.Actor, RoleAdmin); err != nil {
		_ = tx.Rollback()
		return OrgUserMembership{}, err
	}

	current, err := getMembershipByIDForUpdate(ctx, tx, input.Actor.OrgID, membershipID)
	if err != nil {
		_ = tx.Rollback()
		return OrgUserMembership{}, err
	}
	if current.UserID == input.Actor.UserID && !strings.EqualFold(current.RoleCode, roleCode) {
		_ = tx.Rollback()
		return OrgUserMembership{}, ErrProtectedMembership
	}

	updated, err := scanOrgUserMembership(tx.QueryRowContext(ctx, `
UPDATE identityaccess.memberships m
SET role_code = $2
FROM identityaccess.users u
WHERE m.id = $1
  AND m.org_id = $3
  AND u.id = m.user_id
RETURNING
	m.id,
	m.org_id,
	m.user_id,
	u.email,
	u.display_name,
	u.status,
	m.role_code,
	m.status,
	m.created_at`,
		membershipID,
		roleCode,
		input.Actor.OrgID,
	))
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return OrgUserMembership{}, ErrMembershipNotFound
		}
		return OrgUserMembership{}, fmt.Errorf("update membership role: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "identityaccess.membership_role_changed",
		EntityType:  "identityaccess.membership",
		EntityID:    updated.MembershipID,
		Payload: map[string]any{
			"user_id":       updated.UserID,
			"email":         updated.UserEmail,
			"previous_role": current.RoleCode,
			"next_role":     updated.RoleCode,
			"membership_id": updated.MembershipID,
		},
	}); err != nil {
		_ = tx.Rollback()
		return OrgUserMembership{}, err
	}

	if err := tx.Commit(); err != nil {
		return OrgUserMembership{}, fmt.Errorf("commit update membership role: %w", err)
	}
	return updated, nil
}

func getUserByEmailForUpdate(ctx context.Context, tx *sql.Tx, email string) (identityUser, bool, error) {
	user, err := scanIdentityUser(tx.QueryRowContext(ctx, `
SELECT
	id,
	email,
	display_name,
	status,
	created_at
FROM identityaccess.users
WHERE lower(email) = lower($1)
FOR UPDATE`,
		email,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identityUser{}, false, nil
		}
		return identityUser{}, false, fmt.Errorf("load user by email: %w", err)
	}
	return user, true, nil
}

func getMembershipByOrgUserForUpdate(ctx context.Context, tx *sql.Tx, orgID, userID string) (OrgUserMembership, bool, error) {
	item, err := scanOrgUserMembership(tx.QueryRowContext(ctx, `
SELECT
	m.id,
	m.org_id,
	m.user_id,
	u.email,
	u.display_name,
	u.status,
	m.role_code,
	m.status,
	m.created_at
FROM identityaccess.memberships m
JOIN identityaccess.users u
  ON u.id = m.user_id
WHERE m.org_id = $1
  AND m.user_id = $2
FOR UPDATE OF m`,
		orgID,
		userID,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OrgUserMembership{}, false, nil
		}
		return OrgUserMembership{}, false, fmt.Errorf("load membership by org and user: %w", err)
	}
	return item, true, nil
}

func getMembershipByIDForUpdate(ctx context.Context, tx *sql.Tx, orgID, membershipID string) (OrgUserMembership, error) {
	item, err := scanOrgUserMembership(tx.QueryRowContext(ctx, `
SELECT
	m.id,
	m.org_id,
	m.user_id,
	u.email,
	u.display_name,
	u.status,
	m.role_code,
	m.status,
	m.created_at
FROM identityaccess.memberships m
JOIN identityaccess.users u
  ON u.id = m.user_id
WHERE m.org_id = $1
  AND m.id = $2
FOR UPDATE OF m`,
		orgID,
		membershipID,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OrgUserMembership{}, ErrMembershipNotFound
		}
		return OrgUserMembership{}, fmt.Errorf("load membership by id: %w", err)
	}
	return item, nil
}

func scanIdentityUser(row rowScanner) (identityUser, error) {
	var user identityUser
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.Status,
		&user.CreatedAt,
	)
	if err != nil {
		return identityUser{}, err
	}
	return user, nil
}

func scanOrgUserMembership(row rowScanner) (OrgUserMembership, error) {
	var membership OrgUserMembership
	err := row.Scan(
		&membership.MembershipID,
		&membership.OrgID,
		&membership.UserID,
		&membership.UserEmail,
		&membership.UserDisplayName,
		&membership.UserStatus,
		&membership.RoleCode,
		&membership.MembershipStatus,
		&membership.CreatedAt,
	)
	if err != nil {
		return OrgUserMembership{}, err
	}
	return membership, nil
}

func isValidRoleCode(roleCode string) bool {
	switch strings.TrimSpace(roleCode) {
	case RoleAdmin, RoleOperator, RoleApprover:
		return true
	default:
		return false
	}
}
