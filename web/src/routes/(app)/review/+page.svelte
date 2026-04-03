<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import PlaceholderPanel from '$lib/components/primitives/PlaceholderPanel.svelte';
</script>

<PageHeader
	eyebrow="Review"
	title="Review workbench scaffold"
	description="Milestone 13 Slice 2 will migrate the promoted review families onto this route tree without reintroducing handler-local browser business logic."
/>

<PlaceholderPanel
	title="Review route family anchored"
	summary="The high-value review list and detail surfaces will land on this protected Svelte route family after the foundation slice is verified."
	points={[
		'Route family exists at /app/review.',
		'Protected-route bootstrap already enforces the browser session.',
		'Shared list, detail, and feedback primitives can now be layered in without restarting the scaffold.'
	]}
/>
