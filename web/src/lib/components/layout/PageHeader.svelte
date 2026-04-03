<script lang="ts">
	interface Props {
		eyebrow?: string;
		title: string;
		description?: string;
	}

	let { eyebrow, title, description }: Props = $props();
</script>

<header class="page-header">
	{#if eyebrow}
		<p class="eyebrow">{eyebrow}</p>
	{/if}
	<h1>{title}</h1>
	{#if description}
		<p class="description">{description}</p>
	{/if}
</header>

<style>
	.page-header h1 {
		font-size: var(--text-xl);
		font-weight: 600;
		margin: 0.25rem 0 0;
	}

	.description {
		color: var(--ink-soft);
		margin: 0.625rem 0 0;
		max-width: 62rem;
	}
</style>
