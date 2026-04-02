package parties

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
)

var (
	ErrPartyNotFound   = errors.New("party not found")
	ErrContactNotFound = errors.New("contact not found")
	ErrInvalidParty    = errors.New("invalid party")
	ErrInvalidContact  = errors.New("invalid contact")
)

const (
	PartyKindCustomer       = "customer"
	PartyKindVendor         = "vendor"
	PartyKindCustomerVendor = "customer_vendor"
	PartyKindOther          = "other"

	StatusActive   = "active"
	StatusInactive = "inactive"
)

type Party struct {
	ID              string
	OrgID           string
	PartyCode       string
	DisplayName     string
	LegalName       string
	PartyKind       string
	Status          string
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Contact struct {
	ID              string
	OrgID           string
	PartyID         string
	FullName        string
	RoleTitle       string
	Email           sql.NullString
	Phone           sql.NullString
	IsPrimary       bool
	Status          string
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreatePartyInput struct {
	PartyCode   string
	DisplayName string
	LegalName   string
	PartyKind   string
	Actor       identityaccess.Actor
}

type CreateContactInput struct {
	PartyID   string
	FullName  string
	RoleTitle string
	Email     string
	Phone     string
	IsPrimary bool
	Actor     identityaccess.Actor
}

type ListPartiesInput struct {
	PartyKind string
	Actor     identityaccess.Actor
}

type ListContactsInput struct {
	PartyID string
	Actor   identityaccess.Actor
}

type GetPartyInput struct {
	PartyID string
	Actor   identityaccess.Actor
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateParty(ctx context.Context, input CreatePartyInput) (Party, error) {
	partyKind := strings.TrimSpace(input.PartyKind)
	if strings.TrimSpace(input.PartyCode) == "" || strings.TrimSpace(input.DisplayName) == "" || !isValidPartyKind(partyKind) {
		return Party{}, ErrInvalidParty
	}
	if input.LegalName != "" && strings.TrimSpace(input.LegalName) == "" {
		return Party{}, ErrInvalidParty
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Party{}, fmt.Errorf("begin create party: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Party{}, err
	}

	party, err := scanParty(tx.QueryRowContext(ctx, `
INSERT INTO parties.parties (
	org_id,
	party_code,
	display_name,
	legal_name,
	party_kind,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
	id,
	org_id,
	party_code,
	display_name,
	legal_name,
	party_kind,
	status,
	created_by_user_id,
	created_at,
	updated_at;`,
		input.Actor.OrgID,
		strings.TrimSpace(input.PartyCode),
		strings.TrimSpace(input.DisplayName),
		strings.TrimSpace(input.LegalName),
		partyKind,
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return Party{}, fmt.Errorf("insert party: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "parties.party_created",
		EntityType:  "parties.party",
		EntityID:    party.ID,
		Payload: map[string]any{
			"party_code": party.PartyCode,
			"party_kind": party.PartyKind,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Party{}, err
	}

	if err := tx.Commit(); err != nil {
		return Party{}, fmt.Errorf("commit create party: %w", err)
	}

	return party, nil
}

func (s *Service) CreateContact(ctx context.Context, input CreateContactInput) (Contact, error) {
	if strings.TrimSpace(input.PartyID) == "" || strings.TrimSpace(input.FullName) == "" {
		return Contact{}, ErrInvalidContact
	}
	if input.RoleTitle != "" && strings.TrimSpace(input.RoleTitle) == "" {
		return Contact{}, ErrInvalidContact
	}

	email := strings.TrimSpace(input.Email)
	phone := strings.TrimSpace(input.Phone)
	if email == "" && phone == "" {
		return Contact{}, ErrInvalidContact
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Contact{}, fmt.Errorf("begin create contact: %w", err)
	}

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator); err != nil {
		_ = tx.Rollback()
		return Contact{}, err
	}

	party, err := getPartyForUpdate(ctx, tx, input.Actor.OrgID, input.PartyID)
	if err != nil {
		_ = tx.Rollback()
		return Contact{}, err
	}
	if party.Status != StatusActive {
		_ = tx.Rollback()
		return Contact{}, ErrInvalidContact
	}

	if input.IsPrimary {
		if _, err := tx.ExecContext(ctx, `
UPDATE parties.contacts
SET is_primary = FALSE,
	updated_at = NOW()
WHERE org_id = $1
  AND party_id = $2
  AND is_primary = TRUE`,
			input.Actor.OrgID,
			party.ID,
		); err != nil {
			_ = tx.Rollback()
			return Contact{}, fmt.Errorf("clear existing primary contact: %w", err)
		}
	}

	contact, err := scanContact(tx.QueryRowContext(ctx, `
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
RETURNING
	id,
	org_id,
	party_id,
	full_name,
	role_title,
	email,
	phone,
	is_primary,
	status,
	created_by_user_id,
	created_at,
	updated_at;`,
		input.Actor.OrgID,
		party.ID,
		strings.TrimSpace(input.FullName),
		strings.TrimSpace(input.RoleTitle),
		nullIfEmpty(email),
		nullIfEmpty(phone),
		input.IsPrimary,
		input.Actor.UserID,
	))
	if err != nil {
		_ = tx.Rollback()
		return Contact{}, fmt.Errorf("insert contact: %w", err)
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "parties.contact_created",
		EntityType:  "parties.contact",
		EntityID:    contact.ID,
		Payload: map[string]any{
			"party_id":   contact.PartyID,
			"is_primary": contact.IsPrimary,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Contact{}, err
	}

	if err := tx.Commit(); err != nil {
		return Contact{}, fmt.Errorf("commit create contact: %w", err)
	}

	return contact, nil
}

func (s *Service) ListParties(ctx context.Context, input ListPartiesInput) ([]Party, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list parties: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator, identityaccess.RoleApprover); err != nil {
		return nil, err
	}

	partyKind := strings.TrimSpace(input.PartyKind)
	if partyKind != "" && !isValidPartyKind(partyKind) {
		return nil, ErrInvalidParty
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
	id,
	org_id,
	party_code,
	display_name,
	legal_name,
	party_kind,
	status,
	created_by_user_id,
	created_at,
	updated_at
FROM parties.parties
WHERE org_id = $1
  AND ($2 = '' OR party_kind = $2)
ORDER BY created_at DESC, id DESC`,
		input.Actor.OrgID,
		partyKind,
	)
	if err != nil {
		return nil, fmt.Errorf("query parties: %w", err)
	}
	defer rows.Close()

	parties := make([]Party, 0)
	for rows.Next() {
		party, err := scanParty(rows)
		if err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate parties: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list parties: %w", err)
	}

	return parties, nil
}

func (s *Service) ListContacts(ctx context.Context, input ListContactsInput) ([]Contact, error) {
	if strings.TrimSpace(input.PartyID) == "" {
		return nil, ErrInvalidContact
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list contacts: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator, identityaccess.RoleApprover); err != nil {
		return nil, err
	}

	if _, err := getParty(ctx, tx, input.Actor.OrgID, input.PartyID); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
	id,
	org_id,
	party_id,
	full_name,
	role_title,
	email,
	phone,
	is_primary,
	status,
	created_by_user_id,
	created_at,
	updated_at
FROM parties.contacts
WHERE org_id = $1
  AND party_id = $2
ORDER BY is_primary DESC, created_at DESC, id DESC`,
		input.Actor.OrgID,
		input.PartyID,
	)
	if err != nil {
		return nil, fmt.Errorf("query contacts: %w", err)
	}
	defer rows.Close()

	contacts := make([]Contact, 0)
	for rows.Next() {
		contact, err := scanContact(rows)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate contacts: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list contacts: %w", err)
	}

	return contacts, nil
}

func (s *Service) GetParty(ctx context.Context, input GetPartyInput) (Party, error) {
	if strings.TrimSpace(input.PartyID) == "" {
		return Party{}, ErrInvalidParty
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Party{}, fmt.Errorf("begin get party: %w", err)
	}
	defer tx.Rollback()

	if err := identityaccess.AuthorizeTx(ctx, tx, input.Actor, identityaccess.RoleAdmin, identityaccess.RoleOperator, identityaccess.RoleApprover); err != nil {
		return Party{}, err
	}

	party, err := getParty(ctx, tx, input.Actor.OrgID, strings.TrimSpace(input.PartyID))
	if err != nil {
		return Party{}, err
	}

	if err := tx.Commit(); err != nil {
		return Party{}, fmt.Errorf("commit get party: %w", err)
	}

	return party, nil
}

func getPartyForUpdate(ctx context.Context, tx *sql.Tx, orgID, partyID string) (Party, error) {
	return scanParty(tx.QueryRowContext(ctx, `
SELECT
	id,
	org_id,
	party_code,
	display_name,
	legal_name,
	party_kind,
	status,
	created_by_user_id,
	created_at,
	updated_at
FROM parties.parties
WHERE org_id = $1
  AND id = $2
FOR UPDATE`,
		orgID,
		partyID,
	))
}

func getParty(ctx context.Context, tx *sql.Tx, orgID, partyID string) (Party, error) {
	return scanParty(tx.QueryRowContext(ctx, `
SELECT
	id,
	org_id,
	party_code,
	display_name,
	legal_name,
	party_kind,
	status,
	created_by_user_id,
	created_at,
	updated_at
FROM parties.parties
WHERE org_id = $1
  AND id = $2`,
		orgID,
		partyID,
	))
}

func scanParty(scanner interface {
	Scan(dest ...any) error
}) (Party, error) {
	var party Party
	if err := scanner.Scan(
		&party.ID,
		&party.OrgID,
		&party.PartyCode,
		&party.DisplayName,
		&party.LegalName,
		&party.PartyKind,
		&party.Status,
		&party.CreatedByUserID,
		&party.CreatedAt,
		&party.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Party{}, ErrPartyNotFound
		}
		return Party{}, err
	}
	return party, nil
}

func scanContact(scanner interface {
	Scan(dest ...any) error
}) (Contact, error) {
	var contact Contact
	if err := scanner.Scan(
		&contact.ID,
		&contact.OrgID,
		&contact.PartyID,
		&contact.FullName,
		&contact.RoleTitle,
		&contact.Email,
		&contact.Phone,
		&contact.IsPrimary,
		&contact.Status,
		&contact.CreatedByUserID,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Contact{}, ErrContactNotFound
		}
		return Contact{}, err
	}
	return contact, nil
}

func isValidPartyKind(kind string) bool {
	switch kind {
	case PartyKindCustomer, PartyKindVendor, PartyKindCustomerVendor, PartyKindOther:
		return true
	default:
		return false
	}
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}
