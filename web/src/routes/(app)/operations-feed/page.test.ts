import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import OperationsFeedPage from './+page.svelte';

const baseData = {
	snapshot: {
		items: [
			{
				kind: 'approval',
				title: 'Approval queued',
				status: 'pending',
				summary: 'Decision work is waiting in the shared queue.',
				primary_href: '/app/review/approvals/approval-1',
				primary_label: 'Open approval',
				secondary_href: '/app/review/documents/document-1',
				secondary_label: 'Open document',
				occurred_at: '2026-04-10T09:00:00Z'
			}
		]
	}
};

describe('operations feed page', () => {
	it('keeps recent workflow movement wired to exact downstream review routes', () => {
		render(OperationsFeedPage, { props: { data: baseData } as never });

		expect(screen.getByText('Durable operations feed')).toBeTruthy();
		expect(screen.getByRole('link', { name: 'Open approval' }).getAttribute('href')).toBe(
			'/app/review/approvals/approval-1'
		);
		expect(screen.getByRole('link', { name: 'Open document' }).getAttribute('href')).toBe(
			'/app/review/documents/document-1'
		);
	});
});
