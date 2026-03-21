CREATE UNIQUE INDEX IF NOT EXISTS documents_documents_org_id_unique
	ON documents.documents (org_id, id);

CREATE UNIQUE INDEX IF NOT EXISTS parties_parties_org_id_unique
	ON parties.parties (org_id, id);

CREATE UNIQUE INDEX IF NOT EXISTS parties_contacts_org_id_unique
	ON parties.contacts (org_id, id);

CREATE TABLE accounting.invoice_documents (
	document_id UUID PRIMARY KEY,
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	invoice_role TEXT,
	billed_party_id UUID,
	billing_contact_id UUID,
	currency_code TEXT,
	reference_value TEXT NOT NULL DEFAULT '',
	summary TEXT NOT NULL DEFAULT '',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounting_invoice_documents_document_fk
		FOREIGN KEY (org_id, document_id) REFERENCES documents.documents (org_id, id) ON DELETE RESTRICT,
	CONSTRAINT accounting_invoice_documents_party_fk
		FOREIGN KEY (org_id, billed_party_id) REFERENCES parties.parties (org_id, id) ON DELETE RESTRICT,
	CONSTRAINT accounting_invoice_documents_contact_fk
		FOREIGN KEY (org_id, billing_contact_id) REFERENCES parties.contacts (org_id, id) ON DELETE RESTRICT,
	CONSTRAINT accounting_invoice_documents_role_allowed CHECK (invoice_role IS NULL OR invoice_role IN ('sales', 'purchase')),
	CONSTRAINT accounting_invoice_documents_currency_code_valid CHECK (
		currency_code IS NULL OR currency_code ~ '^[A-Z]{3}$'
	),
	CONSTRAINT accounting_invoice_documents_reference_value_not_blank CHECK (
		reference_value = '' OR btrim(reference_value) <> ''
	),
	CONSTRAINT accounting_invoice_documents_summary_not_blank CHECK (
		summary = '' OR btrim(summary) <> ''
	)
);

CREATE INDEX accounting_invoice_documents_org_created_idx
	ON accounting.invoice_documents (org_id, created_at DESC);

CREATE TABLE accounting.payment_receipt_documents (
	document_id UUID PRIMARY KEY,
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	direction TEXT,
	counterparty_id UUID,
	counterparty_contact_id UUID,
	currency_code TEXT,
	reference_value TEXT NOT NULL DEFAULT '',
	summary TEXT NOT NULL DEFAULT '',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounting_payment_receipt_documents_document_fk
		FOREIGN KEY (org_id, document_id) REFERENCES documents.documents (org_id, id) ON DELETE RESTRICT,
	CONSTRAINT accounting_payment_receipt_documents_party_fk
		FOREIGN KEY (org_id, counterparty_id) REFERENCES parties.parties (org_id, id) ON DELETE RESTRICT,
	CONSTRAINT accounting_payment_receipt_documents_contact_fk
		FOREIGN KEY (org_id, counterparty_contact_id) REFERENCES parties.contacts (org_id, id) ON DELETE RESTRICT,
	CONSTRAINT accounting_payment_receipt_documents_direction_allowed CHECK (
		direction IS NULL OR direction IN ('payment', 'receipt')
	),
	CONSTRAINT accounting_payment_receipt_documents_currency_code_valid CHECK (
		currency_code IS NULL OR currency_code ~ '^[A-Z]{3}$'
	),
	CONSTRAINT accounting_payment_receipt_documents_reference_value_not_blank CHECK (
		reference_value = '' OR btrim(reference_value) <> ''
	),
	CONSTRAINT accounting_payment_receipt_documents_summary_not_blank CHECK (
		summary = '' OR btrim(summary) <> ''
	)
);

CREATE INDEX accounting_payment_receipt_documents_org_created_idx
	ON accounting.payment_receipt_documents (org_id, created_at DESC);

INSERT INTO accounting.invoice_documents (
	document_id,
	org_id,
	currency_code,
	created_by_user_id,
	created_at,
	updated_at
)
SELECT
	d.id,
	d.org_id,
	je.currency_code,
	d.created_by_user_id,
	d.created_at,
	d.updated_at
FROM documents.documents d
LEFT JOIN accounting.invoice_documents aid
	ON aid.document_id = d.id
LEFT JOIN LATERAL (
	SELECT currency_code
	FROM accounting.journal_entries je
	WHERE je.org_id = d.org_id
	  AND je.source_document_id = d.id
	  AND je.entry_kind = 'posting'
	ORDER BY je.posted_at DESC, je.id DESC
	LIMIT 1
) je ON TRUE
WHERE d.type_code = 'invoice'
  AND aid.document_id IS NULL;

INSERT INTO accounting.payment_receipt_documents (
	document_id,
	org_id,
	currency_code,
	created_by_user_id,
	created_at,
	updated_at
)
SELECT
	d.id,
	d.org_id,
	je.currency_code,
	d.created_by_user_id,
	d.created_at,
	d.updated_at
FROM documents.documents d
LEFT JOIN accounting.payment_receipt_documents prd
	ON prd.document_id = d.id
LEFT JOIN LATERAL (
	SELECT currency_code
	FROM accounting.journal_entries je
	WHERE je.org_id = d.org_id
	  AND je.source_document_id = d.id
	  AND je.entry_kind = 'posting'
	ORDER BY je.posted_at DESC, je.id DESC
	LIMIT 1
) je ON TRUE
WHERE d.type_code = 'payment_receipt'
  AND prd.document_id IS NULL;
