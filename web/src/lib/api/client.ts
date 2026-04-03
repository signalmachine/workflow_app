import type { APIError } from '$lib/api/types';

export class APIClientError extends Error {
	status: number;

	constructor(status: number, message: string) {
		super(message);
		this.name = 'APIClientError';
		this.status = status;
	}
}

async function parseResponse<T>(response: Response): Promise<T> {
	if (response.ok) {
		return (await response.json()) as T;
	}

	let message = `Request failed with status ${response.status}`;
	try {
		const payload = (await response.json()) as APIError;
		if (payload?.error) {
			message = payload.error;
		}
	} catch {
		// Ignore JSON parse failures on non-JSON error responses.
	}

	throw new APIClientError(response.status, message);
}

export async function apiRequest<T>(input: RequestInfo | URL, init?: RequestInit, fetcher: typeof fetch = fetch): Promise<T> {
	const response = await fetcher(input, {
		credentials: 'same-origin',
		headers: {
			Accept: 'application/json',
			...(init?.body ? { 'Content-Type': 'application/json' } : {}),
			...(init?.headers ?? {})
		},
		...init
	});

	return parseResponse<T>(response);
}
