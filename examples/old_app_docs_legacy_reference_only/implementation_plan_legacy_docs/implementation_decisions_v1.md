# service_day Implementation Decisions v1

Date: 2026-03-13
Status: Legacy reference
Purpose: preserve the broader pre-thin-v1 locked decisions for historical context and older implementation detail.

Legacy note:
1. active thin-v1 defaults now live in `plan_docs/service_day_implementation_defaults_v1.md`
2. this file should not silently override `plan_docs/`

## 1. Decision Rules

1. This document records implementation defaults, not open brainstorming.
2. If a later decision changes one of these defaults, update this file and the canonical companion document it affects.
3. If code conflicts with this file, either fix the code or explicitly revise the decision here.

## 2. Locked Decisions

### 2.1 Shared Task Engine

Decision:
1. use one shared `tasks` engine across CRM, projects, and work orders
2. shared task orchestration belongs to `workflow`
3. task business meaning remains with the owning domain
4. each task has one primary actionable owner, either a person or a team queue
5. each task keeps one primary business context, while secondary related links may exist where justified for visibility and analytics

Implementation effect:
1. do not create separate `project_tasks` or `work_order_tasks` base tables in v1
2. use typed context links and domain-owned rules for task behavior
3. task completion may trigger domain behavior only through explicit domain services
4. do not allow one task to have several competing primary owners
5. treat team ownership as queue responsibility, not as an implicit many-person assignment model

### 2.1a Activity vs task model

Decision:
1. activities and tasks are separate concepts and should not be collapsed into one record type
2. activities represent factual history such as calls, visits, meetings, emails, field notes, and other completed events
3. tasks represent accountable next actions with ownership, due-state, and completion tracking
4. when an activity creates follow-up accountability, it should link to or spawn a task rather than overloading the activity row with full workflow semantics

Implementation effect:
1. keep activity history queryable without forcing every activity into task-state fields
2. keep task analytics clean by separating factual event logs from actionable commitments
3. allow timeline/read-model surfaces to show both activities and tasks together without making the timeline the owner of either concept

### 2.1b Workflow vs task vs activity terminology

Decision:
1. workflow, task, and activity are related but distinct concepts and should be modeled and described separately
2. a workflow is the process definition or state-governed path through which work moves, including transitions, permissions, reminders, approvals, and automation rules
3. a task is one accountable next action inside or alongside a workflow, with one primary owner, one primary business context, due-state, and completion tracking
4. an activity is a factual record of something that already happened, whether or not it produced a task or changed workflow state

Implementation effect:
1. do not use activity records as the primary owner of workflow state
2. do not use generic workflow state as a substitute for explicit task ownership
3. allow workflows to generate tasks and activities, and allow activities to lead to tasks, without collapsing all three into one table or one overloaded status model

### 2.2 Delivery Milestones vs Billing Milestones

Decision:
1. delivery planning milestones belong to `projects`
2. commercial billing milestones belong to `billing`
3. these are related concepts but not the same record type

Implementation effect:
1. keep `project_milestones` in delivery scope
2. keep `billing_milestones` in finance scope
3. do not overload project progress tracking as the billing source of truth without an explicit mapping rule

### 2.3 Commercial Document Strategy

Decision:
1. estimates and invoices should share naming conventions, line-shape conventions, and posting identifiers
2. do not force a shared physical base table in v1
3. revisit shared document-core storage only after Milestone B proves the CRM estimate flow

Implementation effect:
1. align field names where behavior overlaps
2. keep estimates in CRM ownership and invoices in billing ownership
3. define explicit conversion/mapping contracts between estimate lines and invoice source lines

### 2.3a Service-business-first scope discipline

Decision:
1. execute the current approved plan as a service-business-first product rather than broadening early toward every business type
2. do not add new CRM-universalization or other cross-industry broadening work to the current implementation plan unless that broadening is already present in the approved canonical documents
3. keep the currently planned extensibility and later-phase expansion seams, but do not turn them into present-scope implementation requirements
4. revisit broader CRM or module generalization only after the current planned milestones are implemented and a later explicit canonical decision says that expansion is needed

Implementation effect:
1. keep current CRM MVP acceptance and near-term implementation aligned to the existing service-business-first planning documents
2. do not treat cross-industry CRM breadth as a blocker for the current milestones beyond what the approved plan already requires
3. preserve extensible schema and API seams where the current plan already calls for them, but defer new universalization features, taxonomy broadening, and generalized workflow expansion
4. if later broadening is approved, update the relevant active canonical planning document first and then record the scope change in `plan_docs/service_day_refactor_tracker_v1.md`

Project-scope note:
1. in the current approved plan, CRM has higher near-term delivery priority than `projects`
2. do not let project scope drift turn the current roadmap into an advanced project-management product during the current milestone path
3. keep `projects` limited to the currently planned coordination, grouping, membership, milestone, and roll-up role around the stronger `work_order` execution core
4. advanced project features may be reconsidered in a later `v2` planning wave after the current CRM, execution, billing, and accounting milestones are substantially in place

### 2.4 Communication System of Record

Decision:
1. start with manual communication logging
2. keep the schema import-ready for email or channel sync later
3. communication records are domain truth; timeline is derived and append-oriented

Implementation effect:
1. do not block CRM MVP on inbox sync
2. persist direction, participants, subject/summary, linked context, and timestamps from day one
3. build timeline as a derived customer-facing read model, not the owner of communication data

### 2.5 Workforce and Time Model

Decision:
1. `worker` is the operational contributor record
2. `user` login stays separate
3. `time_entry` is the atomic time fact
4. `timesheet` is the review and approval wrapper, not the primary capture model

Implementation effect:
1. Milestone C can ship with raw time capture before full timesheet governance depth
2. time, cost rate, and bill rate snapshots must live on the source records or approval boundary
3. assignments target workers, not users

### 2.6 AI Write Boundaries

Decision:
1. AI may read, summarize, draft, and recommend in v1
2. AI may only trigger writes through approved tools and normal domain services
3. financially meaningful writes remain human-gated
4. meaningful writes and their audit trail must succeed or fail together; auditability is not best-effort
5. every business-state change must emit an audit record; low-level technical writes are not auto-audited unless they cross an explicitly defined audit boundary

Implementation effect:
1. persist runs, steps, recommendations, approvals, and accepted artifacts
2. do not allow direct table writes from provider adapters or tool handlers
3. use explicit approval records for sensitive actions
4. when a business action requires an audit event, persist both within the same transactional boundary
5. treat create/update/delete operations, state transitions, approvals, financially relevant changes, policy changes, and AI-accepted actions as audit-required business mutations
6. exempt technical writes only by documented category, for example migration bookkeeping, token/session expiry cleanup, projection rebuilds, idempotency housekeeping, and test/support maintenance writes

### 2.7 Bootstrap and identity integrity

Decision:
1. one-time bootstrap must be enforced by the database shape or transaction strategy, not only by an application pre-check
2. tenant-owned identity chains must be schema-safe from `org` to membership, role, and session linkage
3. active membership semantics govern who may authenticate and who may hold actionable workflow assignments in v1

Implementation effect:
1. do not rely on a read-before-write singleton check alone for bootstrap
2. prefer composite tenant-safe foreign keys where the child row already carries tenant context
3. keep workflow assignment and completion actor rules aligned with active membership policy
4. treat task assignee, creator, and completion actor as active-membership-only roles in v1

### 2.7a Workflow ownership semantics

Decision:
1. person assignment and team assignment are both required product capabilities
2. a team assignment is not the same thing as many simultaneous person assignees
3. watchers, followers, requesters, or participants are secondary visibility roles and should not replace the single actionable owner model

Implementation effect:
1. keep task ownership analytically clear so overdue, throughput, and queue-aging metrics remain credible
2. support later claim, delegate, and reassign flows without schema redesign
3. avoid many-assignee ambiguity as the default task contract because it weakens accountability, notifications, and SLA reporting

### 2.8 API validation and public error boundaries

Decision:
1. request-shape and identifier validation must happen at the service or HTTP boundary before persistence calls for externally supplied inputs
2. malformed caller input must map to stable client-visible `4xx` responses, not database-driven `500` errors
3. package-owned invalid-input sentinels should be used for handler error mapping instead of ad hoc raw validation strings
4. auth bearer-scheme parsing should be case-insensitive while still requiring an actual bearer token value

Implementation effect:
1. validate UUID-backed route params and query filters before any store-layer `::uuid` cast is reached
2. keep CRM, workflow, AI, and identity handlers aligned on explicit invalid-input and auth-failure mappings
3. do not rely on SQL cast failures or driver error text as part of the public API contract
4. add regression coverage for malformed identifiers and bearer-header case variants whenever auth or request parsing changes

### 2.9 Mobile/backend readiness stance

Decision:
1. the backend should treat mobile as a first-class client surface over the same domain services, not as a thin afterthought behind portal work
2. mobile-facing API contracts must be explicitly versioned and backward-compatible within a declared support window
3. mobile authentication should use device-scoped sessions with refresh-token rotation and revocation rather than relying only on one long-lived bearer token shape
4. list endpoints consumed by mobile clients must support pagination and a deliberate incremental-sync shape before mobile readiness is treated as complete
5. mobile-originated create and mutation requests that may be retried over unstable networks must support idempotent execution where duplicate effects would be harmful
6. attachment flows for mobile clients should use an explicit transport strategy such as presigned uploads/downloads or equivalent bounded file-transfer contracts, not ad hoc direct binary handling
7. notification delivery and device registration are backend concerns and must be modeled explicitly before mobile task/reminder flows are treated as production-ready
8. unless a later decision changes this, the first mobile client may be online-first, but that stance must remain explicit until an offline/sync contract is designed
9. the first planned mobile client may be built with Flutter, but backend mobile-readiness decisions remain client-agnostic and should not depend on Flutter-specific shortcuts

Implementation effect:
1. add versioning, deprecation, and compatibility rules to the HTTP API surface before mobile release work
2. define separate refresh/session records or equivalent auth primitives for device-aware mobile login lifecycle management
3. standardize pagination, `updated_since` or equivalent sync boundaries, and machine-readable error payloads for mobile-consumed endpoints
4. use idempotency keys for retry-prone mobile writes where duplicate creation or state transitions would be unsafe
5. keep mobile and portal surfaces on the same domain-service boundaries rather than introducing a second hidden business workflow path
6. do not treat Flutter as a reason to skip device-session, sync, upload, notification, or retry-safety backend work

Declared v1 API policy:
1. `/api/v1/...` is the explicit stable major-version path for the current API surface, and `/api/...` remains a same-shape alias for the current version during the pre-production phase
2. responses on versioned and unversioned API paths must stamp `X-Service-Day-API-Version: v1` and `X-Service-Day-API-Compatibility: additive-within-v1`
3. within `v1`, changes should remain backward-compatible and additive for existing clients; removing fields, changing semantics incompatibly, or removing endpoints requires a new major-version path rather than silent mutation
4. when an endpoint or field is deprecated, affected responses should use standard `Deprecation` and `Sunset` headers before removal; do not remove deprecated `v1` behavior without an explicit sunset window
5. until a later canonical decision changes this, the mobile contract remains online-first: sync reads, device-scoped auth, uploads, and idempotent retries are in scope, but offline write queues or local-first conflict resolution are not yet promised

### 2.9b Product interface stance

Decision:
1. the intended product interaction surfaces are mobile clients and later web or portal clients over the same backend domain services
2. CLI or REPL tooling is not a planned first-class product interface for `service_day`
3. small CLI-style commands are allowed where they materially help developer testing, migration, verification, seeding, or operational support
4. internal tooling must not become a hidden second workflow path that bypasses the normal API, approval, audit, or domain-service boundaries for product behavior

Implementation effect:
1. prioritize mobile and web or portal API contracts, auth, sync, upload, notification, and retry safety over any interactive shell tooling
2. keep `cmd/...` utilities narrow and task-specific rather than treating them as an alternate user experience
3. do not create roadmap or milestone language that implies a general CLI or REPL product surface unless a later explicit canonical decision changes that stance

### 2.9c Launch experience and search stance

Decision:
1. the later web client should prefer an activity-centered launch model rather than module-home-first navigation
2. a pinned home tile should launch one exact workflow screen, queue, or document-focused screen directly
3. global search should expose both actions and records across modules, not only entity records from one domain
4. bounded contexts remain backend ownership boundaries; they are not the primary navigation contract users should need to understand
5. per-user pinned-home configuration should remain a user preference layer, not a second owner of business workflow state
6. launch/search visibility must remain permission-aware, tenant-safe, and aligned with the same approval and audit boundaries used elsewhere

Implementation effect:
1. avoid future web-client designs that route users through module home pages before they can reach the actual target workflow
2. plan for stable launchable-activity identifiers or equivalent intent metadata that can survive cross-module growth
3. treat action search and record search as separate but unified result types under one global search experience
4. keep direct-entry navigation on top of normal domain-service and authorization boundaries instead of adding a hidden launcher-only workflow path
5. design per-user home-tile persistence as a preference/configuration concern, not as an ownership escape hatch from domain modules

### 2.9d Data exchange stance

Decision:
1. later CSV and spreadsheet import/export support is in scope as a controlled interoperability and bulk-operation capability, not as the primary day-to-day operating model
2. CSV is the minimum expected exchange format; Excel support may land later where operator value justifies it
3. imports must run through explicit bounded workflows or job-style execution paths, not direct table mutation
4. exports must come from authorized domain or read-model contracts, not informal frontend-only data assembly
5. import/export flows must preserve tenant safety, permission rules, audit visibility, and replay safety where duplicates would be harmful

Implementation effect:
1. keep domain write paths reusable so later import jobs can invoke the same validation and business rules as normal writes
2. keep file transport and attachment infrastructure generic enough to carry future import sources and generated export artifacts
3. preserve explicit read-model and reporting seams so export-friendly datasets do not depend on unstable client-side transformations
4. avoid introducing helper tooling that normalizes direct spreadsheet-to-table writes as an acceptable product pattern
5. leave room for a later job-tracking or batch-processing capability without forcing it into the earliest implementation slices

### 2.9a Mobile local-language speech-capture boundary

Decision:
1. mobile speech capture is a client-input convenience layer, not an alternate unreviewed business write path
2. the only backend command source in this flow is the user-approved transcript text, not raw audio or an unapproved draft transcript
3. approved transcript submissions must carry explicit locale metadata such as language/script plus a client command id or equivalent idempotency token
4. transcript approval is revision-specific: if the transcript text changes materially after recognition or correction, the new revision must be approved again before backend submission
5. raw audio retention is optional and policy-driven; if retained, it must be treated as attachment-like evidence rather than business truth
6. low-confidence or partial recognition output may help the client UX, but it must not bypass the approval gate or be treated as the canonical business instruction

Implementation effect:
1. treat approved transcript text as the canonical input to later business-event interpretation flows
2. require the mobile submission contract to include transcript text, locale/script, approval timestamp or revision marker, and an idempotency boundary
3. store any retained raw audio through bounded file/attachment contracts rather than a hidden AI side store
4. keep downstream interpretation and resulting writes on the same domain-service and audit boundaries used for typed commands
5. do not allow speech-recognition confidence scores or provider-specific audio artifacts to become public business-state dependencies

### 2.10 Accounting kernel baseline

Decision:
1. `accounting` owns the canonical posting boundary and ledger truth; operational modules may prepare posting inputs but may not write ledger state directly
2. `ledger_accounts`, `journal_entries`, and `journal_lines` are the minimum Milestone A accounting kernel records
3. accounting entries must follow an explicit lifecycle such as `proposed -> submitted -> posted -> reversed` or an equivalent constrained model; posting status remains an invariant-heavy system lifecycle and must stay constrained in code and schema rather than tenant-configurable
4. receivable and payable control-account treatment belongs to the accounting kernel, even when the originating document lives in billing or another operational module
5. posting from operational documents into accounting must execute through an explicit idempotent posting contract so retries cannot create duplicate ledger effects
6. the accounting foundation must be solid double-entry accounting even though the product is optimized first for service operations rather than finance-first workflows
7. AI may propose accounting entries by policy and may submit them only when org policy explicitly allows it, but AI may not post ledger entries
8. posting authority belongs to authorized human users with the configured accounting role or approval authority
9. accounting documents and entries must support explicit document types, durable unique numbering, accounting periods, reversal/void controls, credit notes, debit notes, and reversing journals as standard accounting capabilities

Implementation effect:
1. keep accounting posting services as the only path that creates `journal_entries` and `journal_lines`
2. require every future invoice, receipt, payroll, tax, or other operational posting flow to reference a stable source-document identifier and posting idempotency boundary
3. separate pre-posting entry preparation from final posting so finance users can review a submitted queue before ledger finalization when policy requires it
4. model posting lifecycle states explicitly enough to distinguish proposed or submitted operational state from posted, reversed, or otherwise accounting-final state
5. keep control-account assignment and posting validation inside accounting-owned rules rather than duplicating those invariants in billing or CRM
6. do not treat the presence of accounting tables alone as completion; the posting contract, review flow, and invariants remain part of Milestone A foundation quality
7. enforce balanced journal behavior and accounting-safe reversal/correction flows in the accounting layer rather than delegating that responsibility to billing or reporting
8. record proposer, submitter, poster, and timestamps explicitly enough for audit and approval reconstruction
9. keep document numbering durable and globally unique per configured series without financial-year reset unless a later explicit decision introduces a separate compliant numbering policy for a specific document class
10. support accounting-period open/close controls in a way that can fit both India and UAE operations without splitting the accounting core by country
11. distinguish void/cancel semantics, reversal semantics, credit/debit-note semantics, and reversing-journal semantics explicitly enough that user permissions and audit consequences are technically clear

### 2.10a India-first GST and TDS posture

Decision:
1. GST Lite and TDS are part of the service-company-ready finance baseline, not just later reporting extras
2. the first tax baseline should cover the common service-business document and accounting touchpoints cleanly before deeper edge-case statutory automation
3. the tax model must remain extensible so later India depth and later UAE tax depth do not require replacing the core accounting/document structures

Implementation effect:
1. keep tax metadata attachable to parties, documents, document lines, and accounting outputs where required
2. support service-oriented GST treatment on invoices, credit notes, debit notes, and related accounting flows
3. support TDS-related classification, deduction metadata, and accounting linkage suitable for service-company commercial flows
4. do not hard-code narrow report-only assumptions that would block deeper future compliance features

### 2.10b Analytics and dimensions posture

Decision:
1. the product should support advanced analytics because business-efficiency improvement requires measurable operational and financial outcomes
2. do not make `cost_center` the only or primary analytics concept in the system
3. use a typed analytic-dimensions model, with `cost_center` as one supported optional dimension type rather than the center of the design
4. the strongest business analysis should combine operational axes such as `project`, `work_order`, branch, service category, account/site, worker/team, and item/material context with configured analytic dimensions

Implementation effect:
1. keep operational records as the source of truth for execution context
2. allow accounting and reporting flows to carry or derive dimension context where financially and analytically useful
3. keep analytic dimensions controlled and typed instead of using freeform tags as pseudo-accounting structure
4. design reporting so businesses can analyze profitability, utilization, efficiency, and category performance without restructuring the chart of accounts for every management question

### 2.11 Inventory and materials deployment model

Decision:
1. `inventory_ops` exists to support service delivery, project execution, repairs, installed equipment tracking, and light direct item sales without becoming a full procurement-to-stock trading system in v1
2. the canonical operational material flow is `receipt -> project allocation? -> work_order reservation? -> issue/install/use -> bill/cost traceability`
3. item master data and stock movement are distinct from serialized or individually tracked deployed equipment records
4. when a material or equipment class requires identity-level traceability, model the specific unit or serialized instance explicitly rather than overloading generic quantity-only issue rows
5. sourced equipment must retain source-cost and supplier traceability sufficient for costing, warranty/support history, and later billing or project review, even if v1 avoids a full purchasing suite
6. customer-site installed equipment and service-owned deployed assets are operational facts and should remain queryable from delivery records rather than reconstructed from notes or attachments
7. billable versus non-billable material consumption must be explicit on the operational source records that feed costing and billing
8. the inventory foundation should remain usable by small trading companies at limited depth and extensible for later deeper inventory expansion if a future product decision justifies it
9. if a later limited inventory-trading extension is added, it must reuse the same item, receipt, stock-movement, billing, tax, and accounting foundations rather than introducing a second trading-specific inventory or ledger model
10. if a later small marketplace-seller extension is added, it must also reuse the same item, receipt, stock-movement, billing, tax, and accounting foundations rather than introducing a second commerce-specific inventory or ledger model
11. marketplace-specific scope should stay intentionally narrow for small sellers and must not broaden into listing-management, ads, or marketplace-operations-first product direction without an explicit later decision

Implementation effect:
1. keep `inventory_ops` responsible for item master data, stock movement, reservations or allocations, serialized-unit tracking where needed, and installed-asset traceability
2. do not make projects or work orders the owner of stock balances; they consume inventory through explicit allocation, reservation, issue, and usage records
3. allow quantity-based consumables and identity-tracked equipment to coexist in the same domain model without forcing every item into serialization
4. preserve supplier/source-document references and unit-cost traceability on receipts, receipt lines, or equivalent source records even if purchase orders remain out of scope in v1
5. add customer-site or installed-asset linkage so deployed networking gear, field devices, and service parts can be tied back to account/site/project/work_order context
6. keep billing sourced from explicit estimate, issue, usage, direct-sale fulfillment, or billing-preparation records rather than from ad hoc stock summaries
7. design item and commercial-line references so estimate/invoice lines can support service-led documents first while remaining extensible for later broader product-sale scenarios
8. do not broaden v1 into full vendor procurement approvals, replenishment planning, warehouse-optimization workflows, or advanced trading-company controls unless a later canonical decision changes the scope
9. if later direct-sale order, fulfillment, return, or inventory-adjustment records are added for trading-oriented flows, map them into the shared billing/tax/accounting contracts instead of building a separate trade-side finance model
10. if later marketplace order, return, fee, or settlement records are added, map them into the shared billing/tax/accounting contracts instead of building a separate commerce-side finance model

### 2.12 Serviced-asset and equipment-history model

Decision:
1. the system must distinguish the asset being serviced from the parts or materials consumed during service delivery
2. serviced assets may include customer vehicles, installed field equipment, customer-owned machines, or other maintainable units
3. when service history, warranty context, odometer or meter readings, serial identity, or installed-part traceability matters, the serviced asset must be modeled explicitly rather than inferred from notes or work-order free text
4. work orders should reference serviced assets when the execution target is a specific maintainable unit
5. installed parts or replacement components that materially affect future diagnostics, warranty review, or maintenance history should remain traceable from the serviced asset record

Implementation effect:
1. add a serviced-asset record family or equivalent aggregate that can represent vehicles and non-vehicle equipment without forcing the whole system into vehicle-only terminology
2. keep identifying fields extensible so domain-specific details such as VIN, registration number, chassis/engine number, serial number, site tag, or meter reading can be captured without weakening the generic model
3. let work orders, inspections, diagnostics, installed parts, and service history query surfaces link back to the serviced asset
4. keep parts consumption owned by `inventory_ops`, but let serviced-asset history expose which parts were installed, replaced, removed, or returned through linked operational records
5. do not treat CRM account/contact data as a substitute for the serviced-asset system of record

## 3. Additional Locked Decisions

These decisions are now also locked for implementation:

### 3.1 Task Context Shape

Decision:
1. each task has exactly one primary context
2. the primary context is one of the supported domain record types such as `lead`, `opportunity`, `project`, or `work_order`
3. secondary related links may exist for search, timeline projection, analytics, and cross-context visibility, but they do not replace the primary context

Implementation effect:
1. model task ownership with `context_type` and `context_id`
2. do not allow one task to directly own multiple primary business contexts
3. if a task is created from CRM but is really executing project or work-order follow-up, decide which record owns the task and keep the others as linked related context

### 3.2 Service Intake Shape

Decision:
1. use early `work_order` statuses for intake and triage
2. do not create a separate `service_requests` aggregate in v1

Implementation effect:
1. start `work_order` lifecycle with intake-oriented states such as `intake` or `triage`
2. keep customer/site/problem description fields on `work_orders`
3. revisit a separate intake record only if portal, SLA, or dispatch workflows later demand separate invariants

### 3.3 Status Modeling Policy

Decision:
1. use lookup tables for business-managed, admin-configurable progressions
2. use typed enums or constrained code constants for invariant-heavy system lifecycle states

Implementation effect:
1. keep sales pipeline concepts such as `opportunity_stages` configurable through lookup tables
2. keep states such as approval, posting, idempotency, and agent run lifecycle constrained in code and schema
3. do not make invariant-heavy workflow states tenant-configurable without a deliberate redesign

### 3.4 Timesheet Depth

Decision:
1. Milestone C ships `time_entries` plus light approval support
2. full timesheet submission and approval workflow is deferred until later usage proves the need

Implementation effect:
1. prioritize atomic time capture, rate snapshots, and correction rules in Milestone C
2. allow simple review or approval on time entries without requiring period-based timesheet UX
3. treat `timesheets` as a later governance layer, not as the first delivery dependency
