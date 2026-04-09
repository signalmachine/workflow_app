# Documentation

Date: 2026-04-01
Status: Active index
Purpose: explain the role of the `docs/` tree and direct contributors to the correct durable documentation layer.

## 1. Role of this folder

This folder holds durable supporting documentation for `workflow_app`.

It complements the canonical planning set in `new_app_docs/`, but it does not replace it.

Use `docs/` for:

1. technical reference material that should stay useful across implementation sessions
2. workflow-reference and workflow-validation material
3. user-facing guides for supported application behavior
4. companion summary material that explains high-level objectives
5. archived reference material that should stay available without cluttering the active docs surface

## 2. Boundary with `new_app_docs/`

Keep document roles distinct.

`new_app_docs/` remains the canonical planning source for:

1. scope
2. architecture guardrails and implementation defaults
3. execution order
4. active implementation slices
5. implementation status and next steps

`docs/` is the durable supporting layer for:

1. technical implementation guidance
2. workflow-reference and validation material
3. user guides
4. companion objective summaries
5. archived documentation that should remain available as reference-only material

If a planning document and a `docs/` document disagree, use `new_app_docs/` as the source of truth for active implementation status and planned next work, then update the relevant `docs/` material if it has drifted.

## 3. Folder map

1. `technical_guides/`
   Durable technical reference for architecture, boundaries, persistence, AI-agent behavior, and verification guidance.
2. `workflows/`
   Durable workflow catalog, workflow validation checklists, and live-review evidence.
3. `user_guides/`
   Operator-facing and task-oriented guidance for supported application behavior.
4. `implementation_objectives/`
   High-level companion summary material for implementation objectives and principles.
5. `archive/`
   Reference-only historical material that should not be part of the default active working set.

## 4. Suggested starting points

Use these entry points depending on the task:

1. implementation context: `technical_guides/README.md`
2. workflow validation or workflow continuity review: `workflows/README.md`
3. operator-facing usage guidance: `user_guides/README.md`
4. high-level objective summary: `implementation_objectives/implementation_objectives.md`

For active implementation planning, return to `new_app_docs/README.md`.
