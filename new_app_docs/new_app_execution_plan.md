# workflow_app Execution Plan

Date: 2026-03-19
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

Exit criteria:

1. AI tools operate only through domain services
2. approvals exist for sensitive actions
3. auditable writes are transaction-safe
4. supported document families have one canonical identity model
5. the v1 document kernel explicitly supports work-order, invoice, payment or receipt, inventory receipt, issue, and adjustment, journal, and AI-draft document families
6. the multi-agent model is observable through durable run, step, artifact, recommendation, approval, and delegation records
7. project-linked inventory consumption in v1 reuses supported inventory issue or adjustment document flows with execution-context linkage rather than introducing a separate project-document family

Current implementation checkpoint:

1. identity memberships, sessions, and role-aware service authorization are implemented
2. audit and idempotency foundations are implemented from Milestone 0
3. the shared document kernel now has central document identity, lifecycle state, and document-family registration
4. approvals, approval queue entries, and approval decisions are implemented with transactional audit writes
5. the next Milestone 1 implementation target is AI run history, tool policy, recommendations, artifacts, and delegation traces

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
5. proposer, submitter, poster, and timestamps remain reconstructible for audit review

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

## 7. Execution warning

Do not add CRM breadth, advanced projects, portal work, payroll, broad UI work, or advanced agent-autonomy features during milestones 0 through 5.

## 8. Quality and sophistication rule

`workflow_app` is allowed to be thin in breadth, but it is not allowed to be weak in foundation design.

Implementation rule:

1. solve the foundational modeling problems in v1
2. defer breadth, not rigor
3. let v2 inherit a strong schema and control model rather than a quick MVP patchwork
