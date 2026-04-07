# Web UI Design Guide

Date: 2026-04-03  
Status: Active — canonical UI design reference  
Scope: Applies to all web and mobile client surfaces built on the `workflow_app` backend. Not specific to any implementation technology.

---

## 1. Purpose and scope

This document defines the **durable design standards** for the `workflow_app` operator interface. It is not a migration guide or a component implementation plan — those live in `docs/svelte_web_guides/svelte_web_ui_migration_plan.md`. This guide defines *how the UI should feel and be structured*, regardless of implementation technology.

The standards here apply to:
- The Svelte SPA being built now
- Any future mobile client (iOS, Android, React Native, Flutter)
- Any future web UI iteration

If a design decision conflicts with this guide, update the guide intentionally with a rationale — do not silently diverge.

---

## 2. Design philosophy: what this application must feel like

This is an **enterprise operator tool**. Operators use it for high-stakes, high-frequency work: approving financial documents, reviewing AI-generated proposals, tracking workflow continuity, managing accounting entries. The UI must communicate seriousness, control, and clarity.

### 2.1 The target feeling

> Calm authority. Everything in its place. The operator knows what needs attention without being shouted at.

Reference points: Linear, GitHub Issues view, Notion database view, Vercel dashboard. These are not consumer apps — they are tools that operators trust to organize complex information without overwhelming them.

### 2.2 What the current UI does wrong (the "newspaper" problem)

The existing Go template UI feels like a newspaper with banner ads because:

1. **Two header bands consume ~120px before content begins.** A top bar plus a nav bubble strip, both full-width, both demanding attention.
2. **Navigation labels itself.** The strip is headed "Workflow destinations" — explaining its own purpose rather than just working.
3. **Every section has equal visual weight.** Page header, status summary cards, filter form, data table, sub-sections — all compete at the same volume.
4. **Content stretches the full viewport width** with minimal padding, making text lines too long and panels too wide to scan quickly.
5. **Flash messages are inline page sections** rather than transient feedback.
6. **Filters are always visible**, occupying prime viewport real estate even when the user is reading results, not filtering.

Every design decision in this guide addresses one or more of these problems.

---

## 3. Navigation model: sidebar plus contextual section tabs

### 3.1 Decision

**Use a fixed left sidebar for primary area navigation.** Do not use a horizontal top-strip of navigation bubbles or tabs as the primary global navigation.

The application may also use a **contextual second-level tab row** tied to the currently selected sidebar area. That secondary tab row must change the operator's view within the current area. It must not become a second competing global navigation system.

This is the single most impactful structural decision for enterprise feel.

### 3.2 Why sidebar over top-strip

| Concern | Top-strip nav | Left sidebar |
|---|---|---|
| Vertical content space | Loses 40–60px to nav | Zero vertical space lost |
| Nav item capacity | Runs out at ~6–8 items | Comfortably handles 10–14 items with labels |
| Active state clarity | Subtle highlight on a row of pills | Full-height background on a sidebar item — unambiguous |
| Screen real estate on wide monitors | Wastes horizontal space; nav spreads thin | Sidebar uses space that would otherwise be white |
| Admin sub-navigation | Requires a second nav strip | Admin items collapse/expand in the same sidebar |
| Mobile | Nav pills overflow-scroll or wrap | Sidebar slides in as a drawer |

The top-strip model was designed for narrow-section apps (4–6 top-level destinations). This application has enough primary areas, workflow modes, and admin sub-sections that the sidebar scales better for the global layer.

### 3.3 Sidebar layout spec

```
┌─────────────────────────────────────────────────────┐
│  [Brand mark]  Workflow App          [User menu ↓ ]  │  ← TopBar (48px fixed)
├──────────┬──────────────────────────────────────────┤
│          │                                           │
│ [Agent]  │   [Context tabs for active sidebar area]  │
│ [Acct.]  │   [Page content — full height, full       │
│ [Inv.]   │    width of right panel]                  │
│ [Ops.]   │                                           │
│  ───     │                                           │
│  [Admin] │                                           │
│  [Set.]  │                                           │
│          │                                           │
│  220px   │   calc(100vw - 220px)                     │
└──────────┴──────────────────────────────────────────┘
```

**Sidebar dimensions:**
- Width: `220px` fixed
- Background: `var(--shell-ink)` (dark navy — contrasts with the light content area)
- Nav item height: `40px`
- Nav item padding: `0 var(--space-4)`
- Active item: `var(--shell-nav-active-bg)` background, `var(--shell-nav-active-text)` color
- Hover item: `var(--shell-nav-hover)` background
- Icon: optional 16px icon to the left of label; text label always visible

**TopBar:**
- Height: `48px` fixed
- Full width, sits above both sidebar and content
- Contains: brand mark (left), user menu (right)
- Background: `var(--shell-ink-strong)`
- No primary global navigation in the top bar itself

**Context tab row:**
- Sits below the TopBar and above page content inside the main content column
- Belongs to the currently selected sidebar area
- Uses concise mode labels such as `Overview`, `Workflows`, `Actions`, `Lists`, `Reports`, `Search`
- May vary by sidebar area; for example the `Agent` area can lead with `Messages` and `Requests`
- Must remain a single restrained row, horizontally scrollable on narrow widths rather than wrapping into a heavy second header band
- Must not duplicate the sidebar's major-area choices

**Content area:**
- Starts at `top: 48px` (below TopBar), `left: 220px` (right of sidebar)
- Background: the page gradient
- Inner padding: `var(--space-8) var(--space-8)` (32px all sides) on desktop
- Max content width within the content area: `1100px` — not 1320px (see §6.1)

### 3.4 Major-area model

The left sidebar should be reserved for the most stable major areas of the application. The intended forward model is:

```
  Agent
  Accounting
  Inventory
  Operations
  ───
  Admin
  Settings
```

Rules:

1. sidebar items represent durable areas of responsibility, not one-off actions or page shortcuts
2. action pages such as request submission, feeds, search, and grouped reports belong under contextual tabs within an area rather than as peer sidebar items
3. `AR` and `AP` should remain under `Accounting` until they become real first-class route families with their own coherent operator surfaces
4. `Agent` is an agent-observability and agent-control area, not a second owner of accounting, inventory, or execution truth
5. contextual tabs may expose domain workflows through the selected area's lens, but they must not reassign canonical workflow ownership

### 3.5 Admin and sub-navigation

Admin sub-sections (Accounting setup, Party setup, Access controls, Inventory setup) are collapsible items under an "Admin" parent in the sidebar. There is **no second nav strip** for admin pages. The `AdminLayout.svelte` double-nav pattern from the original plan is replaced by sidebar expand/collapse.

The sidebar item for Admin expands to reveal:
```
  Admin                    ▾
    Accounting setup
    Party setup
    Access controls
    Inventory setup
```

This eliminates the entire `AdminLayout.svelte` component — admin pages use the same `AppLayout.svelte` as all other pages.

### 3.6 Mobile nav

On viewport widths below `768px`:
- The sidebar collapses to hidden
- A hamburger button appears in the TopBar (left side)
- Tapping it slides the sidebar in as a full-height drawer overlay
- Tapping outside or a nav item closes it

On desktop widths:
- The sidebar may support a collapsible compact state
- The collapsed state should reduce the rail to icon-only or narrow-label presentation rather than hiding the area model completely
- The collapsed or expanded preference should persist per user where practical

---

## 4. Typography scale

### 4.1 Font

**IBM Plex Sans** for all UI text. IBM Plex Mono for all code, IDs, references, and technical values.

Load weights: 400 (regular), 500 (medium), 600 (semibold). Do not load bold (700) — semibold at 600 is sufficient and renders more cleanly at small sizes.

Self-host from Google Fonts or use the `@fontsource/ibm-plex-sans` npm package for the Svelte build to avoid external network requests.

### 4.2 Type scale (tokens)

Add these tokens to `:root`:

```css
:root {
  /* Type scale */
  --text-2xl: 1.5rem;     /* 24px — page hero, used sparingly */
  --text-xl:  1.25rem;    /* 20px — major page titles */
  --text-lg:  1.0625rem;  /* 17px — section headings */
  --text-base: 0.875rem;  /* 14px — default body text */
  --text-sm:  0.8125rem;  /* 13px — table cell body, secondary text */
  --text-xs:  0.75rem;    /* 12px — meta text, timestamps, labels */
  --text-2xs: 0.6875rem;  /* 11px — eyebrow labels, status pill text */

  /* Line heights */
  --leading-tight: 1.25;
  --leading-base:  1.5;
  --leading-loose: 1.75;
}
```

### 4.3 Usage rules

| Element | Size token | Weight | Transform | Color |
|---|---|---|---|---|
| Page H1 | `--text-xl` | 600 | none | `var(--ink)` |
| Section H2 | `--text-lg` | 600 | none | `var(--ink)` |
| Eyebrow label | `--text-2xs` | 700 | `uppercase`, `letter-spacing: 0.08em` | `var(--ink-faint)` |
| Body / table cell | `--text-base` | 400 | none | `var(--ink)` |
| Secondary text | `--text-sm` | 400 | none | `var(--ink-soft)` |
| Meta / timestamp | `--text-xs` | 400 | none | `var(--ink-faint)` |
| Status pill | `--text-2xs` | 700 | `uppercase`, `letter-spacing: 0.06em` | semantic |
| Code / ID / ref | `--text-sm` | 400 | none | `var(--ink-soft)`, monospace |
| Button label | `--text-sm` | 600 | none | contextual |

### 4.4 Rules

1. **Use no more than three type sizes on any single page view.** If you have page title (xl), body (base), and meta (xs), that is already three. Adding a fourth creates visual noise.
2. **Never use `font-weight: 700` (bold) for headings or body text.** Semibold (`600`) is the maximum for all prose, labels, and headings. The sole exceptions are eyebrow labels and status pills (both `--text-2xs`, uppercase, heavily tracked) where 700 improves legibility at tiny sizes. Do not use 700 anywhere else.
3. **Eyebrow labels are the correct way to label sections** — not H3 or H4 tags. Eyebrows are `--text-2xs`, uppercase, tracked, in `var(--ink-faint)`. They are subordinate to the content below them.
4. **References, IDs, and machine values** (request refs like `REQ-0042`, UUIDs, amounts) always render in `var(--font-mono)`. This makes them scannable as technical values, distinct from prose.

---

## 5. Design tokens

### 5.1 Color tokens (`:root`)

```css
:root {
  /* Background */
  --bg: #e9eef2;
  --bg-deep: #dbe5ea;

  /* Surfaces */
  --surface: rgba(252, 251, 247, 0.94);
  --surface-strong: rgba(255, 255, 255, 0.98);
  --surface-muted: #eef2f4;
  --surface-accent: #eff4f8;

  /* Shell (sidebar + topbar) */
  --shell-ink: #1d3343;
  --shell-ink-strong: #122533;
  --shell-nav-hover: rgba(255, 255, 255, 0.07);
  --shell-nav-active-bg: rgba(220, 233, 242, 0.15);
  --shell-nav-active-text: #c8dcea;

  /* Text */
  --ink: #18232d;
  --ink-soft: #5a6975;
  --ink-faint: #7a8894;

  /* Borders */
  --line: rgba(96, 117, 132, 0.26);
  --line-strong: rgba(59, 80, 95, 0.42);

  /* Accent (primary interactive) */
  --accent: #2f617f;
  --accent-strong: #1d4359;
  --accent-soft: #dce9f2;
  --accent-faint: rgba(47, 97, 127, 0.10);

  /* Semantic status */
  --good: #21634a;
  --good-soft: #e6f2ec;
  --bad: #a64553;
  --bad-soft: #f9ebee;
  --warn: #7a5418;
  --warn-soft: #fdf2e0;
  --neutral: #4d6274;
  --neutral-soft: #e7edf2;

  /* Shadows */
  --shadow-sm: 0 1px 4px rgba(18, 37, 51, 0.08), 0 4px 12px rgba(18, 37, 51, 0.04);
  --shadow:    0 2px 8px rgba(18, 37, 51, 0.08), 0 12px 32px rgba(18, 37, 51, 0.08);
  --shadow-lg: 0 4px 16px rgba(18, 37, 51, 0.10), 0 24px 48px rgba(18, 37, 51, 0.12);
}
```

> **Note on shadows:** The new shadow scale is more layered than the current template. A combined ambient + directional shadow reads as more premium and is how Linear, Vercel, and similar enterprise apps compose their card surfaces.

### 5.2 Spacing tokens

```css
:root {
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 20px;
  --space-6: 24px;
  --space-8: 32px;
  --space-10: 40px;
  --space-12: 48px;
  --space-16: 64px;
}
```

### 5.3 Radius tokens

```css
:root {
  --radius-xl: 20px;   /* Large panels, modals */
  --radius-lg: 14px;   /* Cards, panels */
  --radius-md: 10px;   /* Inputs, smaller surfaces */
  --radius-sm: 999px;  /* Pills, badges */
}
```

> **Note:** Radius values are reduced from the current template (`--radius-xl: 24px`, `--radius-lg: 18px`). Excessively large radii look playful rather than professional. The adjusted values are still generous without feeling consumer.

### 5.4 Layout tokens

```css
:root {
  --sidebar-width: 220px;
  --topbar-height: 48px;
  --content-max-width: 1100px;
  --page-gutter: var(--space-8);   /* 32px */
  --panel-padding: var(--space-6); /* 24px */
}
```

### 5.5 Z-index scale

```css
:root {
  --z-shell: 100;
  --z-sidebar: 110;
  --z-overlay: 200;
  --z-toast: 300;
}
```

### 5.6 Motion tokens

```css
:root {
  --transition-fast: 100ms ease;
  --transition-base: 160ms ease;
  --transition-slow: 240ms ease;
}
```

### 5.7 Token rules

1. **Never hardcode hex values in component styles.** Always reference a token.
2. **Never use a numeric value for spacing without a token.** Use `var(--space-4)` not `16px`.
3. **Semantic names over descriptive names.** `var(--good)` is correct. `var(--green)` is not.
4. When a new semantic need arises (e.g., a warning state), add a new token — do not reuse an existing token with a different intent.

---

## 6. Layout and whitespace doctrine

### 6.1 Content width

**Maximum content width is `1100px`**, not the full viewport width. On wide monitors, content is centered within the right-side content area.

The current template uses `--max-width: 1320px`. This is too wide — at 1320px, a table row stretches across nearly the full screen, columns spread thin, and row content is hard to track across. `1100px` is closer to the sweet spot of most enterprise dashboards.

### 6.2 Minimum panel padding rule

> **Every surface that contains content must have at least `var(--panel-padding)` (24px) of inner padding on all sides.**

This is the single most impactful whitespace rule. When panels breathe, the whole page breathes. Violating this rule — padding below 24px on any visible card or section — is the primary cause of the "cramped" feeling.

Apply this to: cards, table containers, detail panels, filter panels, modals, sidebar group containers.

Exception: table rows themselves — cell padding is `var(--space-3) var(--space-4)` (12px 16px) to keep row density appropriate for data scanning.

### 6.3 Section gap rule

**Gap between stacked page sections is always `var(--space-8)` (32px) minimum.**

When sections are close together they feel like one blob of noise. 32px gap makes the hierarchy visually obvious — each section is its own distinct island.

### 6.4 Table row height

**Minimum table row height is `48px`.**

Current template rows are compact and dense. 48px rows are still data-dense (you can fit 12 rows in a 600px table) while being comfortable to read and click.

### 6.5 Line length cap

**Body text and detail field values cap at `65ch`** (approximately 65 characters). Very wide text blocks are hard to read. Apply `max-width: 65ch` to paragraphs and description fields within detail cards.

---

## 7. Information hierarchy: one primary job per page

### 7.1 The principle

> **Every page has one primary job. Everything else is secondary.**

The operator arrives at a page with a task. The visual design must make the primary job immediately obvious and reduce friction to completing it.

### 7.2 Hierarchy by page type

**List pages** — the primary job is *finding the right row*:
- The table dominates: it should take the most screen space
- The filter panel is secondary: collapsed by default, activated when needed
- Status summary cards (if present) are tertiary: compact, above the table, not competing for height

**Detail pages** — the primary job is *reading this entity*:
- The entity hero card dominates (status, key identifiers, timestamps)
- Actions are secondary: grouped in a clear action panel, not scattered across the page
- Sub-sections (AI runs, attachments, accounting links) are tertiary: collapsed by default, expanded on demand
- JSON/raw payload viewers are collapsed by default — they are never the primary job

**Dashboard** — the primary job is *knowing what needs my attention*:
- The highest-priority action item dominates (pending approvals badge, queued requests count)
- Role-aware action cards are primary content
- Status grids are secondary: compact, below the action cards, not a sea of equal-weight tiles

**Admin pages** — the primary job is *managing the list*:
- The list table is primary
- The create form is secondary: in a collapsible panel or right-side drawer, not above the list by default

### 7.3 Visual weight rules

| Priority | Visual treatment |
|---|---|
| Primary content | Full `--ink` color, `--text-base` or larger, prominent position |
| Secondary content | `--ink-soft` color, `--text-sm`, visually below or less wide |
| Tertiary / collapsed | `--ink-faint` color, eyebrow label treatment, collapsed until explicit user action |
| Actions (primary) | Full accent button, visible but not dominating |
| Actions (secondary/destructive) | Ghost/outline button or text link; never the first visual focus |

---

## 8. Color usage: restraint is the rule

### 8.1 Semantic color only

Color is used **only** to communicate meaning, never for decoration:

| Color intent | Token | Use |
|---|---|---|
| Success / healthy / approved | `--good` / `--good-soft` | Status pills, positive counts |
| Error / failed / rejected | `--bad` / `--bad-soft` | Status pills, error states |
| Warning / pending / attention | `--warn` / `--warn-soft` | Status pills, caution indicators |
| Neutral / inactive / default | `--neutral` / `--neutral-soft` | Default status pill, inactive states |
| Interactive / primary action | `--accent` / `--accent-strong` | Buttons, links, active sidebar items |

### 8.2 Rules

1. **No decorative color.** Background sections, dividers, and containers must use surface tokens (`--surface`, `--surface-muted`) not color tokens.
2. **Status pills are the only colored elements in a table row.** Every other cell is `--ink` or `--ink-soft` on a white/surface background.
3. **Limit red (`--bad`) to genuine error states.** Do not use `--bad` for warnings or pending states. `--warn` exists for that.
4. **The sidebar uses its own ink system** (`--shell-ink`, `--shell-nav-active-bg`) — not the main token set. The sidebar is intentionally a different visual register.

---

## 9. Component behavior standards

### 9.1 Loading states

Every component that fetches data must show a loading state. No blank intermediate states.

- **Tables:** render 3 skeleton rows (pulsing gray bars) while `loading: true`
- **Detail cards:** render skeleton placeholder blocks while `loading: true`
- **Counts and summary cards:** show a dash (`—`) while loading, not zero

### 9.2 Empty states

Every list or table must show an explicit empty state when `rows.length === 0` and `loading === false`.

Empty states must have:
- A short title ("No inbound requests match this filter")
- A brief body with a constructive hint ("Try adjusting the status filter or clearing the search term")
- Optionally, a direct action link

An empty table with no message is not acceptable.

### 9.3 Confirmation before side effects

**Any action that mutates persistent state must be confirmed before submission.** This includes:
- Approve / reject approval
- Cancel inbound request
- Delete inbound draft
- Mark items inactive
- Any admin create or delete action

Use the `ConfirmAction` modal component. Never use the browser's native `confirm()` dialog.

### 9.4 Filters: collapsed by default on detail pages, open by default on list pages

- List pages: filter panel opens by default (operators come to list pages to filter)
- Detail pages: filter panels (if any) collapsed by default (operators come to detail pages to read one entity)
- When a filter panel is collapsed and has active filters applied, show an active filter count badge on the collapse toggle

### 9.5 Progressive disclosure on detail pages

Detail pages for complex entities (inbound request detail, proposal detail, approval detail) contain many sub-sections. Apply progressive disclosure:

- **Always visible:** entity hero card (status, key fields, dates), primary action panel
- **Collapsed by default, expandable:** AI run traces, artifacts, attachments, delegation chains, audit trail
- **Never visible by default:** raw JSON payloads, internal system IDs (collapsed inside `JsonBlock`)

The operator should not need to scroll past a wall of expanded sections to find the action they need.

### 9.6 Toast notifications, not inline flash banners

Feedback from user actions (approval submitted, draft saved, request queued) must appear as **auto-dismissing toast notifications**, not as inline page banners that push content down. Toasts:
- Appear in the top-right corner
- Auto-dismiss after 4 seconds
- Can be manually dismissed
- Persist across soft navigations within the SPA
- Stack if multiple notifications appear simultaneously (max 3 visible)

Inline banners are reserved for page-load error states (e.g., the API returned an error loading the page — this persists until resolved, it is not transient feedback).

---

## 10. Shell and brand

### 10.1 TopBar

- 48px height, full width
- Dark background (`var(--shell-ink-strong)`)
- Left: brand mark ("WA" monogram) + app name
- Right: user menu (display name + role, dropdown for settings and logout)
- No navigation in the TopBar — all navigation is in the sidebar

### 10.2 Sidebar

- 220px width, full height (below TopBar), fixed position
- Dark background (`var(--shell-ink)`)
- Top section: primary nav items (Home, Intake, Operations, Review, Inventory)
- Separator
- Bottom section: Admin (collapsible with sub-items), Settings
- Each item: 40px height, `--text-sm` weight 500, icon optional
- Active item: `var(--shell-nav-active-bg)` background, `var(--shell-nav-active-text)` color
- Hover item: `var(--shell-nav-hover)` background

### 10.3 Brand identity rules

- App name: "Workflow App" — sentence case, never all-caps
- The brand mark (monogram "WA" or any future logo asset) appears only in the TopBar — do not repeat it in page content
- Subtext in the TopBar below the app name: omit the current "Persist-first operator shell" — this is implementation jargon, not useful UX copy

---

## 11. Enterprise UI anti-patterns to avoid

These patterns are present in the current Go template UI and must **not** be carried forward into the Svelte implementation:

| Anti-pattern | Why | Correct alternative |
|---|---|---|
| Navigation that labels itself ("Workflow destinations") | Self-evident UI doesn't need explanation | Remove the label; the nav items speak for themselves |
| Two header bands before content | Wastes vertical space; newspaper feeling | One TopBar + sidebar; content starts at the top of the content area |
| Flash banners as full-width page sections | Push content down; feel alarming even for routine notices | Toast notifications top-right, auto-dismiss |
| Filter forms always visible on list pages | Correct default is already filtered; form visibility wastes space | Collapsible filter panel; show filter summary when collapsed |
| Equal visual weight for every section | Everything competes; nothing communicates priority | Information hierarchy: primary content visually dominant |
| Create forms above list tables on admin pages | Admin is primarily manage, not create | Create in a right-side panel or collapsible section; list table is primary |
| Raw `<pre>` blocks for JSON payloads | Monospace walls in the middle of page content | `JsonBlock`: collapsible, syntax-highlighted, copyable, collapsed by default |
| Approval decisions as raw HTML form POST | No confirmation, no feedback loop | `ConfirmAction` modal + toast on completion |
| Dense inline link rows in table cells (`view · edit · approve · continue`) | Hard to scan, takes cell space | `ContinuityLinks` component: icon-differentiated, compact flex-wrap |
| Page content stretching to browser edge | Overly wide lines; nothing organized | Max content width `1100px`, centered, with `32px` page gutter |
| Hard-coded hex colors in component styles | Breaks when design system changes | Always `var(--token-name)` |

---

## 12. Decision log

Decisions made here should be recorded for traceability.

| Date | Decision | Rationale |
|---|---|---|
| 2026-04-03 | Left sidebar navigation over top nav-bubble strip | Eliminates double-header; scales to 10+ nav items; standard enterprise pattern |
| 2026-04-03 | Max content width `1100px` (not `1320px`) | Prevents over-stretched tables and text lines on wide monitors |
| 2026-04-03 | Font weight max `600` semibold, no `700` bold | Bold renders aggressively at `--text-sm`; semibold is sufficient |
| 2026-04-03 | Toast notifications, not inline flash banners | Inline banners push content and feel alarming for routine feedback |
| 2026-04-03 | No Tailwind CSS | Token-first design system; scoped CSS; enterprise evolveability |
| 2026-04-03 | IBM Plex Sans | Professional, neutral, high legibility; consistent with IBM Carbon, Linear |
| 2026-04-03 | Radius values reduced (xl: 20px, lg: 14px) | `24px`/`18px` radii felt playful; adjusted values are professional but still generous |
| 2026-04-03 | Minimum panel padding `24px` | Sub-24px padding was a primary cause of the "cramped" feeling |
| 2026-04-03 | Table row minimum height `48px` | Dense-but-readable; touch-friendly for potential tablet operators |
