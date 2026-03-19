# service_day Thin v1 Scope

Date: 2026-03-19
Status: Active canonical thin-v1 scope
Purpose: define the strict v1 scope after applying the refactor principles.

## 1. V1 objective

Build the minimum serious system that can:

1. accept human requests through an AI agent
2. create and manage reviewable business documents
3. post approved documents into strong financial and inventory ledgers
4. support foundational GST and TDS handling on tax-relevant document and posting flows
5. track operational execution through work orders and tasks
6. expose report and review views for humans

This is the thin v1. Everything else is optional or v2.

At foundation depth, v1 should be usable by:

1. a very small service company
2. a light-trading operator using the same shared foundations

V1 scope rule:

1. if a capability can safely wait for v2, it should wait for v2
2. v1 should include only what is required to make the foundation operational and coherent

## 2. In-scope v1 foundations

### 2.1 Platform and safety

1. org, identity, roles, sessions
2. audit events
3. idempotency
4. attachments where they support documents or approvals
5. shared approval records, approval queues, and approval orchestration
6. AI run history, recommendations, AI causation linkage into shared approvals, and tool policy
7. agent observability records and evaluation-friendly execution traces

### 2.2 Master records

1. parties
2. contacts as supporting identity detail, not as a full CRM workspace
3. items with enough classification to distinguish resale stock, service-delivery materials, installed or traceable equipment, and non-stock consumables
4. ledger accounts
5. inventory locations
6. worker records and labor cost-rate support sufficient for assignment, labor capture, and work-order costing
7. tax profiles, tax codes, or equivalent foundation records needed for GST and TDS handling

### 2.3 Document engine

Required document kernel:

1. shared document identity
2. document typing
3. source-document linkage into downstream posting and reporting
4. durable numbering where accounting, tax, or operational correctness requires it
5. one central `documents` row per supported document
6. one-to-one linkage between the central document row and the owning module payload for adopted document families
7. shared document authority for identity, lifecycle, numbering, and posting linkage, while payload truth remains in the owning module

Required document families:

1. work-order document participation where a work order enters the shared document lifecycle
2. invoice documents
3. receipt or payment documents
4. inventory receipt documents
5. inventory issue or adjustment documents
6. accounting journal proposal or entry documents where needed
7. GST/TDS-relevant tax treatment and withholding metadata on supported documents

Required document lifecycle:

1. draft
2. submitted
3. approved or rejected
4. posted where applicable
5. reversed or voided by explicit corrective flow where applicable

### 2.4 Ledger foundations

1. double-entry financial ledger
2. inventory movement ledger
3. posting service for document to ledger transformation
4. reversal and correction rules
5. period and numbering basics only if needed for safe posting
6. foundational GST and TDS posting support for supported invoice, payment, and withholding flows
7. cost-traceable linkage so resale stock, service-material consumption, and direct-expense consumables do not collapse into one undifferentiated inventory behavior

### 2.5 Execution context

1. work orders
2. task engine
3. assignees and accountable ownership
4. worker-linked labor capture and labor-cost visibility
5. operational status history
6. links from execution records to source documents and resulting postings
7. links from work orders to labor, material usage, billable versus non-billable consumption, and serviced or installed units where relevant
8. explicit serviced-asset and installed-unit records with clear module ownership and traceability rules where relevant

### 2.6 Reports and review surfaces

1. invoice lists
2. inventory stock list, including separation of resale stock versus service-delivery material where stocked
3. inventory movement history, including movement purpose and execution linkage where relevant
4. journal entry list
5. ledger account balance views
6. work-order queue and status views
7. approval queue
8. audit trail lookup
9. GST/TDS-relevant review and summary views at foundation depth
10. work-order labor and material-consumption visibility at foundation depth

## 3. Explicit v1 interaction model

The interaction model is:

1. human asks AI agent to do work
2. AI reads context and creates or updates draft documents through domain tools
3. human reviews the proposal in a report or approval surface
4. approved documents move through controlled submission and posting services
5. humans inspect outcomes through reports

Human UI in v1 should stay minimal and focused on:

1. approval
2. review
3. inspection
4. reporting

Not part of v1:

1. form-first invoice authoring by humans
2. direct human ledger editing
3. AI direct posting to ledgers
4. deep tax breadth beyond foundational GST and TDS
5. broad human operational UI for routine business data entry

## 3.1 Existing implementation handling rule

Already-implemented application parts do not need to be removed from v1 solely because they are broader than the new plan.

Rules:

1. keep already-implemented slices when they do not conflict with the thin-v1 architecture
2. do not expand those slices as v1 priorities unless they directly support the thin-v1 foundation
3. if an already-implemented slice conflicts with ledger, document, execution, approval, or reporting rules, refactor it to align with the new plan

## 4. High-priority business capabilities in v1

Priority order:

1. AI agent control boundary
2. shared approval and document kernel
3. accounting engine
4. foundational GST and TDS support
5. inventory engine
6. work-order layer
7. labor capture and costing baseline
8. task management
9. reports

These capabilities define the product value of v1.

## 5. Low-priority or deferred areas

### 5.1 Push to v2

1. broad CRM pipeline management
2. leads and opportunity-heavy sales workflow
3. advanced projects module
4. customer portal
5. mobile speech capture productization
6. external communication channels
7. spreadsheet exchange flows
8. payroll
9. deep tax localization beyond foundational GST and TDS
10. rental, marketplace, and trading extensions

### 5.2 Allowed in v1 only as support, not as a module center

1. party and contact lookup
2. minimal account context
3. basic notes or attachments
4. lightweight customer references on work orders and invoices

## 6. CRM reset for v1

CRM should be reduced from a primary module to a support layer.

V1 should keep only:

1. party registry
2. contact detail
3. customer reference on documents and execution records
4. simple relationship lookup where needed by AI or reports

V1 should defer:

1. lead funnels
2. opportunity management
3. estimate-heavy pre-sales workflows
4. CRM timelines as a major product surface

## 6.1 Work-order modeling rule

`work_order` is the primary execution record in thin v1.

Rules:

1. `work_orders` owns execution truth, status progression, and execution-specific business rules
2. when a work order participates in shared draft, approval, posted, reversed, or voided document lifecycle, it does so through the canonical `documents` layer
3. `work_orders` must not introduce a second competing lifecycle authority for the same business document
4. a work-order payload may exist without every execution-state transition being modeled as a separate human-authored document, but any supported document lifecycle still flows through the shared document kernel

## 7. Projects reset for v1

`project` should not be a major v1 planning concern.

Rule:

1. if grouping is needed, keep it minimal and subordinate to work orders
2. do not let project management absorb execution focus
3. work orders remain the main operational context
4. labor, material, billing, and reporting truth should attach to work orders directly even when optional project roll-up exists

## 8. Acceptance test for thin v1

Thin v1 is successful if a business can do this:

1. ask the AI agent to create a work order, invoice draft, inventory receipt, inventory issue, or tax-relevant posting draft
2. review the resulting draft document
3. approve and post where allowed
4. capture labor and material usage against work orders with credible costing linkage
5. inspect financial, inventory, tax, and work-order status through reports
6. trust that truth comes from ledgers and audited postings rather than from mutable summary fields

If v1 does more than this, that is optional. If it cannot do this, v1 is incomplete.
