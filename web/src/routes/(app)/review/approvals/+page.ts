import type { PageLoad } from './$types';

import { listApprovalQueue } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const status = url.searchParams.get('status') ?? '';
	const queueCode = url.searchParams.get('queue_code') ?? '';
	const approvalID = url.searchParams.get('approval_id') ?? '';
	const approvals = await listApprovalQueue({ status, queueCode, approvalID, limit: 50 }, fetch);

	return {
		filters: { status, queueCode, approvalID },
		approvals: approvals.items
	};
};
