/**
 * IRIS Payroll - Monkey/Crawler E2E Test
 *
 * PURPOSE:
 * - Automatically navigate through the app like a human user
 * - Click interactive elements, fill forms, explore pages
 * - EXCLUDE dangerous/destructive actions (admin, delete, payments, etc.)
 *
 * SAFETY GUARANTEES:
 * - Blocks ALL DELETE requests
 * - Blocks admin/permission routes
 * - Blocks payment/billing endpoints
 * - Blocks dangerous UI actions (delete buttons, danger zones)
 * - Only runs on localhost
 *
 * USAGE:
 *   npm run test:e2e         # Headless mode
 *   npm run test:e2e:headed  # With browser UI
 */

import { test, expect, Page, Route } from '@playwright/test';

// ============================================================================
// CONFIGURATION
// ============================================================================

const CONFIG = {
  MAX_STEPS: 100, // Maximum number of actions before stopping
  MAX_PAGES: 20, // Maximum number of different pages to visit
  MAX_CLICKS_PER_PAGE: 10, // Maximum clicks on same page
  NAVIGATION_DELAY: 500, // ms to wait between actions
  FORM_FILL_PROBABILITY: 0.7, // 70% chance to fill a form
};

// ============================================================================
// DANGEROUS PATTERNS (DENYLISTS)
// ============================================================================

// Backend endpoint patterns to BLOCK
const DANGEROUS_ENDPOINT_PATTERNS = [
  // Admin & Permissions
  /\/admin\//i,
  /\/permission/i,
  /\/role-inheritance/i,
  /\/users\/\d+$/i, // User management endpoints

  // Destructive actions
  /\/delete/i,
  /\/destroy/i,
  /\/drop/i,
  /\/purge/i,
  /\/reset/i,
  /\/seed/i,
  /\/migrate/i,

  // Financial/Billing
  /\/billing/i,
  /\/payment/i,
  /\/invoice/i,
  /\/reimburse/i,
  /\/mark-paid/i,
  /\/issue.*payment/i,
  /\/advance-payments\/\d+\/(approve|issue)/i,

  // Termination actions
  /\/terminate/i,
  /\/decline/i,
  /\/reject/i,
  /\/archive/i,
  /\/revoke/i,

  // Sensitive operations
  /\/webhook/i,
  /\/secret/i,
  /\/key/i,
  /\/oauth/i,
];

// Frontend route patterns to AVOID
const DANGEROUS_ROUTE_PATTERNS = [
  /\/admin\//i,
  /\/configuration\/role-inheritance/i,
  /\/configuration\/permissions/i,
  /\/billing/i,
  /\/payment/i,
  /\/subscription/i,
];

// UI element text patterns to AVOID clicking
const DANGEROUS_UI_TEXT_PATTERNS = [
  /delete/i,
  /remove/i,
  /eliminar/i,
  /borrar/i,
  /destroy/i,
  /drop/i,
  /purge/i,
  /reset/i,
  /wipe/i,
  /pagar/i,
  /payment/i,
  /subscribe/i,
  /cancel.*subscription/i,
  /refund/i,
  /terminar/i,
  /rechazar/i,
];

// UI class patterns to AVOID (danger buttons)
const DANGEROUS_UI_CLASS_PATTERNS = [
  /\bdanger\b/i,
  /\bdestructive\b/i,
  /\bbg-red-/i,
  /\btext-red-/i,
  /\bdelete/i,
];

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Check if a URL/path is dangerous
 */
function isDangerousRoute(url: string): boolean {
  const path = new URL(url).pathname;
  return DANGEROUS_ROUTE_PATTERNS.some((pattern) => pattern.test(path));
}

/**
 * Check if an API endpoint is dangerous
 */
function isDangerousEndpoint(url: string, method: string): boolean {
  // Block ALL DELETE requests
  if (method === 'DELETE') {
    return true;
  }

  // Check against patterns
  return DANGEROUS_ENDPOINT_PATTERNS.some((pattern) => pattern.test(url));
}

/**
 * Check if UI element is dangerous to click
 */
async function isDangerousElement(locator: any): Promise<boolean> {
  try {
    // Check text content
    const text = await locator.textContent().catch(() => '');
    if (text && DANGEROUS_UI_TEXT_PATTERNS.some((p) => p.test(text))) {
      return true;
    }

    // Check aria-label
    const ariaLabel = await locator.getAttribute('aria-label').catch(() => '');
    if (ariaLabel && DANGEROUS_UI_TEXT_PATTERNS.some((p) => p.test(ariaLabel))) {
      return true;
    }

    // Check data-testid
    const testId = await locator.getAttribute('data-testid').catch(() => '');
    if (testId && DANGEROUS_UI_TEXT_PATTERNS.some((p) => p.test(testId))) {
      return true;
    }

    // Check classes
    const className = await locator.getAttribute('class').catch(() => '');
    if (className && DANGEROUS_UI_CLASS_PATTERNS.some((p) => p.test(className))) {
      return true;
    }

    // Check href for dangerous routes
    const href = await locator.getAttribute('href').catch(() => '');
    if (href && href.startsWith('/')) {
      const fullUrl = `${process.env.BASE_URL || 'http://localhost:3000'}${href}`;
      if (isDangerousRoute(fullUrl)) {
        return true;
      }
    }

    return false;
  } catch {
    return false;
  }
}

/**
 * Generate dummy data for form inputs
 */
function getDummyValue(inputType: string, name: string): string {
  const lowerName = name.toLowerCase();

  if (lowerName.includes('email')) {
    return `test.${Date.now()}@example.com`;
  }
  if (lowerName.includes('password')) {
    return 'TestPassword123!';
  }
  if (lowerName.includes('phone') || lowerName.includes('tel')) {
    return '5551234567';
  }
  if (lowerName.includes('name')) {
    return `Test User ${Date.now()}`;
  }
  if (lowerName.includes('date')) {
    return '2024-01-15';
  }
  if (lowerName.includes('number') || lowerName.includes('amount')) {
    return '100';
  }

  switch (inputType) {
    case 'email':
      return `test.${Date.now()}@example.com`;
    case 'password':
      return 'TestPassword123!';
    case 'number':
      return '42';
    case 'tel':
      return '5551234567';
    case 'date':
      return '2024-01-15';
    case 'time':
      return '10:00';
    case 'url':
      return 'https://example.com';
    default:
      return `Test ${Date.now()}`;
  }
}

/**
 * Check if user is on login page
 */
async function isLoginPage(page: Page): boolean {
  const url = page.url();
  return url.includes('/auth/login') || url.includes('/login');
}

/**
 * Attempt login with credentials from environment
 */
async function attemptLogin(page: Page): Promise<boolean> {
  const email = process.env.E2E_EMAIL;
  const password = process.env.E2E_PASSWORD;

  if (!email || !password) {
    console.log('⚠️  No E2E_EMAIL or E2E_PASSWORD provided - continuing as guest');
    return false;
  }

  try {
    // Find password input (indicates login form)
    const passwordInput = page.locator('input[type="password"]').first();
    const emailInput = page.locator('input[type="email"], input[name="email"]').first();

    if ((await emailInput.count()) === 0 || (await passwordInput.count()) === 0) {
      console.log('⚠️  Login form not found - continuing as guest');
      return false;
    }

    console.log('🔐 Attempting login...');
    await emailInput.fill(email);
    await passwordInput.fill(password);

    // Find and click submit button
    const submitButton = page.locator('button[type="submit"]').first();
    await submitButton.click();

    // Wait for navigation or error
    await page.waitForTimeout(2000);

    // Check if still on login page (login failed)
    if (await isLoginPage(page)) {
      console.log('❌ Login failed - continuing as guest');
      return false;
    }

    console.log('✅ Login successful');
    return true;
  } catch (error) {
    console.log(`❌ Login error: ${error} - continuing as guest`);
    return false;
  }
}

// ============================================================================
// MONKEY TEST REPORT
// ============================================================================

interface MonkeyReport {
  pagesVisited: Set<string>;
  clicksExecuted: number;
  clicksBlocked: number;
  requestsBlocked: number;
  blockedActions: Array<{ type: string; reason: string; target: string }>;
  errors: Array<{ type: string; message: string; url: string }>;
}

function createReport(): MonkeyReport {
  return {
    pagesVisited: new Set(),
    clicksExecuted: 0,
    clicksBlocked: 0,
    requestsBlocked: 0,
    blockedActions: [],
    errors: [],
  };
}

function printReport(report: MonkeyReport): void {
  console.log('\n' + '='.repeat(80));
  console.log('🐵 MONKEY TEST REPORT');
  console.log('='.repeat(80));

  console.log('\n📊 STATISTICS:');
  console.log(`  Pages visited: ${report.pagesVisited.size}`);
  console.log(`  Clicks executed: ${report.clicksExecuted}`);
  console.log(`  Clicks blocked (dangerous): ${report.clicksBlocked}`);
  console.log(`  Requests blocked: ${report.requestsBlocked}`);
  console.log(`  Errors detected: ${report.errors.length}`);

  console.log('\n📄 PAGES VISITED:');
  Array.from(report.pagesVisited).forEach((url, i) => {
    console.log(`  ${i + 1}. ${url}`);
  });

  if (report.blockedActions.length > 0) {
    console.log('\n🛡️  BLOCKED ACTIONS (dangerous):');
    report.blockedActions.slice(0, 10).forEach((action, i) => {
      console.log(`  ${i + 1}. [${action.type}] ${action.reason}`);
      console.log(`     Target: ${action.target}`);
    });
    if (report.blockedActions.length > 10) {
      console.log(`  ... and ${report.blockedActions.length - 10} more`);
    }
  }

  if (report.errors.length > 0) {
    console.log('\n❌ ERRORS:');
    report.errors.forEach((error, i) => {
      console.log(`  ${i + 1}. [${error.type}] ${error.message}`);
      console.log(`     URL: ${error.url}`);
    });
  }

  console.log('\n' + '='.repeat(80) + '\n');
}

// ============================================================================
// MAIN MONKEY TEST
// ============================================================================

test.describe('Monkey/Crawler E2E Test', () => {
  let report: MonkeyReport;

  test.beforeEach(async ({ page, baseURL }) => {
    report = createReport();

    // ========================================================================
    // SECURITY CHECK: Only localhost
    // ========================================================================
    if (!baseURL?.includes('localhost') && !baseURL?.includes('127.0.0.1')) {
      throw new Error(`❌ SECURITY VIOLATION: Tests can only run on localhost. Current: ${baseURL}`);
    }

    // ========================================================================
    // NETWORK INTERCEPTION: Block dangerous requests
    // ========================================================================
    await page.route('**/*', async (route: Route) => {
      const request = route.request();
      const url = request.url();
      const method = request.method();

      // Allow static assets
      if (
        url.match(/\.(css|js|png|jpg|jpeg|gif|svg|woff|woff2|ttf|ico)$/) ||
        url.includes('/_next/')
      ) {
        return route.continue();
      }

      // Check if endpoint is dangerous
      if (isDangerousEndpoint(url, method)) {
        report.requestsBlocked++;
        report.blockedActions.push({
          type: 'REQUEST_BLOCKED',
          reason: `Dangerous ${method} request`,
          target: url,
        });

        console.log(`🛡️  BLOCKED: ${method} ${url}`);

        // Mock 403 response
        return route.fulfill({
          status: 403,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'Blocked by E2E safety guard',
            reason: 'Dangerous endpoint detected',
          }),
        });
      }

      // Continue safe requests
      return route.continue();
    });

    // ========================================================================
    // ERROR DETECTION: Catch console errors and page errors
    // ========================================================================
    page.on('pageerror', (error) => {
      report.errors.push({
        type: 'PAGE_ERROR',
        message: error.message,
        url: page.url(),
      });
      console.error(`❌ Page error: ${error.message}`);
    });

    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        report.errors.push({
          type: 'CONSOLE_ERROR',
          message: msg.text(),
          url: page.url(),
        });
      }
    });

    // ========================================================================
    // RESPONSE MONITORING: Catch 5xx errors
    // ========================================================================
    page.on('response', (response) => {
      const status = response.status();
      if (status >= 500) {
        report.errors.push({
          type: 'SERVER_ERROR',
          message: `${status} ${response.statusText()}`,
          url: response.url(),
        });
        console.error(`❌ Server error: ${status} on ${response.url()}`);
      }
    });
  });

  test('should crawl app safely without destructive actions', async ({ page, baseURL }) => {
    console.log('\n🐵 Starting Monkey Test...\n');

    // ========================================================================
    // STEP 1: Navigate to home
    // ========================================================================
    await page.goto(baseURL || 'http://localhost:3000');
    report.pagesVisited.add(page.url());

    // ========================================================================
    // STEP 2: Handle login if needed
    // ========================================================================
    if (await isLoginPage(page)) {
      await attemptLogin(page);
      await page.waitForTimeout(1000);
      report.pagesVisited.add(page.url());
    }

    // ========================================================================
    // STEP 3: Monkey crawling
    // ========================================================================
    let steps = 0;
    const visitedTargets = new Set<string>();
    let clicksOnCurrentPage = 0;
    let currentUrl = page.url();

    while (steps < CONFIG.MAX_STEPS && report.pagesVisited.size < CONFIG.MAX_PAGES) {
      steps++;

      // Reset counter if page changed
      if (page.url() !== currentUrl) {
        currentUrl = page.url();
        clicksOnCurrentPage = 0;
        report.pagesVisited.add(currentUrl);

        // Skip if dangerous route
        if (isDangerousRoute(currentUrl)) {
          console.log(`🛡️  SKIPPING dangerous route: ${currentUrl}`);
          await page.goBack();
          continue;
        }
      }

      // Prevent infinite loops on same page
      if (clicksOnCurrentPage >= CONFIG.MAX_CLICKS_PER_PAGE) {
        console.log(`⚠️  Max clicks on page reached, going back...`);
        await page.goBack({ waitUntil: 'domcontentloaded' }).catch(() => {});
        await page.waitForTimeout(CONFIG.NAVIGATION_DELAY);
        continue;
      }

      console.log(`\n[Step ${steps}/${CONFIG.MAX_STEPS}] on ${page.url()}`);

      // ======================================================================
      // Find clickable elements
      // ======================================================================
      const clickableSelectors = [
        'a[href]:visible',
        'button:visible:not([disabled])',
        '[role="button"]:visible',
        '[onclick]:visible',
        '[tabindex]:visible',
      ];

      const clickables = page.locator(clickableSelectors.join(', '));
      const count = await clickables.count();

      if (count === 0) {
        console.log('  No clickable elements found, going back...');
        await page.goBack({ waitUntil: 'domcontentloaded' }).catch(() => {});
        await page.waitForTimeout(CONFIG.NAVIGATION_DELAY);
        continue;
      }

      // ======================================================================
      // Pick a random element to click
      // ======================================================================
      const randomIndex = Math.floor(Math.random() * count);
      const element = clickables.nth(randomIndex);

      // Create unique identifier for this target
      const tag = await element.evaluate((el) => el.tagName).catch(() => 'UNKNOWN');
      const text = (await element.textContent().catch(() => '')) || '';
      const href = (await element.getAttribute('href').catch(() => '')) || '';
      const targetId = `${tag}:${text.slice(0, 30)}:${href}`;

      // Skip if already visited
      if (visitedTargets.has(targetId)) {
        continue;
      }

      // ======================================================================
      // Safety check: Is element dangerous?
      // ======================================================================
      if (await isDangerousElement(element)) {
        report.clicksBlocked++;
        report.blockedActions.push({
          type: 'CLICK_BLOCKED',
          reason: 'Dangerous UI element detected',
          target: targetId,
        });
        console.log(`  🛡️  BLOCKED click on: ${targetId}`);
        visitedTargets.add(targetId);
        continue;
      }

      // ======================================================================
      // Safety check: Avoid logout (unless explicitly allowed)
      // ======================================================================
      if (text.toLowerCase().includes('logout') || text.toLowerCase().includes('salir')) {
        report.clicksBlocked++;
        report.blockedActions.push({
          type: 'CLICK_BLOCKED',
          reason: 'Logout button - would terminate session',
          target: targetId,
        });
        console.log(`  🛡️  BLOCKED logout button`);
        visitedTargets.add(targetId);
        continue;
      }

      // ======================================================================
      // Safety check: External links
      // ======================================================================
      if (href && (href.startsWith('http://') || href.startsWith('https://'))) {
        const urlObj = new URL(href);
        if (
          !urlObj.hostname.includes('localhost') &&
          !urlObj.hostname.includes('127.0.0.1')
        ) {
          console.log(`  ⏭️  SKIPPING external link: ${href}`);
          visitedTargets.add(targetId);
          continue;
        }
      }

      // ======================================================================
      // SAFE TO CLICK
      // ======================================================================
      try {
        console.log(`  ✅ Clicking: ${targetId}`);

        // Scroll into view
        await element.scrollIntoViewIfNeeded().catch(() => {});

        // Click
        await element.click({ timeout: 5000 });
        report.clicksExecuted++;
        clicksOnCurrentPage++;
        visitedTargets.add(targetId);

        // Wait for any navigation/loading
        await page.waitForTimeout(CONFIG.NAVIGATION_DELAY);

        // ====================================================================
        // Handle forms (if any appeared)
        // ====================================================================
        if (Math.random() < CONFIG.FORM_FILL_PROBABILITY) {
          const inputs = page.locator(
            'input:visible:not([type="submit"]):not([type="button"]):not([disabled])'
          );
          const inputCount = await inputs.count();

          if (inputCount > 0) {
            console.log(`  📝 Filling ${inputCount} form inputs...`);

            for (let i = 0; i < Math.min(inputCount, 5); i++) {
              const input = inputs.nth(i);
              const inputType =
                (await input.getAttribute('type').catch(() => 'text')) || 'text';
              const inputName =
                (await input.getAttribute('name').catch(() => '')) || '';

              // Skip hidden and dangerous inputs
              if (inputType === 'hidden' || inputType === 'file') continue;

              const value = getDummyValue(inputType, inputName);

              try {
                await input.fill(value);
                console.log(`    - Filled ${inputName || inputType}: ${value}`);
              } catch (e) {
                // Input not editable, skip
              }
            }

            // Fill textareas
            const textareas = page.locator('textarea:visible:not([disabled])');
            const textareaCount = await textareas.count();
            for (let i = 0; i < Math.min(textareaCount, 2); i++) {
              try {
                await textareas.nth(i).fill('Test comment generated by E2E');
              } catch (e) {
                // Skip
              }
            }

            // Fill selects
            const selects = page.locator('select:visible:not([disabled])');
            const selectCount = await selects.count();
            for (let i = 0; i < Math.min(selectCount, 3); i++) {
              try {
                const options = await selects.nth(i).locator('option').count();
                if (options > 1) {
                  // Select second option (skip placeholder)
                  await selects.nth(i).selectOption({ index: 1 });
                }
              } catch (e) {
                // Skip
              }
            }

            // NOTE: We do NOT submit forms automatically to avoid creating data
            // Forms are filled to test validation/rendering but not submitted
            console.log(`  ⚠️  Forms filled but NOT submitted (safety measure)`);
          }
        }
      } catch (error: any) {
        console.log(`  ⚠️  Click failed: ${error.message}`);
      }

      // Wait between actions
      await page.waitForTimeout(CONFIG.NAVIGATION_DELAY);
    }

    // ========================================================================
    // STEP 4: Generate and display report
    // ========================================================================
    printReport(report);

    // ========================================================================
    // STEP 5: Assertions
    // ========================================================================
    expect(
      report.pagesVisited.size,
      'Should have visited at least 1 page'
    ).toBeGreaterThanOrEqual(1);

    expect(
      report.clicksExecuted,
      'Should have executed at least 1 click'
    ).toBeGreaterThanOrEqual(1);

    // Only fail on actual PAGE_ERROR (not console 404s from missing assets)
    const pageErrors = report.errors.filter((e) => e.type === 'PAGE_ERROR');
    if (pageErrors.length > 0) {
      console.error('❌ Page errors detected:', pageErrors);
    }
    expect(pageErrors.length).toBe(0);

    console.log('\n✅ Monkey test completed successfully!\n');
  });
});
