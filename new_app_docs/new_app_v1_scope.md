# New App Thin v1 Scope

Date: 2026-03-19
Status: Draft canonical scope
Purpose: define the strict v1 boundary for the replacement codebase.

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

### 2.2 Master records

1. parties
2. contacts as support detail only
3. items
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

### 2.4 Accounting foundation

1. journal entries and lines
2. balanced posting validation
3. append-only posted truth
4. reversal and correction path
5. GST and TDS-aware posting seams at foundation depth

### 2.5 Inventory foundation

1. inventory movement ledger
2. receipt, issue, and adjustment support
3. derived on-hand quantity
4. item-role and movement-purpose classification
5. service-material and resale-stock distinction

### 2.6 Execution foundation

1. work orders
2. tasks
3. assignment ownership
4. labor capture
5. execution status history
6. linkage to documents, inventory, and accounting outcomes

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

## 4. Scope test

Before adding a feature, ask:

1. does it strengthen documents, ledgers, execution, approvals, or reports
2. does deferring it force a schema rewrite later
3. does deferring it force a posting-model rewrite later
4. does deferring it force an approval-model rewrite later

If the answer is no, defer it to v2.
