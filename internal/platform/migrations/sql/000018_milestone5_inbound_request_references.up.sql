CREATE TABLE ai.inbound_request_numbering_series (
	org_id UUID PRIMARY KEY REFERENCES identityaccess.orgs (id) ON DELETE CASCADE,
	next_number BIGINT NOT NULL DEFAULT 1,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT ai_inbound_request_numbering_series_next_number_positive CHECK (next_number > 0)
);

ALTER TABLE ai.inbound_requests
	ADD COLUMN request_number BIGINT,
	ADD COLUMN request_reference TEXT;

WITH existing AS (
	SELECT
		id,
		org_id,
		ROW_NUMBER() OVER (PARTITION BY org_id ORDER BY created_at ASC, id ASC) AS request_number
	FROM ai.inbound_requests
)
UPDATE ai.inbound_requests r
SET request_number = existing.request_number,
	request_reference = 'REQ-' || LPAD(existing.request_number::text, 6, '0')
FROM existing
WHERE existing.id = r.id;

INSERT INTO ai.inbound_request_numbering_series (org_id, next_number)
SELECT org_id, COALESCE(MAX(request_number), 0) + 1
FROM ai.inbound_requests
GROUP BY org_id
ON CONFLICT (org_id) DO UPDATE
SET next_number = EXCLUDED.next_number,
	updated_at = NOW();

ALTER TABLE ai.inbound_requests
	ALTER COLUMN request_number SET NOT NULL,
	ALTER COLUMN request_reference SET NOT NULL;

ALTER TABLE ai.inbound_requests
	ADD CONSTRAINT ai_inbound_requests_request_number_positive CHECK (request_number > 0),
	ADD CONSTRAINT ai_inbound_requests_request_reference_not_blank CHECK (btrim(request_reference) <> ''),
	ADD CONSTRAINT ai_inbound_requests_request_reference_format CHECK (request_reference ~ '^REQ-[0-9]{6,}$');

CREATE UNIQUE INDEX ai_inbound_requests_org_request_number_unique
	ON ai.inbound_requests (org_id, request_number);

CREATE UNIQUE INDEX ai_inbound_requests_org_request_reference_unique
	ON ai.inbound_requests (org_id, lower(request_reference));
