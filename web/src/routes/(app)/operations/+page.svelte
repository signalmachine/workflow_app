<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import PlaceholderPanel from '$lib/components/primitives/PlaceholderPanel.svelte';
</script>

<PageHeader
	eyebrow="Operations"
	title="Operations landing scaffold"
	description="This placeholder locks the route family, shell continuity, and visual composition that later workflow-surface migration will fill in."
/>

<PlaceholderPanel
	title="Operations surface pending workflow parity"
	summary="Later Milestone 13 work will move the operations landing, durable feed, and coordinator chat onto explicit JSON snapshot reads."
	points={[
		'Primary route exists at /app/operations.',
		'The shell preserves sidebar continuity across grouped destinations.',
		'No business logic moved into the browser in this foundation slice.'
	]}
/>
