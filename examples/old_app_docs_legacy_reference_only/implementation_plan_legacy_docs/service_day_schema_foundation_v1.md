# service_day Schema Foundation v1

Date: 2026-03-11
Status: Legacy reference
Purpose: preserve the broader pre-thin-v1 schema foundation for historical context and older implementation detail.

Legacy note:
1. active thin-v1 schema guidance now lives in `plan_docs/service_day_schema_and_module_boundaries_v1.md`
2. use this file only when a specific older schema decision needs clarification

## 1. Foundation Goals

The schema foundation should:
1. support CRM-first execution without trapping the product in CRM-only assumptions
2. support both small service businesses and more structured project-led businesses
3. keep `work_order` as the single primary execution object and the strongest long-term operating capability in the product
4. keep `project` optional but first-class
5. treat workforce, time capture, costing, and billing traceability as core foundations
6. preserve tenant safety, auditability, and future country expansion
7. keep worker, compensation, and organization foundations compatible with later lightweight payroll support for India and UAE
8. preserve tight integration across CRM, projects, `work_orders`, billing, and accounting so handoffs do not require duplicate state or lossy translation
9. ensure the accounting foundation is a real double-entry ledger model rather than a reporting-only summary layer
10. keep file, audit, idempotency, and read-model seams strong enough that later CSV and spreadsheet import/export can be added without foundational rewrites
11. keep task/activity ownership, queueing, and efficiency-analytics seams strong enough that later operational reporting does not require task-model redesign

## 2. Core Domain Spine

The primary business flow should be:

`lead -> opportunity -> estimate -> project? -> work_order -> time/material/expense -> invoice -> receipt -> ledger`

Important rules:
1. `project` is optional
2. `work_order` is the operational execution record
3. `service_request` is not a long-lived core aggregate in v1
4. time, materials, expenses, and assignment data should attach to execution records cleanly
5. billing should originate from structured commercial or execution records, not freehand reconstruction
6. CRM, project, and `work_order` records should connect through stable foreign keys and explicit conversion/handoff contracts rather than ad hoc duplicated fields
7. when work targets a specific vehicle, machine, or installed device, the serviced asset should remain a first-class linked record rather than being reduced to account-level free text
8. financial truth must be represented through balanced journal structures and explicit posting services rather than inferred only from invoices, receipts, or operational totals
9. accounting preparation state and posted ledger truth must remain distinguishable in schema and queries so review-before-post flows are safe

## 3. Naming Decisions

Use these names consistently in code, schema, and product language:

1. `project`
   - umbrella engagement, commercial container, or delivery program
2. `work_order`
   - primary unit of executable work
3. `task`
   - checklist or assignment item under CRM, project, or work order context
   - accountable next action with one primary owner and one primary business context
4. `activity`
   - factual event or interaction that already happened
   - may link to the same business records as tasks, but does not own workflow state
5. `workflow`
   - process and state-transition model around how work moves
   - may generate, govern, or consume tasks and activities, but is not itself a substitute for either
6. `worker`
   - person contributing operational work, whether employee or contractor
7. `time_entry`
   - atomic unit of logged time
8. `timesheet`
   - approval/review wrapper for many time entries, added after raw time capture

Avoid these as primary domain terms in v1:
1. `job_order`
2. `service_order`
3. `resource` as the primary user-facing label

`resource` may still appear internally where generic assignment modeling is useful, but the main workforce concept should be `worker`.

## 4. Tenancy and Identity Rules

1. Every tenant-relevant operational table should carry `org_id`.
2. Foreign keys that cross tenant-owned records should include tenant-safe ownership where practical.
3. Tables without `org_id` should be rare and intentionally global.
4. Internal identity and external party identity must remain separate.
5. Worker records should not depend on payroll design.
6. Worker and time foundations should not block later payroll support.
7. Country-specific payroll or tax fields should not be hard-coded into generic worker tables unless they are truly universal.
8. identity tables are part of the tenant-safety surface: memberships, roles, sessions, and other auth records must not permit logically cross-tenant linkage through raw ids alone.
9. when a business write is required to be auditable, the schema and write path should support a single transactional boundary for the business row and its audit event.
10. audit scope should be defined around business mutations, not every low-level maintenance write; the schema should preserve high-signal forensic history rather than generate indiscriminate noise.
11. one-time system bootstrap must be enforced by the database shape or transaction strategy, not only by an application-level read-before-write check; a singleton bootstrap-claim record is an acceptable v1 mechanism.

## 5. Recommended Foundational Tables

These are logical contracts across the v1 roadmap, not final migration files.

Milestone rule:
1. inclusion in this document does not mean the table must land in Milestone A
2. Milestone A should create only the kernel tables needed for boot, auth, audit, attachments, parties, workers, accounting kernel, and AI run/approval persistence
3. delivery, billing, and governance-layer tables may be defined here but land in later milestones

### 5.1 Cross-Cutting

1. `orgs`
2. `users`
3. `roles`
4. `user_org_memberships`
5. `auth_sessions`
6. `device_sessions`
7. `refresh_tokens`
8. `audit_events`
9. `idempotency_keys`
10. `attachments`
11. `attachment_links`
12. `comments`
13. `tags`
14. `custom_fields`
15. `parties`
16. `party_roles`
17. `external_principals`
18. `portal_memberships`
19. `analytic_dimensions`
20. `analytic_dimension_values`
21. `notification_devices`
22. `notification_events`
23. `notification_deliveries`
24. `teams`
25. `team_members`
26. `data_exchange_jobs`
27. `data_exchange_job_items`

Note:
1. v1 user-facing "notes" should use the shared `comments` primitive unless a later decision explicitly introduces a separate note aggregate
2. `audit_events` is the mandatory ledger for business-state changes, including AI-originated accepted actions that change product state through normal domain services
3. mobile/backend readiness should reuse the same tenant-safe identity spine; device or notification records must not become a second disconnected auth surface
4. `external_principals` and `portal_memberships` are part of the canonical identity surface for later portal or delegated customer access; they should not be modeled as CRM-owned contact variants
5. analytic-dimension definition records are controlled shared configuration, not freeform reporting tags or portal-only metadata
6. later `data_exchange_jobs` and related row/result metadata are acceptable shared operational-support records when spreadsheet or CSV import/export lands, but they should track exchange execution rather than become a second owner of imported business truth
7. `teams` and `team_members` are shared operational-support records so workflow, delivery, notifications, and reporting can reference a stable team concept without duplicating team identity inside each domain

Mobile/backend readiness rule:
1. mobile sessions should be modeled as device-aware extensions of the identity surface rather than as opaque bearer-token blobs
2. refresh-token records must be revocable, rotated explicitly, and linked back to the owning org, user, and device session chain
3. notification device registration and delivery bookkeeping belong in the canonical schema once mobile readiness moves from planning into implementation
4. API versioning and pagination are contract-level concerns, but any persisted compatibility or sync cursor state must remain tenant-safe and traceable
5. if later mobile speech-entry support is enabled, approved transcript text should be persisted as a typed AI artifact or equivalent approved-command record, while any retained raw audio should remain optional supporting evidence through attachment-like storage rather than becoming the canonical business command

Data exchange rule:
1. later import jobs should reference source attachments or equivalent bounded file artifacts rather than hidden local-only files
2. exchange-job tracking should stay metadata-oriented and tenant-safe
3. imported business records should still live in their owning domain tables rather than being duplicated in exchange tables

### 5.2 CRM

1. `accounts`
2. `contacts`
3. `account_contacts`
4. `leads`
5. `opportunities`
6. `opportunity_stages`
7. `account_lifecycle_stages`
8. `account_lifecycle_history`
9. `activities`
10. `communications`
11. `communication_participants`
12. `estimates`
13. `estimate_lines`

CRM read model rule:
1. relationship timeline is a derived CRM-facing read model
2. do not treat `timeline_events` as required source-of-truth storage in v1 unless a later decision explicitly adds an append table for projection support

CRM modeling rules:
1. customer lifecycle and sales-stage progression should be configurable through tenant-safe lookup data where business variation is expected
2. the current lifecycle state should live on the owning business record, while meaningful transition history should remain reconstructable without depending on timeline projection alone
3. quote or estimate records must preserve commercial correctness for pricing, discount, tax, currency, subtotal, and total fields
4. quote or estimate changes must be revision-safe through immutable supersession, explicit versioning, or an equivalent auditable mechanism rather than silent in-place history loss
5. later invoice, project, or work-order conversion must reference stable commercial-record identifiers rather than re-deriving the quote from timeline events or freeform text
6. customer lifecycle history should be persisted through an explicit append table such as `account_lifecycle_history`, not inferred only from the current account state and projected timeline rows

### 5.3 Delivery

1. `projects`
2. `project_members`
3. `project_milestones`
4. `work_orders`
5. `work_order_status_history`
6. `work_order_assignments`
7. `work_order_lines`
8. `work_order_cost_entries`
9. `change_orders`
10. `scope_items`

### 5.3a Serviced assets

1. `service_assets`
2. `service_asset_identifiers`
3. `service_asset_readings`
4. `service_asset_history` or equivalent derived/append support if needed

Serviced-asset rule:
1. the serviced asset is distinct from both the customer account and the parts consumed during repair or maintenance
2. the model should support vehicle-like and non-vehicle assets through extensible typed identifiers rather than hard-coding every domain into one narrow shape
3. work orders should be able to link to a serviced asset when work is performed on a specific maintainable unit

Recommended v1 serviced-asset shape:
1. `service_assets`
   - `id`
   - `org_id`
   - `account_id`
   - `site_id` nullable or equivalent customer-location reference when available
   - `asset_type_code` such as `vehicle`, `machine`, `installed_equipment`, or `other`
   - `display_name`
   - `manufacturer` nullable
   - `model` nullable
   - `status`
   - `commissioned_on` nullable
   - `decommissioned_on` nullable
   - `notes_summary` nullable
2. `service_asset_identifiers`
   - `id`
   - `org_id`
   - `service_asset_id`
   - `identifier_type_code` such as `vin`, `registration_number`, `serial_number`, `engine_number`, `chassis_number`, `site_tag`, or `customer_asset_code`
   - `identifier_value`
   - uniqueness should be tenant-safe for identifiers that must not repeat within an org
3. `service_asset_readings`
   - `id`
   - `org_id`
   - `service_asset_id`
   - `reading_type_code` such as `odometer_km`, `hours_run`, `cycle_count`, or `meter_reading`
   - `reading_value`
   - `captured_at`
   - `captured_by_membership_id` nullable
4. `work_orders`
   - add `service_asset_id` nullable for asset-centric work
5. optional derived history support
   - expose inspections, diagnostics, completed work orders, and installed-part changes through linked source records or a derived history surface rather than duplicating business truth in one monolithic asset-history row

### 5.4 Workforce and Time

1. `workers`
2. `worker_roles`
3. `worker_skills`
4. `worker_cost_rates`
5. `time_entries`

Deferred governance layer:
1. `timesheets`
2. `timesheet_entries`

Optional later payroll layer:
1. `payroll_profiles`
2. `compensation_components`
3. `worker_compensation_profiles`
4. `payroll_periods`
5. `payroll_runs`
6. `payroll_run_lines`
7. `payroll_payments`

### 5.5 Operational Cost Capture

1. `expense_entries`
2. `items`
3. `item_locations`
4. `material_receipts`
5. `material_issues`
6. `work_order_material_usage`
7. `project_material_allocations`
8. `service_parts_usage`
9. `inventory_reservations` or equivalent reservation records when material is committed before issue
10. `inventory_units` or equivalent serialized/item-instance records when identity-level equipment tracking matters
11. `installed_assets` or equivalent customer-site deployment records when delivered equipment must remain queryable after installation

Operational materials rule:
1. quantity-based consumables and identity-tracked equipment should share one coherent inventory model without forcing every item into serialization
2. source-cost and supplier-reference traceability should survive from receipt into costing and delivery review even if a full purchasing module is deferred
3. installed or deployed equipment should be traceable to account/site/project/work_order context through explicit records rather than inferred later from notes or attachments
4. the same item and stock foundation should support light direct item sales without requiring a second inventory model
5. item and commercial-document references should remain extensible enough that small trading-company usage and later deeper inventory expansion do not require replacing the foundational tables
6. if a later limited inventory-trading extension is added, the same foundation should support direct-sale fulfillment, customer returns, stock adjustments, and inventory-backed margin reporting without introducing a second stock ledger
7. if a later small marketplace-seller extension is added, the same foundation should support channel-order fulfillment, returns, and settlement reconciliation without introducing a second stock ledger

### 5.5a Optional later inventory-trading layer

1. `sales_orders`
2. `sales_order_lines`
3. `sales_fulfillments`
4. `sales_returns`
5. `inventory_adjustments`

Inventory-trading rules:
1. this later layer is intentionally narrow and should support limited buy-hold-sell workflows for small operators without turning the product into a warehouse-first or distribution-first ERP
2. sales-order lines should reference the shared item and commercial foundations rather than introducing a separate product catalog truth
3. fulfillments, returns, and adjustments should remain reconcilable back to shared inventory, billing, tax, and accounting records
4. broad purchase-order approvals, route distribution, warehouse slotting, and large-trading-company controls are out of scope unless a later deliberate product decision expands the direction

### 5.5aa Optional later marketplace-seller layer

1. `channel_accounts`
2. `channel_orders`
3. `channel_order_lines`
4. `channel_returns`
5. `channel_settlements`
6. `channel_settlement_lines`

Marketplace-seller rules:
1. this later layer is intentionally narrow and should support limited small-seller flows such as Amazon India without turning the product into a marketplace-first ERP
2. channel-order lines should reference the shared item and commercial foundations rather than introducing a separate catalog truth
3. channel returns and settlement lines should remain reconcilable back to shared inventory, billing, tax, and accounting records
4. listing-management, advertising, and broad marketplace-operations depth are out of scope unless a later deliberate product decision expands the direction

### 5.5b Optional later rental-operations layer

1. `properties`
2. `property_units`
3. `owner_agreements` or equivalent upstream lease contracts
4. `occupancy_agreements` or equivalent occupant/tenant contracts
5. `recurring_charge_schedules`
6. `deposit_liabilities`
7. `owner_settlements` or equivalent payable/settlement records
8. `property_operating_expenses`

Rental-operations rules:
1. rental support is a later extension and should reuse shared `accounts`/`parties`, billing, receipt, and accounting primitives rather than introducing a disconnected property-only subledger
2. rentable buildings, floors, rooms, apartments, or beds should be modeled through a flexible property/unit hierarchy rather than hard-coding one occupancy shape
3. upstream owner obligations and downstream occupant receivables should remain distinguishable so operator margin is measurable rather than inferred from net cash movement alone
4. deposit, utility recharge, food-package charges, and similar recurring or semi-recurring items should fit the same commercial and accounting controls as other finance documents
5. if a rental business also performs maintenance or service work, those operational events should link to the shared `work_order` model instead of creating a second maintenance workflow stack inside the rental tables

### 5.6 Finance

1. `invoices`
2. `invoice_lines`
3. `receipts`
4. `receipt_allocations`
5. `ledger_accounts`
6. `journal_entries`
7. `journal_lines`
8. `accounting_entry_submissions` or equivalent submission metadata if that is the chosen lifecycle shape
9. `accounting_periods`
10. `accounting_document_sequences` or equivalent durable numbering state
11. `credit_notes`
12. `credit_note_lines`
13. `debit_notes`
14. `debit_note_lines`
15. `tax_profiles`
16. `tds_sections` or equivalent withholding-reference configuration if needed for the India baseline
17. `billing_milestones`
18. `retention_terms`
19. `billing_certificates`

### 5.7 AI

1. `agent_runs`
2. `agent_run_steps`
3. `agent_tool_policies`
4. `agent_artifacts`
5. `agent_approvals`
6. `agent_recommendations`

Optional later support:
1. `agent_memories`

### 5.8 Mobile and client-delivery support

1. `device_sessions`
2. `refresh_tokens`
3. `notification_devices`
4. `notification_events`
5. `notification_deliveries`

Mobile support rules:
1. `device_sessions` should belong to `identity_access` and represent one install or signed-in device context, not a second user model
2. `refresh_tokens` should be rotation-friendly and should not be the only source of session truth; revocation should remain possible at the device-session level
3. `notification_devices` should capture per-device push-registration state and delivery preferences without duplicating CRM or workflow business truth
4. `notification_events` and `notification_deliveries` should track fan-out and delivery bookkeeping only; they should not replace the underlying workflow, CRM, or approval events that triggered them
5. if limited offline support arrives later, any sync cursors, client mutation queues, or conflict markers should be modeled as adjunct client-state records rather than contaminating the core business aggregates

## 6. Aggregate Intent

### 6.1 Projects

`project` should represent grouped or long-running work.

`project` is the coordination and aggregation layer, not the default execution record.

Use `project` when you need:
1. multi-phase delivery
2. roll-up visibility across several work orders
3. project-level budget or profitability
4. commercial grouping for milestones or progress billing
5. a client-facing engagement or delivery program that spans several concrete execution steps

Do not force a project for every service interaction.

Project characteristics typically include:
1. overall budget or target profitability
2. broad start and end dates
3. milestones or phase coordination
4. contract or commercial grouping
5. high-level progress tracking across multiple execution units

### 6.2 Work Orders

`work_order` is the operational core.

`work_order` is the concrete executable unit of work. It should be the place where scheduling, assignment, labor, materials, and completion actually happen.

It should support:
1. assignment
2. scheduling
3. execution status
4. customer and site context
5. time capture
6. material usage
7. expense capture
8. operational costing
9. billing readiness

Work-order characteristics typically include:
1. a specific scope or task description
2. assigned technician, worker, or team
3. scheduled execution date or window
4. labor, material, and expense capture
5. completion status and execution notes
6. serviced-asset linkage where the job targets a specific vehicle, machine, or installed unit

Recommended structural rule:
1. `work_orders.project_id` should be nullable
2. `work_orders.opportunity_id` may be nullable
3. `work_orders.account_id` should usually be present

This allows:
1. project-led execution
2. direct service work without a project
3. repair-shop and field-service flows without unnecessary overhead
4. asset-centric repair flows where the service history of one vehicle or machine matters across many work orders

Recommended hierarchy:
1. `lead -> opportunity -> project? -> work_order -> time/material/expense -> invoice`
2. `project` remains optional
3. one project may group many work orders
4. one small service job may use a work order without any project at all

Design rule:
1. do not mix project coordination semantics into the work-order execution record
2. do not use projects as the place where labor, materials, and execution truth are captured
3. do use projects as the place where several work orders are grouped, coordinated, and summarized

### 6.3 Service Request

`service_request` should not be a first-class bounded context in v1.

Recommended treatment:
1. use it as intake terminology in UI where needed
2. represent intake as an early `work_order` status
3. do not build parallel costing, assignment, or billing semantics around it

## 7. Workforce Model

The platform should model operational contributors without requiring payroll depth.

### 7.1 Worker Principles

1. A `worker` may be an employee or contractor.
2. A `worker` may or may not have a login `user`.
3. Costing should use rate snapshots on time/expense records, not only current worker defaults.
4. Assignment should reference workers, not only users.
5. Payroll support may be added later, but payroll-specific state should layer on top of workers rather than replacing the worker model.

### 7.2 Worker Fields

Recommended baseline fields:
1. `id`
2. `org_id`
3. `party_id` nullable
4. `user_id` nullable
5. `code`
6. `display_name`
7. `employment_type`
8. `active`
9. `default_cost_rate`
10. `default_bill_rate` nullable
11. `team_id` nullable
12. `branch_id` nullable
13. `hired_on` nullable

Recommended team shape:
1. `teams`
   - `id`
   - `org_id`
   - `code`
   - `display_name`
   - `active`
2. `team_members`
   - `id`
   - `org_id`
   - `team_id`
   - `worker_id`
   - `role_code`
   - `active`

### 7.3 Assignment Model

Use explicit assignment tables where ownership matters:
1. `project_members`
2. `work_order_assignments`

Assignments should support:
1. assigned role
2. primary/secondary responsibility
3. planned start/end
4. allocation percentage or expected hours where useful

Workflow-assignment rule:
1. work-order assignment remains domain-owned in `work_orders`
2. task ownership may target either a `worker` or a `team`
3. if a team owns a task first, later claim or delegate actions may assign it to one worker without losing the original queue history

## 8. Time Tracking and Timesheets

Time tracking is foundational. Timesheets are a governance layer over time tracking.

### 8.1 Time Entries

`time_entry` is the atomic fact.

Recommended fields:
1. `id`
2. `org_id`
3. `worker_id`
4. `work_order_id` nullable
5. `project_id` nullable
6. `task_id` nullable
7. `entry_date`
8. `start_time` nullable
9. `end_time` nullable
10. `duration_minutes`
11. `description`
12. `billable`
13. `cost_rate_snapshot`
14. `bill_rate_snapshot` nullable
15. `approval_status`
16. `timesheet_id` nullable

Rules:
1. A time entry should point to at least one work context.
2. A time entry should be immutable after approval except through explicit correction flow.
3. Cost and bill rates should be snapshotted at entry or approval time.

### 8.2 Timesheets

`timesheet` is not the same as time tracking.

It exists to support:
1. weekly or monthly review
2. manager approval
3. lock/freeze behavior
4. utilization and labor reporting
5. cleaner billing cutoffs

Recommended rollout:
1. build `time_entries` in the delivery foundation phase
2. add `timesheets` once approval, utilization, or larger team control matters
3. add payroll only after worker, time, compensation, and accounting seams are stable enough to support it cleanly

### 8.3 Billing and Costing Implication

Time entries should feed:
1. work-order costing
2. project profitability
3. billable labor preparation
4. worker utilization reporting later

## 9. Costing Model

Costing should be built around execution records, not only around financial documents.

Recommended cost sources:
1. `time_entries`
2. `expense_entries`
3. `work_order_material_usage`
4. external service/vendor charges later if needed
5. optional later rental-operating costs such as utilities, food service, owner settlements, and property staffing where rental support is enabled

`work_order_cost_entries` can serve as a normalized roll-up or journalized operational costing layer if you want one consolidated cost stream.

Recommended rule:
1. raw source records remain canonical
2. summarized cost records must be derivable or traceable back to source facts
3. material cost should trace back through receipt, reservation or issue, and usage records closely enough to explain delivered-equipment margin and warranty-support history

### 9.1 Analytic dimensions recommendation

Recommendation:
1. do not make `cost_center` the only analytics concept in the system
2. support a typed analytic-dimensions model, with `cost_center` as one optional built-in dimension type rather than the primary design center
3. let operational records and accounting records both carry dimension context where analytically useful
4. keep dimensions controlled and typed, not freeform tags pretending to be accounting truth

Recommended shared shape:
1. `analytic_dimensions`
   - dimension definition per org
   - includes stable code, display name, dimension type, and status
2. `analytic_dimension_values`
   - controlled values under one dimension definition
   - supports stable code, display order, status, and optional parent/child hierarchy where the business genuinely needs it
3. module-owned records may then reference `analytic_dimension_values` either:
   - directly through explicit foreign keys where one fixed dimension slot is justified
   - through a typed assignment table or equivalent owned association where multiple dimensions per record are justified
4. the operational or financial source record remains the system of record; dimension tables provide controlled classification, not a second transactional truth layer

Suggested dimension examples:
1. `cost_center`
2. `business_unit`
3. `department`
4. `service_line`
5. `region`
6. `channel`

Design rule:
1. the strongest analysis axes will still often come from the operational model itself, such as `project`, `work_order`, branch, account/site, service category, worker/team, and item/material context
2. analytic dimensions should complement those operational axes, not replace them
3. reporting should support slicing profitability, utilization, revenue, cost mix, and efficiency by both operational context and configured dimensions
4. dimension definitions should be owned centrally enough that different modules do not invent conflicting local versions of the same business classification

### 9.1 Materials and deployed-equipment implications

Recommended operational flow:
1. material or equipment is received with source-cost traceability
2. quantity may remain available in stock or be allocated to a project
3. specific quantity or serialized unit may be reserved for a work order
4. quantity or unit is issued, installed, consumed, or returned through explicit records
5. billable and non-billable usage stays distinguishable on the source records used for costing and billing

Design rules:
1. do not rely on aggregate stock balance alone to explain what was deployed to a customer site
2. when serial number, warranty, RMA, or installed-base history matters, use identity-tracked unit records rather than only quantity movement rows
3. customer-site installed equipment should remain queryable even after the original project or work order is complete
4. delivery modules may reference inventory records, but inventory ownership of stock movement and equipment traceability remains with `inventory_ops`

### 9.2 Serviced-asset implications

Recommended serviced-asset flow:
1. create or identify the serviced asset such as a vehicle, machine, or installed device
2. link diagnostics, inspections, and work orders to that asset
3. link fitted, removed, or replaced parts through the operational inventory records
4. preserve service history and key readings so later maintenance, warranty review, and repeat visits remain traceable

Design rules:
1. do not force customer-owned maintainable units to live only as CRM notes or attachment metadata
2. keep asset identity extensible enough for VIN, serial number, registration, engine/chassis number, meter readings, or site tags
3. when installed-part history matters, expose it through linked service-asset and inventory records rather than duplicating the truth in a second manual history table

## 10. Commercial and Billing Model

Implementation default:
1. use one shared `tasks` table in `workflow` for CRM and delivery work
2. each task has exactly one primary business context through `context_type` and `context_id`
3. represent domain-specific task meaning through typed context links and module rules, not duplicate task tables
4. allow secondary related links where search, timeline, analytics, or cross-context visibility require project, work-order, account, site, or serviced-asset tagging in addition to the primary context
5. keep activities as separate domain records rather than collapsing them into the task table

Recommended workflow task shape:
1. `tasks`
   - `id`
   - `org_id`
   - `context_type`
   - `context_id`
   - `owner_type` such as `worker` or `team`
   - `owner_id`
   - `status`
   - `priority`
   - `due_at` nullable
   - `claimed_by_worker_id` nullable when team-queue flows later need explicit claim state
   - `completed_at` nullable
   - `completed_by_worker_id` nullable
2. `task_related_links` later if or when needed
   - keeps secondary links such as `project`, `work_order`, `account`, `site`, or `service_asset`
   - must not replace the single primary-context contract

Recommended activity modeling rule:
1. activities should support one primary linked record plus optional related links where practical
2. activities may reference a responsible person for follow-up, but they should not become the owner of full queue, SLA, or completion semantics
3. if an activity becomes an accountable next action, create or link a task instead of overloading the activity record

### 10.1 Estimates

Estimates should support:
1. service lines
2. milestone lines
3. optional linkage to opportunity and later project/work-order creation

### 10.2 Invoices

Invoices should support:
1. direct invoice lines
2. milestone-based billing
3. progress/proportion billing
4. retention and billing-certificate hooks where relevant

### 10.3 Source of Billing Truth

Billing should originate from:
1. estimate commitments
2. approved milestones
3. approved work completion states
4. approved billable time or billable cost accumulations later
5. optional later recurring schedules such as monthly rent, utility recharge, or service-package billing where rental support is enabled

Do not rely on invoices as the place where execution truth is recreated manually.

## 11. Accounting Kernel Invariants

These invariants should exist before billing is treated as complete:

1. `ledger_accounts` exist with tenant-safe ownership
2. `journal_entries` and `journal_lines` support balanced posting
3. accounting lifecycle state distinguishes proposed or submitted entries from posted ledger truth
4. operational documents post through explicit posting services
5. reversal and correction flows are explicit
6. receivable positions are derivable from posted transactions and receipt allocations
7. document type and unique document number are explicit on accounting and billing records where they materially affect auditability or traceability
8. accounting-period boundaries and period status are enforceable in schema and service rules
9. reversing journals and standard credit/debit-note flows are representable without mutating posted history
10. posted ledger rows are never treated as casually mutable operational summaries
11. proposer, submitter, poster, and related timestamps are reconstructable from canonical records and audit history
12. GST and TDS metadata can attach to the relevant parties, documents, lines, and accounting outputs without overfitting the generic finance core to one country
13. the same accounting lifecycle should be able to represent later rental-operator receivables, owner settlements, deposits, and operating expenses without requiring a second ledger model

## 12. Audit and Mutation Rules

1. Approval-worthy records should use explicit status transitions.
2. Sensitive state changes should emit audit events.
3. AI-originated proposals should never bypass normal domain write paths.
4. Deletions should be rare; prefer status changes and soft archival where appropriate.
5. Idempotency must exist on write-heavy command paths and integrations.

## 13. Non-Goals for v1 Foundation

1. payroll
2. full HR model
3. warehouse-heavy stock management
4. advanced dispatch optimization
5. country-specific deep compliance outside India-first baseline planning

## 14. Immediate Next Schema Decisions

These should be resolved in the next pass before migrations begin:

1. whether estimates and invoices share a common document-core base at the schema level
2. which specific status models deserve lookup tables versus typed enums
3. exact optional secondary task-link design, only if later use cases require it

## 15. Recommendation

The schema foundation should optimize for this mental model:

1. CRM owns customer and sales context
2. `project` owns grouped engagement context when needed
3. `work_order` owns execution
4. `worker`, `time_entry`, materials, and expenses own cost input
5. billing and accounting turn structured operational truth into financial truth
6. payroll, when added later, should consume worker, time, compensation, and accounting facts without distorting those earlier domain boundaries

That model is simple enough for small service businesses and strong enough to support more complex workflows later.
