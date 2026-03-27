package identityaccess

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrSessionNotActive  = errors.New("session not active")
	ErrMembershipMissing = errors.New("membership not found")
	ErrSessionInvalid    = errors.New("session invalid")
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

type StartBrowserSessionInput struct {
	OrgSlug     string
	Email       string
	DeviceLabel string
	ExpiresAt   time.Time
}

type StartTokenSessionInput struct {
	OrgSlug              string
	Email                string
	DeviceLabel          string
	SessionExpiresAt     time.Time
	AccessTokenExpiresAt time.Time
}

type BrowserSession struct {
	Session         Session
	RefreshToken    string
	RoleCode        string
	OrgSlug         string
	OrgName         string
	UserEmail       string
	UserDisplayName string
}

type TokenSession struct {
	Session               Session
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	RoleCode              string
	OrgSlug               string
	OrgName               string
	UserEmail             string
	UserDisplayName       string
}

type SessionContext struct {
	Actor           Actor
	Session         Session
	RoleCode        string
	OrgSlug         string
	OrgName         string
	UserEmail       string
	UserDisplayName string
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

func (s *Service) StartBrowserSession(ctx context.Context, input StartBrowserSessionInput) (BrowserSession, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BrowserSession{}, fmt.Errorf("begin start browser session: %w", err)
	}

	profile, err := loadBrowserProfileTx(ctx, tx, input.OrgSlug, input.Email)
	if err != nil {
		_ = tx.Rollback()
		return BrowserSession{}, err
	}

	refreshToken, refreshTokenHash, err := newRefreshToken()
	if err != nil {
		_ = tx.Rollback()
		return BrowserSession{}, err
	}

	session, err := startSessionTx(ctx, tx, StartSessionInput{
		OrgID:            profile.OrgID,
		UserID:           profile.UserID,
		DeviceLabel:      strings.TrimSpace(input.DeviceLabel),
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        input.ExpiresAt,
	})
	if err != nil {
		_ = tx.Rollback()
		return BrowserSession{}, err
	}

	if err := tx.Commit(); err != nil {
		return BrowserSession{}, fmt.Errorf("commit start browser session: %w", err)
	}

	return BrowserSession{
		Session:         session,
		RefreshToken:    refreshToken,
		RoleCode:        profile.RoleCode,
		OrgSlug:         profile.OrgSlug,
		OrgName:         profile.OrgName,
		UserEmail:       profile.UserEmail,
		UserDisplayName: profile.UserDisplayName,
	}, nil
}

func (s *Service) StartTokenSession(ctx context.Context, input StartTokenSessionInput) (TokenSession, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TokenSession{}, fmt.Errorf("begin start token session: %w", err)
	}

	profile, err := loadBrowserProfileTx(ctx, tx, input.OrgSlug, input.Email)
	if err != nil {
		_ = tx.Rollback()
		return TokenSession{}, err
	}

	refreshToken, refreshTokenHash, err := newRefreshToken()
	if err != nil {
		_ = tx.Rollback()
		return TokenSession{}, err
	}

	session, err := startSessionTx(ctx, tx, StartSessionInput{
		OrgID:            profile.OrgID,
		UserID:           profile.UserID,
		DeviceLabel:      strings.TrimSpace(input.DeviceLabel),
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        input.SessionExpiresAt,
	})
	if err != nil {
		_ = tx.Rollback()
		return TokenSession{}, err
	}

	accessToken, accessTokenHash, err := newAccessToken()
	if err != nil {
		_ = tx.Rollback()
		return TokenSession{}, err
	}

	tokenRecord, err := insertAccessTokenTx(ctx, tx, session.ID, accessTokenHash, input.AccessTokenExpiresAt)
	if err != nil {
		_ = tx.Rollback()
		return TokenSession{}, err
	}

	if err := tx.Commit(); err != nil {
		return TokenSession{}, fmt.Errorf("commit start token session: %w", err)
	}

	return TokenSession{
		Session:               session,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  tokenRecord.ExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: session.ExpiresAt,
		RoleCode:              profile.RoleCode,
		OrgSlug:               profile.OrgSlug,
		OrgName:               profile.OrgName,
		UserEmail:             profile.UserEmail,
		UserDisplayName:       profile.UserDisplayName,
	}, nil
}

func (s *Service) AuthenticateSession(ctx context.Context, sessionID, refreshToken string) (SessionContext, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SessionContext{}, fmt.Errorf("begin authenticate session: %w", err)
	}

	session, err := authenticateSessionTx(ctx, tx, sessionID, refreshToken, true)
	if err != nil {
		_ = tx.Rollback()
		return SessionContext{}, err
	}

	if err := tx.Commit(); err != nil {
		return SessionContext{}, fmt.Errorf("commit authenticate session: %w", err)
	}

	return session, nil
}

func (s *Service) AuthenticateAccessToken(ctx context.Context, accessToken string) (SessionContext, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SessionContext{}, fmt.Errorf("begin authenticate access token: %w", err)
	}

	session, err := authenticateAccessTokenTx(ctx, tx, accessToken, true)
	if err != nil {
		_ = tx.Rollback()
		return SessionContext{}, err
	}

	if err := tx.Commit(); err != nil {
		return SessionContext{}, fmt.Errorf("commit authenticate access token: %w", err)
	}

	return session, nil
}

func (s *Service) RevokeAuthenticatedSession(ctx context.Context, sessionID, refreshToken string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin revoke authenticated session: %w", err)
	}

	session, err := authenticateSessionTx(ctx, tx, sessionID, refreshToken, false)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := revokeSessionTx(ctx, tx, session.Session.ID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit revoke authenticated session: %w", err)
	}

	return nil
}

func (s *Service) RefreshTokenSession(ctx context.Context, sessionID, refreshToken string, accessTokenExpiresAt time.Time) (TokenSession, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TokenSession{}, fmt.Errorf("begin refresh token session: %w", err)
	}

	sessionContext, err := authenticateSessionTx(ctx, tx, sessionID, refreshToken, false)
	if err != nil {
		_ = tx.Rollback()
		return TokenSession{}, err
	}

	nextRefreshToken, nextRefreshTokenHash, err := newRefreshToken()
	if err != nil {
		_ = tx.Rollback()
		return TokenSession{}, err
	}

	nextAccessToken, nextAccessTokenHash, err := newAccessToken()
	if err != nil {
		_ = tx.Rollback()
		return TokenSession{}, err
	}

	tokenRecord, err := rotateSessionTokensTx(ctx, tx, sessionContext.Session.ID, nextRefreshTokenHash, nextAccessTokenHash, accessTokenExpiresAt)
	if err != nil {
		_ = tx.Rollback()
		return TokenSession{}, err
	}

	sessionContext.Session.RefreshTokenHash = nextRefreshTokenHash
	sessionContext.Session.LastSeenAt = tokenRecord.LastSeenAt

	if err := tx.Commit(); err != nil {
		return TokenSession{}, fmt.Errorf("commit refresh token session: %w", err)
	}

	return TokenSession{
		Session:               sessionContext.Session,
		AccessToken:           nextAccessToken,
		AccessTokenExpiresAt:  tokenRecord.ExpiresAt,
		RefreshToken:          nextRefreshToken,
		RefreshTokenExpiresAt: sessionContext.Session.ExpiresAt,
		RoleCode:              sessionContext.RoleCode,
		OrgSlug:               sessionContext.OrgSlug,
		OrgName:               sessionContext.OrgName,
		UserEmail:             sessionContext.UserEmail,
		UserDisplayName:       sessionContext.UserDisplayName,
	}, nil
}

func (s *Service) RevokeAccessTokenSession(ctx context.Context, accessToken string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin revoke access-token session: %w", err)
	}

	session, err := authenticateAccessTokenTx(ctx, tx, accessToken, false)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := revokeSessionTx(ctx, tx, session.Session.ID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit revoke access-token session: %w", err)
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

type browserProfile struct {
	OrgID           string
	UserID          string
	RoleCode        string
	OrgSlug         string
	OrgName         string
	UserEmail       string
	UserDisplayName string
}

func loadBrowserProfileTx(ctx context.Context, tx *sql.Tx, orgSlug, email string) (browserProfile, error) {
	const query = `
SELECT
	o.id,
	u.id,
	m.role_code,
	o.slug,
	o.name,
	u.email,
	u.display_name
FROM identityaccess.memberships m
JOIN identityaccess.orgs o
  ON o.id = m.org_id
JOIN identityaccess.users u
  ON u.id = m.user_id
WHERE lower(o.slug) = lower($1)
  AND lower(u.email) = lower($2)
  AND o.status = 'active'
  AND u.status = 'active'
  AND m.status = 'active'
FOR UPDATE OF m;`

	var profile browserProfile
	if err := tx.QueryRowContext(ctx, query, strings.TrimSpace(orgSlug), strings.TrimSpace(email)).Scan(
		&profile.OrgID,
		&profile.UserID,
		&profile.RoleCode,
		&profile.OrgSlug,
		&profile.OrgName,
		&profile.UserEmail,
		&profile.UserDisplayName,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return browserProfile{}, ErrUnauthorized
		}
		return browserProfile{}, fmt.Errorf("load browser profile: %w", err)
	}

	return profile, nil
}

func authenticateSessionTx(ctx context.Context, tx *sql.Tx, sessionID, refreshToken string, updateLastSeen bool) (SessionContext, error) {
	const query = `
SELECT
	s.id,
	s.org_id,
	s.user_id,
	s.membership_id,
	s.device_label,
	s.refresh_token_hash,
	s.status,
	s.expires_at,
	s.replaced_by_session_id,
	s.issued_at,
	s.last_seen_at,
	m.role_code,
	o.slug,
	o.name,
	u.email,
	u.display_name
FROM identityaccess.sessions s
JOIN identityaccess.memberships m
  ON m.id = s.membership_id
JOIN identityaccess.orgs o
  ON o.id = s.org_id
JOIN identityaccess.users u
  ON u.id = s.user_id
WHERE s.id = $1
  AND s.status = 'active'
  AND s.expires_at > NOW()
  AND m.status = 'active'
  AND o.status = 'active'
  AND u.status = 'active'
FOR UPDATE OF s;`

	var (
		session Session
		context SessionContext
	)
	err := tx.QueryRowContext(ctx, query, strings.TrimSpace(sessionID)).Scan(
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
		&context.RoleCode,
		&context.OrgSlug,
		&context.OrgName,
		&context.UserEmail,
		&context.UserDisplayName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionContext{}, ErrUnauthorized
		}
		return SessionContext{}, fmt.Errorf("load authenticated session: %w", err)
	}

	if !refreshTokenMatches(strings.TrimSpace(refreshToken), session.RefreshTokenHash) {
		return SessionContext{}, ErrUnauthorized
	}

	if updateLastSeen {
		const update = `
UPDATE identityaccess.sessions
SET last_seen_at = NOW()
WHERE id = $1
RETURNING last_seen_at;`
		if err := tx.QueryRowContext(ctx, update, session.ID).Scan(&session.LastSeenAt); err != nil {
			return SessionContext{}, fmt.Errorf("update session last seen: %w", err)
		}
	}

	return SessionContext{
		Actor: Actor{
			OrgID:     session.OrgID,
			UserID:    session.UserID,
			SessionID: session.ID,
		},
		Session:         session,
		RoleCode:        context.RoleCode,
		OrgSlug:         context.OrgSlug,
		OrgName:         context.OrgName,
		UserEmail:       context.UserEmail,
		UserDisplayName: context.UserDisplayName,
	}, nil
}

func authenticateAccessTokenTx(ctx context.Context, tx *sql.Tx, accessToken string, updateLastSeen bool) (SessionContext, error) {
	const query = `
SELECT
	t.id,
	t.session_id,
	t.token_hash,
	t.status,
	t.expires_at,
	t.replaced_by_access_token_id,
	t.issued_at,
	t.last_seen_at,
	s.id,
	s.org_id,
	s.user_id,
	s.membership_id,
	s.device_label,
	s.refresh_token_hash,
	s.status,
	s.expires_at,
	s.replaced_by_session_id,
	s.issued_at,
	s.last_seen_at,
	m.role_code,
	o.slug,
	o.name,
	u.email,
	u.display_name
FROM identityaccess.session_access_tokens t
JOIN identityaccess.sessions s
  ON s.id = t.session_id
JOIN identityaccess.memberships m
  ON m.id = s.membership_id
JOIN identityaccess.orgs o
  ON o.id = s.org_id
JOIN identityaccess.users u
  ON u.id = s.user_id
WHERE t.token_hash = $1
  AND t.status = 'active'
  AND t.expires_at > NOW()
  AND s.status = 'active'
  AND s.expires_at > NOW()
  AND m.status = 'active'
  AND o.status = 'active'
  AND u.status = 'active'
FOR UPDATE OF t, s;`

	var (
		accessTokenRecord accessTokenRecord
		session           Session
		context           SessionContext
	)
	tokenHash := hashAccessToken(strings.TrimSpace(accessToken))
	err := tx.QueryRowContext(ctx, query, tokenHash).Scan(
		&accessTokenRecord.ID,
		&accessTokenRecord.SessionID,
		&accessTokenRecord.TokenHash,
		&accessTokenRecord.Status,
		&accessTokenRecord.ExpiresAt,
		&accessTokenRecord.ReplacedByAccessTokenID,
		&accessTokenRecord.IssuedAt,
		&accessTokenRecord.LastSeenAt,
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
		&context.RoleCode,
		&context.OrgSlug,
		&context.OrgName,
		&context.UserEmail,
		&context.UserDisplayName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionContext{}, ErrUnauthorized
		}
		return SessionContext{}, fmt.Errorf("load authenticated access token: %w", err)
	}

	if updateLastSeen {
		const updateToken = `
UPDATE identityaccess.session_access_tokens
SET last_seen_at = NOW()
WHERE id = $1
RETURNING last_seen_at;`
		if err := tx.QueryRowContext(ctx, updateToken, accessTokenRecord.ID).Scan(&accessTokenRecord.LastSeenAt); err != nil {
			return SessionContext{}, fmt.Errorf("update access token last seen: %w", err)
		}
		const updateSession = `
UPDATE identityaccess.sessions
SET last_seen_at = NOW()
WHERE id = $1
RETURNING last_seen_at;`
		if err := tx.QueryRowContext(ctx, updateSession, session.ID).Scan(&session.LastSeenAt); err != nil {
			return SessionContext{}, fmt.Errorf("update session last seen from access token: %w", err)
		}
	}

	return SessionContext{
		Actor: Actor{
			OrgID:     session.OrgID,
			UserID:    session.UserID,
			SessionID: session.ID,
		},
		Session:         session,
		RoleCode:        context.RoleCode,
		OrgSlug:         context.OrgSlug,
		OrgName:         context.OrgName,
		UserEmail:       context.UserEmail,
		UserDisplayName: context.UserDisplayName,
	}, nil
}

func newRefreshToken() (string, string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	token := hex.EncodeToString(raw[:])
	return token, hashRefreshToken(token), nil
}

func newAccessToken() (string, string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	token := hex.EncodeToString(raw[:])
	return token, hashAccessToken(token), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func hashAccessToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func refreshTokenMatches(rawToken, expectedHash string) bool {
	if strings.TrimSpace(rawToken) == "" || strings.TrimSpace(expectedHash) == "" {
		return false
	}
	actual := hashRefreshToken(rawToken)
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expectedHash)) == 1
}

type rowScanner interface {
	Scan(dest ...any) error
}

type accessTokenRecord struct {
	ID                      string
	SessionID               string
	TokenHash               string
	Status                  string
	ExpiresAt               time.Time
	ReplacedByAccessTokenID sql.NullString
	IssuedAt                time.Time
	LastSeenAt              time.Time
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

func insertAccessTokenTx(ctx context.Context, tx *sql.Tx, sessionID, tokenHash string, expiresAt time.Time) (accessTokenRecord, error) {
	const insert = `
INSERT INTO identityaccess.session_access_tokens (
	session_id,
	token_hash,
	status,
	expires_at
) VALUES ($1, $2, 'active', $3)
RETURNING
	id,
	session_id,
	token_hash,
	status,
	expires_at,
	replaced_by_access_token_id,
	issued_at,
	last_seen_at;`

	var token accessTokenRecord
	if err := tx.QueryRowContext(ctx, insert, sessionID, tokenHash, expiresAt).Scan(
		&token.ID,
		&token.SessionID,
		&token.TokenHash,
		&token.Status,
		&token.ExpiresAt,
		&token.ReplacedByAccessTokenID,
		&token.IssuedAt,
		&token.LastSeenAt,
	); err != nil {
		return accessTokenRecord{}, fmt.Errorf("insert access token: %w", err)
	}

	return token, nil
}

func rotateSessionTokensTx(ctx context.Context, tx *sql.Tx, sessionID, nextRefreshTokenHash, nextAccessTokenHash string, accessTokenExpiresAt time.Time) (accessTokenRecord, error) {
	nextToken, err := insertAccessTokenTx(ctx, tx, sessionID, nextAccessTokenHash, accessTokenExpiresAt)
	if err != nil {
		return accessTokenRecord{}, err
	}

	const rotateAccessTokens = `
UPDATE identityaccess.session_access_tokens
SET status = 'rotated',
	replaced_by_access_token_id = $2,
	last_seen_at = NOW()
WHERE session_id = $1
  AND id <> $2
  AND status = 'active';`
	if _, err := tx.ExecContext(ctx, rotateAccessTokens, sessionID, nextToken.ID); err != nil {
		return accessTokenRecord{}, fmt.Errorf("rotate prior access tokens: %w", err)
	}

	const updateSession = `
UPDATE identityaccess.sessions
SET refresh_token_hash = $2,
	last_seen_at = NOW()
WHERE id = $1
  AND status = 'active'
RETURNING last_seen_at;`
	if err := tx.QueryRowContext(ctx, updateSession, sessionID, nextRefreshTokenHash).Scan(&nextToken.LastSeenAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return accessTokenRecord{}, ErrSessionNotActive
		}
		return accessTokenRecord{}, fmt.Errorf("rotate session refresh token: %w", err)
	}

	return nextToken, nil
}

func revokeSessionTx(ctx context.Context, tx *sql.Tx, sessionID string) error {
	const revokeTokens = `
UPDATE identityaccess.session_access_tokens
SET status = 'revoked',
	last_seen_at = NOW()
WHERE session_id = $1
  AND status = 'active';`
	if _, err := tx.ExecContext(ctx, revokeTokens, sessionID); err != nil {
		return fmt.Errorf("revoke session access tokens: %w", err)
	}

	const statement = `
UPDATE identityaccess.sessions
SET status = 'revoked',
	last_seen_at = NOW()
WHERE id = $1
  AND status = 'active';`

	result, err := tx.ExecContext(ctx, statement, sessionID)
	if err != nil {
		return fmt.Errorf("revoke authenticated session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke authenticated session rows affected: %w", err)
	}
	if rows == 0 {
		return ErrSessionNotActive
	}

	return nil
}
