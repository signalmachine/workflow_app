import type { PageLoad } from './$types';

import { listAccountingPeriods, listLedgerAccounts, listTaxCodes } from '$lib/api/admin';
import { requireAdmin } from '$lib/utils/admin';

export const load: PageLoad = async ({ fetch, parent }) => {
	const { session } = await parent();
	requireAdmin(session);

	const [ledgerAccounts, taxCodes, periods] = await Promise.all([
		listLedgerAccounts(fetch),
		listTaxCodes(fetch),
		listAccountingPeriods(fetch)
	]);

	return {
		ledgerAccounts: ledgerAccounts.items,
		taxCodes: taxCodes.items,
		periods: periods.items
	};
};
