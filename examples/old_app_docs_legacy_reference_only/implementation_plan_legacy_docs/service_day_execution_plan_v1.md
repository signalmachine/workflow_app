# service_day Execution Plan v1

Date: 2026-03-11
Status: Legacy reference
Purpose: preserve the broader pre-thin-v1 execution plan for historical context and implementation back-reference.

Legacy note:
1. this milestone model is not the active canonical thin-v1 execution plan
2. use `plan_docs/service_day_refactor_execution_plan_v1.md` and `plan_docs/service_day_refactor_tracker_v1.md` for current planning and status

## 1. Execution Intent

This document turns the initial `service_day` planning direction into a pragmatic implementation path for a new codebase.

The execution goals are:
1. reach a bootable system quickly without weakening data-model foundations
2. deliver a usable internal CRM before broader service-operations depth, while keeping `work_order` support positioned as the eventual strongest product capability
3. avoid early schema choices that would later force redesign in projects, work orders, workforce, billing, or accounting
4. treat workforce, time capture, and costing inputs as first-class foundations
5. keep AI useful from the first runnable slice while keeping agent writes bounded and auditable
6. keep UAE as the real commercial target market while shipping India-first compliance and payroll sequencing without forcing a later redesign
7. make explicit which decisions must be made now and which can safely be deferred

## 2. Execution Principles

1. Vertical slices beat broad but shallow module scaffolding.
2. Foundational schema decisions should be made before workflow-heavy features, not after.
3. A runnable CRM is the first product goal, but CRM cannot be isolated from future delivery, workforce, billing, and the central `work_order` operating model.
4. Accounting kernel work must define invariants early even if the full accounting surface arrives later.
5. Workforce and time-capture models should be defined before operational costing grows around them.
6. AI must begin with assistive workflows that fit naturally into human-owned operations.
7. Portal readiness is a schema and authorization concern now, not a UI concern later.
8. Country-specific tax and payroll depth should arrive in waves, but the core worker, organization, accounting, and document model must stay country-extensible from the start.
9. India-first GST and TDS support should be designed as service-company-ready operational tax/accounting flows, not as isolated report-only add-ons.
10. The system should be able to measure business efficiency across multiple business categories, so reporting and analytics dimensions must be modeled intentionally rather than improvised in downstream dashboards.
11. Task and activity management should be treated as a first-class cross-product efficiency capability rather than as a narrow CRM convenience or a generic standalone app detached from domain meaning.
12. Activities should represent factual history, while tasks should represent accountable next actions with ownership, due-state, and queue visibility.
13. Mobile and web or portal clients are the intended product interaction surfaces; any CLI-style tooling should remain narrow internal support tooling rather than a competing interface contract.
14. As the web client grows beyond CRM, the default launch model should favor direct workflow or document entry plus global search rather than broad module landing pages.
15. Later spreadsheet import/export should be prepared through explicit exchange seams, auditable batch boundaries, and reusable file transport rather than retrofitted through direct table-loading utilities.

## 3. Decide Now vs Defer Safely

### 3.1 Decide Now

These decisions should be resolved before schema-foundation work is treated as complete.

1. Accounting kernel scope
   - Define the exact invariants that exist before billing starts:
   - chart of accounts model
   - journal entry and journal line structure
   - posting status model, including pre-post review states
   - receivable/payable control-account strategy
   - idempotent posting boundary from operational documents into accounting
   - balanced double-entry enforcement and reversal/correction rules
   - whether the org requires separate submit-versus-post control for accounting entry finalization
   - document-type, document-number, accounting-period, and reversing-entry rules
2. Commercial document strategy
   - Decide whether estimates and invoices share a common document-core abstraction from day one.
   - Recommended direction: shared commercial document concepts, separate domain aggregates where behavior diverges.
3. Execution object strategy
   - The execution relationship is fixed as:
   - `project` is the long-lived commercial and delivery umbrella when grouping is needed
   - `work_order` is the single primary execution record across service businesses
   - `service_request` is optional intake terminology that can map into `work_order`
4. Communication system-of-record strategy
   - Decide what the first communication truth is:
   - manual activity and note logging only
   - manual plus imported email records
   - fully synchronized channel integrations
   - Recommended direction: start with manual records plus a generic communication/event model that can absorb imports later.
5. Task model strategy
   - Use one shared task engine with context-specific wrappers and permissions.
   - Implementation default: task orchestration lives in `workflow`; business meaning remains with the owning domain.
   - Support assignment to either one person or one team queue as the primary actionable owner.
   - Keep one primary business context for ownership and automation, while allowing secondary related links for reporting and visibility such as account, site, project, work order, or serviced asset.
6. Activity model strategy
   - Keep activities separate from tasks.
   - Activities capture what happened; tasks capture the next accountable follow-up.
   - Allow activities to link to the same business contexts as tasks, but do not make activity records the owner of workflow state.
7. Workforce model strategy
   - Recommended direction:
   - `worker` is the primary operational contributor record
   - `user` login remains separate
   - `time_entry` is the atomic time fact
   - `timesheet` is an approval/review wrapper added after raw time capture
8. Launch experience strategy
   - Use an activity-centered launch model for the later web client.
   - Personalized pinned tiles should open exact workflow or document screens directly.
   - Global search should expose both actions and records across modules.
   - Recommended direction: do not make bounded contexts or module landing pages the primary user-navigation primitive.
9. Data exchange strategy
   - Keep CSV as the minimum interoperable exchange format and allow later spreadsheet support where it materially helps operators.
   - Imports should run through explicit bounded workflows, not direct table mutation.
   - Exports should come from stable domain or read-model contracts rather than ad hoc frontend assembly.

### 3.2 Defer Safely

These can remain intentionally incomplete in the first execution wave.

1. UAE statutory depth beyond generic tax/profile readiness
2. customer portal UX and login flows beyond schema and auth foundations
3. advanced quote-to-project automation
4. autonomous agent workflows beyond bounded preparation and recommendation
5. field-service dispatch optimization
6. procurement depth beyond what is needed to support controlled material receipts and issues
7. payroll depth beyond worker and operational costing needs
8. full HR-suite depth beyond lightweight SMB payroll needs
9. rental-operations workflows beyond keeping the commercial, accounting, site/unit, and recurring-charge foundations extensible enough to support them later
10. inventory-trading workflow depth beyond keeping the item, inventory, billing, tax, and accounting foundations extensible enough to support a later limited small-trading extension
11. marketplace-seller workflow depth beyond keeping the item, inventory, billing, tax, and accounting foundations extensible enough to support a later narrow small-seller extension
12. broad end-user documentation breadth beyond keeping `docs/user_guides/` ready and treating user guides as a normal implementation deliverable for shipped workflows
13. full spreadsheet import/export breadth beyond keeping the current foundation ready for later explicit exchange workflows

Inventory/materials interpretation:
1. deferring full procurement depth does not mean omitting source-cost, supplier-reference, reservation, serialized-equipment, or installed-base requirements that materially affect delivery operations
2. Milestone C should be strong enough for equipment-backed service delivery even if replenishment planning and broad warehouse optimization remain out of scope
3. the same foundation should also support light direct item sales and be usable by small trading companies at limited depth without requiring a second inventory architecture
4. deeper trading-company inventory scope should remain an intentional later expansion rather than a v1 optimization target

Inventory-trading interpretation:
1. if a later limited trading-oriented extension is added, it should reuse the same item, receipt, stock-movement, billing, tax, and accounting foundations rather than introducing a parallel sales-order or stock-ledger stack
2. early foundations should therefore preserve direct-sale fulfillment seams, customer-return-safe stock movements, margin-friendly cost traceability, and inventory-to-ledger posting boundaries even though trading workflows are not a current milestone target
3. advanced warehouse slotting, broad procurement approvals, route-scale distribution, and large-trading-company controls remain explicitly out of scope unless a later deliberate product decision changes that direction

Marketplace-lite interpretation:
1. if a later small Amazon-style seller extension is added, it should be built on the same item, receipt, stock-movement, billing, tax, and accounting foundations rather than on a parallel commerce stack
2. early foundations should therefore preserve supplier/source references, SKU-grade item identity, return-safe stock movements, and fee or settlement-friendly accounting seams even though marketplace flows are not a current milestone target
3. channel-listing management, ads, and broad commerce operations remain explicitly out of scope unless a later deliberate product decision changes that direction

## 4. Recommended Domain Shape

The following domain stance should guide the first schema documents and service boundaries.

### 4.1 CRM

Core records:
1. accounts
2. contacts
3. leads
4. opportunities
5. activities
6. communications
7. estimates
8. relationship timeline views
9. workflow-linked follow-up visibility and efficiency projections

Rules:
1. CRM is the customer-context layer, not the long-term center of product strength.
2. Every later execution or billing record should retain a path back to account, contact, and opportunity context where applicable.
3. CRM must hand work into `work_order` and delivery flows through explicit, low-friction transitions rather than disconnected module boundaries.
4. Timeline should be a derived CRM-facing read model, aggregating domain events rather than becoming the canonical owner of every business fact.

### 4.2 Delivery

Recommended model:
1. `project` represents a customer-facing engagement, commercial umbrella, or delivery program
2. `project` is the coordination and aggregation layer for larger work, not the default operational execution record
3. `work_order` represents a concrete unit of execution under a project or directly under a customer relationship
4. `work_order` is the operational unit where assignment, scheduling, labor, materials, and completion are tracked
5. one project may group many work orders, but small jobs may use only a work order with no project at all
6. project and work-order execution can both attach to the shared `tasks` engine through typed context links
7. costing should accumulate from time, materials, expenses, and external service inputs
8. `service_request` is not a separate long-lived core aggregate in v1; it is intake language that can map into a `work_order` state or a small intake record if needed
9. when execution targets a specific customer-owned or installed unit, `work_order` should retain a clean link to that serviced asset rather than only to the customer account

Practical rule:
1. use `project` for planning, coordination, roll-up visibility, and commercial grouping
2. use `work_order` for the specific executable instruction that a team performs
3. treat the quality, speed, and auditability of `work_order` flows as a primary product differentiator rather than a secondary downstream module

### 4.3 Workforce and Time

Recommended model:
1. `worker` is the primary operational contributor record, whether employee or contractor
2. `worker` may exist without a login `user`
3. `time_entry` is the atomic unit of logged time
4. `timesheet` is the approval/review wrapper over time entries and can land after raw time capture
5. assignments should point to workers, not only users

Payroll stance:
1. payroll is a later bounded capability, not part of the first delivery foundation
2. the worker and time model must still be compatible with later payroll support
3. target payroll depth is SMB-oriented and operationally useful, not a full enterprise HR/payroll suite
4. when payroll begins, India should land first and UAE should follow on the same extensible core

### 4.4 Commercial

Recommended model:
1. estimates and invoices should share common structural conventions
2. milestone billing, retention, and change-order effects should be modeled before construction-style workflows are implemented
3. invoices should not become the place where commercial truth is reconstructed manually

### 4.5 Accounting

Recommended kernel:
1. ledger accounts
2. journal entries and lines
3. posting interface from operational modules
4. receivables treatment
5. audit-safe reversal/correction strategy

Accounting quality rule:
1. the accounting module must be a solid double-entry ledger from the beginning, even if the first user-facing accounting surface is narrower than the operations surface
2. operational convenience must not bypass accounting invariants such as balanced posting, controlled reversals, idempotent posting, and source-document traceability
3. accounting entry preparation and ledger posting should remain distinct steps so reviewable submission queues fit naturally where required
4. standard accounting controls such as document typing, durable numbering, credit/debit-note handling, reversing journals, and accounting-period enforcement should be first-class accounting concerns rather than later bolt-ons
5. a minimal accounting service shell that validates balanced draft entries and performs explicit posting is enough for Milestone A foundation, but richer submit/review queues, reversal workflows, and period controls remain phased follow-on work

### 4.6 AI

Initial AI should operate on top of the same domain model rather than sidecar copies.

Rules:
1. AI proposals are records, not only transient responses
2. approval points are explicit
3. tool permissions are org-aware and role-aware
4. prompts, tool calls, artifacts, and accepted outputs are persisted for audit and replay safety

## 5. Milestone Path

## Milestone A: Kernel and Schema Foundation

Objective:
Establish the codebase, tenancy, identity, audit, workflow, workforce foundation, accounting kernel invariants, and AI execution shell.

Build scope:
1. repository skeleton and module boundaries
2. migration framework and baseline schema conventions
3. org, users, roles, memberships, and authn/authz foundation
4. audit events and idempotency
5. attachments foundation
6. parties and external principal foundation
7. worker identity and cost-rate foundation
8. accounting kernel tables and posting contracts
9. AI adapter shell, tool-policy shell, and approval model
10. HTTP server, health, and basic admin bootstrap path
11. mobile/backend readiness seams in the foundation layer:
   - schema placeholders and ownership for device sessions and refresh-token lifecycle
   - attachment transport direction chosen early enough to avoid client-specific rework later
   - notification/device-registration ownership decided even if delivery lands later
   - launchable-activity, direct-entry, and global-search seams should remain possible without introducing a second hidden workflow path
   - import/export, file-artifact, and later batch-job seams should remain possible without bypassing domain-service ownership or audit boundaries
   - the first planned mobile client may use Flutter, but backend mobile-readiness work should remain client-technology-agnostic
12. task/activity operating-foundation seams:
   - one shared workflow task model with explicit person-vs-team ownership semantics
   - activity/event modeling kept separate from task state
   - reminder, overdue, and queue-aging concepts left possible without redesign
   - primary business context plus secondary related-link support left possible for later project, work-order, site, and asset visibility

Exit criteria:
1. app boots locally against PostgreSQL
2. tenant-scoped schema rules are enforced consistently
3. bootstrap singleton guarantees hold under concurrent requests, not only serial ones
4. authenticated internal users can access their org safely
5. audit plumbing exists for business-state changes, meaningful writes are atomic with their audit trail, and exemptions for low-level technical writes are explicit
6. accounting kernel invariants are documented in code and schema
7. balanced double-entry posting rules are explicit enough that later billing, tax, and payroll flows do not need accounting redesign
8. submit-versus-post accounting controls and posting-role boundaries are explicit enough that AI-assisted finance workflows do not need redesign later
9. AI execution shell can persist runs, steps, and approvals even if features are minimal
10. the planning and schema foundation no longer leave mobile-critical auth and notification concerns as undocumented afterthoughts
11. the architecture keeps room for a later approval-gated mobile speech-capture flow where spoken business events in local Indian languages can be transcribed, reviewed in the same language, and only then submitted for backend processing
12. the architecture keeps room for first-class task/activity ownership, queue, reminder, and efficiency-analytics flows without forcing later schema rewrites

## Milestone B: CRM Core Vertical Slice

Objective:
Deliver a usable internal CRM for lead-to-opportunity-to-estimate workflows.

Build scope:
1. accounts and contacts
2. leads and conversion flow
3. opportunities and stage management
4. activities and follow-ups
5. notes, attachments, and communication records
6. shared task engine with CRM linking
7. estimate baseline
8. relationship timeline and search baseline
9. AI summaries, drafting, and next-step proposals
10. first internal-mobile-consumable API baseline:
   - pagination and list-contract shape for the first mobile-critical CRM endpoints
   - machine-readable error semantics and compatibility posture for mobile-consumed APIs
   - idempotent write boundaries where early mobile retries would otherwise create duplicate CRM or workflow effects
   - no Flutter-specific shortcut should weaken the generic backend contract needed by any serious mobile client
11. future mobile speech-entry seam identified:
   - speech-to-text and transcript approval remain later work, but API contracts should assume that some mobile-created commands may originate from approved local-language transcripts rather than only typed forms
12. first user-guide baseline:
   - create `docs/user_guides/README.md`
   - start internal-user workflow guides for the actually shipped CRM surface rather than waiting for all later milestones

Exit criteria:
1. users can manage prospect and customer relationships end-to-end
2. opportunity progression is visible and auditable
3. communication and activity history is meaningfully searchable
4. follow-up ownership is visible through person and team-oriented task views rather than hidden inside ad hoc CRM notes
5. task and activity history can be analyzed later without redesign because ownership, due-state, and context-link seams are already sound
6. estimates can be prepared for service-oriented work
7. AI assistance is useful without creating unreviewed writes
8. API failures are classified intentionally and do not leak raw internal error details
9. malformed route/query identifiers and other caller input are rejected at the application boundary instead of falling through to database cast failures
10. accepted CRM actions are reconstructible from audit events plus AI run history where AI participated
11. the first CRM endpoints that a mobile client depends on have a deliberate pagination, error, and retry-safe write contract rather than implicit internal-only behavior
12. CRM records and estimate state keep a technically solid path into later project and `work_order` initiation without lossy handoff seams
13. current mobile API design does not block a later approved-transcript capture flow for Telugu, Kannada, or similar local-language internal business-event submission
14. at least the first shipped internal-user workflows have corresponding user-guide coverage or an explicit documented deferral

## Milestone C: Delivery Operations Vertical Slice

Objective:
Move won work into execution with minimal but coherent project, work-order, and workforce operations.

Build scope:
1. projects
2. project membership and ownership
3. work orders
4. work-order assignments
5. shared tasks in delivery context and project milestones
6. time-entry baseline
7. light time-entry approval baseline
8. expense and material capture baseline
9. work-order costing baseline
10. change-order and billing-milestone baseline
11. equipment allocation, reservation, and installed-base baseline where delivery use cases require identity-level tracking
12. serviced-asset baseline for vehicle, machine, or installed-equipment history where the work targets a specific maintainable unit

Exit criteria:
1. a won opportunity or accepted estimate can become delivery work without data re-entry
2. work ownership, progress, and status are visible
3. early costing is visible at project/work-order level
4. workforce effort is captured cleanly without payroll coupling
5. construction-style and repair-style use cases can both fit the model at baseline depth
6. material and equipment flows support both quantity-based consumables and identity-tracked deployed gear without forcing manual reconstruction from notes
7. serviced assets such as vehicles or installed equipment can retain durable maintenance history and linked installed-part traceability
8. the integrated flow across CRM, projects, `work_orders`, service-linked inventory, and serviced assets feels like one operating system rather than stitched-together modules
9. the same inventory and commercial foundation can support direct stocked-item sales without distorting the delivery-first model
10. delivery work can appear in person or team work queues without losing the primary project/work-order ownership semantics

Construction-trade interpretation:
1. Milestone C should be suitable for subcontract execution such as electrical, plumbing, low-voltage/networking, fit-out, repair, and similar material-backed work where crews perform work orders under a larger project umbrella
2. the plan should support both project-led construction work and direct work-order service jobs without requiring separate inventory architectures

## Milestone D: Billing and Financial Traceability

Objective:
Turn executed work into billable and financially traceable outcomes.

Build scope:
1. invoices and receipts
2. invoice posting into accounting
3. customer balance visibility
4. milestone and progress-billing support
5. retention and certification support where applicable
6. basic revenue and collections reporting
7. finance review-queue or direct-post path by role and policy
8. standard accounting document and period controls baseline
9. recurring-charge and recurring-settlement primitives strong enough that later rent schedules, owner dues, and utility/service recharge flows do not require redesign

Exit criteria:
1. operational work can be billed from structured source records
2. invoice-to-receipt-to-ledger traceability is intact
3. customer balances are credible
4. contract-commercial controls exist for project-led businesses
5. submitted accounting entries can be reviewed and posted by an authorized human, and AI-originated entries never bypass the human posting boundary
6. accounting documents, credit/debit-note flows, period controls, and reversal/reversing-entry behavior are coherent enough that later regional finance depth does not require redesign
7. recurring commercial/accounting patterns are coherent enough that later rental-operator billing and owner-settlement flows can reuse the same finance core

## Milestone E: Compliance and Portal Readiness

Objective:
Add baseline regulatory depth and prepare external-facing workflows, including the backend capabilities required for a first internal mobile client and later external mobile surfaces.

Build scope:
1. GST Lite
2. TDS baseline
3. portal authorization and visibility policies
4. customer portal API foundations
5. mobile/backend readiness completion for internal clients and reuse for later external mobile surfaces:
   - device-scoped session and refresh-token implementation completed on the earlier schema seam
   - stable API versioning and deprecation policy enforced on mobile-consumed endpoints
   - pagination and incremental-sync contracts expanded beyond the first CRM baseline to the broader covered API surface
   - idempotent mobile-write boundaries and retry-safe semantics extended to the production mobile-critical commands
   - attachment upload/download transport strategy implemented
   - notification and device-registration primitives implemented
   - approved-transcript command submission shape defined for local-language speech capture so mobile voice entry remains reviewable, auditable, and retry-safe
6. deeper AI workflow recommendations around billing and collections
7. service-company-suitable tax treatment across billing and accounting flows

Exit criteria:
1. India-first compliance baseline exists for core billing/accounting flows and is suitable for service-company billing patterns
2. external access foundations are present without duplicating CRM truth
3. the codebase is ready for a first controlled portal slice
4. the backend exposes stable enough auth, API, attachment, and notification contracts for a first internal mobile client without relying on ad hoc app-specific workarounds
5. GST and TDS behavior is extensible enough that later deeper tax features do not require redesign of the core accounting and document model
6. the system can later support speech-captured business-event entry from the mobile client, with same-language transcript approval before backend interpretation, without redesigning the mobile or AI boundaries

Reference speech-entry workflow:
1. user speaks a business event in Telugu, Kannada, or another supported local language
2. the mobile client renders recognized text back in the same language where practical
3. the user approves the specific transcript revision
4. the client submits the approved transcript, locale metadata, and idempotency boundary
5. the backend interprets that approved text through normal AI and domain-service boundaries

## Milestone F: Lightweight Payroll Extension

Objective:
Add SMB-focused payroll support on top of the workforce, time, and accounting foundation without turning the product into a full payroll suite.

Build scope:
1. payroll boundary and schema baseline
2. compensation and earning/deduction component baseline
3. payroll periods and payroll runs
4. payslip and payment-record baseline
5. India payroll baseline first
6. UAE payroll adaptation next on the same core contracts

Exit criteria:
1. the worker/time/accounting foundation supports payroll without collapsing worker identity into login identity
2. India payroll can run credibly for small service businesses at lightweight depth
3. UAE payroll can be added without redesigning the earlier payroll core
4. the payroll layer stays intentionally SMB-focused rather than expanding into full-suite HR complexity

## Milestone G: Customer Messaging Channel Extensions

Objective:
Add later customer-facing external messaging support after portal, communication truth, audit, approval, and delivery/billing foundations are already stable.

Build scope:
1. WhatsApp support for customer-facing communication threads, status updates, reminders, and bounded customer interactions
2. inbound and outbound message capture linked to the canonical communication record model rather than a parallel truth store
3. template, approval, and policy controls for customer-facing outbound messages where needed
4. channel-identity linkage between customer/contact records and approved WhatsApp endpoints or equivalents
5. later automation hooks that may draft or queue outbound customer messages only through explicit approval and audit boundaries

Exit criteria:
1. WhatsApp or equivalent customer-channel messages are traceable through the communication system of record rather than hidden sidecar state
2. customer-facing outbound automation remains bounded by explicit approval and audit rules
3. customer/contact linkage, conversation history, and later workflow-triggered updates can use the channel without weakening the canonical CRM and communication model

## Milestone H: Rental Operations Extension

Objective:
Add later support for operator-led rental businesses that lease whole properties or buildings from owners and rent rooms or units onward to occupants while preserving the service-first core of the product.

Build scope:
1. properties, buildings, and rentable-unit records with owner and site linkage
2. occupant or tenant agreements with start/end dates, recurring rent terms, deposits, and status
3. recurring monthly rent charge generation and collection tracking
4. owner settlement or payable tracking for the upstream lease obligation
5. operating-expense capture for utilities, food service, staffing, maintenance, and similar costs
6. profitability and cash-visibility views across property, unit, owner, and occupant dimensions
7. optional service/work-order linkage for maintenance or hospitality-style operations where rental businesses also run operational execution flows

Exit criteria:
1. a rental operator can model owner, property, unit, and occupant relationships without duplicating customer/accounting truth
2. monthly rent receivables, owner obligations, deposits, and operating expenses are traceable through the shared finance core
3. property-level and unit-level margin is visible after owner dues and operating costs
4. rental support reuses the shared billing, accounting, party, and site foundations rather than introducing a disconnected second commercial stack
5. the broader product still remains centered on service and `work_order` execution; rental support is an intentional extension, not a new architectural center of gravity

## Milestone I: Limited Inventory-Trading Extension

Objective:
Add a later limited extension for small inventory-led businesses that buy stock for resale and sell directly to customers, without weakening the service-business core.

Build scope:
1. supplier relationships and bounded receipt or stock-adjustment flows on top of the shared party and inventory foundations
2. direct-sale order or fulfillment capture at limited depth for stocked-item sales outside service-delivery-only workflows
3. customer return handling and stock-safe reversal flows at limited depth
4. accounting and tax linkage for inventory-backed sales, cost of goods sold, returns, and inventory adjustments through the shared finance core
5. reporting visibility for stock position, sell-through, gross margin, and inventory-backed customer balances at limited depth

Exit criteria:
1. a small trading-oriented operator can trace stock from receipt into direct sale, customer return, and shared ledger impact without a second disconnected inventory or finance model
2. direct item sales and returns remain reconcilable through the shared billing, tax, and accounting core rather than ad hoc spreadsheets
3. the extension remains intentionally narrow and does not introduce trading-first architecture that would weaken the service, project, and `work_order` center of gravity

## Milestone J: Narrow Marketplace-Seller Extension

Objective:
Add a later limited extension for small seller workflows such as a business that sources branded goods from a contract manufacturer and sells them through Amazon India, without weakening the service-business core.

Build scope:
1. supplier and contract-manufacturer party relationships tied to the shared party and inventory foundations
2. simple item/SKU and stock receipt flows sufficient for sourced finished goods at limited depth
3. channel-order, return, and settlement capture at limited depth, whether imported or recorded through bounded operational flows
4. fee, refund, and settlement accounting linkage through the shared billing, tax, and accounting core
5. reporting visibility for basic inventory position, channel sales, returns, settlement variance, and product-level margin

Exit criteria:
1. a small seller can trace sourced finished goods from receipt into channel sale, return, and settlement without a second disconnected inventory or ledger model
2. channel fees, refunds, and net settlements are visible through the shared accounting core rather than ad hoc spreadsheets
3. the extension remains intentionally narrow and does not introduce marketplace-first architecture that would weaken the service, project, and `work_order` center of gravity

## 6. Recommended Vertical Slices

These slices are intentionally cross-cutting. Each slice should produce a user-visible outcome and a clean persistence boundary.

### Slice 1: Org-Aware Boot

Outcome:
1. internal user signs in
2. org context is resolved
3. basic authorization works
4. audit events are emitted

Included domains:
1. identity_access
2. workflow
3. audit

### Slice 2: Relationship Workspace

Outcome:
1. user creates account and contact
2. user logs notes and activities
3. timeline becomes useful immediately
4. activities are kept distinct from accountable follow-up tasks

Included domains:
1. crm
2. attachments
3. search

### Slice 3: Lead to Opportunity

Outcome:
1. lead is qualified
2. converted into account/contact/opportunity
3. follow-up tasks are created
4. AI proposes next actions
5. the resulting follow-up can be owned by a person or routed to a team queue

Included domains:
1. crm
2. ai
3. workflow

### Slice 4: Opportunity to Estimate

Outcome:
1. user prepares service estimate
2. estimate lines support service scope or milestone-style quoting
3. AI can draft a first estimate from notes

Included domains:
1. crm
2. billing foundation
3. ai

### Slice 5: Won Work to Project/Work Order

Outcome:
1. accepted commercial intent becomes executable work
2. project and/or work order is created without losing CRM context

Included domains:
1. crm
2. projects
3. work_orders
4. workforce

### Slice 6: Delivery Cost Capture

Outcome:
1. users capture time, materials, and expenses
2. early work-order costing becomes visible

Included domains:
1. projects
2. work_orders
3. workforce
4. inventory_ops
5. accounting kernel

### Slice 7: Bill and Collect

Outcome:
1. invoice is created from delivery/commercial context
2. receipt is recorded
3. receivable position is visible

Included domains:
1. billing
2. accounting
3. reporting

## 7. Suggested Build Order Inside Milestones

### 7.1 Milestone A

1. skeleton, config, HTTP boot
2. migration framework
3. org and user membership model
4. roles and authorization wrappers
5. audit and idempotency
6. parties and external principals
7. attachments
8. worker identity and cost-rate schema
9. accounting kernel schema
10. AI run and approval schema shell

### 7.2 Milestone B

1. accounts and contacts
2. leads
3. opportunities and stages
4. activities and notes
5. communications and derived timeline views
6. shared tasks with CRM linking through `workflow`
7. search
8. estimates
9. AI assistance
10. first person/team ownership and queue semantics for shared follow-up

### 7.3 Milestone C

1. project aggregate
2. work order aggregate
3. work-order assignments
4. delivery use of shared tasks plus project milestones
5. time capture
6. light time-entry approval rules
7. expense capture
8. material issue/consumption
9. costing views
10. change orders and delivery milestone support
11. serialized equipment or installed-asset support where the delivery use case requires it
12. serviced-asset history support where repair or maintenance work targets a specific maintainable unit

### 7.4 Milestone D

1. invoices
2. receipts
3. posting into accounting
4. balances and revenue views
5. retention/certification controls
6. first analytic-dimension-backed profitability and management views

## 8. Architecture Guardrails

1. Keep one `org_id` strategy across all tenant-relevant business tables.
2. Prefer tenant-safe composite foreign keys where ownership matters.
3. Keep append-only audit history separate from mutable business state.
4. Avoid turning AI tables into alternate workflow state stores.
5. Keep search/indexing derivable from domain truth where possible.
6. Do not let inventory/material flows become procurement-first or stock-led.
7. Do not allow billing flows to bypass agreed commercial or delivery source records.

## 9. Testing Strategy By Phase

### Milestone A

1. migration tests
2. authz and tenant-isolation tests
3. worker/user separation tests
4. accounting posting contract tests
5. audit/idempotency tests

### Milestone B

1. CRM service tests
2. tenant-boundary tests across core CRM records
3. estimate lifecycle tests
4. AI tool-policy and approval tests

### Milestone C

1. project/work-order conversion path tests
2. worker assignment tests
3. time-entry and approval tests
4. costing aggregation tests
5. material usage tests
6. change-order and milestone tests
7. serialized-equipment or installed-asset traceability tests where identity-tracked deployment is supported
8. serviced-asset history tests where repair or maintenance slices support vehicle or equipment records

### Milestone D

1. invoice posting tests
2. receipt allocation tests
3. balance correctness tests
4. reporting correctness tests

## 10. Completed Companion Documents

These companion documents are now completed:

1. `service_day_module_boundaries_v1.md`
   - define bounded contexts, allowed dependencies, and ownership of cross-cutting services
2. `service_day_schema_foundation_v1.md`
   - specify tables, keys, tenancy, and accounting/commercial invariants
3. `service_day_crm_mvp_scope_v1.md`
   - define exact acceptance scope, workflows, and non-goals for the first runnable CRM
4. `service_day_ai_architecture_v1.md`
   - define providers, tool loop, approval flow, storage, and safety boundaries

## 11. Immediate Recommendation

Begin execution with Milestone A and do not treat it as mere scaffolding.

Before Milestone A is declared complete, explicitly lock these design points:
1. accounting kernel invariants
2. commercial document strategy
3. project/work-order/service-request relationship
4. communication model
5. shared task-engine approach
6. worker/time-entry/timesheet approach

If those six points are settled early, the remaining milestones can move quickly without forcing foundational rewrites.
