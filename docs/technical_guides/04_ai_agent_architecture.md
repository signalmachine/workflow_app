# AI Agent Architecture

Date: 2026-03-31
Status: Active technical guide
Purpose: explain how the coordinator, provider, tool policy, run history, and recommendation flow work together in `workflow_app`.

## 1. The AI architecture in one sentence

`workflow_app` uses a bounded coordinator-plus-specialist AI model where the provider can summarize, inspect, and recommend, but does not directly write business truth.

## 2. Main AI packages

The core package is `internal/ai`.

It owns:

1. AI runs
2. run steps
3. tool registrations
4. tool policy
5. artifacts
6. recommendations
7. delegation traces
8. provider integration

The surrounding packages provide intake context, reporting reads, and shared backend orchestration.

## 3. The coordinator flow

The coordinator is the entry point for processing the next queued request.

At a high level it does this:

1. claim the next queued request
2. create a coordinator run
3. load request, attachment, and derived-text context
4. call the provider with a bounded tool loop
5. persist the resulting artifact and recommendation
6. mark the request processed or failed

The most important rule is that the coordinator is still constrained by the database-backed control model. It is not a general autonomous agent.

## 4. Coordinator and provider split

The coordinator is the domain workflow.

The provider is the model-specific execution adapter.

That split matters because the coordinator owns workflow intent, while the provider only executes a bounded review task.

### 4.1 Coordinator input

```go
type ProcessNextQueuedInput struct {
	Channel string
	Actor   identityaccess.Actor
}
```

### 4.2 Provider input

```go
type CoordinatorProviderInput struct {
	CapabilityCode   string
	Actor            identityaccess.Actor
	RequestReference string
	Channel          string
	OriginType       string
	Metadata         json.RawMessage
	Messages         []CoordinatorMessage
	Attachments      []CoordinatorAttachment
	DerivedTexts     []CoordinatorDerivedText
}
```

The provider receives a compact snapshot of persisted request context. It does not receive open-ended system access.

## 5. Bounded tool use

The OpenAI provider uses the Responses API with a hard cap on tool loops.

Important properties:

1. the loop is bounded
2. tool calls are explicit
3. tool policy is enforced
4. output is validated before it becomes a durable recommendation
5. the default path favors request-scoped evidence over queue-summary evidence

Example configuration:

```go
return responses.ResponseNewParams{
	Model:             p.model,
	Store:             openai.Bool(false),
	Temperature:       openai.Float(0.1),
	MaxOutputTokens:   openai.Int(900),
	MaxToolCalls:      openai.Int(1),
	ParallelToolCalls: openai.Bool(false),
}
```

The low temperature, single-call loop, and `store: false` setting are all deliberate safety choices.

## 6. Tool policy

Tool policy is a first-class control boundary.

The package supports policies such as:

1. `allow`
2. `approval_required`
3. `deny`

This is how the application keeps model behavior aligned with business risk. The model may ask for a tool, but the policy layer decides whether that tool is actually allowed in the current capability context.

## 7. Specialist delegation

The coordinator can optionally route one allowlisted specialist capability through a durable child run and delegation record.

That means:

1. the coordinator can stay narrow
2. specialist routing is persisted
3. the operator can inspect the delegation trail later
4. the specialist still lives inside the same controlled workflow model

The allowlisted specialist capabilities are intentionally bounded. The provider is not allowed to invent arbitrary sub-agents.

## 8. Durable artifacts and recommendations

The provider response becomes durable only after validation.

The key persisted outputs are:

1. coordinator run
2. provider execution step
3. artifact
4. recommendation
5. optional delegation record

Example output shape:

```go
type CoordinatorProviderOutput struct {
	ProviderResponseID   string
	ProviderName         string
	Model                string
	Summary              string
	Priority             string
	ArtifactTitle        string
	ArtifactBody         string
	Rationale            []string
	NextActions          []string
	SpecialistDelegation *CoordinatorSpecialistDelegation
}
```

The recommendation is review material, not a direct write instruction.

## 9. Live OpenAI adapter

`internal/ai/openai_provider.go` is the provider adapter for OpenAI.

It:

1. loads the model and API key from config
2. builds a Responses API request
3. executes a bounded tool loop
4. validates and repairs structured output where necessary
5. records tool executions and usage

The adapter uses request-scoped read tools such as request detail and current proposal continuity before falling back to queue summary context.

## 10. Entry points worth knowing

The main entry points are:

1. `ai.NewCoordinator`
2. `Coordinator.ProcessNextQueued`
3. `ai.NewOpenAIProvider`
4. `app.NewOpenAIAgentProcessorFromEnv`
5. `cmd/verify-agent`

If you are changing the AI path, start with these.

## 11. Example of the safety model

The provider prompt explicitly tells the model not to write business truth directly.

That is the right pattern for this application:

1. provider produces a bounded recommendation
2. the system stores the recommendation
3. the operator decides whether to act on it
4. domain services still enforce approval and posting boundaries

## 12. What can break easily

Be careful with:

1. tool loop limits
2. provider output validation
3. request-centered prompt instructions
4. specialist delegation constraints
5. run status transitions
6. live-provider test assumptions

If any of those drift, the AI path may still "work" but stop being safe enough for controlled workflow use.

