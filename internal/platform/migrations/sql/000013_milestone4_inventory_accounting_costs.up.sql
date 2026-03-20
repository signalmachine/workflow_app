ALTER TABLE inventory_ops.accounting_handoffs
	ADD COLUMN cost_minor BIGINT,
	ADD COLUMN cost_currency_code TEXT,
	ADD COLUMN posted_at TIMESTAMPTZ;

ALTER TABLE inventory_ops.accounting_handoffs
	DROP CONSTRAINT inventory_accounting_handoffs_posted_consistent;

ALTER TABLE inventory_ops.accounting_handoffs
	ADD CONSTRAINT inventory_accounting_handoffs_posted_consistent CHECK (
		(handoff_status = 'pending' AND journal_entry_id IS NULL AND posted_at IS NULL)
		OR (handoff_status = 'posted' AND journal_entry_id IS NOT NULL AND posted_at IS NOT NULL)
	);

ALTER TABLE inventory_ops.accounting_handoffs
	ADD CONSTRAINT inventory_accounting_handoffs_cost_snapshot_consistent CHECK (
		(cost_minor IS NULL AND cost_currency_code IS NULL)
		OR (cost_minor IS NOT NULL AND cost_minor > 0 AND cost_currency_code IS NOT NULL AND btrim(cost_currency_code) <> '')
	);

CREATE INDEX inventory_accounting_handoffs_journal_idx
	ON inventory_ops.accounting_handoffs (journal_entry_id)
	WHERE journal_entry_id IS NOT NULL;
