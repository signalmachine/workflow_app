import type { PageLoad } from './$types';

import { getAuditEventDetail } from '$lib/api/review';

export const prerender = false;

export const load: PageLoad = async ({ fetch, params }) => {
	return {
		event: await getAuditEventDetail(params.eventID, fetch)
	};
};
