package identityaccess

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrSessionNotActive  = errors.New("session not active")
	ErrMembershipMissing = errors.New("membership not found")
)

const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleApprover = "approver"
)

type Actor struct {
	OrgID     string
	UserID    string
	SessionID string
}

type Session struct {
	ID                  string
	OrgID               string
	UserID              string
	MembershipID        string
	DeviceLabel         string
	RefreshTokenHash    string
	Status              string
	ExpiresAt           time.Time
	ReplacedBySessionID sql.NullString
	IssuedAt            time.Time
	LastSeenAt          time.Time
}

type StartSessionInput struct {
	OrgID            string
	UserID           string
	DeviceLabel      string
	RefreshTokenHash string
	ExpiresAt        time.Time
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) StartSession(ctx context.Context, input StartSessionInput) (Session, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, fmt.Errorf("begin start session: %w", err)
	}

	session, err := startSessionTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Session{}, err
	}

	if err := tx.Commit(); err != nil {
		return Session{}, fmt.Errorf("commit start session: %w", err)
	}

	return session, nil
}

func (s *Service) RevokeSession(ctx context.Context, actor Actor, sessionID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin revoke session: %w", err)
	}

	if err := AuthorizeTx(ctx, tx, actor, RoleAdmin, RoleOperator, RoleApprover); err != nil {
		_ = tx.Rollback()
		return err
	}

	const statement = `
UPDATE identityaccess.sessions
SET status = 'revoked',
	last_seen_at = NOW()
WHERE id = $1
  AND org_id = $2
  AND user_id = $3
  AND status = 'active';`

	result, err := tx.ExecContext(ctx, statement, sessionID, actor.OrgID, actor.UserID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("revoke session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("revoke session rows affected: %w", err)
	}
	if rows == 0 {
		_ = tx.Rollback()
		return ErrSessionNotActive
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit revoke session: %w", err)
	}

	return nil
}

func AuthorizeTx(ctx context.Context, tx *sql.Tx, actor Actor, allowedRoles ...string) error {
	const query = `
SELECT m.role_code
FROM identityaccess.sessions s
JOIN identityaccess.memberships m
  ON m.id = s.membership_id
WHERE s.id = $1
  AND s.org_id = $2
  AND s.user_id = $3
  AND s.status = 'active'
  AND s.expires_at > NOW()
  AND m.status = 'active'
FOR UPDATE;`

	var roleCode string
	if err := tx.QueryRowContext(ctx, query, actor.SessionID, actor.OrgID, actor.UserID).Scan(&roleCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUnauthorized
		}
		return fmt.Errorf("authorize actor: %w", err)
	}

	for _, allowedRole := range allowedRoles {
		if roleCode == allowedRole {
			return nil
		}
	}

	return ErrUnauthorized
}

func startSessionTx(ctx context.Context, tx *sql.Tx, input StartSessionInput) (Session, error) {
	const membershipQuery = `
SELECT id
FROM identityaccess.memberships
WHERE org_id = $1
  AND user_id = $2
  AND status = 'active'
FOR UPDATE;`

	var membershipID string
	if err := tx.QueryRowContext(ctx, membershipQuery, input.OrgID, input.UserID).Scan(&membershipID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrMembershipMissing
		}
		return Session{}, fmt.Errorf("load membership for session: %w", err)
	}

	const insert = `
INSERT INTO identityaccess.sessions (
	org_id,
	user_id,
	membership_id,
	device_label,
	refresh_token_hash,
	status,
	expires_at
) VALUES ($1, $2, $3, $4, $5, 'active', $6)
RETURNING
	id,
	org_id,
	user_id,
	membership_id,
	device_label,
	refresh_token_hash,
	status,
	expires_at,
	replaced_by_session_id,
	issued_at,
	last_seen_at;`

	return scanSession(tx.QueryRowContext(
		ctx,
		insert,
		input.OrgID,
		input.UserID,
		membershipID,
		input.DeviceLabel,
		input.RefreshTokenHash,
		input.ExpiresAt,
	))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSession(row rowScanner) (Session, error) {
	var session Session
	err := row.Scan(
		&session.ID,
		&session.OrgID,
		&session.UserID,
		&session.MembershipID,
		&session.DeviceLabel,
		&session.RefreshTokenHash,
		&session.Status,
		&session.ExpiresAt,
		&session.ReplacedBySessionID,
		&session.IssuedAt,
		&session.LastSeenAt,
	)
	if err != nil {
		return Session{}, err
	}
	return session, nil
}
