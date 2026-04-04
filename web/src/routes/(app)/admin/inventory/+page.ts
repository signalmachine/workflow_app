import type { PageLoad } from './$types';

import { listInventoryItems, listInventoryLocations } from '$lib/api/admin';
import { requireAdmin } from '$lib/utils/admin';

export const load: PageLoad = async ({ fetch, parent, url }) => {
	const { session } = await parent();
	requireAdmin(session);

	const itemRole = url.searchParams.get('item_role') ?? '';
	const locationRole = url.searchParams.get('location_role') ?? '';
	const [items, locations] = await Promise.all([
		listInventoryItems(itemRole, fetch),
		listInventoryLocations(locationRole, fetch)
	]);

	return {
		filters: { itemRole, locationRole },
		items: items.items,
		locations: locations.items
	};
};
