<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	import FlashBanner from '$lib/components/feedback/FlashBanner.svelte';
	import { login } from '$lib/api/session';
	import { routes } from '$lib/utils/routes';

	let orgSlug = $state('');
	let email = $state('');
	let password = $state('');
	let submitting = $state(false);
	let errorMessage = $state('');

	let notice = $derived(page.url.searchParams.get('notice') ?? '');
	let next = $derived(page.url.searchParams.get('next') ?? routes.home);

	async function submit(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		submitting = true;
		errorMessage = '';

		try {
			await login({
				org_slug: orgSlug.trim(),
				email: email.trim(),
				password,
				device_label: 'browser'
			});
			await goto(next);
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Failed to sign in';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>workflow_app sign in</title>
</svelte:head>

<main class="login-shell">
	<section class="login-panel">
		<p class="eyebrow">Milestone 13 Slice 1</p>
		<h1>Operator sign-in</h1>
		<p class="intro">
			This Svelte login surface uses the existing cookie-auth session seam at <span class="mono">POST /api/session/login</span>.
		</p>

		{#if notice}
			<FlashBanner kind="notice" message={notice} />
		{/if}
		{#if errorMessage}
			<FlashBanner kind="error" message={errorMessage} />
		{/if}

		<form class="login-form" onsubmit={submit}>
			<label>
				<span>Org slug</span>
				<input bind:value={orgSlug} autocomplete="organization" placeholder="north-harbor" required />
			</label>

			<label>
				<span>Email</span>
				<input bind:value={email} autocomplete="email" placeholder="admin@example.com" required type="email" />
			</label>

			<label>
				<span>Password</span>
				<input bind:value={password} autocomplete="current-password" required type="password" />
			</label>

			<button disabled={submitting} type="submit">
				{submitting ? 'Signing in...' : 'Sign in'}
			</button>
		</form>
	</section>
</main>

<style>
	.login-shell {
		align-items: center;
		display: grid;
		min-height: 100vh;
		padding: 1.5rem;
	}

	.login-panel {
		background: var(--surface-strong);
		border: 1px solid var(--line);
		border-radius: 24px;
		box-shadow: var(--shadow);
		margin: 0 auto;
		max-width: 28rem;
		padding: 2rem;
		width: 100%;
	}

	.login-panel h1 {
		font-size: var(--text-xl);
		font-weight: 600;
		margin: 0.3rem 0 0.625rem;
	}

	.intro {
		color: var(--ink-soft);
		margin: 0 0 1.25rem;
	}

	.login-form {
		display: grid;
		gap: 1rem;
	}

	.login-form label {
		display: grid;
		gap: 0.4rem;
	}

	.login-form span {
		color: var(--ink-soft);
		font-size: var(--text-xs);
		font-weight: 500;
	}

	.login-form input {
		background: white;
		border: 1px solid var(--line);
		border-radius: 12px;
		padding: 0.8rem 0.9rem;
	}

	.login-form button {
		background: var(--accent-strong);
		border: 0;
		border-radius: 999px;
		color: white;
		font-weight: 600;
		padding: 0.8rem 1rem;
	}

	.login-form button:disabled {
		opacity: 0.7;
	}
</style>
