<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import type { PageProps } from './$types';

	import FlashBanner from '$lib/components/feedback/FlashBanner.svelte';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { createContact, updatePartyStatus } from '$lib/api/admin';
	import { formatDateTime } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	let notice = $state('');
	let error = $state('');
	let submitting = $state(false);
	let contactForm = $state({
		full_name: '',
		role_title: '',
		email: '',
		phone: '',
		is_primary: false
	});

	function setErrorMessage(value: unknown, fallback: string): void {
		error = value instanceof Error ? value.message : fallback;
		notice = '';
	}

	async function submitContact(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		submitting = true;
		try {
			await createContact(data.party.id, {
				...contactForm,
				role_title: contactForm.role_title.trim() || undefined,
				email: contactForm.email.trim() || undefined,
				phone: contactForm.phone.trim() || undefined
			});
			contactForm = { full_name: '', role_title: '', email: '', phone: '', is_primary: false };
			await invalidateAll();
			notice = 'Party contact created.';
			error = '';
		} catch (cause) {
			setErrorMessage(cause, 'Failed to create contact.');
		} finally {
			submitting = false;
		}
	}

	async function changePartyStatus(status: string): Promise<void> {
		try {
			await updatePartyStatus(data.party.id, status);
			await invalidateAll();
			notice = `Party marked ${status}.`;
			error = '';
		} catch (cause) {
			setErrorMessage(cause, 'Failed to update party status.');
		}
	}
</script>

<PageHeader
	eyebrow="Admin"
	title={`${data.party.party_code} · ${data.party.display_name}`}
	description="Exact party maintenance and contact creation now continue in Svelte on the same shared parties seam."
/>

{#if notice}
	<FlashBanner kind="notice" message={notice} />
{/if}
{#if error}
	<FlashBanner kind="error" message={error} />
{/if}

<div class="page-stack">
	<SurfaceCard>
		<div class="detail-grid">
			<div><strong>Kind</strong><div>{data.party.party_kind}</div></div>
			<div><strong>Status</strong><div><StatusBadge status={data.party.status} /></div></div>
			<div><strong>Legal name</strong><div>{data.party.legal_name ?? '-'}</div></div>
			<div><strong>Updated</strong><div>{formatDateTime(data.party.updated_at)}</div></div>
		</div>
		<div class="action-row">
			<a href={routes.adminParties}>Back to parties</a>
			<button onclick={() => changePartyStatus(data.party.status === 'active' ? 'inactive' : 'active')} type="button">
				Mark {data.party.status === 'active' ? 'inactive' : 'active'}
			</button>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Create contact</p>
		<form class="admin-form" onsubmit={submitContact}>
			<input bind:value={contactForm.full_name} placeholder="Full name" required />
			<input bind:value={contactForm.role_title} placeholder="Role title" />
			<input bind:value={contactForm.email} placeholder="Email" type="email" />
			<input bind:value={contactForm.phone} placeholder="Phone" />
			<label class="checkbox-row"><input bind:checked={contactForm.is_primary} type="checkbox" />Primary contact</label>
			<button disabled={submitting} type="submit">Create contact</button>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Contacts</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Name</th><th>Role</th><th>Email</th><th>Phone</th><th>Primary</th><th>Status</th></tr></thead>
				<tbody>
					{#each data.contacts as contact (contact.id)}
						<tr>
							<td>{contact.full_name}</td>
							<td>{contact.role_title ?? '-'}</td>
							<td>{contact.email ?? '-'}</td>
							<td>{contact.phone ?? '-'}</td>
							<td>{contact.is_primary ? 'Yes' : 'No'}</td>
							<td><StatusBadge status={contact.status} /></td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>
</div>

<style>
	.detail-grid {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
	}

	.action-row {
		align-items: center;
		display: flex;
		gap: 0.75rem;
		justify-content: space-between;
		margin-top: 1rem;
	}

	.admin-form {
		display: grid;
		gap: 0.75rem;
		max-width: 28rem;
	}

	.checkbox-row {
		align-items: center;
		display: flex;
		gap: 0.5rem;
	}
</style>
