import type { PageLoad } from './$types';

import { listDocuments } from '$lib/api/review';

export const load: PageLoad = async ({ fetch, url }) => {
	const status = url.searchParams.get('status') ?? '';
	const typeCode = url.searchParams.get('type_code') ?? '';
	const documentID = url.searchParams.get('document_id') ?? '';
	const documents = await listDocuments({ status, typeCode, documentID, limit: 50 }, fetch);

	return {
		filters: { status, typeCode, documentID },
		documents: documents.items
	};
};
