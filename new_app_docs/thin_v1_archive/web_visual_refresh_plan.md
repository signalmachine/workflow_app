# workflow_app Web Visual Refresh Plan

Date: 2026-03-30
Status: Implemented in code on 2026-03-30; bounded corrective follow-up and manual browser-review evidence still pending
Purpose: define one bounded UI-refresh pass that makes the promoted web layer look more professional and operator-ready before the deferred live workflow review resumes.

## 0. Current implementation status

As of 2026-03-30, the implementation pass for this slice is landed in code:

1. the shared web template now uses a low-glare blue-gray enterprise visual system instead of the earlier warm beige palette
2. the typography has moved to a cleaner sans-serif stack
3. the targeted page set in section 3.1 now has refreshed navigation, hierarchy, cards, forms, tables, and filter layout treatment
4. focused `internal/app` HTTP coverage now asserts the refreshed dashboard shell, login surface, inbound-request review filters, and proposal-review filters
5. repository verification is complete through `go build ./cmd/... ./internal/...` and `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
6. implementation review then found one bounded corrective slice: align the real login route with the scoped page list and fix shared-template narrow-width overflow on remaining unwrapped table pages; that work is now planned in `web_visual_refresh_follow_up_plan.md`
7. the remaining open items for full closeout are that corrective slice plus the bounded manual browser review on desktop and narrow-width layouts required by section 7

## 1. Why this plan exists

The current thin-v1 web layer is functionally credible but visually underpowered.

The browser surface already supports the main operator path, but the current presentation still carries avoidable friction:

1. the typography feels dated for a modern business application
2. the warm beige palette does not read as a crisp professional enterprise surface
3. forms, tables, dashboard cards, and review sections are usable but not yet visually clear enough for sustained operator use
4. the real browser validation and user-testing pass will produce better signal if the application first looks intentionally maintained rather than technically exposed

This should be handled as one bounded pre-validation slice rather than left as an implicit “later polish” note.

## 2. Planning decision

Current decision:

1. land one inexpensive but meaningful visual-refresh pass before the deferred Phase 1 workflow validation resumes
2. keep the work on the existing Go server-rendered web layer without introducing a frontend rewrite, Node toolchain, design-system framework, or dark-mode branch
3. keep the slice focused on presentation, readability, and light usability refinement rather than turning it into broad workflow redesign
4. keep the overall product in a light visual mode with calmer low-glare surfaces rather than bright white or dark backgrounds

This slice is intended to:

1. improve first-run operator trust and perceived product quality
2. make the real workflow-validation pass easier to execute and easier to judge
3. avoid unnecessary churn later when users react first to presentation quality instead of workflow continuity

## 3. Scope

In scope:

1. a refreshed global color-token set for the existing web template
2. improved typography for forms, dashboard sections, tables, review pages, and navigation
3. cleaner panel, card, table, form-control, button, and status-pill styling
4. improved spacing, grouping, and hierarchy for the exact pages listed in section 3.1
5. small cheap usability improvements that fit naturally into the same pass, such as clearer labels, button emphasis, placeholder text tuning, empty-state clarity, and stronger filter-form readability
6. light responsive cleanup where obvious browser friction exists on desktop or narrower screens

Out of scope:

1. a SPA or separate frontend stack
2. dark mode
3. deep workflow redesign
4. broad copywriting or product-terminology rewrites
5. broad interaction-pattern changes that require new backend contracts
6. a formal enterprise design system buildout

## 3.1 Exact page scope

This slice should touch only the following required browser surfaces unless implementation reveals one narrowly coupled template dependency:

1. `/app` dashboard
2. `/app/login` sign-in surface rendered through the shared unauthenticated shell
3. `/app/inbound-requests/{request_reference_or_id}`
4. `/app/review/inbound-requests`
5. one downstream review surface only, with `/app/review/proposals` as the default first choice unless a more heavily used downstream page is discovered during implementation review

Defer all other page-specific styling follow-ups unless:

1. the shared template changes already improve them automatically, or
2. one page becomes an obvious visual outlier after the bounded pass above

## 4. Visual direction

The target should feel closer to a modern enterprise cloud product than to a prototype.

Preferred direction:

1. light blue-gray or slate-tinted backgrounds rather than warm beige or bright white
2. restrained enterprise accent colors, likely blue or blue-teal, rather than earthy tones
3. dark slate text rather than heavy pure-black contrast
4. sans-serif UI typography rather than default serif presentation
5. subtle depth through border, surface, and shadow control rather than heavy gradients or decorative effects

Practical style rule:

1. borrow the discipline of products such as SAP, Salesforce, or Workday in tone and clarity
2. do not attempt to imitate any one product exactly
3. keep the implementation cheap by improving tokens, spacing, and component states first

## 5. Required outcomes

This slice is complete only when:

1. the shared web template uses the refreshed light enterprise-style color tokens instead of the current warm beige palette
2. the shared UI typography has moved from the current serif-heavy presentation to a cleaner sans-serif stack
3. `/app` and the unauthenticated sign-in surface have updated spacing, hierarchy, controls, and panel styling
4. `/app/inbound-requests/{request_reference_or_id}`, `/app/review/inbound-requests`, and one bounded downstream review page have updated table or filter or card readability on the same visual system
5. buttons, alerts, status indicators, and form controls use one coherent refreshed visual language across the pages in section 3.1
6. the final result stays in a low-glare light theme and avoids both bright white and dark-background presentation
7. bounded browser review evidence is recorded for desktop and narrow-width layouts on the exact pages in section 3.1

## 6. Recommended implementation order

Implement in this order:

1. refresh the global CSS token set in the shared web template
2. replace the current serif-heavy typography with a cleaner sans-serif stack
3. restyle the login page, top shell, buttons, inputs, cards, tables, notices, and status chips
4. tighten spacing and section structure on the exact pages listed in section 3.1
5. add any cheap clarity improvements discovered while touching those templates
6. then rerun bounded browser review and the deferred Phase 1 workflow validation

Execution rule:

1. prefer one coherent style pass over many tiny visual tweaks
2. keep all changes inside the existing web template and shared page rendering seam unless a narrow template decomposition is clearly needed for maintainability
3. do not reopen backend milestone scope under the cover of UI polish

Suggested implementation posture:

1. prioritize the highest-visibility low-hanging improvements first, especially shared color tokens, typography, spacing, panels, buttons, forms, tables, and status styling
2. prefer cheap global wins from the shared template over page-by-page redesign
3. use the bounded page list in section 3.1 to avoid visual-scope drift
4. avoid pixel-level perfection work when a simpler shared refinement already achieves the objective of a more credible professional operator surface

## 7. Validation expectations

Required verification for this slice:

1. `gopls` diagnostics on edited Go files
2. `go build ./cmd/... ./internal/...`
3. `set -a; source .env; set +a; GOCACHE=/tmp/go-build go test -p 1 ./cmd/... ./internal/...`
4. bounded browser review of the sign-in page, dashboard, inbound-request detail, and at least one downstream review surface on desktop and narrow-width layouts
5. documentation sync in the canonical planning docs if sequencing or status changes during the pass

## 8. Sequencing impact on the current validation plan

This visual-refresh plan should land before the deferred live workflow validation resumes.

Reason:

1. the remaining validation work is explicitly browser- and operator-facing
2. the current visual layer is not blocking correctness, but it is weak enough to reduce testing quality and user confidence
3. the refresh is bounded and cheaper now than after another round of live workflow review

Priority rule:

1. implement this visual-refresh slice next
2. then resume the paused Phase 1 workflow validation from `post_checkpoint_validation_and_user_testing_plan.md`
3. if the refresh uncovers only narrow browser regressions, fix them inside the same bounded slice rather than reopening broader web-layer scope
