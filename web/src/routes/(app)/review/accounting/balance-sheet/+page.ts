import type { PageLoad } from './$types';

import { getBalanceSheet } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const asOf = url.searchParams.get('as_of') ?? '';
	const report = await getBalanceSheet({ asOf }, fetch);

	return {
		filters: { asOf },
		report
	};
};
