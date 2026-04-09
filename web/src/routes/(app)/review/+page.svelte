<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import MetricTile from '$lib/components/primitives/MetricTile.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { approvalDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review"
	title="Review workbench"
	description="The promoted review-entry surfaces now live in Svelte with filters and continuity links kept on the shared reporting seam."
/>

<div class="page-stack">
	<div class="metric-grid">
		<MetricTile detail="Grouped lifecycle counts across the inbound request queue." label="Inbound requests" value={data.snapshot.inbound_request_count} href={routes.reviewInboundRequests} />
		<MetricTile detail="Processed recommendations ready for approval or document follow-through." label="Proposals" value={data.snapshot.proposal_count} href={routes.reviewProposals} />
		<MetricTile detail="Pending governed decisions on downstream document truth." label="Pending approvals" value={data.snapshot.pending_approvals.length} href={withQuery(routes.reviewApprovals, { status: 'pending' })} />
	</div>

	<SurfaceCard>
		<p class="eyebrow">Review surfaces</p>
		<div class="card-grid">
			<a class="action-link" href={routes.reviewInboundRequests}>Inbound requests</a>
			<a class="action-link" href={routes.reviewProposals}>Processed proposals</a>
			<a class="action-link" href={routes.reviewApprovals}>Approval queue</a>
			<a class="action-link" href={routes.reviewDocuments}>Documents</a>
			<a class="action-link" href={routes.reviewAccounting}>Accounting</a>
			<a class="action-link" href={routes.reviewInventory}>Inventory</a>
			<a class="action-link" href={routes.reviewWorkOrders}>Work orders</a>
			<a class="action-link" href={routes.reviewAudit}>Audit</a>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Pending approvals</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Approval</th>
						<th>Status</th>
						<th>Queue</th>
						<th>Document</th>
						<th>Requested</th>
					</tr>
				</thead>
				<tbody>
					{#each data.snapshot.pending_approvals as item (item.approval_id)}
						<tr>
							<td><a href={approvalDetail(item.approval_id)}>{item.approval_id}</a></td>
							<td><StatusBadge status={item.approval_status} /></td>
							<td>{item.queue_code}</td>
							<td>{item.document_title}</td>
							<td>{formatDateTime(item.requested_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>

<style>
	.action-link {
		background: var(--surface-strong);
		border: 1px solid var(--line);
		border-radius: 14px;
		color: var(--ink);
		display: block;
		font-weight: 600;
		padding: 0.95rem 1rem;
		text-decoration: none;
	}
</style>
