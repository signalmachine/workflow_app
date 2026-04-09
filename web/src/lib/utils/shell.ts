const desktopSidebarPreferenceKey = 'workflow_app.desktop_sidebar_collapsed';

interface StorageLike {
	getItem(key: string): string | null;
	setItem(key: string, value: string): void;
}

export function readDesktopSidebarCollapsed(storage: StorageLike | undefined): boolean {
	if (!storage) {
		return false;
	}

	return storage.getItem(desktopSidebarPreferenceKey) === 'true';
}
