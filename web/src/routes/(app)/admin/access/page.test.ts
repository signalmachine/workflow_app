import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AdminAccessPage from './+page.svelte';

const baseData = {
	users: [
		{
			membership_id: 'membership-1',
			user_display_name: 'Jamie Admin',
			user_email: 'jamie@example.com',
			role_code: 'operator',
			user_status: 'active',
			membership_status: 'active',
			created_at: '2026-04-10T09:00:00Z'
		}
	]
};

describe('admin access page', () => {
	it('keeps provisioning and role changes visible on the shared identity seam', () => {
		render(AdminAccessPage, { props: { data: baseData } as never });

		expect(screen.getByText('Access controls')).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Provision user' })).toBeTruthy();
		expect(screen.getAllByDisplayValue('operator')).toHaveLength(2);
		expect(screen.getByText(/Jamie Admin · jamie@example.com/)).toBeTruthy();
	});
});
