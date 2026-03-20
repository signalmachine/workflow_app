DROP INDEX IF EXISTS accounting.accounting_journal_entries_org_effective_on_idx;

ALTER TABLE accounting.journal_entries
DROP COLUMN IF EXISTS effective_on;

DROP INDEX IF EXISTS accounting.accounting_periods_org_dates_idx;
DROP INDEX IF EXISTS accounting.accounting_periods_org_code_unique;

DROP TABLE IF EXISTS accounting.periods;
