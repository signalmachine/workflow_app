# workflow_app Implementation Objectives

Date: 2026-03-22
Status: High-level multi-version implementation summary
Purpose: capture the high-level rules, principles, objectives, requirements, specifications, and invariants that implementation work should preserve across the application across v1, v2, and later versions.

## 1. Canonical source rule

This document is a high-level multi-version summary, not a replacement for the active planning set.

Rules:

1. `new_app_docs/` is the active canonical planning source for current v1 work.
2. legacy planning material is reference-only unless an active `new_app_docs/` document promotes a rule forward.
3. When active and legacy planning differ, implementation should follow `new_app_docs/`.
4. this document may describe long-term implementation objectives that extend beyond thin v1
5. thin v1 remains the active implementation priority and should be treated as the minimum serious foundation version, not the full product endpoint
6. this document is a companion summary and is not mandatory reading for every implementation session
7. `docs/implementation_objectives/implementation_principles.md` is a reference-only companion note that provides implementation-principles guidance, but it is not the sole source of implementation principles and it is not part of the canonical planning set
8. the normal session-start source of truth remains `AGENTS.md`, `README.md`, `new_app_docs/`, and optional reference material when needed
9. when high-level objectives, rules, principles, specifications, or invariants change in those canonical sources, this summary should be reviewed and updated if needed
10. when implementation-time codebase review surfaces drift, an issue, an inconsistency, or a conflict, contributors should report it and either fix it in the same change when appropriate or document it in the canonical implementation plan docs for a future session

## 2. Product identity

`workflow_app` is intended to be:

1. an AI-agent-first business operating system
2. a database-first and SQL-first application
3. a documents-plus-ledgers-plus-execution-context system
4. a review-and-report product for humans rather than a form-heavy manual-entry product
5. a service-business-first foundation that also remains usable for light trading at limited depth

It is explicitly not intended to become:

1. a CRM-first product
2. a portal-first product
3. a broad manual UI-first ERP
4. a workflow suite without strong accounting, inventory, and audit truth

## 3. Core doctrine

The application is shaped by three foundational layers:

1. documents as intent
2. ledgers as truth
3. execution context as operational reality

The primary interaction model should be:

1. persisted inbound request as intake truth
2. AI processing as asynchronous proposal generation
3. human review and controlled action through explicit queues and review surfaces

Selective inspiration may be taken from systems such as OpenClaw where it strengthens the business-control architecture:

1. durable persisted intake rather than transient prompt handling
2. queue-oriented asynchronous processing
3. modular tool or skill packaging with explicit capability boundaries
4. browser-first control surfaces that can later support mobile clients on the same backend model

Those patterns should be adapted, not copied blindly:

1. `workflow_app` is stricter on approvals, posting boundaries, auditability, and database truth
2. consumer-assistant style always-on autonomy is not the target operating model
3. broad self-directed agent behavior remains out of scope unless a foundation need explicitly justifies it

Implementation consequences:

1. every important capability should map to a document, a ledger effect, an execution context, an approval path, and a review/report surface
2. documents explain what the business decided
3. ledgers explain what changed in measurable terms
4. execution records explain what actually happened operationally
5. reports are derived views, not truth owners
6. inbound requests should be durable records rather than transient prompts
7. inbound requests are not business documents and should not consume the `documents` numbering or lifecycle model merely because they may later cause document creation
8. the same intake model should be reusable for human and non-human upstream systems
9. every meaningful workflow and control state should be durably reconstructible from database records rather than transient process state
10. parked inbound requests should support draft and queued states explicitly rather than depending on implicit message completeness
11. queued or otherwise submitted-but-unprocessed request removal should normally be soft cancel rather than hard deletion so auditability and recovery remain intact
12. draft requests may still be hard-deleted completely because they have not yet entered the AI processing queue
13. queued or cancelled pre-processing requests may return to `draft` for amendment and later resubmission while preserving the same intake identity

## 4. Versioning stance and thin-v1 objective

The implementation objectives in this document span multiple versions.

Rules:

1. not every objective in this document is intended to land in v1
2. thin v1 is the current foundation release target
3. v2 and later versions may deepen localization, workflow breadth, vertical extensions, and operator surfaces on top of the v1 foundation
4. implementation should preserve extension room for later versions without expanding thin-v1 scope prematurely

The active thin v1 aims to deliver the minimum serious system that can:

1. accept human requests through AI
2. create reviewable business documents
3. post approved documents into financial and inventory truth layers
4. support foundational GST and TDS handling
5. track operations through work orders and tasks
6. expose approval, review, inspection, and reporting surfaces for humans
7. persist inbound requests and process them through a review-oriented queue model rather than relying on immediate AI response as the default operating path

Near-term success is defined more by safe and observable AI-assisted operation on strong foundations than by broad product breadth.

## 5. Highest-priority v1 capabilities

The highest thin-v1 priorities are:

1. identity, org, roles, sessions, and tenant safety
2. audit events, approvals, and idempotent write boundaries
3. persisted inbound request intake plus AI coordinator, specialist-agent routing, tool policy, and run observability
4. party, contact-support, item, location, ledger-account, and tax-foundation records
5. workforce foundation for assignment, labor capture, and labor costing
6. accounting and posting foundations
7. inventory movement foundations
8. adopted document-family payload ownership for work-order, invoice, payment or receipt, and inventory flows
9. work-order and task execution foundations
10. report and review surfaces
11. provider-backed AI execution using the OpenAI Go SDK plus a shared backend processing contract, explicit live-verification command, and the minimum backend API and attachment-transport contract required to exercise that path in real testing

CRM and project depth may remain in the repository where already implemented, but they are support concerns in thin v1 rather than the product center.

## 6. AI-agent-first requirements

AI is the primary operator interface, but not the authority over truth.

Required rules:

1. AI acts through explicit tools over normal domain services.
2. AI may read, summarize, draft, recommend, and request approval.
3. AI may execute only bounded writes allowed by policy.
4. financially meaningful writes remain policy-gated, with human gating as the default control posture in thin v1
5. AI may never write ledger rows directly
6. AI may never bypass posting, approval, audit, or schema constraints
7. AI execution must be observable through durable run history, steps, artifacts, recommendations, approvals, and delegation traces
8. the preferred architecture is multi-agent: one coordinator routes bounded work to specialist agents
9. the preferred interaction model may persist inbound user requests first and process them asynchronously rather than relying on immediate AI response as the default path
10. queue-oriented processing is preferred because it preserves durability, supports clearer human review, and extends cleanly to requests originating from external systems as well as humans
11. modular tool or skill boundaries are preferred where they keep agent capabilities explicit, reviewable, and policy-gated rather than hidden inside prompt-only behavior
12. AI workers should not process requests that remain in draft or have been cancelled before pickup
13. thin v1 may stay narrow in workflow breadth, but the AI execution layer should still be foundation-complete enough to run real provider-backed agent flows safely
14. the preferred first live provider path for v1 is OpenAI through the Go SDK and Responses API
15. `workflow_app` should use modern workflow AI agent architectures that are suitable for strictly controlled business workflows rather than autonomy-heavy general-agent patterns
16. tool calling should be the primary AI execution pattern, and AI tool handlers should route into the existing domain services in the codebase rather than duplicating business logic inside the AI layer
17. when implementing or verifying `internal/ai` against the OpenAI Go SDK, contributors should prefer official OpenAI documentation and the official `openai/openai-go` repository for exact SDK and API details rather than relying on memory alone

The short-term AI objective is to observe, evaluate, and improve agent behavior on real bounded business tasks.

## 7. Human-interface stance

Human surfaces in thin v1 should now include a usable web application layer while still preserving the AI-agent-first operating model.

Allowed primary human surfaces:

1. approval queues
2. review screens
3. inspection and query surfaces
4. reports
5. a usable web application layer over the shared backend contracts
6. a thin-v1 browser layer that stays server-rendered by default and may use lightweight progressive-enhancement libraries without adopting a separate frontend build stack

Not intended as core thin-v1 behavior:

1. broad manual operational data entry
2. direct human ledger editing
3. broad human operational UI replacing agent-driven workflows
4. a separate backend for web versus mobile clients
5. thin v1 should still preserve one shared backend foundation that later mobile and web clients can both use, even where their capture and presentation layers differ

Preferred thin-v1 web-implementation stance:

1. keep Go-native server-rendered HTML as the baseline browser delivery model
2. prefer progressive enhancement such as `htmx` where it materially improves operator continuity while preserving server ownership of rendering and workflow state
3. use small client-state helpers such as `Alpine.js` only where local interaction needs justify them
4. avoid introducing a separate SPA architecture, Node dependency chain, or frontend-specific build pipeline in thin v1 unless the canonical planning set explicitly changes that decision
5. when the promoted web layer proves a concrete need, backend corrections and narrow shared-backend enhancements should still be made, but they should stay in service of the same shared engine and must not become a pretext for unrelated backend feature expansion or a second web-specific backend
6. the same principle applies when work is centered on other non-backend layers such as the AI-agent layer: backend bugs, missing support seams, and narrow capability gaps should still be corrected when they materially block the active slice, but they should remain tied to that slice rather than expanding backend scope opportunistically
7. during Milestone 7 execution, prefer larger coherent browser-workflow slices over many tiny continuity patches, while still keeping each slice bounded to one related operator path rather than mixing unrelated areas into one delivery

## 8. Data and database principles

The database is the main safety system.

Implementation rules:

1. enforce invariants in PostgreSQL wherever practical
2. treat Go services as the second enforcement layer
3. prefer constraints, foreign keys, unique rules, and transactional boundaries over application-only correctness
4. tenant-relevant tables should carry `org_id`
5. tenant-crossing references must be blocked by schema design, not only by handler checks
6. one-time bootstrap must be database-safe
7. invalid states, invalid transitions, and unsafe postings should be rejected by default
8. sophisticated PostgreSQL-native modeling is preferred when it materially improves correctness, auditability, performance, or operability, provided it does not create unnecessary implementation or operational pain
9. meaningful workflow and control-state transitions should persist durably enough that intake, processing, review, approval, posting, execution, and recoverable failure history can be reconstructed from database records
10. during thin-v1 development and testing, storing inbound-request attachments in PostgreSQL is acceptable if the design preserves a later move to external object storage
11. original uploaded artifacts such as voice recordings should remain durably available even when derived records such as transcriptions are created

## 9. Module and ownership boundaries

High-level ownership rules must remain explicit even though the product should feel integrated.

Rules:

1. each module owns its tables, write paths, and invariants
2. other modules should not write directly into another module's tables
3. cross-module reads should prefer explicit services, exported queries, or read models
4. shared infrastructure does not create shared business ownership
5. product flows should feel integrated without weakening ownership boundaries
6. shared document identity, lifecycle, numbering, and posting-linkage rules need one canonical ownership path
7. shared approval records and approval queues need one canonical ownership path rather than AI-only special handling

## 10. Core truth invariants

The following are core application invariants:

1. ledger truth is append-only
2. financial postings are balanced double-entry movements
3. inventory truth is append-only movement history from explicit source to destination
4. balances, stock, outstanding values, and reports are derived rather than stored as mutable truth
5. documents do not directly store financial or stock truth
6. posting is explicit, deterministic, idempotent, and transactional
7. a business action that requires audit must succeed or fail with its audit event
8. AI traceability supplements audit and does not replace it

## 11. Document requirements

Document handling should preserve:

1. stable document identifiers
2. explicit document types
3. explicit lifecycle states such as draft, submitted, approved or rejected, posted where applicable, and reversed or voided where applicable
4. source-document linkage into downstream postings
5. durable numbering where accounting, tax, or operational correctness requires it
6. one canonical shared document model for identity, lifecycle, numbering, and posting linkage even when payload ownership stays with domain modules
7. inbound request references should remain outside the document-numbering model because they identify intake records rather than downstream business documents

High-level document families expected in thin v1:

1. work-order documents
2. invoice documents
3. payment or receipt documents
4. inventory receipt documents
5. inventory issue or adjustment documents
6. journal proposal or journal-entry documents where needed
7. AI-created draft proposals and pending actions

Thin-v1 completion rule for adopted document families:

1. a supported thin-v1 document family is not considered complete merely because its type is registered in `documents`
2. adopted thin-v1 work-order, invoice, payment or receipt, and inventory document families should each have their owning payload truth implemented with one-to-one linkage back to the central document row

Canonical numbering rules:

1. accounting documents and entries should support explicit document types
2. document numbering should be durable and unique per configured series
3. numbering should not reset every financial year unless a later explicit compliant policy is adopted for a specific document class

## 12. Accounting and posting objectives

Accounting is a foundational v1 capability, not a later reporting add-on.

Requirements:

1. `accounting` owns ledger truth and posting boundaries
2. operational modules may prepare posting inputs but may not write posted ledger state directly
3. posting must remain centralized, balanced, idempotent, and correction-safe
4. reversal and correction flows must be explicit
5. accounting-period and numbering controls should remain possible on the shared core
6. receivable and payable control-account treatment belongs in accounting
7. the normal control lifecycle should remain draft -> submitted -> approved -> posted where posting applies
8. proposer, submitter, approver, poster, and timestamps should remain reconstructible for audit and approval review
9. separation of duties between approver and poster should be policy-configurable rather than imposed as one hard rule for every org and document class

## 13. Tax objectives

Foundational GST and TDS support are part of v1.

Rules:

1. GST and TDS belong in supported document and posting flows, not only in reports
2. tax metadata should be attachable to parties, documents, document lines, and accounting outputs where needed
3. tax flows must be suitable for a very small service company and a light-trading operator on the same foundation
4. deeper localization breadth and broader country-specific depth remain later concerns

## 14. Inventory and material-flow objectives

Inventory should use one shared truth model across service-led and light-trading scenarios.

Requirements:

1. one inventory foundation, not separate service and trading inventory engines
2. on-hand stock must be derived from movements
3. movement source and destination must be explicit
4. resale stock, service-delivery materials, installed or traceable equipment, and direct-expense consumables must be distinguished explicitly
5. billable versus non-billable service-material usage must be explicit where costing or billing depends on it
6. serialized, lot-tracked, or installed-unit traceability should exist where the delivery use case requires it
7. inventory usage should link cleanly into execution context and accounting outcomes

## 15. Execution-context objectives

`work_order` is the strongest long-term operational capability and should remain the primary execution record.

Required execution rules:

1. work orders are the main execution context
2. projects are optional and subordinate when present
3. one shared task engine exists across contexts
4. each task has one primary actionable owner
5. team ownership is a queue concept, not many simultaneous primary assignees
6. tasks and activities are distinct concepts
7. worker-linked labor capture and labor costing belong in thin v1
8. execution history should remain explicit and auditable
9. work should link to documents, labor usage, material usage, and financial outcomes
10. serviced assets or installed units should remain first-class linked records when work targets a specific maintainable unit

## 16. Parties, items, and supporting master data

High-level master-data requirements:

1. parties are unified external entities and should not be split into isolated customer/vendor truth models
2. shared foundation entities should be referenced across modules through one canonical identity rather than duplicated into module-local truth models
3. contacts exist as supporting identity detail, not as the center of product scope
4. items are a shared foundation across service and light-trading scenarios
5. behavior differences should come from attributes and classification rather than separate core item models
6. inventory locations, ledger accounts, and tax-foundation records are required shared foundations
7. internal identity, external party identity, and worker identity must remain separate concerns

## 17. Audit, approvals, and idempotency

These are non-negotiable control boundaries.

Rules:

1. every meaningful business mutation must be auditable
2. approvals should be explicit and persisted for sensitive actions
3. idempotent write execution is required where retries could create duplicate effects
4. accepted AI-originated actions should carry causation links into audit metadata
5. the system should preserve enough trace data to reconstruct who proposed, approved, submitted, and posted an action

## 18. API, mobile, and client requirements

The backend is intended to serve AI, mobile, and later web or portal clients through the same domain-service boundaries.

Current high-level client rules:

1. `/api/v1/...` is the explicit stable major-version path
2. `/api/...` may remain a same-shape pre-production alias for the current version
3. v1 changes should remain additive and backward-compatible
4. request validation should happen before persistence calls
5. device-scoped sessions and refresh-token rotation are the intended mobile-auth model
6. retry-prone writes should use idempotent boundaries where duplicates would be harmful
7. list endpoints intended for mobile use should support pagination and incremental-sync-friendly shapes
8. attachment transport should use explicit bounded upload/download contracts
9. notification registration and delivery bookkeeping are backend responsibilities
10. the current mobile stance is online-first unless a later canonical decision changes it
11. the first planned mobile client may use Flutter, but backend decisions must remain client-agnostic
12. if thin v1 needs real browser-based user testing, the minimum promoted client slice should be persist-first request ingest, queued AI processing, and review-oriented web support rather than broad operational UI
13. the queued persisted-request model should also be usable by non-human upstream systems so integrations do not require a second intake architecture

## 19. Reporting objectives

Reports are first-class outputs but not first-class truth owners.

Required reporting categories at high level:

1. approval views
2. document lists
3. journal and ledger views
4. inventory stock and movement views
5. work-order queues and status views
6. tax summary and review views
7. audit lookup views

Humans should inspect system state through these derived views rather than by mutating truth tables directly.

## 20. Later-version objectives and deferred areas

The repository preserves several future-facing seams and later-version objectives, but these are not the center of thin v1.

Common v2-or-later objectives include:

1. broad CRM pipeline depth
2. advanced project-management breadth
3. customer portal
4. broad launch/navigation UX
5. CSV and spreadsheet exchange workflows
6. payroll
7. deeper tax localization
8. external communication-channel integrations
9. rental-operator extensions
10. narrow marketplace-seller or broader trading extensions
11. mobile speech-capture workflows using approved local-language transcripts
12. full UAE statutory and localization support beyond shared-core compatibility

These should remain future-compatible where practical, but they should not distort thin-v1 implementation priorities.

## 21. Multi-version completion framing

This document does not define one single version-completion gate for the whole product.

Rules:

1. thin-v1 completion should be judged against the active `new_app_docs/` scope and foundation checklist
2. later-version objectives should be delivered only after they are promoted into active canonical planning for that version
3. multi-version ambition should not be used to weaken thin-v1 discipline

## 22. High-level completion test

At a high level, the intended system should let a business:

1. submit a bounded inbound request from a human or another system
2. persist that request before AI processing begins
3. let AI process the queued request into a draft, proposal, or recommended action
4. review the resulting proposal through explicit review surfaces
5. approve and post through controlled services where policy allows
6. inspect document, execution, inventory, accounting, tax, and audit outcomes through reports
7. trust that correctness comes from constrained documents, ledgers, execution links, approvals, queued request traceability, and database-enforced invariants rather than mutable convenience fields
