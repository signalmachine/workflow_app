<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import ActionCard from '$lib/components/primitives/ActionCard.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { routes } from '$lib/utils/routes';
</script>

<PageHeader
	eyebrow="Admin"
	title="Lists"
	description="Maintained records are grouped here so status review starts from directory pages rather than crowded contextual tabs."
/>

<div class="page-stack">
	<div class="card-grid">
		<ActionCard
			title="Ledger, tax, and period lists"
			summary="Review accounting setup records and update active, inactive, or period-close status from the accounting destination."
			href={routes.adminAccounting}
		/>
		<ActionCard
			title="Party list"
			summary="Review support parties, status controls, and exact party detail continuity from the party destination."
			href={routes.adminParties}
		/>
		<ActionCard
			title="Item and location lists"
			summary="Review inventory setup records and active or inactive status controls from the inventory destination."
			href={routes.adminInventory}
		/>
	</div>

	<SurfaceCard tone="muted">
		<p class="eyebrow">Boundary</p>
		<p class="muted-copy">Lists expose admin-maintained records only; workflow history and posting review stay in their promoted review areas.</p>
	</SurfaceCard>
</div>

<style>
	.card-grid {
		display: grid;
		gap: var(--space-4);
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
	}
</style>
