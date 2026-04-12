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
8. documentation updates to keep the review plans, active tracker, workflow docs, and user guides aligned as capabilities land

Out of scope by default:

1. weakening the persist-first, database-truth, human-approval operating model
2. moving workflow/business truth into Svelte-only code
3. novelty-driven agent autonomy that bypasses approvals or auditability
4. broad module-path or deployment-policy changes without a confirmed repository and deployment target
5. implementing proactive AI before failed-run recovery and specialist execution are reliable

## 4. Delivery Slices

### 4.1 Slice 1: App structural cleanup

Implement the low-risk app cleanup items that remain after Milestone 15:

1. P2-3: remove or clarify duplicate/test-only API constructors after updating tests that use the served aliases
2. P2-4: replace private `...any` optional service injection with explicit typed parameters
3. P2-5: remove duplicate `adminPartyContactsPath`
4. P3-2: compute embedded web `fs.Sub` once at startup
5. P3-3: replace dot-based static asset detection with an extension allowlist
6. P3-4: parse bearer token once in logout
7. P3-5: decide whether to keep strict single-JSON-value decoding or simplify it; do not treat it as a validated whitespace defect
8. P4-1: fix implicit JSON `Content-Type` detection

### 4.2 Slice 2: Read-path and session write-hotspot hardening

Implement:

1. P2-6 remainder: apply `AuthorizeReadOnlyTx` to reporting and accounting read-only paths where the operation truly does not mutate state
2. P3-7: rate-limit `last_seen_at` writes or otherwise reduce per-request session write contention
3. focused concurrency and org-boundary tests around the changed read paths

### 4.3 Slice 3: File decomposition refactors

Implement as same-package mechanical refactors with focused verification after each package:

1. P2-1: split `internal/app/api.go` by constants, interfaces, types, constructors, helpers, mappers, and errors
2. P2-2: split `internal/accounting/service.go` by accounting ownership area
3. P4-3: split `web/src/lib/api/types.ts` into domain-scoped type modules with a barrel export
4. AI Improvement A: split `internal/ai/openai_provider.go` by provider, tools, prompt, schema, parse, and repair concerns

### 4.4 Slice 4: Attachment storage seam

Implement:

1. P3-6: `AttachmentStore` abstraction in `internal/attachments`
2. PostgreSQL-backed implementation preserving current behavior
3. tests proving create, read, and download behavior remains unchanged
4. documentation update confirming the clean path to external object storage

### 4.5 Slice 5: AI prompt and context architecture

Implement:

1. Gap 3: `RequestContextLoader` interface and `DatabaseRequestContextLoader`
2. Gap 6: `PromptConfig`, default prompt constants, instruction hash/version recording, and optional configured prompt loading only if operationally justified
3. Improvement C: registry-driven specialist capability allowlist and dynamic schema enum
4. Improvement D: documented stopword maintenance rule and focused tests
5. Improvement H: derived-text and message prompt-size bounds with explicit truncation markers

### 4.6 Slice 6: Real specialist execution

Implement:

1. Gap 1: `SpecialistProvider` interface and coordinator registry
2. `inbound_request.operations_triage` specialist execution
3. `inbound_request.approval_triage` specialist execution
4. fallback behavior that is explicit and auditable when no specialist provider is registered or a specialist fails
5. verification that child specialist artifacts differ from coordinator artifacts when a specialist actually runs

### 4.7 Slice 7: Operator recovery and feedback loop

Implement:

1. Improvement I: failed-run visibility and admin requeue endpoint for failed or transient-failed requests
2. transient OpenAI retry if it was not already pulled into Milestone 15
3. Improvement J: coordinator brief feedback table, backend handler, and minimal Svelte gesture
4. reporting or admin review surface enough to correlate feedback with instruction version

### 4.8 Slice 8: Queue triggering and AI tool breadth

Implement only after the coordinator is stable:

1. Architectural Gap 1: PostgreSQL `LISTEN/NOTIFY` queue wakeup with polling fallback
2. Improvement K: first additional read tools, starting with inventory item levels, budget position, approval authority, and vendor summary where backend seams are ready
3. policy entries and tests for each new tool

### 4.9 Slice 9: Production policy and observability cleanup

Implement after deployment requirements are clear:

1. P4-2: global 401 handling and refresh behavior appropriate to the active auth mode
2. P5-2: explicit same-origin/CORS policy middleware if deployment topology requires it
3. P5-3: improved Svelte error boundary and structured client error logging once observability tooling is chosen
4. P5-1: Go module path correction only after a canonical VCS module path is confirmed

### 4.10 Slice 10: Proactive AI capability

Implement Architectural Gap 2 only after the earlier AI slices are stable:

1. scheduled/proactive run data model and processor path
2. operator alert/review surface for runs that are not tied to an inbound request
3. explicit safeguards preserving approval gates, auditability, and database truth

## 5. Verification

Milestone 16 verification should scale by slice:

1. mechanical same-package file splits require `go build`, package tests, gopls diagnostics, and `git diff --check`
2. accounting/reporting read-path changes require focused org-boundary and report correctness tests plus serialized canonical Go verification
3. AI provider/context/prompt/specialist work requires focused `internal/ai` tests, coordinator integration tests, and live or mocked provider validation as appropriate
4. frontend type and error-boundary changes require `npm --prefix web run check`, `npm --prefix web run build`, and focused Svelte tests
5. workflow-visible behavior changes require `docs/workflows/` updates before downstream user-guide updates

## 6. Completion Rule

Milestone 16 is complete when the non-urgent review findings have either been implemented, explicitly deferred with rationale, or promoted into a later focused plan. Do not mark the milestone complete while review-plan items remain ambiguous between “not started” and “consciously deferred.”
