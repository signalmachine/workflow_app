# workflow_app Implementation Defaults

Date: 2026-03-19
Status: Draft canonical implementation defaults
Purpose: record the active defaults that implementation should preserve unless the canonical `workflow_app` planning set is explicitly updated.

## 1. Default rules

1. this document records active defaults, not open brainstorming
2. if a new decision changes one of these defaults, update this file and the relevant companion planning doc
3. if code conflicts with this file, either fix the code or revise the active planning docs explicitly
4. legacy reference docs may be used for slice detail, but only the active `new_app_docs/` set is canonical

## 2. Locked defaults

### 2.1 Workflow ownership

1. use one shared `tasks` engine across v1 contexts
2. task orchestration belongs to `workflow`
3. each task has one primary actionable owner
4. team ownership is a queue concept, not many simultaneous primary assignees
5. tasks and activities are different concepts and must not be collapsed
6. shared approval orchestration and approval queues belong to `workflow`, even when AI or domain modules trigger the approval need

### 2.2 AI write boundary

1. AI may read, summarize, draft, recommend, and request approval
2. AI may execute bounded writes only through approved tools and normal domain services
3. financially meaningful writes remain human-gated
4. meaningful business writes and their audit trail must succeed or fail together
5. AI traceability records supplement audit; they do not replace it

### 2.3 Document identity and ownership

1. supported business and accounting documents should have stable identifiers
2. document types remain explicit
3. durable numbering exists where accounting, tax, or operational correctness requires it
4. numbering should be unique per configured series
5. numbering should not reset every financial year unless a later explicit compliant policy is adopted for a specific document class
6. `documents` owns shared document identifiers, lifecycle state, numbering, and posting-linkage contracts
7. `accounting`, `inventory_ops`, and `work_orders` own their domain-specific payloads and business rules
8. every supported business document family uses exactly one central `documents` row per document
9. the preferred table shape is a direct `document_id` link from the domain payload row to the central `documents` row, with one-to-one semantics enforced
10. central ownership-routing fields may exist in `documents`, but they do not replace the one-to-one contract between document identity and module-owned payload truth

### 2.4 Accounting and posting

1. `accounting` owns the posting boundary and ledger truth
2. operational modules may prepare posting inputs but may not write posted ledger state directly
3. posting must be explicit, idempotent, balanced, and correction-safe
4. AI may propose and, where policy allows, submit; AI may never perform final human-controlled posting

### 2.5 Inventory and execution flow

1. use one shared inventory foundation for service-led and light-trading operations
2. do not create a second trading-specific inventory model
3. distinguish resale stock, service-delivery materials, installed or traceable equipment, and direct-expense consumables explicitly
4. billable versus non-billable material usage must be explicit where costing or billing depends on it
5. `work_order` is the primary execution record when work-order context exists
6. `project` is optional and subordinate if it exists
7. inventory consumption may attach to work orders or other minimal supported execution contexts without requiring a broad projects module
8. serialized, lot-tracked, or installed-equipment identity should be preserved where the delivery use case requires it

### 2.6 Workforce and identity

1. worker identity remains distinct from login identity
2. external party identity remains distinct from worker identity
3. worker-linked labor capture is part of thin-v1 foundation, not a v2-only extension
4. assignment, time capture, and labor costing should fit together without requiring payroll to exist first

### 2.7 Tenant and session model

1. `org` is the canonical tenant unit
2. a deployed instance may host multiple orgs on one shared application foundation
3. a user may have memberships in multiple orgs
4. role and access decisions belong to the membership within the active org
5. a user may have different roles in different orgs
6. one session carries one active org context at a time
7. default-org selection may exist as a convenience, but it is not an authorization source
8. org switching must be explicit and must re-establish the active org context safely
9. tenant-owned reads and writes must always execute against the active `org_id`

### 2.8 Interface stance

1. the intended product surfaces are AI plus minimal review, approval, inspection, and reporting surfaces
2. broad human operational UI is not a thin-v1 priority
3. CLI tooling may exist for developer or support work, but it is not a first-class product interface
