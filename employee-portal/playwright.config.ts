import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright Configuration for IRIS Employee Portal E2E Tests
 */

const BASE_URL = process.env.BASE_URL || 'http://localhost:3001';
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

// Validate localhost only
if (!BASE_URL.includes('localhost') && !BASE_URL.includes('127.0.0.1')) {
  throw new Error(
    `SECURITY VIOLATION: Tests can only run against localhost. Current BASE_URL: ${BASE_URL}`
  );
}

export default defineConfig({
  testDir: './tests',

  // Test execution settings
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 2, // Retry flaky tests
  workers: 1,

  // Reporter configuration
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
    ['json', { outputFile: 'test-results/test-results.json' }]
  ],

  // Global settings
  use: {
    baseURL: BASE_URL,

    // Trace and screenshot on failure
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',

    // Browser settings
    viewport: { width: 1280, height: 720 },
    ignoreHTTPSErrors: true,

    // Timeouts
    actionTimeout: 10000,
    navigationTimeout: 15000,
  },

  // Timeout settings
  timeout: 120000,
  expect: {
    timeout: 5000,
  },

  // Projects for different browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
