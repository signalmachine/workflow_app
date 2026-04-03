<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
		tone?: 'default' | 'muted';
	}

	let { children, tone = 'default' }: Props = $props();
</script>

<section class:muted={tone === 'muted'} class="surface-card">
	{@render children()}
</section>

<style>
	.surface-card {
		background: var(--surface);
		border: 1px solid var(--line);
		border-radius: 18px;
		box-shadow: var(--shadow);
		padding: 1.25rem;
	}

	.muted {
		background: var(--surface-muted);
	}
</style>
