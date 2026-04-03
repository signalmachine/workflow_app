<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime, formatMilliQuantity } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader eyebrow="Review" title="Inventory" description="Stock, movement history, and reconciliation exceptions stay grouped under one reporting-led route family." />

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewInventory} class="filter-row" method="get">
			<input name="item_id" placeholder="item id" value={data.filters.itemID} />
			<input name="location_id" placeholder="location id" value={data.filters.locationID} />
			<input name="document_id" placeholder="document id" value={data.filters.documentID} />
			<input name="movement_type" placeholder="movement type" value={data.filters.movementType} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewInventory}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

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
							<td>{movement.movement_number}</td>
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
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
