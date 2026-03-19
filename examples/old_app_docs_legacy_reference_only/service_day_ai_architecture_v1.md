# service_day AI Architecture v1

Date: 2026-03-18
Status: Active thin-v1 AI plan
Purpose: define the AI architecture for the thin-v1 system where AI is the main operator over strict business foundations.

## 1. AI objective

The near-term objective is not broad production rollout first.

The near-term objective is:

1. watch AI agents operate on real business foundations
2. verify that their planning, tool use, approvals, and posting boundaries are correct
3. make agent behavior observable, reviewable, and safe

This means v1 should optimize for controlled agent execution quality, not for maximum product breadth.

## 2. Core AI stance

AI is:

1. the main operator interface
2. a client of domain services
3. never the authority over truth

The database, posting rules, and approval boundaries remain the authority.

The active v1 architecture should use a multi-agent model:

1. one coordinator agent receives the user request
2. the coordinator decomposes the request into bounded sub-tasks
3. the coordinator routes sub-tasks to specialist agents
4. specialist agents return structured outputs, not uncontrolled side effects
5. the coordinator assembles the final plan, approvals, and execution sequence

## 3. Required thin-v1 agent capabilities

V1 agents should be able to:

1. interpret user intent into document-oriented tasks
2. gather context through explicit tools
3. draft work-order, invoice, payment, inventory, and tax-relevant document actions
4. request approvals where policy requires them
5. submit approved actions through normal domain services
6. explain what they did through durable run history and artifacts

Coordinator responsibilities:

1. classify user intent
2. select the correct workflow path
3. route work to the right specialist capability
4. request missing approvals
5. consolidate outputs into a final bounded action or review package

Specialist-agent responsibilities:

1. accounting and posting preparation
2. GST/TDS-aware document and posting assistance
3. inventory movement planning
4. work-order and task planning
5. reporting and reconciliation support

## 4. Required thin-v1 agent architecture

The v1 AI architecture should be modern and explicit, not a thin prompt wrapper.

Required components:

1. capability layer
2. tool registry
3. tool policy layer
4. coordinator agent layer
5. specialist agent layer
6. run orchestration layer
7. persistence layer
8. observability layer

Minimum expectations:

1. explicit read, draft, approval, and bounded action tools
2. org-aware and role-aware tool policy
3. coordinator routing over named specialist capabilities
4. run start, step execution, approval waits, and completion states
5. durable run, step, artifact, recommendation, and approval records
6. audit linkage and evaluation hooks
7. narrow typed tool inputs that map cleanly onto constrained domain services and database-enforced invariants
8. reuse of the shared workflow approval model rather than an AI-local approval system

## 4.1 Coordinator agent

The coordinator agent is the workflow router.

It should:

1. accept the user request
2. inspect current context
3. decide whether the request is accounting, tax, inventory, work-order, reporting, or mixed-domain work
4. call one or more specialist agents
5. enforce that cross-domain workflows still go through one consistent approval and execution plan

The coordinator should not:

1. bypass specialist logic for complex domain actions
2. post directly
3. turn into an unconstrained general-purpose writer

## 4.2 Specialist agents

Specialist agents are capability-bounded workers.

Each specialist should:

1. operate on one bounded domain capability
2. expose structured outputs
3. use only the tools allowed for that capability
4. return results to the coordinator for consolidation

Recommended initial specialists:

1. `accounting.posting_preparation`
2. `tax.gst_tds_review`
3. `inventory.movement_planning`
4. `workorders.execution_planning`
5. `reporting.reconciliation`

## 5. Agent quality rules

1. agents must reason over documents, ledgers, execution context, and tax context explicitly
2. agents must never write ledger rows directly
3. agents must never bypass approval or posting services
4. agents must produce durable artifacts that explain proposed actions
5. agents must leave enough trace data that failures can be inspected and corrected
6. agent tools should prefer constrained structured inputs over open-ended mutation surfaces
7. coordinator-to-specialist delegation should be explicit and traceable in stored run history

## 6. Approval and posting boundary

1. agents may draft
2. agents may recommend
3. agents may request approval in the shared workflow approval system
4. agents may submit only where policy allows
5. agents may never perform final human-controlled posting

This rule applies especially to:

1. accounting posting
2. inventory-affecting actions
3. GST-relevant invoice actions
4. TDS-relevant withholding actions

## 7. GST and TDS handling in AI flows

V1 agents do not need deep tax expertise across every edge case, but they must be foundation-aware.

Required behavior:

1. identify when GST treatment is required on a document
2. identify when TDS withholding context is relevant
3. draft tax-relevant document fields through structured outputs
4. route final tax effects through normal posting and approval boundaries
5. expose tax reasoning in reviewable artifacts when the action is materially tax-relevant

## 8. Observability-first design

Because the short-term objective is to watch agents in action, v1 should preserve:

1. full run and step history
2. tool call inputs and outputs where safe
3. approval records
4. causation links from accepted AI actions into audit records
5. enough evaluation metadata to compare expected versus actual outcomes later
6. delegation traces showing which coordinator request called which specialist capability

Approval ownership rule:

1. shared approval records and approval queues are part of the thin-v1 workflow and control boundary and are owned by `workflow`, not by `ai`
2. the `ai` module stores AI causation, recommendation, and artifact linkage into those shared approval flows
3. approval state for non-AI-originated sensitive actions should still use the same shared approval model

## 9. What thin v1 should avoid

1. hidden sidecar memory as business truth
2. direct provider-to-database write paths
3. agent flows that cannot be replayed or inspected
4. autonomous financial posting
5. autonomous tax finalization
6. UI-heavy AI experiences that distract from agent execution quality
7. broad human operational UIs that duplicate what the agent is supposed to operate
8. one giant undifferentiated agent with no bounded specialist roles for non-trivial workflows

## 10. Thin-v1 success condition for AI

The AI architecture is successful when:

1. a human can ask the agent to perform a business operation
2. the agent uses explicit tools and produces a reviewable proposal
3. approval and posting boundaries remain intact
4. resulting document, ledger, execution, and tax effects are traceable
5. the team can inspect how the agent behaved and improve it iteratively
