# workflow_app Milestone 15 Urgent Review and AI Hardening Plan

Date: 2026-04-12
Status: Planned next implementation milestone
Purpose: define the urgent corrective milestone after the April 2026 application-wide and AI-layer reviews, separating defects and current-runtime hardening from larger structural and capability work.

## 1. Validation Summary

The two review documents are directionally valid and should drive the next work:

1. `code_review_and_improvement_plan.md` correctly identifies urgent production-readiness defects in API auth status handling, wrong method responses, session error leakage, the missing inventory navigation endpoint, and unbounded attachment request bodies.
2. `ai_layer_improvement_plan.md` correctly identifies current-runtime AI hardening needs in tool-policy authorization, timeout budgeting, tool failure visibility, provider identity, output token budget, and read-tool budget accounting.
3. the high-severity specialist delegation finding is valid, but the full real-specialist execution path is larger than an urgent defect slice; Milestone 15 should remove misleading specialist attribution or mark fallback explicitly, while Milestone 16 should implement full specialist execution.
4. the `AuthorizeReadOnlyTx` recommendation is shared by both plans and should be implemented once, then used by the AI tool-policy path and the read-heavy accounting/reporting paths as scope permits.
5. P3-5 in `code_review_and_improvement_plan.md` is overstated: the current decoder does not reject trailing whitespace because the second decode returns `io.EOF` after whitespace. Treat the recommendation as a non-urgent API strictness decision, not as a validated defect.

## 2. Milestone Position

Milestone 15 should happen before broad production use and before larger Milestone 16 refactors. It is justified because the validated issues affect:

1. authentication semantics and Svelte 401 handling
2. API contract consistency for review detail routes
3. internal error exposure
4. inventory hub runtime completeness
5. attachment upload resource bounds
6. current coordinator reliability and operator visibility
7. misleading specialist-run attribution if delegation is recorded without a real specialist provider

## 3. Scope

In scope:

1. implement P1-1 through P1-4 from `code_review_and_improvement_plan.md`
2. implement P3-1 attachment request-size bounds
3. implement P4-4 only as the frontend/API-client half required by P1-4
4. implement shared read-only authorization support needed by AI Gap 2 and app P2-6
5. implement AI Phase 1 current-runtime hardening: Gap 2, Gap 4, Gap 5, Improvement B, Improvement F, and Improvement G
6. add a bounded specialist-truth correction so specialist fallback is explicit and no operator-visible artifact implies a real specialist executed when it did not
7. add focused regression coverage for each defect family and run the documented verification commands
8. update the active review plans and tracker status as items land

Out of scope:

1. full `api.go`, `accounting/service.go`, `types.ts`, or `openai_provider.go` decomposition
2. real specialist provider execution and new specialist prompts
3. prompt externalization, prompt versioning, or prompt file loading
4. new AI tool-suite breadth
5. proactive or scheduled AI runs
6. broad CORS, telemetry, module-path, or frontend error-boundary policy work
7. changing the JSON decoder strictness policy unless a concrete client failure is reproduced

## 4. Delivery Slices

### 4.1 Slice 1: API contract and auth defect pass

Implement:

1. P1-1: shared `writeActorError` and replacement of review, inbound, and approval handler `400` auth failures with sanitized `401`
2. P1-2: `405` JSON responses for wrong methods on GET-only review detail handlers
3. P1-3: sanitized `500` for unexpected `GET /api/session` failures, preserving `401` for unauthorized
4. focused tests proving unauthenticated review/inbound/approval requests return `401`, detail wrong-method calls return `405`, and session non-auth failures do not leak raw errors

### 4.2 Slice 2: Inventory landing runtime completion

Implement:

1. P1-4: backend `/api/navigation/inventory` endpoint using `GetInventoryLandingSnapshot`
2. P4-4: `InventoryLandingSnapshot` frontend type, API client function, and Svelte inventory hub data loading
3. focused backend and frontend tests proving the route is registered, org-scoped, and consumed by the promoted inventory hub

### 4.3 Slice 3: Attachment resource bound

Implement:

1. P3-1: `http.MaxBytesReader` for request bodies that accept base64 attachments
2. `http.MaxBytesError` handling in `writeJSONBodyError`
3. focused tests proving oversized attachment payloads return `413` without attempting unbounded decode

### 4.4 Slice 4: AI runtime hardening

Implement:

1. `identityaccess.AuthorizeReadOnlyTx` and `authorizeReadTx` in the AI service
2. AI Gap 2: read-only transaction for `ResolveToolPolicy`
3. AI Gap 4: per-call OpenAI timeouts rather than one exhausted outer timeout
4. AI Gap 5: `DegradedMode` and `DegradedReasons` in provider output, coordinator step payload, and artifact payload when tools are blocked or fail
5. Improvement B: provider `Name()` on `CoordinatorProvider`, removing the concrete-type switch
6. Improvement F: named output-token limits raised from `900` to the planned primary and repair budgets
7. Improvement G: unique read-tool counting in `coordinatorReadToolBudgetExhausted`
8. focused AI tests for policy resolution, timeout behavior through mocks, degraded-mode payloads, provider naming, output-token constants, and repeated-tool budget counting

### 4.5 Slice 5: Specialist truth correction

Implement one bounded correction:

1. either disable specialist delegation until a real registered specialist provider exists, or
2. keep the fallback path but record and expose `used_fallback: true` and a clear coordinator-output fallback label in child-run step and artifact metadata

The goal is not to implement real specialists in Milestone 15. The goal is to prevent the system from implying that a domain specialist performed analysis when the coordinator output was reused.

## 5. Verification

Use `docs/technical_guides/07_testing_and_verification.md` for exact command shape. Expected closeout evidence:

1. focused `internal/app` tests for API contract defects
2. focused `internal/ai` tests for AI hardening
3. affected frontend tests for inventory hub loading
4. `npm --prefix web run check` and `npm --prefix web run build` when frontend files change
5. `go build ./cmd/... ./internal/...`
6. serialized canonical Go verification when shared auth, app API, reporting, or AI coordinator seams change
7. gopls diagnostics on edited Go files
8. `git diff --check`

## 6. Completion Rule

Milestone 15 is complete only when urgent API defects, attachment bounds, inventory landing continuity, AI Phase 1 hardening, and specialist-truth correction are implemented and verified. Any remaining structural or capability recommendations move to Milestone 16 unless implementation uncovers a new blocking defect.
