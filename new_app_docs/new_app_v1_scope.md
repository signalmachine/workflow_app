# workflow_app Thin v1 Scope

Date: 2026-03-19
Status: Draft canonical scope
Purpose: define the strict v1 boundary for the `workflow_app` replacement codebase.

## 1. V1 objective

Build the minimum serious system that can:

1. accept human requests through AI
2. create reviewable business documents
3. move approved documents through controlled submission and posting paths
4. maintain strong financial and inventory truth
5. track execution through work orders and tasks
6. expose review and reporting surfaces for humans

V1 is intentionally thin in breadth but not thin in rigor.

Rules:

1. v1 must be structurally strong enough that v2 adds breadth rather than forcing foundational rewrites
2. when a hard modeling problem belongs to foundation, solve it in v1 instead of deferring it under schedule pressure

## 2. In-scope foundations

### 2.1 Platform and control

1. org, users, roles, sessions
2. audit events
3. idempotency
4. attachments where they support evidence or documents
5. shared approvals, approval queues, and approval decisioning
6. AI runs, tool policy, recommendations, artifacts, and delegation traces
7. coordinator-plus-specialist multi-agent routing at bounded foundation depth

Tenant model rule:

1. `org` is the tenant boundary for v1
2. one deployed instance may host multiple orgs through `org_id`-scoped data
3. one user may hold memberships in more than one org
4. role assignment belongs to the user membership in an org, not to the user globally
5. one session or request operates in one active org context at a time
6. switching org context should be explicit and should not weaken tenant-safety guarantees

### 2.2 Master records

1. parties
2. contacts as support detail only
3. items with enough classification to distinguish resale stock, service-delivery materials, installed or traceable equipment, and direct-expense consumables
4. inventory locations
5. ledger accounts
6. workers
7. tax-foundation records for GST and TDS baseline

### 2.3 Document foundation

1. shared document identity
2. document typing
3. shared lifecycle state
4. numbering where required
5. source-document linkage
6. one central document row per supported business document
7. minimum supported v1 document families:
8. work-order documents
9. invoice documents
10. payment or receipt documents
11. inventory receipt, issue, and adjustment documents
12. journal proposal or journal-entry documents where needed
13. AI-created draft proposals and pending actions
14. documents remain editable only until finalized and posted where applicable
15. project-linked inventory consumption in v1 uses the same supported inventory issue or adjustment document families with execution-context linkage rather than a separate project-document family

### 2.4 Accounting foundation

1. journal entries and lines
2. balanced posting validation
3. append-only posted truth
4. reversal and correction path
5. GST and TDS-aware posting seams at foundation depth
6. explicit centralized posting from approved documents into accounting truth
7. receivable and payable control-account readiness on the shared core
8. accounting-period and numbering controls kept possible on the shared core

Tax scope rule:

1. v1 includes foundational GST and TDS support in documents, posting, and baseline review/reporting seams
2. deep localization breadth and full statutory edge-case tooling are deferred to v2 unless a specific item is required for v1 foundation correctness

### 2.5 Inventory foundation

1. inventory movement ledger
2. receipt, issue, and adjustment support
3. derived on-hand quantity
4. item-role and movement-purpose classification
5. service-material and resale-stock distinction
6. explicit movement source and destination
7. baseline billable versus non-billable material-usage distinction where costing or billing depends on it
8. support for two distinct inventory uses on one shared foundation:
9. buy-and-sell trading inventory flows
10. inventory consumption into service delivery or execution flows
11. work-order-linked material consumption with optional bill-through on the related customer-facing document
12. project-linked inventory consumption where a project is the execution context, without requiring a broad projects module in v1
13. schema room for identity-level traceability of serialized, lot-tracked, or installed equipment classes where the delivery use case requires it

### 2.6 Execution foundation

1. work orders
2. tasks
3. assignment ownership
4. labor capture
5. execution status history
6. linkage to documents, inventory, and accounting outcomes
7. one shared task engine
8. tasks and activities remain distinct concepts
9. work orders remain the primary execution record when work-order context exists
10. inventory consumption may also attach to a non-work-order execution context such as project execution where required by the business flow

### 2.7 Reporting foundation

1. approval queue
2. document lists
3. journal and ledger views
4. inventory views
5. work-order views
6. audit lookup
7. GST/TDS summary views at baseline depth

## 3. Explicitly out of scope for v1

1. broad CRM pipeline management
2. opportunity-heavy sales workflow
3. estimate-heavy pre-sales product depth
4. advanced projects module
5. customer portal
6. payroll
7. spreadsheet exchange
8. broad tax localization
9. large human operational UI
10. broad project-management product depth beyond the minimum execution-context linkage needed for inventory consumption
11. full statutory edge-case tooling beyond the foundational GST/TDS baseline

## 4. Scope test

Before adding a feature, ask:

1. does it strengthen documents, ledgers, execution, approvals, or reports
2. does deferring it force a schema rewrite later
3. does deferring it force a posting-model rewrite later
4. does deferring it force an approval-model rewrite later
5. does deferring it break the document -> approval -> posting -> ledger/execution chain

If the answer is no, defer it to v2.
