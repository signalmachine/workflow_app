# workflow_app

`workflow_app` is an AI-agent-first, database-first business operating system centered on documents, ledgers, execution context, approvals, and reports.

This repository has completed Milestone 0, Milestone 2, Milestone 3, and Milestone 4 from the canonical planning set in [`new_app_docs/`](./new_app_docs). Milestone 1 remains partially complete because the shared control boundary is in place but thin-v1 still lacks adopted document ownership for invoice and payment or receipt documents plus persist-first inbound request and attachment foundations. Work orders now adopt the shared document kernel through a one-to-one `work_orders.documents` bridge. The repository is now continuing through those remaining thin-v1 foundation gaps before any broader implementation proceeds:

1. bootstrap the Go module
2. add a migration runner
3. create the first control-boundary schema for orgs, users, memberships, audit, and idempotency
4. add the first shared document, approval, session/auth, and AI traceability foundations
5. complete the accounting foundation with ledger accounts, append-only journals, centralized document posting, reversals, tax seams, accounting periods, and review queries
6. add the first `inventory_ops` foundation with items, locations, append-only movements, stock derivation, source and destination semantics, inventory document payload ownership, and explicit accounting/execution handoff seams
7. add the first `work_orders` foundation with work-order truth, append-only status history, transactional consumption of pending inventory execution links into work-order material usage, workflow-owned tasks with one accountable worker, workforce-owned labor capture with cost snapshots, and centralized accounting consumption of both labor and work-order-linked inventory handoffs
8. add support-depth `parties` and `contacts` records needed by thin-v1 invoice, payment or receipt, trading inventory, and service execution flows without reviving a primary CRM module

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

Immediate next steps:

1. complete adopted document-family ownership for invoice and payment or receipt payloads with one-to-one linkage back to the shared `documents` kernel, reusing shared support-record identities where applicable
2. implement minimum persist-first inbound request intake, attachment references, queue-oriented AI processing, and browser-usable review visibility for thin-v1 user testing
3. finish the remaining thin-v1 reporting polish after those foundation gaps land
