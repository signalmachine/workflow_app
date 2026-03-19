# workflow_app Schema And Module Boundaries

Date: 2026-03-19
Status: Draft canonical boundaries
Purpose: define the intended first-class modules and their ownership boundaries for the `workflow_app` replacement codebase.

## 1. First-class thin-v1 modules

The replacement codebase should start with only these first-class modules:

1. `identityaccess`
2. `ai`
3. `documents`
4. `workflow`
5. `accounting`
6. `inventory_ops`
7. `workforce`
8. `work_orders`
9. `attachments`
10. `reporting`

## 2. Support records, not primary modules

These may exist, but they are not primary product centers in v1:

1. parties
2. contacts
3. tax-foundation records
4. notifications where needed for approval or task visibility

Do not create a primary `crm` module in the replacement codebase.

## 3. Ownership rules

1. each module owns its tables, write paths, and invariants
2. other modules must not write directly into another module's tables
3. shared infrastructure does not create shared business ownership
4. cross-module workflows must stay integrated through explicit identifiers and handoff contracts
5. core business flows should compose as document -> approval -> posting -> ledger and execution outcomes

## 4. Core ownership map

### 4.1 `documents`

Owns:

1. shared document identity
2. document type
3. shared lifecycle state
4. numbering
5. posting-linkage contracts
6. canonical supported document-family registration and lifecycle participation

### 4.2 `workflow`

Owns:

1. tasks
2. approvals
3. approval queues
4. approval decisions
5. non-posted review orchestration
6. shared approval truth even when AI or domain modules trigger the approval need

### 4.3 `accounting`

Owns:

1. ledger accounts
2. journal truth
3. posting invariants
4. reversal and correction boundaries
5. centralized posting execution for accounting truth
6. receivable and payable control-account treatment

### 4.4 `inventory_ops`

Owns:

1. items
2. locations
3. inventory movements
4. receipt and issue truth
5. material-usage source records
6. movement source-destination integrity
7. billable and non-billable material-usage classification support
8. one shared inventory truth model for both trading resale flows and execution-consumption flows
9. linkage from material usage into work-order or other supported execution contexts
10. traceable-unit or installed-equipment identity support where the delivery use case requires it

### 4.5 `workforce`

Owns:

1. workers
2. labor time entries
3. labor cost-rate support

### 4.6 `work_orders`

Owns:

1. work-order execution truth
2. execution status history
3. execution-facing readiness and linkage records
4. work execution linkage to labor usage and material usage context
5. work-order-primary handling for execution flows that require work-order context

### 4.7 `reporting`

Owns:

1. derived views
2. read models
3. operator-facing inspection surfaces

## 5. Schema rules

1. tenant-relevant tables carry `org_id`
2. tenant-crossing references are blocked by schema design
3. posted ledger truth is append-only
4. inventory movement truth is append-only
5. balances and stock are derived, not stored as mutable truth
6. auditable writes and audit records should succeed or fail transactionally together
7. supported document families use one central document row each
8. duplicate posting must be blocked by database-backed constraints
9. invalid lifecycle transitions should be rejected by schema-backed or transaction-backed enforcement
10. balanced accounting truth should be enforced at the database boundary, not trusted to handlers alone
11. inventory movements must preserve explicit source and destination semantics
12. inventory usage records must distinguish trading/resale movement from service or project execution consumption
13. internal login identity, external party identity, and worker identity remain distinct concerns
14. `ai` may link to approvals, but must not become a second approval system
15. each adopted supported document family should use one-to-one linkage between its owning payload row and the central `documents` row
16. the preferred physical shape is a direct `document_id` link from the owning payload row to the central `documents` row, with one-to-one semantics enforced
