CREATE TABLE identityaccess.sessions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	membership_id UUID NOT NULL REFERENCES identityaccess.memberships (id) ON DELETE RESTRICT,
	device_label TEXT NOT NULL,
	refresh_token_hash TEXT NOT NULL,
	status TEXT NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	replaced_by_session_id UUID REFERENCES identityaccess.sessions (id) ON DELETE RESTRICT,
	issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT identityaccess_sessions_device_label_not_blank CHECK (btrim(device_label) <> ''),
	CONSTRAINT identityaccess_sessions_refresh_token_hash_not_blank CHECK (btrim(refresh_token_hash) <> ''),
	CONSTRAINT identityaccess_sessions_status_allowed CHECK (status IN ('active', 'revoked', 'rotated', 'expired'))
);

CREATE INDEX identityaccess_sessions_org_user_status_idx
	ON identityaccess.sessions (org_id, user_id, status, expires_at DESC);

CREATE INDEX identityaccess_sessions_membership_idx
	ON identityaccess.sessions (membership_id, status, expires_at DESC);
