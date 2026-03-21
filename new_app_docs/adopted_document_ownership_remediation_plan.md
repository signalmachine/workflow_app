# workflow_app Adopted Document Ownership Remediation Plan

Date: 2026-03-21
Status: Completed implementation slice
Purpose: record the completed implementation slice that closed the remaining thin-v1 adopted document-family ownership gaps before later interaction foundations land on top of an unstable document model.

## 1. Problem statement

The canonical thin-v1 plan now requires one-to-one payload ownership for adopted `work_order`, `invoice`, and `payment or receipt` document families.

Current state before the slice:

1. inventory document families already have inventory-owned payload rows keyed one-to-one by `document_id`
2. `work_orders` now uses a thin work-order document payload bridge keyed by `document_id` with a unique link into execution truth
3. `invoice` and `payment_receipt` are registered document types but do not yet have owning payload tables
4. tests and posting flows still treat invoices as bare central documents rather than module-owned payload truth

## 2. Remediation objective

Land one consistent thin-v1 document-adoption model across all adopted document families so:

1. `documents` remains the canonical owner of shared identity, lifecycle, numbering, and posting linkage
2. each adopted document family has exactly one owning payload row with direct `document_id` linkage back to `documents`
3. shared support-record identities such as `parties` and `contacts` are reused rather than copied into document-local truth
4. downstream reporting, approval, posting, and AI flows can reason over one stable document adoption pattern

## 3. Scope

In scope:

1. invoice document ownership implementation
2. payment or receipt document ownership implementation
3. any follow-up tightening needed after the landed work-order ownership bridge
4. reporting and posting read-path adjustments required by the new payload tables
5. migration and service updates needed to preserve current document lifecycle and approval/posting semantics
6. integration tests proving one-to-one payload ownership and existing posting behavior

Out of scope:

1. broad CRM depth
2. portal or manual-entry UI
3. full invoicing breadth beyond the minimum payload truth needed by thin v1
4. cash-management or receivables/payables breadth beyond the minimum payment or receipt payload truth needed by thin v1

## 4. Recommended target model

### 4.1 Shared document rule

1. keep one central `documents.documents` row per adopted document
2. keep lifecycle, numbering, and posting-linkage decisions in `documents`
3. keep payload truth in the owning domain module with one `document_id` primary or unique key

### 4.2 Work-order adoption shape

Recommended shape:

1. add a work-order-owned document payload table keyed by `document_id`
2. preserve a direct one-to-one relationship between the work-order payload row and the execution truth row
3. avoid making `documents` the owner of work-order execution fields
4. decide explicitly whether the execution truth row should itself adopt `document_id` as the primary key or whether a thin work-order document header should bridge document identity to existing execution truth

Preferred direction:

1. use `document_id` as the canonical work-order identity if the migration impact stays manageable
2. if migration risk is materially higher, use a thin work-order document header keyed by `document_id` plus a unique link into the existing execution truth row, then collapse identities later only if still justified

Implemented direction:

1. the current codebase now uses the thin work-order document header approach through `work_orders.documents`
2. work-order creation creates the shared `documents` row and the work-order payload bridge transactionally
3. work-order review now surfaces the linked document identity and status while execution-facing foreign keys remain stable on `work_orders.work_orders`

### 4.3 Invoice adoption shape

Recommended minimum payload:

1. `document_id`
2. invoice role or subtype if needed
3. billed party identity
4. billing contact reference where applicable
5. currency code
6. summary or reference fields needed by posting and review
7. created-by and timestamps

Notes:

1. keep line-item breadth thin unless a concrete posting or reporting invariant requires more in v1
2. use shared `parties` and `contacts` identities rather than document-local customer/contact copies

### 4.4 Payment or receipt adoption shape

Recommended minimum payload:

1. `document_id`
2. direction or subtype needed to distinguish payment from receipt
3. counterparty reference through shared support records where applicable
4. currency code
5. payment reference fields needed for approval, review, and posting
6. created-by and timestamps

## 5. Milestone breakdown

### 5.1 Schema work

1. add module-owned payload tables for invoice and payment or receipt documents, with the work-order payload bridge already landed
2. enforce one-to-one linkage back to `documents.documents`
3. enforce tenant-safe foreign keys to shared support records and other referenced rows
4. backfill or bridge any remaining legacy work-order records into the adopted model without breaking current review and posting paths

### 5.2 Service-layer work

1. add explicit create and load flows for adopted payload rows, extending the landed work-order pattern where it still fits
2. ensure document creation and payload creation happen transactionally together
3. preserve current approval and posting flows while switching downstream reads to payload-aware joins
4. reject attempts to post or review document families that are still missing required payload truth

### 5.3 Reporting work

1. extend reporting read models to surface the adopted payload information where needed
2. keep document review centered on the shared document chain rather than spreading lifecycle state into module-local read paths
3. add any missing read models needed for thin-v1 invoice and payment or receipt inspection

### 5.4 Test work

1. migration coverage for one-to-one ownership constraints
2. integration tests for create -> submit -> approve -> post using adopted invoice payloads
3. integration tests for payment or receipt payload ownership and approval compatibility

## 6. Risks and technical challenges

1. invoice and payment or receipt payload design can easily sprawl if convenience fields are added before strict v1 scope discipline is enforced
2. reporting queries may need careful refactoring so they do not accidentally duplicate truth between `documents` and the new payload tables
3. the landed work-order bridge intentionally avoids a direct `document_id` primary-key migration, so later collapse should be reconsidered only if it delivers clear value

## 7. Outcome

Landed implementation:

1. `accounting.invoice_documents` now owns invoice payload truth through a one-to-one `document_id` bridge back to `documents.documents`
2. `accounting.payment_receipt_documents` now owns payment or receipt payload truth through the same one-to-one `document_id` bridge pattern
3. accounting service flows now create the shared document row and the module-owned payload row transactionally for both adopted accounting document families
4. posting now rejects bare `invoice` documents that bypass adopted payload ownership
5. integration tests now cover invoice and payment-or-receipt payload ownership plus migration backfill behavior

## 8. Success criteria

This remediation slice is complete only when:

1. adopted `invoice` and `payment_receipt` families each have one owning payload path with one-to-one `document_id` linkage, with `work_order` already using that pattern through `work_orders.documents`
2. current reporting and posting flows still work against the adopted model
3. shared support identities are reused rather than copied into module-local truth
4. integration tests prove the new ownership model and guard against regression back to bare-document usage
