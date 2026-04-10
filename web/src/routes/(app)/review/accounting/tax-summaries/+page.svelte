<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDate, formatMinorUnits } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review"
	title="Tax summaries"
	description="Tax-summary review stays on a dedicated destination for effective-range and tax-code filtering."
/>

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewAccountingTaxSummaries} class="filter-row" method="get">
			<input name="start_on" placeholder="start date" value={data.filters.startOn} />
			<input name="end_on" placeholder="end date" value={data.filters.endOn} />
			<input name="tax_type" placeholder="tax type" value={data.filters.taxType} />
			<input name="tax_code" placeholder="tax code" value={data.filters.taxCode} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewAccountingTaxSummaries}>Clear</a>
			</div>
		</form>
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
