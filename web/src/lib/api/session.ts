import { apiRequest } from '$lib/api/client';
import type { SessionContext, SessionLoginRequest } from '$lib/api/types';

export function getCurrentSession(fetcher: typeof fetch = fetch): Promise<SessionContext> {
	return apiRequest<SessionContext>('/api/session', undefined, fetcher);
}

export function login(payload: SessionLoginRequest): Promise<SessionContext> {
	return apiRequest<SessionContext>('/api/session/login', {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export function logout(): Promise<{ revoked: boolean }> {
	return apiRequest<{ revoked: boolean }>('/api/session/logout', {
		method: 'POST'
	});
}
