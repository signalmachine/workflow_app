# workflow_app

`workflow_app` is an AI-agent-first, database-first business operating system centered on documents, ledgers, execution context, approvals, and reports.

This repository has completed Milestone 0 and started Milestone 1 from the canonical planning set in [`new_app_docs/`](./new_app_docs):

1. bootstrap the Go module
2. add a migration runner
3. create the first control-boundary schema for orgs, users, memberships, audit, and idempotency
4. add the first shared document, approval, and session/auth foundations

The planning documents in [`new_app_docs/`](./new_app_docs) remain the canonical source for scope, sequencing, and module boundaries.

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

Next:

1. complete the remaining Milestone 1 AI foundation slice for run history, tool policy, recommendations, artifacts, and delegation traces
