<script lang="ts">
	import type { PageProps } from './$types';

	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();
</script>

<PageHeader eyebrow="Review" title="Documents" description="Document truth, approval linkage, and journal continuity remain visible without leaving the Svelte workbench." />

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.reviewDocuments} class="filter-row" method="get">
			<input name="status" placeholder="status" value={data.filters.status} />
			<input name="type_code" placeholder="type code" value={data.filters.typeCode} />
			<input name="document_id" placeholder="document id" value={data.filters.documentID} />
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.reviewDocuments}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Title</th>
						<th>Status</th>
						<th>Type</th>
						<th>Approval</th>
						<th>Journal</th>
						<th>Updated</th>
					</tr>
				</thead>
				<tbody>
					{#each data.documents as document (document.document_id)}
						<tr>
							<td>{document.title}</td>
							<td><StatusBadge status={document.status} /></td>
							<td>{document.type_code}</td>
							<td>{document.approval_status ?? '-'}</td>
							<td>{document.journal_entry_number ?? '-'}</td>
							<td>{formatDateTime(document.updated_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
