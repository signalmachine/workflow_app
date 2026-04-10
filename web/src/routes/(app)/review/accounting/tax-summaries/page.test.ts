import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import TaxSummariesPage from './+page.svelte';

describe('tax summaries page', () => {
	it('keeps tax summaries on a dedicated report destination', () => {
		render(TaxSummariesPage, {
			props: {
				data: {
					filters: { startOn: '', endOn: '', taxType: 'gst', taxCode: 'GST' },
					taxes: [
						{
							tax_type: 'gst',
							tax_code: 'GST',
							tax_name: 'Goods and Services Tax',
							document_count: 2,
							net_minor: 25000,
							last_effective_on: '2026-04-10'
						}
					]
				}
			} as never
		});

		expect(screen.getByDisplayValue('GST')).toBeTruthy();
		expect(screen.getByText('GST · Goods and Services Tax')).toBeTruthy();
	});
});
