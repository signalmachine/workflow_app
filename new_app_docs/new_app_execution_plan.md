# New App Execution Plan

Date: 2026-03-19
Status: Draft canonical execution plan
Purpose: define the narrow implementation path for the replacement codebase.

## 1. Milestone 0: New repo bootstrap

Goal:

1. create the new repository with the correct thin-v1 shape from day one

Outputs:

1. Go module
2. migration runner
3. first migrations for org, users, memberships, audit, idempotency
4. canonical `new_app_docs` planning set moved into the new repo

## 2. Milestone 1: Control boundary and document kernel

Goal:

1. establish the first safe business-control layer

Scope:

1. identity and auth
2. audit
3. idempotency
4. attachments
5. approvals and approval queue
6. AI run history and tool policy
7. document kernel

Exit criteria:

1. AI tools operate only through domain services
2. approvals exist for sensitive actions
3. auditable writes are transaction-safe
4. supported document families have one canonical identity model

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

Exit criteria:

1. unbalanced posting cannot persist
2. posted truth is append-only
3. posting is explicit and idempotent

## 4. Milestone 3: Inventory foundation

Goal:

1. make stock truth real early

Scope:

1. items
2. locations
3. movements
4. receipt and issue flows
5. movement-purpose classification

Exit criteria:

1. stock is derived from movements
2. movements are append-only
3. service-material and resale-stock flows are explicitly distinguishable

## 5. Milestone 4: Execution foundation

Goal:

1. make work execution a first-class truth layer

Scope:

1. work orders
2. tasks
3. assignment
4. labor capture
5. execution history

Exit criteria:

1. work orders are the primary execution record
2. tasks have one clear accountable owner
3. labor and material context can attach to work execution cleanly

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

Exit criteria:

1. humans can inspect current truth without raw database access
2. reports reconcile to source documents and ledgers
3. thin v1 is operationally reviewable

## 7. Execution warning

Do not add CRM breadth, advanced projects, portal work, payroll, or broad UI work during milestones 0 through 5.

## 8. Quality and sophistication rule

The replacement app is allowed to be thin in breadth, but it is not allowed to be weak in foundation design.

Implementation rule:

1. solve the foundational modeling problems in v1
2. defer breadth, not rigor
3. let v2 inherit a strong schema and control model rather than a quick MVP patchwork
