import type { PageLoad } from './$types';

import { getDashboardSnapshot } from '$lib/api/navigation';

export const load: PageLoad = async ({ fetch }) => {
	return {
		dashboard: await getDashboardSnapshot(fetch)
	};
};
