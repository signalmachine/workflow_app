<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import ActionCard from '$lib/components/primitives/ActionCard.svelte';
	import MetricTile from '$lib/components/primitives/MetricTile.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { inboundRequestDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	let requestCount = $derived(data.dashboard.inbound_summary.reduce((total, item) => total + item.request_count, 0));
	let proposalCount = $derived(data.dashboard.proposal_summary.reduce((total, item) => total + item.proposal_count, 0));
</script>

<PageHeader
	eyebrow="Operator home"
	title={data.dashboard.role_headline}
	description={data.dashboard.role_body}
/>

<div class="page-stack">
	<div class="metric-grid">
		<MetricTile detail="Current lifecycle totals across the shared request queue." label="Requests" value={requestCount} href={routes.reviewInboundRequests} />
		<MetricTile detail="Coordinator recommendations ready for follow-through." label="Proposals" value={proposalCount} href={routes.reviewProposals} />
		<MetricTile detail="Approval work should stay close to the first click." label="Pending approvals" value={data.dashboard.approvals.length} href={withQuery(routes.reviewApprovals, { status: 'pending' })} />
	</div>

	<SurfaceCard>
		<p class="eyebrow">Primary actions</p>
		<div class="card-grid">
			{#each data.dashboard.primary_actions as action (action.href)}
				<ActionCard {...action} />
			{/each}
		</div>
	</SurfaceCard>

	<SurfaceCard tone="muted">
		<p class="eyebrow">Secondary actions</p>
		<div class="card-grid">
			{#each data.dashboard.secondary_actions as action (action.href)}
				<ActionCard {...action} />
			{/each}
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Recent requests</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Reference</th>
						<th>Status</th>
						<th>Channel</th>
						<th>Messages</th>
						<th>Updated</th>
					</tr>
				</thead>
				<tbody>
					{#each data.dashboard.inbound_requests as request (request.request_id)}
						<tr>
							<td><a href={inboundRequestDetail(request.request_reference)}>{request.request_reference}</a></td>
							<td><StatusBadge status={request.status} /></td>
							<td>{request.channel}</td>
							<td>{request.message_count} / {request.attachment_count}</td>
							<td>{formatDateTime(request.updated_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Recent proposals</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Request</th>
						<th>Status</th>
						<th>Type</th>
						<th>Summary</th>
						<th>Created</th>
					</tr>
				</thead>
				<tbody>
					{#each data.dashboard.proposals as proposal (proposal.recommendation_id)}
						<tr>
							<td><a href={withQuery(routes.reviewProposals, { request_reference: proposal.request_reference })}>{proposal.request_reference}</a></td>
							<td><StatusBadge status={proposal.recommendation_status} /></td>
							<td>{proposal.recommendation_type}</td>
							<td class="muted-copy">{proposal.summary}</td>
							<td>{formatDateTime(proposal.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
