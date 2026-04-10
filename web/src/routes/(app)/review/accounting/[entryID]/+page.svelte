<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDate, formatDateTime, formatMinorUnits } from '$lib/utils/format';
	import { approvalDetail, documentDetail, inboundRequestDetail, proposalDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review detail"
	title={`Journal ${data.journal.entry_number}`}
	description="Centralized posting detail remains exact and traceable back through document, approval, and request continuity where those links exist."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="detail-grid">
			<div><strong>Status</strong><div><StatusBadge status={data.journal.approval_status ?? 'posted'} /></div></div>
			<div><strong>Kind</strong><div>{data.journal.entry_kind}</div></div>
			<div><strong>Effective</strong><div>{formatDate(data.journal.effective_on)}</div></div>
			<div><strong>Posted</strong><div>{formatDateTime(data.journal.posted_at)}</div></div>
			<div><strong>Debit</strong><div>{formatMinorUnits(data.journal.total_debit_minor)}</div></div>
			<div><strong>Credit</strong><div>{formatMinorUnits(data.journal.total_credit_minor)}</div></div>
		</div>
		<p>{data.journal.summary}</p>
		<div class="action-row">
			<a href={withQuery(routes.reviewAccountingJournalEntries, { entry_id: data.journal.entry_id })}>Filtered accounting view</a>
			{#if data.journal.source_document_id}
				<a href={documentDetail(data.journal.source_document_id)}>Document detail</a>
			{/if}
			{#if data.journal.approval_id}
				<a href={approvalDetail(data.journal.approval_id)}>Approval detail</a>
			{/if}
			{#if data.journal.request_reference}
				<a href={inboundRequestDetail(data.journal.request_reference)}>Inbound request</a>
			{/if}
			{#if data.journal.recommendation_id}
				<a href={proposalDetail(data.journal.recommendation_id)}>Proposal detail</a>
			{/if}
		</div>
	</SurfaceCard>
</div>
