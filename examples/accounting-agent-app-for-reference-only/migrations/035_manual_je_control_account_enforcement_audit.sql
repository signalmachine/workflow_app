-- Migration 035: add enforcement/override audit fields for manual JE control-account attempts.

ALTER TABLE manual_je_control_account_audits
    ADD COLUMN IF NOT EXISTS enforcement_mode VARCHAR(20) NOT NULL DEFAULT 'warn';

ALTER TABLE manual_je_control_account_audits
    ADD COLUMN IF NOT EXISTS override_control_accounts BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE manual_je_control_account_audits
    ADD COLUMN IF NOT EXISTS override_reason TEXT;

ALTER TABLE manual_je_control_account_audits
    ADD COLUMN IF NOT EXISTS is_blocked BOOLEAN NOT NULL DEFAULT false;

DO $$ BEGIN
    ALTER TABLE manual_je_control_account_audits
        ADD CONSTRAINT chk_manual_je_control_account_audits_mode
        CHECK (enforcement_mode IN ('off', 'warn', 'enforce'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
