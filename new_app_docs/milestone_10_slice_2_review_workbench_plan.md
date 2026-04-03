# workflow_app Milestone 10 Slice 2 Plan

Date: 2026-04-03
Status: Implemented historical slice for the earlier Go-template browser rebuild; superseded as forward stack guidance by `../docs/svelte_web_guides/svelte_web_ui_migration_plan.md`
Purpose: record the second large Milestone 10 implementation slice from the earlier Go-template browser rebuild so the review-workbench product scope remains available as migration context without serving as the active stack plan.

## 1. Slice role

This slice rebuilds the review workbench family on top of the Slice 1 architecture and shell.

It should land one coherent page model across the promoted review surfaces.

## 2. Why this slice exists

The current risk is not that one review page looks slightly worse than another.

The real risk is that operators move across many list and summary surfaces that currently carry too much local variation in filters, headers, summary blocks, table treatment, and row actions.

This slice exists to:

1. standardize the review browsing model in one coordinated pass
2. prevent a prolonged mixed state where some review pages use the new structure and others still use the old one
3. make the browse-to-drill-down flow coherent before the detail rebuild begins

## 3. In scope

In scope:

1. `/app/review/inbound-requests`
2. `/app/review/approvals`
3. `/app/review/proposals`
4. `/app/review/documents`
5. `/app/review/accounting`, including summary-table pivots that already exist on the current seam
6. `/app/review/inventory`
7. `/app/review/work-orders`
8. `/app/review/audit`
9. shared review-page primitives for filters, summary blocks, table containment, row actions, pagination or empty states, and page-level cross-links

Out of scope:

1. exact detail-page rebuilds for `/app/inbound-requests/{request_reference_or_id}`, `/app/review/approvals/{approval_id}`, `/app/review/proposals/{recommendation_id}`, `/app/review/documents/{document_id}`, `/app/review/accounting/{entry_id}`, `/app/review/accounting/control-accounts/{account_id}`, `/app/review/accounting/tax-summaries/{tax_code}`, `/app/review/inventory/{movement_id}`, `/app/review/inventory/items/{item_id}`, `/app/review/inventory/locations/{location_id}`, `/app/review/work-orders/{work_order_id}`, or `/app/review/audit/{event_id}`
2. new review data models or new broad backend capabilities
3. broad workflow-policy changes under the label of UI cleanup

## 4. Required design outcomes

This slice is complete only when:

1. all promoted review list surfaces share one clear hierarchy for page header, filters, summary, table, and row actions
2. filters are visually subordinate to the review data they control
3. summary cards or summary tables appear only where they materially improve operator scanning or decision speed
4. review rows still provide strong continuity links without turning every table into a wall of equally weighted actions
5. narrow-width table behavior is handled consistently through shared containment rather than local patches

## 5. Required implementation outcomes

Required implementation outcomes:

1. all review pages listed above render from the new modular template structure
2. shared review primitives are reused across the family rather than copied page by page
3. legacy review-template duplication is reduced materially by the end of the slice
4. the browser route map remains stable unless a change is explicitly documented before implementation

## 6. Suggested implementation order inside the slice

Implement in this order:

1. finalize shared review-page primitives
2. migrate inbound-request, approvals, and proposals review together
3. migrate documents, accounting, inventory, work orders, and audit as the downstream review family
4. remove or reduce legacy review-template duplication made obsolete by the migration

Implementation naming rule:

1. keep Slice 2 labels in planning docs, tracker rows, and review notes only
2. do not create production code filenames or symbols that embed phase labels such as `slice2`, `milestone10`, or similar sequencing markers
3. name code by the review workbench responsibility it owns, such as review templates, review filters, review summaries, review tables, or route families

## 7. Verification

Before closing this slice:

1. update focused web tests for the shared review-page model and representative routes across the family
2. run `go build ./cmd/... ./internal/...`
3. run `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
4. run `gopls` diagnostics on edited Go files
5. run bounded browser review on desktop and narrow-width layouts for all rebuilt review routes
6. record bounded review-route browser-review evidence on the `docs/workflows/` track if this slice materially changes operator navigation or validation expectations

## 8. Stop rule

Stop this slice when:

1. the promoted review list family is rebuilt on the new shared review-page model
2. review-page continuity works cleanly into the still-existing detail surfaces
3. the slice does not widen into a broad detail-page redesign beyond any narrowly required compatibility seam
4. none of the in-scope review routes above still depends on the legacy review-page layout

If a page needs substantive detail-hierarchy work, that belongs in Slice 3.

## 9. Documentation sync

When this slice lands:

1. update `new_app_tracker.md` with implementation status and verification state
2. update `milestone_10_web_rebuild_plan.md` if actual review-family scope or stop rules drift
3. update relevant workflow validation docs if review navigation or browser-validation expectations materially change
