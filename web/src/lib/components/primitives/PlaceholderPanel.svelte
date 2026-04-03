<script lang="ts">
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';

	interface Props {
		title: string;
		summary: string;
		points?: string[];
	}

	let { title, summary, points = [] }: Props = $props();
</script>

<SurfaceCard>
	<div class="placeholder">
		<p class="eyebrow">Milestone 13 Slice 1</p>
		<h2>{title}</h2>
		<p>{summary}</p>
		{#if points.length > 0}
			<ul>
				{#each points as point (point)}
					<li>{point}</li>
				{/each}
			</ul>
		{/if}
	</div>
</SurfaceCard>

<style>
	.placeholder h2 {
		font-size: var(--text-lg);
		font-weight: 600;
		margin: 0.25rem 0 0.625rem;
	}

	.placeholder p,
	.placeholder li {
		color: var(--ink-soft);
	}

	.placeholder ul {
		margin: 0.875rem 0 0;
		padding-left: 1.125rem;
	}
</style>
