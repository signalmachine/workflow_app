import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import AgentChatPage from './+page.svelte';

const baseData = {
	snapshot: {
		request_reference: 'REQ-1001',
		request_status: 'queued',
		recent_requests: [],
		recent_proposals: [
			{
				request_id: 'request-1',
				request_reference: 'REQ-1001',
				request_status: 'processed',
				recommendation_id: 'proposal-1',
				run_id: 'run-1',
				recommendation_type: 'proposal',
				recommendation_status: 'approval_requested',
				summary: 'Review and request approval.',
				suggested_queue_code: 'ops',
				created_at: '2026-04-09T12:00:00Z'
			}
		]
	}
};

describe('agent chat page', () => {
	it('uses exact drill-down links for recent coordinator proposals', () => {
		render(AgentChatPage, { props: { data: baseData } as never });

		expect(screen.getAllByRole('link', { name: 'REQ-1001' }).at(-1)?.getAttribute('href')).toBe(
			'/app/inbound-requests/REQ-1001'
		);
		expect(screen.getByRole('link', { name: 'proposal-1' }).getAttribute('href')).toBe(
			'/app/review/proposals/proposal-1'
		);
	});
});
