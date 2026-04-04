import { describe, expect, it } from 'vitest';

import type { SessionContext } from '$lib/api/types';

import { buildSettingsViewModel } from './settings';

function makeSession(roleCode: string): SessionContext {
	return {
		session_id: 'session-123',
		org_id: 'org-123',
		org_slug: 'acme',
		org_name: 'Acme Corp',
		user_id: 'user-123',
		user_email: 'operator@example.com',
		user_display_name: 'Alex Operator',
		role_code: roleCode,
		device_label: 'Browser',
		expires_at: '2026-04-04T12:00:00Z',
		issued_at: '2026-04-04T10:00:00Z',
		last_seen_at: '2026-04-04T10:15:00Z'
	};
}

describe('buildSettingsViewModel', () => {
	it('includes admin continuation for admin actors', () => {
		const model = buildSettingsViewModel(makeSession('admin'));

		expect(model.adminContinuation).toMatchObject({
			title: 'Open admin maintenance hub',
			href: '/app/admin'
		});
		expect(model.personalUtilityLinks).toHaveLength(2);
		expect(model.settingsPrinciples).toHaveLength(3);
	});

	it('keeps admin continuation hidden for non-admin actors', () => {
		const model = buildSettingsViewModel(makeSession('operator'));

		expect(model.adminContinuation).toBeUndefined();
	});
});
