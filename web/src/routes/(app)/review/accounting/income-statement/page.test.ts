import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import IncomeStatementPage from './+page.svelte';

describe('income statement page', () => {
	it('renders revenue, expense, and net income totals', () => {
		render(IncomeStatementPage, {
			props: {
				data: {
					filters: { startOn: '2026-04-01', endOn: '2026-04-10' },
					report: {
						start_on: '2026-04-01',
						end_on: '2026-04-10',
						lines: [
							{
								line_key: 'revenue-1',
								account_id: 'revenue-1',
								account_code: '4105',
								account_name: 'Service Revenue',
								account_class: 'revenue',
								section: 'revenue',
								amount_minor: 100000,
								is_synthetic: false
							}
						],
						total_revenue_minor: 100000,
						total_expenses_minor: 12600,
						net_income_minor: 87400
					}
				}
			} as never
		});

		expect(screen.getByText('Income statement')).toBeTruthy();
		expect(screen.getByDisplayValue('2026-04-01')).toBeTruthy();
		expect(screen.getByDisplayValue('2026-04-10')).toBeTruthy();
		expect(screen.getByText('4105 · Service Revenue')).toBeTruthy();
		expect(screen.getAllByText('Net income')).toHaveLength(2);
	});
});
