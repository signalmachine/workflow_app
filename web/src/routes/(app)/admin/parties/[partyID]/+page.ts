import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

import { APIClientError } from '$lib/api/client';
import { getPartyDetail } from '$lib/api/admin';
import { requireAdmin } from '$lib/utils/admin';

export const prerender = false;

export const load: PageLoad = async ({ fetch, params, parent }) => {
	const { session } = await parent();
	requireAdmin(session);

	try {
		const detail = await getPartyDetail(params.partyID, fetch);
		return detail;
	} catch (cause) {
		if (cause instanceof APIClientError && cause.status === 404) {
			throw error(404, 'Party not found');
		}
		throw cause;
	}
};
