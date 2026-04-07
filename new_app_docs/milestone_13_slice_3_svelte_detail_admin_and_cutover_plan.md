# workflow_app Milestone 13 Slice 3 Plan

Date: 2026-04-04
Status: In progress; settings continuity, admin-parity, inbound-request detail, promoted review-detail checkpoints, and default `/app` cutover to the built Svelte frontend are implemented in code on 2026-04-04 while bounded post-cutover workflow validation and final legacy-browser retirement stay open
Purpose: define the third Milestone 13 implementation slice so detail surfaces, admin surfaces, parity closeout, cutover, and legacy browser retirement happen together instead of being left as an indefinite cleanup tail.

## 1. Slice role

This slice finishes the migration.

It should:

1. migrate the promoted detail-route family
2. migrate settings and admin-maintenance surfaces needed for current operator continuity
3. complete bounded parity review against the existing browser route family
4. switch Go serving from the old template-based `/app` to the built Svelte frontend
5. retire superseded Go-template browser code and tests

## 2. Scope

In scope:

1. exact detail routes across inbound requests, approvals, proposals, documents, accounting, inventory, work orders, and audit where they remain part of the promoted browser family
2. `/app/settings`
3. `/app/admin`
4. current bounded admin-maintenance surfaces that already exist on the shared backend
5. Go cutover work for serving the Svelte build under `/app`
6. deletion or retirement of old template-based browser code that becomes dead after cutover
7. end-to-end validation and workflow-checklist updates tied to the new browser surface

Out of scope:

1. new workflow breadth unrelated to current browser parity
2. reopening completed backend milestones under the label of migration
3. a long-lived production dual-stack browser model

## 3. Required outcomes

This slice is complete only when:

1. the promoted detail-route family is available in Svelte
2. current admin and settings continuity required by the product is available in Svelte
3. the Go binary serves the Svelte build at `/app`
4. the earlier Go-template browser implementation is no longer the active serving path
5. workflow validation on the new browser surface is recorded on the `docs/workflows/` track

## 4. Cutover rule

1. switch serving only after the bounded parity checklist for the promoted route family passes
2. once serving switches, prefer immediate dead-code retirement for the old browser path rather than preserving a silent fallback
3. if one small fallback seam must remain temporarily, document it explicitly and give it a bounded removal plan

## 5. Verification

Before closing this slice:

1. run frontend verification on all migrated detail and admin routes
2. run canonical Go verification for the cutover and any backend changes
3. run bounded end-to-end workflow validation on the real `/app` plus `/api/...` seam after cutover
4. update workflow checklists and route expectations under `docs/workflows/`

## 6. Current implementation checkpoint

The active Slice 3 product checkpoints are now landed in code.

Landed result:

1. `/app/settings` now has Svelte continuity parity for current session context, ownership guidance, workload-aware recommended routes, admin continuation for privileged actors, and personal utility shortcuts instead of remaining a placeholder shell checkpoint
2. the Svelte admin route family now covers `/app/admin`, `/app/admin/accounting`, `/app/admin/parties`, `/app/admin/parties/{party_id}`, `/app/admin/access`, and `/app/admin/inventory` against the existing shared `/api/admin/...` seams
3. the Svelte shell is now role-aware for privileged maintenance: admin destinations stay visible only to admin actors, while non-admin access attempts redirect back to `/app` with an explicit error message
4. admin maintenance parity now includes ledger-account, tax-code, accounting-period, party, contact, org-user, role-assignment, inventory-item, and inventory-location flows on the same shared backend ownership boundaries already used by the old browser layer
5. the promoted exact inbound-request detail route now runs in Svelte at `/app/inbound-requests/{request_reference_or_id}` with direct continuity for request-reference, run, step, and delegation lookups on the existing shared `/api/review/inbound-requests/{lookup}` seam, and the main Svelte request-entry surfaces now link to that exact detail route instead of stopping at filtered list views
6. the remaining promoted exact review-detail family now also runs in Svelte at `/app/review/approvals/{approval_id}`, `/app/review/proposals/{recommendation_id}`, `/app/review/documents/{document_id}`, `/app/review/accounting/{entry_id}`, `/app/review/inventory/{movement_id}`, `/app/review/work-orders/{work_order_id}`, and `/app/review/audit/{event_id}` on explicit shared `/api/review/.../{id}` detail seams rather than handler-local HTML composition
7. the migrated list routes now link into those exact Svelte detail routes instead of stopping at list-only continuity
8. `npm --prefix web run check`, `npm --prefix web run test`, `npm --prefix web run build`, `mcp__svelte__svelte_autofixer` on the new settings component and the earlier inventory-movement detail component, `go build ./cmd/... ./internal/...`, and the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` verification all completed cleanly for these checkpoints
9. focused Go integration coverage now also clears `WORKFLOW_WEB_FRONTEND` and asserts the promoted `/app` route family serves the embedded Svelte shell by default at `/app`, `/app/login`, the promoted utility and landing routes, and the promoted detail-route family, so the default-cutover serving path has explicit code-level coverage beyond the narrower unit checks in `internal/app/web_test.go`

Remaining Slice 3 work:

1. bounded post-cutover workflow validation on the real `/app` plus `/api/...` seam still remains open on the `docs/workflows/` track before the temporary `WORKFLOW_WEB_FRONTEND=templates` fallback can be retired
2. the black-box integration suite still carries legacy template-HTML continuity assertions behind `WORKFLOW_WEB_FRONTEND=templates`, so those assertions need to move onto the shared API seam or other Svelte-compatible coverage before old template-browser retirement and dead-code cleanup can finish cleanly
