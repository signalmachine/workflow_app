import { apiRequest } from '$lib/api/client';
import type {
	AccountingPeriod,
	Contact,
	InventoryItem,
	InventoryLocation,
	LedgerAccount,
	OrgUserMembership,
	Party,
	PartyDetailResponse,
	TaxCode
} from '$lib/api/types';

export function listLedgerAccounts(fetcher: typeof fetch = fetch): Promise<{ items: LedgerAccount[] }> {
	return apiRequest('/api/admin/accounting/ledger-accounts', undefined, fetcher);
}

export function createLedgerAccount(
	payload: {
		code: string;
		name: string;
		account_class: string;
		control_type: string;
		allows_direct_posting: boolean;
		tax_category_code?: string;
	},
	fetcher: typeof fetch = fetch
): Promise<LedgerAccount> {
	return apiRequest('/api/admin/accounting/ledger-accounts', { method: 'POST', body: JSON.stringify(payload) }, fetcher);
}

export function updateLedgerAccountStatus(accountID: string, status: string, fetcher: typeof fetch = fetch): Promise<LedgerAccount> {
	return apiRequest(`/api/admin/accounting/ledger-accounts/${accountID}/status`, {
		method: 'POST',
		body: JSON.stringify({ status })
	}, fetcher);
}

export function listTaxCodes(fetcher: typeof fetch = fetch): Promise<{ items: TaxCode[] }> {
	return apiRequest('/api/admin/accounting/tax-codes', undefined, fetcher);
}

export function createTaxCode(
	payload: {
		code: string;
		name: string;
		tax_type: string;
		rate_basis_points: number;
		receivable_account_id?: string;
		payable_account_id?: string;
	},
	fetcher: typeof fetch = fetch
): Promise<TaxCode> {
	return apiRequest('/api/admin/accounting/tax-codes', { method: 'POST', body: JSON.stringify(payload) }, fetcher);
}

export function updateTaxCodeStatus(taxCodeID: string, status: string, fetcher: typeof fetch = fetch): Promise<TaxCode> {
	return apiRequest(`/api/admin/accounting/tax-codes/${taxCodeID}/status`, {
		method: 'POST',
		body: JSON.stringify({ status })
	}, fetcher);
}

export function listAccountingPeriods(fetcher: typeof fetch = fetch): Promise<{ items: AccountingPeriod[] }> {
	return apiRequest('/api/admin/accounting/periods', undefined, fetcher);
}

export function createAccountingPeriod(
	payload: { period_code: string; start_on: string; end_on: string },
	fetcher: typeof fetch = fetch
): Promise<AccountingPeriod> {
	return apiRequest('/api/admin/accounting/periods', { method: 'POST', body: JSON.stringify(payload) }, fetcher);
}

export function closeAccountingPeriod(periodID: string, fetcher: typeof fetch = fetch): Promise<AccountingPeriod> {
	return apiRequest(`/api/admin/accounting/periods/${periodID}/close`, { method: 'POST', body: JSON.stringify({}) }, fetcher);
}

export function listParties(partyKind = '', fetcher: typeof fetch = fetch): Promise<{ items: Party[] }> {
	const params = new URLSearchParams();
	if (partyKind.trim() !== '') {
		params.set('party_kind', partyKind.trim());
	}
	const suffix = params.size > 0 ? `?${params.toString()}` : '';
	return apiRequest(`/api/admin/parties${suffix}`, undefined, fetcher);
}

export function createParty(
	payload: { party_code: string; display_name: string; legal_name?: string; party_kind: string },
	fetcher: typeof fetch = fetch
): Promise<Party> {
	return apiRequest('/api/admin/parties', { method: 'POST', body: JSON.stringify(payload) }, fetcher);
}

export function getPartyDetail(partyID: string, fetcher: typeof fetch = fetch): Promise<PartyDetailResponse> {
	return apiRequest(`/api/admin/parties/${partyID}`, undefined, fetcher);
}

export function createContact(
	partyID: string,
	payload: { full_name: string; role_title?: string; email?: string; phone?: string; is_primary: boolean },
	fetcher: typeof fetch = fetch
): Promise<Contact> {
	return apiRequest(`/api/admin/parties/${partyID}/contacts`, { method: 'POST', body: JSON.stringify(payload) }, fetcher);
}

export function updatePartyStatus(partyID: string, status: string, fetcher: typeof fetch = fetch): Promise<Party> {
	return apiRequest(`/api/admin/parties/${partyID}/status`, { method: 'POST', body: JSON.stringify({ status }) }, fetcher);
}

export function listOrgUsers(fetcher: typeof fetch = fetch): Promise<{ items: OrgUserMembership[] }> {
	return apiRequest('/api/admin/access/users', undefined, fetcher);
}

export function provisionOrgUser(
	payload: { email: string; display_name: string; role_code: string; password: string },
	fetcher: typeof fetch = fetch
): Promise<OrgUserMembership> {
	return apiRequest('/api/admin/access/users', { method: 'POST', body: JSON.stringify(payload) }, fetcher);
}

export function updateMembershipRole(
	membershipID: string,
	roleCode: string,
	fetcher: typeof fetch = fetch
): Promise<OrgUserMembership> {
	return apiRequest(`/api/admin/access/users/${membershipID}/role`, {
		method: 'POST',
		body: JSON.stringify({ role_code: roleCode })
	}, fetcher);
}

export function listInventoryItems(itemRole = '', fetcher: typeof fetch = fetch): Promise<{ items: InventoryItem[] }> {
	const params = new URLSearchParams();
	if (itemRole.trim() !== '') {
		params.set('item_role', itemRole.trim());
	}
	const suffix = params.size > 0 ? `?${params.toString()}` : '';
	return apiRequest(`/api/admin/inventory/items${suffix}`, undefined, fetcher);
}

export function createInventoryItem(
	payload: { sku: string; name: string; item_role: string; tracking_mode: string },
	fetcher: typeof fetch = fetch
): Promise<InventoryItem> {
	return apiRequest('/api/admin/inventory/items', { method: 'POST', body: JSON.stringify(payload) }, fetcher);
}

export function updateInventoryItemStatus(itemID: string, status: string, fetcher: typeof fetch = fetch): Promise<InventoryItem> {
	return apiRequest(`/api/admin/inventory/items/${itemID}/status`, {
		method: 'POST',
		body: JSON.stringify({ status })
	}, fetcher);
}

export function listInventoryLocations(locationRole = '', fetcher: typeof fetch = fetch): Promise<{ items: InventoryLocation[] }> {
	const params = new URLSearchParams();
	if (locationRole.trim() !== '') {
		params.set('location_role', locationRole.trim());
	}
	const suffix = params.size > 0 ? `?${params.toString()}` : '';
	return apiRequest(`/api/admin/inventory/locations${suffix}`, undefined, fetcher);
}

export function createInventoryLocation(
	payload: { code: string; name: string; location_role: string },
	fetcher: typeof fetch = fetch
): Promise<InventoryLocation> {
	return apiRequest('/api/admin/inventory/locations', { method: 'POST', body: JSON.stringify(payload) }, fetcher);
}

export function updateInventoryLocationStatus(
	locationID: string,
	status: string,
	fetcher: typeof fetch = fetch
): Promise<InventoryLocation> {
	return apiRequest(`/api/admin/inventory/locations/${locationID}/status`, {
		method: 'POST',
		body: JSON.stringify({ status })
	}, fetcher);
}
