<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime, formatMilliQuantity, formatMinorUnits } from '$lib/utils/format';
	import { accountingEntryDetail, approvalDetail, documentDetail, inboundRequestDetail, proposalDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review detail"
	title={data.workOrder.work_order_code}
	description="Exact execution review keeps work-order state, cost rollups, and upstream workflow continuity on one page."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="detail-grid">
			<div><strong>Status</strong><div><StatusBadge status={data.workOrder.status} /></div></div>
			<div><strong>Document status</strong><div>{data.workOrder.document_status}</div></div>
			<div><strong>Open tasks</strong><div>{data.workOrder.open_task_count}</div></div>
			<div><strong>Completed tasks</strong><div>{data.workOrder.completed_task_count}</div></div>
			<div><strong>Labor</strong><div>{data.workOrder.total_labor_minutes} min · {formatMinorUnits(data.workOrder.total_labor_cost_minor)}</div></div>
			<div><strong>Material</strong><div>{formatMilliQuantity(data.workOrder.material_quantity_milli)} · {formatMinorUnits(data.workOrder.posted_material_cost_minor)}</div></div>
		</div>
		<p>{data.workOrder.summary}</p>
		<div class="action-row">
			<a href={withQuery(routes.reviewWorkOrders, { work_order_id: data.workOrder.work_order_id })}>Filtered work-order view</a>
			<a href={documentDetail(data.workOrder.document_id)}>Document detail</a>
			{#if data.workOrder.request_reference}
				<a href={inboundRequestDetail(data.workOrder.request_reference)}>Inbound request</a>
			{/if}
			{#if data.workOrder.recommendation_id}
				<a href={proposalDetail(data.workOrder.recommendation_id)}>Proposal detail</a>
			{/if}
			{#if data.workOrder.approval_id}
				<a href={approvalDetail(data.workOrder.approval_id)}>Approval detail</a>
			{/if}
			{#if data.workOrder.last_accounting_posted_at}
				<a href={withQuery(routes.reviewAccountingJournalEntries, { document_id: data.workOrder.document_id })}>Accounting review</a>
			{/if}
		</div>
		<p class="muted-copy">
			Last status change {formatDateTime(data.workOrder.last_status_changed_at)} | Last posted {formatDateTime(data.workOrder.last_accounting_posted_at)}
		</p>
	</SurfaceCard>
</div>
