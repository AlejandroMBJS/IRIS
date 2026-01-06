import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright Configuration for IRIS Monkey/Crawler E2E Tests
 *
 * SECURITY NOTICE:
 * - This config is ONLY for local/dev environments (localhost)
 * - Tests will automatically abort if baseURL is not localhost
 * - Dangerous routes and actions are blocked by the test implementation
 */

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

// Validate localhost only
if (!BASE_URL.includes('localhost') && !BASE_URL.includes('127.0.0.1')) {
  throw new Error(
    `❌ SECURITY VIOLATION: Tests can only run against localhost. Current BASE_URL: ${BASE_URL}`
  );
}

export default defineConfig({
  testDir: './tests',

  // Test execution settings
  fullyParallel: false, // Run tests sequentially for monkey test
  forbidOnly: !!process.env.CI,
  retries: 2, // Retry flaky tests
  workers: 1, // Single worker for monkey test to avoid conflicts

  // Reporter configuration
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
    ['json', { outputFile: 'test-results/monkey-test-results.json' }]
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
  timeout: 120000, // 2 minutes per test
  expect: {
    timeout: 5000,
  },

  // Projects for different browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    // Uncomment to test on other browsers
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],

  // Dev server (optional - use if you want Playwright to start Next.js)
  // webServer: {
  //   command: 'npm run dev',
  //   url: BASE_URL,
  //   reuseExistingServer: !process.env.CI,
  //   timeout: 120000,
  // },
});
