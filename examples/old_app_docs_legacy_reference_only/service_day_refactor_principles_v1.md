# service_day Refactor Principles v1

Date: 2026-03-18
Status: Draft refactor principle set
Purpose: translate `docs/IMP_PRINCIPLES.md` into concrete product and architecture rules for a thin `service_day` v1.

## 1. Product statement

`service_day` v1 should be a database-first, AI-agent-first operations system built on:

1. documents
2. ledgers
3. execution context
4. approvals
5. reports

It should behave more like a disciplined business operating database wrapped by an AI agent than like a form-heavy line-of-business app.

At foundational depth, a very small service company or light trading business should be able to use the same v1 foundation without forcing broader v1 scope.

## 2. Core doctrine for `service_day`

For every capability, answer these questions first:

1. what is the document
2. what ledger movement does it create
3. what execution context does it reference
4. what approval is required
5. what report or projection will humans read
6. what tax treatment, if any, applies

If one of those answers is unclear, the feature is not ready for v1.

Because AI agents are the main users of the system in v1, the database and data model must be designed so invalid or unsafe states are rejected by default rather than relying on agent correctness.

## 3. V1 system model

### 3.1 Documents

Documents are business intent and review surfaces.

V1 document categories should be limited to:

1. operational documents
   - work orders
   - task or assignment instructions where a document shape is justified
2. commercial and finance documents
   - invoices
   - payments or receipts
   - inventory receipt or issue documents
   - tax-relevant withholding or adjustment records where needed
   - journal proposals or posting batches when needed
3. AI-created proposals
   - proposed document drafts
   - proposed actions awaiting approval

Rules:

1. documents can be drafted by AI
2. humans review and approve meaningful actions
3. documents do not store balance truth
4. document lifecycle must be explicit
5. posting must be deterministic, idempotent, and transactional
6. supported document families should share one canonical identity, lifecycle, numbering, and posting-linkage model even when payload ownership stays with domain modules

### 3.2 Ledgers

V1 ledgers are the truth layer.

Required ledgers:

1. financial ledger
2. inventory ledger
3. tax-relevant posting flows over the same accounting truth

Rules:

1. ledger rows are append-only
2. financial movements are double-entry and balanced
3. inventory movements are source-to-destination quantity movements
4. balances, stock, COGS, service-material consumption, and outstanding positions are derived
5. AI never writes ledger rows directly

Inventory modeling rule:

1. use one shared item and inventory foundation for both service companies and light trading operators
2. do not create separate service-side and trading-side inventory engines or ledgers
3. classify items and stock-affecting movements by economic purpose so the system can distinguish at least:
   - resale stock
   - service-delivery materials or installed components
   - non-stock consumables expensed directly where stock control is unnecessary
4. service-company material flow and trading-company resale flow must share the same truth model while preserving different costing, billing, and reporting outcomes
5. where a material or equipment class needs identity-level traceability, use explicit serial, lot, or installed-unit records rather than overloading generic quantity-only movements

### 3.3 Execution context

Execution context explains real work but does not replace ledger truth.

Required v1 execution context:

1. work orders
2. tasks
3. assignment and accountable ownership
4. worker-linked labor capture and labor-cost visibility
5. optional activity history where needed for audit or operations
6. linkage from work to inventory usage and installed or serviced units where the delivery use case requires it

Execution context is high priority because it explains why inventory moved, why labor happened, and why a document was created.

### 3.4 Reports

Reports are first-class v1 output, but not first-class truth.

Required report categories:

1. invoice list and invoice status views
2. journal and ledger views
3. inventory on-hand and movement views
4. work-order queues and status views
5. approval queues

Rules:

1. reports are derived from documents, ledgers, and execution context
2. report tables may exist as projections or caches, but not as truth owners
3. humans inspect and approve through reports and review surfaces, not by mutating ledger state directly

### 3.5 AI

AI is the main operator interface in v1.

Rules:

1. humans ask the AI agent to do business work
2. AI gathers context and proposes document actions
3. AI uses explicit tools over normal domain services
4. AI never bypasses approval, posting, audit, or schema constraints
5. AI is a client of the system, not the system authority

### 3.5a Agent-safe data modeling

V1 data modeling must assume agents will occasionally produce imperfect requests.

Rules:

1. important states should be represented explicitly rather than inferred loosely
2. invalid transitions should be blocked in schema and service boundaries
3. typed references and constrained enums should be preferred over freeform agent-written state
4. posting, tax, and inventory effects should be impossible to create outside controlled boundaries
5. the database should reject bad actions even if an agent proposes them confidently

### 3.6 Tax foundation

Foundational GST and TDS support are part of thin v1.

Rules:

1. GST and TDS belong in v1 only at foundation depth
2. tax treatment must live on documents and posting logic, not as ad hoc report-only math
3. tax flows must remain compatible with very small service-company and light-trading usage
4. deeper localization breadth and country expansion remain v2 concerns

## 4. V1 priority order

The thin-v1 priority order should be:

1. identity, auth, org, roles, audit, idempotency
2. AI run, tool, and approval framework
3. party and item foundations
4. worker foundation for assignment, labor capture, and costing
5. financial ledger, GST/TDS foundation, and posting engine
6. inventory ledger and movement engine
7. work-order and task execution context
8. document layer over operations, finance, and tax
9. report layer

## 5. V1 de-prioritization rules

These are not foundation priorities for v1:

1. CRM pipeline depth
2. marketing-style lead management
3. project-management depth
4. portal UX
5. broad web navigation design
6. cross-industry expansion planning
7. deep tax breadth beyond foundational GST and TDS

Minimal party/contact data may still exist, but only as support for documents, approvals, and execution context.

## 6. Architecture rules

1. PostgreSQL enforces invariants wherever possible.
2. Go services are the second enforcement layer, not the first.
3. Posting boundaries are centralized.
4. Every meaningful business mutation is auditable.
5. Human approval remains required for final posting and other materially risky actions.
6. The public v1 human experience is review, approve, query, and report. It is not manual form-first document authoring.
7. The AI architecture should be modern, explicit, and observable; tool orchestration, approval state, artifacts, and audit linkage are core v1 quality concerns.
8. The short-term objective is to observe and improve agent behavior on real bounded business tasks, not to maximize production breadth.
9. Human operational UI in v1 should stay minimal and limited mainly to review, approval, inspection, and reporting surfaces.
10. Multi-agent execution is preferred over one giant generalist agent; coordinator routing plus bounded specialist agents should be the default architecture for workflow execution.

## 7. Product identity rule

`service_day` v1 is not:

1. a CRM-first app
2. a portal-first app
3. a form-first ERP

`service_day` v1 is:

1. an AI-operated business system
2. a ledger-and-documents engine
3. an execution-context system centered on work orders and tasks
4. a GST/TDS-aware foundation for small service-company and light-trading operations
5. a review-and-report product for humans
