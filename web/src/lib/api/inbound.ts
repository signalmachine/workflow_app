import { apiRequest } from '$lib/api/client';
import type { ProcessNextQueuedResponse, SubmitInboundRequestPayload, SubmitInboundRequestResponse } from '$lib/api/types';

export function submitInboundRequest(
	payload: SubmitInboundRequestPayload,
	fetcher: typeof fetch = fetch
): Promise<SubmitInboundRequestResponse> {
	return apiRequest<SubmitInboundRequestResponse>(
		'/api/inbound-requests',
		{
			method: 'POST',
			body: JSON.stringify(payload)
		},
		fetcher
	);
}

export function processNextQueuedInboundRequest(fetcher: typeof fetch = fetch): Promise<ProcessNextQueuedResponse> {
	return apiRequest<ProcessNextQueuedResponse>(
		'/api/agent/process-next-queued-inbound-request',
		{
			method: 'POST',
			body: JSON.stringify({ channel: 'browser' })
		},
		fetcher
	);
}
