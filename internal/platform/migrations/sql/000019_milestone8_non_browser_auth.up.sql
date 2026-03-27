CREATE TABLE identityaccess.session_access_tokens (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	session_id UUID NOT NULL REFERENCES identityaccess.sessions (id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL,
	status TEXT NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	replaced_by_access_token_id UUID REFERENCES identityaccess.session_access_tokens (id) ON DELETE RESTRICT,
	issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT identityaccess_session_access_tokens_token_hash_not_blank CHECK (btrim(token_hash) <> ''),
	CONSTRAINT identityaccess_session_access_tokens_status_allowed CHECK (status IN ('active', 'revoked', 'rotated', 'expired'))
);

CREATE UNIQUE INDEX identityaccess_session_access_tokens_token_hash_idx
	ON identityaccess.session_access_tokens (token_hash);

CREATE INDEX identityaccess_session_access_tokens_session_status_idx
	ON identityaccess.session_access_tokens (session_id, status, expires_at DESC);
