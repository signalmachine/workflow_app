DROP INDEX IF EXISTS ai_agent_runs_inbound_request_idx;

ALTER TABLE ai.agent_runs
	DROP CONSTRAINT IF EXISTS ai_agent_runs_inbound_request_fk;

ALTER TABLE ai.agent_runs
	DROP COLUMN IF EXISTS inbound_request_id;

DROP TABLE IF EXISTS attachments.derived_texts;
DROP TABLE IF EXISTS attachments.request_message_links;
DROP TABLE IF EXISTS attachments.attachments;
DROP TABLE IF EXISTS ai.inbound_request_messages;
DROP TABLE IF EXISTS ai.inbound_requests;

DROP INDEX IF EXISTS ai_agent_runs_org_id_unique;
DROP INDEX IF EXISTS identityaccess_sessions_org_id_unique;
