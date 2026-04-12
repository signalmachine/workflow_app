# workflow_app Planning README

Date: 2026-04-10
Status: Canonical v2 planning surface for active implementation
Purpose: keep the active `new_app_docs/` working set small, directive, and focused on pending work.

## Why this folder exists

This folder is the canonical implementation-planning surface for active `workflow_app` work.

Use the root of `new_app_docs/` only for:

1. current status
2. active scope and architecture rules
3. active execution order
4. active implementation defaults

Historical planning detail now lives in:

1. `v2_archive/` for the previous heavy v2 planning set
2. `thin_v1_archive/` for completed thin-v1 planning history
3. `app_v2_plans/` for proposal and parking-lot material rather than active control docs
4. `future_plans/` for future implementation candidates that are not active implementation work

## Active reading order

Read these in order for active implementation work:

1. `new_app_tracker_v2.md`
2. `new_app_scope_v2.md`
3. `new_app_architecture_v2.md`
4. `new_app_implementation_defaults_v2.md`
5. `new_app_execution_plan_v2.md`
6. `milestone_15_urgent_review_and_ai_hardening_plan.md` when working on the current urgent corrective milestone
7. `milestone_16_structural_and_ai_capability_plan.md` when working on follow-on structural and AI capability improvements after Milestone 15
8. `milestone_14_production_readiness_and_workflow_validation_plan.md` only when historical Milestone 14 detail is needed
9. `future_plans/data_exchange_and_bulk_import_plan.md` for the future data-exchange candidate; do not start data-exchange implementation until the Milestone 15 urgent review findings have been closed and user-testing findings have been triaged
10. `ai_layer_improvement_plan.md` when working on AI-agent layer improvement findings
11. `code_review_and_improvement_plan.md` when working on application-wide code-review findings

Then use:

1. `../docs/technical_guides/07_testing_and_verification.md` for exact verification workflow
2. `../docs/workflows/README.md` for durable workflow validation and workflow-reference material

## Archive rule

1. do not expand the root planning docs back into long historical records
2. when a root doc becomes history-heavy, move that detail into `v2_archive/` and keep the root version thin
3. use archived docs only when the current task needs historical rationale, previous milestone detail, or old corrective slices that are not already captured in the thin root docs
4. do not treat archived docs as the default active reading surface
