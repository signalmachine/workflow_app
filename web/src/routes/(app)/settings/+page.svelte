<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import ActionCard from '$lib/components/primitives/ActionCard.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { buildSettingsViewModel } from '$lib/utils/settings';

	let { data }: PageProps = $props();

	let viewModel = $derived(buildSettingsViewModel(data.session));
</script>

<PageHeader
	eyebrow="Settings"
	title="User-scoped settings and continuity"
	description="Keep personal utility context here, and keep org-scoped maintenance under the separate admin surface."
/>

<div class="page-stack">
	<SurfaceCard>
		<p class="eyebrow">Current session</p>
		<p class="session-copy">
			{data.session.user_display_name}, {data.session.user_email} in {data.session.org_name} as {data.session.role_code}.
		</p>
	</SurfaceCard>

	<SurfaceCard tone="muted">
		<p class="eyebrow">Ownership split</p>
		<div class="principle-list">
			{#each viewModel.settingsPrinciples as principle (principle)}
				<p>{principle}</p>
			{/each}
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Current recommended routes</p>
		<div class="card-grid">
			{#each data.dashboard.primary_actions as action (action.href)}
				<ActionCard {...action} />
			{/each}
		</div>
	</SurfaceCard>

	{#if viewModel.adminContinuation}
		<SurfaceCard>
			<p class="eyebrow">Admin continuity</p>
			<div class="card-grid">
				<ActionCard {...viewModel.adminContinuation} />
			</div>
		</SurfaceCard>
	{/if}

	<SurfaceCard>
		<p class="eyebrow">Personal utility routes</p>
		<div class="card-grid">
			{#each viewModel.personalUtilityLinks as action (action.href)}
				<ActionCard {...action} />
			{/each}
		</div>
	</SurfaceCard>
</div>

<style>
	.card-grid {
		display: grid;
		gap: var(--space-4);
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
	}

	.principle-list {
		display: grid;
		gap: 0.9rem;
	}

	.principle-list p,
	.session-copy {
		color: var(--ink-soft);
		margin: 0.75rem 0 0;
	}
</style>
