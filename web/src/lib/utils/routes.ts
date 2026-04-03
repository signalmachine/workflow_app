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
	operations: withBase('/operations'),
	review: withBase('/review'),
	inventory: withBase('/inventory'),
	settings: withBase('/settings'),
	admin: withBase('/admin'),
	adminAccounting: withBase('/admin/accounting'),
	adminParties: withBase('/admin/parties'),
	adminAccess: withBase('/admin/access'),
	adminInventory: withBase('/admin/inventory')
};
