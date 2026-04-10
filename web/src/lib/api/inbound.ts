import { apiRequest } from '$lib/api/client';
import type {
	DeleteInboundDraftResponse,
	ProcessNextQueuedResponse,
	SaveInboundDraftPayload,
	SubmitInboundRequestPayload,
	SubmitInboundRequestResponse
} from '$lib/api/types';

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

export function saveInboundDraft(
	requestID: string,
	payload: SaveInboundDraftPayload,
	fetcher: typeof fetch = fetch
): Promise<SubmitInboundRequestResponse> {
	return apiRequest<SubmitInboundRequestResponse>(
		`/api/inbound-requests/${encodeURIComponent(requestID)}/draft`,
		{
			method: 'PUT',
			body: JSON.stringify(payload)
		},
		fetcher
	);
}

export function queueInboundRequest(requestID: string, fetcher: typeof fetch = fetch): Promise<SubmitInboundRequestResponse> {
	return apiRequest<SubmitInboundRequestResponse>(
		`/api/inbound-requests/${encodeURIComponent(requestID)}/queue`,
		{
			method: 'POST'
		},
		fetcher
	);
}

export function cancelInboundRequest(
	requestID: string,
	reason: string,
	fetcher: typeof fetch = fetch
): Promise<SubmitInboundRequestResponse> {
	return apiRequest<SubmitInboundRequestResponse>(
		`/api/inbound-requests/${encodeURIComponent(requestID)}/cancel`,
		{
			method: 'POST',
			body: JSON.stringify({ reason })
		},
		fetcher
	);
}

export function amendInboundRequest(requestID: string, fetcher: typeof fetch = fetch): Promise<SubmitInboundRequestResponse> {
	return apiRequest<SubmitInboundRequestResponse>(
		`/api/inbound-requests/${encodeURIComponent(requestID)}/amend`,
		{
			method: 'POST'
		},
		fetcher
	);
}

export function deleteInboundDraft(requestID: string, fetcher: typeof fetch = fetch): Promise<DeleteInboundDraftResponse> {
	return apiRequest<DeleteInboundDraftResponse>(
		`/api/inbound-requests/${encodeURIComponent(requestID)}/delete`,
		{
			method: 'DELETE'
		},
		fetcher
	);
}
