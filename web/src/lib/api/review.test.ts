import { afterEach, describe, expect, it, vi } from 'vitest';

import { APIClientError } from './client';
import { getInboundRequestDetail } from './review';

describe('review api', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('requests inbound request detail by encoded lookup', async () => {
		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			expect(String(input)).toBe('/api/review/inbound-requests/run%3Arun-123');
			return new Response(
				JSON.stringify({
					request: {
						request_id: 'req-1',
						request_reference: 'REQ-1',
						origin_type: 'human',
						channel: 'browser',
						status: 'processed',
						metadata: {},
						received_at: '2026-04-04T00:00:00Z',
						created_at: '2026-04-04T00:00:00Z',
						updated_at: '2026-04-04T00:00:00Z',
						message_count: 1,
						attachment_count: 0
					},
					messages: [],
					attachments: [],
					runs: [],
					steps: [],
					delegations: [],
					artifacts: [],
					recommendations: [],
					proposals: []
				}),
				{ status: 200, headers: { 'Content-Type': 'application/json' } }
			);
		});

		const detail = await getInboundRequestDetail('run:run-123', fetchMock as typeof fetch);

		expect(detail.request.request_reference).toBe('REQ-1');
		expect(fetchMock).toHaveBeenCalledTimes(1);
	});

	it('surfaces not found errors for missing detail routes', async () => {
		const fetchMock = vi.fn(async () => new Response(JSON.stringify({ error: 'record not found' }), { status: 404, headers: { 'Content-Type': 'application/json' } }));

		await expect(getInboundRequestDetail('REQ-404', fetchMock as typeof fetch)).rejects.toEqual(new APIClientError(404, 'record not found'));
	});
});
