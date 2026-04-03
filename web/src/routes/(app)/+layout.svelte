<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	import FlashBanner from '$lib/components/feedback/FlashBanner.svelte';
	import AppShell from '$lib/components/shell/AppShell.svelte';
	import { logout } from '$lib/api/session';
	import { routes } from '$lib/utils/routes';

	let { data, children } = $props();

	let currentPath = $derived(page.url.pathname);
	let notice = $derived(page.url.searchParams.get('notice') ?? '');
	let error = $derived(page.url.searchParams.get('error') ?? '');

	async function handleLogout(): Promise<void> {
		await logout();
		await goto(`${routes.login}?notice=${encodeURIComponent('Signed out.')}`);
	}
</script>

<svelte:head>
	<title>workflow_app</title>
</svelte:head>

<AppShell
	currentPath={currentPath}
	onLogout={handleLogout}
	orgName={data.session.org_name}
	userDisplayName={data.session.user_display_name}
>
	{#if notice}
		<FlashBanner kind="notice" message={notice} />
	{/if}
	{#if error}
		<FlashBanner kind="error" message={error} />
	{/if}
	{@render children()}
</AppShell>
