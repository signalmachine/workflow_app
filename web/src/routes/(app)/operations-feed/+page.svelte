<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';

	let { data }: PageProps = $props();
</script>

<PageHeader
	eyebrow="Operations"
	title="Durable operations feed"
	description="Recent request, proposal, and approval movement stays visible here without dropping back to the old template layer."
/>

<SurfaceCard>
	<div class="page-stack">
		{#each data.snapshot.items as item (`${item.kind}-${item.occurred_at}-${item.title}`)}
			<section class="feed-item">
				<div class="feed-head">
					<div>
						<p class="eyebrow">{item.kind}</p>
						<h3>{item.title}</h3>
					</div>
					<StatusBadge status={item.status} />
				</div>
				<p class="muted-copy">{item.summary}</p>
				<div class="feed-actions">
					<a href={item.primary_href}>{item.primary_label}</a>
					{#if item.secondary_href && item.secondary_label}
						<a href={item.secondary_href}>{item.secondary_label}</a>
					{/if}
					<span class="muted-copy">{formatDateTime(item.occurred_at)}</span>
				</div>
			</section>
		{/each}
	</div>
</SurfaceCard>

<style>
	.feed-item {
		border-top: 1px solid var(--line);
		padding-top: 1rem;
	}

	.feed-item:first-child {
		border-top: 0;
		padding-top: 0;
	}

	.feed-head {
		align-items: start;
		display: flex;
		gap: 1rem;
		justify-content: space-between;
	}

	.feed-head h3 {
		margin: 0.2rem 0 0;
	}

	.feed-actions {
		align-items: center;
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
	}
</style>
