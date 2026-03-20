CREATE SCHEMA IF NOT EXISTS inventory_ops;

CREATE TABLE inventory_ops.items (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	sku TEXT NOT NULL,
	name TEXT NOT NULL,
	item_role TEXT NOT NULL,
	tracking_mode TEXT NOT NULL DEFAULT 'none',
	status TEXT NOT NULL DEFAULT 'active',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT inventory_items_sku_not_blank CHECK (btrim(sku) <> ''),
	CONSTRAINT inventory_items_name_not_blank CHECK (btrim(name) <> ''),
	CONSTRAINT inventory_items_item_role_allowed CHECK (item_role IN ('resale', 'service_material', 'traceable_equipment', 'direct_expense_consumable')),
	CONSTRAINT inventory_items_tracking_mode_allowed CHECK (tracking_mode IN ('none', 'serial', 'lot')),
	CONSTRAINT inventory_items_status_allowed CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX inventory_items_org_sku_unique
	ON inventory_ops.items (org_id, lower(sku));

CREATE INDEX inventory_items_org_role_status_idx
	ON inventory_ops.items (org_id, item_role, status, created_at DESC);

CREATE TABLE inventory_ops.locations (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	code TEXT NOT NULL,
	name TEXT NOT NULL,
	location_role TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT inventory_locations_code_not_blank CHECK (btrim(code) <> ''),
	CONSTRAINT inventory_locations_name_not_blank CHECK (btrim(name) <> ''),
	CONSTRAINT inventory_locations_role_allowed CHECK (location_role IN ('warehouse', 'van', 'site', 'vendor', 'customer', 'adjustment', 'installed')),
	CONSTRAINT inventory_locations_status_allowed CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX inventory_locations_org_code_unique
	ON inventory_ops.locations (org_id, lower(code));

CREATE INDEX inventory_locations_org_role_status_idx
	ON inventory_ops.locations (org_id, location_role, status, created_at DESC);

CREATE TABLE inventory_ops.movement_numbering_series (
	org_id UUID PRIMARY KEY REFERENCES identityaccess.orgs (id) ON DELETE CASCADE,
	next_number BIGINT NOT NULL DEFAULT 1,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT inventory_movement_numbering_series_next_number_positive CHECK (next_number > 0)
);

CREATE TABLE inventory_ops.movements (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	movement_number BIGINT NOT NULL,
	document_id UUID REFERENCES documents.documents (id) ON DELETE RESTRICT,
	item_id UUID NOT NULL REFERENCES inventory_ops.items (id) ON DELETE RESTRICT,
	movement_type TEXT NOT NULL,
	movement_purpose TEXT NOT NULL,
	usage_classification TEXT NOT NULL DEFAULT 'not_applicable',
	source_location_id UUID REFERENCES inventory_ops.locations (id) ON DELETE RESTRICT,
	destination_location_id UUID REFERENCES inventory_ops.locations (id) ON DELETE RESTRICT,
	quantity_milli BIGINT NOT NULL,
	reference_note TEXT NOT NULL DEFAULT '',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT inventory_movements_number_positive CHECK (movement_number > 0),
	CONSTRAINT inventory_movements_type_allowed CHECK (movement_type IN ('receipt', 'issue', 'adjustment')),
	CONSTRAINT inventory_movements_purpose_allowed CHECK (movement_purpose IN ('resale', 'service_consumption', 'installed_equipment', 'direct_expense', 'stock_adjustment')),
	CONSTRAINT inventory_movements_usage_allowed CHECK (usage_classification IN ('not_applicable', 'billable', 'non_billable')),
	CONSTRAINT inventory_movements_quantity_positive CHECK (quantity_milli > 0),
	CONSTRAINT inventory_movements_reference_note_not_blank CHECK (reference_note = '' OR btrim(reference_note) <> ''),
	CONSTRAINT inventory_movements_locations_distinct CHECK (
		source_location_id IS NULL
		OR destination_location_id IS NULL
		OR source_location_id <> destination_location_id
	),
	CONSTRAINT inventory_movements_shape_valid CHECK (
		(movement_type = 'receipt' AND source_location_id IS NULL AND destination_location_id IS NOT NULL)
		OR (movement_type = 'issue' AND source_location_id IS NOT NULL AND destination_location_id IS NULL)
		OR (movement_type = 'adjustment' AND (
			(source_location_id IS NULL AND destination_location_id IS NOT NULL)
			OR (source_location_id IS NOT NULL AND destination_location_id IS NULL)
		))
	),
	CONSTRAINT inventory_movements_usage_consistent CHECK (
		(usage_classification = 'not_applicable' AND movement_purpose IN ('resale', 'stock_adjustment', 'installed_equipment'))
		OR (usage_classification IN ('billable', 'non_billable') AND movement_purpose IN ('service_consumption', 'direct_expense'))
	)
);

CREATE UNIQUE INDEX inventory_movements_org_number_unique
	ON inventory_ops.movements (org_id, movement_number);

CREATE INDEX inventory_movements_item_created_idx
	ON inventory_ops.movements (org_id, item_id, created_at DESC);

CREATE INDEX inventory_movements_source_location_idx
	ON inventory_ops.movements (org_id, source_location_id, created_at DESC);

CREATE INDEX inventory_movements_destination_location_idx
	ON inventory_ops.movements (org_id, destination_location_id, created_at DESC);

CREATE OR REPLACE FUNCTION inventory_ops.prevent_movement_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
	RAISE EXCEPTION 'inventory movement truth is append-only';
END;
$$;

CREATE TRIGGER inventory_movements_no_update
BEFORE UPDATE OR DELETE ON inventory_ops.movements
FOR EACH ROW
EXECUTE FUNCTION inventory_ops.prevent_movement_mutation();
