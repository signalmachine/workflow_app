<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime, formatMilliQuantity, formatMinorUnits } from '$lib/utils/format';
	import { accountingEntryDetail, approvalDetail, documentDetail, inboundRequestDetail, proposalDetail, routes, withQuery, workOrderDetail } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review detail"
	title={`Movement ${data.movement.review.movement_number}`}
	description="Exact inventory movement continuity now includes linked document, approval, execution, and accounting reconciliation context."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="detail-grid">
			<div><strong>Type</strong><div><StatusBadge status={data.movement.review.movement_type} /></div></div>
			<div><strong>Purpose</strong><div>{data.movement.review.movement_purpose}</div></div>
			<div><strong>Item</strong><div>{data.movement.review.item_sku} · {data.movement.review.item_name}</div></div>
			<div><strong>Quantity</strong><div>{formatMilliQuantity(data.movement.review.quantity_milli)}</div></div>
			<div><strong>Created</strong><div>{formatDateTime(data.movement.review.created_at)}</div></div>
			<div><strong>Reference</strong><div>{data.movement.review.reference_note || '-'}</div></div>
		</div>
		<div class="action-row">
			<a href={withQuery(routes.reviewInventory, { movement_id: data.movement.review.movement_id })}>Filtered inventory view</a>
			{#if data.movement.review.request_reference}
				<a href={inboundRequestDetail(data.movement.review.request_reference)}>Inbound request</a>
			{/if}
			{#if data.movement.review.recommendation_id}
				<a href={proposalDetail(data.movement.review.recommendation_id)}>Proposal detail</a>
			{/if}
			{#if data.movement.review.approval_id}
				<a href={approvalDetail(data.movement.review.approval_id)}>Approval detail</a>
			{/if}
			{#if data.movement.review.document_id}
				<a href={documentDetail(data.movement.review.document_id)}>Document detail</a>
			{/if}
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Reconciliation</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Document</th><th>Execution</th><th>Accounting</th><th>Cost</th></tr></thead>
				<tbody>
					{#each data.movement.reconciliation as item (item.document_line_id)}
						<tr>
							<td>{item.document_title}</td>
							<td>
								{#if item.work_order_id}
									<a href={workOrderDetail(item.work_order_id)}>{item.work_order_code ?? item.work_order_id}</a>
								{:else}
									{item.execution_link_status ?? '-'}
								{/if}
							</td>
							<td>
								{#if item.journal_entry_id}
									<a href={accountingEntryDetail(item.journal_entry_id)}>{item.journal_entry_number ?? 'Journal detail'}</a>
								{:else}
									{item.accounting_handoff_status ?? '-'}
								{/if}
							</td>
							<td>{item.cost_minor !== undefined ? formatMinorUnits(item.cost_minor) : '-'}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>

