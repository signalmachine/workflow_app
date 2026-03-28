# Sales Document Flow Improvement Plan (March 7, 2026)

## Implementation Review Notes (March 7, 2026)

This plan is intentionally constrained to what is practical in the current application architecture. It is not a full implementation of `docs/new_application/new_app_blueprint_smb_gst_multi_agent_2026_03_06.md`.

### Current Implementation Snapshot

1. Sales lifecycle currently runs as:
   - `DRAFT -> CONFIRMED -> SHIPPED -> INVOICED -> PAID` on `sales_orders`.
2. Shipping currently posts accounting/inventory impact as `GI` (goods issue) via inventory shipment path.
3. Invoicing currently posts `SI` after order is in `SHIPPED` state.
4. Customer receipt currently posts `RC`.
5. `SO`, `SI`, `GI`, `RC` document types are operational; `DN` and `SE` are not yet active in current code/policies.

### Key Constraint

1. We should preserve current `sales_orders`-centric flow and incrementally introduce missing document-sequenced behavior.
2. We should avoid broad workflow-engine redesign in this application.
3. We should keep backward compatibility for existing `GI`-based shipment records.

## Goal

Improve sales workflow in this repository so it better matches SMB operational reality while staying compatible with current design:

1. Introduce delivery-first inventory sales flow (`DN -> SI`) without breaking existing order flow.
2. Allow delivery-only posting when invoice is pending.
3. Introduce direct non-inventory sales invoice path (`SE`) for services/other non-stock sales.
4. Keep optional `SO` reference behavior.
5. Keep accounting integrity and idempotent posting behavior.

## Recommended System (Least Technical Risk)

1. Add `DN` document type and use it for shipment posting in new path.
2. Keep current order lifecycle table (`sales_orders`) but enrich it with linkage metadata:
   - `delivery_document_id` (DN)
   - `invoice_document_id` (SI already exists)
3. Add explicit linkage between delivery and invoice documents (`DN -> SI`) using a small linkage table.
4. Keep `GI` support for legacy records and compatibility read paths.
5. Add `SE` direct invoice path for non-inventory sales without delivery posting.
6. Keep `RC` payment flow unchanged except invoice-link enrichments.

## Current State (as of March 7, 2026)

1. Sales posting is split across:
   - shipment (`GI`) for inventory/COGS,
   - invoice (`SI`) for AR/revenue,
   - receipt (`RC`) for collections.
2. Shipment accounting is tied to `ShipOrder` and inventory service shipment movements.
3. There is no dedicated `DN` document or explicit `DN -> SI` reference enforcement.
4. There is no direct service-sales invoice path using `SE`.
5. Document status model remains `DRAFT/POSTED/CANCELLED` (no `SUBMITTED` state in current app).

## Desired End State (for this repository)

### Path A: Inventory Sale With Immediate Invoice (`DN -> SI`, same action sequence)

1. User confirms delivery+invoice intent for an inventory order.
2. System posts `DN` first (inventory issue + COGS impact).
3. System posts `SI` second (AR + revenue + tax-ready extension points).
4. `SI` must reference posted `DN` to prevent duplicate inventory issue posting.

### Path B: Inventory Delivery Without Invoice (`DN` only)

1. User posts delivery when invoice is not yet ready.
2. System posts `DN` and marks order as invoice-pending.
3. Later `SI` must reference open `DN`.

### Path C: Non-Inventory Sales Invoice (Direct)

1. Service/other non-stock sales can post direct invoice without delivery posting.
2. Use `SE` document type for direct non-inventory invoice path.
3. Posting remains AR/revenue (and future GST-specific splits where configured).

### Path D: Optional SO Reference

1. `SO` remains optional planning/reference document.
2. `DN` and `SI` may carry optional `so_id`/`so_number` metadata.
3. No strict SO matching enforcement in v1 of this plan.

### Path E: Customer Receipt

1. `RC` posting remains payment settlement path.
2. Add optional allocation/link metadata to `SI` for better traceability (without full subledger rewrite).

## Scope Guardrails (v1)

1. In-scope:
   - `DN` introduction for inventory delivery posting.
   - `DN -> SI` linkage and basic enforcement.
   - `SE` direct non-inventory invoice path.
   - compatibility mode for legacy `GI` shipment history.
2. Out-of-scope:
   - full workflow-engine/state-machine platform redesign.
   - full GST domain redesign in this iteration.
   - full maker-checker `SUBMITTED` state rollout across all documents.
   - broad AI multi-agent topology changes.

## Functional Requirements

1. Add inventory delivery operation:
   - `RecordDeliveryNote` (`DN`) for inventory-goods delivery posting.
2. Update invoice operation:
   - `RecordInventorySalesInvoice` (`SI`) requiring reference to `DN` in new flow.
3. Add direct non-inventory invoice operation:
   - `RecordDirectSalesInvoice` (`SE`).
4. Keep receipt operation:
   - `RecordCustomerReceipt` (`RC`) with optional invoice linkage.
5. Enforce company scoping and idempotency on all write operations.
6. Keep existing `ShipOrder` path operational during migration; progressively route new flows to `DN`.

## Accounting Rules

### Inventory Delivery Note (`DN`)

- Posting:
  - `DR` COGS
  - `CR` Inventory
- Header fields:
  - `DocumentTypeCode = DN`
  - `PostingDate`, `DocumentDate`
  - optional `so_id` / `so_number`

### Inventory Sales Invoice (`SI`, linked to `DN`)

- Posting:
  - `DR` AR
  - `CR` Revenue (plus tax accounts when enabled)
- Must reference posted `DN` in new flow.
- Header fields:
  - `DocumentTypeCode = SI`
  - optional `so_id` / required `dn_id` (in DN-based inventory flow)

### Direct Service/Expense Sales Invoice (`SE`)

- Posting:
  - `DR` AR
  - `CR` Revenue/service income
- No `DN` required.
- Header fields:
  - `DocumentTypeCode = SE`
  - optional order/reference metadata

### Customer Receipt (`RC`)

- Posting:
  - `DR` Bank/Cash
  - `CR` AR
- Optional allocation to one or more sales invoices (minimal linkage first).

## Data Model Changes

## 1) Document types and policy seeds

1. Add `DN` and `SE` to `document_types` with global numbering policy.
2. Extend `document_type_policies` intent mapping for:
   - `goods_issue` -> `DN` (new preferred)
   - `sales_invoice` -> `SI`
   - `service_sales_invoice` -> `SE` (new intent)
3. Keep legacy `GI` active temporarily for historical compatibility and safe rollout.

## 2) Sales document linkage

1. Add `document_links` table (minimal shape):
   - `company_id`
   - `from_document_id`
   - `to_document_id`
   - `link_type` (`delivery_to_invoice`, `invoice_to_receipt`, etc.)
   - `created_at`
2. Add unique guardrails for `delivery_to_invoice` to prevent double invoicing a DN beyond policy.

## 3) Sales order metadata extension

1. Add optional columns on `sales_orders`:
   - `delivery_document_id` (FK `documents.id`, nullable)
   - `invoice_pending` boolean default false
2. Keep existing status values for compatibility; do not force global status redesign in this phase.

## 4) Optional service invoice headers (if needed)

1. Either:
   - reuse current posting path with proposal metadata only, or
   - add lightweight `sales_invoices` header for service/direct flow traceability.
2. Recommendation: start with proposal + `documents` + `document_links`; introduce dedicated table only if reporting requires it.

## Service Layer Changes

## `internal/core`

1. Add `RecordDeliveryNote(...)` to post `DN` and inventory issue/COGS entries.
2. Refactor shipment logic:
   - keep existing `ShipStockTx` internals for movement/cost computation,
   - switch document type from `GI` to `DN` for new path,
   - maintain compatibility mode for existing `GI` idempotency keys.
3. Update `InvoiceOrder(...)`:
   - accept/resolve `dn_id` linkage,
   - enforce `DN` reference when invoicing inventory-goods flow.
4. Add `RecordDirectSalesInvoice(...)` for `SE` path.
5. Keep `RecordPayment(...)` (`RC`) and add optional invoice-link persistence.

## `internal/app`

1. Add request/result DTOs:
   - `RecordDeliveryNoteRequest`
   - `RecordInventorySalesInvoiceRequest`
   - `RecordDirectSalesInvoiceRequest`
2. Add service methods in `ApplicationService`.
3. Add feature gate:
   - `SALES_DOCUMENT_FLOW_MODE = legacy|dn_si`
   - `legacy`: current behavior
   - `dn_si`: new `DN -> SI` behavior
4. Extend document type policy intent detection for `SE` and `DN` path ids.

## Adapter/API Changes

## Web

1. Add endpoint:
   - `POST /api/companies/{code}/sales/delivery-notes`
2. Add endpoint:
   - `POST /api/companies/{code}/sales/invoices/inventory` (`SI`, requires `dn_id` in `dn_si` mode)
3. Add endpoint:
   - `POST /api/companies/{code}/sales/invoices/direct` (`SE`)
4. Keep existing order endpoints and internally route by `SALES_DOCUMENT_FLOW_MODE`.
5. Add minimal UI actions:
   - “Post Delivery Note”
   - “Post Inventory Invoice”
   - “Post Direct Service Invoice”

## REPL/CLI

1. Add commands:
   - `/delivery note ...`
   - `/sales-invoice inventory --dn <id> ...`
   - `/sales-invoice direct ...`

## AI Agent Tools

1. Add write tools:
   - `record_delivery_note`
   - `record_inventory_sales_invoice`
   - `record_direct_sales_invoice`
2. Update AI document type schema allowlist to include `DN` and `SE`.
3. Keep legacy `GI` handling during transition.

## Policy and Governance

1. Default remains safe backward-compatible rollout:
   - `SALES_DOCUMENT_FLOW_MODE=legacy` initially.
2. Pilot tenants can enable:
   - `SALES_DOCUMENT_FLOW_MODE=dn_si`.
3. Keep `DOCUMENT_TYPE_POLICY_MODE` behavior unchanged.
4. Ensure violations are audited when legacy/new paths mismatch policy expectations.

## Migration Plan

1. Add migration for new document types (`DN`, `SE`) and policy rows.
2. Add migration for `document_links` table + indexes.
3. Add migration for optional `sales_orders.delivery_document_id` + `invoice_pending`.
4. Add compatibility backfill script (optional but recommended):
   - map historical shipment `GI` docs to synthetic delivery-link records where possible.
5. Use dual-read compatibility in reporting until rollout stabilizes.

## Test Plan

## Core tests

1. `DN` posting success for inventory order with stock deductions and COGS.
2. Inventory `SI` posting requires valid `DN` in `dn_si` mode.
3. `DN`-only flow keeps invoice pending until later SI.
4. Direct `SE` invoice success without delivery note.
5. Idempotency tests for `DN`, `SI`, `SE`, `RC`.
6. Concurrency tests: duplicate invoice attempts for same DN are blocked.

## App/adapter tests

1. API validation and error mapping for new endpoints.
2. Auth/company isolation checks.
3. Compatibility tests for `legacy` vs `dn_si` mode routing.
4. CLI/REPL command coverage.
5. AI tool argument validation and document type governance behavior.

## Regression checks

1. Existing order lifecycle tests remain green in `legacy` mode.
2. Existing reporting/tests for AR, inventory, and COGS remain stable.
3. No duplicate inventory issue between delivery and invoice flows.

## Rollout Sequence

1. Ship schema + core + app + API under `SALES_DOCUMENT_FLOW_MODE`.
2. Enable `dn_si` mode for pilot tenant(s).
3. Monitor:
   - document type governance logs,
   - inventory/COGS reconciliation,
   - AR and receipt settlement behavior.
4. Expand rollout gradually after reconciliation sign-off.
5. Mark schema rollout complete only after target DB `verify-db` and `verify-db-health` succeed.

## Gap Analysis vs Blueprint (Scoped for Current App)

1. Achievable now:
   - `DN -> SI` sequence for inventory sales,
   - `DN`-only pending invoice,
   - direct service invoice `SE`,
   - optional SO reference,
   - phased AI/tooling update.
2. Deferred from blueprint:
   - global `Draft -> Submitted -> Posted` workflow model,
   - full GST model and tax continuity engine,
   - full workflow engine state authority and multi-agent architecture,
   - complete service-delivery-linked sales model.

## Validation Commands (Local)

1. `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db`
2. `DATABASE_URL=$TEST_DATABASE_URL go run ./cmd/verify-db-health`
3. `go test -p 1 ./...`

## Validation Commands (Cloud/Target DB)

1. `DATABASE_URL=<CLOUD_DATABASE_URL> go run ./cmd/verify-db`
2. `DATABASE_URL=<CLOUD_DATABASE_URL> go run ./cmd/verify-db-health`

## Risks and Mitigations

1. Legacy/new flow divergence:
   - Mitigation: feature gate + compatibility path + phased rollout.
2. Duplicate inventory/COGS postings:
   - Mitigation: enforce `DN -> SI` linkage and idempotency constraints.
3. Policy mismatch during transition (`GI` vs `DN`):
   - Mitigation: transitional policy allowlist + audit visibility.
4. Reporting inconsistency during migration:
   - Mitigation: dual-read and explicit document-link joins in sales diagnostics.

## Suggested Implementation Phases

1. Phase 1: Schema and policy baseline
   - add `DN`/`SE`, links table, order metadata columns.
2. Phase 2: Core delivery-note posting
   - implement `DN` posting path and compatibility switch.
3. Phase 3: Invoice linkage enforcement
   - require/link `DN` for inventory `SI` in `dn_si` mode.
4. Phase 4: Direct service invoice path
   - implement `SE` API/CLI/AI path.
5. Phase 5: Pilot rollout and governance hardening
   - production pilot, reconciliations, policy tightening.

## Execution Backlog (For Later Implementation)

Status key:
- `todo`: not started
- `ready`: fully specified and unblocked
- `blocked`: waiting on dependency
- `done`: implemented
- `partial`: partially implemented

### Phase 1 tickets (Schema + policy)

1. `DB-SALES-001` (`todo`)
- Title: Add `DN` and `SE` document types + policy seeds
- Acceptance:
  - migrations are idempotent
  - policy table supports transition mode (`GI` + `DN` temporarily)

2. `DB-SALES-002` (`todo`)
- Title: Add `document_links` table for `DN -> SI`
- Acceptance:
  - link uniqueness constraints prevent invalid duplicates
  - company scoping enforced by FK/index strategy

3. `DB-SALES-003` (`todo`)
- Title: Extend `sales_orders` with delivery metadata
- Acceptance:
  - `delivery_document_id` and `invoice_pending` migrate cleanly

### Phase 2 tickets (Core + app)

1. `CORE-SALES-001` (`todo`)
- Title: Implement `RecordDeliveryNote` using inventory shipment mechanics
- Acceptance:
  - posts `DN` with correct journal impact and inventory movement

2. `CORE-SALES-002` (`todo`)
- Title: Enforce inventory `SI` linkage to `DN` in `dn_si` mode
- Acceptance:
  - inventory invoice blocked without valid DN link

3. `APP-SALES-001` (`todo`)
- Title: Add sales flow feature gate and app service methods
- Acceptance:
  - `legacy` and `dn_si` behavior selectable via config

### Phase 3 tickets (API + UX + AI)

1. `API-SALES-001` (`todo`)
- Title: Add delivery note and inventory/direct sales invoice endpoints
- Acceptance:
  - auth, validation, company isolation, error mapping complete

2. `WEB-SALES-001` (`todo`)
- Title: Add minimal sales workflow UI actions for DN/SI/SE
- Acceptance:
  - operators can complete new sales flows without manual SQL/CLI

3. `AI-SALES-001` (`todo`)
- Title: Add/route AI tools for DN and SE flows
- Acceptance:
  - tool routing prefers `DN -> SI` for inventory and `SE` for non-inventory

### Phase 4 tickets (Rollout + reporting)

1. `OPS-SALES-001` (`todo`)
- Title: Pilot `dn_si` mode and monitor reconciliation
- Acceptance:
  - no material inventory/COGS/AR reconciliation drift

2. `RPT-SALES-001` (`todo`)
- Title: Add sales diagnostics by document path (`legacy_gi_si|dn_si|se_direct`)
- Acceptance:
  - finance can audit adoption and detect anomalies quickly

3. `DOC-SALES-001` (`todo`)
- Title: Update user/contributor docs after rollout sign-off
- Acceptance:
  - operational and technical docs reflect final behavior

## Deferred Note

This plan is documented for phased implementation.
No immediate code rollout should start until these tickets are prioritized into an execution sprint.
