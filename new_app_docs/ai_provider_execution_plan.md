# workflow_app AI Provider Execution Plan

Date: 2026-03-22
Status: In-progress thin-v1 implementation slice
Purpose: define the foundation-complete provider-backed AI execution layer required after Milestone 5 so `workflow_app` is usable as an AI-agent-first application in v1 rather than only having AI persistence and observability scaffolding.

## 1. Problem statement

The current thin-v1 codebase has durable AI control-boundary foundations:

1. AI runs, steps, artifacts, recommendations, tool policy, and delegation traces persist durably
2. persisted inbound requests can link to AI runs
3. reporting can review the request -> run -> recommendation -> approval -> document chain
4. bounded coordinator-to-specialist routing is modeled in the schema and service layer

What is still missing:

1. integration tests that can exercise the live OpenAI path when an OpenAI API key is configured
2. a focused provider-verification command and the promoted backend or web contracts needed to exercise the live path outside direct service calls

This gap is now important for thin v1 because the application is intended to be AI-agent-first. Without a live provider-backed path, the current AI layer remains an observability and control scaffold rather than a usable operator interface.

Current implementation checkpoint:

1. `internal/ai` now loads optional OpenAI provider configuration from `OPENAI_API_KEY` and `OPENAI_MODEL`
2. the official OpenAI Go SDK is now part of the active codebase
3. the first provider-backed adapter now uses the Responses API with strict structured output for queued inbound-request review
4. the first coordinator flow can claim one queued inbound request, assemble request, attachment, and derived-text context, create a coordinator run and step, persist a provider brief artifact and operator-review recommendation, and mark the request `processed` or `failed`
5. the coordinator path now includes a hard-capped Responses tool loop, per-capability tool-policy enforcement, and the first reporting read tool for inbound-request status summaries, with tool-execution metadata persisted in the coordinator step, artifact, and recommendation payloads
6. the coordinator can now optionally route one allowlisted specialist delegation through a durable child run and delegation record before the final artifact and recommendation are persisted
7. provider-backed business writes still terminate at artifact and recommendation persistence rather than bypassing approvals, postings, or normal domain services

## 2. V1 objective

Land the provider-backed AI execution foundations required for a usable v1 agent layer so:

1. queued inbound requests can be processed through a real OpenAI-backed coordinator flow
2. the coordinator can route bounded work to one allowlisted specialist capability while preserving the existing durable run history and delegation model
3. AI still acts only through explicit tools and normal domain services
4. approval, posting, audit, and write boundaries remain unchanged
5. the implementation remains narrow in product breadth while still being complete in provider, tool-loop, safety, configuration, and verification foundations

## 3. V1 scope

In scope:

1. OpenAI Go SDK integration in `internal/ai`
2. Responses API as the execution primitive for new agent flows
3. config and environment support for `OPENAI_API_KEY`, model selection, and bounded provider settings
4. a provider-backed coordinator path that can read persisted inbound requests, assemble request and attachment context, and produce artifacts or recommendations through normal domain services
5. bounded coordinator-to-specialist delegation built on the existing durable run and delegation records
6. explicit tool-registration, tool-policy enforcement, and approval-aware tool-loop handling in the provider-backed execution path
7. structured-output and validation boundaries where provider output drives domain proposals or recommendations
8. provider timeout, retry, and error handling rules that preserve business-state safety
9. one minimal HTTP or API surface for session-auth, request submission, attachment upload and download, and review-oriented reads around the live AI path
10. real integration tests gated on configured OpenAI credentials
11. a repository-level verification command for provider-backed AI behavior when credentials are present

Out of scope:

1. multiple providers in v1 unless a concrete foundation need appears
2. broad autonomous agent behavior
3. chat-product UX, conversational memory breadth, or broad conversation-history product depth
4. rich multimodal client capture beyond the already-approved inbound-request and attachment foundation
5. broad browser UI work

Companion promotion:

1. the minimum backend API and attachment transport contract required to exercise the live AI path is now treated as active v1 work rather than remaining only in `app_v2_plans/v2_client_and_multimodal_surfaces.md`

Thin-v1 rule:

1. the AI layer may stay narrow in workflow breadth, but it should not stay thin in provider-execution foundation depth

## 4. Required architecture

### 4.1 Provider stance

V1 should be OpenAI-first:

1. the first live provider is OpenAI
2. the implementation should use the OpenAI Go SDK
3. the implementation should use the Responses API for agent execution flows
4. old chat-completions-style agent execution should not be introduced as the primary path

### 4.2 Control model

The provider-backed execution path must preserve the existing business-control architecture:

1. provider calls may draft, summarize, classify, recommend, and select tools
2. provider calls must not write business state directly
3. all writes still go through domain services and their existing audit, approval, and posting boundaries
4. provider failures must not mutate business truth
5. model outputs that drive writes must remain bounded, validated, and policy-gated
6. the AI layer should use modern workflow AI agent architectures that are suitable for controlled business workflows rather than open-ended autonomy patterns
7. tool calling should be the primary execution mode for provider-backed agent behavior
8. AI tool handlers should call the existing domain services in the current codebase and should remain thin orchestration adapters rather than implementing duplicate workflow logic inside `internal/ai`

### 4.3 First target flow

First usable flow:

1. a queued inbound request is claimed into processing
2. the coordinator run is created and linked to that request
3. request text, attachments, and derived texts are assembled into the provider input
4. the coordinator uses the Responses API to classify the work and either:
5. produce a bounded recommendation directly
6. or delegate to one specialist capability through the existing child-run path
7. read tools may execute automatically where policy allows
8. write or approval-requiring tools must terminate into recommendation or approval-request paths rather than silently committing business changes
9. the resulting artifact, recommendation, approval linkage, and downstream document chain remain reviewable through the existing reporting layer

## 5. Configuration requirements

V1 configuration should include:

1. `OPENAI_API_KEY`
2. `OPENAI_MODEL`
3. optional provider tuning only where needed for safe operation
4. `.env.example` entries for the provider-backed AI path
5. documentation for local setup and verification when provider-backed AI is enabled
6. a clear contract for local development without provider credentials versus live-provider verification with credentials
7. API-facing configuration notes only where needed for the minimum live request-ingest and review path

Configuration rules:

1. provider-backed AI must fail clearly when credentials are missing
2. the codebase should still build and default tests should still run when provider credentials are absent
3. real-provider tests must be opt-in and clearly separated from default database-backed test runs

## 6. Test and verification plan

Required coverage:

1. unit or service tests for provider-config parsing and missing-key behavior
2. tests for coordinator-run creation and status transitions around provider execution
3. tests for tool-policy enforcement inside the provider-backed path
4. tests for request -> provider -> artifact or recommendation -> approval linkage persistence
5. tests for the minimum API contract around the live request-ingest and review path
6. OpenAI-backed integration tests gated on `OPENAI_API_KEY`
7. one explicit verification command for the provider-backed AI flow against a live API when configured
8. tests for structured-output validation and refusal or failure handling where provider output does not satisfy the expected contract

Expected verification shape:

1. default `go test ./...` remains provider-independent
2. provider-backed tests run only when the required environment is present or an explicit integration tag is used
3. the repository should gain a focused verification entrypoint such as `cmd/verify-agent` rather than burying all live-provider checks inside broad test runs

## 7. Planned implementation order

Recommended sequence after Milestone 5:

1. add configuration and `.env.example` support for OpenAI credentials and model selection
2. add bounded specialist delegation on top of the now-live coordinator and tool-loop execution path
3. add live-provider integration tests plus the explicit verification command
4. add the minimum API surface and attachment transport contracts needed to exercise that path outside direct service calls
5. update reporting or operational docs only where needed to explain the now-live provider-backed path

Execution rule:

1. treat this milestone as an umbrella for multiple small vertical slices, not as one monolithic implementation push
2. prefer delivering one real end-to-end workflow path first, then widening capability depth incrementally on the same controlled architecture

## 8. Success criteria

This slice is complete only when:

1. the active codebase uses the OpenAI Go SDK in `internal/ai`
2. the Responses API is the primary execution path for the new agent flow
3. `.env.example` documents the required OpenAI configuration
4. a real queued inbound request can drive a provider-backed coordinator run
5. that run can produce durable artifacts, recommendations, approval linkage, and specialist delegation through the existing control model
6. provider-backed execution includes the core safety and reliability foundations needed for a usable v1 AI layer rather than stopping at a thin happy-path demo
7. the provider-backed path is reachable through shared backend contracts that the promoted v1 web layer and later mobile client can both use
8. live-provider verification exists and is opt-in rather than silently required for all contributors
9. the resulting implementation remains bounded and auditable rather than becoming a broad autonomy or chat-product expansion
