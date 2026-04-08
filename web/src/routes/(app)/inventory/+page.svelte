<script lang="ts">
	import type { PageProps } from './$types';

	import ActionCard from '$lib/components/primitives/ActionCard.svelte';
	import MetricTile from '$lib/components/primitives/MetricTile.svelte';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime, formatMilliQuantity, formatMinorUnits, humanizeStatus } from '$lib/utils/format';
	import {
		documentDetail,
		inventoryMovementDetail,
		routes,
		withQuery,
		workOrderDetail
	} from '$lib/utils/routes';

	let { data }: PageProps = $props();

	let quickActions = $derived.by(() => {
		const actions = [
			{
				title: 'Review stock positions',
				summary: 'Open the grouped inventory review surface for on-hand balances and location-level stock truth.',
				href: routes.reviewInventory,
				id: 'review-stock'
			},
			{
				title: 'Review movement history',
				summary: 'Inspect exact movement records with shared filters for item, location, document, and movement type.',
				href: routes.reviewInventory,
				id: 'review-movements'
			},
			{
				title: 'Pending execution handoffs',
				summary: 'Focus inventory rows that still need to link into a downstream execution context.',
				href: withQuery(routes.reviewInventory, { only_pending_execution: 'true' }),
				badge: String(data.pendingExecution.length),
				id: 'pending-execution'
			},
			{
				title: 'Pending accounting handoffs',
				summary: 'Review cost-bearing inventory rows that still need a centralized accounting posting handoff.',
				href: withQuery(routes.reviewInventory, { only_pending_accounting: 'true' }),
				badge: String(data.pendingAccounting.length),
				id: 'pending-accounting'
			}
		];

		if (data.roleCode === 'admin') {
			actions.push({
				title: 'Inventory setup',
				summary: 'Open governed item and location maintenance on the shared admin seam.',
				href: routes.adminInventory,
				id: 'inventory-setup'
			});
		}

		return actions;
	});

	function describeMovementRoute(movement: (typeof data.movements)[number]): string {
		const source = movement.source_location_code ?? movement.source_location_name;
		const destination = movement.destination_location_code ?? movement.destination_location_name;
		if (source && destination) {
			return `${source} -> ${destination}`;
		}
		if (source) {
			return `${source} -> -`;
		}
		if (destination) {
			return `- -> ${destination}`;
		}
		return '-';
	}

	function describeExecutionTarget(item: (typeof data.pendingExecution)[number]): string {
		if (item.work_order_id) {
			return item.work_order_code ?? item.work_order_id;
		}
		if (item.execution_context_type && item.execution_context_id) {
			return `${humanizeStatus(item.execution_context_type)} ${item.execution_context_id}`;
		}
		return '-';
	}
</script>

<PageHeader
	eyebrow="Inventory"
	title="Inventory landing"
	description="Stock position, movement continuity, and pending execution or accounting handoffs now have one real area landing on the shared backend seam."
/>

<div class="page-stack">
	<div class="metric-grid">
		<MetricTile
			label="Visible stock positions"
			value={data.stock.length}
			detail="Current stock snapshot rows from the shared inventory review seam."
			href={routes.reviewInventory}
		/>
		<MetricTile
			label="Recent movements"
			value={data.movements.length}
			detail="Latest movement records with exact drill-down continuity into movement detail."
			href={routes.reviewInventory}
		/>
		<MetricTile
			label="Pending execution"
			value={data.pendingExecution.length}
			detail="Inventory rows still waiting to link into downstream execution context."
			href={withQuery(routes.reviewInventory, { only_pending_execution: 'true' })}
		/>
		<MetricTile
			label="Pending accounting"
			value={data.pendingAccounting.length}
			detail="Inventory rows still waiting for a centralized accounting handoff."
			href={withQuery(routes.reviewInventory, { only_pending_accounting: 'true' })}
		/>
	</div>

	<SurfaceCard>
			<p class="eyebrow">Inventory actions</p>
			<div class="card-grid">
			{#each quickActions as action (action.id)}
				<ActionCard {...action} />
			{/each}
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Recent movement snapshot</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Movement</th>
						<th>Type</th>
						<th>Item</th>
						<th>Route</th>
						<th>Qty</th>
						<th>Document</th>
						<th>Created</th>
					</tr>
				</thead>
				<tbody>
					{#each data.movements as movement (movement.movement_id)}
						<tr>
							<td><a href={inventoryMovementDetail(movement.movement_id)}>{movement.movement_number}</a></td>
							<td><StatusBadge status={movement.movement_type} /></td>
							<td>{movement.item_sku} · {movement.item_name}</td>
							<td>{describeMovementRoute(movement)}</td>
							<td>{formatMilliQuantity(movement.quantity_milli)}</td>
							<td>
								{#if movement.document_id}
									<a href={documentDetail(movement.document_id)}>{movement.document_title ?? movement.document_id}</a>
								{:else}
									-
								{/if}
							</td>
							<td>{formatDateTime(movement.created_at)}</td>
						</tr>
					{:else}
						<tr>
							<td class="muted-copy" colspan="7">No movement rows are currently available in the landing snapshot.</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Pending execution handoffs</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Movement</th>
						<th>Document</th>
						<th>Item</th>
						<th>Execution target</th>
						<th>Status</th>
					</tr>
				</thead>
				<tbody>
					{#each data.pendingExecution as item (item.document_line_id)}
						<tr>
							<td><a href={inventoryMovementDetail(item.movement_id)}>{item.movement_number}</a></td>
							<td><a href={documentDetail(item.document_id)}>{item.document_title}</a></td>
							<td>{item.item_sku} · {item.item_name}</td>
							<td>
								{#if item.work_order_id}
									<a href={workOrderDetail(item.work_order_id)}>{describeExecutionTarget(item)}</a>
								{:else}
									{describeExecutionTarget(item)}
								{/if}
							</td>
							<td>
								{#if item.execution_link_status}
									<StatusBadge status={item.execution_link_status} />
								{:else}
									-
								{/if}
							</td>
						</tr>
					{:else}
						<tr>
							<td class="muted-copy" colspan="5">No pending execution handoffs are currently visible in the landing snapshot.</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Pending accounting handoffs</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Movement</th>
						<th>Document</th>
						<th>Item</th>
						<th>Cost</th>
						<th>Status</th>
						<th>Review</th>
					</tr>
				</thead>
				<tbody>
					{#each data.pendingAccounting as item (item.document_line_id)}
						<tr>
							<td><a href={inventoryMovementDetail(item.movement_id)}>{item.movement_number}</a></td>
							<td><a href={documentDetail(item.document_id)}>{item.document_title}</a></td>
							<td>{item.item_sku} · {item.item_name}</td>
							<td>{formatMinorUnits(item.cost_minor)}</td>
							<td>
								{#if item.accounting_handoff_status}
									<StatusBadge status={item.accounting_handoff_status} />
								{:else}
									-
								{/if}
							</td>
							<td><a href={withQuery(routes.reviewAccounting, { document_id: item.document_id })}>Open accounting review</a></td>
						</tr>
					{:else}
						<tr>
							<td class="muted-copy" colspan="6">No pending accounting handoffs are currently visible in the landing snapshot.</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard tone="muted">
		<p class="eyebrow">Stock snapshot</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Item</th>
						<th>Role</th>
						<th>Location</th>
						<th>On hand</th>
					</tr>
				</thead>
				<tbody>
					{#each data.stock as item (item.item_id + item.location_id)}
						<tr>
							<td>{item.item_sku} · {item.item_name}</td>
							<td>{humanizeStatus(item.item_role)}</td>
							<td>{item.location_code} · {item.location_name}</td>
							<td>{formatMilliQuantity(item.on_hand_milli)}</td>
						</tr>
					{:else}
						<tr>
							<td class="muted-copy" colspan="4">No stock rows are currently available in the landing snapshot.</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
