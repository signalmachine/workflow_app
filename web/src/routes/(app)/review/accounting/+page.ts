import type { PageLoad } from './$types';

import { listControlAccountBalances, listJournalEntries, listTaxSummaries } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const startOn = url.searchParams.get('start_on') ?? '';
	const endOn = url.searchParams.get('end_on') ?? '';
	const asOf = url.searchParams.get('as_of') ?? '';
	const entryID = url.searchParams.get('entry_id') ?? '';
	const documentID = url.searchParams.get('document_id') ?? '';
	const taxType = url.searchParams.get('tax_type') ?? '';
	const taxCode = url.searchParams.get('tax_code') ?? '';
	const controlType = url.searchParams.get('control_type') ?? '';
	const accountID = url.searchParams.get('account_id') ?? '';

	const [journals, balances, taxes] = await Promise.all([
		listJournalEntries({ startOn, endOn, entryID, documentID, limit: 50 }, fetch),
		listControlAccountBalances({ asOf, controlType, accountID }, fetch),
		listTaxSummaries({ startOn, endOn, taxType, taxCode }, fetch)
	]);

	return {
		filters: { startOn, endOn, asOf, entryID, documentID, taxType, taxCode, controlType, accountID },
		journals: journals.items,
		balances: balances.items,
		taxes: taxes.items
	};
};
