# Accounting Foundations Fix Plan (Deferred)

Date: March 4, 2026
Status: Deferred backlog document for later implementation
Scope: Data modeling and accounting-foundation hardening

This document captures follow-up fixes for the accounting/data-model review. It incorporates both the original findings and the clarified verdicts/severity levels below.

## Priority Summary

1. Critical
- ReceivePO non-atomic posting (Issue 1)

2. Medium
- AP variance handling when invoice differs from receipt (Issue 2)
- Missing DB-level ledger invariants (Issue 4)
- `account_rules` temporal model mismatch (Issue 6)
- Account references stored as codes without FK integrity (Issue 7)
- Idempotency key uniqueness not tenant-scoped (Issue 8)

3. Low
- Multi-currency in procurement path currently unimplemented (Issue 3)
- CreatePO missing strict quantity/cost validation (Issue 5)

## Issue 1: ReceivePO Non-Atomic Posting

Verdict: Genuine critical issue.

Current risk:
- Receipt line processing commits inventory and journal effects per line.
- A failure mid-loop can leave inventory/ledger updated but PO status not fully transitioned.
- This creates subledger-document divergence.

Target fix:
- Make ReceivePO one transaction boundary for all side effects:
  - inventory movements
  - journal postings
  - PO status transition to `RECEIVED`
- If any line fails, rollback everything.

Required architecture:

```text
BEGIN TRANSACTION
  process all received lines
  write all inventory movements
  write all accounting entries
  update purchase_orders.status = RECEIVED
COMMIT
```

Implementation notes:
- Introduce TX-scoped receive helpers similar to existing TX-scoped inventory/ledger methods.
- Avoid any standalone `Commit()`/`ReceiveStock()` calls inside the line loop unless they operate on the same `pgx.Tx`.

Acceptance criteria:
- Simulated failure on line N causes zero persisted side effects for lines 1..N-1.
- PO status, inventory, and ledger are always mutually consistent.
- Integration test explicitly asserts rollback behavior under injected fault.

## Issue 2: AP Not Clearing on Invoice-Receipt Difference

Verdict: Legitimate issue; policy-dependent.
Severity: Medium.

Current risk:
- System appears to recognize AP at goods receipt (`Dr Inventory / Cr AP`).
- Vendor invoice variance is only warned/logged, not posted.
- Payment may use invoice amount, potentially leaving residual AP.

Policy options:

1) Enforce strict match
- Require `invoice_amount == received/PO amount` and reject mismatch.
- Simplest operational model for SMB usage.

2) Allow variance with accounting entry
- Post variance adjustment entry at invoice capture.
- Example (Flow A style):
  - If invoice > receipt: additional liability/variance posting.
  - If invoice < receipt: reverse variance accordingly.
- Define exact accounts (e.g., Purchase Price Variance) via `account_rules`.

Acceptance criteria:
- After full PO lifecycle (`RECEIVED -> INVOICED -> PAID`), AP for that PO nets to zero under chosen policy.
- Tests cover both positive and negative variance scenarios, or strict-reject behavior.

## Issue 3: Procurement Path Single-Currency Behavior

Verdict: Not a bug; currently unfinished functionality.
Severity: Low.

Current state:
- PO/payment path currently defaults to INR and exchange rate 1 in key places.
- Schema supports currency/exchange rate, but implementation is not fully using it.

Decision:
- Track as planned enhancement, not defect.

Future implementation:
- Respect PO/header currency + rate throughout receipt/invoice/payment accounting paths.
- Ensure base conversion logic remains consistent with ledger model.

Acceptance criteria:
- End-to-end foreign-currency PO test passes with correct base amounts and clearing behavior.

## Issue 4: Missing DB-Level Ledger Invariants

Verdict: Strong and valid improvement.
Severity: Medium.

Current risk:
- Journal line correctness is enforced at application layer only.
- Database currently lacks defensive constraints for invalid debit/credit combinations.

Target DB constraints:

```sql
CHECK (debit_base >= 0 AND credit_base >= 0)
CHECK (
    (debit_base > 0 AND credit_base = 0)
 OR (credit_base > 0 AND debit_base = 0)
)
```

Optional hardening:
- Trigger/constraint to reject unbalanced entry totals per `entry_id` at commit boundary.

Acceptance criteria:
- Invalid lines (`debit=credit`, both zero, both positive) are rejected by DB.
- Existing valid integration tests remain green.

## Issue 5: CreatePO Validation Gaps

Verdict: Valid but minor.
Severity: Low.

Current risk:
- PO lines may allow non-sensical values (e.g., non-positive quantity, negative unit cost).

Target validation:
- `quantity > 0`
- `unit_cost >= 0`
- Optional: disallow both `product_id` and `expense_account_code` being absent in contexts where that is invalid.

Acceptance criteria:
- Invalid line inputs fail fast before DB writes.
- Unit tests cover negative/zero edge cases.

## Issue 6: account_rules Temporal Model Mismatch

Verdict: Correct observation.
Severity: Medium.

Current mismatch:
- Table has `effective_from`/`effective_to`.
- Unique index shape prevents multiple versions for same rule over time.

Resolution options:

1) No temporal behavior needed
- Remove effective date columns and keep single active row model.

2) Temporal behavior required
- Change uniqueness to include version key (e.g., `effective_from`).
- Update resolver query to evaluate both `effective_from` and `effective_to`.

Acceptance criteria:
- Chosen model is explicit and consistent in schema + resolver logic.
- Tests cover rule resolution across date boundaries when temporal mode is enabled.

## Issue 7: Account References Stored as Plain Codes (No FK)

Verdict: Very good catch.
Severity: Medium.

Current risk:
- Several domain tables store account references as text codes.
- This weakens referential integrity and creates drift risk if account definitions change.

Future direction:
- Migrate from code columns to account ID references where practical.
- Keep code as optional denormalized display field if needed, but enforce FK on ID.

Candidate fields:
- `products.revenue_account_code`
- `vendors.ap_account_code`
- `vendors.default_expense_account_code`
- `purchase_order_lines.expense_account_code`
- `account_rules.account_code` (or redesign to reference account IDs)

Migration strategy (safe rollout):
1. Add nullable `*_account_id` columns with FK.
2. Backfill IDs from existing codes (company-scoped).
3. Switch service logic to IDs.
4. Add NOT NULL where required.
5. Deprecate/remove code fields where feasible.

Acceptance criteria:
- Referential integrity enforced by DB FKs for all critical account references.
- No unresolved account links remain after backfill.

## Issue 8: Idempotency Key Not Tenant-Scoped

Verdict: Correct.
Severity: Medium.

Current risk:
- Global uniqueness on `idempotency_key` can collide across companies.

Target fix:
- Replace global uniqueness with scoped uniqueness:

```sql
UNIQUE (company_id, idempotency_key)
```

Implementation notes:
- Keep behavior for empty/null idempotency keys unchanged as intended.
- Update any conflict clauses in SQL to match new index shape.

Acceptance criteria:
- Same key can be used independently by different companies.
- Duplicate key in same company is still rejected.

## Recommended Execution Order

1. Issue 1 (atomic ReceivePO)
2. Issue 8 (tenant-scoped idempotency)
3. Issue 4 (DB ledger constraints)
4. Issue 2 (AP variance policy + implementation)
5. Issue 6 (account_rules temporal decision)
6. Issue 7 (account reference FK migration)
7. Issue 5 (CreatePO validation)
8. Issue 3 (multi-currency procurement enhancement)

## Definition of Done (Program-Level)

- No workflow can leave PO state, inventory, and ledger out of sync after error.
- AP clears predictably under documented invoice variance policy.
- Ledger data shape is guarded by DB constraints, not only app logic.
- Tenant isolation is preserved for idempotency semantics.
- Account reference integrity is enforceable and auditable.
