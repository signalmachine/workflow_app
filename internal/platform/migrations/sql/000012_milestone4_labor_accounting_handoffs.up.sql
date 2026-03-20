CREATE TABLE workforce.labor_accounting_handoffs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	labor_entry_id UUID NOT NULL UNIQUE REFERENCES workforce.labor_entries (id) ON DELETE RESTRICT,
	work_order_id UUID NOT NULL REFERENCES work_orders.work_orders (id) ON DELETE RESTRICT,
	task_id UUID REFERENCES workflow.tasks (id) ON DELETE RESTRICT,
	journal_entry_id UUID REFERENCES accounting.journal_entries (id) ON DELETE RESTRICT,
	handoff_status TEXT NOT NULL DEFAULT 'pending',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	posted_at TIMESTAMPTZ,
	CONSTRAINT workforce_labor_accounting_handoffs_status_allowed CHECK (handoff_status IN ('pending', 'posted')),
	CONSTRAINT workforce_labor_accounting_handoffs_posted_consistent CHECK (
		(handoff_status = 'pending' AND journal_entry_id IS NULL AND posted_at IS NULL)
		OR (handoff_status = 'posted' AND journal_entry_id IS NOT NULL AND posted_at IS NOT NULL)
	)
);

CREATE INDEX workforce_labor_accounting_handoffs_org_work_order_status_idx
	ON workforce.labor_accounting_handoffs (org_id, work_order_id, handoff_status, created_at DESC);

CREATE INDEX workforce_labor_accounting_handoffs_journal_idx
	ON workforce.labor_accounting_handoffs (journal_entry_id)
	WHERE journal_entry_id IS NOT NULL;
