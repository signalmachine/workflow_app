import type { HomeAction, SessionContext } from '$lib/api/types';
import { adminRoleCode } from '$lib/utils/admin';
import { routes } from '$lib/utils/routes';

export interface SettingsViewModel {
	settingsPrinciples: string[];
	personalUtilityLinks: HomeAction[];
	adminContinuation?: HomeAction;
}

export function buildSettingsViewModel(session: SessionContext): SettingsViewModel {
	return {
		settingsPrinciples: [
			'Settings stays user-scoped: session context, personal continuity, and safe workflow shortcuts belong here.',
			'Org-scoped maintenance, access-sensitive setup, and governed controls belong under Admin for authorized actors.',
			'Workflow pages remain the primary working surfaces; utility pages should route back into exact operational or review paths.'
		],
		personalUtilityLinks: [
			{
				title: 'Open route catalog',
				summary: 'Search grouped route families when the next workflow surface is not obvious from the shell.',
				href: routes.routeCatalog
			},
			{
				title: 'Open home',
				summary: 'Return to the role-aware home surface for workload-prioritized entry points.',
				href: routes.home
			}
		],
		adminContinuation:
			session.role_code === adminRoleCode
				? {
						title: 'Open admin maintenance hub',
						summary: 'Use the separate admin surface for org-scoped setup families, governance review, and later privileged controls.',
						href: routes.admin
					}
				: undefined
	};
}
