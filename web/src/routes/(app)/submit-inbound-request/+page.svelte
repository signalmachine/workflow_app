<script lang="ts">
	import { goto } from '$app/navigation';

	import { submitInboundRequest } from '$lib/api/inbound';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { fileToBase64 } from '$lib/utils/files';
	import { inboundRequestDetail, routes } from '$lib/utils/routes';
	import type { SubmitInboundRequestResponse } from '$lib/api/types';

	let submitterLabel = $state('');
	let messageText = $state('');
	let attachments = $state<FileList | null>(null);
	let saving = $state(false);
	let errorMessage = $state('');
	let result = $state<SubmitInboundRequestResponse | null>(null);

	async function buildAttachments() {
		const files = Array.from(attachments ?? []);
		return Promise.all(
			files.map(async (file) => ({
				original_file_name: file.name,
				media_type: file.type || 'application/octet-stream',
				content_base64: await fileToBase64(file),
				link_role: 'evidence'
			}))
		);
	}

	async function handleSubmit(queueForReview: boolean): Promise<void> {
		saving = true;
		errorMessage = '';

		try {
			result = await submitInboundRequest({
				origin_type: 'human',
				channel: 'browser',
				metadata: { submitter_label: submitterLabel },
				message: {
					message_role: 'request',
					text_content: messageText
				},
				attachments: await buildAttachments(),
				queue_for_review: queueForReview
			});

			if (queueForReview) {
				await goto(inboundRequestDetail(result.request_reference));
				return;
			}
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Failed to save the inbound request.';
		} finally {
			saving = false;
		}
	}
</script>

<PageHeader
	eyebrow="Intake"
	title="Submit inbound request"
	description="Create a persisted request on the shared backend, optionally keep it in draft, or queue it for coordinator review."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="page-stack">
			<label>
				<span>Submitter label</span>
				<input bind:value={submitterLabel} placeholder="front desk, dispatch, field tech" />
			</label>
			<label>
				<span>Request message</span>
				<textarea bind:value={messageText} placeholder="Describe the issue or requested work..."></textarea>
			</label>
			<label>
				<span>Attachments</span>
				<input bind:files={attachments} multiple type="file" />
			</label>
			<div class="filter-actions">
				<button disabled={saving} onclick={() => handleSubmit(true)} type="button">{saving ? 'Saving...' : 'Queue for review'}</button>
				<button class="secondary" disabled={saving} onclick={() => handleSubmit(false)} type="button">Save as draft</button>
			</div>
			{#if errorMessage}
				<p class="muted-copy">{errorMessage}</p>
			{/if}
		</div>
	</SurfaceCard>

	{#if result}
		<SurfaceCard tone="muted">
			<p class="eyebrow">Latest result</p>
			<h3>{result.request_reference}</h3>
			<p class="muted-copy">Current status: {result.status}</p>
			<a href={inboundRequestDetail(result.request_reference)}>Open exact request detail</a>
		</SurfaceCard>
	{/if}
</div>
