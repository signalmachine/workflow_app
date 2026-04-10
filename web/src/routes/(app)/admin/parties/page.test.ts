import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AdminPartiesPage from './+page.svelte';

const baseData = {
	filters: {
		partyKind: 'customer'
	},
	parties: [
		{
			id: 'party-1',
			party_code: 'CUST-100',
			display_name: 'North Harbor Supply',
			party_kind: 'customer',
			status: 'active',
			updated_at: '2026-04-10T09:00:00Z'
		}
	]
};

describe('admin parties page', () => {
	it('keeps exact detail continuity available for support-party maintenance', () => {
		render(AdminPartiesPage, { props: { data: baseData } as never });

		expect(screen.getByText('Party setup')).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Create party' })).toBeTruthy();
		expect(screen.getByRole('link', { name: 'Open detail' }).getAttribute('href')).toBe(
			'/app/admin/parties/party-1'
		);
		expect(screen.getByRole('button', { name: 'Mark inactive' })).toBeTruthy();
	});
});
