<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime, formatMilliQuantity } from '$lib/utils/format';
	import { inventoryMovementDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	let activeScopes = $derived.by(() => {
		const scopes: { id: string; label: string; href?: string }[] = [];
		if (data.filters.onlyPendingExecution) {
			scopes.push({
				id: 'pending-execution',
				label: 'Pending execution handoffs',
				href: withQuery(routes.reviewInventory, { only_pending_execution: 'true' })
			});
		}
		if (data.filters.onlyPendingAccounting) {
			scopes.push({
				id: 'pending-accounting',
				label: 'Pending accounting handoffs',
				href: withQuery(routes.reviewInventory, { only_pending_accounting: 'true' })
			});
		}
		return scopes;
	});
</script>

<PageHeader eyebrow="Review" title="Inventory" description="Stock, movement history, and reconciliation exceptions stay grouped under one reporting-led route family." />

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewInventory} class="filter-row" method="get">
			<input name="item_id" placeholder="item id" value={data.filters.itemID} />
			<input name="location_id" placeholder="location id" value={data.filters.locationID} />
			<input name="document_id" placeholder="document id" value={data.filters.documentID} />
			<input name="movement_type" placeholder="movement type" value={data.filters.movementType} />
			<label class="filter-toggle">
				<input
					checked={data.filters.onlyPendingExecution}
					name="only_pending_execution"
					type="checkbox"
					value="true"
				/>
				Pending execution only
			</label>
			<label class="filter-toggle">
				<input
					checked={data.filters.onlyPendingAccounting}
					name="only_pending_accounting"
					type="checkbox"
					value="true"
				/>
				Pending accounting only
			</label>
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewInventory}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	{#if activeScopes.length > 0}
		<SurfaceCard tone="muted">
			<p class="eyebrow">Active reconciliation scope</p>
			<div class="scope-list">
				{#each activeScopes as scope (scope.id)}
					<a class="scope-chip" href={scope.href}>{scope.label}</a>
				{/each}
			</div>
			<p class="muted-copy">
				These scoped links come directly from the inventory landing so pending downstream handoffs stay easy to review without rebuilding the query by hand.
			</p>
		</SurfaceCard>
	{/if}

	<SurfaceCard>
		<p class="eyebrow">Stock</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Item</th><th>Location</th><th>On hand</th></tr></thead>
				<tbody>
					{#each data.stock as item (item.item_id + item.location_id)}
						<tr>
							<td>{item.item_sku} · {item.item_name}</td>
							<td>{item.location_code} · {item.location_name}</td>
							<td>{formatMilliQuantity(item.on_hand_milli)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Movements</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Movement</th><th>Status</th><th>Item</th><th>Qty</th><th>Created</th></tr></thead>
				<tbody>
					{#each data.movements as movement (movement.movement_id)}
						<tr>
							<td><a href={inventoryMovementDetail(movement.movement_id)}>{movement.movement_number}</a></td>
							<td><StatusBadge status={movement.movement_type} /></td>
							<td>{movement.item_sku}</td>
							<td>{formatMilliQuantity(movement.quantity_milli)}</td>
							<td>{formatDateTime(movement.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Reconciliation</p>
		{#if activeScopes.length > 0}
			<p class="muted-copy">
				Showing {activeScopes.map((scope) => scope.label.toLowerCase()).join(' and ')}.
			</p>
		{/if}
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Document</th><th>Item</th><th>Execution</th><th>Accounting</th><th>Created</th></tr></thead>
				<tbody>
					{#each data.reconciliation as item (item.document_line_id)}
						<tr>
							<td>{item.document_title}</td>
							<td>{item.item_sku}</td>
							<td>
								{#if item.execution_link_status}
									<StatusBadge status={item.execution_link_status} />
								{:else}
									-
								{/if}
							</td>
							<td>
								{#if item.accounting_handoff_status}
									<StatusBadge status={item.accounting_handoff_status} />
								{:else}
									-
								{/if}
							</td>
							<td>{formatDateTime(item.movement_created_at)}</td>
						</tr>
					{:else}
						<tr>
							<td class="muted-copy" colspan="5">
								No reconciliation rows match the current inventory review scope.
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>

<style>
	.filter-toggle {
		align-items: center;
		color: var(--ink-soft);
		display: inline-flex;
		gap: 0.55rem;
		min-height: 2.75rem;
	}

	.scope-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		margin-top: 0.9rem;
	}

	.scope-chip {
		background: var(--surface-strong);
		border: 1px solid var(--line);
		border-radius: 999px;
		color: var(--ink);
		display: inline-flex;
		font-size: 0.92rem;
		font-weight: 600;
		padding: 0.6rem 0.9rem;
		text-decoration: none;
	}
</style>
