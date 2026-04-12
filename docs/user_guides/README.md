# User Guides

Date: 2026-04-12
Status: Active
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

`docs/workflows/` remains the durable workflow-reference layer for supported workflow definitions and validation checklists, and it is the canonical workflow-documentation source for this folder.

This folder should translate the supported application behavior into user-consumable guidance without becoming the live implementation plan.

The primary source material for these guides should be the workflow catalog and related durable workflow documents in `docs/workflows/`.

Workflow-doc source rule:

1. define or update workflow truth in `docs/workflows/` first
2. then translate that approved workflow truth into operator-facing guidance here
3. do not let user guides become an independent source of workflow truth

When a supported operator workflow is added or materially changed, review whether one or more focused user guides in this folder should be created or updated.

## 3. Initial content direction

Good starting guides for this folder include:

1. browser operator getting started
2. inbound request draft, queue, amend, and cancel flows
3. review and approval surfaces
4. document lookup and report lookup surfaces
5. session and sign-in basics for supported clients

The supported review guides should cover both list-level filters and exact detail pages for the workflows already exposed in the browser.

Recommended reading order:

1. `01_running_the_application.md`
2. `02_browser_sign_in_and_admin_bootstrap.md`
3. `03_browser_operator_getting_started.md`
4. `04_user_testing_readiness.md`
5. `05_inbound_request_lifecycle.md`
6. `06_failed_processing_visibility.md`
7. `07_processed_proposal_review.md`
8. `08_request_approval_from_processed_proposal.md`
9. `09_approval_decision_workflow.md`
10. `10_operations_feed.md`
11. `11_agent_chat.md`
12. `12_admin_accounting_setup.md`
13. `13_admin_party_setup.md`
14. `14_admin_access_maintenance.md`
15. `15_admin_inventory_setup.md`
16. `16_document_review.md`
17. `17_accounting_review.md`
18. `18_inventory_review.md`
19. `19_work_order_review.md`
20. `20_audit_lookup.md`

## 3.1 Workflow coverage map

Use this map when checking whether the user-guide layer still reflects the workflow catalog.

1. browser session login and active-session continuity: `01_running_the_application.md`, `02_browser_sign_in_and_admin_bootstrap.md`, `03_browser_operator_getting_started.md`
2. inbound request submit and queue processing: `05_inbound_request_lifecycle.md`
3. draft save, amend, queue, cancel, and delete lifecycle: `05_inbound_request_lifecycle.md`
4. processed proposal review and continuity: `07_processed_proposal_review.md`
5. processed proposal to approval request: `08_request_approval_from_processed_proposal.md`
6. approval decision and downstream continuity: `09_approval_decision_workflow.md`
7. failed provider or failed processing visibility: `06_failed_processing_visibility.md`
8. downstream document review: `16_document_review.md`
9. accounting report and journal review: `17_accounting_review.md`
10. inventory review: `18_inventory_review.md`
11. work-order review: `19_work_order_review.md`
12. audit lookup: `20_audit_lookup.md`
13. operations feed and agent-chat continuity: `10_operations_feed.md`, `11_agent_chat.md`
14. admin accounting setup maintenance: `12_admin_accounting_setup.md`
15. admin party setup maintenance: `13_admin_party_setup.md`
16. admin access maintenance: `14_admin_access_maintenance.md`
17. admin inventory setup maintenance: `15_admin_inventory_setup.md`

Each workflow guide should include at least one concrete example. Use examples as operator guidance only; keep canonical workflow state and validation status in `docs/workflows/`.

## 4. Organization rule

Prefer many small and clearly named guides over one or two large documents.

As this folder grows:

1. split guides by operator task or workflow
2. keep each guide scoped to one primary user goal or workflow family
3. align guide names and content with the workflow terminology used in `docs/workflows/application_workflow_catalog.md`
