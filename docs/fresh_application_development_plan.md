# Fresh Application Development Plan

Date: 2026-04-16
Status: Draft planning document for a fresh split-application implementation
Purpose: define the architecture and phased rollout for a new AI-agent-first application stack that uses a traditional ERP simulator backend as its integration target.

## 1. Decision Summary

Start a fresh implementation rather than continuing directly with the current combined `workflow_app` runtime.

The current codebase contains useful lessons and reference implementations, but it did not prove the most important workflow early enough:

```text
natural language accounting request
-> persisted AI request
-> structured AI proposal
-> backend-owned business document
-> human approval or explicit posting action
-> accounting journal entry
-> traceability back to the original request
```

The new implementation should make that workflow the first meaningful delivery checkpoint.

The new system should be split into separate applications:

1. backend application: a deterministic traditional ERP/accounting simulator
2. AI agent application: the primary product and intelligence layer

The backend application is not intended for production deployment as a real ERP. It is a realistic stand-in for systems like Tally, Zoho, QuickBooks, or SAP so that the AI agent application can be developed and tested against credible business constraints.

This document is not an implementation plan for continuing the current `workflow_app` codebase. It is a fresh-start plan that uses the current codebase and its Milestone 15 and Milestone 16 review findings as lessons.

## 2. Product Boundary

### 2.1 Backend Application Role

The backend is a traditional business-system simulator.

It should support only the kinds of capabilities a normal accounting, inventory, order-management, or ERP-style application would expose:

1. master data
2. accounting documents
3. inventory documents and movements
4. order documents
5. tax rules and tax codes
6. approval or posting status where traditional systems require it
7. reports
8. REST APIs
9. a development REPL that uses the REST API

The backend should be dumb about natural language and AI, but it should not be a toy. It must enforce normal business-system invariants so that the AI app is tested against realistic constraints.

The backend should enforce:

1. double-entry accounting balance
2. document lifecycle and posting boundaries
3. master-data validation
4. GST and TDS rules where supported
5. inventory movement rules
6. idempotent mutation behavior where API retries are likely
7. audit events for material state transitions
8. stable report outputs

The backend should not contain:

1. chat messages
2. natural language interpretation
3. router agents
4. specialist agents
5. prompts or prompt versions
6. model-provider integrations
7. agent memory
8. agent runs, steps, tool calls, or recommendations
9. AI-specific proposal approval logic

Traditional approval or posting status is allowed in the backend when it models what a normal business application would expose. AI-specific review, recommendation, and agent-confidence concepts belong only in the AI application.

### 2.2 AI Agent Application Role

The AI agent application is the main product.

It should:

1. receive user natural language messages and structured system messages
2. persist every message before processing
3. classify intent through a router agent
4. reject unsupported or no-impact requests clearly
5. enqueue supported work for specialist agents
6. use OpenAI Responses API with strict structured outputs
7. validate extracted proposals against backend REST APIs
8. create or update backend records only through backend REST APIs
9. persist agent runs, tool calls, model outputs, proposals, failures, and backend API calls
10. expose a REPL first, then a web UI later

The AI application may propose and prepare work. It must not own accounting truth, inventory truth, or order truth.

The AI application should treat the backend as an external system. Even during local development, it should call the backend through REST APIs rather than importing backend packages or writing to the backend database directly.

## 3. High-Level Architecture

```text
AI Agent REPL / future web UI
        |
        v
AI Agent API
        |
        v
AI Agent Worker
router agent + specialist agents
        |
        | REST API calls, idempotency keys, API-client identity
        v
Business Backend API
traditional ERP simulator
        |
        v
Business PostgreSQL database
```

Separate database ownership:

```text
Business DB:
  orgs, users, parties, accounting, tax, inventory, orders, workflow states, reports, audit

AI Agent DB:
  messages, requests, runs, steps, tool calls, proposals, model outputs, backend API calls, failures, feedback
```

The AI agent application must not import backend Go packages. Integration should happen through REST contracts and generated or hand-written typed clients.

Each cross-application request should carry a correlation identifier so a workflow can be traced across both databases:

```text
ai_request_id
ai_run_id
idempotency_key
backend_document_id
backend_journal_entry_id
```

## 4. Repository Layout

Recommended repositories:

```text
business-backend/
ai-agent-app/
```

Optional later repository:

```text
api-contracts/
```

The contract repository is optional at the start. If not created separately, keep the backend OpenAPI contract in the backend repo and generate or copy the AI app client from it.

Each repo should be independently runnable and testable. A local developer script may start both applications together, but the applications should not depend on one shared process or one shared database.

## 5. Backend Application Shape

Suggested backend structure:

```text
business-backend/
  cmd/
    api/
    migrate/
    repl/
    seed-demo/
  internal/
    platform/
      config/
      db/
      audit/
      idempotency/
      httpx/
    identity/
    parties/
    accounting/
      ledger/
      tax/
      periods/
      documents/
      posting/
      reports/
    inventory/
      items/
      locations/
      movements/
    orders/
      purchase/
      sales/
    workflow/
      approvals/
    api/
      rest/
      openapi/
```

Use Go and native `pgxpool` for PostgreSQL access.

The backend API should look like a traditional business application API, not an AI helper API.

Good endpoint examples:

```text
GET  /api/v1/ledger-accounts
POST /api/v1/ledger-accounts
GET  /api/v1/tax-codes
GET  /api/v1/parties
POST /api/v1/vendor-invoices
POST /api/v1/vendor-invoices/{id}/submit
POST /api/v1/vendor-invoices/{id}/approve
POST /api/v1/vendor-invoices/{id}/post
GET  /api/v1/journal-entries
GET  /api/v1/reports/trial-balance
```

Avoid AI-shaped backend endpoints:

```text
POST /api/v1/book-natural-language-invoice
POST /api/v1/classify-business-event
POST /api/v1/create-ai-proposal
```

Those belong in the AI agent application.

## 6. AI Agent Application Shape

Suggested AI app structure:

```text
ai-agent-app/
  cmd/
    api/
    worker/
    repl/
    migrate/
  internal/
    platform/
      config/
      db/
      queue/
      httpx/
      audit/
    intake/
      messages/
      requests/
    agents/
      router/
      accounting/
        vendorinvoice/
        manualjournal/
    llm/
      openai/
      prompts/
      schemas/
    backendclient/
      businessapi/
    proposals/
    policy/
    recovery/
    api/
      rest/
```

Use OpenAI Go SDK with the Responses API.

The AI app should use strict structured outputs for every agent result. Free-text model output may be stored for traceability, but it should not be the data shape used to call the backend.

## 7. Lessons From Current Milestone 15 And Milestone 16

The new app should treat the current Milestone 15 and Milestone 16 findings as design requirements.

### 7.0 Current-App Issue Mapping

The following current-app issues should directly shape the fresh implementation:

| Current issue | Fresh-app design response |
|---|---|
| API auth and method responses drifted from client expectations | Define JSON API error contracts in Phase 0 and test them in Phase 1 |
| UI routes existed without complete API backing | Do not add REPL or UI commands unless the real REST endpoint exists |
| Attachment bodies were not bounded | Add request-size limits before file or attachment support enters scope |
| AI tool-policy reads used write-style locking | Separate read and write paths from the start |
| OpenAI timeout budget covered too much work | Use per-call model timeouts and explicit job-level deadlines |
| Tool failures were not visible to operators | Persist degraded-mode and tool-failure metadata in agent runs |
| Specialist delegation could imply work that did not happen | Persist router and specialist runs separately only when both actually execute |
| Generic AI recommendation did not become an accounting proposal | Phase 1 must create a backend vendor invoice draft, not only a text recommendation |
| End-to-end accounting workflow arrived too late | Phase 1 completion requires a posted journal entry from natural language input |

### 7.1 API And Runtime Lessons

Build these rules into Phase 1:

1. unauthorized requests return sanitized `401`
2. wrong methods return JSON `405`, not HTML `404`
3. internal errors are not leaked to clients
4. request bodies have explicit size limits before attachments are introduced
5. API routes must not be advertised unless a real handler exists
6. JSON error contracts are tested
7. idempotency is required on mutating endpoints that the AI app may retry

### 7.2 AI Runtime Lessons

Build these rules into Phase 1:

1. every message is persisted before model processing
2. every agent run has persisted steps and outputs
3. every backend API call made by the agent is recorded
4. tool failures and blocked tools are visible in the run result
5. model timeouts are per call, not one exhausted timeout over a multi-call loop
6. router and specialist agents are separate only when both actually execute
7. do not imply a specialist ran when the router output was reused
8. strict schemas are mandatory for accounting proposals
9. missing data produces a reviewable missing-data state, not a half-created document
10. unsupported non-accounting requests do not create accounting placeholders

### 7.3 Workflow Lessons

The first phase must prove an end-to-end workflow. Do not build months of foundation before discovering whether the core agent workflow works.

The first proof workflow should be:

```text
vendor invoice natural language request
-> AI extraction
-> backend vendor invoice draft
-> submit/approve/post
-> balanced journal entry
-> trial balance reflects the transaction
```

## 8. Phased Implementation Plan

### Phase 0: Contracts And Local Harness

Goal: create a thin but real split-system harness.

Deliver:

1. `business-backend` repo initialized
2. `ai-agent-app` repo initialized
3. Go modules using `pgxpool`
4. migration commands for both databases
5. health endpoints for both apps
6. backend OpenAPI contract for Phase 1 endpoints
7. typed backend client in the AI app
8. local run instructions for both apps
9. demo seed command for backend
10. basic test database setup

Exit criteria:

1. backend starts and passes health check
2. AI app starts and passes health check
3. AI app can call backend health endpoint
4. migrations are repeatable on clean databases
5. seed command can create the Phase 1 demo baseline idempotently

### Phase 1: Vendor Invoice To Posted Journal Entry

Goal: prove the full split-system architecture with one minimum workflow.

Phase 1 should be intentionally narrow. It is allowed to include only the backend capability needed to test one AI-driven vendor invoice workflow. Anything not needed for this workflow should be deferred.

Backend deliverables:

1. org or company record
2. API client or user identity sufficient for local development
3. parties with vendor support
4. ledger accounts
5. GST tax codes
6. accounting periods
7. vendor invoice document
8. submit, approve, and post actions
9. journal entry posting
10. trial balance report
11. audit events for material state transitions

AI app deliverables:

1. message intake API or REPL command
2. persisted request record
3. router agent
4. vendor invoice specialist agent
5. strict vendor invoice extraction schema
6. backend validation calls for parties, accounts, tax codes, and open periods
7. create backend vendor invoice draft through REST
8. persist proposal and backend document ID
9. visible missing-data state
10. queue worker and `process-next` REPL command

Required demo request:

```text
Book vendor invoice from ABC Traders for office supplies Rs 10,000 plus 18% GST dated 15 Apr 2026.
```

Expected accounting effect:

```text
Dr Office Supplies Expense        10000
Dr Input GST                       1800
    Cr Accounts Payable           11800
```

Exit criteria:

1. natural-language request is persisted in the AI database
2. router classifies it as an accounting vendor invoice
3. specialist extracts structured proposal data
4. backend vendor invoice draft is created through REST
5. proposal links to backend document ID
6. document can be submitted, approved, and posted
7. journal entry is balanced
8. trial balance remains balanced
9. traceability exists from AI message to request, run, proposal, backend document, and journal entry

This phase is not complete if it only creates a generic recommendation or only creates an unposted draft.

Recommended Phase 1 build order:

1. implement backend seed data and read APIs for parties, accounts, tax codes, and periods
2. implement backend vendor invoice create, submit, approve, post, journal-entry, and trial-balance behavior
3. verify backend workflow manually through REST or backend REPL without AI
4. implement AI app message persistence and queue claim
5. implement router classification with a deterministic fake provider first
6. implement vendor invoice specialist with a deterministic fake provider first
7. connect the AI app to backend validation and draft creation APIs
8. run the full workflow with fake providers until it passes repeatedly
9. replace fake provider with OpenAI Responses API behind the same schema
10. keep fake-provider tests as regression coverage

Minimum Phase 1 backend tables:

1. companies or orgs
2. API clients or users
3. parties
4. ledger accounts
5. tax codes
6. accounting periods
7. vendor invoices
8. vendor invoice lines
9. journal entries
10. journal lines
11. audit events
12. idempotency keys

Minimum Phase 1 AI tables:

1. messages
2. requests
3. queue jobs
4. agent runs
5. agent run steps
6. tool calls or backend API calls
7. proposals
8. proposal-backend links
9. run failures

Phase 1 non-goals:

1. customer invoices
2. payments and receipts
3. inventory
4. purchase orders
5. TDS
6. attachments
7. browser UI
8. multi-agent orchestration beyond router plus vendor-invoice specialist
9. autonomous posting by the AI app
10. broad admin screens

### Phase 2: Failure, Recovery, And Safety

Goal: make the first workflow trustworthy under bad input and retry conditions.

Deliver:

1. no-accounting-impact classification
2. unsupported-business-event classification
3. needs-clarification state
4. failed-job state
5. retry and requeue commands
6. idempotency for backend draft creation
7. duplicate-request protection for AI retries
8. OpenAI timeout and retry policy
9. visible tool-failure and degraded-mode metadata
10. API contract tests for auth, method, and error response behavior

Exit criteria:

1. missing vendor or date does not create a backend invoice
2. no-accounting-impact request creates no accounting proposal
3. retrying after a transient failure does not create duplicate invoices
4. failed requests can be inspected and requeued
5. model/tool failure is visible in the AI run output

### Phase 3: Manual Journal Workflow

Goal: prove a second accounting path without broad ERP expansion.

Deliver:

1. manual journal classification
2. strict debit/credit line schema
3. balance validation before backend calls
4. backend journal voucher or accounting document
5. approval before posting
6. traceability from AI proposal to posted journal entry

Exit criteria:

1. balanced manual journal request posts correctly after approval
2. unbalanced model output is rejected before backend persistence
3. missing account references produce needs-clarification state

### Phase 4: Payment Or Receipt Continuity

Goal: prove document continuity after an invoice exists.

Recommended first choice: vendor payment against a posted or approved vendor invoice.

Deliver:

1. payable lookup
2. payment document
3. cash or bank account selection
4. payment posting
5. payable balance report

Exit criteria:

1. payment links to the original vendor invoice
2. accounts payable balance changes correctly
3. trial balance remains balanced

### Phase 5: Minimal Inventory Workflow

Goal: add the first non-accounting-only business workflow after accounting integration is proven.

Deliver:

1. items
2. locations
3. inventory movement document
4. stock summary report
5. AI inventory request classification

Exit criteria:

1. natural-language stock adjustment or receipt creates a backend inventory movement
2. stock report changes correctly
3. inventory workflow does not create accounting records unless a defined backend handoff exists

### Phase 6: Web UI

Goal: add browser UX after REPL workflows prove the API and worker behavior.

Start with AI app UI, not backend UI.

Initial screens:

1. chat/request entry
2. request list
3. run detail
4. proposal detail
5. backend document link
6. failed/requeue view

Backend UI can remain minimal or absent. The backend is a simulator and can be operated through seed commands, REST calls, and the backend REPL during development.

## 9. Phase 1 Acceptance Test

Create an automated end-to-end test script or test harness for Phase 1.

Test flow:

```text
1. start backend
2. start AI app
3. seed backend:
   - ABC Traders vendor
   - Office Supplies expense account
   - Input GST account
   - Accounts Payable account
   - GST 18 percent tax code
   - open accounting period
4. submit AI message:
   "Book vendor invoice from ABC Traders for office supplies Rs 10,000 plus 18% GST dated 15 Apr 2026."
5. process AI queue
6. assert AI proposal exists
7. assert backend vendor invoice exists
8. submit, approve, and post backend document
9. assert journal entry lines:
   - debit expense 10000
   - debit input GST 1800
   - credit accounts payable 11800
10. assert trial balance total debit equals total credit
11. assert traceability from AI request to backend document and journal entry
```

This test is the gate that prevents another long build from failing to prove the core workflow.

## 10. REPL Requirements

Both applications should have REPL interfaces, but REPLs must call REST APIs rather than internal services.

Backend REPL examples:

```text
/accounts
/parties
/tax-codes
/trial-balance
/vendor-invoices
/vendor-invoice <id>
/post-vendor-invoice <id>
```

AI REPL examples:

```text
/message Book vendor invoice from ABC Traders for office supplies Rs 10,000 plus 18% GST dated 15 Apr 2026
/process-next
/requests
/request <id>
/runs
/run <id>
/proposal <id>
/requeue <request-id>
```

The AI REPL is the first operator interface for the important product. The backend REPL is a development tool for inspecting and controlling the simulator.

## 11. Backend Capability Guardrails

Because the backend is only a simulator, avoid building capabilities that real traditional systems would not normally own.

Do not add:

1. natural-language workflows
2. model-backed classification
3. AI-generated recommendations
4. agent-specific approval semantics
5. prompt-aware business APIs
6. autonomous business workflow execution
7. chat-native backend concepts

Do add traditional system capabilities only when they support AI-agent testing:

1. vendor invoice
2. customer invoice
3. payment and receipt
4. journal voucher
5. inventory item and movement
6. purchase order and sales order
7. basic GST and TDS support
8. reports needed by agents for validation and feedback

## 12. AI Application Guardrails

The AI app should not become a hidden ERP.

Do not store final business truth in the AI app:

1. no final ledger balances
2. no final inventory stock
3. no final document lifecycle state independent of backend
4. no backend-master-data duplicates except cached read models with clear refresh behavior

The AI app may store:

1. user messages
2. system-originated messages
3. AI requests
4. run steps
5. model inputs and outputs
6. tool calls
7. extracted proposal data
8. backend document IDs
9. backend API call records
10. errors and feedback

## 13. Recommended First Milestone Name

Use a workflow outcome as the milestone name:

```text
Phase 1: Vendor Invoice To Posted Journal Entry
```

Avoid foundation-only milestone names such as:

```text
Phase 1: Platform Foundation
```

The foundation should be built only to the depth needed to prove the first workflow.

## 14. Completion Standard

Do not treat any phase as complete based only on code structure, package tests, or a working prompt.

Every phase must close with:

1. a working end-to-end workflow
2. database evidence
3. API evidence
4. automated or scripted verification
5. documented known limitations

For Phase 1, the required evidence is a posted balanced journal entry created from a natural-language vendor invoice request through the split AI app plus backend app architecture.
