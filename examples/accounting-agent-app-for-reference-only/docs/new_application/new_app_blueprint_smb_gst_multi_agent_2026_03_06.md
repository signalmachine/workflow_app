# business_day Blueprint (SMB Core: Accounting + Inventory + GST + Multi-Agent)

Date: 2026-03-06  
Status: Draft for review
Application name: `business_day`

## Confirmed Product Decisions (2026-03-06)

1. Geography and tax scope for v1:
   - India-only product in v1.
   - GST-first implementation.
   - No multi-country tax engine in first release.
2. Tenanting model for v1:
   - Multi-company capable data model with strict `org_id`/company scoping on all business and accounting tables.
   - Product is not planned as a SaaS business model.
   - Primary deployment mode is per-company instance (local or cloud), while still supporting multiple company codes in one instance when needed.
3. User-company association for v1:
   - A logged-in user is associated to one default company only.
   - No multi-company switching in v1.
   - User access is enforced by company code scoping.
   - Active company context must be visibly shown on all transactional and reporting surfaces (for example: company code on journal proposals, posted documents, and reports).
4. Agent write safety model for v1:
   - Agent proposes action.
   - Human can confirm, amend, or cancel.
   - No autonomous financial write execution.
5. Target SMB operating band:
   - Around 1,000 invoices/month per tenant.
   - Up to 3 active accounting/inventory operators.
6. Inventory valuation in v1:
   - Weighted average costing.
   - Design must keep path open to add FIFO later.
7. Delivery capacity assumption:
   - One primary developer using AI coding tools.
8. Currency model for v1:
   - Single base currency per `org_id`/company.
   - No multi-currency transactions in v1.
9. Subledger and control-account policy:
   - Control accounts (AR/AP/Inventory) are system-managed.
   - Direct manual posting to control accounts is blocked by default.
10. Document numbering policy:
   - Document types are explicit and extensive.
   - Document type code format is strict two-letter uppercase (example: `GR`, `PI`, `PE`, `PS`).
   - Document numbers are globally unique and do not reset at year close.
11. Procurement and sales workflow model:
   - Workflow is document-sequenced by transaction type (inventory: receipt/delivery + invoice; non-inventory: direct invoice).
   - Purchase order and sales order/delivery are optional planning/logistics documents.
   - Invoice booking is independent of PO/SO closure flow.
   - Inventory-goods procurement booking is mandatory two-document flow: `GR` then `PI`.
   - `GR`-only booking is allowed when goods are received but vendor invoice is pending.
   - `GR` posts inventory receipt and temporary procurement liability (GR clearing payable, not vendor AP).
   - `PI` clears GR clearing payable and transfers liability to vendor AP.
   - Service/other-expense invoices are booked directly without `GR` using separate document types.
   - Service Receipt Note is a future workflow, not part of current v1 booking flow.
   - Purchase invoice and `GR` can optionally reference a PO number.
   - Inventory-goods sales booking is two-document flow: `DN` then `SI` (mostly same workflow action).
   - `DN`-only booking is allowed when goods are delivered but invoice is pending.
   - `DN` records inventory issue/COGS impact; `SI` records customer AR/revenue/tax.
   - Service/other non-inventory sales invoices are booked directly without `DN` using separate document types.
   - Service Delivery/Service Receipt style workflow for sales is future scope.
   - Sales invoice and `DN` can optionally reference a SO number.
12. Period-close capability:
   - Period close states are required in v1.
   - Before hard close, limited editable fields allowed for high-privilege users with full audit logging.
13. Access control model for v1:
   - Instance-level `super_admin` role is mandatory.
   - `super_admin` can create users, assign roles, and manage access control policies.
   - Expected business roles include `finance_manager`, `accountant`, and `auditor` (view-only).
14. Primary operating model:
   - Application is AI-agent-operated by default for day-to-day workflows.
   - Human users primarily provide intent, approvals, and exception handling inputs to agents.
   - Fully manual end-to-end execution remains available as fallback/continuity mode.
15. Interface and web stack direction:
   - Primary web stack: Go + Go templates + HTMX + Alpine.js (plus compatible minimal JS/CSS utilities as needed).
   - Web UI must be highly modular (component/partial-driven) to support rapid workflow evolution.
   - REPL and stateless CLI are secondary interfaces for testing, diagnostics, and power-user operations.
   - Architecture should be API-ready for a lightweight Flutter mobile client (focused operational use, not full desktop parity in v1).

## 1. Objective

Design and build `business_day` as a general SMB business operating application, starting with accounting, inventory, and GST as v1 core, with:

1. Strong foundational data model that can absorb future features without schema churn.
2. GST-first workflows (India-ready) for SMB operations.
3. SMB-friendly UX and defaults (low setup burden, clear guardrails, minimal accounting jargon).
4. Workflow-focused multi-agent architecture (not a single monolithic agent).
5. Product quality inspired by SAP strengths, without SAP-level operational complexity.
6. Clear expansion path to service-business operations and non-financial workflows (for example CRM capabilities).
7. AI-agent-first operating experience, with manual UI flows as secondary backup path.

## 2. Product Principles

1. Accounting integrity first: every financial workflow must map to balanced journal postings.
2. Operational events first: users do sales/purchase/payment/inventory tasks, not manual journals.
3. Policy-driven behavior: tax, numbering, approvals, and posting rules must be configurable by policy tables, not hardcoded logic.
4. Append-only finance records: corrections are reversals/adjustments; no destructive edits to posted entries.
5. Explainability by default: every auto-posting includes "why this was posted" metadata.
6. Agent safety by design: agents can propose actions; critical writes require policy-gated confirmation.
7. Platform extensibility by design: finance and non-finance domains must coexist under shared tenant, workflow, policy, and audit foundations.
8. Company-context visibility by default: each transaction/proposal/report must display the active company code to prevent cross-company user confusion.
9. AI-primary operations by default: business workflows should be executable end-to-end through modular agent orchestration; manual flows are resilience fallback.

## 3. SMB Scope (v1 Finance-First Release)

### In Scope

1. Multi-company (tenant-scoped), single-country initial rollout (India GST).
2. Sales cycle: estimate (optional), delivery note (`DN`), sales invoice (`SI`), receipt, credit note, debit note.
   - Inventory-goods sales booking uses `DN` then `SI` in v1.
   - `DN` can be posted independently before invoice receipt; `SI` can be posted later against the `DN`.
   - `DN` captures goods issue and valuation impact; `SI` captures AR/revenue/output-GST.
   - Service and other non-inventory sales invoices are direct invoice postings (no `DN`) under separate document types.
   - Sales order is optional and informational/planning by default.
3. Purchase cycle: purchase order (optional), goods receipt (`GR`), purchase invoice (`PI`), vendor payment, debit/credit notes.
   - Inventory-goods purchase booking requires `GR` followed by `PI` in v1.
   - `GR` can be posted independently before invoice receipt; `PI` can be posted later against the `GR`.
   - `GR` creates liability in GR clearing payable (not vendor AP).
   - `PI` clears GR clearing payable and recognizes vendor AP liability.
   - Service and other non-inventory expense invoices are direct invoice postings (no `GR`) under separate document types.
   - Purchase invoice does not require PO clearing in v1.
   - Purchase order closure is a separate manual action.
4. Inventory: item master, stock receipts/issues, valuation (weighted average in v1), warehouse support.
   - Goods inventory receipt is posted through `GR`; `PI` does not create a second inventory receipt.
5. Banking and cash books: receipt/payment workflows with reconciliation support.
6. GST workflows: tax determination, posting, return datasets, reconciliation diagnostics.
7. AI copilots: workflow assistants, data retrieval assistants, compliance assistants.
8. Instance administration and RBAC:
   - user creation and role assignment by `super_admin`,
   - company-code based access boundaries,
   - role-based permissions for workflow actions and reports.
9. Company context display requirements:
   - reports must display company code in header/context area,
   - transaction and journal-entry proposal screens must display active company code,
   - document print/export views must include company code metadata.
10. Dashboard and KPI control tower:
   - v1 includes a primary dashboard for business + system status.
   - dashboard must highlight pending actions, overdue items, and urgency-ranked exceptions.
   - dashboard must support agent-driven "what should I do next" recommendations with clear action links.

### Out of Scope (v1)

1. Full SAP-grade cross-module controls (complex release strategies, matrix pricing engines, deep transport layers).
2. Multi-country tax engines in v1.
3. Manufacturing MRP and production accounting in v1.
4. Deep CRM automation (lead scoring/campaign orchestration) in v1.
5. Full service-job lifecycle execution in v1 (kept as post-v1 expansion).

### Future Expansion Direction (Post-v1)

1. Service-business workflows:
   - service job/ticket lifecycle,
   - service delivery milestones,
   - service billing and revenue recognition hooks.
2. CRM workflows:
   - lead, contact, account timeline,
   - opportunity and quotation tracking,
   - activity logging (calls/meetings/tasks) with optional finance links.
3. Non-financial operational workflows:
   - workflow/task tracking and SLA checkpoints,
   - customer communication history,
   - policy-governed operational approvals beyond accounting.

## 4. "SAP Strengths, SMB Simplicity" Target

Adopt:

1. Document-type discipline with strict two-letter codes (SI, GR, PI, PE, PS, RC, PV, CN, DN, JE, etc.).
2. Deterministic posting rules (rule/policy engine).
3. Auditability with full traceability (who/what/why/when).
4. Reversal-first correction model.
5. Numbering and idempotency guarantees.

Avoid:

1. Excessive mandatory configuration before first transaction.
2. Rigid process steps that block SMB speed.
3. Internal-only accounting language in end-user workflow screens.

## 5. High-Level Architecture

## 5.1 System Shape

1. Modular monolith in Go for v1 (fast delivery, simpler operations, strong transactional boundaries).
2. Postgres as system of record.
3. Outbox/event log for reliable async processing and future service extraction.
4. Thin adapters (Web, API, CLI) over a strict application service layer.
5. Domain modules with explicit contracts and no adapter leakage.
6. Web adapter architecture:
   - server-rendered Go templates as baseline rendering model,
   - HTMX for partial updates and workflow-step interactions,
   - Alpine.js for lightweight local interactivity/state.
7. UI composition rule:
   - reusable template partials/components for KPI cards, workflow queues, approval panels, and document forms,
   - avoid page-specific business logic duplication.
8. Interface prioritization:
   - primary: web + agent interaction surfaces,
   - secondary: REPL/stateless CLI for testing and power usage.
9. Mobile readiness:
   - keep application-service/API contracts stable and DTO-first so Flutter client can consume the same workflow capabilities.
10. Deployment model:
   - local single-instance deployment for an individual company,
   - cloud single-instance deployment for an individual company,
   - optional multi-company configuration inside one deployed instance via company scoping.
11. Commercial model note: architecture supports multi-tenancy, but v1 planning is not SaaS-first.

## 5.2 Core Modules

1. `identity_access` (users, roles, company-code scoping, instance `super_admin` capabilities)
2. `master_data` (customers, vendors, items, accounts, tax profiles)
3. `documents` (document lifecycle, numbering, references)
4. `workflow_engine` (state machine transitions + policy checks)
5. `ledger` (journal entries/lines, reversals, balances)
6. `tax_gst` (tax determination, postings, filing datasets, reconciliation)
7. `inventory` (stock, valuation, movements)
8. `payments` (cash/bank, allocations, settlement)
9. `reporting` (financial + operational + tax diagnostics)
10. `dashboard_insights` (KPI cards, pending queues, urgency scoring, recommended actions)
11. `agent_platform` (orchestration, tool contracts, guardrails, observability)
12. `service_ops` (service jobs, milestones, service delivery artifacts) [post-v1]
13. `crm` (leads, contacts, opportunities, activities) [post-v1]

## 5.3 Foundation Rule

No UI adapter may import domain internals directly.  
All adapters must depend only on application DTO/contracts.

## 5.4 Access Control Model (RBAC)

1. `super_admin` (instance scope):
   - create users,
   - assign/revoke roles,
   - manage company access mappings and security settings.
2. `finance_manager`:
   - submit/approve/post/reverse financial documents per policy,
   - access financial and compliance reports.
3. `accountant`:
   - create and update operational/financial documents within role limits,
   - execute day-to-day accounting and reconciliation workflows,
   - may be configured as submit-only (no post permission) by role policy.
4. `auditor` (view-only):
   - read-only access to books, workflow history, and audit trails,
   - no posting, approval, or configuration rights.
5. Role assignments are enforced with company-code boundaries and full audit logs.
6. Access-denied and validation messages should include company context where relevant (without leaking other-company data).
7. Role keys are canonical lowercase in both DB and API contracts:
   - `super_admin`, `finance_manager`, `accountant`, `auditor`.
8. Authorization must be enforced at application command/use-case boundary (server-side), not only at web route middleware.
9. Adapter-level role checks remain as defense-in-depth, but cannot be the only enforcement layer.
10. Authorization decisions must be auditable with actor, role, company, action, and decision reason metadata.

## 6. Foundational Data Model (Extensibility-First)

## 6.1 Core Design Decisions

1. Every row is tenant/company scoped.
2. Separate business documents from accounting documents/journals.
3. Explicit reference links (`source_document_type`, `source_document_id`) for traceability.
4. Policy tables for behavior variation, not code branching.
5. Temporal validity (`effective_from`, `effective_to`) for rule/version changes.
6. Support nullable optional columns with companion policy flags instead of "future hard migrations everywhere".

## 6.2 Recommended Table Families

1. `companies`, `company_settings`, `financial_periods`
2. `users`, `roles`, `user_company_roles`
3. `accounts`, `account_groups`, `account_rules`, `posting_policies`
4. `document_types`, `document_sequences`, `documents`, `document_links`
5. `delivery_notes`, `delivery_note_lines`, `sales_invoices`, `sales_invoice_lines`, `goods_receipt_notes`, `goods_receipt_note_lines`, `purchase_invoices`, `purchase_invoice_lines`
6. `receipts`, `receipt_allocations`, `vendor_payments`, `vendor_payment_allocations`
7. `items`, `warehouses`, `inventory_items`, `inventory_movements`, `inventory_valuation_layers`
8. `journal_entries`, `journal_lines`, `journal_entry_links`
9. `tax_codes`, `tax_rates`, `tax_rules`, `tax_determination_logs`, `tax_line_details`
10. `gst_returns`, `gst_return_lines`, `gst_reconciliation_items`
11. `workflow_instances`, `workflow_tasks`, `workflow_transitions`
12. `agent_sessions`, `agent_actions`, `agent_tool_calls`, `agent_handoffs`, `agent_audits`
13. `idempotency_keys`, `integration_outbox`, `webhook_deliveries`
14. `period_statuses`, `period_lock_rules`, `period_override_audits`
15. `service_jobs`, `service_job_tasks`, `service_job_events`, `service_billing_links` [post-v1]
16. `crm_accounts`, `crm_contacts`, `crm_leads`, `crm_opportunities`, `crm_activities` [post-v1]

## 6.3 Posting Abstraction

Introduce a posting layer:

1. Business workflow emits a normalized `PostingIntent` object.
2. Posting engine resolves accounts/tax via policies.
3. Posting engine emits balanced `JournalEntry`.
4. Persist posting explanation snapshot (rules used + decision path).
5. Enforce subledger-control rules:
   - AR postings require customer context.
   - AP postings require vendor context.
   - Inventory postings require item/movement context.
   - Direct manual postings to control accounts are rejected by policy.

This isolates workflow expansion from ledger redesign.

## 6.4 Inventory Valuation Extensibility (Weighted Average -> FIFO Later)

1. Keep valuation method as company policy (`inventory_valuation_method`) not hardcoded in service logic.
2. Persist enough movement metadata to support multiple valuation strategies:
   - movement type, quantity in/out, unit cost, total cost, timestamp, source document link.
3. Introduce valuation layer storage from v1 (`inventory_valuation_layers`) even if weighted average is active.
4. Implement valuation engine interface:
   - `ComputeIssueCost(method, item, quantity, as_of)` with method-specific strategies.
5. For FIFO rollout later:
   - backfill/opening layers for existing stock,
   - revaluation adjustment entry policy for transition date,
   - per-tenant cutover controls.

Conclusion: adding FIFO later is feasible with moderate effort if the above foundations are included in v1.

## 7. GST-First Domain Model

## 7.1 GST Data Primitives

1. Party GST profile: registration type, GSTIN, place of supply defaults.
2. Item/service tax profile: HSN/SAC, default GST rate class, exempt/nil flags.
3. Tax code master: CGST/SGST/IGST/Cess combinations.
4. Supply context: intra/inter-state, B2B/B2C/export/SEZ/reverse-charge.
5. Tax calculation snapshot per line and header.

## 7.2 GST Workflow Coverage (v1)

1. Sales invoice with GST split logic.
2. DN to SI tax and quantity continuity for inventory-goods sales.
3. Purchase invoice with input credit eligibility flags.
4. GR to PI tax data continuity for goods procurements.
5. Reverse charge scenarios (policy-driven).
6. Credit/debit notes tax impact.
7. Return-ready data views (GSTR-1/GSTR-3B oriented datasets).
8. Reconciliation diagnostics:
   - Books vs return aggregates
   - Input tax credit exceptions
   - Missing/invalid GSTIN checks
9. Period-aware GST controls:
   - restrict GST-sensitive edits after soft close,
   - force adjustment-document path after hard close.

## 7.3 Compliance Design Rule

Tax determination must be reproducible:

1. Store tax rule version used.
2. Store exact basis values and rounding decisions.
3. Support recalculation only via explicit adjustment docs, never hidden rewrites.

## 8. Multi-Agent Architecture (Workflow-Focused)

Design rule:
1. Prefer workflow-specialized agents over broad module agents wherever practical.
2. Keep agent responsibilities narrow (single primary workflow family).
3. Allow a larger number of agents if that improves determinism, auditability, and testability.
4. Optimize architecture for long-lived modular upgrades: add/replace agents, tools, and policies without core workflow rewrites.

## 8.1 Agent Topology

1. `CoordinatorAgent`:
   - Intent routing
   - Context assembly
   - Safety policy enforcement
2. `RouterAgent` (optional):
   - deterministic rule-first workflow selection
   - ambiguity detection and fallback to `CoordinatorAgent` clarification flow
3. `GoodsProcurementAgent` (`GR`, `PI` flow)
4. `ExpenseProcurementAgent` (`PE` direct invoice flow)
5. `InventorySalesAgent` (`DN`, `SI` flow)
6. `ServiceSalesAgent` (`SE` direct invoice flow)
7. `ReceiptsAgent` (customer receipts + allocations)
8. `VendorPaymentsAgent` (vendor payments + allocations)
9. `InventoryAdjustmentsAgent` (stock adjustments/reclassifications per policy)
10. `GSTComplianceAgent`
11. `ReportingAnalystAgent`
12. `MasterDataAgent`

## 8.2 Agent Contract Model

1. All agents operate via typed tool schemas (JSON schema strict mode).
2. Read tools can execute autonomously.
3. Write tools are gated by:
   - policy checks
   - confidence threshold
   - human confirmation (configurable by risk level)
4. Agent handoffs are explicit records (`agent_handoffs` table).
5. Human approval state for write actions must be persisted in PostgreSQL (not in-memory process state) so confirmation survives restarts and supports multi-instance web deployment from day one.

## 8.3 Workflow Engine + Agents

Agents should not own process state.  
Workflow engine is source of truth for state transitions; agents only propose/execute allowed transition actions.

Procurement and sales specialized execution rules:
1. On "book goods purchase invoice" intent, `CoordinatorAgent` (or optional `RouterAgent`) hands off to `GoodsProcurementAgent`.
2. `GoodsProcurementAgent` executes a linked two-step posting sequence after human confirmation:
   - Post `GR` first.
   - Post `PI` second.
3. `PI` must reference the created/selected `GR` and perform clearing from GR clearing payable to vendor AP.
4. If a `PO` reference is present, `GoodsProcurementAgent` may load defaults from PO; if absent, workflow still proceeds.
5. On "record goods receipt without invoice" intent, `GoodsProcurementAgent` posts `GR` only and marks procurement as invoice-pending until PI is booked.
6. On "book service/expense vendor invoice" intent, `CoordinatorAgent` (or optional `RouterAgent`) hands off to `ExpenseProcurementAgent` which posts `PE` directly (no `GR` path).
7. On "book inventory sales invoice" intent, `CoordinatorAgent` (or optional `RouterAgent`) hands off to `InventorySalesAgent` for linked two-step posting:
   - Post `DN` first.
   - Post `SI` second.
8. `SI` must reference the created/selected `DN` to prevent duplicate inventory issue posting.
9. On "record delivery without invoice" intent, `InventorySalesAgent` posts `DN` only and marks sales as invoice-pending until SI is booked.
10. On "book service/other non-inventory sales invoice" intent, `CoordinatorAgent` (or optional `RouterAgent`) hands off to `ServiceSalesAgent` which posts `SE` directly (no `DN` path).
11. Specialized agents must have strict tool allowlists aligned to their workflow document types.
12. Cross-workflow actions require explicit handoff records and are never performed by implicit tool expansion.

## 8.4 AI-Primary Operating Model (Business Workflows)

1. Agent-first execution path is primary for operational workflows (procure-to-pay, order-to-cash, reconciliation, compliance checks).
2. Human interaction model is primarily:
   - provide intent/context,
   - resolve ambiguities/exceptions,
   - approve gated financial actions.
3. Manual UI/API workflow execution is required as business continuity fallback, not the primary operating path.
4. All core business actions must expose stable agent-consumable tool contracts before manual-UI optimization.

## 8.5 Modern Agent Pattern Requirements (Modular + Upgrade-Ready)

1. Use a planner-executor-verifier loop where appropriate:
   - planner decomposes workflow steps,
   - executor performs tool actions,
   - verifier checks policy/invariant satisfaction before completion.
2. Keep deterministic tool boundary:
   - business state transitions happen through workflow/posting services, not free-form agent reasoning.
3. Maintain short-lived execution context + persistent memory separation:
   - transient conversation state in session context,
   - durable business facts and audit records in PostgreSQL.
4. Version agent contracts explicitly:
   - tool schemas and response contracts include version metadata for safe upgrades.
5. Support model/provider evolution without workflow redesign:
   - agent orchestration contracts remain stable while model choice/config can change independently.
6. Require replayability:
   - each agent run can be reconstructed from inputs, tool calls, decisions, and policy versions.
7. Require graceful degradation:
   - if an agent/tool is unavailable, workflow can route to backup agent or deterministic manual fallback.

## 8.6 Safety and Reliability

1. Client-supplied stable idempotency key is required for all write actions.
2. Deterministic retry behavior for tool execution.
3. Full request/response/action audit logs.
4. Role-based tool permissions.
5. Fallback to deterministic non-agent flows for critical business operations.
6. V1 policy: all write actions remain human-gated via confirm/amend/cancel flow.
7. Missing idempotency key on external write APIs is a validation error; server-generated timestamp-based fallback keys are not allowed for external callers.

## 9. API and Contract Strategy

1. Versioned API from day one (`/api/v1`).
2. DTO-first contracts (never expose raw internal/domain structs).
3. Stable error envelope with machine-readable codes.
4. OpenAPI + contract tests for high-change workflows.
5. Event versioning for outbox payloads.
6. Agent tool contracts are first-class API artifacts with schema versioning and compatibility policy.
7. Mobile-client readiness:
   - keep endpoints stateless and token-based,
   - design response payloads for low-bandwidth consumption (summary + drill-down),
   - preserve stable pagination/filter/sort contracts for Flutter list/detail screens.
8. Write-command APIs must include actor context and explicit authorization evaluation before business execution.
9. Error catalog must include stable authorization/idempotency policy codes (for example: `FORBIDDEN`, `IDEMPOTENCY_REQUIRED`, `IDEMPOTENCY_CONFLICT`).

## 9.1 Dashboard and KPI Contract (Action-Oriented)

1. Provide a unified dashboard API/view model for:
   - business health (cash, receivables, payables, sales/purchase run-rate),
   - compliance health (GST filing readiness, reconciliation exceptions),
   - operations health (pending workflow counts, approval queues, failed agent actions).
2. KPI blocks must include drill-down links to actionable queues, not only static metrics.
3. Pending and urgency model must be explicit:
   - status buckets (`pending`, `due_today`, `overdue`, `blocked`),
   - severity buckets (`info`, `warning`, `critical`),
   - recommended next action and owner hint.
4. Dashboard must be tenant/company scoped and always display active company context.
5. Agent and manual workflows should both emit queue events so dashboard status remains consistent regardless of execution mode.

## 10. Testing and Quality Gates

1. Unit tests for policy engines, posting engine, and tax determination.
2. Integration tests per workflow against test DB.
3. Ledger invariants tests:
   - debit == credit
   - no cross-tenant posting leakage
4. GST scenario matrix tests (intra/inter-state, B2B/B2C, reverse charge, notes).
5. Agent simulation tests with golden tool-call traces.
6. Architecture tests:
   - adapter -> app only
   - no forbidden package dependencies

Release gate:

1. `go test -p 1 ./...` for DB-shared suites.
2. migration verification + DB health verification.
3. workflow regression suite.
4. agent safety suite.

## 11. Delivery Roadmap

## Phase 0: Foundations (4-6 weeks)

1. Repo skeleton, module boundaries, architectural guardrails.
2. Core schema v1 + migrations framework.
3. Posting engine + ledger invariants.
4. Document and sequence services.

## Phase 1: Business Core (6-8 weeks)

1. Sales + receipts workflows.
2. Purchase + vendor payment workflows.
3. Inventory basics and valuation.
4. Standard reports (TB, P&L, BS, AR/AP aging).

## Phase 2: GST Core (4-6 weeks)

1. GST master data and tax engine.
2. GST postings and reporting datasets.
3. Reconciliation diagnostics and exception queues.

## Phase 3: Multi-Agent Layer (4-6 weeks)

1. Coordinator + optional Router + first 4 specialized workflow agents:
   - `GoodsProcurementAgent`, `ExpenseProcurementAgent`, `InventorySalesAgent`, `ServiceSalesAgent`.
2. Tooling contracts, strict per-agent allowlists, guardrails, audit UI.
3. Extend specialized coverage to receipts, vendor payments, inventory adjustments, and GST workflows.

## Phase 4: SMB UX Hardening (3-5 weeks)

1. Guided setup wizard.
2. Workflow-centric UI simplification.
3. Rule recommendations and health diagnostics.

## Phase 5: business_day Expansion (Post-v1)

1. Service operations foundation:
   - service job lifecycle,
   - service-to-billing integration contracts.
2. CRM foundation:
   - accounts/contacts/leads/opportunities/activity timeline.
3. Cross-domain linking:
   - optional links from CRM and service events to documents and postings for full traceability.

## 11.1 Delivery Reality for One Developer + AI Tooling

1. Prefer thinner v1 scope with strong invariants over broad feature list.
2. Sequence work in strictly testable vertical slices:
   - schema + policies -> one workflow -> posting -> reports -> agent tooling.
3. Avoid parallel unfinished modules; complete one workflow end-to-end before next.
4. Keep "later-ready" extension points but defer non-critical complexity.
5. Plan explicit hardening time for migration scripts, diagnostics, and test reliability.

Recommended practical target: deliver robust v1 in fewer workflows first, then expand.

## 11.2 Recommended Workflow Posting Mode for SMB (`Save`, `Submit`, and `Post`)

1. Yes, support explicit `Submit` (parking) and `Post` in v1.
2. Enforced state sequence:
   - `Draft` (editable, no ledger impact)
   - `Submitted` (parked/ready for posting; no ledger impact)
   - `Posted` (immutable financial impact)
   - `Cancelled`/`Reversed` for correction paths
3. Transition rule:
   - no document may enter `Posted` unless it has entered `Submitted` first.
4. UX behavior:
   - users without post permission can only move `Draft -> Submitted`,
   - if a user with post permission clicks `Post` on a `Draft`, system performs `Submit` then `Post` in one action sequence.
5. Why this is useful for SMB:
   - allows invoice preparation and parking before final confirmation,
   - supports maker-checker behavior even in small teams,
   - prevents accidental direct posting bypassing review intent.
6. Keep UX lightweight:
   - avoid SAP-style heavy approval chains by default,
   - optional "second approval required" policy can be added later per tenant.
7. Agent alignment:
   - agent can create/update draft and propose submit/post actions,
   - post execution remains policy-gated with explicit human confirmation in v1.

## 11.3 Invoice-First Operational Design (Confirmed)

1. Goods procurement posting uses mandatory two-document sequence in v1:
   - `GR` first (inventory/expense + GR clearing payable),
   - `PI` second (clear GR clearing payable + vendor AP).
2. PO is optional and informational/planning by default.
3. PO support is convenience mode:
   - "Load from PO" can prefill both GR and PI data,
   - no strict 3-way matching in v1 unless enabled later by policy.
4. PO close is manual and separate from GR/PI posting lifecycle.
5. `GR` and `PI` both include optional PO reference fields.
6. GR-only mode is supported for goods received before invoice receipt; later PI must reference/clear pending GR liability.
7. Service/other-expense vendor invoices are direct booking without `GR`, using separate non-inventory document types.
8. Service Receipt Note is planned for a future phase and is not in the current v1 workflow.
9. Inventory-goods sales posting uses two-document sequence in v1:
   - `DN` first (inventory issue/COGS impact),
   - `SI` second (AR + revenue + output GST).
10. DN-only mode is supported for goods delivered before invoice booking; later SI must reference pending DN.
11. SO is optional and informational/planning by default.
12. Service/other non-inventory sales invoices are direct booking without `DN`, using separate non-inventory document types.

## 11.3.1 Procurement Document Type Map (Two-Letter Policy)

1. `GR`: Goods Receipt (inventory receipt + GR clearing payable).
2. `PI`: Inventory Purchase Invoice (must clear linked `GR` liability to vendor AP).
3. `PE`: Expense Invoice (non-inventory direct AP booking; no `GR`).
4. `PS`: Service Invoice with Service Receipt linkage (future; reserved document type).
5. V1 active procurement document types: `GR`, `PI`, `PE`.
6. V1 future-reserved procurement document type: `PS`.

## 11.3.2 Sales Document Type Map (Two-Letter Policy)

1. `DN`: Delivery Note (inventory issue + quantity fulfillment reference).
2. `SI`: Inventory Sales Invoice (customer AR + revenue/output tax; references `DN`).
3. `SE`: Sales Expense/Service Invoice (non-inventory direct AR/revenue posting; no `DN`).
4. `SS`: Service Invoice with service-delivery linkage (future; reserved document type).
5. V1 active sales document types: `DN`, `SI`, `SE`.
6. V1 future-reserved sales document type: `SS`.

## 11.4 Period Close Model for SMB + GST

1. Period states:
   - `OPEN`
   - `SOFT_CLOSED`
   - `HARD_CLOSED`
2. Mandatory pre-close validation (for both `SOFT_CLOSED` and `HARD_CLOSED` transitions):
   - all documents in the target period must be `Posted` or `Cancelled`/`Reversed` per policy,
   - period close is blocked if any `Submitted` but not `Posted` document exists,
   - period close is blocked if any `Draft` document is still open in the target period.
3. `SOFT_CLOSED` behavior:
   - high-privilege role can edit only whitelisted fields,
   - mandatory reason capture and audit record,
   - GST-impacting field edits trigger tax-delta warning.
4. `HARD_CLOSED` behavior:
   - no direct edits to posted financial documents,
   - corrections only via reversal/credit note/debit note/adjustment JE per policy.
5. GST recommendation:
   - lock periods used in filed returns to `HARD_CLOSED`,
   - keep current month `OPEN`, previous month `SOFT_CLOSED` until return filing cutover.

## 12. Migration Strategy from Current App

1. Do not big-bang migrate live users in one step.
2. Build import bridge:
   - master data import
   - open balances
   - open AR/AP items
   - inventory opening
3. Dual-run pilot for selected tenants.
4. Reconciliation sign-off before cutover.

## 12.1 Legacy Capture Before Repo Cutover (Mandatory)

Because the current `accounting-agent-app` repository may not remain easily accessible during `business_day` implementation, capture the following as baseline references before cutover:

1. Current operational document types and numbering behavior used in production-like flows.
2. Current control-account governance behavior:
   - warning/enforce behavior,
   - override/audit expectations,
   - reconciliation diagnostics outputs.
3. Current document-type governance behavior:
   - policy mode behavior (`off|warn|enforce`),
   - violation audit payload shape and reporting expectations.
4. Current idempotency behavior and known weak points:
   - where idempotency is enforced strongly,
   - where fallback key generation exists and must not be carried forward.
5. Current integration test and DB validation gate behavior:
   - migration verification flow,
   - DB health verification flow,
   - serial integration test execution expectations.
6. Current AI integration constraints to carry forward:
   - OpenAI Responses API usage pattern,
   - strict tool-schema contracts,
   - human-gated write model.

Recommended artifact set to copy into new repo (`docs/legacy_reference/`):
1. Schema snapshot (tables, indexes, constraints) and migration manifest.
2. Policy behavior matrix (control-account + document-type governance).
3. Report contract samples (JSON/CSV examples) for key diagnostics.
4. Tool contract catalog (read/write tool names, input schemas, expected outputs).
5. Runbook commands currently used for DB verify, DB health, and test gating.

## 13. Non-Functional Requirements

1. Performance baseline for v1 SMB target:
   - Designed for up to 1,000 invoices/month per tenant.
   - Up to 3 concurrent accounting/inventory users per tenant.
   - sub-300ms p95 for common reads; controlled async for heavy reports.
2. Security: tenant isolation guarantees, encrypted secrets, RBAC + audit.
3. Operability: structured logs, trace IDs, metrics, SLO dashboards.
4. Recoverability: PITR-capable DB backups, tested restore playbooks.
5. Explainability: "show me why" endpoint for all auto postings and tax decisions.
6. Cross-domain auditability: finance and non-finance workflow actions share actor/timestamp/reason trace standards.
7. Context safety: UI/API responses for transaction/report/proposal reads must carry active company code context fields.
8. Agent operability: monitor agent success rate, fallback rate, approval latency, and workflow completion SLA as primary runtime KPIs.
9. Dashboard freshness and usability:
   - near-real-time refresh for operational queues,
   - predictable refresh SLA for heavy KPI aggregates,
   - mobile-friendly and desktop-friendly visibility for urgency indicators.
10. Browser security hardening:
   - CSRF protection required for cookie-authenticated state-changing endpoints,
   - origin/referer validation and anti-CSRF token strategy must be defined before go-live.

## 14. Key Risks and Mitigations

1. Overengineering early:
   - Mitigation: modular monolith first, strict v1 scope.
2. GST rule complexity growth:
   - Mitigation: rule tables + versioned determination logs.
3. Agent unpredictability:
   - Mitigation: typed tool contracts + workflow engine authority + human gates.
4. SMB adoption friction:
   - Mitigation: progressive setup defaults + language simplification + guided tasks.

## 15. Decisions Needed Before Build Start

1. Timeline and phase targets based on one-developer capacity.

Resolved decisions (2026-03-07):
1. Support multi-instance web deployment from day one.
2. Store role names as canonical lowercase in DB and API contracts.

## 16. Immediate Next Step (Recommended)

Create and sign off an Architecture Decision Record (ADR) set before coding:

1. ADR-001: Modular monolith boundaries
2. ADR-002: Posting engine contract and invariants
3. ADR-003: GST tax determination model
4. ADR-004: Multi-agent orchestration and safety model
5. ADR-005: Tenanting and migration/cutover strategy
6. ADR-006: Workflow sequencing policy (PO/SO optional reference model)
7. ADR-007: Period close and GST lock policy

## 17. Pre-Backlog Analysis Checklist (Deferred)

Before converting this blueprint into a milestone-by-milestone execution backlog, complete these design reviews:

1. Data model deep-dive:
   - document tables,
   - subledger linkage model,
   - period lock/override tables,
   - GST tax snapshot and determination-log schema.
2. Workflow contracts:
   - exact state machines for GR, PI, DN, SI, Payment, Credit Note, Debit Note, PO, SO.
3. Posting matrix:
   - document type -> allowed posting patterns -> control account restrictions.
4. Period-close policy matrix:
   - fields editable in `SOFT_CLOSED`,
   - eligible roles,
   - mandatory reason/audit requirements.
5. Agent boundary design:
   - workflow-specialized agent responsibilities (split wherever practical),
   - tool availability per agent,
   - confirm/amend/cancel behavior by action type.
6. Reporting contract:
   - GST reconciliation outputs,
   - subledger reconciliation outputs,
   - month-end operational diagnostics.
7. Post-v1 expansion contracts:
   - service-job lifecycle/state model,
   - CRM entity model and optional links to finance workflows.

## 18. Implementation Readiness Notes (Added 2026-03-07)

The blueprint is strong directionally, but implementation should start only after the following are completed:

1. Status and sign-off:
   - move status from draft to approved-for-build,
   - complete and sign off ADR-001 to ADR-007.
2. Pre-backlog items currently deferred must be completed before build slicing:
   - exact workflow state machines for GR/PI/DN/SI/Payment/Credit Note/Debit Note/PO/SO,
   - posting matrix (document type -> allowed posting patterns -> control account restrictions),
   - period-close policy matrix and role/field edit boundaries.
3. Data model specifications must be promoted from table-family level to implementation schema contracts:
   - required columns, data types, nullability,
   - PK/FK/unique/check constraints,
   - critical indexes and tenancy/company isolation constraints.
4. API contract strategy must be expanded to endpoint-level contracts:
   - per-endpoint request/response DTOs,
   - stable machine-readable error code catalog,
   - OpenAPI-first review for high-change workflows.
5. Delivery and quality gates should include explicit acceptance criteria per phase/workflow:
   - functional acceptance conditions,
   - invariants and regression expectations,
   - minimum automated test coverage required for completion.
6. Legacy baseline capture must be complete before current-repo freeze:
   - copy required legacy reference artifacts into the new repo,
   - record known defects intentionally not carried forward,
   - mark which legacy behaviors are compatibility requirements vs deliberate redesign.

## 19. PostgreSQL-First Enforcement Model (SQL-First, AI-Agent-First, Business-First)

Design intent:
1. Keep PostgreSQL as an active business-rules engine, not passive storage.
2. Push core integrity controls to schema constraints, indexes, and transactional SQL functions.
3. Keep AI agents bounded by database-enforced safety controls.
4. Apply these controls as default guardrails, not rigid doctrine, with explicit and auditable deviation paths where implementation realities require it.

### 19.1 Tenant Isolation by Construction

1. Mandatory `org_id` on every business, accounting, policy, and audit table.
2. Prefer composite tenancy-safe keys and references:
   - parent identity: `(org_id, id)`,
   - child foreign keys must include `org_id` and match parent `org_id`.
3. Enable row-level security (RLS) for tenant-scoped tables in production deployment modes.
4. Enforce company context in every write path:
   - API/session sets active `org_id`,
   - SQL write functions assert active `org_id` consistency.
5. Add migration-time verification query that fails if any required table is missing `org_id`, PK/UK tenancy scope, or required FK tenancy scope.
6. Maintain a tenant-safe FK matrix as a tracked artifact:
   - every cross-table relationship must document tenancy columns and expected FK shape,
   - CI schema-lint check fails if a required FK omits tenancy key columns.

### 19.2 Ledger and Posting Invariants in Database

1. Balanced journal invariant:
   - enforce `sum(debit) = sum(credit)` per `journal_entry_id` using deferred constraint trigger at transaction commit.
2. Append-only financial records:
   - block update/delete on posted `journal_entries` and `journal_lines`,
   - corrections only through reversal/adjustment documents with explicit link fields.
3. Control-account restrictions:
   - reject direct manual postings to AR/AP/Inventory control accounts by policy-aware SQL guard function.
4. Period lock enforcement:
   - database function checks period state (`OPEN`, `SOFT_CLOSED`, `HARD_CLOSED`) before posting/editing,
   - hard-close blocks direct mutation and allows only approved correction document types.
5. Deterministic posting evidence:
   - store posting rule version, policy snapshot, and explanation payload with each posted document/journal.

### 19.3 Workflow Sequencing Constraints (GR/PI, DN/SI)

1. Model parent-child document line linkage tables for mandatory sequence pairs:
   - `gr_pi_line_links`,
   - `dn_si_line_links`.
2. Enforce one-way quantity ceilings in SQL:
   - invoiced quantity cannot exceed linked received/delivered quantity.
3. Enforce dependency checks before posting:
   - `PI` must reference valid `GR` (except policy-explicit non-inventory direct flows),
   - `SI` must reference valid `DN` for inventory-goods flow.
4. Enforce idempotent linkage uniqueness:
   - unique constraints to prevent duplicate linkage for same source-target combination.
5. Use transactional posting functions that create workflow documents + journal + links atomically.

### 19.4 Idempotency and Concurrency Control

1. All write commands require client-provided stable idempotency key persisted in `idempotency_keys`.
2. Recommended uniqueness shape:
   - `(org_id, action_type, actor_id, client_idempotency_key)`.
3. Store canonical request hash and resulting resource reference for replay-safe responses.
4. For inventory valuation and sequence generation, use explicit locking discipline:
   - `SELECT ... FOR UPDATE` on affected item/sequence rows in posting transactions.
5. Retries must be safe by design:
   - duplicate key on idempotency returns prior result, not a second posting.
6. External write APIs must reject missing keys and must not auto-generate timestamp-based idempotency values.

### 19.5 AI Agent Guardrails at Database Layer

1. Agent writes must call audited SQL functions/procedures, not raw table mutations.
2. Persist immutable agent audit chain:
   - `agent_sessions`, `agent_actions`, `agent_tool_calls`, `agent_handoffs`, `agent_audits`,
   - include correlation IDs and human confirmation references.
3. Introduce database-controlled write kill switch:
   - policy table keyed by `(org_id, workflow_family, tool_name, enabled)` checked inside write functions.
4. Enforce role-policy checks in SQL function preconditions for high-risk writes.
5. Preserve deterministic fallback:
   - non-agent workflow APIs call the same posting SQL functions used by agents.
6. Runtime policy decisions must come from versioned policy/config tables; environment variables are bootstrap defaults only and cannot be the sole source of business-policy enforcement.

### 19.6 Reporting, GST, and Performance with PostgreSQL Features

1. Build GST and reconciliation datasets as stable SQL views/materialized views with refresh strategy.
2. Use partial indexes for high-frequency status filters (for example `Submitted`, `Posted`, `invoice-pending` states).
3. Use generated columns where helpful for normalized search keys and reporting groupings.
4. Partition high-volume append tables by period when needed:
   - `journal_entries`, `journal_lines`, `inventory_movements`, `agent_tool_calls`.
5. Capture explainability data as queryable JSONB snapshots with targeted GIN indexes for diagnostics use-cases.

### 19.7 Migration and Rollout Rules (SQL-First)

1. Additive-first schema changes only; never rewrite applied migrations.
2. For online-safe evolution:
   - add nullable columns first,
   - backfill in controlled batches,
   - add constraints as `NOT VALID`,
   - validate after data cleanup.
3. Every migration affecting financial integrity must include:
   - rollback strategy notes,
   - post-migration verification SQL checks,
   - updated integration tests for affected invariants.
4. Required validation sequence for schema changes:
   - run `verify-db` on test DB,
   - run `verify-db-health` on test DB,
   - run serial test suite (`go test -p 1 ./...`),
   - run `verify-db` then `verify-db-health` on target cloud/live DB before marking complete.

### 19.8 Definition of Done for SQL-First Architecture

A workflow is not complete until:
1. Domain rules are enforced by constraints/functions in PostgreSQL (not only Go service logic).
2. Tenant isolation is verifiable by schema + RLS/policy tests.
3. Idempotency and retry behavior are validated with concurrent integration tests.
4. Agent and non-agent write paths converge on the same audited SQL posting functions.
5. Reporting/GST outputs are reproducible from stored snapshots and versioned policies.
6. Human approval/confirmation flows are durable across restarts and multi-instance deployments.
7. Authorization is enforced at application command boundary and validated by tests (not only adapter middleware tests).

### 19.9 Pragmatic Deviation Policy (Avoid Rigidity)

Principle:
1. SQL-first controls are the default architecture direction.
2. Teams may deviate when a control materially harms delivery, operability, or correctness in current phase constraints.
3. Deviation must preserve accounting integrity and tenant safety, even if implementation shape differs.

Deviation process:
1. Record deviation in ADR/change note with:
   - policy being deviated from,
   - reason and practical constraint,
   - risk and compensating controls,
   - planned convergence target (if temporary).
2. Classify deviations:
   - temporary (must include review date/exit criteria),
   - intentional long-term (must include justification and alternative controls).
3. Require explicit sign-off for high-risk deviations:
   - tenant isolation,
   - posting invariants,
   - period lock controls,
   - idempotency guarantees.
4. Track deviations in implementation backlog so they are visible and test-covered.

Allowed implementation flexibility examples:
1. If RLS adds unacceptable complexity in early phase, use strict app + SQL function scoping first, with RLS as scheduled hardening.
2. If a hard DB constraint causes unsafe migration risk, phase it with `NOT VALID` plus compensating verification queries until validation.
3. If a stored procedure boundary slows iteration for low-risk paths, allow app-layer orchestration temporarily while keeping invariant checks in DB.

Non-negotiables:
1. No deviation may allow unscoped cross-company access.
2. No deviation may allow unbalanced posted journal entries.
3. No deviation may bypass auditable confirmation for agent-initiated financial writes in v1 policy.
