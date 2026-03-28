-- Migration 033: control-account metadata and backfill from active account rules.

ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS is_control_account BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS control_type VARCHAR(20);

DO $$ BEGIN
    ALTER TABLE accounts
        ADD CONSTRAINT chk_accounts_control_type
        CHECK (control_type IN ('AR', 'AP', 'INVENTORY') OR control_type IS NULL);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE accounts
        ADD CONSTRAINT chk_accounts_control_flag_consistency
        CHECK (is_control_account OR control_type IS NULL);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Backfill control-account flags from active AR/AP/INVENTORY rules.
WITH active_control_rules AS (
    SELECT DISTINCT ON (ar.company_id, ar.account_code)
        ar.company_id,
        ar.account_code,
        ar.rule_type
    FROM account_rules ar
    WHERE ar.rule_type IN ('AR', 'AP', 'INVENTORY')
      AND ar.effective_from <= CURRENT_DATE
      AND (ar.effective_to IS NULL OR ar.effective_to >= CURRENT_DATE)
    ORDER BY ar.company_id, ar.account_code, ar.priority DESC, ar.effective_from DESC, ar.id DESC
)
UPDATE accounts a
SET is_control_account = true,
    control_type = acr.rule_type
FROM active_control_rules acr
WHERE a.company_id = acr.company_id
  AND a.code = acr.account_code
  AND (
      a.is_control_account IS DISTINCT FROM true
      OR a.control_type IS DISTINCT FROM acr.rule_type
  );
