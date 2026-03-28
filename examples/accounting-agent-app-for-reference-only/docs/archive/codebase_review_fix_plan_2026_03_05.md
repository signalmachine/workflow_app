# Codebase Review Fix Plan (2026-03-05)

## Context
A full end-to-end review of recent uncommitted changes surfaced a small set of high-value issues and inconsistencies. Core integration tests and verification tools currently pass, but several defects remain that can cause incorrect financial outputs or weaken migration/runtime consistency.

This plan defines the implementation sequence to fix those issues with minimal regression risk.

## Goals
1. Eliminate mixed-currency aggregation bugs in AP-facing operational/AI read outputs.
2. Make `account_rules.account_id` effective in runtime rule resolution (not migration-only).
3. Move schema evolution assumptions out of tests and into migrations-only validation paths.
4. Align DB tooling behavior for `DATABASE_URL` handling.
5. Clean up migration/tooling hygiene to reduce operator confusion.

## Non-Goals
1. Re-architecting AP/AR domain models.
2. Building full FX revaluation workflows.
3. Changing external API contracts beyond what is needed for correctness.

## Findings to Fix
1. Mixed currency totals in AP reporting helpers (`invoice_amount` tx currency mixed with `total_base`).
2. Rule engine always prefers `account_code`, making `account_id` effectively ignored.
3. Integration tests mutate schema with DDL in setup, masking migration defects.
4. Inconsistent DB URL policy across `verify-db` and `verify-db-health`.
5. Stray migration helper file (`migrations/apply_patch.go`) outside standard migration flow.

## Execution Strategy
Use a phased rollout. Each phase must compile and pass tests before moving to the next phase.

### Phase 1: Currency-Safe AP Operational Outputs

#### Scope
- `internal/app/app_service.go`
- Optional: adapter/tool docs if output JSON changes

#### Problems
- Outstanding invoice totals and payment history can combine values from different currencies into one field.

#### Implementation Tasks
1. **Define output contract for AP operational reads**
1. Keep ledger-based AP balance (`get_ap_balance`) untouched.
2. For PO operational outputs, return either:
   - Base-currency-only totals (`*_base` fields), or
   - Per-currency grouped totals (`totals_by_currency`), or
   - Both.
3. Add explicit currency metadata for every amount field.

2. **Refactor outstanding invoice aggregation**
1. Replace current `SUM(CASE WHEN invoice_amount ... ELSE total_base)` with one of:
   - `SUM(COALESCE(invoice_amount * po.exchange_rate, po.total_base))` as `outstanding_invoice_total_base`, and/or
   - grouped sums by `po.currency` for `invoice_amount`.
2. Return JSON with unambiguous names, e.g.:
   - `outstanding_invoice_total_base`
   - `base_currency`
   - `outstanding_by_currency: [{currency, amount_transaction}]`

3. **Refactor payment history fields**
1. Include `currency`, `exchange_rate`, and both transaction/base amounts where available.
2. Avoid fallback from transaction amount to base amount in the same field.

4. **Update tool descriptions and docs**
1. Ensure descriptions mention if values are base-currency or grouped by transaction currency.

#### Validation
1. Add/extend tests in `internal/core/purchase_order_integration_test.go` or app-level tests for:
   - multi-currency outstanding invoices
   - payment history with INR + USD rows
2. Manual check through AI tool invocation path (`go run ./cmd/verify-agent`) where applicable.

#### Acceptance Criteria
1. No response field mixes currencies implicitly.
2. Every amount field has currency semantics in name or companion metadata.

---

### Phase 2: Make `account_id` Authoritative in Rule Resolution

#### Scope
- `internal/core/rule_engine.go`
- `cmd/verify-db-health/main.go`
- migration backfill checks (if needed)

#### Problems
- `COALESCE(ar.account_code, a.code)` keeps `account_id` dormant because `account_code` is non-null.

#### Implementation Tasks
1. **Adjust resolution precedence**
1. Change query selection to `COALESCE(a.code, ar.account_code)`.
2. Keep company-scoped join safety: `a.company_id = ar.company_id`.

2. **Strengthen health checks**
1. Add a new check in `verify-db-health` for rows where both are present but mismatch:
   - `ar.account_code <> a.code` when `ar.account_id IS NOT NULL`.
2. Decide strictness:
   - start as warning for one release, then fail hard.

3. **Data consistency hardening (optional migration)**
1. Add migration to sync stale code from `account_id` source where mismatch exists.
2. Optionally add trigger or constraint strategy later if dual-write remains.

#### Validation
1. Extend `internal/core/rule_engine_integration_test.go` with cases:
   - `account_id` set, `account_code` stale => resolver returns account from `account_id`.
   - temporal window behavior remains intact.

#### Acceptance Criteria
1. Resolver returns the account referenced by `account_id` when present.
2. Health check reports (or blocks) code/id drift.

---

### Phase 3: Remove Schema-DML/DDL from Integration Test Bootstrap

#### Scope
- `internal/core/ledger_integration_test.go` test bootstrap helper
- any other test setup using `ALTER TABLE` / `CREATE INDEX`

#### Problems
- Tests self-apply schema changes, so they no longer prove migration correctness.

#### Implementation Tasks
1. **Strip schema mutations from tests**
1. Remove DDL from `setupTestDB` bootstrap.
2. Keep only data truncate + deterministic seed inserts.

2. **Enforce migration-first precondition**
1. Add clear failure message if expected columns/indexes are missing.
2. Optionally perform a quick smoke query and call `t.Skip`/`t.Fatalf` with remediation:
   - `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db`

3. **Document test workflow**
1. Update README/testing docs to require test DB migration before integration tests.

#### Validation
1. Run:
   - `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db`
   - `make test`
2. Confirm tests fail clearly on an un-migrated DB.

#### Acceptance Criteria
1. No integration test performs schema migration work.
2. Migration defects can no longer be hidden by test setup.

---

### Phase 4: Unify `DATABASE_URL` Policy in Verification Tools

#### Scope
- `cmd/verify-db-health/main.go`
- README docs for verification commands

#### Problems
- `verify-db` requires `DATABASE_URL`; `verify-db-health` falls back to hardcoded local DSN.

#### Implementation Tasks
1. Change `verify-db-health` to fail fast when `DATABASE_URL` is empty.
2. Keep `.env` loading support.
3. Update README examples and troubleshooting text.

#### Validation
1. `go run ./cmd/verify-db-health` without `DATABASE_URL` should fail with explicit config message.
2. With `DATABASE_URL`, command should pass/fail based on real DB state.

#### Acceptance Criteria
1. Both verification tools have the same configuration contract.

---

### Phase 5: Migration Tooling Hygiene Cleanup

#### Scope
- `migrations/apply_patch.go`
- docs/repo hygiene

#### Problems
- Extra ad-hoc migration runner in `migrations/` creates confusion and bypasses checksum-based flow.

#### Implementation Tasks
1. Remove `migrations/apply_patch.go` if unused.
2. If needed for emergency ops, relocate to `cmd/` with explicit name and README warning.
3. Ensure migration execution guidance only references `cmd/verify-db`.

#### Validation
1. `go test ./...` remains green.
2. `go run ./cmd/verify-db` still handles migration pipeline end-to-end.

#### Acceptance Criteria
1. Single blessed migration execution path remains.

## Cross-Phase Verification Matrix
Run after each phase and once at the end:
1. `go test ./...`
2. `make test`
3. `go run ./cmd/verify-db-health`
4. `go run ./cmd/verify-agent` (especially after Phase 1 output changes)

For DB-dependent commands, use explicit env:
- `DATABASE_URL=...`
- `TEST_DATABASE_URL=...`

## Rollout Order
1. Phase 2 (rule engine correctness)
2. Phase 1 (currency-safe outputs)
3. Phase 4 (tool policy consistency)
4. Phase 3 (test bootstrap cleanup)
5. Phase 5 (hygiene)

Rationale: fix runtime correctness first, then operational clarity, then test/process hardening.

## Risks and Mitigations
1. **Risk:** API/tool JSON shape changes may impact consumers.
   - **Mitigation:** Add additive fields first; deprecate old ambiguous fields with clear notes.
2. **Risk:** Existing data may have `account_code` and `account_id` mismatch.
   - **Mitigation:** introduce health-check visibility before strict enforcement.
3. **Risk:** Integration tests may fail on developer machines with stale test DBs.
   - **Mitigation:** improve error messaging and document one-command remediation.

## Deliverables Checklist
1. Code changes for all 5 phases.
2. New/updated tests for currency handling and rule resolution precedence.
3. Updated README verification instructions.
4. Clean migration/tooling directory conventions.
5. Final review note with commands run and outcomes.

