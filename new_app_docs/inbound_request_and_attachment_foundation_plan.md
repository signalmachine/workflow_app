# workflow_app Inbound Request And Attachment Foundation Plan

Date: 2026-03-21
Status: Completed implementation slice with follow-up guidance
Purpose: record the implemented persist-first interaction foundation for thin v1 and capture the remaining follow-up guidance for later reporting polish.

## 1. Problem statement

The revised thin-v1 direction required persisted inbound request intake, queue-oriented AI processing, attachment references, and browser-usable review visibility.

State before the slice:

1. AI run history, artifacts, recommendations, and delegation traces existed
2. there was no persisted inbound request model
3. there was no durable request lifecycle status for queued or processed handling
4. there was no attachment persistence model beyond an empty schema
5. reporting did not yet expose inbound-request or processed-proposal review paths

Agreed operating constraints for this slice:

1. inbound requests may start as drafts and must not be processed by AI until explicitly submitted or queued
2. queued or otherwise submitted-but-unprocessed request deletion should be implemented as soft cancel rather than unrestricted hard delete
3. strong auditability should be preserved even when users cancel parked requests before AI picks them up
4. for thin-v1 development and testing, attachment binary content may be stored in PostgreSQL first, with the design leaving room to move blobs to external storage later
5. voice input should preserve the original audio attachment and store transcription as a derived record rather than replacing the original artifact

## 2. Remediation objective

Land the minimum persist-first interaction foundation so:

1. inbound user or system requests persist durably before AI processing begins
2. AI runs can link back to the request that caused them
3. attachment references can be stored safely for approval evidence, document support flows, and inbound request intake
4. humans can review request status, resulting proposals, and downstream document outcomes through minimal browser-usable review support
5. parked requests can be drafted, queued, cancelled, and reviewed without weakening auditability or queue correctness

## 3. Scope

In scope:

1. inbound request persistence
2. durable request lifecycle status
3. linkage from inbound requests to AI runs and downstream proposals or actions
4. attachment metadata and bounded attachment-reference contracts
5. minimal reporting surfaces for inbound-request and processed-proposal review
6. queue-oriented processing seams even if the first implementation still uses a simple worker or synchronous trigger path internally
7. draft editing, draft hard deletion, and pre-processing amend-back-to-draft behavior

Out of scope:

1. broad web application product depth
2. multimodal client breadth beyond the minimum attachment-reference and browser-testing seam
3. consumer-style chat UX
4. advanced autonomous agent behavior

## 4. Recommended target model

### 4.1 Inbound requests

Recommended minimum fields:

1. request id
2. stable user-visible request reference or request number
3. `org_id`
4. active session or actor linkage where present
5. request origin type such as human or system
6. request channel such as browser or api
7. original request text or payload
8. lifecycle status
9. queue timestamps for received, started, completed, failed, and acted-on states
10. optional resulting proposal, recommendation, or document references
11. created and updated timestamps
12. cancellation or soft-delete metadata where applicable

Recommended status set:

1. draft
2. queued
3. processing
4. processed
5. acted_on
6. completed
7. failed
8. cancelled

Recommended control rules:

1. a request may remain editable while in `draft`
2. AI workers may only claim requests in `queued`
3. submitting a completed draft transitions it into `queued`
4. a user may cancel a request while it is `draft` or `queued`
5. once processing starts, normal user behavior should not hard-delete the request
6. hidden or cancelled requests must not be eligible for worker pickup
7. the system should return the stable request reference immediately when a request is submitted or queued so the user can track it without depending on raw UUIDs
8. the preferred design is to allocate that reference at request creation time rather than waiting until queueing so draft, audit, support, and recovery flows all refer to one stable identifier
9. while a request is still `draft`, a user may add detail or edit existing detail before queueing
10. a `draft` request may be hard-deleted completely because it has not yet entered the AI processing queue
11. a `queued` or cancelled pre-processing request may return to `draft` for amendment and later resubmission, but that amend path must be blocked once AI processing has started

Recommended message model:

1. a request may contain one or more persisted messages
2. each message may include text, voice, pictures, or document attachments
3. request eligibility for queueing should be explicit rather than inferred from partial message state

Recommended user-facing submission response:

1. when a request is submitted into the queue, the system should return a response equivalent to `request submitted for processing with reference REQ-000123`
2. raw database ids may still exist internally, but the default user-facing acknowledgment should prefer the stable request reference or number
3. in the current codebase, this acknowledgment is implemented as service and API semantics, not as proof that a full browser UI is already shipped

### 4.2 Attachment support

Recommended minimum fields:

1. attachment id
2. `org_id`
3. PostgreSQL-backed blob content or an abstracted storage locator contract that can later point to external storage
4. original file name
5. media type
6. size metadata
7. uploaded-by actor linkage where present
8. created timestamp

Recommended thin-v1 attachment rule:

1. store attachment content in PostgreSQL during development and early thin-v1 testing
2. keep attachment metadata and attachment-content addressing explicit so the storage backend can move later without changing request semantics
3. enforce bounded size and media-type validation so database growth remains controlled during thin-v1

Recommended linkage model:

1. use explicit join tables or typed reference rows rather than overloading attachments with many nullable foreign keys
2. support at least inbound-request attachment references in the first slice
3. allow later attachment references from approvals or documents without changing the core attachment metadata model
4. preserve the original uploaded artifact even when derivative artifacts such as transcriptions are generated

Recommended voice handling:

1. keep the original audio attachment as an auditable artifact
2. store transcription output as a derived record linked back to the original request message or attachment
3. explicitly decide whether queue eligibility waits for transcription completion or allows a later enrichment step, rather than leaving that behavior implicit

### 4.3 AI linkage

1. add explicit linkage from inbound requests to AI runs
2. allow one request to produce one or more runs while preserving causation
3. preserve current run, step, artifact, recommendation, and delegation observability
4. keep approval truth in `workflow`, with AI only linking into it

### 4.4 Reporting and review

Minimum review outputs:

1. inbound request list with status and timestamps
2. inbound request detail with linked AI runs, recommendations, approvals, and resulting documents where present
3. processed-proposal review sufficient to inspect what the system produced from a request
4. attachment-reference visibility sufficient for operator review without broad file-management product depth
5. these outputs may exist first as reporting read models and API-ready service seams before a browser surface is shipped

## 5. Milestone breakdown

### 5.1 Schema work

1. add inbound-request tables with durable status support
2. add stable user-visible request reference or numbering support
3. add attachment metadata and attachment-reference tables
4. add foreign keys or reference rows linking inbound requests to attachments
5. add explicit inbound-request linkage into AI runs or adjacent causation tables
6. add explicit draft, cancellation, and worker-claim support so queue pickup rules stay database-visible
7. add derivative-artifact support for transcriptions or equivalent extracted text

### 5.2 Service-layer work

1. add persist-first intake service APIs
2. separate request persistence from AI processing initiation
3. ensure request status transitions are transactional and auditable where required
4. keep queue semantics explicit even if the first execution engine is intentionally simple
5. implement soft cancel for queued or otherwise parked requests instead of unrestricted hard delete
6. prevent workers from claiming cancelled, hidden, or incomplete draft requests
7. return the stable request reference to the caller when a request is submitted or queued
8. keep draft requests editable before queueing, allow draft hard deletion, and allow queued or cancelled pre-processing requests to return to `draft` for amendment

### 5.3 Reporting work

1. add inbound-request list and detail read models
2. add processed-proposal review models that join request, AI, approval, and document outcomes
3. keep the browser-testing surface review-oriented, not operational-UI-heavy

### 5.4 Test work

1. integration tests for request persistence before AI processing
2. tests for request status transitions
3. tests for attachment reference linkage and tenant safety
4. tests for inbound request to AI run causation and reporting joins
5. tests for draft requests remaining unprocessed until explicitly queued
6. tests for pre-processing cancellation preventing worker pickup while preserving reviewability
7. tests for voice attachment plus transcription linkage
8. tests for stable request-reference allocation and submission acknowledgment behavior
9. tests for draft hard deletion and draft-only artifact cleanup
10. tests for queued-request amend-back-to-draft behavior before AI pickup

## 6. Risks and technical challenges

1. queue semantics can sprawl if request lifecycle states are added without a clear control model
2. attachment design can become over-generalized unless the first slice stays limited to metadata plus explicit references
3. AI causation joins can become ambiguous if inbound-request linkage is split across too many tables
4. browser-testing support must stay narrowly review-oriented or it will pull the codebase back toward broad UI work
5. storing attachment content in PostgreSQL can increase database size, backup cost, and query-operability risk unless bounds stay explicit
6. soft cancel semantics need careful worker-claim logic so cancelled or hidden requests cannot race into processing
7. transcription introduces a second asynchronous lifecycle that must not leave queue eligibility ambiguous
8. assigning a user-visible reference only at queue time can create audit and support confusion if drafts already exist without the later stable reference

## 7. Recommended sequencing

This slice should land after adopted document ownership.

Reason:

1. request intake and processed-proposal review will sit on top of document and approval flows
2. stabilizing adopted document ownership first reduces the chance that request review surfaces need a second major redesign
3. the current biggest structural inconsistency is document-family adoption, not AI traceability

## 8. Success criteria

Landed implementation:

1. `ai.inbound_requests` and `ai.inbound_request_messages` now persist request intake before AI processing
2. stable `REQ-...` request references are now allocated at draft creation time and preserved through queueing, cancellation, and amendment
3. queued requests can be claimed into processing, and AI runs now link back to the originating request
4. `attachments` now stores PostgreSQL-backed attachments, request-message links, and transcription-derived text records
5. `reporting` now exposes inbound-request list and detail review plus processed-proposal review through service-level read models
6. drafts remain editable, drafts may be hard-deleted completely, queued requests may be soft-cancelled before pickup, and queued or cancelled pre-processing requests may return to `draft` for amendment and later resubmission
7. the current browser-ready foundation is implemented as intake and reporting service semantics rather than a shipped browser UI

This remediation slice is complete only when:

1. a request can persist durably before AI processing
2. a request can carry durable status through queue-oriented handling
3. attachment references exist for the allowed thin-v1 use cases
4. AI runs and resulting proposals or documents can be traced back to the originating request
5. reporting exposes inbound-request and processed-proposal review paths sufficient for thin-v1 browser testing
6. draft requests are not processed until explicitly queued
7. pre-processing cancellation is supported through soft delete or cancel semantics rather than hard deletion
8. original voice or file attachments remain auditable even when derivative records such as transcriptions are created
9. a stable user-visible request reference exists and is returned to the caller when the request is submitted for processing
