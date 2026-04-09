# Repository Guidelines

## Project Structure & Module Organization

This repository is the active `workflow_app` implementation codebase.

The canonical planning set lives in `new_app_docs/`. Start with `new_app_docs/README.md`, then use `new_app_docs/new_app_tracker_v2.md` for current status and next implementation steps. The active thin v2 control docs are:

1. `new_app_docs/new_app_scope_v2.md`
2. `new_app_docs/new_app_architecture_v2.md`
3. `new_app_docs/new_app_implementation_defaults_v2.md`
4. `new_app_docs/new_app_execution_plan_v2.md`

Treat these as reference-only by default unless the task specifically needs them:

1. `new_app_docs/thin_v1_archive/`
2. `new_app_docs/app_v2_plans/`
3. `new_app_docs/v2_archive/`
4. `docs/implementation_objectives/implementation_principles.md`
5. everything under `examples/`

Use `docs/workflows/` for durable workflow-reference material, reusable validation checklists, and live-review evidence. It is not the canonical implementation-planning surface, but it is the canonical workflow-documentation source for downstream user guides and workflow-facing technical documentation.

`docs/implementation_objectives/implementation_objectives.md` is a companion summary, not a replacement for the canonical planning set.

## Document Hygiene

Keep `AGENTS.md` short and durable. Put repository-wide contributor rules here, and move detailed plans, troubleshooting notes, and session-specific material into the appropriate document under `new_app_docs/` or `docs/`.

After every implementation change:

1. review the canonical active docs in top-level `new_app_docs/` and update status, completed work, and next steps when they have drifted
2. review `docs/workflows/` when user-visible workflow behavior, durable workflow status, or reusable live-validation checklists changed materially
3. review `docs/implementation_objectives/implementation_objectives.md` when high-level rules, principles, scope boundaries, or invariants changed materially
4. update `README.md` when setup, commands, architecture shape, or user-visible capabilities changed materially

Keep implementation planning in the active top-level `new_app_docs/` surface. Keep workflow validation and evidence in `docs/workflows/`. Do not silently mix those tracks.

Workflow-doc source rule:

1. when supported operator workflows, workflow continuity, or workflow validation status change, update `docs/workflows/` first
2. then update `docs/user_guides/` and any workflow-facing `docs/technical_guides/` material that derives from that workflow truth
3. do not let downstream guides become independent sources of workflow truth

## Go Session Rules

For every Go implementation session:

1. start with Go workspace context plus the relevant canonical docs in `new_app_docs/`
2. use `new_app_docs/new_app_tracker_v2.md` as the live implementation-status reference
3. use `gopls` MCP as the default path for workspace summary, symbol discovery, references, diagnostics, and safe refactors
4. use the dedicated `mcp__gopls__...` tools such as `mcp__gopls__go_workspace` rather than assuming `workspace://...` resources
5. run diagnostics on edited Go files before completion
6. use vulnerability checks when dependencies or security-sensitive code change
7. when implementing or verifying `internal/ai` against the OpenAI Go SDK, prefer official OpenAI docs and the official `openai/openai-go` repository via MCP or approved web lookup for exact SDK and API details
8. do not treat implementation as complete until the required verification has run or an explicit blocker has been documented

For planning-only or Markdown-only sessions, do not force MCP usage when local document reads are sufficient.

`docs/technical_guides/07_testing_and_verification.md` is the canonical source for exact verification command shapes and verification workflow requirements.

Prefer a local disposable PostgreSQL instance for `TEST_DATABASE_URL` during DB-backed verification. If the serialized suite appears hung, inspect and clean up stale advisory-lock holder sessions on the disposable test DB before treating the symptom as a product defect.

## Svelte Session Rules

When working on the promoted Svelte-based web application and its remaining migration or follow-on work:

1. use the Svelte MCP server as the default documentation source for Svelte 5 and SvelteKit work
2. start Svelte or SvelteKit research with `mcp__svelte__list_sections`
3. analyze the returned section metadata, especially `use_cases`, then fetch all relevant sections with `mcp__svelte__get_documentation`
4. when writing Svelte code, run `mcp__svelte__svelte_autofixer` before presenting the code and continue until it reports no issues or suggestions
5. ask the user before generating a playground link, and use `mcp__svelte__playground_link` only after user confirmation
6. do not generate a playground link when the code has been written directly into repository files

## Writing Style & Naming Conventions

Write concise Markdown with clear headings and short paragraphs or numbered rules. Follow the existing lowercase snake_case filename pattern, for example `new_app_execution_plan_v2.md` or `v2_scope_overview.md`. Use date-stamped filenames only when the date is materially part of the record. Keep terminology aligned with the planning set: documents, ledgers, execution context, approvals, reports, completed thin v1, and active v2.

Keep milestone, slice, and checkpoint labels in planning docs, tracker rows, commits, and review notes rather than in long-lived production source filenames or exported identifiers. Name code by owned domain responsibility, route family, or technical role instead of implementation phase labels.

## Engineering Standards

Follow industry-standard best practices by default unless there is a concrete repository-specific, product-specific, or technical reason to deviate. When deviating, make the reason explicit in code, docs, or review notes as appropriate.

Contributors should push back on weaker architectural or implementation choices, guide the user toward best-practice system design by default, and not proceed with a materially weaker path until the downsides and tradeoffs have been made explicit to the user and the user has clearly confirmed that deviation.

During implementation, if a codebase review surfaces drift, an issue, an inconsistency, or a conflict, contributors should report it and either fix it in the same change when appropriate or document it in the canonical implementation plan docs for a future session rather than leaving it as silent drift.

When working primarily in a non-backend layer such as the web UI, browser application flow, or AI-agent layer, contributors should still fix backend bugs, missing support seams, inconsistencies, or underbuilt architecture that materially block correctness, continuity, maintainability, or usability. Those backend changes should stay within coherent ownership boundaries, but contributors should not preserve weak code merely because it predates the current slice.

The active implementation posture is now ambitious and best-practice-driven rather than thin-scope-driven:

1. thin-v1 is complete and is no longer the limiting delivery posture for active implementation choices
2. contributors are not operating under artificial thin-phase limits and may implement stronger patterns, broader capability, and deeper refactors when those changes materially improve correctness, operability, maintainability, performance, or production readiness
3. `go wild` means freedom to apply proven engineering and business-software best practices without artificial narrowness
4. `go wild` does not mean novelty for its own sake, experimental architecture without justification, or changes that weaken auditability, workflow control, or shared-backend truth
5. contributors should continuously review the codebase for opportunities to refactor, simplify, modularize, or rebuild weak areas, and should either perform that work in the current change when appropriate or promote an explicit plan for it in the canonical docs
6. when a code area is visibly underbuilt, structurally weak, or misaligned with established best practices, contributors should prefer correcting or rebuilding it over layering more work on top of it
7. large monolithic files, `God` files, and concentrated orchestration hotspots should be treated as active refactor targets; contributors should break them into clearer ownership slices, and when the work is too large for the current change they should record a canonical refactor plan rather than leaving the debt implicit

## Architecture & Scope Guardrails

`workflow_app` is intentionally AI-agent-first, database-first, and centered on documents, ledgers, and execution context. Do not let CRM, portal, or broad manual-entry UI concerns become the center of gravity again. Thin v1 is complete; active work should now optimize for the strongest production-shape implementation that still preserves the workflow-centered doctrine, shared truth model, and approval or audit boundaries.

Everything meaningful in the system should tie to one or more workflows. Not every component is itself a workflow, but every meaningful feature, state transition, support seam, review surface, and operational control should support, constrain, observe, or expose a workflow. If a proposed capability cannot be tied clearly to one or more workflows, treat it as suspect until that relationship is made explicit in code, docs, or planning material.

For the promoted web layer, the active direction is now a Svelte-based web application served on the same shared Go backend and auth model. Treat Svelte as the display and interaction layer, keep business logic, workflow state, and business truth on the shared Go backend seams, use the approved Svelte toolchain rather than extending the old Go `html/template` layer as the forward path, and avoid Tailwind CSS by default unless the canonical planning docs are explicitly updated again.

The promoted web layer and the later mobile client should continue to share one backend foundation and auth model rather than splitting into web-specific versus mobile-specific backends.

The promoted web UI should now be treated as desktop-first. Keep the web surface strong for desktop operator use, avoid spending active implementation effort on mobile-web optimization unless a shared correctness issue requires it, and treat future mobile-specific UX depth as the responsibility of the separate mobile client rather than the served web runtime.

Shared foundation entities should have one canonical identity reused across modules. Do not let accounting, inventory, execution, CRM-style support flows, or later features create duplicate module-local truth models when they should reference the same underlying record.

The primary app working model is persist-first and queue-oriented. Inbound requests should be stored durably before AI processing begins, AI processing should usually run asynchronously from that queue, and humans should review resulting proposals or actions from explicit review surfaces rather than depending on immediate AI response as the default path.

Keep these product-model invariants intact:

1. the same persisted-request model must support both human-originated and system-originated requests
2. request records are distinct from downstream business documents, and `REQ-...` identifiers are request-tracking identifiers rather than business-document numbers
3. draft requests must not be processed by AI, and user-visible removal should normally be soft cancel or soft delete rather than unrestricted hard deletion
4. attachment handling may start in PostgreSQL for v1 or early v2, but must preserve a clean path to external object storage and retain original uploaded artifacts
5. every meaningful workflow and control state must be durably reconstructible from database records rather than transient process or client state

Use the technical guides for the detailed system shape:

1. `docs/technical_guides/03_inbound_request_lifecycle.md`
2. `docs/technical_guides/08_document_lifecycle_and_posting_boundaries.md`
3. `docs/technical_guides/13_attachments_and_derived_text.md`
4. `docs/technical_guides/14_data_modeling_and_database_schema.md`

It is acceptable to adopt selective OpenClaw-style patterns where they strengthen this architecture, especially durable intake, queue-oriented async processing, modular tool or skill boundaries, and browser-first control surfaces. Do not copy consumer-assistant or autonomy-heavy behavior where it would weaken approvals, posting boundaries, auditability, or database truth.

## Testing & Review Guidelines

For planning-only work, validation is document-focused: check heading structure, cross-file consistency, scope alignment, and broken references. If scope, sequencing, or status changes, update the canonical planning file first and only then update summaries or companion docs. Do not mark tracker items done without concrete evidence in the same change. If you find an inconsistency, resolve it, call it out, or document it explicitly rather than leaving silent drift.

For implementation work:

1. every behavior change should include tests appropriate to the change
2. workflow-critical changes are not adequately verified by unit or package tests alone when the real risk is end-to-end operator continuity, control-boundary behavior, approval transitions, or operator-visible state
3. for workflow-critical changes, prefer end-to-end review and live testing on the real `/app` plus `/api/...` seam after focused code review and the fixes needed to reach a defensible production-quality result
4. keep end-to-end workflow testing bounded by an explicit checklist with pass or fail evidence and blocker tracking
5. when a durable workflow or validation checklist exists in `docs/workflows/`, use it and update it if implemented workflow support or testing policy has drifted
6. if any verification command fails, investigate the cause before proceeding
7. do not continue past a failing check without either fixing the issue and rerunning the relevant verification successfully, or documenting the blocker explicitly in the same change
8. run the repository's documented verification commands before closing out implementation work

`docs/technical_guides/07_testing_and_verification.md` is the canonical source for:

1. canonical verification command lines
2. focused rerun patterns such as `-race`, `-shuffle`, and `-count=1`
3. DB-backed test environment requirements
4. lock-contention, sandbox, and disposable test-database troubleshooting rules

## Commit & Pull Request Guidelines

Current Git history is minimal, so use short imperative commit subjects that describe the actual change, for example `docs: update v2 implementation posture`. Unless the user says otherwise, commit completed implementation checkpoints after verification and documentation sync so progress is captured in reviewable units. Keep each commit and pull request focused on one coherent concern. PRs should explain the purpose, list the canonical files touched, and note any decision, scope, sequencing, refactor, or posture change. Avoid mixing unrelated planning edits in one review.

## Security & Configuration Tips

Do not commit `.env`, `.envrc`, logs, coverage output, or generated artifacts; `.gitignore` already excludes them. Keep local configuration rules and any new operational guidance aligned between `AGENTS.md`, the active planning set, and the top-level setup docs as the repository evolves.
