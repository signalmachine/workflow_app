# User Guides

Date: 2026-03-28
Status: Initialized
Purpose: provide operator-facing and reader-friendly guidance for using `workflow_app` through its supported application surfaces.

## 1. Role of this folder

This folder is for durable user-facing guidance.

It should contain multiple focused guides rather than one large catch-all manual.

Use it for:

1. operator onboarding guides
2. task-oriented how-to guides
3. workflow walkthroughs for supported application flows
4. troubleshooting notes for common user-visible issues
5. release-aligned usage notes when behavior changes materially

## 2. Boundary with other documentation

Keep document roles distinct.

`new_app_docs/` remains the canonical planning source for scope, sequencing, status, and next steps.

`docs/workflows/` remains the durable workflow-reference layer for supported workflow definitions and validation checklists.

This folder should translate the supported application behavior into user-consumable guidance without becoming the live implementation plan.

The primary source material for these guides should be the workflow catalog and related durable workflow documents in `docs/workflows/`.

When a supported operator workflow is added or materially changed, review whether one or more focused user guides in this folder should be created or updated.

## 3. Initial content direction

Good starting guides for this folder include:

1. browser operator getting started
2. inbound request draft, queue, amend, and cancel flows
3. review and approval surfaces
4. document and report lookup basics
5. session and sign-in basics for supported clients

Current guides:

1. `running_the_application.md`
2. `browser_sign_in_and_admin_bootstrap.md`

## 4. Organization rule

Prefer many small and clearly named guides over one or two large documents.

As this folder grows:

1. split guides by operator task or workflow
2. keep each guide scoped to one primary user goal or workflow family
3. align guide names and content with the workflow terminology used in `docs/workflows/application_workflow_catalog.md`
