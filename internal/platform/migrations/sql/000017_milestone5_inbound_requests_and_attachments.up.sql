CREATE UNIQUE INDEX IF NOT EXISTS identityaccess_sessions_org_id_unique
	ON identityaccess.sessions (org_id, id);

CREATE UNIQUE INDEX IF NOT EXISTS ai_agent_runs_org_id_unique
	ON ai.agent_runs (org_id, id);

CREATE TABLE ai.inbound_requests (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	session_id UUID,
	actor_user_id UUID REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	origin_type TEXT NOT NULL,
	channel TEXT NOT NULL,
	status TEXT NOT NULL,
	metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
	cancellation_reason TEXT NOT NULL DEFAULT '',
	failure_reason TEXT NOT NULL DEFAULT '',
	received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	queued_at TIMESTAMPTZ,
	processing_started_at TIMESTAMPTZ,
	processed_at TIMESTAMPTZ,
	acted_on_at TIMESTAMPTZ,
	completed_at TIMESTAMPTZ,
	failed_at TIMESTAMPTZ,
	cancelled_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT ai_inbound_requests_session_fk
		FOREIGN KEY (org_id, session_id) REFERENCES identityaccess.sessions (org_id, id) ON DELETE RESTRICT,
	CONSTRAINT ai_inbound_requests_origin_type_allowed CHECK (origin_type IN ('human', 'system')),
	CONSTRAINT ai_inbound_requests_channel_not_blank CHECK (btrim(channel) <> ''),
	CONSTRAINT ai_inbound_requests_status_allowed CHECK (
		status IN ('draft', 'queued', 'processing', 'processed', 'acted_on', 'completed', 'failed', 'cancelled')
	),
	CONSTRAINT ai_inbound_requests_cancellation_reason_not_blank CHECK (
		cancellation_reason = '' OR btrim(cancellation_reason) <> ''
	),
	CONSTRAINT ai_inbound_requests_failure_reason_not_blank CHECK (
		failure_reason = '' OR btrim(failure_reason) <> ''
	)
);

CREATE UNIQUE INDEX ai_inbound_requests_org_id_unique
	ON ai.inbound_requests (org_id, id);

CREATE INDEX ai_inbound_requests_org_status_queue_idx
	ON ai.inbound_requests (org_id, status, queued_at NULLS LAST, received_at DESC);

CREATE TABLE ai.inbound_request_messages (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	request_id UUID NOT NULL,
	message_index INT NOT NULL,
	message_role TEXT NOT NULL,
	text_content TEXT NOT NULL DEFAULT '',
	created_by_user_id UUID REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT ai_inbound_request_messages_request_fk
		FOREIGN KEY (org_id, request_id) REFERENCES ai.inbound_requests (org_id, id) ON DELETE CASCADE,
	CONSTRAINT ai_inbound_request_messages_message_index_positive CHECK (message_index > 0),
	CONSTRAINT ai_inbound_request_messages_message_role_allowed CHECK (
		message_role IN ('request', 'system', 'assistant', 'transcription')
	)
);

CREATE UNIQUE INDEX ai_inbound_request_messages_org_id_unique
	ON ai.inbound_request_messages (org_id, id);

CREATE UNIQUE INDEX ai_inbound_request_messages_request_index_unique
	ON ai.inbound_request_messages (request_id, message_index);

CREATE TABLE attachments.attachments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	storage_backend TEXT NOT NULL,
	storage_locator TEXT NOT NULL,
	original_file_name TEXT NOT NULL,
	media_type TEXT NOT NULL,
	size_bytes BIGINT NOT NULL,
	checksum_sha256 TEXT NOT NULL,
	content BYTEA,
	uploaded_by_user_id UUID REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT attachments_attachments_storage_backend_allowed CHECK (storage_backend IN ('postgres')),
	CONSTRAINT attachments_attachments_storage_locator_not_blank CHECK (btrim(storage_locator) <> ''),
	CONSTRAINT attachments_attachments_original_file_name_not_blank CHECK (btrim(original_file_name) <> ''),
	CONSTRAINT attachments_attachments_media_type_not_blank CHECK (btrim(media_type) <> ''),
	CONSTRAINT attachments_attachments_size_bytes_nonnegative CHECK (size_bytes >= 0),
	CONSTRAINT attachments_attachments_checksum_sha256_valid CHECK (checksum_sha256 ~ '^[a-f0-9]{64}$'),
	CONSTRAINT attachments_attachments_content_required_for_postgres CHECK (
		(storage_backend = 'postgres' AND content IS NOT NULL)
	)
);

CREATE UNIQUE INDEX attachments_attachments_org_id_unique
	ON attachments.attachments (org_id, id);

CREATE INDEX attachments_attachments_org_created_idx
	ON attachments.attachments (org_id, created_at DESC);

CREATE TABLE attachments.request_message_links (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	request_message_id UUID NOT NULL,
	attachment_id UUID NOT NULL,
	link_role TEXT NOT NULL DEFAULT 'source',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT attachments_request_message_links_message_fk
		FOREIGN KEY (org_id, request_message_id) REFERENCES ai.inbound_request_messages (org_id, id) ON DELETE CASCADE,
	CONSTRAINT attachments_request_message_links_attachment_fk
		FOREIGN KEY (org_id, attachment_id) REFERENCES attachments.attachments (org_id, id) ON DELETE CASCADE,
	CONSTRAINT attachments_request_message_links_link_role_allowed CHECK (link_role IN ('source', 'evidence'))
);

CREATE UNIQUE INDEX attachments_request_message_links_message_attachment_unique
	ON attachments.request_message_links (request_message_id, attachment_id);

CREATE INDEX attachments_request_message_links_org_message_idx
	ON attachments.request_message_links (org_id, request_message_id, created_at DESC);

CREATE TABLE attachments.derived_texts (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	source_attachment_id UUID NOT NULL,
	request_message_id UUID,
	created_by_run_id UUID,
	derivative_type TEXT NOT NULL,
	content_text TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT attachments_derived_texts_attachment_fk
		FOREIGN KEY (org_id, source_attachment_id) REFERENCES attachments.attachments (org_id, id) ON DELETE CASCADE,
	CONSTRAINT attachments_derived_texts_message_fk
		FOREIGN KEY (org_id, request_message_id) REFERENCES ai.inbound_request_messages (org_id, id) ON DELETE SET NULL,
	CONSTRAINT attachments_derived_texts_run_fk
		FOREIGN KEY (org_id, created_by_run_id) REFERENCES ai.agent_runs (org_id, id) ON DELETE SET NULL,
	CONSTRAINT attachments_derived_texts_derivative_type_allowed CHECK (derivative_type IN ('transcription')),
	CONSTRAINT attachments_derived_texts_content_text_not_blank CHECK (btrim(content_text) <> '')
);

CREATE INDEX attachments_derived_texts_org_attachment_created_idx
	ON attachments.derived_texts (org_id, source_attachment_id, created_at DESC);

ALTER TABLE ai.agent_runs
	ADD COLUMN inbound_request_id UUID;

ALTER TABLE ai.agent_runs
	ADD CONSTRAINT ai_agent_runs_inbound_request_fk
	FOREIGN KEY (org_id, inbound_request_id) REFERENCES ai.inbound_requests (org_id, id) ON DELETE SET NULL;

CREATE INDEX ai_agent_runs_inbound_request_idx
	ON ai.agent_runs (org_id, inbound_request_id, started_at DESC);
