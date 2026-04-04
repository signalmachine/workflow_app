<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import type { PageProps } from './$types';

	import FlashBanner from '$lib/components/feedback/FlashBanner.svelte';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { provisionOrgUser, updateMembershipRole } from '$lib/api/admin';
	import { formatDateTime } from '$lib/utils/format';
	import { roleOptions } from '$lib/utils/admin';

	let { data }: PageProps = $props();

	let notice = $state('');
	let error = $state('');
	let submitting = $state(false);
	let userForm = $state({
		email: '',
		display_name: '',
		role_code: roleOptions[1],
		password: ''
	});

	function setErrorMessage(value: unknown, fallback: string): void {
		error = value instanceof Error ? value.message : fallback;
		notice = '';
	}

	async function submitUser(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		submitting = true;
		try {
			await provisionOrgUser(userForm);
			userForm = { email: '', display_name: '', role_code: roleOptions[1], password: '' };
			await invalidateAll();
			notice = 'Org user provisioned.';
			error = '';
		} catch (cause) {
			setErrorMessage(cause, 'Failed to provision org user.');
		} finally {
			submitting = false;
		}
	}

	async function changeRole(membershipID: string, roleCode: string): Promise<void> {
		try {
			await updateMembershipRole(membershipID, roleCode);
			await invalidateAll();
			notice = 'Membership role updated.';
			error = '';
		} catch (cause) {
			setErrorMessage(cause, 'Failed to update membership role.');
		}
	}
</script>

<PageHeader
	eyebrow="Admin"
	title="Access controls"
	description="Provision org users and update role assignments without splitting the shared session and membership model."
/>

{#if notice}
	<FlashBanner kind="notice" message={notice} />
{/if}
{#if error}
	<FlashBanner kind="error" message={error} />
{/if}

<div class="page-stack">
	<SurfaceCard>
		<p class="eyebrow">Provision org user</p>
		<form class="admin-form" onsubmit={submitUser}>
			<input bind:value={userForm.email} placeholder="Email" required type="email" />
			<input bind:value={userForm.display_name} placeholder="Display name" required />
			<select bind:value={userForm.role_code}>
				{#each roleOptions as option (option)}
					<option value={option}>{option}</option>
				{/each}
			</select>
			<input bind:value={userForm.password} minlength="8" placeholder="Password" required type="password" />
			<button disabled={submitting} type="submit">Provision user</button>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>User</th><th>Role</th><th>User status</th><th>Membership</th><th>Created</th><th>Change role</th></tr></thead>
				<tbody>
					{#each data.users as user (user.membership_id)}
						<tr>
							<td>{user.user_display_name} · {user.user_email}</td>
							<td>{user.role_code}</td>
							<td><StatusBadge status={user.user_status} /></td>
							<td><StatusBadge status={user.membership_status} /></td>
							<td>{formatDateTime(user.created_at)}</td>
							<td>
								<select onchange={(event) => changeRole(user.membership_id, (event.currentTarget as HTMLSelectElement).value)} value={user.role_code}>
									{#each roleOptions as option (option)}
										<option value={option}>{option}</option>
									{/each}
								</select>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>

<style>
	.admin-form {
		display: grid;
		gap: 0.75rem;
		max-width: 28rem;
	}
</style>
