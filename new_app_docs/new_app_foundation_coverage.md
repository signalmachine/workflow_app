# workflow_app Foundation Coverage

Date: 2026-03-19
Status: Draft canonical foundation checklist
Purpose: define what foundation-complete means for `workflow_app` v1 so later features do not fail because of missing structural work.

## 1. Foundation-complete rule

V1 is foundation-complete only when the following are true:

1. later features can reuse the core data model without forcing foundational rewrites
2. core posting and truth boundaries are already present
3. AI agents can operate safely through bounded workflows
4. missing work is mostly feature breadth, not missing structural primitives

## 2. Required foundation coverage

### 2.1 Identity and control

Required:

1. org and tenant model
2. users, roles, memberships, sessions, and auth boundaries
3. audit events
4. idempotency
5. approval records and approval state
6. explicit boundary that shared approval truth belongs to `workflow`, while `ai` stores only causation and recommendation linkage into that shared model
7. explicit support for one user holding memberships in multiple orgs with role assignment scoped to membership
8. one active org context per session or request so tenant-owned reads and writes stay unambiguous
9. persisted inbound request intake so original user intent is durable before AI processing begins
10. request lifecycle status suitable for queued, processing, processed, acted-on, completed, or failed handling
11. stable user-visible request references suitable for submission acknowledgments and later support or review lookup

### 2.2 Core master data

Required:

1. parties
2. contacts at support depth
3. items
4. inventory locations
5. ledger accounts
6. workers and labor cost-rate support records
7. tax-foundation records needed for GST and TDS handling

Interpretation rule:

1. the minimum party and contact foundation must be sufficient to support thin-v1 invoice, payment or receipt, trading inventory, and service execution document flows
2. support depth is enough for v1, but total absence of party or contact support is not
3. shared foundation entities should be referenced across modules through one canonical identity rather than duplicated into module-local records

### 2.3 Document foundation

Required:

1. stable document identifiers
2. document typing
3. draft, submitted, approved, posted, reversed, or voided lifecycle where applicable
4. source-document linkage into downstream postings
5. durable numbering strategy where needed for accounting and tax-safe operation
6. one canonical ownership path for shared document identity and lifecycle rules across supported document families
7. one-to-one linkage between the central document row and the owning payload row for every adopted supported document family
8. central document authority for identity, lifecycle, numbering, and posting linkage, with payload truth remaining in the owning domain module

Adopted-family rule:

1. for thin v1, adopted work-order, invoice, payment or receipt, and inventory document families must have their owning payload rows implemented rather than only document-type registration

### 2.4 Accounting foundation

Required:

1. double-entry journal model
2. balanced posting validation
3. append-only posted truth
4. centralized posting service
5. reversal and correction strategy
6. source-document and idempotent-posting boundaries

### 2.5 Tax foundation

Required:

1. GST treatment on relevant documents
2. TDS withholding context on relevant documents
3. tax-aware posting rules
4. tax-relevant review and reporting seams

Not required in v1:

1. deep localization breadth
2. full statutory edge-case tooling

### 2.6 Inventory foundation

Required:

1. item and location model
2. inventory movement ledger
3. source and destination movement semantics
4. receipt, issue, and adjustment support
5. stock derived from movements, not mutable truth fields
6. item-role and movement-purpose classification sufficient to distinguish resale stock, service-delivery materials, installed or traceable equipment, and direct-expense consumables
7. cost-traceable linkage into accounting and execution
8. explicit billable versus non-billable material usage where billing or job costing depends on that distinction
9. support for both trading resale flows and service or project execution consumption on one shared inventory model
10. identity-level traceability for serialized, lot-tracked, or installed equipment classes when the delivery use case requires it

### 2.7 Execution foundation

Required:

1. work orders
2. tasks
3. one clear accountable owner model
4. worker-linked assignment and labor capture
5. labor cost-traceability into operational costing and accounting outcomes
6. execution status history
7. links from execution to documents, inventory, and accounting outcomes
8. work-order-primary handling where work-order context exists, while still allowing minimal non-work-order execution linkage where foundation workflows require it

### 2.8 AI foundation

Required:

1. coordinator agent
2. specialist agents
3. capability routing
4. tool registry
5. tool policy
6. run history
7. artifact, recommendation, and approval persistence
8. delegation traces
9. linkage from persisted inbound requests into AI runs and downstream proposals or actions
10. provider-backed AI execution foundations using the OpenAI Go SDK and Responses API so the v1 AI layer is usable beyond persistence scaffolding
11. environment, safety, and verification support for provider-backed AI without making default local development depend on external credentials
12. backend API and transport foundations for session-auth, request submission, attachment upload and download, and review-oriented reads so the promoted v1 web layer and later mobile client can share one backend model

### 2.9 Reporting foundation

Required:

1. approval views
2. document lists
3. accounting views
4. inventory views
5. execution views
6. tax summary views
7. audit lookup views
8. inbound-request and processed-proposal review views sufficient for the promoted v1 web layer and later shared-backend client evolution

## 3. Missing-foundation test

When deciding whether something belongs in v1, ask:

1. if we defer this, would later features force a schema rewrite
2. if we defer this, would later features force a posting-model rewrite
3. if we defer this, would later features force an AI-tooling or approval-model rewrite
4. if we defer this, would later features lose correctness rather than convenience

If the answer is yes, it is probably foundation.

## 4. V1 completion rule

Do not call v1 complete just because a few workflows work.

Call v1 complete only when:

1. the foundation checklist in this document is covered
2. the thin-v1 scope remains controlled
3. the remaining deferred work is mostly feature breadth, UX depth, or localization depth
