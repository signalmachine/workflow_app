import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AdminListsPage from './+page.svelte';

describe('admin lists directory page', () => {
	it('routes maintained-record review through grouped destinations', () => {
		render(AdminListsPage, { props: {} as never });

		expect(screen.getByText('Lists')).toBeTruthy();
		expect(screen.getByRole('link', { name: /Ledger, tax, and period lists/i }).getAttribute('href')).toBe(
			'/app/admin/accounting'
		);
		expect(screen.getByRole('link', { name: /Party list/i }).getAttribute('href')).toBe('/app/admin/parties');
		expect(screen.getByRole('link', { name: /Item and location lists/i }).getAttribute('href')).toBe(
			'/app/admin/inventory'
		);
	});
});
