<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDate, formatDateTime, formatMinorUnits } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader eyebrow="Review" title="Accounting" description="Journal entries, control-account balances, and tax summaries stay centralized in one Svelte review surface." />

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewAccounting} class="filter-row" method="get">
			<input name="start_on" placeholder="start date" value={data.filters.startOn} />
			<input name="end_on" placeholder="end date" value={data.filters.endOn} />
			<input name="entry_id" placeholder="entry id" value={data.filters.entryID} />
			<input name="document_id" placeholder="document id" value={data.filters.documentID} />
			<input name="tax_type" placeholder="tax type" value={data.filters.taxType} />
			<input name="control_type" placeholder="control type" value={data.filters.controlType} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewAccounting}>Clear</a>
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
							<td>{journal.entry_number}</td>
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

	<SurfaceCard>
		<p class="eyebrow">Control balances</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Account</th><th>Control type</th><th>Net</th><th>Last effective</th></tr></thead>
				<tbody>
					{#each data.balances as balance (balance.account_id)}
						<tr>
							<td>{balance.account_code} · {balance.account_name}</td>
							<td>{balance.control_type}</td>
							<td>{formatMinorUnits(balance.net_minor)}</td>
							<td>{formatDate(balance.last_effective_on)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Tax summaries</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Tax</th><th>Documents</th><th>Net</th><th>Last effective</th></tr></thead>
				<tbody>
					{#each data.taxes as tax (tax.tax_type + tax.tax_code)}
						<tr>
							<td>{tax.tax_code} · {tax.tax_name}</td>
							<td>{tax.document_count}</td>
							<td>{formatMinorUnits(tax.net_minor)}</td>
							<td>{formatDate(tax.last_effective_on)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
