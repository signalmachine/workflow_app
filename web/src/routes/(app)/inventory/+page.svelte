<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import PlaceholderPanel from '$lib/components/primitives/PlaceholderPanel.svelte';
</script>

<PageHeader
	eyebrow="Inventory"
	title="Inventory landing scaffold"
	description="This route family placeholder preserves the promoted inventory landing and leaves all stock, movement, and reconciliation truth on shared backend seams."
/>

<PlaceholderPanel
	title="Inventory migration placeholder"
	summary="Later slices will read the current inventory landing and review data through explicit JSON contracts rather than old template composition."
	points={[
		'The destination exists at /app/inventory.',
		'The foundation keeps the inventory route family distinct from the broader review surface.',
		'Future parity work can add filters and drill-down routes without changing the shell contract.'
	]}
/>
