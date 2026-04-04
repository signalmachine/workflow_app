<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import type { PageProps } from './$types';

	import FlashBanner from '$lib/components/feedback/FlashBanner.svelte';
	import PageHeader from '$lib/components/layout/PageHeader.svelte';
	import StatusBadge from '$lib/components/primitives/StatusBadge.svelte';
	import SurfaceCard from '$lib/components/primitives/SurfaceCard.svelte';
	import {
		closeAccountingPeriod,
		createAccountingPeriod,
		createLedgerAccount,
		createTaxCode,
		updateLedgerAccountStatus,
		updateTaxCodeStatus
	} from '$lib/api/admin';
	import { accountClassOptions, controlTypeOptions, statusOptions, taxTypeOptions } from '$lib/utils/admin';
	import { formatDate, formatDateTime } from '$lib/utils/format';

	let { data }: PageProps = $props();

	let notice = $state('');
	let error = $state('');
	let submitting = $state(false);

	let accountForm = $state({
		code: '',
		name: '',
		account_class: accountClassOptions[0],
		control_type: controlTypeOptions[0],
		allows_direct_posting: false,
		tax_category_code: ''
	});
	let taxForm = $state({
		code: '',
		name: '',
		tax_type: taxTypeOptions[0],
		rate_basis_points: 0,
		receivable_account_id: '',
		payable_account_id: ''
	});
	let periodForm = $state({
		period_code: '',
		start_on: '',
		end_on: ''
	});

	function setErrorMessage(value: unknown, fallback: string): void {
		error = value instanceof Error ? value.message : fallback;
		notice = '';
	}

	async function refresh(message: string): Promise<void> {
		await invalidateAll();
		notice = message;
		error = '';
	}

	async function submitLedgerAccount(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		submitting = true;
		try {
			await createLedgerAccount({
				...accountForm,
				tax_category_code: accountForm.tax_category_code.trim() || undefined
			});
			accountForm = {
				code: '',
				name: '',
				account_class: accountClassOptions[0],
				control_type: controlTypeOptions[0],
				allows_direct_posting: false,
				tax_category_code: ''
			};
			await refresh('Ledger account created.');
		} catch (cause) {
			setErrorMessage(cause, 'Failed to create ledger account.');
		} finally {
			submitting = false;
		}
	}

	async function submitTaxCode(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		submitting = true;
		try {
			await createTaxCode({
				...taxForm,
				receivable_account_id: taxForm.receivable_account_id.trim() || undefined,
				payable_account_id: taxForm.payable_account_id.trim() || undefined
			});
			taxForm = {
				code: '',
				name: '',
				tax_type: taxTypeOptions[0],
				rate_basis_points: 0,
				receivable_account_id: '',
				payable_account_id: ''
			};
			await refresh('Tax code created.');
		} catch (cause) {
			setErrorMessage(cause, 'Failed to create tax code.');
		} finally {
			submitting = false;
		}
	}

	async function submitAccountingPeriod(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		submitting = true;
		try {
			await createAccountingPeriod(periodForm);
			periodForm = { period_code: '', start_on: '', end_on: '' };
			await refresh('Accounting period created.');
		} catch (cause) {
			setErrorMessage(cause, 'Failed to create accounting period.');
		} finally {
			submitting = false;
		}
	}

	async function changeLedgerStatus(accountID: string, status: string): Promise<void> {
		try {
			await updateLedgerAccountStatus(accountID, status);
			await refresh(`Ledger account marked ${status}.`);
		} catch (cause) {
			setErrorMessage(cause, 'Failed to update ledger account status.');
		}
	}

	async function changeTaxStatus(taxCodeID: string, status: string): Promise<void> {
		try {
			await updateTaxCodeStatus(taxCodeID, status);
			await refresh(`Tax code marked ${status}.`);
		} catch (cause) {
			setErrorMessage(cause, 'Failed to update tax code status.');
		}
	}

	async function closePeriod(periodID: string): Promise<void> {
		try {
			await closeAccountingPeriod(periodID);
			await refresh('Accounting period closed.');
		} catch (cause) {
			setErrorMessage(cause, 'Failed to close accounting period.');
		}
	}
</script>

<PageHeader
	eyebrow="Admin"
	title="Accounting setup"
	description="Ledger-account, tax-code, and accounting-period maintenance now runs directly from the Svelte shell against the shared admin seam."
/>

{#if notice}
	<FlashBanner kind="notice" message={notice} />
{/if}
{#if error}
	<FlashBanner kind="error" message={error} />
{/if}

<div class="page-stack">
	<div class="admin-grid">
		<SurfaceCard>
			<p class="eyebrow">Create ledger account</p>
			<form class="admin-form" onsubmit={submitLedgerAccount}>
				<input bind:value={accountForm.code} placeholder="Code" required />
				<input bind:value={accountForm.name} placeholder="Name" required />
				<select bind:value={accountForm.account_class}>
					{#each accountClassOptions as option (option)}
						<option value={option}>{option}</option>
					{/each}
				</select>
				<select bind:value={accountForm.control_type}>
					{#each controlTypeOptions as option (option)}
						<option value={option}>{option}</option>
					{/each}
				</select>
				<input bind:value={accountForm.tax_category_code} placeholder="Tax category code" />
				<label class="checkbox-row"><input bind:checked={accountForm.allows_direct_posting} type="checkbox" />Allow direct posting</label>
				<button disabled={submitting} type="submit">Create account</button>
			</form>
		</SurfaceCard>

		<SurfaceCard>
			<p class="eyebrow">Create tax code</p>
			<form class="admin-form" onsubmit={submitTaxCode}>
				<input bind:value={taxForm.code} placeholder="Code" required />
				<input bind:value={taxForm.name} placeholder="Name" required />
				<select bind:value={taxForm.tax_type}>
					{#each taxTypeOptions as option (option)}
						<option value={option}>{option}</option>
					{/each}
				</select>
				<input bind:value={taxForm.rate_basis_points} min="0" placeholder="Rate basis points" required type="number" />
				<input bind:value={taxForm.receivable_account_id} placeholder="Receivable account id" />
				<input bind:value={taxForm.payable_account_id} placeholder="Payable account id" />
				<button disabled={submitting} type="submit">Create tax code</button>
			</form>
		</SurfaceCard>

		<SurfaceCard>
			<p class="eyebrow">Create accounting period</p>
			<form class="admin-form" onsubmit={submitAccountingPeriod}>
				<input bind:value={periodForm.period_code} placeholder="Period code" required />
				<input bind:value={periodForm.start_on} required type="date" />
				<input bind:value={periodForm.end_on} required type="date" />
				<button disabled={submitting} type="submit">Create period</button>
			</form>
		</SurfaceCard>
	</div>

	<SurfaceCard>
		<p class="eyebrow">Ledger accounts</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Code</th><th>Name</th><th>Class</th><th>Control</th><th>Status</th><th>Updated</th><th>Action</th></tr></thead>
				<tbody>
					{#each data.ledgerAccounts as account (account.id)}
						<tr>
							<td>{account.code}</td>
							<td>{account.name}</td>
							<td>{account.account_class}</td>
							<td>{account.control_type}</td>
							<td><StatusBadge status={account.status} /></td>
							<td>{formatDateTime(account.updated_at)}</td>
							<td>
								<button onclick={() => changeLedgerStatus(account.id, account.status === 'active' ? 'inactive' : 'active')} type="button">
									Mark {account.status === 'active' ? 'inactive' : 'active'}
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Tax codes</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Code</th><th>Name</th><th>Type</th><th>Rate</th><th>Status</th><th>Updated</th><th>Action</th></tr></thead>
				<tbody>
					{#each data.taxCodes as code (code.id)}
						<tr>
							<td>{code.code}</td>
							<td>{code.name}</td>
							<td>{code.tax_type}</td>
							<td>{code.rate_basis_points}</td>
							<td><StatusBadge status={code.status} /></td>
							<td>{formatDateTime(code.updated_at)}</td>
							<td>
								<button onclick={() => changeTaxStatus(code.id, code.status === 'active' ? 'inactive' : 'active')} type="button">
									Mark {code.status === 'active' ? 'inactive' : 'active'}
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</SurfaceCard>

	<SurfaceCard>
		<p class="eyebrow">Accounting periods</p>
		<div class="table-wrap">
			<table class="surface-table">
				<thead><tr><th>Period</th><th>Range</th><th>Status</th><th>Closed</th><th>Action</th></tr></thead>
				<tbody>
					{#each data.periods as period (period.id)}
						<tr>
							<td>{period.period_code}</td>
							<td>{formatDate(period.start_on)} to {formatDate(period.end_on)}</td>
							<td><StatusBadge status={period.status} /></td>
							<td>{formatDateTime(period.closed_at)}</td>
							<td>
								{#if period.status === 'open'}
									<button onclick={() => closePeriod(period.id)} type="button">Close period</button>
								{:else}
									<span class="muted-copy">Closed</span>
								{/if}
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

	.checkbox-row {
		align-items: center;
		display: flex;
		gap: 0.5rem;
	}
</style>
