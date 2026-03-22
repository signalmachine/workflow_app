# workflow_app

`workflow_app` is an AI-agent-first, database-first business operating system centered on documents, ledgers, execution context, approvals, reports, and the first real browser operator layer.

This repository has completed Milestone 0 through Milestone 6 from the canonical planning set in [`new_app_docs/`](./new_app_docs), and Milestone 7 is now underway with multiple browser slices landed under `/app`. The shared control boundary now includes adopted document ownership for work orders, invoices, and payment or receipt documents plus persist-first inbound request and attachment foundations with stable `REQ-...` inbound-request references for submission acknowledgments and review. Draft requests can now be edited or hard-deleted before queueing, while queued pre-processing requests can be soft-cancelled or returned to draft for amendment and resubmission. The provider-backed AI foundation is now live: `internal/ai` includes optional OpenAI configuration loading, the official OpenAI Go SDK, a Responses-API-backed provider adapter, and a coordinator flow that can claim one queued inbound request, execute a hard-capped tool loop with per-capability tool-policy enforcement, auto-run the first reporting read tool when policy allows, optionally route the result through one allowlisted specialist capability with a durable child run and delegation record, and persist the resulting run, step, artifact, and recommendation without making the default build and test flow depend on external credentials. `internal/app` now provides one shared backend seam for browser-session auth, queue processing, inbound-request submission, attachment download, operator review, approval decisions, document review, accounting review, proposal review, inventory review, work-order review, and audit lookup, plus server-rendered operator surfaces for sign-in, request intake, queue processing, request-detail inspection, approval actions, and downstream document, accounting, proposal, inventory, work-order, and audit inspection. Exact `document_id` drill-down, exact work-order review filtering by `document_id`, exact accounting journal drill-down by source `document_id`, dedicated processed-proposal review with request-reference filtering and proposal-status summary, cross-links between proposals, approvals, documents, accounting, inventory movements, work-order detail, and audit lookup, and browser rendering of persisted AI step and delegation detail on inbound-request pages are now part of that browser flow, while `cmd/verify-agent` and `cmd/app` expose the live path through focused verification and the widened runnable application server.

## Web stack

The current and preferred thin-v1 web stack is:

1. Go `net/http` on the shared application backend
2. Go `html/template` for server-rendered HTML
3. standard HTML forms and browser behavior as the baseline interaction model
4. optional `htmx` for progressive enhancement where partial-page updates materially improve operator flow
5. optional `Alpine.js` only for small local UI-state needs

Thin-v1 default rule:

1. do not introduce a separate SPA frontend or a Node-based frontend build pipeline unless the canonical planning documents are explicitly updated to require it

1. bootstrap the Go module
2. add a migration runner
3. create the first control-boundary schema for orgs, users, memberships, audit, and idempotency
4. add the first shared document, approval, session/auth, and AI traceability foundations
5. complete the accounting foundation with ledger accounts, append-only journals, centralized document posting, reversals, tax seams, accounting periods, and review queries
6. add the first `inventory_ops` foundation with items, locations, append-only movements, stock derivation, source and destination semantics, inventory document payload ownership, and explicit accounting/execution handoff seams
7. add the first `work_orders` foundation with work-order truth, append-only status history, transactional consumption of pending inventory execution links into work-order material usage, workflow-owned tasks with one accountable worker, workforce-owned labor capture with cost snapshots, and centralized accounting consumption of both labor and work-order-linked inventory handoffs
8. add support-depth `parties` and `contacts` records needed by thin-v1 invoice, payment or receipt, trading inventory, and service execution flows without reviving a primary CRM module
9. add persist-first inbound requests, request-message attachments, transcription derivatives, queue-oriented AI request processing seams, and reporting-visible request -> AI -> approval -> document review

The planning documents in [`new_app_docs/`](./new_app_docs) remain the canonical source for scope, sequencing, and module boundaries.

Testing guidance for collaborating with Codex on Go tests lives in [`docs/testing/README.md`](./docs/testing/README.md).

## Current commands

Apply migrations:

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/workflow_app?sslmode=disable go run ./cmd/migrate
```

Build the current workspace:

```bash
go build ./...
```

Run tests with the configured test database:

```bash
set -a; source .env; set +a; go test -p 1 ./...
```

Optional OpenAI configuration for the Milestone 6 live-provider path:

```bash
OPENAI_API_KEY=...
OPENAI_MODEL=...
```

Run the first application API surface:

```bash
set -a; source .env; set +a; go run ./cmd/app
```

Open the first browser operator surface:

```text
http://127.0.0.1:8080/app
```

Open the current downstream browser review surfaces:

```text
http://127.0.0.1:8080/app/review/documents
http://127.0.0.1:8080/app/review/accounting
http://127.0.0.1:8080/app/review/proposals
http://127.0.0.1:8080/app/review/inventory
http://127.0.0.1:8080/app/review/work-orders
http://127.0.0.1:8080/app/review/audit
```

Start a browser-usable session and capture cookies:

```bash
curl -c cookies.txt -X POST http://127.0.0.1:8080/api/session/login \
  -H "Content-Type: application/json" \
  -d '{
    "org_slug":"<org-slug>",
    "email":"<user-email>",
    "device_label":"browser"
  }'
```

Inspect the active browser session:

```bash
curl -b cookies.txt http://127.0.0.1:8080/api/session
```

Trigger the queued-request AI processor through HTTP:

```bash
curl -X POST http://127.0.0.1:8080/api/agent/process-next-queued-inbound-request \
  -H "Content-Type: application/json" \
  -H "X-Workflow-Org-ID: <org-uuid>" \
  -H "X-Workflow-User-ID: <user-uuid>" \
  -H "X-Workflow-Session-ID: <session-uuid>" \
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

curl "http://127.0.0.1:8080/api/review/accounting/control-account-balances" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/accounting/tax-summaries" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/inventory/stock" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/inventory/movements" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/inventory/reconciliation" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/work-orders" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/work-orders?document_id=<document-uuid>" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/work-orders/<work-order-uuid>" \
  -b cookies.txt

curl "http://127.0.0.1:8080/api/review/audit-events?entity_type=work_orders.work_order&entity_id=<work-order-uuid>" \
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
25. the first widened HTTP API contract set over that seam at `POST /api/session/login`, `GET /api/session`, `POST /api/session/logout`, `POST /api/agent/process-next-queued-inbound-request`, `POST /api/inbound-requests`, `GET /api/attachments/{attachment_id}/content`, `GET /api/review/inbound-requests`, `GET /api/review/inbound-request-status-summary`, `GET /api/review/inbound-requests/{request_reference_or_id}`, `GET /api/review/processed-proposals`, `GET /api/review/processed-proposal-status-summary`, `GET /api/review/approval-queue`, and `POST /api/approvals/{approval_id}/decision`, including browser-session cookies, explicit active-org session promotion from org slug plus user email, compatibility with the existing UUID request-actor headers for automation, queued-request processing, one-workflow request submission with optional inline attachments, attachment download, reporting-backed operator review reads, approval decisions routed through the existing workflow boundary, provider-not-configured and queue-empty handling, and a minimal `cmd/app` server entrypoint for browser or API-driven testing
26. the first real browser application slice at `/app`, including server-rendered browser sign-in, inbound-request submission with file attachments, process-next queue execution, recent inbound-request and pending-approval review, inbound-request detail with attachment, AI, and proposal inspection, and browser-driven approval decisions on the same shared backend foundation
27. the next downstream browser review slice at `/app/review/documents` and `/app/review/accounting`, plus shared backend review endpoints at `GET /api/review/documents`, `GET /api/review/accounting/journal-entries`, `GET /api/review/accounting/control-account-balances`, and `GET /api/review/accounting/tax-summaries`, all available through the same browser session-cookie auth path so operators can continue from approvals into document and accounting review without leaving the app
28. the next widened browser review slice at `/app/review/inventory`, `/app/review/work-orders`, `/app/review/work-orders/{work_order_id}`, and `/app/review/audit`, plus shared backend review endpoints at `GET /api/review/inventory/stock`, `GET /api/review/inventory/movements`, `GET /api/review/inventory/reconciliation`, `GET /api/review/work-orders`, `GET /api/review/work-orders/{work_order_id}`, and `GET /api/review/audit-events`, all available through the same browser session-cookie auth path so operators can continue from financial review into stock, execution, and audit inspection without leaving the app
29. the latest browser continuity slice adds `/app/review/proposals`, driven by the existing `GET /api/review/processed-proposals` and `GET /api/review/processed-proposal-status-summary` backend reads, so operators can filter processed proposals by request reference, inspect proposal-status summary, and continue cleanly between inbound requests, approvals, and downstream documents without dropping back to the dashboard

Immediate next steps:

1. tighten operator continuity and drill-downs on top of the landed `/app` review surfaces on the same backend contracts
2. continue widening backend contracts only where the browser layer proves a concrete need, without creating a second truth owner
3. keep Milestone 7 focused on one coherent operator loop at a time on backend contracts that a later v2 mobile client will also reuse
