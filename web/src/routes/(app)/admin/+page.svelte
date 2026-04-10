<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import ActionCard from '$lib/components/primitives/ActionCard.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { routes } from '$lib/utils/routes';
</script>

<PageHeader
	eyebrow="Admin"
	title="Privileged maintenance hub"
	description="Privileged maintenance starts from grouped directories before operators enter concrete setup, list, and access destinations."
/>

<div class="page-stack">
	<div class="card-grid">
		<ActionCard
			title="Master data"
			summary="Choose accounting, party, or inventory setup from one grouped directory before making maintenance changes."
			href={routes.adminMasterData}
		/>
		<ActionCard
			title="Lists"
			summary="Review maintained records, exact detail continuity, and status controls from a grouped list directory."
			href={routes.adminLists}
		/>
		<ActionCard
			title="Access controls"
			summary="Provision org users and update membership roles while keeping one shared auth foundation."
			href={routes.adminAccess}
		/>
	</div>

	<SurfaceCard tone="muted">
		<p class="eyebrow">Guardrails</p>
		<ul class="principles">
			<li>Admin surfaces stay bounded to maintenance and setup, not manual workflow truth editing.</li>
			<li>All writes still flow through the shared Go services and audit boundaries.</li>
			<li>Directory pages group setup and list destinations without turning admin into a broad manual-entry workspace.</li>
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
