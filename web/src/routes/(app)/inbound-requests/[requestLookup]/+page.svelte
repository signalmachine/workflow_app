<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import type { PageProps } from './$types';

	import {
		amendInboundRequest,
		cancelInboundRequest,
		deleteInboundDraft,
		queueInboundRequest,
		saveInboundDraft
	} from '$lib/api/inbound';
	import type { InboundRequestMessage, ProcessedProposalReview } from '$lib/api/types';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { fileToBase64 } from '$lib/utils/files';
	import { formatDateTime } from '$lib/utils/format';
	import {
		accountingEntryDetail,
		approvalDetail,
		documentDetail,
		inboundRequestDetail,
		proposalDetail,
		routes,
		withQuery
	} from '$lib/utils/routes';

	let { data }: PageProps = $props();

	let savingDraft = $state(false);
	let mutatingLifecycle = $state(false);
	let mutationError = $state('');
	let mutationNotice = $state('');
	let newAttachments = $state<FileList | null>(null);

	const latestProposal = $derived.by(() =>
		data.proposals.reduce<ProcessedProposalReview | null>((latest, proposal) => {
			if (latest === null) {
				return proposal;
			}
			return new Date(proposal.created_at).getTime() > new Date(latest.created_at).getTime() ? proposal : latest;
		}, null)
	);

	const latestMessage = $derived.by<InboundRequestMessage | null>(() =>
		data.messages.reduce<InboundRequestMessage | null>((latest, message) => {
			if (latest === null) {
				return message;
			}
			return message.message_index > latest.message_index ? message : latest;
		}, null)
	);

	const canEditDraft = $derived(data.request.status === 'draft');
	const canCancelQueued = $derived(data.request.status === 'queued');
	const canDeleteDraft = $derived(data.request.status === 'draft');
	const canAmendRequest = $derived(['queued', 'cancelled', 'failed'].includes(data.request.status));
	const currentSubmitterLabel = $derived(
		typeof data.request.metadata.submitter_label === 'string' ? data.request.metadata.submitter_label : ''
	);
	const currentDraftMessageText = $derived(latestMessage?.text_content ?? '');
	const currentCancelReason = $derived(data.request.cancellation_reason ?? '');

	type RequestFieldOverride = { requestID: string; value: string } | null;

	let submitterLabelOverride = $state<RequestFieldOverride>(null);
	let draftMessageTextOverride = $state<RequestFieldOverride>(null);
	let cancelReasonOverride = $state<RequestFieldOverride>(null);

	const submitterLabel = $derived(
		submitterLabelOverride?.requestID === data.request.request_id
			? submitterLabelOverride.value
			: currentSubmitterLabel
	);
	const draftMessageText = $derived(
		draftMessageTextOverride?.requestID === data.request.request_id
			? draftMessageTextOverride.value
			: currentDraftMessageText
	);
	const cancelReason = $derived(
		cancelReasonOverride?.requestID === data.request.request_id
			? cancelReasonOverride.value
			: currentCancelReason
	);

	function formatJSON(value: Record<string, unknown>): string {
		return JSON.stringify(value, null, 2);
	}

	async function buildAttachments() {
		const files = Array.from(newAttachments ?? []);
		return Promise.all(
			files.map(async (file) => ({
				original_file_name: file.name,
				media_type: file.type || 'application/octet-stream',
				content_base64: await fileToBase64(file),
				link_role: 'evidence'
			}))
		);
	}

	async function refreshDetail(notice: string): Promise<void> {
		mutationNotice = notice;
		submitterLabelOverride = null;
		draftMessageTextOverride = null;
		cancelReasonOverride = null;
		newAttachments = null;
		await invalidateAll();
	}

	async function handleSaveDraft(queueAfterSave: boolean): Promise<void> {
		if (!canEditDraft) {
			return;
		}
		savingDraft = true;
		mutationError = '';
		mutationNotice = '';
		try {
			await saveInboundDraft(data.request.request_id, {
				message_id: latestMessage?.message_id,
				origin_type: data.request.origin_type,
				channel: data.request.channel,
				metadata: {
					...data.request.metadata,
					submitter_label: submitterLabel
				},
				message: {
					message_role: latestMessage?.message_role ?? 'request',
					text_content: draftMessageText
				},
				attachments: await buildAttachments()
			});
			if (queueAfterSave) {
				await queueInboundRequest(data.request.request_id);
				await refreshDetail('Draft saved and queued for review.');
				return;
			}
			await refreshDetail('Draft changes saved.');
		} catch (error) {
			mutationError = error instanceof Error ? error.message : 'Failed to save draft changes.';
		} finally {
			savingDraft = false;
		}
	}

	async function handleCancelRequest(): Promise<void> {
		mutatingLifecycle = true;
		mutationError = '';
		mutationNotice = '';
		try {
			await cancelInboundRequest(data.request.request_id, cancelReason);
			await refreshDetail('Queued request cancelled.');
		} catch (error) {
			mutationError = error instanceof Error ? error.message : 'Failed to cancel the request.';
		} finally {
			mutatingLifecycle = false;
		}
	}

	async function handleAmendRequest(): Promise<void> {
		mutatingLifecycle = true;
		mutationError = '';
		mutationNotice = '';
		try {
			await amendInboundRequest(data.request.request_id);
			await refreshDetail('Request moved back to draft for amendment.');
		} catch (error) {
			mutationError = error instanceof Error ? error.message : 'Failed to amend the request.';
		} finally {
			mutatingLifecycle = false;
		}
	}

	async function handleDeleteDraft(): Promise<void> {
		mutatingLifecycle = true;
		mutationError = '';
		mutationNotice = '';
		try {
			await deleteInboundDraft(data.request.request_id);
			await goto(`${routes.reviewInboundRequests}?notice=${encodeURIComponent(`${data.request.request_reference} deleted.`)}`);
		} catch (error) {
			mutationError = error instanceof Error ? error.message : 'Failed to delete the draft.';
		} finally {
			mutatingLifecycle = false;
		}
	}
</script>

<PageHeader
	eyebrow="Request detail"
	title={data.request.request_reference}
	description="Exact inbound-request continuity now runs in Svelte on the shared reporting seam, including messages, attachments, AI runs, and downstream proposals."
/>

<div class="page-stack">
	<SurfaceCard>
		<div class="detail-grid">
			<div>
				<strong>Status</strong>
				<div><StatusBadge status={data.request.status} /></div>
			</div>
			<div>
				<strong>Channel</strong>
				<div>{data.request.channel}</div>
			</div>
			<div>
				<strong>Origin</strong>
				<div>{data.request.origin_type}</div>
			</div>
			<div>
				<strong>Messages</strong>
				<div>{data.request.message_count}</div>
			</div>
			<div>
				<strong>Attachments</strong>
				<div>{data.request.attachment_count}</div>
			</div>
			<div>
				<strong>Updated</strong>
				<div>{formatDateTime(data.request.updated_at)}</div>
			</div>
		</div>
		<div class="action-row">
			<a href={routes.reviewInboundRequests}>Back to inbound review</a>
			{#if data.request.last_run_id}
				<a href={inboundRequestDetail(`run:${data.request.last_run_id}`)}>Open latest run</a>
			{/if}
		</div>
	</SurfaceCard>

	{#if mutationError}
		<SurfaceCard tone="muted">
			<p>{mutationError}</p>
		</SurfaceCard>
	{:else if mutationNotice}
		<SurfaceCard tone="muted">
			<p>{mutationNotice}</p>
		</SurfaceCard>
	{/if}

	{#if canEditDraft}
		<SurfaceCard>
			<p class="eyebrow">Draft controls</p>
			<div class="page-stack">
				<label>
					<span>Submitter label</span>
					<input
						oninput={(event) => {
							submitterLabelOverride = {
								requestID: data.request.request_id,
								value: event.currentTarget.value
							};
						}}
						placeholder="front desk, dispatch, field tech"
						value={submitterLabel}
					/>
				</label>
				<label>
					<span>Request message</span>
					<textarea
						oninput={(event) => {
							draftMessageTextOverride = {
								requestID: data.request.request_id,
								value: event.currentTarget.value
							};
						}}
						placeholder="Describe the issue or requested work..."
						value={draftMessageText}
					></textarea>
				</label>
				<label>
					<span>Add attachments</span>
					<input bind:files={newAttachments} multiple type="file" />
				</label>
				<div class="filter-actions">
					<button disabled={savingDraft || mutatingLifecycle} onclick={() => handleSaveDraft(false)} type="button">
						{savingDraft ? 'Saving...' : 'Save draft changes'}
					</button>
					<button class="secondary" disabled={savingDraft || mutatingLifecycle} onclick={() => handleSaveDraft(true)} type="button">
						{savingDraft ? 'Saving...' : 'Queue updated draft'}
					</button>
					{#if canDeleteDraft}
						<button class="danger" disabled={savingDraft || mutatingLifecycle} onclick={handleDeleteDraft} type="button">
							{mutatingLifecycle ? 'Deleting...' : 'Delete draft'}
						</button>
					{/if}
				</div>
			</div>
		</SurfaceCard>
	{:else if canCancelQueued || canAmendRequest}
		<SurfaceCard>
			<p class="eyebrow">Lifecycle controls</p>
			<div class="page-stack">
				{#if canCancelQueued}
					<label>
						<span>Cancellation reason</span>
						<textarea
							oninput={(event) => {
								cancelReasonOverride = {
									requestID: data.request.request_id,
									value: event.currentTarget.value
								};
							}}
							placeholder="Explain why this queued request should stop before processing."
							value={cancelReason}
						></textarea>
					</label>
					<div class="filter-actions">
						<button class="danger" disabled={mutatingLifecycle} onclick={handleCancelRequest} type="button">
							{mutatingLifecycle ? 'Updating...' : 'Cancel queued request'}
						</button>
						<button class="secondary" disabled={mutatingLifecycle} onclick={handleAmendRequest} type="button">
							{mutatingLifecycle ? 'Updating...' : 'Amend back to draft'}
						</button>
					</div>
				{:else if canAmendRequest}
					<p class="muted-copy">
						This request can move back to draft so operators can update the persisted intake record before re-queueing it.
					</p>
					<div class="filter-actions">
						<button disabled={mutatingLifecycle} onclick={handleAmendRequest} type="button">
							{mutatingLifecycle ? 'Updating...' : 'Amend back to draft'}
						</button>
					</div>
				{/if}
			</div>
		</SurfaceCard>
	{/if}

	{#if Object.keys(data.request.metadata).length > 0}
		<SurfaceCard>
			<p class="eyebrow">Metadata</p>
			<pre>{formatJSON(data.request.metadata)}</pre>
		</SurfaceCard>
	{/if}

	<SurfaceCard>
		<p class="eyebrow">Lifecycle</p>
		<div class="detail-grid">
			<div><strong>Received</strong><div>{formatDateTime(data.request.received_at)}</div></div>
			<div><strong>Queued</strong><div>{formatDateTime(data.request.queued_at)}</div></div>
			<div><strong>Processing started</strong><div>{formatDateTime(data.request.processing_started_at)}</div></div>
			<div><strong>Processed</strong><div>{formatDateTime(data.request.processed_at)}</div></div>
			<div><strong>Completed</strong><div>{formatDateTime(data.request.completed_at)}</div></div>
			<div><strong>Cancelled</strong><div>{formatDateTime(data.request.cancelled_at)}</div></div>
			<div><strong>Failed</strong><div>{formatDateTime(data.request.failed_at)}</div></div>
			<div><strong>Cancellation reason</strong><div>{data.request.cancellation_reason ?? '-'}</div></div>
			<div><strong>Failure reason</strong><div>{data.request.failure_reason ?? '-'}</div></div>
		</div>
	</SurfaceCard>

	{#if latestProposal}
		<SurfaceCard>
			<p class="eyebrow">Workflow continuity</p>
			<div class="detail-grid continuity-summary">
				<div>
					<strong>Latest proposal</strong>
					<div>
						<a href={proposalDetail(latestProposal.recommendation_id)}>{latestProposal.recommendation_id}</a>
					</div>
				</div>
				<div>
					<strong>Proposal status</strong>
					<div><StatusBadge status={latestProposal.recommendation_status} /></div>
				</div>
				<div>
					<strong>Approval</strong>
					<div>{latestProposal.approval_status ?? '-'}</div>
				</div>
				<div>
					<strong>Document</strong>
					<div>{latestProposal.document_status ?? '-'}</div>
				</div>
			</div>
			<p class="muted-copy">
				Keep the next exact workflow surfaces close to the request evidence so request, proposal, approval, and document continuity stay on one readable path.
			</p>
			<div class="action-row">
				<a href={proposalDetail(latestProposal.recommendation_id)}>Open latest proposal</a>
				{#if latestProposal.approval_id}
					<a href={approvalDetail(latestProposal.approval_id)}>Open approval detail</a>
				{/if}
				{#if latestProposal.document_id}
					<a href={documentDetail(latestProposal.document_id)}>Open document detail</a>
					{#if latestProposal.journal_entry_id}
						<a href={accountingEntryDetail(latestProposal.journal_entry_id)}>
							Open accounting entry{#if latestProposal.journal_entry_number}
								{` #${latestProposal.journal_entry_number}`}
							{/if}
						</a>
					{:else}
						<a href={withQuery(routes.reviewAccountingJournalEntries, { document_id: latestProposal.document_id })}>Open accounting review</a>
					{/if}
				{/if}
			</div>
		</SurfaceCard>
	{/if}

	<SurfaceCard>
		<p class="eyebrow">Messages</p>
		<div class="page-stack">
			{#each data.messages as message (message.message_id)}
				<section class="timeline-item">
					<div class="timeline-head">
						<div>
							<strong>{message.message_role}</strong>
							<span class="muted-copy">#{message.message_index}</span>
						</div>
						<span class="muted-copy">{formatDateTime(message.created_at)}</span>
					</div>
					<p>{message.text_content}</p>
					<p class="muted-copy">Attachments: {message.attachment_count} | Created by: {message.created_by_user_id ?? '-'}</p>
				</section>
			{/each}
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Attachments</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead>
					<tr>
						<th>File</th>
						<th>Role</th>
						<th>Media type</th>
						<th>Derived text</th>
						<th>Created</th>
					</tr>
				</thead>
				<tbody>
					{#each data.attachments as attachment (attachment.attachment_id)}
						<tr>
							<td><a href={`/api/attachments/${attachment.attachment_id}/content`}>{attachment.original_file_name}</a></td>
							<td>{attachment.link_role}</td>
							<td>{attachment.media_type}</td>
							<td>{attachment.derived_text_count}</td>
							<td>{formatDateTime(attachment.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">AI runs</p>
		<div class="page-stack">
			{#each data.runs as run (run.run_id)}
				<section class="timeline-item" id={`run-${run.run_id}`}>
					<div class="timeline-head">
						<div>
							<a href={inboundRequestDetail(`run:${run.run_id}`)}>{run.run_id}</a>
							<span class="muted-copy">{run.agent_role} · {run.capability_code}</span>
						</div>
						<StatusBadge status={run.status} />
					</div>
					<p>{run.summary}</p>
					<p class="muted-copy">Started {formatDateTime(run.started_at)} | Completed {formatDateTime(run.completed_at)}</p>
				</section>
			{/each}
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Recommendations and proposals</p>
		<div class="page-stack">
			{#each data.recommendations as recommendation (recommendation.recommendation_id)}
				<section class="timeline-item">
					<div class="timeline-head">
						<div>
							<strong>{recommendation.recommendation_type}</strong>
							<span class="muted-copy">{recommendation.recommendation_id}</span>
						</div>
						<StatusBadge status={recommendation.status} />
					</div>
					<p>{recommendation.summary}</p>
					<p class="muted-copy">Run {recommendation.run_id} | Updated {formatDateTime(recommendation.updated_at)}</p>
				</section>
			{/each}

			{#each data.proposals as proposal (proposal.recommendation_id)}
				<section class="timeline-item">
					<div class="timeline-head">
						<div>
							<a href={proposalDetail(proposal.recommendation_id)}>Processed proposal</a>
							<span class="muted-copy">{proposal.recommendation_id}</span>
						</div>
						<StatusBadge status={proposal.recommendation_status} />
					</div>
					<p>{proposal.summary}</p>
					<p class="muted-copy">
						Approval: {proposal.approval_status ?? '-'} | Document: {proposal.document_status ?? '-'} | Created {formatDateTime(proposal.created_at)}
					</p>
					<div class="action-row action-row--tight">
						<a href={proposalDetail(proposal.recommendation_id)}>Proposal detail</a>
						{#if proposal.approval_id}
							<a href={approvalDetail(proposal.approval_id)}>Approval detail</a>
						{/if}
						{#if proposal.document_id}
							<a href={documentDetail(proposal.document_id)}>Document detail</a>
							{#if proposal.journal_entry_id}
								<a href={accountingEntryDetail(proposal.journal_entry_id)}>
									Accounting entry{#if proposal.journal_entry_number}
										{` #${proposal.journal_entry_number}`}
									{/if}
								</a>
							{:else}
								<a href={withQuery(routes.reviewAccountingJournalEntries, { document_id: proposal.document_id })}>Accounting review</a>
							{/if}
						{/if}
					</div>
				</section>
			{/each}
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Steps and delegations</p>
		<div class="page-stack">
			{#each data.steps as step (step.step_id)}
				<section class="timeline-item" id={`step-${step.step_id}`}>
					<div class="timeline-head">
						<div>
							<a href={inboundRequestDetail(`step:${step.step_id}`)}>{step.step_title}</a>
							<span class="muted-copy">{step.step_type} · #{step.step_index}</span>
						</div>
						<StatusBadge status={step.status} />
					</div>
					<p class="muted-copy">Run {step.run_id} | Created {formatDateTime(step.created_at)}</p>
				</section>
			{/each}

			{#each data.delegations as delegation (delegation.delegation_id)}
				<section class="timeline-item" id={`delegation-${delegation.delegation_id}`}>
					<div class="timeline-head">
						<div>
							<a href={inboundRequestDetail(`delegation:${delegation.delegation_id}`)}>{delegation.delegation_id}</a>
							<span class="muted-copy">{delegation.child_agent_role} · {delegation.child_capability_code}</span>
						</div>
						<StatusBadge status={delegation.child_run_status} />
					</div>
					<p>{delegation.reason}</p>
					<p class="muted-copy">Parent run {delegation.parent_run_id} | Child run {delegation.child_run_id} | Created {formatDateTime(delegation.created_at)}</p>
				</section>
			{/each}
		</div>
	</SurfaceCard>

	{#if data.artifacts.length > 0}
		<SurfaceCard>
			<p class="eyebrow">Artifacts</p>
			<div class="page-stack">
				{#each data.artifacts as artifact (artifact.artifact_id)}
					<section class="timeline-item">
						<div class="timeline-head">
							<div>
								<strong>{artifact.title}</strong>
								<span class="muted-copy">{artifact.artifact_type}</span>
							</div>
							<span class="muted-copy">{formatDateTime(artifact.created_at)}</span>
						</div>
						<p class="muted-copy">Run {artifact.run_id} | Step {artifact.step_id ?? '-'}</p>
						<pre>{formatJSON(artifact.payload)}</pre>
					</section>
				{/each}
			</div>
		</SurfaceCard>
	{/if}
</div>

<style>
	.detail-grid {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
	}

	.action-row {
		align-items: center;
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		justify-content: space-between;
		margin-top: 1rem;
	}

	.action-row--tight {
		justify-content: flex-start;
		margin-top: 0.75rem;
	}

	.continuity-summary {
		margin-bottom: 0.75rem;
	}

	.timeline-item {
		border-top: 1px solid var(--line);
		padding-top: 1rem;
	}

	.timeline-item:first-child {
		border-top: 0;
		padding-top: 0;
	}

	.timeline-head {
		align-items: start;
		display: flex;
		gap: 1rem;
		justify-content: space-between;
	}

	pre {
		background: var(--surface-muted);
		border: 1px solid var(--line);
		border-radius: 12px;
		font-size: 0.85rem;
		margin: 0;
		overflow-x: auto;
		padding: 0.9rem;
		white-space: pre-wrap;
		word-break: break-word;
	}

	button.danger {
		background: var(--bad);
	}
</style>
