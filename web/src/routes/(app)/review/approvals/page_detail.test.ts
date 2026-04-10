import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import ApprovalDetailPage from './[approvalID]/+page.svelte';

const baseData = {
	approval: {
		approval_id: 'approval-1',
		approval_status: 'approved',
		queue_code: 'finance-review',
		queue_status: 'completed',
		requested_at: '2026-04-10T09:00:00Z',
		decided_at: '2026-04-10T09:10:00Z',
		document_id: 'document-1',
		document_title: 'Warehouse pump repair invoice',
		request_reference: 'REQ-1001',
		recommendation_id: 'proposal-1',
		journal_entry_id: 'entry-1'
	}
};

describe('approval detail page', () => {
	it('keeps exact request, proposal, document, and accounting continuity together', () => {
		render(ApprovalDetailPage, { props: { data: baseData } as never });

		expect(screen.getByRole('link', { name: 'Inbound request' }).getAttribute('href')).toBe(
			'/app/inbound-requests/REQ-1001'
		);
		expect(screen.getByRole('link', { name: 'Proposal detail' }).getAttribute('href')).toBe(
			'/app/review/proposals/proposal-1'
		);
		expect(screen.getByRole('link', { name: 'Document detail' }).getAttribute('href')).toBe(
			'/app/review/documents/document-1'
		);
		expect(screen.getByRole('link', { name: 'Accounting detail' }).getAttribute('href')).toBe(
			'/app/review/accounting/entry-1'
		);
	});
});
