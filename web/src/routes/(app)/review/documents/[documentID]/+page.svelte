<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { accountingEntryDetail, approvalDetail, inboundRequestDetail, proposalDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review detail"
	title={data.document.title}
	description="Document identity, approval state, and downstream posting continuity stay explicit on the shared review seam."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="detail-grid">
			<div><strong>Status</strong><div><StatusBadge status={data.document.status} /></div></div>
			<div><strong>Type</strong><div>{data.document.type_code}</div></div>
			<div><strong>Number</strong><div>{data.document.number_value ?? '-'}</div></div>
			<div><strong>Created</strong><div>{formatDateTime(data.document.created_at)}</div></div>
			<div><strong>Submitted</strong><div>{formatDateTime(data.document.submitted_at)}</div></div>
			<div><strong>Approved</strong><div>{formatDateTime(data.document.approved_at)}</div></div>
		</div>
		<div class="action-row">
			<a href={withQuery(routes.reviewDocuments, { document_id: data.document.document_id })}>Filtered document view</a>
			{#if data.document.request_reference}
				<a href={inboundRequestDetail(data.document.request_reference)}>Inbound request</a>
			{/if}
			{#if data.document.recommendation_id}
				<a href={proposalDetail(data.document.recommendation_id)}>Proposal detail</a>
			{/if}
			{#if data.document.approval_id}
				<a href={approvalDetail(data.document.approval_id)}>Approval detail</a>
			{/if}
			{#if data.document.journal_entry_id}
				<a href={accountingEntryDetail(data.document.journal_entry_id)}>Accounting detail</a>
			{/if}
		</div>
	</SurfaceCard>
</div>

