import type { PageLoad } from './$types';

import { getProcessedProposalDetail } from '$lib/api/review';

export const prerender = false;

export const load: PageLoad = async ({ fetch, params }) => {
	return {
		proposal: await getProcessedProposalDetail(params.recommendationID, fetch)
	};
};
