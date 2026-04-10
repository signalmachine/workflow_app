<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDate, formatMinorUnits, humanizeStatus } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	const orderedSections = ['revenue', 'expense'];
</script>

<PageHeader
	eyebrow="Review"
	title="Income statement"
	description="Income statement review keeps revenue, expenses, and net income on the backend-owned reporting seam."
/>

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewAccountingIncomeStatement} class="filter-row" method="get">
			<input name="start_on" placeholder="start date" value={data.filters.startOn} />
			<input name="end_on" placeholder="end date" value={data.filters.endOn} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewAccountingIncomeStatement}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<div class="metric-grid">
		<SurfaceCard>
			<p class="eyebrow">Revenue</p>
			<p class="metric-value">{formatMinorUnits(data.report.total_revenue_minor)}</p>
		</SurfaceCard>
		<SurfaceCard>
			<p class="eyebrow">Expenses</p>
			<p class="metric-value">{formatMinorUnits(data.report.total_expenses_minor)}</p>
		</SurfaceCard>
		<SurfaceCard>
			<p class="eyebrow">Net income</p>
			<p class="metric-value">{formatMinorUnits(data.report.net_income_minor)}</p>
		</SurfaceCard>
	</div>

	<SurfaceCard>
		<p class="eyebrow">{formatDate(data.report.start_on)} to {formatDate(data.report.end_on)}</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Section</th><th>Account</th><th>Amount</th></tr></thead>
				<tbody>
					{#each orderedSections as section (section)}
						{#each data.report.lines.filter((line) => line.section === section) as line (line.line_key)}
							<tr>
								<td>{humanizeStatus(section)}</td>
								<td>{line.account_code} · {line.account_name}</td>
								<td>{formatMinorUnits(line.amount_minor)}</td>
							</tr>
						{/each}
					{/each}
				</tbody>
				<tfoot>
					<tr><th colspan="2">Revenue</th><td>{formatMinorUnits(data.report.total_revenue_minor)}</td></tr>
					<tr><th colspan="2">Expenses</th><td>{formatMinorUnits(data.report.total_expenses_minor)}</td></tr>
					<tr><th colspan="2">Net income</th><td>{formatMinorUnits(data.report.net_income_minor)}</td></tr>
				</tfoot>
			</table>
		</div>
	</SurfaceCard>
</div>
