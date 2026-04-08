import type { PageLoad } from './$types';

import { listInventoryMovements, listInventoryReconciliation, listInventoryStock } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, parent }) => {
	const { session } = await parent();

	const [stock, movements, pendingExecution, pendingAccounting] = await Promise.all([
		listInventoryStock({ limit: 8 }, fetch),
		listInventoryMovements({ limit: 8 }, fetch),
		listInventoryReconciliation({ onlyPendingExecution: true, limit: 8 }, fetch),
		listInventoryReconciliation({ onlyPendingAccounting: true, limit: 8 }, fetch)
	]);

	return {
		roleCode: session.role_code,
		stock: stock.items,
		movements: movements.items,
		pendingExecution: pendingExecution.items,
		pendingAccounting: pendingAccounting.items
	};
};
