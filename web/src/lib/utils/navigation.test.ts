import { describe, expect, it } from 'vitest';

import { getNavigationModel, isNavigationItemActive } from './navigation';

describe('navigation model', () => {
	it('maps inbound request detail continuity into the agent area', () => {
		const model = getNavigationModel('/app/inbound-requests/REQ-000123', 'operator');

		expect(model.activeArea.id).toBe('agent');
		expect(model.activeArea.tabs.map((tab) => tab.label)).toEqual(['Messages', 'Requests', 'Lists']);
	});

	it('maps accounting review routes into the accounting area', () => {
		const model = getNavigationModel('/app/review/accounting/entry-123', 'operator');

		expect(model.activeArea.id).toBe('accounting');
		expect(model.activeArea.tabs.map((tab) => tab.label)).toContain('Reports');
	});

	it('keeps admin hidden for non-admin actors', () => {
		const model = getNavigationModel('/app/settings', 'operator');

		expect(model.areas.some((area) => area.id === 'admin')).toBe(false);
	});

	it('includes admin tabs for admin actors', () => {
		const model = getNavigationModel('/app/admin/inventory', 'admin');

		expect(model.activeArea.id).toBe('admin');
		expect(model.activeArea.tabs.map((tab) => tab.label)).toEqual([
			'Overview',
			'Accounting',
			'Parties',
			'Access',
			'Inventory'
		]);
	});
});

describe('isNavigationItemActive', () => {
	it('matches prefixes without falsely activating the home route', () => {
		expect(
			isNavigationItemActive('/app/review/inventory/movement-123', [{ path: '/app/review/inventory' }])
		).toBe(true);
		expect(isNavigationItemActive('/app/review', [{ path: '/app', exact: true }])).toBe(false);
	});
});
