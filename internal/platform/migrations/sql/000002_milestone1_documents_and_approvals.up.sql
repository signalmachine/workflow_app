CREATE TABLE documents.document_types (
	code TEXT PRIMARY KEY,
	display_name TEXT NOT NULL,
	numbering_policy TEXT NOT NULL DEFAULT 'optional',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT documents_document_types_code_not_blank CHECK (btrim(code) <> ''),
	CONSTRAINT documents_document_types_display_name_not_blank CHECK (btrim(display_name) <> ''),
	CONSTRAINT documents_document_types_numbering_policy_allowed CHECK (numbering_policy IN ('optional', 'required'))
);

INSERT INTO documents.document_types (code, display_name, numbering_policy) VALUES
	('work_order', 'Work Order', 'optional'),
	('invoice', 'Invoice', 'required'),
	('payment_receipt', 'Payment or Receipt', 'required'),
	('inventory_receipt', 'Inventory Receipt', 'required'),
	('inventory_issue', 'Inventory Issue', 'required'),
	('inventory_adjustment', 'Inventory Adjustment', 'required'),
	('journal', 'Journal', 'required'),
	('ai_draft', 'AI Draft Proposal', 'optional')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE documents.numbering_series (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	document_type_code TEXT NOT NULL REFERENCES documents.document_types (code) ON DELETE RESTRICT,
	series_code TEXT NOT NULL,
	next_number BIGINT NOT NULL DEFAULT 1,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT documents_numbering_series_series_code_not_blank CHECK (btrim(series_code) <> ''),
	CONSTRAINT documents_numbering_series_next_number_positive CHECK (next_number > 0)
);

CREATE UNIQUE INDEX documents_numbering_series_org_type_code_unique
	ON documents.numbering_series (org_id, document_type_code, lower(series_code));

CREATE TABLE documents.documents (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	type_code TEXT NOT NULL REFERENCES documents.document_types (code) ON DELETE RESTRICT,
	status TEXT NOT NULL,
	title TEXT NOT NULL DEFAULT '',
	number_series_id UUID REFERENCES documents.numbering_series (id) ON DELETE RESTRICT,
	number_value TEXT,
	source_document_id UUID REFERENCES documents.documents (id) ON DELETE RESTRICT,
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	submitted_by_user_id UUID REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	submitted_at TIMESTAMPTZ,
	approved_at TIMESTAMPTZ,
	rejected_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT documents_documents_status_allowed CHECK (status IN ('draft', 'submitted', 'approved', 'rejected', 'posted', 'reversed', 'voided')),
	CONSTRAINT documents_documents_number_value_not_blank CHECK (number_value IS NULL OR btrim(number_value) <> ''),
	CONSTRAINT documents_documents_submit_fields_consistent CHECK (
		(status = 'submitted' AND submitted_at IS NOT NULL AND submitted_by_user_id IS NOT NULL)
		OR (status <> 'submitted')
		OR (submitted_at IS NULL AND submitted_by_user_id IS NULL)
	),
	CONSTRAINT documents_documents_approved_at_consistent CHECK (
		(status = 'approved' AND approved_at IS NOT NULL)
		OR (status <> 'approved')
	),
	CONSTRAINT documents_documents_rejected_at_consistent CHECK (
		(status = 'rejected' AND rejected_at IS NOT NULL)
		OR (status <> 'rejected')
	)
);

CREATE INDEX documents_documents_org_status_idx
	ON documents.documents (org_id, status, created_at DESC);

CREATE INDEX documents_documents_org_type_idx
	ON documents.documents (org_id, type_code, created_at DESC);

CREATE UNIQUE INDEX documents_documents_series_number_unique
	ON documents.documents (number_series_id, lower(number_value))
	WHERE number_series_id IS NOT NULL AND number_value IS NOT NULL;

CREATE TABLE workflow.approvals (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	document_id UUID NOT NULL REFERENCES documents.documents (id) ON DELETE RESTRICT,
	status TEXT NOT NULL,
	queue_code TEXT NOT NULL,
	requested_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	decided_by_user_id UUID REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	decision_note TEXT,
	requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	decided_at TIMESTAMPTZ,
	CONSTRAINT workflow_approvals_status_allowed CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
	CONSTRAINT workflow_approvals_queue_code_not_blank CHECK (btrim(queue_code) <> ''),
	CONSTRAINT workflow_approvals_decision_fields_consistent CHECK (
		(status = 'pending' AND decided_by_user_id IS NULL AND decided_at IS NULL)
		OR (status IN ('approved', 'rejected', 'cancelled') AND decided_at IS NOT NULL)
	)
);

CREATE UNIQUE INDEX workflow_approvals_one_pending_per_document
	ON workflow.approvals (document_id)
	WHERE status = 'pending';

CREATE INDEX workflow_approvals_org_status_idx
	ON workflow.approvals (org_id, status, requested_at DESC);

CREATE TABLE workflow.approval_queue_entries (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	approval_id UUID NOT NULL UNIQUE REFERENCES workflow.approvals (id) ON DELETE CASCADE,
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	queue_code TEXT NOT NULL,
	status TEXT NOT NULL,
	enqueued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	closed_at TIMESTAMPTZ,
	CONSTRAINT workflow_approval_queue_entries_queue_code_not_blank CHECK (btrim(queue_code) <> ''),
	CONSTRAINT workflow_approval_queue_entries_status_allowed CHECK (status IN ('pending', 'closed')),
	CONSTRAINT workflow_approval_queue_entries_closed_at_consistent CHECK (
		(status = 'pending' AND closed_at IS NULL)
		OR (status = 'closed' AND closed_at IS NOT NULL)
	)
);

CREATE INDEX workflow_approval_queue_entries_org_queue_status_idx
	ON workflow.approval_queue_entries (org_id, queue_code, status, enqueued_at DESC);

CREATE TABLE workflow.approval_decisions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	approval_id UUID NOT NULL REFERENCES workflow.approvals (id) ON DELETE CASCADE,
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	document_id UUID NOT NULL REFERENCES documents.documents (id) ON DELETE RESTRICT,
	decision TEXT NOT NULL,
	actor_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	note TEXT,
	decided_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT workflow_approval_decisions_decision_allowed CHECK (decision IN ('approved', 'rejected', 'cancelled'))
);

CREATE INDEX workflow_approval_decisions_approval_idx
	ON workflow.approval_decisions (approval_id, decided_at DESC);
