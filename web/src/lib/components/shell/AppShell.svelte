<script lang="ts">
	import type { Snippet } from 'svelte';

	import ContextTabs from '$lib/components/shell/ContextTabs.svelte';
	import SideNav from '$lib/components/shell/SideNav.svelte';
	import TopBar from '$lib/components/shell/TopBar.svelte';
	import { getNavigationModel } from '$lib/utils/navigation';

	interface Props {
		children: Snippet;
		currentPath: string;
		orgName: string;
		roleCode: string;
		userDisplayName: string;
		onLogout: () => void;
	}

	let { children, currentPath, orgName, roleCode, userDisplayName, onLogout }: Props = $props();

	let navOpen = $state(false);
	let navigation = $derived(getNavigationModel(currentPath, roleCode));

	function toggleNav(): void {
		navOpen = !navOpen;
	}

	function closeNav(): void {
		navOpen = false;
	}
</script>

<TopBar {onLogout} {orgName} {roleCode} {userDisplayName} onToggleNav={toggleNav} />
<ContextTabs currentPath={currentPath} tabs={navigation.activeArea.tabs} />
<SideNav areas={navigation.areas} currentPath={currentPath} isOpen={navOpen} onClose={closeNav} />

<div class="shell">
	<main class="content">
		{@render children()}
	</main>
</div>

<style>
	.shell {
		padding-left: 240px;
	}

	.content {
		margin: 0 auto;
		max-width: 1100px;
		min-height: calc(100vh - 102px);
		padding: var(--space-8);
	}

	@media (max-width: 767px) {
		.shell {
			padding-left: 0;
		}

		.content {
			padding: 1.25rem 1rem 2rem;
		}
	}
</style>
