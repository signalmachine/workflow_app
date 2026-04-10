import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import LoginPage from './+page.svelte';

describe('login page', () => {
	it('keeps the sign-in surface thin and centered on the shared session seam', () => {
		render(LoginPage, { props: {} as never });

		expect(screen.getByRole('heading', { name: 'Operator sign-in' })).toBeTruthy();
		expect(
			screen.getByText(/POST \/api\/session\/login/, {
				exact: false
			})
		).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Sign in' })).toBeTruthy();
	});
});
