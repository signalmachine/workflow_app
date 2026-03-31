# workflow_app Web UI Streamlining Plan

Date: 2026-03-31
Status: Draft canonical planning slice
Purpose: capture the specific web-UI issues observed in the current promoted browser layer so they can be fixed later as one bounded usability pass without changing workflow behavior or backend ownership.

## 1. Why this plan exists

The current web layer is usable, but it still feels busier than it should.

The application already has the right backend foundation and the right workflow control model. The remaining problem is presentation density:

1. too many competing visual elements are present on the same page
2. repeated callouts, helper text, and secondary links make pages feel noisy
3. list pages and detail pages often expose more visual weight than the operator needs at first glance
4. the global chrome is improved, but several workflow pages still need a calmer hierarchy

This plan exists so those issues are recorded as a bounded future cleanup slice rather than left as vague “UI polish” drift.

## 2. Planning decision

Current decision:

1. treat the current browser layer as functionally usable but visually over-stimulated
2. plan one bounded streamlining pass that reduces noise, clarifies hierarchy, and improves scanability
3. keep the work inside the existing Go server-rendered stack
4. do not widen the UI work into a workflow redesign, backend contract rewrite, or separate frontend architecture

The goal is not visual novelty.

The goal is:

1. calmer pages
2. clearer primary actions
3. less competing chrome
4. easier operator scanning

## 3. Observed issues

These are the concrete issues visible in the current browser layer.

### 3.1 Dashboard density

The dashboard still carries too many simultaneous focal points:

1. hero copy
2. quick links
3. coordinator queue action
4. status summary cards
5. recent requests
6. pending approvals
7. processed proposals

That is technically functional, but it reads as busy rather than focused.

### 3.2 Navigation weight

The top navigation is improved, but it still presents many destinations at equal visual weight.

This makes the browser shell feel more like a dense admin menu than a restrained operator control surface.

### 3.3 Repeated helper text

Several pages repeat explanatory copy in multiple blocks:

1. hero copy
2. section notes
3. card captions
4. table lead-ins
5. inline action notes

The result is that operators see the same idea several times on one page.

### 3.4 Table and filter noise

List pages often expose:

1. filter controls
2. summary cards
3. table headers
4. row actions
5. secondary metadata lines

This is useful for completeness, but the visual hierarchy is too flat.

### 3.5 Detail-page overload

Several detail pages try to show everything at once:

1. identity or status summaries
2. related links
3. lifecycle state
4. audit links
5. related records
6. action buttons

The operator can still use the page, but the layout does not clearly distinguish primary from secondary information.

### 3.6 Inconsistent page rhythm

Some pages are card-heavy, some are table-heavy, and some are detail-heavy.

That is expected in a workflow app, but the spacing and section rhythm are not yet unified enough to make the whole surface feel intentionally designed.

## 4. Design intent

The future cleanup should make the browser layer feel:

1. quieter
2. more structured
3. less repetitive
4. more confident about what matters first

Specific intent:

1. reduce the number of visible simultaneous calls to action on each screen
2. keep the first visible action obvious
3. compress secondary controls into lower emphasis areas
4. make status, ownership, and state transitions easier to scan
5. give detail pages one clear primary content block and then secondary supporting blocks

## 5. Scope

In scope:

1. dashboard hierarchy cleanup
2. navigation simplification or regrouping
3. tighter page-header and hero-copy usage
4. more disciplined use of cards, callout blocks, and helper text
5. better separation between primary and secondary actions
6. table/list-page hierarchy cleanup
7. detail-page visual de-cluttering
8. consistency improvements across review pages

Out of scope:

1. backend workflow changes
2. new routes or new data models
3. redesigning the workflow architecture
4. introducing a separate frontend stack
5. turning the browser surface into a product-marketing site

## 6. Bounded implementation themes

The eventual fix should likely be broken into a few bounded themes.

### 6.1 Shell and navigation simplification

Likely improvements:

1. reduce the visual weight of the global nav
2. group high-level destinations more deliberately
3. distinguish primary workflow paths from secondary review paths
4. keep the shell present but not dominant

### 6.2 Dashboard hierarchy pass

Likely improvements:

1. reduce simultaneous panels
2. collapse less important guidance into lighter notes
3. make status summaries more compact
4. ensure the coordinator action and primary intake continuation are the main focal points

### 6.3 Review page decluttering

Likely improvements:

1. shorten repeated helper text
2. reduce duplicate card-level explanation
3. keep filters together and visually subordinate to the data they control
4. make row actions less visually noisy

### 6.4 Detail page hierarchy pass

Likely improvements:

1. make one section the obvious primary fact block
2. move supporting links and audit references into lower-emphasis sections
3. reduce the number of equally weighted detail cards
4. keep long detail pages scannable rather than visually dense

## 7. Guardrails

Do not:

1. change workflow semantics under the label of UI cleanup
2. move logic into the browser that belongs in backend services
3. introduce a second frontend architecture
4. turn this into open-ended styling experimentation
5. widen scope into unrelated features because the browser still feels busy

The point is to remove noise, not to invent a new look.

## 8. Success criteria

This plan is complete when:

1. the browser shell feels calmer and more intentional
2. the dashboard has a clearer primary story
3. review pages are easier to scan at a glance
4. detail pages distinguish primary facts from secondary context more clearly
5. the operator can move through the application with fewer competing visual cues
6. the underlying workflow behavior remains unchanged

## 9. Relation to the current web layer

This plan should be treated as a follow-on to the already-landed visual refresh work.

That means:

1. it starts from the improved low-glare shell, not the earlier warm-beige version
2. it focuses on density, hierarchy, and scanability rather than color-token experimentation
3. it should be implemented only after the team decides that the remaining visual noise is worth another bounded pass
4. it should remain subordinate to the current workflow and correctness priorities

## 10. Notes for future implementation

When this slice is eventually implemented, prefer:

1. one shared cleanup pass over many tiny page-specific edits
2. removing visual weight before adding new component structure
3. fewer repeated explanation blocks
4. stronger section hierarchy
5. lighter secondary actions

If the cleanup requires broader page restructuring, split that into a follow-on slice rather than letting it expand silently.
