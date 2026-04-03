# Svelte 5 Guide for workflow_app

Date: 2026-04-03  
Status: Active — reference for all Svelte implementation work  
Scope: All components and pages in `web/src/`. Apply these patterns consistently throughout.

> **Why this document exists:** Many AI coding tools default to Svelte 4 syntax, which produces broken or deprecated code in a Svelte 5 project. This guide specifies the correct Svelte 5 runes syntax and the specific patterns used in this application. Always reference this document before generating or reviewing Svelte code.

---

## 1. Svelte 4 → Svelte 5 quick reference

This is the most-needed section. If an AI tool generates any Svelte 4 pattern from the left column, replace it with the Svelte 5 equivalent from the right.

| Concept | ❌ Svelte 4 (do not use) | ✅ Svelte 5 runes (always use) |
|---|---|---|
| Reactive state | `let count = 0;` | `let count = $state(0);` |
| Derived value | `$: doubled = count * 2;` | `let doubled = $derived(count * 2);` |
| Side effect | `$: { doSomething(count); }` | `$effect(() => { doSomething(count); });` |
| Component prop | `export let name: string;` | `let { name } = $props();` |
| Prop with default | `export let size = 'md';` | `let { size = 'md' } = $props();` |
| Bindable prop | `export let value = '';` | `let { value = $bindable('') } = $props();` |
| Click event | `on:click={handler}` | `onclick={handler}` |
| Submit event | `on:submit|preventDefault={handler}` | `onsubmit={(e) => { e.preventDefault(); handler(e); }}` |
| Component event | `createEventDispatcher` + `on:customevent` | Callback prop: `let { onSubmit } = $props();` then `onclick={onSubmit}` |
| Mount lifecycle | `onMount(() => { ... })` | `$effect(() => { ... })` — or keep `onMount` for async init |
| Before/after update | `beforeUpdate` / `afterUpdate` | `$effect` (runs after state changes) |
| Store auto-subscribe | `$store` | `$store` — **unchanged, still works the same** |

> **Key rule:** If you see `export let`, `$:`, `on:`, or `createEventDispatcher` in a Svelte file — it is Svelte 4 syntax. Replace it.

---

## 2. Runes reference

### 2.1 `$state` — reactive variables

Declares reactive state. Any mutation triggers a re-render.

```svelte
<script lang="ts">
  let count = $state(0);
  let loading = $state(false);
  let data = $state<InboundRequestReview[]>([]);
  let error = $state<string | null>(null);
</script>
```

**Objects and arrays:** Svelte 5 makes objects and arrays deeply reactive when declared with `$state`. You can either mutate properties directly or reassign the whole variable — both work and both trigger reactivity. What does NOT work is mutating a **destructured copy** outside of `$state`.

```svelte
<script lang="ts">
  let filters = $state({ status: '', reference: '' });

  // ✅ Both of these work correctly in Svelte 5:

  // Option A — mutate property directly (preferred for targeted changes)
  function clearStatus() {
    filters.status = '';
  }

  // Option B — reassign the whole object (also valid; reactive)
  function clearAll() {
    filters = { status: '', reference: '' };
  }

  // ❌ Wrong — destructuring escapes the reactive proxy
  let { status } = filters;  // 'status' is now a plain string, not reactive
  status = 'queued';         // This does NOT update filters.status
</script>
```

**Exception:** Primitive reassignment is fine:
```svelte
let count = $state(0);
count++;       // ✅ fine — primitive replacement
count = 42;    // ✅ fine
```

### 2.2 `$derived` — computed values

Replaces `$: derived = expression`. Recalculates whenever its dependencies change.

```svelte
<script lang="ts">
  let rows = $state<InboundRequestReview[]>([]);
  let filterStatus = $state('');

  // Simple expression
  let totalCount = $derived(rows.length);

  // Complex expression — use $derived.by() with a function
  let filteredRows = $derived.by(() => {
    if (!filterStatus) return rows;
    return rows.filter(r => r.status === filterStatus);
  });

  // Derived from derived
  let hasResults = $derived(filteredRows.length > 0);
</script>
```

> Use `$derived(expression)` for simple one-liners. Use `$derived.by(() => { ... })` for multi-line logic.

### 2.3 `$effect` — side effects

Runs after the DOM updates. Re-runs when any reactive value it reads changes. Replaces `$: { ... }` blocks and most `onMount`/`afterUpdate` patterns.

```svelte
<script lang="ts">
  let searchQuery = $state('');
  let results = $state<string[]>([]);

  // Runs on mount AND whenever searchQuery changes
  $effect(() => {
    if (!searchQuery) {
      results = [];
      return;
    }
    // Async inside $effect: use a flag to handle cleanup
    let cancelled = false;
    fetchSearch(searchQuery).then(data => {
      if (!cancelled) results = data;
    });
    return () => { cancelled = true; };  // cleanup function
  });
</script>
```

**Key rules for `$effect`:**
- Do not use `$effect` solely to sync two reactive values — use `$derived` instead
- `$effect` cleanup: return a function from the effect; it runs before the next effect execution and on component destroy
- For one-time initialization (fetch on mount), `onMount` is still acceptable and often cleaner

### 2.4 `$props` — component props

Replaces all `export let` declarations. Props are destructured from the `$props()` call.

```svelte
<script lang="ts">
  // Basic props
  let { title, body } = $props();

  // With types
  interface Props {
    label: string;
    value: number;
    variant?: 'primary' | 'secondary';
    disabled?: boolean;
  }
  let { label, value, variant = 'primary', disabled = false }: Props = $props();

  // Spread rest props (for passing HTML attributes through)
  let { label, ...restProps } = $props();
</script>
```

### 2.5 `$bindable` — two-way binding props

Only needed when a parent wants to use `bind:propName`. Use sparingly — prefer callback props for most cases.

```svelte
<!-- FilterInput.svelte -->
<script lang="ts">
  let { value = $bindable(''), placeholder = '' } = $props();
</script>

<input bind:value {placeholder} />
```

```svelte
<!-- Parent usage -->
<FilterInput bind:value={filters.status} placeholder="Status" />
```

### 2.6 `$inspect` — debugging (dev only)

Log reactive values during development. Removed automatically in production builds.

```svelte
<script lang="ts">
  let count = $state(0);
  $inspect(count);  // logs count and its source on every change
</script>
```

---

## 3. Event handling in Svelte 5

### 3.1 DOM events

Events are now standard HTML attributes — no `on:` prefix, no modifiers.

```svelte
<!-- ✅ Svelte 5 -->
<button onclick={handleClick}>Click me</button>
<input oninput={(e) => value = e.currentTarget.value} />
<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>

<!-- ❌ Svelte 4 — do not use -->
<button on:click={handleClick}>Click me</button>
<form on:submit|preventDefault={handleSubmit}>
```

### 3.2 Component events — use callback props, not dispatch

Svelte 5 drops `createEventDispatcher`. Pass callbacks as props instead.

```svelte
<!-- ConfirmAction.svelte -->
<script lang="ts">
  interface Props {
    title: string;
    body: string;
    onConfirm: () => Promise<void>;
    onCancel?: () => void;
    danger?: boolean;
  }
  let { title, body, onConfirm, onCancel, danger = false }: Props = $props();

  let confirming = $state(false);

  async function handleConfirm() {
    confirming = true;
    try {
      await onConfirm();
    } finally {
      confirming = false;
    }
  }
</script>

<dialog open>
  <h2>{title}</h2>
  <p>{body}</p>
  <div class="actions">
    <button onclick={() => onCancel?.()}>Cancel</button>
    <button class:danger onclick={handleConfirm} disabled={confirming}>
      {confirming ? 'Working…' : 'Confirm'}
    </button>
  </div>
</dialog>
```

```svelte
<!-- Parent usage -->
<ConfirmAction
  title="Approve this request?"
  body="This will mark the approval as accepted."
  onConfirm={handleApprove}
  onCancel={() => showConfirm = false}
  danger={false}
/>
```

---

## 4. TypeScript integration

### 4.1 Script tag

Always use `lang="ts"`:

```svelte
<script lang="ts">
  import type { InboundRequestReview } from '$lib/api/types';
</script>
```

### 4.2 Typing props

```svelte
<script lang="ts">
  interface Props {
    rows: InboundRequestReview[];
    loading?: boolean;
    emptyTitle?: string;
    emptyBody?: string;
  }
  let { rows, loading = false, emptyTitle = 'No results', emptyBody = '' }: Props = $props();
</script>
```

### 4.3 Typing state

```svelte
<script lang="ts">
  import type { SessionContext, InboundRequestDetail } from '$lib/api/types';

  let session = $state<SessionContext | null>(null);
  let detail = $state<InboundRequestDetail | null>(null);
  let error = $state<string | null>(null);
</script>
```

### 4.4 Path aliases

The Vite config defines `$lib` as an alias for `web/src/lib`. Use this consistently:

```svelte
<script lang="ts">
  import { apiFetch } from '$lib/api/client';
  import type { InboundRequestReview } from '$lib/api/types';
  import { session } from '$lib/stores/session';
  import StatusPill from '$lib/components/data/StatusPill.svelte';
</script>
```

---

## 5. Application-specific patterns

### 5.1 API fetch pattern (data loading on mount)

This is the standard pattern for every route-level page component that loads data:

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { apiFetch } from '$lib/api/client';
  import { flash } from '$lib/stores/flash';
  import type { InboundRequestReview } from '$lib/api/types';

  let rows = $state<InboundRequestReview[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    try {
      rows = await apiFetch<InboundRequestReview[]>('/api/review/inbound-requests');
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load';
      flash.error(error);
    } finally {
      loading = false;
    }
  });
</script>
```

> **Why `onMount` and not `$effect`?** `onMount` runs once on mount and clearly expresses intent. `$effect` for initial data loading is also valid but prone to accidental re-runs if any reactive dependency changes. Use `onMount` for one-shot loads; use `$effect` when the fetch should re-run on filter changes.

### 5.2 Reactive filter pattern (re-fetches on filter change)

For list pages where filters trigger new API calls:

```svelte
<script lang="ts">
  import { apiFetch } from '$lib/api/client';
  import { flash } from '$lib/stores/flash';
  import type { InboundRequestReview } from '$lib/api/types';

  let filters = $state({ status: '', reference: '' });
  let rows = $state<InboundRequestReview[]>([]);
  let loading = $state(false);
  let error = $state<string | null>(null);

  // Re-runs whenever filters changes
  $effect(() => {
    // Read filters inside the effect to register the dependency
    const { status, reference } = filters;
    let cancelled = false;

    loading = true;
    error = null;

    const qs = new URLSearchParams();
    if (status) qs.set('status', status);
    if (reference) qs.set('request_reference', reference);

    apiFetch<InboundRequestReview[]>(`/api/review/inbound-requests?${qs}`)
      .then(data => { if (!cancelled) rows = data; })
      .catch(e => { if (!cancelled) { error = e.message; flash.error(e.message); } })
      .finally(() => { if (!cancelled) loading = false; });

    return () => { cancelled = true; };
  });

  function clearFilters() {
    filters.status = '';
    filters.reference = '';
  }
</script>
```

### 5.3 URL param sync for filter state

Keeps filter state in the URL so filtered views are shareable. Use the `location` store from `@svelte-spa-router`:

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { push, location } from 'svelte-spa-router';

  let filters = $state({ status: '', reference: '' });

  // Initialise from URL params on mount
  onMount(() => {
    const hash = window.location.hash; // e.g. #/review/inbound-requests?status=queued
    const qIndex = hash.indexOf('?');
    if (qIndex !== -1) {
      const qs = new URLSearchParams(hash.slice(qIndex + 1));
      filters.status = qs.get('status') ?? '';
      filters.reference = qs.get('request_reference') ?? '';
    }
  });

  // Write filters back to URL when they change
  $effect(() => {
    const { status, reference } = filters;
    const qs = new URLSearchParams();
    if (status) qs.set('status', status);
    if (reference) qs.set('request_reference', reference);
    const query = qs.toString();
    const base = '/review/inbound-requests';
    const newHash = query ? `${base}?${query}` : base;
    // Update hash without triggering router navigation
    history.replaceState(null, '', `#${newHash}`);
  });
</script>
```

### 5.4 API mutation pattern (actions that change data)

For lifecycle controls, approval decisions, form submissions:

```svelte
<script lang="ts">
  import { apiFetch } from '$lib/api/client';
  import { flash } from '$lib/stores/flash';

  let submitting = $state(false);
  let showConfirm = $state(false);

  async function handleApprove() {
    submitting = true;
    try {
      await apiFetch(`/api/approvals/${approvalId}/decision`, {
        method: 'POST',
        body: JSON.stringify({ decision: 'approved', note: decisionNote }),
      });
      flash.notice('Approval submitted');
      showConfirm = false;
      // Refresh the page data
      await loadDetail();
    } catch (e) {
      flash.error(e instanceof Error ? e.message : 'Failed to submit approval');
    } finally {
      submitting = false;
    }
  }
</script>

{#if showConfirm}
  <ConfirmAction
    title="Approve this request?"
    body="The proposal will be marked as approved and posted."
    onConfirm={handleApprove}
    onCancel={() => showConfirm = false}
  />
{/if}

<button onclick={() => showConfirm = true} disabled={submitting}>
  Approve
</button>
```

### 5.5 Skeleton loading pattern

The `DataTable` component handles this internally via its `loading` prop. For custom loading states:

```svelte
{#if loading}
  <div class="skeleton-stack">
    {#each { length: 3 } as _}
      <div class="skeleton-row"></div>
    {/each}
  </div>
{:else if error}
  <div class="inline-error">{error}</div>
{:else if rows.length === 0}
  <EmptyState title="No requests found" body="Try adjusting your filters." />
{:else}
  <DataTable {columns} {rows} />
{/if}
```

---

## 6. Standard component templates

### 6.1 Shared UI component (no API calls)

```svelte
<!-- web/src/lib/components/data/StatusPill.svelte -->
<script lang="ts">
  import { statusClass } from '$lib/utils/status';

  interface Props {
    status: string;
  }
  let { status }: Props = $props();

  let cls = $derived(statusClass(status)); // returns 'good' | 'bad' | 'warn' | 'neutral'
</script>

<span class="status-pill" class:good={cls === 'good'} class:bad={cls === 'bad'} class:warn={cls === 'warn'} class:neutral={cls === 'neutral'}>
  {status}
</span>

<style>
  .status-pill {
    display: inline-flex;
    align-items: center;
    padding: 3px 10px;
    border-radius: var(--radius-sm);
    font-size: var(--text-2xs);
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    background: var(--neutral-soft);
    color: var(--neutral);
  }
  .status-pill.good   { background: var(--good-soft);   color: var(--good); }
  .status-pill.bad    { background: var(--bad-soft);    color: var(--bad); }
  .status-pill.warn   { background: var(--warn-soft);   color: var(--warn); }
</style>
```

### 6.2 Route-level page component (list page)

```svelte
<!-- web/src/routes/review/InboundRequests.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { apiFetch } from '$lib/api/client';
  import { flash } from '$lib/stores/flash';
  import type { InboundRequestReview } from '$lib/api/types';
  import DataTable from '$lib/components/data/DataTable.svelte';
  import StatusPill from '$lib/components/data/StatusPill.svelte';
  import FilterPanel from '$lib/components/forms/FilterPanel.svelte';
  import SelectField from '$lib/components/forms/SelectField.svelte';
  import FormField from '$lib/components/forms/FormField.svelte';
  import PageHeader from '$lib/components/layout/PageHeader.svelte';

  // Route params from @svelte-spa-router (if any)
  // let { params } = $props();

  // State
  let rows = $state<InboundRequestReview[]>([]);
  let loading = $state(false);
  let filters = $state({ status: '', reference: '' });

  // Table column definitions
  const columns = [
    { key: 'request_reference', label: 'Reference' },
    { key: 'status', label: 'Status', render: (row: InboundRequestReview) => row.status },
    { key: 'channel', label: 'Channel' },
    { key: 'updated_at', label: 'Last updated' },
  ];

  // Fetch re-runs whenever filters changes (reactive filter pattern from §5.2).
  // If you need a one-shot initial load with no reactive filters, use onMount instead (see §5.1).
  $effect(() => {
    const { status, reference } = filters; // register deps
    let cancelled = false;
    loading = true;

    const qs = new URLSearchParams();
    if (status) qs.set('status', status);
    if (reference) qs.set('request_reference', reference);

    apiFetch<InboundRequestReview[]>(`/api/review/inbound-requests?${qs}`)
      .then(data => { if (!cancelled) rows = data; })
      .catch(e => { if (!cancelled) flash.error(e.message); })
      .finally(() => { if (!cancelled) loading = false; });

    return () => { cancelled = true; };
  });
</script>

<PageHeader
  eyebrow="Review"
  title="Inbound Requests"
  body="All inbound requests received by the system."
/>

<FilterPanel title="Filter requests" initialOpen={true}>
  <SelectField
    label="Status"
    bind:value={filters.status}
    options={[
      { value: '', label: 'All statuses' },
      { value: 'draft', label: 'Draft' },
      { value: 'queued', label: 'Queued' },
      { value: 'processed', label: 'Processed' },
    ]}
  />
  <FormField label="Reference" bind:value={filters.reference} placeholder="REQ-..." />
</FilterPanel>

<DataTable
  {columns}
  {rows}
  {loading}
  emptyTitle="No requests match"
  emptyBody="Try adjusting the filters above."
/>

<style>
  /* Page-specific layout only. Design tokens from :root. */
</style>
```

### 6.3 Route-level page component (detail page)

```svelte
<!-- web/src/routes/intake/RequestDetail.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { apiFetch } from '$lib/api/client';
  import { flash } from '$lib/stores/flash';
  import type { InboundRequestDetail } from '$lib/api/types';
  import PageHeader from '$lib/components/layout/PageHeader.svelte';
  import StatusPill from '$lib/components/data/StatusPill.svelte';
  import DetailGrid from '$lib/components/data/DetailGrid.svelte';
  import ConfirmAction from '$lib/components/forms/ConfirmAction.svelte';
  import BreadcrumbBar from '$lib/components/navigation/BreadcrumbBar.svelte';

  // @svelte-spa-router passes route params as a prop
  interface Props {
    params: { id: string };
  }
  let { params }: Props = $props();

  let detail = $state<InboundRequestDetail | null>(null);
  let loading = $state(true);
  let showCancelConfirm = $state(false);
  let submitting = $state(false);

  async function loadDetail() {
    try {
      detail = await apiFetch<InboundRequestDetail>(
        `/api/review/inbound-requests/${params.id}`
      );
    } catch (e) {
      flash.error('Failed to load request detail');
    } finally {
      loading = false;
    }
  }

  onMount(loadDetail);

  async function handleCancel() {
    submitting = true;
    try {
      await apiFetch(`/api/inbound-requests/${params.id}/cancel`, { method: 'POST' });
      flash.notice('Request cancelled');
      showCancelConfirm = false;
      await loadDetail();
    } catch (e) {
      flash.error(e instanceof Error ? e.message : 'Cancel failed');
    } finally {
      submitting = false;
    }
  }
</script>

<BreadcrumbBar crumbs={[
  { label: 'Inbound Requests', href: '#/review/inbound-requests' },
  { label: detail?.request_reference ?? '…' },
]} />

{#if loading}
  <div class="skeleton-block"></div>
{:else if detail}
  <PageHeader
    eyebrow="Inbound Request"
    title={detail.request_reference}
  />

  <!-- Hero card: status + key fields -->
  <section class="hero-card">
    <StatusPill status={detail.status} />
    <DetailGrid items={[
      { label: 'Channel', value: detail.channel },
      { label: 'Received', value: detail.created_at },
      { label: 'Last updated', value: detail.updated_at },
    ]} />
  </section>

  <!-- Lifecycle controls -->
  {#if detail.status === 'queued'}
    <section class="action-panel">
      <button class="secondary" onclick={() => showCancelConfirm = true}>
        Cancel request
      </button>
    </section>
  {/if}

  <!-- Collapsed sub-sections -->
  <details>
    <summary>Messages and attachments</summary>
    <!-- ... -->
  </details>
{/if}

{#if showCancelConfirm}
  <ConfirmAction
    title="Cancel this request?"
    body="The request will be marked as cancelled and removed from the processing queue."
    confirmLabel="Cancel request"
    danger={true}
    onConfirm={handleCancel}
    onCancel={() => showCancelConfirm = false}
  />
{/if}

<style>
  .hero-card {
    background: var(--surface-strong);
    border-radius: var(--radius-lg);
    padding: var(--panel-padding);
    box-shadow: var(--shadow-sm);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .action-panel {
    display: flex;
    gap: var(--space-3);
  }
</style>
```

---

## 7. `@svelte-spa-router` integration

### 7.1 Route definition in `App.svelte`

```svelte
<!-- web/src/App.svelte -->
<script lang="ts">
  import Router from 'svelte-spa-router';
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { session } from '$lib/stores/session';
  import { apiFetch } from '$lib/api/client';
  import type { SessionContext } from '$lib/api/types';
  import AppLayout from './layouts/AppLayout.svelte';
  import PublicLayout from './layouts/PublicLayout.svelte';
  import Login from './routes/auth/Login.svelte';
  import Dashboard from './routes/dashboard/Dashboard.svelte';
  import InboundRequests from './routes/review/InboundRequests.svelte';
  import RequestDetail from './routes/intake/RequestDetail.svelte';
  // ... more imports

  // Check session on startup
  onMount(async () => {
    try {
      const ctx = await apiFetch<SessionContext>('/api/session');
      session.set(ctx);
    } catch {
      push('/login');
    }
  });

  // Route map — hash-based (#/dashboard, #/review/inbound-requests, etc.)
  // All non-login routes should be wrapped with authGuard — see §7.5 for the wrap() pattern.
  // The example below shows bare routes for readability; production App.svelte applies wrap() to all authenticated routes.
  const routes = {
    '/login': Login,
    '/': Dashboard,
    '/dashboard': Dashboard,
    '/review/inbound-requests': InboundRequests,
    '/intake/requests/:id': RequestDetail,
    // ... all routes
  };
</script>

<Router {routes} />
```

### 7.2 Reading route params

`@svelte-spa-router` passes URL params as a `params` prop to the matched component:

```svelte
<!-- For route '/intake/requests/:id' -->
<script lang="ts">
  interface Props {
    params: { id: string };
  }
  let { params }: Props = $props();

  // Use params.id to fetch the entity
</script>
```

### 7.3 Programmatic navigation

```svelte
<script lang="ts">
  import { push, pop, replace } from 'svelte-spa-router';

  function goToDashboard() {
    push('/dashboard');
  }

  function goBack() {
    pop();
  }

  function replaceWithDetail(id: string) {
    replace(`/intake/requests/${id}`);
  }
</script>
```

### 7.4 Reading current location

```svelte
<script lang="ts">
  import { location } from 'svelte-spa-router';
  // $location is a readable store — auto-subscribed with $ prefix
</script>

<!-- $location contains the current hash path, e.g. '/dashboard' -->
<nav>
  <a href="#/dashboard" class:active={$location === '/dashboard'}>Home</a>
  <a href="#/review/inbound-requests" class:active={$location.startsWith('/review')}>Review</a>
</nav>
```

### 7.5 Route guards (protecting authenticated routes)

Wrap authenticated routes with a guard using `wrap()`:

```svelte
<!-- App.svelte -->
<script lang="ts">
  import { wrap } from 'svelte-spa-router/wrap';
  import { get } from 'svelte/store';
  import { session } from '$lib/stores/session';
  import { push } from 'svelte-spa-router';

  function authGuard(): boolean {
    if (!get(session)) {
      push('/login');
      return false;
    }
    return true;
  }

  const routes = {
    '/login': Login,
    '/dashboard': wrap({ component: Dashboard, conditions: [authGuard] }),
    '/review/inbound-requests': wrap({ component: InboundRequests, conditions: [authGuard] }),
    // ...
  };
</script>
```

---

## 8. Svelte stores

### 8.1 Session store

```typescript
// web/src/lib/stores/session.ts
import { writable } from 'svelte/store';
import type { SessionContext } from '$lib/api/types';

export const session = writable<SessionContext | null>(null);
```

Usage in components (auto-subscribe with `$`):

```svelte
<script lang="ts">
  import { session } from '$lib/stores/session';
</script>

{#if $session}
  <p>Logged in as {$session.user_email}</p>
{/if}
```

Usage in non-component TypeScript files (use `get()`):

```typescript
import { get } from 'svelte/store';
import { session } from '$lib/stores/session';

const currentSession = get(session);
if (!currentSession) throw new Error('Not authenticated');
```

### 8.2 Flash store (toast notifications)

```typescript
// web/src/lib/stores/flash.ts
import { writable } from 'svelte/store';

interface FlashMessage {
  id: string;
  kind: 'notice' | 'error';
  text: string;
}

function createFlashStore() {
  const { subscribe, update } = writable<FlashMessage[]>([]);

  function add(kind: FlashMessage['kind'], text: string) {
    const id = crypto.randomUUID();
    update(msgs => [...msgs, { id, kind, text }]);
    setTimeout(() => dismiss(id), 4000);  // Auto-dismiss after 4s
  }

  function dismiss(id: string) {
    update(msgs => msgs.filter(m => m.id !== id));
  }

  return {
    subscribe,
    notice: (text: string) => add('notice', text),
    error: (text: string) => add('error', text),
    dismiss,
  };
}

export const flash = createFlashStore();
```

Usage:

```svelte
<script lang="ts">
  import { flash } from '$lib/stores/flash';
  flash.notice('Draft saved');
  flash.error('Failed to load data');
</script>
```

### 8.3 Stores vs runes

Both approaches are valid in Svelte 5. This application uses **stores** for shared cross-component state (session, flash) because they work cleanly across both Svelte components and plain TypeScript modules. Runes (`$state`) are used for local component state only.

| State type | Use |
|---|---|
| Component-local state | `$state()` rune |
| Shared/global state (session, flash) | Svelte `writable` store |
| Computed from local state | `$derived()` rune |
| Computed from a store (in `.svelte` file) | `import { derived } from 'svelte/store'` then `const doubled = derived(count, $c => $c * 2)` |

---

## 9. Snippets (new Svelte 5 feature — use with care)

Snippets replace some slot use cases. They are useful but optional — the application uses slots for most layout composition. Know snippets exist but do not overuse them.

```svelte
<!-- Snippet definition (inside a component) -->
{#snippet header(title: string)}
  <div class="panel-header">
    <h2>{title}</h2>
  </div>
{/snippet}

<!-- Rendering a snippet -->
{@render header('My Panel')}
```

Snippets can be passed as props to child components:

```svelte
<!-- Parent -->
<DataTable {rows} {columns}>
  {#snippet empty()}
    <EmptyState title="No results" body="Try a different filter." />
  {/snippet}
</DataTable>

<!-- DataTable.svelte receives it — import the Snippet type for correct TypeScript -->
<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    rows: unknown[];
    columns: unknown[];
    empty?: Snippet;  // typed correctly
  }
  let { rows, columns, empty }: Props = $props();
</script>

{#if rows.length === 0}
  {@render empty?.()}
{/if}
```

> For most components in this application, use typed props over snippets. Snippets are best for cases where the parent needs to inject custom markup (like a custom empty state or custom row actions in `DataTable`).

---

## 10. Common mistakes to reject in code review

These are the most frequent errors AI tools produce when generating Svelte 5 code. Reject any PR or generated file that contains these patterns.

### 10.1 Svelte 4 prop declarations

```svelte
<!-- ❌ Reject — Svelte 4 syntax -->
<script lang="ts">
  export let title: string;
  export let count = 0;
</script>

<!-- ✅ Correct -->
<script lang="ts">
  let { title, count = 0 }: { title: string; count?: number } = $props();
</script>
```

### 10.2 Svelte 4 reactive declarations

```svelte
<!-- ❌ Reject -->
<script>
  $: doubled = count * 2;
  $: {
    if (count > 10) console.log('big');
  }
</script>

<!-- ✅ Correct -->
<script lang="ts">
  let doubled = $derived(count * 2);
  $effect(() => {
    if (count > 10) console.log('big');
  });
</script>
```

### 10.3 Svelte 4 event directives

```svelte
<!-- ❌ Reject -->
<button on:click={handleClick}>Click</button>
<form on:submit|preventDefault={handleSubmit}>

<!-- ✅ Correct -->
<button onclick={handleClick}>Click</button>
<form onsubmit={(e) => { e.preventDefault(); handleSubmit(e); }}>
```

### 10.4 `createEventDispatcher`

```svelte
<!-- ❌ Reject -->
<script>
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
  function handleClick() {
    dispatch('select', { id });
  }
</script>

<!-- ✅ Correct — use callback prop -->
<script lang="ts">
  let { onSelect }: { onSelect: (id: string) => void } = $props();
</script>
<button onclick={() => onSelect(id)}>Select</button>
```

### 10.5 Missing `lang="ts"` on script tag

```svelte
<!-- ❌ Reject — untyped -->
<script>
  let count = $state(0);
</script>

<!-- ✅ Correct -->
<script lang="ts">
  let count = $state(0);
</script>
```

### 10.6 Hardcoded CSS values instead of tokens

```svelte
<!-- ❌ Reject -->
<style>
  .card { background: #ffffff; padding: 24px; border-radius: 14px; }
</style>

<!-- ✅ Correct -->
<style>
  .card { background: var(--surface-strong); padding: var(--panel-padding); border-radius: var(--radius-lg); }
</style>
```

### 10.7 `goto()` from SvelteKit

```typescript
// ❌ Reject — SvelteKit API, not available in this project
import { goto } from '$app/navigation';
goto('/dashboard');

// ✅ Correct — @svelte-spa-router
import { push } from 'svelte-spa-router';
push('/dashboard');
```

---

## 11. Toolchain and configuration reference

### 11.1 `vite.config.ts`

```typescript
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import path from 'path';

export default defineConfig({
  plugins: [svelte()],
  resolve: {
    alias: {
      '$lib': path.resolve('./src/lib'),
    },
  },
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

### 11.2 `svelte.config.js`

```javascript
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

export default {
  preprocess: vitePreprocess(),
  compilerOptions: {
    runes: true,  // Enforce runes mode globally — prevents accidental Svelte 4 syntax
  },
};
```

> Setting `runes: true` in compiler options causes the compiler to **error** on any Svelte 4 syntax (`export let`, `$:`, `on:`). This is the recommended setting — it turns accidental legacy syntax from a runtime bug into a build error.

### 11.3 `tsconfig.json`

```json
{
  "extends": "@tsconfig/svelte/tsconfig.json",
  "compilerOptions": {
    "target": "ESNext",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "resolveJsonModule": true,
    "allowImportingTsExtensions": true,
    "checkJs": true,
    "paths": {
      "$lib": ["./src/lib"],
      "$lib/*": ["./src/lib/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.svelte"],
  "exclude": ["node_modules"]
}
```

### 11.4 `package.json` key dependencies

```json
{
  "dependencies": {
    "svelte": "^5.0.0",
    "svelte-spa-router": "^4.0.0"
  },
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^4.0.0",
    "@tsconfig/svelte": "^5.0.0",
    "typescript": "^5.0.0",
    "vite": "^6.0.0",
    "vitest": "^3.0.0",
    "@testing-library/svelte": "^5.0.0",
    "@fontsource/ibm-plex-sans": "^5.0.0",
    "@fontsource/ibm-plex-mono": "^5.0.0"
  },
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "test": "vitest"
  }
}
```

> **Version note:** Always verify package versions against [npmjs.com](https://www.npmjs.com/) before running `npm install`. The versions above reflect the known-correct range at the time of writing but npm packages increment frequently. The critical compatibility constraint is: `svelte@^5` requires `@sveltejs/vite-plugin-svelte@^4` (not `^5`). Run `npm create svelte@latest` with the versions you intend to pin and verify the generated `package.json` before committing.

---

## 12. Testing with Vitest

### 12.1 Test file placement

Co-locate test files with the component being tested:

```
web/src/lib/components/data/
├── StatusPill.svelte
├── StatusPill.test.ts       ← unit test
├── DataTable.svelte
└── DataTable.test.ts
```

### 12.2 Component test pattern

```typescript
// StatusPill.test.ts
import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import StatusPill from './StatusPill.svelte';

describe('StatusPill', () => {
  it('renders status text', () => {
    render(StatusPill, { props: { status: 'queued' } });
    expect(screen.getByText('queued')).toBeTruthy();
  });

  it('applies good class for approved status', () => {
    render(StatusPill, { props: { status: 'approved' } });
    const el = screen.getByText('approved');
    expect(el.classList.contains('good')).toBe(true);
  });
});
```

### 12.3 API client test pattern

```typescript
// client.test.ts
import { describe, it, expect, vi } from 'vitest';
import { apiFetch } from './client';

describe('apiFetch', () => {
  it('throws on 401 and redirects to login', async () => {
    global.fetch = vi.fn().mockResolvedValue({ status: 401, ok: false });
    await expect(apiFetch('/api/protected')).rejects.toThrow('Unauthorized');
  });
});
```

---

## 13. Quick reference card

Copy this into your AI tool's context window at the start of a component generation session:

```
This project uses Svelte 5 with runes mode enforced (runes: true in svelte.config.js).

ALWAYS use:
- $state() for reactive variables
- $derived() or $derived.by() for computed values  
- $effect() for side effects
- let { prop } = $props() for component props
- onclick={} not on:click={}  
- Callback props not createEventDispatcher
- import { push } from 'svelte-spa-router' for navigation (NOT goto from $app/navigation)
- lang="ts" on all <script> tags
- var(--token-name) for all CSS values (never hardcoded hex or px)
- $lib/* for import paths

NEVER use:
- export let (Svelte 4)
- $: (Svelte 4 reactive)
- on:click or on:submit (Svelte 4 event directives)
- createEventDispatcher (Svelte 4)
- goto() from SvelteKit
- Hardcoded colors: use var(--accent), var(--good), etc.
- Tailwind CSS classes
```
