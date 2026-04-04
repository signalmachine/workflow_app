<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import ActionCard from '$lib/components/primitives/ActionCard.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { routes } from '$lib/utils/routes';
</script>

<PageHeader
	eyebrow="Admin"
	title="Privileged maintenance hub"
	description="Bounded accounting, party, access, and inventory maintenance stays on the shared backend seam and now runs from the Svelte shell."
/>

<div class="page-stack">
	<div class="card-grid">
		<ActionCard
			title="Accounting setup"
			summary="Maintain ledger accounts, tax codes, and accounting periods without reopening business truth inside the browser."
			href={routes.adminAccounting}
		/>
		<ActionCard
			title="Party setup"
			summary="Create support parties and continue into exact contact maintenance on the same shared party model."
			href={routes.adminParties}
		/>
		<ActionCard
			title="Access controls"
			summary="Provision org users and update membership roles while keeping one shared auth foundation."
			href={routes.adminAccess}
		/>
		<ActionCard
			title="Inventory setup"
			summary="Maintain item and location masters with explicit status governance on the inventory foundation."
			href={routes.adminInventory}
		/>
	</div>

	<SurfaceCard tone="muted">
		<p class="eyebrow">Guardrails</p>
		<ul class="principles">
			<li>Admin surfaces stay bounded to maintenance and setup, not manual workflow truth editing.</li>
			<li>All writes still flow through the shared Go services and audit boundaries.</li>
			<li>Slice 3 admin parity is now live; review-detail and cutover work remains separate follow-through.</li>
		</ul>
	</SurfaceCard>
</div>

<style>
	.card-grid {
		display: grid;
		gap: var(--space-4);
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
	}

	.principles {
		color: var(--ink-soft);
		margin: 0.75rem 0 0;
		padding-left: 1rem;
	}
</style>
