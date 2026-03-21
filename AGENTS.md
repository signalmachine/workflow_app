# Repository Guidelines

## Project Structure & Module Organization

This repository is the active `workflow_app` implementation codebase, with the canonical planning set living in `new_app_docs/`. Start with `new_app_docs/README.md`, then use `new_app_tracker.md` for current status and next steps. Core doctrine, scope, defaults, and execution order are defined in `new_app_v1_principles.md`, `new_app_v1_scope.md`, `new_app_implementation_defaults.md`, and `new_app_execution_plan.md`. Use `new_app_docs/app_v2_plans/` only for explicitly deferred v2 work. `docs/implementation_objectives/implementation_objectives.md` is a companion summary, not a replacement for the canonical planning set. Treat `docs/implementation_objectives/implementation_principles.md` and `examples/old_app_docs_legacy_reference_only/` as reference only.

## Document Hygiene

Keep `AGENTS.md` short and durable. Put repository-wide contributor rules here, and move detailed plans, review notes, or session-specific material into the appropriate document under `new_app_docs/` or `docs/`. After every implementation change, review the canonical docs in `new_app_docs/` and update status, completed work, and next steps in the same change whenever they have drifted. When high-level rules, principles, scope boundaries, or invariants change in `AGENTS.md`, `new_app_docs/`, or `README` files, review `docs/implementation_objectives/implementation_objectives.md` and update it if the summary has drifted. `docs/implementation_objectives/implementation_principles.md` is reference-only and does not need maintenance sync when canonical docs change. After implementation changes, update `README.md` when setup, commands, architecture shape, or user-visible capabilities have changed.

## MCP Usage

Use MCP tools when they materially improve accuracy or speed.

For every Go coding session in this repository once the Go workspace exists:

- `gopls` MCP is required and should be the default path through the session
- start with Go workspace context
- use `gopls` for workspace summary, symbol search, package context, references, safe renames, and diagnostics whenever it materially fits the task
- if a session includes Go code changes, run diagnostics on edited files before completion
- use vulnerability checks when dependencies or security-sensitive code change

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
- `go build ./...` to verify the current workspace builds
- `set -a; source .env; set +a; go test -p 1 ./...` to run the current automated test suite against the configured test database without package-level advisory-lock contention
- `git diff --check` to catch whitespace and Markdown formatting issues

## Writing Style & Naming Conventions

Write concise Markdown with clear headings and short paragraphs or numbered rules. Follow the existing lowercase snake-case filename pattern, for example `new_app_execution_plan.md` or `v2_scope_overview.md`. Use date-stamped filenames only when the date is materially part of the record. Keep terminology aligned with the planning set: documents, ledgers, execution context, approvals, reports, thin v1, and v2.

## Engineering Standards

Follow industry-standard best practices by default unless there is a concrete repository-specific, product-specific, or technical reason to deviate. When deviating, make the reason explicit in code, docs, or review notes as appropriate.

## Architecture & Scope Guardrails

`workflow_app` is intentionally AI-agent-first, database-first, and centered on documents, ledgers, and execution context. Do not let CRM, portal, or broad manual-entry UI concerns become the center of gravity again. If a capability can wait until v2 without weakening the foundation, put it under `new_app_docs/app_v2_plans/` instead of expanding v1. Thin v1 means narrow breadth, not weak modeling or low quality.

Shared foundation entities should have one canonical identity reused across modules. Do not let accounting, inventory, execution, CRM-style support flows, or later features create duplicate module-local truth models when they should reference the same underlying record.

The primary app working model is persist-first and queue-oriented. Inbound requests should be stored durably before AI processing begins, AI processing should usually run asynchronously from that queue, and humans should review the resulting proposals or actions from explicit review surfaces rather than depending on immediate AI response as the default path.

The same persisted-request model should be suitable for both human-originated and system-originated requests so later integrations can use the same controlled intake path without inventing a second processing model.

As a database-first application, every meaningful workflow and control state should be durably reconstructible from database records. Do not rely on transient process memory or client state for the authoritative record of request intake, AI processing, review, approval, document lifecycle, posting, execution, or failure states that matter to business control or recovery.

It is acceptable to adopt selective OpenClaw-style patterns where they strengthen this architecture, especially durable intake, queue-oriented async processing, modular tool or skill boundaries, and browser-first control surfaces. Do not copy consumer-assistant or autonomy-heavy behavior where it would weaken approvals, posting boundaries, auditability, or database truth.

## Testing & Review Guidelines

For planning-only work, validation is document-focused: check heading structure, cross-file consistency, scope alignment, and broken references. If scope, sequencing, or status changes, update the canonical planning file first and only then update summaries or companion docs. Do not mark tracker items done without concrete evidence in the same change. If you find an inconsistency, resolve it, call it out, or document it explicitly rather than leaving silent drift.

For implementation work:

- every behavior change should include tests appropriate to the change
- run `go build ./...` before closing out the task
- run `set -a; source .env; set +a; go test -p 1 ./...` before closing out the task when code or persistence behavior changed
- if migrations or persistence behavior change, verify against the configured development and test databases unless an explicit blocker is documented
- while the application remains pre-production, it is acceptable to drop and recreate the configured test database to recover from schema drift, failed migration experiments, or other disposable development-state issues

## Commit & Pull Request Guidelines

Current Git history is minimal, so use short imperative commit subjects that describe the actual change, for example `docs: tighten thin-v1 scope rules`. Unless the user says otherwise, commit completed implementation slices after verification and documentation sync so progress is captured in small, reviewable checkpoints. Keep each commit and pull request focused on one slice. PRs should explain the purpose, list the canonical files touched, and note any decision, scope, sequencing, or status change. Avoid mixing unrelated planning edits in one review.

## Security & Configuration Tips

Do not commit `.env`, `.envrc`, logs, coverage output, or generated artifacts; `.gitignore` already excludes them. Keep local configuration rules and any new operational guidance aligned between `AGENTS.md`, the active planning set, and the top-level setup docs as the repository evolves.
