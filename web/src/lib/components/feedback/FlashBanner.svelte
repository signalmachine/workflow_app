<script lang="ts">
	interface Props {
		kind?: 'notice' | 'error';
		message: string;
	}

	let { kind = 'notice', message }: Props = $props();
</script>

{#if message}
	<div class:error={kind === 'error'} class:notice={kind !== 'error'} class="flash">
		{message}
	</div>
{/if}

<style>
	.flash {
		border: 1px solid var(--line);
		border-radius: 16px;
		box-shadow: var(--shadow);
		margin-bottom: var(--space-5);
		padding: 0.875rem 1rem;
	}

	.notice {
		background: var(--accent-soft);
		color: var(--accent-strong);
	}

	.error {
		background: var(--bad-soft);
		color: var(--bad);
	}
</style>
