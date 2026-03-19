CREATE TABLE accounting.ledger_accounts (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	code TEXT NOT NULL,
	name TEXT NOT NULL,
	account_class TEXT NOT NULL,
	control_type TEXT NOT NULL DEFAULT 'none',
	allows_direct_posting BOOLEAN NOT NULL DEFAULT TRUE,
	status TEXT NOT NULL DEFAULT 'active',
	tax_category_code TEXT,
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounting_ledger_accounts_code_not_blank CHECK (btrim(code) <> ''),
	CONSTRAINT accounting_ledger_accounts_name_not_blank CHECK (btrim(name) <> ''),
	CONSTRAINT accounting_ledger_accounts_account_class_allowed CHECK (account_class IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
	CONSTRAINT accounting_ledger_accounts_control_type_allowed CHECK (control_type IN ('none', 'receivable', 'payable', 'gst_input', 'gst_output', 'tds_receivable', 'tds_payable')),
	CONSTRAINT accounting_ledger_accounts_status_allowed CHECK (status IN ('active', 'inactive')),
	CONSTRAINT accounting_ledger_accounts_tax_category_not_blank CHECK (tax_category_code IS NULL OR btrim(tax_category_code) <> '')
);

CREATE UNIQUE INDEX accounting_ledger_accounts_org_code_unique
	ON accounting.ledger_accounts (org_id, lower(code));

CREATE INDEX accounting_ledger_accounts_org_class_status_idx
	ON accounting.ledger_accounts (org_id, account_class, status, created_at DESC);

CREATE TABLE accounting.journal_numbering_series (
	org_id UUID PRIMARY KEY REFERENCES identityaccess.orgs (id) ON DELETE CASCADE,
	next_number BIGINT NOT NULL DEFAULT 1,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounting_journal_numbering_series_next_number_positive CHECK (next_number > 0)
);

CREATE TABLE accounting.journal_entries (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	entry_number BIGINT NOT NULL,
	entry_kind TEXT NOT NULL,
	source_document_id UUID REFERENCES documents.documents (id) ON DELETE RESTRICT,
	reversal_of_entry_id UUID REFERENCES accounting.journal_entries (id) ON DELETE RESTRICT,
	posting_fingerprint TEXT NOT NULL,
	currency_code TEXT NOT NULL DEFAULT 'INR',
	tax_scope_code TEXT NOT NULL DEFAULT 'none',
	summary TEXT NOT NULL DEFAULT '',
	reversal_reason TEXT,
	posted_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounting_journal_entries_entry_number_positive CHECK (entry_number > 0),
	CONSTRAINT accounting_journal_entries_entry_kind_allowed CHECK (entry_kind IN ('posting', 'reversal')),
	CONSTRAINT accounting_journal_entries_currency_code_valid CHECK (currency_code ~ '^[A-Z]{3}$'),
	CONSTRAINT accounting_journal_entries_tax_scope_allowed CHECK (tax_scope_code IN ('none', 'gst', 'tds', 'gst_tds')),
	CONSTRAINT accounting_journal_entries_posting_fingerprint_not_blank CHECK (btrim(posting_fingerprint) <> ''),
	CONSTRAINT accounting_journal_entries_posting_summary_not_blank CHECK (btrim(summary) <> ''),
	CONSTRAINT accounting_journal_entries_reversal_reason_consistent CHECK (
		(entry_kind = 'reversal' AND reversal_of_entry_id IS NOT NULL AND reversal_reason IS NOT NULL AND btrim(reversal_reason) <> '')
		OR (entry_kind = 'posting' AND reversal_of_entry_id IS NULL AND reversal_reason IS NULL)
	),
	CONSTRAINT accounting_journal_entries_source_document_consistent CHECK (
		(entry_kind = 'posting' AND source_document_id IS NOT NULL)
		OR (entry_kind = 'reversal' AND source_document_id IS NULL)
	)
);

CREATE UNIQUE INDEX accounting_journal_entries_org_number_unique
	ON accounting.journal_entries (org_id, entry_number);

CREATE UNIQUE INDEX accounting_journal_entries_source_document_unique
	ON accounting.journal_entries (source_document_id)
	WHERE entry_kind = 'posting';

CREATE UNIQUE INDEX accounting_journal_entries_reversal_of_unique
	ON accounting.journal_entries (reversal_of_entry_id)
	WHERE reversal_of_entry_id IS NOT NULL;

CREATE INDEX accounting_journal_entries_org_posted_idx
	ON accounting.journal_entries (org_id, posted_at DESC);

CREATE TABLE accounting.journal_lines (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	entry_id UUID NOT NULL REFERENCES accounting.journal_entries (id) ON DELETE RESTRICT,
	line_number INT NOT NULL,
	account_id UUID NOT NULL REFERENCES accounting.ledger_accounts (id) ON DELETE RESTRICT,
	description TEXT NOT NULL DEFAULT '',
	debit_minor BIGINT NOT NULL DEFAULT 0,
	credit_minor BIGINT NOT NULL DEFAULT 0,
	tax_code TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounting_journal_lines_line_number_positive CHECK (line_number > 0),
	CONSTRAINT accounting_journal_lines_one_sided_amount CHECK (
		(debit_minor > 0 AND credit_minor = 0)
		OR (credit_minor > 0 AND debit_minor = 0)
	),
	CONSTRAINT accounting_journal_lines_tax_code_not_blank CHECK (tax_code IS NULL OR btrim(tax_code) <> '')
);

CREATE UNIQUE INDEX accounting_journal_lines_entry_line_unique
	ON accounting.journal_lines (entry_id, line_number);

CREATE INDEX accounting_journal_lines_account_created_idx
	ON accounting.journal_lines (account_id, created_at DESC);

CREATE OR REPLACE FUNCTION accounting.validate_journal_entry_balance(target_entry_id UUID)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
	line_count INT;
	debit_total BIGINT;
	credit_total BIGINT;
BEGIN
	SELECT
		COUNT(*),
		COALESCE(SUM(debit_minor), 0),
		COALESCE(SUM(credit_minor), 0)
	INTO
		line_count,
		debit_total,
		credit_total
	FROM accounting.journal_lines
	WHERE entry_id = target_entry_id;

	IF line_count < 2 THEN
		RAISE EXCEPTION 'journal entry % must have at least two lines', target_entry_id;
	END IF;

	IF debit_total <> credit_total THEN
		RAISE EXCEPTION 'journal entry % is unbalanced', target_entry_id;
	END IF;
END;
$$;

CREATE OR REPLACE FUNCTION accounting.enforce_journal_entry_balance()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
	PERFORM accounting.validate_journal_entry_balance(COALESCE(NEW.entry_id, OLD.entry_id));
	RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION accounting.prevent_journal_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
	RAISE EXCEPTION 'posted accounting truth is append-only';
END;
$$;

CREATE CONSTRAINT TRIGGER accounting_journal_lines_balance_check
AFTER INSERT OR UPDATE OR DELETE ON accounting.journal_lines
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION accounting.enforce_journal_entry_balance();

CREATE TRIGGER accounting_journal_entries_no_update
BEFORE UPDATE OR DELETE ON accounting.journal_entries
FOR EACH ROW
EXECUTE FUNCTION accounting.prevent_journal_mutation();

CREATE TRIGGER accounting_journal_lines_no_update
BEFORE UPDATE OR DELETE ON accounting.journal_lines
FOR EACH ROW
EXECUTE FUNCTION accounting.prevent_journal_mutation();
