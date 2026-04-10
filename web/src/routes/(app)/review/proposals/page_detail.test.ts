import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import ProposalDetailPage from './[recommendationID]/+page.svelte';

const baseData = {
	proposal: {
		recommendation_id: 'proposal-1',
		recommendation_status: 'approval_requested',
		request_reference: 'REQ-1001',
		request_status: 'processed',
		recommendation_type: 'proposal',
		suggested_queue_code: 'ops',
		created_at: '2026-04-09T10:03:00Z',
		summary: 'Prepare and approve the document.',
		approval_id: 'approval-1',
		document_id: 'document-1',
		journal_entry_id: 'entry-1',
		journal_entry_number: 912
	}
};

describe('proposal detail page', () => {
	it('prefers exact accounting-entry continuity over filtered accounting review links', () => {
		render(ProposalDetailPage, { props: { data: baseData } as never });

		expect(screen.getByRole('link', { name: 'Inbound request' }).getAttribute('href')).toBe(
			'/app/inbound-requests/REQ-1001'
		);
		expect(screen.getByRole('link', { name: 'Accounting entry #912' }).getAttribute('href')).toBe(
			'/app/review/accounting/entry-1'
		);
		expect(screen.queryByRole('link', { name: 'Accounting review' })).toBeNull();
	});
});
