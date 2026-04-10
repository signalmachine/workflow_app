import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AdminPage from './+page.svelte';

describe('admin hub page', () => {
	it('keeps privileged maintenance bounded to the promoted setup surfaces', () => {
		render(AdminPage, { props: {} as never });

		expect(screen.getByText('Privileged maintenance hub')).toBeTruthy();
		expect(screen.getByRole('link', { name: /Accounting setup/i }).getAttribute('href')).toBe(
			'/app/admin/accounting'
		);
		expect(screen.getByRole('link', { name: /Party setup/i }).getAttribute('href')).toBe(
			'/app/admin/parties'
		);
		expect(screen.getByRole('link', { name: /Access controls/i }).getAttribute('href')).toBe(
			'/app/admin/access'
		);
		expect(screen.getByRole('link', { name: /Inventory setup/i }).getAttribute('href')).toBe(
			'/app/admin/inventory'
		);
	});
});
