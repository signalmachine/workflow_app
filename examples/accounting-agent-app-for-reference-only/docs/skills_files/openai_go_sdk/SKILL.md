---
name: openai_go_sdk
description: Project-specific OpenAI Go SDK patterns for Responses API, strict JSON schema output, and agentic tool loops. Use when editing internal/ai code or diagnosing OpenAI SDK issues.
---

# OpenAI Go SDK (Project Skill)

Use this skill when changing AI behavior in `internal/ai/`.

## Scope
- Primary implementation: `internal/ai/agent.go`, `internal/ai/tools.go`
- SDK version source of truth: `go.mod`
- Local SDK reference mirror: `examples/openai-go-sdk-reference/`

## Hard Rules
1. Use OpenAI **Responses API** for agent flows.
   - Allowed: `client.Responses.New`, `client.Responses.NewStreaming`
   - Disallowed for agent flows: `client.Chat.Completions.New`
2. Keep integration changes in `internal/ai/` and do not move accounting validation into AI.
3. Preserve human-in-the-loop write behavior: AI proposes, core services validate/commit.

## Current Repo Patterns (Follow Exactly)

### A) Structured output for journal entry interpretation
- Path: `Agent.InterpretEvent`
- Pattern:
  - `responses.ResponseNewParams`
  - `Model: openai.ChatModelGPT4o`
  - `Text.Format.OfJSONSchema` with `Strict: openai.Bool(true)`
  - Hand-built strict schema map (`generateSchema` / `proposalSchema`)

### B) Agentic tool loop for domain actions
- Path: `Agent.InterpretDomainAction`
- Pattern:
  - `Tools: []responses.ToolUnionParam{...}`
  - `PreviousResponseID` chaining across iterations
  - Loop cap (`maxLoops`) to avoid runaway tool cycles
  - Read tools auto-execute; write/meta tools terminate loop and return for confirmation/routing

### C) Input shapes
- Text-only input: `ResponseNewParamsInputUnion{OfString: ...}`
- Tool output turn: `ResponseNewParamsInputUnion{OfInputItemList: ...}`
- Image attachments: `ResponseInputImageParam{ImageURL: param.NewOpt(dataURL)}`

## Strict Schema Notes
- For strict JSON schema mode:
  - Use `additionalProperties: false` on every object.
  - Ensure `required` lists all properties.
  - Represent nullable values with `anyOf: [{...}, {"type":"null"}]`.
- Avoid relying on loose optional behavior for structured accounting outputs.

## Error and Reliability Baseline
- Wrap API calls with timeout contexts.
- Inspect `*openai.Error` via `errors.As`.
- Log usage metrics (`resp.Usage`) when present.
- Normalize/validate model outputs before core commit paths.

## Change Checklist
1. Confirm no Chat Completions calls were introduced.
2. Confirm tool loops still have a hard max iteration cap.
3. Confirm schema strictness did not regress.
4. Run:
   - `go test ./internal/core -v`
   - `go run ./cmd/verify-agent` (if `OPENAI_API_KEY` is set)

## Common Pitfalls
- Mixing old SDK snippets that use outdated helper conventions.
- Breaking loop continuity by omitting `PreviousResponseID` where needed.
- Returning free-form text where structured output is expected.
- Removing defensive validation after model output parsing.
