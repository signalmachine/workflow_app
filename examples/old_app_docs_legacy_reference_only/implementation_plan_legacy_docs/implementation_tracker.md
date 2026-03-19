# service_day Implementation Tracker

Date: 2026-03-13
Status: Legacy execution record
Purpose: preserve the broader pre-thin-v1 implementation record and remediation history.

Legacy note:
1. active thin-v1 milestone status now lives in `plan_docs/service_day_refactor_tracker_v1.md`
2. keep this file for historical evidence, older remediation context, and already-shipped slice details

## Status Legend

1. `not_started`
2. `in_progress`
3. `blocked`
4. `done`

## Global Rules

1. update this file when milestone state changes
2. keep status factual; do not mark `done` without evidence
3. record blockers at the decision-gate level, not as vague notes
4. if scope changes, update the canonical planning document first, then update this tracker

## Current Snapshot

1. planning documentation is implementation-ready
2. implementation defaults and decision gates are locked
3. Milestone A implementation has started with a Go application skeleton, embedded migrations, and the first kernel schema pass
4. Slice 1 org-aware boot is in progress; health endpoints, PostgreSQL connectivity, bootstrap/login/session auth, and an authenticated identity endpoint are wired, bootstrap singleton enforcement now holds at the database layer under concurrent first-run attempts, but deeper authorization and broader org-scoped request handling are not complete yet
5. Milestone B has started with authenticated CRM accounts, contacts, notes, activities, communications, leads, opportunities, shared tasks, lead conversion, and a basic derived timeline on top of the kernel auth layer
6. Authenticated CRM quick search now covers accounts, contacts, leads, opportunities, activities, communications, and attachments with tenant-safe filtering; relationship timelines now also project shared tasks
7. CRM-specific tenancy remediation is complete for the currently shipped CRM scope: dirty-data migration upgrades fail fast with actionable diagnostics, non-polymorphic child tables are tenant-safe, and polymorphic CRM context links now resolve through the schema-backed `linked_records` registry; the newer platform/auth/audit hardening items remain open in the remediation plan below
8. Canonical `implementation_plan` review is complete for the reinforced database-first/SQL-first and AI-agent-first objectives, with follow-up corrections now also applied for portal-identity ownership, customer-lifecycle history canon, and analytic-dimension ownership
9. Canonical delivery wording now explicitly differentiates `project` as the coordination and aggregation layer from `work_order` as the concrete operational execution unit; projects remain optional and may group many work orders
10. Canonical strategy now states that UAE is the real commercial target market while India remains the first delivered compliance and payroll baseline
11. Payroll is now explicitly planned as a later SMB-focused capability: India first, then UAE on the same extensible worker/time/accounting core, rather than as a full-suite HR/payroll system
12. Opportunity-linked estimates now include bounded AI recommendation flows: estimate drafts can be generated from opportunity context, persisted through `agent_runs` and `agent_recommendations`, and then either accepted into live CRM estimates through the normal estimate write path or discarded while preserving approval and audit causation
13. The current implementation now has the originally planned correctness remediation for the current scope closed through API-boundary hardening: bootstrap singleton guarantees, audit-required identity/workflow/CRM atomic writes, tenant-safe identity linkage, workflow active-membership enforcement, malformed-identifier rejection, CRM invalid-input sentinel consistency, and bearer-scheme normalization are in place
14. The canonical audit policy is now explicit: every business-state change must emit an audit record atomically, while low-level technical writes are exempt only by documented category
15. Canonical CRM planning now makes customer lifecycle and quote management explicit as technically solid, extensible foundations rather than market-specific CRM mimicry; configurable progression and strict commercial invariants are the active design standard
16. Opportunity-linked AI follow-up recommendations are now available as bounded task suggestions with review-state transitions: authenticated users can persist next-step task proposals from CRM opportunity context through `agent_runs`, `agent_artifacts`, and `agent_recommendations`, then accept them into live shared workflow tasks or discard them while keeping AI causation in the audit trail
17. The current backend now has a broader but still incomplete mobile-readiness foundation: `/api/v1/...` aliases resolve to the current HTTP surface, API responses now stamp explicit v1 compatibility headers, core CRM list endpoints support `limit`, `updated_since`, and cursor pagination, authenticated `/api/device-sessions/login`, `/api/device-sessions/refresh`, `GET /api/device-sessions`, and `POST /api/device-sessions/{id}/revoke` now provide device-scoped session login, rotation, visibility, and revocation, authenticated `GET /api/notification-devices` plus `POST /api/notification-devices` add the first notification-device registration primitive on top of tenant-safe `notification_devices`, authenticated `GET /api/notifications` plus `POST /api/notifications/{id}/read` now expose the first user-visible notification inbox on top of persisted `notification_events` and `notification_deliveries`, cross-user workflow task assignment now fans out inbox plus queued push deliveries atomically with the task write, an optional background dispatcher can now deliver those queued push notifications through configured provider HTTP endpoints, and current idempotency conflicts now expose machine-readable retry headers/body hints while retry-safe coverage now also includes the current lead and estimate state-transition endpoints; richer push retry/backoff behavior and any remaining retry-prone endpoint coverage still remain open, so the tracker should not imply a fully mobile-ready backend yet
18. Canonical planning now states more explicitly that the product's strongest long-term capability should be `work_order` support, while CRM, projects, billing, and accounting are expected to integrate tightly around that execution core rather than behaving like separate apps
19. Canonical mobile planning now records Flutter as the intended first mobile client technology while keeping backend mobile-readiness requirements explicitly client-agnostic
20. The accounting kernel baseline is now explicitly locked in the implementation decisions: accounting owns posting truth, posting lifecycle is invariant-heavy, and future ledger writes must use explicit idempotent posting contracts rather than ad hoc module-local rules
21. Canonical inventory/materials planning now states more explicitly that equipment-heavy service delivery is in scope: sourced gear can be received, allocated or reserved, deployed to customer sites, cost-traced, and tracked as installed equipment without redefining the product as a procurement-first ERP
22. Canonical serviced-asset planning now states more explicitly that repair and maintenance businesses may need a first-class record of the vehicle, machine, or installed device being worked on, with durable history and linked installed-part traceability
23. Canonical delivery planning now also states more explicitly that construction-style subcontract execution such as electrical, plumbing, low-voltage/networking, and similar material-backed trades should fit the shared project/work_order/inventory model rather than requiring a separate product direction
24. Canonical roadmap now places WhatsApp as a later customer-facing channel capability at the end of the roadmap, after portal, communication truth, and approval/audit boundaries are already in place
25. Canonical inventory/commercial planning now states more explicitly that the product remains service-first, but the same foundation should support direct stocked-item sales and limited-depth small-trading-company usage without forcing a second inventory architecture
26. Canonical accounting planning now states more explicitly that the product remains operations-first but still requires a solid double-entry accounting foundation with balanced postings, explicit posting boundaries, and correction-safe ledger behavior
27. Canonical accounting planning now also locks a submit-versus-post control model: AI may propose and, by policy, submit entries, but final posting remains a human-controlled action for authorized finance users
28. Canonical accounting planning now also states more explicitly that standard accounting controls are in scope: document types, durable unique numbering, accounting periods, credit/debit notes, reversal/void controls, and reversing journals must fit the shared accounting core for both India and UAE operations
29. Canonical tax planning now also states more explicitly that GST Lite and TDS must be suitable for service-company billing/accounting workflows while remaining extensible for deeper future India and UAE tax features
30. Canonical analytics planning now also states more explicitly that business-efficiency improvement requires deliberate measurement/reporting design, with typed analytic dimensions preferred over a cost-center-only model and `cost_center` retained as one optional supported dimension type
31. Canonical planning now also records a later rental-operations extension: operator-led building or property leasing, rentable rooms/units, recurring occupant rent collection, owner settlement, and property operating-expense tracking should fit the shared finance core without displacing the service-first `work_order` product center
32. Canonical planning now also records a later approval-gated mobile speech-capture objective: an internal user should be able to speak a business event in a local Indian language such as Telugu or Kannada, review the rendered text in the same language, approve it, and only then send the approved transcript to the backend for interpretation and bounded processing
33. CRM customer lifecycle baseline is now implemented on accounts: schema-backed lifecycle stages are seeded per org, authenticated lifecycle-stage list/create and account-transition/history APIs are available, lead conversion and opportunity creation advance lifecycle state where relevant, and account timelines now project lifecycle transitions
34. Canonical planning now also records a later narrow marketplace-seller extension: a small seller sourcing branded goods from a contract manufacturer and selling through channels such as Amazon India should fit the shared item, inventory, billing, tax, and accounting foundations at limited depth without displacing the service-first `work_order` product center
35. Canonical planning now also records a later limited inventory-trading extension: a small operator buying stock for resale and selling directly to customers should fit the shared item, inventory, billing, tax, and accounting foundations at limited depth without displacing the service-first `work_order` product center
36. Canonical planning now also records end-user documentation as an implementation concern: real user-visible workflows should accumulate maintained guides under `docs/user_guides/` rather than relying only on developer-facing docs
37. Canonical planning now assigns later portal-facing identity records such as `external_principals` and `portal_memberships` to `identity_access`, adds explicit `account_lifecycle_history` canon to match the implemented CRM lifecycle baseline, and makes typed analytic-dimension configuration a reporting-owned shared schema concern rather than leaving it as unowned prose
38. Canonical interface planning now states explicitly that mobile clients and later web or portal clients are the intended product surfaces, while any CLI-style tooling remains narrow internal support tooling rather than a first-class product interface
39. Canonical planning now also defines the later web launch experience as activity-centered: a signed-in user should get a bounded pinned-tile home plus global search for actions and records, with tiles launching exact workflow or document screens directly rather than broad module home pages
40. The 2026-03-14 review memo is now archived under `docs/archive/review_remediation_2026_03_14.md`, while user workflow-guide sequencing is tracked separately in `implementation_plan/service_day_user_workflow_docs_plan_v1.md` so documentation planning does not get mixed into the code-remediation sequence
41. Canonical scope discipline is now explicit: the current implementation should stay service-business-first and should not broaden CRM or other modules early toward every business type unless that broader scope is already present in the approved plan; later universalization decisions remain deferred until after the current planned milestones are implemented
42. Canonical planning now also records later CSV and spreadsheet data exchange support, with an explicit early-foundation stance: import/export should land through reusable file transport, reporting/read-model, audit, and domain-service seams rather than through direct table-loading shortcuts
43. Canonical scope discipline now also states that CRM has higher near-term delivery priority than `projects`, project scope should not drift into advanced project-management depth during the current milestone path, and richer project features are a possible later `v2` planning topic rather than part of the current roadmap
44. Canonical workflow planning now also states more explicitly that task/activity management is a first-class efficiency capability across CRM, projects, work orders, and later site or asset workflows; tasks remain distinct from activities, support one primary person-or-team owner plus one primary business context, and may later expose secondary related links for cross-context visibility and analytics
45. Canonical terminology now defines `workflow`, `task`, and `activity` separately so process/state rules, accountable next actions, and factual event history do not drift into one overloaded concept
46. A focused future-session remediation design now exists in [implementation_plan/workflow_task_model_alignment_remediation_2026_03_16.md](/home/vinod/PROJECTS/service_day/implementation_plan/workflow_task_model_alignment_remediation_2026_03_16.md) for aligning the current code with the updated workflow model, including shared team records, person-or-team task ownership, secondary task related links, and follow-through on activity/task separation
47. A focused future-session remediation design now exists in [implementation_plan/estimate_line_shape_alignment_remediation_2026_03_16.md](/home/vinod/PROJECTS/service_day/implementation_plan/estimate_line_shape_alignment_remediation_2026_03_16.md) for broadening the current estimate line-type contract beyond only `service` and `milestone` so the commercial baseline remains compatible with scoped-work and stocked-item quoting

## Decision Gates

| ID | Decision Gate | Status | Notes |
| --- | --- | --- | --- |
| DG-01 | Shared task engine direction | done | Locked in `implementation_decisions_v1.md` |
| DG-02 | Delivery milestone vs billing milestone ownership | done | Locked in `implementation_decisions_v1.md` |
| DG-03 | Commercial document strategy for estimate/invoice overlap | done | Shared conventions, separate ownership in v1 |
| DG-04 | Communication system-of-record strategy | done | Manual-first, import-ready |
| DG-05 | Worker and time-entry baseline model | done | Locked in `implementation_decisions_v1.md` |
| DG-06 | AI write-boundary and approval model | done | Locked in `implementation_decisions_v1.md` |
| DG-07 | Task primary-context rule | done | One primary context per task; secondary links may exist for visibility/analytics without replacing primary ownership |
| DG-08 | Service-request intake shape | done | Intake uses early `work_order` statuses in v1 |
| DG-09 | Status lookup-table vs enum policy | done | Hybrid policy locked in `implementation_decisions_v1.md` |
| DG-10 | Timesheet depth timing | done | Milestone C uses `time_entries` plus light approval |
| DG-11 | Payroll expansion posture | done | Later SMB-focused payroll capability; India first, UAE next, on top of the shared worker/time/accounting foundation |
| DG-12 | Mobile/backend readiness stance | done | Locked in `implementation_decisions_v1.md` |
| DG-13 | Accounting kernel baseline and posting boundary | done | Locked in `implementation_decisions_v1.md` |
| DG-14 | Inventory/materials deployment model for equipment-backed service delivery | done | Locked in `implementation_decisions_v1.md` |
| DG-15 | Serviced-asset and equipment-history model | done | Locked in `implementation_decisions_v1.md` |
| DG-16 | Mobile local-language speech-capture boundary | done | Approved transcript is the only backend command source; locale metadata, revision-specific approval, optional raw-audio retention, and idempotent submission rules are locked in `implementation_decisions_v1.md` |
| DG-17 | Activity-centered web launch and global-search model | done | Locked in `service_day_navigation_launch_experience_v1.md` and `implementation_decisions_v1.md` |
| DG-18 | Controlled CSV/spreadsheet data-exchange posture | done | Locked in `service_day_data_exchange_plan_v1.md` and `implementation_decisions_v1.md` |

## Milestone Tracker

| Milestone | Objective | Status | Depends On | Exit Evidence |
| --- | --- | --- | --- | --- |
| A | Kernel and schema foundation | in_progress | DG-01 to DG-06 and DG-13 done | App boots, tenant-safe auth works, bootstrap singleton guarantees hold, auditable writes are atomic, accounting kernel invariants and posting boundaries are explicit, and core schema is merged |
| B | CRM core vertical slice | in_progress | A materially complete enough to support CRM slice execution | Lead-to-opportunity-to-estimate workflow demonstrated |
| C | Delivery operations vertical slice | not_started | B done | Won work reaches executable delivery and cost capture |
| D | Billing and financial traceability | not_started | C done | Invoice-to-receipt-to-ledger traceability demonstrated |
| E | Compliance and portal readiness | not_started | D done | India-first compliance baseline, portal API foundation, and the broader mobile/backend readiness contract are demonstrated while keeping UAE adaptation straightforward |
| F | Lightweight payroll extension | not_started | C and D foundations substantially ready | India-first SMB payroll baseline works on top of worker/time/accounting foundations, and UAE payroll can follow without redesign |
| G | Customer messaging channel extensions | not_started | E done | Customer-facing WhatsApp support is added on top of the canonical communication model without creating a parallel channel-specific truth store |
| H | Rental operations extension | not_started | D done | Later operator-led rental support reuses the shared finance, party, and site/unit foundations for occupant billing, owner settlement, and property-level profitability without displacing the service-first `work_order` core |
| I | Limited inventory-trading extension | not_started | D and E foundations materially ready | Later limited direct-sale and stock-for-resale support reuses the shared item, inventory, billing, tax, and accounting foundations for receipt-to-sale-to-return-to-ledger traceability without turning the product into a trading-first ERP |
| J | Narrow marketplace-seller extension | not_started | I and E foundations materially ready | Later limited small-seller support reuses the shared item, inventory, billing, tax, and accounting foundations for channel orders, returns, settlements, and sourced-finished-goods traceability without turning the product into a marketplace-first ERP |

## Slice Tracker

| Slice | Outcome | Status | Milestone | Notes |
| --- | --- | --- | --- | --- |
| S1 | Org-aware boot | in_progress | A | HTTP boot, PostgreSQL connection, migration runner, bootstrap/login/session auth, and authenticated `GET /api/me` added; deeper authorization still pending |
| S2 | Relationship workspace | in_progress | B | Authenticated accounts, contacts, notes, activities, communications, attachments, account/contact/lead/opportunity timeline flows, and CRM quick search are implemented; both non-polymorphic child tables and polymorphic CRM context links now have schema-backed tenant enforcement for the currently supported linked record types, and CRM request-validation paths now reject malformed UUID-backed filters as `400 invalid_input` before store-layer casts |
| S3 | Lead to opportunity | in_progress | B | Authenticated leads create/list, qualify/disqualify, convert, opportunities create/list/update, shared workflow task create/list/complete flows, and AI next-step task recommendations from opportunity context are added; task recommendations can now be accepted into live workflow tasks through persisted approval records and tool-policy seeding or discarded with AI causation captured in audit metadata, while lead-context AI suggestions remain pending and workflow/AI identifier validation now rejects malformed UUID-like inputs before store casts; the current idempotent retry contract now covers task creation, opportunity updates, and AI recommendation acceptance rather than only the earlier safe transition endpoints |
| S4 | Opportunity to estimate | in_progress | B | Authenticated opportunity-linked estimate create/list, revision-safe revise, and final-ready flows are added with service or milestone lines, tenant-safe schema enforcement, commercial numbering and dates, currency/tax/discount persistence, timeline/search projection, persisted AI estimate-draft recommendations that can now be accepted into live CRM estimates or discarded with approval and audit causation retained, and replay-safe estimate creation under scoped `Idempotency-Key` handling |
| S5 | Won work to project/work order | not_started | C | No loss of CRM context |
| S6 | Delivery cost capture | not_started | C | Time, materials, expenses, costing |
| S7 | Bill and collect | not_started | D | Invoice, receipt, receivable visibility |

## Workstream Checklist

### Foundation

| Item | Status | Notes |
| --- | --- | --- |
| Repository skeleton and module layout | done | Go module, `cmd/app`, `cmd/migrate`, and internal platform packages added |
| Migration framework | done | Embedded SQL migration runner added under `internal/platform/migrations` |
| Org, user, role, membership model | in_progress | Baseline tables plus bootstrap and login flows added; `000013_identity_tenant_safety.up.sql` now enforces tenant-safe membership-role and auth-session-membership linkage, `000014_bootstrap_singleton.up.sql` adds the schema-backed one-time bootstrap gate, and `000019_identity_device_sessions.up.sql` now adds tenant-safe `device_sessions` and `refresh_tokens` plus device-linked auth sessions for external-client auth lifecycle work |
| Authorization wrappers | in_progress | Bearer-token auth middleware and identity context added; auth header parsing now accepts valid bearer-scheme case variants consistently, while role-aware policy enforcement is still pending |
| Audit events and idempotency | in_progress | Audit-required identity, workflow, and CRM writes now persist atomically with their audit events; `000018_idempotency_scope_and_recovery.up.sql` scopes idempotency keys per operation and adds lease-based recovery, and the first retry-safe `Idempotency-Key` boundaries now cover workflow task creation, CRM estimate creation, CRM opportunity updates, and AI recommendation acceptance while broader endpoint coverage still remains open |
| Parties and external principals | in_progress | Baseline tables added |
| Attachments foundation | in_progress | Metadata and linking tables added |
| Worker baseline | in_progress | `workers` and `worker_cost_rates` tables added |
| Payroll-ready workforce foundation | not_started | Worker, time, compensation, and accounting seams should remain compatible with later India-first and then UAE payroll support |
| Accounting kernel schema | in_progress | `ledger_accounts`, `journal_entries`, and `journal_lines` added; `000022_documents_kernel.up.sql` now also adds shared `documents`, backfills accounting journals into that canonical document kernel, and links current journal draft/post flows transactionally into shared document lifecycle state, while the repo still has no public accounting or documents HTTP surface |
| Canonical double-entry accounting posture clarified | done | Updated the initial plan, execution plan, schema foundation, implementation decisions, and tracker on 2026-03-13 so `service_day` remains service-operations-first while still requiring a technically solid double-entry accounting foundation rather than a lightweight finance approximation |
| Canonical accounting submit-versus-post control clarified | done | Updated the execution plan, schema foundation, AI architecture, module boundaries, implementation decisions, and tracker on 2026-03-13 so accounting entries follow a proposal/submission/posting lifecycle, AI may only propose or policy-allowed submit, and final posting remains a human-controlled action for authorized finance roles |
| Canonical standard accounting controls clarified | done | Updated the initial plan, execution plan, schema foundation, module boundaries, implementation decisions, and tracker on 2026-03-13 so standard accounting capabilities such as document types, durable numbering, accounting periods, credit/debit notes, reversal or void controls, and reversing journals are explicit accounting-core requirements rather than implied future enhancements |
| Canonical GST and TDS service-company baseline clarified | done | Updated the initial plan, execution plan, schema foundation, module boundaries, implementation decisions, and tracker on 2026-03-13 so GST Lite and TDS are explicitly service-company-suitable finance capabilities with room for deeper future India and UAE tax depth |
| Canonical analytics-dimensions posture clarified | done | Updated the initial plan, execution plan, schema foundation, module boundaries, implementation decisions, and tracker on 2026-03-13 so business-efficiency reporting is treated as a core product objective, typed analytic dimensions are preferred over a cost-center-only model, and `cost_center` remains one optional supported dimension type |
| AI run and approval schema shell | in_progress | Agent run, step, artifact, policy, approval, and recommendation tables added |
| Mobile/backend readiness | in_progress | `/api/v1/...` aliases now expose the current API surface under explicit `X-Service-Day-API-Version: v1` and `X-Service-Day-API-Compatibility: additive-within-v1` headers, accounts, contacts, leads, opportunities, and estimates support `limit`, `updated_since`, and cursor pagination, retry-safe idempotent writes now cover workflow task creation, CRM account/contact/lead/opportunity/estimate/communication/note/activity creation, CRM opportunity updates, and AI recommendation acceptance with scoped-key recovery semantics, authenticated `/api/device-sessions/login`, `/api/device-sessions/refresh`, `GET /api/device-sessions`, and `POST /api/device-sessions/{id}/revoke` now provide device-scoped session login, rotation, visibility, and revocation on tenant-safe `device_sessions` and `refresh_tokens`, authenticated `GET /api/notification-devices` plus `POST /api/notification-devices` add the first notification-device registration primitive on tenant-safe `notification_devices`, authenticated `GET /api/notifications` plus `POST /api/notifications/{id}/read` add the first user-visible notification inbox on persisted `notification_events` and `notification_deliveries`, cross-user workflow task assignment now fans out inbox plus queued push deliveries atomically with the task write, the app can now optionally dispatch queued push deliveries through configured provider HTTP endpoints, current idempotency conflicts now expose machine-readable retry guidance in headers and JSON bodies, and CRM attachments now have a server-owned `POST /api/attachments/upload` plus `GET /api/attachments/{id}/download` transport baseline backed by local filesystem storage in development; richer push retry/backoff behavior and any remaining retry-prone endpoint coverage still remain open |
| Canonical launch/home/search experience added | done | Updated the initial plan, execution plan, module boundaries, implementation decisions, user workflow docs plan, tracker, and a new canonical `service_day_navigation_launch_experience_v1.md` on 2026-03-16 so the later web client is expected to use a personalized pinned-tile home plus global action-and-record search, with tiles launching exact workflow or document screens directly rather than module home pages |
| Canonical data-exchange posture added | done | Updated the initial plan, execution plan, schema foundation, module boundaries, implementation decisions, tracker, and a new canonical `service_day_data_exchange_plan_v1.md` on 2026-03-16 so later CSV and spreadsheet import/export support is explicitly planned and the current implementation is expected to preserve the necessary file, audit, read-model, and domain-service seams early |
| User-guide baseline | in_progress | `docs/user_guides/README.md` is now added and the planning set treats user guides as an implementation deliverable; the first real shipped-workflow guides remain explicitly deferred and still missing for `admin_bootstrap_and_login.md`, `crm_accounts_and_contacts.md`, `lead_to_opportunity.md`, `estimates.md`, and `tasks_and_follow_up.md` while the auth, mobile-readiness, and CRM relationship flows settle enough to avoid immediate churn |
| Canonical inventory/materials delivery-model refinement | done | Updated the initial plan, execution plan, module boundaries, schema foundation, and implementation decisions on 2026-03-12 so service-linked inventory covers sourced equipment, allocation/reservation, serialized-unit tracking where needed, installed customer-site assets, and billable vs non-billable material traceability without broadening the product into a procurement-first ERP |
| Canonical service-first inventory posture clarified for direct item sales and small-trading-company fit | done | Updated the initial plan, CRM MVP scope, execution plan, module boundaries, schema foundation, implementation decisions, and tracker on 2026-03-13 so service companies remain the primary target, direct stocked-item sales are explicitly supported, and the same inventory foundation is expected to stay usable for small trading companies at limited depth without redefining v1 as a trading ERP |
| Canonical later inventory-trading extension added | done | Updated the initial plan, execution plan, module boundaries, schema foundation, implementation decisions, and tracker on 2026-03-18 so a later limited buy-hold-sell trading workflow now has explicit roadmap, ownership, schema, and milestone language distinct from the narrower marketplace-seller extension while still reusing the shared item, inventory, billing, tax, and accounting foundations |
| Canonical serviced-asset model refinement | done | Updated the initial plan, execution plan, module boundaries, schema foundation, and implementation decisions on 2026-03-12 so repair and maintenance workflows can link work orders to explicit serviced assets such as vehicles, installed equipment, or customer-owned machines rather than relying on account-level notes |
| Canonical construction-trade fit clarified | done | Updated the initial plan and execution plan on 2026-03-12 so construction subcontractors, including electrical, plumbing, low-voltage/networking, fit-out, and similar material-backed trades, are explicitly covered by the shared project/work_order/inventory direction |
| Canonical future customer WhatsApp roadmap entry added | done | Updated the initial plan and execution plan on 2026-03-12 so WhatsApp is explicitly planned as a later customer-facing communication/channel capability at the end of the roadmap rather than as part of the current CRM communications baseline |
| Canonical plan review for database-first and AI-agent-first alignment | done | Reviewed the canonical planning set on 2026-03-12; current execution, schema, CRM, module-boundary, and AI documents align with the reinforced objectives, and delivery wording now also states more explicitly that `project` is coordination/aggregation while `work_order` is the concrete execution unit |
| Canonical CRM scope refinement for lifecycle and quote extensibility | done | Updated the initial plan, CRM MVP scope, and schema foundation on 2026-03-12 so customer lifecycle is explicit, quote management is commercially stronger, and both stay generic, technically solid, and extensible rather than overfit to a specific regional CRM convention |
| Canonical later rental-operations extension added | done | Updated the initial plan, execution plan, schema foundation, and tracker on 2026-03-13 so operator-led rental businesses can be supported later through shared property/unit, recurring-charge, settlement, expense, and accounting foundations without shifting the product away from its service-first `work_order` core |
| Canonical mobile local-language speech-capture objective added | done | Updated the initial plan, execution plan, AI architecture, and tracker on 2026-03-13 so a later mobile flow can capture spoken business events in languages such as Telugu or Kannada, render the transcript back in the same language for approval, and only submit approved text to the backend for interpretation and bounded processing |
| Canonical narrow marketplace-seller extension added | done | Updated the initial plan, execution plan, module boundaries, schema foundation, implementation decisions, and tracker on 2026-03-13 so a later small Amazon-style seller workflow can reuse the shared item, inventory, billing, tax, and accounting foundations at limited depth without shifting the product away from its service-first `work_order` core |
| Canonical user-guide implementation expectation added | done | Updated the initial plan, execution plan, tracker, contributor guidance, and workflow docs on 2026-03-13 so `docs/user_guides/` exists, shipped user-visible workflows are expected to accumulate maintained end-user guides, and documentation close-out now includes user-guide updates where relevant |
| Canonical ownership and schema gaps closed after review | done | Updated the module boundaries, schema foundation, and tracker on 2026-03-13 so portal-facing identity records are owned by `identity_access`, `account_lifecycle_history` is part of the canonical CRM schema, `payroll` appears in the top-level bounded-context inventory, and typed analytic-dimension definitions now have explicit ownership and schema shape instead of remaining unowned planning prose |
| Canonical product-interface stance clarified | done | Updated the initial plan, execution plan, implementation decisions, and tracker on 2026-03-15 so mobile clients and later web or portal clients are the intended product surfaces, while any CLI-style commands remain narrow internal tooling for migration, verification, seeding, testing, or support rather than a first-class product interface |

### CRM MVP

| Item | Status | Notes |
| --- | --- | --- |
| Accounts and contacts | in_progress | Authenticated accounts and contacts create/list APIs added with tenant-scoped persistence and schema-backed cross-tenant link prevention for `account_contacts`; direct account and contact creation now also supports scoped `Idempotency-Key` replay for retry-safe duplicate suppression, contacts now surface explicit account-link arrays plus a replace-links API for primary designation behavior, and the direct-store `PrimaryAccountID` path now also maps cross-tenant account references back to the intended application-level `account not found` contract so the broad CRM integration suite is coherent again; fuller contact editing remains later work |
| Leads and conversion flow | in_progress | Authenticated leads create/list and qualify/disqualify/convert APIs added with tenant-scoped persistence; direct lead creation and the current lead status-transition endpoints now support scoped `Idempotency-Key` replay for retry-safe duplicate suppression, conversion now requires qualified leads, clears contradictory disqualification fields, and sits on top of the upgrade-safe tenant-remediation path |
| Opportunities and stages | in_progress | Opportunities and tenant-stage seeding added with create/list/update APIs, including audited stage, expected-close, probability, and owner changes; broader stage-management UI and policy-heavy progression rules still pending |
| Customer lifecycle baseline | done | Added schema-backed `account_lifecycle_stages` plus `account_lifecycle_history`, account lifecycle state on `accounts`, authenticated `/api/account-lifecycle-stages`, `/api/accounts/{id}/lifecycle`, and `/api/accounts/{id}/lifecycle-history` flows, lifecycle timeline projection, and automatic lifecycle advancement on lead conversion and opportunity creation; verified with `go build ./cmd/app ./cmd/migrate`, `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run TestStoreIntegration"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/platform/migrations"`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"` |
| Activities and notes | in_progress | Authenticated notes use shared `comments`; direct note and activity creation now support scoped `Idempotency-Key` replay for retry-safe duplicate suppression, and derived timeline APIs are added for account/contact/lead/opportunity context with schema-backed `linked_records` enforcement for supported linked record types and shared-task timeline projection |
| Communications baseline | done | Authenticated manual communications create/list APIs added with participant capture and derived timeline projection; direct communication creation now also supports scoped `Idempotency-Key` replay for retry-safe duplicate suppression, and `communication_participants` has tenant-safe schema enforcement against its parent communication and linked CRM target |
| Attachments baseline | done | Authenticated CRM attachment create/list APIs added on top of kernel `attachments` and `attachment_links`; timeline and search now project linked attachments, `attachment_links` now has tenant-safe schema enforcement against both its parent attachment and linked CRM target, and authenticated `POST /api/attachments/upload` plus `GET /api/attachments/{id}/download` now provide a first server-owned file transport contract with generated storage keys and opaque download URLs |
| Shared tasks | in_progress | Shared `workflow` tasks table plus authenticated `/api/tasks` create/list and `/api/tasks/{id}/complete` flows added for current CRM linked-record contexts; assignee, creator, and completion actor eligibility now require active memberships through both store validation and schema-backed task-write checks; direct task creation now supports scoped `Idempotency-Key` replay for retry-safe duplicate suppression, cross-user assignment now persists a notification inbox event plus queued push deliveries atomically with the task write and can dispatch those queued pushes when provider endpoints are configured, and AI task recommendations now persist through `agent_runs`, `agent_artifacts`, and `agent_recommendations`, with authenticated `/api/ai/task-proposals/{id}/accept` or `/discard` decisions able to create live workflow tasks or close the proposal while delivery-context reuse remains pending |
| Search baseline | done | Authenticated `/api/search` added with tenant-safe CRM search across accounts, contacts, leads, opportunities, activities, communications, and attachments |
| Estimates baseline | in_progress | Authenticated `/api/estimates` create/list, `/api/estimates/{id}/revise`, and `/api/estimates/{id}/finalize` flows now persist commercial numbering, issue/expiry dates, currency, tax, discount, and immutable revision lineage on top of service and milestone lines, tenant-safe schema enforcement, timeline/search projection, AI estimate-draft acceptance can now create or revise live CRM estimates through the owning estimate write path, and estimate creation plus the current revise/finalize transitions now support scoped `Idempotency-Key` replay for retry-safe duplicate suppression |
| AI summaries and drafting | in_progress | `/api/ai/estimate-drafts` and `/api/ai/task-proposals` now persist bounded estimate-draft and next-step task recommendations from CRM opportunity context through `agent_runs`, `agent_artifacts`, and `agent_recommendations`; estimate-draft and task-proposal acceptance now persist `agent_approvals`, seed the current tool-policy shell, carry `approval_id` plus AI causation into resulting audit metadata, and support idempotent retry replay on acceptance, while lead-context suggestions remain pending for CRM MVP acceptance |

## Active Remediation Plan

The following implementation corrections are required before Milestone A and the current Milestone B baseline can be treated as technically solid.

### RP-01 Audit-atomic write paths

Status:
1. `done`

Problem:
1. several CRM, workflow, and identity write paths commit the business row before or separately from the audit event
2. this allows persisted state to exist without the required audit trail when the audit insert fails
3. that violates the locked proposal-first and auditable-write boundary

Required changes:
1. move meaningful write paths and their audit-event persistence into the same database transaction
2. remove helper patterns that write business state through `db` and then attempt best-effort audit inserts afterward
3. review bootstrap, login session creation, CRM writes, workflow task writes, and estimate-finalization paths for the same bug class
4. codify and implement the boundary between audit-required business mutations and explicitly exempt technical writes

Required tests:
1. regression coverage proving an injected audit-write failure rolls back the primary business write
2. coverage for both create and state-transition paths, not only one happy-path insert

Acceptance criteria:
1. no meaningful write can succeed without its paired audit record
2. API responses and database state stay consistent under audit-write failure
3. the implemented audit policy is explicit enough that new slices can tell whether a write must emit an audit event without relying on guesswork

Implementation notes:
1. identity bootstrap and login session creation, workflow task create/complete, and the current CRM create/state-transition paths now write business state and audit events inside one database transaction
2. the shared CRM audit helper now writes through the active transaction executor rather than always using `db`
3. tagged PostgreSQL integration coverage now forces audit insert failures with a trigger and proves the primary business write rolls back for representative create and state-transition paths

### RP-02 Bootstrap singleton enforcement

Status:
1. `done`

Problem:
1. bootstrap currently relies on an application-level `HasAnyUser` pre-check
2. concurrent first-run requests can both pass that check and create multiple bootstrap tenants/admins
3. the one-time bootstrap rule is not enforced at the database layer

Required changes:
1. define one concrete database-enforced bootstrap singleton strategy
2. implement the bootstrap path so the singleton guarantee holds under concurrent requests, not only serial ones
3. document the chosen mechanism in the relevant schema or execution notes when it lands

Required tests:
1. concurrency-focused regression coverage or equivalent store-level proof for the singleton path
2. migration or schema verification proving the singleton guarantee is enforced by the database design

Acceptance criteria:
1. only one bootstrap initialization can ever succeed on an empty system
2. later bootstrap attempts fail deterministically without creating partial state

Implementation notes:
1. `000014_bootstrap_singleton.up.sql` adds a singleton `system_bootstrap` table and backfills it for already-initialized databases so future bootstrap attempts are schema-gated
2. `identityaccess.StoreDB.Bootstrap` now claims that singleton row inside the bootstrap transaction and maps the database uniqueness conflict to `ErrBootstrapAlreadyCompleted`
3. tagged integration coverage now proves both raw schema enforcement of the singleton row and concurrent service-level bootstrap attempts where exactly one request succeeds

### RP-03 Tenant-safe identity relationships

Status:
1. `done`

Problem:
1. core identity tables still contain raw-id foreign keys that do not enforce tenant ownership consistently
2. `user_org_memberships.role_id` and `auth_sessions.membership_id` can currently point at logically inconsistent rows
3. authenticated identity reads therefore trust relationships that the schema has not fully protected

Required changes:
1. add tenant-safe schema enforcement for identity relationships where the child row already carries tenant context
2. make session identity linkage prove that `org_id`, `user_id`, membership, and role all refer to the same tenant-owned membership chain
3. review identity queries after the schema fix so they rely on the enforced shape, not hopeful assumptions

Required tests:
1. regression integration coverage proving mismatched-org identity references are rejected at the schema layer
2. coverage proving valid login and session reads still work after the stronger constraints

Acceptance criteria:
1. identity and session rows cannot express cross-tenant or logically inconsistent linkage
2. tenant-safe identity guarantees are true in schema, not only in query conventions

Implementation notes:
1. `000013_identity_tenant_safety.up.sql` adds actionable dirty-data upgrade checks, composite uniqueness for tenant-safe reference targets, a composite membership-role foreign key, and a composite auth-session-membership-user foreign key
2. `identityaccess` login and session reads now join roles and memberships through the tenant-owned chain that the schema enforces
3. tagged integration coverage now proves both rejection of cross-tenant identity linkages and successful login/session reads under the stronger constraints

### RP-04 Workflow assignee eligibility

Status:
1. `done`

Problem:
1. workflow task creation currently accepts any org membership row
2. disabled or invited members can therefore be assigned live work
3. this conflicts with the active-user semantics used by authentication

Required changes:
1. decide and implement the v1 rule for who may be assignee, creator, or completion actor
2. at minimum, active memberships must be required where the task should be actionable
3. reflect the same rule in both store validation and schema constraints where practical

Required tests:
1. regression integration test proving invited or disabled memberships cannot receive new actionable tasks
2. coverage for completion attempts by ineligible actors

Acceptance criteria:
1. task assignment and completion semantics match active membership policy
2. workflow records do not target identities that cannot legitimately act in the system

Implementation notes:
1. workflow now treats assignee, creator, and completion actor as active-membership-only roles in v1
2. `000015_workflow_active_memberships.up.sql` adds a `tasks` trigger that rejects persisted task rows whose assignee or creator membership is not active
3. `workflow.StoreDB` now rejects non-active assignees, creators, and completion actors before write/transition attempts, and tagged integration coverage proves invited or disabled memberships cannot be assigned or complete work

### RP-05 HTTP error classification and exposure

Status:
1. `done`

Problem:
1. CRM, workflow, AI, and identity handlers currently collapse many unexpected failures into `400`
2. some handlers also return raw internal error text to clients
3. this makes operational failures look like caller mistakes and leaks implementation detail

Required changes:
1. introduce explicit handler error mapping for validation, not-found, transition, conflict, and internal failures
2. stop returning raw store or SQL error text in public API responses
3. add or update tests so unexpected backend failures surface as stable internal-error responses

Required tests:
1. HTTP handler tests proving internal store failures return `500` with stable error codes
2. handler tests proving domain validation and not-found errors still map to the intended status codes

Acceptance criteria:
1. client-visible errors are stable and intentionally classified
2. internal infrastructure and SQL details are not exposed in API payloads

Implementation notes:
1. identity, CRM, workflow, and AI handlers now map validation, not-found, transition, and conflict cases explicitly and return `internal_error` for unexpected backend failures
2. identity and AI services now expose package-level invalid-input sentinels so handler classification does not depend on matching raw validation strings
3. handler tests now prove both stable `500 internal_error` responses for unexpected backend failures and preservation of intended domain-specific statuses for representative validation, not-found, and transition cases
4. the follow-up hardening pass has now landed for CRM validation sentinels, malformed identifier rejection ahead of database `::uuid` casts, and case-insensitive bearer-scheme parsing

### RP-06 API validation and identifier-boundary hardening

Status:
1. `done`

Problem:
1. several CRM service methods still return plain validation errors rather than package-level invalid-input sentinels, so handlers classify caller mistakes as `500 internal_error`
2. CRM, workflow, and AI request paths accept non-empty identifier strings and let malformed UUID-like input reach store-layer `::uuid` casts, which again turns caller mistakes into `500`s
3. auth middleware currently requires the exact `Bearer ` scheme casing and rejects otherwise valid authorization headers that use a different case variant

Required changes:
1. normalize CRM validation failures onto explicit package-level invalid-input sentinels instead of ad hoc `errors.New(...)` values
2. introduce shared identifier validation at the service boundary for UUID-backed route params and query filters before store calls are attempted
3. apply that validation to CRM list/filter and state-transition paths, workflow task create/list/complete paths, and AI opportunity-scoped recommendation inputs
4. update auth middleware so bearer-scheme parsing is case-insensitive while still requiring a real bearer token
5. keep handler error mapping stable so malformed caller input yields intentional `400` or `401` responses rather than infrastructure-looking `500`s

Required tests:
1. handler tests proving malformed CRM search/filter and estimate/opportunity identifiers return `400 invalid_input`
2. workflow handler or service tests proving malformed `context_id`, `assignee_user_id`, and task IDs are rejected before store-layer SQL casts
3. AI handler or service tests proving malformed `opportunity_id` input returns `400 invalid_input`
4. auth middleware tests proving lowercase and mixed-case bearer schemes are accepted while empty or malformed auth headers still fail with `401`

Acceptance criteria:
1. malformed or missing caller input is classified as a caller error at the service boundary
2. database cast failures are no longer the normal path for invalid external identifiers
3. auth token parsing is interoperable with valid bearer-scheme case variants
4. Milestone B API hardening claims are only treated as complete after the new regression coverage is in place

Implementation notes:
1. CRM service validation failures now unwrap to the package-level `ErrInvalidInput` sentinel while preserving field-specific error text, so handlers classify caller mistakes as `400 invalid_input` instead of `500 internal_error`
2. shared UUID-shape validation now rejects malformed CRM route/query identifiers plus CRM/workflow/AI service-boundary identifiers before store-layer `::uuid` casts are attempted
3. workflow HTTP coverage now proves malformed task IDs are rejected at the handler boundary, and CRM/AI service and handler coverage now prove malformed estimate/opportunity/account identifiers surface as caller errors
4. identity auth middleware now accepts valid bearer-scheme case variants such as `bearer` while still rejecting missing or malformed bearer tokens

### RP-07 Mobile/backend readiness foundations

Status:
1. `in_progress`

Problem:
1. the current backend has versioned and paginated CRM list groundwork, but it still behaves like an internal web/API slice rather than a deliberate mobile-ready backend
2. the first device-scoped refresh/session lifecycle now exists, but logout/revocation visibility, richer per-device session management, and broader mobile-client auth ergonomics are still incomplete
3. list/read APIs now have an initial pagination and incremental-sync contract on the first CRM endpoints, while machine-readable client guidance exists on the covered retry-sensitive paths; explicit compatibility-policy wording was the remaining documented API-contract gap
4. notification and retry/idempotency contracts are not yet explicit enough for a production mobile client, and the new attachment transport baseline still needs later hardening such as object-store or presigned-transfer support if local filesystem storage stops being sufficient

Required changes:
1. extend the mobile auth model with per-device session visibility, logout/revocation, and clear token-expiry behavior beyond the current login/refresh baseline
2. define API versioning, deprecation, and compatibility rules for mobile-consumed endpoints
3. extend pagination and incremental-sync contracts to the main list/query endpoints that a mobile client will depend on first
4. add machine-readable error payload conventions and retryability expectations for client handling
5. harden the new attachment upload/download transport strategy for mobile networks and background transfers as storage needs mature
6. add notification/device-registration primitives for tasks, reminders, approvals, and key CRM events
7. add idempotency-key support where retry-prone mobile writes would otherwise create duplicate records or duplicate transitions
8. explicitly document whether the first mobile client is online-only or supports limited offline drafts/cache behavior

Required tests:
1. auth and store coverage for device-session revocation, expired-session handling, and any later logout/session-list flows on top of the current refresh-token rotation baseline
2. endpoint tests for pagination and sync-parameter validation on the first mobile-critical list APIs
3. attachment and idempotency tests proving retry-safe behavior under repeated client requests
4. notification registration tests covering token replacement and logout cleanup semantics

Acceptance criteria:
1. the backend can support a first internal Android client without ad hoc client-specific workarounds in auth, list loading, or file handling
2. mobile releases are protected by a declared API compatibility policy rather than best-effort assumptions
3. retry-prone mobile writes do not create unsafe duplicate business effects in the covered endpoints
4. the tracker does not claim mobile readiness until auth, API, attachment, and notification foundations are implemented with verification evidence

Implementation notes:
1. the current HTTP server now accepts `/api/v1/...` as a stable alias for the existing `/api/...` surface and stamps `X-Service-Day-API-Version: v1` plus `X-Service-Day-API-Compatibility: additive-within-v1` on API responses
2. `GET /api/accounts`, `GET /api/contacts`, `GET /api/leads`, `GET /api/opportunities`, and `GET /api/estimates` now expose a first mobile-facing list contract with validated `limit`, `updated_since`, and cursor parameters plus a `page` response object
3. CRM attachments now also expose a first server-owned binary transport baseline through authenticated `POST /api/attachments/upload` multipart upload and `GET /api/attachments/{id}/download`, with generated storage keys, checksum/size capture, and opaque download URLs backed by local filesystem storage in development
4. `POST /api/device-sessions/login` and `POST /api/device-sessions/refresh` now provide the first device-scoped session and refresh-token rotation baseline on top of tenant-safe `device_sessions` and `refresh_tokens`
5. authenticated `GET /api/device-sessions` and `POST /api/device-sessions/{id}/revoke` now expose the first per-device session visibility and logout/revocation management path, with revocation cascading to access sessions, active refresh tokens, and notification-device registrations
6. authenticated `GET /api/notification-devices` plus `POST /api/notification-devices` now add the first notification-device registration primitive, including registration-token replacement on the same session/provider pair and logout cleanup semantics through device-session-linked revocation
7. authenticated `GET /api/notifications` plus `POST /api/notifications/{id}/read` now add the first user-visible notification inbox, while cross-user workflow task assignment persists `notification_events` plus inbox and queued push-delivery bookkeeping atomically through `notification_deliveries`
8. current idempotency conflict responses now expose machine-readable retry guidance through `X-Service-Day-Retryable`, `Retry-After` on in-progress conflicts, and matching JSON fields so mobile clients can distinguish safe retry from non-retryable key reuse
9. CRM store queries now page and incrementally sync by `updated_at` with deterministic `updated_at DESC, id DESC` ordering, while lead state transitions now also maintain `updated_at` so sync consumers can observe those changes

### RP-08 CRM acceptance and public-contract drift

Status:
1. `done`

Problem:
1. top-level implementation summaries had read closer to CRM MVP acceptance than the actual shipped behavior warranted
2. the authenticated identity endpoint had exposed `GET /api/me` in documentation before the handler enforced that method shape

Required changes:
1. keep canonical CRM and contract requirements explicit, but align tracker and summary wording so the repo does not imply CRM MVP acceptance before the remaining identity-contract gap is closed
2. tighten the authenticated identity endpoint contract so non-`GET` requests to `/api/me` are rejected deliberately rather than succeeding accidentally

Required tests:
1. HTTP regression coverage proving `/api/me` rejects non-`GET` methods with `405 method_not_allowed`
2. documentation review coverage in the next close-out pass so README and tracker wording match the shipped slice precisely

Acceptance criteria:
1. the tracker no longer implies CRM MVP acceptance because the `/api/me` contract hardening is no longer missing
2. the documented `/api/me` method contract matches the actual runtime behavior

Implementation notes:
1. the canonical CRM MVP scope and schema foundation remain correct on lifecycle and estimate expectations; the estimate/commercial baseline and AI estimate-acceptance path are now implemented, and the formerly remaining gap in this remediation item was the small authenticated-identity method-contract leak, which is now closed
2. the customer lifecycle baseline is now in place through schema-backed `account_lifecycle_stages`, `accounts.lifecycle_stage_code`, `account_lifecycle_history`, authenticated lifecycle APIs, and lifecycle timeline projection
3. `/api/me` now rejects non-`GET` requests with `405 {"error":"method_not_allowed"}` before auth middleware runs, so the documented method contract matches runtime behavior again

### RP-09 Workflow task-model alignment to canonical person-or-team ownership

Status:
1. `in_progress`

Problem:
1. the updated canonical planning set now requires one primary person-or-team owner per task, optional secondary related links, and a stronger distinction between tasks and activities
2. the current implementation still uses `tasks.assignee_user_id` as a user-only ownership model, has no canonical `teams` or `team_members`, and has no `task_related_links`
3. CRM `activities` still carry some task-like due/owner/completion semantics, which creates drift from the clarified task-vs-activity model

Required changes:
1. implement shared `teams` and `team_members` records
2. evolve workflow task ownership from single-user assignee to an explicit primary owner model
3. add secondary task-related links without weakening the one-primary-context rule
4. update AI task acceptance, notifications, and reporting seams to the stronger task model
5. stop deepening CRM activity rows as the main accountable-follow-up record shape

Required tests:
1. migration and tenant-safety coverage for `teams`, `team_members`, and task-owner evolution
2. workflow integration coverage for person-owned and team-owned tasks
3. queue/claim/delegate behavior tests where team ownership is introduced
4. AI and notification regression coverage on the evolved task contract

Acceptance criteria:
1. the workflow model matches the current canonical plan rather than the earlier user-only assignee model
2. team ownership is represented through canonical shared records rather than freeform worker text
3. task primary ownership, primary context, and optional secondary links are all explicit and technically coherent
4. activities remain factual history while accountable follow-up lives in shared tasks

Implementation notes:
1. use [workflow_task_model_alignment_remediation_2026_03_16.md](/home/vinod/PROJECTS/service_day/implementation_plan/workflow_task_model_alignment_remediation_2026_03_16.md) as the detailed future-session implementation guide
2. this remediation should land before broader team-based mobile task flows, queue analytics, or deeper work-order execution surfaces depend on the older user-only task contract

### RP-10 Estimate line-shape alignment to canonical commercial breadth

Status:
1. `in_progress`

Problem:
1. the canonical commercial plan keeps the estimate baseline service-business-first, but still expects line shapes that can cover services, milestones, scoped work, and stocked items where direct item sales matter
2. the current implementation only accepts `service` and `milestone` line types in both CRM estimate handling and AI estimate-acceptance handling
3. this narrows the commercial contract more than the current plan intends, and that narrower limit was not previously indexed as an explicit deferred correction

Required changes:
1. broaden the canonical estimate line-type contract in code from the current two-type limit to the approved future-safe minimum set
2. keep CRM estimate create, revise, and AI acceptance flows aligned on the same line-type validation rules
3. update summary and tracker wording so the shipped estimate baseline is not described more narrowly than intended once the remediation lands

Required tests:
1. CRM service and integration coverage for the expanded supported line types
2. AI estimate-acceptance regression coverage for the same expanded line-type set
3. invalid-input coverage proving unknown line types still fail intentionally

Acceptance criteria:
1. the estimate model no longer hard-codes only `service` and `milestone` as the allowed commercial shapes
2. scoped-work and stocked-item quoting can fit the estimate baseline without forcing inventory or billing depth into the same session
3. the implemented contract matches the canonical CRM and commercial planning language

Implementation notes:
1. use [estimate_line_shape_alignment_remediation_2026_03_16.md](/home/vinod/PROJECTS/service_day/implementation_plan/estimate_line_shape_alignment_remediation_2026_03_16.md) as the detailed future-session guide
2. this should follow RP-09 or another closely related commercial hardening session rather than being coupled to inventory or billing milestone work by default

### Delivery

| Item | Status | Notes |
| --- | --- | --- |
| Projects | not_started | |
| Project milestones | not_started | |
| Work orders | not_started | |
| Work-order assignments | not_started | |
| Time entries | not_started | |
| Timesheet-ready approval baseline | not_started | Full timesheet workflow intentionally deferred |
| Expense capture | not_started | |
| Material issue and usage | not_started | |
| Operational costing views | not_started | |
| Change orders | not_started | |
| Payroll baseline | not_started | India-first SMB payroll later, then UAE payroll on the same extensible contracts |

### Billing and Finance

| Item | Status | Notes |
| --- | --- | --- |
| Invoices | not_started | |
| Billing milestones | not_started | |
| Receipts and allocations | not_started | |
| Accounting posting integration | not_started | |
| Customer balances | not_started | |
| Revenue and collections reporting | not_started | |

## Acceptance Evidence Log

Use this section to record concrete proof as milestones complete.

| Date | Item | Evidence |
| --- | --- | --- |
| 2026-03-12 | Planning set finalized | README, tracker, and implementation decisions added; task, intake, status, and timesheet decisions locked |
| 2026-03-12 | Planning set aligned for implementation | Shared `tasks` clarified as `workflow`-owned; relationship timeline clarified as derived; worker/time sequencing clarified across milestones |
| 2026-03-12 | Milestone A started | Go module, HTTP service skeleton, PostgreSQL connector, embedded migration runner, and initial kernel schema implemented; `go test ./...` passed |
| 2026-03-12 | Slice 1 auth baseline added | Bootstrap, login, auth session persistence, bearer-token middleware, and authenticated identity endpoint implemented; `go test ./...` passed |
| 2026-03-12 | Slice 2 CRM baseline started | Accounts, contacts, and account-contact linking schema added; authenticated `/api/accounts` and `/api/contacts` create/list flows implemented; `go test ./...` passed |
| 2026-03-12 | Slice 2 relationship follow-up advanced | Shared `comments`-backed CRM notes, CRM activities schema, and authenticated `/api/notes`, `/api/activities`, and `/api/timeline` flows implemented for account/contact context; `go test ./...` passed |
| 2026-03-12 | Slice 3 lead pipeline started | CRM `leads`, `opportunities`, and `opportunity_stages` schema added; authenticated `/api/leads`, `/api/leads/{id}/qualify`, `/api/leads/{id}/disqualify`, `/api/leads/{id}/convert`, and `/api/opportunities` flows implemented; notes/activities/timeline now support lead and opportunity context; `go test ./...` passed |
| 2026-03-12 | Communications baseline started | CRM `communications` and `communication_participants` schema added; authenticated `/api/communications` create/list flow implemented; timeline now projects communications; `.env` remains the canonical local config source and `TEST_DATABASE_URL` support was documented for integration-test setup; `go test ./...` passed |
| 2026-03-12 | CRM PostgreSQL integration baseline added | Canonical `.env` plus optional `direnv` wrapper workflow verified for `TEST_DATABASE_URL`; tagged CRM store integration tests added for lead conversion, communications, timeline reads, and tenant isolation; `go test -tags integration -count=1 ./internal/crm -run TestStoreIntegration` passed in a `direnv`-managed shell |
| 2026-03-12 | CRM search baseline added | Authenticated `/api/search?q=...` implemented with tenant-safe search across accounts, contacts, leads, opportunities, activities, and communications; `go test ./...` and `go test -tags integration -count=1 ./internal/crm -run TestStoreIntegration` passed |
| 2026-03-12 | CRM attachments baseline added | Authenticated `/api/attachments` create/list flow implemented on top of kernel attachment tables; relationship timeline and CRM search now project linked attachments; `go test ./...` and `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run TestStoreIntegration"` passed |
| 2026-03-12 | End-to-end review findings recorded | Tenant-boundary enforcement gaps, invalid lead conversion state handling, and tracker/doc completion drift identified; remediation plan added in `docs/review_remediation_2026_03_12.md`; `go test ./...` and `gopls` diagnostics were clean, but no regression coverage exists yet for the review findings |
| 2026-03-12 | CRM review remediation partially advanced | Added `000008_crm_tenant_safe_relationships.up.sql`, store-level org-aware contact validation, strict lead conversion state checks, migration upgrade coverage for a clean legacy schema, and CRM regression tests for cross-tenant references and invalid conversion attempts; `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run TestStoreIntegration"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/platform/migrations -run TestUpgradesApplyTenantSafeCRMRelationshipsFromLegacyState"`, and `gopls` diagnostics on edited non-tagged files passed, but review follow-up is still open for dirty-data upgrade handling and remaining child/link-table tenancy constraints |
| 2026-03-12 | CRM review remediation advanced for dirty upgrades and child tables | `000008_crm_tenant_safe_relationships.up.sql` now fails fast with actionable dirty-data diagnostics before composite tenant FKs are enforced, `000009_crm_child_table_tenant_safety.up.sql` adds tenant-safe parent FKs for `communication_participants` and `attachment_links`, and new regression coverage exercises both paths; `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run TestStoreIntegration"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/platform/migrations"`, workspace `gopls` diagnostics, and `gopls` `go_vulncheck` passed from a tooling perspective; direct `gopls` file diagnostics were unavailable for integration-tagged tests, and `go_vulncheck` reported unrelated Go standard-library SSH advisories rather than CRM-specific findings |
| 2026-03-12 | CRM polymorphic context-link remediation completed for current scope | Added `000010_linked_record_refs.up.sql` to backfill and maintain a shared schema-backed `linked_records` registry for current CRM aggregates, wired `comments`, `activities`, `communications`, and `attachment_links` target references through composite foreign keys, switched store linked-record validation to the registry, and added dirty-upgrade plus raw-SQL regression coverage for polymorphic context links; `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run TestStoreIntegration"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/platform/migrations"`, `gopls` diagnostics on edited Go files, and `gopls` `go_vulncheck` ran successfully; `go_vulncheck` again reported unrelated Go standard-library SSH advisories rather than CRM-specific findings |
| 2026-03-12 | Shared workflow tasks baseline added | Added `000011_workflow_tasks.up.sql`, a new `internal/workflow` module, authenticated `/api/tasks` create/list and `/api/tasks/{id}/complete` flows, tenant-safe task context and assignee enforcement, and CRM timeline projection for linked tasks; `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/workflow"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run TestStoreIntegration"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/platform/migrations"`, `gopls` diagnostics, and `gopls` `go_vulncheck` ran successfully; integration suites currently share one `TEST_DATABASE_URL`, so they were verified serially after an invalid parallel run caused cross-suite interference |
| 2026-03-12 | Canonical planning-set review reaffirmed database-first and AI-agent-first alignment | Reviewed the canonical implementation-plan documents against the reinforced objectives before continuing implementation; current execution, schema, CRM, module-boundary, and AI wording already matched closely enough that no canonical changes were required |
| 2026-03-12 | Project versus work-order distinction clarified in canonical planning docs | Updated the canonical execution, schema, and module-boundary documents to state more explicitly that `project` is the optional engagement-level coordination and aggregation layer, while `work_order` is the concrete operational execution unit for assignment, scheduling, labor, materials, and completion tracking |
| 2026-03-12 | UAE-target and payroll sequencing clarified in canonical planning docs | Updated the canonical strategy, execution, schema, and module-boundary documents so UAE is the real commercial target market, India remains the first delivered compliance and payroll baseline, and payroll is a later SMB-focused capability rather than a full-suite HR/payroll expansion |
| 2026-03-12 | Opportunity-linked estimate baseline added | Added `000012_crm_estimates.up.sql`, authenticated `/api/estimates` create/list and `/api/estimates/{id}/finalize` flows, tenant-safe estimate and estimate-line schema rules, estimate-linked record support in the shared registry, and CRM timeline/search projection for estimates; `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; timeout 120s go test -v -tags integration -count=1 ./internal/crm -run TestStoreIntegration"`, `/bin/bash -lc "set -a; source .env; set +a; timeout 120s go test -tags integration -count=1 ./internal/platform/migrations"`, and `gopls` diagnostics on edited non-tagged Go files passed |
| 2026-03-12 | CRM AI estimate draft recommendation shell added | Added `internal/ai` with authenticated `POST /api/ai/estimate-drafts`, CRM opportunity-context reads, proposal-only estimate draft generation, and persisted `agent_runs`, `agent_run_steps`, `agent_artifacts`, and `agent_recommendations` records without creating live estimates; `go test ./...`, `go test -tags integration -count=1 ./internal/ai`, `timeout 120s go test -v -tags integration -count=1 ./internal/crm -run TestStoreIntegration`, and `gopls` diagnostics on edited Go files passed |
| 2026-03-12 | Comprehensive implementation review recorded new remediation plan | Canonical tracker updated with RP-01 through RP-05 covering audit-atomic writes, bootstrap singleton enforcement, tenant-safe identity linkage, workflow assignee eligibility, and HTTP error classification; `go test ./...` and workspace `gopls` diagnostics were clean, but these findings remain open until code and regression coverage land |
| 2026-03-12 | Canonical audit policy clarified | Implementation-plan docs now state that every business-state change must emit an atomic audit record, AI traceability records do not replace audit events, and low-level technical writes are exempt only by documented category |
| 2026-03-12 | RP-01 audit-atomic write paths completed for current identity, workflow, and CRM scope | Converted current audit-required identity, workflow, and CRM write paths to transactional write-plus-audit persistence, added trigger-backed rollback integration coverage for representative create and state-transition failures, and verified the changed packages serially against the shared `TEST_DATABASE_URL`; `go test ./internal/identityaccess ./internal/workflow ./internal/crm`, `go test -tags integration -count=1 ./internal/identityaccess -run 'TestStoreIntegration(FindLoginCreateSessionAndFindSession|CreateSessionRollsBackWhenAuditInsertFails|RejectsCrossTenantIdentityLinkage)'`, `go test -tags integration -count=1 ./internal/workflow -run 'TestStoreIntegration(CreateListAndCompleteTasks|CreateTaskRollsBackWhenAuditInsertFails|CompleteTaskRollsBackWhenAuditInsertFails|RejectsCrossTenant(TaskContext|Assignee))'`, `go test -tags integration -count=1 ./internal/crm -run 'TestStoreIntegration(EstimateLifecycleAndContext|CreateAccountRollsBackWhenAuditInsertFails|FinalizeEstimateRollsBackWhenAuditInsertFails)'`, `go test ./...`, `gopls` diagnostics on edited non-tagged Go files, and `gopls` `go_vulncheck` passed from a tooling perspective; the integration suites still must run serially because they share one `TEST_DATABASE_URL`, and `go_vulncheck` again reported unrelated Go standard-library SSH advisories rather than slice-specific findings |
| 2026-03-12 | RP-03 tenant-safe identity linkage completed | Added `000013_identity_tenant_safety.up.sql` to fail fast on dirty legacy identity rows, enforce composite tenant-safe membership-role and auth-session-membership-user references, and tightened `identityaccess` login/session joins to rely on those enforced chains; `go test ./internal/identityaccess ./internal/platform/migrations`, `go test -tags integration -count=1 ./internal/identityaccess -run TestStoreIntegration`, `go test -tags integration -count=1 ./internal/platform/migrations -run 'TestUpgrades(ApplyTenantSafeCRMRelationshipsFromLegacyState|RejectDirtyLegacyCrossTenantIdentityRowsWithActionableError)'`, `go test ./...`, and `gopls` diagnostics on edited non-tagged Go files passed; `gopls` file diagnostics could not load metadata for the `//go:build integration` test files in the current workspace setup, and `go_vulncheck` reported unrelated Go standard-library SSH advisories rather than package-specific findings |
| 2026-03-12 | RP-02 bootstrap singleton enforcement completed | Added `000014_bootstrap_singleton.up.sql` to create a schema-backed singleton bootstrap gate with legacy backfill, removed the application-only `HasAnyUser` bootstrap pre-check in favor of transactional singleton claiming, and added integration coverage for both raw schema enforcement and concurrent bootstrap races; `go test ./internal/identityaccess ./internal/platform/migrations`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/identityaccess ./internal/platform/migrations"`, `go test ./...`, workspace `gopls` diagnostics, and `gopls` `go_vulncheck` ran successfully; `go_vulncheck` again reported unrelated Go standard-library SSH advisories rather than bootstrap-specific findings |
| 2026-03-12 | RP-07 mobile API contract work started | Added `/api/v1/...` aliases with `X-Service-Day-API-Version: v1`, introduced validated `limit`, `updated_since`, and cursor pagination on `GET /api/accounts`, `GET /api/contacts`, `GET /api/leads`, `GET /api/opportunities`, and `GET /api/estimates`, and updated CRM store reads to sync and page deterministically by `updated_at`; `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; timeout 120s go test -v -tags integration -count=1 ./internal/crm -run TestStoreIntegration"`, and `gopls` diagnostics on edited Go files passed |
| 2026-03-12 | Product-priority planning updated around `work_order` strength and cross-module integration | Updated the canonical initial plan, execution plan, module boundaries, schema foundation, CRM MVP scope, and planning README so `work_order` support is the strongest long-term product capability and CRM/projects/billing/accounting are expected to integrate tightly around that execution core; planning-only change, review-based validation completed |
| 2026-03-12 | Mobile planning clarified for Flutter client choice | Updated the canonical initial plan, execution plan, and implementation decisions to record Flutter as the intended first mobile client technology while keeping backend mobile-readiness decisions explicitly client-agnostic; planning-only change, review-based validation completed |
| 2026-03-12 | Canonical documentation disconnects reconciled | Locked the accounting kernel baseline and posting boundary in `implementation_decisions_v1.md`, added explicit CRM tracker visibility for the still-pending customer lifecycle baseline and AI recommendation acceptance/discard work, clarified README wording so current AI capabilities are described as proposal-only, and corrected minor CRM scope numbering errors; planning-only change, review-based validation completed |
| 2026-03-12 | Canonical inventory/materials planning refined for equipment-backed delivery | Updated the canonical initial plan, execution plan, module boundaries, schema foundation, implementation decisions, and tracker so service-linked inventory now explicitly covers sourced equipment receipts, project allocation/work-order reservation, serialized-unit tracking where needed, installed customer-site assets, and billable vs non-billable material traceability without expanding v1 into a full procurement-first ERP; planning-only change, review-based validation completed |
| 2026-03-13 | Canonical service-first inventory posture clarified for direct item sales and small-trading-company fit | Updated the canonical initial plan, CRM MVP scope, execution plan, module boundaries, schema foundation, implementation decisions, and tracker so service companies remain the primary target, direct stocked-item sales are explicitly supported, and the same inventory foundation should also be usable by small trading companies at limited depth without redefining v1 as a trading ERP; planning-only change, review-based validation completed |
| 2026-03-13 | Review-driven tracker corrections recorded for next implementation session | Updated the canonical tracker so mobile API groundwork is no longer described as mobile readiness, added RP-08 for the still-missing customer lifecycle baseline, commercial estimate-shape gaps, and the `/api/me` method-contract leak, and tightened the next-session note to prioritize those fixes explicitly; planning-only change, review-based validation completed |
| 2026-03-13 | Canonical double-entry accounting posture clarified | Updated the canonical initial plan, execution plan, schema foundation, implementation decisions, and tracker so `service_day` remains service-operations-first while still requiring a solid double-entry accounting foundation with balanced journals, explicit posting boundaries, and correction-safe ledger behavior; planning-only change, review-based validation completed |
| 2026-03-13 | Canonical accounting submit-versus-post control clarified | Updated the canonical execution plan, schema foundation, AI architecture, module boundaries, implementation decisions, and tracker so accounting entries follow a proposal/submission/posting lifecycle, AI may only propose or policy-allowed submit, and final posting remains a human-controlled action for authorized finance roles; planning-only change, review-based validation completed |
| 2026-03-13 | Canonical standard accounting controls clarified | Updated the canonical initial plan, execution plan, schema foundation, module boundaries, implementation decisions, and tracker so standard accounting capabilities such as document types, durable numbering, accounting periods, credit/debit notes, reversal or void controls, and reversing journals are explicit requirements of the accounting core for India- and UAE-compatible operations; planning-only change, review-based validation completed |
| 2026-03-13 | Canonical GST and TDS service-company baseline clarified | Updated the canonical initial plan, execution plan, schema foundation, module boundaries, implementation decisions, and tracker so GST Lite and TDS are explicitly service-company-suitable baseline finance capabilities with room for deeper future India and UAE tax depth; planning-only change, review-based validation completed |
| 2026-03-13 | Canonical analytics-dimensions posture clarified | Updated the canonical initial plan, execution plan, schema foundation, module boundaries, implementation decisions, and tracker so business-efficiency improvement is treated as a measurable product objective, typed analytic dimensions are preferred over a cost-center-only model, and `cost_center` remains one optional supported dimension type; planning-only change, review-based validation completed |
| 2026-03-12 | Canonical serviced-asset planning refined for repair and maintenance workflows | Updated the canonical initial plan, execution plan, module boundaries, schema foundation, implementation decisions, and tracker so work orders can target explicit serviced assets such as vehicles, machines, and installed devices, with durable service history and linked installed-part traceability rather than relying on customer notes alone; planning-only change, review-based validation completed |
| 2026-03-12 | Canonical serviced-asset v1 shape and construction-trade fit clarified | Updated the canonical schema foundation with a concrete v1 `service_assets` shape plus typed identifiers, readings, and nullable `work_orders.service_asset_id`, and updated the planning docs to state more explicitly that construction subcontractors such as electrical, plumbing, low-voltage/networking, and similar material-backed trades fit the shared project/work_order/inventory execution model; planning-only change, review-based validation completed |
| 2026-03-12 | Canonical future customer WhatsApp support added to roadmap | Updated the canonical initial plan, execution plan, and tracker so WhatsApp is explicitly planned as a later customer-facing channel capability after portal and communication foundations are stable, with channel messages remaining linked to the canonical communication model rather than becoming a parallel truth store; planning-only change, review-based validation completed |
| 2026-03-12 | Verification workflow clarified for runnable app checks | Updated `AGENTS.md`, `docs/go_workflow.md`, and `README.md` so implementation sessions now explicitly require `go build ./cmd/app ./cmd/migrate`, keep `go test ./...` as the default behavioral gate, use a short app smoke run for runtime-facing changes, and treat integration verification as package-specific rather than CRM-only; planning-only change, review-based validation completed |
| 2026-03-12 | SSH vulnerability applicability review completed | Reviewed current `go_vulncheck` SSH findings against the repo and module graph: the application does not implement SSH server/client/agent flows, repo search found no `golang.org/x/crypto/ssh` usage, and the direct `golang.org/x/crypto` dependency is currently used for `bcrypt` from `internal/identityaccess`; treat the reported SSH advisories as not applicable to the current `service_day` codebase and re-evaluate only if SSH functionality is introduced or dependency posture changes materially |
| 2026-03-12 | RP-04 workflow active-membership eligibility completed | Added `000015_workflow_active_memberships.up.sql` to reject task writes whose assignee or creator membership is not active, updated `workflow.StoreDB` to require active assignee, creator, and completion-actor memberships, and added tagged integration coverage for invited/disabled assignment, creation, and completion attempts; `go test ./internal/workflow ./internal/platform/migrations`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/workflow ./internal/platform/migrations"`, `go test ./...`, workspace `gopls` diagnostics, and `gopls` `go_vulncheck` ran successfully; `go_vulncheck` again reported unrelated Go standard-library SSH advisories rather than workflow-specific findings |
| 2026-03-12 | RP-05 HTTP error classification completed | Updated identity, CRM, workflow, and AI handlers to map expected domain failures explicitly and return stable `internal_error` payloads for unexpected backend failures without exposing raw internal text; added HTTP regression tests for representative invalid-input, not-found, transition, membership-conflict, and internal-error cases; `go test ./internal/identityaccess ./internal/ai ./internal/workflow ./internal/crm`, `go test ./...`, workspace `gopls` diagnostics, and `gopls` `go_vulncheck` ran successfully; `go_vulncheck` again reported unrelated Go standard-library SSH advisories rather than API-layer findings |
| 2026-03-12 | CRM AI next-step task recommendation shell added | Added authenticated `POST /api/ai/task-proposals`, opportunity-context task-proposal generation in `internal/ai`, and persisted `crm.next_step_task` runs, task-proposal artifacts, and task-proposal recommendations without creating live workflow tasks; `go test ./internal/ai`, `go test -tags integration -count=1 ./internal/ai`, `go test ./...`, and `gopls` diagnostics on edited non-tagged Go files passed; `gopls` file diagnostics still could not load metadata for the `//go:build integration` AI test file in the current workspace setup |
| 2026-03-12 | Provider-backed AI estimate drafting added | Added optional OpenAI Responses API estimate-draft generation with strict JSON schema output, config-driven model/API-key wiring, provider/model persistence in `agent_runs`, and heuristic fallback when OpenAI is not configured; `go test ./internal/ai ./internal/platform/config ./internal/platform/httpserver`, `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/ai"`, and `gopls` diagnostics on edited non-tagged Go files passed; `gopls` file diagnostics still could not load metadata for the `//go:build integration` AI test file in the current workspace setup, and `gopls` `go_vulncheck` could not run because the installed scanner was built against an older Go toolchain than the current `go1.26.1` environment |
| 2026-03-12 | Local Go developer tools rebuilt for Go 1.26.1 | Rebuilt `gopls` and `staticcheck` against the current Go `1.26.1` toolchain and installed side-by-side as `/home/vinod/go/bin/gopls-go1.26.1` and `/home/vinod/go/bin/staticcheck-go1.26.1`; the default binaries were then replaced with those rebuilt versions while keeping backups as `/home/vinod/go/bin/gopls.pre-go1.26.1` and `/home/vinod/go/bin/staticcheck.pre-go1.26.1`, but any already-running MCP/tool-host process still needs a restart before in-session `go_vulncheck` uses the newer scanner, and live vuln scanning still requires outbound access to `vuln.go.dev` |
| 2026-03-12 | OpenAI estimate-draft local default switched to GPT-4o | Updated the optional provider-backed estimate-drafting path so local development and testing now default `OPENAI_MODEL` to `gpt-4o` instead of `gpt-5-mini`, while keeping the model runtime-configurable for later changes; aligned tests plus local configuration docs with that choice, and `go test ./internal/ai ./internal/platform/config ./internal/platform/httpserver` plus `gopls` diagnostics on edited non-tagged Go files passed |
| 2026-03-12 | RP-06 API validation and identifier-boundary hardening completed | Added shared UUID-shape validation for CRM/workflow/AI service-boundary identifiers plus CRM/workflow handler route/query guards, normalized CRM validation failures onto `ErrInvalidInput` while preserving field-specific messages, and updated auth middleware to accept case-insensitive bearer schemes; `go test ./internal/crm ./internal/workflow ./internal/ai ./internal/identityaccess`, `go test ./...`, and `gopls` diagnostics on edited Go files passed; MCP-hosted `gopls` `go_vulncheck` still reflected the pre-restart tool host, while direct CLI vuln scanning via the rebuilt default `gopls` got past the Go-version mismatch and then stopped on blocked outbound access to `vuln.go.dev` |
| 2026-03-13 | CRM customer lifecycle baseline completed | Added `000016_crm_account_lifecycle.up.sql`, schema-backed tenant lifecycle stages and transition history for accounts, authenticated `/api/account-lifecycle-stages`, `/api/accounts/{id}/lifecycle`, and `/api/accounts/{id}/lifecycle-history` flows, lifecycle timeline projection, and automatic lifecycle advancement on lead conversion and opportunity creation; `go build ./cmd/app ./cmd/migrate`, `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run TestStoreIntegration"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/platform/migrations"`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"` passed; `gopls` file diagnostics for the `//go:build integration` CRM test remain unavailable in the default workspace, so the tagged suite remained the authoritative check for that file |
| 2026-03-13 | AI task-proposal acceptance and discard flow added | Added authenticated `POST /api/ai/task-proposals/{id}/accept` and `/api/ai/task-proposals/{id}/discard`, transactional acceptance into live shared workflow tasks through the workflow task write path, accepted-artifact persistence, and AI recommendation accept/discard audit events with causation metadata on both the AI and workflow sides; verified with `go build ./cmd/app ./cmd/migrate`, `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/ai"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/workflow"`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"` |
| 2026-03-14 | Review-remediation priority slice implemented | Added opportunity update support, many-to-many contact account-link surfacing plus replacement, estimate commercial fields and revision lineage with `000017_crm_estimate_commercials.up.sql`, AI acceptance approval persistence and tool-policy seeding, an initial `internal/accounting` posting shell with balanced-journal validation, and then closed the follow-up review findings with `000018_idempotency_scope_and_recovery.up.sql`, scoped idempotency-key handling on the safe CRM and AI transition endpoints, replay-safe no-op opportunity updates, replay-safe AI task-proposal acceptance, and explicit invalid-stage classification for CRM lookup-backed transitions; verified with `go test ./...`, `go build ./cmd/app ./cmd/migrate`, workspace `gopls` diagnostics on edited non-tagged Go files, `gopls` `go_vulncheck ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/platform/migrations"`, `/bin/bash -lc "set -a; source .env; set +a; timeout 120s go test -v -tags integration -count=1 ./internal/ai -run \"TestStoreIntegration(AcceptTaskProposalCreatesWorkflowTaskAndMarksRecommendationAccepted|AcceptTaskProposalReplaysAcceptedDecision|DiscardTaskProposalMarksRecommendationDiscarded)\""`, `/bin/bash -lc "set -a; source .env; set +a; timeout 120s go test -v -tags integration -count=1 ./internal/crm -run \"TestStoreIntegration(AccountLifecycleFlow|TransitionAccountLifecycleRejectsUnknownStage|UpdateOpportunityNoOpDoesNotDuplicateAudit|UpdateOpportunityRejectsUnknownStage)\""`, `/bin/bash -lc "set -a; source .env; set +a; timeout 120s go test -v -tags integration -count=1 ./internal/platform/idempotency"`, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; `gopls` diagnostics still do not load the `//go:build integration` files in the default workspace, so the tagged integration suites remained the authoritative checks for those edited files |
| 2026-03-15 | AI estimate-draft acceptance and discard flow added | Added authenticated `POST /api/ai/estimate-drafts/{id}/accept` and `/api/ai/estimate-drafts/{id}/discard`, transactional acceptance into live CRM estimates through CRM-owned create/revise write helpers, accepted-estimate artifact persistence, approval records and tool-policy seeding for estimate actions, audit metadata carrying AI causation into resulting CRM estimate writes, and scoped idempotent retry replay on estimate acceptance; verified with `gofmt -w internal/ai/service.go internal/ai/http.go internal/ai/store.go internal/ai/service_test.go internal/ai/http_test.go internal/ai/store_integration_test.go internal/crm/store.go`, `go test ./internal/ai ./internal/crm`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/ai"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run 'TestStoreIntegration(EstimateLifecycleAndContext|FinalizeEstimateRollsBackWhenAuditInsertFails)'"`, workspace `gopls` diagnostics on edited non-tagged Go files, `gopls` `go_vulncheck ./...`, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; the broader `./internal/crm -run TestStoreIntegration` suite still has an existing unrelated failing `TestStoreIntegrationRejectsCrossTenantContactLinking` in the current dirty worktree, so the targeted estimate-related CRM integration tests were the authoritative verification for this slice |
| 2026-03-15 | Idempotent create-path coverage expanded for tasks and estimates | Extended the scoped `Idempotency-Key` contract to authenticated `POST /api/tasks` and `POST /api/estimates`, refactored CRM, workflow, and AI handlers to depend on a small idempotency executor interface so HTTP replay behavior can be tested cleanly, and added regression coverage for replayed responses plus request-hash mismatch conflicts on the newly covered create endpoints; verified with `gofmt -w internal/crm/http.go internal/crm/http_test.go internal/workflow/http.go internal/workflow/http_test.go internal/ai/http.go`, `go test ./internal/crm ./internal/workflow`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"` |
| 2026-03-15 | `/api/me` method contract hardened | Updated `internal/identityaccess` so `/api/me` now rejects non-`GET` requests with `405 {"error":"method_not_allowed"}` before auth runs, added HTTP regression coverage for the method guard, and clarified `README.md` so the documented identity endpoint contract matches runtime behavior; verified with `gofmt -w internal/identityaccess/http.go internal/identityaccess/http_test.go`, `go test ./internal/identityaccess`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"` |
| 2026-03-15 | Device-scoped session and refresh-token baseline added | Added `000019_identity_device_sessions.up.sql` with tenant-safe `device_sessions`, `refresh_tokens`, and device-linked `auth_sessions`, extended `internal/identityaccess` with authenticated `POST /api/device-sessions/login` and `POST /api/device-sessions/refresh`, preserved the existing `POST /api/login` flow for current clients, and added rotation-focused unit plus PostgreSQL integration coverage for the new refresh-token path; verified with `gofmt -w internal/identityaccess/service.go internal/identityaccess/store.go internal/identityaccess/http.go internal/identityaccess/service_test.go internal/identityaccess/http_test.go internal/identityaccess/store_integration_test.go`, `go test ./internal/identityaccess`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/identityaccess"`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/platform/migrations"`, workspace `gopls` diagnostics on edited non-tagged Go files, `gopls` `go_vulncheck ./internal/identityaccess`, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; the smoke run exited via the expected `timeout` status after successful startup, and current `go_vulncheck` findings remain the existing non-applicable Go SSH advisories rather than auth-slice-specific issues |
| 2026-03-15 | Attachment upload/download transport baseline added | Added a small `internal/attachments` local-storage package, wired authenticated `POST /api/attachments/upload` multipart upload and `GET /api/attachments/{id}/download` into the CRM slice, kept the existing metadata-only attachment create/list contract for compatibility, generated storage keys server-side, returned opaque download URLs, and added attachment storage unit tests plus CRM HTTP/store regression coverage for upload/download and attachment lookup; verified with `gofmt -w internal/attachments/storage.go internal/attachments/spool.go internal/attachments/storage_test.go internal/crm/http.go internal/crm/http_test.go internal/crm/service.go internal/crm/service_test.go internal/crm/store.go internal/crm/store_integration_test.go internal/platform/config/config.go internal/platform/httpserver/server.go`, `go test ./internal/attachments ./internal/crm ./internal/platform/httpserver`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run 'TestStoreIntegration(LeadConversionAndTimeline|RejectsCrossTenantAttachmentLinking|RejectsCrossTenantAttachmentTargetLinking)'"`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; the broad `./internal/crm -run TestStoreIntegration` suite still has the existing unrelated `TestStoreIntegrationRejectsCrossTenantContactLinking` failure in the current worktree, so the targeted attachment-related CRM integration tests were the authoritative database checks for this slice |
| 2026-03-15 | Cross-tenant contact-link error-contract drift recorded for follow-up | Added `implementation_plan/crm_contact_link_error_contract_remediation_2026_03_15.md` to keep the failing `TestStoreIntegrationRejectsCrossTenantContactLinking` visible; the current evidence suggests tenant safety still holds but the store or service error contract has drifted from an application-mapped `account not found` outcome to a raw schema/FK failure, and the next implementation session should resolve that contract explicitly before relying on the broad CRM integration suite as fully healthy again |
| 2026-03-16 | CRM contact-link error contract restored | Updated `internal/crm/store.go` so `CreateContact` derives one canonical account-link set from either `AccountLinks` or direct `PrimaryAccountID`, pre-validates those links through the existing linked-record guard before any `account_contacts` insert, and therefore preserves the intended application-level `account not found` behavior for cross-tenant contact-link attempts even for direct store callers; verified with `gofmt -w internal/crm/store.go`, `go test ./internal/crm`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; go test -tags integration -count=1 ./internal/crm -run TestStoreIntegrationRejectsCrossTenantContactLinking"` |
| 2026-03-16 | Canonical workflow-efficiency posture strengthened | Updated the canonical initial plan, execution plan, implementation decisions, schema foundation, module boundaries, CRM MVP scope, and tracker so task/activity management is treated as a first-class efficiency capability, tasks remain distinct from activities, shared workflow supports one primary person-or-team owner plus one primary business context, and secondary related links are allowed for project/work-order/site/asset visibility and analytics; planning-only change, review-based validation completed |
| 2026-03-16 | Canonical workflow terminology clarified | Updated `implementation_decisions_v1.md`, `service_day_schema_foundation_v1.md`, and the tracker so `workflow`, `task`, and `activity` are now defined explicitly and consistently in the canonical planning set; planning-only change, review-based validation completed |
| 2026-03-16 | Device-session management and notification registration baseline added | Added `000020_notifications_device_registrations.up.sql` with tenant-safe `notification_devices` plus revocation-sync trigger behavior, extended `internal/identityaccess` with authenticated `GET /api/device-sessions` and `POST /api/device-sessions/{id}/revoke`, added a new `internal/notifications` module with authenticated `GET /api/notification-devices` and `POST /api/notification-devices`, and updated CRM/workflow/AI idempotency conflicts to emit machine-readable retry hints through headers and JSON payload fields; verified with `gofmt -w internal/notifications/service.go internal/notifications/store.go internal/notifications/http.go internal/notifications/service_test.go internal/notifications/http_test.go internal/notifications/store_integration_test.go internal/identityaccess/service.go internal/identityaccess/http.go internal/identityaccess/store.go internal/identityaccess/service_test.go internal/identityaccess/http_test.go internal/identityaccess/store_integration_test.go internal/platform/httpserver/server.go internal/platform/migrations/migrations_test.go internal/workflow/http.go internal/workflow/http_test.go internal/crm/http.go internal/crm/http_test.go internal/ai/http.go`, `go test ./internal/notifications ./internal/identityaccess ./internal/platform/httpserver ./internal/workflow ./internal/crm ./internal/ai ./internal/platform/migrations`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, `go test -tags integration -count=1 ./internal/notifications`, `go test -tags integration -count=1 ./internal/identityaccess`, `go test -tags integration -count=1 ./internal/platform/migrations`, workspace `gopls` diagnostics on edited non-tagged Go files, `gopls` `go_vulncheck ./internal/identityaccess`, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; the smoke run exited via the expected `timeout` status after successful startup, and current `go_vulncheck` findings remain the existing non-applicable Go SSH advisories rather than app-specific issues |
| 2026-03-16 | Notification inbox and task-assignment fan-out baseline added | Added `000021_notifications_events_and_deliveries.up.sql` with tenant-safe `notification_events` and `notification_deliveries`, extended `internal/notifications` with authenticated `GET /api/notifications` and `POST /api/notifications/{id}/read`, and wired `internal/workflow` task creation so cross-user assignment now persists a notification event plus inbox and queued push-delivery bookkeeping atomically with the task write; verified with `gofmt -w internal/notifications/service.go internal/notifications/store.go internal/notifications/http.go internal/notifications/service_test.go internal/notifications/http_test.go internal/notifications/store_integration_test.go internal/workflow/store.go internal/workflow/store_integration_test.go internal/platform/migrations/migrations_test.go`, `go test ./internal/notifications ./internal/workflow ./internal/platform/migrations`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, `go test -tags integration -count=1 ./internal/notifications`, `go test -tags integration -count=1 ./internal/workflow`, `go test -tags integration -count=1 ./internal/platform/migrations`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; the smoke run exited via the expected `timeout` status after successful startup |
| 2026-03-16 | Push-dispatch worker baseline added for queued notifications | Added `internal/notifications/dispatch.go` plus app-start wiring so `service_day` can now optionally poll queued `notification_deliveries`, route them by `push_provider`, and send provider-specific HTTP JSON requests to configured FCM or APNS endpoints, with delivery rows marked `delivered` on success or `failed` with provider error codes on failure; verified with `gofmt -w cmd/app/main.go internal/notifications/dispatch.go internal/notifications/dispatch_test.go internal/notifications/store.go internal/notifications/store_integration_test.go internal/platform/config/config.go internal/platform/config/config_test.go`, `go test ./internal/notifications ./internal/platform/config ./cmd/app`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, `go test -tags integration -count=1 ./internal/notifications`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; the smoke run exited via the expected `timeout` status after successful startup |
| 2026-03-16 | API compatibility policy made explicit in the HTTP contract | Updated `internal/platform/httpserver` so API responses now stamp both `X-Service-Day-API-Version: v1` and `X-Service-Day-API-Compatibility: additive-within-v1`, and updated the canonical implementation decisions, tracker narrative, and root `README.md` to record the declared pre-production v1 policy: additive-only backward-compatible changes within `v1`, new major-version paths for breaking changes, standard `Deprecation` plus `Sunset` headers before future removals, and an explicit online-first mobile contract; verified with `gofmt -w internal/platform/httpserver/server.go internal/platform/httpserver/server_test.go`, `go test ./internal/platform/httpserver`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; the smoke run exited via the expected `timeout` status after successful startup |
| 2026-03-16 | CRM create endpoints gained broader retry-safe idempotency coverage | Updated `internal/crm/http.go` so scoped `Idempotency-Key` replay now covers authenticated JSON create flows for accounts, contacts, leads, opportunities, estimates, communications, notes, and activities in addition to the earlier estimate create and opportunity update paths, using operation-scoped replay keys and the existing machine-readable retry guidance contract; verified with `gofmt -w internal/crm/http.go internal/crm/http_test.go`, `go test ./internal/crm`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; the smoke run exited via the expected `timeout` status after successful startup |
| 2026-03-16 | CRM retry-safe idempotency now covers lead and estimate transitions | Updated `internal/crm/http.go` so scoped `Idempotency-Key` replay now also covers lead qualify/disqualify/convert plus estimate revise/finalize, closing the remaining duplicate-prone CRM state transitions on the current mobile-consumable surface; verified with `gofmt -w internal/crm/http.go internal/crm/http_test.go`, `go test ./internal/crm`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app"`; the smoke run exited via the expected `timeout` status after successful startup |
| 2026-03-16 | Documentation handoff aligned to current remediation priority | Updated `implementation_plan/README.md`, `implementation_plan/implementation_tracker.md`, `README.md`, and `AGENTS.md` so remediation-note usage is explicit, the repository handoff points back to the tracker as the canonical execution record, the example CRM flow no longer treats activities as the accountable follow-up record, and the next-session priority now clearly starts with RP-09 workflow task-model alignment before broader mobile/task expansion; planning-only change, review-based validation completed |
| 2026-03-16 | Estimate line-shape remediation queued | Added `implementation_plan/estimate_line_shape_alignment_remediation_2026_03_16.md` and indexed RP-10 in the tracker so the current estimate-line contract drift is explicitly queued for a future implementation session: the canonical plan still expects estimate lines to remain usable for scoped work and stocked items where relevant, while the shipped code currently only accepts `service` and `milestone`; planning-only change, review-based validation completed |
| 2026-03-18 | Canonical later inventory-trading extension added | Updated the canonical initial plan, execution plan, module boundaries, schema foundation, implementation decisions, and tracker so a later limited inventory-trading workflow now has explicit roadmap, ownership, schema, and milestone language distinct from the narrower marketplace-seller extension while still reusing the shared item, inventory, billing, tax, and accounting foundations; planning-only change, review-based validation completed |
| 2026-03-19 | Thin-v1 document kernel implementation started | Added `internal/documents` with canonical document identity plus lifecycle validation, added `000022_documents_kernel.up.sql` to create shared `documents`, backfill existing accounting journals into that kernel, and register document linked-record refs, and updated `internal/accounting` so draft journal creation and posting now create and transition shared document records transactionally instead of relying only on journal-local lifecycle state; verified with `gofmt -w internal/documents/service.go internal/documents/store.go internal/documents/service_test.go internal/accounting/service.go internal/accounting/store.go internal/platform/migrations/migrations_test.go internal/platform/migrations/migrations_integration_test.go`, `go test ./internal/documents ./internal/accounting ./internal/platform/migrations`, `go build ./cmd/app ./cmd/migrate`, `go test ./...`, workspace `gopls` diagnostics on edited non-tagged Go files, and `/bin/bash -lc "set -a; source .env; set +a; timeout 120s go test -tags integration -count=1 ./internal/platform/migrations"` |
| 2026-03-16 | User-workflow documentation plan synced to notification inbox baseline | Updated `implementation_plan/service_day_user_workflow_docs_plan_v1.md` so the documentation plan now reflects the shipped notification inbox/read baseline more accurately: auth guidance includes inbox/read alongside notification-device registration, task/follow-up guidance calls out the current cross-user assignment notification behavior, and the remaining notification gap is narrowed to push-provider dispatch plus broader event coverage rather than treating all notification fan-out as still missing; planning-only change, review-based validation completed |
| 2026-03-16 | User-workflow documentation plan refreshed for the current shipped surface | Updated `implementation_plan/service_day_user_workflow_docs_plan_v1.md` so the documentation wave now reflects the current implementation more accurately: auth guidance explicitly includes device-session list/revoke plus notification-device registration, the current online-first/mobile-contract limitations are called out, and `estimates.md` is now part of the next expected real guide set alongside the existing CRM/auth guides; planning-only change, review-based validation completed |
| 2026-03-13 | Canonical narrow marketplace-seller extension added | Updated the canonical initial plan, execution plan, module boundaries, schema foundation, implementation decisions, and tracker so a later small Amazon-style seller workflow can reuse the shared item, inventory, billing, tax, and accounting foundations at limited depth without shifting the product away from its service-first core; planning-only change, review-based validation completed |
| 2026-03-13 | User-guides documentation baseline added | Added `docs/user_guides/README.md` and updated the canonical initial plan, execution plan, tracker, `AGENTS.md`, `docs/go_workflow.md`, and root `README.md` so end-user guides are treated as part of implementation quality for shipped user-visible workflows; planning-only change, review-based validation completed |
| 2026-03-13 | Canonical ownership and schema inconsistencies patched after plan review | Updated the canonical module boundaries, schema foundation, and tracker so `identity_access` now owns later portal-facing principal and membership records, `account_lifecycle_history` is explicit in the CRM schema canon, `payroll` is present in the top-level module inventory, and typed analytic-dimension configuration now has explicit shared ownership and schema shape; planning-only change, review-based validation completed |

## Next-session note

Use the following order unless a higher-priority bug interrupts it.

### 1. Recommended next implementation slice

Primary recommendation:
1. land RP-09 workflow task-model alignment first, using [workflow_task_model_alignment_remediation_2026_03_16.md](/home/vinod/PROJECTS/service_day/implementation_plan/workflow_task_model_alignment_remediation_2026_03_16.md) as the implementation guide for shared `teams`, explicit person-or-team task ownership, secondary task links, and activity/task cleanup
2. continue mobile/backend readiness immediately after that alignment work, focusing on the next notification-event wave, stronger push retry or backoff behavior, and any remaining retry-sensitive endpoint coverage that should build on the stronger task model rather than the older user-only contract
3. start the first real user workflow guides under `docs/user_guides/` once the auth, CRM relationship, and current task flows are stable enough to document accurately
4. implement RP-10 estimate line-shape alignment once RP-09 is landed or a related commercial hardening session is underway, so the estimate baseline no longer hard-codes only `service` and `milestone`
5. add lead-context AI suggestions only if they fit the same persisted run, recommendation, approval, and audit model cleanly after the workflow-model correction

Why this is next:
1. RP-09 is now the clearest documented implementation drift between the updated canonical planning set and the shipped code, and the remediation note explicitly says it should land before broader team-based mobile task flows, queue analytics, or deeper work-order execution surfaces depend on the older user-only task contract
2. the contact-link contract drift is closed, and the latest mobile-readiness wave already landed session-management, notification registration, notification inbox/read, push-dispatch, retry-guidance primitives, and explicit compatibility-policy wording, so the next mobile hardening work should build on the corrected workflow ownership model instead of extending the older one further
3. the estimate-line contract drift is narrower in code than in the canonical commercial plan, so it now has an explicit queued remediation and should be corrected before later commercial breadth is assumed from the current implementation
4. user guides should follow once the shipped auth, CRM, and current workflow behavior are stable enough that the docs will not churn immediately
5. lead-context AI suggestions remain valuable, but they should follow the same persisted approval and audit model and should not overtake the more structural workflow-model correction

### 2. Concrete implementation expectations

For the next mobile-facing or retry-sensitive endpoint wave:
1. scope idempotency keys per operation
2. use lease-based reclaim semantics consistently
3. avoid claiming generic create-path idempotency unless the endpoint truly supports replay-safe duplicate suppression
4. keep runtime method behavior, docs, and tests aligned so public-contract drift stays visible early

### 3. Secondary follow-on items

After the current mobile-auth and attachment baseline:
1. continue mobile/backend readiness with broader retry-safe endpoint coverage and the next notification step, likely broader event coverage beyond task assignment plus stronger push retry/backoff handling, after RP-09 is in place
2. implement RP-10 estimate line-shape alignment when the session is touching CRM commercial behavior or AI estimate acceptance
3. add the first real user workflow guides under `docs/user_guides/` once the auth, CRM relationship, and current workflow flows are stable enough to document accurately
4. add lead-context AI suggestions only if they fit the same persisted run, recommendation, approval, and audit model cleanly

### 4. Verification baseline for the next session

If the next session changes Go code in the current CRM or AI slice, run at minimum:
1. `gofmt -w` on changed Go files
2. `go build ./cmd/app ./cmd/migrate`
3. `go test ./...`
4. `gopls` diagnostics on edited non-integration Go files
5. targeted integration suites for the touched store or migration paths
6. `APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app` when HTTP wiring or runtime-visible behavior changes

### 5. Ongoing guardrails

1. treat the current SSH-related `go_vulncheck` findings as non-applicable unless SSH functionality is introduced later
2. keep the first planned Flutter client separate from backend mobile-readiness concerns; Flutter is the client choice, not the backend contract
3. preserve payroll-ready seams as delivery and accounting foundations advance
4. do not let later `inventory_ops`, `service_assets`, rental, WhatsApp, or marketplace-extension planning blur the factual status of the current CRM-plus-AI baseline

## Blockers

No hard blockers are currently recorded.

Visible follow-up issue:
1. no additional blocker is currently recorded for the broad CRM integration baseline; resume from RP-09 workflow task-model alignment first, then continue the mobile/backend readiness and user-guide follow-on work unless a higher-priority bug interrupts
