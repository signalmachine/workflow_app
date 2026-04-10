import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import DocumentDetailPage from './[documentID]/+page.svelte';

const baseData = {
	document: {
		document_id: 'document-1',
		title: 'Warehouse pump repair invoice',
		status: 'approved',
		type_code: 'invoice',
		number_value: 'INV-1001',
		created_at: '2026-04-10T09:00:00Z',
		submitted_at: '2026-04-10T09:05:00Z',
		approved_at: '2026-04-10T09:10:00Z',
		request_reference: 'REQ-1001',
		recommendation_id: 'proposal-1',
		approval_id: 'approval-1',
		journal_entry_id: 'entry-1'
	}
};

describe('document detail page', () => {
	it('keeps upstream request and proposal provenance visible beside posting continuity', () => {
		render(DocumentDetailPage, { props: { data: baseData } as never });

		expect(screen.getByRole('link', { name: 'Inbound request' }).getAttribute('href')).toBe(
			'/app/inbound-requests/REQ-1001'
		);
		expect(screen.getByRole('link', { name: 'Proposal detail' }).getAttribute('href')).toBe(
			'/app/review/proposals/proposal-1'
		);
		expect(screen.getByRole('link', { name: 'Approval detail' }).getAttribute('href')).toBe(
			'/app/review/approvals/approval-1'
		);
		expect(screen.getByRole('link', { name: 'Accounting detail' }).getAttribute('href')).toBe(
			'/app/review/accounting/entry-1'
		);
	});
});
