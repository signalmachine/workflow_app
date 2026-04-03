<script lang="ts">
	import { page } from '$app/state';
	import { routes } from '$lib/utils/routes';
</script>

<svelte:head>
	<title>Application error</title>
</svelte:head>

<main class="error-page">
	<p class="eyebrow">Application state</p>
	<h1>{page.status}: {page.error?.message ?? 'Unexpected error'}</h1>
	<p>The Svelte shell could not resolve the requested route or load state for it.</p>
	<p><a href={routes.home}>Return to the operator shell</a></p>
</main>

<style>
	.error-page {
		margin: 0 auto;
		max-width: 48rem;
		padding: 4rem 1.5rem;
	}

	.error-page h1 {
		font-size: var(--text-xl);
		font-weight: 600;
	}
</style>
