CREATE SCHEMA IF NOT EXISTS ai;

CREATE TABLE ai.agent_tools (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tool_name TEXT NOT NULL UNIQUE,
	display_name TEXT NOT NULL,
	module_code TEXT NOT NULL,
	mutates_state BOOLEAN NOT NULL DEFAULT FALSE,
	status TEXT NOT NULL DEFAULT 'active',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT ai_agent_tools_tool_name_not_blank CHECK (btrim(tool_name) <> ''),
	CONSTRAINT ai_agent_tools_display_name_not_blank CHECK (btrim(display_name) <> ''),
	CONSTRAINT ai_agent_tools_module_code_not_blank CHECK (btrim(module_code) <> ''),
	CONSTRAINT ai_agent_tools_status_allowed CHECK (status IN ('active', 'inactive'))
);

CREATE TABLE ai.agent_runs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	session_id UUID NOT NULL REFERENCES identityaccess.sessions (id) ON DELETE RESTRICT,
	actor_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	agent_role TEXT NOT NULL,
	capability_code TEXT NOT NULL,
	status TEXT NOT NULL,
	request_text TEXT NOT NULL DEFAULT '',
	summary TEXT NOT NULL DEFAULT '',
	metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
	parent_run_id UUID REFERENCES ai.agent_runs (id) ON DELETE RESTRICT,
	started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	completed_at TIMESTAMPTZ,
	CONSTRAINT ai_agent_runs_agent_role_allowed CHECK (agent_role IN ('coordinator', 'specialist')),
	CONSTRAINT ai_agent_runs_capability_code_not_blank CHECK (btrim(capability_code) <> ''),
	CONSTRAINT ai_agent_runs_status_allowed CHECK (status IN ('running', 'completed', 'failed', 'cancelled')),
	CONSTRAINT ai_agent_runs_completion_consistent CHECK (
		(status = 'running' AND completed_at IS NULL)
		OR (status IN ('completed', 'failed', 'cancelled') AND completed_at IS NOT NULL)
	)
);

CREATE INDEX ai_agent_runs_org_status_started_idx
	ON ai.agent_runs (org_id, status, started_at DESC);

CREATE INDEX ai_agent_runs_parent_idx
	ON ai.agent_runs (parent_run_id, started_at DESC);

CREATE TABLE ai.agent_run_steps (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	run_id UUID NOT NULL REFERENCES ai.agent_runs (id) ON DELETE CASCADE,
	step_index INT NOT NULL,
	step_type TEXT NOT NULL,
	step_title TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	input_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
	output_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT ai_agent_run_steps_step_index_positive CHECK (step_index > 0),
	CONSTRAINT ai_agent_run_steps_step_type_not_blank CHECK (btrim(step_type) <> ''),
	CONSTRAINT ai_agent_run_steps_status_allowed CHECK (status IN ('completed', 'failed'))
);

CREATE UNIQUE INDEX ai_agent_run_steps_run_step_index_unique
	ON ai.agent_run_steps (run_id, step_index);

CREATE TABLE ai.agent_tool_policies (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	capability_code TEXT NOT NULL,
	tool_id UUID NOT NULL REFERENCES ai.agent_tools (id) ON DELETE RESTRICT,
	policy TEXT NOT NULL,
	rationale TEXT NOT NULL DEFAULT '',
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT ai_agent_tool_policies_capability_code_not_blank CHECK (btrim(capability_code) <> ''),
	CONSTRAINT ai_agent_tool_policies_policy_allowed CHECK (policy IN ('allow', 'approval_required', 'deny'))
);

CREATE UNIQUE INDEX ai_agent_tool_policies_org_capability_tool_unique
	ON ai.agent_tool_policies (org_id, capability_code, tool_id);

CREATE INDEX ai_agent_tool_policies_org_capability_idx
	ON ai.agent_tool_policies (org_id, capability_code, updated_at DESC);

CREATE TABLE ai.agent_artifacts (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	run_id UUID NOT NULL REFERENCES ai.agent_runs (id) ON DELETE CASCADE,
	step_id UUID REFERENCES ai.agent_run_steps (id) ON DELETE SET NULL,
	artifact_type TEXT NOT NULL,
	title TEXT NOT NULL DEFAULT '',
	payload JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT ai_agent_artifacts_artifact_type_not_blank CHECK (btrim(artifact_type) <> '')
);

CREATE INDEX ai_agent_artifacts_run_created_idx
	ON ai.agent_artifacts (run_id, created_at DESC);

CREATE TABLE ai.agent_recommendations (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	run_id UUID NOT NULL REFERENCES ai.agent_runs (id) ON DELETE CASCADE,
	artifact_id UUID REFERENCES ai.agent_artifacts (id) ON DELETE SET NULL,
	approval_id UUID REFERENCES workflow.approvals (id) ON DELETE SET NULL,
	recommendation_type TEXT NOT NULL,
	status TEXT NOT NULL,
	summary TEXT NOT NULL DEFAULT '',
	payload JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_by_user_id UUID NOT NULL REFERENCES identityaccess.users (id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT ai_agent_recommendations_type_not_blank CHECK (btrim(recommendation_type) <> ''),
	CONSTRAINT ai_agent_recommendations_status_allowed CHECK (status IN ('proposed', 'approval_requested', 'accepted', 'rejected'))
);

CREATE INDEX ai_agent_recommendations_org_status_created_idx
	ON ai.agent_recommendations (org_id, status, created_at DESC);

CREATE TABLE ai.agent_delegations (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	org_id UUID NOT NULL REFERENCES identityaccess.orgs (id) ON DELETE RESTRICT,
	parent_run_id UUID NOT NULL REFERENCES ai.agent_runs (id) ON DELETE CASCADE,
	child_run_id UUID NOT NULL UNIQUE REFERENCES ai.agent_runs (id) ON DELETE CASCADE,
	requested_by_step_id UUID REFERENCES ai.agent_run_steps (id) ON DELETE SET NULL,
	capability_code TEXT NOT NULL,
	reason TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT ai_agent_delegations_capability_code_not_blank CHECK (btrim(capability_code) <> '')
);

CREATE INDEX ai_agent_delegations_org_parent_created_idx
	ON ai.agent_delegations (org_id, parent_run_id, created_at DESC);
