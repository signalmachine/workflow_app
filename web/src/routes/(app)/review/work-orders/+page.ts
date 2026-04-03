import type { PageLoad } from './$types';

import { listWorkOrders } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const status = url.searchParams.get('status') ?? '';
	const workOrderID = url.searchParams.get('work_order_id') ?? '';
	const documentID = url.searchParams.get('document_id') ?? '';
	const workOrders = await listWorkOrders({ status, workOrderID, documentID, limit: 50 }, fetch);

	return {
		filters: { status, workOrderID, documentID },
		workOrders: workOrders.items
	};
};
