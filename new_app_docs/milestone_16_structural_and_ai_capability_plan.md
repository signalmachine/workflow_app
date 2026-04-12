# workflow_app Milestone 16 Structural and AI Capability Plan

Date: 2026-04-12
Status: Planned follow-on milestone after Milestone 15
Purpose: define the full AI-layer overhaul and workflow-execution milestone after urgent defects and current-runtime hardening are closed.

## 1. Source Review Documents

Milestone 16 is a sequencing and implementation overlay on top of the original validated review documents. It does not replace them.

Use these original documents as the detailed issue source of truth while implementing Milestone 16:

1. `new_app_docs/code_review_and_improvement_plan.md`
2. `new_app_docs/ai_layer_improvement_plan.md`

When a Milestone 16 item references a priority, gap, or improvement label such as P2-1, P3-7, Gap 1, or Improvement K, interpret that label by reading the matching section in the original review document first. Keep any implementation status updates synchronized between this milestone plan and the relevant original review document so the two tracks do not drift.

Milestone 16 also incorporates the 2026-04-12 comparison against `https://github.com/signalmachine/accounting-agent-app`. That comparison exposed one additional workflow-critical capability gap that is not explicit enough in the original review documents: `workflow_app` can currently turn an accounting-style inbound request into a generic operator-review proposal, but it does not yet turn that request into a structured, approval-ready accounting document or journal proposal.

## 2. Milestone Position

Milestone 16 should start only after Milestone 15 has closed. Its job is to overhaul the AI layer enough to support reliable end-to-end workflow execution, not merely to complete a fixed checklist of non-urgent refactors.

The planned slices below are the starting structure, not the delivery ceiling. During implementation, code review and workflow testing may expose additional gaps in the AI layer, backend workflow seams, document lifecycle, approval lifecycle, posting path, observability, or operator recovery flow. Add those activities to Milestone 16 when they materially block or weaken end-to-end workflow execution.

This milestone should still avoid novelty-driven or unrelated refactors. New work belongs in Milestone 16 when it is necessary to make the foundational workflow pass from request submission through AI processing, proposal/document creation, approval, accounting posting, and final accounting entry persistence in the database.

The first Milestone 16 implementation checkpoint should target the foundational accounting workflow directly. Do not make that checkpoint wait behind every structural cleanup slice. Implement only the enabling cleanup, context-loading, prompt/schema, or specialist-registry work needed to deliver and verify the request-to-accounting-entry path first; then continue with the remaining structural and capability slices. For non-accounting business events, the first checkpoint only needs classification and safe no-accounting behavior, not persistence into new non-accounting tables.

## 3. Scope

In scope:

1. remaining P2/P3/P4/P5 items from `code_review_and_improvement_plan.md` that were not pulled into Milestone 15
2. AI Gap 1 full specialist execution
3. AI Gap 3 request-context loader extraction
4. AI Gap 6 prompt configuration and instruction-version tracking
5. AI Improvements A, C, D, H, I, J, and K
6. Architectural Gap 1 PostgreSQL `LISTEN/NOTIFY` queue triggering
7. Architectural Gap 2 scheduled/proactive AI runs only after the coordinator and operator-recovery slices are stable
8. accounting-intent workflow generation from inbound requests: detect accounting-entry, invoice, payment, or receipt intent; produce structured backend-owned document or journal proposal data; link the AI recommendation to the resulting document when enough information exists; and expose a missing-data state when it does not
9. accounting-impact triage for business events: determine whether the request is an accounting event, a supported non-accounting event, or an unsupported event; return a clear no-accounting-impact comment without creating an accounting proposal when it is not an accounting event; and decompose requests that contain multiple accounting events into separate proposal/document candidates
10. documentation updates to keep the review plans, active tracker, workflow docs, and user guides aligned as capabilities land
11. additional AI-layer, workflow, approval, document, accounting, recovery, and observability work discovered during implementation when it is required to make the foundational end-to-end workflow execute successfully

Out of scope by default:

1. weakening the persist-first, database-truth, human-approval operating model
2. moving workflow/business truth into Svelte-only code
3. novelty-driven agent autonomy that bypasses approvals or auditability
4. broad module-path or deployment-policy changes without a confirmed repository and deployment target
5. implementing proactive AI before failed-run recovery and specialist execution are reliable
6. allowing AI to directly post journal entries without explicit human approval
7. treating `REQ-...` request identifiers as accounting document or journal-entry numbers

## 4. Accounting-Agent Comparison Finding

The smaller `accounting-agent-app` succeeds at the accounting-entry scenario because it has a direct financial-event path:

1. route natural-language accounting events to a dedicated journal-entry interpreter
2. require strict structured output containing document type, company, currency, exchange rate, posting date, document date, summary, confidence, reasoning, and balanced debit/credit lines
3. store the proposed write as a pending action
4. require explicit user confirmation before committing to the ledger
5. commit through the deterministic application/ledger service, not through the model

`workflow_app` currently takes a different path:

1. the queued request is processed by the general coordinator
2. the coordinator response schema produces only an operator-review brief: summary, priority, artifact title/body, rationale, next actions, and optional specialist delegation
3. the persisted recommendation has `recommendation_type = operator_review`
4. the recommendation is not linked to a `documents.documents` row
5. the proposal-approval path correctly rejects documentless proposals because it requires a linked `document_id`

The result is visible in the `REQ-000002` smoke test: the AI correctly understood the accounting intent, but only recommended that an operator create an entry. It did not create an accounting document, approval-ready proposal, or journal proposal.

Milestone 16 should also make the AI layer explicitly classify business events before proposing writes:

1. if the event has no accounting impact and is not a supported non-accounting event, the agent should return a clear operator comment that no accounting proposal is needed and no supported persistence path exists yet
2. if the event has accounting impact, the agent should determine whether it represents one accounting event or multiple separate accounting events
3. if a single request includes multiple invoices or otherwise distinct accounting events, the system should prepare separate proposal/document candidates rather than collapsing them into one journal entry
4. if the event is a supported non-accounting event in the future, the agent may propose updates through the specifically instructed non-accounting workflow and backend service
5. future support for storing non-accounting business events should use a database-owned event/workflow model outside the accounting tables, not placeholder rows in accounting truth tables
6. unsupported non-accounting events must not invent persistence paths; they should produce a comment or missing-capability result until prompts, event types, services, and tables are explicitly defined

Milestone 16 must close that integration gap without copying the smaller app's weaker shortcut of committing ledger truth immediately after chat confirmation. The `workflow_app` target remains:

```text
inbound request -> AI structured proposal -> backend-owned document draft/submission -> human approval -> accounting posting -> journal entry
```

AI may prepare or propose. Humans approve. The accounting package posts.

## 5. Delivery Slices

These slices are planning anchors. They should be expanded during implementation when codebase review or workflow testing exposes a gap that blocks the foundational workflow from completing correctly. Do not treat the slice list as a reason to defer a required AI-layer or workflow-seam fix that is discovered while implementing Milestone 16.

Recommended execution order:

1. start with the minimum enabling subset of Slice 5.5 and Slice 5.7 only where needed for a clean accounting-proposal path
2. implement Slice 5.6 as the first externally meaningful Milestone 16 checkpoint
3. run the live or seeded continuity proof through final accounting-entry persistence
4. implement Slice 5.8 recovery if the first workflow pass exposes stuck-run or requeue gaps
5. continue through the remaining cleanup, refactor, storage, queue, tool-breadth, policy, and proactive-AI slices
6. defer persistence for supported non-accounting event types until after the accounting path is proven, unless one such event type is explicitly promoted with its own prompt, service, table, and workflow acceptance criteria

This order preserves the original plan while making the end-to-end workflow the milestone's controlling priority.

### 5.1 Slice 1: App structural cleanup

Implement the low-risk app cleanup items that remain after Milestone 15:

1. P2-3: remove or clarify duplicate/test-only API constructors after updating tests that use the served aliases
2. P2-4: replace private `...any` optional service injection with explicit typed parameters
3. P2-5: remove duplicate `adminPartyContactsPath`
4. P3-2: compute embedded web `fs.Sub` once at startup
5. P3-3: replace dot-based static asset detection with an extension allowlist
6. P3-4: parse bearer token once in logout
7. P3-5: decide whether to keep strict single-JSON-value decoding or simplify it; do not treat it as a validated whitespace defect
8. P4-1: fix implicit JSON `Content-Type` detection

### 5.2 Slice 2: Read-path and session write-hotspot hardening

Implement:

1. P2-6 remainder: apply `AuthorizeReadOnlyTx` to reporting and accounting read-only paths where the operation truly does not mutate state
2. P3-7: rate-limit `last_seen_at` writes or otherwise reduce per-request session write contention
3. focused concurrency and org-boundary tests around the changed read paths

### 5.3 Slice 3: File decomposition refactors

Implement as same-package mechanical refactors with focused verification after each package:

1. P2-1: split `internal/app/api.go` by constants, interfaces, types, constructors, helpers, mappers, and errors
2. P2-2: split `internal/accounting/service.go` by accounting ownership area
3. P4-3: split `web/src/lib/api/types.ts` into domain-scoped type modules with a barrel export
4. AI Improvement A: split `internal/ai/openai_provider.go` by provider, tools, prompt, schema, parse, and repair concerns

### 5.4 Slice 4: Attachment storage seam

Implement:

1. P3-6: `AttachmentStore` abstraction in `internal/attachments`
2. PostgreSQL-backed implementation preserving current behavior
3. tests proving create, read, and download behavior remains unchanged
4. documentation update confirming the clean path to external object storage

### 5.5 Slice 5: AI prompt and context architecture

Implement:

1. Gap 3: `RequestContextLoader` interface and `DatabaseRequestContextLoader`
2. Gap 6: `PromptConfig`, default prompt constants, instruction hash/version recording, and optional configured prompt loading only if operationally justified
3. Improvement C: registry-driven specialist capability allowlist and dynamic schema enum
4. Improvement D: documented stopword maintenance rule and focused tests
5. Improvement H: derived-text and message prompt-size bounds with explicit truncation markers

### 5.6 Slice 6: Accounting proposal generation from inbound requests

Implement the missing accounting workflow capability exposed by the `REQ-000002` smoke test and by comparison with `accounting-agent-app`.

Goal:

1. convert sufficiently specific accounting-style inbound requests into structured backend-owned proposals that can move into approval and posting
2. distinguish accounting-impacting events, supported non-accounting events, and unsupported events
3. decompose requests with multiple accounting events into separate proposal/document candidates
4. keep AI output advisory and structured
5. keep document creation, approval, and posting under deterministic Go services

Implement:

1. add an accounting-impact and accounting-intent classification path for inbound requests, covering at least:
   - no accounting impact and no supported persistence path
   - supported non-accounting business event, reserved for future explicitly configured event types and not required for the first checkpoint
   - manual journal or expense entry requests
   - purchase/vendor invoice requests
   - customer invoice or revenue requests when supported by current document seams
   - payment or receipt requests when supported by current document seams
2. add a strict structured AI output schema for accounting triage and proposals, separate from the generic coordinator brief schema
3. include fields needed for deterministic validation, such as:
   - accounting impact classification
   - intent type
   - event count and event boundaries
   - document type code
   - title and summary
   - effective or posting date
   - currency
   - tax scope and tax-code suggestion when applicable
   - proposed debit and credit lines when a journal-style proposal is appropriate
   - party, contact, or counterparty references when an invoice or payment document is appropriate
   - confidence, rationale, and missing-data questions
4. return a reviewable no-accounting-impact comment when the request has no accounting impact and no supported non-accounting persistence path, without creating an accounting document, journal proposal, or accounting approval
5. split multiple accounting events into separate proposal/document candidates:
   - two or more invoices should normally become two or more invoice proposals or accounting documents
   - unrelated payments and invoices should not be collapsed into one journal entry unless the backend accounting service explicitly models that combined event
   - ambiguous grouping must produce a missing-data or operator-review result rather than guessing
6. validate structured accounting output before persistence:
   - account codes must exist and belong to the actor's org
   - debit and credit totals must balance for journal-style proposals
   - tax codes must exist when used
   - document type must be supported by the backend document/accounting service
   - missing required fields must produce a reviewable missing-data result rather than a half-created document
7. create or prepare the correct backend-owned downstream record when validation succeeds:
   - for invoice/payment-style flows, create the appropriate document payload through the accounting/document services when the current backend seam supports the required structure
   - for journal-style flows, create a reviewable accounting document or structured proposal record rather than directly posting a journal entry
   - link the `ai.agent_recommendations` row to the resulting document so `RequestProcessedProposalApproval` can proceed
8. expose missing-data, no-accounting-impact, and unsupported-event states in the proposal UI so operators can understand why no accounting proposal was created, amend the request, or fill required fields instead of silently receiving a generic brief
9. ensure the exact request, proposal, approval, document, and accounting detail routes preserve the same `REQ-...` continuity chain
10. defer true bulk upload and batch-management UX until the single-request decomposition path is proven, but design the schema and services so a future batch of vendor invoices can create one child proposal/document candidate per invoice

Guardrails:

1. do not let AI call `PostDocument` or create `accounting.journal_entries` directly
2. do not bypass approval when the proposal affects accounting truth
3. do not store accounting truth only in an AI artifact payload
4. do not collapse the request record into the accounting document; keep `REQ-...` as request tracking only
5. prefer typed Go DTOs and service methods over ad hoc JSON maps for accounting proposal persistence
6. do not create accounting-table rows for non-accounting business events; future non-accounting event persistence must live in a separate database-owned workflow/event model
7. do not collapse multiple invoices or separate accounting events into one accounting entry merely because they arrived in the same request or upload batch
8. do not record every non-accounting event by default; only explicitly supported non-accounting event types with defined prompts, backend services, and non-accounting tables may create database updates

Acceptance evidence:

1. a request like `REQ-000002` for office supplies plus GST no longer stops at a generic operator-review recommendation
2. when enough structured data exists, the processed proposal has a linked `document_id`
3. requesting approval for that proposal creates a workflow approval
4. approval changes the linked document state through the existing workflow seam
5. posting, when explicitly performed through the accounting seam, creates a journal entry linked back to the same document, recommendation, run, and request
6. when the request lacks required accounting data, the proposal clearly lists the missing data and does not create an invalid document
7. a request with no accounting impact and no supported non-accounting persistence path returns a clear no-accounting-impact or unsupported-event comment and creates no accounting document, journal proposal, or accounting approval
8. a request containing two vendor invoices produces two separate proposal/document candidates or a clear missing-data result explaining why the split cannot be made safely

### 5.7 Slice 7: Real specialist execution

Implement:

1. Gap 1: `SpecialistProvider` interface and coordinator registry
2. `inbound_request.operations_triage` specialist execution
3. `inbound_request.approval_triage` specialist execution
4. accounting specialist execution if Slice 6 needs a separate capability such as `inbound_request.accounting_triage`
5. fallback behavior that is explicit and auditable when no specialist provider is registered or a specialist fails
6. verification that child specialist artifacts differ from coordinator artifacts when a specialist actually runs

### 5.8 Slice 8: Operator recovery and feedback loop

Implement:

1. Improvement I: failed-run visibility and admin requeue endpoint for failed or transient-failed requests
2. transient OpenAI retry if it was not already pulled into Milestone 15
3. Improvement J: coordinator brief feedback table, backend handler, and minimal Svelte gesture
4. reporting or admin review surface enough to correlate feedback with instruction version

### 5.9 Slice 9: Queue triggering and AI tool breadth

Implement only after the coordinator is stable:

1. Architectural Gap 1: PostgreSQL `LISTEN/NOTIFY` queue wakeup with polling fallback
2. Improvement K: first additional read tools, starting with inventory item levels, budget position, approval authority, vendor summary, chart-of-accounts lookup, tax-code lookup, open accounting periods, and existing party/contact lookup where backend seams are ready
3. policy entries and tests for each new tool

### 5.10 Slice 10: Production policy and observability cleanup

Implement after deployment requirements are clear:

1. P4-2: global 401 handling and refresh behavior appropriate to the active auth mode
2. P5-2: explicit same-origin/CORS policy middleware if deployment topology requires it
3. P5-3: improved Svelte error boundary and structured client error logging once observability tooling is chosen
4. P5-1: Go module path correction only after a canonical VCS module path is confirmed

### 5.11 Slice 11: Proactive AI capability

Implement Architectural Gap 2 only after the earlier AI slices are stable:

1. scheduled/proactive run data model and processor path
2. operator alert/review surface for runs that are not tied to an inbound request
3. explicit safeguards preserving approval gates, auditability, and database truth

## 6. Verification

Milestone 16 verification should scale by slice:

1. mechanical same-package file splits require `go build`, package tests, gopls diagnostics, and `git diff --check`
2. accounting/reporting read-path changes require focused org-boundary and report correctness tests plus serialized canonical Go verification
3. accounting proposal generation requires focused tests for classification, schema validation, missing-data handling, document linkage, approval linkage, and posting continuity
4. no-accounting-impact triage requires focused tests proving no accounting proposal/document/approval is created
5. multi-event decomposition requires focused tests proving two distinct invoices or accounting events do not collapse into one accounting entry
6. AI provider/context/prompt/specialist work requires focused `internal/ai` tests, coordinator integration tests, and live or mocked provider validation as appropriate
7. frontend type and error-boundary changes require `npm --prefix web run check`, `npm --prefix web run build`, and focused Svelte tests
8. workflow-visible behavior changes require `docs/workflows/` updates before downstream user-guide updates
9. at least one live or seeded continuity proof should cover:
   - submit an accounting-style inbound request
   - process it
   - confirm the processed proposal has a linked document when data is sufficient
   - request approval
   - approve
   - post through the accounting seam
   - open the exact accounting entry and trace it back to the same request and recommendation

The milestone cannot close on package tests alone. The foundational workflow must be successfully tested and recorded as passed from request submission to the final accounting entry hitting the database. Any defect that prevents that path from completing belongs in Milestone 16 unless there is an explicit, documented reason that it is outside the AI/workflow foundation.

## 7. Open Implementation Questions

Resolve these before implementing Slice 5.6:

1. Should the first accounting proposal path create a generic `journal` document, an `invoice` document, or a new explicit accounting-proposal record that later materializes into a document?
2. Which document types are supported enough today for AI-prepared documents without adding broad manual-entry UI?
3. For office-supplies-plus-GST requests, should the first implementation prefer a vendor invoice flow, a payment/receipt flow, or a journal-style expense accrual when the vendor/payment state is unspecified?
4. Which fields should be required before creating a document, and which fields should remain operator-editable before approval?
5. Should accounting-intent detection run inside the general coordinator, a registered accounting specialist, or a dedicated provider method that the coordinator invokes after classification?
6. What is the first durable non-accounting business-event model, and should Milestone 16 only return no-accounting-impact comments until that model exists? The default answer for the first checkpoint is yes: classify safely first and defer non-accounting persistence.
7. How should child proposal/document candidates be numbered and displayed when one inbound request contains multiple invoices or accounting events?

## 8. Completion Rule

Milestone 16 is complete only when the foundational AI-driven accounting workflow has been implemented and tested end to end: request submission, durable queue processing, AI proposal generation, backend-owned document or accounting proposal creation, human approval, accounting posting, and the resulting accounting entry persisted in the database with traceability back to the original request.

Do not mark Milestone 16 complete merely because the initial planned slices are done. If implementation review or workflow testing exposes additional AI-layer, workflow-seam, approval, posting, persistence, recovery, or operator-continuity gaps that block this end-to-end path, add those activities to Milestone 16 and complete them before closure.

The non-urgent review findings must also be implemented, explicitly deferred with rationale, or promoted into a later focused plan. Do not mark the milestone complete while review-plan items remain ambiguous between “not started” and “consciously deferred.”
