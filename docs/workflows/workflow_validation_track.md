# workflow_app Workflow Validation Track

Date: 2026-04-10
Status: Active validation track, separate from implementation planning; the current browser runtime is the Milestone 13 served Svelte frontend with the contextual-navigation shell, the promoted workflow, utility, admin, and detail-route families now run on that one Go-served `/app` surface, the inventory landing now hands off into an explicit scoped inventory-review UI for pending execution and accounting follow-through, the operator home plus coordinator chat plus review landing snapshots now also route known proposal and approval rows directly into exact detail pages, exact inbound-request detail plus exact proposal detail now also prefer direct downstream accounting-entry drill-down when a linked proposal document already has a posted journal entry, exact inbound-request detail now also exposes the parked-request lifecycle controls directly on the promoted Svelte route, the desktop shell now keeps a persisted collapsed-sidebar preference and the contextual-tab row starts over the main content column instead of across the far-left edge, the mobile drawer model stays as secondary compatibility rather than an active optimization target, focused automated coverage now also asserts route-catalog operator-intent continuity plus promoted admin accounting and inventory status controls plus exact downstream accounting-entry drill-down plus the promoted login and utility and admin and operations and exact approval or document or accounting detail surfaces, the served handler returns real `/app/_app/...` assets and `404` for missing bundle paths, the live provider seam has been reconfirmed after the 2026-04-09 bounded coordinator corrective pass, `cmd/verify-agent -database-url "$DATABASE_URL" -approval-flow` now also confirms one exact request -> proposal -> approval -> document -> accounting chain on the shared session plus `/api/...` seam for its verification org, and the Milestone 13 desktop browser-review closeout now has explicit real-browser pass evidence rather than an open blocker
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

The implementation track is currently prioritized ahead of resumed live workflow review while the post-cutover Milestone 13 served browser surface settles.

Current order:

1. treat the Milestone 13 browser-review sweep as complete evidence and use the same split admin plus verification-org pattern for future workflow-critical browser reruns
2. resume the deferred live workflow validation backlog on the real seam from a post-closeout baseline rather than treating Milestone 13 browser review as still open work
3. if later browser review finds real defects, group tightly related findings into one bounded corrective fix plan in `new_app_docs/` rather than scattering many tiny follow-up slices across the browser surface
4. record evidence against the current served Svelte runtime and shared API seams, not against the retired template-browser behavior

Implementation note recorded on 2026-04-09:

1. the codebase now has explicit full-handler integration coverage for `/app/_app/version.json` plus missing `/app/_app/...` asset `404` behavior on top of the earlier unit-level shell tests
2. the operator home plus coordinator chat plus review landing snapshots now also use exact proposal and approval drill-down routes when the exact record is already known, which reduces one source of post-cutover continuity friction before the larger real-seam sweep runs
3. exact inbound-request detail plus exact proposal detail now also keep direct downstream accounting follow-through visible when a linked proposal document already exists, which reduces one more source of continuity friction before the larger real-seam sweep runs
4. exact inbound-request detail plus exact proposal detail now also prefer direct downstream accounting-entry detail when the shared review seam already exposes the posted journal entry for that linked proposal document, which reduces one more source of continuity friction before the larger real-seam sweep runs
5. this implementation work reduces serving-path and drill-down ambiguity, but it does not replace the remaining browser-review and workflow-continuity evidence required in section 3.1
6. the desktop shell now also supports a persisted collapsed-sidebar preference, so the post-cutover browser sweep should explicitly review both expanded and collapsed desktop states instead of only the default expanded view
7. narrow-width review may still be used as a bounded compatibility check when warranted, but it is no longer an equal-priority optimization target for the served web runtime because future mobile-first UX belongs to the separate mobile client
8. the OpenAI coordinator-provider path needed one bounded corrective slice before the live provider seam was stable again: after three read-tool executions the coordinator now disables further tool availability and forces the next turn to return the structured brief, which prevents repeated read-loop drift without reopening open-ended autonomy
9. `cmd/verify-agent` now creates its verification actor through the shared browser-session auth path, which removes one verification-harness-only auth mismatch from the live provider check
10. the canonical repo verification, `cmd/verify-agent`, and the dedicated `TestOpenAIAgentProcessorLiveIntegration` live-provider test all passed again on 2026-04-09 after that corrective slice
11. the later focused verification for the exact downstream-accounting detail pass succeeded through `npm --prefix web run check`, focused Svelte route tests, `go build ./cmd/... ./internal/...`, compile coverage for `./internal/reporting`, and focused non-DB `internal/app` tests
12. the full canonical DB-backed suite `set -a; source .env; set +a; timeout 300s go test -p 1 ./cmd/... ./internal/...` also passed cleanly on 2026-04-09, so the earlier `create tax code: unauthorized` and `reset test database: deadlock detected` failures did not reproduce and should currently be treated as transient environment or test-state noise rather than active Milestone 13 blockers
13. a follow-up real-seam validation run on 2026-04-09 also confirmed that `/app` returns the served Svelte shell, `/app/_app/version.json` returns a real static asset, missing `/app/_app/...` assets return `404`, browser-session login still works through `/api/session/login`, route-catalog search still returns `Approval review` for `pending approvals`, and one live request (`REQ-000001`) now successfully moved through submit -> queue -> provider-backed processing -> exact request detail -> exact proposal detail on the shared browser-session seam
14. `cmd/verify-agent -approval-flow` also passed on 2026-04-09 and confirmed that the same live verification request (`REQ-000001`) can continue through one deterministic exact proposal -> approval request -> approval decision -> exact approval detail -> exact document detail chain on the shared browser-session plus `/api/...` seam
15. focused Svelte component coverage added on 2026-04-10 now asserts multi-term route-catalog continuity for `pending approvals`, visible promoted admin accounting and inventory status controls, and exact accounting-entry drill-down from request and proposal detail when the posted journal entry is already known
16. focused Go web-serving coverage added on 2026-04-10 now asserts SPA fallback across the full promoted `/app` route family and exact detail routes, which reduces one more cutover-regression blind spot before the remaining desktop browser sweep
17. an additional focused Svelte route pass added on 2026-04-10 now asserts the promoted login, settings, admin hub, admin access, admin party setup, operations landing, submit-inbound-request, operations feed, and exact approval, document, and accounting detail surfaces before the remaining desktop browser sweep
18. the 2026-04-10 real-browser closeout uncovered one real admin defect and one verification-harness mismatch: admin accounting tax-code creation needed governed control-account selectors rather than raw account-id text fields, and `cmd/verify-agent` needed an explicit `-database-url "$DATABASE_URL"` run plus emitted verification-org credentials and a posted invoice seed so the browser continuity proof could run against the same backend and a real journal-entry chain
19. the 2026-04-10 closeout Playwright sweep then passed on the served `/app` seam with four checks covering desktop shell persistence, route-catalog continuity, admin maintenance plus exact party detail, promoted route-family rendering, and exact request -> proposal -> approval -> document -> accounting continuity using the dedicated verification org seeded by `cmd/verify-agent`
20. a follow-on Milestone 14 Slice 1 checkpoint on 2026-04-10 then restored the parked-request lifecycle controls on exact inbound-request detail and corrected the desktop shell tab-column layout, with focused frontend tests passing through `npm --prefix web test -- src/lib/api/inbound.test.ts src/routes/(app)/inbound-requests/page_detail.test.ts` and `npm --prefix web run check`

## 3.1 Milestone 13 post-cutover checklist

Milestone 13 should be treated as ready for closeout only when the larger post-cutover sweep below has explicit pass or blocker evidence recorded on this workflow track.

Closeout sweep:

1. browser-review pass on desktop for `/app/login`, `/app`, `/app/routes`, `/app/settings`, `/app/admin` for an admin actor, `/app/admin/accounting` for an admin actor, `/app/admin/parties` for an admin actor, `/app/admin/parties/{party_id}` for exact-detail contact creation, `/app/admin/access` for an admin actor, `/app/admin/inventory` for an admin actor, `/app/operations`, `/app/review`, `/app/inventory`, `/app/submit-inbound-request`, `/app/operations-feed`, `/app/agent-chat`, `/app/inbound-requests/{request_reference_or_id}`, `/app/review/inbound-requests`, `/app/review/approvals`, `/app/review/proposals`, `/app/review/documents`, `/app/review/accounting`, `/app/review/inventory`, `/app/review/work-orders`, and `/app/review/audit`
2. optional narrow-width compatibility pass only where a real concern exists for navigation reachability, contained overflow, or obvious operator blockage on the served web surface
3. focused continuity pass from exact request detail into proposal detail, approval detail, and document detail
4. focused continuity pass from exact request detail or proposal detail into one downstream accounting or inventory or work-order drill-down surface, preferring a direct accounting-entry route when the linked document already has a posted journal entry
5. explicit confirmation that no promoted route still depends on the retired template browser path and that missing static-asset requests return a real asset result or a `404` rather than silently falling back to the SPA shell
6. explicit confirmation that any defect found during this review is either fixed and revalidated or recorded as a blocker before milestone closeout

Current evidence recorded on 2026-04-09:

1. `pass: served /app shell - HTTP 200 with the embedded Svelte runtime and /app-based asset imports on the real app server`
2. `pass: static asset handling - /app/_app/version.json returned JSON and missing /app/_app/immutable/entry/missing.js returned HTTP 404 instead of the SPA shell`
3. `pass: browser-session auth - /api/session/login and /api/session both succeeded for a bootstrapped admin actor on the shared auth seam`
4. `pass: route catalog search - /api/navigation/routes?q=pending%20approvals returned the exact Approval review destination`
5. `pass: request to proposal continuity - REQ-000001 submitted through /api/inbound-requests, processed through /api/agent/process-next-queued-inbound-request, and remained review-visible through exact /api/review/inbound-requests/REQ-000001 plus exact processed-proposal detail`
6. `pass: browser-review sweep 2026-04-10 - real Playwright desktop pass on the served /app seam covered the promoted route family, persisted sidebar state, route-catalog intent search, and exact admin continuity without shell or rendering regressions`
7. `pass: request -> approval -> document chain - cmd/verify-agent -approval-flow created a deterministic approval-ready proposal on REQ-000001 and confirmed exact request, proposal, approval, and document continuity through the shared session plus /api/... seam`
8. `pass: focused Svelte closeout coverage 2026-04-10 - route catalog intent search, promoted admin accounting and inventory status controls, and request/proposal exact accounting-entry continuity all have explicit automated assertions`
9. `pass: focused Go route-family coverage 2026-04-10 - promoted /app route families and exact detail routes now all return the served Svelte shell through the shared SPA fallback handler`
10. `pass: focused Svelte route-family coverage 2026-04-10 - login, settings, admin hub, admin access, admin parties, operations landing, submit-inbound-request, operations feed, and exact approval/document/accounting detail surfaces all have explicit component assertions`
11. `pass: request -> proposal -> approval -> document -> accounting chain 2026-04-10 - cmd/verify-agent -database-url "$DATABASE_URL" -approval-flow seeded a dedicated verification org with a posted invoice and journal entry, and the real-browser Playwright pass followed exact request, proposal, approval, document, and accounting detail routes on that served org without continuity breaks`

Evidence rule:

1. a short pass or fail note per checklist item is sufficient
2. if the sweep finds defects, record the failing route family or workflow edge and the promoted grouped fix-plan reference in `new_app_docs/`
3. do not mark Milestone 13 workflow-validation closeout complete in `new_app_docs/` until all six items above have pass evidence

## 3.2 Browser-review execution plan

Use one bounded browser-review pass instead of broad exploratory clicking.

Review posture:

1. run the real app on the shared `/app` seam with a real admin actor available
2. review every promoted route family on desktop first and treat that as the primary browser judgment path
3. treat this as a presentation and operator-continuity review, not just a route-smoke check
4. record short pass or blocker evidence as you go instead of waiting for one summary at the end

Desktop viewport target:

1. use a full desktop browser window around 1280 to 1440 pixels wide

For every reviewed route, confirm:

1. the page renders without clipped chrome, overlapping controls, or obvious spacing collapse
2. the active top-bar shell remains readable and the current destination is visually clear
3. the desktop review should be checked once with the sidebar expanded and once with the persisted collapsed state so the narrower shell still preserves clear major-area navigation
4. the primary page action or workflow start point is visible without hunting through decorative framing
5. tables, filters, route-directory links, and continuity actions remain visually primary over supporting copy
6. no table, code block, metadata row, or inline link band forces unreadable horizontal overflow beyond the intended contained scroll areas
7. the route still reads like an operator application surface rather than a card-heavy editorial page

## 3.3 Route-family assertions

Use the following assertions during the section 3.1 sweep on the current served Svelte runtime.

### 3.3.1 Operator entry and utility surfaces

Routes:

1. `/app/login`
2. `/app`
3. `/app/routes`
4. `/app/settings`
5. `/app/admin` for an admin actor
6. `/app/admin/accounting` for an admin actor
7. `/app/admin/parties` for an admin actor
8. `/app/admin/parties/{party_id}` for exact-detail contact creation
9. `/app/admin/access` for an admin actor
10. `/app/admin/inventory` for an admin actor
11. `/app/submit-inbound-request`
12. `/app/operations-feed`
13. `/app/agent-chat`

Assertions:

1. login is visibly simple and thin, with no promotional split-layout posture
2. home behaves like an operator start surface with clear next actions, not a generic dashboard mosaic
3. route catalog search returns useful route matches for operator-intent queries such as `pending approvals` and `failed requests`
4. the major-area sidebar and contextual section tabs make the current area and local view obvious without reintroducing one flat global route list
5. settings, admin, admin accounting setup including status controls, admin party setup plus the exact party-detail contact surface, admin access controls, and admin inventory setup including status controls feel secondary to workflow destinations and do not compete with the main shell
6. the access-maintenance page keeps provisioning and role updates bounded to shared identity control rather than reading like a broad identity console
7. the inventory-maintenance page stays bounded to item and location setup rather than drifting into generic stock editing or movement correction
8. intake, operations-feed, and agent-chat each present one clear primary action without burying it under supporting copy

### 3.3.2 Landing pages and navigation scaling

Routes:

1. `/app/operations`
2. `/app/review`
3. `/app/inventory`

Assertions:

1. `/app/operations` and `/app/review` behave as compact route directories first, while `/app/inventory` behaves as a thin domain landing with stock, movement, and handoff-entry snapshots
2. grouped links and filtered follow-through actions are easier to scan than the older card-mosaic posture
3. summary content stays subordinate to route selection or the next workflow-follow-through action
4. the sidebar plus contextual-tab composition makes route discovery easier without turning the shell into a flat site map

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
3. table headers, status pills, and exact drill-down links remain readable on desktop, and any optional narrow-width compatibility check should confirm only that the page does not become obviously unusable
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
2. request evidence, execution trace, and downstream continuity links remain easy to find, with request detail keeping the latest proposal plus direct approval, document, and exact downstream accounting-entry drill-down actions near the top of the page when that posted journal entry already exists
3. upstream and downstream exact links can be followed without losing context
4. no detail section becomes unreadable on desktop because of inline metadata density or uncontained content, and any optional narrow-width compatibility check should only guard against obvious breakage
5. any validation note that still refers to server-rendered page composition is rewritten in terms of the served Svelte shell, shared `/api/...` data seams, or explicit workflow continuity behavior

## 3.4 Evidence format

Record evidence in this document or the active validation notes using one short line per checked item.

Preferred format:

1. `pass: <route or route family> - <short reason>`
2. `blocker: <route or workflow edge> - <short defect summary> - <follow-up plan if promoted>`

Example pass notes:

1. `pass: /app/review/inbound-requests desktop - filters and contained table stay visually primary`
2. `pass: /app/login compatibility - narrow-width fallback still avoids obvious overlap or blocked sign-in`
3. `pass: live provider seam 2026-04-09 - cmd/verify-agent, cmd/verify-agent -approval-flow, and TestOpenAIAgentProcessorLiveIntegration all completed after the bounded coordinator read-loop and verify-agent auth-path corrective pass`

Example blocker notes:

1. `blocker: /app/review/accounting narrow - table action links wrap into unreadable stacks - promote one grouped review-table corrective slice`
2. `blocker: request detail to proposal detail continuity - top continuity actions missing or still buried below AI trace sections - promote one grouped detail-page hierarchy corrective slice`

## 4. Current workflow-validation backlog

Deferred live workflow validation should resume with:

1. draft request -> continue editing -> queue -> process -> downstream request and proposal continuity
2. failed provider or failed processing path -> failure visibility -> operator troubleshooting continuity

Immediate Milestone 13-first order before the broader backlog resumes:

1. complete the larger Milestone 13 post-cutover sweep in section 3.1
2. if the sweep passes, mark the Milestone 13 browser-validation closeout complete in the canonical planning docs
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
