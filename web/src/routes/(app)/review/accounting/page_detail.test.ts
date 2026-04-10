import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AccountingDetailPage from './[entryID]/+page.svelte';

const baseData = {
	journal: {
		entry_id: 'entry-1',
		entry_number: 912,
		approval_status: 'approved',
		entry_kind: 'sales_invoice',
		effective_on: '2026-04-10',
		posted_at: '2026-04-10T09:10:00Z',
		total_debit_minor: 125000,
		total_credit_minor: 125000,
		summary: 'Posted invoice entry for the warehouse pump repair.',
		source_document_id: 'document-1',
		approval_id: 'approval-1',
		request_reference: 'REQ-1001',
		recommendation_id: 'proposal-1'
	}
};

describe('accounting detail page', () => {
	it('keeps posting detail traceable back through the full workflow chain', () => {
		render(AccountingDetailPage, { props: { data: baseData } as never });

		expect(screen.getByRole('link', { name: 'Document detail' }).getAttribute('href')).toBe(
			'/app/review/documents/document-1'
		);
		expect(screen.getByRole('link', { name: 'Approval detail' }).getAttribute('href')).toBe(
			'/app/review/approvals/approval-1'
		);
		expect(screen.getByRole('link', { name: 'Inbound request' }).getAttribute('href')).toBe(
			'/app/inbound-requests/REQ-1001'
		);
		expect(screen.getByRole('link', { name: 'Proposal detail' }).getAttribute('href')).toBe(
			'/app/review/proposals/proposal-1'
		);
		expect(screen.getByRole('link', { name: 'Filtered accounting view' }).getAttribute('href')).toBe(
			'/app/review/accounting/journal-entries?entry_id=entry-1'
		);
	});
});
