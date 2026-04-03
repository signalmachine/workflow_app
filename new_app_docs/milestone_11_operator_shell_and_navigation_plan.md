# workflow_app Milestone 11 Operator Shell and Navigation Plan

Date: 2026-04-03
Status: Implemented historical follow-on milestone for the earlier Go-template browser layer; its information-architecture lessons still matter, but the approved forward browser direction is now the Svelte-based replacement documented in `../docs/svelte_web_guides/svelte_web_ui_migration_plan.md`
Purpose: record the follow-on browser-planning layer after the Milestone 10 rebuild so the promoted web UI could move from a structurally sound modular shell to a calmer SAP-style operator shell; this now serves as historical product-direction context for the Svelte migration rather than the active stack prescription.

## 1. Why this milestone exists

Milestone 10 is the correct architectural rebuild, but the rebuilt shell should not be treated as the final operator-navigation posture.

Implementation review on the new shell shows a clear product-direction question:

1. the current left-side control rail is structurally valid, but it is visually heavy for the current app size
2. the promoted operator-entry routes are still few enough that a persistent left rail can feel more distracting than helpful
3. future route growth should not be solved by adding more always-visible links at equal weight
4. the next navigation layer should prepare for a larger application without turning every page into a dense ERP chrome surface

This milestone exists to choose the stronger follow-on direction:

1. use a lighter top-level navigation model for the true primary destinations
2. group broader route families behind bundle landing pages rather than exposing every route directly in the global shell
3. add searchable route discovery instead of scaling visible chrome linearly with route count
4. turn `/app` from a generic dashboard into a personalized operator home over time
5. define the place of `Settings` and `Admin` explicitly so they do not become accidental navigation clutter

## 2. Planning decision

The current left rail should not be treated as a locked final pattern.

The preferred follow-on direction is:

1. use a top bar as the primary shell, but do not force an immediate hard cap on top-level destinations while the browser surface is still being exposed broadly
2. expose every currently supported workflow through the web UI in some way, either directly on the top bar or through landing pages reached from the top bar
3. use top-level destinations as landing pages for related route bundles where appropriate
4. keep dense review and browse surfaces available, but subordinate many of them under bundle pages, workbench entry points, or search-driven discovery rather than forcing every route into one flat strip
5. reserve a searchable route catalog or command palette as the scaling mechanism for broader navigation
6. evolve `/app` into a personalized home surface with role-aware and user-aware workflow shortcuts rather than a permanently generic dashboard
7. keep other pages standardized by route family and role rather than making the whole application user-customizable page by page

This is intentionally SAP-inspired in information architecture, not in visual heaviness.

The target is not:

1. a large static mega-menu with every route exposed globally
2. a copy of legacy enterprise UI chrome
3. a second frontend architecture
4. a route explosion disguised as convenience

## 3. Architecture stance

This milestone should build on the modular Milestone 10 browser foundation rather than reopen it.

It should preserve:

1. the shared Go backend and auth foundation
2. the earlier rebuilt browser information architecture, shared primitives, and route continuity patterns that remain relevant during migration
3. the same route continuity and shared backend truth
4. the same workflow-first, review-first, approval-aware doctrine

It should change:

1. global shell emphasis
2. top-level navigation structure
3. route-discovery model
4. dashboard and home-surface responsibilities
5. settings and admin entry-point posture
6. visual-theme direction for the promoted shell and landing pages

## 4. Product direction

### 4.1 Global shell

The default shell should move toward a lighter top bar.

Preferred top-level posture at this stage:

1. expose the current major workflow entry points through the top bar
2. allow some workflows to appear directly as top-level pills or bubble-style links while breadth is still being surfaced
3. use landing-page destinations such as `Review`, `Operations`, or `Inventory` where they help organize downstream routes cleanly
4. defer hard reduction of top-level items until the next streamlining pass proves which destinations should stay global

Rule:

1. the global shell should orient the operator, not act as a full site map even when current breadth is more exposed than the long-term target
2. the top bar may temporarily carry more than the long-term target count, but downstream route families should still collapse into landing pages where that improves scanability
3. top-level navigation items should use a pill or bubble-style treatment with clear active-state differentiation so operators can tell which destination is currently active
4. the top-bar bubble set should be allowed to wrap into more than one row when needed rather than forcing a single-row overflow strip
5. later user-level hiding or pinning of some bubble destinations is acceptable, but it should remain additive preference behavior rather than the primary route-discovery mechanism

### 4.1.1 Visual direction

The promoted shell should avoid both stark white and dark heavy backgrounds as the dominant page treatment.

Preferred visual posture:

1. use restrained light backgrounds in a blue-gray family rather than the current green-heavy palette
2. keep typography dark, preferably black or near-black, except where contrast rules require another color
3. use stronger colors primarily for active states, emphasis, status meaning, and bounded visual hierarchy rather than as full-page background blocks
4. keep the overall effect calm, readable, and enterprise-credible without reverting to dark chrome or portal-style promotional framing

Rule:

1. avoid pure white as the dominant page background where a softer low-glare light tone works
2. avoid dark-mode-style full-page backgrounds and heavy dark panels as the main visual identity
3. keep text contrast high and do not use grey-on-grey combinations that weaken readability
4. when background colors vary by section or destination, preserve one coherent palette rather than turning each surface into an unrelated theme
5. avoid letting gradients, hero panels, or card mosaics become the dominant visual language on ordinary operator pages

### 4.2 Bundle landing pages

Some top-level destinations should act as landing pages for related route families.

Examples:

1. `Inventory` as a landing page for stock review, movement review, item review, location review, and later bounded inventory actions
2. `Review` as a landing page for inbound requests, approvals, proposals, documents, accounting, work orders, audit, and later additional review queues
3. `Operations` as a landing page for the operations feed, agent chat, and later bounded queue-driven operational controls

Rule:

1. landing pages should summarize the route family, expose the key next actions, and provide the most common continuity links
2. landing pages should not duplicate every downstream page inline at equal visual weight
3. where possible, landing pages should behave as compact grouped route directories rather than card galleries

### 4.3 Searchable route catalog

The app should gain a searchable route-discovery surface in the near future.

Preferred shape:

1. one searchable command palette or route catalog available from the shell
2. supports search by route title, workflow term, domain term, and common operator intent
3. returns destinations, not arbitrary mutations
4. may later include recent destinations, pinned links, and role-aware suggestions

Guardrail:

1. the first version should remain navigation-oriented rather than becoming a consumer-chat command box or broad action launcher

### 4.4 Personalized home

`/app` should evolve into a personalized operator home.

Preferred responsibilities:

1. role-aware workflow shortcuts
2. user-pinned or org-pinned quick links
3. workload summaries relevant to the current operator
4. bounded reminders for stalled approvals, failed requests, or queue-health issues
5. continuity into the operator's most common workflows
6. future user preferences for showing or hiding selected top-bar bubble destinations if that proves useful

Rule:

1. `/app` should remain workflow-centered and review-centered, not become a generic portal or marketing-style home page
2. `Home` should be the primary user-configurable surface; other pages should stay standard by route and role

### 4.5 Settings and admin surfaces

The app should explicitly distinguish personal settings from tenant or system administration.

Preferred shape:

1. `Settings` is a user-scoped surface for profile, session, notification, and home-page preferences that are safe for the current actor to control
2. `Admin` is a privileged surface for org-scoped or system-scoped configuration, access management, policy configuration, and operational controls
3. both surfaces should vary by access level, but the pages themselves should remain standardized rather than per-user redesigned
4. foundational org-scoped maintenance such as ledger-account, tax-code, accounting-period, customer, and party setup should be planned under `Admin` rather than repurposing `Settings`

Rule:

1. `Settings` and `Admin` should not compete with the primary workflow destinations for top-level attention
2. `Settings` should usually be available from the user or session menu rather than the primary global navigation
3. `Admin` should appear only for actors with the relevant access, and it may live behind the same user or utility menu unless later scale justifies a stronger entry point
4. page variation should come from role-based visibility, sections, and actions rather than from user-specific page layouts outside `Home`

## 5. Scope

In scope for this milestone:

1. replacing the persistent left rail with a lighter top-level shell where justified by browser review
2. introducing top-bar navigation that can expose the current workflow breadth more directly during the transition period
3. defining the first bundle landing pages for grouped domains or workflow families
4. introducing a searchable route catalog or command palette for navigation discovery
5. rebuilding `/app` into a personalized home surface
6. defining the first `Settings` and `Admin` surface posture and their relationship to the global shell
7. defining pill or bubble-style navigation treatment with active-state contrast for the top bar
8. allowing the top-bar bubble set to wrap across multiple rows where needed
9. defining the preferred restrained blue-gray visual palette for shell backgrounds, lines, and landing-page treatment
10. adding the minimum supporting preference model needed for pinned links, role-aware defaults, user-specific shortcuts, or later bubble-visibility preferences if the current shared foundation lacks it

Out of scope for this historical milestone:

1. introducing a SPA or Node-based navigation framework during that milestone phase
2. turning the search surface into a broad write-action launcher on day one
3. exposing every route as permanent global chrome
4. broad CRM-style workspace personalization unrelated to workflows
5. consumer-style dashboard widgets with weak workflow relevance
6. user-specific customization of all page layouts outside the home surface

## 6. Required outcomes

This milestone is complete only when:

1. the global shell is calmer than the current left-rail presentation
2. all currently supported workflows are exposed through the web UI either directly from the top bar or through reachable landing pages
3. top-level navigation items use a pill or bubble-style treatment with clear active-state differentiation
4. the top-bar bubble set remains readable through multi-row wrapping instead of hidden overflow as route breadth grows
5. the shell and landing pages use a restrained blue-gray palette instead of the current green-heavy treatment, stark white, or dark heavy backgrounds as the dominant visual treatment
6. broader route families are discoverable through bundle landing pages and the route catalog rather than through dense undifferentiated chrome alone
7. `/app` behaves as a personalized operator home rather than only a generic shared dashboard
8. the new shell and home remain aligned with the workflow-first doctrine and shared backend truth
9. the resulting navigation model can later be streamlined without another shell rewrite, including future user-level hide or pin preferences if justified
10. `Settings` and `Admin` have an explicit standardized posture consistent with role-based access rather than ad hoc page growth
11. the visual system stays coherent across pages even when light background tones vary by surface
12. the promoted browser layer no longer feels like an editorial or portal surface built from stacked banners and mixed-emphasis cards

## 7. Suggested implementation slices

This milestone should remain bounded by three related slices rather than one open-ended UX bucket.

### 7.1 Slice 1: top-bar shell and primary-destination restructure

Goal:

1. replace the current persistent left rail with a lighter top shell and establish the final primary-destination set

Scope:

1. top bar with pill or bubble-style navigation items
2. shell spacing and page chrome reduction
3. multi-row wrapping behavior for the bubble set
4. restrained blue-gray palette and active-state contrast rules for shell chrome
5. current workflow-entry naming and ordering
6. user-menu or utility-menu placement for `Settings` and access-scoped `Admin`
7. compatibility handling for review and detail pages that still need local secondary navigation

### 7.2 Slice 2: bundle landing pages and route taxonomy

Goal:

1. introduce landing pages that organize related route families cleanly

Scope:

1. `Review` landing page
2. `Operations` landing page
3. first domain landing pages such as `Inventory` only where the route family already justifies it
4. route taxonomy, labels, and bundle descriptions

### 7.3 Slice 3: searchable route catalog and personalized home

Goal:

1. complete the navigation-scaling posture and establish the next-generation operator start surface

Scope:

1. searchable route catalog or command palette
2. role-aware and user-aware `/app` home composition
3. pinned links or quick links
4. user-scoped settings for home-page preferences
5. optional later user preferences for hiding selected top-bar bubbles without changing route truth
6. bounded preference persistence and validation

## 8. Sequence rule

This milestone normally follows Milestone 10 closeout, but the bounded Slice 1 shell shift may be promoted earlier when implementation priority is made explicit and the separate Milestone 10 validation track remains intact.

Sequence:

1. keep Milestone 10 browser-review and workflow-continuity closeout active on the separate `docs/workflows/` track until that evidence is complete
2. when implementation priority is explicit, it is acceptable to promote the bounded Slice 1 shell change ahead of that closeout instead of blocking all browser work behind manual review evidence
3. land the shell change first and expose current workflow breadth cleanly
4. then land bundle landing pages and route taxonomy so broader route families stop depending on flat global chrome alone
5. then land searchable route discovery once the route taxonomy is explicit enough to search cleanly
6. streamline and reduce top-level items only after real usage shows which destinations deserve to remain global
7. land personalization only after the information architecture and route taxonomy are stable enough to personalize safely

Reason:

1. the rebuild foundation should still remain stable, but the shell shift is now bounded enough to proceed without reopening the Milestone 10 modular architecture
2. broad route exposure is acceptable in the short term if landing pages and active-state cues keep it understandable
3. multi-row top-bar wrapping is preferable to horizontal overflow or hidden navigation during the broad-exposure phase
4. search and personalization are stronger once the route taxonomy is explicit
5. landing personalization before route taxonomy stabilizes would create noisy shortcuts and early preference drift
6. the visual palette should be decided early enough that later landing pages inherit one coherent low-glare treatment

Current implementation checkpoint:

1. Slice 1 is now implemented in code on the modular embedded bundle under `internal/app/web_templates`
2. the heavy persistent left rail has been replaced with a lighter top shell using wrapped bubble navigation, a soft-light palette, and a utility session menu that reserves `Settings` plus privileged `Admin` for secondary placement
3. Slice 2 is now also implemented in code: `/app/operations`, `/app/review`, and `/app/inventory` act as bundle landing pages, the top shell now groups route families under those landings, and the narrower direct queue band now keeps only the most frequent upstream review queues globally exposed
4. focused `go test ./internal/app -run '^TestHandleWeb' -count=1`, `go build ./cmd/... ./internal/...`, and `gopls` diagnostics passed after the landing-page slice
5. after switching `TEST_DATABASE_URL` to a fresh local PostgreSQL database, applying migrations to that fresh test DB, and tightening the harness so the disposable advisory lock is held only during setup work, the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...` verification also passed cleanly
6. Slice 3 is now implemented in code: the shell utility path exposes `/app/routes`, `/app/settings`, and access-scoped `/app/admin`, the route catalog provides searchable destination-only discovery on the shared browser seam with tokenized multi-term operator-intent matching instead of one raw substring requirement, and `/app` now uses role-aware workload shortcuts instead of a permanently generic dashboard while keeping preference persistence deferred
7. focused `go test ./internal/app -run '^TestHandleWeb' -count=1`, `go build ./cmd/... ./internal/...`, the canonical `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`, and `gopls` diagnostics all passed after the Slice 3 implementation plus the later route-catalog corrective slice
8. the later bounded `internal/app` transport-boundary cleanup review is now also complete for the remaining role-aware home and agent-chat continuity paths: dashboard recent-request or proposal composition now comes from shared `reporting.DashboardSnapshot`, and agent-chat continuity now comes from shared `reporting.AgentChatSnapshot` instead of route-local multi-read orchestration
9. the separate Milestone 10 browser-review closeout still remains required before the broader browser-validation track can be treated as complete
10. implementation review after Slice 3 found one more bounded browser follow-up need: the promoted shell felt too card-heavy and editorial, and that follow-up is now implemented through `web_ui_erp_style_density_correction_plan.md` rather than opening a new milestone
11. the later density-correction follow-up now also removes the remaining card-like landing and utility-page drift: `/app`, `/app/operations`, `/app/review`, `/app/inventory`, `/app/routes`, `/app/settings`, and `/app/admin` now default to plain hyperlink-first route directories with lightweight workflow counts instead of summary-card or result-card mosaics

## 9. Open design rules

The following decisions are now the default planning posture unless later evidence changes them:

1. prefer top navigation over the current persistent left rail for the global shell
2. prefer full workflow exposure in the near term through a mix of top-level items and landing pages rather than hiding routes too early
3. prefer bundle landing pages over forcing all downstream routes into one flat navigation strip
4. prefer multi-row wrapped top-bar bubbles over horizontally clipped or hidden navigation during the current broad-exposure phase
5. prefer a searchable route catalog over endlessly widening visible navigation as the long-term scaling tool
6. prefer a personalized `/app` home over a permanently generic dashboard
7. prefer soft low-glare light backgrounds with dark readable text over stark white or dark-theme-heavy full-page treatment
8. prefer route discovery and workflow continuity over decorative enterprise chrome
9. prefer `Settings` and `Admin` as secondary utility surfaces rather than primary workflow destinations
10. prefer standardized role-aware pages over broad user-specific customization outside the home surface
11. prefer bounded ongoing refactoring that improves modularity and lowers churn over preserving avoidable structural debt during the active v2 phase

## 10. Documentation sync

When this milestone or any of its slices are promoted:

1. update `new_app_tracker.md` with acceptance, status, and sequencing
2. update `new_app_execution_plan.md` so the milestone order remains explicit
3. update `new_app_implementation_defaults.md` if the navigation or home-surface posture becomes a locked default
4. update `docs/workflows/application_workflow_catalog.md` and validation checklists when bundle landing pages, route catalog behavior, or home-surface workflow continuity materially change
