# workflow_app Execution Plan V2

Date: 2026-04-10
Status: Active execution order after Milestone 13 closeout and the follow-on implementation review that promoted Milestone 14 as the active next bounded v2 milestone, the user-testing-readiness milestone, the first baseline accounting-reporting milestone on the promoted runtime, the first broader navigation and information-architecture cleanup milestone for crowded promoted areas, and the milestone that seeds the demo entity for realistic testing
Purpose: define the current execution order without carrying the full completed milestone narrative in the default context.

## 1. Completed baseline

1. thin-v1 foundation is complete
2. Milestone 10 browser rebuild history is complete and archived
3. Milestone 11 shell and navigation history is complete and archived
4. Milestone 12 admin-maintenance planning history is archived after implementation progress moved beyond it
5. Milestone 13 is implemented history, including Slice 3 cutover and browser-closeout work

## 2. Active execution order

1. treat Milestone 14 as the active next milestone
2. start with the bounded review-gap corrective pass on inbound-request lifecycle support and documentation truth
3. include the bounded shared-shell desktop layout correction in that first corrective pass because it is a real promoted-runtime UX issue rather than a later optional polish item
4. then apply the grouped-directory contextual-tab model in the most crowded promoted areas, starting with Admin and then Accounting, so tabs lead to focused directory pages and dedicated destinations
5. then add baseline accounting reports for trial balance, balance sheet, and income statement on the shared reporting seam and close clearly similar adjacent reporting gaps when they are exposed during the same work
6. then seed `North Harbor Works` with a standard chart of accounts and the minimum realistic master-data baseline needed for reports, admin/list surfaces, and bounded user testing
7. then expand production-readiness testing and verification where the current risk surface still exceeds the current coverage
8. then execute and record the deferred live workflow-validation backlog on the real served `/app` plus `/api/...` seam
9. then make the user-testing posture explicit so the application can be handed to bounded testers with clear supported workflows, exclusions, and guidance
10. update durable workflow-validation material in `docs/workflows/` whenever supported workflow truth or validation evidence changes
11. then update user guides, technical guides, and setup docs so the documentation set reflects the corrected current state

## 2.1 Delivered Milestone 13 baseline

The implemented Slice 3 baseline already includes:

1. Svelte continuity for `/app/settings`
2. Svelte admin continuity for `/app/admin`, `/app/admin/accounting`, `/app/admin/parties`, `/app/admin/parties/{party_id}`, `/app/admin/access`, and `/app/admin/inventory`
3. exact Svelte detail routes for inbound requests, approvals, proposals, documents, accounting entries, inventory movements, work orders, and audit events
4. direct detail-route continuity from the migrated list, landing, home, and coordinator-chat surfaces where exact IDs are already known
5. direct downstream accounting follow-through from exact request and proposal detail where the linked document already exists
6. Go serving of the embedded Svelte runtime at `/app`
7. retirement of the old template-based `/app` serving path and its compatibility branch

Milestone 13 Slice 3 closeout was intentionally narrow and is complete:

1. bounded real-seam desktop browser review on the current `/app` runtime, with Playwright used as the default execution path
2. workflow-checklist and evidence updates in `docs/workflows/`
3. one grouped corrective slice where the browser sweep exposed a real defect or missing support seam

Completed prerequisite recorded on 2026-04-09:

1. the OpenAI coordinator-provider loop now disables additional read tools after the bounded read budget is consumed instead of spinning on repeated reads
2. `cmd/verify-agent` now creates its verification actor through the shared browser-session auth path instead of only the lower-level session bootstrap
3. `cmd/verify-agent` now also supports `-approval-flow`, which creates one deterministic approval-ready proposal on the processed verification request and confirms request -> proposal -> approval -> document continuity through the shared session plus `/api/...` seam
4. `go build ./cmd/... ./internal/...`, `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`, `set -a; source .env; set +a; go run ./cmd/verify-agent`, `set -a; source .env; set +a; go run ./cmd/verify-agent -approval-flow`, and `set -a; source .env; set +a; go test -tags integration -count=1 ./internal/app -run TestOpenAIAgentProcessorLiveIntegration -v` all passed after that corrective slice

Completed prerequisite recorded on 2026-04-10:

1. focused Svelte component coverage now asserts multi-term route-catalog search continuity, promoted admin accounting and inventory status controls, and exact accounting-entry drill-down from request and proposal detail when the journal entry is already known
2. focused Go web-serving coverage now asserts SPA fallback across the full promoted `/app` route family and exact detail routes instead of relying on only a few shell-path checks
3. an additional focused Svelte route pass now asserts the promoted login, settings, admin hub, admin access, admin party setup, operations landing, submit-inbound-request, operations feed, and exact approval, document, and accounting detail surfaces
4. Playwright browser automation is now available locally, so the next closeout session should spend its browser-validation effort on the real served runtime instead of widening indirect coverage again unless the Playwright sweep exposes a concrete new gap
5. `npm --prefix web test -- page.test.ts page_detail.test.ts navigation.test.ts inventory/page.test.ts agent-chat/page.test.ts review/page.test.ts admin/accounting/page.test.ts admin/inventory/page.test.ts routes/page.test.ts`, `npm --prefix web test -- 'src/routes/(public)/login/page.test.ts' 'src/routes/(app)/settings/page.test.ts' 'src/routes/(app)/admin/page.test.ts' 'src/routes/(app)/admin/access/page.test.ts' 'src/routes/(app)/admin/parties/page.test.ts' 'src/routes/(app)/operations/page.test.ts' 'src/routes/(app)/submit-inbound-request/page.test.ts' 'src/routes/(app)/operations-feed/page.test.ts' 'src/routes/(app)/review/approvals/page_detail.test.ts' 'src/routes/(app)/review/documents/page_detail.test.ts' 'src/routes/(app)/review/accounting/page_detail.test.ts'`, focused `go test ./internal/app -run '^(TestRegisterWebRoutesServesSPAFallback|TestRegisterWebRoutesServesSPAFallbackAcrossPromotedRouteFamilies|TestHandleSvelteAppServesIndexAtAppRoot|TestHandleSvelteAppServesHeadRequests|TestHandleSvelteAppServesEmbeddedJSAsset|TestHandleSvelteAppDoesNotFallbackForMissingStaticAsset|TestNewAgentAPIHandlerWithDependenciesServesSvelteShell)$'`, `npm --prefix web run check`, and `git diff --check` all passed for that closeout-coverage slice
6. the final Playwright closeout also established four durable rules for future sessions:
   a. run the served app and any verification seed command against the same backend explicitly, especially when `cmd/verify-agent` would otherwise prefer `TEST_DATABASE_URL`
   b. rebuild `web/` into `internal/app/web_dist` and restart `cmd/app` before diagnosing stale-browser failures on the embedded runtime
   c. seed commands that create dedicated verification orgs must print the exact browser credentials and continuity ids needed for the browser proof
   d. browser route sweeps should assert stable page contracts and exact ids rather than overfitting to editable copy

The post-closeout implementation review then found the first Milestone 14 corrective scope:

1. the backend parked-request lifecycle seams remain implemented, but the promoted Svelte runtime does not yet expose the full draft-edit, queue, cancel, amend, and delete contract that older docs still describe
2. stale workflow and user-guide references still point at retired or nonexistent browser paths
3. the shared desktop shell still needs one bounded layout correction so the sidebar begins directly beneath the top app bar and the contextual tab row starts over the content column
4. the promoted accounting area is still underbuilt for a production-oriented operator surface until baseline financial statements exist on the shared reporting seam
5. the current contextual-tab and admin/accounting page model is still too crowded, so the promoted runtime should move toward grouped directory pages with dedicated destination pages behind them
6. the default demo entity is still too empty for realistic testing, so Milestone 14 should seed `North Harbor Works` with a standard chart of accounts and minimum master data through the shared backend-owned setup path
7. Milestone 14 should also absorb clearly similar adjacent readiness gaps that are discovered during implementation review when those additions stay coherent, bounded, and production-shape-driven
8. the next production-readiness push should prioritize workflow-critical correctness tests, failure-path validation, documentation truth, backend-owned reporting depth, operator-usable information architecture, and a realistic demo baseline rather than another broad browser-architecture milestone
9. the first 2026-04-10 Slice 1 checkpoint is now implemented in code and focused frontend verification: exact inbound-request detail exposes draft save plus queue plus delete and queued cancel plus amend controls through the shared backend seam, and the desktop shell layout now starts the contextual-tab row over the content column instead of above the sidebar
10. the follow-up Slice 1 documentation-truth checkpoint on 2026-04-10 realigned the workflow catalog, inbound-request lifecycle guide, agent-chat guide, inbound-request technical guide, and active scope note with the current route contract
11. the grouped navigation checkpoint on 2026-04-10 then added Admin `Master Data` and `Lists` directory pages, changed the Admin contextual tabs to grouped directories plus Access, and changed `/app/review/accounting` into an accounting report directory with dedicated `journal-entries`, `control-balances`, and `tax-summaries` destinations
12. the baseline accounting-report checkpoint on 2026-04-10 then added backend-owned trial balance, balance sheet, and income statement report contracts plus dedicated Svelte destinations under `/app/review/accounting`, with focused reporting integration tests, focused route-serving tests, focused Svelte component tests, frontend check/build, Go build, and gopls diagnostics passing for that checkpoint
13. the `North Harbor Works` demo-data baseline checkpoint on 2026-04-11 added a backend-owned idempotent setup seed behind `cmd/bootstrap-admin` for the minimum chart of accounts, GST tax codes, FY2026-27 accounting period, sample customer and vendor parties with primary contacts, starter inventory items, and starter inventory locations
14. the first production-readiness test-expansion checkpoint on 2026-04-11 added `internal/app` API integration coverage for the promoted inbound-request lifecycle mutation seam: browser-session draft create, draft update, queue, cross-org cancel rejection without state mutation, same-org cancel, amend-back-to-draft, and draft delete
15. the remaining Milestone 14 implementation work should now continue production-readiness test expansion where risk still exceeds coverage, then move to deferred workflow validation rather than reopening the request-detail, shell-layout, inbound-request route-vocabulary, first grouped-directory, baseline accounting-report, first demo-baseline, or first lifecycle API-coverage checkpoints

## 2.2 Milestone 15 future direction

If Milestone 14 closes cleanly, the next planned direction should be Milestone 15 data exchange:

1. CSV-first bulk master-data import on the shared backend seam
2. CSV and Excel-compatible export for promoted lists and reports
3. import and export flows that remain backend-owned, auditable, and operator-usable

## 3. Promotion rule for the next milestone

1. do not reopen completed milestone buckets broadly
2. if a real defect is found in completed work, handle it as one bounded corrective slice
3. choose the next milestone based on the strongest remaining production-shape need, not on historical sequence inertia

## 4. Verification rule

1. do not treat implementation as complete without running the required verification or recording an explicit blocker
2. use `../docs/technical_guides/07_testing_and_verification.md` for exact verification commands and workflow
3. for Milestone 13 closeout, verification must include frontend checks, canonical Go verification, and bounded end-to-end validation on the real served Svelte runtime
4. browser-serving changes must be checked for real asset serving, correct SPA fallback behavior, and one bounded browser smoke pass on `/app`
5. when Playwright is available locally, prefer it for the Milestone 13 real browser-review sweep rather than treating manual route clicking as the default first path
6. for future workflow-critical browser closeout, record the exact seed command, backend target, org slug, actor credentials, and continuity ids used for the pass in the same change that records the evidence
7. when documentation claims supported browser workflow behavior, verify that claim against the promoted Svelte runtime rather than against older archived implementation history alone
