# Document Numbering Global Uniqueness Plan (2026-03-05)

Status: completed and archived.

## 1) Objective

Adopt a single numbering policy across the application:

- Document numbers are globally unique per `(company_id, type_code)`.
- Numbering does not reset at financial year end.
- `financial_year` remains reporting metadata only, not sequence scope.

## 2) Implementation Review (Completed)

Implementation was completed in one controlled change set (migration + code + tests + ops checks).

### 2.1 Database Changes

Completed via migration:

- `migrations/036_document_numbering_global_uniqueness.sql`

What was implemented:

1. Normalized go-live document types to global numbering policy:
   - `numbering_strategy = 'global'`
   - `resets_every_fy = false`
   - Applied to: `JE`, `SI`, `PI`, `SO`, `PO`, `GR`, `GI`, `RC`, `PV` (if present).
2. Replaced sequence uniqueness to global scope only:
   - `document_sequences_unique_idx` on `(company_id, type_code)`.
3. Replaced document-number uniqueness to global scope only:
   - `documents_unique_number_idx` on `(company_id, type_code, document_number)`
   - partial index: `WHERE document_number IS NOT NULL`.
4. Added migration guard for duplicate `(company_id, type_code, document_number)` before index rebuild.
5. Consolidated legacy `document_sequences` rows to one row per `(company_id, type_code)` using `MAX(last_number)`.

### 2.2 Application Changes

Completed in runtime logic:

- `internal/core/document_service.go`
  - Posting sequence now always uses global scope `(company_id, type_code)`.
  - Effective sequence FY/branch are always `nil`.
  - Added guardrail to reject non-global document type configuration at runtime.
- `internal/core/ledger.go`
  - Removed FY-scoped numbering derivation.
  - Enforces global numbering policy before draft/post flow.

### 2.3 Test and Seed Changes

Completed:

- Updated seed flow:
  - `cmd/restore-seed/main.go` (`SI`/`PI` now `global + resets_every_fy=false`).
- Updated integration tests:
  - `internal/core/document_integration_test.go`
  - `internal/core/ledger_integration_test.go`
  - `internal/core/order_integration_test.go`
  - `internal/core/purchase_order_integration_test.go`
- Added/validated coverage for:
  - cross-year sequence continuation (no reset),
  - concurrent posting uniqueness,
  - policy-drift rejection for non-global document type config,
  - test schema precondition checks for new index shape.

### 2.4 Documentation and Ops Updates

Completed:

- `docs/deployment/deployment.md` updated with go-live numbering verification gate.
- `cmd/verify-db-health/main.go` updated with blocking checks for:
  - go-live document types configured as `global + no FY reset`,
  - `document_sequences_unique_idx` shape,
  - `documents_unique_number_idx` shape.
- `AGENTS.md` updated with mandatory completion rule:
  - schema-change work is not complete until migration + health checks succeed on target live/dev `DATABASE_URL`.

## 3) Environment Execution Record

Execution date: March 5, 2026.

### 3.1 Test DB Validation

Executed:

1. `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db`
2. `go test -p 1 ./...`
3. `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db-health`

Result: passed after applying migration and using clean test seed state.

### 3.2 Live/Dev Cloud DB Validation (`DATABASE_URL`)

Executed:

1. `go run ./cmd/verify-db`
   - Result: migration `036_document_numbering_global_uniqueness.sql` applied.
2. `go run ./cmd/verify-db-health`
   - Result: passed, including new numbering/index checks.

## 4) Acceptance Criteria Review

All target criteria are satisfied:

1. Cross-year sequence continuation: validated by integration tests.
2. No duplicate posted document numbers per `(company_id, type_code)`: enforced by new index.
3. Open-item uniqueness semantics aligned to `(company_id, type_code, document_number)`.
4. Migration, health checks, and full serial test run executed successfully.
5. Seed/restore and runtime behavior now align with no FY-reset policy.

## 5) Notes

- `financial_year` and `branch_id` are retained for compatibility/reporting metadata and are not used for numbering scope.
- Future branch/FY-specific legal exceptions must be explicitly isolated and should not alter the default global numbering policy.
