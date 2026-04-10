# workflow_app Execution Plan V2

Date: 2026-04-09
Status: Active execution order after archive reset, with the 2026-04-09 coordinator-provider corrective slice, approval-continuity verification pass, the 2026-04-10 focused Svelte closeout coverage pass, and verification reruns completed ahead of the remaining live browser-review closeout
Purpose: define the current execution order without carrying the full completed milestone narrative in the default context.

## 1. Completed baseline

1. thin-v1 foundation is complete
2. Milestone 10 browser rebuild history is complete and archived
3. Milestone 11 shell and navigation history is complete and archived
4. Milestone 12 admin-maintenance planning history is archived after implementation progress moved beyond it
5. Milestone 13 Slice 1 and Slice 2 are implemented

## 2. Active execution order

1. complete the remaining Milestone 13 Slice 3 implementation work
2. verify the resulting Svelte cutover and shared backend continuity
3. update durable workflow-validation material in `docs/workflows/`
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

Remaining Slice 3 closeout is intentionally narrow:

1. bounded real-seam desktop browser review on the current `/app` runtime
2. workflow-checklist and evidence updates in `docs/workflows/`
3. one grouped corrective slice only if that browser sweep exposes a real defect or missing support seam

Completed prerequisite recorded on 2026-04-09:

1. the OpenAI coordinator-provider loop now disables additional read tools after the bounded read budget is consumed instead of spinning on repeated reads
2. `cmd/verify-agent` now creates its verification actor through the shared browser-session auth path instead of only the lower-level session bootstrap
3. `cmd/verify-agent` now also supports `-approval-flow`, which creates one deterministic approval-ready proposal on the processed verification request and confirms request -> proposal -> approval -> document continuity through the shared session plus `/api/...` seam
4. `go build ./cmd/... ./internal/...`, `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`, `set -a; source .env; set +a; go run ./cmd/verify-agent`, `set -a; source .env; set +a; go run ./cmd/verify-agent -approval-flow`, and `set -a; source .env; set +a; go test -tags integration -count=1 ./internal/app -run TestOpenAIAgentProcessorLiveIntegration -v` all passed after that corrective slice

Completed prerequisite recorded on 2026-04-10:

1. focused Svelte component coverage now asserts multi-term route-catalog search continuity, promoted admin accounting and inventory status controls, and exact accounting-entry drill-down from request and proposal detail when the journal entry is already known
2. focused Go web-serving coverage now asserts SPA fallback across the full promoted `/app` route family and exact detail routes instead of relying on only a few shell-path checks
3. `npm --prefix web test -- page.test.ts page_detail.test.ts navigation.test.ts inventory/page.test.ts agent-chat/page.test.ts review/page.test.ts admin/accounting/page.test.ts admin/inventory/page.test.ts routes/page.test.ts`, focused `go test ./internal/app -run '^(TestRegisterWebRoutesServesSPAFallback|TestRegisterWebRoutesServesSPAFallbackAcrossPromotedRouteFamilies|TestHandleSvelteAppServesIndexAtAppRoot|TestHandleSvelteAppServesHeadRequests|TestHandleSvelteAppServesEmbeddedJSAsset|TestHandleSvelteAppDoesNotFallbackForMissingStaticAsset|TestNewAgentAPIHandlerWithDependenciesServesSvelteShell)$'`, `npm --prefix web run check`, and `git diff --check` all passed for that closeout-coverage slice

## 3. Promotion rule for the next milestone

1. do not reopen completed milestone buckets broadly
2. if a real defect is found in completed work, handle it as one bounded corrective slice
3. choose the next milestone based on the strongest remaining production-shape need, not on historical sequence inertia

## 4. Verification rule

1. do not treat implementation as complete without running the required verification or recording an explicit blocker
2. use `../docs/technical_guides/07_testing_and_verification.md` for exact verification commands and workflow
3. for Milestone 13 closeout, verification must include frontend checks, canonical Go verification, and bounded end-to-end validation on the real served Svelte runtime
4. browser-serving changes must be checked for real asset serving, correct SPA fallback behavior, and one bounded browser smoke pass on `/app`
