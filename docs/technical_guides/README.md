# Technical Guides

Date: 2026-04-01
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

`docs/workflows/` is the canonical workflow-documentation source for supported operator workflows, workflow continuity, and workflow-validation status. Workflow-facing technical guidance in this folder should derive from that source rather than restating independent workflow truth.

These guides are intended to become part of the durable technical context that both humans and AI coding agents can use during implementation, review, debugging, and extension work.

## 3. Reading order

Start with the most important concepts first:

1. [Application architecture overview](./01_application_architecture_overview.md)
2. [Module boundaries and shared truth](./02_module_boundaries_and_shared_truth.md)
3. [Inbound request lifecycle and queue processing](./03_inbound_request_lifecycle.md)
4. [AI agent architecture](./04_ai_agent_architecture.md)
5. [Web and API seams](./05_web_and_api_seams.md)
6. [Identity, sessions, and authentication](./06_identity_session_auth.md)
7. [Testing and verification](./07_testing_and_verification.md)
8. [Document lifecycle and posting boundaries](./08_document_lifecycle_and_posting_boundaries.md)
9. [Workflow approvals and task model](./09_workflow_approvals_and_task_model.md)
10. [Accounting journal, control accounts, and reversals](./10_accounting_journal_control_accounts_and_reversals.md)
11. [Inventory movements and reconciliation](./11_inventory_movements_and_reconciliation.md)
12. [Reporting read model](./12_reporting_read_model.md)
13. [Attachments and derived text](./13_attachments_and_derived_text.md)
14. [Data modeling and database schema](./14_data_modeling_and_database_schema.md)
15. [Production readiness and release checklist](./15_production_readiness_and_release_checklist.md)

## 4. Organization rule

Prefer many focused technical guides over a small number of broad documents.

As this folder grows:

1. split guides by architecture concern, subsystem, or cross-cutting pattern
2. keep each guide narrow enough to stay maintainable as the application evolves
3. use stable terminology aligned with the canonical planning and workflow-reference documents
