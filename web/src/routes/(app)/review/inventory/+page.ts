import type { PageLoad } from './$types';

import { listInventoryMovements, listInventoryReconciliation, listInventoryStock } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const movementID = url.searchParams.get('movement_id') ?? '';
	const itemID = url.searchParams.get('item_id') ?? '';
	const locationID = url.searchParams.get('location_id') ?? '';
	const documentID = url.searchParams.get('document_id') ?? '';
	const movementType = url.searchParams.get('movement_type') ?? '';
	const onlyPendingAccounting = url.searchParams.get('only_pending_accounting') === 'true';
	const onlyPendingExecution = url.searchParams.get('only_pending_execution') === 'true';

	const [stock, movements, reconciliation] = await Promise.all([
		listInventoryStock({ itemID, locationID, limit: 50 }, fetch),
		listInventoryMovements({ movementID, itemID, locationID, documentID, movementType, limit: 50 }, fetch),
		listInventoryReconciliation({ documentID, onlyPendingAccounting, onlyPendingExecution, limit: 50 }, fetch)
	]);

	return {
		filters: { movementID, itemID, locationID, documentID, movementType, onlyPendingAccounting, onlyPendingExecution },
		stock: stock.items,
		movements: movements.items,
		reconciliation: reconciliation.items
	};
};
