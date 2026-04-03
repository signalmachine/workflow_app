# workflow_app Milestone 13 Slice 2 Plan

Date: 2026-04-03
Status: Planned
Purpose: define the second Milestone 13 implementation slice so the high-value workflow surfaces migrate together on top of the Svelte foundation instead of as disconnected page-by-page rewrites.

## 1. Slice role

This slice migrates the core operator workflows that make the browser useful day to day.

It should land one coherent Svelte route family across:

1. operator home and workload entry
2. inbound request intake
3. operations landing and operations feed
4. agent chat
5. review landing and the main review list surfaces
6. route discovery where it materially supports those flows

## 2. Why this slice exists

The migration should not start with edge routes or admin polish.

The highest leverage comes from moving the routes that:

1. authenticate the operator
2. show current workload
3. let operators submit inbound requests
4. let operators review queues and core workflow lists
5. prove that the Svelte shell works against the real shared backend under normal operator flow

## 3. Scope

In scope:

1. `/app`
2. `/app/login`
3. `/app/submit-inbound-request`
4. `/app/operations`
5. `/app/operations-feed`
6. `/app/agent-chat`
7. `/app/review`
8. the promoted review list surfaces and route directories that operators use to begin drill-down work

Support work in scope:

1. additive snapshot or reporting endpoints required for those routes
2. list filters, pagination assumptions, continuity links, and list-table primitives
3. route-state and search-parameter handling aligned with SvelteKit navigation rather than hash-router patterns

Out of scope:

1. full detail-route parity
2. full admin-surface parity
3. final cutover and deletion of the old template layer

## 4. Required outcomes

This slice is complete only when:

1. the main operator starting surfaces work in Svelte against the real backend
2. review lists and intake flows no longer depend on the old Go-template implementation for normal operator use
3. filters, continuity links, and workflow-entry cues behave consistently across the Svelte routes
4. any snapshot API gaps needed by these routes are closed on the shared `/api/...` seam

## 5. Guardrails

1. keep merge, sort, and workflow-summary logic server-side where the current backend already owns it
2. do not move reporting composition into client-only code merely to avoid adding a shared API endpoint
3. keep the review workbench workflow-centered and table-first rather than turning it into decorative dashboarding

## 6. Verification

Before closing this slice:

1. run frontend verification for the migrated routes
2. run bounded live workflow checks for login, home, intake, operations feed, agent chat, and core review-list continuity
3. run canonical Go verification for any new API endpoints or backend changes
