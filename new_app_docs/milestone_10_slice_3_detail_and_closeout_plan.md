# workflow_app Milestone 10 Slice 3 Plan

Date: 2026-04-01
Status: Accepted next implementation slice after the rebuilt operator-entry and review-workbench families
Purpose: define the third large Milestone 10 implementation slice so the detail surfaces, parity closeout, and retirement of the legacy web structure are finished together instead of being left as an indefinite cleanup tail.

## 1. Slice role

This slice completes the rebuild.

It should:

1. rebuild the promoted detail surfaces on the new modular architecture
2. close browser parity on the promoted thin-v1 route family
3. remove the legacy monolithic template structure and duplicate styling

## 2. Why this slice exists

The review workbench family can become coherent before the detail surfaces do, but Milestone 10 is not complete until exact drill-down and continuity pages also match the new browser model.

This slice exists to:

1. finish the browse-to-detail operator path in one coordinated pass
2. prevent a long-lived mixed state where list pages are rebuilt but detail pages still carry the legacy structure
3. force cleanup and legacy removal to happen as part of the same promoted slice, not as a vague future tidy-up

## 3. In scope

In scope:

1. `/app/inbound-requests/{request_reference_or_id}`
2. `/app/review/approvals/{approval_id}`
3. `/app/review/proposals/{recommendation_id}`
4. `/app/review/documents/{document_id}`
5. `/app/review/accounting/{entry_id}`
6. `/app/review/accounting/control-accounts/{account_id}`
7. `/app/review/accounting/tax-summaries/{tax_code}`
8. `/app/review/inventory/{movement_id}`
9. `/app/review/inventory/items/{item_id}`
10. `/app/review/inventory/locations/{location_id}`
11. `/app/review/work-orders/{work_order_id}`
12. `/app/review/audit/{event_id}`
13. shared detail-page primitives for primary summaries, lifecycle or action panels, related-links blocks, secondary metadata, payload or trace treatment, and expandable deep-context sections
14. legacy template and style removal required to make the rebuilt architecture the only active baseline

Out of scope:

1. new browser-product breadth unrelated to the existing promoted route family
2. backend workflow redesign
3. indefinite post-rebuild polish not tied to parity or continuity closeout

## 4. Required design outcomes

This slice is complete only when:

1. every promoted detail route has a clear primary-versus-secondary hierarchy
2. request detail remains a strong continuity hub without reading as one long undifferentiated stack
3. deep payloads, trace material, and verbose metadata are de-emphasized or collapsible where appropriate
4. detail pages still preserve exact workflow continuity into audit, approval, document, accounting, inventory, and execution context where those links already exist
5. the final rebuilt browser surface feels like one coherent application rather than a shell refresh wrapped around old detail pages

## 5. Required implementation outcomes

Required implementation outcomes:

1. all promoted detail routes render from the new modular template structure
2. shared detail primitives are reused across the family rather than recreated locally
3. the legacy monolithic template path is retired from active use by the end of the slice
4. duplicate styling or obsolete partials made unnecessary by the rebuild are removed
5. final doc and workflow references reflect the rebuilt browser baseline

## 6. Suggested implementation order inside the slice

Implement in this order:

1. finalize shared detail-page primitives
2. migrate inbound-request, approval, and proposal detail as the upstream continuity cluster
3. migrate document, accounting, inventory, work-order, and audit detail as the downstream continuity cluster
4. remove legacy template code and duplicate styling
5. complete final parity and browser-validation closeout

Implementation naming rule:

1. keep Slice 3 labels in planning docs, tracker rows, and review notes only
2. do not create production code filenames or symbols that embed phase labels such as `slice3`, `milestone10`, or similar sequencing markers
3. name code by the detail or closeout responsibility it owns, such as request detail, approval detail, review detail templates, legacy template cleanup, or shared detail primitives

## 7. Verification

Before closing this slice:

1. update focused web tests for the rebuilt detail family and any legacy-removal-sensitive seams
2. run `go build ./cmd/... ./internal/...`
3. run `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
4. run `gopls` diagnostics on edited Go files
5. run bounded browser review on desktop and narrow-width layouts for all promoted detail routes
6. run a focused workflow continuity pass across the rebuilt `/app` route family
7. record browser-review and workflow-continuity evidence on the `docs/workflows/` track, or explicitly record the blocker there if closeout evidence cannot yet be completed

## 8. Stop rule

Stop this slice only when:

1. the promoted detail family is rebuilt
2. the new modular web architecture is the only active baseline
3. legacy monolithic template structure needed for the old browser path is retired
4. documentation and workflow-validation references are synced to the rebuilt surface
5. the in-scope detail routes above all render on the new architecture

Milestone 10 should close immediately after this slice rather than staying open for vague visual follow-up.

## 9. Documentation sync

When this slice lands:

1. update `new_app_tracker.md` with implementation status, verification, and closeout state
2. update `milestone_10_web_rebuild_plan.md` to record completion status
3. update `README.md` if architecture shape or browser usage guidance materially changes
4. update relevant `docs/workflows/` files if browser routes, workflow navigation, or validation checklists drift
