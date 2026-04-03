<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import PlaceholderPanel from '$lib/components/primitives/PlaceholderPanel.svelte';
</script>

<PageHeader
	eyebrow="Operator home"
	title="Role-aware home shell"
	description="This Slice 1 route establishes the Svelte home entry point, shell state, and shared session bootstrap on the same Go auth origin."
/>

<div class="stack">
	<PlaceholderPanel
		title="Home migration placeholder"
		summary="Milestone 13 Slice 2 will replace this foundation placeholder with dashboard and workload parity built from shared reporting snapshots instead of Go template rendering."
		points={[
			'Session bootstrap already resolves through GET /api/session.',
			'The shell uses the approved left-sidebar and fixed top-bar structure.',
			'Route groups are in place for protected app routes and public auth routes.'
		]}
	/>
</div>

<style>
	.stack {
		display: grid;
		gap: 1rem;
	}
</style>
