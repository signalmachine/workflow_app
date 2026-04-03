// @ts-nocheck
import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

import { APIClientError } from '$lib/api/client';
import { getCurrentSession } from '$lib/api/session';
import { routes } from '$lib/utils/routes';

export const load = async ({ fetch }: Parameters<PageLoad>[0]) => {
	try {
		await getCurrentSession(fetch);
		throw redirect(307, routes.home);
	} catch (error) {
		if (error instanceof APIClientError && error.status === 401) {
			return {};
		}
		throw error;
	}
};
