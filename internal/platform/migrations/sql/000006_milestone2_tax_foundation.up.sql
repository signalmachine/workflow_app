CREATE TABLE accounting.tax_codes (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	code TEXT NOT NULL,
	name TEXT NOT NULL,
	tax_type TEXT NOT NULL,
	rate_basis_points INT NOT NULL,
	receivable_account_id UUID REFERENCES accounting.ledger_accounts (id) ON DELETE RESTRICT,
	payable_account_id UUID REFERENCES accounting.ledger_accounts (id) ON DELETE RESTRICT,
	status TEXT NOT NULL DEFAULT 'active',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounting_tax_codes_code_not_blank CHECK (btrim(code) <> ''),
	CONSTRAINT accounting_tax_codes_name_not_blank CHECK (btrim(name) <> ''),
	CONSTRAINT accounting_tax_codes_type_allowed CHECK (tax_type IN ('gst', 'tds')),
	CONSTRAINT accounting_tax_codes_rate_basis_points_valid CHECK (rate_basis_points >= 0 AND rate_basis_points <= 10000),
	CONSTRAINT accounting_tax_codes_status_allowed CHECK (status IN ('active', 'inactive')),
	CONSTRAINT accounting_tax_codes_control_account_required CHECK (
		receivable_account_id IS NOT NULL OR payable_account_id IS NOT NULL
	)
);

CREATE UNIQUE INDEX accounting_tax_codes_org_code_unique
	ON accounting.tax_codes (org_id, lower(code));

CREATE INDEX accounting_tax_codes_org_type_status_idx
	ON accounting.tax_codes (org_id, tax_type, status, created_at DESC);
