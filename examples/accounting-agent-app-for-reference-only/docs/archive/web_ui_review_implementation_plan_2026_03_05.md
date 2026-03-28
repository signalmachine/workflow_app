# Web UI Review Implementation Plan (2026-03-05)

## Purpose
This document defines a complete implementation plan to address web UI issues and inconsistencies identified in the March 5, 2026 review, and to raise baseline quality for accessibility, security, UX consistency, and maintainability.

## Implementation Closure
Status: completed and archived on 2026-03-05.

Implemented phases:
- Phase 1 through Phase 7 completed.

Final validation run summary:
- `make css` passed.
- `make build` passed.
- `make test` passed.
- `go test -p 1 ./...` passed.
- `go run ./cmd/verify-db-health` passed.
- `go run ./cmd/verify-agent` passed.

## Scope
In scope:
- `web/templates/layouts/*.templ`
- `web/templates/pages/*.templ`
- `web/static/css/input.css` and generated `web/static/css/app.css`
- `internal/adapters/web/*.go` where routing or UI behavior must be fixed

Out of scope:
- Domain/business logic changes unrelated to web UX
- Large visual redesigns that change product direction
- Generated templ files (`*_templ.go`) as direct edit targets

## Review Findings Mapped To Workstreams

### Critical
1. Broken vendor "View" link points to missing route (`/purchases/vendors/{code}`)
2. Potential XSS path in chat by rendering `marked.parse(...)` output via `x-html`

### High
3. Icon-only buttons missing accessible names
4. `x-cloak` used without CSS rule, causing Alpine pre-init flicker

### Medium
5. Chat upload errors are silent in UI
6. Journal Entry page does not set active nav state
7. Hardcoded FY badge (`FY 2025-26`) will become stale

## Codebase Status Snapshot (updated after implementation on 2026-03-05)
- Vendor list no longer exposes a dead `/purchases/vendors/{code}` browser link.
- Chat markdown render path now uses escape + sanitize flow before `x-html`.
- `buildAppLayoutData` now computes fiscal-year badge dynamically (env-configurable start month fallback).
- Journal Entry page now uses `reports` active nav state.
- Global `[x-cloak]` CSS rule is present in source CSS.

## Target Outcomes
- No dead-end UI routes from primary user flows
- Safe markdown rendering in chat
- WCAG-oriented accessibility baseline for interactive controls
- Consistent, predictable behavior for Alpine-driven UI states
- Clear user feedback for upload failures
- Consistent nav highlighting and date-related UI metadata

## Implementation Phases

## Phase 1: Route Integrity And Navigation Consistency
Objective: eliminate broken navigation and align active-nav behavior.

Status: completed (2026-03-05)

Implementation decision:
- Choose Option B for now: remove vendor `View` CTA that links to `/purchases/vendors/{code}` until a dedicated vendor detail page/route is implemented in a later phase.

Tasks:
1. Remove vendor `View` CTA in `vendors_list.templ` (or replace with non-link placeholder text) so no dead route is exposed.
2. Keep vendor detail as deferred scope for follow-up implementation.
3. Set non-empty active nav for Journal Entry page (`reports`).

Acceptance criteria:
- No 404 from vendor list actions.
- Journal Entry shows coherent active section in sidebar.

Validation:
- Manual click-path check across vendor list and journal entry pages.
- Route grep confirms registered endpoints match rendered links.

## Phase 2: Chat Rendering Security Hardening
Objective: prevent HTML/script injection from AI-generated content.

Status: completed (2026-03-05)

Tasks:
1. Replace raw `marked.parse()` + `x-html` trust chain with one of:
   - sanitize HTML before assignment (preferred), or
   - render markdown as escaped text if sanitization unavailable.
2. Explicitly disable raw HTML passthrough in markdown rendering config if supported.
3. Add defensive fallback for parse failures.
4. Add regression tests (unit/integration for render helper where feasible).

Acceptance criteria:
- Script tags and inline event handlers are never executed from AI output.
- Legitimate markdown formatting still renders correctly.

Validation:
- Manual test with malicious payloads (`<script>`, `onerror=`, `javascript:` links).
- Confirm rendering behavior for tables/lists/code blocks remains acceptable.

## Phase 3: Accessibility Baseline Pass
Objective: improve keyboard/screen-reader usability of controls.

Status: completed (2026-03-05)

Tasks:
1. Add `aria-label` to icon-only buttons and controls in:
   - `app_layout.templ`
   - `chat_layout.templ`
   - `modal_shell.templ`
   - `chat_home.templ`
2. Ensure all close/remove/toggle controls have explicit accessible names.
3. Verify focus styles are visible and consistent for keyboard navigation.
4. Optional quick win: add landmark/semantic improvements where trivial.

Acceptance criteria:
- Icon-only interactive controls have programmatic names.
- Keyboard user can identify and operate key controls.

Validation:
- Keyboard-only walkthrough.
- Basic screen-reader label check.

## Phase 4: Alpine UX Stability
Objective: remove pre-init flicker and inconsistent state exposure.

Status: completed (2026-03-05)

Tasks:
1. Add `[x-cloak] { display: none !important; }` to source CSS (`input.css`).
2. Rebuild `app.css` from Tailwind input.
3. Verify `x-cloak` locations behave as intended (e.g., Journal Entry status blocks).

Acceptance criteria:
- No visible flash of cloaked elements during initial page load.

Validation:
- Hard refresh on affected pages (slow network emulation if available).

## Phase 5: Chat UX Error Transparency
Objective: users receive immediate feedback on upload failures.

Status: completed (2026-03-05)

Tasks:
1. Capture upload failure responses and show inline message near attachment area.
2. Handle non-2xx responses with surfaced reason (file type/size/general error).
3. Keep successful uploads unchanged.
4. Ensure error message clears on next successful upload attempt.

Acceptance criteria:
- Failed upload attempts produce visible, actionable error feedback.

Validation:
- Test invalid mime type, oversized file, network failure scenarios.

## Phase 6: Date And Metadata Robustness
Objective: avoid stale hardcoded temporal labels.

Status: completed (2026-03-05)

Tasks:
1. Replace hardcoded FY badge with computed value.
2. Prefer deriving from company FY config if available; fallback to date-based computed FY.
3. Keep display format stable (`FY YYYY-YY`).

Acceptance criteria:
- FY badge is correct across fiscal-year boundaries without code edits.

Validation:
- Unit tests for FY formatter across boundary dates.

## Phase 7: UI Consistency Sweep (Recommended Enhancements)
Objective: standardize behavior and reduce future drift.

Status: completed (2026-03-05)

Tasks:
1. Standardize button semantics:
   - Add explicit `type` on all buttons inside forms.
2. Standardize status/loading/error patterns across Alpine pages (`order_detail`, `po_detail`, `chat_home`, `journal_entry`).
3. Normalize flash/error component usage where practical.
4. Add compact "UI checklist" in docs for future contributors.

Acceptance criteria:
- Consistent interaction patterns and fewer implicit browser defaults.

Validation:
- Spot checks on all lifecycle action pages.

## Execution Order
1. Phase 1 (route/nav correctness)
2. Phase 2 (security hardening)
3. Phase 3 (accessibility labels)
4. Phase 4 (x-cloak stability)
5. Phase 5 (chat upload feedback)
6. Phase 6 (FY badge robustness)
7. Phase 7 (consistency sweep)

Rationale:
- Fixing broken paths and security risks first prevents user-facing failures and risk exposure.
- Accessibility and UX consistency then improve operability and quality.

## Risk Assessment
- Phase 2 has highest regression risk (chat rendering behavior changes).
- Phase 1 can affect navigation expectations if vendor detail is deferred.
- Phase 6 may require business decision on FY definition if company-level FY differs from default assumption.

Mitigations:
- Keep changes incremental and behind clear helper functions.
- Validate each phase independently before proceeding.
- Avoid mixing styling refactors with security/path fixes in same PR.

## Suggested PR Breakdown
PR 1:
- Phase 1 (vendor route fix + journal entry active nav)

PR 2:
- Phase 2 (chat rendering security)

PR 3:
- Phase 3 + Phase 4 (a11y labels + x-cloak CSS)

PR 4:
- Phase 5 + Phase 6 (upload feedback + FY badge)

PR 5:
- Phase 7 (consistency sweep + contributor checklist doc)

## Validation Checklist (Per PR)
1. `make generate`
2. `make css`
3. `make test`
4. `go run ./cmd/verify-db-health`
5. `go run ./cmd/verify-agent` (if chat/AI flow touched)
6. Manual UI smoke test for changed screens

## Manual UI Smoke Test Matrix
- Login/Register: happy path + invalid credentials
- Sidebar navigation and active state
- Vendors list actions
- Journal Entry validation/post flow
- Chat:
  - simple answer
  - proposal/action confirm/cancel
  - file upload success/failure
- Modal open/close and keyboard focus behavior
- Flash message auto-dismiss and manual close

## Definition Of Done
- All acceptance criteria for completed phases pass.
- No new dead links in templates.
- No unsafe HTML execution path in chat messages.
- Accessibility baseline controls labeled.
- CSS/build/test checks pass locally.
- Documentation updated for any new UI conventions.

## Follow-Up (Post-Implementation)
1. Add lightweight automated link check for rendered routes.
2. Add frontend-focused CI checks for accessibility linting (if tooling added later).
3. Consider introducing a reusable component pattern for icon buttons and alert banners.
