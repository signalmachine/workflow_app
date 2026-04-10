import type { PageLoad } from './$types';

import { getIncomeStatement } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const startOn = url.searchParams.get('start_on') ?? '';
	const endOn = url.searchParams.get('end_on') ?? '';
	const report = await getIncomeStatement({ startOn, endOn }, fetch);

	return {
		filters: { startOn, endOn },
		report
	};
};
