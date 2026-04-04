import type { PageLoad } from './$types';

import { getApprovalQueueDetail } from '$lib/api/review';

export const prerender = false;

export const load: PageLoad = async ({ fetch, params }) => {
	return {
		approval: await getApprovalQueueDetail(params.approvalID, fetch)
	};
};
