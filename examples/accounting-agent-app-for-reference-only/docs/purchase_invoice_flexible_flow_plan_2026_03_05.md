# Purchase Invoice Flexible Flow Plan (March 5, 2026)

## Implementation Review Notes (March 5, 2026)

This section reflects current implementation status after code changes in this repository.

### Implemented

1. Schema and migration:
   - Added migration `042_purchase_invoice_flexible_flow.sql` with:
     - PO close metadata columns (`closed_at`, `close_reason`, `closed_by_user_id`)
     - `vendor_invoices`, `vendor_invoice_lines`, `vendor_invoice_payments`
     - normalized uniqueness/indexes and idempotency uniqueness for `vendor_invoices`
2. Core domain/service:
   - Added `RecordDirectVendorInvoice`, `PayVendorInvoice`, `ClosePO`, `GetVendorInvoice`.
   - Added direct invoice payment status transitions (`OPEN -> PARTIALLY_PAID -> PAID`).
3. App service contract/wiring:
   - Added request/result DTOs and `ApplicationService` methods for direct invoice, direct payment, and PO close.
   - Added `PURCHASE_INVOICE_POLICY_MODE` gate (`strict` default, `flexible` required for new flows).
4. Adapter/API:
   - Added:
     - `POST /api/companies/{code}/vendor-invoices`
     - `POST /api/companies/{code}/vendor-invoices/{id}/pay`
     - `POST /api/companies/{code}/purchase-orders/{id}/close`
5. AI tool registration/execution:
   - Added write tools:
     - `record_direct_vendor_invoice`
     - `pay_vendor_invoice`
     - `close_purchase_order`
6. CLI/REPL:
   - Added direct commands for record/pay vendor invoice and PO close.
7. Tests:
   - Added core integration test coverage for direct invoice, partial/full payment, and PO close.

### Partially Implemented

1. AI behavior updates:
   - Tool schema + execution mapping are implemented.
   - Prompt strategy tuning for when to prefer strict vs direct flow still needs targeted prompt updates and validation.
2. Validation execution:
   - `verify-db` on test DB was executed successfully with migration `042` applied.
   - `verify-db-health` currently fails due pre-existing environment/data health issues not introduced by this feature.
   - `go test ./...` currently has pre-existing `internal/app` integration failures unrelated to this feature path.

### Pending

1. Web UI workflow page for direct vendor invoice (`WEB-PI-001`) is not implemented.
2. Backfill script for legacy strict PO invoices into `vendor_invoices` is not implemented.
3. Dual-read compatibility window in reporting/app reads is not implemented.
4. Reporting/governance slices by invoice source (`direct|po_strict|po_bypass`) are not implemented.
5. Rollout/pilot/docs update tasks remain pending.

### Business Clarification (March 5, 2026)

This is the agreed business direction and should be treated as the source of truth:

1. For inventory purchases, booking should follow two-document sequence:
   - `GR` (Goods Receipt) first,
   - `PI` (Purchase Invoice - inventory) second.
2. `GR` creates temporary procurement liability in GR clearing payable (not vendor AP).
3. `PI` clears GR clearing payable and recognizes vendor AP liability.
4. `GR`-only booking must be allowed when invoice is not yet received from vendor.
5. PO is optional and treated as convenience reference/prefill, not an enforcement object.
6. PO reference fields should be supported on both `GR` and `PI` (editable as needed).
7. Booking invoice against a PO must not auto-close PO; PO close remains separate manual action.
8. For non-inventory invoices (expenses/services), booking can proceed without `GR` using non-inventory invoice document types.
9. Service Receipt Note style flow is future scope, not current scope.

Implementation note:
- Current application has partial support for direct vendor invoices, payments, and optional PO linkage, but does not yet fully implement GR clearing flow for inventory procurements. This plan defines phased adoption to avoid breaking existing behavior.

## Goal
Support SMB purchase reality where:

1. Inventory procurement is controlled by receipt + invoice sequencing (`GR -> PI`) with clear liability transfer.
2. Non-inventory invoices remain direct and lightweight.
3. PO is optional, non-controlling, and used as planning/reference only.
4. Current strict PO lifecycle remains compatible during migration.

## Recommended System (Least Technical Risk)

1. Introduce procurement document-sequenced mode for inventory:
   - `GR` (inventory receipt + GR clearing payable),
   - `PI` (clear GR clearing + vendor AP).
2. Keep direct non-inventory invoice path for expense/service invoices:
   - `PE` for expense invoice (current scope),
   - `PS` reserved for future service-receipt-linked flow.
3. Keep strict PO lifecycle as backward-compatible fallback until migration is complete.
4. Do not implement strict PO↔invoice matching engine in v1 of this plan.
5. Keep PO reference optional on both receipt/invoice documents:
   - allow prefill from PO when selected,
   - allow manual edit/override of PO reference fields,
   - keep PO status unchanged by default posting,
   - close PO only via explicit manual close action.

## Current State (as of March 5, 2026)

- Strict PO invoice path exists and remains compatible.
- New direct vendor-invoice path and payments exist in current implementation (`vendor_invoices` path).
- Inventory-specific `GR -> PI` clearing model is not yet complete in current application.
- Current direct invoice lines are still oriented toward expense-account allocation; inventory-receipt coupling remains pending.

## Desired End State

### Path A: Inventory Purchase with Invoice Available (`GR -> PI`, same workflow action)
- User submits inventory purchase invoice booking intent.
- System posts `GR` first, then `PI` in linked sequence.
- `GR` records inventory receipt and GR clearing payable.
- `PI` references `GR`, clears GR clearing payable, and posts vendor AP.

### Path B: Inventory Goods Receipt Without Invoice (`GR` only)
- User can post goods receipt before vendor invoice is received.
- System posts `GR` only and marks invoice as pending.
- Later `PI` must reference pending `GR` and clear GR liability.

### Path C: Non-Inventory Invoice (Direct)
- Expense/service invoices can be posted directly without `GR`.
- Document type:
  - `PE` for expense invoices (current scope),
  - `PS` reserved for future service-receipt-linked flow.
- Direct non-inventory invoices create vendor AP directly.

### Path D: Optional PO Reference (No strict matching)
- PO can be selected for prefill on `GR`, `PI`, or `PE`.
- PO references remain editable metadata fields.
- Booking does not auto-close PO.

### Path E: Vendor Payment
- User settles AP against posted purchase invoices.
- System posts payment with document type `PV` (`DR AP`, `CR bank/cash`).
- Partial/full payment statuses remain supported.

## Scope Guardrails (v1)

1. Inventory flow target: `GR -> PI` with GR clearing mechanism.
2. Non-inventory flow target: direct invoice (`PE`) without `GR`.
3. PO is metadata/prefill only; no strict PO amount/line enforcement in flexible mode.
4. Existing strict PO lifecycle remains available as fallback compatibility path.
5. Service-receipt-linked invoice flow (`PS`) is out of current scope.

## Functional Requirements

1. Add inventory receipt operation for flexible mode:
   - `RecordGoodsReceipt` (`GR`)
2. Add inventory invoice operation linked to receipt:
   - `RecordInventoryPurchaseInvoice` (`PI`) with required `gr_id` in flexible inventory flow
3. Keep direct non-inventory invoice operation:
   - `RecordDirectVendorInvoice` (`PE`)
4. Keep vendor payment operation:
   - `PayVendorInvoice` (`PV`)
5. Introduce PO close as explicit manual operation only.
6. Enforce company-scoping and idempotency for all write operations.
7. Keep document type governance consistent for `GR`, `PI`, `PE`, `PV` (and reserve `PS` for future).

## Accounting Rules

### Inventory Goods Receipt (`GR`)
- Posting:
  - `DR` inventory account(s)
  - `CR` GR clearing payable account
- Header fields:
  - `DocumentTypeCode = GR`
  - `PostingDate`, `DocumentDate`, optional PO reference

### Inventory Purchase Invoice (`PI`, linked to `GR`)
- Posting:
  - `DR` GR clearing payable account
  - `CR` vendor AP account
- Invoice tax/rounding/amount details captured on `PI`.
- `PI` must reference related `GR` in flexible inventory flow.

### Direct Non-Inventory Invoice (`PE`)
- Posting:
  - `DR` expense/service account(s)
  - `CR` vendor AP account
- Header fields:
  - `DocumentTypeCode = PE`
  - `PostingDate`, `DocumentDate`, optional PO reference
- `PS` document type is reserved for future service-receipt-linked flow.

### Vendor Payment for Direct/Bypass Invoice
- Payment posting:
  - `DR` vendor AP account
  - `CR` selected cash/bank account
- Header fields:
  - `DocumentTypeCode = PV`
  - `PostingDate`, `DocumentDate`
- Payment can be full or partial; invoice status derives from total paid vs invoice amount (`OPEN`, `PARTIALLY_PAID`, `PAID`).

## Data Model Changes

## 1) Purchase order status
- Add `CLOSED` to supported PO statuses.
- Add columns:
  - `closed_at TIMESTAMPTZ NULL`
  - `close_reason TEXT NULL`
  - `closed_by_user_id INT NULL` (FK users, optional if available in context)

## 2) Vendor invoice linkage
- Keep/extend vendor invoice header table for direct + PO-bypass + inventory-invoice records:
  - `vendor_invoices`
  - Key fields:
    - `id`
    - `company_id`
    - `vendor_id`
    - `invoice_number`
    - `invoice_date`
    - `currency`
    - `exchange_rate`
    - `invoice_amount`
    - `idempotency_key`
    - `invoice_document_number`
    - `invoice_document_type` (`PI`, `PE`, future `PS`)
    - `journal_entry_id` (or idempotency key linkage)
    - `source` (`inventory_gr_pi`, `direct_non_inventory`, `po_strict`, `po_bypass`)
    - `po_id NULL`
    - `po_number NULL` (editable persisted reference)
    - `gr_id NULL` (required for inventory invoice in flexible mode)
    - `status` (`OPEN`, `PARTIALLY_PAID`, `PAID`, `VOID`)
    - `amount_paid`
    - `last_paid_at`
    - `created_by_user_id`
    - `created_at`
- Add payment linkage table:
  - `vendor_invoice_payments` with `vendor_invoice_id`, `payment_document_number`, `payment_amount`, `payment_date`, `journal_entry_id`, `created_at`.
- Keep line table in v1 for multi-line allocations:
  - `vendor_invoice_lines` with line type support (`inventory` | `expense` | `service`) and account/item references as applicable.

## 3) Goods receipt linkage (new)
- Introduce goods receipt header and lines for inventory flexible flow:
  - `goods_receipts`
  - `goods_receipt_lines`
- Key fields:
  - company/vendor references
  - receipt date/posting date
  - optional PO reference (`po_id`, `po_number`)
  - `gr_document_number`
  - `journal_entry_id`
  - status (`OPEN`, `INVOICED`, `CANCELLED`) as policy-defined
- Link invoices to receipts through `vendor_invoices.gr_id`.

## 4) Constraints and indexes
- Unique: normalized invoice number `(company_id, vendor_id, lower(trim(invoice_number)))` to prevent case/spacing duplicates.
- Unique: `(company_id, idempotency_key)` on `vendor_invoices`.
- Unique: `(company_id, idempotency_key)` on `goods_receipts`.
- FK scope checks through `company_id` joins.
- Index on `(company_id, created_at desc)` and `(company_id, po_id)`.
- Index on `(company_id, status)` for open AP lookups.
- Index on `(company_id, gr_id)` for invoice-to-receipt linkage lookups.

## Service Layer Changes

## `internal/core`
- Add `RecordGoodsReceipt(...)` for inventory receipts (`GR`).
- Add `RecordInventoryPurchaseInvoice(...)` for inventory invoice posting (`PI`) with receipt linkage.
- Keep/extend `RecordDirectVendorInvoice(...)` for non-inventory direct invoices (`PE`).
- Add `PayVendorInvoice(...)` for direct/bypass AP settlement.
- Add `ClosePO(...)` with validations:
  - allowed from `APPROVED`/`RECEIVED`/`INVOICED` (policy-defined),
  - disallow close if already `PAID`.
- Keep existing strict `RecordVendorInvoice` unchanged for backward compatibility.
- Add compatibility write rule: strict PO invoice flow writes/maintains existing PO invoice columns and optionally mirrors to `vendor_invoices` with `source=po_strict` (feature-flagged rollout).

## `internal/app`
- Add new request/result types:
  - `GoodsReceiptRequest`
  - `InventoryPurchaseInvoiceRequest`
  - `DirectVendorInvoiceRequest`
  - `PayVendorInvoiceRequest`
  - `ClosePurchaseOrderRequest`
- Expose methods in `ApplicationService`.
- Ensure document type policy enforcement applies for `GR`, `PI`, `PE`, `PV`.

## Adapter/API Changes

## Web
- Add endpoint:
  - `POST /api/companies/{code}/goods-receipts` (inventory receipt, optional `po_id`)
- Add endpoint:
  - `POST /api/companies/{code}/vendor-invoices/inventory` (`PI`, requires `gr_id` in flexible inventory mode)
- Add endpoint:
  - `POST /api/companies/{code}/vendor-invoices` (direct non-inventory, optional `po_id`)
- Add endpoint:
  - `POST /api/companies/{code}/vendor-invoices/{id}/pay`
- Add endpoint:
  - `POST /api/companies/{code}/purchase-orders/{id}/close`
- Add UI screen/form under Purchases:
  - “Record Goods Receipt” (inventory path),
  - “Record Vendor Invoice” (non-inventory direct path),
  - “Book Inventory Invoice” (links to pending/open GR).
- On PO detail page:
  - “Close PO” action with required reason.
- Invoice form requirements:
  - fields to capture `po_number` (auto-filled when booking against PO, editable manually),
  - for inventory flow, invoice form must require/select `gr_id`,
  - for non-inventory flow, no `gr_id` is required.

## REPL/CLI
- Add direct commands:
  - `/goods-receipt record ...`
  - `/vendor-invoice record-inventory --gr <id> ...`
  - `/vendor-invoice record ...`
  - `/vendor-invoice pay <id> ...`
  - `/po close <id> --reason "..."`

## AI Agent Tools
- Add write tool:
  - `record_goods_receipt`
- Add write tool:
  - `record_inventory_purchase_invoice`
- Add write tool:
  - `record_direct_vendor_invoice`
- Add write tool:
  - `pay_vendor_invoice`
- Update prompts so purchase-invoice events can choose:
  - inventory `GR -> PI` flow (default for inventory invoices),
  - `GR`-only when invoice is pending,
  - direct `PE` flow for non-inventory invoices,
  - strict PO flow only when explicitly requested/policy-required.

## Policy and Governance

1. Preferred SMB default should be flexible mode with document sequencing by invoice type.
2. Keep config:
   - `PURCHASE_INVOICE_POLICY_MODE = strict|flexible`
   - `flexible` maps to:
     - inventory: `GR -> PI`,
     - non-inventory: direct `PE`.
3. Keep `DOCUMENT_TYPE_POLICY_MODE` behavior unchanged; flexible flows must pass document-type governance for `GR`/`PI`/`PE`/`PV`.
4. Audit every posting/payment with source metadata and linkage (`gr_id`, `po_id`, `po_number` as applicable).

## Migration Plan

1. Add migration for PO `CLOSED` status + close metadata columns.
2. Add migration for `vendor_invoices`.
3. Add migration for `vendor_invoice_lines` and `vendor_invoice_payments`.
4. Add migration for constraints/indexes (including normalized invoice uniqueness + idempotency key uniqueness).
5. Backfill script: create `vendor_invoices` rows for existing PO-invoiced records (`source=po_strict`) to keep reporting continuity.
6. Add rollback-safe dual-read window in app/reporting code (prefer `vendor_invoices`; fallback to PO header fields where backfill not present).

## Test Plan

## Core tests
- Direct PI success (single and multi-line).
- Duplicate invoice number rejection per company/vendor.
- PO-linked invoice keeps PO open; PO close remains manual and separate.
- Strict mode still enforces existing PO invoice rules.
- Direct/bypass payment flow (full + partial) updates invoice status correctly.
- Idempotency tests for both invoice creation and payment (replay-safe).
- Concurrency test: two parallel requests for same vendor/invoice number -> exactly one succeeds.

## App/adapter tests
- API validation and error mapping.
- Auth and company isolation.
- REPL/CLI command coverage.
- AI write-tool argument validation.
- Compatibility read-path tests during dual-read/backfill window.

## Regression checks
- Existing PO strict lifecycle tests remain green.
- AP balance/reporting remains consistent.

## Rollout Sequence

1. Ship schema + core + app + API behind feature flags.
2. Enable flexible procurement UI/API for pilot tenants with `PURCHASE_INVOICE_POLICY_MODE=flexible`.
3. Review audit logs and reconciliation report.
4. Make flexible procurement sequencing the default for SMB tenant profile if no issues.
5. Mark schema-change rollout complete only after cloud `verify-db` and cloud `verify-db-health` both pass and are recorded in change notes.

## Gap Analysis vs Flexible Procurement Plan (as of March 5, 2026)

1. Inventory `GR -> PI` flow:
   - Plan: mandatory in flexible mode for inventory invoices.
   - Current: pending/partial. Current implementation has direct invoice path but not full GR clearing model and linkage enforcement.
2. `GR`-only pending invoice mode:
   - Plan: allowed for goods received before invoice.
   - Current: pending. No dedicated goods-receipt-first lifecycle in current flexible path.
3. Non-inventory direct flow (`PE`):
   - Plan: direct booking without GR.
   - Current: partial. Direct invoice exists but line/type governance and document-type distinctions need tightening.
4. Optional PO references on both receipt and invoice:
   - Plan: editable `po_number` and optional `po_id` on relevant docs.
   - Current: partial. PO optional linking exists in some paths; persisted editable `po_number` is pending.
5. PO prefill without strict matching:
   - Plan: convenience only.
   - Current: pending in web UX and partially available in backend pathways.
6. Flexible-mode default:
   - Plan: preferred SMB default after pilot.
   - Current: strict remains default; flexible must be explicitly enabled.
7. AI orchestration preference:
   - Plan: inventory defaults to `GR -> PI`, non-inventory defaults to direct `PE`.
   - Current: partial. New tools exist; routing/prompt policy still requires targeted tuning.

## Validation Commands (Local)

1. `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db`
2. `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db-health`
3. `go test -p 1 ./...`

## Validation Commands (Cloud/Target DB)

1. `DATABASE_URL=<CLOUD_DATABASE_URL> go run ./cmd/verify-db`
2. `DATABASE_URL=<CLOUD_DATABASE_URL> go run ./cmd/verify-db-health`

## Risks and Mitigations

1. Duplicate liabilities from mixed flows:
   - Mitigation: unique invoice constraint + clear source tagging.
2. Policy confusion between strict and flexible tenants:
   - Mitigation: explicit env/config and UI badges.
3. Reporting drift:
   - Mitigation: unify both flows through standard `PI` document posting and AP account rules.
4. Legacy/new source divergence during transition:
   - Mitigation: backfill + dual-read compatibility window + reconciliation checks before removing legacy dependencies.
5. Unsettled AP for direct invoices:
   - Mitigation: include `PayVendorInvoice` in same rollout scope (not a later optional phase).

## Suggested Implementation Phases

1. Phase 1: Compatibility baseline
   - keep strict flow stable,
   - keep existing direct invoice/payment paths operational,
   - add missing schema fields (`po_number`, invoice type/source normalization).
2. Phase 2: Inventory receipt model
   - introduce `GR` documents/tables and posting,
   - implement `GR` API/CLI/AI path,
   - add `GR` idempotency and linkage.
3. Phase 3: Inventory invoice clearing model
   - implement `PI` linked-to-`GR` posting,
   - enforce `gr_id` requirement for flexible inventory invoices,
   - prevent duplicate receipt/invoice liability effects.
4. Phase 4: Non-inventory tightening + UX
   - formalize `PE` flow,
   - add purchase UI pages for `GR`, `PI`, and `PE`,
   - support PO prefill/editable references.
5. Phase 5: Policy + AI + rollout
   - tune AI routing for inventory vs non-inventory,
   - pilot with flexible mode tenants,
   - promote flexible mode to default after reconciliation sign-off.

## Execution Backlog (For Later Implementation)

Status key:
- `todo`: not started
- `ready`: fully specified and unblocked
- `blocked`: waiting on dependency
- `done`: implemented
- `partial`: partially implemented

### Phase 1 tickets (Schema + compatibility)

1. `DB-PI-001` (`done`)
- Title: Add PO `CLOSED` lifecycle fields
- Scope: migration for PO status extension + `closed_at`, `close_reason`, `closed_by_user_id`
- Acceptance:
  - migration applies cleanly on new and existing DBs
  - no break in existing PO lifecycle tests

2. `DB-PI-002` (`done`)
- Title: Add `vendor_invoices` header table
- Scope: schema + indexes + FKs + normalized uniqueness on invoice number
- Acceptance:
  - duplicates rejected at DB level
  - cross-company FK misuse blocked

3. `DB-PI-003` (`done`)
- Title: Add `vendor_invoice_lines` table
- Scope: line-level allocations for multi-line direct invoices
- Acceptance:
  - each invoice line links to valid expense account
  - invoice header total equals sum of lines

4. `DB-PI-004` (`done`)
- Title: Add `vendor_invoice_payments` table
- Scope: payment linkage for direct/bypass invoice settlements
- Acceptance:
  - supports partial and full payment tracking
  - supports payment history query by invoice

### Phase 2 tickets (Core + app + API)

1. `CORE-PI-001` (`partial`)
- Title: Implement inventory `GR -> PI` posting in `internal/core`
- Scope: `RecordGoodsReceipt` + `RecordInventoryPurchaseInvoice` linkage, idempotency, validations
- Acceptance:
  - posts `GR` and linked `PI` documents + JEs correctly
  - GR clearing and vendor AP balances reconcile

2. `APP-PI-001` (`partial`)
- Title: Add app service contract/types for `GR -> PI` and direct `PE`
- Scope: new request/result DTOs + `ApplicationService` methods for inventory and non-inventory paths
- Acceptance:
  - adapters can call inventory and non-inventory flows without strict PO dependency

3. `API-PI-001` (`partial`)
- Title: Add `GR` and inventory-invoice endpoints in flexible mode
- Scope: request validation, auth, company scoping, error mapping for:
  - `POST /api/companies/{code}/goods-receipts`
  - `POST /api/companies/{code}/vendor-invoices/inventory`
- Acceptance:
  - inventory flow enforces `GR -> PI` linkage
  - returns created document numbers and JE references

4. `CORE-PI-002` (`done`)
- Title: Implement `PayVendorInvoice` in `internal/core`
- Scope: `PV` posting + partial/full payment status updates
- Acceptance:
  - payment JE posted with idempotent key
  - invoice transitions `OPEN -> PARTIALLY_PAID -> PAID`

5. `API-PI-002` (`done`)
- Title: Add `POST /api/companies/{code}/vendor-invoices/{id}/pay`
- Scope: request validation, auth, company scoping, error mapping
- Acceptance:
  - supports partial/full payments
  - returns payment document and updated invoice status

### Phase 3 tickets (PO close + UI)

1. `CORE-PO-001` (`done`)
- Title: Implement `ClosePO`
- Scope: domain rule for allowed statuses + closure reason enforcement
- Acceptance:
  - closed POs cannot be further received/invoiced/paid unless explicitly reopened (if supported)

2. `API-PO-001` (`done`)
- Title: Add `POST /api/companies/{code}/purchase-orders/{id}/close`
- Scope: endpoint + audit logging
- Acceptance:
  - closure reason stored
  - company and role checks enforced

3. `WEB-PI-001` (`todo`)
- Title: Add purchase workflow pages for `GR`, inventory `PI`, and direct `PE`
- Scope: form UX + optional PO link + PO number capture/edit + inventory receipt + GR-linked invoice flow
- Acceptance:
  - works on desktop/mobile
  - clear strict/flexible mode messaging

### Phase 4 tickets (Policy + AI + CLI/REPL)

1. `CFG-PI-001` (`done`)
- Title: Add `PURCHASE_INVOICE_POLICY_MODE=strict|flexible`
- Scope: policy gate in app layer
- Acceptance:
  - strict preserves current behavior
  - flexible enables direct + bypass flows

2. `AI-PI-001` (`partial`)
- Title: Add flexible procurement write tools (`GR`, inventory `PI`, direct `PE`, payment)
- Scope: tool schemas, execution mapping, prompt updates
- Acceptance:
  - AI defaults inventory flow to `GR -> PI`
  - AI uses direct `PE` for non-inventory invoices
  - AI can settle direct/bypass invoice via payment tool
  - still routes strict PO path when requested

3. `CLI-PI-001` (`done`)
- Title: Add REPL/CLI commands for `GR`, inventory `PI`, direct `PE`, and PO close
- Scope: `/goods-receipt record`, `/vendor-invoice record-inventory`, `/vendor-invoice record`, `/vendor-invoice pay`, `/po close`
- Acceptance:
  - parity with web API capability

### Phase 5 tickets (Pilot + governance)

1. `OPS-PI-001` (`todo`)
- Title: Pilot rollout in selected tenant(s)
- Scope: enable flexible mode for pilot only
- Acceptance:
  - no critical reconciliation drift for pilot period

2. `RPT-PI-001` (`todo`)
- Title: Add reporting slices by invoice source (`direct|po_strict|po_bypass`)
- Scope: diagnostics in governance/reporting layer
- Acceptance:
  - finance can audit non-PO invoice usage easily

3. `DOC-PI-001` (`todo`)
- Title: Update user guide + contributor docs after release
- Scope: workflows, controls, and policy mode explanation
- Acceptance:
  - docs reflect final behavior and guardrails

## Deferred Note

This plan is intentionally documented for later implementation.  
No code rollout should start until backlog items are selected and sequenced into an execution sprint.
