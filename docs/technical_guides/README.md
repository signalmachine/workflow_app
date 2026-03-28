# Technical Guides

Date: 2026-03-28
Status: Initialized
Purpose: provide durable technical reference material for humans and AI coding agents working on `workflow_app`, especially architecture, system patterns, and implementation conventions.

## 1. Role of this folder

This folder is for technical implementation guidance that should remain useful beyond a single planning or implementation session.

It should contain multiple focused guides rather than one or two oversized technical documents.

Use it for:

1. application architecture overviews
2. cross-cutting technical patterns
3. module interaction and ownership notes
4. persistence, queue, and workflow control references
5. operational and integration guidance for developers

These guides should be written so they help both:

1. human contributors understand the system and change it safely
2. AI coding agents such as Codex build context quickly and follow the intended architecture and patterns

## 2. Boundary with other documentation

Keep document roles distinct.

`new_app_docs/` remains the canonical planning source for active milestones, execution order, scope, and implementation status.

`docs/implementation_objectives/` remains the high-level companion summary layer for canonical principles and objectives.

This folder should capture stable technical understanding of the application as built, not replace the canonical planning set.

These guides are intended to become part of the durable technical context that both humans and AI coding agents can use during implementation, review, debugging, and extension work.

## 3. Initial content direction

Good starting guides for this folder include:

1. application architecture overview
2. request intake and async processing patterns
3. document, approval, and posting lifecycle patterns
4. browser layer and shared `/app` plus `/api/...` seam guidance
5. testing and verification patterns for workflow-critical changes

## 4. Organization rule

Prefer many focused technical guides over a small number of broad documents.

As this folder grows:

1. split guides by architecture concern, subsystem, or cross-cutting pattern
2. keep each guide narrow enough to stay maintainable as the application evolves
3. use stable terminology aligned with the canonical planning and workflow-reference documents
