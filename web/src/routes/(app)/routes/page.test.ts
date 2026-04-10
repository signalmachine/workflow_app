import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import RouteCatalogPage from './+page.svelte';

const baseData = {
	snapshot: {
		query: 'pending approvals',
		items: [
			{
				title: 'Approval review',
				href: '/app/review/approvals',
				category: 'Review',
				summary: 'Pending governed decisions on downstream document truth.'
			}
		]
	}
};

describe('route catalog page', () => {
	it('keeps multi-term operator-intent searches wired to the exact route family', () => {
		render(RouteCatalogPage, { props: { data: baseData } as never });

		expect(screen.getByDisplayValue('pending approvals')).toBeTruthy();
		expect(
			screen
				.getByRole('link', { name: /Approval review/ })
				.getAttribute('href')
		).toBe('/app/review/approvals');
	});
});
