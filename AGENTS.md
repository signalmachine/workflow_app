# Repository Guidelines

## Project Structure & Module Organization

This repository is the active `workflow_app` implementation codebase, with the canonical planning set living in `new_app_docs/`. Start with `new_app_docs/README.md`, then use `new_app_tracker.md` for current status and next implementation steps. Core doctrine, scope, defaults, and execution order are defined in `new_app_v1_principles.md`, `new_app_v1_scope.md`, `new_app_implementation_defaults.md`, and `new_app_execution_plan.md`. Use `new_app_docs/app_v2_plans/` only for explicitly deferred v2 work. `docs/workflows/` is the separate workflow-reference and workflow-validation layer for supported operator workflows, reusable validation checklists, live-review evidence, and later user-guide preparation; it is not the live implementation-planning source, and active workflow testing should be tracked there rather than in `new_app_docs/`. `docs/implementation_objectives/implementation_objectives.md` is a companion summary, not a replacement for the canonical planning set. Treat `docs/implementation_objectives/implementation_principles.md` and everything under `examples/` as read-only reference material from older implementations or planning eras; nothing in `examples/` is part of the active `workflow_app` implementation surface.

## Document Hygiene

Keep `AGENTS.md` short and durable. Put repository-wide contributor rules here, and move detailed plans, review notes, or session-specific material into the appropriate document under `new_app_docs/` or `docs/`. After every implementation change, review the canonical docs in `new_app_docs/` and update status, completed work, and next steps in the same change whenever they have drifted. Keep implementation planning in `new_app_docs/` and keep active workflow validation, reusable workflow checklists, and live-review evidence in `docs/workflows/`; do not silently mix those tracks again. When user-visible workflow behavior, durable workflow status, or reusable live-validation checklists change materially, review `docs/workflows/` and update it if the workflow reference has drifted. When high-level rules, principles, scope boundaries, or invariants change in `AGENTS.md`, `new_app_docs/`, or `README` files, review `docs/implementation_objectives/implementation_objectives.md` and update it if the summary has drifted. `docs/implementation_objectives/implementation_principles.md` is reference-only and does not need maintenance sync when canonical docs change. After implementation changes, update `README.md` when setup, commands, architecture shape, or user-visible capabilities have changed.

## MCP Usage

Use MCP tools when they materially improve accuracy or speed.

For every Go coding session in this repository once the Go workspace exists:

- `gopls` MCP is required and should be the default path through the session
- start with Go workspace context
- use the dedicated `mcp__gopls__...` tools such as `mcp__gopls__go_workspace` for Go workspace context rather than assuming `gopls` exposes `workspace://...` resources through `read_mcp_resource`
- use `gopls` for workspace summary, symbol search, package context, references, safe renames, and diagnostics whenever it materially fits the task
- if a session includes Go code changes, run diagnostics on edited files before completion
- use vulnerability checks when dependencies or security-sensitive code change
- when implementing or verifying `internal/ai` against the OpenAI Go SDK, prefer official OpenAI docs and the official `openai/openai-go` repository via MCP or approved web lookup for exact SDK and API details rather than relying on memory or local skills alone

For planning-only or Markdown-only sessions, do not force MCP usage when local document reads are sufficient.

## Required Go Workflow

For every Go implementation session:

- start with Go workspace context plus the relevant canonical docs in `new_app_docs/`
- use `new_app_docs/new_app_tracker.md` as the live implementation-status reference
- do not treat implementation as complete until the required verification has run or an explicit blocker has been documented
- prefer `gopls` for package understanding, symbol discovery, references, diagnostics, and safe refactors when it materially fits

## Build, Test, and Development Commands

Current useful commands:

- `git status --short` to review local changes before editing shared planning files
- `rg --files new_app_docs docs examples` to list the working document set
- `sed -n '1,160p' new_app_docs/README.md` to check the canonical reading order
- `go run ./cmd/migrate` to apply embedded PostgreSQL migrations
- `go build ./cmd/... ./internal/...` to verify the active implementation workspace builds without pulling the read-only `examples/` tree into module verification
- `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` to run the current automated test suite against the configured test database without package-level advisory-lock contention and without pulling the read-only `examples/` tree into module verification
- `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test ./path/to/package -count=1` to run a focused DB-backed package test with the same configured environment when a narrow rerun is needed; do not treat plain `go test ./path/to/package` as the normal path for DB-backed packages in this repository
- `go test -race ./path/to/package` to run targeted race detection for concurrency-sensitive packages; this is not yet part of the repository's standard full-suite verification path
- `go test -shuffle=on ./path/to/package` to detect hidden test-order coupling in focused packages when test isolation is in doubt
- `go test -count=1 ./path/to/package` to disable cached test results for focused reruns and flake investigation
- `git diff --check` to catch whitespace and Markdown formatting issues

## Writing Style & Naming Conventions

Write concise Markdown with clear headings and short paragraphs or numbered rules. Follow the existing lowercase snake-case filename pattern, for example `new_app_execution_plan.md` or `v2_scope_overview.md`. Use date-stamped filenames only when the date is materially part of the record. Keep terminology aligned with the planning set: documents, ledgers, execution context, approvals, reports, thin v1, and v2.

## Engineering Standards

Follow industry-standard best practices by default unless there is a concrete repository-specific, product-specific, or technical reason to deviate. When deviating, make the reason explicit in code, docs, or review notes as appropriate.

Contributors should push back on weaker architectural or implementation choices, guide the user toward best-practice system design by default, and not proceed with a materially weaker path until the downsides and tradeoffs have been made explicit to the user and the user has clearly confirmed that deviation.

During implementation, if a codebase review surfaces drift, an issue, an inconsistency, or a conflict, contributors should report it and either fix it in the same change when appropriate or document it in the canonical implementation plan docs for a future session rather than leaving it as silent drift.

When working primarily in a non-backend layer such as the web UI, browser application flow, or AI-agent layer, contributors should still fix backend bugs, missing support seams, inconsistencies, or narrow capability gaps that materially block correctness, continuity, or usability. Those backend changes should stay within existing ownership boundaries and should remain in service of the active implementation slice rather than becoming unrelated backend feature expansion.

## Architecture & Scope Guardrails

`workflow_app` is intentionally AI-agent-first, database-first, and centered on documents, ledgers, and execution context. Do not let CRM, portal, or broad manual-entry UI concerns become the center of gravity again. If a capability can wait until v2 without weakening the foundation, put it under `new_app_docs/app_v2_plans/` instead of expanding v1. Thin v1 means narrow breadth, not weak modeling or low quality.

Everything meaningful in the system should tie to one or more workflows. Not every component is itself a workflow, but every meaningful feature, state transition, support seam, review surface, and operational control should support, constrain, observe, or expose a workflow. If a proposed capability cannot be tied clearly to one or more workflows, treat it as suspect until that relationship is made explicit in code, docs, or planning material.

For the promoted web layer, prefer a Go-native server-rendered stack by default. Use Go `html/template` plus standard browser behavior as the baseline, prefer `htmx` for progressive enhancement where partial-page updates materially improve operator flow, and use `Alpine.js` only for small local UI state when plain HTML becomes awkward. Avoid introducing a separate Node or SPA toolchain unless the canonical planning docs are explicitly updated to require it.

The promoted web layer and the later mobile client should continue to share one backend foundation and auth model rather than splitting into web-specific versus mobile-specific backends.

Shared foundation entities should have one canonical identity reused across modules. Do not let accounting, inventory, execution, CRM-style support flows, or later features create duplicate module-local truth models when they should reference the same underlying record.

The primary app working model is persist-first and queue-oriented. Inbound requests should be stored durably before AI processing begins, AI processing should usually run asynchronously from that queue, and humans should review the resulting proposals or actions from explicit review surfaces rather than depending on immediate AI response as the default path.

The same persisted-request model should be suitable for both human-originated and system-originated requests so later integrations can use the same controlled intake path without inventing a second processing model.

Persisted inbound requests are request records, not `documents` rows. Their stable `REQ-...` references are request-tracking identifiers, not business-document numbers. When processing produces a draft, approval candidate, or posting candidate, that downstream business record should remain a separate document with its own document lifecycle, approval path, and numbering rules while the original inbound request remains the same request record with its own intake-processing status history.

Persisted inbound requests may remain in `draft` until explicitly queued, and draft requests must not be processed by AI. User-visible removal of parked requests should normally be implemented as soft cancel or soft delete rather than unrestricted hard deletion so auditability and recovery remain intact.

For thin-v1 development and testing, attachment content for inbound requests may be stored in PostgreSQL first, provided the design keeps a clean path to move binary storage to external object storage later. Original uploaded artifacts, including voice recordings, should remain durably available even when derivative records such as transcriptions are created.

As a database-first application, every meaningful workflow and control state should be durably reconstructible from database records. Do not rely on transient process memory or client state for the authoritative record of request intake, AI processing, review, approval, document lifecycle, posting, execution, or failure states that matter to business control or recovery.

It is acceptable to adopt selective OpenClaw-style patterns where they strengthen this architecture, especially durable intake, queue-oriented async processing, modular tool or skill boundaries, and browser-first control surfaces. Do not copy consumer-assistant or autonomy-heavy behavior where it would weaken approvals, posting boundaries, auditability, or database truth.

## Testing & Review Guidelines

For planning-only work, validation is document-focused: check heading structure, cross-file consistency, scope alignment, and broken references. If scope, sequencing, or status changes, update the canonical planning file first and only then update summaries or companion docs. Do not mark tracker items done without concrete evidence in the same change. If you find an inconsistency, resolve it, call it out, or document it explicitly rather than leaving silent drift.

For implementation work:

- every behavior change should include tests appropriate to the change
- workflow-critical changes should not be treated as adequately verified by unit or package tests alone when the real risk is end-to-end operator continuity, control-boundary behavior, approval transitions, or operator-visible state
- for workflow-critical slices, prefer bounded end-to-end review and live testing on the real `/app` plus `/api/...` seam after focused code review and narrow blocker fixes
- keep end-to-end workflow testing bounded by an explicit checklist with pass/fail evidence and blocker tracking; do not rely on broad exploratory manual testing without a documented workflow list and boundary assertions
- when a durable workflow or validation checklist exists in `docs/workflows/`, use it and update it if the implemented workflow support or testing policy has drifted
- if any verification command fails, investigate the cause before proceeding
- do not continue past a failing check without either fixing the issue and rerunning the relevant verification successfully, or documenting the blocker explicitly in the same change
- if a failure is caused by using a non-standard command path for this repository, rerun verification using the documented repository command before treating it as a product defect
- run `go build ./cmd/... ./internal/...` before closing out the task
- run `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` before closing out the task when code or persistence behavior changed
- use `go test -race ./path/to/package` selectively when changing concurrency-sensitive code such as queue claim or processing paths, session or token flows, or other shared in-memory coordination; do not treat full-repo `-race` as the current standard verification path unless the repository guidance is updated explicitly
- use `go test -shuffle=on ./path/to/package` when test-order coupling, shared fixtures, or hidden state leakage is a realistic risk
- use `go test -count=1 ./path/to/package` for focused reruns when you need to bypass cached results or investigate a flaky failure
- database-backed tests in this repository are expected to run with the configured test database loaded from `.env`; do not treat direct `go test` runs without that environment as the normal verification path, even when the tests are not explicitly labeled as integration-only
- when a focused rerun targets a DB-backed package, use the same `.env`-loaded command shape as the canonical suite rather than starting with a bare `go test`
- if DB-backed verification appears hung, check for stale or overlapping sessions holding the disposable test-database advisory lock before treating the symptom as a product defect; document the blocker and cleanup in the canonical planning docs when it materially affects verification
- if a DB-backed verification command fails because the sandbox cannot reach the configured test database, rerun the documented `.env`-loaded repository command with the required approval path before treating the failure as a product defect
- if migrations or persistence behavior change, verify against the configured development and test databases unless an explicit blocker is documented
- while the application remains pre-production, it is acceptable to drop and recreate the configured test database to recover from schema drift, failed migration experiments, or other disposable development-state issues
- the disposable database-reset rule applies only to the configured test database, not to the application or development database

## Commit & Pull Request Guidelines

Current Git history is minimal, so use short imperative commit subjects that describe the actual change, for example `docs: tighten thin-v1 scope rules`. Unless the user says otherwise, commit completed implementation slices after verification and documentation sync so progress is captured in small, reviewable checkpoints. Keep each commit and pull request focused on one slice. PRs should explain the purpose, list the canonical files touched, and note any decision, scope, sequencing, or status change. Avoid mixing unrelated planning edits in one review.

## Security & Configuration Tips

Do not commit `.env`, `.envrc`, logs, coverage output, or generated artifacts; `.gitignore` already excludes them. Keep local configuration rules and any new operational guidance aligned between `AGENTS.md`, the active planning set, and the top-level setup docs as the repository evolves.
