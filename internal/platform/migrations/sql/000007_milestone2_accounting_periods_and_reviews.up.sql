CREATE TABLE accounting.periods (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	period_code TEXT NOT NULL,
	start_on DATE NOT NULL,
	end_on DATE NOT NULL,
	status TEXT NOT NULL DEFAULT 'open',
	closed_by_user_id UUID REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	closed_at TIMESTAMPTZ,
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounting_periods_code_not_blank CHECK (btrim(period_code) <> ''),
	CONSTRAINT accounting_periods_date_range_valid CHECK (end_on >= start_on),
	CONSTRAINT accounting_periods_status_allowed CHECK (status IN ('open', 'closed')),
	CONSTRAINT accounting_periods_closed_state_consistent CHECK (
		(status = 'open' AND closed_by_user_id IS NULL AND closed_at IS NULL)
		OR (status = 'closed' AND closed_by_user_id IS NOT NULL AND closed_at IS NOT NULL)
	)
);

CREATE UNIQUE INDEX accounting_periods_org_code_unique
	ON accounting.periods (org_id, lower(period_code));

CREATE INDEX accounting_periods_org_dates_idx
	ON accounting.periods (org_id, start_on, end_on);

ALTER TABLE accounting.journal_entries
ADD COLUMN effective_on DATE NOT NULL DEFAULT CURRENT_DATE;

CREATE INDEX accounting_journal_entries_org_effective_on_idx
	ON accounting.journal_entries (org_id, effective_on DESC, entry_number DESC);
