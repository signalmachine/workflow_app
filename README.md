# workflow_app

`workflow_app` is an AI-agent-first, database-first business operating system centered on documents, ledgers, execution context, approvals, and reports.

This repository has completed Milestone 0, Milestone 1, and Milestone 2 from the canonical planning set in [`new_app_docs/`](./new_app_docs), completed the first Milestone 3 inventory foundation slice, and is continuing Milestone 4 execution foundation with the first execution-to-accounting bridge:

1. bootstrap the Go module
2. add a migration runner
3. create the first control-boundary schema for orgs, users, memberships, audit, and idempotency
4. add the first shared document, approval, session/auth, and AI traceability foundations
5. complete the accounting foundation with ledger accounts, append-only journals, centralized document posting, reversals, tax seams, accounting periods, and review queries
6. add the first `inventory_ops` foundation with items, locations, append-only movements, stock derivation, source and destination semantics, inventory document payload ownership, and explicit accounting/execution handoff seams
7. add the first `work_orders` foundation with work-order truth, append-only status history, transactional consumption of pending inventory execution links into work-order material usage, workflow-owned tasks with one accountable worker, workforce-owned labor capture with cost snapshots, and labor-accounting handoff consumption through centralized journal posting

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
set -a; source .env; set +a; go test ./...
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
11. pending accounting handoff rows and pending execution-context links for inventory document lines so later modules can consume inventory outcomes without crossing ownership boundaries
12. first-class work-order records with append-only execution status history and material-usage records derived from pending inventory execution links
13. shared workflow tasks linked to work orders with one accountable worker plus workforce workers and append-only labor entries with captured cost snapshots
14. pending labor-accounting handoffs from `workforce` plus centralized `accounting` consumption of approved journal documents for work-order labor costs

Next:

1. define the first narrow accounting consumer for inventory accounting handoff rows while preserving centralized posting ownership
2. keep the thin-v1 module map narrow while inventory and execution foundation expand
3. start the first thin review and reporting surfaces once the remaining inventory-accounting bridge is clear
