import type { PageLoad } from './$types';

import { listOrgUsers } from '$lib/api/admin';
import { requireAdmin } from '$lib/utils/admin';

export const load: PageLoad = async ({ fetch, parent }) => {
	const { session } = await parent();
	requireAdmin(session);

	const users = await listOrgUsers(fetch);
	return { users: users.items };
};
