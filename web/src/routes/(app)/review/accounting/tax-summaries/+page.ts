import type { PageLoad } from './$types';

import { listTaxSummaries } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const startOn = url.searchParams.get('start_on') ?? '';
	const endOn = url.searchParams.get('end_on') ?? '';
	const taxType = url.searchParams.get('tax_type') ?? '';
	const taxCode = url.searchParams.get('tax_code') ?? '';

	const taxes = await listTaxSummaries({ startOn, endOn, taxType, taxCode }, fetch);

	return {
		filters: { startOn, endOn, taxType, taxCode },
		taxes: taxes.items
	};
};
