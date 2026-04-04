<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import type { PageProps } from './$types';

	import FlashBanner from '$lib/components/feedback/FlashBanner.svelte';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import { createParty, updatePartyStatus } from '$lib/api/admin';
	import { partyKindOptions } from '$lib/utils/admin';
	import { formatDateTime } from '$lib/utils/format';
	import { adminPartyDetail, routes, withQuery } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	let notice = $state('');
	let error = $state('');
	let submitting = $state(false);
	let partyForm = $state({
		party_code: '',
		display_name: '',
		legal_name: '',
		party_kind: partyKindOptions[0]
	});

	function setErrorMessage(value: unknown, fallback: string): void {
		error = value instanceof Error ? value.message : fallback;
		notice = '';
	}

	async function submitParty(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		submitting = true;
		try {
			await createParty({
				...partyForm,
				legal_name: partyForm.legal_name.trim() || undefined
			});
			partyForm = { party_code: '', display_name: '', legal_name: '', party_kind: partyKindOptions[0] };
			await invalidateAll();
			notice = 'Party created.';
			error = '';
		} catch (cause) {
			setErrorMessage(cause, 'Failed to create party.');
		} finally {
			submitting = false;
		}
	}

	async function changePartyStatus(partyID: string, status: string): Promise<void> {
		try {
			await updatePartyStatus(partyID, status);
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
	title="Party setup"
	description="Support-party maintenance stays bounded here, with exact contact continuity now continuing into dedicated Svelte party detail routes."
/>

{#if notice}
	<FlashBanner kind="notice" message={notice} />
{/if}
{#if error}
	<FlashBanner kind="error" message={error} />
{/if}

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.adminParties} class="filter-row" method="get">
			<select name="party_kind">
				<option value="">All party kinds</option>
				{#each partyKindOptions as option (option)}
					<option selected={data.filters.partyKind === option} value={option}>{option}</option>
				{/each}
			</select>
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.adminParties}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Create party</p>
		<form class="admin-form" onsubmit={submitParty}>
			<input bind:value={partyForm.party_code} placeholder="Party code" required />
			<input bind:value={partyForm.display_name} placeholder="Display name" required />
			<input bind:value={partyForm.legal_name} placeholder="Legal name" />
			<select bind:value={partyForm.party_kind}>
				{#each partyKindOptions as option (option)}
					<option value={option}>{option}</option>
				{/each}
			</select>
			<button disabled={submitting} type="submit">Create party</button>
		</form>
	</SurfaceCard>

	<SurfaceCard>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Party</th><th>Kind</th><th>Status</th><th>Updated</th><th>Detail</th><th>Action</th></tr></thead>
				<tbody>
					{#each data.parties as party (party.id)}
						<tr>
							<td>{party.party_code} · {party.display_name}</td>
							<td>{party.party_kind}</td>
							<td><StatusBadge status={party.status} /></td>
							<td>{formatDateTime(party.updated_at)}</td>
							<td><a href={adminPartyDetail(party.id)}>Open detail</a></td>
							<td>
								<button onclick={() => changePartyStatus(party.id, party.status === 'active' ? 'inactive' : 'active')} type="button">
									Mark {party.status === 'active' ? 'inactive' : 'active'}
								</button>
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
