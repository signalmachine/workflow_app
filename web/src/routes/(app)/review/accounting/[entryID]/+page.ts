import type { PageLoad } from './$types';

import { getJournalEntryDetail } from '$lib/api/review';

export const prerender = false;

export const load: PageLoad = async ({ fetch, params }) => {
	return {
		journal: await getJournalEntryDetail(params.entryID, fetch)
	};
};
