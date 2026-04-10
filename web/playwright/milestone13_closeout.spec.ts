import { expect, test, type Page } from '@playwright/test';

const env = {
	orgSlug: process.env.PLAYWRIGHT_ORG_SLUG?.trim() ?? '',
	email: process.env.PLAYWRIGHT_EMAIL?.trim() ?? '',
	password: process.env.PLAYWRIGHT_PASSWORD ?? '',
	requestReference: process.env.PLAYWRIGHT_REQUEST_REFERENCE?.trim() ?? '',
	continuityOrgSlug: process.env.PLAYWRIGHT_CONTINUITY_ORG_SLUG?.trim() ?? '',
	continuityEmail: process.env.PLAYWRIGHT_CONTINUITY_EMAIL?.trim() ?? '',
	continuityPassword: process.env.PLAYWRIGHT_CONTINUITY_PASSWORD ?? '',
	continuityRequestReference: process.env.PLAYWRIGHT_CONTINUITY_REQUEST_REFERENCE?.trim() ?? '',
	continuityRecommendationID: process.env.PLAYWRIGHT_CONTINUITY_RECOMMENDATION_ID?.trim() ?? '',
	continuityApprovalID: process.env.PLAYWRIGHT_CONTINUITY_APPROVAL_ID?.trim() ?? '',
	continuityDocumentID: process.env.PLAYWRIGHT_CONTINUITY_DOCUMENT_ID?.trim() ?? '',
	continuityJournalEntryID: process.env.PLAYWRIGHT_CONTINUITY_JOURNAL_ENTRY_ID?.trim() ?? ''
};

const missingEnv = Object.entries({
	PLAYWRIGHT_ORG_SLUG: env.orgSlug,
	PLAYWRIGHT_EMAIL: env.email,
	PLAYWRIGHT_PASSWORD: env.password,
	PLAYWRIGHT_REQUEST_REFERENCE: env.requestReference,
	PLAYWRIGHT_CONTINUITY_ORG_SLUG: env.continuityOrgSlug,
	PLAYWRIGHT_CONTINUITY_EMAIL: env.continuityEmail,
	PLAYWRIGHT_CONTINUITY_PASSWORD: env.continuityPassword,
	PLAYWRIGHT_CONTINUITY_REQUEST_REFERENCE: env.continuityRequestReference,
	PLAYWRIGHT_CONTINUITY_RECOMMENDATION_ID: env.continuityRecommendationID,
	PLAYWRIGHT_CONTINUITY_APPROVAL_ID: env.continuityApprovalID,
	PLAYWRIGHT_CONTINUITY_DOCUMENT_ID: env.continuityDocumentID,
	PLAYWRIGHT_CONTINUITY_JOURNAL_ENTRY_ID: env.continuityJournalEntryID
})
	.filter(([, value]) => value === '')
	.map(([key]) => key);

type RouteExpectation = {
	path: string;
	heading?: string | RegExp;
	marker?: string | RegExp;
};

const staticRouteExpectations: RouteExpectation[] = [
	{ path: '/app', marker: 'Primary actions' },
	{ path: '/app/routes', heading: 'Searchable route discovery', marker: 'Search' },
	{ path: '/app/settings', heading: 'User-scoped settings and continuity', marker: 'Admin continuity' },
	{ path: '/app/admin', heading: 'Privileged maintenance hub', marker: 'Accounting setup' },
	{ path: '/app/admin/access', heading: 'Access controls', marker: 'Provision user' },
	{ path: '/app/admin/accounting', heading: 'Accounting setup', marker: 'Create ledger account' },
	{ path: '/app/admin/parties', heading: 'Party setup', marker: 'Create party' },
	{ path: '/app/admin/inventory', heading: 'Inventory setup', marker: 'Create inventory item' },
	{ path: '/app/operations', heading: 'Operations landing', marker: 'Queue actions' },
	{ path: '/app/review', heading: 'Review workbench', marker: 'Review surfaces' },
	{ path: '/app/inventory', heading: 'Inventory landing', marker: 'Inventory actions' },
	{ path: '/app/submit-inbound-request', heading: 'Submit inbound request', marker: 'Queue for review' },
	{ path: '/app/operations-feed', heading: 'Durable operations feed' },
	{ path: '/app/agent-chat', heading: 'Coordinator chat', marker: 'Recent coordinator proposals' },
	{ path: '/app/review/inbound-requests', marker: 'Filter' },
	{ path: '/app/review/approvals', marker: 'Approval' },
	{ path: '/app/review/proposals', marker: 'Proposal' },
	{ path: '/app/review/documents', marker: 'Document' },
	{ path: '/app/review/accounting', marker: 'Accounting' },
	{ path: '/app/review/inventory', marker: 'Inventory' },
	{ path: '/app/review/work-orders', marker: 'Work order' },
	{ path: '/app/review/audit', marker: 'Audit' }
];

test.describe.serial('Milestone 13 browser closeout', () => {
	test.skip(
		missingEnv.length > 0,
		`missing required environment for live browser review: ${missingEnv.join(', ')}`
	);

	let partyDetailPath = '';

	test('keeps login, route search, and desktop shell persistence working', async ({ page }) => {
		await page.goto('/app/login');
		await expect(page.getByRole('heading', { name: 'Operator sign-in' })).toBeVisible();

		await fillLoginForm(page);
		await page.getByRole('button', { name: 'Sign in' }).click();

		await expect(page).toHaveURL(/\/app\/?$/);
		await expect(page.locator('aside.sidebar')).not.toHaveClass(/collapsed/);
		await expect(page.getByText('Major areas')).toBeVisible();

		await page.getByRole('button', { name: 'Collapse navigation' }).click();
		await expect(page.locator('aside.sidebar')).toHaveClass(/collapsed/);

		await page.reload();
		await expect(page.locator('aside.sidebar')).toHaveClass(/collapsed/);
		await expect(page.getByRole('button', { name: 'Expand navigation' })).toBeVisible();

		await page.goto('/app/routes');
		await searchRouteCatalog(page, 'pending approvals');
		await expect(page.getByRole('link', { name: /Approval review/ })).toBeVisible();

		await searchRouteCatalog(page, 'failed requests');
		await expect(page.getByRole('link', { name: /Inbound requests review/ })).toBeVisible();
	});

	test('keeps admin maintenance status controls and exact party detail usable', async ({ page }) => {
		await login(page);

		const unique = Date.now().toString();
		const uniqueNumber = Number.parseInt(unique, 10);

		await page.goto('/app/admin/accounting');
		const adminForms = page.locator('form.admin-form');
		const accountForm = adminForms.nth(0);
		const taxForm = adminForms.nth(1);
		const periodForm = adminForms.nth(2);

		const ledgerCode = `PW-AR-${unique}`;
		await accountForm.getByPlaceholder('Code', { exact: true }).fill(ledgerCode);
		await accountForm.getByPlaceholder('Name', { exact: true }).fill(`Playwright receivable ${unique}`);
		await accountForm.locator('select').nth(0).selectOption('liability');
		await accountForm.locator('select').nth(1).selectOption('gst_output');
		await accountForm.getByRole('button', { name: 'Create account' }).click();
		await expect(page.getByText('Ledger account created.')).toBeVisible();

		const ledgerRow = page.locator('tr', { hasText: ledgerCode });
		await expect(ledgerRow).toBeVisible();

		const taxCode = `PW-TAX-${unique}`;
		await taxForm.getByPlaceholder('Code', { exact: true }).fill(taxCode);
		await taxForm.getByPlaceholder('Name', { exact: true }).fill(`Playwright tax ${unique}`);
		await taxForm.getByPlaceholder('Rate basis points').fill('750');
		await taxForm.locator('select').nth(2).selectOption({ index: 1 });
		await taxForm.getByRole('button', { name: 'Create tax code' }).click();
		await expect(page.getByText('Tax code created.')).toBeVisible();

		await ledgerRow.getByRole('button', { name: 'Mark inactive' }).click();
		await expect(page.getByText('Ledger account marked inactive.')).toBeVisible();

		const taxRow = page.locator('tr', { hasText: taxCode });
		await expect(taxRow).toBeVisible();
		await taxRow.getByRole('button', { name: 'Mark inactive' }).click();
		await expect(page.getByText('Tax code marked inactive.')).toBeVisible();

		const periodCode = `PW-${unique}`;
		const periodYear = 2030 + (uniqueNumber % 20);
		const periodMonth = String((uniqueNumber % 12) + 1).padStart(2, '0');
		const periodStart = `${periodYear}-${periodMonth}-01`;
		const periodEnd = `${periodYear}-${periodMonth}-28`;
		await periodForm.getByPlaceholder('Period code').fill(periodCode);
		await periodForm.locator('input[type="date"]').nth(0).fill(periodStart);
		await periodForm.locator('input[type="date"]').nth(1).fill(periodEnd);
		await periodForm.getByRole('button', { name: 'Create period' }).click();
		await expect(page.getByText('Accounting period created.')).toBeVisible();

		const periodRow = page.locator('tr', { hasText: periodCode });
		await expect(periodRow).toBeVisible();
		await periodRow.getByRole('button', { name: 'Close period' }).click();
		await expect(page.getByText('Accounting period closed.')).toBeVisible();

		await page.goto('/app/admin/parties');
		const partyCode = `PW-PARTY-${unique}`;
		await page.getByPlaceholder('Party code').fill(partyCode);
		await page.getByPlaceholder('Display name').fill(`Playwright Party ${unique}`);
		await page.getByPlaceholder('Legal name').fill(`Playwright Party Legal ${unique}`);
		await page.getByRole('button', { name: 'Create party' }).click();
		await expect(page.getByText('Party created.')).toBeVisible();

		const partyRow = page.locator('tr', { hasText: partyCode });
		await expect(partyRow).toBeVisible();
		await partyRow.getByRole('link', { name: 'Open detail' }).click();
		await expect(page.getByRole('button', { name: 'Create contact' })).toBeVisible();
		partyDetailPath = new URL(page.url()).pathname;

		await page.getByPlaceholder('Full name').fill(`Playwright Contact ${unique}`);
		await page.getByPlaceholder('Role title').fill('Dispatch lead');
		await page.getByPlaceholder('Email').fill(`playwright-${unique}@example.com`);
		await page.getByPlaceholder('Phone').fill('555-0100');
		await page.getByRole('checkbox', { name: 'Primary contact' }).check();
		await page.getByRole('button', { name: 'Create contact' }).click();
		await expect(page.getByText('Party contact created.')).toBeVisible();
		await expect(page.getByText(`Playwright Contact ${unique}`)).toBeVisible();

		await page.getByRole('button', { name: 'Mark inactive' }).click();
		await expect(page.getByText('Party marked inactive.')).toBeVisible();

		await page.goto('/app/admin/inventory');
		const itemSku = `PW-SKU-${unique}`;
		await page.getByPlaceholder('SKU').fill(itemSku);
		await page.getByPlaceholder('Name', { exact: true }).nth(0).fill(`Playwright Item ${unique}`);
		await page.getByRole('button', { name: 'Create item' }).click();
		await expect(page.getByText('Inventory item created.')).toBeVisible();

		const itemRow = page.locator('tr', { hasText: itemSku });
		await expect(itemRow).toBeVisible();
		await itemRow.getByRole('button', { name: 'Mark inactive' }).click();
		await expect(page.getByText('Inventory item marked inactive.')).toBeVisible();

		const locationCode = `PW-LOC-${unique}`;
		await page.getByPlaceholder('Code', { exact: true }).fill(locationCode);
		await page.getByPlaceholder('Name', { exact: true }).nth(1).fill(`Playwright Location ${unique}`);
		await page.getByRole('button', { name: 'Create location' }).click();
		await expect(page.getByText('Inventory location created.')).toBeVisible();

		const locationRow = page.locator('tr', { hasText: locationCode });
		await expect(locationRow).toBeVisible();
		await locationRow.getByRole('button', { name: 'Mark inactive' }).click();
		await expect(page.getByText('Inventory location marked inactive.')).toBeVisible();
	});

	test('renders the promoted desktop route family on the served Svelte runtime', async ({ page }) => {
		await login(page);

		for (const route of staticRouteExpectations) {
			await assertRoute(page, route);
		}

		await assertRoute(page, {
			path: `/app/inbound-requests/${encodeURIComponent(env.requestReference)}`,
			heading: env.requestReference,
			marker: 'Workflow continuity'
		});

		expect(
			partyDetailPath,
			'expected the earlier admin maintenance test to capture a party detail route'
		).not.toBe('');
		await assertRoute(page, {
			path: partyDetailPath,
			marker: 'Create contact'
		});
	});

	test('keeps exact request, proposal, approval, document, and accounting continuity intact', async ({ page }) => {
		await page.context().clearCookies();
		await login(page, {
			orgSlug: env.continuityOrgSlug,
			email: env.continuityEmail,
			password: env.continuityPassword
		});

		await page.goto(`/app/inbound-requests/${encodeURIComponent(env.continuityRequestReference)}`);
		await expect(page.getByRole('heading', { name: env.continuityRequestReference })).toBeVisible();
		await expect(page.getByText('Workflow continuity')).toBeVisible();
		await expect(page.getByRole('link', { name: 'Open latest proposal' })).toBeVisible();

		await page.goto(`/app/review/proposals/${encodeURIComponent(env.continuityRecommendationID)}`);
		await expect(page).toHaveURL(
			new RegExp(`/app/review/proposals/${escapeRegExp(env.continuityRecommendationID)}$`)
		);
		await expect(page.getByRole('heading', { name: env.continuityRecommendationID })).toBeVisible();
		await expect(page.getByRole('link', { name: /Accounting entry/ })).toBeVisible();

		await page.getByRole('link', { name: 'Approval detail' }).click();
		await expect(page).toHaveURL(
			new RegExp(`/app/review/approvals/${escapeRegExp(env.continuityApprovalID)}$`)
		);
		await expect(page.getByRole('heading', { name: env.continuityApprovalID })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Document detail' })).toBeVisible();

		await page.getByRole('link', { name: 'Document detail' }).click();
		await expect(page).toHaveURL(
			new RegExp(`/app/review/documents/${escapeRegExp(env.continuityDocumentID)}$`)
		);
		await expect(page.getByRole('heading', { name: /Verify agent continuity document/ })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Accounting detail' })).toBeVisible();

		await page.getByRole('link', { name: 'Accounting detail' }).click();
		await expect(page).toHaveURL(
			new RegExp(`/app/review/accounting/${escapeRegExp(env.continuityJournalEntryID)}$`)
		);
		await expect(page.getByRole('heading', { name: /Journal / })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Inbound request' })).toBeVisible();
	});
});

async function login(
	page: Page,
	credentials: { orgSlug: string; email: string; password: string } = {
		orgSlug: env.orgSlug,
		email: env.email,
		password: env.password
	}
): Promise<void> {
	await page.goto('/app/login');
	await expect(page.getByRole('heading', { name: 'Operator sign-in' })).toBeVisible();
	await fillLoginForm(page, credentials);
	await page.getByRole('button', { name: 'Sign in' }).click();
	await expect(page).toHaveURL(/\/app\/?$/);
	await expect(page.locator('main h1')).toBeVisible();
}

async function fillLoginForm(
	page: Page,
	credentials: { orgSlug: string; email: string; password: string } = {
		orgSlug: env.orgSlug,
		email: env.email,
		password: env.password
	}
): Promise<void> {
	await page.getByLabel('Org slug').fill(credentials.orgSlug);
	await page.getByLabel('Email').fill(credentials.email);
	await page.getByLabel('Password').fill(credentials.password);
}

async function searchRouteCatalog(page: Page, query: string): Promise<void> {
	const searchInput = page.getByPlaceholder('Search requests, approvals, inventory, admin...');
	await searchInput.fill(query);
	await page.getByRole('button', { name: 'Search' }).click();
	await expect(searchInput).toHaveValue(query);
}

async function assertRoute(page: Page, route: RouteExpectation): Promise<void> {
	await page.goto(route.path);
	await expect(page).not.toHaveURL(/\/app\/login/);
	if (route.heading) {
		await expect(page.getByRole('heading', { name: route.heading })).toBeVisible();
	} else {
		await expect(page.locator('main h1')).toBeVisible();
	}
	if (route.marker) {
		await expect(page.getByText(route.marker, { exact: false }).first()).toBeVisible();
	}
}

function escapeRegExp(value: string): string {
	return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
