<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { auditEventDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	function formatJSON(value: Record<string, unknown>): string {
		return JSON.stringify(value, null, 2);
	}
</script>

<PageHeader
	eyebrow="Review detail"
	title={data.event.event_type}
	description="Audit detail remains exact and bookmarkable for actor, entity, and payload inspection."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="detail-grid">
			<div><strong>Event</strong><div>{data.event.id}</div></div>
			<div><strong>Entity type</strong><div>{data.event.entity_type}</div></div>
			<div><strong>Entity id</strong><div>{data.event.entity_id}</div></div>
			<div><strong>Actor</strong><div>{data.event.actor_user_id ?? '-'}</div></div>
			<div><strong>Occurred</strong><div>{formatDateTime(data.event.occurred_at)}</div></div>
		</div>
		<div class="action-row">
			<a href={withQuery(routes.reviewAudit, { event_id: data.event.id })}>Filtered audit view</a>
			<a href={auditEventDetail(data.event.id)}>Permalink</a>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Payload</p>
		<pre>{formatJSON(data.event.payload)}</pre>
	</SurfaceCard>
</div>
