-- Migration 030: Accounting foundations hardening (issues 2/4/6/7)

-- ---------------------------------------------------------------------------
-- Issue 4: DB-level ledger invariants on journal_lines
-- ---------------------------------------------------------------------------
ALTER TABLE journal_lines DROP CONSTRAINT IF EXISTS chk_journal_lines_non_negative;
ALTER TABLE journal_lines
    ADD CONSTRAINT chk_journal_lines_non_negative
    CHECK (debit_base >= 0 AND credit_base >= 0);

ALTER TABLE journal_lines DROP CONSTRAINT IF EXISTS chk_journal_lines_one_sided_positive;
ALTER TABLE journal_lines
    ADD CONSTRAINT chk_journal_lines_one_sided_positive
    CHECK (
        (debit_base > 0 AND credit_base = 0)
        OR
        (credit_base > 0 AND debit_base = 0)
    );

-- ---------------------------------------------------------------------------
-- Issue 6: account_rules temporal model consistency
-- ---------------------------------------------------------------------------
DROP INDEX IF EXISTS idx_account_rules_lookup;

CREATE UNIQUE INDEX IF NOT EXISTS idx_account_rules_lookup_temporal
    ON account_rules (
        company_id,
        rule_type,
        COALESCE(qualifier_key, ''),
        COALESCE(qualifier_value, ''),
        effective_from
    );

-- ---------------------------------------------------------------------------
-- Issue 7: Add account-id references with FK integrity (safe rollout)
-- ---------------------------------------------------------------------------
ALTER TABLE products ADD COLUMN IF NOT EXISTS revenue_account_id INT;
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS ap_account_id INT;
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS default_expense_account_id INT;
ALTER TABLE purchase_order_lines ADD COLUMN IF NOT EXISTS expense_account_id INT;
ALTER TABLE account_rules ADD COLUMN IF NOT EXISTS account_id INT;

-- Backfill account IDs from existing code columns (company scoped).
UPDATE products p
SET revenue_account_id = a.id
FROM accounts a
WHERE p.revenue_account_id IS NULL
  AND a.company_id = p.company_id
  AND a.code = p.revenue_account_code;

UPDATE vendors v
SET ap_account_id = a.id
FROM accounts a
WHERE v.ap_account_id IS NULL
  AND a.company_id = v.company_id
  AND a.code = v.ap_account_code;

UPDATE vendors v
SET default_expense_account_id = a.id
FROM accounts a
WHERE v.default_expense_account_id IS NULL
  AND v.default_expense_account_code IS NOT NULL
  AND a.company_id = v.company_id
  AND a.code = v.default_expense_account_code;

UPDATE purchase_order_lines pol
SET expense_account_id = a.id
FROM purchase_orders po, accounts a
WHERE pol.order_id = po.id
  AND a.company_id = po.company_id
  AND a.code = pol.expense_account_code
  AND pol.expense_account_id IS NULL
  AND pol.expense_account_code IS NOT NULL;

UPDATE account_rules ar
SET account_id = a.id
FROM accounts a
WHERE ar.account_id IS NULL
  AND a.company_id = ar.company_id
  AND a.code = ar.account_code;

-- Add FKs.
DO $$ BEGIN
    ALTER TABLE products
        ADD CONSTRAINT fk_products_revenue_account_id
        FOREIGN KEY (revenue_account_id) REFERENCES accounts(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE vendors
        ADD CONSTRAINT fk_vendors_ap_account_id
        FOREIGN KEY (ap_account_id) REFERENCES accounts(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE vendors
        ADD CONSTRAINT fk_vendors_default_expense_account_id
        FOREIGN KEY (default_expense_account_id) REFERENCES accounts(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE purchase_order_lines
        ADD CONSTRAINT fk_pol_expense_account_id
        FOREIGN KEY (expense_account_id) REFERENCES accounts(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE account_rules
        ADD CONSTRAINT fk_account_rules_account_id
        FOREIGN KEY (account_id) REFERENCES accounts(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
