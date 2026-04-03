import type { PageLoad } from './$types';

import { listProcessedProposalStatusSummary, listProcessedProposals } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const status = url.searchParams.get('status') ?? '';
	const requestReference = url.searchParams.get('request_reference') ?? '';
	const [summary, proposals] = await Promise.all([
		listProcessedProposalStatusSummary(fetch),
		listProcessedProposals({ status, requestReference, limit: 50 }, fetch)
	]);

	return {
		filters: { status, requestReference },
		summary: summary.items,
		proposals: proposals.items
	};
};
