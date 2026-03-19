# service_day AI Architecture v1

Date: 2026-03-11
Status: Legacy reference
Purpose: preserve the broader pre-thin-v1 AI architecture for historical context and older implementation detail.

Legacy note:
1. active thin-v1 AI guidance now lives in `plan_docs/service_day_ai_architecture_v1.md`
2. use this file only when a specific older AI decision needs clarification

## 1. AI Goals

The AI subsystem should:
1. deliver useful assistance in the first CRM MVP
2. remain bounded, auditable, and tenant-safe
3. operate through explicit tools and contracts
4. support later expansion into broader service operations
5. avoid creating a second hidden business workflow system
6. support later local-language speech-driven business-event capture on mobile without letting raw speech recognition results mutate product state directly

## 2. Architecture Direction

The AI architecture is:
1. OpenAI-first
2. adapter-based
3. Responses-API-first for new agentic features
4. strict-schema-first for action-oriented outputs
5. approval-aware for sensitive actions

This means:
1. OpenAI is the first implemented provider
2. the application should still define provider-agnostic contracts
3. prompts, tools, approvals, artifacts, and accepted outputs must be persisted

## 3. Design Principles

1. AI must not bypass domain ownership.
2. AI outputs that affect operations are first-class records, not transient text blobs.
3. Human review is default for meaningful writes.
4. Tool access must be org-aware and role-aware.
5. AI should enrich the product workflow, not replace product state models.
6. The system should be debuggable from stored run history.
7. proposal-first and auditable-write claims are only true when accepted business actions and their audit trail persist atomically.
8. AI traceability records do not replace audit events; accepted AI-originated business changes must appear in both systems with causation links.

## 4. Recommended Architecture Shape

The AI subsystem should be split into these layers:

1. `ai` application contracts
2. provider adapter layer
3. tool registry and tool-policy layer
4. run orchestration layer
5. approval layer
6. persistence layer

### 4.1 Application Contracts

The rest of the app should depend on stable AI contracts such as:
1. summarize CRM notes
2. draft follow-up reply
3. propose next task
4. draft estimate from notes
5. evaluate billing readiness later
6. transcribe and normalize an internal business-event utterance from a supported local language into an approval-ready same-language text artifact for later mobile workflows

The application should not depend directly on provider-specific request/response shapes.

### 4.2 Provider Adapter

The first provider adapter should:
1. call the OpenAI Responses API
2. support structured output and tool use
3. support persisted conversation/run context where helpful
4. support background execution for longer runs later

Provider adapters should expose:
1. text generation
2. structured generation
3. tool-capable reasoning runs

### 4.3 Tool Layer

AI tools should be explicit functions over domain services.

Examples:
1. fetch account summary
2. fetch opportunity timeline
3. list recent communications
4. create draft task proposal
5. create draft estimate proposal
6. request approval for sensitive action

Critical rule:
1. tools may call owning module services
2. tools may not write directly to raw tables

## 5. AI Run Lifecycle

Each AI interaction should map to persisted run records.

Recommended flow:
1. request enters AI application contract
2. run record is created
3. prompt, transcript, and context are assembled
4. provider adapter executes the run
5. tool calls, outputs, and approvals are persisted
6. final accepted result is stored as an artifact or recommendation
7. any accepted business action is applied through normal domain services

Speech-capture interpretation:
1. raw audio or low-confidence transcript output is not itself a business command
2. for mobile speech-entry flows, the first durable AI output should be an approval-ready transcript artifact in the same user language where practical
3. only user-approved transcript text may proceed into downstream event interpretation or action recommendation flows

## 6. Persistence Model

Recommended core records:
1. `agent_runs`
2. `agent_run_steps`
3. `agent_artifacts`
4. `agent_recommendations`
5. `agent_tool_policies`
6. `agent_approvals`

Optional later support:
1. `agent_memories`

### 6.1 `agent_runs`

Should capture:
1. org
2. actor/user
3. capability name
4. provider/model
5. status
6. input summary
7. started/completed timestamps

### 6.2 `agent_run_steps`

Should capture:
1. prompt step
2. tool call step
3. tool result step
4. approval request step
5. final-output step
6. failure step

### 6.3 `agent_artifacts`

Should store:
1. generated summaries
2. drafted replies
3. drafted estimates
4. output documents or structured data blobs where appropriate
5. approved-transcript candidates and final approved transcript artifacts for later speech-driven entry flows where enabled

### 6.4 `agent_recommendations`

Should represent:
1. a proposed task
2. a proposed follow-up date
3. a proposed estimate structure
4. a proposed next-best action

Recommendation records matter because accepted proposals should be traceable to an AI-originated recommendation.

Audit interaction rule:
1. `agent_runs`, `agent_run_steps`, `agent_artifacts`, `agent_recommendations`, and `agent_approvals` explain AI reasoning and approval flow
2. `audit_events` remains the canonical ledger for the resulting business-state change
3. accepted AI actions should carry causation links such as `agent_run_id`, `recommendation_id`, and `approval_id` into the audit metadata where applicable

### 6.5 `agent_approvals`

Should support:
1. requested action summary
2. approver identity
3. approval state
4. approved/rejected timestamp
5. link to the recommended action or pending command

### 6.6 `agent_memories`

Use cautiously.

Recommended rule:
1. memory must be scoped by org and purpose
2. memory should not become hidden business truth
3. memory should not replace queryable domain records

## 7. Tool and Permission Model

Every AI tool should be governed by policy.

### 7.1 Tool Policy Inputs

Policy should consider:
1. org
2. actor role
3. capability
4. target record type
5. requested action type

### 7.2 Tool Classes

Use three broad classes:

1. read tools
   - safe reads over domain data
2. draft/proposal tools
   - create artifacts or recommendations
3. action tools
   - request or execute bounded product actions

### 7.3 Default Policy

1. reads are allowed when role and record access allow them
2. draft/proposal tools are broadly usable for internal users
3. action tools require explicit policy and often explicit approval
4. financial posting is never silently agent-executed
5. action-tool execution must use the same domain-service and audit transaction boundaries as non-AI writes
6. for accounting flows, AI may prepare or propose entries by default, may submit entries only when org policy explicitly permits it, and may never perform the final human posting step

## 8. Approval Model

Approvals should be explicit, persisted, and linked to concrete proposed actions.

### 8.1 Approval-Worthy Actions

Examples:
1. sending a customer-facing communication automatically
2. creating or materially changing commercial documents
3. generating operational tasks on behalf of another user
4. initiating financially meaningful changes later
5. submitting AI-prepared accounting entries when that submit action is policy-gated
6. accepting a speech-captured transcript for downstream business-event interpretation when the resulting action could create or change product state

### 8.2 Approval Flow

1. AI creates recommendation or pending action
2. approval record is created
3. authorized human reviews context and proposed action
4. approval or rejection is stored

Pragmatic first implementation note:
1. the first shipped approval-aware action path may create and approve the approval record in one bounded user-driven acceptance step, as long as the persisted approval record, policy lookup, and downstream audit causation are still explicit
2. richer queued approval inboxes and multi-step reviewer routing can be layered later without changing the core run or recommendation causation model
5. only approved actions continue into domain services

Accounting-specific interpretation:
1. AI-originated accounting entries should remain traceable as proposed or submitted finance actions before posting
2. an organization may configure AI to stop at proposal or to submit directly into a finance review queue
3. final posting must still be performed by a human with the configured accounting authority
4. posting directly from an AI proposal is allowed only as a human action through the same posting boundary and audit path used for non-AI work

### 8.3 Replay Safety

Approved action execution must be:
1. idempotent
2. traceable to the approval record
3. linked to resulting business records

## 9. Prompt and Context Assembly

Prompt assembly should be deliberate and capability-specific.

Rules:
1. use only the domain context necessary for the task
2. avoid overloading prompts with full record history when summaries or selected context are enough
3. assemble context through explicit queries or tools
4. keep prompt templates versioned where practical

Recommended context sources:
1. CRM summaries
2. recent activities
3. recent communications
4. relevant estimate lines or opportunity details

## 10. Structured Output Strategy

For action-oriented flows, prefer strict structured output over free text.

Examples:
1. task proposal object
2. follow-up draft object
3. estimate draft object
4. billing-readiness report later

Benefits:
1. safer parsing
2. clearer validation
3. easier audit
4. better replay and testing

## 11. Capability Rollout

### 11.1 Wave 1: CRM Assistance

Build first:
1. contact and lead summarization
2. note cleanup
3. opportunity health summary
4. follow-up draft generation
5. next-best-action suggestion
6. estimate draft assistance

### 11.2 Wave 2: CRM-to-Delivery Support

Build next:
1. conversion-readiness summaries
2. project/work-order setup suggestions
3. assignment suggestions later

### 11.3 Wave 3: Financial and Operational Intelligence

Build later:
1. work-order costing anomaly detection
2. billing readiness checks
3. collections follow-up recommendations
4. project risk summaries

## 12. Failure and Safety Handling

The AI subsystem should fail safely.

Rules:
1. provider failures must not mutate business state
2. partial tool execution must remain traceable
3. malformed structured outputs must be rejected or repaired explicitly
4. fallback behavior should return a recoverable user-facing failure, not silent degradation

## 13. Testing Strategy

Minimum expected tests:
1. tool-policy enforcement tests
2. approval-required flow tests
3. structured-output validation tests
4. tenant-scope enforcement tests
5. run persistence tests
6. idempotent action-execution tests

## 14. Operational Concerns

The AI subsystem should be observable.

Minimum expectations:
1. run status tracking
2. failure logging
3. prompt/template version traceability where practical
4. model/provider visibility per run
5. latency and retry handling later

## 15. Immediate Recommendations

Build the first AI layer with these constraints:
1. OpenAI adapter first
2. Responses API first
3. strict structured outputs for action flows
4. read and draft tools first
5. approvals before sensitive writes
6. full persistence of runs, steps, artifacts, recommendations, and approvals

If those constraints hold, AI can expand into deeper service-business workflows later without undermining core application integrity.
