import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import OperationsPage from './+page.svelte';

const baseData = {
	snapshot: {
		queued_request_count: 4,
		pending_approval_count: 2,
		proposal_review_count: 3,
		recent_feed: [
			{
				kind: 'request',
				title: 'REQ-1001 queued',
				status: 'queued',
				summary: 'Request is ready for coordinator pickup.',
				primary_href: '/app/inbound-requests/REQ-1001',
				primary_label: 'Open request',
				secondary_href: '/app/review/proposals/proposal-1',
				secondary_label: 'Open proposal',
				occurred_at: '2026-04-10T09:00:00Z'
			}
		]
	}
};

describe('operations page', () => {
	it('keeps queue actions and feed continuity ahead of decorative content', () => {
		render(OperationsPage, { props: { data: baseData } as never });

		expect(screen.getByText('Operations landing')).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Process next queued request' })).toBeTruthy();
		expect(screen.getByRole('link', { name: 'Start a new request' }).getAttribute('href')).toBe(
			'/app/submit-inbound-request'
		);
		expect(screen.getByRole('link', { name: 'Open request' }).getAttribute('href')).toBe(
			'/app/inbound-requests/REQ-1001'
		);
	});
});
