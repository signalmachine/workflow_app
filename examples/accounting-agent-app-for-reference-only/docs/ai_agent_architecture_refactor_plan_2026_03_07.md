# AI Agent Architecture Refactor Plan (Current Application)

Date: 2026-03-07  
Status: Proposed implementation plan for the existing `accounting-agent-app` codebase

## 1. Scope and Context

This document defines how to refactor the AI agent architecture in the **current application**.

Important boundary:
1. This is **not** a new-application design document.
2. This plan does **not** change or supersede `docs/new_application/new_app_blueprint_smb_gst_multi_agent_2026_03_06.md`.
3. This plan must fit existing production constraints, current data model, current services, and current UI/API routes.

## 2. Current-State Summary (As Implemented)

1. One general domain-action agent handles most non-deterministic chat behavior.
2. Journal-entry interpretation is a separate structured-output path.
3. Tooling is collected in one broad registry with mixed domains.
4. Chat API primarily exposes a single `/chat` entry point with no explicit domain scope.
5. Write actions are human-confirmed via pending tokens.

Implications:
1. Tool surface is broad, which can reduce routing determinism.
2. Domain specialization is implicit (prompt/tool description), not architectural.
3. There is no first-class "router + dedicated agents + direct page access" model yet.

## 3. Refactor Goals

1. Introduce a router-first orchestration model for chat requests.
2. Introduce workflow-dedicated agents with strict tool allowlists.
3. Support direct access to dedicated agents from specific UI areas (example: purchases pages).
4. Preserve existing human confirmation safety for write actions.
5. Minimize regressions by reusing existing `internal/app` and `internal/core` service methods.

## 4. Non-Goals

1. No major database schema rewrite is required for phase 1.
2. No replacement of existing core accounting/purchase/sales services.
3. No migration to a new application architecture.
4. No change to existing business policies unless explicitly required.

## 5. Constraints (Current Application)

1. Existing adapters (`web`, `repl`, `cli`) depend on `ApplicationService`.
2. Existing tool execution/confirmation flow must remain reliable.
3. Existing RBAC and company-scope checks must remain enforced.
4. Existing JE structured-output path must stay stable during initial phases.
5. Feature rollout should be incremental and reversible via configuration flags.

## 6. Target Architecture (Incremental)

## 6.1 Components

1. `RouterAgent` (or deterministic `RouterService` + LLM fallback):
   - Classifies intent into domain scope.
   - Chooses one target dedicated agent.
   - Requests clarification when confidence is low.
2. Dedicated workflow agents (initial set):
   - `AccountingAgent` (financial event -> JE path handoff)
   - `PurchaseAgent` (vendor/PO/receipt/invoice/payment flows)
   - `SalesAgent` (customer/order/invoice/payment flows)
   - `MasterDataAgent` (accounts/customers/vendors/products lookups/create proposals)
   - `ReportingAgent` (redirect and report-navigation guidance under current behavior)
3. `Coordinator` in app layer:
   - Executes routing decision.
   - Creates domain-specific tool registries.
   - Preserves confirmation and execution path for write tools.

## 6.2 Entry Modes

1. `auto` mode:
   - Default chat behavior.
   - Router selects dedicated agent.
2. `direct` mode:
   - Caller specifies a scope (`purchase`, `sales`, `accounting`, `master_data`, `reporting`).
   - Router is bypassed unless explicit fallback is needed.

## 6.3 Direct Page Access

1. Purchases UI calls chat with `agent_scope=purchase`.
2. Sales UI can call with `agent_scope=sales`.
3. General chat home continues with `agent_scope=auto`.

## 7. Proposed Code Changes

## 7.1 `internal/app` contract changes

1. Extend domain action request shape with optional scope:
   - `AgentScopeAuto`
   - `AgentScopePurchase`
   - `AgentScopeSales`
   - `AgentScopeAccounting`
   - `AgentScopeMasterData`
   - `AgentScopeReporting`
2. Add orchestration method in app service:
   - `InterpretScopedDomainAction(...)`
3. Keep current method as compatibility wrapper (`auto` scope).

## 7.2 `internal/ai` changes

1. Introduce reusable agent runner primitives:
   - common loop executor
   - per-agent prompt/tool bundle
2. Add router implementation:
   - deterministic rules first (keywords + context)
   - LLM fallback for ambiguous intent
3. Add dedicated agent definitions with strict tool lists.

## 7.3 Tool registry refactor

1. Split current monolithic registry builder into domain builders:
   - `buildPurchaseToolRegistry`
   - `buildSalesToolRegistry`
   - `buildAccountingToolRegistry`
   - `buildMasterDataToolRegistry`
   - `buildReportingToolRegistry`
2. Keep shared read tools in a small common module where needed.
3. Preserve existing write-tool execution path in `ExecuteWriteTool`.

## 7.4 Web adapter changes

1. Extend chat request payload with optional `agent_scope`.
2. Keep `/chat` endpoint; no breaking endpoint migration required.
3. Purchases pages pass `agent_scope=purchase`.
4. General chat home sends `agent_scope=auto`.

## 7.5 Security hardening (same refactor track)

1. Bind pending confirmation token to creator user ID.
2. Enforce same-user (or policy-allowed override) on `/chat/confirm`.
3. Keep company-scope checks as-is.

## 8. Phased Implementation Plan

## Phase 1: Scoped Entry + Registry Split

Deliverables:
1. `agent_scope` request support in web + app.
2. Domain-specific tool registry builders.
3. Purchase direct access from purchases pages.
4. Backward-compatible default behavior (`auto`).

Validation:
1. Existing chat flow still works without scope.
2. Purchase-scoped prompts only expose purchase-relevant tools.
3. Confirmation flow remains unchanged for write proposals.

## Phase 2: Router Introduction

Deliverables:
1. Router service with deterministic rules and optional LLM fallback.
2. Orchestration from `auto` scope to dedicated agents.
3. Clarification handling when routing confidence is low.

Validation:
1. Intent classification tests (table-driven).
2. Ambiguous prompts return clarification.
3. Misroute rates monitored via logs.

## Phase 3: Dedicated Agent Prompting

Deliverables:
1. Dedicated prompt/instruction bundles per agent.
2. Common agent loop reused across agents.
3. Reporting behavior kept deterministic and non-fabricated.

Validation:
1. Golden tests for tool-call traces.
2. Negative tests for out-of-domain tool calls.
3. Existing JE routing behavior remains stable.

## Phase 4: Safety + Observability Hardening

Deliverables:
1. Pending token ownership enforcement.
2. Structured audit logs include:
   - selected scope
   - router decision
   - chosen agent
   - handoff reason
3. Operational feature flags for rollback.

Validation:
1. Auth tests for confirm token ownership.
2. Audit fields present in logs for all chat actions.

## 9. Testing Strategy

1. Unit tests:
   - scope parsing
   - router classification
   - registry allowlist integrity
2. Integration tests:
   - scoped chat request -> expected tool proposals
   - confirm/cancel flow across scoped agents
3. Regression tests:
   - existing JE proposal path
   - existing purchase write actions
4. Optional safety suite:
   - adversarial prompts for cross-domain leakage

## 10. Rollout and Backward Compatibility

1. Feature flags:
   - `AI_AGENT_SCOPED_ROUTING_ENABLED`
   - `AI_AGENT_ROUTER_ENABLED`
   - `AI_AGENT_CONFIRM_TOKEN_BINDING_ENABLED`
2. Deployment order:
   - deploy phase 1 with routing disabled by default
   - enable scoped mode for internal users
   - enable router in stages
3. Rollback:
   - disable flags to return to legacy single-agent path.

## 11. Risks and Mitigations

1. Risk: Incorrect routing on ambiguous prompts.
   - Mitigation: low-confidence clarification response and conservative defaults.
2. Risk: Domain fragmentation increases maintenance.
   - Mitigation: shared agent loop and shared tool schema utilities.
3. Risk: UI mismatch in scoped pages.
   - Mitigation: default to `auto` if scope missing/invalid.
4. Risk: Behavior drift in write proposals.
   - Mitigation: retain existing execution backend and add regression tests.

## 12. Acceptance Criteria

1. Router + dedicated-agent orchestration exists for `auto` chat mode.
2. Purchases pages can directly invoke purchase-scoped agent behavior.
3. Dedicated agents have strict tool allowlists by workflow.
4. Human confirmation model remains intact for all writes.
5. Legacy chat usage remains functional without client-breaking changes.
6. New behavior is covered by unit and integration tests.

## 13. Recommended First PR Slice

1. Add `agent_scope` to chat request and app method wiring.
2. Extract purchase tool registry and keep others in legacy registry.
3. Route `agent_scope=purchase` to purchase-only registry.
4. Update purchases page client calls to pass `agent_scope=purchase`.
5. Add tests for:
   - scope parsing
   - purchase allowlist
   - no-regression on default auto flow.

