import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import JournalEntriesPage from './+page.svelte';

describe('journal entries page', () => {
	it('keeps exact journal drill-down on the dedicated report destination', () => {
		render(JournalEntriesPage, {
			props: {
				data: {
					filters: { startOn: '', endOn: '', entryID: '', documentID: 'document-1' },
					journals: [
						{
							entry_id: 'entry-1',
							entry_number: 912,
							entry_kind: 'sales_invoice',
							summary: 'Posted invoice entry.',
							approval_status: 'approved',
							posted_at: '2026-04-10T09:10:00Z',
							total_debit_minor: 125000
						}
					]
				}
			} as never
		});

		expect(screen.getByDisplayValue('document-1')).toBeTruthy();
		expect(screen.getByRole('link', { name: '912' }).getAttribute('href')).toBe('/app/review/accounting/entry-1');
	});
});
