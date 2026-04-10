import type { PageLoad } from './$types';

import { getTrialBalance } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const asOf = url.searchParams.get('as_of') ?? '';
	const report = await getTrialBalance({ asOf }, fetch);

	return {
		filters: { asOf },
		report
	};
};
