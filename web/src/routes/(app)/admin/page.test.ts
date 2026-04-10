import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AdminPage from './+page.svelte';

describe('admin hub page', () => {
	it('keeps privileged maintenance bounded to the promoted setup surfaces', () => {
		render(AdminPage, { props: {} as never });

		expect(screen.getByText('Privileged maintenance hub')).toBeTruthy();
		expect(screen.getByRole('link', { name: /Master data/i }).getAttribute('href')).toBe('/app/admin/master-data');
		expect(screen.getByRole('link', { name: /Lists/i }).getAttribute('href')).toBe('/app/admin/lists');
		expect(screen.getByRole('link', { name: /Access controls/i }).getAttribute('href')).toBe(
			'/app/admin/access'
		);
	});
});
