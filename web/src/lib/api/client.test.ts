import { afterEach, describe, expect, it, vi } from 'vitest';

import { APIClientError, apiRequest } from './client';

describe('apiRequest', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('returns parsed JSON for successful responses', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(async () => new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'Content-Type': 'application/json' } }))
		);

		const result = await apiRequest<{ ok: boolean }>('/api/session');

		expect(result.ok).toBe(true);
	});

	it('surfaces API error payloads', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(async () => new Response(JSON.stringify({ error: 'unauthorized' }), { status: 401, headers: { 'Content-Type': 'application/json' } }))
		);

		await expect(apiRequest('/api/session')).rejects.toEqual(new APIClientError(401, 'unauthorized'));
	});
});
