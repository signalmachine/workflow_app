# `implementation_plan` README

Purpose: preserve the pre-thin-v1 planning set as legacy reference material for historical context, existing implementation detail, and focused remediation follow-up.

Status note:
1. `plan_docs/` is the active canonical planning set
2. this folder is no longer the default reading order for new implementation work
3. use these files only when active `plan_docs/` material points here or when a specific historical decision needs clarification

## Legacy reading order

If older implementation detail needs to be checked, these are the main legacy documents:

1. `service_day_initial_plan_2026_03_11.md`
   - product intent, phase ordering, and strategic boundaries
2. `service_day_execution_plan_v1.md`
   - milestone path, vertical slices, and phase-by-phase testing
3. `service_day_module_boundaries_v1.md`
   - bounded contexts, ownership rules, and dependency direction
4. `service_day_navigation_launch_experience_v1.md`
   - activity-centered launch, direct workflow entry, and global search requirements
5. `service_day_data_exchange_plan_v1.md`
   - later Excel/CSV import-export direction and early foundation requirements
6. `service_day_schema_foundation_v1.md`
   - foundational tables, tenancy rules, and schema invariants
7. `service_day_crm_mvp_scope_v1.md`
   - first runnable product acceptance scope
8. `service_day_ai_architecture_v1.md`
   - AI provider, tool, approval, and persistence shape
9. `implementation_decisions_v1.md`
   - v1 defaults that are now locked for implementation unless explicitly changed
10. `implementation_tracker.md`
   - historical execution evidence, dependency-gate history, and legacy remediation indexing
11. `service_day_user_workflow_docs_plan_v1.md`
   - separate planning and sequencing for end-user workflow guides tied to shipped functionality

## Legacy document roles

Use each document for one purpose only:

1. strategy changes go in `service_day_initial_plan_2026_03_11.md`
2. sequencing and milestone changes go in `service_day_execution_plan_v1.md`
3. ownership and dependency changes go in `service_day_module_boundaries_v1.md`
4. launch experience, direct-entry, and global-search requirements go in `service_day_navigation_launch_experience_v1.md`
5. later spreadsheet import/export direction and early exchange foundations go in `service_day_data_exchange_plan_v1.md`
6. table and invariant changes go in `service_day_schema_foundation_v1.md`
7. CRM acceptance changes go in `service_day_crm_mvp_scope_v1.md`
8. AI execution or safety changes go in `service_day_ai_architecture_v1.md`
9. implementation-default decisions go in `implementation_decisions_v1.md`
10. historical evidence and legacy remediation indexing go in `implementation_tracker.md`; active thin-v1 status lives in `plan_docs/service_day_refactor_tracker_v1.md`
11. end-user workflow-document sequencing and deferral rules go in `service_day_user_workflow_docs_plan_v1.md`

## Legacy defaults

These defaults explain older implementation context, but they do not override `plan_docs/`:

1. one shared `tasks` engine exists for CRM and delivery
2. shared task orchestration belongs to `workflow`
3. delivery milestones and billing milestones are separate concepts with separate ownership
4. `project` is optional; `work_order` is the primary execution record
5. communication starts manual-first but remains import-ready
6. AI can propose and draft, but meaningful writes stay bounded and auditable
7. the strongest long-term product capability is `work_order` support, with CRM, projects, billing, and accounting integrated tightly around that execution core

## Legacy tracker usage

Use `implementation_tracker.md` only when one of these legacy-reference events happens:

1. older implementation evidence needs to stay preserved after the active thin-v1 tracker moved to `plan_docs/service_day_refactor_tracker_v1.md`
2. a focused legacy remediation note under `implementation_plan/` needs an index entry so it does not become orphaned
3. historical milestone or decision-gate context needs to remain readable for older shipped slices
4. a legacy planning note needs a pointer to the newer active thin-v1 document that replaced it

Focused remediation notes:
1. remediation notes under `implementation_plan/` are allowed for slice-specific follow-up that should not rewrite a canonical document
2. if a remediation note is still useful, make its legacy role explicit and prefer linking it from the active `plan_docs` tracker or migration map rather than treating `implementation_tracker.md` as a second live execution record
3. use `plan_docs/service_day_refactor_tracker_v1.md` plus any linked remediation note to resume active implementation after planning updates

## Legacy evidence rule

Do not mark a tracker item complete until all of these exist:

1. schema or code change merged in the codebase
2. tests or verification notes recorded
3. dependency gates satisfied
4. acceptance behavior demonstrated for user-visible slices
5. any new planning deviations documented in the relevant canonical file
