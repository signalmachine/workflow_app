import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AdminMasterDataPage from './+page.svelte';

describe('admin master-data directory page', () => {
	it('routes setup work through grouped destinations', () => {
		render(AdminMasterDataPage, { props: {} as never });

		expect(screen.getByText('Master data')).toBeTruthy();
		expect(screen.getByRole('link', { name: /Accounting master data/i }).getAttribute('href')).toBe(
			'/app/admin/accounting'
		);
		expect(screen.getByRole('link', { name: /Party master data/i }).getAttribute('href')).toBe(
			'/app/admin/parties'
		);
		expect(screen.getByRole('link', { name: /Inventory master data/i }).getAttribute('href')).toBe(
			'/app/admin/inventory'
		);
	});
});
