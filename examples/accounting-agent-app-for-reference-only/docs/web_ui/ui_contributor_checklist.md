# UI Contributor Checklist

Use this quick checklist for web UI changes in `web/templates` and `internal/adapters/web`.

1. Route integrity
- Every rendered link/button target has a registered route or intentional placeholder behavior.

2. Accessibility baseline
- Icon-only controls include `aria-label`.
- Buttons have explicit `type` when relevant.
- Keyboard focus remains visible on interactive elements.

3. Alpine behavior
- Any `x-cloak` usage is supported by global `[x-cloak]` CSS.
- Hidden/visible state transitions do not flash incorrect content on load.

4. Error feedback
- User-triggered async actions surface failures in-page (not only console logs).
- Error messages are cleared when action succeeds on retry.

5. Security
- Untrusted rich text is sanitized before any HTML rendering.
- Links from untrusted content reject unsafe protocols.

6. Temporal metadata
- Date/fiscal labels are computed, not hardcoded.
- Boundary behavior is covered by unit tests.
