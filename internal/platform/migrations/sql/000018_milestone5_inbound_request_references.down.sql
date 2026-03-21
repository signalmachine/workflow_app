DROP INDEX IF EXISTS ai_inbound_requests_org_request_reference_unique;
DROP INDEX IF EXISTS ai_inbound_requests_org_request_number_unique;

ALTER TABLE ai.inbound_requests
	DROP CONSTRAINT IF EXISTS ai_inbound_requests_request_reference_format,
	DROP CONSTRAINT IF EXISTS ai_inbound_requests_request_reference_not_blank,
	DROP CONSTRAINT IF EXISTS ai_inbound_requests_request_number_positive,
	DROP COLUMN IF EXISTS request_reference,
	DROP COLUMN IF EXISTS request_number;

DROP TABLE IF EXISTS ai.inbound_request_numbering_series;
