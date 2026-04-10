# workflow_app Execution Plan V2

Date: 2026-04-10
Status: Active execution order after archive reset, with the 2026-04-09 coordinator-provider corrective slice, the approval-continuity verification pass, the 2026-04-10 focused Svelte closeout coverage passes, the real-browser Playwright closeout, and the verify-agent posted-accounting continuity seed now completed for Milestone 13
Purpose: define the current execution order without carrying the full completed milestone narrative in the default context.

## 1. Completed baseline

1. thin-v1 foundation is complete
2. Milestone 10 browser rebuild history is complete and archived
3. Milestone 11 shell and navigation history is complete and archived
4. Milestone 12 admin-maintenance planning history is archived after implementation progress moved beyond it
5. Milestone 13 Slice 1 and Slice 2 are implemented

## 2. Active execution order

1. treat Milestone 13 as closed with real-browser evidence and posted-accounting continuity proof
2. preserve the closeout lessons below as the default verification shape for the next workflow-critical browser slice
3. update durable workflow-validation material in `docs/workflows/` whenever that next slice changes supported operator truth
4. then promote one next bounded v2 milestone based on real remaining product or architecture need

## 2.1 Milestone 13 Slice 3 active checkpoint

The implemented Slice 3 baseline already includes:

1. Svelte continuity for `/app/settings`
2. Svelte admin continuity for `/app/admin`, `/app/admin/accounting`, `/app/admin/parties`, `/app/admin/parties/{party_id}`, `/app/admin/access`, and `/app/admin/inventory`
3. exact Svelte detail routes for inbound requests, approvals, proposals, documents, accounting entries, inventory movements, work orders, and audit events
4. direct detail-route continuity from the migrated list, landing, home, and coordinator-chat surfaces where exact IDs are already known
5. direct downstream accounting follow-through from exact request and proposal detail where the linked document already exists
6. Go serving of the embedded Svelte runtime at `/app`
7. retirement of the old template-based `/app` serving path and its compatibility branch

Milestone 13 Slice 3 closeout was intentionally narrow and is now complete:

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
