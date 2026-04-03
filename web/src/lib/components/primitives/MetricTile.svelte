<script lang="ts">
	interface Props {
		label: string;
		value: string | number;
		detail?: string;
		href?: string;
	}

	let { label, value, detail, href }: Props = $props();
</script>

<section class="metric-tile">
	<p class="eyebrow">{label}</p>
	<div class="value-row">
		<strong>{value}</strong>
		{#if href}
			<a href={href}>Open</a>
		{/if}
	</div>
	{#if detail}
		<p class="detail">{detail}</p>
	{/if}
</section>

<style>
	.metric-tile {
		background: var(--surface-strong);
		border: 1px solid var(--line);
		border-radius: 16px;
		padding: 1rem;
	}

	.value-row {
		align-items: baseline;
		display: flex;
		gap: 0.75rem;
		justify-content: space-between;
		margin-top: 0.4rem;
	}

	.value-row strong {
		font-size: 1.6rem;
		font-weight: 600;
	}

	.detail {
		color: var(--ink-soft);
		margin: 0.5rem 0 0;
	}
</style>
