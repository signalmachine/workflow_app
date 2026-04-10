import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AdminInventoryPage from './+page.svelte';

const baseData = {
	filters: {
		itemRole: 'service_material',
		locationRole: 'warehouse'
	},
	items: [
		{
			id: 'item-1',
			sku: 'SKU-1',
			name: 'Copper pipe',
			item_role: 'service_material',
			tracking_mode: 'stocked',
			status: 'active',
			updated_at: '2026-04-09T12:00:00Z'
		}
	],
	locations: [
		{
			id: 'location-1',
			code: 'MAIN',
			name: 'Main warehouse',
			location_role: 'warehouse',
			status: 'inactive',
			updated_at: '2026-04-09T12:00:00Z'
		}
	]
};

describe('admin inventory page', () => {
	it('keeps item and location status controls visible alongside filter continuity', () => {
		render(AdminInventoryPage, { props: { data: baseData } as never });

		expect(screen.getByDisplayValue('service_material')).toBeTruthy();
		expect(screen.getAllByDisplayValue('warehouse')).toHaveLength(2);
		expect(screen.getByRole('button', { name: 'Mark inactive' })).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Mark active' })).toBeTruthy();
		expect(screen.getByRole('link', { name: 'Clear' }).getAttribute('href')).toBe(
			'/app/admin/inventory'
		);
	});
});
