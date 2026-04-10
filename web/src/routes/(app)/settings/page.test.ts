import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import SettingsPage from './+page.svelte';

const baseData = {
	session: {
		user_display_name: 'Alex Operator',
		user_email: 'alex@example.com',
		org_name: 'North Harbor',
		role_code: 'admin'
	},
	dashboard: {
		primary_actions: [
			{
				title: 'Review pending approvals',
				summary: 'Keep decisions ahead of downstream document flow.',
				href: '/app/review/approvals?status=pending'
			}
		]
	}
};

describe('settings page', () => {
	it('keeps personal continuity separate from org-scoped admin maintenance', () => {
		render(SettingsPage, { props: { data: baseData } as never });

		expect(screen.getByText('User-scoped settings and continuity')).toBeTruthy();
		expect(screen.getByText(/Alex Operator, alex@example.com in North Harbor as admin\./)).toBeTruthy();
		expect(screen.getByText('Admin continuity')).toBeTruthy();
		expect(screen.getByRole('link', { name: /Review pending approvals/i }).getAttribute('href')).toBe(
			'/app/review/approvals?status=pending'
		);
	});
});
