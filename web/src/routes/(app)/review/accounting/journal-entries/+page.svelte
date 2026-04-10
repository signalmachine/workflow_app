<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime, formatMinorUnits } from '$lib/utils/format';
	import { accountingEntryDetail, routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review"
	title="Journal entries"
	description="Posted journal review stays on a dedicated report destination with exact accounting-entry drill-down."
/>

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewAccountingJournalEntries} class="filter-row" method="get">
			<input name="start_on" placeholder="start date" value={data.filters.startOn} />
			<input name="end_on" placeholder="end date" value={data.filters.endOn} />
			<input name="entry_id" placeholder="entry id" value={data.filters.entryID} />
			<input name="document_id" placeholder="document id" value={data.filters.documentID} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewAccountingJournalEntries}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Journal entries</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Entry</th><th>Kind</th><th>Summary</th><th>Status</th><th>Posted</th><th>Debit</th></tr></thead>
				<tbody>
					{#each data.journals as journal (journal.entry_id)}
						<tr>
							<td><a href={accountingEntryDetail(journal.entry_id)}>{journal.entry_number}</a></td>
							<td>{journal.entry_kind}</td>
							<td class="muted-copy">{journal.summary}</td>
							<td><StatusBadge status={journal.approval_status ?? 'posted'} /></td>
							<td>{formatDateTime(journal.posted_at)}</td>
							<td>{formatMinorUnits(journal.total_debit_minor)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
