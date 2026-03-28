-- Migration 041: persistent audit trail for document type policy violations.

CREATE TABLE IF NOT EXISTS document_type_policy_violation_audits (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    source VARCHAR(30) NOT NULL,
    policy_mode VARCHAR(20) NOT NULL,
    intent_code VARCHAR(40) NOT NULL,
    document_type_code VARCHAR(10) NOT NULL,
    idempotency_key VARCHAR(120) NOT NULL DEFAULT '',
    violation_message TEXT NOT NULL,
    is_enforced BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$ BEGIN
    ALTER TABLE document_type_policy_violation_audits
        ADD CONSTRAINT chk_document_type_policy_violation_mode
        CHECK (policy_mode IN ('warn', 'enforce'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_document_type_policy_violation_audits_company_created
    ON document_type_policy_violation_audits(company_id, created_at DESC);
