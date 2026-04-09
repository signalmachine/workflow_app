import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import InboundRequestDetailPage from './[requestLookup]/+page.svelte';

const baseData = {
	request: {
		request_id: 'request-1',
		request_reference: 'REQ-1001',
		status: 'processed',
		origin_type: 'human',
		channel: 'browser',
		submitter_label: 'Front desk',
		session_id: 'session-1',
		submitted_by_user_id: 'user-1',
		message_count: 1,
		attachment_count: 0,
		latest_run_status: 'completed',
		last_run_id: 'run-1',
		received_at: '2026-04-08T10:00:00Z',
		queued_at: '2026-04-08T10:01:00Z',
		processing_started_at: '2026-04-08T10:02:00Z',
		processed_at: '2026-04-08T10:03:00Z',
		completed_at: '2026-04-08T10:04:00Z',
		cancelled_at: '',
		failed_at: '',
		cancellation_reason: '',
		failure_reason: '',
		updated_at: '2026-04-08T10:04:00Z',
		metadata: {}
	},
	messages: [
		{
			message_id: 'message-1',
			message_index: 1,
			message_role: 'request',
			text_content: 'Inspect the warehouse pump issue.',
			attachment_count: 0,
			created_by_user_id: 'user-1',
			created_at: '2026-04-08T10:00:00Z'
		}
	],
	attachments: [],
	runs: [
		{
			run_id: 'run-1',
			agent_role: 'coordinator',
			capability_code: 'inbound_request.review',
			status: 'completed',
			summary: 'Review complete.',
			started_at: '2026-04-08T10:02:00Z',
			completed_at: '2026-04-08T10:03:00Z'
		}
	],
	steps: [],
	delegations: [],
	artifacts: [],
	recommendations: [],
	proposals: [
		{
			request_id: 'request-1',
			request_reference: 'REQ-1001',
			request_status: 'processed',
			recommendation_id: 'proposal-older',
			run_id: 'run-1',
			recommendation_type: 'proposal',
			recommendation_status: 'processed',
			summary: 'First proposal.',
			suggested_queue_code: 'ops',
			created_at: '2026-04-08T09:59:00Z'
		},
		{
			request_id: 'request-1',
			request_reference: 'REQ-1001',
			request_status: 'processed',
			recommendation_id: 'proposal-latest',
			run_id: 'run-1',
			recommendation_type: 'proposal',
			recommendation_status: 'approval_requested',
			summary: 'Latest proposal.',
			suggested_queue_code: 'ops',
			approval_id: 'approval-1',
			approval_status: 'pending',
			document_id: 'document-1',
			document_status: 'submitted',
			created_at: '2026-04-08T10:05:00Z'
		}
	]
};

describe('inbound request detail page', () => {
	it('keeps exact proposal, approval, and document continuity links visible near the top', () => {
		render(InboundRequestDetailPage, { props: { data: baseData } as never });

		expect(screen.getByText('Workflow continuity')).toBeTruthy();
		expect(screen.getByRole('link', { name: 'Open latest proposal' }).getAttribute('href')).toBe(
			'/app/review/proposals/proposal-latest'
		);
		expect(screen.getByRole('link', { name: 'Open approval detail' }).getAttribute('href')).toBe(
			'/app/review/approvals/approval-1'
		);
		expect(screen.getByRole('link', { name: 'Open document detail' }).getAttribute('href')).toBe(
			'/app/review/documents/document-1'
		);
		expect(screen.getByRole('link', { name: 'Open accounting review' }).getAttribute('href')).toBe(
			'/app/review/accounting?document_id=document-1'
		);
	});

	it('keeps proposal rows wired to exact downstream drill-down routes', () => {
		render(InboundRequestDetailPage, { props: { data: baseData } as never });

		expect(screen.getAllByRole('link', { name: 'Proposal detail' }).at(-1)?.getAttribute('href')).toBe(
			'/app/review/proposals/proposal-latest'
		);
		expect(screen.getAllByRole('link', { name: 'Approval detail' }).at(-1)?.getAttribute('href')).toBe(
			'/app/review/approvals/approval-1'
		);
		expect(screen.getAllByRole('link', { name: 'Document detail' }).at(-1)?.getAttribute('href')).toBe(
			'/app/review/documents/document-1'
		);
		expect(screen.getAllByRole('link', { name: 'Accounting review' }).at(-1)?.getAttribute('href')).toBe(
			'/app/review/accounting?document_id=document-1'
		);
	});
});
