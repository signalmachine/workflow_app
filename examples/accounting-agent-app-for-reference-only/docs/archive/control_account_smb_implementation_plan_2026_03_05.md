# Control Accounts for SMB: Low-Risk, High-Impact Implementation Plan

Date: 2026-03-05  
Status: Proposed (ready for execution)

## 1) Objective

Introduce light control-account guardrails that reduce high-impact posting mistakes while preserving SMB flexibility.

Target control accounts:
- AR (Accounts Receivable)
- AP (Accounts Payable)
- INVENTORY

Non-goal for this phase:
- Full enterprise hard-locking at DB layer for every posting path.
- Complex real-time reconciliation blockers.

## 2) Why this plan

Current system already posts AR/AP/Inventory through domain flows (sales, purchase, inventory), but manual JE can still directly hit these accounts.  
For SMB usage, we should keep flexibility, but protect high-risk areas with soft-then-controlled enforcement.

## 3) Principles

- Keep operational document flows unchanged (`SI`, `GR`, PO payment, etc.).
- Restrict only manual JE behavior around control accounts.
- Allow admin override with mandatory reason.
- Add observability before strict enforcement.
- Make rollout reversible via feature flags.

## 4) Phased Implementation

### Phase 1: Metadata + Visibility (No Blocking)

Scope:
- Add account metadata for control behavior.
- Seed control flags from current rule configuration.
- Add reporting on manual JE usage of control accounts.

Changes:
1. Migration `033_add_control_account_flags.sql`
   - Add `accounts.is_control_account BOOLEAN NOT NULL DEFAULT false`
   - Add `accounts.control_type VARCHAR(20) NULL`
   - Add check:
     - `control_type IN ('AR','AP','INVENTORY') OR control_type IS NULL`
   - Optional consistency check:
     - if `is_control_account = false`, allow `control_type` null
2. Backfill from `account_rules`:
   - Accounts mapped by active `AR/AP/INVENTORY` rules are marked as control accounts.
3. Read-only reporting endpoint/tool:
   - "manual journal entries that post to control accounts"
   - Include JE id, date, user (if available), account code, amount.

Acceptance criteria:
- Migration is idempotent and passes `verify-db`.
- Existing flows continue without behavior change.
- Report correctly lists historical and new manual JE hits on control accounts.

Risk:
- Very low. Additive schema + read-only reporting.

Rollback:
- Keep columns unused if disabled; no destructive rollback needed.

---

### Phase 2: Soft Guardrails (Warn, Still No Blocking)

Scope:
- Warn user in manual JE UI/API when control accounts are included.
- Record audit event for attempts.

Changes:
1. Manual JE validation path:
   - Detect if any JE line references `accounts.is_control_account = true`.
2. UI:
   - Show warning message:
     - "This is a control account. Prefer sales/purchase/inventory flow."
3. API:
   - Return warnings in validation/post response (non-fatal).
4. Audit log:
   - Log attempted direct control-account posting metadata.

Acceptance criteria:
- Users can still post manual JE.
- Warning appears consistently in UI and API responses.
- Audit records are created for each attempt.

Risk:
- Low. No rejection path introduced yet.

Rollback:
- Disable warning logic via feature flag.

---

### Phase 3: Controlled Enforcement (High Impact, SMB-Friendly)

Scope:
- Reject direct control-account manual JE by default.
- Permit admin override with explicit reason.

Changes:
1. Manual JE request contract:
   - `override_control_accounts` (bool, default false)
   - `override_reason` (string, required when override true)
2. Authorization:
   - Only admin role can use override.
3. Validation behavior:
   - If control account used and override missing: reject with clear message.
   - If override used without reason: reject.
4. Audit:
   - Persist who overrode, when, and reason.
5. Feature flag:
   - `CONTROL_ACCOUNT_ENFORCEMENT_MODE`:
     - `off`
     - `warn`
     - `enforce`

Acceptance criteria:
- Non-admin users cannot directly post manual JE to control accounts.
- Admin can post only with non-empty reason.
- All overrides are auditable.
- Domain document flows remain unaffected.

Risk:
- Medium (behavior change for manual JE users).

Mitigation:
- Keep `warn` mode active for 1-2 weeks before `enforce`.
- Publish short user guidance in UI help text.

Rollback:
- Switch mode from `enforce` back to `warn` or `off`.

---

### Phase 4: Lightweight Monthly Reconciliation (No Hard Block)

Scope:
- Add simple diagnostics to detect drift between control accounts and operational balances.

Checks:
1. AR control GL vs receivable operational/open amounts.
2. AP control GL vs payable operational/open amounts.
3. Inventory GL vs computed inventory valuation.

Output:
- Exception report with variance amount and links to detail views.
- No transaction blocking in this phase.

Acceptance criteria:
- Report can run on-demand and monthly.
- Variances are explainable and traceable.

Risk:
- Low to medium (possible noise if business process timing differs).

Mitigation:
- Define clear report cut-off date semantics.

## 5) Technical Design Notes

- Enforcement should apply to `DocumentTypeCode = 'JE'` (manual JE path), not domain-generated docs.
- Primary check should happen in application/core validation path (not only UI) to cover API/AI callers.
- Keep DB constraints minimal at first; strict DB triggers can be deferred.

## 6) Files Likely Affected

- `migrations/033_add_control_account_flags.sql`
- `internal/core/ledger.go` (or new dedicated validator)
- `internal/adapters/web/accounting.go`
- `web/templates/pages/journal_entry.templ`
- `internal/app/app_service.go` (if shared validation hooks are added)
- Integration tests in `internal/core/*_integration_test.go` and adapter tests

## 7) Testing Strategy

Required tests:
1. Manual JE with non-control accounts posts successfully.
2. Manual JE with control account:
   - allowed in `warn`
   - rejected in `enforce` without override
   - rejected for non-admin even with override flag
   - accepted for admin with valid reason
3. Domain flows still succeed:
   - Sales invoice (AR)
   - Vendor receipt/payment (AP)
   - Goods receipt/issue (Inventory)
4. Audit record creation for warnings/overrides.

Validation commands:
1. `make test`
2. `go run ./cmd/verify-db`
3. `go run ./cmd/verify-db-health`

## 8) Rollout Plan

1. Deploy Phase 1 + Phase 2 with mode=`warn`.
2. Monitor warning volume for one accounting cycle.
3. Train users: when to use sales/purchase/inventory flows instead of manual JE.
4. Enable Phase 3 mode=`enforce`.
5. Add monthly Phase 4 report in the next sprint.

## 9) Decision Summary

Recommended now:
- Execute Phases 1-3.
- Keep Phase 4 as immediate follow-up.

This gives the best SMB balance: flexibility for daily operations, guardrails for high-risk control accounts, and minimal implementation risk.
