CREATE TABLE workforce.workers (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	worker_code TEXT NOT NULL,
	display_name TEXT NOT NULL,
	linked_user_id UUID REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	status TEXT NOT NULL DEFAULT 'active',
	default_hourly_cost_minor BIGINT NOT NULL,
	cost_currency_code TEXT NOT NULL,
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT workforce_workers_code_not_blank CHECK (btrim(worker_code) <> ''),
	CONSTRAINT workforce_workers_display_name_not_blank CHECK (btrim(display_name) <> ''),
	CONSTRAINT workforce_workers_status_allowed CHECK (status IN ('active', 'inactive')),
	CONSTRAINT workforce_workers_default_hourly_cost_non_negative CHECK (default_hourly_cost_minor >= 0),
	CONSTRAINT workforce_workers_currency_code_format CHECK (cost_currency_code ~ '^[A-Z]{3}$')
);

CREATE UNIQUE INDEX workforce_workers_org_code_unique
	ON workforce.workers (org_id, lower(worker_code));

CREATE UNIQUE INDEX workforce_workers_org_linked_user_unique
	ON workforce.workers (org_id, linked_user_id)
	WHERE linked_user_id IS NOT NULL;

CREATE INDEX workforce_workers_org_status_created_idx
	ON workforce.workers (org_id, status, created_at DESC);

CREATE TABLE workflow.tasks (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	context_type TEXT NOT NULL,
	context_id UUID NOT NULL,
	title TEXT NOT NULL,
	instructions TEXT NOT NULL DEFAULT '',
	queue_code TEXT,
	status TEXT NOT NULL DEFAULT 'open',
	accountable_worker_id UUID NOT NULL REFERENCES workforce.workers (id) ON DELETE RESTRICT,
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	completed_by_user_id UUID REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	completed_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT workflow_tasks_context_type_allowed CHECK (context_type IN ('work_order')),
	CONSTRAINT workflow_tasks_title_not_blank CHECK (btrim(title) <> ''),
	CONSTRAINT workflow_tasks_instructions_not_blank CHECK (instructions = '' OR btrim(instructions) <> ''),
	CONSTRAINT workflow_tasks_queue_code_not_blank CHECK (queue_code IS NULL OR btrim(queue_code) <> ''),
	CONSTRAINT workflow_tasks_status_allowed CHECK (status IN ('open', 'in_progress', 'completed', 'cancelled')),
	CONSTRAINT workflow_tasks_completion_fields_consistent CHECK (
		(status IN ('open', 'in_progress') AND completed_at IS NULL AND completed_by_user_id IS NULL)
		OR (status IN ('completed', 'cancelled') AND completed_at IS NOT NULL AND completed_by_user_id IS NOT NULL)
	)
);

CREATE INDEX workflow_tasks_org_context_status_idx
	ON workflow.tasks (org_id, context_type, context_id, status, created_at DESC);

CREATE INDEX workflow_tasks_accountable_worker_idx
	ON workflow.tasks (org_id, accountable_worker_id, status, created_at DESC);

CREATE TABLE workforce.labor_entries (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	worker_id UUID NOT NULL REFERENCES workforce.workers (id) ON DELETE RESTRICT,
	work_order_id UUID NOT NULL REFERENCES work_orders.work_orders (id) ON DELETE RESTRICT,
	task_id UUID REFERENCES workflow.tasks (id) ON DELETE RESTRICT,
	started_at TIMESTAMPTZ NOT NULL,
	ended_at TIMESTAMPTZ NOT NULL,
	duration_minutes INT NOT NULL,
	hourly_cost_minor BIGINT NOT NULL,
	cost_minor BIGINT NOT NULL,
	cost_currency_code TEXT NOT NULL,
	note TEXT NOT NULL DEFAULT '',
	captured_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT workforce_labor_entries_duration_positive CHECK (duration_minutes > 0),
	CONSTRAINT workforce_labor_entries_end_after_start CHECK (ended_at > started_at),
	CONSTRAINT workforce_labor_entries_hourly_cost_non_negative CHECK (hourly_cost_minor >= 0),
	CONSTRAINT workforce_labor_entries_cost_non_negative CHECK (cost_minor >= 0),
	CONSTRAINT workforce_labor_entries_currency_code_format CHECK (cost_currency_code ~ '^[A-Z]{3}$'),
	CONSTRAINT workforce_labor_entries_note_not_blank CHECK (note = '' OR btrim(note) <> '')
);

CREATE INDEX workforce_labor_entries_work_order_started_idx
	ON workforce.labor_entries (org_id, work_order_id, started_at DESC, id DESC);

CREATE INDEX workforce_labor_entries_task_started_idx
	ON workforce.labor_entries (org_id, task_id, started_at DESC, id DESC)
	WHERE task_id IS NOT NULL;
