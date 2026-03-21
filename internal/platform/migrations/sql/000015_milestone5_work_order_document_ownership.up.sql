CREATE TABLE work_orders.documents (
	document_id UUID PRIMARY KEY REFERENCES documents.documents (id) ON DELETE RESTRICT,
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	work_order_id UUID NOT NULL UNIQUE REFERENCES work_orders.work_orders (id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX work_orders_documents_org_work_order_unique
	ON work_orders.documents (org_id, work_order_id);

INSERT INTO documents.documents (
	id,
	org_id,
	type_code,
	status,
	title,
	created_by_user_id,
	created_at,
	updated_at
)
SELECT
	wo.id,
	wo.org_id,
	'work_order',
	'draft',
	wo.title,
	wo.created_by_user_id,
	wo.created_at,
	wo.updated_at
FROM work_orders.work_orders wo
LEFT JOIN work_orders.documents wd
	ON wd.work_order_id = wo.id
WHERE wd.work_order_id IS NULL;

INSERT INTO work_orders.documents (
	document_id,
	org_id,
	work_order_id,
	created_at
)
SELECT
	wo.id,
	wo.org_id,
	wo.id,
	wo.created_at
FROM work_orders.work_orders wo
LEFT JOIN work_orders.documents wd
	ON wd.work_order_id = wo.id
WHERE wd.work_order_id IS NULL;
