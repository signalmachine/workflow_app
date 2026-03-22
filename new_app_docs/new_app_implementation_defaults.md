# workflow_app Implementation Defaults

Date: 2026-03-19
Status: Draft canonical implementation defaults
Purpose: record the active defaults that implementation should preserve unless the canonical `workflow_app` planning set is explicitly updated.

## 1. Default rules

1. this document records active defaults, not open brainstorming
2. if a new decision changes one of these defaults, update this file and the relevant companion planning doc
3. if code conflicts with this file, either fix the code or revise the active planning docs explicitly
4. legacy reference docs may be used for slice detail, but only the active `new_app_docs/` set is canonical

## 2. Locked defaults

### 2.1 Workflow ownership

1. use one shared `tasks` engine across v1 contexts
2. task orchestration belongs to `workflow`
3. each task has one primary actionable owner
4. team ownership is a queue concept, not many simultaneous primary assignees
5. tasks and activities are different concepts and must not be collapsed
6. shared approval orchestration and approval queues belong to `workflow`, even when AI or domain modules trigger the approval need

### 2.2 AI write boundary

1. AI may read, summarize, draft, recommend, and request approval
2. AI may execute bounded writes only through approved tools and normal domain services
3. financially meaningful writes remain policy-gated, with human gating as the thin-v1 default
4. meaningful business writes and their audit trail must succeed or fail together
5. AI traceability records supplement audit; they do not replace it
6. inbound user requests should persist before AI processing begins so asynchronous execution does not depend on synchronous request-response handling
7. the primary interaction model may be queue-first and review-oriented rather than immediate-response by default
8. draft inbound requests must not be processed by AI until explicitly queued or submitted
9. cancellation of parked requests should normally be soft cancel or soft delete rather than unrestricted hard deletion
10. AI workers must not claim cancelled, hidden, or incomplete draft requests
11. the first live provider-backed thin-v1 execution path should use the OpenAI Go SDK and the Responses API
12. provider-backed AI verification should be opt-in so the default repository build and test flow does not require external API credentials

### 2.3 Document identity and ownership

1. supported business and accounting documents should have stable identifiers
2. document types remain explicit
3. durable numbering exists where accounting, tax, or operational correctness requires it
4. numbering should be unique per configured series
5. numbering should not reset every financial year unless a later explicit compliant policy is adopted for a specific document class
6. `documents` owns shared document identifiers, lifecycle state, numbering, and posting-linkage contracts
7. `accounting`, `inventory_ops`, and `work_orders` own their domain-specific payloads and business rules
8. every supported business document family uses exactly one central `documents` row per document
9. the preferred table shape is a direct `document_id` link from the domain payload row to the central `documents` row, with one-to-one semantics enforced
10. central ownership-routing fields may exist in `documents`, but they do not replace the one-to-one contract between document identity and module-owned payload truth
11. adopted thin-v1 document families are not complete until their owning payload tables exist, including minimum work-order, invoice, and payment or receipt payload support

### 2.4 Accounting and posting

1. `accounting` owns the posting boundary and ledger truth
2. operational modules may prepare posting inputs but may not write posted ledger state directly
3. posting must be explicit, idempotent, balanced, and correction-safe
4. the normal lifecycle remains draft -> submitted -> approved -> posted where posting applies
5. AI may propose and, where policy allows, submit
6. final posting remains human-controlled by default in thin v1, but the architecture should preserve room for tightly policy-governed AI posting on selected document or entry classes later
7. separation of duties between approver and poster should be policy-configurable so some orgs may require different actors while small operators may allow the same actor to approve and post

### 2.5 Inventory and execution flow

1. use one shared inventory foundation for service-led and light-trading operations
2. do not create a second trading-specific inventory model
3. distinguish resale stock, service-delivery materials, installed or traceable equipment, and direct-expense consumables explicitly
4. billable versus non-billable material usage must be explicit where costing or billing depends on it
5. `work_order` is the primary execution record when work-order context exists
6. `project` is optional and subordinate if it exists
7. inventory consumption may attach to work orders or other minimal supported execution contexts without requiring a broad projects module
8. serialized, lot-tracked, or installed-equipment identity should be preserved where the delivery use case requires it

### 2.6 Support records

1. thin v1 includes minimum party and contact support depth for trading and service document flows
2. party and contact support records do not justify a primary CRM module
3. support-record depth should stay anchored to document, accounting, inventory, and execution correctness rather than commercial CRM breadth
4. shared foundation entities should use one canonical identity across modules rather than separate module-local truth models

### 2.7 Workforce and identity

1. worker identity remains distinct from login identity
2. external party identity remains distinct from worker identity
3. worker-linked labor capture is part of thin-v1 foundation, not a v2-only extension
4. assignment, time capture, and labor costing should fit together without requiring payroll to exist first

### 2.8 Tenant and session model

1. `org` is the canonical tenant unit
2. a deployed instance may host multiple orgs on one shared application foundation
3. a user may have memberships in multiple orgs
4. role and access decisions belong to the membership within the active org
5. a user may have different roles in different orgs
6. one session carries one active org context at a time
7. default-org selection may exist as a convenience, but it is not an authorization source
8. org switching must be explicit and must re-establish the active org context safely
9. tenant-owned reads and writes must always execute against the active `org_id`

### 2.9 Interface stance

1. the intended product surfaces are AI plus a usable web application layer for intake, review, approval, inspection, and reporting
2. the web layer should stay aligned with the AI-agent-first operating model rather than becoming a broad manual-entry product by default
3. CLI tooling may exist for developer or support work, but it is not a first-class product interface
4. the web layer should use backend contracts that a later mobile client can also reuse rather than diverging into a second backend model
5. mobile-product depth, voice-capture UX, and richer multimodal client behavior remain v2 concerns unless a foundation dependency proves otherwise
6. the preferred thin-v1 web stack is Go server-rendered HTML on the shared Go backend without a separate Node build pipeline
7. `htmx` is the preferred progressive-enhancement layer where partial-page updates materially improve operator continuity while preserving server-rendered ownership
8. `Alpine.js` is acceptable only for small local UI-state needs such as disclosure, tabs, or lightweight interaction polish
9. do not introduce a separate SPA frontend, Node dependency tree, or frontend-specific build chain unless the canonical planning set is explicitly updated to justify that cost

### 2.10 Inbound request and attachment handling

1. a persisted inbound request may contain one or more messages and attachments
2. draft requests remain editable until explicitly submitted into the processing queue
3. draft requests may be hard-deleted completely while they remain unprocessed drafts
4. queued or otherwise parked pre-processing requests should default to soft cancel semantics so auditability and recovery remain intact
5. a queued or cancelled pre-processing request may return to `draft` for amendment and later resubmission, but requests that have already started AI processing must not use that amend path
6. original uploaded artifacts, including voice recordings, should remain durably available even when derived text or other extracted artifacts are created
7. for thin-v1 development and early testing, attachment binary content may live in PostgreSQL, but the storage contract should preserve a later move to external object storage
8. persisted inbound requests should have a stable user-visible reference or request number suitable for operator and customer communication rather than relying on raw UUIDs
9. when a request is submitted or queued, the caller should receive that reference immediately in the acknowledgment response
10. if drafts exist, the preferred default is to allocate the stable request reference when the draft is created so later queueing, cancellation, amendment, audit, and support flows all keep one identifier
