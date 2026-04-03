import type { PageLoad } from './$types';

import { getRouteCatalogSnapshot } from '$lib/api/navigation';

export const load: PageLoad = async ({ fetch, url }) => {
	const query = url.searchParams.get('q') ?? '';
	return {
		snapshot: await getRouteCatalogSnapshot(query, fetch)
	};
};
