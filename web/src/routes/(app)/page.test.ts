import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import HomePage from './+page.svelte';

const baseData = {
	dashboard: {
		role_headline: 'Operator home',
		role_body: 'Keep request intake and exact workflow continuity close to the first click.',
		primary_actions: [],
		secondary_actions: [],
		inbound_summary: [{ status: 'queued', request_count: 2 }],
		proposal_summary: [{ recommendation_status: 'processed', proposal_count: 1 }],
		inbound_requests: [],
		proposals: [
			{
				request_id: 'request-1',
				request_reference: 'REQ-1001',
				request_status: 'processed',
				recommendation_id: 'proposal-1',
				run_id: 'run-1',
				recommendation_type: 'proposal',
				recommendation_status: 'processed',
				summary: 'Prepare the downstream document.',
				suggested_queue_code: 'ops',
				created_at: '2026-04-09T12:00:00Z'
			}
		],
		approvals: []
	}
};

describe('home page', () => {
	it('keeps recent proposal rows wired to exact proposal detail routes', () => {
		render(HomePage, { props: { data: baseData } as never });

		expect(screen.getByRole('link', { name: 'REQ-1001' }).getAttribute('href')).toBe(
			'/app/inbound-requests/REQ-1001'
		);
		expect(screen.getByRole('link', { name: 'proposal-1' }).getAttribute('href')).toBe(
			'/app/review/proposals/proposal-1'
		);
	});
});
