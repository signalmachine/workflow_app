<script lang="ts">
	import type { PageProps } from './$types';
	import { goto } from '$app/navigation';

	import { processNextQueuedInboundRequest } from '$lib/api/inbound';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import MetricTile from '$lib/components/primitives/MetricTile.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { formatDateTime } from '$lib/utils/format';
	import { routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	let processing = $state(false);
	let processError = $state('');

	async function handleProcessNext(): Promise<void> {
		processing = true;
		processError = '';

		try {
			const result = await processNextQueuedInboundRequest();
			if (!result.processed) {
				processError = 'No queued inbound requests were available.';
				return;
			}

			await goto(withQuery(routes.reviewInboundRequests, { request_reference: result.request_reference }));
		} catch (error) {
			processError = error instanceof Error ? error.message : 'Failed to process the next queued request.';
		} finally {
			processing = false;
		}
	}
</script>

<PageHeader
	eyebrow="Operations"
	title="Operations landing"
	description="Queue movement, feed continuity, and coordinator chat now run through the Svelte route family on the shared backend seam."
/>

<div class="page-stack">
	<div class="metric-grid">
		<MetricTile detail="Requests currently waiting for coordinator pickup." label="Queued requests" value={data.snapshot.queued_request_count} href={withQuery(routes.reviewInboundRequests, { status: 'queued' })} />
		<MetricTile detail="Approval queue work that blocks downstream movement." label="Pending approvals" value={data.snapshot.pending_approval_count} href={withQuery(routes.reviewApprovals, { status: 'pending' })} />
		<MetricTile detail="Recommendations ready for operational follow-through." label="Proposal review" value={data.snapshot.proposal_review_count} href={routes.reviewProposals} />
	</div>

	<SurfaceCard>
		<p class="eyebrow">Queue actions</p>
		<div class="filter-actions">
			<button disabled={processing} onclick={handleProcessNext} type="button">
				{processing ? 'Processing...' : 'Process next queued request'}
			</button>
			<a href={routes.submitInboundRequest}>Start a new request</a>
			<a href={routes.agentChat}>Open coordinator chat</a>
			<a href={routes.operationsFeed}>Open full feed</a>
		</div>
		{#if processError}
			<p class="muted-copy">{processError}</p>
		{/if}
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Recent movement</p>
		<div class="page-stack">
			{#each data.snapshot.recent_feed as item (`${item.kind}-${item.occurred_at}-${item.title}`)}
				<section class="feed-item">
					<div class="feed-head">
						<div>
							<p class="eyebrow">{item.kind}</p>
							<h3>{item.title}</h3>
						</div>
						<StatusBadge status={item.status} />
					</div>
					<p class="muted-copy">{item.summary}</p>
					<div class="feed-actions">
						<a href={item.primary_href}>{item.primary_label}</a>
						{#if item.secondary_href && item.secondary_label}
							<a href={item.secondary_href}>{item.secondary_label}</a>
						{/if}
						<span class="muted-copy">{formatDateTime(item.occurred_at)}</span>
					</div>
				</section>
			{/each}
		</div>
	</SurfaceCard>
</div>

<style>
	.feed-item {
		border-top: 1px solid var(--line);
		padding-top: 1rem;
	}

	.feed-item:first-child {
		border-top: 0;
		padding-top: 0;
	}

	.feed-head {
		align-items: start;
		display: flex;
		gap: 1rem;
		justify-content: space-between;
	}

	.feed-head h3 {
		margin: 0.2rem 0 0;
	}

	.feed-actions {
		align-items: center;
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
	}
</style>
