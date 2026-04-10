import { cleanup, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';

import InboundRequestDetailPage from './[requestLookup]/+page.svelte';

afterEach(() => {
	cleanup();
});

const baseData = {
	request: {
		request_id: 'req-1',
		request_reference: 'REQ-1001',
		status: 'processed',
		channel: 'browser',
		origin_type: 'operator',
		message_count: 1,
		attachment_count: 0,
		updated_at: '2026-04-09T12:00:00Z',
		last_run_id: 'run-1',
		metadata: {},
		received_at: '2026-04-09T10:00:00Z',
		queued_at: '2026-04-09T10:01:00Z',
		processing_started_at: '2026-04-09T10:02:00Z',
		processed_at: '2026-04-09T10:03:00Z',
		completed_at: '2026-04-09T10:04:00Z',
		cancelled_at: null,
		failed_at: null,
		cancellation_reason: null,
		failure_reason: null
	},
	messages: [
		{
			message_id: 'msg-1',
			message_role: 'user',
			message_index: 1,
			text_content: 'Please prepare the downstream document.',
			attachment_count: 0,
			created_by_user_id: 'user-1',
			created_at: '2026-04-09T10:00:00Z'
		}
	],
	attachments: [],
	runs: [
		{
			run_id: 'run-1',
			agent_role: 'coordinator',
			capability_code: 'proposal',
			status: 'completed',
			summary: 'Prepared the proposal.',
			started_at: '2026-04-09T10:02:00Z',
			completed_at: '2026-04-09T10:03:00Z'
		}
	],
	recommendations: [],
	proposals: [
		{
			recommendation_id: 'proposal-1',
			recommendation_status: 'approval_requested',
			approval_id: 'approval-1',
			approval_status: 'pending',
			document_id: 'document-1',
			document_status: 'submitted',
			journal_entry_id: 'entry-1',
			journal_entry_number: 912,
			summary: 'Prepare and approve the document.',
			created_at: '2026-04-09T10:03:00Z'
		}
	],
	steps: [],
	artifacts: []
};

describe('inbound request detail page', () => {
	it('shows draft lifecycle controls when the request is still editable', () => {
		render(InboundRequestDetailPage, {
			props: {
				data: {
					...baseData,
					request: {
						...baseData.request,
						status: 'draft'
					}
				}
			} as never
		});

		expect(screen.getByRole('button', { name: 'Save draft changes' })).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Queue updated draft' })).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Delete draft' })).toBeTruthy();
	});

	it('shows queued-request cancellation and amendment controls', () => {
		render(InboundRequestDetailPage, {
			props: {
				data: {
					...baseData,
					request: {
						...baseData.request,
						status: 'queued'
					}
				}
			} as never
		});

		expect(screen.getByRole('button', { name: 'Cancel queued request' })).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Amend back to draft' })).toBeTruthy();
	});

	it('prefers exact accounting-entry drill-down when the journal entry is already known', () => {
		render(InboundRequestDetailPage, { props: { data: baseData } as never });

		expect(screen.getAllByRole('link', { name: 'Open latest proposal' }).at(-1)?.getAttribute('href')).toBe('/app/review/proposals/proposal-1');
		expect(
			screen.getByRole('link', { name: 'Open accounting entry #912' }).getAttribute('href')
		).toBe('/app/review/accounting/entry-1');
		expect(screen.queryByRole('link', { name: 'Open accounting review' })).toBeNull();
	});
});
