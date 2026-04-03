import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

import { APIClientError } from '$lib/api/client';
import { getCurrentSession } from '$lib/api/session';
import { routes } from '$lib/utils/routes';

export const load: LayoutLoad = async ({ fetch, url }) => {
	try {
		const session = await getCurrentSession(fetch);
		return { session };
	} catch (error) {
		if (error instanceof APIClientError && error.status === 401) {
			const next = encodeURIComponent(url.pathname + url.search);
			throw redirect(307, `${routes.login}?next=${next}`);
		}

		throw error;
	}
};
