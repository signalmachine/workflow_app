<script lang="ts">
	import type { NavigationTab } from '$lib/utils/navigation';
	import { isNavigationItemActive } from '$lib/utils/navigation';

	interface Props {
		currentPath: string;
		tabs: NavigationTab[];
	}

	let { currentPath, tabs }: Props = $props();
</script>

{#if tabs.length > 0}
	<nav aria-label="Area sections" class="context-tabs">
		{#each tabs as tab (tab.id)}
			<a
				aria-current={isNavigationItemActive(currentPath, tab.matchers) ? 'page' : undefined}
				class:active={isNavigationItemActive(currentPath, tab.matchers)}
				href={tab.href}
			>
				{tab.label}
			</a>
		{/each}
	</nav>
{/if}

<style>
	.context-tabs {
		align-items: center;
		background: rgba(248, 251, 253, 0.86);
		border-bottom: 1px solid var(--line);
		display: flex;
		flex-wrap: wrap;
		gap: 0.55rem;
		padding: 0.85rem var(--space-8) 0.9rem;
		position: sticky;
		top: 48px;
		z-index: 18;
		backdrop-filter: blur(14px);
	}

	.context-tabs a {
		border: 1px solid rgba(47, 97, 127, 0.14);
		border-radius: 999px;
		color: var(--ink-soft);
		font-size: var(--text-xs);
		font-weight: 600;
		padding: 0.5rem 0.82rem;
		text-decoration: none;
		transition:
			background 140ms ease,
			border-color 140ms ease,
			color 140ms ease;
	}

	.context-tabs a:hover {
		background: rgba(47, 97, 127, 0.08);
		border-color: rgba(47, 97, 127, 0.2);
		color: var(--accent-strong);
		text-decoration: none;
	}

	.context-tabs a.active {
		background: var(--accent-strong);
		border-color: var(--accent-strong);
		color: white;
	}

	@media (max-width: 767px) {
		.context-tabs {
			padding: 0.75rem 1rem 0.85rem;
			top: 48px;
		}
	}
</style>
