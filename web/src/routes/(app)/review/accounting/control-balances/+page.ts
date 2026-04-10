import type { PageLoad } from './$types';

import { listControlAccountBalances } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const asOf = url.searchParams.get('as_of') ?? '';
	const controlType = url.searchParams.get('control_type') ?? '';
	const accountID = url.searchParams.get('account_id') ?? '';

	const balances = await listControlAccountBalances({ asOf, controlType, accountID }, fetch);

	return {
		filters: { asOf, controlType, accountID },
		balances: balances.items
	};
};
