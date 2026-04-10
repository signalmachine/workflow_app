<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDate, formatMinorUnits, humanizeStatus } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	const orderedSections = ['asset', 'liability', 'equity'];
</script>

<PageHeader
	eyebrow="Review"
	title="Balance sheet"
	description="Balance sheet review keeps assets, liabilities, equity, and current earnings on the backend-owned reporting seam."
/>

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewAccountingBalanceSheet} class="filter-row" method="get">
			<input name="as_of" placeholder="as of date" value={data.filters.asOf} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewAccountingBalanceSheet}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<div class="metric-grid">
		<SurfaceCard>
			<p class="eyebrow">Assets</p>
			<p class="metric-value">{formatMinorUnits(data.report.total_assets_minor)}</p>
		</SurfaceCard>
		<SurfaceCard>
			<p class="eyebrow">Liabilities and equity</p>
			<p class="metric-value">{formatMinorUnits(data.report.total_liabilities_and_equity_minor)}</p>
		</SurfaceCard>
		<SurfaceCard>
			<p class="eyebrow">Imbalance</p>
			<p class="metric-value">{formatMinorUnits(data.report.imbalance_minor)}</p>
		</SurfaceCard>
	</div>

	<SurfaceCard>
		<p class="eyebrow">As of {formatDate(data.report.as_of)}</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Section</th><th>Account</th><th>Amount</th></tr></thead>
				<tbody>
					{#each orderedSections as section (section)}
						{#each data.report.lines.filter((line) => line.section === section) as line (line.line_key)}
							<tr>
								<td>{humanizeStatus(section)}</td>
								<td>
									{#if line.account_code}
										{line.account_code} · {line.account_name}
									{:else}
										{line.account_name}
									{/if}
								</td>
								<td>{formatMinorUnits(line.amount_minor)}</td>
							</tr>
						{/each}
					{/each}
				</tbody>
				<tfoot>
					<tr><th colspan="2">Assets</th><td>{formatMinorUnits(data.report.total_assets_minor)}</td></tr>
					<tr><th colspan="2">Liabilities</th><td>{formatMinorUnits(data.report.total_liabilities_minor)}</td></tr>
					<tr><th colspan="2">Equity</th><td>{formatMinorUnits(data.report.total_equity_minor)}</td></tr>
				</tfoot>
			</table>
		</div>
	</SurfaceCard>
</div>
