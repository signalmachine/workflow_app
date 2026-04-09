import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import ProposalDetailPage from './[recommendationID]/+page.svelte';

const baseData = {
	proposal: {
		request_id: 'request-1',
		request_reference: 'REQ-1001',
		request_status: 'processed',
		recommendation_id: 'proposal-1',
		run_id: 'run-1',
		recommendation_type: 'proposal',
		recommendation_status: 'approval_requested',
		summary: 'Review and request approval.',
		suggested_queue_code: 'finance-review',
		approval_id: 'approval-1',
		approval_status: 'pending',
		document_id: 'document-1',
		document_type_code: 'invoice',
		document_title: 'Submitted invoice',
		document_number: 'INV-1001',
		document_status: 'submitted',
		created_at: '2026-04-09T12:00:00Z'
	}
};

describe('proposal detail page', () => {
	it('keeps downstream accounting review continuity visible when a document already exists', () => {
		render(ProposalDetailPage, { props: { data: baseData } as never });

		expect(screen.getByRole('link', { name: 'Inbound request' }).getAttribute('href')).toBe(
			'/app/inbound-requests/REQ-1001'
		);
		expect(screen.getByRole('link', { name: 'Approval detail' }).getAttribute('href')).toBe(
			'/app/review/approvals/approval-1'
		);
		expect(screen.getByRole('link', { name: 'Document detail' }).getAttribute('href')).toBe(
			'/app/review/documents/document-1'
		);
		expect(screen.getByRole('link', { name: 'Accounting review' }).getAttribute('href')).toBe(
			'/app/review/accounting?document_id=document-1'
		);
	});
});
