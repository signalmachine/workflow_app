<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Route catalog"
	title="Searchable route discovery"
	description="Search by workflow term, route family, or operator intent without leaving the protected shell."
/>

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.routeCatalog} class="filter-row" method="get">
			<input name="q" placeholder="Search requests, approvals, inventory, admin..." value={data.snapshot.query} />
			<div class="filter-actions">
				<button type="submit">Search</button>
				<a href={routes.routeCatalog}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<div class="page-stack">
			{#each data.snapshot.items as item (item.href)}
				<a class="catalog-item" href={item.href}>
					<div>
						<p class="eyebrow">{item.category}</p>
						<h3>{item.title}</h3>
					</div>
					<p class="muted-copy">{item.summary}</p>
				</a>
			{/each}
		</div>
	</SurfaceCard>
</div>

<style>
	.catalog-item {
		border-top: 1px solid var(--line);
		display: block;
		padding-top: 1rem;
		text-decoration: none;
	}

	.catalog-item:first-child {
		border-top: 0;
		padding-top: 0;
	}

	.catalog-item h3 {
		color: var(--ink);
		margin: 0.2rem 0 0;
	}
</style>
