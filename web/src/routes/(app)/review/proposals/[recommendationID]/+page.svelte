<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { accountingEntryDetail, approvalDetail, documentDetail, inboundRequestDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review detail"
	title={data.proposal.recommendation_id}
	description="Processed proposal review now has an exact Svelte drill-down for recommendation status, approval readiness, and downstream document continuity."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="detail-grid">
			<div><strong>Status</strong><div><StatusBadge status={data.proposal.recommendation_status} /></div></div>
			<div><strong>Request</strong><div>{data.proposal.request_reference}</div></div>
			<div><strong>Request status</strong><div>{data.proposal.request_status}</div></div>
			<div><strong>Type</strong><div>{data.proposal.recommendation_type}</div></div>
			<div><strong>Suggested queue</strong><div>{data.proposal.suggested_queue_code ?? '-'}</div></div>
			<div><strong>Created</strong><div>{formatDateTime(data.proposal.created_at)}</div></div>
		</div>
		<p>{data.proposal.summary}</p>
		<div class="action-row">
			<a href={withQuery(routes.reviewProposals, { recommendation_id: data.proposal.recommendation_id })}>Filtered proposal view</a>
			<a href={inboundRequestDetail(data.proposal.request_reference)}>Inbound request</a>
			{#if data.proposal.approval_id}
				<a href={approvalDetail(data.proposal.approval_id)}>Approval detail</a>
			{/if}
			{#if data.proposal.document_id}
				<a href={documentDetail(data.proposal.document_id)}>Document detail</a>
				{#if data.proposal.journal_entry_id}
					<a href={accountingEntryDetail(data.proposal.journal_entry_id)}>
						Accounting entry{#if data.proposal.journal_entry_number}
							{` #${data.proposal.journal_entry_number}`}
						{/if}
					</a>
				{:else}
					<a href={withQuery(routes.reviewAccountingJournalEntries, { document_id: data.proposal.document_id })}>Accounting review</a>
				{/if}
			{/if}
		</div>
	</SurfaceCard>
</div>
