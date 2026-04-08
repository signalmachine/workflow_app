# workflow_app

`workflow_app` is an AI-agent-first, database-first business operating system centered on documents, ledgers, execution context, approvals, reports, and the first real browser operator layer.

`new_app_docs/` remains the canonical implementation-planning source for scope, sequencing, status, and next implementation steps. `docs/workflows/` is now the separate workflow-reference and workflow-validation track for supported operator paths, reusable validation checklists, live review evidence, and later user-guide preparation. Do not treat `new_app_docs/` as the active live workflow-testing tracker.

The post-checkpoint validation work remains important, but it is now tracked separately from implementation planning. The Milestone 10 web rebuild is implemented in code across the modular operator-entry, review-workbench, and detail-route families, while browser-review and workflow-continuity evidence still remain open on the separate `docs/workflows/` track. Active implementation has now also moved through the full Milestone 11 slice set, the first six Milestone 12 maintenance slices, and the active Milestone 13 cutover checkpoint: `web/` now contains the landed SvelteKit scaffold with a major-area sidebar plus contextual section tabs beneath the top bar, protected app-route and public login-route structure, shared primitives, minimal session bootstrap against `/api/session`, unit-test coverage for the frontend API client plus the navigation model, direct build output into `internal/app/web_dist`, the migrated Svelte workflow surfaces for home, inbound request intake, operations landing and feed, coordinator chat, route discovery, review landing, inventory landing, and the promoted review-list families, the migrated Svelte admin route family for `/app/admin`, `/app/admin/accounting`, `/app/admin/parties`, `/app/admin/parties/<party-id>`, `/app/admin/access`, and `/app/admin/inventory`, and the promoted exact Svelte detail routes for inbound requests, approvals, proposals, documents, accounting journal entries, inventory movements, work orders, and audit events on shared `/api/review/...` detail seams. `/app/inventory` now serves as a real inventory area landing backed by the shared review APIs for stock positions, recent movement continuity, and pending execution or accounting handoffs instead of stopping at a scaffold checkpoint. `cmd/app` now serves that built SPA under `/app` on the promoted runtime path, the generic `internal/app` handler constructors now default to that same served Svelte runtime shape, and the retired Go-template `/app` layer has now been removed from the active codebase instead of lingering as a compatibility path. The current shell now groups navigation by durable application areas: Agent, Accounting, Inventory, Operations, Settings, and access-scoped Admin, while contextual tabs keep second-level navigation local to the active area instead of flattening the app back into one long route list. `/app/routes` still provides searchable destination-only route discovery, `/app/settings` still shows live session context plus workflow-continuity shortcuts, access-scoped `/app/admin` remains the privileged maintenance hub, and `/app` still behaves as a role-aware operator home rather than a permanently generic dashboard. Slice 2 also adds shared `/api/navigation/...` snapshot endpoints so route-catalog, home, operations, review landing, and agent-chat composition stay server-shaped instead of being rebuilt in client-only browser code. The `internal/app` test-suite performance and reliability pass is now also landed: the shared DB-backed harness no longer reruns schema migrations on every `dbtest.Open` call inside one test process, and the disposable advisory lock is now held only during migration or reset setup instead of for the full lifetime of each DB-backed test, while per-test resets keep the suite isolated. The live-provider checkpoint itself remains healthy: Step 1 live-provider verification was restored on 2026-03-28, the Milestone 9 closeout on 2026-03-29 reconfirmed the live path, and the first post-review rerun on 2026-03-29 also passed. The OpenAI Responses loop uses provider-safe function-tool names plus stateless continuation compatible with `store: false`, focused `go test ./internal/app -run '^TestHandleWeb' -count=1` plus `go build ./cmd/... ./internal/...` passed for the earlier shell slices, and the new inventory-landing checkpoint now also passes through frontend type-check, frontend tests, frontend build, Go build, and the canonical DB-backed Go suite.

This repository has completed thin-v1 through Milestone 9 from the canonical planning set in [`new_app_docs/`](./new_app_docs). Starting at Milestone 10, active implementation work is v2 work: broader application enhancement plus production-readiness work on top of the completed shared foundation. That promotion does not change the core doctrine. The application remains AI-agent-first, database-first, workflow-centered, approval-gated, and built on one shared backend truth for browser and later clients. What changes in v2 is execution posture: the repository is no longer limited by thin-v1 breadth, active work can now improve usability, expand capability deliberately, harden the application toward production readiness across backend, AI, browser, validation, and operational layers, and proactively refactor or rebuild weak areas when that is the stronger best-practice engineering path.

`Go wild` in this repository means:

1. apply strong established best practices freely when they improve the product or codebase materially
2. do not preserve weak implementation merely because it already exists
3. continuously review the codebase for refactor or rebuild opportunities and either implement them or plan them explicitly
4. avoid experimental, fashion-driven, or novelty-first choices that are not justified by sound business-software or engineering practice
5. treat large monolithic code files and `God` files as active refactor candidates, and split them into clearer responsibility-based modules when that improves maintainability, scalability, or safety

## Web stack

The active web direction for the v2 phase is:

1. Go `net/http` on the shared application backend
2. a Svelte-based web client served on the same backend and auth origin as `/api/...`
3. Go remains the owner of business logic, workflow rules, approvals, reporting composition, and durable state while Svelte acts as the display and interaction layer
4. a frontend build pipeline that supports the approved Svelte migration while keeping deployment centered on the Go application boundary
5. repo-owned application styles rather than Tailwind CSS by default

Current rule:

1. treat the earlier Go-template browser layer as implemented migration source material rather than the forward target architecture
2. do not introduce a separate browser-only backend or split the auth and API model between web and mobile clients

## Thin-v1 foundation delivered

1. bootstrap the Go module
2. add a migration runner
3. create the first control-boundary schema for orgs, users, memberships, audit, and idempotency
4. add the first shared document, approval, session/auth, and AI traceability foundations
5. complete the accounting foundation with ledger accounts, append-only journals, centralized document posting, reversals, tax seams, accounting periods, and review queries
6. add the first `inventory_ops` foundation with items, locations, append-only movements, stock derivation, source and destination semantics, inventory document payload ownership, and explicit accounting/execution handoff seams
7. add the first `work_orders` foundation with work-order truth, append-only status history, transactional consumption of pending inventory execution links into work-order material usage, workflow-owned tasks with one accountable worker, workforce-owned labor capture with cost snapshots, and centralized accounting consumption of both labor and work-order-linked inventory handoffs
8. add support-depth `parties` and `contacts` records needed by the thin-v1 invoice, payment or receipt, trading inventory, and service execution flows without reviving a primary CRM module
9. add persist-first inbound requests, request-message attachments, transcription derivatives, queue-oriented AI request processing seams, and reporting-visible request -> AI -> approval -> document review

The planning documents in [`new_app_docs/`](./new_app_docs) remain the canonical source for scope, sequencing, and module boundaries.

Anything under [`examples/`](./examples) is reference-only, read-only material from older implementations or planning eras and is not part of the active `workflow_app` implementation surface. The retired accounting-agent proof-of-concept now lives only as an external historical reference at https://github.com/signalmachine/accounting-agent-app.

The durable workflow-reference layer for supported operator paths, reusable end-to-end validation checklists, live workflow-review evidence, and later user-guide preparation lives in [`docs/workflows/`](./docs/workflows).

Testing and verification guidance lives in [`docs/technical_guides/07_testing_and_verification.md`](./docs/technical_guides/07_testing_and_verification.md).

## Current commands

Apply migrations:

```bash
go run ./cmd/migrate
```

Build the current workspace:

```bash
go build ./cmd/... ./internal/...
```

Bootstrap a friendly main-database admin login for browser sign-in:

```bash
go run ./cmd/bootstrap-admin -password 'choose-a-strong-password'
```

The bootstrap command is idempotent. By default it ensures:

1. org name `North Harbor Works`
2. org slug `north-harbor`
3. admin email `admin@northharbor.local`
4. admin display name `North Harbor Admin`

You can override any of those with flags such as `-org-name`, `-org-slug`, `-email`, and `-display-name`.

Run tests with the configured test database:

```bash
set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...
```

The test suite still uses `TEST_DATABASE_URL`; direct shell exports continue to win over `.env` when both are present. Do not point the canonical test command at the main `DATABASE_URL`. Prefer a local disposable PostgreSQL instance for `TEST_DATABASE_URL`; the DB-backed suite is serialized and materially more stable and faster on local Postgres than on a remote shared test database.

Run targeted race detection for concurrency-sensitive packages when needed:

```bash
go test -race ./path/to/package
```

`go test -race` is not currently part of the repository's standard full-suite verification path.

Optional OpenAI configuration for the Milestone 6 live-provider path:

```bash
OPENAI_API_KEY=...
OPENAI_MODEL=...
```

Run the first application API surface:

```bash
go run ./cmd/app
```

`cmd/app`, `cmd/bootstrap-admin`, `cmd/migrate`, `cmd/set-password`, and `cmd/verify-agent` now auto-load `.env` from the repository root when it is present, without overriding any variables already exported in the shell or passed by flag.

Install the frontend toolchain for the Svelte browser work:

```bash
npm --prefix web install
```

Run the frontend verification commands:

```bash
npm --prefix web run check
npm --prefix web run test
npm --prefix web run build
```

`cmd/app` serves the embedded Svelte frontend at `/app` on the promoted runtime path.

Repository rule for generated frontend artifacts:

1. `web/.svelte-kit/` is generated workspace state and is not part of the committed source surface
2. `internal/app/web_dist/` remains the committed built frontend artifact for the current Go embed and `/app` static-serving path unless the repository later moves that build responsibility fully into CI or release packaging

Open the first browser operator surface:

```text
http://127.0.0.1:8080/app
http://127.0.0.1:8080/app/routes
http://127.0.0.1:8080/app/settings
http://127.0.0.1:8080/app/admin
http://127.0.0.1:8080/app/admin/accounting
http://127.0.0.1:8080/app/admin/parties
http://127.0.0.1:8080/app/admin/access
http://127.0.0.1:8080/app/admin/inventory
http://127.0.0.1:8080/app/operations
http://127.0.0.1:8080/app/review
http://127.0.0.1:8080/app/inventory
http://127.0.0.1:8080/app/operations-feed
http://127.0.0.1:8080/app/agent-chat
http://127.0.0.1:8080/app/submit-inbound-request
```

Open the current downstream browser review surfaces:

```text
http://127.0.0.1:8080/app/review/inbound-requests
http://127.0.0.1:8080/app/review/documents
http://127.0.0.1:8080/app/review/documents/<document-uuid>
http://127.0.0.1:8080/app/review/accounting
http://127.0.0.1:8080/app/review/accounting/<journal-entry-uuid>
http://127.0.0.1:8080/app/review/approvals
http://127.0.0.1:8080/app/review/approvals/<approval-uuid>
http://127.0.0.1:8080/app/review/proposals
http://127.0.0.1:8080/app/review/proposals/<recommendation-uuid>
http://127.0.0.1:8080/app/review/inventory
http://127.0.0.1:8080/app/review/inventory/<movement-uuid>
http://127.0.0.1:8080/app/review/work-orders
http://127.0.0.1:8080/app/review/work-orders/<work-order-uuid>
http://127.0.0.1:8080/app/review/audit
http://127.0.0.1:8080/app/review/audit/<audit-event-uuid>
http://127.0.0.1:8080/app/inbound-requests/step:<agent-step-uuid>
```

Start a browser-usable session and capture cookies:

```bash
curl -c cookies.txt -X POST http://127.0.0.1:8080/api/session/login \
  -H "Content-Type: application/json" \
  -d '{
    "org_slug":"<org-slug>",
    "email":"<user-email>",
    "password":"<password>",
    "device_label":"browser"
  }'
```

Inspect the active browser session:

```bash
curl -b cookies.txt http://127.0.0.1:8080/api/session
```

Start a non-browser bearer session and capture the returned tokens:

```bash
curl -X POST http://127.0.0.1:8080/api/session/token \
  -H "Content-Type: application/json" \
  -d '{
    "org_slug":"<org-slug>",
    "email":"<user-email>",
    "password":"<password>",
    "device_label":"mobile"
  }'
```

Set or rotate a test user password:

```bash
go run ./cmd/set-password -user-id <user-uuid> -password '<password>'
```

For first-run browser access on the main database, prefer `go run ./cmd/bootstrap-admin -password '<password>'` over manual user-ID lookups and password rotation.

Refresh a non-browser bearer session:

```bash
curl -X POST http://127.0.0.1:8080/api/session/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "session_id":"<session-uuid>",
    "refresh_token":"<refresh-token>"
  }'
```

Inspect the active session through bearer auth:

```bash
curl http://127.0.0.1:8080/api/session \
  -H "Authorization: Bearer <access-token>"
```

Trigger the queued-request AI processor through HTTP with bearer auth:

```bash
curl -X POST http://127.0.0.1:8080/api/agent/process-next-queued-inbound-request \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access-token>" \
  -d '{"channel":"browser"}'
```

Run the focused live-provider verification command:

```bash
set -a; source .env; set +a; go run ./cmd/verify-agent
```

Submit an inbound request with an inline attachment through HTTP:

```bash
curl -X POST http://127.0.0.1:8080/api/inbound-requests \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "channel":"browser",
    "metadata":{"submitter_label":"front desk"},
    "message":{"message_role":"request","text_content":"Urgent pump issue reported from the warehouse."},
    "attachments":[
      {
        "original_file_name":"note.txt",
        "media_type":"text/plain",
        "content_base64":"dXJnZW50IHB1bXAgZmFpbHVyZSBkZXRhaWxz"
      }
    ]
  }'
```

Submit an inbound request through the same shared API seam with bearer auth:

```bash
curl -X POST http://127.0.0.1:8080/api/inbound-requests \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access-token>" \
  -d '{
    "origin_type":"human",
    "channel":"mobile",
    "metadata":{"submitter_label":"field app"},
    "message":{"message_role":"request","text_content":"Need urgent operator review from the mobile client."}
  }'
```

Download a persisted attachment through HTTP:

```bash
curl -L http://127.0.0.1:8080/api/attachments/<attachment-uuid>/content \
  -b cookies.txt \
  -o attachment.bin
```

List inbound requests queued or processed for operator review:

```bash
curl "http://127.0.0.1:8080/api/review/inbound-requests?status=processed" \
  -b cookies.txt
```

Load one inbound request review detail by stable request reference:

```bash
curl "http://127.0.0.1:8080/api/review/inbound-requests/REQ-000001" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/inbound-requests/step:<agent-step-uuid>" \
  -b cookies.txt
```

List processed proposals and approval queue entries:

```bash
curl "http://127.0.0.1:8080/api/review/processed-proposals?request_reference=REQ-000001" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/approval-queue?status=pending" \
  -b cookies.txt
```

List document, accounting, inventory, work-order, and audit review surfaces through the same browser-session-backed API seam:

```bash
curl "http://127.0.0.1:8080/api/review/documents" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/documents?document_id=<document-uuid>" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/accounting/journal-entries" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/accounting/journal-entries?document_id=<document-uuid>" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/accounting/journal-entries?entry_id=<journal-entry-uuid>" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/accounting/control-account-balances" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/accounting/control-account-balances?control_type=gst_output" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/accounting/tax-summaries" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/accounting/tax-summaries?tax_type=gst" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/accounting/tax-summaries?tax_code=GST18" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/inventory/stock" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/inventory/movements" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/inventory/reconciliation" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/work-orders" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/work-orders?work_order_id=<work-order-uuid>" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/work-orders?document_id=<document-uuid>" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/work-orders/<work-order-uuid>" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/audit-events?entity_type=work_orders.work_order&entity_id=<work-order-uuid>" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/audit-events?event_id=<audit-event-uuid>" \
  -b cookies.txt
```

Decide an approval through the same API surface:

```bash
curl -X POST http://127.0.0.1:8080/api/approvals/<approval-uuid>/decision \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"decision":"approved","decision_note":"Looks correct."}'
```

Revoke the active browser session:

```bash
curl -b cookies.txt -X POST http://127.0.0.1:8080/api/session/logout
```

Revoke a bearer-authenticated session:

```bash
curl -X POST http://127.0.0.1:8080/api/session/logout \
  -H "Authorization: Bearer <access-token>"
```

## Current implementation status

Implemented:

1. migration runner with applied-migration tracking
2. control-boundary schema for orgs, users, memberships, audit, and idempotency
3. shared document kernel for thin-v1 document families
4. workflow approvals, approval queue entries, and approval decisions
5. device-scoped sessions and role-aware authorization around document and approval service actions
6. AI tool registry, tool policy, run history, artifacts, recommendations, approval linkage, and delegation traces
7. accounting ledger accounts plus centralized, idempotent, document-linked journal posting and reversal with database-backed balance enforcement
8. GST and TDS tax foundation records with tax-aware posting validation against active tax codes and control accounts
9. accounting periods with effective-date posting control, journal review queries, and control-account balance views for receivable/payable readiness
10. inventory items, locations, movement numbering, append-only movement truth, derived stock queries, inventory document payload ownership, and receipt/issue/adjustment movement recording with purpose and usage classification
11. pending execution-context links plus costed inventory accounting handoffs for inventory document lines so later modules can consume inventory outcomes without crossing ownership boundaries
12. first-class work-order records with append-only execution status history and material-usage records derived from pending inventory execution links
13. shared workflow tasks linked to work orders with one accountable worker plus workforce workers and append-only labor entries with captured cost snapshots
14. pending labor-accounting handoffs from `workforce` plus centralized `accounting` consumption of approved journal documents for work-order labor costs
15. centralized `accounting` consumption of costed inventory handoffs for work-order material usage through approved journal documents
16. first-class `reporting` read surfaces for approval queue review, document review, accounting journal review, control-account balance review, GST/TDS tax summaries, inventory stock review, inventory movement review, inventory reconciliation review, work-order review, and audit lookup
17. support-depth `parties` records plus tenant-safe `contacts` for thin-v1 trading and service document flows
18. one-to-one work-order document ownership through `work_orders.documents`, with transactional creation of the shared document row plus work-order execution truth
19. one-to-one invoice and payment or receipt document ownership through accounting-owned payload tables keyed by `document_id`
20. persist-first inbound request drafts, org-scoped `REQ-...` request references, draft editing and hard deletion, queued-request amend and cancel handling before pickup, messages, queue claim and status transitions, PostgreSQL-backed attachments, attachment transcription derivatives, and AI run causation linked back to the originating request
21. `reporting` review surfaces for inbound requests, request attachments, linked AI runs, AI step traces, delegation traces, AI artifacts, recommendation payloads, and processed proposals joined to approvals and documents, with stable request references exposed for operator tracking and used directly for request-detail and processed-proposal lookup, persisted cancellation and failure reasons visible for operator troubleshooting, submitter, session, metadata, attachment provenance, and document-context fields surfaced for richer operator review, and queue-oriented status summary reads for inbound requests and processed proposals
22. optional OpenAI provider configuration loading and validation in `internal/ai`, keeping live-provider setup explicit while default repository verification remains provider-independent
23. the official OpenAI Go SDK plus a Responses-API-backed coordinator provider in `internal/ai`, with a queued inbound-request execution path that claims one request, assembles request, attachment, and derived-text context, runs a hard-capped tool loop, enforces per-capability tool policy, auto-executes the first reporting read tool when allowed, can optionally persist one allowlisted specialist child run plus delegation record, persists coordinator or specialist step, artifact, and recommendation tool-execution metadata, and marks the request `processed` or `failed` according to the provider-backed outcome
24. a shared backend-facing `internal/app` agent-processing contract that drives the queued coordinator path outside direct package wiring, plus an opt-in `cmd/verify-agent` live-provider verification command and integration coverage built on that shared seam
25. the first widened HTTP API contract set over that seam at `POST /api/session/login`, `GET /api/session`, `POST /api/session/logout`, `POST /api/agent/process-next-queued-inbound-request`, `POST /api/inbound-requests`, `GET /api/attachments/{attachment_id}/content`, `GET /api/review/inbound-requests`, `GET /api/review/inbound-request-status-summary`, `GET /api/review/inbound-requests/{request_reference_or_id}`, `GET /api/review/processed-proposals`, `GET /api/review/processed-proposal-status-summary`, `GET /api/review/approval-queue`, and `POST /api/approvals/{approval_id}/decision`, including browser-session cookies, bearer-session auth, explicit active-org session promotion from org slug plus user email plus password, one-workflow request submission with optional inline attachments, attachment download, reporting-backed operator review reads, approval decisions routed through the existing workflow boundary, provider-not-configured and queue-empty handling, and a minimal `cmd/app` server entrypoint for browser or API-driven testing
26. the first real browser application slice at `/app`, including browser sign-in on the shared session seam, inbound-request submission with file attachments, process-next queue execution, recent inbound-request and pending-approval review, inbound-request detail with attachment, AI, and proposal inspection, and browser-driven approval decisions on the same shared backend foundation
27. the next downstream browser review slice at `/app/review/documents` and `/app/review/accounting`, plus shared backend review endpoints at `GET /api/review/documents`, `GET /api/review/accounting/journal-entries`, `GET /api/review/accounting/control-account-balances`, and `GET /api/review/accounting/tax-summaries`, all available through the same browser session-cookie auth path so operators can continue from approvals into document and accounting review without leaving the app
28. the next widened browser review slice at `/app/review/inventory`, `/app/review/work-orders`, `/app/review/work-orders/{work_order_id}`, and `/app/review/audit`, plus shared backend review endpoints at `GET /api/review/inventory/stock`, `GET /api/review/inventory/movements`, `GET /api/review/inventory/reconciliation`, `GET /api/review/work-orders`, `GET /api/review/work-orders/{work_order_id}`, and `GET /api/review/audit-events`, all available through the same browser session-cookie auth path so operators can continue from financial review into stock, execution, and audit inspection without leaving the app
29. the latest browser continuity slice adds `/app/review/proposals`, driven by the existing `GET /api/review/processed-proposals` and `GET /api/review/processed-proposal-status-summary` backend reads, so operators can filter processed proposals by request reference, inspect proposal-status summary, and continue cleanly between inbound requests, approvals, and downstream documents without dropping back to the dashboard
30. the latest browser continuity slice adds `/app/review/approvals`, driven by the existing `GET /api/review/approval-queue` backend read, so operators can filter pending-versus-closed approval rows by queue code, act on approvals from a dedicated review page, and continue from proposal or document review into the matching approval queue slice instead of relying only on dashboard snippets
31. the latest browser continuity slice adds `/app/review/inbound-requests`, driven by the existing `GET /api/review/inbound-requests` and `GET /api/review/inbound-request-status-summary` backend reads, so operators can filter by request status or exact `REQ-...` reference, jump from request-status summary cards into the matching filtered browser list, inspect request-level AI run and recommendation status context, and continue into exact request detail without relying only on the dashboard snippet
32. the latest browser continuity slice adds exact `approval_id` and `recommendation_id` drill-down on the existing approval and proposal review seams, and extends audit lookup with direct links back into exact inbound-request, approval, and proposal review so operators can move from an audit trace into the precise browser review context instead of reopening broad lists by hand
33. the latest browser continuity slice adds exact accounting `entry_id` drill-down on the shared journal-review seam plus a dedicated `/app/review/accounting/{entry_id}` browser detail page, and extends document, approval, inventory-reconciliation, accounting-list, and audit surfaces with direct journal-entry links so operators can move from downstream financial context or audit traces into one exact posting record instead of reopening broader accounting lists by hand
34. the latest browser continuity slice turns inbound-request detail into a stronger review hub by linking request-level AI recommendations and downstream proposals into exact proposal review, exact approval review, filtered request review, and direct inbound-request or recommendation audit lookup so operators can continue from intake evidence into downstream control decisions without reconstructing context by hand
35. the latest browser continuity slice turns inventory stock review into an active browser pivot by linking stock rows into anchored filtered stock, movement-history, and reconciliation states and by routing inventory item and location audit entities back into those focused inventory views instead of leaving stock review as a dead-end table
36. the latest browser continuity slice extends work-order review with exact `work_order_id` filtering on both `/api/review/work-orders` and `/app/review/work-orders`, and it turns `/app/review/work-orders/{work_order_id}` into a stronger continuity stop by linking back into focused work-order review plus direct accounting review on the same shared seam instead of leaving work-order detail as a dead-end page
37. the latest browser continuity slice turns `/app/review/inventory/{movement_id}` into a stronger review stop by linking exact movement detail into item-focused stock, movement-history, and reconciliation views plus source-document reconciliation, source-document accounting review, and source or destination location movement history so operators can continue from one movement into adjacent stock, document, execution, and posting context without reopening broad inventory lists
38. the latest browser continuity slice extends exact inbound-request detail lookup on the shared browser and API seams to resolve `run:<agent-run-id>` and `delegation:<delegation-id>`, and it adds audit-page entity continuation for `ai.agent_run` plus `ai.agent_delegation` so provider-execution audit events now return operators to the exact inbound-request execution trail instead of dead-ending on generic audit results
39. the latest browser continuity slice extends exact inbound-request detail lookup on the shared browser and API seams to resolve `step:<agent-step-id>` as well, and it adds step-level audit continuation plus step-section audit links so `ai.agent_run_step` or `ai.agent_step` entities can land operators on the exact persisted execution step instead of only the broader request page

Immediate next steps:

1. keep the separate Milestone 10 browser-review and workflow-continuity closeout moving on `docs/workflows/`
2. promote later Milestone 12 policy or operational controls only when workflow evidence justifies another bounded admin slice
3. continue widening shared backend contracts only where they strengthen one shared truth model for browser and later clients
