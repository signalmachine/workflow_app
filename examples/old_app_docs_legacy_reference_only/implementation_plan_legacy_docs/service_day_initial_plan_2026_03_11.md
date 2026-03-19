# service_day Initial Plan

Date: 2026-03-11
Status: Legacy reference
Scope: historical broad planning for the pre-thin-v1 direction

Legacy note:
1. this document is retained for historical context and older implementation rationale
2. `plan_docs/` is the active canonical planning set for current work

## 1. Product Intent

`service_day` is a Go + PostgreSQL application for service businesses, project-led businesses, and custom work-order businesses. It is optimized first for service-led operations rather than inventory-trading workflows where goods are purchased for stock and resold as the primary operating model, but the foundation should remain easy to use for small trading companies at limited depth and extensible if broader inventory scope is justified later.

The product direction is:
1. AI-agent-first operations platform with `work_order` support as the strongest core execution capability, even though CRM lands first as the earliest runnable slice.
2. Strong foundations early: org-aware data model, identity, audit, workflow, and AI execution boundaries.
3. Fast path to a runnable MVP before all later modules are implemented.
4. Progressive expansion into tightly integrated CRM, project coordination, work orders and costing, billing, accounting, GST Lite, and TDS.
5. UAE is the real commercial target market, while India remains the first delivered statutory and localization baseline.
6. Support controlled inventory and materials usage for service delivery, work orders, and projects, including light direct item sales, without making trading/inventory resale the primary product center of gravity in v1.
7. Support equipment-heavy delivery businesses where sourced gear is procured, allocated, deployed, installed at customer sites, and costed as part of projects or work orders without turning the product into a general trading ERP.
8. Support workforce tracking, time capture, and timesheet-ready workflows first, while intentionally preparing for a later SMB-focused payroll layer.
9. Support serviced-asset-centric businesses where work is performed against a specific customer vehicle, machine, installed device, or other maintainable unit with durable service history.
10. Even though the product is a service-business operations platform first, it must still include a technically solid double-entry accounting foundation rather than a lightweight or memo-only finance layer.
11. Because the product objective is improved business efficiency, the model must also support strong measurement, reporting, and analytics rather than treating operational reporting as a later afterthought.
12. Support a later rental-operations business mode where an operator leases a building or similar property from an owner, rents rooms or units onward to occupants, collects recurring rent, settles owner dues, and tracks margin plus operating expenses such as food service, utilities, and staffing.
13. Support a later mobile-first conversational capture flow where a small business owner speaks a business event in a local Indian language such as Telugu or Kannada, the app renders the recognized text back in the same language for approval, and only the approved text is submitted to the backend for interpretation and bounded processing.
14. Support a later narrow small-seller extension for businesses that source branded goods from contract manufacturers and sell through channels such as Amazon India, but keep that extension explicitly secondary to the service-business operating core.
15. Support later controlled CSV and spreadsheet import/export workflows where bulk migration, interoperability, reporting extracts, or finance-friendly data exchange are useful, while keeping the application itself as the system of record.
16. Treat end-user documentation as part of implementation quality so major user-facing workflows eventually ship with maintained user guides rather than relying only on developer-oriented docs.

## 2. Strategic Decisions

These decisions are fixed for the first planning pass.

1. Geography order:
   - Commercial target market: UAE.
   - First delivered statutory baseline: India.
   - Architecture priority: UAE and India must both fit the core model without redesign, but India lands first for compliance and payroll sequencing.
2. Deployment model:
   - Primary model is one self-hosted instance per business.
   - Multi-tenancy is still retained through `org_id` on all business data so one deployment can host multiple orgs if needed.
3. Tenant model:
   - Single PostgreSQL database.
   - `org_id` required by default on all tenant-relevant business, workflow, audit, AI, authz-membership, and integration tables.
   - Deliberately global reference/config tables may omit `org_id`, but only when the table is truly system-wide and the choice is explicitly justified.
   - Follow `business_day` tenanting discipline: tenant-scoped foreign keys and tenant-safe uniqueness.
4. User model:
   - Internal users first.
   - Customer login and portal foundations are designed early but implemented later.
   - The first planned mobile client is expected to be built with Flutter.
   - The primary product delivery surfaces are mobile clients and later web or portal clients over the same backend domain services.
   - The later web client should prefer an activity-centered launch experience with direct workflow or document entry rather than a broad module-home navigation model.
   - Small CLI-style utilities may exist for developer testing, verification, migration, seeding, or support tasks, but CLI or REPL tooling is not a planned first-class product interface.
   - That client choice does not replace backend mobile-readiness requirements; auth, sync, attachment, notification, and retry semantics remain backend concerns.
   - Later mobile workflows may include local-language speech capture for internal business-event entry, but transcript approval must occur before backend actioning so speech recognition does not become an unreviewed write path.
5. Payroll strategy:
   - Payroll is in scope later as an SMB-focused capability, not as an enterprise HR/payroll suite.
   - India payroll support should land before UAE payroll support if the sequencing stays technically practical.
   - UAE payroll readiness must still be baked into worker, time, organization, and accounting foundations from the beginning.
6. AI control model:
   - AI may assist, propose, summarize, and prepare actions.
   - AI may execute internal actions only through bounded tools and with explicit human approval or pre-approved workflow policy.
   - Financially significant writes remain human-gated and fully auditable.
7. Tax/regulatory strategy:
   - Minimal tax support at first.
   - Accounting foundation should arrive before or together with billing-adjacent operational flows that would otherwise create retrofit risk.
   - India-first GST Lite and TDS support should be suitable for service-company billing, expense, deduction, and accounting workflows while remaining extensible for deeper future tax features.
8. Architecture strategy:
   - Reuse `business_day` repo structure, modular-monolith boundaries, direct SQL posture, test discipline, and audit posture unless service-business realities demand a different choice.

## 3. First Business Archetypes

The initial design should optimize for these four archetypes while keeping extension paths open:

1. Agency / professional services firm
   - CRM-heavy, pipeline-heavy, proposal-heavy, project delivery after sale.
2. Construction subcontractor / field-service project business
   - Opportunity to estimate to project to progress billing to retention / collection.
   - Includes electrical, plumbing, low-voltage/networking, fit-out, and similar material-backed subcontract execution where projects and work orders both matter.
3. Repair / service center
   - Intake, diagnosis, work order, parts or consumables as incidental support, work tracking, invoice, and service history.
   - Often requires a durable record of the serviced vehicle or machine, plus installed-part and maintenance history.
4. Rental operator / managed-occupancy business, later in the roadmap
   - Operator leases a building or property from an owner, manages rooms or apartments as rentable units, collects recurring rent from occupants, pays the owner, and tracks operating margin after utilities, staffing, food, and similar expenses.
   - Depends more heavily on recurring commercial schedules, occupancy/unit records, payables/settlement discipline, and profitability reporting than on `work_order` execution, so it should land after the core service-business operating model is already stable.

Why this mix:
1. Together they force the model to support both pure services and service-plus-execution workflows.
2. They expose the need for CRM, project/tasking, work orders, costing, and billing without collapsing into trading ERP.
3. They create a reusable abstraction set for later service domains.

## 4. Core Product Boundary

`service_day` should support:
1. CRM
2. customer and relationship history
3. pipeline and quotation support
4. project and task management
5. cross-product task and activity management as a first-class efficiency capability
6. work orders
7. work-order costing
8. billing and collections
9. accounting foundation
10. GST Lite and TDS for India
11. UAE tax and compliance extensions later
12. lightweight SMB payroll later, starting with India and then UAE
13. customer portal later
14. AI-assisted and AI-orchestrated workflows throughout
15. controlled materials / consumables / spare-parts support for service execution, repairs, and project delivery
16. serviced assets such as vehicles, installed equipment, or customer-owned machines where execution and maintenance history attach to a specific unit
17. later rental-operations support for managed buildings, rooms, apartments, recurring rent billing, owner settlements, and operating-expense visibility where the business model is leasing capacity and re-renting it at a margin
18. later approval-gated mobile speech capture for business-event intake in local Indian languages, with same-language transcript review before backend processing
19. later limited inventory-based trading support for small operator-led buy-hold-sell flows built on the shared item, inventory, billing, tax, and accounting foundations without redefining the product as a trading-first ERP
20. later limited marketplace-seller support for small operator-led resale flows built on the shared item, inventory, billing, tax, and accounting foundations without redefining the product as a marketplace ERP
21. maintained user guides for real user-visible workflows under `docs/user_guides/`, written against shipped behavior rather than future intent
22. small internal CLI utilities where they materially help migration, verification, seeding, or support work without becoming a second product interaction model

`service_day` should not optimize for:
1. stock-led trading workflows
2. warehouse-heavy inventory operations as a primary design center
3. procurement-to-stock-to-resale flows as core architecture drivers in v1
4. full real-estate developer, brokerage, or large-facility property-management ERP scope as a primary design center in v1
5. marketplace-first seller tooling such as listing optimization, advertising operations, or multi-channel commerce orchestration as a primary design center in v1

Note:
Incidental materials, spare parts, consumables, direct item sales, and project/work-order stock are allowed. The architecture should treat them primarily as service-delivery support, work-order cost inputs, and lightweight commercial inventory support, not as the main business identity of the product.

Practical interpretation:
Equipment-heavy service providers such as networking integrators should be able to source vendor gear, receive it into controlled stock, allocate or reserve it for a project/work order, deploy it to a customer site, and keep source-cost plus installed-base traceability without requiring the application to become a full procurement-led resale ERP.

Repair-heavy businesses such as motor vehicle workshops should also be able to track the serviced asset itself, the work performed against it, the parts fitted or removed, and durable maintenance history without reducing the asset to freeform notes on the customer record.

Service companies that also sell parts, accessories, consumables, or equipment directly should be able to quote, invoice, and fulfill those item sales from the same item and stock foundation used for delivery work.

Small trading companies should also be able to operate on the same foundation at limited depth, as long as v1 remains clearly optimized for service-led operations rather than warehouse-heavy or procurement-heavy trading workflows.

Inventory-trading interpretation:
The shared item, stock-movement, billing, tax, and accounting foundations should also be extensible enough that a later limited trading-oriented business can buy stock for resale, hold quantity on hand, fulfill direct customer sales, process customer returns, and measure gross margin without introducing a second item or finance architecture. That support should remain intentionally bounded to small inventory-led operators and should not displace service delivery, `work_order` execution, or customer/project flows as the product center of gravity.

Small-marketplace-seller interpretation:
The shared commercial and inventory foundation should also be extensible enough that a later small Amazon-style seller can record supplier or contract-manufacturer sourcing, receive stock, capture or import channel orders and returns at limited depth, and reconcile channel settlements and fees through the shared finance core. That support should remain intentionally narrow and should not displace service delivery, `work_order` execution, or customer/project flows as the product center of gravity.

Accounting interpretation:
The product is still an operations platform first, but invoices, receipts, tax-relevant postings, payroll-ready obligations, and financial corrections must land on a real double-entry accounting model with balanced journals, controlled posting boundaries, and auditable reversal/correction flows.

Standard accounting interpretation:
The accounting layer should also include standard ledger controls expected in a serious business system: document types, globally unique document numbers that do not reset every financial year, controlled void/cancel or reversal behavior, credit notes, debit notes, reversing journals for accrual/adjustment workflows, and accounting-period controls that can support both India and UAE operating requirements on the same core model.

Tax interpretation:
GST and TDS should be strong enough for service-company operations rather than treated as superficial tax labels. The initial tax model should support service-oriented invoice and note flows, tax metadata on counterparties and documents, tax treatment on billable lines, and accounting linkage, while remaining extensible for deeper statutory workflows later.

Future expansion rule:
The foundational item, stock-movement, costing, and commercial-document model should not paint the product into a corner. If a later deliberate product decision broadens inventory depth, the current service-first model should be able to expand without foundational schema rewrites.

Rental-operations interpretation:
The same finance, party, site/unit, and recurring commercial foundations should be extensible enough that a later rental operator can model building owners, rentable properties and units, occupant agreements, recurring monthly rent charges, owner settlements, and operating-expense capture without requiring a second accounting architecture. That later support should remain intentionally bounded to operator-led rental businesses, not a broad real-estate ERP.

## 5. Product Principles

1. `work_order` support is the strongest product capability and should remain the clearest execution center of the system.
2. CRM, projects, work orders, billing, and accounting must connect cleanly through shared identifiers, explicit handoff flows, and derived read models rather than behaving like isolated products.
3. The system must become usable early, not only after all modules exist.
4. AI is a first-class subsystem, not a bolt-on assistant.
5. Human oversight, auditability, and tenant safety are non-negotiable.
6. Data model quality matters more than shipping narrow workflow shortcuts.
7. Foundations must make it easier, not harder, to add later countries, business types, and delivery channels.
8. Core business models should follow technically solid domain best practice first, not superficial alignment with any one market's common software habits.
9. Business-managed progression models should stay configurable where variation is normal, while monetary correctness, auditability, and linkage invariants remain enforced in schema and code.
10. Accounting must be technically solid double-entry accounting, not an approximate reporting layer bolted onto operations later.
11. Efficiency improvement requires measurable operational and financial outcomes, so analytics dimensions and reporting seams must be designed deliberately rather than inferred later from inconsistent transactional fields.
12. Task and activity management should be treated as a first-class efficiency layer across CRM, projects, work orders, billing follow-up, and later service-asset/site workflows rather than as a narrow CRM convenience feature.
13. Activities and tasks are different concepts: activities record what happened, while tasks record what should happen next and who owns that follow-through.
14. Shared workflow should support assignment to one person or one team queue as the actionable owner, with additional participants or watchers only as secondary visibility roles.
15. Tasks and activities should keep one primary business context for ownership and automation, but may also carry secondary related links such as account, site, project, work order, or serviced asset where visibility and analytics require them.
16. The later web launch experience should center on direct workflow or document entry, with personalized activity tiles and global search preferred over module-home sprawl.
17. Later spreadsheet import/export should remain easy to add through explicit exchange, attachment, audit, and domain-service seams rather than through direct table-loading shortcuts.

## 6. Recommended Module Order

This is the recommended build order.

### Phase 0: Foundation Kernel
1. Project guardrails and modular boundaries
2. org, identity, RBAC, audit, idempotency, workflow primitives
3. accounting kernel foundation
4. AI platform foundation
5. communication/activity timeline foundation
6. shared task/activity ownership, queueing, reminder, and analytics seams
7. party/external-identity foundation for future portal
8. workforce and assignment foundation
9. master data foundation for service businesses
10. service-linked inventory/materials foundation
11. keep import/export and batch-processing seams technically possible for later controlled data exchange

### Phase 1: CRM MVP
1. leads
2. organizations / accounts
3. contacts
4. opportunities
5. activities and follow-ups
6. notes, attachments, and conversation summaries
7. shared tasks linked to CRM records
8. quotations / estimates baseline
9. searchable unified relationship timeline
10. first user guides for shipped CRM/internal-user flows under `docs/user_guides/`

### Phase 2: Delivery Operations Foundation
1. projects
2. shared tasks in delivery context and project milestones
3. work orders, with optional service-request intake terminology
4. workers, assignments, and time-entry baseline
5. work-order costing baseline
6. light time-entry approval baseline
7. project/work-order material consumption and service-part usage baseline

### Phase 3: Commercial Operations
1. billing / invoicing
2. collections
3. basic revenue reporting
4. contract-commercial controls for project and work-order billing

### Phase 4: Finance and Compliance
1. double-entry accounting baseline
2. GST Lite for India
3. TDS baseline
4. UAE tax/compliance adaptation

### Phase 5: Customer Experience and Advanced Automation
1. customer portal
2. customer login and delegated access
3. customer request tracking and approvals
4. deeper AI workflow automation

### Phase 6: External Channel Integrations
1. WhatsApp support for customer-facing communication, updates, reminders, and later bounded customer workflow interactions
2. channel-specific communication import/sync rules where justified
3. approval and audit rules for any customer-facing outbound automation over external channels

### Phase 7: Limited Inventory-Trading Extension
1. limited stock-for-resale workflows on top of the shared party, inventory, billing, tax, and accounting foundations
2. direct-sale order or fulfillment records, customer returns, and stock-backed margin visibility for small trading-oriented operators
3. bounded replenishment, stock-adjustment, and inventory-accounting linkage needed for limited buy-hold-sell operations without turning the product into a warehouse-first ERP
4. keep the scope intentionally narrow and operationally simple rather than broadening into a trading-first ERP or distribution suite

### Phase 8: Narrow Marketplace-Seller Extension
1. limited supplier and contract-manufacturer sourcing support on top of the shared party and inventory foundations
2. simple channel-order, return, and settlement capture for small seller workflows such as Amazon India
3. accounting and tax linkage for channel fees, receivables/settlements, and inventory-backed item sales
4. keep the scope intentionally narrow and operationally simple rather than broadening into a commerce-first ERP or marketplace-operations suite

## 7. Why This Order

1. CRM first gives an early usable application and customer-context entry point, but it is not the long-term center of product differentiation.
2. Project and work execution should become the strongest operating layer because service businesses win or lose in delivery after sale.
3. Accounting kernel cannot be left too late; otherwise invoices, collections, and work profitability are built on weak foundations.
4. Portal should be designed early but built after internal workflows stabilize.
5. Service-linked inventory support should exist as a controlled operational layer, without allowing the product to drift into stock-trading ERP design.
6. external messaging channels such as WhatsApp should land after core CRM, execution, billing, portal, and audit boundaries are already strong enough that channel integration does not become a backdoor source of business truth
7. any later inventory-trading or marketplace-seller support should reuse the shared item, inventory, billing, tax, and accounting foundations rather than introducing a separate commerce stack or pushing the product toward a resale-first identity
8. user guides should be written incrementally with shipped workflows so implementation completeness is visible to real operators, not only to developers reading technical docs

## 8. CRM MVP Scope

For the first runnable CRM MVP, the minimum recommended scope is:

1. Accounts
   - customers, prospects, partner organizations, branch/location identities
   - explicit customer lifecycle stage, owner, and next-action visibility without requiring a separate customer-success module
2. Contacts
   - multiple contacts per account, role labels, communication preferences
3. Leads
   - source, qualification, assignment, conversion into account/contact/opportunity
4. Opportunities
   - value, stage, expected close date, probability, service category, owner
5. Activities
   - calls, meetings, emails, visits, reminders, follow-ups
   - activity records capture factual history; reminders and actionable follow-up should promote into shared workflow tasks where accountability matters
6. Notes and attachments
   - human notes, AI summaries, linked files, extracted metadata
7. Communications baseline
   - conversation records, message direction, summaries, and relationship-linked history
8. Quotations / estimates
   - line items oriented to services, milestones, work scope, or stocked items where the business also sells inventory directly
   - issue/expiry dates, commercial status, taxes, totals, and revision-safe progression
9. Tasks
   - user tasks created manually or proposed by AI
   - assignable to one person or one team queue
   - one primary context plus optional secondary related links for project, work order, site, asset, or customer visibility
10. Timeline
   - all customer-facing history in one tenant-safe stream
11. Search
   - fast lookup across account, contact, lead, opportunity, and activity data

Workflow-efficiency interpretation:
1. task and activity management should become a measurable operating layer across the product, not remain trapped inside CRM screens
2. managers should be able to review overdue follow-up, aging, queue load, completion throughput, and handoff friction across customer, project, and work-order contexts
3. later efficiency analytics should cut across person, team, project, work order, account/site, service category, and configured analytic dimensions

Not in CRM MVP:
1. full marketing automation
2. customer portal UI
3. advanced quote-to-project automation
4. full field-service dispatch optimization

## 9. Early Data Model Foundations

These foundations should exist before module sprawl begins.

### 9.1 Cross-Cutting Tables
1. `orgs`
2. `users`
3. `roles`
4. `user_org_memberships`
5. `audit_events`
6. `idempotency_keys`
7. `attachments`
8. `attachment_links`
9. `comments`
10. `tags`
11. `custom_fields`
12. `parties`
13. `party_roles`
14. `external_principals`
15. `portal_memberships`

Note:
The table lists in this section are logical planning anchors, not a literal rule that every listed table must always have identical tenancy behavior. The default is `org_id` on every tenant-relevant table. Any table without `org_id` must be intentionally global and documented as such.

For CRM notes in v1, use the shared `comments` primitive unless a later schema decision introduces a separate note aggregate.

### 9.2 CRM Core
1. `accounts`
2. `contacts`
3. `account_contacts`
4. `leads`
5. `opportunities`
6. `opportunity_stages`
7. `account_lifecycle_stages`
8. `activities`
9. `communications`
10. `communication_participants`
11. `estimates`
12. `estimate_lines`

Note:
Shared `tasks` belong to `workflow`, even when surfaced in CRM pages.
Relationship timeline should be treated as a derived CRM-facing read model, not the source of truth for underlying records.

### 9.3 Delivery Core
1. `projects`
2. `project_members`
3. `project_milestones`
4. `work_orders`
5. `work_order_status_history`
6. `work_order_assignments`
7. `work_order_lines`
8. `work_order_cost_entries`
9. `time_entries`
10. `change_orders`
11. `scope_items`

### 9.4 Workforce and Time Core
1. `workers`
2. `worker_roles`
3. `worker_skills`
4. `worker_cost_rates`
5. `time_entries`
6. `timesheets`
7. `timesheet_entries`

### 9.5 Finance Core
1. `invoices`
2. `invoice_lines`
3. `receipts`
4. `ledger_accounts`
5. `journal_entries`
6. `journal_lines`
7. `tax_profiles`
8. `billing_milestones`
9. `retention_terms`
10. `billing_certificates`

### 9.6 Service Inventory / Materials Core
1. `items`
2. `item_locations`
3. `material_receipts`
4. `material_issues`
5. `work_order_material_usage`
6. `project_material_allocations`
7. `service_parts_usage`

### 9.7 AI Core
1. `agent_runs`
2. `agent_run_steps`
3. `agent_tool_policies`
4. `agent_artifacts`
5. `agent_approvals`
6. `agent_recommendations`
7. `agent_memories`

## 10. Key Data Modeling Rules

1. `org_id` is the default on every tenant-relevant table.
2. Tables without `org_id` should be rare, intentionally global, and explicitly justified in schema docs.
3. Every tenant-scoped foreign key should include `org_id` where applicable.
4. Uniqueness should usually be tenant-scoped, for example `UNIQUE (org_id, id)` and business-key variants such as `UNIQUE (org_id, code)`.
5. Workflow records should separate durable business state from append-only audit history.
6. AI outputs that influence operations must be persisted as first-class records, not only logs.
7. Portal-readiness should be achieved through party and access abstractions early, not by retrofitting customer identity into internal-user tables later.
8. Inventory support must remain service-linked and work-order/project-linked by default, not modeled as a trading-first stock engine.
9. Workforce and time-tracking foundations should support service operations without assuming payroll depth.
10. Shared `tasks` belong to `workflow`, even when used by CRM, project, or work-order flows.
11. Relationship timeline is a derived CRM-facing read model, not the source of truth for communications, activities, tasks, or estimates.

## 10A. Tenancy Rule

The tenancy rule for `service_day` should be:

1. Design for future multi-company support from day one.
2. Keep `org_id` on all organization-owned operational data even if the first deployments are usually one business per instance.
3. Use tenant-scoped composite foreign keys wherever practical.
4. Treat cross-org data sharing as an explicit later feature, not an accidental side effect of schema shortcuts.
5. Require justification before introducing any global table that could otherwise create tenant leakage or later migration pain.

## 11. AI Architecture Direction

The initial AI architecture should be OpenAI-first but adapter-based.

### 11.1 Recommended Technical Shape
1. Provider-agnostic application contracts.
2. OpenAI adapter implemented first.
3. Responses API as the primary execution primitive for new agent features.
4. Strict JSON schema outputs and function/tool calling for all action-oriented flows.
5. Human approval checkpoints before sensitive writes.
6. Background execution support for longer-running reasoning tasks.
7. Full traceability of prompts, tool calls, approvals, outputs, and final accepted actions.
8. `agent_memories` are optional later support, not a Milestone A requirement.

### 11.2 Initial AI Capabilities
1. contact and lead summarization
2. call note cleanup
3. opportunity health summaries
4. follow-up draft generation
5. estimate draft assistance
6. next-best-action suggestions
7. conversation summarization across customer history
8. bounded workflow proposals, for example:
   - suggest next task
   - suggest follow-up date
   - draft customer reply
   - prepare estimate from notes

### 11.3 Later AI Capabilities
1. workflow-triggered autonomous preparation with human approval
2. project risk summaries
3. work-order costing anomaly detection
4. billing readiness checks
5. collections follow-up recommendations
6. customer-portal agent experiences

### 11.4 AI Safety Rules
1. No silent financial posting by agents.
2. No hidden write path outside explicit tools.
3. Tool permissions must be role-aware and org-aware.
4. Every agent action must preserve replay safety and audit evidence.
5. Agent memory must be scoped by org and purpose.

## 12. CRM to Portal Foundation Strategy

Portal is not in the first implementation slice, but these design choices should be present early:

1. model parties separately from auth principals where useful
2. allow one account to have multiple external contacts later
3. keep document and activity visibility policy-driven
4. treat portal as a separate delivery surface over the same core domain services
5. avoid portal-only tables that duplicate CRM truth unless required for security isolation

## 13. Country Strategy

### India First Delivery
1. company identity with Indian tax and statutory metadata fields
2. GST Lite later, not full GST engine at the start
3. TDS later after accounting baseline
4. lightweight India payroll should be the first payroll implementation when payroll begins

### UAE Commercial Target
1. support country-aware company profile and address model from day one
2. support tax registration metadata generically, not India-only column naming everywhere
3. keep document numbering and reporting adaptable for UAE later
4. keep worker, compensation, company-policy, and payroll-adjacent foundations generic enough for later UAE payroll support
5. do not hard-code India-only statutory assumptions into worker, time, compensation, or accounting foundations

## 14. Recommended Architecture Parity with business_day

The new codebase should mirror these `business_day` patterns unless a service-business need disproves them:

1. modular monolith
2. direct SQL with PostgreSQL, no ORM
3. explicit application contracts
4. authorized service wrappers
5. append-only audit posture
6. idempotent command handling
7. pre-start schema specs before workflow-heavy slices
8. strict testing gates
9. tenant-safe foreign key patterns

Recommended initial bounded contexts:
1. `identity_access`
2. `crm`
3. `projects`
4. `work_orders`
5. `workforce`
6. `inventory_ops`
7. `service_assets`
8. `billing`
9. `accounting`
10. `tax`
11. `ai`
12. `attachments`
13. `workflow`
14. `reporting`

## 15. First Implementation Milestones

This is the recommended milestone path for the separate `service_day` codebase.

### Milestone A: Bootable Foundation
1. repository skeleton
2. PostgreSQL migrations baseline
3. org, user, role, RBAC, audit, idempotency
4. attachments foundation
5. party and external identity foundation
6. worker identity and cost-rate foundation
7. accounting kernel foundation
8. AI adapter shell and policy framework
9. basic HTTP server and health boot

Exit condition:
Application boots, tenants are isolated, users can authenticate, audit/event plumbing exists, and the foundational attachment, worker, accounting, and party models are in place.

### Milestone B: CRM MVP Running
1. accounts, contacts, leads, opportunities
2. activities, communications, shared tasks, notes, attachments
3. estimate baseline
4. AI-assisted summaries and drafting

Exit condition:
A service business can manage its pipeline and relationship work end-to-end in a usable internal CRM.

### Milestone C: Delivery Operations Running
1. projects
2. shared tasks in delivery context and project milestones
3. work orders
4. workers, assignments, and time-entry capture baseline
5. material and cost capture baseline
6. progress/commercial controls baseline for project-led delivery

Exit condition:
Won work can move into execution with visible ownership, progress, and early costing.

### Milestone D: Billing and Accounting Baseline
1. invoices and receipts
2. customer balances and revenue visibility
3. retention / certification / contract-billing support where applicable

Exit condition:
Operational work can become billable and financially traceable.

## 16. Main Risks and Design Traps

1. Building CRM without project and work-order seams will cause expensive refactors later.
2. Delaying accounting too long will create invoice and profitability redesign pain.
3. Over-designing AI autonomy early will slow delivery and weaken safety.
4. Hard-coding India-only tax semantics into generic core tables will make UAE support messy.
5. Treating customer portal as only a UI concern will create identity and authorization debt.
6. Allowing service workflows to mimic trading ERP flows too closely will distort the product.
7. Under-modeling communication history will weaken both CRM usefulness and AI quality.
8. Under-modeling project-commercial controls will make construction and work-order-led businesses a poor fit.

## 17. Completed Supporting Documents

The planning set now includes:

1. `service_day_execution_plan_v1.md`
   - phased build plan and vertical slices
2. `service_day_module_boundaries_v1.md`
   - bounded contexts and dependency rules
3. `service_day_schema_foundation_v1.md`
   - tenant, identity, CRM, AI, and accounting baseline schema contracts
4. `service_day_crm_mvp_scope_v1.md`
   - detailed CRM acceptance scope
5. `service_day_ai_architecture_v1.md`
   - OpenAI-first adapter design, tools, approvals, memory, and audit model

## 18. Open Questions For The Next Pass

These do not block the current plan document or current implementation, but they should be answered before the later affected milestones are treated as implementation-ready:

1. Whether estimates and invoices should share one commercial document-core abstraction from the start.
2. How far Milestone C should go on progress billing, retention, and variation/change-order handling for construction-style service businesses.
3. Whether the customer portal later needs approval workflows, document exchange, ticketing, or only read-and-respond access first.

## 19. Immediate Recommendation

Start the separate `service_day` codebase with:
1. the same architectural discipline as `business_day`
2. a service-business-specific schema foundation
3. a CRM MVP target that becomes runnable early
4. an OpenAI-first AI layer built through provider adapters
5. explicit preparation for projects, work orders, service-linked inventory, billing, accounting, and portal expansion

This gives the fastest path to a usable product without locking the design into trading-ERP assumptions.
