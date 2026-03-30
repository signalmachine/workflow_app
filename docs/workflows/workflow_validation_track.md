# workflow_app Workflow Validation Track

Date: 2026-03-30
Status: Active validation track, separate from implementation planning
Purpose: keep workflow testing, live review, and readiness evidence on a workflow-validation track in `docs/workflows/` rather than inside the normal product-implementation planning stream in `new_app_docs/`.

## 1. Why this document exists

Workflow validation and product implementation are related, but they are not the same track.

This document exists to keep them separate:

1. `new_app_docs/` remains the implementation-planning source for product changes
2. `docs/workflows/` becomes the workflow-validation and review track
3. when workflow review finds a real product gap, the resulting fix plan should be added back into `new_app_docs/`
4. this track may be deferred temporarily when urgent implementation fixes or bounded product slices need to land first

## 2. Track rule

Use this folder for:

1. workflow validation order
2. bounded live-testing and review checklists
3. pass or fail evidence
4. readiness conclusions
5. validation blockers discovered on the real `/app` plus `/api/...` seam
6. workflow-level status such as fixed, pending, blocked, or deferred

Do not use this folder for:

1. broad product implementation planning
2. architectural scope expansion
3. feature-milestone sequencing unrelated to workflow validation
4. detailed fix implementation tracking once a workflow issue has been identified

## 3. Current deferred validation order

The implementation track is currently prioritized ahead of resumed live workflow review.

Current order:

1. complete the bounded web visual-refresh slice from `new_app_docs/web_visual_refresh_plan.md`
2. complete the next planned browser-surface product slices from `new_app_docs/operator_communication_and_intake_surfaces_plan.md`
3. then resume the deferred live workflow validation on the real seam

## 4. Current workflow-validation backlog

Deferred live workflow validation should resume with:

1. draft request -> continue editing -> queue -> process -> downstream request and proposal continuity
2. processed proposal -> request approval -> approval decision -> downstream approval and document continuity
3. failed provider or failed processing path -> failure visibility -> operator troubleshooting continuity

## 5. Issue-handling rule

When workflow review finds a real defect or missing support seam:

1. record the validation result here and in the relevant checklist evidence
2. add the bounded fix plan to `new_app_docs/`
3. implement and verify that fix on the implementation track
4. keep this workflow track limited to issue status and validation evidence while that implementation work happens
5. then return here and rerun the affected workflow validation

## 6. Related documents

Use this document together with:

1. `application_workflow_catalog.md`
2. `end_to_end_validation_checklist.md`
3. the active implementation plans in `new_app_docs/`
