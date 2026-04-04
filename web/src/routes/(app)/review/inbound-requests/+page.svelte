<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import MetricTile from '$lib/components/primitives/MetricTile.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { inboundRequestDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader eyebrow="Review" title="Inbound requests" description="Review draft, queued, processing, failed, and completed request states through the shared reporting seam." />

<div class="page-stack">
	<div class="metric-grid">
		{#each data.summary as item (item.status)}
			<MetricTile label={item.status} value={item.request_count} href={withQuery(routes.reviewInboundRequests, { status: item.status })} />
		{/each}
	</div>

	<SurfaceCard>
		<form action={routes.reviewInboundRequests} class="filter-row" method="get">
			<input name="status" placeholder="status" value={data.filters.status} />
			<input name="request_reference" placeholder="REQ-..." value={data.filters.requestReference} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewInboundRequests}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Reference</th>
						<th>Status</th>
						<th>Channel</th>
						<th>Messages</th>
						<th>Latest recommendation</th>
						<th>Updated</th>
					</tr>
				</thead>
				<tbody>
					{#each data.requests as request (request.request_id)}
						<tr>
							<td><a href={inboundRequestDetail(request.request_reference)}>{request.request_reference}</a></td>
							<td><StatusBadge status={request.status} /></td>
							<td>{request.channel}</td>
							<td>{request.message_count} / {request.attachment_count}</td>
							<td>{request.last_recommendation_status ?? '-'}</td>
							<td>{formatDateTime(request.updated_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
