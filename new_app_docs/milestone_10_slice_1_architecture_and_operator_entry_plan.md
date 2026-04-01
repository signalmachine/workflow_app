# workflow_app Milestone 10 Slice 1 Plan

Date: 2026-04-01
Status: Proposed pre-implementation slice plan
Purpose: define the first large Milestone 10 implementation slice so the web rebuild starts with structural foundation work plus the shell and operator-entry surfaces, not with scattered page-by-page redesign.

## 1. Slice role

This slice establishes the rebuild foundation and the browser surfaces that define first-use operator flow.

It should land:

1. the modular template and shared-style architecture
2. the rebuilt shell and navigation model
3. the rebuilt sign-in, dashboard, intake, operations-feed, and agent-chat surfaces

It should not yet attempt to rebuild the broad review and detail families.

## 2. Why this slice exists

Milestone 10 fails if implementation begins by touching many downstream pages before the new architecture, shell language, and entry-surface rhythm exist.

This slice exists to:

1. create the new shared UI ownership model first
2. define the shell and navigation language before downstream page families inherit it
3. replace the highest-traffic entry surfaces together so operators do not start on an old shell and then jump into a new one immediately

## 3. In scope

In scope:

1. modular server-rendered template layout and partial structure
2. shared CSS or asset reorganization for shell and primitive reuse
3. shared primitives for shell, navigation, page header, flash messaging, action groups, cards, table wrappers, empty states, and detail blocks
4. rebuilt sign-in surface at `/app/login`
5. rebuilt dashboard at `/app`
6. rebuilt inbound-request intake surface at `/app/submit-inbound-request`
7. rebuilt operations feed at `/app/operations-feed`
8. rebuilt agent chat at `/app/agent-chat`
9. route-compatible migration support so old and new templates can coexist temporarily where needed

Out of scope:

1. broad review-list rebuild
2. detail-page rebuild
3. route-map expansion beyond narrowly justified cleanup
4. backend feature work unrelated to browser continuity or rebuild support

## 4. Required design outcomes

This slice is complete only when:

1. the active shell clearly distinguishes primary workflow destinations from secondary review destinations
2. the dashboard has one primary operator story centered on request intake, queue continuation, and current operational state
3. sign-in, intake, operations-feed, and agent-chat all visibly belong to the same rebuilt application family
4. helper text and repeated chrome are materially reduced on all touched surfaces
5. the resulting shell and entry surfaces are calmer without weakening workflow continuity

## 5. Required implementation outcomes

Required implementation outcomes:

1. pages touched by this slice render from the new modular template structure rather than the legacy monolithic template
2. shared shell and shared primitives are reusable by later review and detail slices without another structural rewrite
3. old and new rendering paths can coexist temporarily without creating duplicate truth logic
4. responsive behavior is intentional on desktop and narrow-width layouts for all touched surfaces

## 6. Suggested implementation order inside the slice

Implement in this order:

1. introduce the new template and shared-style foundation
2. land the rebuilt shell and navigation
3. migrate sign-in and dashboard
4. migrate intake, operations feed, and agent chat
5. clean up slice-local legacy template or style duplication that the migration makes obsolete

## 7. Verification

Before closing this slice:

1. update focused web tests for the new shell plus the rebuilt entry surfaces
2. run `go build ./cmd/... ./internal/...`
3. run `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
4. run `gopls` diagnostics on edited Go files
5. run bounded browser review on desktop and narrow-width layouts for `/app`, `/app/login`, `/app/submit-inbound-request`, `/app/operations-feed`, and `/app/agent-chat`
6. record operator-visible browser-review evidence on the `docs/workflows/` track if this slice changes navigation, page flow, or validation expectations materially

## 8. Stop rule

Stop this slice when:

1. the architecture foundation and all operator-entry surfaces above are migrated
2. the new shell is the active baseline for those surfaces
3. the slice does not widen into review pages or detail pages beyond any narrowly unavoidable compatibility seam
4. no remaining promoted operator-entry route still depends on the legacy shell

If review pages need substantive rebuilding, that work belongs in Slice 2.

## 9. Documentation sync

When this slice lands:

1. update `new_app_tracker.md` with implementation status and verification state
2. update `milestone_10_web_rebuild_plan.md` if actual slice scope or stop rules drift
3. update workflow docs only if operator entry behavior or validation expectations materially change
