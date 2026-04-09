import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import InventoryReviewPage from './+page.svelte';

const baseData = {
	filters: {
		movementID: '',
		itemID: '',
		locationID: '',
		documentID: '',
		movementType: '',
		onlyPendingAccounting: false,
		onlyPendingExecution: false
	},
	stock: [],
	movements: [],
	reconciliation: []
};

describe('inventory review page', () => {
	it('surfaces active pending-handoff scope from inventory landing links', () => {
		render(InventoryReviewPage, {
			props: {
				data: {
					...baseData,
					filters: {
						...baseData.filters,
						onlyPendingExecution: true,
						onlyPendingAccounting: true
					}
				}
			} as never
		});

		expect(screen.getByText('Active reconciliation scope')).toBeTruthy();
		expect(screen.getByLabelText('Pending execution only')).toHaveProperty('checked', true);
		expect(screen.getByLabelText('Pending accounting only')).toHaveProperty('checked', true);
		expect(screen.getByRole('link', { name: 'Pending execution handoffs' }).getAttribute('href')).toBe(
			'/app/review/inventory?only_pending_execution=true'
		);
		expect(screen.getByRole('link', { name: 'Pending accounting handoffs' }).getAttribute('href')).toBe(
			'/app/review/inventory?only_pending_accounting=true'
		);
		expect(screen.getByText(/showing pending execution handoffs and pending accounting handoffs/i)).toBeTruthy();
	});

	it('shows a scoped empty state when no reconciliation rows match', () => {
		render(InventoryReviewPage, { props: { data: baseData } as never });

		expect(screen.getAllByText('No reconciliation rows match the current inventory review scope.').length).toBeGreaterThan(0);
	});
});
