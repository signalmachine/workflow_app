<script lang="ts">
	import type { Snippet } from 'svelte';
	import { onMount } from 'svelte';

	import ContextTabs from '$lib/components/shell/ContextTabs.svelte';
	import SideNav from '$lib/components/shell/SideNav.svelte';
	import TopBar from '$lib/components/shell/TopBar.svelte';
	import { getNavigationModel } from '$lib/utils/navigation';
	import { readDesktopSidebarCollapsed } from '$lib/utils/shell';

	interface Props {
		children: Snippet;
		currentPath: string;
		orgName: string;
		roleCode: string;
		userDisplayName: string;
		onLogout: () => void;
	}

	let { children, currentPath, orgName, roleCode, userDisplayName, onLogout }: Props = $props();

	let mobileNavOpen = $state(false);
	let desktopSidebarCollapsed = $state(false);
	let mobileViewport = $state(false);
	let navigation = $derived(getNavigationModel(currentPath, roleCode));

	function toggleNav(): void {
		if (mobileViewport) {
			mobileNavOpen = !mobileNavOpen;
			return;
		}

		desktopSidebarCollapsed = !desktopSidebarCollapsed;
		window.localStorage.setItem(
			'workflow_app.desktop_sidebar_collapsed',
			desktopSidebarCollapsed ? 'true' : 'false'
		);
	}

	function closeNav(): void {
		mobileNavOpen = false;
	}

	onMount(() => {
		const mediaQuery = window.matchMedia('(max-width: 767px)');

		function syncViewport(): void {
			mobileViewport = mediaQuery.matches;
			if (mobileViewport) {
				mobileNavOpen = false;
			}
		}

		desktopSidebarCollapsed = readDesktopSidebarCollapsed(window.localStorage);
		syncViewport();
		mediaQuery.addEventListener('change', syncViewport);

		return () => {
			mediaQuery.removeEventListener('change', syncViewport);
		};
	});
</script>

<TopBar
	{onLogout}
	{orgName}
	{roleCode}
	{userDisplayName}
	navExpanded={mobileViewport ? mobileNavOpen : !desktopSidebarCollapsed}
	onToggleNav={toggleNav}
/>
<div class="shell">
	<SideNav
		areas={navigation.areas}
		collapsed={desktopSidebarCollapsed}
		currentPath={currentPath}
		isOpen={mobileNavOpen}
		onClose={closeNav}
	/>
	<div class:collapsed={desktopSidebarCollapsed} class="main-column">
		<ContextTabs currentPath={currentPath} tabs={navigation.activeArea.tabs} />
		<main class="content">
			{@render children()}
		</main>
	</div>
</div>

<style>
	.shell {
		min-height: calc(100vh - 48px);
	}

	.main-column {
		margin-left: 240px;
		min-height: calc(100vh - 48px);
		transition: margin-left 160ms ease;
	}

	.main-column.collapsed {
		margin-left: 132px;
	}

	.content {
		margin: 0 auto;
		max-width: 1100px;
		min-height: calc(100vh - 102px);
		padding: var(--space-8);
	}

	@media (max-width: 767px) {
		.main-column,
		.main-column.collapsed {
			margin-left: 0;
		}

		.content {
			padding: 1.25rem 1rem 2rem;
		}
	}
</style>
