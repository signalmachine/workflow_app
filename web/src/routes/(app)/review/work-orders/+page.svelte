<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime, formatMinorUnits } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader eyebrow="Review" title="Work orders" description="Execution status, labor, material usage, and posting continuity remain visible on one route family." />

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewWorkOrders} class="filter-row" method="get">
			<input name="status" placeholder="status" value={data.filters.status} />
			<input name="work_order_id" placeholder="work order id" value={data.filters.workOrderID} />
			<input name="document_id" placeholder="document id" value={data.filters.documentID} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewWorkOrders}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Code</th><th>Status</th><th>Tasks</th><th>Labor</th><th>Material</th><th>Updated</th></tr></thead>
				<tbody>
					{#each data.workOrders as item (item.work_order_id)}
						<tr>
							<td>{item.work_order_code}</td>
							<td><StatusBadge status={item.status} /></td>
							<td>{item.open_task_count} open / {item.completed_task_count} done</td>
							<td>{item.total_labor_minutes} min · {formatMinorUnits(item.total_labor_cost_minor)}</td>
							<td>{formatMinorUnits(item.posted_material_cost_minor)}</td>
							<td>{formatDateTime(item.updated_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
