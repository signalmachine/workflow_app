<script lang="ts">
	import type { PageProps } from './$types';
	import { goto } from '$app/navigation';

	import { submitInboundRequest } from '$lib/api/inbound';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { fileToBase64 } from '$lib/utils/files';
	import { formatDateTime } from '$lib/utils/format';
	import { routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	let submitterLabel = $state('');
	let messageText = $state('');
	let attachments = $state<FileList | null>(null);
	let saving = $state(false);
	let errorMessage = $state('');

	async function handleSubmit(): Promise<void> {
		saving = true;
		errorMessage = '';

		try {
			const payloadAttachments = await Promise.all(
				Array.from(attachments ?? []).map(async (file) => ({
					original_file_name: file.name,
					media_type: file.type || 'application/octet-stream',
					content_base64: await fileToBase64(file),
					link_role: 'evidence'
				}))
			);

			const result = await submitInboundRequest({
				origin_type: 'human',
				channel: 'agent_chat',
				metadata: { submitter_label: submitterLabel },
				message: {
					message_role: 'request',
					text_content: messageText
				},
				attachments: payloadAttachments,
				queue_for_review: true
			});

			await goto(withQuery(routes.agentChat, { request_reference: result.request_reference, request_status: result.status }));
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Failed to submit the coordinator chat request.';
		} finally {
			saving = false;
		}
	}
</script>

<PageHeader
	eyebrow="Operations"
	title="Coordinator chat"
	description="Request-centered chat stays persist-first: every message becomes a real inbound request on the shared queue."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="page-stack">
			<label>
				<span>Submitter label</span>
				<input bind:value={submitterLabel} placeholder="dispatch desk" />
			</label>
			<label>
				<span>Message</span>
				<textarea bind:value={messageText} placeholder="Need coordinator guidance on this request..."></textarea>
			</label>
			<label>
				<span>Attachments</span>
				<input bind:files={attachments} multiple type="file" />
			</label>
			<div class="filter-actions">
				<button disabled={saving} onclick={handleSubmit} type="button">{saving ? 'Submitting...' : 'Send to coordinator'}</button>
				<a href={routes.reviewInboundRequests}>Open request review</a>
			</div>
			{#if errorMessage}
				<p class="muted-copy">{errorMessage}</p>
			{/if}
			{#if data.snapshot.request_reference}
				<div class="filter-actions">
					<span class="muted-copy">Latest chat request:</span>
					<a href={withQuery(routes.reviewInboundRequests, { request_reference: data.snapshot.request_reference })}>
						{data.snapshot.request_reference}
					</a>
					{#if data.snapshot.request_status}
						<StatusBadge status={data.snapshot.request_status} />
					{/if}
				</div>
			{/if}
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Recent chat requests</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Reference</th>
						<th>Status</th>
						<th>Messages</th>
						<th>Updated</th>
					</tr>
				</thead>
				<tbody>
					{#each data.snapshot.recent_requests as request (request.request_id)}
						<tr>
							<td><a href={withQuery(routes.reviewInboundRequests, { request_reference: request.request_reference })}>{request.request_reference}</a></td>
							<td><StatusBadge status={request.status} /></td>
							<td>{request.message_count} / {request.attachment_count}</td>
							<td>{formatDateTime(request.updated_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Recent coordinator proposals</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>Request</th>
						<th>Status</th>
						<th>Summary</th>
						<th>Created</th>
					</tr>
				</thead>
				<tbody>
					{#each data.snapshot.recent_proposals as proposal (proposal.recommendation_id)}
						<tr>
							<td><a href={withQuery(routes.reviewProposals, { request_reference: proposal.request_reference })}>{proposal.request_reference}</a></td>
							<td><StatusBadge status={proposal.recommendation_status} /></td>
							<td class="muted-copy">{proposal.summary}</td>
							<td>{formatDateTime(proposal.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>
