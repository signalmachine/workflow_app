<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader eyebrow="Review" title="Audit lookup" description="Actor, entity, and timestamp provenance remain queryable without leaving the shared browser seam." />

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewAudit} class="filter-row" method="get">
			<input name="event_id" placeholder="event id" value={data.filters.eventID} />
			<input name="entity_type" placeholder="entity type" value={data.filters.entityType} />
			<input name="entity_id" placeholder="entity id" value={data.filters.entityID} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewAudit}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Event</th><th>Entity</th><th>Actor</th><th>Occurred</th></tr></thead>
				<tbody>
					{#each data.events as event (event.id)}
						<tr>
							<td>{event.event_type}</td>
							<td>{event.entity_type} · {event.entity_id}</td>
							<td>{event.actor_user_id ?? '-'}</td>
							<td>{formatDateTime(event.occurred_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
