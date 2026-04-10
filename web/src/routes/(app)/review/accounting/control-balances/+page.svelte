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
	title="Control balances"
	description="Control-account balance review stays on a dedicated destination for account and control-type filtering."
/>

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewAccountingControlBalances} class="filter-row" method="get">
			<input name="as_of" placeholder="as of date" value={data.filters.asOf} />
			<input name="control_type" placeholder="control type" value={data.filters.controlType} />
			<input name="account_id" placeholder="account id" value={data.filters.accountID} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewAccountingControlBalances}>Clear</a>
			</div>
		</form>
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
</div>
