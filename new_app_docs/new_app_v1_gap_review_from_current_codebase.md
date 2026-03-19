# New App v1 Gap Review From Current Codebase

Date: 2026-03-19
Status: Draft review note
Purpose: compare the current repository implementation against the replacement thin-v1 foundation plan so the new codebase starts with a realistic gap view.

## 1. Review conclusion

The current codebase proves some important foundation slices, but it does not yet satisfy the planned thin-v1 foundation shape for the replacement app.

The biggest issue is not lack of sophistication. The biggest issue is uneven sophistication:

1. identity, auth, audit, idempotency, AI traceability, and parts of workflow are already credible
2. accounting and documents have a first serious kernel
3. inventory, workforce, work-order execution, and reporting remain materially missing
4. CRM depth is much heavier than the missing foundation layers

This confirms that the replacement app should start new rather than trying to trim the current codebase into shape.

## 2. What already exists that is worth carrying forward conceptually

The current codebase already demonstrates useful patterns for the replacement app:

1. tenant-safe identity and membership handling
2. audit events written transactionally with business actions
3. idempotent retry-safe write boundaries
4. shared approval ownership in `workflow` rather than AI-local approval truth
5. AI run, recommendation, artifact, and tool-policy persistence
6. first shared document kernel
7. first accounting posting boundary with balanced journal validation

These ideas should inform the new codebase, but the implementation should not be copied forward blindly.

## 3. What is still missing for the planned thin-v1 foundation app

### 3.1 Platform and control gaps

Still missing or not complete enough:

1. broader approval orchestration depth beyond the current queue and decision baseline
2. stronger review-oriented approval surface planning from the start of the new repo

### 3.2 Document foundation gaps

Still missing or not complete enough:

1. broader supported document-family adoption beyond the current accounting-linked kernel
2. stronger shared lifecycle participation for later invoice, payment, inventory, and work-order document families
3. fuller shared numbering strategy for all supported foundation document families

### 3.3 Accounting and tax gaps

Still missing or not complete enough:

1. broader accounting truth beyond the current journal shell
2. fuller posting flows from operational documents into accounting truth
3. explicit GST and TDS baseline implementation at the intended foundation depth

### 3.4 Inventory gaps

Currently missing as a first-class implementation area:

1. item-role and movement-purpose modeling at inventory-foundation depth
2. receipt, issue, and adjustment flows
3. movement ledger as the source of stock truth
4. service-material versus resale-stock separation

### 3.5 Workforce and execution gaps

Currently missing as first-class implementation areas:

1. worker master records as a distinct bounded context
2. labor capture
3. labor costing baseline
4. work orders as the primary execution record
5. execution status history
6. execution linkage to documents, inventory, and accounting outcomes

### 3.6 Reporting gaps

Currently missing as a first-class implementation area:

1. approval review views
2. accounting views
3. inventory views
4. work-order views
5. audit lookup views as a coherent reporting surface

## 4. Main shape mismatch

The current repository is most advanced where the replacement thin-v1 plan wants only support depth, and least advanced where the replacement thin-v1 plan wants the deepest foundation work.

That mismatch is:

1. too much CRM depth
2. not enough inventory depth
3. not enough workforce depth
4. not enough work-order depth
5. not enough reporting depth

## 5. Replacement-codebase implication

The replacement app should preserve quality and sophistication, but redirect that sophistication into the correct layers first:

1. stronger first migrations
2. stronger document kernel
3. stronger accounting and inventory foundations
4. stronger execution and labor foundations
5. stronger reporting and review surfaces

The replacement app should not spend early sophistication budget on CRM breadth.
