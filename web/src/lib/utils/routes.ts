import { base } from '$app/paths';

function withBase(path: string): string {
	if (path === '/') {
		return base || '/';
	}
	return `${base}${path}`;
}

export const routes = {
	login: withBase('/login'),
	home: withBase('/'),
	inboundRequests: withBase('/inbound-requests'),
	routeCatalog: withBase('/routes'),
	submitInboundRequest: withBase('/submit-inbound-request'),
	operations: withBase('/operations'),
	operationsFeed: withBase('/operations-feed'),
	agentChat: withBase('/agent-chat'),
	review: withBase('/review'),
	reviewInboundRequests: withBase('/review/inbound-requests'),
	reviewProposals: withBase('/review/proposals'),
	reviewApprovals: withBase('/review/approvals'),
	reviewDocuments: withBase('/review/documents'),
	reviewAccounting: withBase('/review/accounting'),
	reviewInventory: withBase('/review/inventory'),
	reviewWorkOrders: withBase('/review/work-orders'),
	reviewAudit: withBase('/review/audit'),
	inventory: withBase('/inventory'),
	settings: withBase('/settings'),
	admin: withBase('/admin'),
	adminMasterData: withBase('/admin/master-data'),
	adminLists: withBase('/admin/lists'),
	adminAccounting: withBase('/admin/accounting'),
	adminParties: withBase('/admin/parties'),
	adminAccess: withBase('/admin/access'),
	adminInventory: withBase('/admin/inventory'),
	reviewAccountingJournalEntries: withBase('/review/accounting/journal-entries'),
	reviewAccountingControlBalances: withBase('/review/accounting/control-balances'),
	reviewAccountingTaxSummaries: withBase('/review/accounting/tax-summaries')
};

export function adminPartyDetail(partyID: string): string {
	return withBase(`/admin/parties/${partyID}`);
}

export function inboundRequestDetail(requestLookup: string): string {
	return withBase(`/inbound-requests/${encodeURIComponent(requestLookup)}`);
}

export function approvalDetail(approvalID: string): string {
	return withBase(`/review/approvals/${encodeURIComponent(approvalID)}`);
}

export function proposalDetail(recommendationID: string): string {
	return withBase(`/review/proposals/${encodeURIComponent(recommendationID)}`);
}

export function documentDetail(documentID: string): string {
	return withBase(`/review/documents/${encodeURIComponent(documentID)}`);
}

export function accountingEntryDetail(entryID: string): string {
	return withBase(`/review/accounting/${encodeURIComponent(entryID)}`);
}

export function inventoryMovementDetail(movementID: string): string {
	return withBase(`/review/inventory/${encodeURIComponent(movementID)}`);
}

export function workOrderDetail(workOrderID: string): string {
	return withBase(`/review/work-orders/${encodeURIComponent(workOrderID)}`);
}

export function auditEventDetail(eventID: string): string {
	return withBase(`/review/audit/${encodeURIComponent(eventID)}`);
}

export function withQuery(path: string, query: Record<string, string | number | undefined>): string {
	const params = new URLSearchParams();
	for (const [key, value] of Object.entries(query)) {
		if (value === undefined) {
			continue;
		}
		const text = String(value).trim();
		if (text === '') {
			continue;
		}
		params.set(key, text);
	}
	const suffix = params.size > 0 ? `?${params.toString()}` : '';
	return `${path}${suffix}`;
}
