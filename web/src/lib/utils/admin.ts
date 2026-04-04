import { redirect } from '@sveltejs/kit';

import type { SessionContext } from '$lib/api/types';
import { routes } from '$lib/utils/routes';

export const adminRoleCode = 'admin';
export const statusOptions = ['active', 'inactive'] as const;
export const accountClassOptions = ['asset', 'liability', 'equity', 'revenue', 'expense'] as const;
export const controlTypeOptions = ['none', 'receivable', 'payable', 'gst_input', 'gst_output', 'tds_receivable', 'tds_payable'] as const;
export const taxTypeOptions = ['gst', 'tds'] as const;
export const partyKindOptions = ['customer', 'vendor', 'customer_vendor', 'other'] as const;
export const roleOptions = ['admin', 'operator', 'approver'] as const;
export const itemRoleOptions = ['resale', 'service_material', 'traceable_equipment', 'direct_expense_consumable'] as const;
export const trackingModeOptions = ['none', 'serial', 'lot'] as const;
export const locationRoleOptions = ['warehouse', 'van', 'site', 'vendor', 'customer', 'adjustment', 'installed'] as const;

export function requireAdmin(session: SessionContext): void {
	if (session.role_code !== adminRoleCode) {
		throw redirect(307, `${routes.home}?error=${encodeURIComponent('Admin access required.')}`);
	}
}
