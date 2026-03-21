# workflow_app v1 Gap Review From Current Codebase

Date: 2026-03-21
Status: Draft review note
Purpose: compare the current repository implementation against the `workflow_app` thin-v1 foundation plan so the new codebase starts with a realistic gap view.

## 1. Review conclusion

The current codebase proves some important foundation slices, but it does not yet satisfy the planned thin-v1 foundation shape for `workflow_app`.

The biggest issue is not lack of sophistication. The biggest issue is uneven sophistication:

1. identity, auth, audit, idempotency, AI traceability, and parts of workflow are already credible
2. accounting, documents, and the first inventory movement foundation now have serious kernels
3. workforce, work-order execution, and the inventory-to-accounting bridge now have credible foundation slices
4. persisted inbound-request intake, attachment references, queue-oriented AI processing seams, and browser-usable reporting review now have their first required slice
5. the remaining thin-v1 gaps are now mostly reporting polish and future breadth controls rather than missing control-boundary primitives
6. CRM depth is still much heavier than the remaining missing foundation layers

This confirms that `workflow_app` should start new rather than trying to trim the current codebase into shape.

## 2. What already exists that is worth carrying forward conceptually

The current codebase already demonstrates useful patterns for `workflow_app`:

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

Now implemented at the required first slice:

1. persisted inbound request intake now exists with durable request-status handling, request messages, and attachment references
2. queue-oriented AI processing can now claim queued requests and link AI runs back to the originating persisted request
3. browser-usable reporting review now exists for inbound requests, processed proposals, linked approvals, and downstream documents without broad client-product breadth

### 3.2 Document foundation gaps

Still missing or not complete enough:

1. broader supported document-family adoption beyond the current accounting-linked kernel
2. owning payload completion for adopted work-order, invoice, and payment or receipt document families so the one-to-one document ownership rule is actually satisfied
3. stronger shared lifecycle participation for later invoice, payment, inventory, and work-order document families
4. fuller shared numbering strategy for all supported foundation document families

### 3.3 Support-record gaps

Now implemented at the required first slice:

1. minimum party support required by invoice, payment or receipt, trading inventory, and service execution flows now exists through tenant-safe `parties` records
2. contact support depth now exists as support detail on top of those party records rather than as revived CRM breadth
3. remaining support-record work is now downstream wiring into adopted document payload ownership rather than absence of the support records themselves

### 3.4 Accounting and tax gaps

Remaining accounting work is now concentrated around later operational breadth rather than missing thin-v1 ownership primitives:

1. adopted invoice and payment or receipt payload ownership now exists on the intended one-to-one document-ownership path
2. remaining accounting work is no longer absence of journal, tax, period, review, or adopted-document foundation

### 3.5 Inventory gaps

Partially addressed, but not yet complete enough:

1. item-role and movement-purpose modeling now exist in the first `inventory_ops` slice
2. receipt, issue, and adjustment recording paths now exist on one shared movement ledger
3. stock truth is now derived from movements rather than stored mutably
4. service-material versus resale-stock separation now exists, and inventory document payload ownership plus execution handoff seams and costed accounting handoffs now exist on top of the shared movement ledger
5. pending work-order-linked inventory handoffs can now be consumed through centralized accounting posting without crossing ownership boundaries
6. remaining inventory depth is now concentrated around future traceable-unit detail and downstream adoption wiring rather than absence of core review/reporting surfaces or basic document ownership

### 3.6 Workforce and execution gaps

Partially addressed, but not yet complete enough:

1. worker master records now exist as a distinct `workforce` bounded context
2. labor capture now exists with first cost-rate snapshotting on append-only labor entries
3. task and accountable-owner depth now exists through shared `workflow.tasks` linked to work orders with one accountable worker
4. execution linkage now reaches accounting outcomes for both labor and the first work-order material-usage slice
5. work-order execution truth now has canonical one-to-one document ownership completion rather than document-type support alone
6. remaining execution depth is now concentrated around later broader costing breadth and non-work-order execution linkage rather than absence of the first review/reporting slice

### 3.7 Reporting gaps

Partially addressed, but not yet complete enough:

1. approval review, document review, inventory stock review, work-order review, and audit lookup now exist through the first `reporting` module slice
2. accounting journal review and control-account balance review now exist through coherent reporting-oriented read surfaces rather than only domain-local list methods
3. GST and TDS summary views now exist as explicit first-class reporting outputs
4. inventory movement review and document-line inventory reconciliation now exist for inventory execution and accounting handoff inspection
5. inbound-request and processed-proposal review surfaces now exist for thin-v1 browser testing
6. remaining reporting depth is now narrower and concentrated around final operator-facing polish rather than missing request-ingress review primitives

## 4. Main shape mismatch

The current repository is most advanced where the replacement thin-v1 plan wants only support depth, and least advanced where the replacement thin-v1 plan wants the deepest foundation work.

That mismatch is:

1. too much CRM depth
2. the old codebase still carries more CRM depth than the thin-v1 plan wants
3. the replacement foundation now closes the former adopted-document and interaction-ingress gaps, so the remaining work is mainly reporting polish and controlled breadth deferral

## 5. Replacement-codebase implication

`workflow_app` should preserve quality and sophistication, but redirect that sophistication into the correct layers first:

1. stronger first migrations
2. stronger document kernel with adopted payload ownership completed
3. minimum persist-first request intake, attachment-reference handling, queued AI processing, and browser-testing ingress
4. stronger accounting, inventory, execution, labor, and reporting foundations where the remaining work is now mostly adoption wiring and final polish rather than missing kernels

`workflow_app` should not spend early sophistication budget on CRM breadth.
