import type { PageLoad } from './$types';

import { listInboundRequestStatusSummary, listInboundRequests } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const status = url.searchParams.get('status') ?? '';
	const requestReference = url.searchParams.get('request_reference') ?? '';
	const [summary, requests] = await Promise.all([
		listInboundRequestStatusSummary(fetch),
		listInboundRequests({ status, requestReference, limit: 50 }, fetch)
	]);

	return {
		filters: { status, requestReference },
		summary: summary.items,
		requests: requests.items
	};
};
