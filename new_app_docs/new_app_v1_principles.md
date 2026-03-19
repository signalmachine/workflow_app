# workflow_app v1 Principles

Date: 2026-03-19
Status: Draft canonical principles
Purpose: define the non-negotiable doctrine for the `workflow_app` thin-v1 application.

## 1. Product identity

`workflow_app` is:

1. an AI-agent-first business operating system
2. a database-first and SQL-first application
3. a documents-plus-ledgers-plus-execution-context system
4. a review-and-report product for humans

`workflow_app` is not:

1. a CRM-first product
2. a broad manual-entry ERP
3. a portal-first or launch-UX-first product
4. a workflow shell without accounting, inventory, and audit truth

## 2. Core doctrine

The application is built around:

1. documents as intent
2. ledgers as truth
3. execution context as operational reality
4. approvals as the human control boundary
5. reports as derived views

## 3. Thin-v1 discipline

Thin v1 means:

1. foundation before breadth
2. no module earns v1 priority unless it strengthens documents, ledgers, execution, approvals, or reports
3. already-common SaaS expectations are not automatic v1 scope
4. operator convenience is secondary to durable correctness

Thin v1 does not mean:

1. simplistic modeling
2. rushed weak schema work
3. low-quality implementation
4. postponing hard foundation problems to v2

## 4. AI rule

AI is the main operator interface, but AI is not the authority over truth.

Rules:

1. AI may read, summarize, draft, recommend, and request approval
2. AI may execute bounded writes only through explicit tools and normal domain services
3. AI may never write ledger truth directly
4. meaningful writes must remain auditable
5. financially meaningful writes must remain human-gated
6. v1 still uses a multi-agent architecture, but only with bounded coordinator-to-specialist routing and durable observability
7. advanced agent autonomy, speculative delegation depth, and broad self-directed workflow expansion belong in v2 unless required by a foundation invariant

## 5. Human-interface rule

Human UI in v1 stays intentionally minimal.

Allowed primary human surfaces:

1. approval queue
2. review views
3. inspection views
4. reporting views

Not part of v1:

1. broad operational data-entry UI
2. direct ledger editing
3. full CRM workspace
4. broad project-management UI

## 6. Reset warning

The previous codebase showed that CRM-heavy implementation creates planning drag.

`workflow_app` must therefore:

1. exclude CRM as a primary module
2. treat party and contact data as support records only
3. refuse sales-workflow expansion unless a true foundation dependency requires it
