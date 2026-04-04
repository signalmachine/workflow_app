import type { PageLoad } from './$types';

import { getWorkOrderReview } from '$lib/api/review';

export const prerender = false;

export const load: PageLoad = async ({ fetch, params }) => {
	return {
		workOrder: await getWorkOrderReview(params.workOrderID, fetch)
	};
};
