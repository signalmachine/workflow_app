# workflow_app Workflows

Date: 2026-03-28
Status: Active durable workflow reference
Purpose: provide a durable operator-workflow and feature-reference layer that survives beyond the thin-v1 implementation phase and supports testing, review, onboarding, and later user-guide preparation.

## 1. Role of this folder

This folder is for durable workflow-reference material.

It is intended to survive after the current implementation and validation phase is complete.

Use it for:

1. supported end-to-end workflow catalogues
2. durable feature and operator-surface references
3. reusable workflow checklists for live review and testing
4. source material for later user guides, onboarding guides, and release-readiness review

## 2. Boundary with `new_app_docs/`

Keep the document roles distinct.

`new_app_docs/` remains the canonical planning source for:

1. scope
2. milestones
3. sequencing
4. active implementation slices
5. implementation status and next steps

`docs/workflows/` is the durable reference layer for:

1. workflows that exist or are intentionally documented
2. supported operator paths and review surfaces
3. workflow-level validation checklists
4. workflow-validation and live-review tracking
5. feature continuity that should remain understandable after implementation planning moves on

Operational policy:

1. workflow testing and validation may be deferred when urgent product fixes or bounded implementation slices need to land first
2. those product fixes should be planned and tracked in `new_app_docs/`, not in this folder
3. this folder should record the workflow-level result as fixed, pending, blocked, or deferred rather than becoming the implementation-plan source for the fix itself

If a planning document and a workflow-reference document disagree:

1. use `new_app_docs/` as the source of truth for active implementation status and planned next work
2. update `docs/workflows/` when the implemented or supported workflow reference has drifted

## 2.1 Workflow-mapping rule

Everything meaningful in `workflow_app` should tie to one or more workflows.

That does not mean every component is itself a workflow.

It means:

1. every meaningful feature should support, constrain, observe, or expose one or more workflows
2. every meaningful state transition should be understandable in workflow terms
3. every review surface should exist in service of workflow continuity, control, or inspection
4. support seams such as auth, audit, queue control, attachment handling, tool policy, and reporting joins should be justified by the workflows they enable or protect

If a capability cannot be tied clearly to one or more workflows:

1. treat it as a design smell by default
2. require explicit justification before expanding it

## 3. Current starting documents

1. `application_workflow_catalog.md`
2. `end_to_end_validation_checklist.md`
3. `workflow_validation_track.md`

## 4. Maintenance rule

When user-visible workflow behavior, workflow status, or review-surface continuity changes materially:

1. update the relevant canonical planning docs in `new_app_docs/`
2. update the relevant workflow-reference docs in this folder when the durable workflow reference has drifted
3. if the change came from a workflow-review finding, keep the fix-plan details in `new_app_docs/` and keep only workflow-level status and evidence here
