CREATE TABLE inventory_ops.documents (
	document_id UUID PRIMARY KEY REFERENCES documents.documents (id) ON DELETE RESTRICT,
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	movement_type TEXT NOT NULL,
	reference_note TEXT NOT NULL DEFAULT '',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT inventory_documents_movement_type_allowed CHECK (movement_type IN ('receipt', 'issue', 'adjustment')),
	CONSTRAINT inventory_documents_reference_note_not_blank CHECK (reference_note = '' OR btrim(reference_note) <> '')
);

CREATE INDEX inventory_documents_org_type_created_idx
	ON inventory_ops.documents (org_id, movement_type, created_at DESC);

CREATE TABLE inventory_ops.document_lines (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	document_id UUID NOT NULL REFERENCES inventory_ops.documents (document_id) ON DELETE CASCADE,
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	line_number INT NOT NULL,
	movement_id UUID NOT NULL UNIQUE REFERENCES inventory_ops.movements (id) ON DELETE RESTRICT,
	item_id UUID NOT NULL REFERENCES inventory_ops.items (id) ON DELETE RESTRICT,
	movement_purpose TEXT NOT NULL,
	usage_classification TEXT NOT NULL DEFAULT 'not_applicable',
	source_location_id UUID REFERENCES inventory_ops.locations (id) ON DELETE RESTRICT,
	destination_location_id UUID REFERENCES inventory_ops.locations (id) ON DELETE RESTRICT,
	quantity_milli BIGINT NOT NULL,
	reference_note TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT inventory_document_lines_line_number_positive CHECK (line_number > 0),
	CONSTRAINT inventory_document_lines_purpose_allowed CHECK (movement_purpose IN ('resale', 'service_consumption', 'installed_equipment', 'direct_expense', 'stock_adjustment')),
	CONSTRAINT inventory_document_lines_usage_allowed CHECK (usage_classification IN ('not_applicable', 'billable', 'non_billable')),
	CONSTRAINT inventory_document_lines_quantity_positive CHECK (quantity_milli > 0),
	CONSTRAINT inventory_document_lines_reference_note_not_blank CHECK (reference_note = '' OR btrim(reference_note) <> ''),
	CONSTRAINT inventory_document_lines_locations_distinct CHECK (
		source_location_id IS NULL
		OR destination_location_id IS NULL
		OR source_location_id <> destination_location_id
	),
	CONSTRAINT inventory_document_lines_usage_consistent CHECK (
		(usage_classification = 'not_applicable' AND movement_purpose IN ('resale', 'stock_adjustment', 'installed_equipment'))
		OR (usage_classification IN ('billable', 'non_billable') AND movement_purpose IN ('service_consumption', 'direct_expense'))
	)
);

CREATE UNIQUE INDEX inventory_document_lines_document_line_unique
	ON inventory_ops.document_lines (document_id, line_number);

CREATE INDEX inventory_document_lines_item_created_idx
	ON inventory_ops.document_lines (org_id, item_id, created_at DESC);

CREATE TABLE inventory_ops.accounting_handoffs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	document_id UUID NOT NULL REFERENCES inventory_ops.documents (document_id) ON DELETE CASCADE,
	document_line_id UUID NOT NULL UNIQUE REFERENCES inventory_ops.document_lines (id) ON DELETE CASCADE,
	journal_entry_id UUID REFERENCES accounting.journal_entries (id) ON DELETE RESTRICT,
	handoff_status TEXT NOT NULL DEFAULT 'pending',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT inventory_accounting_handoffs_status_allowed CHECK (handoff_status IN ('pending', 'posted')),
	CONSTRAINT inventory_accounting_handoffs_posted_consistent CHECK (
		(handoff_status = 'pending' AND journal_entry_id IS NULL)
		OR (handoff_status = 'posted' AND journal_entry_id IS NOT NULL)
	)
);

CREATE INDEX inventory_accounting_handoffs_org_status_created_idx
	ON inventory_ops.accounting_handoffs (org_id, handoff_status, created_at DESC);

CREATE TABLE inventory_ops.execution_links (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	document_id UUID NOT NULL REFERENCES inventory_ops.documents (document_id) ON DELETE CASCADE,
	document_line_id UUID NOT NULL UNIQUE REFERENCES inventory_ops.document_lines (id) ON DELETE CASCADE,
	execution_context_type TEXT NOT NULL,
	execution_context_id TEXT NOT NULL,
	linkage_status TEXT NOT NULL DEFAULT 'pending',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT inventory_execution_links_context_type_allowed CHECK (execution_context_type IN ('work_order', 'project')),
	CONSTRAINT inventory_execution_links_context_id_not_blank CHECK (btrim(execution_context_id) <> ''),
	CONSTRAINT inventory_execution_links_status_allowed CHECK (linkage_status IN ('pending', 'linked'))
);

CREATE INDEX inventory_execution_links_org_context_created_idx
	ON inventory_ops.execution_links (org_id, execution_context_type, created_at DESC);
