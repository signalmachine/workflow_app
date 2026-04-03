# workflow_app Milestone 13 Svelte Web Migration Plan

Date: 2026-04-03
Status: Active next-browser milestone, planned for implementation
Purpose: define the next browser implementation milestone that replaces the earlier Go-template `/app` layer with a Svelte-based web application on the same shared Go backend, auth foundation, and workflow truth model.

## 1. Why this milestone exists

The browser direction is now settled.

The repository should not continue investing in the earlier Go-template browser stack as the forward implementation path.

That earlier layer still matters because:

1. it proved the shared `/app` plus `/api/...` seam is operational
2. it validated route families, operator flows, and continuity expectations
3. it provides the migration reference for feature parity and workflow validation

But it should no longer absorb new browser architecture work because:

1. additional shell, template, or styling work there would be high-risk throwaway effort
2. the product now needs stronger browser modularity than the Go-template layer is likely to provide efficiently
3. the Svelte migration guides already define the intended UI and implementation direction, even though some stack details inside them need correction

## 2. Planning decision

Milestone 13 is the next browser implementation milestone.

It should:

1. create the new Svelte browser application in `web/`
2. keep the Go backend, `/api/...` seam, session cookies, and domain-service ownership intact
3. migrate the promoted `/app` route family in bounded slices rather than as one unsafe cutover
4. keep workflow validation and parity judgment tied to the current durable route and workflow expectations

Milestone 13 is not:

1. a backend rewrite
2. a separate browser-only backend
3. a relaxation of workflow, approval, audit, or reporting truth boundaries
4. a redesign excuse for product-scope drift

## 3. Architecture corrections from the source guides

The source guides under `docs/svelte_web_guides/` remain valuable reference material, but the canonical implementation plan in `new_app_docs/` should correct these stack-level points:

1. use SvelteKit rather than a plain Svelte plus `@svelte-spa-router` application
2. use SvelteKit filesystem routing, nested layouts, and route groups rather than a manually maintained hash-router map
3. use SvelteKit SPA mode with `adapter-static` and a fallback page served by Go, rather than hash-based `#/...` routing
4. use SvelteKit navigation primitives such as `goto`, `beforeNavigate`, and layout-driven route protection where appropriate, rather than treating `goto()` as forbidden
5. keep the deployed application as one Go-served surface, but accept the Node-based frontend build toolchain required by SvelteKit during development and build
6. keep Svelte stores limited to truly shared cross-cutting state; prefer Svelte 5 runes and local component state by default

Reason:

1. the Svelte docs support SvelteKit SPA mode cleanly for a frontend served by another backend
2. SvelteKit gives this repository route structure, layouts, and type-safe navigation that would otherwise need to be rebuilt manually
3. the Go backend can still serve one built frontend artifact under `/app` without introducing a long-running Node runtime in production

## 4. Target architecture

The target browser architecture is:

1. `web/` as a SvelteKit + TypeScript project
2. `adapter-static` SPA output with a fallback page
3. Go serving the built frontend under `/app` and static assets under the corresponding asset path
4. same-origin API calls from the Svelte app to `/api/...`
5. existing cookie-based browser auth preserved as the primary browser auth model
6. current `/api/...` JSON seams reused as the business API surface

The target browser architecture is not:

1. SvelteKit SSR with a separate Node server process
2. hash-based SPA routing
3. duplicated workflow state in the browser
4. direct browser ownership of business rules that belong in domain services or reporting reads

## 5. Scope

In scope:

1. scaffold the SvelteKit app in `web/`
2. establish the design-token and shell foundation from the Svelte design guides
3. migrate the promoted route family from the earlier `/app` implementation
4. add any missing additive `/api/...` endpoints required for the Svelte routes
5. add frontend test and build tooling appropriate to the new stack
6. cut over Go serving from the old template layer to the built Svelte frontend once bounded parity is reached
7. retire superseded Go-template browser code after cutover

Out of scope:

1. changing accounting, inventory, execution, approval, or AI truth ownership
2. building a browser-only write path that bypasses shared services
3. introducing Tailwind CSS by default
4. promoting SvelteKit server routes as a second backend
5. broad workflow expansion unrelated to the migration

## 6. Delivery and cutover model

During implementation:

1. the existing Go-served `/app` remains the live reference implementation until the final cutover slice
2. day-to-day Svelte implementation should use the SvelteKit dev server with proxying to the Go backend for `/api/...`
3. the Svelte app should be validated continuously against the same backend seams and workflow expectations as the existing browser layer

Cutover rule:

1. do not switch Go production serving from the template layer to the built Svelte frontend until the core route family and auth flow have bounded parity
2. once cutover begins, prefer one decisive switch plus bounded cleanup rather than a long-lived dual-browser production state

## 7. Suggested slices

This milestone should execute through three large slices:

1. `milestone_13_slice_1_svelte_foundation_and_shell_plan.md`
2. `milestone_13_slice_2_svelte_workflow_surfaces_plan.md`
3. `milestone_13_slice_3_svelte_detail_admin_and_cutover_plan.md`

## 8. Required backend and platform work

This milestone may require additive backend work, but only where it strengthens the same shared seam.

Likely required work:

1. additive JSON snapshot endpoints currently used only by the old Go web layer
2. Go static-file serving and fallback handling for the built Svelte frontend
3. build and developer-workflow updates for the frontend toolchain
4. parity-oriented API response cleanup where the old web layer relied on handler-local data assembly rather than explicit shared API shapes

Guardrail:

1. if a browser screen needs new business branching, prefer adding or tightening a shared reporting or service contract rather than rebuilding that logic in Svelte

## 9. Verification

Before this milestone closes:

1. run the canonical Go verification commands for any backend changes
2. run SvelteKit type-check and test verification for the frontend
3. run bounded end-to-end workflow validation against the real `/app` plus `/api/...` seam after cutover
4. compare the Svelte routes against the earlier browser route family so parity judgment is explicit rather than assumed

## 10. Documentation sync

When this milestone begins or changes:

1. update `new_app_tracker.md` with queue position and slice status
2. update `new_app_execution_plan.md` so Milestone 13 is explicit in milestone order
3. keep `docs/svelte_web_guides/` as reference guides unless an urgent factual correction is required there
4. update `docs/workflows/` when browser validation expectations or route-entry expectations materially change
