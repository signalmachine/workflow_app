<script lang="ts">
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import PlaceholderPanel from '$lib/components/primitives/PlaceholderPanel.svelte';
</script>

<PageHeader
	eyebrow="Admin"
	title="Privileged maintenance hub scaffold"
	description="Milestone 13 Slice 3 will migrate the bounded admin families onto this protected shell once workflow-facing surfaces are in place."
/>

<PlaceholderPanel
	title="Admin route family anchored"
	summary="The sidebar already carries the intended grouped privileged-maintenance posture. Later slices will attach the existing admin APIs and detail routes."
	points={[
		'Admin navigation is grouped under one sidebar branch.',
		'Bounded shared backend seams remain the source of truth.',
		'No second admin-only shell or route strip was introduced.'
	]}
/>
