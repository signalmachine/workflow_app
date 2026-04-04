import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

import { APIClientError } from '$lib/api/client';
import { getInboundRequestDetail } from '$lib/api/review';

export const prerender = false;

export const load: PageLoad = async ({ fetch, params }) => {
	try {
		return await getInboundRequestDetail(params.requestLookup, fetch);
	} catch (cause) {
		if (cause instanceof APIClientError && cause.status === 404) {
			throw error(404, 'Inbound request not found');
		}
		throw cause;
	}
};
