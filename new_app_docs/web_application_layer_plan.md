# workflow_app Web Application Layer Plan

Date: 2026-03-21
Status: Planned v1 implementation slice
Purpose: define the promoted web-layer work required so `workflow_app` is usable as an application in v1 rather than only through service-layer seams, minimal API contracts, or direct developer tooling.

## 1. Promotion decision

The earlier thin-v1 planning set intentionally kept web work minimal and deferred broader operational surfaces to v2.

That is now changed.

Current decision:

1. v1 should include a usable web application layer
2. the web layer should support the real provider-backed AI path, not only report on its outputs after the fact
3. this does not remove the approval, posting, audit, or database-first control model
4. this does mean web work is no longer limited to minimal browser-testing seams
5. the preferred thin-v1 implementation stack for that web layer is Go server-rendered HTML on the shared backend, optionally enhanced with `htmx` for partial-page updates and `Alpine.js` for small local UI state, without introducing a Node toolchain

## 2. V1 objective

Land the web-layer foundations and operator-facing surfaces required so users can actually operate the v1 system through the browser with the AI path, review path, approval path, and downstream inspection path all reachable in one coherent application layer.

## 3. V1 scope

In scope:

1. a real HTTP or API application surface
2. session-auth and active-org handling suitable for browser use
3. attachment upload and download contracts on top of the existing attachment persistence model
4. inbound request submission and tracking through browser-usable paths
5. approval queue and approval decision surfaces
6. inbound-request, processed-proposal, document, accounting, inventory, work-order, and audit review surfaces
7. browser-usable initiation and review of the provider-backed AI flow
8. enough navigation, page structure, and operator workflow continuity that the application is usable in practice rather than only technically exposed

Out of scope:

1. portal-product depth
2. polished consumer-chat UX
3. full mobile-product work
4. broad CRM revival
5. broad manual-entry ERP behavior that displaces the AI-agent-first model

## 4. Required web-layer characteristics

The promoted v1 web layer should preserve the core doctrine:

1. AI remains the main operator interface
2. humans still review through explicit queues and read models
3. financially meaningful actions stay approval-gated and auditable
4. the browser layer should orchestrate domain services rather than become a second truth owner
5. the web layer should make the request -> AI -> recommendation -> approval -> document -> posting or execution chain understandable without raw SQL or developer-only tools
6. the promoted web layer should use backend contracts that a later v2 mobile client can also reuse, even if the web and mobile clients differ in interaction design and presentation
7. the promoted web layer should prefer progressive enhancement over SPA-style client ownership so the server remains the primary truth owner for rendering and flow control

## 4.1 Preferred implementation stack

Preferred thin-v1 stack:

1. Go `net/http` plus Go `html/template` as the rendering baseline
2. shared backend handlers and reporting reads as the only truth-owning web backend
3. `htmx` for incremental updates such as filter refresh, inline detail loads, queue actions, and other partial-page interactions
4. `Alpine.js` only where a small amount of local client state materially improves usability
5. no required Node, npm, or separate frontend build pipeline in thin v1 unless a later canonical planning update explicitly changes that rule

## 5. Minimum usable v1 web outcomes

V1 web work is not complete until the following are true:

1. a user can sign in and operate in one active org context through the browser
2. a user can submit an inbound request with attachments
3. a user can observe queued, processing, processed, failed, cancelled, and completed request states
4. a user can inspect resulting AI artifacts, recommendations, and approval needs
5. an authorized user can act on approval queues through the browser
6. downstream documents and review surfaces are reachable without dropping back to service-layer-only tooling
7. the application is usable enough that real operator testing does not depend on bespoke scripts or direct database access

## 6. Sequencing guidance

Recommended order after the remaining Milestone 5 reporting polish:

1. complete the provider-backed AI execution milestone
2. land the minimal API and attachment transport contracts required for the live path
3. build the usable web application layer on top of those live contracts and existing reporting reads
4. keep broad mobile, chat-product, portal, and multimodal-polish work deferred until after this v1 web layer is credible

Execution rule:

1. treat this milestone as an umbrella for multiple small vertical slices, not as one monolithic web build
2. prefer one usable end-to-end operator loop first, then widen surrounding web coverage incrementally on the same shared backend contracts
3. Milestone 7 is primarily about connecting the existing core application engine and shared backend seams to a usable browser layer, not about adding unrelated new backend features
4. backend corrections and narrow shared-backend enhancements are still required when the web layer proves a concrete need for correctness, continuity, or operator usability
5. those backend changes must stay inside existing ownership boundaries and must not create a second truth owner, a separate web backend, or unrelated scope expansion

## 7. Success criteria

This slice is complete only when:

1. the browser layer is a real application surface rather than only a testing seam
2. it supports the live AI path and approval flow end to end
3. it remains aligned with the database-first and approval-first control model
4. the application is materially usable by operators in v1

## 8. Current implementation checkpoint

The first Milestone 7 browser slice is now landed.

Current checkpoint:

1. `/app` now provides a real server-rendered browser surface on top of the shared backend seam rather than only API contracts
2. operators can sign in with browser-session auth, submit inbound requests with file attachments, process the next queued request, review recent requests and pending approvals, inspect inbound-request detail, and act on approval decisions from that browser surface
3. inbound-request detail now exposes the request message trail, attachment downloads and derived text, AI runs, artifacts, recommendations, and downstream proposals in one page flow
4. the next browser slice now widens that same surface into downstream document and accounting review through `/app/review/documents` and `/app/review/accounting`, so operators can continue from proposal and approval work into document review plus journal-entry, control-account, and tax-summary review without dropping back to scripts
5. the shared backend seam now also exposes `GET /api/review/documents`, `GET /api/review/accounting/journal-entries`, `GET /api/review/accounting/control-account-balances`, and `GET /api/review/accounting/tax-summaries` through the same session-cookie auth path as the rest of the browser flow
6. approval queue reads and approval decisions now accept the same browser session-cookie auth path as the rest of the browser flow, closing the previous browser-versus-API inconsistency
7. the next browser slice now widens that same surface into inventory, work-order, and audit review through `/app/review/inventory`, `/app/review/work-orders`, `/app/review/work-orders/{work_order_id}`, and `/app/review/audit`, so operators can continue from approvals and accounting into stock, movement, reconciliation, execution-rollup, and audit inspection on the same session-backed application surface
8. the shared backend seam now also exposes `GET /api/review/inventory/stock`, `GET /api/review/inventory/movements`, `GET /api/review/inventory/reconciliation`, `GET /api/review/work-orders`, `GET /api/review/work-orders/{work_order_id}`, and `GET /api/review/audit-events` through the same session-cookie auth path as the rest of the browser flow
9. document review now also supports exact `document_id` drill-down through the same reporting read path, so browser and API callers can land on one downstream business document rather than reopening the full list
10. browser review templates now include direct cross-links from proposals and approval queue rows into the exact downstream document, from document and work-order review into audit lookup, and from inventory reconciliation into the related document and work-order detail, tightening the operator loop without adding a second backend or a new truth owner
11. the latest continuity refinement adds exact work-order review filtering by `document_id` on the shared backend seam plus direct document-to-execution links for work-order documents, movement-level audit links in inventory review, and audit-page entity links back into document, work-order, and inventory inspection so operators can keep traversing one browser flow instead of reopening broad lists by hand
12. the next continuity slice is now landed at `/app/review/proposals`, using the already-landed `GET /api/review/processed-proposals` and `GET /api/review/processed-proposal-status-summary` reads to give operators a dedicated browser review page for processed proposals with proposal-status summary, request-reference filtering, and direct continuation back into request detail and forward into downstream documents
13. inbound-request detail now also renders the already-persisted AI step payloads and delegation traces from the shared reporting seam, so operators can inspect provider-execution and specialist-routing context in the browser instead of falling back to raw API detail
14. the latest continuity slice now extends accounting review with exact source-`document_id` journal drill-down on the shared backend seam plus browser cross-links from document, inventory-reconciliation, and work-order accounting context into the matching accounting page, so operators can stay on one downstream continuity path instead of reopening broad accounting lists by hand
15. the latest continuity slice now extends inventory review with exact `movement_id` drill-down on the shared backend seam plus audit-page movement links into that filtered inventory surface, so operators can move from one audit event back into the exact movement and reconciliation context instead of staying on generic audit results or reopening the full inventory list by hand
16. the latest continuity slice now adds a dedicated `/app/review/approvals` page on top of the already-landed `GET /api/review/approval-queue` seam, with pending-versus-closed and queue-code filtering, browser approval actions from the full review page, dashboard continuity into full approval review, and cross-links from proposal and document review into the matching approval-queue slice
17. the latest continuity slice now adds a dedicated `/app/review/inbound-requests` page on top of the already-landed `GET /api/review/inbound-requests` and `GET /api/review/inbound-request-status-summary` reads, with status and `REQ-...` filtering, status-summary cards that jump into exact browser filters, request-level AI-status context, and direct continuation into request detail so operators can work beyond the dashboard snippet without leaving the shared backend seam
18. the latest continuity slice now extends proposal and approval review with exact `recommendation_id` and `approval_id` drill-down on the shared backend seam, and it extends audit lookup with direct entity links back into exact inbound-request detail, exact approval review, and exact proposal review so operators can move from audit traces back into the precise browser review surface instead of manually reconstructing context
19. the latest continuity slice now adds dedicated `/app/review/approvals/{approval_id}` and `/app/review/proposals/{recommendation_id}` browser detail pages on top of the already-landed approval-queue and processed-proposal read seams, so exact browser drill-downs become real review stops with decision, control-chain, document, request, and audit continuity instead of staying as filtered list states
20. the latest continuity slice now adds a dedicated `/app/review/documents/{document_id}` browser detail page on top of the already-landed document-review read seam, so operators can open one business document as a real browser stop with lifecycle timestamps, approval state, accounting links, execution or inventory continuation where relevant, and audit continuity without bouncing through filtered list pages
21. the latest continuity slice now extends the shared journal-review seam with exact `entry_id` drill-down through both `/api/review/accounting/journal-entries?entry_id=...` and `/app/review/accounting?entry_id=...`, adds a dedicated `/app/review/accounting/{entry_id}` browser detail page, and wires document, approval, inventory-reconciliation, accounting-list, and audit surfaces directly into that exact posting review so operators can move from downstream control context into one journal entry instead of reopening broad accounting lists
22. the latest continuity slice now adds a dedicated `/app/review/inventory/{movement_id}` browser detail page on top of the already-landed inventory-movement and reconciliation reads, and it wires inventory-list plus audit movement links into that exact browser stop so operators can inspect one movement with document, execution, accounting, and audit continuity instead of staying at filtered-list depth
23. the latest continuity slice now turns inbound-request detail into a stronger review hub by wiring request-level AI recommendations and processed proposals into exact proposal review, exact approval review, filtered request review, and direct inbound-request or recommendation audit lookup, so operators can stay on one browser path from intake evidence into downstream control decisions instead of stalling at the request page
24. the latest continuity slice now extends the shared audit-review seam with exact `event_id` drill-down through both `/api/review/audit-events?event_id=...` and `/app/review/audit?event_id=...`, adds a dedicated `/app/review/audit/{event_id}` browser detail page, and wires audit search results into that exact event stop so operators can inspect one persisted audit event with payload and direct entity continuation instead of staying on a broad search list
25. the latest continuity slice now turns accounting tax-summary and control-account tables into browser-usable review pivots by exposing tax-type, control-type, and exact control-account filtering on top of the existing shared review seams, so operators can stay in one accounting review path instead of treating those sections as dead-end summaries
26. the latest continuity slice now turns inventory stock balances into active browser pivots by adding anchored filtered links from stock rows into stock, movement-history, and reconciliation views and by routing inventory item and location audit entities back into those same focused inventory review states instead of leaving stock review as a dead-end table
27. the latest continuity slice now turns accounting control-account balances and tax summaries into dedicated browser review stops through `/app/review/accounting/control-accounts/{account_id}` and `/app/review/accounting/tax-summaries/{tax_code}`, and the shared accounting-review seam now also supports exact `tax_code` filtering so browser and API callers can land on one tax summary directly instead of treating those sections as summary-only tables
28. the remaining Milestone 7 work is now operator continuity, richer drill-downs, and focused refinement on top of those landed review surfaces without creating a second backend or reviving broad manual-entry UI scope
