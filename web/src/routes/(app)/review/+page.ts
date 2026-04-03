import type { PageLoad } from './$types';

import { getReviewLandingSnapshot } from '$lib/api/navigation';

export const load: PageLoad = async ({ fetch }) => {
	return {
		snapshot: await getReviewLandingSnapshot(fetch)
	};
};
