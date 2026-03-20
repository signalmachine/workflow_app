DROP INDEX IF EXISTS inventory_ops.inventory_accounting_handoffs_journal_idx;

ALTER TABLE inventory_ops.accounting_handoffs
	DROP CONSTRAINT IF EXISTS inventory_accounting_handoffs_cost_snapshot_consistent;

ALTER TABLE inventory_ops.accounting_handoffs
	DROP CONSTRAINT IF EXISTS inventory_accounting_handoffs_posted_consistent;

ALTER TABLE inventory_ops.accounting_handoffs
	ADD CONSTRAINT inventory_accounting_handoffs_posted_consistent CHECK (
		(handoff_status = 'pending' AND journal_entry_id IS NULL)
		OR (handoff_status = 'posted' AND journal_entry_id IS NOT NULL)
	);

ALTER TABLE inventory_ops.accounting_handoffs
	DROP COLUMN IF EXISTS posted_at,
	DROP COLUMN IF EXISTS cost_currency_code,
	DROP COLUMN IF EXISTS cost_minor;
