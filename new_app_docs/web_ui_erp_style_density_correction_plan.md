# workflow_app Web UI ERP-Style Density Correction Plan

Date: 2026-04-03
Status: Implemented historical corrective slice for the earlier Go-template browser layer; superseded as forward stack guidance by `../docs/svelte_web_guides/svelte_web_ui_migration_plan.md`
Purpose: record one focused corrective slice from the earlier browser layer so the visual and density lessons remain available to the Svelte migration without leaving this document positioned as the active implementation plan.

## 1. Why this slice exists

Implementation review on the promoted browser layer found a real presentation problem:

1. the current shell and landing pages still feel too much like promotional or editorial surfaces instead of operator work surfaces
2. repeated hero cards, summary cards, explanatory copy blocks, and similarly weighted panels make pages read more like a news or portal layout than an ERP-style application
3. the current soft-green palette weakens the intended enterprise feel and does not match the preferred visual direction raised during review
4. landing pages such as `Inventory` still spend too much space on card framing instead of acting as compact route directories into related workflow activities
5. the login and session entry posture is still too large and presentational for the desired thin operator shell

This slice exists to correct the promoted browser presentation before more browser work proceeds.

## 2. Planning decision

This should not open a new milestone yet.

It should be treated as one bounded corrective slice under the current browser work because:

1. the problem is presentation quality and operator density on already-promoted routes
2. the required changes stay within the existing modular web bundle and route set
3. the slice changes shell hierarchy, visual system, and page composition rather than product capability or backend ownership

If this slice later exposes a broader second-phase browser program, that later work can be promoted explicitly with evidence.

## 3. Objective

Move the promoted browser layer toward an ERPNext-like operator feel without copying another product directly.

That means:

1. thin top application bar
2. restrained blue-gray theme instead of the current green-heavy palette
3. compact route-oriented landing pages
4. much less card framing
5. much less explanatory copy above the work surface
6. tables, filters, exact links, and workflow actions visible earlier on the page
7. no return to a left-side navigation panel
8. single-column stacked workflow pages rather than two-column page composition
9. dedicated workflow pages should behave like calm start screens for that one workflow, not mini dashboards

## 4. Design stance

Use the existing `accounting-agent-app` web layer only as directional reference, not as a template to copy.

The target is:

1. denser operator framing
2. calmer enterprise hierarchy
3. clearer distinction between shell chrome and page content
4. stronger ERP-style practicality

The target is not:

1. pixel or component copying from ERPNext or the reference repo
2. a branded redesign exercise
3. a card-gallery dashboard
4. another broad “visual refresh” without hierarchy correction

## 5. Scope

In scope:

1. replace the current oversized shell header with a thin top bar
2. show the app name as `Workflow App` on the left side of the top bar
3. move session access to a compact round user or login control at the top-right
4. shift the shared visual system to a restrained blue-gray palette
5. remove hero-first composition from the promoted dashboard, landing pages, route catalog, settings, and admin surfaces
6. replace card-heavy landing-page composition with compact grouped link lists and small supporting summaries where materially useful
7. reduce or remove promotional or explanatory copy that competes with workflow actions
8. simplify the sign-in surface so it matches the thinner shell language
9. preserve existing route continuity, backend contracts, and page ownership boundaries
10. standardize dedicated activity and workflow pages around a single-column stacked layout instead of two-column compositions
11. keep dedicated workflow pages focused on the primary start action and the minimum supporting context for that workflow
12. where a landing or utility surface only needs navigation, prefer plain hyperlink-first route directories over summary cards, result cards, or button mosaics

Out of scope:

1. backend workflow changes
2. new route families
3. browser personalization expansion
4. a separate frontend toolchain
5. client-side widget systems or dashboard gadgets
6. reintroducing a left-side navigation rail or side panel as the primary shell

## 6. Required outcomes

This slice is complete only when:

1. the shell no longer feels like a stacked banner system
2. the dominant theme is blue-gray rather than green
3. `Inventory`, `Review`, `Operations`, and similar landing pages behave as route directories first
4. major workflow pages show the work surface earlier with less decorative framing
5. login uses the same restrained application language instead of a two-column promotional layout
6. the result feels closer to an ERP-style operator application than to a content or portal site
7. dedicated workflow pages use a single-column stacked layout as the default application posture
8. dedicated workflow pages feel calm and specific, with one clear start point for that workflow rather than many competing sections

## 7. Target surfaces

Primary surfaces for this slice:

1. shared shell and shared styles under `internal/app/web_templates/_layout.tmpl` and `internal/app/web_templates/_styles.tmpl`
2. `/app`
3. `/app/login`
4. `/app/operations`
5. `/app/review`
6. `/app/inventory`
7. `/app/routes`
8. `/app/settings`
9. `/app/admin`
10. the review workbench family where stacked framing and explanatory copy still dominate over filters and tables

## 8. Implementation rules

1. prefer line, spacing, typography, and hierarchy changes over decorative surfaces
2. reserve strong panel or card treatment for true exceptions, not default page structure
3. use compact section labels and link groups on landing pages instead of mosaics of mixed-size cards
4. keep tables and exact workflow continuity links visually primary on review pages
5. keep the shell visually thin even when the route family remains broad
6. do not reintroduce any left-side navigation panel as part of this correction slice, whether heavy or visually minimized
7. keep the bundle modular and responsibility-based rather than adding another one-off styling layer
8. treat major landing or gateway pages as route directories that link to dedicated activity or workflow pages
9. make single-column stacked composition the default for dedicated activity, workflow, review, and detail pages across the application
10. on a dedicated workflow or activity page, prefer one primary start action with only the minimum adjacent context needed to begin that workflow correctly
11. when counts or queue state matter on a landing page, present them as lightweight hyperlink rows before adding any richer framing

## 9. Verification

Minimum verification for this slice:

1. `go test ./internal/app -run '^TestHandleWeb' -count=1`
2. `go build ./cmd/... ./internal/...`
3. `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
4. bounded browser review on desktop and narrow width for the target surfaces in section 7

## 10. Documentation sync

When this slice is implemented:

1. update `new_app_tracker.md` with implementation and verification status
2. update `milestone_11_operator_shell_and_navigation_plan.md` if the visual and shell posture in this slice becomes the active default
3. update `new_app_implementation_defaults.md` so the browser defaults no longer point at the superseded green card-heavy posture
4. update the workflow validation checklist if the shell or landing-page expectations used in browser review change materially
