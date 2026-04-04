import type { PageLoad } from './$types';

import { listParties } from '$lib/api/admin';
import { requireAdmin } from '$lib/utils/admin';

export const load: PageLoad = async ({ fetch, parent, url }) => {
	const { session } = await parent();
	requireAdmin(session);

	const partyKind = url.searchParams.get('party_kind') ?? '';
	const parties = await listParties(partyKind, fetch);

	return {
		filters: { partyKind },
		parties: parties.items
	};
};
