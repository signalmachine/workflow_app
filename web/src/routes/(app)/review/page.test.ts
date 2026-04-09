import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import ReviewLandingPage from './+page.svelte';

const baseData = {
	snapshot: {
		inbound_summary: [],
		proposal_summary: [],
		pending_approvals: [
			{
				approval_id: 'approval-1',
				queue_code: 'ops',
				queue_status: 'pending',
				approval_status: 'pending',
				document_id: 'document-1',
				document_title: 'Repair work order',
				request_reference: 'REQ-1001',
				recommendation_id: 'proposal-1',
				recommendation_status: 'approval_requested',
				requested_at: '2026-04-09T12:00:00Z'
			}
		],
		inbound_request_count: 0,
		proposal_count: 0
	}
};

describe('review landing page', () => {
	it('uses exact approval detail links for pending approval rows', () => {
		render(ReviewLandingPage, { props: { data: baseData } as never });

		expect(screen.getByRole('link', { name: 'approval-1' }).getAttribute('href')).toBe(
			'/app/review/approvals/approval-1'
		);
	});
});
