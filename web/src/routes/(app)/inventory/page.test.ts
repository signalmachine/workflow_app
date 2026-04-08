import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import InventoryLandingPage from './+page.svelte';

const baseData = {
	session: {
		session_id: 'session-1',
		org_id: 'org-1',
		org_slug: 'north-harbor',
		org_name: 'North Harbor Works',
		user_id: 'user-1',
		user_email: 'operator@example.com',
		user_display_name: 'Operator Example',
		role_code: 'operator',
		device_label: 'browser',
		expires_at: '2026-04-09T00:00:00Z',
		issued_at: '2026-04-08T00:00:00Z',
		last_seen_at: '2026-04-08T00:00:00Z'
	},
	roleCode: 'operator',
	stock: [
		{
			item_id: 'item-1',
			item_sku: 'SKU-1',
			item_name: 'Copper pipe',
			item_role: 'material',
			location_id: 'loc-1',
			location_code: 'MAIN',
			location_name: 'Main warehouse',
			location_role: 'warehouse',
			on_hand_milli: 12500
		}
	],
	movements: [
		{
			movement_id: 'move/1',
			movement_number: 101,
			document_id: 'doc-1',
			document_title: 'Inventory issue',
			item_id: 'item-1',
			item_sku: 'SKU-1',
			item_name: 'Copper pipe',
			item_role: 'material',
			movement_type: 'issue',
			movement_purpose: 'execution',
			usage_classification: 'billable',
			source_location_code: 'MAIN',
			source_location_name: 'Main warehouse',
			quantity_milli: 5000,
			reference_note: '',
			created_by_user_id: 'user-1',
			created_at: '2026-04-08T10:00:00Z'
		}
	],
	pendingExecution: [
		{
			document_id: 'doc-1',
			document_type_code: 'inventory_issue',
			document_title: 'Inventory issue',
			document_status: 'approved',
			document_line_id: 'line-1',
			line_number: 1,
			movement_id: 'move/1',
			movement_number: 101,
			movement_type: 'issue',
			movement_purpose: 'execution',
			usage_classification: 'billable',
			item_id: 'item-1',
			item_sku: 'SKU-1',
			item_name: 'Copper pipe',
			item_role: 'material',
			quantity_milli: 5000,
			execution_link_status: 'pending',
			work_order_id: 'wo-1',
			work_order_code: 'WO-001',
			work_order_status: 'open',
			movement_created_at: '2026-04-08T10:00:00Z'
		}
	],
	pendingAccounting: [
		{
			document_id: 'doc-1',
			document_type_code: 'inventory_issue',
			document_title: 'Inventory issue',
			document_status: 'approved',
			document_line_id: 'line-1',
			line_number: 1,
			movement_id: 'move/1',
			movement_number: 101,
			movement_type: 'issue',
			movement_purpose: 'execution',
			usage_classification: 'billable',
			item_id: 'item-1',
			item_sku: 'SKU-1',
			item_name: 'Copper pipe',
			item_role: 'material',
			quantity_milli: 5000,
			accounting_handoff_status: 'pending',
			cost_minor: 125000,
			movement_created_at: '2026-04-08T10:00:00Z'
		}
	]
};

describe('inventory landing page', () => {
	it('renders live inventory actions and exact continuity links for operators', () => {
		render(InventoryLandingPage, { props: { data: baseData } as never });

		expect(screen.getByText('Inventory landing')).toBeTruthy();
		expect(screen.getByText('Visible stock positions')).toBeTruthy();
		expect(screen.getByRole('link', { name: /Pending execution handoffs/ }).getAttribute('href')).toBe(
			'/app/review/inventory?only_pending_execution=true'
		);
		expect(screen.getByRole('link', { name: /Pending accounting handoffs/ }).getAttribute('href')).toBe(
			'/app/review/inventory?only_pending_accounting=true'
		);
		expect(screen.getAllByRole('link', { name: '101' })[0]?.getAttribute('href')).toBe('/app/review/inventory/move%2F1');
		expect(screen.getByRole('link', { name: 'WO-001' }).getAttribute('href')).toBe('/app/review/work-orders/wo-1');
		expect(screen.queryByRole('link', { name: /Inventory setup/ })).toBeNull();
	});

	it('shows inventory setup continuity for admin actors', () => {
		render(InventoryLandingPage, {
			props: {
				data: {
					...baseData,
					roleCode: 'admin'
				}
			} as never
		});

		expect(screen.getByRole('link', { name: /Inventory setup/ }).getAttribute('href')).toBe('/app/admin/inventory');
	});
});
