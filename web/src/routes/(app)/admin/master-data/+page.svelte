<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import ActionCard from '$lib/components/primitives/ActionCard.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { routes } from '$lib/utils/routes';
</script>

<PageHeader
	eyebrow="Admin"
	title="Master data"
	description="Setup starts from grouped destinations so operators can choose the owned foundation before making maintenance changes."
/>

<div class="page-stack">
	<div class="card-grid">
		<ActionCard
			title="Accounting master data"
			summary="Create ledger accounts, tax codes, and accounting periods on the backend-owned accounting setup seam."
			href={routes.adminAccounting}
		/>
		<ActionCard
			title="Party master data"
			summary="Create customer, vendor, and organization support records with exact contact continuity behind each party."
			href={routes.adminParties}
		/>
		<ActionCard
			title="Inventory master data"
			summary="Create item and location masters while keeping stock truth and movement review in the inventory workflow area."
			href={routes.adminInventory}
		/>
	</div>

	<SurfaceCard tone="muted">
		<p class="eyebrow">Boundary</p>
		<p class="muted-copy">Master-data setup remains separate from workflow review, posting, inventory movement, and approval decisions.</p>
	</SurfaceCard>
</div>

<style>
	.card-grid {
		display: grid;
		gap: var(--space-4);
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
	}
</style>
