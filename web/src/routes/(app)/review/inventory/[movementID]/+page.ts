import type { PageLoad } from './$types';

import { getInventoryMovementDetail } from '$lib/api/review';

export const prerender = false;

export const load: PageLoad = async ({ fetch, params }) => {
	return {
		movement: await getInventoryMovementDetail(params.movementID, fetch)
	};
};
