<script lang="ts">
	interface Props {
		title: string;
		summary: string;
		href: string;
		badge?: string;
	}

	let { title, summary, href, badge }: Props = $props();
</script>

<a class="action-card" href={href}>
	<div class="title-row">
		<h3>{title}</h3>
		{#if badge}
			<span>{badge}</span>
		{/if}
	</div>
	<p>{summary}</p>
</a>

<style>
	.action-card {
		background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(239, 244, 248, 0.9));
		border: 1px solid var(--line);
		border-radius: 16px;
		display: block;
		padding: 1rem;
		text-decoration: none;
	}

	.action-card:hover {
		border-color: var(--line-strong);
		text-decoration: none;
	}

	.title-row {
		align-items: center;
		display: flex;
		gap: 0.75rem;
		justify-content: space-between;
	}

	.title-row h3 {
		color: var(--ink);
		font-size: var(--text-lg);
		margin: 0;
	}

	.title-row span {
		background: var(--accent-soft);
		border-radius: 999px;
		color: var(--accent-strong);
		font-size: var(--text-xs);
		font-weight: 700;
		padding: 0.2rem 0.55rem;
	}

	p {
		color: var(--ink-soft);
		margin: 0.55rem 0 0;
	}
</style>
