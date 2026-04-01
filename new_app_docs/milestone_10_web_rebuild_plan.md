# workflow_app Milestone 10 Web Rebuild Plan

Date: 2026-04-01
Status: Proposed canonical milestone
Purpose: define the first active v2 milestone: a full web-layer rebuild that replaces the current monolithic browser implementation with a modular, scalable server-rendered application layer while preserving the shared backend truth model and forcing implementation through a small number of large related slices.

## 1. Why this milestone exists

The current browser layer has crossed the point where another bounded polish pass is the weaker engineering path.

Thin-v1 is complete.

The current implementation proved that the shared backend seam is credible and that the thin-v1 browser surface is operational.

It also proved that the current web implementation shape is not a good long-term foundation:

1. the browser surface is still concentrated into one very large shared template and style block
2. meaningful UI improvement now requires high-churn edits across too many unrelated screens at once
3. the current shell, page hierarchy, and page-composition model are too dense to scale cleanly as the app matures
4. the implementation cost of continued incremental cleanup is now higher than a disciplined rebuild on the same backend seam

This milestone exists so the repository explicitly chooses the stronger path:

1. rebuild the web layer structure
2. keep the shared backend foundation
3. preserve workflow semantics and approval boundaries
4. avoid treating a major rebuild as vague “UI polish”
5. begin the active v2 phase with a milestone that materially improves usability and production-readiness leverage without reopening foundation design

## 2. Planning decision

Milestone 10 is the first active v2 milestone.

It is the correct next-planning shape if the team wants a materially stronger browser application and a stronger production-ready application surface without changing the backend truth model.

This milestone should:

1. replace the current monolithic template architecture
2. establish a modular page and component structure for the server-rendered web layer
3. simplify shell, navigation, hierarchy, and page rhythm across the browser surface
4. preserve the shared `/app` plus `/api/...` seam and the existing workflow doctrine

This milestone is not:

1. a SPA rewrite
2. a Node or frontend-build-pipeline promotion
3. a second business backend
4. a workflow-model redesign
5. a disguised foundation rewrite

## 3. Objective

Rebuild the completed thin-v1 web layer so it becomes a maintainable, scalable, and production-ready operator application surface rather than a large single-template implementation that is merely functionally complete.

Milestone 10 should produce:

1. a modular server-rendered web architecture
2. a calmer and more structured operator shell
3. a coherent page taxonomy for dashboard, intake, review, detail, and communication surfaces
4. a reusable UI-composition model that lowers future implementation and regression risk
5. a browser surface that remains aligned with the AI-agent-first, review-first, approval-first doctrine
6. the first explicit v2 usability and production-readiness baseline for later broader application enhancement

## 4. Architectural stance

The rebuild should keep the existing product and backend doctrine intact.

The correct target is:

1. the same shared Go backend
2. the same domain services and reporting read paths
3. the same request -> AI -> recommendation -> approval -> document -> posting or execution continuity
4. the same browser-session and shared auth foundation
5. the same server-rendered stack, optionally with `htmx` or `Alpine.js` only where the canonical defaults already allow them

The rebuild should change:

1. template architecture
2. page composition
3. shared UI primitives
4. shell and navigation model
5. page hierarchy and density management

The rebuild should not change the truth owner.

The browser layer must still orchestrate and expose the existing shared backend rather than becoming a second system of workflow logic.

## 5. Scope

In scope:

1. replacing the current inline-template web implementation with modular templates and shared UI partials
2. moving shared styling into a maintainable asset structure suitable for server-rendered reuse
3. defining one explicit shell and navigation model for the promoted browser layer
4. rebuilding dashboard, intake, review, detail, operations-feed, and agent-chat pages on the new structure
5. standardizing page-header, filter, action, card, table, detail-block, and empty-state patterns
6. improving responsive behavior and narrow-width continuity as part of the rebuild rather than as a later patch
7. preserving route continuity where practical and documenting intentional route or template changes when needed

Out of scope:

1. introducing a separate frontend toolchain
2. introducing a separate browser-only backend
3. changing workflow semantics under the label of UI work
4. broad new feature addition unrelated to the rebuild
5. v2 mobile-product work
6. consumer-chat-style product expansion

## 6. Promoted route baseline

Unless the canonical planning set is updated explicitly before implementation, Milestone 10 should treat the following as the promoted thin-v1 browser baseline that must be covered by the rebuild:

1. `/app`
2. `/app/login`
3. `/app/submit-inbound-request`
4. `/app/operations-feed`
5. `/app/agent-chat`
6. `/app/inbound-requests/{request_reference_or_id}`
7. `/app/review/inbound-requests`
8. `/app/review/approvals`
9. `/app/review/approvals/{approval_id}`
10. `/app/review/proposals`
11. `/app/review/proposals/{recommendation_id}`
12. `/app/review/documents`
13. `/app/review/documents/{document_id}`
14. `/app/review/accounting`
15. `/app/review/accounting/{entry_id}`
16. `/app/review/accounting/control-accounts/{account_id}`
17. `/app/review/accounting/tax-summaries/{tax_code}`
18. `/app/review/inventory`
19. `/app/review/inventory/{movement_id}`
20. `/app/review/inventory/items/{item_id}`
21. `/app/review/inventory/locations/{location_id}`
22. `/app/review/work-orders`
23. `/app/review/work-orders/{work_order_id}`
24. `/app/review/audit`
25. `/app/review/audit/{event_id}`

Route rule:

1. these routes define the parity baseline for Slice 1 through Slice 3
2. route additions or removals during Milestone 10 require a canonical plan update before implementation
3. if a route is intentionally replaced, document the redirect or compatibility handling in the same change

## 7. Required outcomes

Milestone 10 is complete only when:

1. the monolithic web template is retired from active use
2. the browser surface is composed from modular templates or partials with clear ownership boundaries
3. the browser shell distinguishes primary workflow destinations from secondary review and support destinations
4. the dashboard has one clear operator story rather than many equal-weight focal points
5. review pages follow one coherent hierarchy for header, filters, summary, table, and row actions
6. detail pages clearly separate primary facts, next actions, secondary context, and deep trace or payload material
7. desktop and narrow-width browser review pass on the rebuilt surfaces
8. the rebuilt web layer remains on the same shared backend truth and workflow semantics

## 8. Slice planning rule

Milestone 10 must not begin as an open-ended page-by-page rebuild.

Implementation rule:

1. implementation must proceed through a small number of large related slices rather than many medium or page-local passes
2. each promoted slice must have its own canonical plan in `new_app_docs/` before implementation begins
3. each slice plan must define scope, included surfaces, excluded surfaces, verification, stop rule, and documentation-sync expectations
4. if a proposed change does not fit the currently accepted slice plan, either defer it to the next slice or update the canonical slice plan first
5. Milestone 10 remains one milestone, but execution should stay tightly governed by these pre-written slice plans

Source-naming rule:

1. milestone and slice labels are planning language only and must not be encoded into long-lived production source filenames, package names, template names, or exported identifiers
2. name implementation files and symbols by the domain surface or responsibility they own, for example `web_review_templates` or `review_workbench`, rather than `slice1`, `slice2`, or `milestone10`
3. temporary migration helpers may reference old versus new rendering paths, but they should still use responsibility-based names instead of phase labels so the names remain valid after the milestone closes

## 9. Planned large slices

Milestone 10 should be implemented through three large slices of related activity.

### 9.1 Slice 1: architecture, shell, and operator-entry surfaces

Goal:

1. establish the new modular rendering foundation and land the new shell on the operator-entry page family that defines the app's first impression and navigation language

Required outcomes:

1. shared shell, navigation, flash-message, page-header, filter-bar, action-group, table-container, summary-card, detail-block, and empty-state partials exist
2. browser pages are no longer authored as one giant inline template
3. shared CSS and shared UI primitives have one maintainable home and one clear ownership model
4. the migration path allows temporary coexistence while pages are rebuilt incrementally
5. the rebuilt shell, sign-in surface, dashboard, intake page, operations feed, and agent-chat surfaces all run on the new structure
6. the dashboard establishes one clear operator story and the shell distinguishes primary workflow destinations from secondary review destinations

Canonical slice plan:

1. `milestone_10_slice_1_architecture_and_operator_entry_plan.md`

### 9.2 Slice 2: review workbench family rebuild

Goal:

1. rebuild the list and summary driven review surfaces as one coherent operator workbench family on top of the new shell and shared review-page primitives

Required outcomes:

1. inbound requests, approvals, proposals, documents, accounting, inventory, work orders, and audit pages use one disciplined review-page structure
2. filters are compact and subordinate to the data they control
3. summary cards appear only where they materially improve scanability and decision speed
4. row actions and cross-links remain strong continuity tools without dominating the page
5. responsive behavior and narrow-width containment are rebuilt consistently across the review family rather than patched page by page

Canonical slice plan:

1. `milestone_10_slice_2_review_workbench_plan.md`

### 9.3 Slice 3: detail continuity and closeout rebuild

Goal:

1. rebuild the detail surfaces and complete parity closeout so the new architecture becomes the only active browser baseline

Required outcomes:

1. request detail uses one clear primary summary area, one lifecycle-action area, and lower-emphasis trace sections
2. approval, proposal, document, accounting, inventory, work-order, and audit detail pages follow a more disciplined primary-versus-secondary structure
3. deep payloads, metadata, and verbose trace blocks are de-emphasized or collapsed where appropriate
4. the detail pages remain strong continuity hubs without reading as one long flat stack
5. rebuilt detail surfaces cover the promoted thin-v1 browser routes cleanly
6. obsolete shared template code and duplicate styling are removed
7. focused tests, docs, and workflow references are updated for the new browser baseline

Canonical slice plan:

1. `milestone_10_slice_3_detail_and_closeout_plan.md`

## 10. Recommended execution order

Implement in this order:

1. Slice 1: architecture, shell, and operator-entry surfaces
2. Slice 2: review workbench family rebuild
3. Slice 3: detail continuity and closeout rebuild

Reason:

1. the main risk is structural, so the architecture foundation and shell should land before broad page-family migration begins
2. the shell and entry surfaces define the page language the later review and detail surfaces should inherit
3. review pages should move together as one workbench family so operators do not bounce between old and new list-page models for too long
4. detail pages and legacy removal should close out together so parity, cleanup, and browser validation happen on one coherent final slice

## 11. Route and compatibility stance

Route stability matters because the current browser surface already underpins user guides, workflow validation, and integration coverage.

Default rule:

1. keep the current route map unless a route change clearly improves long-term coherence

If a route change is proposed:

1. document it in the milestone plan before implementation
2. update user guides, workflow docs, and tests in the same change
3. prefer redirects or compatibility handling where the old path already appears in active docs or test flows

## 12. Validation expectations

For every Milestone 10 implementation slice:

1. update tests appropriate to the rebuilt page family or shared template seam
2. run `gopls` diagnostics on edited Go files
3. run `go build ./cmd/... ./internal/...`
4. run `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
5. run bounded browser review on desktop and narrow-width layouts for the affected rebuilt surfaces
6. sync the canonical docs in `new_app_docs/` and the relevant workflow docs when browser behavior or validation expectations drift
7. when the slice materially changes operator-visible browser behavior, capture bounded browser-review evidence on the `docs/workflows/` track rather than leaving review as undocumented local knowledge

Milestone closeout additionally requires:

1. browser review of the full promoted `/app` surface family
2. a focused workflow continuity pass on the rebuilt browser layer
3. documentation closeout confirming the legacy web architecture is no longer the active baseline
4. durable closeout evidence in `docs/workflows/` for the rebuilt browser baseline or an explicit blocker recorded there

## 13. Guardrails

During Milestone 10:

1. keep the shared backend seam as the single truth owner
2. do not broaden into backend feature work unless the rebuild exposes a real shared-backend blocker
3. do not add a separate frontend runtime or build system without an explicit canonical decision change
4. do not let visual experimentation outrun operator clarity and workflow continuity
5. do not mix unrelated new product breadth into the rebuild milestone
6. keep each promoted slice bounded and documented before implementation starts
7. do not split Slice 2 or Slice 3 into many page-local implementation units unless a concrete blocker forces a replan in the canonical docs

## 14. Relationship to prior UI plans

This milestone supersedes the narrower streamlining posture captured in `thin_v1_archive/web_ui_streamlining_plan.md`.

That older plan remains useful as a problem statement, but it is no longer the preferred implementation shape because the repository has now explicitly chosen a full rebuild instead of another bounded density-cleanup pass on the current structure.

`thin_v1_archive/web_visual_refresh_plan.md` and `thin_v1_archive/web_visual_refresh_follow_up_plan.md` remain historically accurate for the currently landed browser baseline.

Milestone 10 starts from that landed baseline, but replaces its structure.

## 15. Implementation gate

Milestone 10 planning is not implementation-ready until all three slice-plan documents listed above exist and are accepted as the active execution path.

Current state:

1. the milestone is now defined at the canonical level
2. implementation should begin only after the slice-plan documents are present and the tracker records this milestone as the accepted next promoted slice

## 16. Stop rule

Milestone 10 is complete when:

1. the new modular web architecture is the active baseline
2. the rebuilt browser surface covers the promoted thin-v1 workflows cleanly
3. the operator experience is materially calmer and clearer than the current baseline
4. the legacy monolithic template structure is retired
5. the repository is ready to resume browser-led workflow validation on the rebuilt surface rather than continuing indefinite web-layer churn

Milestone 10 must not remain open as an unbounded “frontend modernization” bucket.
