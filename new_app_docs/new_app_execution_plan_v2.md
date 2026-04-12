# workflow_app Execution Plan V2

Date: 2026-04-12
Status: Milestone 14 is closed after the core implementation checkpoints, documentation-truth closeout, and final confidence gate; April 2026 review validation now promotes Milestone 15 urgent review and AI hardening before broader Milestone 16 structural and AI capability work
Purpose: define the current execution order without carrying the full completed milestone narrative in the default context.

## 1. Completed baseline

1. thin-v1 foundation is complete
2. Milestone 10 browser rebuild history is complete and archived
3. Milestone 11 shell and navigation history is complete and archived
4. Milestone 12 admin-maintenance planning history is archived after implementation progress moved beyond it
5. Milestone 13 is implemented history, including Slice 3 cutover and browser-closeout work

## 2. Active execution order

1. treat Milestone 14 as closed baseline rather than active implementation scope
2. implement `milestone_15_urgent_review_and_ai_hardening_plan.md` first: urgent API defects, attachment bounds, inventory landing continuity, current-runtime AI hardening, and specialist-truth correction
3. implement `milestone_16_structural_and_ai_capability_plan.md` next as a full AI-layer and foundational workflow-execution overhaul: start from the planned non-urgent refactors, AI architecture improvements, accounting proposal generation from inbound requests, specialist execution, recovery loops, and capability expansion, but make the first externally meaningful checkpoint the request-to-accounting-entry workflow rather than waiting for every structural cleanup slice to land
4. continue extensive user testing on the corrected promoted runtime, but do not let user testing defer the urgent Milestone 15 defects once they are validated
5. record pass, fail, blocker, and deferral evidence in `docs/workflows/` before promoting workflow-facing follow-up implementation work
6. group real product defects found during testing into bounded corrective slices rather than reopening broad milestone buckets
7. update user guides, technical guides, setup docs, and active planning docs when user testing or review implementation changes supported workflow truth
8. keep data exchange as future implementation until Milestone 15 closes and user-testing findings have been triaged

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
2. stale workflow and user-guide references pointed at retired or nonexistent browser paths before the Milestone 14 documentation-truth correction
3. the shared desktop shell needed one bounded layout correction before the shell-layout checkpoint
4. the promoted accounting area was underbuilt for a production-oriented operator surface before baseline financial statements landed on the shared reporting seam
5. the contextual-tab and admin/accounting page model was too crowded before the grouped-directory checkpoints
6. the default demo entity was too empty for realistic testing before Milestone 14 seeded `North Harbor Works` with a standard chart of accounts and minimum master data through the shared backend-owned setup path
7. Milestone 14 should also absorb clearly similar adjacent readiness gaps that are discovered during implementation review when those additions stay coherent, bounded, and production-shape-driven
8. the next production-readiness push should prioritize workflow-critical correctness tests, failure-path validation, documentation truth, backend-owned reporting depth, operator-usable information architecture, and a realistic demo baseline rather than another broad browser-architecture milestone
9. the first 2026-04-10 Slice 1 checkpoint is now implemented in code and focused frontend verification: exact inbound-request detail exposes draft save plus queue plus delete and queued cancel plus amend controls through the shared backend seam, and the desktop shell layout now starts the contextual-tab row over the content column instead of above the sidebar
10. the follow-up Slice 1 documentation-truth checkpoint on 2026-04-10 realigned the workflow catalog, inbound-request lifecycle guide, agent-chat guide, inbound-request technical guide, and active scope note with the current route contract
11. the grouped navigation checkpoint on 2026-04-10 then added Admin `Master Data` and `Lists` directory pages, changed the Admin contextual tabs to grouped directories plus Access, and changed `/app/review/accounting` into an accounting report directory with dedicated `journal-entries`, `control-balances`, and `tax-summaries` destinations
12. the baseline accounting-report checkpoint on 2026-04-10 then added backend-owned trial balance, balance sheet, and income statement report contracts plus dedicated Svelte destinations under `/app/review/accounting`, with focused reporting integration tests, focused route-serving tests, focused Svelte component tests, frontend check/build, Go build, and gopls diagnostics passing for that checkpoint
13. the `North Harbor Works` demo-data baseline checkpoint on 2026-04-11 added a backend-owned idempotent setup seed behind `cmd/bootstrap-admin` for the minimum chart of accounts, GST tax codes, FY2026-27 accounting period, sample customer and vendor parties with primary contacts, starter inventory items, and starter inventory locations
14. the first production-readiness test-expansion checkpoint on 2026-04-11 added `internal/app` API integration coverage for the promoted inbound-request lifecycle mutation seam: browser-session draft create, draft update, queue, cross-org cancel rejection without state mutation, same-org cancel, amend-back-to-draft, and draft delete
15. the next production-readiness test-expansion checkpoint on 2026-04-11 added `internal/app` API integration coverage for cross-org approval-decision rejection without mutating the pending approval, approval decision metadata, or submitted document state
16. the next production-readiness test-expansion checkpoint on 2026-04-11 added `internal/app` API integration coverage for failed provider processing through the browser-session process-next endpoint and exact inbound-request review detail
17. the next production-readiness test-expansion checkpoint on 2026-04-11 added `internal/app` API integration coverage for cross-org proposal approval requests without mutating the recommendation approval link, creating an approval row, or changing submitted document state
18. the next production-readiness test-expansion checkpoint on 2026-04-11 added `internal/app` API integration coverage proving the promoted browser-session accounting-report endpoints keep trial balance, balance sheet, and income statement data scoped to the authenticated org
19. the next production-readiness test-expansion checkpoint on 2026-04-11 added `internal/app` API integration coverage proving the promoted browser-session inventory and work-order review endpoints preserve same-org stock, movement detail, reconciliation, and work-order continuity while hiding exact foreign-org records through empty list results or not-found detail responses
20. the next production-readiness test-expansion checkpoint on 2026-04-11 added `internal/app` API integration coverage proving promoted browser-session Admin exact-record actions reject foreign-org ledger-account status, tax-code status, accounting-period close, party detail/status/contact creation, inventory item/location status, and access membership-role changes without mutating those records, and corrected accounting and inventory status-update not-found translation for foreign exact ids
21. the Milestone 14 documentation-truth closeout and final confidence gate are now complete; avoid additional broad pre-user-testing production-readiness expansion unless user testing or a concrete blocker promotes one bounded corrective slice

## 2.2 Post-Milestone 14 user testing and future data exchange

Milestone 14 has closed cleanly. Do not start data-exchange implementation immediately. The April 2026 review findings now create a nearer-term Milestone 15 urgent corrective step before future data exchange or larger feature work.

Milestone 15 should close the urgent review findings first. Milestone 16 should then handle the remaining non-urgent structural and AI capability recommendations. User testing on the promoted runtime should continue, with findings triaged into bounded corrective work as needed.

Data exchange remains the future implementation candidate after that user-testing period. Its planned direction is:

1. CSV-first bulk master-data import on the shared backend seam
2. CSV and Excel-compatible export for promoted lists and reports
3. import and export flows that remain backend-owned, auditable, and operator-usable

## 3. Promotion rule for the next milestone

1. do not reopen completed milestone buckets broadly
2. if a real defect is found in completed work, handle it as one bounded corrective slice
3. choose the next milestone based on findings from the user-testing period, not on historical sequence inertia or additional pre-testing validation expansion

## 4. Verification rule

1. do not treat implementation as complete without running the required verification or recording an explicit blocker
2. use `../docs/technical_guides/07_testing_and_verification.md` for exact verification commands and workflow
3. for post-Milestone-14 corrective slices, verification should be limited to the affected runtime surface, canonical Go verification when code or persistence changes warrant it, and a single small live-smoke workflow validation pass only when needed to support current documentation claims
4. browser-serving changes must be checked for real asset serving, correct SPA fallback behavior, and one bounded browser smoke pass on `/app`
5. when Playwright is available locally, prefer it for the Milestone 13 real browser-review sweep rather than treating manual route clicking as the default first path
6. for future workflow-critical browser closeout, record the exact seed command, backend target, org slug, actor credentials, and continuity ids used for the pass in the same change that records the evidence
7. when documentation claims supported browser workflow behavior, verify that claim against the promoted Svelte runtime rather than against older archived implementation history alone
