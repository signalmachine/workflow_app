# Svelte Web UI Migration Plan

Date: 2026-04-03  
Status: Implementation planning — approved for execution  
Purpose: Define the complete architecture, component structure, API seam, and phased execution plan for replacing the Go `html/template` web layer with a modular Svelte SPA.

**Design standards companion document:** [`docs/svelte_web_guides/web_ui_design_guide.md`](web_ui_design_guide.md) — the durable canonical reference for navigation model, typography, whitespace, information hierarchy, and enterprise UI principles. All design decisions during implementation must align with that guide.

**Svelte 5 implementation guide:** [`docs/svelte_web_guides/svelte_guide.md`](svelte_guide.md) — correct Svelte 5 runes syntax, common AI-tool mistakes to reject, application-specific patterns, router integration, stores, and component templates. Read this before writing any Svelte code.

---

## 1. Context and motivation

### 1.1 What is being replaced

The current browser layer lives entirely inside `internal/app` and is rendered by the Go server:

| File / area | Size | Function |
|---|---|---|
| `web.go` | 34 KB | ~40 typed page-data structs, feed builders, template helpers |
| `web_*.go` handler files | ~130 KB total | Route handlers, session auth, data assembly, HTML rendering |
| `web_templates/*.tmpl` | ~160 KB total | Go `html/template` markup |
| `web_templates_bundle.go` | 10 KB | Template parse and dispatch |
| CSS (inline in `_styles.tmpl`) | 18 KB | Design system, inline in every page's `<style>` block |

The backend API (`api.go`, `api_*.go`) is a **separate, clean `/api/...` JSON seam** that is already:
- Issuing both browser session cookies and bearer tokens
- Handling all business logic in domain services
- Returning typed JSON for every route family

The API was deliberately built to be **client-neutral**. It is the correct foundation for a Svelte client.

### 1.2 Known current UI deficiencies

Analysis of the existing template layer surfaces these structural problems that the Svelte migration should fix:

1. **No component reuse.** The 11 template files duplicate filter forms, status pills, continuity-link patterns, and empty-state blocks dozens of times with no shared component abstraction.

2. **Fat `webPageData` union struct.** A single 42-field union struct carries the entire page payload for every route. Every handler touches the same struct whether it needs 2 fields or 20.

3. **CSS delivered inline.** All 18 KB of CSS is embedded in every page's `<style>` tag via `{{template "web_styles" .}}`. No caching, no splitting, no per-page scope.

4. **No client-side reactivity.** Filters require a full round-trip GET. Status updates require a form POST and full page reload. There is no optimistic UI or incremental loading.

5. **Continuation-link patterns are verbose and inconsistently hand-coded.** Every table row in review pages manually assembles 4–7 inline links with conditional Go template logic, making the markup dense and hard to scan.

6. **Approval decision forms are inlined in detail pages** using raw HTML `<form method="post">` without any confirmation UX or feedback cycle.

7. **No accessible routing.** Navigation state is expressed through CSS class helpers (`navClass`, `navSectionClass`) applied to `<a>` tags. There is no router, no history management, and no back-button continuity in complex filter states.

8. **Admin pages mix list and create forms on the same surface.** Eight separate admin pages inline creation forms above their list tables, making the surface noisy rather than task-centered.

9. **Page hierarchy is flat.** Every page uses the same layout template with no concept of nested layouts for section groups (e.g., all Review pages sharing a Review shell, all Admin pages sharing an Admin shell).

10. **Typography and visual system embedded in Go templates.** Fonts, color tokens, and spacing are not portable or extractable — they live inside a `.tmpl` file.

### 1.3 What will not change

The Go backend is **unchanged**:
- All business logic stays in domain services under `internal/`
- All routes in `api.go` remain as-is
- Session auth (cookies and bearer tokens) remains identical
- Database schema, migrations, and domain models are unaffected
- `cmd/app/` binary serves both the Go APIs and the Svelte static build

---

## 2. Architectural decisions

### 2.1 SPA with Go serving static files

**Decision: Svelte SPA (client-rendered) built to static files, served by Go via `go:embed`.**

Rationale:
- This application is an internal authenticated operator tool — not a public page with SEO or TTFB requirements.
- SvelteKit SSR would require a permanent Node server process alongside Go. Two deployment artifacts, two processes, complex CORS/proxy rules in development.
- The static output model (Vite build → `dist/`) is simple to embed and deploy. Go serves `GET /app*` from the embedded `dist/` directory. Routing is delegated to Svelte's router inside the SPA.
- No Node runtime at deploy time. Node only needed for `npm run build` during the build step.

### 2.2 Auth: same-origin cookies

**Decision: Go serves the Svelte `dist/` files under `/app` on the same origin as `/api`. Existing `HttpOnly` session cookies flow automatically.**

How it works:
- `GET /app` and `GET /app/*` are handled by Go: serve `dist/index.html` (the SPA shell).
- Static assets (`dist/assets/*`) are served by Go from the embedded bundle.
- All API calls from the Svelte client go to `/api/...` on the same origin. No CORS config needed, no token storage in `localStorage`, no XSS exposure from token persistence.
- The existing cookie-based session flow (`POST /api/session/login` sets `workflow_session_id` and `workflow_refresh_token` cookies) works unchanged.
- 401 responses from any API call trigger the Svelte router to redirect to the login view.

### 2.3 Toolchain

| Tool | Role |
|---|---|
| [Vite](https://vite.dev/) | Build tool, dev server |
| [Svelte 5](https://svelte.dev/) | UI framework (runes-based reactivity) |
| [`@svelte-spa-router`](https://github.com/ItalyPaleAle/svelte-spa-router) | Hash-based client-side router |
| TypeScript | Type safety on API contracts |
| Native CSS (scoped per-component) | Styling — no Tailwind, no CSS-in-JS |

> **Router decision: use `@svelte-spa-router` (hash-based routing).** With hash routing, `GET /app` always serves `index.html` and Svelte handles `#/dashboard`, `#/review/inbound-requests`, etc. internally — no Go catch-all configuration needed. `svelte-routing` is an HTML5 history-based library that would require Go to serve `index.html` for every `GET /app/*` path that isn't a static asset; avoid it for this deployment model.

### 2.4 Type contract: Go → TypeScript

A hand-maintained `web/src/lib/api/types.ts` file mirrors the Go JSON response types from `api.go`. This is the first place to update when the API contract changes.

An optional improvement (post-migration) is to generate these types automatically using [tygo](https://github.com/gzuidhof/tygo) or `quicktype` against the Go source. This is not required for the initial build.

---

## 3. Repository structure

```
workflow_app/
├── cmd/
│   └── app/                # Go binary — unchanged
├── internal/               # All Go backend — unchanged
├── web/                    # NEW: Svelte project root
│   ├── src/
│   │   ├── lib/            # Shared library code
│   │   │   ├── api/        # API client and type definitions
│   │   │   ├── components/ # Shared UI components
│   │   │   ├── stores/     # Svelte stores (session, flash, etc.)
│   │   │   └── utils/      # Formatting, date helpers
│   │   ├── routes/         # Route-level page components
│   │   │   ├── auth/
│   │   │   ├── dashboard/
│   │   │   ├── intake/
│   │   │   ├── operations/
│   │   │   ├── review/
│   │   │   ├── inventory/
│   │   │   ├── admin/
│   │   │   └── settings/
│   │   ├── layouts/        # Shared layout components
│   │   ├── App.svelte      # Root app with router
│   │   └── main.ts         # Entry point
│   ├── dist/               # Built output — git-ignored; go:embed target
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── go.mod
├── go.sum
└── Makefile                # Orchestrates npm build + go build
```

### 3.1 Go embedding (additions to existing Go code)

A new file `internal/app/web_static.go` handles Svelte SPA serving. It strips the `/app` URL prefix before passing requests to the file server, and falls back to `index.html` for any non-asset path (hash-based SPA routing). The full implementation is in §10.2.

The `web/dist/` output is copied to `internal/app/web_dist/` during the build step. The Makefile controls this.

### 3.2 Makefile build order

```makefile
.PHONY: build-web build-go build

build-web:
    cd web && npm ci && npm run build
    rm -rf internal/app/web_dist
    cp -r web/dist internal/app/web_dist

build-go:
    go build ./cmd/...

build: build-web build-go

dev-web:
    cd web && npm run dev

dev-go:
    go run ./cmd/app
```

---

## 4. Svelte project internals

### 4.1 Directory map

```
web/src/
├── lib/
│   ├── api/
│   │   ├── client.ts           # Fetch wrapper: auth, error handling, 401 redirect
│   │   ├── types.ts            # All TypeScript types mirroring Go JSON responses
│   │   ├── session.ts          # Session API calls
│   │   ├── inbound.ts          # Inbound request API calls
│   │   ├── review.ts           # Review/reporting API calls
│   │   ├── approvals.ts        # Approval decision API calls
│   │   ├── admin.ts            # Admin API calls (accounting, parties, access, inventory)
│   │   └── agent.ts            # Agent processing API calls
│   ├── components/
│   │   ├── shell/
│   │   │   ├── AppShell.svelte         # Sidebar + TopBar + content slot
│   │   │   ├── TopBar.svelte           # Brand mark + user menu (no nav)
│   │   │   ├── SideNav.svelte          # Fixed left sidebar with nav items
│   │   │   ├── SideNavItem.svelte      # Single sidebar nav entry (icon + label)
│   │   │   └── UserMenu.svelte         # User display + settings/logout dropdown
│   │   ├── layout/
│   │   │   ├── PageHeader.svelte       # Eyebrow + H1 + body paragraph
│   │   │   ├── SectionHead.svelte      # Eyebrow + H2 + optional action slot
│   │   │   ├── TwoUp.svelte            # Two-column grid
│   │   │   ├── ThreeUp.svelte          # Three-column grid
│   │   │   └── PageStack.svelte        # Vertical gap-stacked content
│   │   ├── data/
│   │   │   ├── DataTable.svelte        # Sortable table with optional empty state
│   │   │   ├── StatusPill.svelte       # Colored status badge
│   │   │   ├── SummaryCard.svelte      # Count card with eyebrow and action link
│   │   │   ├── DetailCard.svelte       # Eyebrow + value + optional meta
│   │   │   ├── DetailGrid.svelte       # Auto-fit grid of DetailCard items
│   │   │   ├── ContinuityLinks.svelte  # Compact inline hyperlinks row
│   │   │   ├── EmptyState.svelte       # Empty table / zero-result state
│   │   │   └── JsonBlock.svelte        # Pretty-printed JSON payload viewer
│   │   ├── forms/
│   │   │   ├── FilterPanel.svelte      # Collapsible filter form panel
│   │   │   ├── FormField.svelte        # Label + input with consistent spacing
│   │   │   ├── SelectField.svelte      # Label + select
│   │   │   ├── TextareaField.svelte    # Label + textarea
│   │   │   ├── ButtonRow.svelte        # Flex row of buttons/links
│   │   │   ├── ConfirmAction.svelte    # Confirm dialog before destructive actions
│   │   │   └── FileUpload.svelte       # File attachment input with media-type hint
│   │   ├── feedback/
│   │   │   ├── FlashBanner.svelte      # Notice / error flash message
│   │   │   ├── LoadingSpinner.svelte   # Inline loading indicator
│   │   │   └── ErrorBoundary.svelte    # Route-level error fallback
│   │   └── navigation/
│   │       ├── RouteLinkCard.svelte    # Route directory entry card
│   │       └── BreadcrumbBar.svelte    # Contextual breadcrumb trail
│   ├── stores/
│   │   ├── session.ts          # Writable session context; populated at app start
│   │   ├── flash.ts            # Notice / error message queue (auto-dismiss)
│   │   └── navigation.ts       # Active route tracking for nav active state
│   └── utils/
│       ├── format.ts           # Date formatting, number formatting, status labels
│       ├── routes.ts           # Route path constants and href builders
│       └── status.ts           # Status-to-CSS-class mapping (good/bad/neutral)
├── routes/
│   ├── auth/
│   │   └── Login.svelte        # Login page (public shell)
│   ├── dashboard/
│   │   └── Dashboard.svelte    # Role-aware home with workload summary
│   ├── intake/
│   │   ├── Submit.svelte       # Request intake form
│   │   └── RequestDetail.svelte # Inbound request detail + lifecycle controls
│   ├── operations/
│   │   ├── OperationsLanding.svelte   # Operations hub with queued/pending counts
│   │   ├── OperationsFeed.svelte      # Durable timeline feed
│   │   └── AgentChat.svelte           # Agent guidance request surface
│   ├── review/
│   │   ├── ReviewLanding.svelte
│   │   ├── InboundRequests.svelte     # Filtered inbound request list
│   │   ├── Proposals.svelte           # Filtered proposal list
│   │   ├── ProposalDetail.svelte
│   │   ├── Approvals.svelte           # Filtered approval queue
│   │   ├── ApprovalDetail.svelte
│   │   ├── Documents.svelte           # Filtered document list
│   │   ├── DocumentDetail.svelte
│   │   ├── Accounting.svelte          # Journal + control balances + tax summaries
│   │   ├── AccountingEntryDetail.svelte
│   │   ├── ControlAccountDetail.svelte
│   │   ├── TaxSummaryDetail.svelte
│   │   ├── Inventory.svelte           # Stock + movements + reconciliation
│   │   ├── InventoryMovementDetail.svelte
│   │   ├── InventoryItemDetail.svelte
│   │   ├── InventoryLocationDetail.svelte
│   │   ├── WorkOrders.svelte
│   │   ├── WorkOrderDetail.svelte
│   │   ├── Audit.svelte
│   │   └── AuditEventDetail.svelte
│   ├── inventory/
│   │   └── InventoryLanding.svelte    # Inventory hub
│   ├── admin/
│   │   ├── AdminHub.svelte            # Privileged maintenance hub
│   │   ├── AdminAccounting.svelte
│   │   ├── AdminParties.svelte
│   │   ├── AdminPartyDetail.svelte
│   │   ├── AdminAccess.svelte
│   │   └── AdminInventory.svelte
│   ├── settings/
│   │   └── Settings.svelte
│   └── utility/
│       └── RouteCatalog.svelte        # Searchable route discovery
├── layouts/
│   ├── AppLayout.svelte        # Authenticated shell (AppShell + slot)
│   └── PublicLayout.svelte     # Public shell for login page (no sidebar)
├── App.svelte
└── main.ts
```

---

## 5. Design system

The design system — color tokens, typography scale, spacing, whitespace doctrine, navigation model, and enterprise UI principles — is defined in the canonical companion document:

> **[`docs/svelte_web_guides/web_ui_design_guide.md`](web_ui_design_guide.md)** — read this before building any component.

Key decisions from that guide that directly affect Svelte implementation:

| Decision | Impact on implementation |
|---|---|
| **Left sidebar navigation** (§3) | `AppShell` is a sidebar layout, not a top nav-strip. `NavBubbles.svelte` is removed. `SideNav.svelte` and `SideNavItem.svelte` replace it. `AdminLayout.svelte` is eliminated — admin sub-nav is a collapsible group in the sidebar. |
| **Max content width `1100px`** (§6.1) | The content area inside `AppLayout` has `max-width: var(--content-max-width)` centered. |
| **Minimum panel padding `24px`** (§6.2) | Every `<section>`, card, or panel component uses `padding: var(--panel-padding)` (24px) as a floor. |
| **Typography scale** (§4) | Add `--text-2xl` through `--text-2xs` tokens to `:root`. Use the usage table from §4.3 for every heading and body text element. |
| **Toast notifications** (§9.6) | `FlashBanner.svelte` renders as a toast stack in the top-right corner, not as an inline page section. |
| **Progressive disclosure** (§9.5) | Detail page sub-sections (AI runs, attachments, audit trail) start collapsed. |
| **Warn token added** | Add `--warn: #7a5418` and `--warn-soft: #fdf2e0` to the token set for pending/caution states. |
| **Shadow scale updated** (§5.1) | Layered ambient + directional shadow. Replace the current two-value scale with the three-level scale from the guide. |

### 5.1 Visual improvements table

The migration must fix these specific behavioral and visual problems in the current Go template UI:

| Issue | Current behavior | Svelte target |
|---|---|---|
| Double header band before content | TopBar + nav bubble strip consuming ~120px | Single TopBar (48px) + sidebar (zero vertical cost) |
| Navigation labels itself | Strip headed "Workflow destinations" | Sidebar items self-evident; no labels on the nav container |
| Flash messages as inline page sections | Full-width banner pushing content down | Toast notifications, top-right, auto-dismiss 4s |
| No loading state on data fetches | Full-page blank during navigation | `LoadingSpinner` during API calls; skeleton rows in `DataTable` |
| Filter forms always visible | Filters occupy prime viewport space at all times | Collapsed by default on detail pages; open by default on list pages |
| Approval decisions use raw form POST | No confirmation; outcome only visible after reload | `ConfirmAction` modal before submission; toast on completion |
| Continuity links are dense inline-link rows | Hard to scan; takes full table cell | `ContinuityLinks` component with icon-differentiated link types |
| Admin pages mix create form + list | Noisy; disorienting | Create in a collapsible right-side panel; list table is primary |
| Detail page sub-sections always expanded | Wall of sections before reaching actions | Progressive disclosure: AI runs, attachments, audit collapsed by default |
| Detail pages use `<pre>` for JSON payloads | Monospace walls without expand/collapse | `JsonBlock.svelte`: collapsible, syntax-highlighted, copyable |
| No back-navigation continuity | Browser back works but loses filter state | Filter state synced to URL query params |
| No breadcrumbs | Only page heading for hierarchy | `BreadcrumbBar.svelte` injected by each route page |

---

## 6. API client design

### 6.1 Base client (`web/src/lib/api/client.ts`)

```typescript
// web/src/lib/api/client.ts
// On 401: clear session store, redirect to login.
// On 5xx: surface error to flash store.
import { session } from '$lib/stores/session';
import { push } from 'svelte-spa-router';

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

const BASE = '';  // Same origin — no CORS, no token storage in localStorage

export async function apiFetch<T>(
  path: string,
  init?: RequestInit
): Promise<T> {
  const res = await fetch(BASE + path, {
    ...init,
    credentials: 'same-origin',  // Sends HttpOnly cookies automatically
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (res.status === 401) {
    // Clear session store and redirect to login via @svelte-spa-router
    session.set(null);        // writable store — use .set(null), not .clear()
    push('/login');           // push() from @svelte-spa-router — not SvelteKit's goto()
    throw new Error('Unauthorized');
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error ?? 'Unknown error');
  }

  return res.json() as Promise<T>;
}
```

### 6.2 Types file structure (`web/src/lib/api/types.ts`)

This single file mirrors all Go JSON response types. Kept in one place to make drift visible.

```typescript
// Session
export interface SessionContext {
  user_id: string;
  user_email: string;
  user_display_name: string;
  org_id: string;
  org_slug: string;
  org_name: string;
  role_code: string;
  session_id: string;
}

// Inbound requests
export interface InboundRequestReview {
  request_id: string;
  request_reference: string;
  status: string;
  channel: string;
  message_count: number;
  attachment_count: number;
  failure_reason: string;
  cancellation_reason: string;
  last_run_id: string | null;
  last_recommendation_id: string | null;
  last_recommendation_status: string | null;
  updated_at: string;
}

// ... all other types matching api.go response structs ...
```

### 6.3 API module per domain

Each API module (`session.ts`, `inbound.ts`, `review.ts`, `approvals.ts`, `admin.ts`) exports typed async functions that call `apiFetch`. Example:

```typescript
// review.ts
export async function listInboundRequests(params: {
  status?: string;
  request_reference?: string;
}): Promise<InboundRequestReview[]> {
  const qs = new URLSearchParams(
    Object.fromEntries(
      Object.entries(params).filter(([, v]) => v !== undefined && v !== '')
    )
  );
  return apiFetch(`/api/review/inbound-requests?${qs}`);
}
```

---

## 7. Component specifications

### 7.1 `AppShell.svelte`

The application shell uses a **left sidebar layout** as defined in `docs/svelte_web_guides/web_ui_design_guide.md` §3. This is a fundamental departure from the current top nav-strip model.

Structure:
```
┌─────────────────────────────────────────────────────┐
│ TopBar (48px fixed height, full width)               │
├──────────┬──────────────────────────────────────────┤
│ SideNav  │ <slot> — page content                    │
│ (220px   │                                          │
│  fixed)  │ max-width: var(--content-max-width)      │
│          │ padding: var(--page-gutter)              │
└──────────┴──────────────────────────────────────────┘
```

Components:
- **`TopBar.svelte`** — brand mark (left) + `UserMenu` (right). **No navigation in the top bar.**
- **`SideNav.svelte`** — fixed 220px sidebar, dark background (`var(--shell-ink)`). Contains `SideNavItem` entries for: Home, Intake, Operations, Review, Inventory, then a separator, then Admin (collapsible group with sub-items), Settings.
- **`SideNavItem.svelte`** — single item: optional 16px icon + text label. Active state from router match.
- **`UserMenu.svelte`** — user display name + role pill in TopBar; dropdown with Settings link and Sign out.

> **`NavBubbles.svelte` is not built.** The top nav-strip approach is replaced entirely by the sidebar. **`AdminLayout.svelte` is not built** — admin sub-navigation is a collapsible group within `SideNav`.

Props:
- `activePath: string` — derived from router; passed to `SideNav` for active-state highlighting

Content:
- Default children snippet: page content. In Svelte 5, render with `{@render children()}` inside the shell. Do not use `<slot />`.

### 7.2 `DataTable.svelte`

The most-reused component. Replaces all `<div class="table-wrap"><table>` patterns.

Props:
```typescript
interface Column<T> {
  key: keyof T | string;
  label: string;
  align?: 'left' | 'right';
  width?: string;
  // For plain text/formatted values. For complex cell content, use the rowActions snippet.
  render?: (row: T) => string;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  rows: T[];
  emptyTitle?: string;
  emptyBody?: string;
  loading?: boolean;
  // Svelte 5 snippet for per-row action buttons — replaces Svelte 4 named slot.
  // Usage: <DataTable ...>{#snippet rowActions(row)}<button>...</button>{/snippet}</DataTable>
  rowActions?: import('svelte').Snippet<[T]>;
}
```

Features:
- Sticky `<thead>` on scroll
- Row hover highlight
- Empty state (uses `EmptyState` component when `rows.length === 0` and `loading === false`)
- Loading skeleton rows (3 placeholder rows during fetch)
- `rowActions` snippet prop for per-row action buttons (Svelte 5 — not a named slot)

### 7.3 `FilterPanel.svelte`

Replaces all `<section class="panel section-stack">` filter forms.

Props:
- `title: string`
- `initialOpen?: boolean` — collapsed by default on detail pages, open by default on list pages

Content:
- Default children snippet: form fields (using `FormField`, `SelectField`, etc.)
- `actions` snippet prop: submit and clear buttons

Behavior: On mobile, collapses to a "Show filters" disclosure. Always shows current active filter count as a badge when collapsed.

### 7.4 `SummaryCard.svelte`

Replaces `{{template "web_summary_card" ...}}`. Statically typed props:

```typescript
interface SummaryCardProps {
  label: string;
  value: number | string;
  body?: string;
  actionHref?: string;
  actionLabel?: string;
}
```

### 7.5 `ContinuityLinks.svelte`

Replaces the pattern of `<td class="inline-links"><a ...>...<a ...>...</td>` with a typed, icon-differentiated component.

```typescript
interface ContinuityLink {
  label: string;
  href: string;
  kind?: 'request' | 'proposal' | 'approval' | 'document' | 'entry' | 'audit' | 'default';
}

interface ContinuityLinksProps {
  links: ContinuityLink[];
}
```

Renders as a compact flex-wrap row. Each link kind gets a small icon prefix and appropriate color treatment.

### 7.6 `ConfirmAction.svelte`

A modal dialog to gate approval and destructive actions. Replaces bare `<form method="post">` on approval decision forms.

```typescript
interface ConfirmActionProps {
  title: string;
  body: string;
  confirmLabel?: string;  // Default: "Confirm"
  cancelLabel?: string;   // Default: "Cancel"
  danger?: boolean;       // Red confirm button
  onConfirm: () => Promise<void>;
}
```

Uses `<dialog>` element. Accessible. Focus-trapped when open.

### 7.7 `FlashBanner.svelte`

Subscribes to `flash.ts` store. Renders notice/error messages as dismissible toasts in a fixed position top-right stack. Auto-dismissed after 4 seconds. At most 3 visible at once.

---

## 8. Route pages — implementation notes

### 8.1 Dashboard (`Dashboard.svelte`)

> **Backend gap — requires a new API endpoint before Phase 1 can complete.**

The current Go dashboard handler calls `reviewService.GetDashboardSnapshot()`, a single batched service call that returns: request/proposal status summaries, pending approvals, recent requests, and recent proposals — all in one round-trip. This method **exists in `reporting.Service` but is not exposed as a JSON API endpoint**.

Before implementing `Dashboard.svelte`, add:
```
GET /api/review/dashboard-snapshot
```
This endpoint calls `GetDashboardSnapshot(ctx, actor, 10, 20, 10)` and returns the full snapshot as JSON. The Svelte dashboard then makes one call and distributes the payload to its child components.

Alternative (not recommended): assemble the dashboard by making three independent parallel API calls (`status-summary`, `proposal-summary`, `approval-queue`). This misses the recent requests and proposals panels and multiplies latency.

Once the endpoint exists, `Dashboard.svelte` renders:
- Role-aware heading block (`role_code` from `session` store determines copy via `roleAwareHomeIntro` logic ported to TypeScript)
- Status summary grid (`SummaryCard` per status — clicking pre-filters the list page)
- Role-and-workload-aware primary/secondary action cards (port `buildHomeActions` logic to TypeScript)
- Recent requests and proposals tables

Improvement over current: status cards are clickable, action cards badge live counts, and the surface feels alive rather than static.

### 8.2 Inbound request list (`InboundRequests.svelte`)

Filter state is synced to URL query params (`?status=queued&request_reference=REQ-001`). This means:
- The filter form is reactive (no full POST needed)
- Filtered URLs are shareable
- Browser back/forward works correctly

Data fetched on filter-change with debounce (300ms). Loading state via `DataTable` skeleton.

### 8.3 Inbound request detail (`RequestDetail.svelte`)

Contains multiple sub-sections rendered only when data is present:
- Primary hero card (status, channel, received-at)
- Lifecycle control panel (draft/queued/cancelled-specific actions via `ConfirmAction`)
- Messages and attachments (`TwoUp`)
- AI runs, steps, artifacts (`ThreeUp`)
- Delegations, recommendations, proposals (`ThreeUp`)

The draft edit form is a dedicated collapsible section — no longer requires a separate page load.

### 8.4 Approval detail (`ApprovalDetail.svelte`)

The approval decision panel (Approve/Reject) renders:
1. Current approval status pill
2. Decision note input
3. **Confirm before submit**: clicking Approve/Reject opens a `ConfirmAction` dialog with the decision summary
4. On confirm: calls `POST /api/approvals/{id}/decision`, shows flash message, refreshes page data inline

This eliminates the current full-page POST-redirect cycle for approval decisions.

### 8.5 Accounting review (`Accounting.svelte`)

The complex multi-pivot accounting page (journal entries + control balances + tax summaries on one surface) benefits from Svelte tabs:
- Three tab panels: **Journal entries**, **Control balances**, **Tax summaries**
- Shared filter form at top; each tab shows filtered results for its pivot
- Tab counts shown as badges reflecting current filter
- No more `<div class="two-up">` forcing balances and tax side-by-side below the journal table

### 8.6 Admin pages (`AdminAccounting.svelte`, `AdminParties.svelte`, etc.)

Current pattern: create form above the list table on one page.

Svelte improvement: two-panel layout within the standard app shell:
- Left: list table (primary content, full width on mobile)
- Right: create/edit form in a secondary panel (collapses to a collapsible section or `<details>` on mobile)

Admin pages use the standard `AppLayout.svelte` — **not** a separate `AdminLayout`. Admin sub-navigation (Accounting setup, Party setup, Access, Inventory setup) is handled by the collapsible Admin group in `SideNav`. There is no second navigation strip on admin pages.

### 8.7 Route catalog (`RouteCatalog.svelte`)

Route search becomes reactive: as the user types, results filter client-side from a static catalog embedded in the component. No server round-trip needed for route discovery since the catalog is a fixed list.

---

## 9. Session and auth flow

### 9.1 App startup

On app mount, `App.svelte` calls `GET /api/session`:
- **200 OK**: populate `session` store; route to the requested page (or dashboard)
- **401**: route to Login page

### 9.2 Login flow

`Login.svelte` (rendered with `PublicLayout.svelte`):
1. User submits email, password, org slug.
2. `POST /api/session/login` with JSON body `{org_slug, email, password, device_label}`.
3. Go sets `HttpOnly` cookies and **returns the `SessionContext` directly in the 201 response body**. No second GET to `/api/session` is required on login.
4. `session` store populated from the login response body.
5. Router navigates to `#/dashboard` via `push('/dashboard')`.

Note: `GET /api/session` is still used on **app startup** (§9.1) to restore an existing session from cookies. It is not an extra step after login.

### 9.3 Logout

User clicks Sign out:
1. `POST /api/session/logout` (Go clears cookies)
2. `session` store cleared via `session.set(null)`
3. Router navigates to login via `push('/login')` (hash-based — same SPA context)

### 9.4 401 handling

Any API call returning 401:
1. `apiFetch` intercepts
2. Clears `session` store
3. Saves current route to `flash` store as "Session expired, please sign in again"
4. Navigates to `/login`

---

## 10. Go backend changes required

The Go backend changes split into two parts: (a) new snapshot API endpoints needed before certain Svelte phases, and (b) static file serving for the SPA.

### 10.1 New API endpoints required (additive)

Five snapshot service methods exist in `reporting.Service` but are not yet exposed as JSON endpoints. These must be added **before** the corresponding Svelte pages are built:

| New endpoint | Service method | Required for |
|---|---|---|
| `GET /api/review/dashboard-snapshot` | `GetDashboardSnapshot(ctx, actor, 10, 20, 10)` | Phase 1 — Dashboard |
| `GET /api/review/operations-feed-snapshot` | `GetOperationsFeedSnapshot(ctx, actor, 20)` + merge/sort | Phase 3 — OperationsFeed |
| `GET /api/review/operations-landing-snapshot` | `GetOperationsLandingSnapshot(ctx, actor, 10, 20)` | Phase 3 — OperationsLanding |
| `GET /api/review/agent-chat-snapshot` | `GetAgentChatSnapshot(ctx, actor, 40, 40)` | Phase 3 — AgentChat |
| `GET /api/review/inventory-landing-snapshot` | `GetInventoryLandingSnapshot(ctx, actor, 20)` | Phase 2 — InventoryLanding |

For the operations feed endpoint specifically, the Go merge-and-sort logic from `handleWebOperationsFeed` (building `webOperationsFeedItem` slices then sorting by `OccurredAt`) should stay server-side and be returned as a unified sorted JSON array. Do not push this merge logic into Svelte.

### 10.2 Static file serving (`internal/app/web_static.go` — NEW)

The embedded Svelte build is served under the `/app` prefix. The handler must strip the `/app` prefix before passing to the file server, then fall back to `index.html` for any non-asset path (SPA shell behaviour):

```go
package app

import (
    "embed"
    "io/fs"
    "net/http"
    "strings"
)

//go:embed web_dist
var webDist embed.FS

func newWebStaticHandler() http.Handler {
    sub, _ := fs.Sub(webDist, "web_dist")
    fileServer := http.FileServer(http.FS(sub))
    // Strip /app prefix so the file server sees paths relative to web_dist root.
    stripped := http.StripPrefix("/app", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // After prefix stripping, asset paths look like /assets/index-abc123.js.
        // Non-asset paths (router routes) get index.html — the SPA shell.
        if !strings.HasPrefix(r.URL.Path, "/assets/") && r.URL.Path != "/" {
            r2 := r.Clone(r.Context())
            r2.URL.Path = "/"
            fileServer.ServeHTTP(w, r2)
            return
        }
        fileServer.ServeHTTP(w, r)
    }))
    return stripped
}
```

### 10.3 Route registration changes (`internal/app/api.go`)

The SPA replaces all `web_*.go` handler registrations. The mux changes are:

**Remove all `/app/*` web handler registrations** (lines 556–607 in current `api.go`).

**Add SPA catch-all handler** for `GET /app` and `GET /app/*`:

```go
webStatic := newWebStaticHandler()
mux.Handle("/app/", webStatic)
mux.Handle("/app", webStatic)
```

**All existing `/api/*` routes remain unchanged.**

### 10.4 Vite dev proxy config (`web/vite.config.ts`)

During development, the Vite dev server runs on `localhost:5173` and the Go server on `localhost:8080`. The Vite proxy forwards all `/api` requests from the browser to Go — from Go's perspective the request arrives from `localhost:5173` but with `changeOrigin: true` the Host header is rewritten to `localhost:8080`. **No CORS configuration is needed on the Go side**: the browser sends API requests to Vite (same origin as the dev page), and Vite proxies them. In production, Go serves everything from one origin — CORS is equally unnecessary.

```typescript
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
});
```

Developers run both `make dev-go` and `make dev-web` simultaneously. The browser talks only to Vite at port 5173; Vite proxies `/api/*` calls to Go at port 8080.

---

## 11. What gets deleted from the current codebase

Once the Svelte SPA covers all routes, the following Go code is deleted:

| File | Size | Deletion reason |
|---|---|---|
| `internal/app/web.go` | 34 KB | All page-data structs and template helpers replaced by TypeScript types |
| `internal/app/web_access_handlers.go` | 4 KB | Replaced by Svelte login page |
| `internal/app/web_admin_handlers.go` | 27 KB | Replaced by Svelte admin pages |
| `internal/app/web_bundle_handlers.go` | 5 KB | Template bundle loading replaced by static file serving |
| `internal/app/web_navigation_handlers.go` | 22 KB | Replaced by Svelte operations/inventory/review landing pages |
| `internal/app/web_review_handlers.go` | 16 KB | Replaced by Svelte review pages |
| `internal/app/web_review_more_handlers.go` | 17 KB | Replaced by Svelte review detail pages |
| `internal/app/web_session_inbound_handlers.go` | 23 KB | Replaced by Svelte intake and lifecycle pages |
| `internal/app/web_templates/` (dir) | 163 KB | All templates replaced by Svelte components |
| `internal/app/web_templates_bundle.go` | 10 KB | Template parsing no longer needed |
| `internal/app/web_test.go` | 176 KB | HTML rendering tests fully replaced — see section 13 |

**Total deletion: ~497 KB of Go code and templates.**

**Pre-deletion checklist — logic that currently exists only in the web layer:**

| Logic | Current location | Migration path |
|---|---|---|
| `buildOperationsFeedFromRequests/Proposals/Approvals` + sort | `web_session_inbound_handlers.go` | Move into new `GET /api/review/operations-feed-snapshot` handler |
| `buildHomeActions()` role-aware action prioritization | `web_navigation_handlers.go` | Port to TypeScript in `Dashboard.svelte` (pure display logic) |
| `roleAwareHomeIntro()` | `web_navigation_handlers.go` | Port to TypeScript |
| `routeCatalogEntries()` + `filterRouteCatalogEntries()` scoring | `web_navigation_handlers.go` | Port full scoring algorithm to TypeScript in `RouteCatalog.svelte` |
| `editableInboundMessageID/Text()` | `web.go` | Already backed by `GET /api/review/inbound-requests/{id}` detail which includes message data — port extraction logic to TypeScript |
| `parseMultipartAttachments()` | `web.go` | Not needed — JSON API uses base64 attachments, not multipart |
| `inboundRequestMetadataString()` | `web.go` | Port to TypeScript (simple map key extraction) |

None of these prevent deletion — they are either pure display logic portworthy to TypeScript, or replaced by new API endpoints. Do not delete any `web_*.go` files until the corresponding API endpoints or TypeScript equivalents are verified.

---

## 12. Phased execution plan

### Phase 1: Foundation (week 1–2)

**Goal**: Running Svelte project with login, session management, and dashboard.

Steps:
1. Initialize `web/` Svelte + Vite + TypeScript project
2. Implement `app.css` with full design token set from `docs/svelte_web_guides/web_ui_design_guide.md` §5; import IBM Plex Sans
3. Build `AppShell`, `TopBar`, `SideNav`, `SideNavItem`, `UserMenu` shell components (sidebar layout — no `NavBubbles`); build `PublicLayout`
4. Build `FlashBanner` (toast stack, top-right), `LoadingSpinner`, `EmptyState` feedback components
5. Implement `client.ts`, `session.ts`, `types.ts` in `lib/api/`
6. Implement `session` and `flash` Svelte stores
7. Build `Login.svelte` (calls `POST /api/session/login`; no second GET needed — response body contains session context)
8. Add `GET /api/review/dashboard-snapshot` to Go backend (§10.1). Build `Dashboard.svelte` using this single endpoint.
9. Add Go `web_static.go` (prefix-stripping handler from §10.2); register SPA handler on `/app/` and `/app`
10. Verify: login → dashboard → logout cycle works end-to-end

**Acceptance**: Login, session persistence, dashboard with real data, logout — all working against the Go backend.

### Phase 2: Review workbench (week 3–4)

**Goal**: All review list and detail pages working.

Steps:
1. Build shared components: `DataTable`, `StatusPill`, `SummaryCard`, `FilterPanel`, `ContinuityLinks`, `DetailCard`, `DetailGrid`
2. Build `review.ts` API module (all list and detail calls)
3. Build all review list pages: `InboundRequests`, `Proposals`, `Approvals`, `Documents`, `Accounting`, `Inventory`, `WorkOrders`, `Audit`
4. Build all review detail pages
5. Build `ReviewLanding.svelte`
6. Implement URL-param-synced filter state on all list pages
7. Implement `ApprovalDetail.svelte` with `ConfirmAction` for approve/reject

**Acceptance**: All review routes render correct data; filter state syncs to URL; approval decisions work with confirmation dialog.

### Phase 3: Intake and operations (week 5)

**Goal**: Intake submission, request detail lifecycle controls, operations pages.

Steps:
1. Add `GET /api/review/operations-feed-snapshot` to Go backend (§10.1 — includes merge/sort requirement).
2. Add `GET /api/review/agent-chat-snapshot` to Go backend (§10.1).
3. Build `inbound.ts` and `agent.ts` API modules.
4. Build `Submit.svelte` — uses `POST /api/inbound-requests` with **JSON body** (not multipart). Attachments must be base64-encoded client-side before submission. Use `FileUpload.svelte` to read files via `FileReader.readAsDataURL()` and strip the data-URI prefix before populating the `attachments[]` array.
5. Build `RequestDetail.svelte` with full lifecycle section:
   - Draft: save via `PUT /api/inbound-requests/{id}/draft`, queue via `POST /api/inbound-requests/{id}/queue`
   - Queued: cancel via `POST /api/inbound-requests/{id}/cancel`, amend via `POST /api/inbound-requests/{id}/amend`
   - Delete draft: **`DELETE /api/inbound-requests/{id}/delete`** (the JSON API uses HTTP `DELETE`, not `POST`)
6. Build `OperationsLanding.svelte`, `OperationsFeed.svelte`, `AgentChat.svelte`.
7. Build `FormField`, `TextareaField`, `FileUpload`, `ButtonRow` form components.

**Acceptance**: New request submission (JSON with base64 attachments), draft editing, queue, cancel, amend, delete — all working. Operations feed renders correct timeline.

### Phase 4: Admin and settings (week 6)

**Goal**: All admin maintenance pages and settings.

Steps:
1. Build `admin.ts` API module
2. Build all admin pages using standard `AppLayout`: `AdminHub`, `AdminAccounting`, `AdminParties`, `AdminPartyDetail`, `AdminAccess`, `AdminInventory`
3. Admin sub-navigation (between admin pages) is provided by the collapsible Admin section in `SideNav` — no separate `AdminLayout` needed
4. Build `Settings.svelte`
5. Build `RouteCatalog.svelte` with client-side route search
6. Implement `ConfirmAction` on status toggle controls (mark active/inactive)

**Acceptance**: All admin routes working; create and status-toggle flows operational; admin-only access gate enforced client-side (backed by server 403).

### Phase 5: Deletion and cleanup (week 7)

**Goal**: Remove all old Go template code; verify the backend is clean.

Steps:
1. Audit that all API endpoints used by the old template layer have JSON equivalents
2. Identify any data currently assembled only server-side (e.g., operations feed build logic in `web.go`) — confirm the relevant `/api/...` endpoints return equivalent data
3. Delete all `web_*.go` handler files, `web_templates/`, `web_templates_bundle.go`, `web.go`
4. Remove now-dead mux handler registrations from `api.go`
5. Run `go build ./...` and `go vet ./...` to confirm clean build
6. Update `Makefile`, `README.md`, and `AGENTS.md`

**Acceptance**: `go build ./...` passes; no references to deleted code; complete SPA works end-to-end.

---

## 13. Test strategy

### 13.1 What `web_test.go` tested and what replaces it

`internal/app/web_test.go` (176 KB) tests HTTP response codes, page content snippets, cookie behavior, redirect logic, and session gating purely at the HTTP response level.

After the migration, the HTML rendering layer no longer exists in Go, so these tests become irrelevant. They should be deleted in Phase 5.

Replacement strategy:

| Old coverage | Replacement |
|---|---|
| HTTP 200 for authenticated pages | **No Go test needed** — Go now only serves static files; API endpoints already have coverage |
| HTTP 302 redirect for unauthenticated web routes | **No Go test needed** — auth redirect is handled client-side by the Svelte `session` store |
| Cookie set/clear on login/logout | **Existing** `api_integration_test.go` coverage for `POST /api/session/login` and `POST /api/session/logout` |
| Page content assertions (template rendering) | **Svelte component tests** via Vitest + `@testing-library/svelte` |
| Form POST handling for lifecycle actions | **Existing** `api_integration_test.go` coverage for API endpoints |

### 13.2 Svelte component testing

Use [Vitest](https://vitest.dev/) + [`@testing-library/svelte`](https://testing-library.com/docs/svelte-testing-library/intro/) for component-level testing.

Priorities for test coverage:
1. `StatusPill` — correct CSS class for each status string
2. `DataTable` — empty state renders when rows is empty; skeleton renders when loading
3. `FilterPanel` — filter values sync to URL params
4. `apiFetch` — 401 triggers session clear and redirect; 5xx surfaces flash error
5. `session.ts` store — populated and cleared correctly
6. `Login.svelte` — form validation; API error displayed in flash

### 13.3 End-to-end validation

The existing `docs/workflows/end_to_end_validation_checklist.md` should be updated to reflect the Svelte frontend after Phase 5. The same operator workflow checklist applies; only the browser surface changes.

---

## 14. Canonical doc updates required

These documents must be updated **after** the migration is complete (not before — changes are not yet implemented):

| Document | Required change |
|---|---|
| `new_app_docs/new_app_implementation_defaults.md` | Update rules 2.9.6, 2.9.10, 2.9.11 to reflect Svelte SPA as the new canonical web stack |
| `new_app_docs/new_app_tracker.md` | Add new milestone entry for Svelte migration |
| `AGENTS.md` | Update Architecture & Scope Guardrails; remove Go html/template preference |
| `README.md` | Update setup instructions, build commands, architecture description |
| `docs/workflows/end_to_end_validation_checklist.md` | Update browser validation steps to reflect Svelte SPA |

**Do not update the canonical planning docs before implementation is complete.** The current defaults remain authoritative until the migration is verified in production.

---

## 15. Resolved questions and decisions

All questions from the initial draft have been resolved by validating against the codebase:

1. **Router choice**: **Decided — `@svelte-spa-router` (hash-based).** This eliminates any Go catch-all configuration. `svelte-routing` was removed from the toolchain table as it is HTML5-history-based and incompatible with this deployment shape.

2. **Svelte version**: **Decided — Svelte 5.** Current stable release; use runes-based reactivity throughout.

3. **Type generation automation**: **Decided — manual `types.ts` initially.** Revisit automation with `tygo` after migration stabilizes.

4. **Snapshot API endpoints — confirmed as gaps, resolved in §10.1**: Five reporting service methods (`GetDashboardSnapshot`, `GetOperationsFeedSnapshot`, `GetOperationsLandingSnapshot`, `GetAgentChatSnapshot`, `GetInventoryLandingSnapshot`) exist in `reporting.Service` but are not wired to any JSON API route. They are used exclusively by Go web handlers being deleted. These five endpoints have been added to the Go backend changes plan (§10.1) with explicit phase sequencing.

5. **Attachment submission format**: The JSON API (`POST /api/inbound-requests`) requires **base64-encoded attachments in JSON body**. The old web handler used multipart form upload — that approach is being deleted with the web layer. `FileUpload.svelte` must use the browser's `FileReader.readAsDataURL()` API and strip the data-URI prefix before populating the JSON request.

6. **Delete draft HTTP method**: The JSON API handler for draft deletion uses `DELETE /api/inbound-requests/{id}/delete`, not `POST`. The Svelte client must issue an HTTP `DELETE` for this action.

7. **CORS**: Not needed. Confirmed that Vite proxy rewrites the Host header; Go never sees a cross-origin request in development. No CORS middleware required on the Go side in any environment.

---

## 16. Design principles for the Svelte layer

These principles govern all implementation decisions. For the full design rationale, visual standards, and anti-patterns, see [`docs/svelte_web_guides/web_ui_design_guide.md`](web_ui_design_guide.md).

1. **Left sidebar, not top-strip nav.** `AppShell` uses a fixed 220px sidebar. The TopBar contains only brand and user menu. No nav bubbles, no double header band.

2. **One API client, typed responses.** All API calls go through `apiFetch`. All response types live in `types.ts`. No ad-hoc fetch calls in page components.

3. **URL-first filter state.** All filterable pages sync their filter state to URL query params so filtered views are shareable and navigable.

4. **No business logic in components.** Components render data and dispatch events. They do not contain business rules, status machine logic, or entity-relationship knowledge.

5. **Confirm before side effects.** Any action that changes persistent state (approve, reject, cancel, delete, mark inactive) must use `ConfirmAction` to prevent accidental submissions.

6. **Loading visible everywhere.** Any component waiting for API data shows either a skeleton or spinner. No blank intermediate states.

7. **Toast notifications, not inline banners.** Transient feedback (draft saved, approval submitted) goes to the `flash` store and renders as auto-dismissing toasts. Inline banners are reserved for persistent load errors only.

8. **One primary job per page.** The primary content block is visually dominant. Secondary sections are smaller, lower contrast, or collapsed. Tertiary content (AI traces, JSON payloads, audit detail) is collapsed by default.

9. **Minimum panel padding `24px`.** Every card or panel that contains content has at least `var(--panel-padding)` (24px) inner padding. Sub-24px padding is not acceptable.

10. **Typography scale, not ad-hoc sizes.** Use only the `--text-2xl` through `--text-2xs` tokens defined in the design guide. Do not introduce arbitrary font sizes.

11. **Component props over slots for simple cases.** Reserve slots for true layout composition. Data display components take typed props.

12. **Shared layout, not duplicated shell.** Every authenticated page uses `AppLayout`. No `AdminLayout` wrapper — admin navigation is a collapsible group in the sidebar.

13. **CSS tokens only.** No component uses hardcoded hex values. Everything references `:root` custom properties. This makes theme changes a single-file edit.

14. **Progressive disclosure for complex detail pages.** AI run traces, artifacts, delegation details, attachment bodies — all start collapsed. Operators expand what they need.

15. **Filter panels: collapsed on detail pages, open on list pages.** List pages are for finding; filters should be ready. Detail pages are for reading; filters don't need screen space.

16. **Deletion-first.** When old Go template code is replaced, delete it immediately. Do not carry dead code alongside the Svelte implementation.
