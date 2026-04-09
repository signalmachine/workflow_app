import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import InboundRequestDetailPage from './[requestLookup]/+page.svelte';

const baseData = {
	request: {
		request_id: 'request-1',
		request_reference: 'REQ-1001',
		session_id: 'session-1',
		actor_user_id: 'user-1',
		origin_type: 'human',
		channel: 'browser',
		status: 'processed',
		metadata: {},
		received_at: '2026-04-09T10:00:00Z',
		queued_at: '2026-04-09T10:01:00Z',
		processing_started_at: '2026-04-09T10:02:00Z',
		processed_at: '2026-04-09T10:03:00Z',
		completed_at: '2026-04-09T10:04:00Z',
		created_at: '2026-04-09T10:00:00Z',
		updated_at: '2026-04-09T10:04:00Z',
		message_count: 1,
		attachment_count: 0,
		last_run_id: 'run-1',
		last_run_status: 'completed'
	},
	messages: [],
	attachments: [],
	runs: [],
	steps: [],
	delegations: [],
	artifacts: [],
	recommendations: [],
	proposals: [
		{
			request_id: 'request-1',
			request_reference: 'REQ-1001',
			request_status: 'processed',
			recommendation_id: 'proposal-1',
			run_id: 'run-1',
			recommendation_type: 'proposal',
			recommendation_status: 'approved',
			summary: 'Prepare the posted accounting follow-through.',
			approval_id: 'approval-1',
			approval_status: 'approved',
			document_id: 'document-1',
			document_title: 'Submitted invoice',
			document_status: 'submitted',
			journal_entry_id: 'entry-1',
			journal_entry_number: 42,
			created_at: '2026-04-09T10:03:00Z'
		}
	]
};

describe('inbound request detail page', () => {
	it('prefers exact downstream accounting entry continuity when the latest proposal already has a posted entry', () => {
		render(InboundRequestDetailPage, { props: { data: baseData } as never });

		expect(screen.getByRole('link', { name: 'Open latest proposal' }).getAttribute('href')).toBe(
			'/app/review/proposals/proposal-1'
		);
		expect(screen.getByRole('link', { name: 'Open approval detail' }).getAttribute('href')).toBe(
			'/app/review/approvals/approval-1'
		);
		expect(screen.getByRole('link', { name: 'Open document detail' }).getAttribute('href')).toBe(
			'/app/review/documents/document-1'
		);
		expect(screen.getByRole('link', { name: 'Open accounting entry #42' }).getAttribute('href')).toBe(
			'/app/review/accounting/entry-1'
		);
		expect(screen.getByRole('link', { name: 'Accounting entry #42' }).getAttribute('href')).toBe(
			'/app/review/accounting/entry-1'
		);
	});
});
