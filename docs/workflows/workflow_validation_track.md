# workflow_app Workflow Validation Track

Date: 2026-04-01
Status: Active validation track, separate from implementation planning; Milestone 10 Slice 1 through Slice 3 remain in code on the rebuilt modular browser bundle, Milestone 11 Slice 1 has now shifted the promoted shell to the lighter top-bar bubble model, and bounded browser-review plus workflow-continuity evidence still need to close out the rebuilt route family before broader live workflow review resumes
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

The implementation track is currently prioritized ahead of resumed live workflow review while the rebuilt Milestone 10 route family settles.

Current order:

1. record bounded manual browser-review evidence for the rebuilt route family on desktop and a narrow-width viewport using the current lighter top-bar shell: `/app/login`, `/app`, `/app/submit-inbound-request`, `/app/operations-feed`, `/app/agent-chat`, `/app/inbound-requests/{request_reference_or_id}`, `/app/review/inbound-requests`, `/app/review/approvals`, `/app/review/proposals`, `/app/review/documents`, `/app/review/accounting`, `/app/review/inventory`, `/app/review/work-orders`, and `/app/review/audit`
2. if that browser review is clean, run one focused workflow-continuity pass across request detail -> proposal -> approval -> document or accounting or inventory or work-order drill-down on the rebuilt route family
3. if those checks are clean, resume the deferred live workflow validation on the real seam with the rebuilt browser baseline
4. if browser review or workflow continuity finds a real defect, add the bounded corrective fix plan back into `new_app_docs/` before treating Milestone 10 as closed

## 3.1 Milestone 10 closeout checklist

Milestone 10 should be treated as closed only when the checklist below has explicit pass or blocker evidence recorded on this workflow track.

Closeout checklist:

1. browser-review pass on desktop for `/app/login`, `/app`, `/app/submit-inbound-request`, `/app/operations-feed`, `/app/agent-chat`, `/app/inbound-requests/{request_reference_or_id}`, `/app/review/inbound-requests`, `/app/review/approvals`, `/app/review/proposals`, `/app/review/documents`, `/app/review/accounting`, `/app/review/inventory`, `/app/review/work-orders`, and `/app/review/audit`
2. browser-review pass on a narrow-width viewport for that same promoted route family
3. focused continuity pass from exact request detail into proposal detail, approval detail, and document detail
4. focused continuity pass from exact request detail or proposal detail into one downstream accounting or inventory or work-order drill-down surface
5. explicit confirmation that no promoted Milestone 10 route still depends on the retired legacy active template baseline
6. explicit confirmation that any defect found during this review is either fixed and revalidated or recorded as a blocker before milestone closeout

Evidence rule:

1. a short pass or fail note per checklist item is sufficient
2. if an item fails, record the failing route or workflow edge and the promoted fix-plan reference in `new_app_docs/`
3. do not mark Milestone 10 complete in `new_app_docs/` until all six items above have pass evidence

## 4. Current workflow-validation backlog

Deferred live workflow validation should resume with:

1. draft request -> continue editing -> queue -> process -> downstream request and proposal continuity
2. processed proposal -> request approval -> approval decision -> downstream approval and document continuity
3. failed provider or failed processing path -> failure visibility -> operator troubleshooting continuity

Immediate Milestone 10-first order before the broader backlog resumes:

1. complete the Milestone 10 closeout checklist in section 3.1
2. if the checklist passes, mark Milestone 10 complete in the canonical planning docs
3. then resume the broader deferred workflow-validation backlog listed above

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
