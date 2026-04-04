<script lang="ts">
	import type { Snippet } from 'svelte';

	import SideNav from '$lib/components/shell/SideNav.svelte';
	import TopBar from '$lib/components/shell/TopBar.svelte';

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

	function toggleNav(): void {
		navOpen = !navOpen;
	}

	function closeNav(): void {
		navOpen = false;
	}
</script>

<TopBar {onLogout} {orgName} {roleCode} {userDisplayName} onToggleNav={toggleNav} />
<SideNav currentPath={currentPath} isOpen={navOpen} onClose={closeNav} {roleCode} />

<div class="shell">
	<main class="content">
		{@render children()}
	</main>
</div>

<style>
	.shell {
		padding-left: 220px;
	}

	.content {
		margin: 0 auto;
		max-width: 1100px;
		min-height: calc(100vh - 48px);
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
