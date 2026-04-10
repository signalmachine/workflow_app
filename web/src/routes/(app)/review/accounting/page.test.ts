import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AccountingDirectoryPage from './+page.svelte';

describe('accounting directory page', () => {
	it('routes accounting review through dedicated report destinations', () => {
		render(AccountingDirectoryPage, { props: {} as never });

		expect(screen.getByText('Accounting reports')).toBeTruthy();
		expect(screen.getByRole('link', { name: /Journal entries/i }).getAttribute('href')).toBe(
			'/app/review/accounting/journal-entries'
		);
		expect(screen.getByRole('link', { name: /Control balances/i }).getAttribute('href')).toBe(
			'/app/review/accounting/control-balances'
		);
		expect(screen.getByRole('link', { name: /Tax summaries/i }).getAttribute('href')).toBe(
			'/app/review/accounting/tax-summaries'
		);
		expect(screen.getByRole('link', { name: /Trial balance/i }).getAttribute('href')).toBe(
			'/app/review/accounting/trial-balance'
		);
		expect(screen.getByRole('link', { name: /Balance sheet/i }).getAttribute('href')).toBe(
			'/app/review/accounting/balance-sheet'
		);
		expect(screen.getByRole('link', { name: /Income statement/i }).getAttribute('href')).toBe(
			'/app/review/accounting/income-statement'
		);
	});
});
