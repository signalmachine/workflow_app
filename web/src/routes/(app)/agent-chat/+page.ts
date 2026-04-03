import type { PageLoad } from './$types';

import { getAgentChatSnapshot } from '$lib/api/navigation';

export const load: PageLoad = async ({ fetch, url }) => {
	return {
		snapshot: await getAgentChatSnapshot(
			{
				requestReference: url.searchParams.get('request_reference') ?? undefined,
				requestStatus: url.searchParams.get('request_status') ?? undefined
			},
			fetch
		)
	};
};
