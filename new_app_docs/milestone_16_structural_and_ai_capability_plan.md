# workflow_app Milestone 16 Structural and AI Capability Plan

Date: 2026-04-12
Status: Planned follow-on milestone after Milestone 15
Purpose: define the non-urgent structural, recommended, and capability-expansion work from the April 2026 review plans after urgent defects and current-runtime hardening are closed.

## 1. Source Review Documents

Milestone 16 is a sequencing and implementation overlay on top of the original validated review documents. It does not replace them.

Use these original documents as the detailed issue source of truth while implementing Milestone 16:

1. `new_app_docs/code_review_and_improvement_plan.md`
2. `new_app_docs/ai_layer_improvement_plan.md`

When a Milestone 16 item references a priority, gap, or improvement label such as P2-1, P3-7, Gap 1, or Improvement K, interpret that label by reading the matching section in the original review document first. Keep any implementation status updates synchronized between this milestone plan and the relevant original review document so the two tracks do not drift.

Milestone 16 also incorporates the 2026-04-12 comparison against `https://github.com/signalmachine/accounting-agent-app`. That comparison exposed one additional workflow-critical capability gap that is not explicit enough in the original review documents: `workflow_app` can currently turn an accounting-style inbound request into a generic operator-review proposal, but it does not yet turn that request into a structured, approval-ready accounting document or journal proposal.

## 2. Milestone Position

Milestone 16 should start only after Milestone 15 has closed. Its job is to reduce structural drag, make the AI layer easier to evolve, and implement the larger capability improvements that are too broad for an urgent corrective milestone.

This milestone should remain bounded. It should not become a dumping ground for every possible refactor; each slice should be promoted when it materially improves correctness, operability, maintainability, or AI output quality.

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
9. documentation updates to keep the review plans, active tracker, workflow docs, and user guides aligned as capabilities land

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

Milestone 16 must close that integration gap without copying the smaller app's weaker shortcut of committing ledger truth immediately after chat confirmation. The `workflow_app` target remains:

```text
inbound request -> AI structured proposal -> backend-owned document draft/submission -> human approval -> accounting posting -> journal entry
```

AI may prepare or propose. Humans approve. The accounting package posts.

## 5. Delivery Slices

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
2. keep AI output advisory and structured
3. keep document creation, approval, and posting under deterministic Go services

Implement:

1. add an accounting-intent classification path for inbound requests, covering at least:
   - manual journal or expense entry requests
   - purchase/vendor invoice requests
   - customer invoice or revenue requests when supported by current document seams
   - payment or receipt requests when supported by current document seams
2. add a strict structured AI output schema for accounting proposals, separate from the generic coordinator brief schema
3. include fields needed for deterministic validation, such as:
   - intent type
   - document type code
   - title and summary
   - effective or posting date
   - currency
   - tax scope and tax-code suggestion when applicable
   - proposed debit and credit lines when a journal-style proposal is appropriate
   - party, contact, or counterparty references when an invoice or payment document is appropriate
   - confidence, rationale, and missing-data questions
4. validate structured accounting output before persistence:
   - account codes must exist and belong to the actor's org
   - debit and credit totals must balance for journal-style proposals
   - tax codes must exist when used
   - document type must be supported by the backend document/accounting service
   - missing required fields must produce a reviewable missing-data result rather than a half-created document
5. create or prepare the correct backend-owned downstream record when validation succeeds:
   - for invoice/payment-style flows, create the appropriate document payload through the accounting/document services when the current backend seam supports the required structure
   - for journal-style flows, create a reviewable accounting document or structured proposal record rather than directly posting a journal entry
   - link the `ai.agent_recommendations` row to the resulting document so `RequestProcessedProposalApproval` can proceed
6. expose missing-data states in the proposal UI so operators can amend the request or fill required fields instead of silently receiving a generic brief
7. ensure the exact request, proposal, approval, document, and accounting detail routes preserve the same `REQ-...` continuity chain

Guardrails:

1. do not let AI call `PostDocument` or create `accounting.journal_entries` directly
2. do not bypass approval when the proposal affects accounting truth
3. do not store accounting truth only in an AI artifact payload
4. do not collapse the request record into the accounting document; keep `REQ-...` as request tracking only
5. prefer typed Go DTOs and service methods over ad hoc JSON maps for accounting proposal persistence

Acceptance evidence:

1. a request like `REQ-000002` for office supplies plus GST no longer stops at a generic operator-review recommendation
2. when enough structured data exists, the processed proposal has a linked `document_id`
3. requesting approval for that proposal creates a workflow approval
4. approval changes the linked document state through the existing workflow seam
5. posting, when explicitly performed through the accounting seam, creates a journal entry linked back to the same document, recommendation, run, and request
6. when the request lacks required accounting data, the proposal clearly lists the missing data and does not create an invalid document

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
4. AI provider/context/prompt/specialist work requires focused `internal/ai` tests, coordinator integration tests, and live or mocked provider validation as appropriate
5. frontend type and error-boundary changes require `npm --prefix web run check`, `npm --prefix web run build`, and focused Svelte tests
6. workflow-visible behavior changes require `docs/workflows/` updates before downstream user-guide updates
7. at least one live or seeded continuity proof should cover:
   - submit an accounting-style inbound request
   - process it
   - confirm the processed proposal has a linked document when data is sufficient
   - request approval
   - approve
   - post through the accounting seam
   - open the exact accounting entry and trace it back to the same request and recommendation

## 7. Open Implementation Questions

Resolve these before implementing Slice 5.6:

1. Should the first accounting proposal path create a generic `journal` document, an `invoice` document, or a new explicit accounting-proposal record that later materializes into a document?
2. Which document types are supported enough today for AI-prepared documents without adding broad manual-entry UI?
3. For office-supplies-plus-GST requests, should the first implementation prefer a vendor invoice flow, a payment/receipt flow, or a journal-style expense accrual when the vendor/payment state is unspecified?
4. Which fields should be required before creating a document, and which fields should remain operator-editable before approval?
5. Should accounting-intent detection run inside the general coordinator, a registered accounting specialist, or a dedicated provider method that the coordinator invokes after classification?

## 8. Completion Rule

Milestone 16 is complete when the non-urgent review findings have either been implemented, explicitly deferred with rationale, or promoted into a later focused plan. Do not mark the milestone complete while review-plan items remain ambiguous between “not started” and “consciously deferred.”
