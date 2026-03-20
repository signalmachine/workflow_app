# `plan_docs` README

Date: 2026-03-19
Status: Active canonical planning set
Purpose: define the active thin-v1 planning direction for `service_day` based on the doctrine in `docs/implementation_objectives/implementation_principles.md`.

## Why this folder exists

The current `implementation_plan/` set is broad, CRM-heavy, and mixes v1 foundations with many later product ambitions. This folder defines the accepted thin-v1 reset shaped by:

1. documents as intent
2. ledgers as truth
3. execution context as operations
4. AI as the primary interface
5. reports as derived views

These docs are the primary planning source for new implementation work.

For a high-level cross-cutting summary of the implementation rules and invariants that sit across these documents, see [implementation_objectives.md](/home/vinod/PROJECTS/service_day/docs/implementation_objectives/implementation_objectives.md).

`implementation_plan/` is maintained as legacy reference material:

1. use it for historical context
2. use it when a specific older implementation detail needs to be checked
3. do not let it override `plan_docs/` for new scope or priority decisions

## Reading order

Read these in order:

1. `current_state_review_2026_03_18.md`
   - summary of what the current codebase and plan set are optimized for now
2. `implementation_plan_migration_map_v1.md`
   - which current canonical docs should be kept, collapsed, or moved to v2
3. `service_day_refactor_principles_v1.md`
   - application-specific interpretation of the ledger + documents + execution-context doctrine
4. `service_day_ai_architecture_v1.md`
   - active thin-v1 AI agent architecture and observability goals
5. `service_day_foundation_coverage_v1.md`
   - required foundation checklist so v1 does not ship with hidden structural gaps
6. `service_day_schema_and_module_boundaries_v1.md`
   - active thin-v1 ownership and schema-boundary baseline
7. `service_day_implementation_defaults_v1.md`
   - active thin-v1 locked defaults that implementation should preserve
8. `service_day_thin_v1_scope_v1.md`
   - hard v1 scope, explicit cuts, and priority reset
9. `service_day_refactor_execution_plan_v1.md`
   - milestone path for the thin v1
10. `service_day_refactor_tracker_v1.md`
   - active thin-v1 milestone and cleanup tracker

## Implementation tracking rule

Use `service_day_refactor_tracker_v1.md` as the single live implementation-status document inside `plan_docs/`.

Rules:

1. record implementation status, evidence, and near-term next-step notes in `service_day_refactor_tracker_v1.md`
2. use `service_day_refactor_execution_plan_v1.md` for milestone order and exit criteria, not day-to-day status tracking
3. use `service_day_foundation_coverage_v1.md` for completion criteria, not live progress notes
4. treat `current_state_review_2026_03_18.md` as a dated review snapshot, not as the central implementation tracker
5. update other `plan_docs/` files only when canonical scope, architecture, ownership, sequencing, or invariants change

## Intended use

Use this folder to:

1. guide new implementation work
2. decide what stays in v1 and what moves to v2
3. interpret legacy `implementation_plan/` material safely

## Refactor stance

This refactor assumes:

1. the product is AI-agent-first, not form-first
2. humans approve important actions, especially posting
3. v1 should optimize for foundations, not breadth
4. accounting, inventory, work-order execution, tasks, audit, and AI control boundaries matter more than CRM depth in v1
5. foundational GST and TDS support are part of v1 because a very small service company or light trading business should be able to operate on the shared foundation
6. reports are necessary in v1, but manual data-entry UIs for most business documents are not
7. the primary short-term objective is to observe AI agents operating correctly on top of strong foundations, not to rush broad production rollout

## Legacy reference rule

When `plan_docs/` and `implementation_plan/` differ:

1. follow `plan_docs/`
2. treat `implementation_plan/` as historical or slice-specific reference
3. only preserve legacy direction from `implementation_plan/` when it does not conflict with the thin-v1 plan

## Implementation clarification rule

During implementation work:

1. use `plan_docs/` as the first source for scope, priority, and canonical thin-v1 rules
2. when active docs leave a modeling detail, ownership detail, sequencing detail, or edge-case interpretation unclear, consult the relevant `implementation_plan/` document for additional clarity
3. when active and legacy docs appear to conflict, resolve the work in favor of `plan_docs/`
4. if a legacy rule is still needed for current implementation and does not conflict with thin-v1 scope, promote or restate that rule in `plan_docs/` rather than relying on the legacy file implicitly
5. do not use legacy detail to quietly re-expand thin-v1 scope, but do use it to avoid avoidable ambiguity and to resolve slice-specific questions pragmatically
