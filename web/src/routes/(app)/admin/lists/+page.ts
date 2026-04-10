import type { PageLoad } from './$types';

import { requireAdmin } from '$lib/utils/admin';

export const load: PageLoad = async ({ parent }) => {
	const { session } = await parent();
	requireAdmin(session);
	return {};
};
