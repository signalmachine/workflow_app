<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import type { PageProps } from './$types';

	import FlashBanner from '$lib/components/feedback/FlashBanner.svelte';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import {
		createInventoryItem,
		createInventoryLocation,
		updateInventoryItemStatus,
		updateInventoryLocationStatus
	} from '$lib/api/admin';
	import { itemRoleOptions, locationRoleOptions, trackingModeOptions } from '$lib/utils/admin';
	import { formatDateTime } from '$lib/utils/format';
	import { routes } from '$lib/utils/routes';

	let { data }: PageProps = $props();

	let notice = $state('');
	let error = $state('');
	let submitting = $state(false);
	let itemForm = $state({
		sku: '',
		name: '',
		item_role: itemRoleOptions[0],
		tracking_mode: trackingModeOptions[0]
	});
	let locationForm = $state({
		code: '',
		name: '',
		location_role: locationRoleOptions[0]
	});

	function setErrorMessage(value: unknown, fallback: string): void {
		error = value instanceof Error ? value.message : fallback;
		notice = '';
	}

	async function submitItem(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		submitting = true;
		try {
			await createInventoryItem(itemForm);
			itemForm = { sku: '', name: '', item_role: itemRoleOptions[0], tracking_mode: trackingModeOptions[0] };
			await invalidateAll();
			notice = 'Inventory item created.';
			error = '';
		} catch (cause) {
			setErrorMessage(cause, 'Failed to create inventory item.');
		} finally {
			submitting = false;
		}
	}

	async function submitLocation(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		submitting = true;
		try {
			await createInventoryLocation(locationForm);
			locationForm = { code: '', name: '', location_role: locationRoleOptions[0] };
			await invalidateAll();
			notice = 'Inventory location created.';
			error = '';
		} catch (cause) {
			setErrorMessage(cause, 'Failed to create inventory location.');
		} finally {
			submitting = false;
		}
	}

	async function changeItemStatus(itemID: string, status: string): Promise<void> {
		try {
			await updateInventoryItemStatus(itemID, status);
			await invalidateAll();
			notice = `Inventory item marked ${status}.`;
			error = '';
		} catch (cause) {
			setErrorMessage(cause, 'Failed to update inventory item status.');
		}
	}

	async function changeLocationStatus(locationID: string, status: string): Promise<void> {
		try {
			await updateInventoryLocationStatus(locationID, status);
			await invalidateAll();
			notice = `Inventory location marked ${status}.`;
			error = '';
		} catch (cause) {
			setErrorMessage(cause, 'Failed to update inventory location status.');
		}
	}
</script>

<PageHeader
	eyebrow="Admin"
	title="Inventory setup"
	description="Item and location maintenance now runs from Svelte while keeping the inventory foundation and status controls on the shared backend."
/>

{#if notice}
	<FlashBanner kind="notice" message={notice} />
{/if}
{#if error}
	<FlashBanner kind="error" message={error} />
{/if}

<div class="page-stack">
	<SurfaceCard>
		<form action={routes.adminInventory} class="filter-row" method="get">
			<select name="item_role">
				<option value="">All item roles</option>
				{#each itemRoleOptions as option (option)}
					<option selected={data.filters.itemRole === option} value={option}>{option}</option>
				{/each}
			</select>
			<select name="location_role">
				<option value="">All location roles</option>
				{#each locationRoleOptions as option (option)}
					<option selected={data.filters.locationRole === option} value={option}>{option}</option>
				{/each}
			</select>
			<div class="filter-actions">
				<button type="submit">Filter</button>
				<a href={routes.adminInventory}>Clear</a>
			</div>
		</form>
	</SurfaceCard>

	<div class="admin-grid">
		<SurfaceCard>
			<p class="eyebrow">Create inventory item</p>
			<form class="admin-form" onsubmit={submitItem}>
				<input bind:value={itemForm.sku} placeholder="SKU" required />
				<input bind:value={itemForm.name} placeholder="Name" required />
				<select bind:value={itemForm.item_role}>
					{#each itemRoleOptions as option (option)}
						<option value={option}>{option}</option>
					{/each}
				</select>
				<select bind:value={itemForm.tracking_mode}>
					{#each trackingModeOptions as option (option)}
						<option value={option}>{option}</option>
					{/each}
				</select>
				<button disabled={submitting} type="submit">Create item</button>
			</form>
		</SurfaceCard>

		<SurfaceCard>
			<p class="eyebrow">Create inventory location</p>
			<form class="admin-form" onsubmit={submitLocation}>
				<input bind:value={locationForm.code} placeholder="Code" required />
				<input bind:value={locationForm.name} placeholder="Name" required />
				<select bind:value={locationForm.location_role}>
					{#each locationRoleOptions as option (option)}
						<option value={option}>{option}</option>
					{/each}
				</select>
				<button disabled={submitting} type="submit">Create location</button>
			</form>
		</SurfaceCard>
	</div>

	<SurfaceCard>
		<p class="eyebrow">Items</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>SKU</th><th>Name</th><th>Role</th><th>Tracking</th><th>Status</th><th>Updated</th><th>Action</th></tr></thead>
				<tbody>
					{#each data.items as item (item.id)}
						<tr>
							<td>{item.sku}</td>
							<td>{item.name}</td>
							<td>{item.item_role}</td>
							<td>{item.tracking_mode}</td>
							<td><StatusBadge status={item.status} /></td>
							<td>{formatDateTime(item.updated_at)}</td>
							<td>
								<button onclick={() => changeItemStatus(item.id, item.status === 'active' ? 'inactive' : 'active')} type="button">
									Mark {item.status === 'active' ? 'inactive' : 'active'}
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Locations</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Code</th><th>Name</th><th>Role</th><th>Status</th><th>Updated</th><th>Action</th></tr></thead>
				<tbody>
					{#each data.locations as location (location.id)}
						<tr>
							<td>{location.code}</td>
							<td>{location.name}</td>
							<td>{location.location_role}</td>
							<td><StatusBadge status={location.status} /></td>
							<td>{formatDateTime(location.updated_at)}</td>
							<td>
								<button onclick={() => changeLocationStatus(location.id, location.status === 'active' ? 'inactive' : 'active')} type="button">
									Mark {location.status === 'active' ? 'inactive' : 'active'}
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
	.admin-grid {
		display: grid;
		gap: var(--space-4);
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
	}

	.admin-form {
		display: grid;
		gap: 0.75rem;
	}
</style>
