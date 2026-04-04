<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import MetricTile from '$lib/components/primitives/MetricTile.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { proposalDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader eyebrow="Review" title="Processed proposals" description="Coordinator recommendations, approval readiness, and downstream document continuity stay grouped here." />

<div class="page-stack">
	<div class="metric-grid">
		{#each data.summary as item (item.recommendation_status)}
			<MetricTile label={item.recommendation_status} value={item.proposal_count} href={withQuery(routes.reviewProposals, { status: item.recommendation_status })} />
		{/each}
	</div>

	<SurfaceCard>
		<form action={routes.reviewProposals} class="filter-row" method="get">
			<input name="status" placeholder="status" value={data.filters.status} />
			<input name="request_reference" placeholder="REQ-..." value={data.filters.requestReference} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewProposals}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Request</th>
						<th>Status</th>
						<th>Type</th>
						<th>Summary</th>
						<th>Downstream</th>
						<th>Created</th>
					</tr>
				</thead>
				<tbody>
					{#each data.proposals as proposal (proposal.recommendation_id)}
						<tr>
							<td><a href={proposalDetail(proposal.recommendation_id)}>{proposal.request_reference}</a></td>
							<td><StatusBadge status={proposal.recommendation_status} /></td>
							<td>{proposal.recommendation_type}</td>
							<td class="muted-copy">{proposal.summary}</td>
							<td>{proposal.document_status ?? proposal.approval_status ?? '-'}</td>
							<td>{formatDateTime(proposal.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
