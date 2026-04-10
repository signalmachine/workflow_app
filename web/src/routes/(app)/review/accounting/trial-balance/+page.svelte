<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDate, formatMinorUnits, humanizeStatus } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Review"
	title="Trial balance"
	description="Trial balance review stays on the shared reporting seam with explicit debit, credit, and imbalance totals."
/>

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewAccountingTrialBalance} class="filter-row" method="get">
			<input name="as_of" placeholder="as of date" value={data.filters.asOf} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewAccountingTrialBalance}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<div class="metric-grid">
		<SurfaceCard>
			<p class="eyebrow">Debit balance</p>
			<p class="metric-value">{formatMinorUnits(data.report.total_debit_balance_minor)}</p>
		</SurfaceCard>
		<SurfaceCard>
			<p class="eyebrow">Credit balance</p>
			<p class="metric-value">{formatMinorUnits(data.report.total_credit_balance_minor)}</p>
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
				<thead>
					<tr><th>Account</th><th>Class</th><th>Debits</th><th>Credits</th><th>Debit balance</th><th>Credit balance</th><th>Last effective</th></tr>
				</thead>
				<tbody>
					{#each data.report.lines as line (line.account_id)}
						<tr>
							<td>{line.account_code} · {line.account_name}</td>
							<td>{humanizeStatus(line.account_class)}</td>
							<td>{formatMinorUnits(line.total_debit_minor)}</td>
							<td>{formatMinorUnits(line.total_credit_minor)}</td>
							<td>{formatMinorUnits(line.debit_balance_minor)}</td>
							<td>{formatMinorUnits(line.credit_balance_minor)}</td>
							<td>{formatDate(line.last_effective_on)}</td>
						</tr>
					{:else}
						<tr><td colspan="7">No active ledger accounts found.</td></tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
