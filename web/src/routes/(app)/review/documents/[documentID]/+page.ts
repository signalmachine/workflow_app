import type { PageLoad } from './$types';

import { getDocumentReview } from '$lib/api/review';

export const prerender = false;

export const load: PageLoad = async ({ fetch, params }) => {
	return {
		document: await getDocumentReview(params.documentID, fetch)
	};
};
