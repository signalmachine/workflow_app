import { defineConfig } from '@playwright/test';

const outputDir = process.env.PLAYWRIGHT_OUTPUT_DIR ?? '/tmp/workflow_app_playwright';

export default defineConfig({
	testDir: './playwright',
	timeout: 60_000,
	fullyParallel: false,
	reporter: [['list']],
	outputDir,
	use: {
		baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:18080',
		headless: true,
		viewport: { width: 1400, height: 1100 },
		trace: 'retain-on-failure',
		screenshot: 'only-on-failure',
		video: 'retain-on-failure'
	}
});
