import type { PageLoad } from './$types';

import { getOperationsFeedSnapshot } from '$lib/api/navigation';

export const load: PageLoad = async ({ fetch }) => {
	return {
		snapshot: await getOperationsFeedSnapshot(fetch)
	};
};
