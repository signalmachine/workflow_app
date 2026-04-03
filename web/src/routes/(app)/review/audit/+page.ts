import type { PageLoad } from './$types';

import { listAuditEvents } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const eventID = url.searchParams.get('event_id') ?? '';
	const entityType = url.searchParams.get('entity_type') ?? '';
	const entityID = url.searchParams.get('entity_id') ?? '';
	const events = await listAuditEvents({ eventID, entityType, entityID, limit: 50 }, fetch);

	return {
		filters: { eventID, entityType, entityID },
		events: events.items
	};
};
