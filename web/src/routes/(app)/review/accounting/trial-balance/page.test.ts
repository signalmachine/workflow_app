import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import TrialBalancePage from './+page.svelte';

describe('trial balance page', () => {
	it('renders trial balance totals and account rows', () => {
		render(TrialBalancePage, {
			props: {
				data: {
					filters: { asOf: '2026-04-10' },
					report: {
						as_of: '2026-04-10',
						lines: [
							{
								account_id: 'account-1',
								account_code: '1105',
								account_name: 'Accounts Receivable',
								account_class: 'asset',
								total_debit_minor: 108000,
								total_credit_minor: 0,
								net_minor: 108000,
								debit_balance_minor: 108000,
								credit_balance_minor: 0,
								last_effective_on: '2026-04-10'
							}
						],
						total_debit_balance_minor: 108000,
						total_credit_balance_minor: 108000,
						imbalance_minor: 0
					}
				}
			} as never
		});

		expect(screen.getByText('Trial balance')).toBeTruthy();
		expect(screen.getByDisplayValue('2026-04-10')).toBeTruthy();
		expect(screen.getByText('1105 · Accounts Receivable')).toBeTruthy();
		expect(screen.getByText('Asset')).toBeTruthy();
	});
});
