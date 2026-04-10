import { routes } from '$lib/utils/routes';

export interface NavigationTab {
	id: string;
	label: string;
	href: string;
	matchers: NavigationMatcher[];
}

export interface NavigationArea {
	id: string;
	label: string;
	href: string;
	description: string;
	requiresAdmin?: boolean;
	matchers: NavigationMatcher[];
	tabs: NavigationTab[];
}

export interface NavigationModel {
	activeArea: NavigationArea;
	areas: NavigationArea[];
}

export interface NavigationMatcher {
	path: string;
	exact?: boolean;
}

function exact(path: string): NavigationMatcher {
	return { path, exact: true };
}

function prefix(path: string): NavigationMatcher {
	return { path };
}

function matchesPath(pathname: string, matcher: NavigationMatcher): boolean {
	const path = pathname.trim();
	const candidate = matcher.path.trim();
	if (candidate === '') {
		return false;
	}
	if (matcher.exact || candidate === routes.home) {
		return path === candidate;
	}
	return path === candidate || path.startsWith(`${candidate}/`);
}

export function isNavigationItemActive(pathname: string, matchers: NavigationMatcher[]): boolean {
	return matchers.some((matcher) => matchesPath(pathname, matcher));
}

function buildNavigationAreas(roleCode: string): NavigationArea[] {
	const areas: NavigationArea[] = [
		{
			id: 'agent',
			label: 'Agent',
			href: routes.agentChat,
			description: 'Coordinator messages, requests, and AI-generated proposals.',
			matchers: [
				prefix(routes.agentChat),
				prefix(routes.operationsFeed),
				prefix(routes.inboundRequests),
				prefix(routes.reviewProposals)
			],
			tabs: [
				{
					id: 'messages',
					label: 'Messages',
					href: routes.operationsFeed,
					matchers: [prefix(routes.operationsFeed)]
				},
				{
					id: 'requests',
					label: 'Requests',
					href: routes.agentChat,
					matchers: [prefix(routes.agentChat), prefix(routes.inboundRequests)]
				},
				{
					id: 'lists',
					label: 'Lists',
					href: routes.reviewProposals,
					matchers: [prefix(routes.reviewProposals)]
				}
			]
		},
		{
			id: 'accounting',
			label: 'Accounting',
			href: routes.reviewAccounting,
			description: 'Approvals, documents, journal review, and accounting reports.',
			matchers: [exact(routes.review), prefix(routes.reviewApprovals), prefix(routes.reviewDocuments), prefix(routes.reviewAccounting)],
			tabs: [
				{
					id: 'overview',
					label: 'Overview',
					href: routes.review,
					matchers: [exact(routes.review)]
				},
				{
					id: 'workflows',
					label: 'Workflows',
					href: routes.reviewApprovals,
					matchers: [prefix(routes.reviewApprovals)]
				},
				{
					id: 'lists',
					label: 'Lists',
					href: routes.reviewDocuments,
					matchers: [prefix(routes.reviewDocuments)]
				},
				{
					id: 'reports',
					label: 'Reports',
					href: routes.reviewAccounting,
					matchers: [prefix(routes.reviewAccounting)]
				}
			]
		},
		{
			id: 'inventory',
			label: 'Inventory',
			href: routes.inventory,
			description: 'Stock, movement, and reconciliation work on shared inventory truth.',
			matchers: [prefix(routes.inventory), prefix(routes.reviewInventory)],
			tabs: [
				{
					id: 'overview',
					label: 'Overview',
					href: routes.inventory,
					matchers: [prefix(routes.inventory)]
				},
				{
					id: 'workflows',
					label: 'Workflows',
					href: routes.reviewInventory,
					matchers: [prefix(routes.reviewInventory)]
				}
			]
		},
		{
			id: 'operations',
			label: 'Operations',
			href: routes.home,
			description: 'Workflow entry, execution continuity, and route discovery.',
			matchers: [
				exact(routes.home),
				prefix(routes.operations),
				prefix(routes.submitInboundRequest),
				prefix(routes.reviewInboundRequests),
				prefix(routes.reviewWorkOrders),
				prefix(routes.reviewAudit),
				prefix(routes.routeCatalog)
			],
			tabs: [
				{
					id: 'overview',
					label: 'Overview',
					href: routes.home,
					matchers: [exact(routes.home)]
				},
				{
					id: 'workflows',
					label: 'Workflows',
					href: routes.operations,
					matchers: [prefix(routes.operations)]
				},
				{
					id: 'actions',
					label: 'Actions',
					href: routes.submitInboundRequest,
					matchers: [prefix(routes.submitInboundRequest)]
				},
				{
					id: 'lists',
					label: 'Lists',
					href: routes.reviewInboundRequests,
					matchers: [prefix(routes.reviewInboundRequests), prefix(routes.reviewWorkOrders)]
				},
				{
					id: 'reports',
					label: 'Reports',
					href: routes.reviewAudit,
					matchers: [prefix(routes.reviewAudit)]
				},
				{
					id: 'search',
					label: 'Search',
					href: routes.routeCatalog,
					matchers: [prefix(routes.routeCatalog)]
				}
			]
		},
		{
			id: 'settings',
			label: 'Settings',
			href: routes.settings,
			description: 'Session context, ownership guidance, and personal continuity.',
			matchers: [exact(routes.settings)],
			tabs: [
				{
					id: 'overview',
					label: 'Overview',
					href: routes.settings,
					matchers: [exact(routes.settings)]
				}
			]
		}
	];

	if (roleCode === 'admin') {
		areas.push({
			id: 'admin',
			label: 'Admin',
			href: routes.admin,
			description: 'Privileged maintenance for accounting, parties, access, and inventory.',
			requiresAdmin: true,
			matchers: [prefix(routes.admin)],
			tabs: [
				{
					id: 'overview',
					label: 'Overview',
					href: routes.admin,
					matchers: [exact(routes.admin)]
				},
				{
					id: 'master-data',
					label: 'Master Data',
					href: routes.adminMasterData,
					matchers: [prefix(routes.adminMasterData), prefix(routes.adminAccounting), prefix(routes.adminParties), prefix(routes.adminInventory)]
				},
				{
					id: 'lists',
					label: 'Lists',
					href: routes.adminLists,
					matchers: [prefix(routes.adminLists)]
				},
				{
					id: 'access',
					label: 'Access',
					href: routes.adminAccess,
					matchers: [prefix(routes.adminAccess)]
				}
			]
		});
	}

	return areas;
}

export function getNavigationModel(pathname: string, roleCode: string): NavigationModel {
	const areas = buildNavigationAreas(roleCode);
	const activeArea =
		areas.find((area) => isNavigationItemActive(pathname, area.matchers)) ??
		areas.find((area) => area.id === 'operations') ??
		areas[0];

	return {
		activeArea,
		areas
	};
}
