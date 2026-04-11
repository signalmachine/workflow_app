# workflow_app Tracker V2

Date: 2026-04-10
Status: Thin-v1 is complete, Milestones 10 through 13 are implemented history, and Milestone 14 is now in fast-close mode after its core implementation checkpoints landed; remaining work should avoid broad new testing or workflow-validation expansion, run only a small risk-based verification gate needed for confidence, and prioritize the no-compromise documentation truth and user-testing-readiness closeout
Purpose: track the active implementation state, current sequencing, and immediate next steps without carrying full milestone history in the default context.

## 1. Current state

1. thin-v1 foundation is complete and should be treated as closed baseline rather than active open scope
2. the earlier Go-template browser rebuild and follow-on browser correction work are implemented history, not the forward planning surface
3. the promoted web direction is the Svelte-based web application on the shared Go backend
4. Milestone 12 admin-maintenance depth is implemented enough to stop being the active planning focus
5. Milestone 13 Slice 1 foundation, Slice 2 workflow-surface migration, and Slice 3 cutover or browser-closeout work are implemented history
6. the promoted Svelte runtime, the shared Go `/api/...` seam, the embedded `/app` serving path, the bounded admin family, and the exact detail-route family remain the current delivered baseline
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
19. the post-closeout codebase review then found one active product gap and two active documentation gaps that should now be treated as Milestone 14 work rather than as silent drift:
20. the backend still supports parked-request draft update, queue, cancel, amend, and delete actions, but the promoted Svelte runtime currently exposes only new-request creation plus read-only request detail continuity
21. the inbound-request user guide still points at a nonexistent browser path for queue processing instead of the current operations-page action
22. route and workflow docs still mix `/app/inbound-requests` and `/app/review/inbound-requests` in ways that no longer reflect the current promoted route family cleanly
23. the shared desktop shell still needs one bounded layout correction so the sidebar starts directly under the top app bar and the contextual tabs begin over the main content column instead of across the far-left edge
24. the promoted accounting area still lacks baseline operator financial statements such as trial balance, balance sheet, and income statement
25. Milestone 14 should also review the application for nearby production-shape gaps and add clearly similar bounded capability where those additions materially improve readiness, correctness, or operator continuity
26. the current contextual-tab and admin/accounting information architecture is still too noisy on several promoted pages, so Milestone 14 should move those areas toward grouped directory pages such as `Master Data`, `Lists`, and `Reports` with dedicated destination pages behind them
27. the demo entity `North Harbor Works` still needs a standard chart of accounts and other essential master data so user testing and report review do not begin from an unrealistically empty org
28. the first strong future milestone candidate remains structured data exchange: CSV-first bulk master-data import plus CSV/Excel-compatible export for promoted lists and reports, but Milestone 15 should not start immediately after Milestone 14 because the application should first go through extensive user testing on the Milestone 14 runtime
29. the first Milestone 14 Slice 1 checkpoint is now landed on 2026-04-10: exact inbound-request detail exposes draft save plus queue plus delete controls and queued cancel plus amend controls through the existing shared `/api/inbound-requests/{request_id}/{action}` seam, the desktop shell now places contextual tabs over the main content column instead of across the far-left edge, and focused Svelte checks for those new browser contracts now pass
30. a follow-up Slice 1 documentation-truth checkpoint on 2026-04-10 realigned the durable workflow catalog, inbound-request lifecycle guide, agent-chat guide, inbound-request technical guide, and active scope note with the current route contract: `/app/review/inbound-requests` is the list surface, `/app/inbound-requests/{request_reference_or_id}` is the exact detail and lifecycle-action surface, `/app/operations` owns the browser process-next action, and `/api/...` remains the mutation seam
31. the grouped-directory navigation checkpoint is now landed on 2026-04-10: the Admin contextual tabs now route through grouped `Master Data` and `Lists` directory pages before concrete accounting, party, inventory, or access destinations, and `/app/review/accounting` now acts as an accounting report directory with dedicated `journal-entries`, `control-balances`, and `tax-summaries` destinations behind it
32. focused Svelte component tests, `npm --prefix web run check`, `npm --prefix web run build`, focused Go served-route tests, and gopls diagnostics passed for that grouped navigation checkpoint on 2026-04-10
33. the baseline accounting-report checkpoint is now landed on 2026-04-10: the shared reporting seam exposes backend-owned trial balance, balance sheet, and income statement contracts with explicit sign conventions, current-earnings balance-sheet treatment, effective-date filtering, and imbalance totals; `/app/review/accounting` now links to dedicated `trial-balance`, `balance-sheet`, and `income-statement` destinations; focused reporting integration tests, focused app serving tests, focused Svelte component tests, `npm --prefix web run check`, `npm --prefix web run build`, `go build ./cmd/... ./internal/...`, `git diff --check`, and gopls diagnostics passed for that checkpoint
34. the `North Harbor Works` demo-baseline checkpoint is now landed on 2026-04-11: `cmd/bootstrap-admin` now seeds an idempotent minimum chart of accounts, GST tax codes, FY2026-27 accounting period, customer and vendor parties with primary contacts, starter inventory items, and starter inventory locations through a backend-owned setup package; focused setup integration tests, focused bootstrap/setup package tests, and gopls diagnostics passed for that checkpoint
35. the first production-readiness test-expansion checkpoint is now landed on 2026-04-11: `internal/app` API integration coverage now exercises the promoted inbound-request lifecycle mutation seam end-to-end through browser-session cookies for draft create, draft update, queue, cross-org cancel rejection without state mutation, same-org cancel, amend-back-to-draft, and draft delete; focused `internal/app` verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, and gopls diagnostics passed for that checkpoint
36. an additional production-readiness test-expansion checkpoint is now landed on 2026-04-11: `internal/app` approval-decision API integration coverage now proves cross-org browser-session approval decisions return not found without mutating the pending approval, decision metadata, or submitted document state; the focused approval-decision test group, full `internal/app` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, `git diff --check`, and gopls diagnostics passed for that checkpoint
37. another production-readiness test-expansion checkpoint is now landed on 2026-04-11: `internal/app` API integration coverage now proves the browser-session process-next endpoint exposes failed-provider continuity by returning the exact failed request reference and run id, then showing the failed request, failed run, and failed provider-execution step through exact inbound-request review detail; the focused failure-continuity test, full `internal/app` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, and gopls diagnostics passed for that checkpoint
38. another production-readiness test-expansion checkpoint is now landed on 2026-04-11: `internal/app` API integration coverage now proves cross-org browser-session proposal approval requests return not found without linking the recommendation, creating an approval row, or mutating the submitted document state; the focused cross-org proposal-approval boundary test, full `internal/app` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, `git diff --check`, and gopls diagnostics passed for that checkpoint
39. another production-readiness test-expansion checkpoint is now landed on 2026-04-11: `internal/app` API integration coverage now proves the promoted browser-session accounting-report endpoints keep trial balance, balance sheet, and income statement data scoped to the authenticated org; the focused accounting-report org-boundary test, full `internal/app` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, and gopls diagnostics passed for that checkpoint
40. another production-readiness test-expansion checkpoint is now landed on 2026-04-11: `internal/app` API integration coverage now proves promoted browser-session inventory and work-order review endpoints preserve same-org stock, movement detail, reconciliation, and work-order continuity while hiding the same exact records from another org through empty list results or not-found detail responses; the focused inventory/work-order org-boundary test, full `internal/app` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, `git diff --check`, and gopls diagnostics passed for that checkpoint
41. another production-readiness test-expansion checkpoint is now landed on 2026-04-11: `internal/app` API integration coverage now proves promoted browser-session Admin exact-record actions reject cross-org access without mutation across ledger-account status, tax-code status, accounting-period close, party detail/status/contact creation, inventory item/location status, and access membership role changes; the checkpoint also normalized accounting and inventory setup-status not-found errors so foreign exact ids return not-found rather than generic server failures; the focused admin-boundary test, full `internal/app`, affected `internal/accounting` and `internal/inventoryops` package verification, `go build ./cmd/... ./internal/...`, serialized canonical Go verification, `git diff --check`, and gopls diagnostics passed for that checkpoint

## 2. Active implementation order

1. treat Milestone 13 implementation as delivered baseline and Milestone 14 as the active next bounded milestone
2. treat Milestone 14 Slice 1 as closed through the request-detail lifecycle, shell-layout, inbound-request documentation-truth, and first grouped-directory navigation checkpoints
3. treat the first baseline accounting-report checkpoint as landed through trial balance, balance sheet, and income statement on the shared reporting seam
4. treat the first `North Harbor Works` demo-baseline checkpoint as landed through the bootstrap-owned chart-of-accounts, tax-code, accounting-period, party/contact, and inventory master-data seed
5. treat the current production-readiness test-expansion checkpoints as sufficient for Milestone 14 closeout unless a concrete blocker or high-risk defect is discovered; do not add more broad test-expansion slices by default
6. run only a bounded final verification gate for closeout: the canonical Go verification, any focused frontend check needed for edited docs-adjacent runtime files, and at most one small live-smoke workflow validation pass on the promoted runtime if the documentation claims need fresh evidence
7. then make user-testing readiness explicit by documenting the supported testing posture, remaining exclusions, workflow guidance, and known validation limits; this documentation closeout is required and should not be reduced
8. defer deeper workflow-validation backlog execution and additional exploratory testing to the user-testing period unless a current blocker would make bounded testing misleading
9. use the updated Playwright plus `cmd/verify-agent -database-url "$DATABASE_URL" -approval-flow` pattern as the default real-browser continuity proof for future workflow-critical Svelte changes

## 2.1 Current delivered baseline

The next session should assume this pre-Milestone-14 baseline is already landed:

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

## 2.2 Future milestone candidate after user testing

After Milestone 14, the immediate next operating step is extensive user testing on the corrected promoted runtime. Milestone 15 remains the next planned implementation milestone candidate after that user-testing period, not the immediate follow-on work:

1. start with CSV-first bulk master-data import rather than native Excel workbook import
2. use CSV upload for bounded bulk creation of records such as chart-of-accounts rows, parties, inventory items, and locations
3. support export of promoted lists and reports in CSV
4. support Excel-compatible export for key reports and lists where justified
5. keep import validation, persistence, and auditability on the shared backend seam

## 3. Current decision gate

The next implementation session should answer this in code and verification, not in more planning expansion:

1. does the promoted Svelte runtime regain full parked-request lifecycle support, or is the browser contract intentionally narrowed and documented as such
2. does the shared desktop shell land the bounded sidebar-versus-contextual-tab layout correction cleanly on the promoted runtime
3. do baseline accounting reports land in a production-shape form on the shared reporting seam, and what adjacent similar report gaps should be closed in the same milestone
4. does `North Harbor Works` gain the minimum realistic chart of accounts and master-data baseline needed for bounded user testing and meaningful report review
5. what small final verification gate is still required to avoid handing testers an obviously broken runtime
6. what documentation updates are required to close Milestone 14 cleanly, including explicit notes for any workflow-validation evidence deferred to user testing

## 4. Working rules

1. treat completed milestones as historical context unless a real defect forces a bounded corrective slice
2. keep new planning bounded to one coherent active concern at a time
3. prefer implementation and verification over planning expansion when the next step is already clear
4. keep the default active reading surface limited to the v2 docs in the root of `new_app_docs/`
5. if a detail is needed often during active work, keep it in the thin root docs rather than relying on repeated archive lookup
