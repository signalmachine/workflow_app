# service_day Foundation Coverage v1

Date: 2026-03-19
Status: Active thin-v1 foundation checklist
Purpose: define what foundation-complete means for v1 so later features do not fail because of missing structural work.

## 1. Why this document exists

Thin v1 must not become broad.

But thin v1 also must not leave out foundations that later features depend on.

This document is the control:

1. if something is required by many later features, it belongs in v1 foundation
2. if something is only a workflow or feature variant, it belongs in v2

## 2. Foundation-complete rule

V1 is foundation-complete only when the following are true:

1. later features can reuse the core data model without forcing foundational rewrites
2. core posting and truth boundaries are already present
3. AI agents can operate safely through bounded workflows
4. missing work is mostly feature breadth, not missing structural primitives

## 3. Required foundation coverage

### 3.1 Identity and control

Required:

1. org and tenant model
2. users, roles, and memberships
3. auth sessions and device sessions
4. audit events
5. idempotency
6. approval records and approval state
7. explicit boundary that shared approval truth belongs to `workflow`, while AI stores only causation and recommendation linkage into that shared model

These are foundation because every serious workflow depends on them.

### 3.2 Core master data

Required:

1. parties
2. contacts at support depth
3. items
4. inventory locations
5. ledger accounts
6. workers and labor cost-rate support records
7. tax profiles, tax codes, or equivalent foundation records

These are foundation because documents, ledgers, tax, and execution all depend on them.

### 3.3 Document foundation

Required:

1. stable document identifiers
2. document typing
3. draft, submitted, approved, posted, reversed or voided lifecycle where applicable
4. source-document linkage into downstream postings
5. durable numbering strategy where needed for accounting and tax-safe operation
6. one canonical ownership path for shared document identity and lifecycle rules across supported document families
7. one-to-one linkage between the central document row and the owning payload row for every adopted supported document family
8. central document authority for identity, lifecycle, numbering, and posting linkage, with payload truth remaining in the owning domain module

These are foundation because later invoices, receipts, stock flows, and tax flows all depend on them.

### 3.4 Accounting foundation

Required:

1. double-entry journal model
2. balanced posting validation
3. append-only posted truth
4. centralized posting service
5. reversal and correction strategy
6. source-document and idempotent-posting boundaries

These are foundation because later billing, tax, payment, and reporting all depend on them.

### 3.5 Tax foundation

Required:

1. GST treatment on relevant documents
2. TDS withholding context on relevant documents
3. tax-aware posting rules
4. tax-relevant review and reporting seams

Not required in v1:

1. every edge-case localization workflow
2. full statutory breadth
3. deep jurisdiction-specific operational tooling

### 3.6 Inventory foundation

Required:

1. item and location model
2. inventory movement ledger
3. source and destination movement semantics
4. receipt, issue, and adjustment support
5. stock derived from movements, not mutable truth fields
6. item-role and movement-purpose classification sufficient to distinguish resale stock, service-delivery materials, and direct-expense consumables
7. cost-traceable linkage into accounting and execution
8. explicit billable versus non-billable service-material usage on operational source records where billing or job costing depends on that distinction
9. identity-level traceability for serialized, lot-tracked, or installed equipment classes when the delivery use case requires it

These are foundation because later light trading, material usage, and stocked billing depend on them.

### 3.7 Execution foundation

Required:

1. work orders
2. tasks
3. one clear accountable owner model
4. worker-linked assignment and labor capture
5. labor cost-traceability into operational costing and accounting outcomes
6. execution status history
7. links from execution to documents, inventory, and accounting outcomes
8. serviced-asset or installed-unit linkage where work is performed on or results in a specific maintainable unit
9. explicit ownership split between serviced-asset records and inventory-installed-unit records when both are required

These are foundation because operational workflows need a durable execution context.

### 3.8 AI foundation

Required:

1. coordinator agent
2. specialist agents
3. capability routing
4. tool registry
5. tool policy
6. run history
7. artifact, recommendation, and approval persistence
8. delegation traces

These are foundation because the AI agent is the main operator in v1.

### 3.9 Reporting foundation

Required:

1. approval views
2. document lists
3. accounting views
4. inventory views
5. execution views
6. tax summary views
7. audit lookup views

These are foundation because humans need a safe way to inspect what the system and agents did.

## 4. What is not foundation

The following are not v1 foundation even if useful later:

1. deep CRM workflow breadth
2. advanced projects features
3. customer portal
4. broad human operational UI
5. spreadsheet exchange flows
6. payroll
7. broad tax localization edge cases
8. channel integrations
9. marketplace, rental, and broader business-mode extensions

## 5. Missing-foundation test

When deciding whether something belongs in v1, ask:

1. if we defer this, would later features force a schema rewrite
2. if we defer this, would later features force a posting-model rewrite
3. if we defer this, would later features force an AI-tooling or approval-model rewrite
4. if we defer this, would later features lose correctness rather than convenience

If the answer is yes, it is probably foundation.

If the answer is only “later UX would be nicer,” it is probably v2.

## 6. V1 completion rule

Do not call v1 complete just because a few workflows work.

Call v1 complete only when:

1. the foundation checklist in this document is covered
2. the thin-v1 scope remains controlled
3. the remaining deferred work is mostly feature breadth, UX depth, or localization depth
