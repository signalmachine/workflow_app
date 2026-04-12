# workflow_app Architecture V2

Date: 2026-04-09
Status: Active architecture guardrail
Purpose: keep the current architectural truth clear without requiring the older detailed planning set in normal sessions.

## 1. Core doctrine

1. `workflow_app` is workflow-centered, AI-agent-first, and database-first
2. documents, ledgers, execution context, approvals, and reports remain the center of gravity
3. every meaningful capability should support, constrain, observe, or expose one or more workflows

## 2. Shared backend rule

1. Go remains the owner of business logic, workflow rules, approvals, reporting composition, and durable state
2. the promoted web runtime is a Svelte application served on the same shared Go backend and auth model
3. `/api/...` is the stable application seam shared across the promoted browser runtime and any later client reuse
4. browser implementation must not fork business truth into browser-only logic

## 3. Ownership boundaries

1. core first-class modules remain `identityaccess`, `ai`, `documents`, `workflow`, `accounting`, `inventory_ops`, `workforce`, `work_orders`, `attachments`, and `reporting`
2. each module owns its own truth, write paths, and invariants
3. support records such as parties and contacts may deepen where needed, but they do not justify a primary CRM module
4. shared foundation entities should keep one canonical identity across modules rather than module-local duplicates

## 4. Persist-first operating model

1. inbound requests should persist durably before AI processing starts
2. AI processing should usually run asynchronously from that persisted queue
3. request records remain distinct from downstream business documents
4. every meaningful workflow and control state must be reconstructible from durable database records
5. AI triage must classify submitted business events before proposing writes: accounting events may move toward accounting proposals through approval and posting boundaries, while non-accounting events must not create accounting documents, journal proposals, or ledger entries
6. future non-accounting business-event persistence must be explicit and selective: only supported event types with defined prompts, backend services, and non-accounting tables may be recorded, and unsupported events should return a comment or missing-capability result rather than inventing persistence

## 5. Active implementation implication

1. current browser work should favor Svelte-surface completion plus backend seam hardening over more planning churn
2. when an existing area is visibly weak, prefer refactoring or rebuilding it over layering more work on top
