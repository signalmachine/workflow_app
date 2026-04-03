import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			pages: '../internal/app/web_dist',
			assets: '../internal/app/web_dist',
			fallback: '200.html'
		}),
		paths: {
			base: '/app'
		}
	}
};

export default config;
