# workflow_app Execution Plan

Date: 2026-03-21
Status: Draft canonical execution plan
Purpose: define the narrow implementation path for the `workflow_app` replacement codebase.

## 1. Milestone 0: New repo bootstrap

Goal:

1. create the `workflow_app` repository with the correct thin-v1 shape from day one

Outputs:

1. Go module
2. migration runner
3. first migrations for org, users, memberships, audit, idempotency
4. canonical `new_app_docs` planning set moved into the new repo
5. initial repository naming and product identity established as `workflow_app`

## 2. Milestone 1: Control boundary and document kernel

Goal:

1. establish the first safe business-control layer

Scope:

1. identity and auth
2. audit
3. idempotency
4. attachments
5. approvals and approval queue
6. AI run history, tool policy, and bounded coordinator-to-specialist routing
7. document kernel with named minimum document families
8. persisted inbound request intake plus queue-oriented AI processing seams needed for later browser and mobile clients

Exit criteria:

1. AI tools operate only through domain services
2. approvals exist for sensitive actions
3. auditable writes are transaction-safe
4. supported document families have one canonical identity model
5. the v1 document kernel explicitly supports work-order, invoice, payment or receipt, inventory receipt, issue, and adjustment, journal, and AI-draft document families
6. the multi-agent model is observable through durable run, step, artifact, recommendation, approval, and delegation records
7. project-linked inventory consumption in v1 reuses supported inventory issue or adjustment document flows with execution-context linkage rather than introducing a separate project-document family
8. inbound user requests can persist durably before AI processing, so later asynchronous AI operation does not depend on synchronous request-response handling

Current implementation checkpoint:

1. identity memberships, sessions, and role-aware service authorization are implemented
2. audit and idempotency foundations are implemented from Milestone 0
3. the shared document kernel now has central document identity, lifecycle state, and document-family registration
4. approvals, approval queue entries, and approval decisions are implemented with transactional audit writes
5. AI tool registry, tool policy, run history, artifacts, recommendations, approval linkage, and delegation traces are implemented for bounded coordinator-to-specialist routing
6. persisted inbound-request intake, durable attachment handling, and queue-oriented AI request processing are not yet implemented in the active codebase
7. remaining thin-v1 document-kernel completion is concentrated around adopted payload ownership for work-order, invoice, and payment or receipt document families, with shared party/contact support records reused where applicable rather than duplicated into document-local truth
8. this milestone is complete in its first control-boundary slice, but thin-v1 still requires adopted payload ownership plus minimum inbound-request and attachment support before the overall interaction foundation can be considered complete

## 3. Milestone 2: Accounting and tax foundation

Goal:

1. make financial truth real early

Scope:

1. ledger accounts
2. journal entries and lines
3. balanced posting
4. posting lifecycle
5. reversal path
6. GST and TDS baseline seams
7. receivable and payable control-account readiness
8. period and numbering control seams

Exit criteria:

1. unbalanced posting cannot persist
2. posted truth is append-only
3. posting is explicit and idempotent
4. accounting posting occurs only through centralized posting paths from supported documents
5. proposer, submitter, approver, poster, and timestamps remain reconstructible for audit review
6. approval and posting remain distinct control boundaries, with approver-versus-poster separation left policy-configurable rather than hard-coded globally

Current implementation checkpoint:

1. ledger accounts now exist as a first-class `accounting` foundation record
2. append-only journal entries and journal lines are implemented with database-backed balance enforcement
3. approved documents can post through one centralized accounting service with duplicate-post protection
4. reversals create explicit reversal entries rather than mutating posted truth
5. GST and TDS foundation records plus tax-aware posting validation are now implemented
6. accounting periods, effective-date control, journal review queries, and control-account balance views are now implemented
7. this milestone is complete and now serves as the accounting base for the remaining thin-v1 support-record and adopted-document work

## 4. Milestone 3: Inventory foundation

Goal:

1. make stock truth real early

Scope:

1. items
2. locations
3. movements
4. receipt and issue flows
5. movement-purpose classification
6. source and destination modeling
7. baseline billable and non-billable material usage classification
8. explicit support for both trading resale flows and execution-consumption flows
9. schema room for serialized, lot-tracked, or installed-equipment traceability where the delivery case requires it

Exit criteria:

1. stock is derived from movements
2. movements are append-only
3. service-material and resale-stock flows are explicitly distinguishable
4. inventory documents can feed both execution context and accounting outcomes through explicit handoff paths
5. inventory foundation supports both buy-sell trading and service or project execution consumption on one shared movement model
6. traceable equipment classes can preserve identity-level linkage where the business flow requires it

Current implementation checkpoint:

1. `inventory_ops` now exists as a first-class module with item and location master records
2. append-only inventory movement truth is implemented with database-backed mutation blocking
3. stock is now derived from movements rather than stored as mutable truth
4. source and destination semantics, movement-purpose classification, and billable/non-billable usage classification are enforced in the first inventory service layer
5. approved inventory document references can now validate compatible receipt, issue, and adjustment movement recording paths
6. inventory document payload ownership now exists through inventory-owned document header and line records with one-to-one `document_id` linkage back to the shared `documents` kernel
7. pending accounting handoff rows and pending execution-context links now make downstream inventory outcomes explicit without shifting ownership away from `accounting` or future execution modules
8. this milestone is complete and now serves as the inventory base for the remaining thin-v1 support-record and adopted-document work

## 5. Milestone 4: Execution foundation

Goal:

1. make work execution a first-class truth layer

Scope:

1. work orders
2. tasks
3. assignment
4. labor capture
5. execution history
6. material-usage linkage
7. support for inventory consumption against work orders and other minimal supported execution contexts

Exit criteria:

1. work orders are the primary execution record
2. tasks have one clear accountable owner
3. labor and material context can attach to work execution cleanly
4. execution records link back to source documents and forward to accounting or inventory outcomes where applicable
5. work orders are primary where present, but the model can also attach inventory consumption to project execution without requiring a broad projects module

Current implementation checkpoint:

1. `work_orders` now exists as a first-class module with work-order identity, code, title, summary, and lifecycle status
2. execution status history is now append-only and recorded transactionally with work-order state changes
3. pending `inventory_ops.execution_links` for `work_order` context can now be consumed transactionally into first-class `work_orders.material_usages`
4. inventory execution-link consumption now marks the originating inventory linkage as `linked` without shifting ownership away from `inventory_ops`
5. `workflow` now owns shared work-order task records with one clear accountable worker and append-only task lifecycle participation
6. `workforce` now owns worker master records plus first labor-entry capture with cost-rate snapshots linked to work orders and optional work-order tasks
7. `workforce` now creates pending labor-accounting handoffs from append-only labor truth without writing ledger state directly
8. `accounting` can now consume approved journal documents plus pending labor handoffs into centralized work-order labor postings while preserving idempotent posting ownership
9. `inventory_ops` accounting handoffs now carry explicit cost snapshots so `accounting` can consume pending work-order-linked inventory handoffs into centralized material-cost postings without weakening posting ownership
10. remaining thin-v1 execution completion includes linking adopted work-order payload truth back to the shared document kernel with one-to-one ownership semantics
11. this milestone is complete in its first execution slice, but thin-v1 still requires adopted work-order document ownership completion

## 6. Milestone 5: Reports and review

Goal:

1. make the system usable for human review without broad UI scope

Scope:

1. approval queue
2. document lists
3. accounting views
4. inventory views
5. work-order views
6. audit lookup
7. baseline GST and TDS summary views

Exit criteria:

1. humans can inspect current truth without raw database access
2. reports reconcile to source documents and ledgers
3. thin v1 is operationally reviewable
4. humans can review the document -> approval -> posting chain without reconstructing it manually

Current implementation checkpoint:

1. `reporting` now exists as a first-class read-only module for operator-facing inspection surfaces
2. approval queue review now joins queue, approval, document, and posting state in one reporting read path
3. document review now exposes the document -> approval -> posting chain in one read model
4. accounting journal review and control-account balance review now sit behind `reporting` read surfaces rather than only domain-local queries
5. baseline GST and TDS summary views now exist as first-class reporting outputs with tax-code and control-account linkage
6. inventory stock review now exposes item and location labels on derived stock truth without requiring raw SQL
7. inventory movement review now exposes movement history with source and destination context plus linked document metadata in one reporting read path
8. inventory reconciliation review now exposes document-line accounting and execution handoff state, linked work-order context, and posted journal linkage for inventory review without raw SQL
9. work-order review now exposes task, labor, material-usage, and posted-cost rollups in one inspection surface
10. audit lookup now exists as a coherent reporting read path scoped to tenant and entity filters
11. minimum thin-v1 party and contact support depth now exists through tenant-safe `parties` support records and support-depth contacts
12. remaining thin-v1 completion is now concentrated around two narrow foundation gaps: adopted-document ownership for work-order, invoice, and payment or receipt flows, plus minimum inbound-request and attachment support for browser-based user testing
13. the next implementation target is to complete those remaining thin-v1 foundation gaps before any v2 breadth work begins

## 7. Execution warning

Do not add CRM breadth, advanced projects, portal work, payroll, broad UI work, or advanced agent-autonomy features during milestones 0 through 5.

## 8. Quality and sophistication rule

`workflow_app` is allowed to be thin in breadth, but it is not allowed to be weak in foundation design.

Implementation rule:

1. solve the foundational modeling problems in v1
2. defer breadth, not rigor
3. let v2 inherit a strong schema and control model rather than a quick MVP patchwork
