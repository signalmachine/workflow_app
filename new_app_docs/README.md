# workflow_app Planning README

Date: 2026-03-19
Status: Draft canonical planning set for the `workflow_app` thin-v1 codebase
Purpose: define the planning baseline for the new repository that replaces the current CRM-heavy codebase direction.

## Why this folder exists

The current repository proved the product idea, but it also proved that the shipped CRM-heavy implementation shape creates ongoing scope tension against the intended thin-foundation v1.

This folder exists to prepare a clean restart with:

1. stricter module boundaries
2. stricter thin-v1 scope control
3. documents, ledgers, execution context, approvals, and reports as the real center of gravity
4. AI-agent-first operation without broad human operational UI

This planning set is the repository's canonical planning source.

For the durable workflow-reference layer that should survive after the active thin-v1 planning phase, use `docs/workflows/`.

`workflow_app` still plans for multi-agent architecture in v1, but only at foundation depth:

1. one coordinator routing bounded work to specialist agents
2. durable run history, tool policy, artifacts, and delegation traces
3. no advanced autonomy features unless they are required for foundation correctness

## Recommended reading order

Read these in order:

1. `new_app_v1_principles.md`
2. `new_app_v1_scope.md`
3. `new_app_schema_and_module_boundaries.md`
4. `new_app_implementation_defaults.md`
5. `new_app_foundation_coverage.md`
6. `new_app_execution_plan.md`
7. `new_app_v1_gap_review_from_current_codebase.md`
8. `adopted_document_ownership_remediation_plan.md`
9. `inbound_request_and_attachment_foundation_plan.md`
10. `ai_provider_execution_plan.md`
11. `web_application_layer_plan.md`
12. `non_browser_auth_evolution_plan.md`
13. `new_app_tracker.md`
14. `app_v2_plans/README.md`
15. `../docs/workflows/README.md` for the durable workflow-reference layer after the active planning read

## Reset decision

The intended reset is:

1. start a new codebase
2. do not continue the current codebase by deleting large slices and trying to reshape the remaining implementation in place
3. treat the current repository as a reference source for selective ideas, not as the implementation base

Reason:

1. the current implementation shape carries too much CRM-first gravitational pull
2. deleting parts of the existing codebase would still leave the new effort spending time on untangling legacy assumptions
3. a clean repository will make it easier to enforce thin-v1 rules from the first migration onward

## Core reset rule

If a capability can wait until v2 without weakening the foundation, it must wait until v2.

## Quality rule

Thin v1 does not mean low quality, low rigor, or simplistic modeling.

It means:

1. foundation-complete first
2. narrow module count
3. narrow workflow breadth
4. high technical quality in the layers that do land
