import { apiRequest } from '$lib/api/client';
import type {
	AIDelegation,
	AIArtifact,
	AIRecommendation,
	AIRun,
	AIStep,
	ApprovalQueueEntry,
	AuditEvent,
	ControlAccountBalance,
	DocumentReview,
	InventoryMovementDetail,
	InboundRequestDetail,
	InboundRequestMessage,
	InboundRequestReview,
	InboundRequestStatusSummary,
	InventoryMovementReview,
	InventoryReconciliationItem,
	InventoryStockItem,
	JournalEntryReview,
	ProcessedProposalReview,
	ProcessedProposalStatusSummary,
	RequestAttachment,
	TaxSummary,
	WorkOrderReview
} from '$lib/api/types';

function withSearch(path: string, params: Record<string, string | number | boolean | undefined>): string {
	const search = new URLSearchParams();
	for (const [key, value] of Object.entries(params)) {
		if (value === undefined) {
			continue;
		}
		const text = String(value).trim();
		if (text === '') {
			continue;
		}
		search.set(key, text);
	}
	return search.size > 0 ? `${path}?${search.toString()}` : path;
}

export function listInboundRequestStatusSummary(fetcher: typeof fetch = fetch): Promise<{ items: InboundRequestStatusSummary[] }> {
	return apiRequest('/api/review/inbound-request-status-summary', undefined, fetcher);
}

export function listInboundRequests(
	params: { status?: string; requestReference?: string; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: InboundRequestReview[] }> {
	return apiRequest(
		withSearch('/api/review/inbound-requests', {
			status: params.status,
			request_reference: params.requestReference,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function getInboundRequestDetail(
	lookup: string,
	fetcher: typeof fetch = fetch
): Promise<{
	request: InboundRequestReview;
	messages: InboundRequestMessage[];
	attachments: RequestAttachment[];
	runs: AIRun[];
	steps: AIStep[];
	delegations: AIDelegation[];
	artifacts: AIArtifact[];
	recommendations: AIRecommendation[];
	proposals: ProcessedProposalReview[];
}> {
	return apiRequest<InboundRequestDetail>(`/api/review/inbound-requests/${encodeURIComponent(lookup)}`, undefined, fetcher);
}

export function listProcessedProposalStatusSummary(fetcher: typeof fetch = fetch): Promise<{ items: ProcessedProposalStatusSummary[] }> {
	return apiRequest('/api/review/processed-proposal-status-summary', undefined, fetcher);
}

export function getProcessedProposalDetail(recommendationID: string, fetcher: typeof fetch = fetch): Promise<ProcessedProposalReview> {
	return apiRequest(`/api/review/processed-proposals/${encodeURIComponent(recommendationID)}`, undefined, fetcher);
}

export function listProcessedProposals(
	params: { status?: string; requestReference?: string; recommendationID?: string; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: ProcessedProposalReview[] }> {
	return apiRequest(
		withSearch('/api/review/processed-proposals', {
			status: params.status,
			request_reference: params.requestReference,
			recommendation_id: params.recommendationID,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function listApprovalQueue(
	params: { status?: string; queueCode?: string; approvalID?: string; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: ApprovalQueueEntry[] }> {
	return apiRequest(
		withSearch('/api/review/approval-queue', {
			status: params.status,
			queue_code: params.queueCode,
			approval_id: params.approvalID,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function getApprovalQueueDetail(approvalID: string, fetcher: typeof fetch = fetch): Promise<ApprovalQueueEntry> {
	return apiRequest(`/api/review/approval-queue/${encodeURIComponent(approvalID)}`, undefined, fetcher);
}

export function listDocuments(
	params: { status?: string; typeCode?: string; documentID?: string; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: DocumentReview[] }> {
	return apiRequest(
		withSearch('/api/review/documents', {
			status: params.status,
			type_code: params.typeCode,
			document_id: params.documentID,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function getDocumentReview(documentID: string, fetcher: typeof fetch = fetch): Promise<DocumentReview> {
	return apiRequest(`/api/review/documents/${encodeURIComponent(documentID)}`, undefined, fetcher);
}

export function listJournalEntries(
	params: { startOn?: string; endOn?: string; entryID?: string; documentID?: string; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: JournalEntryReview[] }> {
	return apiRequest(
		withSearch('/api/review/accounting/journal-entries', {
			start_on: params.startOn,
			end_on: params.endOn,
			entry_id: params.entryID,
			document_id: params.documentID,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function getJournalEntryDetail(entryID: string, fetcher: typeof fetch = fetch): Promise<JournalEntryReview> {
	return apiRequest(`/api/review/accounting/journal-entries/${encodeURIComponent(entryID)}`, undefined, fetcher);
}

export function listControlAccountBalances(
	params: { asOf?: string; controlType?: string; accountID?: string },
	fetcher: typeof fetch = fetch
): Promise<{ items: ControlAccountBalance[] }> {
	return apiRequest(
		withSearch('/api/review/accounting/control-account-balances', {
			as_of: params.asOf,
			control_type: params.controlType,
			account_id: params.accountID
		}),
		undefined,
		fetcher
	);
}

export function listTaxSummaries(
	params: { startOn?: string; endOn?: string; taxType?: string; taxCode?: string },
	fetcher: typeof fetch = fetch
): Promise<{ items: TaxSummary[] }> {
	return apiRequest(
		withSearch('/api/review/accounting/tax-summaries', {
			start_on: params.startOn,
			end_on: params.endOn,
			tax_type: params.taxType,
			tax_code: params.taxCode
		}),
		undefined,
		fetcher
	);
}

export function listInventoryStock(
	params: { itemID?: string; locationID?: string; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: InventoryStockItem[] }> {
	return apiRequest(
		withSearch('/api/review/inventory/stock', {
			item_id: params.itemID,
			location_id: params.locationID,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function listInventoryMovements(
	params: { movementID?: string; itemID?: string; locationID?: string; documentID?: string; movementType?: string; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: InventoryMovementReview[] }> {
	return apiRequest(
		withSearch('/api/review/inventory/movements', {
			movement_id: params.movementID,
			item_id: params.itemID,
			location_id: params.locationID,
			document_id: params.documentID,
			movement_type: params.movementType,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function getInventoryMovementDetail(movementID: string, fetcher: typeof fetch = fetch): Promise<InventoryMovementDetail> {
	return apiRequest(`/api/review/inventory/movements/${encodeURIComponent(movementID)}`, undefined, fetcher);
}

export function listInventoryReconciliation(
	params: { documentID?: string; onlyPendingAccounting?: boolean; onlyPendingExecution?: boolean; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: InventoryReconciliationItem[] }> {
	return apiRequest(
		withSearch('/api/review/inventory/reconciliation', {
			document_id: params.documentID,
			only_pending_accounting: params.onlyPendingAccounting,
			only_pending_execution: params.onlyPendingExecution,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function listWorkOrders(
	params: { status?: string; workOrderID?: string; documentID?: string; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: WorkOrderReview[] }> {
	return apiRequest(
		withSearch('/api/review/work-orders', {
			status: params.status,
			work_order_id: params.workOrderID,
			document_id: params.documentID,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function getWorkOrderReview(workOrderID: string, fetcher: typeof fetch = fetch): Promise<WorkOrderReview> {
	return apiRequest(`/api/review/work-orders/${encodeURIComponent(workOrderID)}`, undefined, fetcher);
}

export function listAuditEvents(
	params: { eventID?: string; entityType?: string; entityID?: string; limit?: number },
	fetcher: typeof fetch = fetch
): Promise<{ items: AuditEvent[] }> {
	return apiRequest(
		withSearch('/api/review/audit-events', {
			event_id: params.eventID,
			entity_type: params.entityType,
			entity_id: params.entityID,
			limit: params.limit
		}),
		undefined,
		fetcher
	);
}

export function getAuditEventDetail(eventID: string, fetcher: typeof fetch = fetch): Promise<AuditEvent> {
	return apiRequest(`/api/review/audit-events/${encodeURIComponent(eventID)}`, undefined, fetcher);
}
