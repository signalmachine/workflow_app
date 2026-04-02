# workflow_app Workflow Validation Track

Date: 2026-04-02
Status: Active validation track, separate from implementation planning; Milestone 10 Slice 1 through Slice 3 remain in code on the rebuilt modular browser bundle, the grouped render-baseline corrective slice now also enforces the modular bundle as the only active render path, Milestone 11 Slice 1 through Slice 3 have now shifted the promoted shell to the lighter grouped top-bar plus landing-page model with route catalog and utility surfaces, Milestone 12 Slice 2 added the admin-only accounting setup surface at `/app/admin/accounting` plus shared `/api/admin/accounting/...` maintenance seams, Milestone 12 Slice 3 now also adds the admin-only party setup surface at `/app/admin/parties` plus shared `/api/admin/parties` maintenance seams, and the remaining Milestone 10 closeout work should now run as one larger browser-review plus workflow-continuity sweep with one grouped corrective follow-up only if that sweep finds tightly related defects
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

1. run one larger Milestone 10 closeout sweep that covers desktop browser review, narrow-width browser review, and focused workflow continuity across the full promoted route family using the current lighter grouped top-bar shell, landing pages, route catalog, and utility surfaces
2. if that sweep is clean, mark the Milestone 10 closeout evidence complete and resume the deferred live workflow validation on the real seam with the rebuilt browser baseline
3. if that sweep finds real defects, group tightly related findings into one bounded corrective fix plan in `new_app_docs/` rather than scattering many tiny follow-up slices across the browser surface

## 3.1 Milestone 10 closeout checklist

Milestone 10 should be treated as closed only when the larger closeout sweep below has explicit pass or blocker evidence recorded on this workflow track.

Closeout sweep:

1. browser-review pass on desktop for `/app/login`, `/app`, `/app/routes`, `/app/settings`, `/app/admin` for an admin actor, `/app/admin/accounting` for an admin actor, `/app/admin/parties` for an admin actor, `/app/operations`, `/app/review`, `/app/inventory`, `/app/submit-inbound-request`, `/app/operations-feed`, `/app/agent-chat`, `/app/inbound-requests/{request_reference_or_id}`, `/app/review/inbound-requests`, `/app/review/approvals`, `/app/review/proposals`, `/app/review/documents`, `/app/review/accounting`, `/app/review/inventory`, `/app/review/work-orders`, and `/app/review/audit`
2. browser-review pass on a narrow-width viewport for that same promoted route family
3. focused continuity pass from exact request detail into proposal detail, approval detail, and document detail
4. focused continuity pass from exact request detail or proposal detail into one downstream accounting or inventory or work-order drill-down surface
5. explicit confirmation that no promoted Milestone 10 route still depends on the retired legacy active template baseline; the code path now enforces this by rejecting unmapped web page data instead of falling back to the old monolithic template, so the remaining work is route-family review evidence rather than render-path ambiguity
6. explicit confirmation that any defect found during this review is either fixed and revalidated or recorded as a blocker before milestone closeout

Evidence rule:

1. a short pass or fail note per checklist item is sufficient
2. if the sweep finds defects, record the failing route family or workflow edge and the promoted grouped fix-plan reference in `new_app_docs/`
3. do not mark Milestone 10 complete in `new_app_docs/` until all six items above have pass evidence

## 3.2 Browser-review execution plan

Use one bounded browser-review pass instead of broad exploratory clicking.

Review posture:

1. run the real app on the shared `/app` seam with a real admin actor available
2. review every promoted route family on desktop first, then rerun the same family on a narrow-width viewport
3. treat this as a presentation and operator-continuity review, not just a route-smoke check
4. record short pass or blocker evidence as you go instead of waiting for one summary at the end

Desktop viewport target:

1. use a full desktop browser window around 1280 to 1440 pixels wide

Narrow-width viewport target:

1. use one narrow responsive width around 390 to 430 pixels wide

For every reviewed route, confirm:

1. the page renders without clipped chrome, overlapping controls, or obvious spacing collapse
2. the active top-bar shell remains readable and the current destination is visually clear
3. the primary page action or workflow start point is visible without hunting through decorative framing
4. tables, filters, route-directory links, and continuity actions remain visually primary over supporting copy
5. no table, code block, metadata row, or inline link band forces unreadable horizontal overflow beyond the intended contained scroll areas
6. the route still reads like an operator application surface rather than a card-heavy editorial page

## 3.3 Route-family assertions

Use the following assertions during the section 3.1 sweep.

### 3.3.1 Operator entry and utility surfaces

Routes:

1. `/app/login`
2. `/app`
3. `/app/routes`
4. `/app/settings`
5. `/app/admin` for an admin actor
6. `/app/admin/accounting` for an admin actor
7. `/app/admin/parties` for an admin actor
8. `/app/submit-inbound-request`
9. `/app/operations-feed`
10. `/app/agent-chat`

Assertions:

1. login is visibly simple and thin, with no promotional split-layout posture
2. home behaves like an operator start surface with clear next actions, not a generic dashboard mosaic
3. route catalog search returns useful route matches for operator-intent queries such as `pending approvals` and `failed requests`
4. settings, admin, admin accounting setup, and admin party setup feel secondary to workflow destinations and do not compete with the main shell
5. intake, operations-feed, and agent-chat each present one clear primary action without burying it under supporting copy

### 3.3.2 Landing pages and navigation scaling

Routes:

1. `/app/operations`
2. `/app/review`
3. `/app/inventory`

Assertions:

1. each landing page behaves as a compact route directory first
2. grouped links are easier to scan than the older card-mosaic posture
3. summary content stays subordinate to route selection
4. the top shell plus landing-page composition makes route discovery easier without turning the shell into a flat site map

### 3.3.3 Review workbench family

Routes:

1. `/app/review/inbound-requests`
2. `/app/review/approvals`
3. `/app/review/proposals`
4. `/app/review/documents`
5. `/app/review/accounting`
6. `/app/review/inventory`
7. `/app/review/work-orders`
8. `/app/review/audit`

Assertions:

1. filters appear early enough on the page to support scanning and narrowing without excess scrolling
2. summary bands do not overpower the main review table
3. table headers, status pills, and exact drill-down links remain readable at desktop and narrow width
4. any horizontal overflow is contained inside the intended table wrapper rather than breaking the page
5. the page hierarchy still makes the exact continuity link obvious for the next operator step

### 3.3.4 Detail-route family and workflow continuity

Routes:

1. `/app/inbound-requests/{request_reference_or_id}`
2. one approval detail route
3. one proposal detail route
4. one document detail route
5. at least one downstream accounting or inventory or work-order detail route

Assertions:

1. the detail page remains single-column and readable rather than collapsing into equal-weight panels
2. request evidence, execution trace, and downstream continuity links remain easy to find
3. upstream and downstream exact links can be followed without losing context
4. no detail section becomes unreadable on narrow width because of inline metadata density or uncontained content

## 3.4 Evidence format

Record evidence in this document or the active validation notes using one short line per checked item.

Preferred format:

1. `pass: <route or route family> - <short reason>`
2. `blocker: <route or workflow edge> - <short defect summary> - <follow-up plan if promoted>`

Example pass notes:

1. `pass: /app/review/inbound-requests desktop - filters and contained table stay visually primary`
2. `pass: /app/login narrow - login remains thin and readable without promotional framing`

Example blocker notes:

1. `blocker: /app/review/accounting narrow - table action links wrap into unreadable stacks - promote one grouped review-table corrective slice`
2. `blocker: request detail to proposal detail continuity - proposal link visually buried below AI trace sections - investigate detail-page hierarchy follow-up`

## 4. Current workflow-validation backlog

Deferred live workflow validation should resume with:

1. draft request -> continue editing -> queue -> process -> downstream request and proposal continuity
2. processed proposal -> request approval -> approval decision -> downstream approval and document continuity
3. failed provider or failed processing path -> failure visibility -> operator troubleshooting continuity

Immediate Milestone 10-first order before the broader backlog resumes:

1. complete the larger Milestone 10 closeout sweep in section 3.1
2. if the sweep passes, mark Milestone 10 complete in the canonical planning docs
3. if the sweep fails, promote one grouped corrective slice for the related defects and then rerun the affected parts of the sweep
4. then resume the broader deferred workflow-validation backlog listed above

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
