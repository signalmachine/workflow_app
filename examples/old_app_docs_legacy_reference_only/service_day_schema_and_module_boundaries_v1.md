# service_day Schema and Module Boundaries v1

Date: 2026-03-19
Status: Active thin-v1 schema and ownership baseline
Purpose: define the minimum canonical ownership and schema rules that current thin-v1 implementation work must preserve.

## 1. Why this document exists

The thin-v1 reset is not complete if ownership, write boundaries, and foundational schema rules still live only in legacy documents.

This document is the reduced canonical baseline for those rules.

## 2. Thin-v1 module inventory

Treat these as the active thin-v1 first-class modules:

1. `identityaccess`
2. `ai`
3. `documents`
4. `accounting`
5. `inventory_ops`
6. `workflow`
7. `workforce`
8. `work_orders`
9. `attachments`
10. `reporting`

Support modules or support surfaces in thin v1:

1. party and contact support records
2. tax-foundation records
3. notifications where they support review, approval, or task visibility
4. minimal CRM records only where the already-implemented slice must be maintained or a foundation workflow still depends on them

Not a thin-v1 module center:

1. broad CRM
2. advanced `projects`
3. portal
4. payroll
5. launch/navigation UX
6. data exchange

## 3. Ownership rules

1. each module owns its tables, write paths, and invariants
2. other modules must not write directly into another module's tables
3. cross-module reads should prefer explicit services, read models, or exported queries
4. shared infrastructure does not create shared business ownership
5. product flows should feel integrated, but ownership boundaries must stay explicit

## 4. Active module responsibilities

### 4.1 `identityaccess`

Owns:

1. users, roles, memberships, sessions, device sessions, and refresh-token lifecycle
2. actor identity and authorization boundaries
3. later external principals and portal memberships if those land

Does not own:

1. worker operational identity
2. customer-party truth
3. workflow or document truth

### 4.2 `ai`

Owns:

1. run history
2. run steps
3. artifacts
4. recommendations
5. tool policy and delegation traces
6. AI-specific causation links into shared approval and audit flows

Does not own:

1. ledger truth
2. document truth outside AI artifacts and recommendations
3. direct business-state mutation outside normal domain services

### 4.3 `documents`

Owns:

1. shared document headers or equivalent document identity records
2. document typing
3. shared document lifecycle state where a supported document participates in draft, submitted, approved, posted, reversed, or voided flow
4. durable numbering strategy and series rules where accounting, tax, or operational correctness requires them
5. source-document linkage contracts into downstream posting and reporting
6. the canonical one-row-per-supported-document identity contract for adopted document families

Does not own:

1. ledger truth
2. inventory movement truth
3. work-order execution truth
4. domain-specific payload fields that belong to `accounting`, `inventory_ops`, or `work_orders`

### 4.4 `accounting`

Owns:

1. ledger accounts
2. journal entries and lines
3. posting lifecycle
4. posting invariants, reversal, and correction boundaries

Does not own:

1. operational document authoring
2. work-order execution truth
3. stock balances

### 4.5 `inventory_ops`

Owns:

1. items and inventory locations
2. inventory receipt, issue, adjustment, and movement truth
3. item-role and movement-purpose classification
4. work-order material-usage source records
5. serialized, lot-tracked, or installed-unit records and traceability where inventory-linked unit identity must persist
6. optional reservation or allocation records only if a later canonical thin-v1 update explicitly promotes them from schema room into active scope

Does not own:

1. a second trading-specific inventory model
2. broad procurement approval depth
3. warehouse-optimization scope

### 4.6 `workflow`

Owns:

1. shared task orchestration
2. task ownership and due-state rules
3. queue and reminder-ready workflow support
4. shared approval records, approval state, approval queues, and reviewer action boundaries for non-posted business actions
5. approval orchestration where business actions require review before submission or before entering the posting boundary

Does not own:

1. CRM activity truth
2. work-order lifecycle truth
3. accounting posting truth

### 4.7 `workforce`

Owns:

1. worker master records distinct from login identity
2. worker-role or skill support records where assignment quality depends on them
3. labor time-entry facts
4. labor cost-rate defaults or snapshots used for operational costing
5. later timesheet or review wrappers if governance depth is added without changing the source time fact

Does not own:

1. auth login identity
2. work-order lifecycle truth
3. payroll as a separate later bounded capability

### 4.8 `work_orders`

Owns:

1. work-order lifecycle
   - meaning execution-state progression and execution-specific state rules, not shared document identity or numbering
2. execution status history
3. execution-facing assignment or readiness rules where those are work-order-specific
4. serviced-asset records representing the maintainable unit a work order is performed on when that unit is part of the operational truth model
5. links from work orders to inventory-owned installed units where stocked traceable equipment becomes part of the maintained context
6. linkage to labor, material, and document context needed to explain execution and costing

Does not own:

1. worker master records
2. stock truth
3. ledger truth
4. shared document identity or lifecycle truth

### 4.9 `attachments`

Owns:

1. attachment metadata
2. attachment transport/storage boundary
3. attachment linkage records

### 4.10 `reporting`

Owns:

1. derived views and projections
2. approval, accounting, inventory, execution, tax, and audit read surfaces

Does not own:

1. source-of-truth business state

## 5. Foundational schema rules

1. tenant-relevant tables should carry `org_id`
2. tenant-crossing references must be guarded by schema design, not only handler logic
3. append-only truth is required for posted ledger state and inventory movements
4. stock and balances must be derived from movements and postings rather than stored as mutable summary truth
5. auditable business writes must support one transactional boundary for the business row and its audit event
6. one-time bootstrap must be database-safe, not only application-safe
7. AI traceability rows do not replace audit events for accepted business changes
8. internal identity, external party identity, and worker identity must remain separate concerns
9. supported business documents must have one canonical ownership path for identity, lifecycle, numbering, and posting linkage rather than ad hoc per-screen state
10. shared approval truth must live in `workflow`; the `ai` module may link to approvals but must not become a second approval system
11. serviced-asset truth and inventory-installed-unit truth must remain distinct even when they refer to the same real-world maintained equipment

## 5.1 Canonical shared-document contract

For every supported business document family that participates in the shared document kernel:

1. there must be exactly one central `documents` row for each supported business document
2. there must be at most one owning domain payload row for each central document
3. the `documents` row is the authority for:
   - document identity
   - document type
   - shared lifecycle state
   - numbering
   - source-document linkage
   - posting-linkage metadata or contract fields
4. the owning domain module remains the authority for domain-specific payload fields and business rules
5. supported document families must not keep a second competing lifecycle or numbering authority in their module-local tables
6. posting services for supported document families must require the central document row rather than bypassing the document kernel

Preferred physical shape for adopted document families:

1. the domain payload table carries `document_id`
2. `document_id` is tenant-safe and unique within that payload table
3. central ownership-routing fields may still exist in `documents`, but they must not weaken the one-to-one document contract

## 6. Thin-v1 foundational records

The active thin-v1 baseline should preserve schema room for:

1. orgs, users, roles, memberships, sessions, device sessions, refresh tokens
2. audit events and idempotency keys
3. attachments and attachment links
4. parties and supporting contacts
5. items and inventory locations
6. tax-foundation records
7. AI run, artifact, recommendation, approval, and policy records
8. shared document identity, lifecycle, and numbering records or equivalent canonical document structures
9. direct one-to-one payload linkage from adopted document families into the shared document kernel
10. journal entries and lines
11. workers, labor time entries, and labor cost-rate records
12. work orders and execution status history
13. shared tasks and shared approvals
14. serviced assets or installed units where the work requires that linkage

## 7. Thin-v1 modeling rules that must stay explicit

1. `work_order` is the primary execution record
2. `project` is optional and subordinate if it exists
3. activities and tasks are distinct
4. one task has one primary actionable owner
5. service-material usage, resale flows, and direct-expense consumables must not collapse into one undifferentiated inventory behavior
6. billable versus non-billable material usage must be explicit where costing or billing depends on it
7. serviced assets are distinct from both customer accounts and consumed parts
8. labor time and labor cost capture are part of thin-v1 execution foundation, not a v2-only add-on
9. assignment targets should be compatible with worker-owned execution even when a user is the acting approver or reviewer
10. inventory reservations and allocations are not active thin-v1 scope unless a canonical scope update promotes them explicitly

## 8. Legacy-reference rule

If a schema or ownership rule appears in `implementation_plan/` but not here:

1. do not assume it is still active canon
2. check whether it fits the thin-v1 scope and principles
3. promote it into `plan_docs/` before treating it as a current implementation requirement
