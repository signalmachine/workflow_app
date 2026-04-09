import { afterEach, describe, expect, it, vi } from 'vitest';

import { APIClientError } from './client';
import {
	getInboundRequestDetail,
	getInventoryMovementDetail,
	getProcessedProposalDetail,
	listInventoryReconciliation
} from './review';

describe('review api', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('requests inbound request detail by encoded lookup', async () => {
		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			expect(String(input)).toBe('/api/review/inbound-requests/run%3Arun-123');
			return new Response(
				JSON.stringify({
					request: {
						request_id: 'req-1',
						request_reference: 'REQ-1',
						origin_type: 'human',
						channel: 'browser',
						status: 'processed',
						metadata: {},
						received_at: '2026-04-04T00:00:00Z',
						created_at: '2026-04-04T00:00:00Z',
						updated_at: '2026-04-04T00:00:00Z',
						message_count: 1,
						attachment_count: 0
					},
					messages: [],
					attachments: [],
					runs: [],
					steps: [],
					delegations: [],
					artifacts: [],
					recommendations: [],
					proposals: []
				}),
				{ status: 200, headers: { 'Content-Type': 'application/json' } }
			);
		});

		const detail = await getInboundRequestDetail('run:run-123', fetchMock as typeof fetch);

		expect(detail.request.request_reference).toBe('REQ-1');
		expect(fetchMock).toHaveBeenCalledTimes(1);
	});

	it('surfaces not found errors for missing detail routes', async () => {
		const fetchMock = vi.fn(async () => new Response(JSON.stringify({ error: 'record not found' }), { status: 404, headers: { 'Content-Type': 'application/json' } }));

		await expect(getInboundRequestDetail('REQ-404', fetchMock as typeof fetch)).rejects.toEqual(new APIClientError(404, 'record not found'));
	});

	it('requests processed proposal detail by encoded id', async () => {
		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			expect(String(input)).toBe('/api/review/processed-proposals/rec%3A123');
			return new Response(
				JSON.stringify({
					request_id: 'req-1',
					request_reference: 'REQ-1',
					request_status: 'processed',
					recommendation_id: 'rec:123',
					run_id: 'run-1',
					recommendation_type: 'draft_document',
					recommendation_status: 'proposed',
					summary: 'Draft invoice',
					created_at: '2026-04-04T00:00:00Z'
				}),
				{ status: 200, headers: { 'Content-Type': 'application/json' } }
			);
		});

		const detail = await getProcessedProposalDetail('rec:123', fetchMock as typeof fetch);

		expect(detail.recommendation_id).toBe('rec:123');
		expect(fetchMock).toHaveBeenCalledTimes(1);
	});

	it('requests inventory movement detail by encoded id', async () => {
		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			expect(String(input)).toBe('/api/review/inventory/movements/move%2F123');
			return new Response(
				JSON.stringify({
					review: {
						movement_id: 'move/123',
						movement_number: 42,
						item_id: 'item-1',
						item_sku: 'MAT-1',
						item_name: 'Copper pipe',
						item_role: 'material',
						movement_type: 'issue',
						movement_purpose: 'execution',
						usage_classification: 'billable',
						quantity_milli: 500,
						reference_note: '',
						created_by_user_id: 'user-1',
						created_at: '2026-04-04T00:00:00Z'
					},
					reconciliation: []
				}),
				{ status: 200, headers: { 'Content-Type': 'application/json' } }
			);
		});

		const detail = await getInventoryMovementDetail('move/123', fetchMock as typeof fetch);

		expect(detail.review.movement_id).toBe('move/123');
		expect(fetchMock).toHaveBeenCalledTimes(1);
	});

	it('preserves pending handoff filters for inventory reconciliation review', async () => {
		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			expect(String(input)).toBe(
				'/api/review/inventory/reconciliation?document_id=doc-1&only_pending_accounting=true&only_pending_execution=true&limit=25'
			);
			return new Response(JSON.stringify({ items: [] }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' }
			});
		});

		const response = await listInventoryReconciliation(
			{
				documentID: 'doc-1',
				onlyPendingAccounting: true,
				onlyPendingExecution: true,
				limit: 25
			},
			fetchMock as typeof fetch
		);

		expect(response.items).toEqual([]);
		expect(fetchMock).toHaveBeenCalledTimes(1);
	});
});
