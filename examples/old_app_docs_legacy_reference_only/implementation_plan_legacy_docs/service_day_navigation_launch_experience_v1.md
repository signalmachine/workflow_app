# service_day Navigation and launch experience v1

Date: 2026-03-16
Status: Legacy reference
Purpose: preserve the later launch/search planning from the broader roadmap as historical reference.

## 1. Intent

The later web client should not default to a module-first experience where users enter CRM, projects, work orders, or accounting through broad landing pages and then navigate through another layer of lists.

The intended direction is:
1. a personalized home page with a bounded number of pinned tiles
2. each tile launches a specific workflow screen, queue, or document-focused screen directly
3. global search is the second primary entry path and should expose both actions and records
4. the same launch model must scale from the early CRM slice into projects, work orders, billing, accounting, workforce, inventory, and later modules

This direction is closer to an activity launcher than to a traditional module switcher.

## 2. Product stance

Rules:
1. the primary user question should be "what do I need to do now?" rather than "which module am I in?"
2. the home page should optimize for repeated real workflows, not for showing every capability equally
3. clicking a tile should open the intended workflow or document screen directly, not a module home page containing more links
4. global search should remain available as the fallback path for less-common actions and records that are not pinned
5. CRM must not become the permanent visual center of gravity once projects, work orders, billing, and accounting exist
6. `work_order` and other execution-critical workflows must be able to sit beside CRM workflows in the same launch model without feeling like a second product

## 3. Core concepts

### 3.1 Launchable activity

A launchable activity is the unit that the home page and global search should expose.

Examples:
1. create lead
2. my tasks
3. open opportunities
4. create estimate
5. new work order
6. draft journals awaiting review

A launchable activity is not the same thing as a bounded context or module.

### 3.2 Direct-entry destination

An activity should resolve to one exact destination type:
1. a create flow
2. a filtered list or queue
3. a document/detail workspace
4. a review or approval screen
5. another explicit workflow screen with one clear purpose

Avoid:
1. tile to module landing page to sub-menu to actual target
2. search result to generic category page when the exact destination is already known

### 3.3 Personalized tile home

Each user should be able to pin a bounded number of tiles to a personal home page.

Rules:
1. pinning is per user, not a global one-layout-for-everyone rule
2. the product may still provide role-aware defaults or suggested starter tiles
3. the pinned home is a launch surface, not the canonical owner of workflow state
4. the home should stay intentionally small enough that search remains valuable

### 3.4 Global search

Global search should become a first-class cross-module capability.

Search should be able to return:
1. actions or workflow launches
2. records or documents
3. user work queues or saved operational views
4. later reports or dashboards where those are truly user-facing destinations

Search results should be:
1. permission-aware
2. tenant-safe
3. ranked for practical operator use rather than raw module taxonomy

## 4. Requirement set

### 4.1 Home experience requirements

The planned home experience should:
1. show pinned launch tiles for the current user
2. allow a user to search for additional launchable activities and pin selected ones
3. let a user open the target workflow directly from a tile
4. support a small number of meaningful defaults when a user has not configured their home yet
5. remain useful across desktop and later mobile-adjacent web or portal surfaces without requiring a different product-navigation philosophy per module

### 4.2 Search requirements

The planned search experience should:
1. search across launchable activities and records, not only CRM entities
2. expose direct-entry results for actions such as create, review, approve, or resume
3. keep record search available for existing and later domain data
4. support future modules without forcing a rewrite of the search contract whenever a new bounded context lands
5. avoid becoming a hidden bypass around domain permissions, approval rules, audit rules, or normal service boundaries

### 4.3 Cross-module requirements

The launch model must support:
1. CRM relationship workflows
2. project and work-order execution workflows
3. billing and accounting review flows
4. workforce, time, and notification-driven follow-up flows
5. later portal-safe or delegated-access surfaces where appropriate

The user should not need to understand the backend bounded-context map in order to find a common task.

## 5. Non-goals and guardrails

Do not interpret this direction as:
1. a requirement to make every possible feature tile-worthy
2. a requirement to replace domain ownership with a generic launcher-owned state model
3. a reason to let navigation metadata bypass normal authorization
4. a reason to build a broad dashboard-first analytics homepage before the core workflow launcher is solid
5. a reason to let module teams add inconsistent ad hoc home cards with no shared launch contract

## 6. Backend and architecture implications

This direction implies several backend-facing requirements even though the current repo is still API-first.

Needed concepts:
1. a launchable activity catalog with stable identifiers
2. permission-aware visibility checks per activity
3. per-user pinned-home persistence
4. a global search contract that can return both action and record results
5. stable route or intent mapping so clients can deep-link into an exact workflow or document destination

Architecture rules:
1. domain modules still own business truth and workflow invariants
2. the launch/search layer should consume module-owned contracts, read models, and permission checks rather than reaching directly into other modules' tables arbitrarily
3. direct-entry navigation should stay aligned with the same domain-service boundaries used by mobile, web, portal, AI tools, and notifications
4. search indexing and ranking should remain derivable from domain truth where possible rather than creating a second hidden state system

## 7. Sequencing posture

This is not a requirement to pause current milestone execution and build a web UI immediately.

Recommended sequencing:
1. keep the current backend and domain slices moving
2. use this document to shape future web-client architecture and backend seams early
3. when the later web client begins, treat the tile home plus global search as the intended default launch experience rather than as a late cosmetic enhancement
4. implement the first launcher/search experience in a way that can absorb projects, work orders, billing, and accounting cleanly instead of being hard-coded around CRM only

## 8. Acceptance direction for the later web client

The later web launch experience should be considered directionally correct when:
1. a signed-in user lands on a personal tile home rather than a module-home menu
2. tiles open exact workflow or document destinations directly
3. search can find both actions and records
4. unpinned workflows remain discoverable through search without forcing module navigation
5. the same model still works once non-CRM modules become prominent
