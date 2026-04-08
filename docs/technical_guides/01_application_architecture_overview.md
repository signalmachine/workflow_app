# Application Architecture Overview

Date: 2026-03-31
Status: Active technical guide
Purpose: explain the end-to-end shape of `workflow_app`, the shared backend seam, and the major subsystems contributors need to understand before changing code.

## 1. What this application is

`workflow_app` is an AI-agent-first, database-first business operating system. The product is not organized around a generic CRM or a chat-first UX. It is organized around durable records, explicit review surfaces, and controlled transitions.

The center of gravity is:

1. persisted inbound requests
2. AI processing on queued requests
3. operator review of proposals, approvals, documents, accounting, inventory, work orders, and audit trails
4. browser and API surfaces that both reuse the same backend truth

The important design choice is that the browser layer does not own business truth. It renders and orchestrates shared services that already own the data model.

## 2. Major layers

The application is easiest to reason about in four layers:

1. transport layer
2. orchestration layer
3. domain service layer
4. persistence and reporting layer

### 2.1 Transport layer

This is `cmd/app/main.go` plus `internal/app`. The transport layer handles HTTP requests, route selection, auth extraction, request validation, and response formatting.

Example:

```go
server := &http.Server{
	Addr:    listenAddr,
	Handler: app.NewServedAgentAPIHandler(db),
}
```

That is intentionally thin. It does not decide business rules; it delegates to services.

### 2.2 Orchestration layer

`internal/app` is the integration surface that ties the UI, HTTP API, session auth, submission flow, queued AI processing, and review reads together.

The orchestration layer:

1. loads the right service implementations
2. normalizes input from browser forms or JSON
3. maps shared domain results into browser redirects or JSON responses
4. keeps the web and API paths on one backend contract

### 2.3 Domain service layer

The main domain packages are:

1. `internal/intake`
2. `internal/ai`
3. `internal/documents`
4. `internal/workflow`
5. `internal/accounting`
6. `internal/inventoryops`
7. `internal/workorders`
8. `internal/attachments`
9. `internal/identityaccess`
10. `internal/reporting`

Each module owns its own write path and invariants. Cross-module work happens by explicit handoff, not by one package writing directly into another package's tables.

### 2.4 Persistence and reporting layer

Persistence is PostgreSQL-backed. Reporting is a read model layer that joins module-owned records into operator-facing review views.

This matters because the browser does not query raw module tables directly. It uses reporting reads for the review surfaces so the same inspection contract can support both browser and API callers.

## 3. Startup shape

The runnable application starts in `cmd/app/main.go`.

```go
func main() {
	if err := envload.LoadDefaultIfPresent(); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	server := &http.Server{
		Addr:    listenAddr,
		Handler: app.NewServedAgentAPIHandler(db),
	}
}
```

The important thing here is the assembly model:

1. environment loading is separate from business logic
2. database wiring is separate from request handling
3. `internal/app` is the single place where the HTTP surface is composed

## 4. Core request flow

The most important product flow is:

1. create or submit an inbound request
2. queue it for AI processing
3. claim it from the queue
4. produce a provider-backed coordinator run
5. persist an artifact and recommendation
6. expose the result through review surfaces
7. optionally continue into approval, posting, or downstream execution

The architecture is designed so each step is durably persisted before the next step depends on it.

## 5. Shared truth model

The application uses canonical identifiers across modules:

1. `document_id` identifies the central business document
2. `request_id` and `REQ-...` identify inbound requests
3. `run_id`, `step_id`, and `delegation_id` identify AI execution records
4. `approval_id` identifies approval truth
5. module-owned records point back to those canonical identifiers instead of duplicating them

That shared identity model is the main reason the reporting layer can cross-link request, AI, document, approval, accounting, inventory, and execution data cleanly.

## 6. What to change carefully

The highest-risk changes are the ones that affect:

1. request lifecycle state
2. approval transitions
3. posting boundaries
4. AI tool policy or coordinator behavior
5. session auth or actor resolution
6. reporting reads that feed browser continuity

These areas are workflow-critical because they control operator trust. If one of them changes, it usually deserves integration tests and a review of the corresponding documentation.

## 7. Practical reading path

If you are new to the codebase, read in this order:

1. this guide
2. the module-boundary guide
3. the inbound-request lifecycle guide
4. the AI agent guide
5. the web and API seam guide
6. the auth guide
7. the testing guide
