<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import ActionCard from '$lib/components/primitives/ActionCard.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { routes } from '$lib/utils/routes';
</script>

<PageHeader
	eyebrow="Review"
	title="Accounting reports"
	description="Accounting review starts from report directories so journal, control-balance, and tax-summary work stay on dedicated destinations."
/>

<div class="page-stack">
	<div class="card-grid">
		<ActionCard
			title="Journal entries"
			summary="Review posted journal entries, filter by document or entry id, and continue into exact posting detail."
			href={routes.reviewAccountingJournalEntries}
		/>
		<ActionCard
			title="Control balances"
			summary="Review control-account balances by date, account, or control type on the shared accounting review seam."
			href={routes.reviewAccountingControlBalances}
		/>
		<ActionCard
			title="Tax summaries"
			summary="Review tax-code summaries by effective range, tax type, or tax code without mixing them into the journal list."
			href={routes.reviewAccountingTaxSummaries}
		/>
	</div>

	<SurfaceCard tone="muted">
		<p class="eyebrow">Next reporting step</p>
		<p class="muted-copy">Trial balance, balance sheet, and income statement remain the next reporting additions behind this directory.</p>
	</SurfaceCard>
</div>

<style>
	.card-grid {
		display: grid;
		gap: var(--space-4);
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
	}
</style>
