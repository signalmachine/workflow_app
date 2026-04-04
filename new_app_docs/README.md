# workflow_app Planning README

Date: 2026-04-03
Status: Canonical planning set for the completed thin-v1 foundation and the active ambitious v2 implementation phase
Purpose: define the planning baseline that started the repository reset, records thin-v1 completion, and now governs the active ambitious v2 implementation phase.

## Why this folder exists

The current repository proved the product idea, but it also proved that the shipped CRM-heavy implementation shape creates ongoing scope tension against the intended thin-foundation v1.

This folder exists to prepare and now continue the repository after the clean restart with:

1. stricter module boundaries
2. stricter thin-v1 scope control
3. documents, ledgers, execution context, approvals, and reports as the real center of gravity
4. AI-agent-first operation with strong human review and operator surfaces built on one shared backend truth

This planning set is the repository's canonical implementation-planning source.

Use this folder for:

1. product implementation plans
2. issue-fix planning discovered during implementation or workflow validation
3. implementation status, sequencing, and next implementation steps
4. activity tracking for work that changes the product

For the separate workflow-reference and workflow-validation track, use `docs/workflows/`.

Track rule:

1. keep workflow testing and review evidence out of `new_app_docs/` unless it directly promotes a concrete product fix plan or implementation change
2. if workflow validation discovers a real product issue, add the resulting fix plan here and track implementation here
3. workflow validation may be paused or deferred when urgent product fixes need implementation first without treating that pause as drift

Thin-v1 is now complete at foundation depth.

The active planning phase is now v2 beginning at Milestone 10.

V2 in this repository means:

1. enhance the application across backend, AI, browser, validation, and operational readiness layers on top of the completed foundation
2. broaden capability deliberately and ambitiously where it strengthens real operator usefulness, maintainability, and production readiness
3. continuously review the codebase for weak seams, underbuilt architecture, or outdated implementation choices and refactor or rebuild them when appropriate
4. keep the completed thin-v1 foundation, shared truth model, and workflow doctrine intact rather than reopening foundational modeling under a different label

The active browser-direction update is now explicit:

1. the earlier Go-template browser rebuild established the shared `/app` plus `/api/...` seam and validated the workflow-centered browser model
2. the approved forward web direction is now the Svelte-based replacement documented in `../docs/svelte_web_guides/svelte_web_ui_migration_plan.md`
3. the canonical implementation-planning surface for that migration now lives in the Milestone 13 planning docs inside `new_app_docs/`
4. the shared Go backend, session model, API seam, and workflow doctrine remain the truth foundation for that migration

## Active posture

The active posture is no longer thin, conservative expansion.

It is an ambitious best-practice implementation phase:

1. contributors may apply proven engineering, architecture, product, and operational best practices without artificial thin-scope limits
2. contributors should not preserve weak code, awkward module boundaries, or underbuilt seams merely because they already exist
3. contributors should proactively identify refactor or rebuild opportunities and either implement them in the current change when appropriate or promote them into the canonical planning docs
4. `go wild` means freedom to use strong, established practices, not freedom to introduce novelty, experimentation, or avoidable complexity that is not justified by business-software best practice
5. large monolithic code files and other `God` file concentrations should be treated as explicit refactor or rebuild candidates during active implementation planning

`workflow_app` still keeps the same multi-agent doctrine moving into v2:

1. one coordinator routing bounded work to specialist agents
2. durable run history, tool policy, artifacts, and delegation traces
3. no advanced autonomy features unless they are required for foundation correctness

Anything in `examples/` is read-only and reference-only:

1. these folders are typically preserved from older implementations or planning eras
2. they must not be changed or updated as part of active `workflow_app` implementation work
3. they are not part of the active `workflow_app` implementation surface

The former accounting-agent proof-of-concept may still be used as external historical reference material during implementation review and selective borrowing, but it is no longer kept inside this repository. Use https://github.com/signalmachine/accounting-agent-app when that comparison is materially useful, and keep the same boundary:

1. it was a proof-of-concept application rather than the architectural target for this repository
2. it used one single-agent pattern rather than the coordinator-plus-specialist architecture planned for `workflow_app`
3. it still matters because it worked within its intended limits and achieved its objectives, so implementation decisions in `workflow_app` may continue to compare against it for practical operator usefulness and readiness signal
4. it must not be changed, updated, or treated as part of the active `workflow_app` implementation surface

## Recommended reading order

For the active v2 planning surface, read these in order:

1. `new_app_v1_principles.md`
2. `new_app_v1_scope.md`
3. `new_app_schema_and_module_boundaries.md`
4. `new_app_implementation_defaults.md`
5. `new_app_foundation_coverage.md`
6. `new_app_execution_plan.md`
7. `new_app_tracker.md`
8. `milestone_13_svelte_web_migration_plan.md`
9. `milestone_13_slice_1_svelte_foundation_and_shell_plan.md`
10. `milestone_13_slice_2_svelte_workflow_surfaces_plan.md`
11. `milestone_13_slice_3_svelte_detail_admin_and_cutover_plan.md`
12. `browser_testing_lessons.md`
13. `../docs/svelte_web_guides/svelte_web_ui_migration_plan.md`
14. `milestone_12_admin_maintenance_and_master_data_plan.md`
15. `../docs/workflows/README.md` for the durable workflow-reference and validation-track layer after the active planning read

Reference-only rule:

1. `thin_v1_archive/` is historical context only and should not be part of the default active working set
2. `app_v2_plans/` is idea and proposal reference material only and should not be part of the default active working set
3. the earlier browser-stack planning docs in top-level `new_app_docs/` for Milestone 10, Milestone 11, and the ERP-style density correction are also historical context now that the Svelte migration is the approved forward web direction
4. use historical material only when the current task specifically needs prior browser-product decisions, route-family intent, or implementation history rather than active stack guidance

## Reset decision

The intended reset is:

1. start a new codebase
2. do not continue the current codebase by deleting large slices and trying to reshape the remaining implementation in place
3. treat the current repository as a reference source for selective ideas, not as the implementation base

Reason:

1. the current implementation shape carries too much CRM-first gravitational pull
2. deleting parts of the existing codebase would still leave the new effort spending time on untangling legacy assumptions
3. a clean repository will make it easier to enforce thin-v1 rules from the first migration onward

## Phase rule

Thin-v1 is complete.

Starting at Milestone 10, active implementation work is v2 work.

That means:

1. the repository is no longer restricted to thin-v1 breadth
2. new work should focus on stronger operator usability, broader capability where justified, and production readiness
3. v2 must still preserve the core doctrine, shared backend truth, approval boundaries, and module-discipline rules established by thin-v1

## Quality rule

Thin v1 does not mean low quality, low rigor, or simplistic modeling.

It means:

1. foundation-complete first
2. narrow module count
3. narrow workflow breadth
4. high technical quality in the layers that do land
