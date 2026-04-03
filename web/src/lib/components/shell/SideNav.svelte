<script lang="ts">
	import { routes } from '$lib/utils/routes';

	interface NavItem {
		label: string;
		href: string;
		match: (pathname: string) => boolean;
	}

	interface Props {
		currentPath: string;
		isOpen: boolean;
		onClose: () => void;
	}

	let { currentPath, isOpen, onClose }: Props = $props();

	const primaryItems: NavItem[] = [
		{ label: 'Home', href: routes.home, match: (pathname) => pathname === routes.home },
		{ label: 'Operations', href: routes.operations, match: (pathname) => pathname.startsWith(routes.operations) },
		{ label: 'Review', href: routes.review, match: (pathname) => pathname.startsWith(routes.review) },
		{ label: 'Inventory', href: routes.inventory, match: (pathname) => pathname.startsWith(routes.inventory) }
	];

	const utilityItems: NavItem[] = [
		{ label: 'Routes', href: routes.routeCatalog, match: (pathname) => pathname.startsWith(routes.routeCatalog) },
		{ label: 'Settings', href: routes.settings, match: (pathname) => pathname.startsWith(routes.settings) },
		{ label: 'Admin', href: routes.admin, match: (pathname) => pathname.startsWith(routes.admin) }
	];

	function active(item: NavItem): boolean {
		return item.match(currentPath);
	}
</script>

<div aria-hidden={!isOpen} class:open={isOpen} class="sidebar-backdrop" onclick={onClose}></div>
<aside class:open={isOpen} class="sidebar">
	<nav>
		<div class="nav-group">
			{#each primaryItems as item (item.href)}
				<a aria-current={active(item) ? 'page' : undefined} class:active={active(item)} href={item.href}>
					{item.label}
				</a>
			{/each}
			<div class="nav-subgroup">
				<a href={routes.submitInboundRequest}>Submit request</a>
				<a href={routes.operationsFeed}>Operations feed</a>
				<a href={routes.agentChat}>Agent chat</a>
			</div>
		</div>
		<div class="nav-divider"></div>
		<div class="nav-group">
			{#each utilityItems as item (item.href)}
				<a aria-current={active(item) ? 'page' : undefined} class:active={active(item)} href={item.href}>
					{item.label}
				</a>
			{/each}
			<div class="nav-subgroup">
				<a href={routes.adminAccounting}>Accounting setup</a>
				<a href={routes.adminParties}>Party setup</a>
				<a href={routes.adminAccess}>Access controls</a>
				<a href={routes.adminInventory}>Inventory setup</a>
			</div>
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
		padding: 1rem 0.85rem 1.5rem;
		position: fixed;
		top: 48px;
		width: 220px;
	}

	.nav-group {
		display: grid;
		gap: 0.25rem;
	}

	.sidebar a {
		border-radius: 12px;
		color: inherit;
		display: block;
		font-size: var(--text-sm);
		padding: 0.7rem 0.85rem;
		text-decoration: none;
	}

	.sidebar a:hover {
		background: var(--shell-nav-hover);
		text-decoration: none;
	}

	.sidebar a.active {
		background: var(--shell-nav-active-bg);
		color: var(--shell-nav-active-text);
	}

	.nav-subgroup {
		display: grid;
		gap: 0.125rem;
		margin-left: 0.5rem;
		margin-top: 0.35rem;
	}

	.nav-subgroup a {
		color: rgba(255, 255, 255, 0.74);
		font-size: var(--text-xs);
		padding: 0.5rem 0.75rem;
	}

	.nav-divider {
		border-top: 1px solid rgba(255, 255, 255, 0.08);
		margin: 1rem 0;
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
