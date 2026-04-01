# Repository Guidelines

## Project Structure & Module Organization

This repository is still plan-first, but it also contains a runnable Go backend slice. The active planning working set now lives in `plan_docs/`. Start with `plan_docs/README.md` for the required reading order. `implementation_plan/` is legacy reference material only; use it when `plan_docs/` points to it or when a specific historical implementation detail needs to be checked. Use `docs/` only for supporting material that is not part of the active planning set. `docs/implementation_objectives/implementation_objectives.md` is a companion high-level summary and is useful when a cross-cutting objectives or invariants summary is needed, but it is not mandatory reading for every implementation session.

Key files:
- `plan_docs/README.md`: active planning reading order and legacy-reference rule
- `plan_docs/service_day_refactor_tracker_v1.md`: single live implementation-status tracker inside `plan_docs`; start here for current progress, evidence, and next-step status
- `docs/implementation_objectives/implementation_objectives.md`: high-level summary of implementation rules, objectives, requirements, and invariants
- `plan_docs/current_state_review_2026_03_18.md`: current codebase and planning-shape review
- `plan_docs/service_day_refactor_principles_v1.md`: active thin-v1 doctrine
- `plan_docs/service_day_ai_architecture_v1.md`: active thin-v1 AI architecture and observability plan
- `plan_docs/service_day_foundation_coverage_v1.md`: active thin-v1 foundation checklist
- `plan_docs/service_day_schema_and_module_boundaries_v1.md`: active thin-v1 ownership and schema-boundary baseline
- `plan_docs/service_day_implementation_defaults_v1.md`: active thin-v1 locked implementation defaults
- `plan_docs/service_day_thin_v1_scope_v1.md`: hard v1 scope and explicit v2 cuts
- `plan_docs/service_day_refactor_execution_plan_v1.md`: active milestone path
- `plan_docs/service_day_refactor_tracker_v1.md`: active thin-v1 execution status
- `plan_docs/implementation_plan_migration_map_v1.md`: mapping from legacy docs to the active plan
- `implementation_plan/service_day_*`: legacy strategy, scope, schema, and architecture references

Focused remediation notes under `implementation_plan/` are allowed when needed for a narrow follow-up on older slices, but they are not active canonical planning by themselves.

## Document Hygiene

Keep this file concise and durable. Store only repository-wide contributor rules here; move long reference maps or session notes into `docs/` or a canonical planning document.
Do not remove essential operating instructions from `AGENTS.md` purely to make it shorter; brevity is secondary to preserving important repository-wide guidance.
When high-level objectives, rules, principles, specifications, or invariants change in `AGENTS.md`, `README.md`, or the active `plan_docs/` set, review `docs/implementation_objectives/implementation_objectives.md` and update it when needed so the summary stays aligned.
Treat `docs/archive/` as deprecated legacy material. Do not update archived documents during normal implementation or documentation work unless the user explicitly asks for archive maintenance.

For codebase review requests, review `AGENTS.md` itself for stale guidance and update it when the repository has outgrown an instruction.

## MCP Usage

Use MCP tools when they materially improve accuracy or speed.

- `gopls` MCP is required for every Go coding session.
- Start every Go coding session with `go_workspace`.
- Use `gopls` as the default path throughout the session for Go workspace summary, symbol search, package context, references, safe renames, diagnostics, and vulnerability checks whenever it materially fits the task.
- If a session includes Go code changes, run `go_diagnostics` on edited files before completion and use `go_vulncheck` when dependencies or security-sensitive code change.
- GitHub MCP should be used for repository history, pull request context, issue context, and upstream verification when local files are incomplete.
- For planning-only sessions, `plan_docs/` is the active source of truth; do not force MCP usage when the task is purely document maintenance.

## Required Go Workflow

For every Go implementation session:

- review `plan_docs/README.md`, then `plan_docs/service_day_refactor_tracker_v1.md`, and then the relevant active planning documents for the slice being changed
- consult `implementation_plan/` only when a specific historical slice or legacy decision needs clarification
- follow the detailed workflow in `docs/go_workflow.md`
- do at least a targeted code review of the slice being changed either before implementation, after implementation, or both when the slice is risky
- do not treat implementation work as complete until the required verification has run or an explicit blocker has been documented
- prefer `gopls` over ad hoc text searching whenever the task involves package understanding, symbol discovery, cross-file references, API inspection, diagnostics, or safe rename/refactor operations
- use `staticcheck` as an additional non-routine verification tool for dedicated hardening, cleanup, and code-quality passes or when requested; do not treat it as a replacement for the standard required verification workflow

Continuous review rule:
- use implementation sessions to review the touched code paths continuously rather than waiting for a single late full-codebase review
- check that the current code still aligns with `plan_docs/`, any active implementation decisions, relevant legacy context when needed, and adjacent code contracts
- when you find a bug, inconsistency, missing guardrail, or implementation-plan drift, resolve it in the same session when practical
- if the issue should not be fixed immediately, document it in the active planning material or in a focused legacy remediation note when the issue belongs to an older slice
- keep these reviews pragmatic and slice-scoped by default; do not turn every session into a full broad review unless the task explicitly calls for that level of review

## Build, Test, and Development Commands

The repository now includes a runnable Go application and migration entrypoint.

- `git status --short`: review local changes before editing shared planning files
- `sed -n '1,200p' plan_docs/README.md`: confirm active document order
- `rg "v1|v2|legacy" plan_docs implementation_plan/`: trace active versus legacy planning references quickly
- `go run ./cmd/migrate`: apply embedded PostgreSQL migrations
- `go run ./cmd/app`: start the HTTP application
- `go build ./cmd/app ./cmd/migrate`: verify the runnable binaries build cleanly
- `APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app`: short startup smoke check for runtime-facing changes
- `go test ./...`: run the current automated test suite
- `go vet ./...`: optional low-cost extra verification pass for dedicated review or hardening sessions
- `staticcheck ./...`: preferred additional static-analysis pass when doing non-routine quality review
- see `docs/go_workflow.md` for the current integration-test command and close-out checklist

## Coding Style & Naming Conventions

Keep Markdown concise and implementation-oriented. Use sentence-case prose, short numbered lists where order matters, and fenced code blocks only for commands or schema examples. Follow existing naming patterns such as `service_day_execution_plan_v1.md`.

Do not create overlapping planning documents. Update the canonical file for the topic instead.

For Go code, keep package boundaries explicit and prefer idiomatic naming.

## Architecture & Safety Guardrails

Preserve the active thin-v1 defaults in `plan_docs/`. Keep AI actions proposal-first, route meaningful writes through normal domain services, and preserve auditability for financially significant changes. Treat reports and timelines as derived read models, not systems of record.

Product-priority rule:
- treat `work_order` support as the strongest long-term capability of `service_day`
- do not let CRM, project management, or any other single surrounding layer displace `work_order` execution as the core operating center of the product
- when planning or implementing features, prefer designs that strengthen the handoff into, out of, and around `work_orders`

Thin-v1 priority rule:
- accounting, foundational GST/TDS support, inventory, `work_orders`, tasks, AI control boundaries, approvals, reports, and agent observability are the highest v1 priorities
- CRM and `projects` are support concerns in v1 unless a specific implemented slice is being maintained or a foundation dependency requires them
- if a capability can safely move to v2 without weakening the foundation, keep it out of v1

Cross-module-integration rule:
- treat tight integration across parties, documents, ledgers, `work_orders`, tasks, inventory/material flows, accounting, reporting, and AI as a core product strength
- avoid designs that make modules feel like separate apps joined only by superficial links
- preserve clear ownership boundaries, but require explicit identifiers, handoff contracts, and read models so cross-module workflows stay low-friction and technically coherent

Mobile-client rule:
- the first planned mobile client may use Flutter
- do not treat Flutter as a substitute for backend mobile-readiness work
- keep backend auth, versioning, sync, attachment, notification, and retry/idempotency plans client-agnostic unless a later canonical decision explicitly narrows them

Database-first rule:
- treat `service_day` as a database-first and SQL-first application
- when an invariant, constraint, linkage rule, derivation, or safety check can be implemented in PostgreSQL, prefer implementing it in PostgreSQL before relying on Go code alone
- use Go code as the second enforcement layer, not the only one; the database should still protect correctness if application code is buggy
- take full advantage of modern PostgreSQL features when they materially improve correctness, auditability, performance, or operability
- prefer sophisticated PostgreSQL-native models when they materially improve the system, as long as that sophistication does not make the implementation meaningfully harder to build, evolve, or operate; the right bar is strong modeling with painless implementation rather than simplistic modeling by default
- because AI agents are expected to be the main operators in v1, design the schema and data model so invalid states, invalid transitions, and unsafe postings are rejected by default rather than relying on agent correctness

AI-agent-first rule:
- treat `service_day` as an AI-agent-first application within the locked write-boundary and audit rules
- use AI agents for product tasks that can be delegated safely through explicit tools, approvals, and normal domain services
- design AI agent architecture and workflows using modern agent patterns rather than ad hoc prompt wrappers or hidden sidecar state
- optimize near-term planning for observing and improving agent behavior on bounded business work, not for maximizing production breadth
- prefer a multi-agent architecture with one coordinator agent routing work to bounded specialist agents for workflow execution

Minimal-human-UI rule:
- do not build broad human operational UIs for v1
- keep v1 human surfaces focused on review, approval, inspection, and reporting

Already-implemented-slice rule:
- do not remove or de-scope already-implemented application parts just because the new thin-v1 plan would not choose to build them first
- keep already-implemented slices when they do not conflict with the active thin-v1 architecture
- do not let those existing slices re-expand the active v1 plan unless they directly support the thin-v1 foundation

Technical-rigor rule:
- prioritize technically sound data modeling, boundaries, invariants, and operational correctness over flashy or fashionable features
- if a requested design or feature direction is not technically solid, creates avoidable long-term complexity, or conflicts with best practice for software of this kind, push back clearly and correct the direction before implementing it
- prefer durable architecture and maintainable execution paths even when that means narrowing or delaying a feature
- always follow what is technically solid and what is best practice in the relevant domain rather than copying weak conventions from comparable products or target markets
- design the data model so it can adapt later to justified custom requirements without forcing foundational rewrites

## Testing Guidelines

For planning-only work, validation is currently review-based. Before marking work complete, make sure:
- the relevant canonical document is updated
- any affected active planning guidance in `plan_docs/` is internally consistent
- if legacy material was consulted or contradicted, that relationship is made explicit in the updated docs

For implementation work:
- every behavior change should include tests appropriate to the change
- bug fixes should include a regression test when practical
- run `go build ./cmd/app ./cmd/migrate` before closing out the task
- run `go test ./...` before closing out the task
- when the change touches startup, config loading, HTTP wiring, middleware, or another runnable application path, do a short app smoke run such as `APP_LISTEN_ADDR=127.0.0.1:18080 timeout 3s go run ./cmd/app` unless an explicit blocker prevents it
- if persistence behavior or migrations change, add or update integration coverage and run the relevant PostgreSQL integration suite when the affected area is covered there
- when running a dedicated quality or review pass, prefer adding `staticcheck ./...` if the tool is available locally and the session benefits from deeper static analysis
- record the exact verification command or result in the active planning or delivery notes when it materially advances a slice or milestone

Go quality-tool guidance:
- `staticcheck` is the preferred additional Go quality tool for this repository beyond the default workflow
- `go vet` is a reasonable optional extra pass when a lightweight second opinion is useful
- `errcheck` can be valuable later if the codebase grows more IO-heavy or concurrency-heavy, but it is not currently part of the recommended routine
- `ineffassign` is lower priority because `staticcheck` already covers much of that signal
- avoid adding broad style-only lint stacks by default; favor low-noise tools that improve correctness and maintainability signal

## Commit & Pull Request Guidelines

Git history is minimal, with short subjects such as `Update`, but contributors should use specific imperative messages instead, for example `Clarify Milestone A schema exit criteria`.

Commit timing note:
- committing changes to git may sometimes be delayed during active work
- do not treat an uncommitted but present local change as missing implementation by itself; the checked-out codebase is the source of truth for current status

Pull requests should:
- explain the purpose and affected canonical files
- note any decision, scope, or sequencing change
- include active planning-status updates when milestone state or evidence changes
- avoid mixing unrelated planning edits in one review

## Planning Workflow Notes

If scope changes, update the active `plan_docs/` document first. Do not mark items `done` without concrete evidence.

Implementation-tracking rule:
- use `plan_docs/service_day_refactor_tracker_v1.md` as the single live implementation-status document in `plan_docs/`
- update other `plan_docs/` files for scope, architecture, ownership, sequencing, or invariant changes, not routine progress notes
- use `implementation_plan/` only as legacy reference material, not as the active implementation tracker

Whenever you find a bug, issue, or inconsistency in the codebase, do not move past it silently. Resolve it, inform the user about it, or document it in the appropriate active planning material before proceeding.

If you create a separate remediation or review note under `implementation_plan/` for deferred legacy issues, make its legacy role explicit and do not let it silently override `plan_docs/`.

Pre-production database reset rule:
- while the product is still pre-production, `TEST_DATABASE_URL` should be treated as disposable and may be dropped and recreated when schema work requires a clean reset
- the active local development database may also be treated as disposable when a foundational data-model correction is clearly better than preserving a weak interim shape
- use that freedom to improve the model decisively, not to excuse sloppy schema design or avoid careful migration thinking
- if a radical schema or data-model change requires resetting databases or materially changing migration direction, update the relevant active planning document first and then record the decision in the same planning wave
- once external users, shared demo environments, or integrations depend on the data shape, stop treating the main development database as casually disposable and raise the compatibility bar accordingly

After implementation changes, update `README.md` when setup, commands, architecture shape, or user-visible capabilities have changed.
When user-visible behavior, setup steps, or operational workflows change materially, update the relevant guides under `docs/user_guides/` in the same implementation wave or explicitly record why the guide update is deferred.

When the repository has a working `.env` and `TEST_DATABASE_URL`, treat `.env` as the canonical local configuration source so local app runs, migrations, and integration tests all use the same values instead of ad hoc shell exports.
If `.envrc` is present, treat it as a thin `direnv` wrapper around `.env`, not as a second independent config source. Prefer `direnv`-managed shells for routine local runs, but keep commands compatible with explicit `.env` sourcing when needed.
