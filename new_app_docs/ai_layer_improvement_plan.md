# AI Layer Improvement Plan

Date: 2026-04-12  
Validated: 2026-04-12 (code-verified against live source)  
Status: Active — pending implementation  
Scope: `internal/ai/` — coordinator, provider, service, and surrounding seams  

---

## Overview

The AI agent layer in `internal/ai/` is built on a genuinely strong foundation. The persist-first data model, provider abstraction, tool policy governance, and output validation layer are all well designed and significantly better than most AI application layers in production Go codebases.

However, six concrete gaps limit the layer's correctness, extensibility, and production readiness. This document records every finding from a full code review of the three core files and prescribes actionable fixes for each.

### Files reviewed

| File | Lines | Role |
|---|---|---|
| `internal/ai/coordinator.go` | 937 | Coordinator orchestration, validation, request loading |
| `internal/ai/openai_provider.go` | 1088 | OpenAI Responses API integration, tool loop, prompt builder |
| `internal/ai/service.go` | 1528 | Domain data model, transactions, policy resolution, audit |

### Rating summary

| Area | Rating | Verdict |
|---|---|---|
| Data model and persistence | 9/10 | Excellent — durable, auditable, ACID at every step |
| Provider abstraction | 9/10 | Clean interface — easy to swap providers |
| Tool policy governance | 8/10 | Right architecture — one transaction overhead bug |
| Output validation and self-repair | 8/10 | Rare and well-implemented |
| Tool loop budget control | 6/10 | Timeout flawed; budget exhaustion has counting bug; `MaxOutputTokens` too low |
| Context loading | 6/10 | Bypasses service layer with inline SQL |
| Specialist delegation execution | 4/10 | Database scaffold exists, actual execution is a stub |
| Prompt management | 5/10 | Hardcoded strings, no config or versioning |
| **Overall** | **7/10** | Strong foundation — real gaps in specialist execution and hot-path efficiency |

---

## Architectural Assessment

This section evaluates the AI agent architecture at the design level — the structural decisions that determine how well the system will scale, operate under compliance requirements, and evolve over time. These findings are independent of the code-level gaps documented below.

### What the architecture gets fundamentally right

**1. Persist-first async queue — not synchronous inline AI calls**

Every AI operation goes through a durable queue. The HTTP layer accepts and stores the request; the AI processor picks it up separately and asynchronously. This is the most important architectural decision in the system.

The alternative — calling the AI model inline in the HTTP request path — produces: 10–45 second response times, lost processing on process restart, no retry capability, and no audit trail. Every serious production AI application eventually migrates to the queue-based model. This application starts there.

**2. Database as source of truth, not the provider's session**

Every coordinator action writes committed transactions to PostgreSQL: runs, steps, tool executions, artifacts, recommendations, delegations, audit events. The entire AI execution history is reconstructable from database records alone, without provider cooperation.

This is the posture required for software with compliance, audit, and regulatory obligations. Storing state in `thread_id` or `session_id` on the provider side creates a dependency on the provider's data retention policy and eliminates the ability to reproduce or audit what the AI saw.

**3. AI proposes; humans approve; accounting commits**

This three-layer separation is the correct doctrine for business AI:

```
AI proposes     →  coordinator output → recommendation
Human approves  →  operator reviews and approves in the UI
System commits  →  accounting entries, workflow state changes
```

The tool policy governance system (`allow / approval_required / deny`) makes the approval gate configurable per capability and per tool without a code change. This is not a safety afterthought — it is a correctly designed approval boundary that reflects how business operations actually work.

**4. Provider abstraction with stateless continuation**

The `CoordinatorProvider` interface and `Store: false` stateless continuation mean the application owns all conversation state. The provider is a stateless function: input in, output out. Swapping providers, auditing executions, and reproducing results require no provider cooperation. This is the compliance-friendly and operationally robust posture.

**5. Coordinator + specialist routing pattern**

A general coordinator that routes to domain specialists is the established best practice for multi-domain business AI. The pattern is validated by OpenAI's orchestration guidance, Anthropic's multi-agent research, and production enterprise AI implementations. The current codebase has the correct scaffold; it needs real specialist execution (Gap 1).

**6. Structured output with schema enforcement and self-repair**

Forcing LLM output into a strict JSON schema eliminates the "parse free-text AI response" problem at the source. The three-layer validation guard (structural → semantic → repair → fail) is rare in production AI applications and directly reduces silent quality failures.

---

### Architectural Gap 1 — Queue triggering is polling-based, not event-driven

**Severity: Medium (operability and scalability)**

The coordinator queue processor almost certainly runs on a polling loop — waking up every N seconds to call `ProcessNextQueued`. Polling has two concrete problems:

- **Response latency**: A request submitted halfway through a poll interval waits N/2 seconds before processing begins. Combined with AI processing time, operators experience visible lag before a brief appears in the review surface.
- **Idle resource use**: The worker wakes and queries the database on every cycle regardless of whether there is work to do. Under low request volume, this is constant low-level database noise.

**Fix — PostgreSQL `LISTEN/NOTIFY`**

PostgreSQL provides native event notification with no additional infrastructure. When `intakeService.AdvanceRequest` transitions a request to `queued`, it sends a notification:

```go
// internal/intake/service.go — in the same transaction as the status change
// pg_notify inside a transaction is delivery-deferred: the notification fires only
// if the transaction commits, so missed notifications on rollback are not possible.
_, _ = tx.ExecContext(ctx, "SELECT pg_notify('ai_coordinator_queue', $1)", requestID)
```

The coordinator worker replaces its sleep loop with a `LISTEN` connection:

```go
// internal/ai/coordinator_worker.go
listenConn, _ := db.Conn(ctx)
_, _ = listenConn.ExecContext(ctx, "LISTEN ai_coordinator_queue")
// Block on WaitForNotification; wake immediately when a request is queued
// Fall back to polling every 30s as a safety net for missed notifications
```

This reduces average processing latency from N/2 seconds to near-zero with no new infrastructure dependency. For higher throughput (hundreds of concurrent requests), a message broker (NATS, Redis Streams) is the next step — but `LISTEN/NOTIFY` is the right first upgrade.

---

### Architectural Gap 2 — No proactive or scheduled AI capability

**Severity: Low-Medium (feature ceiling)**

The entire system is reactive — the AI only processes inbound requests submitted by users or systems. There is no mechanism for the coordinator to:

- Alert operators when a purchase order is approaching its budget limit
- Flag a high-priority request that has been queued for too long
- Produce a weekly queue-volume summary for operations review
- Detect anomalous patterns across recent requests

The `Run` data model already anticipates this: `InboundRequestID` is `sql.NullString` — nullable — which means a run does not require an inbound request. The foundation for proactive runs exists in the schema.

**Fix — Add a `ScheduledRun` concept**

```go
// internal/ai/service.go
type ScheduledRunInput struct {
    CapabilityCode string
    TriggerType    string   // "scheduled", "threshold", "anomaly_detection"
    TriggerReason  string
    Metadata       any
    Actor          identityaccess.Actor
}
```

The coordinator's `ProcessNextQueued` pattern is extended to a `ProcessNextScheduledRun` that does not require an `InboundRequestID`. The provider receives a scheduled-run-specific instruction set and prompt, and the output goes to an operator alert surface rather than a request review surface.

This is a Phase 4+ capability — the system needs reliable request processing before proactive intelligence adds value. But the architectural path is clear and does not require structural changes to the existing model.

---

### Comparison to common alternatives

| Architecture | Business suitability | Assessment |
|---|---|---|
| Synchronous inline AI calls | Poor | Timeout risk, no audit trail, no retry, no human review gate |
| LangChain/LangGraph in-memory | Moderate | Flexible for prototyping; requires durable state bolted on for production |
| OpenAI Assistants API with threads | Moderate | Provider-owned state, no custom governance, vendor lock-in |
| CrewAI multi-agent | Moderate | Role-based agents, but typically in-memory and not durable |
| Temporal/Conductor workflow engine | High | Excellent for long-running durable workflows; significant operational complexity |
| **This application's architecture** | **High** | Durable, auditable, governance-gated, provider-agnostic, human-approval model |

The only pattern that outperforms this application's architecture for business AI is a **workflow engine** (Temporal, Conductor) as the coordinator, with AI models as task executors within workflow steps. That pattern adds durable execution with automatic retry, workflow versioning, and sophisticated timeout handling — at the cost of significant infrastructure and operational complexity. For this application at this stage, the current architecture is the better tradeoff.

### Architectural verdict

The architecture is **correct and opinionated in the right ways**. The three foundational decisions — persist-first async queue, human-in-the-loop approval gate, database-as-truth — are exactly right for business AI and are the decisions most teams get wrong.

The two architectural gaps (event-driven triggering, proactive runs) are not present-day blockers. They become important as request volume grows and as the use case evolves from reactive triage to active operational intelligence. Both are addable without structural changes.

The application correctly restrains AI autonomy. The hardest architectural mistake to undo in business AI is having built autonomy where approval gates should have been. This application built the approval gates correctly from the beginning.

---

## What Is Working Well

Before addressing the code-level gaps, the following implementation decisions are explicitly correct and must be preserved during any refactor.

### Persisted run model is ACID at every step

Every coordinator action — starting a run, appending a step, creating an artifact, recording a delegation — is a separate, committed database transaction with an audit event written inside the same transaction boundary. This means the audit log is always consistent with domain state. A mid-run failure leaves a fully reconstructable record.

```go
// service.go — every domain action follows this shape
func (s *Service) StartRun(ctx context.Context, input StartRunInput) (Run, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil { ... }
    run, err := startRunTx(ctx, tx, input)
    if err := audit.WriteTx(ctx, tx, audit.Event{...}); err != nil { ... }
    if err := tx.Commit(); err != nil { ... }
    return run, nil
}
```

### Provider abstraction is correctly isolated

The coordinator knows nothing about OpenAI or any other API. It talks through a single interface:

```go
type CoordinatorProvider interface {
    ExecuteInboundRequest(ctx context.Context, input CoordinatorProviderInput) (CoordinatorProviderOutput, error)
}
```

Swapping to Claude, Gemini, or a local model is a one-struct change with no coordinator changes required.

### Tool policy governance is real governance

The `allow / approval_required / deny` policy resolver backed by a database table gives operators control over which tools a coordinator can call, without a code change or redeploy. Mutating tools default to `approval_required`. The policy source (`default` vs. `database`) is included in the execution record.

### Output validation is layered and rare

The coordinator validates structural completeness, semantic grounding (output mentions concrete request content), and attempts a self-repair API call before failing. This three-layer guard — structural → semantic → repair → fail — is a design pattern that most production AI applications don't have.

### Tool budget exhaustion intent is correct; implementation has a counting bug

The intent — once all read tools have been used, remove the tool list to force a final output — is the right design:

```go
if coordinatorReadToolBudgetExhausted(toolDefs, toolExecutions) {
    params.Tools = nil
}
```

However, `coordinatorReadToolBudgetExhausted` counts **total tool calls** against `len(toolDefs)`, not unique tools called. If the LLM calls the same tool twice (which the instructions discourage but do not prevent at the API level), the budget is marked exhausted after 3 total calls even if only one of the three tools was ever invoked. See Improvement G for the fix.

### Stateless tool loop continuation is correct

The `buildStatelessContinuationInput` function reconstructs prior response output for each subsequent API call (`Store: false`). The application owns the conversation state, not the provider. This is audit-friendly and provider-agnostic.

---

## Gap 1 — Specialist delegation is a stub, not an execution path

**Severity: High**  
**File:** `internal/ai/coordinator.go` — `createDelegatedSpecialistRun` (line 407)

### Problem

When the coordinator's LLM decides to delegate to a specialist (e.g., `inbound_request.operations_triage`), the coordinator calls `createDelegatedSpecialistRun`, which:

1. Creates a child run record with `RunRoleSpecialist`
2. Records a delegation record
3. Appends a step to the child run
4. **Returns the same `providerOutput` from the coordinator — no separate AI execution occurs**

The specialist run is then immediately marked as completed using the coordinator's own artifact and recommendation. The `capability_code` in the delegation record is validated against an allowlist, but no code actually routes to a different prompt, model, instruction set, or tool suite.

```go
// coordinator.go line 286 — after createDelegatedSpecialistRun
artifact, err := c.aiService.CreateArtifact(ctx, CreateArtifactInput{
    RunID:   artifactRun.ID,   // ← specialist run ID
    Payload: map[string]any{
        "body": providerOutput.ArtifactBody, // ← coordinator's output, not specialist's
    },
})
```

The system gives the operator the impression that a specialist did deeper domain analysis (operations triage, approval triage). It did not. The coordinator's general-purpose brief is recorded under the specialist's run identity.

### What a correct specialist execution looks like

A specialist should receive the same request context plus the coordinator's brief as framing context, execute with a domain-specific instruction set and (optionally) a different tool suite, and produce its own structured artifact.

### Fix — Implement a SpecialistProvider interface and execution path

**Step 1: Define a `SpecialistProvider` interface**

```go
// internal/ai/coordinator.go

// SpecialistProvider executes a domain-specific specialist review
// for a delegated inbound request.
type SpecialistProvider interface {
    ExecuteSpecialistReview(ctx context.Context, input SpecialistProviderInput) (CoordinatorProviderOutput, error)
}

type SpecialistProviderInput struct {
    CapabilityCode      string
    Actor               identityaccess.Actor
    RequestReference    string
    Channel             string
    OriginType          string
    Metadata            json.RawMessage
    Messages            []CoordinatorMessage
    Attachments         []CoordinatorAttachment
    DerivedTexts        []CoordinatorDerivedText
    CoordinatorSummary  string   // coordinator's brief — used as framing context
    CoordinatorPriority string
    CoordinatorRationale []string
    DelegationReason    string
}
```

**Step 2: Add a specialist provider registry to the Coordinator**

```go
// internal/ai/coordinator.go

type Coordinator struct {
    intakeService      *intake.Service
    aiService          *Service
    provider           CoordinatorProvider
    specialistRegistry map[string]SpecialistProvider // keyed by capability_code
    capabilityCode     string
    requestLoaderDB    *sql.DB
}

func NewCoordinator(db *sql.DB, provider CoordinatorProvider) *Coordinator {
    return &Coordinator{
        intakeService:      intake.NewService(db),
        aiService:          NewService(db),
        provider:           provider,
        specialistRegistry: make(map[string]SpecialistProvider),
        capabilityCode:     DefaultCoordinatorCapabilityCode,
        requestLoaderDB:    db,
    }
}

// RegisterSpecialist registers a provider for a specific capability code.
// Call this after NewCoordinator before first use.
func (c *Coordinator) RegisterSpecialist(capabilityCode string, provider SpecialistProvider) {
    if c.specialistRegistry == nil {
        c.specialistRegistry = make(map[string]SpecialistProvider)
    }
    c.specialistRegistry[strings.TrimSpace(capabilityCode)] = provider
}
```

**Step 3: Execute the specialist provider in `createDelegatedSpecialistRun`**

```go
// Replace the current createDelegatedSpecialistRun with:

func (c *Coordinator) createDelegatedSpecialistRun(
    ctx context.Context,
    request intake.InboundRequest,
    parentRun Run,
    parentStep RunStep,
    requestContext CoordinatorProviderInput,
    coordinatorOutput CoordinatorProviderOutput,
    actor identityaccess.Actor,
) (Run, RunStep, Delegation, CoordinatorProviderOutput, error) {

    specialist := coordinatorOutput.SpecialistDelegation
    if specialist == nil {
        return Run{}, RunStep{}, Delegation{}, CoordinatorProviderOutput{}, ErrInvalidCoordinatorOutput
    }

    // Start child run
    childRun, err := c.aiService.StartRun(ctx, StartRunInput{
        AgentRole:        RunRoleSpecialist,
        CapabilityCode:   specialist.CapabilityCode,
        InboundRequestID: request.ID,
        ParentRunID:      parentRun.ID,
        RequestText:      fmt.Sprintf("Specialist review (%s) for %s", specialist.CapabilityCode, request.RequestReference),
        Metadata: map[string]any{
            "request_reference": request.RequestReference,
            "parent_run_id":     parentRun.ID,
            "delegation_reason": specialist.Reason,
        },
        Actor: actor,
    })
    if err != nil {
        return Run{}, RunStep{}, Delegation{}, CoordinatorProviderOutput{}, err
    }

    // Record delegation link
    delegation, err := c.aiService.RecordDelegation(ctx, RecordDelegationInput{
        ParentRunID:       parentRun.ID,
        ChildRunID:        childRun.ID,
        RequestedByStepID: parentStep.ID,
        CapabilityCode:    specialist.CapabilityCode,
        Reason:            specialist.Reason,
        Actor:             actor,
    })
    if err != nil {
        _, _ = c.aiService.CompleteRun(ctx, CompleteRunInput{RunID: childRun.ID, Status: RunStatusFailed, Summary: "failed to record delegation", Actor: actor})
        return Run{}, RunStep{}, Delegation{}, CoordinatorProviderOutput{}, err
    }

    // Execute specialist provider if one is registered
    specialistOutput := coordinatorOutput // fallback: use coordinator output
    specialistProvider, hasSpecialist := c.specialistRegistry[strings.TrimSpace(specialist.CapabilityCode)]
    if hasSpecialist {
        specialistInput := SpecialistProviderInput{
            CapabilityCode:       specialist.CapabilityCode,
            Actor:                actor,
            RequestReference:     requestContext.RequestReference,
            Channel:              requestContext.Channel,
            OriginType:           requestContext.OriginType,
            Metadata:             requestContext.Metadata,
            Messages:             requestContext.Messages,
            Attachments:          requestContext.Attachments,
            DerivedTexts:         requestContext.DerivedTexts,
            CoordinatorSummary:   coordinatorOutput.Summary,
            CoordinatorPriority:  coordinatorOutput.Priority,
            CoordinatorRationale: coordinatorOutput.Rationale,
            DelegationReason:     specialist.Reason,
        }
        var execErr error
        specialistOutput, execErr = specialistProvider.ExecuteSpecialistReview(ctx, specialistInput)
        if execErr != nil {
            // Log the failure but do not abandon — fall back to coordinator output
            specialistOutput = coordinatorOutput
            specialistOutput.ProviderResponseID = "specialist_fallback"
        }
    }

    // Append step to child run recording the specialist execution
    step, err := c.aiService.AppendStep(ctx, AppendStepInput{
        RunID:     childRun.ID,
        StepType:  specialistStepTypeDelegatedReview,
        StepTitle: fmt.Sprintf("Specialist review: %s", specialist.CapabilityCode),
        Status:    StepStatusCompleted,
        InputPayload: map[string]any{
            "request_reference":   request.RequestReference,
            "parent_run_id":       parentRun.ID,
            "delegation_id":       delegation.ID,
            "delegation_reason":   specialist.Reason,
            "has_specialist_provider": hasSpecialist,
        },
        OutputPayload: map[string]any{
            "provider":             specialistOutput.ProviderName,
            "provider_response_id": specialistOutput.ProviderResponseID,
            "model":                specialistOutput.Model,
            "priority":             specialistOutput.Priority,
            "summary":              specialistOutput.Summary,
            "tool_loop_iterations": specialistOutput.ToolLoopIterations,
            "used_fallback":        !hasSpecialist,
        },
        Actor: actor,
    })
    if err != nil {
        _, _ = c.aiService.CompleteRun(ctx, CompleteRunInput{RunID: childRun.ID, Status: RunStatusFailed, Summary: "failed to record specialist step", Actor: actor})
        return Run{}, RunStep{}, Delegation{}, CoordinatorProviderOutput{}, err
    }

    return childRun, step, delegation, specialistOutput, nil
}
```

**Step 4: Update the coordinator's `ProcessNextQueued` to use the returned specialist output**

The coordinator currently uses `providerOutput` for both the coordinator artifact and the specialist artifact. After this change, the specialist artifact should use `specialistOutput`.

**Step 5: Implement the first real specialist**

The `inbound_request.operations_triage` specialist should use a domain-specific instruction set focused on:
- Routing the request to the correct workflow queue
- Identifying the right approval path
- Providing operations-specific `next_actions` (e.g., "Open purchase order workflow" vs. "Escalate to finance team")

The OpenAI provider can implement `SpecialistProvider` for each capability code:

```go
// internal/ai/openai_provider.go

func (p *OpenAIProvider) ExecuteSpecialistReview(ctx context.Context, input SpecialistProviderInput) (CoordinatorProviderOutput, error) {
    switch strings.TrimSpace(input.CapabilityCode) {
    case "inbound_request.operations_triage":
        return p.executeOperationsTriageReview(ctx, input)
    case "inbound_request.approval_triage":
        return p.executeApprovalTriageReview(ctx, input)
    default:
        return CoordinatorProviderOutput{}, fmt.Errorf("unsupported specialist capability: %s", input.CapabilityCode)
    }
}
```

---

## Gap 2 — `ResolveToolPolicy` uses a write transaction on every tool call

**Severity: Medium**  
**File:** `internal/ai/service.go` — `ResolveToolPolicy` (line 427)

### Problem

Policy resolution is a read-only operation — it looks up a policy row and returns it. However, `ResolveToolPolicy` opens a read-write transaction and calls `authorizeWriteTx`, which in turn calls `identityaccess.AuthorizeTx` with `SELECT ... FOR UPDATE` on the session and membership rows:

```go
// service.go line 427
func (s *Service) ResolveToolPolicy(ctx context.Context, input ResolveToolPolicyInput) (ResolvedToolPolicy, error) {
    tx, err := s.db.BeginTx(ctx, nil)     // ← read-write transaction — wrong
    if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil { ... } // ← FOR UPDATE — wrong
    policy, found, err := resolveToolPolicyTx(ctx, tx, ...)
    if err := tx.Commit(); err != nil { ... }
    ...
}
```

This is called inside the tool execution loop — once per tool call per coordinator iteration. A coordinator run with 3 tool executions acquires 3 unnecessary write locks on the session row. Under concurrent load (multiple coordinator workers), this creates lock contention on the sessions table.

### Fix

Add `authorizeReadTx` to `service.go` and use it in `ResolveToolPolicy`:

```go
// internal/ai/service.go

func authorizeReadTx(ctx context.Context, tx *sql.Tx, actor identityaccess.Actor) error {
    return identityaccess.AuthorizeReadOnlyTx(ctx, tx, actor, identityaccess.RoleAdmin, identityaccess.RoleOperator)
}

func (s *Service) ResolveToolPolicy(ctx context.Context, input ResolveToolPolicyInput) (ResolvedToolPolicy, error) {
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}) // ← read-only
    if err != nil {
        return ResolvedToolPolicy{}, fmt.Errorf("begin resolve tool policy: %w", err)
    }

    if err := authorizeReadTx(ctx, tx, input.Actor); err != nil { // ← no FOR UPDATE
        _ = tx.Rollback()
        return ResolvedToolPolicy{}, err
    }

    policy, found, err := resolveToolPolicyTx(ctx, tx, input.Actor.OrgID, input)
    if err != nil {
        _ = tx.Rollback()
        return ResolvedToolPolicy{}, err
    }
    if err := tx.Commit(); err != nil {
        return ResolvedToolPolicy{}, fmt.Errorf("commit resolve tool policy: %w", err)
    }

    if found {
        return policy, nil
    }
    return ResolvedToolPolicy{
        ToolName: normalizeRequired(input.ToolName),
        Policy:   normalizeDefaultToolPolicy(input.DefaultPolicy),
        Source:   "default",
    }, nil
}
```

This requires `identityaccess.AuthorizeReadOnlyTx` to exist. If it does not yet exist, add it alongside `AuthorizeTx` in `internal/identityaccess/auth.go` — same query, same logic, without `FOR UPDATE`. See also the improvement plan in `code_review_and_improvement_plan.md` (P2-6), which prescribes the same fix for the reporting layer.

---

## Gap 3 — Context loading uses inline SQL, bypassing the service layer

**Severity: Medium**  
**File:** `internal/ai/coordinator.go` — `loadRequestRow`, `loadRequestMessages`, `loadRequestAttachments`, `loadRequestDerivedTexts` (lines 537–699)

### Problem

The coordinator loads all request context for the provider using four inline SQL queries executed directly against `c.requestLoaderDB`:

```go
func (c *Coordinator) loadRequestMessages(ctx context.Context, orgID, requestID string) ([]CoordinatorMessage, error) {
    const query = `SELECT message_role, text_content FROM ai.inbound_request_messages ...`
    rows, err := c.requestLoaderDB.QueryContext(ctx, query, ...)
    // ...
}
```

Problems:

1. **Schema drift risk**: Changes to `ai.inbound_request_messages`, `attachments.request_message_links`, or `attachments.derived_texts` must be reflected in two places — the coordinator's inline SQL and any service-layer SQL that queries the same tables.
2. **Untestable in isolation**: The inline queries cannot be tested without a real database. They are woven into the coordinator's orchestration flow.
3. **The `loadRequestDerivedTexts` query is particularly complex** (line 644–673): it uses a correlated `EXISTS` subquery to join across two different link-path patterns. This is the kind of query that belongs in a service or repository function with its own test coverage.
4. **The coordinator bypasses `intake.Service`**: `loadRequestRow` re-scans all 22 columns of `ai.inbound_requests` using a custom `scanInboundRequestContext` function (lines 901–936 of coordinator.go), duplicating what `intake.Service` already does.

### Fix

**Step 1: Introduce a `RequestContextLoader` interface**

```go
// internal/ai/coordinator.go

// RequestContextLoader loads the full request context needed for coordinator execution.
// This interface allows the coordinator to be tested without a real database.
type RequestContextLoader interface {
    LoadRequestContext(ctx context.Context, actor identityaccess.Actor, requestID string) (CoordinatorProviderInput, error)
}
```

**Step 2: Implement a `DatabaseRequestContextLoader`**

Move the four inline SQL methods into a standalone struct that implements this interface. This struct owns its own schema dependencies and can be tested independently.

```go
// internal/ai/request_context_loader.go  (new file)

type DatabaseRequestContextLoader struct {
    db             *sql.DB
    capabilityCode string
}

func NewDatabaseRequestContextLoader(db *sql.DB, capabilityCode string) *DatabaseRequestContextLoader {
    return &DatabaseRequestContextLoader{db: db, capabilityCode: capabilityCode}
}

func (l *DatabaseRequestContextLoader) LoadRequestContext(ctx context.Context, actor identityaccess.Actor, requestID string) (CoordinatorProviderInput, error) {
    // Move the four loadRequest* methods here
    // Keep the existing SQL — just relocate it to a testable seam
}
```

**Step 3: Update the Coordinator to use the interface**

```go
type Coordinator struct {
    intakeService      *intake.Service
    aiService          *Service
    provider           CoordinatorProvider
    specialistRegistry map[string]SpecialistProvider
    contextLoader      RequestContextLoader  // ← replaces requestLoaderDB
    capabilityCode     string
}
```

The `NewCoordinator` factory wires in a `DatabaseRequestContextLoader`:

```go
func NewCoordinator(db *sql.DB, provider CoordinatorProvider) *Coordinator {
    capabilityCode := DefaultCoordinatorCapabilityCode
    return &Coordinator{
        intakeService:      intake.NewService(db),
        aiService:          NewService(db),
        provider:           provider,
        specialistRegistry: make(map[string]SpecialistProvider),
        contextLoader:      NewDatabaseRequestContextLoader(db, capabilityCode),
        capabilityCode:     capabilityCode,
    }
}
```

**Step 4: Remove the four inline `loadRequest*` methods from coordinator.go and `scanInboundRequestContext`**

After moving these to `request_context_loader.go`, the coordinator file shrinks significantly and becomes fully testable with a mock `RequestContextLoader`.

---

## Gap 4 — Single 45-second timeout covers all sequential tool loop iterations

**Severity: Medium**  
**File:** `internal/ai/openai_provider.go` — `ExecuteInboundRequest` (line 69)

### Problem

A single `context.WithTimeout` is applied once for the entire coordinator execution, which may include up to four sequential API calls:

```go
const openAIProviderTimeout = 45 * time.Second

func (p *OpenAIProvider) ExecuteInboundRequest(ctx context.Context, input CoordinatorProviderInput) (CoordinatorProviderOutput, error) {
    requestCtx, cancel := context.WithTimeout(ctx, openAIProviderTimeout)
    defer cancel()

    for iteration := 1; iteration <= p.normalizedMaxToolIterations(); iteration++ {
        resp, err := p.responsesAPI.New(requestCtx, params) // ← same context each time
```

If the first API call takes 40 seconds (possible under GPT-4 load), the remaining 5 seconds are shared across all remaining iterations. The second call will immediately time out, even if the server would have responded in 3 seconds.

Additionally, the repair pass in `repairRequestCenteredOutput` uses the same `requestCtx`, which may already be nearly exhausted when it is called.

### Fix

**Option A — Per-iteration timeout (recommended)**

Apply a fresh timeout for each API call. Use a per-call constant that is shorter than the overall budget:

```go
const (
    openAIPerCallTimeout       = 30 * time.Second
    openAIRepairCallTimeout    = 20 * time.Second
    openAIMaxCoordinatorToolLoops = 4
)

func (p *OpenAIProvider) ExecuteInboundRequest(ctx context.Context, input CoordinatorProviderInput) (CoordinatorProviderOutput, error) {
    // No outer timeout — caller's context is the cancellation boundary
    // Each API call gets its own fresh timeout

    for iteration := 1; iteration <= p.normalizedMaxToolIterations(); iteration++ {
        callCtx, callCancel := context.WithTimeout(ctx, openAIPerCallTimeout)
        resp, err := p.responsesAPI.New(callCtx, params)
        callCancel() // always release immediately after the call
        if err != nil {
            // ...
        }
        // ...
    }
}
```

**Option B — Keep outer timeout but make it large enough**

Increase the outer timeout to account for worst-case sequential calls:

```go
// 45s per call × 4 iterations + 20s buffer = ~200 seconds for pathological cases
// A more realistic worst-case is 30s per call × 4 = 120s
const openAIProviderTimeout = 120 * time.Second
```

Option A is preferable because it gives each call the full budget, prevents cascading timeouts, and allows the caller to impose its own outer deadline (e.g., from an HTTP context or a queue processor's job timeout). Option B is simpler but still vulnerable to cascading exhaustion.

**Update the repair call to use its own context:**

```go
func (p *OpenAIProvider) repairRequestCenteredOutput(ctx context.Context, ...) (...) {
    repairCtx, cancel := context.WithTimeout(ctx, openAIRepairCallTimeout)
    defer cancel()
    resp, err := p.responsesAPI.New(repairCtx, params)
    // ...
}
```

---

## Gap 5 — Tool execution errors are silently swallowed; no degraded-mode signaling

**Severity: Low-Medium**  
**File:** `internal/ai/openai_provider.go` — `executeCoordinatorTool` (line 627)

### Problem

When a tool is blocked by policy, not found, or fails during execution, the coordinator receives a structured JSON error message as the tool output, and the LLM continues:

```go
if resolvedPolicy.Policy != PolicyAllow {
    return marshalToolOutput(map[string]any{
        "status":  "blocked",
        "tool":    toolDef.ToolName,
        "message": "tool execution blocked by policy",
    }), execution
}

output, preview, err := toolDef.Execute(ctx, input)
if err != nil {
    return marshalToolOutput(map[string]any{
        "status":  "error",
        "tool":    toolDef.ToolName,
        "message": "tool execution failed",
    }), execution
}
```

The coordinator instructions say "If a tool call is denied or unavailable, continue without it and produce the best safe review possible." This is a reasonable production stance. However, the current implementation has no way to signal to the coordinator result that tool failures occurred. A brief produced after all tools were denied looks identical in the coordinator output to a brief produced with full data access.

The operator reviewing the recommendation has no visibility into whether the AI had complete information when it produced it.

### Fix

**Step 1: Track degraded mode in the provider output**

```go
// internal/ai/coordinator.go
type CoordinatorProviderOutput struct {
    // ... existing fields ...
    ToolLoopIterations   int
    ToolExecutions       []CoordinatorToolExecution
    SpecialistDelegation *CoordinatorSpecialistDelegation
    DegradedMode         bool   // ← new: true if any tool was blocked or failed
    DegradedReasons      []string // ← new: human-readable reasons
}
```

**Step 2: Populate `DegradedMode` in the provider**

```go
// internal/ai/openai_provider.go — after the tool loop completes
parseResult, err := p.parseCoordinatorResponse(input, resp, totalUsage, iteration, toolExecutions)
if err != nil {
    return CoordinatorProviderOutput{}, err
}

// Check if any tool executions were blocked or failed
output := parseResult.output
for _, exec := range toolExecutions {
    if exec.Outcome == "blocked_by_policy" || exec.Outcome == "execution_failed" || exec.Outcome == "policy_lookup_failed" {
        output.DegradedMode = true
        output.DegradedReasons = append(output.DegradedReasons, fmt.Sprintf("%s: %s", exec.ToolName, exec.Outcome))
    }
}
return output, nil
```

**Step 3: Record `DegradedMode` in the coordinator step and artifact**

When `providerOutput.DegradedMode` is true, include it in the artifact payload and the step output payload so operators can see in the review UI that the AI ran with limited context.

**Step 4 (optional): Surface degraded mode in the Svelte review UI**

The proposal review component can show a warning badge or footnote when `degraded_mode: true` is present in the artifact payload, for example: "⚠️ This brief was produced with limited tool access."

---

## Gap 6 — Prompt and instructions are hardcoded strings in Go source

**Severity: Low**  
**File:** `internal/ai/openai_provider.go` — `coordinatorInstructions` (line 214), `buildProviderPrompt` (line 891)

### Problem

The coordinator system instructions and the user prompt format are hardcoded Go string literals:

```go
func (p *OpenAIProvider) coordinatorInstructions() string {
    return `You are the workflow_app inbound-request coordinator.
Review the persisted request context and produce a structured operator-review brief.
...`
}
```

This has two practical consequences:

1. **Prompt iteration requires a code change and redeploy.** Testing whether a different instruction phrasing improves output quality requires a full development cycle.
2. **No version tracking.** It is not possible to know which instruction version produced a given run's artifact, making quality regression analysis difficult.

### Fix — Externalize instructions into a configuration layer

**Step 1: Define a `PromptConfig` struct**

```go
// internal/ai/provider_config.go

type PromptConfig struct {
    CoordinatorInstructions      string
    OperationsTriageInstructions string
    ApprovalTriageInstructions   string
    MaxKeywordsInPrompt          int
}

func DefaultPromptConfig() PromptConfig {
    return PromptConfig{
        CoordinatorInstructions:      defaultCoordinatorInstructions,
        OperationsTriageInstructions: defaultOperationsTriageInstructions,
        ApprovalTriageInstructions:   defaultApprovalTriageInstructions,
        MaxKeywordsInPrompt:          8,
    }
}

// Keep the current hardcoded strings as package-level constants
// so defaults do not require a file at runtime
const defaultCoordinatorInstructions = `You are the workflow_app inbound-request coordinator...`
```

**Step 2: Pass `PromptConfig` into the provider**

```go
type OpenAIProvider struct {
    responsesAPI      openAIResponsesAPI
    aiService         *Service
    reportingService  *reporting.Service
    model             string
    maxToolIterations int
    promptConfig      PromptConfig  // ← new
}

func NewOpenAIProvider(db *sql.DB, config ProviderConfig) (*OpenAIProvider, error) {
    return &OpenAIProvider{
        // ...
        promptConfig: DefaultPromptConfig(),
    }, nil
}
```

**Step 3: Record the instruction version in the run metadata**

When a run starts, include a hash or short identifier of the instruction text in the run metadata:

```go
"coordinator_instructions_hash": shortHash(p.promptConfig.CoordinatorInstructions),
```

This allows post-hoc analysis of which instruction version produced which artifacts, without storing the full text in every run record.

**Step 4 (optional): Support loading prompts from a file path or environment variable**

```go
func PromptConfigFromEnv() PromptConfig {
    cfg := DefaultPromptConfig()
    if path := os.Getenv("COORDINATOR_INSTRUCTIONS_FILE"); path != "" {
        if content, err := os.ReadFile(path); err == nil {
            cfg.CoordinatorInstructions = string(content)
        }
    }
    return cfg
}
```

This allows prompt development and iteration without recompiling, while falling back to compiled defaults if no file is present.

---

## Additional Improvements

### A — The `openai_provider.go` file needs decomposition (1088 lines)

The provider file currently contains:
- The tool loop execution engine
- All three tool definitions and their execution functions
- The prompt builder
- The response format and schema
- The response parser
- The self-repair logic
- Response validation

These are independent concerns. A natural decomposition:

| Target file | Contents |
|---|---|
| `openai_provider.go` | `OpenAIProvider` struct, `ExecuteInboundRequest`, tool loop |
| `openai_tools.go` | `coordinatorToolDefinitions`, tool execution functions |
| `openai_prompt.go` | `buildProviderPrompt`, `buildRequestCenteredRepairPrompt`, `coordinatorInstructions` |
| `openai_schema.go` | `coordinatorResponseFormat`, `coordinatorResponseSchema` |
| `openai_parse.go` | `parseCoordinatorResponse`, `repairRequestCenteredOutput`, `validateOpenAIResponse` |

This does not change any behavior — it is a structural refactor. It makes each concern reviewable independently and allows specialist execution functions to be added to `openai_tools.go` without growing the provider file further.

### B — `providerName` uses a type switch, should be a method

```go
// current — coordinator.go line 877
func providerName(provider CoordinatorProvider) string {
    switch provider.(type) {
    case *OpenAIProvider:
        return "openai"
    default:
        return "custom"
    }
}
```

This pattern breaks when a new provider is added — the coordinator must know about every concrete provider type. Add a `Name() string` method to `CoordinatorProvider`:

```go
type CoordinatorProvider interface {
    Name() string
    ExecuteInboundRequest(ctx context.Context, input CoordinatorProviderInput) (CoordinatorProviderOutput, error)
}

func (p *OpenAIProvider) Name() string { return "openai" }
```

Remove the `providerName` function entirely.

### C — `isAllowedSpecialistCapability` and the response schema enum are both hardcoded and out of sync with each other

```go
// coordinator.go line 859 — validation function
func isAllowedSpecialistCapability(capabilityCode string) bool {
    switch strings.TrimSpace(capabilityCode) {
    case "inbound_request.operations_triage", "inbound_request.approval_triage":
        return true
    default:
        return false
    }
}

// openai_provider.go lines 859–862 — JSON schema enum (separate, also hardcoded)
"enum": []string{
    "inbound_request.operations_triage",
    "inbound_request.approval_triage",
},
```

There are **two separate hardcoded allowlists** for specialist capabilities: one in the coordinator's validation function and one baked into the JSON response schema sent to OpenAI. Adding a new specialist capability currently requires updating both places independently — and there is no compile-time check that they agree. If one is updated and the other is not, the LLM can produce a capability code that the schema accepted but the validation function rejects (or vice versa).

Once the specialist registry exists (Gap 1 fix), both should derive from the same source.

**Fix — two coordinated changes:**

**Part 1:** Make `isAllowedSpecialistCapability` registry-driven on the `Coordinator`:

```go
func (c *Coordinator) isAllowedSpecialistCapability(capabilityCode string) bool {
    _, ok := c.specialistRegistry[strings.TrimSpace(capabilityCode)]
    return ok
}
```

**Part 2:** Make `coordinatorResponseSchema()` accept the registered capability codes and build the enum dynamically. Pass the registry keys into the schema builder at request-build time:

```go
// openai_provider.go
func coordinatorResponseSchema(allowedCapabilities []string) map[string]any {
    return map[string]any{
        // ... existing fields ...
        "specialist_delegation": map[string]any{
            "anyOf": []any{
                map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "capability_code": map[string]any{
                            "type": "string",
                            "enum": allowedCapabilities, // ← from registry, not hardcoded
                        },
                    },
                },
                map[string]any{"type": "null"},
            },
        },
    }
}
```

The `OpenAIProvider`'s `newCoordinatorResponseParams` should accept the allowed capability codes as a parameter (passed by the coordinator from its registry keys) rather than coupling the provider to the coordinator's internal registry. This keeps the provider stateless with respect to the registry:

```go
// In the coordinator — pass registry keys when building request params:
allowedCaps := make([]string, 0, len(c.specialistRegistry))
for code := range c.specialistRegistry {
    allowedCaps = append(allowedCaps, code)
}
params := p.newCoordinatorResponseParams(input, toolDefs, allowedCaps)
```

Adding a specialist to the registry becomes the single point of change — the schema and validation both derive from it.

### D — The coordinator keyword stopword list needs a maintenance strategy

```go
// coordinator.go line 850
func isCoordinatorStopword(token string) bool {
    switch token {
    case "about", "after", "attached", "because", "browser", "channel", ...:
        return true
    }
}
```

This is a hardcoded list of 30+ words used to filter out generic tokens before building the evidence-specificity check. It is the kind of list that grows organically and becomes a maintenance burden. Document it with a clear comment explaining its purpose and acceptance criteria:

```go
// isCoordinatorStopword filters tokens that are too generic to serve as evidence anchors
// for the request-centered output validation. Add a token here only if it appears in
// nearly every inbound request regardless of content (channel names, lifecycle terms,
// workflow vocabulary) and would cause false positives in the evidence-specificity check.
func isCoordinatorStopword(token string) bool {
```

Add a unit test that validates the stopword list does not accidentally suppress legitimate business terms (e.g., "invoice" should never be a stopword).

### E — No retry logic for transient OpenAI API failures

Currently, any API error immediately aborts the coordinator run and marks the inbound request as failed:

```go
resp, err := p.responsesAPI.New(requestCtx, params)
if err != nil {
    var apiErr *openai.Error
    if errors.As(err, &apiErr) {
        return CoordinatorProviderOutput{}, fmt.Errorf("openai responses api error (status %d): %w", apiErr.StatusCode, err)
    }
    return CoordinatorProviderOutput{}, fmt.Errorf("openai responses api request failed: %w", err)
}
```

Transient errors (rate limits, server errors, network blips) cause permanent request failures visible to operators. The OpenAI SDK returns structured errors with HTTP status codes. 429 (rate limit) and 5xx (server errors) are retryable; 400 and 401 are not.

**Minimal fix — Add one retry for retryable errors:**

```go
func isRetryableOpenAIError(err error) bool {
    var apiErr *openai.Error
    if errors.As(err, &apiErr) {
        return apiErr.StatusCode == 429 || apiErr.StatusCode >= 500
    }
    return false
}

// In the tool loop, add one retry with backoff for retryable errors:
resp, err := p.responsesAPI.New(callCtx, params)
if err != nil && isRetryableOpenAIError(err) {
    time.Sleep(2 * time.Second)
    resp, err = p.responsesAPI.New(callCtx, params)
}
if err != nil {
    return CoordinatorProviderOutput{}, fmt.Errorf("openai responses api request failed: %w", err)
}
```

For a production system, this should use an exponential backoff library. A single retry with a 2-second sleep is a pragmatic starting point.

### F — `MaxOutputTokens: 900` is too low for a rich structured JSON brief

**File:** `internal/ai/openai_provider.go` — lines 165 and 1008

Both the primary coordinator call and the repair call cap output at 900 tokens:

```go
// line 165 — primary call
MaxOutputTokens: openai.Int(900),

// line 1008 — repair call
MaxOutputTokens: openai.Int(900),
```

The coordinator response schema requires all of these fields to be populated in a single JSON object: `summary`, `priority`, `artifact_title`, `artifact_body`, `rationale` (array), `next_actions` (array), and `specialist_delegation`. A thorough `artifact_body` alone — with background, risk assessment, and operator guidance — can easily consume 400–600 tokens. Combined with the other required fields, 900 tokens risks truncation on complex requests, which causes `resp.Status == responses.ResponseStatusIncomplete` and a hard error:

```go
case responses.ResponseStatusIncomplete:
    return fmt.Errorf("openai response incomplete: %s", reason)
```

The repair call at line 1008 uses the same 900-token cap. A repair call needs enough room to produce a revised brief — using the same limit that caused the issue in the first place is counterproductive.

**Fix:**

```go
const (
    openAICoordinatorMaxOutputTokens = 2000  // enough for a thorough brief with all required fields
    openAIRepairMaxOutputTokens      = 1500  // repair call: slightly smaller since the structural shape is already known
)

// primary call (line 165)
MaxOutputTokens: openai.Int(openAICoordinatorMaxOutputTokens),

// repair call (line 1008)
MaxOutputTokens: openai.Int(openAIRepairMaxOutputTokens),
```

These constants should be added alongside the existing `openAIProviderTimeout` and `openAIMaxCoordinatorToolLoops` constants and made configurable via `ProviderConfig` for future tuning without a code change.

### G — `coordinatorReadToolBudgetExhausted` counts total calls, not unique tools

**File:** `internal/ai/openai_provider.go` — `coordinatorReadToolBudgetExhausted` (line 293)

The budget exhaustion function counts how many executed tools exist in `toolDefs`, regardless of which tool was called:

```go
func coordinatorReadToolBudgetExhausted(toolDefs map[string]coordinatorToolDefinition, executions []CoordinatorToolExecution) bool {
    readToolCalls := 0
    for _, execution := range executions {
        if _, ok := toolDefs[execution.ToolName]; !ok {
            continue
        }
        readToolCalls++  // ← counts every call, not unique tools
    }
    return readToolCalls >= len(toolDefs)  // ← triggers when total calls == number of tools
}
```

With 3 tools registered, if the LLM calls `reporting_get_current_inbound_request_detail` three times, `readToolCalls` reaches 3 and the budget is exhausted — even though the other two tools (`reporting_list_inbound_request_status_summary` and `reporting_list_current_processed_proposals`) were never called. The LLM's tool list is then removed for subsequent iterations, preventing it from calling potentially useful tools it hasn't used yet.

The coordinator instructions do say "Do not call the same read tool repeatedly once you already have its result" — but API-level enforcement is not present and LLMs occasionally repeat tool calls.

**Fix — Track unique tools called instead of total calls:**

```go
func coordinatorReadToolBudgetExhausted(toolDefs map[string]coordinatorToolDefinition, executions []CoordinatorToolExecution) bool {
    if len(toolDefs) == 0 || len(executions) == 0 {
        return false
    }

    calledTools := make(map[string]struct{}, len(toolDefs))
    for _, execution := range executions {
        name := strings.TrimSpace(execution.ToolName)
        if name == "" {
            continue
        }
        if _, ok := toolDefs[name]; !ok {
            continue
        }
        calledTools[name] = struct{}{}  // ← track unique tools, not total calls
    }

    return len(calledTools) >= len(toolDefs)  // exhausted only when all tools have been called at least once
}
```

This is a small, safe change. A focused unit test should be added: call the same tool 3 times with 3 tools registered — the budget should not be exhausted.

### H — `buildProviderPrompt` has no size bounds on derived text content

**File:** `internal/ai/openai_provider.go` — `buildProviderPrompt` (line 891)

The prompt builder writes every message, every attachment metadata line, and every derived text — in full — into a single string:

```go
for _, derived := range input.DerivedTexts {
    b.WriteString(fmt.Sprintf(
        "- attachment=%s message=%s type=%s text=%s\n",
        derived.SourceAttachmentID,
        derived.RequestMessageID,
        derived.DerivativeType,
        strings.TrimSpace(derived.ContentText), // ← full OCR/extracted text, no length limit
    ))
}
```

For a request with several PDF attachments that have been OCR'd, `ContentText` for each could be thousands of tokens. A request with 3 PDF attachments of 5 pages each, fully OCR'd, could produce a prompt exceeding 10,000 tokens before the system prompt and tool definitions are added. This risks hitting the model's context window limit, which causes an API error that `validateOpenAIResponse` cannot distinguish from other failures — the run fails permanently.

There is no truncation, no budget tracking, no warning when the prompt is large.

**Fix:**

Add a configurable `MaxDerivedTextChars` limit to `PromptConfig` (Gap 6 fix). Truncate derived text content that exceeds this limit with an explicit marker:

```go
// internal/ai/provider_config.go
type PromptConfig struct {
    // ... existing fields ...
    MaxDerivedTextChars int // per derived text item; 0 means no limit
    MaxMessageChars     int // per message text; 0 means no limit
}

func DefaultPromptConfig() PromptConfig {
    return PromptConfig{
        // ...
        MaxDerivedTextChars: 2000, // ~500 tokens per derived text item
        MaxMessageChars:     1000,
    }
}

// In buildProviderPrompt:
text := strings.TrimSpace(derived.ContentText)
if p.promptConfig.MaxDerivedTextChars > 0 && len(text) > p.promptConfig.MaxDerivedTextChars {
    text = text[:p.promptConfig.MaxDerivedTextChars] + " [truncated]"
}
```

Also add a prompt size warning log when the total prompt exceeds a configurable threshold (e.g., 6000 characters), so oversize prompts are visible in logs before they cause API errors.

### I — Failed coordinator runs leave requests permanently stuck with no alerting or recovery path

**File:** `internal/ai/coordinator.go` — `failRunAndRequest` (line 701), called from 9 sites

When any step of coordinator processing fails, `failRunAndRequest` is called:

```go
func (c *Coordinator) failRunAndRequest(ctx context.Context, run Run, reason string, actor identityaccess.Actor) {
    _, _ = c.aiService.CompleteRun(ctx, ...) // marks run as failed
    c.markRequestFailed(ctx, run.InboundRequestID.String, reason, actor) // marks request as failed
}
```

Then `markRequestFailed` calls `intakeService.AdvanceRequest` with `StatusFailed`. After this:

1. The request is permanently in `status = 'failed'` and will never be picked up by the queue processor again
2. No alert, webhook, or notification is triggered
3. No admin UI signal distinguishes a transient failure (rate limit) from a permanent one (bug)
4. Operators who submitted the request receive no indication that AI processing failed
5. Recovery requires a manual database update or a future admin "requeue" endpoint that does not yet exist

Improvement E (transient retry) reduces the frequency of this happening. But it does not eliminate it — bugs, unexpected response formats, context load failures, and persistent API outages will still trigger `failRunAndRequest`.

**Fix — three coordinated changes:**

**Part 1: Distinguish transient vs. permanent failures**

The `failRunAndRequest` function currently treats all failures identically. Add a `permanent bool` parameter:

```go
func (c *Coordinator) failRunAndRequest(ctx context.Context, run Run, reason string, permanent bool, actor identityaccess.Actor) {
    // If not permanent, mark as 'failed_transient' instead of 'failed'
    // A transient-failed request can be requeued by an admin or an automatic retry job
    status := intake.StatusFailed
    if !permanent {
        status = intake.StatusFailedTransient // new status value
    }
    // ...
}
```

**Part 2: Add an admin requeue endpoint**

Add a handler `handlePostAdminRequeueFailedRequest` that transitions a `failed` or `failed_transient` request back to `queued`, resets the failure reason, and logs an audit event. This allows operators to retry AI processing on a request without manual database intervention.

**Part 3: Make failed AI runs visible in the operator review surface**

The review UI should distinguish requests that failed AI processing from requests that are still queued. A clear operator-visible state (e.g., "AI processing failed — retry available") with the sanitized failure reason is more useful than silently leaving the request in a failed state.

### J — No operator feedback loop for coordinator output quality

The system has no mechanism to collect operator feedback on whether a coordinator brief was useful. This means:

1. Prompt iteration (enabled by the Gap 6 `PromptConfig` fix) has no feedback signal — you are tuning blind
2. There is no way to detect systematic quality regressions after a prompt change
3. There is no way to identify which request types consistently produce poor briefs
4. The instruction hash recorded in run metadata (Gap 6 fix, Step 3) cannot be correlated with quality without a feedback record to correlate it against

**Fix — add a brief feedback table and a simple UI gesture**

**Step 1: Add `ai.coordinator_brief_feedback` table**

```sql
CREATE TABLE ai.coordinator_brief_feedback (
    id                 TEXT        NOT NULL DEFAULT generate_random_id('cbf'),
    org_id             TEXT        NOT NULL,
    artifact_id        TEXT        NOT NULL REFERENCES ai.agent_artifacts(id),
    run_id             TEXT        NOT NULL REFERENCES ai.agent_runs(id),
    inbound_request_id TEXT        NOT NULL,
    rated_by_user_id   TEXT        NOT NULL,
    rating             TEXT        NOT NULL CHECK (rating IN ('useful', 'not_useful', 'misleading')),
    notes              TEXT,
    instructions_hash  TEXT,       -- from run metadata — enables quality/version correlation
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);
```

**Step 2: Add a backend handler**

```
POST /api/review/requests/:id/ai-brief-feedback
Body: { rating: "useful" | "not_useful" | "misleading", notes?: string }
```

**Step 3: Add a minimal UI gesture in the Svelte review surface**

A thumbs-up/thumbs-down affordance on the coordinator brief section. This does not need to be prominent — even a single aggregate metric (percentage of briefs rated "useful" per instruction version) is enough to make prompt iteration measurable.

### K — Tool suite breadth is the binding constraint on output quality

This is not a code defect. It is the most important strategic investment after the Phase 1–3 fixes are complete.

After all code fixes, the coordinator still has access to exactly three read tools:
- **Request detail** — what this request says
- **Queue status summary** — what the current queue looks like
- **Processed proposals** — what was done on recent comparable requests

This is enough for the coordinator to produce a competent general-purpose brief. It is not enough for the coordinator to produce an expert-quality, operationally actionable brief. For most real business scenarios, the coordinator lacks context that would make its output meaningfully more useful:

| Request type | Missing context | Impact on output quality |
|---|---|---|
| Purchase / procurement | Current inventory levels, pending POs for same items, available budget in GL account | Generic quantity recommendation vs. informed quantity recommendation |
| Personnel / HR | Employee record, department headcount, open requisitions | Generic routing vs. correct approval authority identification |
| Vendor or service | Vendor history, contract terms, past approval patterns | Generic risk summary vs. contract-aware risk summary |
| Approval escalation | Approval thresholds by role and value, current approval queue depth | Generic escalation recommendation vs. specific authority identification |

**Recommended expansion roadmap:**

Each new tool requires: a `coordinatorToolDefinition` struct, an `Execute` function backed by the appropriate service, and a policy entry in `ai.agent_tool_policies`. The tool policy system means each tool is opt-in per capability code — you can add a tool without immediately enabling it for all coordinator runs.

```go
// Example: inventory lookup tool
coordinatorToolDefinition{
    ToolName:        "inventory_get_item_levels",
    DisplayName:     "Get current inventory levels for requested items",
    ModuleCode:      "inventory",
    MutatesState:    false,
    DefaultPolicy:   PolicyAllow,
    Description:     "Returns current stock levels, pending orders, and reorder points for items mentioned in the request.",
    InputSchema:     inventoryLookupInputSchema(),
    Execute: func(ctx context.Context, input CoordinatorProviderInput) (string, string, error) {
        // call inventoryService.GetItemLevelsForCoordinator(ctx, input.Actor, extractedItemCodes)
    },
}
```

**Priority order for new tools:**

1. `inventory_get_item_levels` — most inbound requests involve inventory
2. `accounting_get_budget_position` — purchase and expense requests need budget context
3. `approvals_get_authority_for_value` — most requests need routing to the right approver
4. `vendor_get_summary` — vendor-related requests need history and contract context

Each tool expands the coordinator's context window of knowledge and directly improves the specificity and actionability of briefs — without changing any coordinator orchestration code.

---

## Implementation Sequencing

### Phase 1 — Immediate correctness (no new features)

Fix the gaps that affect currently-running coordinator runs. All items here are low-risk, low-effort, and can be shipped together.

| Item | Effort | Risk | Impact |
|---|---|---|---|
| Gap 2: `ResolveToolPolicy` read-only transaction | Low | Low | Removes `FOR UPDATE` lock on every tool call |
| Gap 4: Per-iteration timeouts | Low | Low | Prevents cascading timeout failures across tool loop |
| Gap 5: Degraded-mode signaling | Low | Low | Improves operator visibility when tools were blocked |
| Improvement B: `Name()` method on provider interface | Very low | Very low | Removes type-switch brittleness |
| Improvement F: Raise `MaxOutputTokens` to 2000/1500 | Very low | Very low | Prevents incomplete-response errors on long briefs |
| Improvement G: Fix budget exhaustion unique-tool counting | Very low | Very low | Prevents premature tool removal on repeated calls |

### Phase 2 — Structural refactor (own planned slice)

These are larger changes that should not be mixed with Phase 1.

| Item | Effort | Risk | Impact |
|---|---|---|---|
| Gap 3: `RequestContextLoader` interface + extracted file | Medium | Low | Enables unit testing without DB |
| Gap 6: `PromptConfig` externalization | Medium | Low | Enables prompt iteration without redeploy |
| Improvement A: Decompose `openai_provider.go` | Medium | Low | Reduces cognitive overhead, enables specialist additions |
| Improvement C: Registry-driven allowlist + dynamic schema enum | Low | Low | Single source of truth for specialist capabilities |
| Improvement D: Stopword list documentation + test | Low | Very low | Prevents accidental suppression of business terms |
| Improvement H: Prompt size bounds on derived text | Low | Low | Prevents context-window overflow on large requests |
| Improvement I: Failed run alerting + admin requeue endpoint | Medium | Low | Operators can recover stuck requests without DB intervention |
| **Arch Gap 1: PostgreSQL `LISTEN/NOTIFY` queue triggering** | **Medium** | **Low** | **Near-zero request processing latency; eliminates poll overhead** |

### Phase 3 — Specialist execution (significant new feature)

This is the most impactful gap but also the most work. It should be planned as a dedicated implementation slice after Phase 1 and 2 are stable.

| Item | Effort | Risk | Impact |
|---|---|---|---|
| Gap 1: `SpecialistProvider` interface + registry | High | Medium | Specialist delegation becomes real execution |
| Gap 1: `inbound_request.operations_triage` implementation | High | Medium | Actual domain-specific triage for operations requests |
| Gap 1: `inbound_request.approval_triage` implementation | High | Medium | Actual domain-specific triage for approval-ready requests |
| Improvement E: Transient error retry | Low | Low | Reduces coordinator failure rate from rate limits |
| **Improvement J: Operator brief feedback table + UI gesture** | **Medium** | **Low** | **Enables quality measurement and prompt iteration signal** |

### Phase 4 — Quality expansion (strategic investment)

After the system is reliable and specialist execution is real, the next constraint on output quality is tool suite breadth. These items expand what the coordinator can see and produce measurably better, more actionable briefs.

| Item | Effort | Risk | Impact |
|---|---|---|---|
| Improvement K: `inventory_get_item_levels` tool | Medium | Low | Inventory context in procurement and supply briefs |
| Improvement K: `accounting_get_budget_position` tool | Medium | Low | Budget position in purchase and expense briefs |
| Improvement K: `approvals_get_authority_for_value` tool | Medium | Low | Correct approval authority identification |
| Improvement K: `vendor_get_summary` tool | Medium | Low | Vendor-aware risk assessment in procurement briefs |
| **Arch Gap 2: `ScheduledRun` concept + proactive coordinator triggers** | **High** | **Medium** | **AI can proactively alert, summarize, and detect anomalies** |

---

## Verification Checklist

After each phase, run the following before marking it complete.

```bash
# Build — no regressions
go build ./cmd/... ./internal/...

# Vet
go vet ./cmd/... ./internal/...

# AI package tests (run with race detector)
set -a; source .env; set +a
go test -count=1 -race ./internal/ai/...

# Full suite with race — for Phase 2 structural changes
set -a; source .env; set +a
go test -p 1 -count=1 -race ./cmd/... ./internal/...

# Coordinator integration test specifically
set -a; source .env; set +a
go test -count=1 -race -run TestCoordinator ./internal/ai/...
```

### Phase 1 manual checks

After applying Gap 2 (read-only transaction):
1. Run a coordinator processing cycle against a test inbound request.
2. Confirm in PostgreSQL that no `FOR UPDATE` lock waits appear on `identityaccess.sessions` during tool execution.

After applying Gap 4 (per-iteration timeout):
1. Confirm the coordinator completes normally with fast API responses.
2. Simulate a slow first call (e.g., mock provider with artificial delay) and confirm subsequent iterations are not time-constrained by the first call's latency.

### Phase 3 specialist verification

After implementing Gap 1:
1. Submit a request that causes the coordinator to delegate to `operations_triage`.
2. Confirm the specialist run record has `agent_role = 'specialist'` and its artifact body differs from the coordinator's.
3. Confirm the delegation record links parent and child runs correctly.
4. Confirm that removing the specialist from the registry causes the coordinator to gracefully fall back to its own output, not fail.
5. Run the integration test suite in `internal/ai/coordinator_integration_test.go`.

---

## Relationship to the Broader Improvement Plan

The issues in this document are specific to `internal/ai/`. Several of them connect to cross-cutting improvements described in `code_review_and_improvement_plan.md`:

- **P2-6** (read-only `AuthorizeReadOnlyTx`) is required to implement Gap 2 cleanly.
- **P2-4** (remove `...any` variadic injection) should be coordinated with Gap 3 if the `Coordinator` constructor is being changed in the same pass.
- **P3-1** (attachment upload size limits) is not directly connected to this layer but affects what the coordinator receives in `CoordinatorAttachment.SizeBytes`.

Keep this document and `code_review_and_improvement_plan.md` in sync as implementation progresses. Do not let them diverge into independent sources of truth for the same issues.
