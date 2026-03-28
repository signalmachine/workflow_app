-- Migration 034: audit trail for manual JE control-account posting attempts.

CREATE TABLE IF NOT EXISTS manual_je_control_account_audits (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(100),
    action VARCHAR(20) NOT NULL,
    posting_date DATE,
    narration TEXT,
    account_codes TEXT[] NOT NULL DEFAULT '{}',
    warning_details JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$ BEGIN
    ALTER TABLE manual_je_control_account_audits
        ADD CONSTRAINT chk_manual_je_control_account_audits_action
        CHECK (action IN ('validate', 'post'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_manual_je_control_account_audits_company_created
    ON manual_je_control_account_audits(company_id, created_at DESC);
