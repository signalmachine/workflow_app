# workflow_app Milestone 13 Slice 1 Plan

Date: 2026-04-03
Status: Implemented in code on 2026-04-03
Purpose: define the first Milestone 13 implementation slice so the Svelte migration starts with the correct app foundation, route structure, auth bootstrap, shell, and shared UI primitives rather than with scattered page rewrites.

## 1. Slice role

This slice establishes the Svelte application foundation.

It should land:

1. the SvelteKit project scaffold in `web/`
2. the build, dev, and Go-serving integration needed for the new frontend
3. the root shell, navigation structure, and shared design-token foundation
4. session bootstrap, login surface, and protected-route posture
5. the first route skeletons needed for later workflow migration

It should not yet attempt full workflow-page parity.

## 2. Architecture decision for this slice

The active app structure should be:

1. SvelteKit with `src/routes`
2. root `+layout.svelte` and nested app/auth layouts
3. SPA-mode output with Go fallback serving under `/app`
4. SvelteKit navigation and route groups rather than `@svelte-spa-router`

This slice should explicitly avoid:

1. hash-based routing
2. a manual router registry as the primary route-definition mechanism
3. component patterns that fight Svelte 5 runes or SvelteKit layouts

## 3. Scope

In scope:

1. `web/package.json`, `svelte.config.*`, `vite.config.*`, TypeScript, and test config
2. base path and asset-path decisions needed for serving under `/app`
3. frontend dev proxying to the Go backend
4. root route structure and placeholder route files for the promoted route families
5. global app styles, design tokens, typography, and layout primitives from the design guide
6. app shell including primary navigation, sidebar behavior, top bar, user menu, and mobile-nav treatment
7. session bootstrap flow, login page, logout flow, and unauthorized redirect handling
8. API client foundation, shared error handling, and cross-cutting state such as session and toast or flash behavior

Out of scope:

1. broad review, detail, intake, admin, or inventory page parity
2. deleting the old Go-template `/app` implementation
3. broad backend API additions unrelated to foundation and auth

## 4. Required outcomes

This slice is complete only when:

1. `web/` is a working SvelteKit application with Svelte 5 runes and TypeScript
2. the app can start locally against the existing Go backend
3. session bootstrap and login behavior work against the current cookie-auth seam
4. the root shell reflects the approved sidebar-led design direction from the Svelte guides
5. shared UI primitives exist for later list, detail, form, and feedback pages
6. the repository has a repeatable frontend build path that can later be served by Go

## 5. Suggested implementation order

1. scaffold SvelteKit and lock the base configuration
2. add adapter-static SPA-mode build output and Go-serving assumptions
3. establish the root route tree and layout structure
4. add API client and session bootstrap
5. build login and authenticated app shell
6. add shared primitives needed by later pages

## 6. Verification

Before closing this slice:

1. run SvelteKit type-check and frontend tests for the scaffolded code
2. verify login, logout, and protected-route behavior locally against the Go backend
3. verify the frontend build output shape matches the planned Go-serving integration
4. run canonical Go verification if backend or serving integration changes landed in the same slice

## 7. Implementation result

This slice is now implemented in code.

Landed result:

1. `web/` now contains a SvelteKit + Svelte 5 + TypeScript application scaffold with route groups for protected app routes and the public login route
2. the shared frontend foundation now includes design-token CSS, a sidebar-plus-topbar app shell, shared page-header and feedback primitives, placeholder route surfaces for the promoted browser families, and a minimal API client plus session bootstrap path against `/api/session`
3. `internal/app` now includes embedded static SPA serving for the built frontend, establishing the promoted Svelte shell path under `/app`
4. `npm --prefix web run check`, `npm --prefix web run test`, `npm --prefix web run build`, focused `go test ./internal/app -run 'Test(RegisterWebRoutesSvelteModeServesSPAFallback|HandleSvelteAppServesIndexAtAppRoot|RenderWebPageRejectsUnmappedTemplateData)' -count=1`, `go build ./cmd/... ./internal/...`, and `gopls` diagnostics on the edited Go files all completed cleanly for this slice
5. the next promoted browser implementation step is now Milestone 13 Slice 2 workflow-surface migration rather than additional shell or scaffold work on the old Go-template stack
