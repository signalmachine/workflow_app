# workflow_app Milestone 12 Admin Maintenance and Master Data Plan

Date: 2026-04-02
Status: Active milestone with Slice 1 through Slice 5 implemented in code and only later follow-on controls still queued
Purpose: define the first real privileged maintenance surface for browser operators so `Admin` stops being a placeholder route directory and becomes the controlled entry point for manual master-data and configuration work on the shared backend.

## 1. Why this milestone exists

The current promoted browser layer now exposes `Settings` and access-scoped `Admin`, but those surfaces are still mostly posture rather than real maintenance capability.

Implementation review shows a real operator gap:

1. `/app/admin` is correctly admin-only, but it is still only a directory of review routes
2. `/app/settings` is currently a user utility surface and should not become the default home for org-scoped configuration or privileged maintenance
3. the backend foundation already supports manual creation of some key records such as ledger accounts, tax codes, and parties, but those seams are not yet exposed through browser or API maintenance flows
4. operators still need a bounded manual fallback for foundational setup and exception handling even in an AI-agent-first product

This milestone exists to close that gap without reopening broad CRM or broad manual-entry ERP drift.

## 2. Planning decision

The correct posture is:

1. keep `Settings` as a user-scoped utility surface for personal, session, and home-surface preferences
2. make `Admin` the privileged surface for org-scoped maintenance, access management, policy controls, and operational configuration
3. expose only the minimum manual maintenance breadth that supports foundational setup, controlled exception handling, and operator continuity on the existing shared truth model
4. prefer dedicated admin-maintenance pages and shared service contracts over hiding privileged writes behind the route catalog or broadening ordinary review pages into mixed review-plus-edit surfaces

The correct target is not:

1. turning `Settings` into a catch-all admin workspace
2. introducing a broad CRM console under the label of customer setup
3. bypassing approval, posting, audit, or domain-service boundaries
4. creating a browser-only maintenance backend separate from the shared domain services

## 3. Scope

In scope:

1. turning `/app/admin` into a real admin landing page with grouped maintenance families instead of a placeholder route list
2. keeping `/app/settings` user-scoped while adding clear continuity links to admin-only maintenance for authorized actors
3. browser and API maintenance seams for foundational accounting master data:
4. ledger accounts
5. tax codes
6. accounting periods and close controls where the current services already support them
7. browser and API maintenance seams for foundational party records:
8. customer parties
9. vendor parties where the current shared party model or adjacent support seams justify them
10. bounded listing, exact detail, and create flows for those records
11. explicit admin-only authorization and audit visibility for all privileged writes
12. documentation and workflow-validation updates for the new privileged maintenance posture

Out of scope:

1. broad CRM opportunity, pipeline, or estimate management
2. broad customer-profile enrichment unrelated to document, accounting, inventory, or execution correctness
3. turning review pages into generic spreadsheet-style maintenance consoles
4. bypassing the shared party model by creating duplicate customer truth in a browser-only layer
5. introducing a second frontend architecture or separate frontend toolchain

## 4. Product and architecture rules

1. user-scoped preferences stay in `Settings`
2. org-scoped maintenance belongs in `Admin`
3. admin-maintenance routes should use standard server-rendered forms and shared `internal/app` transport seams
4. business validation, authorization, and write ownership stay in domain services
5. customer maintenance should reuse the existing `parties` model rather than promoting a separate CRM module
6. accounting master-data maintenance should reuse the existing `accounting` service contracts rather than introducing transport-local write logic
7. privileged maintenance writes must remain auditable and role-gated
8. creation flows should be bounded and foundational, not broad data-entry products

## 5. Required outcomes

This milestone is complete only when:

1. `Admin` is a real privileged maintenance hub rather than a placeholder review directory
2. `Settings` remains a user-scoped utility page and does not silently absorb org-scoped maintenance ownership
3. admins can manually create and browse the foundational master-data records needed for controlled setup and exception handling
4. the first browser-admin maintenance flows exist for ledger accounts and customer parties at minimum
5. the shared API seam exposes the same maintenance capabilities for later non-browser or mobile reuse where appropriate
6. all privileged maintenance writes are role-gated and auditable
7. the new maintenance surfaces stay bounded enough that the product remains workflow-centered rather than CRM-centered

## 6. Suggested slices

### 6.1 Slice 1: admin posture correction and maintenance hub

Goal:

1. turn `/app/admin` into the canonical privileged maintenance landing page and keep `/app/settings` clearly user-scoped

Scope:

1. admin landing page taxonomy
2. grouped maintenance families such as accounting setup, party setup, access management, and operational controls
3. settings-page wording and continuity updates for admin actors
4. route-catalog updates so admin-maintenance destinations remain discoverable but access-scoped

Stop rule:

1. stop once the admin-versus-settings ownership split is explicit in routes, page copy, and navigation
2. do not widen into full CRUD implementation until the maintenance families and route posture are accepted

### 6.2 Slice 2: accounting setup maintenance

Goal:

1. expose the first real accounting master-data maintenance flows through the shared browser and API seam

Scope:

1. ledger-account list and create flows
2. tax-code list and create flows
3. accounting-period list, create, and close controls where the current service contract already supports them
4. admin-only API endpoints backing those flows

Guardrail:

1. this slice is about bounded setup and control data, not broad journal-entry editing or posted-truth mutation

### 6.3 Slice 3: customer and party maintenance

Goal:

1. expose the first bounded customer-master maintenance seam without reviving CRM-first product gravity

Scope:

1. customer-party list and create flows
2. exact party detail sufficient for contacts and role or kind visibility
3. vendor-party or adjacent support-record maintenance only where it cleanly fits the shared party model and does not duplicate truth
4. admin-only API endpoints backing those flows

Guardrail:

1. treat these records as support master data for documents and workflows, not as a sales workspace

### 6.4 Slice 4: access and follow-on controls

Goal:

1. add the next bounded set of privileged controls once the core master-data seams above are stable

Scope:

1. user and role management on the promoted browser seam
2. later policy or operational controls that already have shared-backend support
3. only the minimum exact detail or edit surfaces needed for real operator continuity

Guardrail:

1. do not start this slice until the earlier master-data slices are landed or explicitly reprioritized

## 7. Queue position

Recommended queue from the current repository state:

1. complete the separate Milestone 10 browser-review and workflow-continuity closeout sweep first
2. Slice 1 is now implemented in code: `/app/settings` remains explicitly user-scoped, `/app/admin` is now a grouped privileged maintenance hub, and the promoted browser copy now makes the admin-versus-settings ownership split explicit
3. Slice 2 is now implemented in code through `/app/admin/accounting` plus bounded admin-only `/api/admin/accounting/...` seams for ledger-account, tax-code, and accounting-period list or create flows and accounting-period close controls
4. Slice 3 customer and party maintenance is now implemented in code through `/app/admin/parties` plus bounded admin-only `/api/admin/parties` seams for list, create, filtered list, exact detail reads with contact visibility, and exact-detail contact creation
5. Slice 4 access controls are now implemented in code through `/app/admin/access` plus bounded admin-only `/api/admin/access/users` list and provision seams and `/api/admin/access/users/{membership_id}/role` role updates on the shared `identityaccess` service seam
6. Slice 5 inventory master-data maintenance is now implemented in code through `/app/admin/inventory` plus bounded admin-only `/api/admin/inventory/items` and `/api/admin/inventory/locations` seams for item and location list or create flows
7. later follow-on policy or operational controls remain queued only if workflow evidence or operator continuity justifies them after the first five slices settle

## 8. Verification

Before closing any slice in this milestone:

1. add focused `internal/app` HTTP coverage for the new admin-only routes and access control
2. add service or integration coverage for any newly exposed maintenance contracts
3. run `go build ./cmd/... ./internal/...`
4. run the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
5. run `gopls` diagnostics on edited Go files
6. update workflow-validation material when the new admin-maintenance routes materially affect browser-review expectations

## 9. Documentation sync

When this milestone begins or changes:

1. update `new_app_tracker.md` with queue position and slice status
2. update `new_app_execution_plan.md` so milestone order remains explicit
3. update `new_app_implementation_defaults.md` if the settings-versus-admin ownership split changes
4. update `docs/workflows/` once the new maintenance flows affect durable operator behavior or validation checklists

## 10. Current checkpoint

Slice 1 through Slice 5 are now implemented in code.

Implemented outcome:

1. `/app/settings` now states the user-scoped ownership rules directly, keeps the page focused on session context plus personal continuity, and shows an explicit admin handoff only for admin actors
2. `/app/admin` now presents grouped maintenance families for accounting setup, party setup, access or governance posture, and inventory setup instead of acting as only a flat review-link directory
3. `/app/admin/accounting` now exposes the first real admin-only browser maintenance surface for ledger-account, tax-code, and accounting-period setup while keeping posted-truth accounting review separate
4. the shared API seam now exposes bounded admin-only list and create endpoints for ledger accounts, tax codes, and accounting periods plus period-close controls for later non-browser reuse
5. route-catalog and role-aware home copy now describe `Admin` as the privileged maintenance hub rather than a generic utility page
6. focused `internal/app` HTTP coverage plus DB-backed service and API integration coverage now lock the admin accounting maintenance slice in place
7. `/app/admin/parties` now exposes the first real admin-only browser maintenance surface for customer and vendor support records while keeping CRM-style breadth out of scope
8. exact party detail now also supports bounded contact creation so shared support records can be completed without dropping back to service-only tooling
9. the shared API seam now exposes bounded admin-only `/api/admin/parties` list, filtered list, create, exact detail reads with visible contacts, and `/api/admin/parties/{party_id}/contacts` create support for later non-browser reuse
10. focused `internal/app` HTTP coverage plus DB-backed `internal/app` integration coverage now lock the admin party maintenance slice in place
11. `/app/admin/access` now exposes the first real admin-only browser access-maintenance surface for org users and membership roles while keeping the work bounded to shared identity truth rather than broad identity-product depth
12. the shared API seam now exposes bounded admin-only `/api/admin/access/users` list and provision flows plus `/api/admin/access/users/{membership_id}/role` role updates for later non-browser reuse on the same backend foundation
13. focused `internal/app` HTTP coverage plus DB-backed `internal/app` and `internal/identityaccess` integration coverage now lock the admin access-maintenance slice in place
14. the first access slice also blocks the currently signed-in admin from removing their own admin role accidentally during a role update
15. `/app/admin/inventory` now exposes the first real admin-only browser maintenance surface for inventory item and location setup while keeping downstream stock review, movement review, and reconciliation review separate
16. the shared API seam now exposes bounded admin-only `/api/admin/inventory/items` and `/api/admin/inventory/locations` list and create flows for later non-browser reuse on the same inventory foundation
17. focused `internal/app` HTTP coverage plus DB-backed `internal/app` and `internal/inventoryops` integration coverage now lock the admin inventory maintenance slice in place

Next queued slice:

1. later policy or operational controls should be promoted only if workflow evidence or operator continuity justifies them after the current five maintenance slices settle
