-- Drop obsolete materialized views in favor of direct journal_lines queries.
DROP MATERIALIZED VIEW IF EXISTS mv_trial_balance;
DROP MATERIALIZED VIEW IF EXISTS mv_account_period_balances;

-- Support direct trial-balance aggregation queries.
CREATE INDEX IF NOT EXISTS idx_journal_lines_account_id
    ON journal_lines (account_id);

CREATE INDEX IF NOT EXISTS idx_journal_entries_company_posting
    ON journal_entries (company_id, posting_date);
