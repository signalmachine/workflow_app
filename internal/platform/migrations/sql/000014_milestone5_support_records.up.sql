CREATE SCHEMA IF NOT EXISTS parties;

CREATE TABLE parties.parties (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	party_code TEXT NOT NULL,
	display_name TEXT NOT NULL,
	legal_name TEXT NOT NULL DEFAULT '',
	party_kind TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT parties_parties_code_not_blank CHECK (btrim(party_code) <> ''),
	CONSTRAINT parties_parties_display_name_not_blank CHECK (btrim(display_name) <> ''),
	CONSTRAINT parties_parties_legal_name_not_blank CHECK (legal_name = '' OR btrim(legal_name) <> ''),
	CONSTRAINT parties_parties_kind_allowed CHECK (party_kind IN ('customer', 'vendor', 'customer_vendor', 'other')),
	CONSTRAINT parties_parties_status_allowed CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX parties_parties_org_code_unique
	ON parties.parties (org_id, lower(party_code));

CREATE INDEX parties_parties_org_kind_created_idx
	ON parties.parties (org_id, party_kind, created_at DESC);

CREATE TABLE parties.contacts (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	party_id UUID NOT NULL REFERENCES parties.parties (id) ON DELETE CASCADE,
	full_name TEXT NOT NULL,
	role_title TEXT NOT NULL DEFAULT '',
	email TEXT,
	phone TEXT,
	is_primary BOOLEAN NOT NULL DEFAULT FALSE,
	status TEXT NOT NULL DEFAULT 'active',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT parties_contacts_full_name_not_blank CHECK (btrim(full_name) <> ''),
	CONSTRAINT parties_contacts_role_title_not_blank CHECK (role_title = '' OR btrim(role_title) <> ''),
	CONSTRAINT parties_contacts_email_not_blank CHECK (email IS NULL OR btrim(email) <> ''),
	CONSTRAINT parties_contacts_phone_not_blank CHECK (phone IS NULL OR btrim(phone) <> ''),
	CONSTRAINT parties_contacts_method_required CHECK (email IS NOT NULL OR phone IS NOT NULL),
	CONSTRAINT parties_contacts_status_allowed CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX parties_contacts_one_primary_per_party
	ON parties.contacts (party_id)
	WHERE is_primary;

CREATE INDEX parties_contacts_party_created_idx
	ON parties.contacts (party_id, created_at DESC);
