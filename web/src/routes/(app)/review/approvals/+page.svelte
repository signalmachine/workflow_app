<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { approvalDetail, routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader eyebrow="Review" title="Approval queue" description="Explicit decision work stays ahead of downstream document and posting review." />

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewApprovals} class="filter-row" method="get">
			<input name="status" placeholder="status" value={data.filters.status} />
			<input name="queue_code" placeholder="queue code" value={data.filters.queueCode} />
			<input name="approval_id" placeholder="approval id" value={data.filters.approvalID} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewApprovals}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<SurfaceCard>
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
					{#each data.approvals as approval (approval.approval_id)}
						<tr>
							<td><a href={approvalDetail(approval.approval_id)}>{approval.approval_id}</a></td>
							<td><StatusBadge status={approval.approval_status} /></td>
							<td>{approval.queue_code}</td>
							<td>{approval.document_title}</td>
							<td>{formatDateTime(approval.requested_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
