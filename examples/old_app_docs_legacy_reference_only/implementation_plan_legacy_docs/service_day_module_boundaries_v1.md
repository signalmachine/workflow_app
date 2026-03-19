# service_day Module Boundaries v1

Date: 2026-03-11
Status: Legacy reference
Purpose: preserve the broader pre-thin-v1 boundary model for historical context and older implementation detail.

Legacy note:
1. active thin-v1 ownership guidance now lives in `plan_docs/service_day_schema_and_module_boundaries_v1.md`
2. use this file only when a specific older boundary decision needs clarification

## 1. Boundary Goals

This document exists to prevent early module sprawl and accidental coupling.

The boundary goals are:
1. keep business ownership clear
2. stop cross-module table reach-through from becoming the default implementation style
3. support a modular monolith that can grow in complexity without becoming a tangled codebase
4. preserve clean seams for CRM, execution, workforce, billing, accounting, AI, and portal expansion
5. keep the modules tightly integrated at the product-flow level so CRM, projects, `work_orders`, billing, and accounting behave like one system rather than separate apps

## 2. Architecture Stance

The application should be a modular monolith with:
1. one deployable application
2. one PostgreSQL database
3. explicit module ownership of tables and services
4. direct SQL inside owning modules
5. cross-module interaction through application contracts, domain services, and explicit read models

Rules:
1. A module owns its tables, write paths, and invariants.
2. Other modules should not write directly into another module's tables.
3. Cross-module reads should prefer:
   - exported queries/read models
   - service contracts
   - derived reporting views
4. Shared infrastructure does not justify shared business ownership.
5. Tight module integration is required at the contract level even when table ownership stays separate.

## 3. Core Modules

These are the recommended first-class bounded contexts.

1. `identity_access`
2. `crm`
3. `projects`
4. `work_orders`
5. `workforce`
6. `payroll` later
7. `inventory_ops`
8. `service_assets`
9. `billing`
10. `accounting`
11. `tax`
12. `ai`
13. `attachments`
14. `workflow`
15. `reporting`
16. `notifications`

## 4. Module Responsibilities

### 4.1 `identity_access`

Owns:
1. users
2. roles
3. user-org memberships
4. authentication and authorization rules
5. internal actor identity
6. device-scoped sessions, refresh-token lifecycle, and session revocation rules for mobile and other external clients
7. external authentication principals for later portal or delegated customer access
8. portal memberships and delegated-access linkage between an external principal and the customer-facing records they are allowed to act through

Does not own:
1. worker operational identity
2. customer/external party identity
3. CRM account, contact, or portal-visible business truth

Exports:
1. current actor identity and org context
2. authorization checks
3. role and membership lookups
4. client-auth/session introspection contracts for mobile and portal surfaces
5. portal and delegated-access authorization checks over the same domain-service boundaries used by internal clients

### 4.2 `crm`

Owns:
1. accounts
2. contacts
3. leads
4. opportunities
5. activities
6. communications
7. CRM task semantics and task-to-CRM linking rules
8. estimates
9. CRM-facing timeline query surfaces and projection rules

Does not own:
1. work execution state
2. workforce assignment
3. invoice posting
4. shared task orchestration infrastructure

Exports:
1. account/contact/opportunity lookup contracts
2. estimate lifecycle contracts
3. customer relationship timeline/query surfaces
4. conversion flows into project/work-order initiation

Integration rule:
1. `crm` should enrich and hand off work, but it should not become the product's center of gravity once execution begins.

### 4.3 `projects`

Owns:
1. projects
2. project membership
3. project-level delivery milestones if treated as delivery planning
4. project roll-up state

Does not own:
1. the primary execution unit
2. worker identity
3. invoice ledger effects

Exports:
1. project lookup and status
2. project membership/ownership views
3. roll-up context for work orders, costing, and billing

Interpretation rule:
1. `projects` owns engagement-level coordination and aggregation
2. `projects` does not own the concrete execution truth for labor, materials, scheduling, or completion

### 4.4 `work_orders`

Owns:
1. work orders
2. work-order status transitions
3. work-order assignments if assignment is execution-specific
4. work-order lines/scope details
5. work-order operational costing roll-ups
6. work-order readiness for billing

Does not own:
1. worker master data
2. invoice issuance
3. financial posting

Exports:
1. work-order lifecycle services
2. execution status and completion surfaces
3. billable readiness signals
4. operational-cost context for reporting and billing preparation

Interpretation rule:
1. `work_orders` owns the concrete executable instruction
2. `work_orders` is where assignment, scheduling, execution state, and cost capture belong
3. a `work_order` may exist without any `project`
4. if a project exists, it coordinates one or more work orders rather than replacing them
5. `work_orders` should be treated as the strongest operating capability in the product, with other modules integrating around it cleanly
6. when execution targets a specific maintainable unit, `work_orders` should reference a serviced asset rather than collapsing that identity into generic customer notes

### 4.5 `workforce`

Owns:
1. workers
2. worker skills and roles
3. worker cost-rate defaults
4. time entries
5. timesheets and approval flow
6. worker-utilization foundations later

Does not own:
1. auth login identity
2. work-order lifecycle
3. payroll

Exports:
1. worker lookup services
2. assignment eligibility and skill views
3. time-entry and timesheet services
4. labor-cost inputs for costing and billing

Payroll-readiness rule:
1. `workforce` must preserve clean seams for a later payroll layer
2. worker identity and time capture should be reusable by later India and UAE payroll flows without collapsing those concerns into the workforce module itself

### 4.5a `payroll` later

Owns later:
1. payroll profiles and country-specific payroll settings
2. compensation components and payroll-period configuration
3. payroll runs, payroll calculations, and payslip outputs
4. payroll payment records and payroll-specific compliance outputs

Does not own:
1. worker master identity
2. raw time-entry capture
3. work-order lifecycle

Exports later:
1. payroll run services
2. payroll summaries and payslip views
3. payroll journal or payment-preparation outputs where needed

Scope rule:
1. payroll should remain SMB-focused and practical rather than becoming a full enterprise HR suite
2. India payroll should be the first delivered payroll baseline
3. UAE payroll should follow on the same extensible contracts rather than through a second disconnected payroll model

### 4.6 `inventory_ops`

Owns:
1. items
2. item locations
3. material receipts
4. material issues
5. work-order material usage
6. project material allocations
7. service parts usage
8. work-order or delivery reservations where inventory is committed before issue
9. serialized equipment units or item-instance tracking where identity-level traceability matters
10. installed customer-site asset linkage for delivered equipment or tracked service parts
11. item-level fulfillment facts needed when the business also sells stocked items directly

Does not own:
1. warehouse-heavy procurement workflows
2. general trading inventory logic
3. broad vendor purchasing approvals or replenishment planning as a v1 requirement

Exports:
1. material availability and usage views
2. work-order/project material consumption data
3. material-cost inputs for costing
4. reservation, issued-not-installed, and installed-equipment visibility for delivery workflows
5. item availability and fulfillment inputs for direct-sale commercial flows

Interpretation rule:
1. `inventory_ops` must be strong enough for equipment-backed service delivery, including sourced gear that is received, reserved, deployed, and linked to customer-site execution context
2. `inventory_ops` should not expand into a procurement-first ERP center of gravity just because vendor-sourced equipment is common in service projects
3. the same module should still be usable by small trading companies at limited depth and remain extensible if broader inventory scope is later approved
4. if a later limited inventory-trading extension is added, `inventory_ops` should remain the owner of item, receipt, stock-movement, reservation, adjustment, and fulfillment facts rather than splitting stock truth into a second trading-specific inventory layer
5. if a later small marketplace-seller extension is added, `inventory_ops` should remain the owner of item, receipt, stock-movement, and fulfillment facts rather than duplicating those concerns in a commerce-specific stock layer

### 4.6a Optional later `inventory_sales`

Owns later:
1. bounded direct-sale order, fulfillment, and customer-return records for stock-backed sales outside service-delivery-only workflows
2. customer-facing stock-sale state transitions that sit between shared inventory ownership and shared billing/accounting posting
3. trading-oriented operational summaries such as sell-through, fulfillment status, and return reasons at limited depth

Does not own:
1. item master truth
2. stock balances or stock-movement truth
3. invoice posting truth
4. broad trading-company pricing, distribution, or warehouse-optimization logic

Exports later:
1. direct-sale fulfillment demand back to `inventory_ops`
2. source-document facts for billing, tax, and accounting integration
3. trading-oriented operational and margin inputs for reporting

### 4.6b Optional later `marketplace_sales`

Owns:
1. bounded channel-order records for later small-seller flows such as Amazon India
2. channel return, refund, fee, and settlement capture at limited depth
3. channel-account and marketplace reference metadata needed to reconcile operational and financial outcomes

Does not own:
1. item master truth
2. stock balances or stock-movement truth
3. broad listing-management, advertising, or marketplace-growth tooling
4. accounting posting truth

Exports:
1. channel-order and return facts for reporting and reconciliation
2. settlement and fee source data for billing/accounting integration
3. product and channel margin inputs derived from shared inventory and accounting records

Interpretation rule:
1. `marketplace_sales` is a later narrow extension for limited small-seller workflows, not a new product center of gravity
2. item, stock, and fulfillment ownership should stay with `inventory_ops`, while billing, tax, and accounting truth remain with their existing owners
3. if later enabled, this module should stay intentionally small and must not pull the product into a marketplace-first or trading-first architecture

### 4.6a `service_assets`

Owns:
1. serviced assets such as customer vehicles, installed field equipment, and customer-owned maintainable machines
2. asset identity and classification
3. service-history continuity for the maintainable unit
4. asset-specific operational readings or identifiers such as VIN, registration, serial number, engine/chassis number, meter, or site tag where relevant
5. links from work performed back to the serviced unit

Does not own:
1. customer relationship ownership
2. stock balances or material movement
3. invoice posting or accounting truth

Exports:
1. serviced-asset lookup and identity surfaces
2. asset history surfaces for delivery, support, and repeat-service workflows
3. asset-context queries for work orders, diagnostics, installed parts, and maintenance history
4. generic asset typing that can represent vehicles, site-installed equipment, or customer-owned machines without creating separate top-level modules per trade

Interpretation rule:
1. the serviced asset is the thing being worked on; it is distinct from the parts consumed and distinct from the customer account that owns it
2. `service_assets` should stay generic enough to support vehicles, installed equipment, and customer-owned machines without forcing a vehicle-only architecture

### 4.7 `billing`

Owns:
1. invoices
2. invoice lines
3. receipts
4. receipt allocations
5. billing milestones when they are commercial rather than delivery-planning artifacts
6. retention and certification billing controls

Does not own:
1. accounting ledger invariants
2. tax engine ownership
3. customer relationship ownership

Exports:
1. invoice issuance services
2. receipt application services
3. receivable-facing document views
4. billing-source linkage to estimates, projects, work orders, and time/cost accumulations

### 4.8 `accounting`

Owns:
1. ledger accounts
2. journal entries
3. journal lines
4. posting services and posting status
5. reversal/correction flows
6. receivable/payable accounting truth
7. submission-review-post lifecycle rules for accounting entry finalization
8. accounting periods, accounting document types, numbering rules, credit/debit-note behavior, and reversing-journal rules

Does not own:
1. invoice lifecycle
2. operational execution
3. worker tracking

Exports:
1. posting interfaces for billing and later operational modules
2. account-balance and ledger inquiry services
3. accounting invariants used by tax and reporting
4. finance review-queue surfaces for submitted-but-not-posted entries where policy requires them
5. period-state and document-sequence rules consumed by billing and later finance flows

### 4.9 `tax`

Owns:
1. tax profiles
2. tax determination support
3. GST Lite logic later
4. TDS logic later
5. country-specific tax extensions
6. service-company-suitable tax rules for billing/accounting flows in the supported baseline countries

Does not own:
1. invoice issuance
2. ledger posting

Exports:
1. tax-calculation and tax-profile services
2. tax metadata for documents and accounting
3. GST/TDS metadata and determination outputs for billing and accounting flows

### 4.10 `ai`

Owns:
1. agent runs
2. agent run steps
3. agent artifacts
4. tool policies
5. approvals
6. persisted recommendations

Does not own:
1. hidden business writes
2. alternative versions of core domain truth

Exports:
1. bounded AI execution services
2. summarization and drafting services
3. recommendation/proposal flows
4. approval orchestration hooks

Speech-capture rule:
1. `ai` may produce transcript artifacts, normalized utterance interpretations, and downstream action recommendations from approved transcript text
2. `ai` does not own raw audio storage policy, mobile command idempotency, or the resulting business-state mutation

### 4.11 `attachments`

Owns:
1. attachment metadata
2. attachment links
3. storage indirection and attachment access rules
4. bounded upload/download transport contracts for mobile, portal, and internal clients

Does not own:
1. the business meaning of the records the attachment is linked to

Exports:
1. upload/link/unlink services
2. attachment resolution services for other modules
3. presigned-upload, download-resolution, or equivalent file-transfer contracts for client apps

Speech-capture rule:
1. if the product retains raw speech audio for audit, support, or replay review, that binary payload should travel through `attachments`-owned transport/storage contracts rather than a hidden provider-specific side path
2. retained audio remains supporting evidence only; it is not the canonical business command

### 4.12 `workflow`

Owns:
1. idempotency boundaries
2. approval orchestration primitives
3. background-job orchestration primitives
4. shared task orchestration
5. task context-linking rules and lifecycle primitives
6. client-safe retry and command orchestration rules where idempotency is required across unstable networks
7. person-vs-team ownership semantics for shared tasks
8. shared task reminder, aging, reassignment, and queue primitives
9. workflow-facing secondary related-link contracts for cross-context visibility where approved by owning domains

Does not own:
1. business state that belongs to domain modules
2. the factual activity history owned by CRM or other operational modules

Exports:
1. workflow coordination primitives
2. approval infrastructure
3. reusable task/command orchestration support
4. idempotent command boundaries for mobile and portal writes
5. shared queue, assignment, reminder, and ownership services for domain modules that use tasks

Speech-capture rule:
1. `workflow` should own the idempotent approved-transcript submission boundary so repeated mobile sends cannot create duplicate business effects
2. domain modules still own the meaning of the approved transcript after interpretation, just as they own typed-command meaning today

### 4.13 `reporting`

Owns:
1. derived read models
2. reporting projections
3. dashboards and analytics assembly
4. analytic-dimension-aware management views and efficiency reporting
5. analytic-dimension configuration records and typed dimension definitions used across operational and accounting reporting
6. later cross-module launcher/search read-model assembly where the client needs a derived home or discovery surface that should not query raw transactional truth directly

Does not own:
1. transactional truth
2. the business records that carry operational or financial dimension context
3. permission rules or business-state ownership for the activities and records it helps surface

Exports:
1. KPI and reporting queries
2. derived sync/read surfaces where a mobile or portal client should not query transactional truth directly
3. profitability, utilization, and business-category analysis views across operational and accounting data
4. typed analytic-dimension lookup and validation contracts consumed by operational and accounting modules when they need controlled dimension references
5. later cross-module launcher/search projections that assemble user-facing activity and discovery read surfaces from module-owned contracts
6. later export-friendly read surfaces where transactional truth is too raw, too expensive, or too unstable as a direct client-facing extract source

### 4.14 `notifications`

Owns:
1. push-notification registration state
2. device tokens and delivery preferences
3. notification event fan-out policies
4. user-visible notification feed or delivery bookkeeping if later required

Does not own:
1. the business events that trigger notifications
2. mobile-only copies of CRM, workflow, or approval state

Exports:
1. device-registration services
2. notification dispatch contracts for workflow, CRM, AI approvals, and later billing flows
3. read models for user notification visibility where needed
4. cross-module derived views

## 5. Recommended Dependency Direction

The preferred dependency direction is inward toward foundational modules and outward through explicit contracts.

Recommended pattern:
1. `identity_access`, `attachments`, and `workflow` are foundational cross-cutting modules
2. `crm`, `projects`, `work_orders`, `workforce`, `inventory_ops`, and `service_assets` are operational domain modules
3. `billing`, `accounting`, and `tax` are commercial/financial modules
4. `ai` and `reporting` sit as orchestration and derived-consumption layers over the domain model

## 6. Allowed Dependency Rules

These rules should guide imports, service references, and write ownership.

### 6.1 Foundational Rules

1. Most modules may depend on `identity_access` for actor/org context.
2. Most modules may depend on `attachments` via contracts.
3. Most modules may depend on `workflow` primitives, but not surrender business-state ownership to `workflow`.
4. Portal-facing client identity should reuse `identity_access` rather than introduce a second external-auth stack inside `crm` or a portal-only module.
5. Later launch/home/search read surfaces should consume module-owned contracts, read models, and authorization checks rather than bypassing bounded-context ownership through ad hoc table reach-through.
6. Later import/export workflows should consume module-owned write and read contracts rather than bypassing bounded-context ownership through bulk table loaders or unowned extract logic.

### 6.2 CRM Rules

1. `crm` may call `attachments`, `workflow`, and `ai`.
2. `crm` may initiate creation flows into `projects` and `work_orders` through explicit application services.
3. `crm` should not write directly into `billing` or `accounting` tables.

### 6.3 Project and Work-Order Rules

1. `projects` and `work_orders` may reference CRM context through stable foreign keys and read contracts.
2. `work_orders` may depend on `workforce` for assignment and time/cost inputs through explicit contracts.
3. `work_orders` may depend on `inventory_ops` for material usage flows.
4. `work_orders` may depend on `service_assets` when execution targets a specific maintainable unit.
5. `projects` should not own work-order internals.
6. `work_orders` should not own worker master data.

### 6.3b Marketplace-sales Rules

1. `marketplace_sales` may depend on `inventory_ops` for item and fulfillment facts through explicit contracts.
2. `marketplace_sales` may depend on `billing`, `tax`, and `accounting` through explicit posting or reconciliation services.
3. `marketplace_sales` should not own item master data, stock movement, or ledger posting truth.
4. `crm`, `projects`, and `work_orders` should remain independent of `marketplace_sales` unless a later deliberate cross-flow is justified.

### 6.3a Service-asset Rules

1. `service_assets` may reference `crm` ownership context such as account and site through stable foreign keys and read contracts.
2. `service_assets` may surface linked work history from `work_orders` and linked installed-part history from `inventory_ops` through queries or derived views.
3. `service_assets` should not own stock movement, billing decisions, or accounting truth.

### 6.4 Workforce Rules

1. `workforce` may reference `identity_access` but should not collapse worker identity into user identity.
2. `workforce` may attach time entries to `projects`, `work_orders`, or tasks through owned association logic.
3. `workforce` should not own billing or accounting decisions.

### 6.5 Billing, Accounting, and Tax Rules

1. `billing` may consume source context from `crm`, `projects`, `work_orders`, and `workforce`.
2. `billing` posts into `accounting` through explicit posting services only.
3. `tax` supports `billing` and `accounting` but should not become the owner of document lifecycle.
4. `accounting` should not reach back and mutate operational module state.

### 6.6 AI and Reporting Rules

1. `ai` may read domain context through approved tool contracts and read services.
2. `ai` must never bypass owning modules when causing writes.
3. `reporting` may read from all modules through read models, projections, or explicit queries.
4. `reporting` must not become a backdoor write path.
5. operational and accounting modules may carry references to reporting-owned analytic-dimension definitions, but they remain the owners of the underlying business facts those dimensions describe.

## 7. Shared Concepts and Ownership Decisions

Some concepts span many modules. Their ownership must still be clear.

### 7.1 Tasks

Recommended direction:
1. one shared task engine is the v1 default
2. ownership should live in `workflow`
3. context-specific meaning should remain in the owning domain
4. a task should have one primary actionable owner: either one person or one team queue
5. a task should have one primary business context, but may also expose secondary related links for project, work order, account, site, or serviced-asset visibility
6. task/activity analytics should be derived from this shared model rather than reconstructed separately in each domain

Practical rule:
1. a task may reference CRM, project, or work-order context
2. task completion must not directly mutate owning business state unless the owning module authorizes it
3. domain modules may define task templates, statuses, and completion side effects, but not bypass the shared task service
4. do not treat many simultaneous person assignees as the default ownership shape; use team queues, watchers, or linked participants instead

### 7.2 Timeline

Recommended direction:
1. timeline is a derived or append-oriented CRM-facing concept
2. timeline should aggregate events from domain modules
3. timeline is not the owner of the underlying business records
4. activities remain factual history, while tasks remain accountable next actions even when both appear in the same timeline

### 7.3 Assignments

Recommended direction:
1. `workforce` owns worker identity
2. assignment records can live in the domain where the assignment has business meaning
3. for v1, `project_members` belongs to `projects`
4. `work_order_assignments` belongs to `work_orders`

### 7.4 Costing

Recommended direction:
1. raw cost inputs remain in owning modules
2. `workforce` owns labor inputs
3. `inventory_ops` owns material inputs
4. `work_orders` may own operational roll-up cost views
5. `accounting` owns financial truth, not operational costing truth

### 7.5 Analytics dimensions

Recommended direction:
1. prefer typed analytic dimensions over a cost-center-only design
2. support `cost_center` as one optional built-in dimension type for businesses that want it
3. `reporting` should own the dimension-definition catalog and other low-write configuration records for analytic dimensions
4. operational and accounting modules should carry dimension references only where those references materially improve reporting or financial analysis
5. let `reporting` consume dimensions from operational and accounting contexts through derived models rather than inventing parallel transactional truth
6. do not force businesses to encode every analytical need into the chart of accounts

## 8. Cross-Module Interaction Patterns

Use a small number of explicit patterns.

### 8.1 Command Pattern

Use when one module requests a business action in another module.

Examples:
1. `crm` requests conversion of an accepted estimate into a project/work-order setup flow
2. `billing` requests posting of an invoice into `accounting`

### 8.2 Query Pattern

Use when one module needs read-only data from another module.

Examples:
1. `work_orders` fetches customer summary from `crm`
2. `billing` fetches work-order completion/billable summary from `work_orders`
3. `work_orders` fetches serviced-asset identity or history from `service_assets`

### 8.3 Event/Projection Pattern

Use when derived views or timelines need to assemble cross-domain information.

Examples:
1. `reporting` builds profitability views
2. `crm` timeline shows activity from execution and billing modules
3. `service_assets` history surfaces summarize linked work-order and installed-part changes

## 9. Anti-Patterns to Avoid

1. letting `billing` directly update work-order completion state
2. letting `accounting` own invoice lifecycle
3. letting `ai` write directly to domain tables outside module services
4. treating `reporting` tables as a source of transactional truth
5. copying customer, worker, or project data into many modules without clear ownership
6. letting `workflow` become a generic dumping ground for business state
7. creating a generic "operations" module that weakens clear domain ownership

## 10. Suggested Package-Level Layout

The exact repo structure may evolve, but the initial shape should mirror the module boundaries:

1. `internal/identityaccess`
2. `internal/crm`
3. `internal/projects`
4. `internal/workorders`
5. `internal/workforce`
6. `internal/inventoryops`
7. `internal/serviceassets`
8. `internal/billing`
9. `internal/accounting`
10. `internal/tax`
11. `internal/ai`
12. `internal/attachments`
13. `internal/workflow`
14. `internal/reporting`

Each module should prefer:
1. domain/service contracts
2. postgres repository layer owned by the module
3. authorization wrapper where needed
4. tests focused on module invariants and cross-module contract behavior

## 11. Immediate Recommendation

Start implementation with these module ownership assumptions:

1. `crm` owns sales and relationship truth
2. `projects` owns engagement grouping
3. `work_orders` owns execution truth
4. `workforce` owns labor identity and time capture
5. `billing` owns commercial collection documents
6. `accounting` owns ledger truth

If these six lines remain stable, the system can grow materially in complexity without losing architectural coherence.
