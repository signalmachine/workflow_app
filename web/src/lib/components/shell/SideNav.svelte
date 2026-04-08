<script lang="ts">
	import type { NavigationArea } from '$lib/utils/navigation';
	import { isNavigationItemActive } from '$lib/utils/navigation';

	interface Props {
		areas: NavigationArea[];
		currentPath: string;
		isOpen: boolean;
		onClose: () => void;
	}

	let { areas, currentPath, isOpen, onClose }: Props = $props();
</script>

<div aria-hidden={!isOpen} class:open={isOpen} class="sidebar-backdrop" onclick={onClose}></div>
<aside class:open={isOpen} class="sidebar">
	<nav>
		<p class="nav-label">Major areas</p>
		<div class="nav-group">
			{#each areas as area (area.id)}
				<a
					aria-current={isNavigationItemActive(currentPath, area.matchers) ? 'page' : undefined}
					class:active={isNavigationItemActive(currentPath, area.matchers)}
					href={area.href}
				>
					<span class="area-label">{area.label}</span>
					<span class="area-copy">{area.description}</span>
				</a>
			{/each}
		</div>
	</nav>
</aside>

<style>
	.sidebar {
		background: var(--shell-ink);
		border-right: 1px solid rgba(255, 255, 255, 0.08);
		bottom: 0;
		color: #e8eff4;
		left: 0;
		overflow-y: auto;
		padding: 1.1rem 0.85rem 1.5rem;
		position: fixed;
		top: 48px;
		width: 240px;
	}

	.nav-label {
		color: rgba(255, 255, 255, 0.52);
		font-size: var(--text-2xs);
		font-weight: 700;
		letter-spacing: 0.12em;
		margin: 0 0 0.75rem;
		padding: 0 0.8rem;
		text-transform: uppercase;
	}

	.nav-group {
		display: grid;
		gap: 0.45rem;
	}

	.sidebar a {
		border: 1px solid rgba(255, 255, 255, 0.04);
		border-radius: 16px;
		color: inherit;
		display: block;
		padding: 0.8rem 0.9rem;
		text-decoration: none;
	}

	.sidebar a:hover {
		background: rgba(255, 255, 255, 0.045);
		border-color: rgba(255, 255, 255, 0.08);
		text-decoration: none;
	}

	.sidebar a.active {
		background:
			linear-gradient(180deg, rgba(205, 224, 236, 0.16), rgba(205, 224, 236, 0.1)),
			var(--shell-nav-active-bg);
		border-color: rgba(205, 224, 236, 0.14);
		color: #f5fbff;
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);
	}

	.area-label,
	.area-copy {
		display: block;
	}

	.area-label {
		font-size: var(--text-sm);
		font-weight: 600;
	}

	.area-copy {
		color: rgba(255, 255, 255, 0.62);
		font-size: var(--text-2xs);
		line-height: 1.45;
		margin-top: 0.24rem;
	}

	.sidebar a.active .area-copy {
		color: rgba(245, 251, 255, 0.78);
	}

	.sidebar-backdrop {
		display: none;
	}

	@media (max-width: 767px) {
		.sidebar {
			transform: translateX(-100%);
			transition: transform 160ms ease;
			z-index: 25;
		}

		.sidebar.open {
			transform: translateX(0);
		}

		.sidebar-backdrop {
			background: rgba(18, 37, 51, 0.4);
			bottom: 0;
			display: none;
			left: 0;
			position: fixed;
			right: 0;
			top: 48px;
			z-index: 24;
		}

		.sidebar-backdrop.open {
			display: block;
		}
	}
</style>
