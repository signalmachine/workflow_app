import { describe, expect, it } from 'vitest';

import { readDesktopSidebarCollapsed } from './shell';

class MemoryStorage {
	private readonly values = new Map<string, string>();

	getItem(key: string): string | null {
		return this.values.get(key) ?? null;
	}

	setItem(key: string, value: string): void {
		this.values.set(key, value);
	}
}

describe('desktop sidebar preference', () => {
	it('defaults to expanded when storage is unavailable', () => {
		expect(readDesktopSidebarCollapsed(undefined)).toBe(false);
	});

	it('reads persisted collapsed and expanded values', () => {
		const storage = new MemoryStorage();

		storage.setItem('workflow_app.desktop_sidebar_collapsed', 'true');
		expect(readDesktopSidebarCollapsed(storage)).toBe(true);

		storage.setItem('workflow_app.desktop_sidebar_collapsed', 'false');
		expect(readDesktopSidebarCollapsed(storage)).toBe(false);
	});
});
