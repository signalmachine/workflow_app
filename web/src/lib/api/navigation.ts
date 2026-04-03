import { apiRequest } from '$lib/api/client';
import type {
	AgentChatSnapshot,
	DashboardSnapshot,
	OperationsFeedSnapshot,
	OperationsSnapshot,
	ReviewLandingSnapshot,
	RouteCatalogSnapshot
} from '$lib/api/types';

export function getDashboardSnapshot(fetcher: typeof fetch = fetch): Promise<DashboardSnapshot> {
	return apiRequest<DashboardSnapshot>('/api/navigation/dashboard', undefined, fetcher);
}

export function getOperationsSnapshot(fetcher: typeof fetch = fetch): Promise<OperationsSnapshot> {
	return apiRequest<OperationsSnapshot>('/api/navigation/operations', undefined, fetcher);
}

export function getOperationsFeedSnapshot(fetcher: typeof fetch = fetch): Promise<OperationsFeedSnapshot> {
	return apiRequest<OperationsFeedSnapshot>('/api/navigation/operations-feed', undefined, fetcher);
}

export function getReviewLandingSnapshot(fetcher: typeof fetch = fetch): Promise<ReviewLandingSnapshot> {
	return apiRequest<ReviewLandingSnapshot>('/api/navigation/review', undefined, fetcher);
}

export function getAgentChatSnapshot(
	params?: { requestReference?: string; requestStatus?: string },
	fetcher: typeof fetch = fetch
): Promise<AgentChatSnapshot> {
	const search = new URLSearchParams();
	if (params?.requestReference) {
		search.set('request_reference', params.requestReference);
	}
	if (params?.requestStatus) {
		search.set('request_status', params.requestStatus);
	}
	const suffix = search.size > 0 ? `?${search.toString()}` : '';
	return apiRequest<AgentChatSnapshot>(`/api/navigation/agent-chat${suffix}`, undefined, fetcher);
}

export function getRouteCatalogSnapshot(query: string, fetcher: typeof fetch = fetch): Promise<RouteCatalogSnapshot> {
	const search = new URLSearchParams();
	if (query.trim() !== '') {
		search.set('q', query.trim());
	}
	const suffix = search.size > 0 ? `?${search.toString()}` : '';
	return apiRequest<RouteCatalogSnapshot>(`/api/navigation/routes${suffix}`, undefined, fetcher);
}
