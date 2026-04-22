import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  outputDir: './e2e/test-results',
  globalSetup: './e2e/global-setup.ts',
  globalTeardown: './e2e/global-teardown.ts',
  use: {
    baseURL: process.env.PW_BASE_URL ?? 'http://localhost:8081',
    locale: 'ru-RU',
    headless: true,
    screenshot: 'only-on-failure',
  },
  reporter: [['list'], ['html', { open: 'never', outputFolder: './e2e/test-report' }]],
});
