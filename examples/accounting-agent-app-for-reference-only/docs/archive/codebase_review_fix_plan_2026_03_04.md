# Codebase Review (Second Pass) and Fix Plan

Date: 2026-03-04
Reviewer: Codex (second-pass validation)
Scope: `internal/core`, `migrations`, and related adapter/service paths.

## Outcome of second review

The previously reported issues are confirmed. No code changes were applied during this review.

## Confirmed issues

### 1) Global idempotency key collision across companies (Critical)

What is confirmed:
- `journal_entries.idempotency_key` is globally unique and not scoped by company.
- Multiple domain flows generate deterministic keys based only on local IDs (`orderID`, `poID`, `movementID`).

Evidence:
- `migrations/001_init.sql`: `idempotency_key TEXT UNIQUE`
- `internal/core/ledger.go`: `ON CONFLICT (idempotency_key) DO NOTHING`
- `internal/core/order_service.go`: `invoice-order-%d`, `payment-order-%d`
- `internal/core/purchase_order_service.go`: `pay-vendor-po-%d`, `po-%d-line-%d-service-receipt`
- `internal/core/inventory_service.go`: `goods-receipt-mv-%d`, `goods-issue-order-%d`

Risk:
- Tenant A can block Tenant B posting when both produce same idempotency key string.

Fix plan:
1. Add migration to replace global uniqueness with composite uniqueness: `(company_id, idempotency_key)`.
2. Keep `idempotency_key` nullable; enforce composite unique only when key is non-null.
3. Update duplicate error handling in ledger to report company-scoped duplicate.
4. Add integration test with two companies using same idempotency key; both should post successfully.

Acceptance:
- Same idempotency key can exist in different companies.
- Duplicate within same company is still rejected.

---

### 2) Document numbering strategy metadata is not enforced (High)

What is confirmed:
- `document_types.numbering_strategy` and `resets_every_fy` are read but not used to derive effective sequence dimensions.
- `ledger` creates draft documents with `financial_year=NULL` and `branch_id=NULL` for all document types.
- Inconsistent strategy vocabulary exists (`global/per_fy/per_branch` vs `sequential` used in inventory migration).

Evidence:
- `internal/core/document_service.go`: fetches strategy, but sequence always keyed by provided `financial_year/branch_id` values.
- `internal/core/ledger.go`: draft doc insert always passes `NULL, NULL` for FY/branch.
- `migrations/005_document_types_and_numbering.sql`: expected strategies in comment: `global`, `per_fy`, `per_branch`.
- `migrations/009_inventory.sql`: inserts `GR`, `GI` with strategy `sequential`.

Risk:
- Per-FY/per-branch document types can silently behave as global.
- Strategy values are not normalized, increasing maintenance drift.

Fix plan:
1. Add migration to normalize strategy values (map `sequential` -> `global` or introduce an explicit allowed enum/check and use one vocabulary).
2. In `DocumentService`, derive sequence scope from `numbering_strategy` and `resets_every_fy`:
   - `global`: ignore FY/branch.
   - `per_fy`: require/effective FY.
   - `per_branch`: require/effective branch (+ FY if configured).
3. In `Ledger`, compute and pass financial year (from posting date) when document type uses FY-based sequencing.
4. Add tests for numbering behavior per strategy.

Acceptance:
- Document number sequences follow type strategy deterministically.
- Strategy values are consistent and validated.

---

### 3) `ReceivePO` is not end-to-end atomic (High)

What is confirmed:
- `ReceivePO` posts each line via separate write paths (`InventoryService.ReceiveStock` and `ledger.Commit`) and only then updates PO status.
- `ReceiveStock` manages and commits its own transaction.

Evidence:
- `internal/core/purchase_order_service.go`: loops lines and calls `inv.ReceiveStock(...)` / `ledger.Commit(...)`, then updates status at end.
- `internal/core/inventory_service.go`: `ReceiveStock` begins and commits its own transaction.

Risk:
- Partial receipt accounting/inventory writes if one line fails mid-loop.
- PO may remain `APPROVED` while some lines are already posted.

Fix plan:
1. Refactor PO receipt flow to a single transaction owned by `ReceivePO`.
2. Introduce/use TX-scoped inventory receive method (`ReceiveStockTx`) and ledger `CommitInTx` everywhere in this flow.
3. Update PO status and `received_at` inside same transaction.
4. Add rollback integration test: fail one line intentionally and assert no lines/movements/entries were persisted.

Acceptance:
- All line receipts + ledger postings + PO status update commit together or rollback together.

---

### 4) Silent parse failures in ledger core (Medium)

What is confirmed:
- Decimal parse errors for `ExchangeRate` and `Amount` are ignored in `ledger.executeCore` (`_, _ := decimal.NewFromString(...)`).

Evidence:
- `internal/core/ledger.go`: parse operations discard errors.

Risk:
- Defensive correctness depends entirely on upstream validation path.

Fix plan:
1. Handle parse errors explicitly in ledger core and return descriptive errors.
2. Keep `Proposal.Validate()` as primary validation; treat ledger checks as defense-in-depth.
3. Add unit test for malformed numeric strings reaching ledger core.

Acceptance:
- No ignored numeric parse errors in posting path.

---

### 5) `CancelOrder` comment/behavior mismatch (Low)

What is confirmed:
- Code allows cancellation only from `DRAFT`.
- Inline comment discusses release path for confirmed orders, which is not reachable by current guard.

Evidence:
- `internal/core/order_service.go`: state guard at `status != "DRAFT"`.
- Nearby comment suggests confirmed-order cancellation context.

Risk:
- Maintainer confusion; potential future regressions.

Fix plan:
1. Align comment to current behavior, or
2. If business rule should include confirmed cancellations, implement explicit transition + tests.

Acceptance:
- Comments and behavior match exactly.

---

## Recommended implementation order

1. Company-scoped idempotency uniqueness (Issue 1).
2. Atomic PO receive refactor (Issue 3).
3. Document strategy enforcement and normalization (Issue 2).
4. Ledger parse hardening + comment cleanup (Issues 4, 5).

## Regression test checklist

- Multi-company duplicate-key safety.
- Numbering behavior per document strategy.
- PO receipt all-or-nothing transactional integrity.

## Notes

- This document is a fix plan only.
- No production code or migration was changed in this review turn.
