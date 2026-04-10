import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import SubmitInboundRequestPage from './+page.svelte';

describe('submit inbound request page', () => {
	it('keeps draft and queue actions visible on the persisted request seam', () => {
		render(SubmitInboundRequestPage, { props: {} as never });

		expect(screen.getByText('Submit inbound request')).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Queue for review' })).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Save as draft' })).toBeTruthy();
		expect(screen.getByPlaceholderText('Describe the issue or requested work...')).toBeTruthy();
	});
});
