CREATE TABLE work_orders.work_orders (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	work_order_code TEXT NOT NULL,
	title TEXT NOT NULL,
	summary TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'open',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	closed_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT work_orders_work_orders_code_not_blank CHECK (btrim(work_order_code) <> ''),
	CONSTRAINT work_orders_work_orders_title_not_blank CHECK (btrim(title) <> ''),
	CONSTRAINT work_orders_work_orders_summary_not_blank CHECK (summary = '' OR btrim(summary) <> ''),
	CONSTRAINT work_orders_work_orders_status_allowed CHECK (status IN ('open', 'in_progress', 'completed', 'cancelled')),
	CONSTRAINT work_orders_work_orders_closed_consistent CHECK (
		(status IN ('completed', 'cancelled') AND closed_at IS NOT NULL)
		OR (status IN ('open', 'in_progress') AND closed_at IS NULL)
	)
);

CREATE UNIQUE INDEX work_orders_work_orders_org_code_unique
	ON work_orders.work_orders (org_id, lower(work_order_code));

CREATE INDEX work_orders_work_orders_org_status_created_idx
	ON work_orders.work_orders (org_id, status, created_at DESC);

CREATE TABLE work_orders.status_history (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	work_order_id UUID NOT NULL REFERENCES work_orders.work_orders (id) ON DELETE CASCADE,
	from_status TEXT,
	to_status TEXT NOT NULL,
	note TEXT NOT NULL DEFAULT '',
	changed_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT work_orders_status_history_from_status_allowed CHECK (
		from_status IS NULL OR from_status IN ('open', 'in_progress', 'completed', 'cancelled')
	),
	CONSTRAINT work_orders_status_history_to_status_allowed CHECK (to_status IN ('open', 'in_progress', 'completed', 'cancelled')),
	CONSTRAINT work_orders_status_history_note_not_blank CHECK (note = '' OR btrim(note) <> '')
);

CREATE INDEX work_orders_status_history_work_order_changed_idx
	ON work_orders.status_history (work_order_id, changed_at DESC);

CREATE TABLE work_orders.material_usages (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	work_order_id UUID NOT NULL REFERENCES work_orders.work_orders (id) ON DELETE CASCADE,
	inventory_execution_link_id UUID NOT NULL UNIQUE REFERENCES inventory_ops.execution_links (id) ON DELETE RESTRICT,
	inventory_document_id UUID NOT NULL REFERENCES inventory_ops.documents (document_id) ON DELETE RESTRICT,
	inventory_document_line_id UUID NOT NULL UNIQUE REFERENCES inventory_ops.document_lines (id) ON DELETE RESTRICT,
	inventory_movement_id UUID NOT NULL UNIQUE REFERENCES inventory_ops.movements (id) ON DELETE RESTRICT,
	item_id UUID NOT NULL REFERENCES inventory_ops.items (id) ON DELETE RESTRICT,
	movement_purpose TEXT NOT NULL,
	usage_classification TEXT NOT NULL,
	quantity_milli BIGINT NOT NULL,
	linked_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT work_orders_material_usages_purpose_allowed CHECK (movement_purpose IN ('resale', 'service_consumption', 'installed_equipment', 'direct_expense', 'stock_adjustment')),
	CONSTRAINT work_orders_material_usages_usage_allowed CHECK (usage_classification IN ('not_applicable', 'billable', 'non_billable')),
	CONSTRAINT work_orders_material_usages_quantity_positive CHECK (quantity_milli > 0)
);

CREATE INDEX work_orders_material_usages_work_order_linked_idx
	ON work_orders.material_usages (work_order_id, linked_at DESC);
