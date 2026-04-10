# workflow_app Tracker V2

Date: 2026-04-10
Status: Thin-v1 is complete, Milestones 10 through 12 are implemented history, and Milestone 13 including Slice 3 browser-review closeout is now complete after the bounded coordinator-provider corrective pass, the downstream-accounting exact-detail continuity pass, the shared-session approval-continuity verification pass, the real-browser Playwright sweep on the served `/app` runtime, the admin-accounting corrective slice for tax-control selection, and the verify-agent continuity-seeding correction that now emits a posted document plus exact journal-entry continuity on the same browser-reviewed org
Purpose: track the active implementation state, current sequencing, and immediate next steps without carrying full milestone history in the default context.

## 1. Current state

1. thin-v1 foundation is complete and should be treated as closed baseline rather than active open scope
2. the earlier Go-template browser rebuild and follow-on browser correction work are implemented history, not the forward planning surface
3. the promoted web direction is the Svelte-based web application on the shared Go backend
4. Milestone 12 admin-maintenance depth is implemented enough to stop being the active planning focus
5. Milestone 13 Slice 1 foundation and Slice 2 workflow-surface migration are implemented
6. Milestone 13 Slice 3 code checkpoints are largely landed: settings continuity, admin continuity, exact detail-route continuity, served-runtime cutover, and old-template retirement are implemented
7. the coordinator-provider seam needed one bounded corrective pass on 2026-04-09: the OpenAI coordinator now stops offering read tools after the bounded read budget is consumed, and `cmd/verify-agent` now creates its verification actor through the real browser-session auth path
8. request detail and processed-proposal detail now also prefer exact downstream accounting-entry drill-down when the linked document already has a posted journal entry, instead of stopping at filtered accounting-review continuity
9. focused frontend checks, focused Svelte route tests, `go build ./cmd/... ./internal/...`, focused non-DB Go tests in `internal/app`, and `gopls` diagnostics passed on 2026-04-09 for that exact-detail continuity pass
10. the broader DB-backed canonical suite `set -a; source .env; set +a; timeout 300s go test -p 1 ./cmd/... ./internal/...` passed cleanly on 2026-04-09, so the earlier `create tax code: unauthorized` and `reset test database: deadlock detected` failures should be treated as non-reproduced transient environment or test-state noise rather than active blockers
11. an additional real-seam validation pass on 2026-04-09 confirmed the served Svelte shell and asset behavior at `/app`, browser-session login through `/api/session/login`, route-catalog search for `pending approvals`, and one live request submission plus queue processing chain through exact request and proposal review continuity on the shared `/api/...` seam
12. `cmd/verify-agent` now also supports `-approval-flow`, and a 2026-04-09 live run used that shared-session API path to confirm one exact request -> proposal -> approval -> document continuity chain on the same verification request
13. focused Svelte route and component tests added on 2026-04-10 now assert multi-term route-catalog continuity, promoted admin accounting and inventory status controls, and exact accounting-entry drill-down from request and proposal detail before the final live browser sweep
14. focused Go web-serving coverage added on 2026-04-10 now asserts SPA-shell fallback across the full promoted `/app` route family and exact detail routes so cutover regressions do not hide behind only a few shell-path checks
15. additional focused Svelte route coverage added on 2026-04-10 now asserts the promoted login, settings, admin hub, admin access, admin party setup, operations landing, submit-inbound-request, operations feed, and exact approval, document, and accounting detail surfaces
16. Playwright browser automation is now available in the local environment and is the default tool for future browser-rendered `/app` verification work
17. the Milestone 13 desktop browser-review closeout passed on 2026-04-10 with four real-browser Playwright checks on the served `/app` seam covering desktop shell persistence, route-catalog continuity, admin maintenance and exact party detail, promoted route-family rendering, and exact request -> proposal -> approval -> document -> accounting continuity
18. the real-browser closeout also exposed and closed two verification-support gaps on 2026-04-10: admin accounting tax-code creation now uses bounded control-account selectors instead of raw account-id text fields, and `cmd/verify-agent -approval-flow` now emits verification-org credentials and seeds a real posted invoice plus journal entry when run against `DATABASE_URL`, which removed the earlier test-db and throwaway-org mismatch from browser continuity proof

## 2. Active implementation order

1. treat Milestone 13 closeout as complete and keep `docs/workflows/` as the evidence source for that browser-review finish
2. use the updated Playwright plus `cmd/verify-agent -database-url "$DATABASE_URL" -approval-flow` pattern as the default real-browser continuity proof for any future workflow-critical Svelte changes
3. decide the next bounded v2 milestone instead of reopening completed milestone buckets broadly

## 2.1 Active Slice 3 detail now kept in this tracker

The next session should assume this Milestone 13 baseline is already landed:

1. `/app/settings` and the bounded admin family already have promoted Svelte continuity on shared backend seams
2. exact Svelte detail routes already exist for inbound requests, approvals, proposals, documents, accounting, inventory, work orders, and audit
3. migrated list, home, review-landing, and coordinator-chat surfaces already deep-link to exact detail routes where known identifiers exist
4. the served Go runtime already embeds and serves the Svelte frontend at `/app`
5. the retired template-browser `/app` layer has already been removed from the active codebase
6. focused automated closeout coverage now exists for route discovery, admin status controls, exact accounting-entry continuity, promoted operator-entry plus utility plus detail-route browser-review component coverage, promoted-route SPA fallback coverage, and the real-seam desktop browser-review sweep
7. the promoted request and proposal detail surfaces already include exact downstream accounting-entry continuity when the shared reporting seam exposes a posted journal entry for the linked document
8. Playwright-driven browser review on the served `/app` runtime should remain the preferred closeout path over adding more indirect HTTP-only or component-only evidence when future workflow-critical UI slices land

Use these supporting docs for the remaining closeout:

1. `../docs/workflows/end_to_end_validation_checklist.md` for bounded real-seam validation steps
2. `../docs/workflows/workflow_validation_track.md` for explicit validation evidence and blocker tracking

## 3. Current decision gate

The next implementation session should answer this in code and verification, not in more planning expansion:

1. what bounded v2 milestone should follow Milestone 13 now that the served Svelte cutover and browser closeout are complete
2. what backend or workflow-support seams deserve the next production-strength push
3. what new workflow-validation evidence will be required for that next bounded milestone

## 4. Working rules

1. treat completed milestones as historical context unless a real defect forces a bounded corrective slice
2. keep new planning bounded to one coherent active concern at a time
3. prefer implementation and verification over planning expansion when the next step is already clear
4. keep the default active reading surface limited to the thin v2 docs in the root of `new_app_docs/`
5. if a detail is needed often during active work, keep it in the thin root docs rather than relying on repeated archive lookup
