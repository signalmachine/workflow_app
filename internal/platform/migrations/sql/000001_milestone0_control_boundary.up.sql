CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS identityaccess;
CREATE SCHEMA IF NOT EXISTS workflow;
CREATE SCHEMA IF NOT EXISTS documents;
CREATE SCHEMA IF NOT EXISTS accounting;
CREATE SCHEMA IF NOT EXISTS inventory_ops;
CREATE SCHEMA IF NOT EXISTS workforce;
CREATE SCHEMA IF NOT EXISTS work_orders;
CREATE SCHEMA IF NOT EXISTS attachments;
CREATE SCHEMA IF NOT EXISTS reporting;

CREATE TABLE identityaccess.orgs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	slug TEXT NOT NULL,
	name TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT identityaccess_orgs_slug_not_blank CHECK (btrim(slug) <> ''),
	CONSTRAINT identityaccess_orgs_name_not_blank CHECK (btrim(name) <> ''),
	CONSTRAINT identityaccess_orgs_status_allowed CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX identityaccess_orgs_slug_unique
	ON identityaccess.orgs (lower(slug));

CREATE TABLE identityaccess.users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email TEXT NOT NULL,
	display_name TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT identityaccess_users_email_not_blank CHECK (btrim(email) <> ''),
	CONSTRAINT identityaccess_users_display_name_not_blank CHECK (btrim(display_name) <> ''),
	CONSTRAINT identityaccess_users_status_allowed CHECK (status IN ('active', 'disabled'))
);

CREATE UNIQUE INDEX identityaccess_users_email_unique
	ON identityaccess.users (lower(email));

CREATE TABLE identityaccess.memberships (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	role_code TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT identityaccess_memberships_role_code_not_blank CHECK (btrim(role_code) <> ''),
	CONSTRAINT identityaccess_memberships_status_allowed CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX identityaccess_memberships_org_user_unique
	ON identityaccess.memberships (org_id, user_id);

CREATE TABLE platform.audit_events (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	actor_user_id UUID REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	event_type TEXT NOT NULL,
	entity_type TEXT NOT NULL,
	entity_id TEXT NOT NULL,
	payload JSONB NOT NULL DEFAULT '{}'::jsonb,
	occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT platform_audit_events_event_type_not_blank CHECK (btrim(event_type) <> ''),
	CONSTRAINT platform_audit_events_entity_type_not_blank CHECK (btrim(entity_type) <> ''),
	CONSTRAINT platform_audit_events_entity_id_not_blank CHECK (btrim(entity_id) <> '')
);

CREATE INDEX platform_audit_events_org_occurred_idx
	ON platform.audit_events (org_id, occurred_at DESC);

CREATE INDEX platform_audit_events_entity_idx
	ON platform.audit_events (entity_type, entity_id, occurred_at DESC);

CREATE TABLE platform.idempotency_keys (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	idempotency_key TEXT NOT NULL,
	request_fingerprint TEXT NOT NULL,
	response_code INT,
	response_body JSONB,
	first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT platform_idempotency_keys_key_not_blank CHECK (btrim(idempotency_key) <> ''),
	CONSTRAINT platform_idempotency_keys_fingerprint_not_blank CHECK (btrim(request_fingerprint) <> '')
);

CREATE UNIQUE INDEX platform_idempotency_keys_org_key_unique
	ON platform.idempotency_keys (org_id, idempotency_key);
