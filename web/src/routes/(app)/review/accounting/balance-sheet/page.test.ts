import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import BalanceSheetPage from './+page.svelte';

describe('balance sheet page', () => {
	it('renders financial position totals and current earnings', () => {
		render(BalanceSheetPage, {
			props: {
				data: {
					filters: { asOf: '2026-04-10' },
					report: {
						as_of: '2026-04-10',
						lines: [
							{
								line_key: 'asset-1',
								account_id: 'asset-1',
								account_code: '1105',
								account_name: 'Accounts Receivable',
								account_class: 'asset',
								section: 'asset',
								amount_minor: 108000,
								is_synthetic: false
							},
							{
								line_key: 'current_earnings',
								account_name: 'Current earnings',
								account_class: 'equity',
								section: 'equity',
								amount_minor: 87400,
								is_synthetic: true
							}
						],
						total_assets_minor: 108000,
						total_liabilities_minor: 20600,
						total_equity_minor: 87400,
						total_liabilities_and_equity_minor: 108000,
						imbalance_minor: 0
					}
				}
			} as never
		});

		expect(screen.getByText('Balance sheet')).toBeTruthy();
		expect(screen.getByText('1105 · Accounts Receivable')).toBeTruthy();
		expect(screen.getByText('Current earnings')).toBeTruthy();
		expect(screen.getByText('Liabilities and equity')).toBeTruthy();
	});
});
