import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import ControlBalancesPage from './+page.svelte';

describe('control balances page', () => {
	it('keeps control-account balance review on a dedicated destination', () => {
		render(ControlBalancesPage, {
			props: {
				data: {
					filters: { asOf: '2026-04-10', controlType: 'receivable', accountID: '' },
					balances: [
						{
							account_id: 'account-1',
							account_code: 'AR1000',
							account_name: 'Accounts Receivable',
							control_type: 'receivable',
							net_minor: 125000,
							last_effective_on: '2026-04-10'
						}
					]
				}
			} as never
		});

		expect(screen.getByDisplayValue('receivable')).toBeTruthy();
		expect(screen.getByText('AR1000 · Accounts Receivable')).toBeTruthy();
	});
});
