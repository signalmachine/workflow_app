import { afterEach, describe, expect, it, vi } from 'vitest';

import {
	amendInboundRequest,
	cancelInboundRequest,
	deleteInboundDraft,
	queueInboundRequest,
	saveInboundDraft
} from './inbound';

describe('inbound api', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('targets the encoded draft lifecycle mutation endpoints', async () => {
		const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
			const url = String(input);
			expect(url.startsWith('/api/inbound-requests/req%2F1/')).toBe(true);
			return new Response(
				JSON.stringify({
					request_id: 'req/1',
					request_reference: 'REQ-1',
					status: 'draft',
					received_at: '2026-04-04T00:00:00Z',
					created_at: '2026-04-04T00:00:00Z',
					updated_at: '2026-04-04T00:00:00Z'
				}),
				{ status: 200, headers: { 'Content-Type': 'application/json' } }
			);
		});

		await saveInboundDraft(
			'req/1',
			{
				message_id: 'msg-1',
				origin_type: 'human',
				channel: 'browser',
				metadata: {},
				message: { message_role: 'request', text_content: 'updated' },
				attachments: []
			},
			fetchMock as typeof fetch
		);
		expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/inbound-requests/req%2F1/draft');
		expect(fetchMock.mock.calls[0]?.[1]?.method).toBe('PUT');

		await queueInboundRequest('req/1', fetchMock as typeof fetch);
		expect(fetchMock.mock.calls[1]?.[0]).toBe('/api/inbound-requests/req%2F1/queue');
		expect(fetchMock.mock.calls[1]?.[1]?.method).toBe('POST');

		await cancelInboundRequest('req/1', 'duplicate', fetchMock as typeof fetch);
		expect(fetchMock.mock.calls[2]?.[0]).toBe('/api/inbound-requests/req%2F1/cancel');
		expect(fetchMock.mock.calls[2]?.[1]?.method).toBe('POST');

		await amendInboundRequest('req/1', fetchMock as typeof fetch);
		expect(fetchMock.mock.calls[3]?.[0]).toBe('/api/inbound-requests/req%2F1/amend');
		expect(fetchMock.mock.calls[3]?.[1]?.method).toBe('POST');
	});

	it('deletes drafts through the shared inbound-request seam', async () => {
		const fetchMock = vi.fn(async () => new Response(JSON.stringify({ deleted: true }), {
			status: 200,
			headers: { 'Content-Type': 'application/json' }
		}));

		const result = await deleteInboundDraft('req/1', fetchMock as typeof fetch);

		expect(fetchMock).toHaveBeenCalledWith(
			'/api/inbound-requests/req%2F1/delete',
			expect.objectContaining({ method: 'DELETE' })
		);
		expect(result.deleted).toBe(true);
	});
});
