import type { PageLoad } from './$types';

import { listJournalEntries } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const startOn = url.searchParams.get('start_on') ?? '';
	const endOn = url.searchParams.get('end_on') ?? '';
	const entryID = url.searchParams.get('entry_id') ?? '';
	const documentID = url.searchParams.get('document_id') ?? '';

	const journals = await listJournalEntries({ startOn, endOn, entryID, documentID, limit: 50 }, fetch);

	return {
		filters: { startOn, endOn, entryID, documentID },
		journals: journals.items
	};
};
