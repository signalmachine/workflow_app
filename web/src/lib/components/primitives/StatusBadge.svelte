<script lang="ts">
	import { humanizeStatus } from '$lib/utils/format';

	interface Props {
		status: string;
	}

	let { status }: Props = $props();

	let tone = $derived.by(() => {
		switch (status.toLowerCase()) {
			case 'completed':
			case 'processed':
			case 'approved':
			case 'active':
			case 'posted':
				return 'good';
			case 'failed':
			case 'rejected':
			case 'cancelled':
			case 'inactive':
				return 'bad';
			case 'pending':
			case 'queued':
			case 'draft':
			case 'processing':
			case 'approval_requested':
				return 'warn';
			default:
				return 'neutral';
		}
	});
</script>

<span class={`badge ${tone}`}>{humanizeStatus(status)}</span>

<style>
	.badge {
		border-radius: 999px;
		display: inline-flex;
		font-size: var(--text-2xs);
		font-weight: 700;
		letter-spacing: 0.04em;
		padding: 0.22rem 0.6rem;
		text-transform: uppercase;
	}

	.good {
		background: var(--good-soft);
		color: var(--good);
	}

	.bad {
		background: var(--bad-soft);
		color: var(--bad);
	}

	.warn {
		background: var(--warn-soft);
		color: var(--warn);
	}

	.neutral {
		background: var(--neutral-soft);
		color: var(--neutral);
	}
</style>
