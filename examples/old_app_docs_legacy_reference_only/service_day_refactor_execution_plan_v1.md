# service_day Refactor Execution Plan v1

Date: 2026-03-19
Status: Active canonical thin-v1 execution plan
Purpose: provide a narrow implementation sequence for the thin v1.

## 1. Execution intent

The thin-v1 execution plan exists to:

1. stop the roadmap from expanding sideways
2. build truth layers before broad workflows
3. make AI-driven operation real without giving AI authority
4. ship a usable review-and-report product with strong foundations
5. optimize first for observing agent behavior on bounded business work rather than for broad production rollout

## 2. Delivery rules

1. build foundation before breadth
2. each milestone must produce usable system behavior, not only table scaffolding
3. no module earns v1 priority unless it clearly supports documents, ledgers, execution context, approvals, or reports
4. if a feature does not strengthen thin-v1 foundations, move it to v2
5. AI observability and policy quality are foundation work, not optional polish
6. do not mark v1 complete until the foundation checklist in `service_day_foundation_coverage_v1.md` is covered
7. when implementation hits an ambiguity that the active thin-v1 docs do not answer fully, consult the relevant legacy `implementation_plan/` slice for additional detail before inventing a new local rule
8. if legacy detail is used to clarify implementation, keep `plan_docs/` authoritative and promote any now-required active rule into the canonical thin-v1 docs

## 3. Milestones

### Milestone 0: Plan reset

Goal:

1. adopt the thin-v1 scope formally
2. classify current `implementation_plan/` material into:
   - keep for v1
   - collapse into a smaller canonical doc
   - move to v2 backlog

Outputs:

1. approved replacement canonical planning set
2. migration map from old docs to new docs
3. explicit v1 versus v2 boundary
4. explicit foundation-coverage checklist for completion control

### Milestone 1: Kernel and control boundary

Goal:

1. establish the non-negotiable platform and control foundation

Scope:

1. org, auth, roles, sessions
2. audit events
3. idempotency
4. attachment support for document evidence
5. shared approval records, approval queues, and approval orchestration boundaries
6. AI runs, recommendations, AI causation linkage into shared approvals, coordinator routing, specialist capability boundaries, tool policy, and observability hooks
7. party, item, worker, and tax-foundation records
8. document kernel primitives needed before downstream posting and workflow depth:
   - shared document identity
   - document typing
   - base lifecycle states
   - source-document linkage contracts
   - one-to-one linkage contract between central document rows and adopted domain payload rows

Exit criteria:

1. AI tools operate only through domain services
2. meaningful actions are auditable
3. approval paths exist for sensitive actions
4. retry-safe mutation boundaries are established
5. supported downstream modules have one canonical document identity and lifecycle path to build on rather than ad hoc local document state
6. the document kernel contract is explicit: supported document families use one central document row per document, while payload truth remains module-owned

### Milestone 2: Financial and tax foundation

Goal:

1. make accounting truth real early

Scope:

1. ledger accounts
2. journal entries and lines
3. balanced posting validation
4. posting lifecycle
5. reversal and correction strategy
6. durable numbering and series rules where accounting and tax-safe operation require them
7. foundational GST and TDS handling on supported document and posting flows
8. accounting journals participate in the shared document kernel through the canonical one-to-one document contract rather than journal-local lifecycle state alone

Exit criteria:

1. unbalanced postings cannot persist
2. posting is idempotent
3. posted truth is append-only
4. AI can propose or submit only within policy, never directly post
5. GST and TDS effects flow through explicit document and posting boundaries rather than ad hoc report-only calculations
6. accounting posting consumes the shared document kernel rather than introducing an accounting-only document lifecycle
7. accounting journals prove the intended document-kernel pattern for later document families: one central document row, one owning payload row, shared lifecycle authority, and module-owned payload truth

### Milestone 3: Inventory ledger foundation

Goal:

1. make stock truth real early

Scope:

1. items
2. inventory locations
3. inventory movement records
4. inventory receipt and issue flows over the shared document kernel
5. item-role and movement-purpose classification so resale stock, service-material usage, and direct-expense consumables are modeled explicitly
6. quantity derivation from movements
7. linkage from inventory movements to work orders, serviced or installed units where relevant, and finance
8. billable versus non-billable material-consumption support at the operational source-record level
9. identity-level traceability for serialized, lot-tracked, or installed equipment classes where the delivery use case requires it

Exit criteria:

1. on-hand quantity is derived, not stored as truth
2. inventory movements are append-only
3. movement source and destination are explicit
4. stock-affecting operations are auditable and idempotent
5. the system can distinguish resale flows from service-consumption flows without needing a second inventory model
6. service-material usage can be traced into execution context and accounting outcomes
7. inventory flows consume the shared document kernel rather than introducing a parallel inventory-local lifecycle

### Milestone 4: Execution context

Goal:

1. establish the operational layer around work

Scope:

1. work orders
2. tasks
3. worker assignment and ownership rules
4. labor time-entry and labor-cost capture baseline
5. execution status history
6. links to documents, inventory usage, labor usage, and financial outcomes
7. explicit serviced-asset and installed-unit ownership model and linkage rules
8. work-order modeling rule implemented consistently:
   - `work_order` remains the primary execution record
   - when a work order participates in shared document lifecycle, it does so through the shared document kernel rather than a second competing lifecycle authority

Exit criteria:

1. work can be requested and tracked through work orders
2. tasks have one clear accountable owner
3. worker-linked labor capture supports credible work-order costing without payroll coupling
4. execution records link cleanly to documents and ledger effects
5. service-company work can retain labor, material-consumption, and maintained-unit context without notes-only reconstruction
6. work-order execution truth and any shared document lifecycle participation do not compete for authority

### Milestone 5: Document workflows

Goal:

1. make AI-driven business operation practical

Scope:

1. invoice drafts and posting flow
2. payment or receipt capture and posting flow
3. inventory receipt and issue document flows
4. work-order document creation and lifecycle
5. approval queue and document review surfaces
6. GST/TDS-aware invoice and payment review paths at foundation depth
7. shared document identity, lifecycle, and numbering rules applied consistently across supported document families

Exit criteria:

1. humans can review AI-created draft documents
2. approved documents post through centralized services
3. posted effects are visible in ledgers and audit

### Milestone 6: Reports and operator review

Goal:

1. make the thin system usable by humans without adding broad form-heavy UI scope

Scope:

1. approval queue
2. invoice and receipt lists
3. journal and ledger views
4. inventory stock and movement views
5. work-order queue and status views
6. audit lookup views
7. GST/TDS summary and review views at foundation depth

Exit criteria:

1. humans can inspect current truth without raw database access
2. reports reconcile to source documents and ledgers
3. v1 has a credible operator-facing review surface

## 4. Current implementation implications

The current codebase suggests these practical implications:

1. `identityaccess`, `ai`, `workflow`, `documents`, and `accounting` can be reused as v1 foundations
2. current CRM breadth should stop expanding for now
3. the first shared document-kernel slice now exists through canonical `documents` records plus accounting-journal linkage, so the next implementation focus should be approval ownership realignment, broader document-family adoption, inventory, work-order, and reporting depth
4. current CRM records should be treated as support context unless they are required by a foundation-led workflow
5. the existing `agent_runs`, `agent_recommendations`, and capability-based tool-policy shape provide a usable base for coordinator and specialist agent orchestration rather than requiring a second AI foundation

## 5. What not to do during this plan

1. do not continue broadening CRM while calling the product thin v1
2. do not add large project-management scope
3. do not design customer portal or broad web navigation before core document and ledger flows are stable
4. do not let future v2 expansion cases distort v1 milestone acceptance

## 6. Refactor success condition

This refactor is successful when the planning set makes it hard to accidentally build the wrong product.

The wrong product is:

1. a broad CRM-first app
2. a form-heavy manual-entry app
3. a workflow suite without ledger truth

The right product is:

1. an AI-operated document system
2. with strong financial, inventory, and foundational tax truth
3. tied to real execution context
4. exposed through review, approval, and report surfaces
5. instrumented so agent behavior can be observed and improved
