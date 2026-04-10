import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AdminAccountingPage from './+page.svelte';

const baseData = {
	ledgerAccounts: [
		{
			id: 'ledger-1',
			code: 'AR1000',
			name: 'Accounts Receivable',
			account_class: 'asset',
			control_type: 'receivable',
			status: 'active',
			updated_at: '2026-04-09T12:00:00Z'
		},
		{
			id: 'ledger-2',
			code: 'GST2000',
			name: 'GST Payable',
			account_class: 'liability',
			control_type: 'gst_output',
			status: 'active',
			updated_at: '2026-04-09T12:00:00Z'
		}
	],
	taxCodes: [
		{
			id: 'tax-1',
			code: 'GST',
			name: 'Goods and Services Tax',
			tax_type: 'sales',
			rate_basis_points: 1000,
			status: 'inactive',
			updated_at: '2026-04-09T12:00:00Z'
		}
	],
	periods: [
		{
			id: 'period-1',
			period_code: '2026-04',
			start_on: '2026-04-01',
			end_on: '2026-04-30',
			status: 'open',
			closed_at: null
		}
	]
};

describe('admin accounting page', () => {
	it('keeps master-data status controls visible in the promoted Svelte admin surface', () => {
		render(AdminAccountingPage, { props: { data: baseData } as never });

		expect(screen.getByText('Accounting setup')).toBeTruthy();
		expect(screen.getAllByRole('button', { name: 'Mark inactive' })).toHaveLength(2);
		expect(screen.getByRole('button', { name: 'Mark active' })).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Close period' })).toBeTruthy();
		expect(screen.getByRole('option', { name: 'GST2000 · GST Payable' })).toBeTruthy();
		expect(screen.queryByPlaceholderText('Payable account id')).toBeNull();
	});
});
