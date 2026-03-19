# service_day Implementation Defaults v1

Date: 2026-03-19
Status: Active thin-v1 locked defaults
Purpose: record the minimum active defaults that current implementation work should preserve unless the canonical thin-v1 docs are explicitly updated.

## 1. Default rules

1. this document records current implementation defaults, not open brainstorming
2. if a new decision changes one of these defaults, update this file and the relevant companion thin-v1 doc in `plan_docs/`
3. if code conflicts with this file, either fix the code or revise the active planning docs explicitly
4. during implementation, consult relevant `implementation_plan/` legacy docs whenever active thin-v1 docs need additional slice detail or historical clarity
5. when active and legacy docs conflict, `plan_docs/` wins; promote any still-needed legacy clarification into `plan_docs/` before treating it as active canon

## 2. Locked defaults

### 2.1 Workflow ownership

1. use one shared `tasks` engine across thin-v1 contexts
2. task orchestration belongs to `workflow`
3. each task has one primary actionable owner
4. team ownership is a queue concept, not many simultaneous primary assignees
5. tasks and activities are different concepts and must not be collapsed
6. shared approval orchestration and approval queues belong to `workflow`, even when AI or domain modules trigger the approval need

### 2.2 AI write boundary

1. AI may read, summarize, draft, recommend, and request approval
2. AI may execute bounded writes only through approved tools and normal domain services
3. financially meaningful writes remain human-gated
4. meaningful business writes and their audit trail must succeed or fail together
5. AI traceability records supplement audit; they do not replace it

### 2.3 Document identity and numbering

1. supported business and accounting documents should have stable identifiers
2. document types should remain explicit
3. durable numbering should exist where accounting, tax, or operational correctness requires it
4. numbering should be unique per configured series
5. numbering should not reset every financial year unless a later explicit compliant policy is adopted for a specific document class

### 2.4 Document ownership

1. thin v1 should preserve one canonical document-identity and lifecycle model across supported document families
2. `documents` owns shared document identifiers, lifecycle state, numbering, and posting-linkage contracts
3. `accounting`, `inventory_ops`, and `work_orders` own their domain-specific document payloads and business rules
4. no supported business document should rely only on screen-local or module-local ad hoc lifecycle state
5. `work_orders` owns execution-state progression and execution-specific business rules, but any shared document lifecycle participation still flows through `documents`
6. every supported business document family must use exactly one central `documents` row per document
7. for adopted document families, the preferred table shape is a direct `document_id` link from the domain payload row to the central `documents` row, with one-to-one semantics enforced
8. central ownership-routing fields may exist in `documents`, but they do not replace the requirement that central document identity and module-owned payload truth stay one-to-one

### 2.5 Accounting and posting

1. `accounting` owns the posting boundary and ledger truth
2. operational modules may prepare posting inputs but may not write posted ledger state directly
3. posting must be explicit, idempotent, balanced, and correction-safe
4. AI may propose and, where policy allows, submit; AI may never perform final human-controlled posting

### 2.6 Inventory and service-material flow

1. use one shared inventory foundation for service-led and light-trading operations
2. do not create a second trading-specific inventory model
3. distinguish resale stock, service-delivery materials, installed or traceable equipment, and direct-expense consumables explicitly
4. billable versus non-billable service-material usage must be explicit where costing or billing depends on it
5. serviced or installed-unit traceability must be explicit where the delivery use case requires it
6. reservation or allocation records are not part of active thin-v1 scope unless a later canonical update promotes them explicitly

### 2.7 Execution model

1. `work_order` is the primary execution record
2. `project` is optional and subordinate if it exists
3. serviced assets or installed units should remain first-class linked records when work targets a specific maintainable unit

### 2.8 Workforce and labor costing

1. worker identity remains distinct from login identity
2. worker-linked labor capture is part of thin-v1 foundation, not a v2-only extension
3. assignment, time capture, and labor costing should fit together without requiring payroll to exist first
4. labor cost visibility should attach cleanly to work orders and downstream accounting outcomes
5. richer timesheet governance may be phased, but raw labor facts and cost-traceable execution support belong in v1

### 2.9 API and mobile contract

1. `/api/v1/...` is the explicit stable major-version path for the current API surface
2. `/api/...` remains a same-shape alias during the current pre-production phase
3. changes within `v1` should remain additive and backward-compatible
4. request validation must happen before persistence calls for externally supplied inputs
5. mobile remains online-first until a different canonical decision is made
6. mobile auth should use device-scoped session and refresh-token lifecycle rules
7. retry-prone mobile writes should use idempotent execution where duplicate effects would be harmful

### 2.10 Interface stance

1. the intended product surfaces are AI, mobile clients, and later web or portal clients over the same backend domain services
2. CLI tooling may exist for developer or support work, but it is not a first-class product interface
3. human UI in thin v1 stays focused on review, approval, inspection, and reporting

### 2.11 Geography and localization stance

1. the commercial target market remains UAE
2. the first delivered statutory and localization baseline remains India
3. full UAE support is a v2 target and is not a thin-v1 implementation priority
4. the shared accounting, tax, document, and operational foundations should fit both India and UAE without requiring a core-model split
5. thin-v1 implementation should preserve that extensibility even when only India-first GST and TDS depth is implemented initially

## 3. Legacy-reference rule

If an older default still lives only in `implementation_plan/implementation_decisions_v1.md`:

1. do not treat it as current canon automatically
2. preserve it only when it fits the accepted thin-v1 direction
3. promote it here before relying on it as an active implementation default
