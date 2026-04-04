<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { accountingEntryDetail, documentDetail, inboundRequestDetail, proposalDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review detail"
	title={data.approval.approval_id}
	description="Approval status, queue position, downstream document state, and upstream workflow continuity remain on one exact review route."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="detail-grid">
			<div><strong>Status</strong><div><StatusBadge status={data.approval.approval_status} /></div></div>
			<div><strong>Queue</strong><div>{data.approval.queue_code}</div></div>
			<div><strong>Queue status</strong><div>{data.approval.queue_status}</div></div>
			<div><strong>Requested</strong><div>{formatDateTime(data.approval.requested_at)}</div></div>
			<div><strong>Decided</strong><div>{formatDateTime(data.approval.decided_at)}</div></div>
			<div><strong>Document</strong><div>{data.approval.document_title}</div></div>
		</div>
		<div class="action-row">
			<a href={withQuery(routes.reviewApprovals, { approval_id: data.approval.approval_id })}>Filtered queue view</a>
			{#if data.approval.request_reference}
				<a href={inboundRequestDetail(data.approval.request_reference)}>Inbound request</a>
			{/if}
			{#if data.approval.recommendation_id}
				<a href={proposalDetail(data.approval.recommendation_id)}>Proposal detail</a>
			{/if}
			<a href={documentDetail(data.approval.document_id)}>Document detail</a>
			{#if data.approval.journal_entry_id}
				<a href={accountingEntryDetail(data.approval.journal_entry_id)}>Accounting detail</a>
			{/if}
		</div>
	</SurfaceCard>
</div>

