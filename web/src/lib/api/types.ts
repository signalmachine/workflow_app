export interface APIError {
	error: string;
}

export interface SessionContext {
	session_id: string;
	org_id: string;
	org_slug: string;
	org_name: string;
	user_id: string;
	user_email: string;
	user_display_name: string;
	role_code: string;
	device_label: string;
	expires_at: string;
	issued_at: string;
	last_seen_at: string;
}

export interface SessionLoginRequest {
	org_slug: string;
	email: string;
	password: string;
	device_label: string;
}

export interface SubmitInboundRequestAttachment {
	original_file_name: string;
	media_type: string;
	content_base64: string;
	link_role: string;
}

export interface SubmitInboundRequestPayload {
	origin_type: string;
	channel: string;
	metadata: Record<string, unknown>;
	message: {
		message_role: string;
		text_content: string;
	};
	attachments: SubmitInboundRequestAttachment[];
	queue_for_review?: boolean;
}

export interface SubmitInboundRequestResponse {
	request_id: string;
	request_reference: string;
	status: string;
	message_id?: string;
	attachment_ids?: string[];
	cancellation_reason?: string;
	failure_reason?: string;
	received_at: string;
	queued_at?: string;
	processing_started_at?: string;
	processed_at?: string;
	acted_on_at?: string;
	completed_at?: string;
	failed_at?: string;
	cancelled_at?: string;
	created_at: string;
	updated_at: string;
}

export interface ProcessNextQueuedResponse {
	processed: boolean;
	request_reference?: string;
	request_status?: string;
	run_id?: string;
	run_status?: string;
	artifact_id?: string;
	recommendation_id?: string;
	recommendation_summary?: string;
}

export interface HomeAction {
	title: string;
	summary: string;
	href: string;
	badge?: string;
}

export interface InboundRequestStatusSummary {
	status: string;
	request_count: number;
	message_count: number;
	attachment_count: number;
	latest_received_at?: string;
	latest_queued_at?: string;
	latest_updated_at: string;
}

export interface InboundRequestReview {
	request_id: string;
	request_reference: string;
	session_id?: string;
	actor_user_id?: string;
	origin_type: string;
	channel: string;
	status: string;
	metadata: Record<string, unknown>;
	cancellation_reason?: string;
	failure_reason?: string;
	received_at: string;
	queued_at?: string;
	processing_started_at?: string;
	processed_at?: string;
	acted_on_at?: string;
	completed_at?: string;
	failed_at?: string;
	cancelled_at?: string;
	created_at: string;
	updated_at: string;
	message_count: number;
	attachment_count: number;
	last_run_id?: string;
	last_run_status?: string;
	last_recommendation_id?: string;
	last_recommendation_status?: string;
}

export interface ProcessedProposalStatusSummary {
	recommendation_status: string;
	proposal_count: number;
	request_count: number;
	document_count: number;
	latest_created_at: string;
}

export interface ProcessedProposalReview {
	request_id: string;
	request_reference: string;
	request_status: string;
	recommendation_id: string;
	run_id: string;
	recommendation_type: string;
	recommendation_status: string;
	summary: string;
	suggested_queue_code?: string;
	approval_id?: string;
	approval_status?: string;
	approval_queue_code?: string;
	document_id?: string;
	document_type_code?: string;
	document_title?: string;
	document_number?: string;
	document_status?: string;
	created_at: string;
}

export interface ApprovalQueueEntry {
	queue_entry_id: string;
	approval_id: string;
	queue_code: string;
	queue_status: string;
	enqueued_at: string;
	closed_at?: string;
	approval_status: string;
	requested_at: string;
	requested_by_user_id: string;
	decided_at?: string;
	decided_by_user_id?: string;
	document_id: string;
	document_type_code: string;
	document_title: string;
	document_number?: string;
	document_status: string;
	journal_entry_id?: string;
	journal_entry_number?: number;
	journal_entry_posted_at?: string;
}

export interface DocumentReview {
	document_id: string;
	type_code: string;
	title: string;
	number_value?: string;
	status: string;
	source_document_id?: string;
	created_by_user_id: string;
	submitted_by_user_id?: string;
	submitted_at?: string;
	approved_at?: string;
	rejected_at?: string;
	created_at: string;
	updated_at: string;
	approval_id?: string;
	approval_status?: string;
	approval_queue_code?: string;
	approval_requested_at?: string;
	approval_decided_at?: string;
	journal_entry_id?: string;
	journal_entry_number?: number;
	journal_entry_posted_at?: string;
}

export interface JournalEntryReview {
	entry_id: string;
	entry_number: number;
	entry_kind: string;
	source_document_id?: string;
	reversal_of_entry_id?: string;
	currency_code: string;
	tax_scope_code: string;
	summary: string;
	reversal_reason?: string;
	posted_by_user_id: string;
	effective_on: string;
	posted_at: string;
	created_at: string;
	document_type_code?: string;
	document_number?: string;
	document_status?: string;
	approval_id?: string;
	approval_status?: string;
	approval_queue_code?: string;
	request_id?: string;
	request_reference?: string;
	recommendation_id?: string;
	recommendation_status?: string;
	run_id?: string;
	line_count: number;
	total_debit_minor: number;
	total_credit_minor: number;
	has_reversal: boolean;
}

export interface ControlAccountBalance {
	account_id: string;
	account_code: string;
	account_name: string;
	account_class: string;
	control_type: string;
	total_debit_minor: number;
	total_credit_minor: number;
	net_minor: number;
	last_effective_on?: string;
}

export interface TaxSummary {
	tax_type: string;
	tax_code: string;
	tax_name: string;
	rate_basis_points: number;
	entry_count: number;
	document_count: number;
	total_debit_minor: number;
	total_credit_minor: number;
	net_minor: number;
	receivable_account_id?: string;
	receivable_account_code?: string;
	receivable_account_name?: string;
	payable_account_id?: string;
	payable_account_code?: string;
	payable_account_name?: string;
	last_effective_on?: string;
}

export interface InventoryStockItem {
	item_id: string;
	item_sku: string;
	item_name: string;
	item_role: string;
	location_id: string;
	location_code: string;
	location_name: string;
	location_role: string;
	on_hand_milli: number;
}

export interface InventoryMovementReview {
	movement_id: string;
	movement_number: number;
	document_id?: string;
	document_type_code?: string;
	document_title?: string;
	document_number?: string;
	document_status?: string;
	approval_id?: string;
	approval_status?: string;
	approval_queue_code?: string;
	request_id?: string;
	request_reference?: string;
	recommendation_id?: string;
	recommendation_status?: string;
	run_id?: string;
	item_id: string;
	item_sku: string;
	item_name: string;
	item_role: string;
	movement_type: string;
	movement_purpose: string;
	usage_classification: string;
	source_location_id?: string;
	source_location_code?: string;
	source_location_name?: string;
	source_location_role?: string;
	destination_location_id?: string;
	destination_location_code?: string;
	destination_location_name?: string;
	destination_location_role?: string;
	quantity_milli: number;
	reference_note: string;
	created_by_user_id: string;
	created_at: string;
}

export interface InventoryReconciliationItem {
	document_id: string;
	document_type_code: string;
	document_title: string;
	document_number?: string;
	document_status: string;
	approval_id?: string;
	approval_status?: string;
	approval_queue_code?: string;
	request_id?: string;
	request_reference?: string;
	recommendation_id?: string;
	recommendation_status?: string;
	run_id?: string;
	document_line_id: string;
	line_number: number;
	movement_id: string;
	movement_number: number;
	movement_type: string;
	movement_purpose: string;
	usage_classification: string;
	item_id: string;
	item_sku: string;
	item_name: string;
	item_role: string;
	source_location_id?: string;
	source_location_code?: string;
	source_location_name?: string;
	destination_location_id?: string;
	destination_location_code?: string;
	destination_location_name?: string;
	quantity_milli: number;
	execution_link_id?: string;
	execution_context_type?: string;
	execution_context_id?: string;
	execution_link_status?: string;
	work_order_id?: string;
	work_order_code?: string;
	work_order_status?: string;
	accounting_handoff_id?: string;
	accounting_handoff_status?: string;
	cost_minor?: number;
	cost_currency_code?: string;
	journal_entry_id?: string;
	journal_entry_number?: number;
	accounting_posted_at?: string;
	movement_created_at: string;
}

export interface WorkOrderReview {
	work_order_id: string;
	document_id: string;
	document_status: string;
	document_number?: string;
	work_order_code: string;
	title: string;
	summary: string;
	status: string;
	closed_at?: string;
	created_at: string;
	updated_at: string;
	last_status_changed_at: string;
	open_task_count: number;
	completed_task_count: number;
	labor_entry_count: number;
	total_labor_minutes: number;
	total_labor_cost_minor: number;
	posted_labor_entry_count: number;
	posted_labor_cost_minor: number;
	material_usage_count: number;
	material_quantity_milli: number;
	posted_material_usage_count: number;
	posted_material_cost_minor: number;
	last_accounting_posted_at?: string;
}

export interface AuditEvent {
	id: string;
	org_id?: string;
	actor_user_id?: string;
	event_type: string;
	entity_type: string;
	entity_id: string;
	payload: Record<string, unknown>;
	occurred_at: string;
}

export interface OperationsFeedItem {
	occurred_at: string;
	kind: string;
	title: string;
	summary: string;
	status: string;
	primary_label: string;
	primary_href: string;
	secondary_label?: string;
	secondary_href?: string;
}

export interface DashboardSnapshot {
	role_headline: string;
	role_body: string;
	primary_actions: HomeAction[];
	secondary_actions: HomeAction[];
	inbound_summary: InboundRequestStatusSummary[];
	proposal_summary: ProcessedProposalStatusSummary[];
	inbound_requests: InboundRequestReview[];
	proposals: ProcessedProposalReview[];
	approvals: ApprovalQueueEntry[];
}

export interface OperationsSnapshot {
	queued_request_count: number;
	pending_approval_count: number;
	proposal_review_count: number;
	recent_feed: OperationsFeedItem[];
}

export interface OperationsFeedSnapshot {
	items: OperationsFeedItem[];
}

export interface ReviewLandingSnapshot {
	inbound_summary: InboundRequestStatusSummary[];
	proposal_summary: ProcessedProposalStatusSummary[];
	pending_approvals: ApprovalQueueEntry[];
	inbound_request_count: number;
	proposal_count: number;
}

export interface AgentChatSnapshot {
	request_reference?: string;
	request_status?: string;
	recent_requests: InboundRequestReview[];
	recent_proposals: ProcessedProposalReview[];
}

export interface RouteCatalogEntry {
	title: string;
	href: string;
	category: string;
	summary: string;
}

export interface RouteCatalogSnapshot {
	query: string;
	items: RouteCatalogEntry[];
}

export interface LedgerAccount {
	id: string;
	code: string;
	name: string;
	account_class: string;
	control_type: string;
	allows_direct_posting: boolean;
	status: string;
	tax_category_code?: string;
	created_by_user_id: string;
	created_at: string;
	updated_at: string;
}

export interface TaxCode {
	id: string;
	code: string;
	name: string;
	tax_type: string;
	rate_basis_points: number;
	receivable_account_id?: string;
	payable_account_id?: string;
	status: string;
	created_by_user_id: string;
	created_at: string;
	updated_at: string;
}

export interface AccountingPeriod {
	id: string;
	period_code: string;
	start_on: string;
	end_on: string;
	status: string;
	closed_by_user_id?: string;
	closed_at?: string;
	created_by_user_id: string;
	created_at: string;
	updated_at: string;
}

export interface Party {
	id: string;
	party_code: string;
	display_name: string;
	legal_name?: string;
	party_kind: string;
	status: string;
	created_by_user_id: string;
	created_at: string;
	updated_at: string;
}

export interface Contact {
	id: string;
	party_id: string;
	full_name: string;
	role_title?: string;
	email?: string;
	phone?: string;
	is_primary: boolean;
	status: string;
	created_by_user_id: string;
	created_at: string;
	updated_at: string;
}

export interface PartyDetailResponse {
	party: Party;
	contacts: Contact[];
}

export interface OrgUserMembership {
	membership_id: string;
	org_id: string;
	user_id: string;
	user_email: string;
	user_display_name: string;
	user_status: string;
	role_code: string;
	membership_status: string;
	created_at: string;
}

export interface InventoryItem {
	id: string;
	sku: string;
	name: string;
	item_role: string;
	tracking_mode: string;
	status: string;
	created_by_user_id: string;
	created_at: string;
	updated_at: string;
}

export interface InventoryLocation {
	id: string;
	code: string;
	name: string;
	location_role: string;
	status: string;
	created_by_user_id: string;
	created_at: string;
	updated_at: string;
}
