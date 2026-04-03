import type { PageLoad } from './$types';

import { getOperationsSnapshot } from '$lib/api/navigation';

export const load: PageLoad = async ({ fetch }) => {
	return {
		snapshot: await getOperationsSnapshot(fetch)
	};
};
