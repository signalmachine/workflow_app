# Issue 11 Refactor Plan — Remove Direct `internal/core` Imports from Adapters

Date: 2026-03-03
Status: Planned
Priority: Medium (architectural quality)

## Background

Current adapter packages (especially under `internal/adapters/web`) directly import model types from `internal/core`.
This violates the intended layering rule:

- Adapters depend on `internal/app` only.
- `internal/app` orchestrates use-cases and maps domain (`internal/core`) to transport-safe DTOs.
- `internal/core` remains isolated from transport/presentation concerns.

## Goal

Remove adapter-level imports of `internal/core` without changing user-visible behavior.

## Scope

In scope:

1. Web adapter imports of `internal/core`.
2. Application service return/input types used by adapters.
3. Mapping from core domain models to app-level DTOs.
4. Tests impacted by type/signature changes.

Out of scope (separate initiatives):

1. Domain behavior changes.
2. DB schema changes.
3. UI redesign.

## Target Architecture

- `internal/adapters/*` imports:
  - allowed: `internal/app`, stdlib, framework libs.
  - disallowed: `internal/core`.
- `internal/app`:
  - defines adapter-facing DTOs (either in `internal/app` or `internal/app/model`).
  - maps to/from `internal/core` internally.

## Implementation Plan

### Phase 1: Inventory and Contract Design

1. Enumerate all direct `internal/core` imports in adapters.
2. For each imported type, define an app-layer DTO equivalent.
3. Decide DTO location:
   - Keep in `internal/app` if small.
   - Move to `internal/app/model` if DTO set grows.

Deliverable: type mapping table (`core type -> app DTO`) and API contract updates.

### Phase 2: App Service Type Migration

1. Update `ApplicationService` interface to expose only app DTOs.
2. Implement mappings in `app_service.go`:
   - core -> DTO for reads.
   - DTO -> core for write commands where needed.
3. Keep method semantics unchanged.

Deliverable: `internal/app` compiles with updated contracts.

### Phase 3: Adapter Migration

1. Replace all adapter references to `core.*` types with `app.*` DTOs.
2. Remove direct `internal/core` imports from:
   - `internal/adapters/web/chat.go`
   - `internal/adapters/web/accounting.go`
   - `internal/adapters/web/cli.go`
   - `internal/adapters/web/display.go`
   - any additional files found during inventory.
3. Ensure templates/JSON handlers still render identical payloads.

Deliverable: adapters compile with no `internal/core` imports.

### Phase 4: Tests and Safeguards

1. Update unit/integration tests impacted by DTO changes.
2. Add an architectural guard test or lint check:
   - fail build if `internal/adapters/**` imports `internal/core`.
3. Run full test suite.

Deliverable: regression-safe, enforceable layering rule.

## Acceptance Criteria

1. `rg "internal/core" internal/adapters` returns no matches.
2. `go test ./...` passes.
3. No API response regressions for existing endpoints.
4. No behavior change in journal entry proposal/commit flow.

## Risks and Mitigations

1. Risk: subtle JSON shape drift.
   - Mitigation: snapshot tests or golden-response checks for critical endpoints.
2. Risk: large PR becomes hard to review.
   - Mitigation: split into phase-based PRs (contracts, then adapters, then guardrails).
3. Risk: accidental circular dependencies.
   - Mitigation: keep DTO package under `internal/app` only; no adapter package imports back into app internals.

## Suggested Work Breakdown

1. PR 1: DTO definitions + `ApplicationService` signature updates + mapping layer.
2. PR 2: Web adapter migration + template compile/test fixes.
3. PR 3: CLI/display migration + architectural guard check.

## Done Checklist

- [ ] Adapter imports cleaned.
- [ ] App DTO mappings complete.
- [ ] Tests updated and passing.
- [ ] Guardrail added to prevent regression.
