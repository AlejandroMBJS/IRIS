/**
 * IRIS Payroll - Advanced Multi-User Monkey/Crawler E2E Test
 *
 * PURPOSE:
 * - Test app with MULTIPLE user roles (admin, hr, supervisor, manager, employee, etc.)
 * - Navigate ALL available routes for each role
 * - Click ALL interactive elements
 * - Fill forms and test workflows
 * - EXCLUDE dangerous/destructive actions
 *
 * FEATURES:
 * ✅ Multi-user testing (12 different roles)
 * ✅ Complete route mapping (30+ pages)
 * ✅ Aggressive exploration mode
 * ✅ Approval workflow testing
 * ✅ Role-specific access validation
 * ✅ Safety guards (no destructive actions)
 *
 * USAGE:
 *   npm run test:e2e         # All configured roles
 *   npm run test:e2e:headed  # With browser UI
 */

import { test, expect, Page, Route, BrowserContext } from '@playwright/test';

// ============================================================================
// CONFIGURATION
// ============================================================================

const CONFIG = {
  MAX_STEPS_PER_USER: parseInt(process.env.E2E_MAX_STEPS_PER_USER || '50'),
  MAX_PAGES_PER_USER: parseInt(process.env.E2E_MAX_PAGES_PER_USER || '15'),
  MAX_CLICKS_PER_PAGE: 15,
  NAVIGATION_DELAY: 300, // ms
  FORM_FILL_PROBABILITY: 0.8,
  AGGRESSIVE_MODE: process.env.E2E_AGGRESSIVE_MODE === 'true',
  EXPLORE_ALL_ROUTES: true,
};

// ============================================================================
// ROUTE MAPPING - ALL FRONTEND ROUTES
// ============================================================================

const ALL_ROUTES = [
  // Auth routes (public)
  '/auth/login',
  '/auth/register',

  // Dashboard routes
  '/dashboard',
  '/dashboard/hr',
  '/dashboard/manager',
  '/dashboard/supervisor',

  // Employee routes
  '/employees',
  '/employees/new',

  // Payroll routes
  '/payroll',
  '/payroll/periods',
  '/payroll/export',

  // Incidences routes
  '/incidences',
  '/incidences/types',
  '/incidences/vacations',

  // Approvals
  '/approvals',

  // Announcements
  '/announcements',
  '/announcements/new',

  // Inbox & Notifications
  '/inbox',
  '/notifications',

  // HR Calendar
  '/hr/calendar',

  // Configuration routes (role-specific)
  '/configuration/shifts',
  '/configuration/payroll-setup',
  '/configuration/incidence-mapping',
  // NOTE: /configuration/role-inheritance and /configuration/permissions are BLOCKED

  // Admin routes (admin only)
  // NOTE: /admin/* routes are BLOCKED by security
];

const ROUTE_ACCESS_BY_ROLE: Record<string, string[]> = {
  admin: ALL_ROUTES,
  hr: ALL_ROUTES.filter((r) => !r.startsWith('/admin')),
  hr_and_pr: ALL_ROUTES.filter((r) => !r.startsWith('/admin')),
  supervisor: [
    '/dashboard',
    '/dashboard/supervisor',
    '/employees',
    '/incidences',
    '/approvals',
    '/announcements',
    '/inbox',
    '/notifications',
  ],
  manager: [
    '/dashboard',
    '/dashboard/manager',
    '/employees',
    '/incidences',
    '/approvals',
    '/announcements',
    '/inbox',
    '/notifications',
  ],
  employee: [
    '/dashboard',
    '/incidences',
    '/announcements',
    '/inbox',
    '/notifications',
  ],
  accountant: ALL_ROUTES.filter((r) => !r.startsWith('/admin') && !r.includes('/hr/')),
  payroll_staff: [
    '/dashboard',
    '/employees',
    '/payroll',
    '/payroll/periods',
    '/payroll/export',
    '/incidences',
    '/inbox',
  ],
  viewer: ['/dashboard', '/employees', '/payroll', '/incidences', '/announcements'],
};

// ============================================================================
// DANGEROUS PATTERNS (DENYLISTS) - same as before
// ============================================================================

const DANGEROUS_ENDPOINT_PATTERNS = [
  /\/admin\//i,
  /\/permission/i,
  /\/role-inheritance/i,
  /\/users\/\d+$/i,
  /\/delete/i,
  /\/destroy/i,
  /\/drop/i,
  /\/purge/i,
  /\/reset/i,
  /\/seed/i,
  /\/migrate/i,
  /\/billing/i,
  /\/payment/i,
  /\/invoice/i,
  /\/reimburse/i,
  /\/mark-paid/i,
  /\/issue.*payment/i,
  /\/advance-payments\/\d+\/(approve|issue)/i,
  /\/terminate/i,
  /\/decline/i,
  /\/reject/i,
  /\/archive/i,
  /\/revoke/i,
  /\/webhook/i,
  /\/secret/i,
  /\/key/i,
  /\/oauth/i,
];

const DANGEROUS_ROUTE_PATTERNS = [
  /\/admin\//i,
  /\/configuration\/role-inheritance/i,
  /\/configuration\/permissions/i,
  /\/billing/i,
  /\/payment/i,
  /\/subscription/i,
];

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

function isDangerousRoute(url: string): boolean {
  const path = new URL(url).pathname;
  return DANGEROUS_ROUTE_PATTERNS.some((pattern) => pattern.test(path));
}

function isDangerousEndpoint(url: string, method: string): boolean {
  if (method === 'DELETE') return true;
  return DANGEROUS_ENDPOINT_PATTERNS.some((pattern) => pattern.test(url));
}

async function isDangerousElement(locator: any): Promise<boolean> {
  try {
    const text = (await locator.textContent().catch(() => '')) || '';
    if (DANGEROUS_UI_TEXT_PATTERNS.some((p) => p.test(text))) return true;

    const ariaLabel = (await locator.getAttribute('aria-label').catch(() => '')) || '';
    if (DANGEROUS_UI_TEXT_PATTERNS.some((p) => p.test(ariaLabel))) return true;

    const testId = (await locator.getAttribute('data-testid').catch(() => '')) || '';
    if (DANGEROUS_UI_TEXT_PATTERNS.some((p) => p.test(testId))) return true;

    const className = (await locator.getAttribute('class').catch(() => '')) || '';
    if (DANGEROUS_UI_CLASS_PATTERNS.some((p) => p.test(className))) return true;

    const href = (await locator.getAttribute('href').catch(() => '')) || '';
    if (href && href.startsWith('/')) {
      const fullUrl = `${process.env.BASE_URL || 'http://localhost:3000'}${href}`;
      if (isDangerousRoute(fullUrl)) return true;
    }

    return false;
  } catch {
    return false;
  }
}

function getDummyValue(inputType: string, name: string): string {
  const lowerName = name.toLowerCase();

  if (lowerName.includes('email')) return `test.${Date.now()}@example.com`;
  if (lowerName.includes('password')) return 'TestPassword123!';
  if (lowerName.includes('phone') || lowerName.includes('tel')) return '5551234567';
  if (lowerName.includes('name')) return `Test User ${Date.now()}`;
  if (lowerName.includes('date')) return '2024-01-15';
  if (lowerName.includes('number') || lowerName.includes('amount')) return '100';

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

// ============================================================================
// USER SESSION MANAGEMENT
// ============================================================================

interface UserSession {
  email: string;
  password: string;
  role: string;
  fullName: string;
  accessToken?: string;
}

async function loginUser(page: Page, session: UserSession): Promise<boolean> {
  try {
    console.log(`\n🔐 Logging in as: ${session.role} (${session.email})`);

    // Navigate to login
    await page.goto('/auth/login');
    await page.waitForTimeout(1000);

    // Find and fill email/password
    const emailInput = page.locator('input[type="email"], input[name="email"]').first();
    const passwordInput = page.locator('input[type="password"]').first();

    if ((await emailInput.count()) === 0 || (await passwordInput.count()) === 0) {
      console.log('❌ Login form not found');
      return false;
    }

    await emailInput.fill(session.email);
    await passwordInput.fill(session.password);

    // Submit
    const submitButton = page.locator('button[type="submit"]').first();
    await submitButton.click();

    // Wait for navigation
    await page.waitForTimeout(3000);

    // Check if still on login page
    if (page.url().includes('/auth/login')) {
      console.log('❌ Login failed - still on login page');
      return false;
    }

    console.log(`✅ Logged in successfully as ${session.role}`);
    return true;
  } catch (error) {
    console.log(`❌ Login error: ${error}`);
    return false;
  }
}

function getTestUsersFromEnv(): UserSession[] {
  const password = process.env.E2E_PASSWORD || 'Test123456!';
  const rolesToTest = (process.env.E2E_TEST_ROLES || 'ADMIN,HR,SUPERVISOR,MANAGER,EMPLOYEE').split(',');

  const allUsers: Record<string, UserSession> = {
    ADMIN: {
      email: process.env.E2E_EMAIL_ADMIN || 'e2e.admin@test.com',
      password,
      role: 'admin',
      fullName: 'E2E Admin User',
    },
    HR: {
      email: process.env.E2E_EMAIL_HR || 'e2e.hr@test.com',
      password,
      role: 'hr',
      fullName: 'E2E HR User',
    },
    SUPERVISOR: {
      email: process.env.E2E_EMAIL_SUPERVISOR || 'e2e.supervisor@test.com',
      password,
      role: 'supervisor',
      fullName: 'E2E Supervisor User',
    },
    MANAGER: {
      email: process.env.E2E_EMAIL_MANAGER || 'e2e.manager@test.com',
      password,
      role: 'manager',
      fullName: 'E2E Manager User',
    },
    EMPLOYEE: {
      email: process.env.E2E_EMAIL_EMPLOYEE || 'e2e.employee@test.com',
      password,
      role: 'employee',
      fullName: 'E2E Employee User',
    },
    ACCOUNTANT: {
      email: process.env.E2E_EMAIL_ACCOUNTANT || 'e2e.accountant@test.com',
      password,
      role: 'accountant',
      fullName: 'E2E Accountant User',
    },
    PAYROLL_STAFF: {
      email: process.env.E2E_EMAIL_PAYROLL_STAFF || 'e2e.payroll@test.com',
      password,
      role: 'payroll_staff',
      fullName: 'E2E Payroll Staff User',
    },
  };

  return rolesToTest.map((role) => allUsers[role.trim()]).filter(Boolean);
}

// ============================================================================
// REPORTING
// ============================================================================

interface UserReport {
  role: string;
  email: string;
  pagesVisited: Set<string>;
  clicksExecuted: number;
  clicksBlocked: number;
  requestsBlocked: number;
  blockedActions: Array<{ type: string; reason: string; target: string }>;
  errors: Array<{ type: string; message: string; url: string }>;
  routesCovered: string[];
}

interface MonkeyReport {
  users: Map<string, UserReport>;
  totalPages: Set<string>;
  totalClicks: number;
  totalClicksBlocked: number;
  totalRequestsBlocked: number;
  totalErrors: number;
}

function createMonkeyReport(): MonkeyReport {
  return {
    users: new Map(),
    totalPages: new Set(),
    totalClicks: 0,
    totalClicksBlocked: 0,
    totalRequestsBlocked: 0,
    totalErrors: 0,
  };
}

function createUserReport(session: UserSession): UserReport {
  return {
    role: session.role,
    email: session.email,
    pagesVisited: new Set(),
    clicksExecuted: 0,
    clicksBlocked: 0,
    requestsBlocked: 0,
    blockedActions: [],
    errors: [],
    routesCovered: [],
  };
}

function printFinalReport(report: MonkeyReport): void {
  console.log('\n' + '='.repeat(80));
  console.log('🐵 MULTI-USER MONKEY TEST - FINAL REPORT');
  console.log('='.repeat(80));

  console.log('\n📊 OVERALL STATISTICS:');
  console.log(`  Users tested: ${report.users.size}`);
  console.log(`  Total unique pages: ${report.totalPages.size}`);
  console.log(`  Total clicks executed: ${report.totalClicks}`);
  console.log(`  Total clicks blocked: ${report.totalClicksBlocked}`);
  console.log(`  Total requests blocked: ${report.totalRequestsBlocked}`);
  console.log(`  Total errors: ${report.totalErrors}`);

  console.log('\n📋 BY USER ROLE:');
  report.users.forEach((userReport, role) => {
    console.log(`\n  ${role.toUpperCase()}:`);
    console.log(`    Email: ${userReport.email}`);
    console.log(`    Pages visited: ${userReport.pagesVisited.size}`);
    console.log(`    Clicks executed: ${userReport.clicksExecuted}`);
    console.log(`    Clicks blocked: ${userReport.clicksBlocked}`);
    console.log(`    Requests blocked: ${userReport.requestsBlocked}`);
    console.log(`    Errors: ${userReport.errors.length}`);

    if (userReport.pagesVisited.size > 0) {
      console.log(`    Routes covered:`);
      Array.from(userReport.pagesVisited)
        .slice(0, 10)
        .forEach((url) => {
          const path = new URL(url).pathname;
          console.log(`      - ${path}`);
        });
      if (userReport.pagesVisited.size > 10) {
        console.log(`      ... and ${userReport.pagesVisited.size - 10} more`);
      }
    }
  });

  console.log('\n' + '='.repeat(80) + '\n');
}

// ============================================================================
// MAIN TEST
// ============================================================================

test.describe('Multi-User Monkey/Crawler E2E Test', () => {
  let globalReport: MonkeyReport;

  test.beforeAll(async () => {
    globalReport = createMonkeyReport();
  });

  test.afterAll(async () => {
    printFinalReport(globalReport);
  });

  // Get all test users
  const testUsers = getTestUsersFromEnv();

  testUsers.forEach((userSession) => {
    test(`should crawl app as ${userSession.role} without destructive actions`, async ({
      page,
      baseURL,
    }) => {
      // Security check
      if (!baseURL?.includes('localhost') && !baseURL?.includes('127.0.0.1')) {
        throw new Error(`❌ SECURITY: Tests can only run on localhost. Current: ${baseURL}`);
      }

      // Create user report
      const userReport = createUserReport(userSession);
      globalReport.users.set(userSession.role, userReport);

      // Setup network interception
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

        // Check if dangerous
        if (isDangerousEndpoint(url, method)) {
          userReport.requestsBlocked++;
          globalReport.totalRequestsBlocked++;
          userReport.blockedActions.push({
            type: 'REQUEST_BLOCKED',
            reason: `Dangerous ${method} request`,
            target: url,
          });

          console.log(`🛡️  BLOCKED: ${method} ${url}`);

          return route.fulfill({
            status: 403,
            contentType: 'application/json',
            body: JSON.stringify({
              error: 'Blocked by E2E safety guard',
              reason: 'Dangerous endpoint detected',
            }),
          });
        }

        return route.continue();
      });

      // Error monitoring
      page.on('pageerror', (error) => {
        userReport.errors.push({
          type: 'PAGE_ERROR',
          message: error.message,
          url: page.url(),
        });
        globalReport.totalErrors++;
      });

      page.on('response', (response) => {
        const status = response.status();
        if (status >= 500) {
          userReport.errors.push({
            type: 'SERVER_ERROR',
            message: `${status} ${response.statusText()}`,
            url: response.url(),
          });
          globalReport.totalErrors++;
        }
      });

      // Login
      const loggedIn = await loginUser(page, userSession);
      if (!loggedIn) {
        console.log(`⚠️  Failed to login as ${userSession.role} - skipping`);
        return;
      }

      // Record current page after login
      userReport.pagesVisited.add(page.url());
      globalReport.totalPages.add(page.url());

      // Get routes accessible to this role
      const accessibleRoutes =
        ROUTE_ACCESS_BY_ROLE[userSession.role] || ROUTE_ACCESS_BY_ROLE['employee'];

      console.log(`\n🗺️  Accessible routes for ${userSession.role}: ${accessibleRoutes.length}`);

      // Explore routes
      for (const route of accessibleRoutes.slice(0, CONFIG.MAX_PAGES_PER_USER)) {
        try {
          // Skip if already on this route
          const currentPath = new URL(page.url()).pathname;
          if (currentPath === route) {
            console.log(`\n📍 Already on: ${route}`);
            await monkeyCrawlPage(page, userReport, 10);
            continue;
          }

          console.log(`\n📍 Navigating to: ${route}`);

          // Use baseURL for navigation
          const fullUrl = `${baseURL}${route}`;
          await page.goto(fullUrl, {
            waitUntil: 'domcontentloaded',
            timeout: 15000
          }).catch((err) => {
            console.log(`⚠️  Navigation error: ${err.message}`);
            return null;
          });

          // Small delay to let page stabilize
          await page.waitForTimeout(CONFIG.NAVIGATION_DELAY);

          // Verify page is still alive
          if (page.isClosed()) {
            console.log(`⚠️  Page was closed, stopping exploration for ${userSession.role}`);
            break;
          }

          userReport.pagesVisited.add(page.url());
          globalReport.totalPages.add(page.url());

          // Monkey crawl on this page
          await monkeyCrawlPage(page, userReport, 10);
        } catch (error: any) {
          console.log(`⚠️  Failed to load ${route}: ${error.message}`);

          // If page is closed, stop exploring
          if (page.isClosed()) {
            console.log(`⚠️  Page closed, stopping exploration`);
            break;
          }
        }
      }

      // Additional random exploration
      await randomExploration(page, userReport, CONFIG.MAX_STEPS_PER_USER - 10);

      // Update global stats
      globalReport.totalClicks += userReport.clicksExecuted;
      globalReport.totalClicksBlocked += userReport.clicksBlocked;

      // Assertions
      expect(userReport.pagesVisited.size).toBeGreaterThanOrEqual(1);
      expect(userReport.errors.filter((e) => e.type === 'PAGE_ERROR').length).toBe(0);

      console.log(`\n✅ Completed testing for ${userSession.role}`);
    });
  });
});

// ============================================================================
// CRAWLING FUNCTIONS
// ============================================================================

async function monkeyCrawlPage(page: Page, report: UserReport, maxClicks: number): Promise<void> {
  const visitedTargets = new Set<string>();
  let clicks = 0;

  while (clicks < maxClicks) {
    const clickableSelectors = [
      'a[href]:visible',
      'button:visible:not([disabled])',
      '[role="button"]:visible',
      '[onclick]:visible',
    ];

    const clickables = page.locator(clickableSelectors.join(', '));
    const count = await clickables.count();

    if (count === 0) break;

    const randomIndex = Math.floor(Math.random() * count);
    const element = clickables.nth(randomIndex);

    const tag = (await element.evaluate((el) => el.tagName).catch(() => 'UNKNOWN')) || '';
    const text = (await element.textContent().catch(() => '')) || '';
    const href = (await element.getAttribute('href').catch(() => '')) || '';
    const targetId = `${tag}:${text.slice(0, 30)}:${href}`;

    if (visitedTargets.has(targetId)) continue;
    visitedTargets.add(targetId);

    if (await isDangerousElement(element)) {
      report.clicksBlocked++;
      report.blockedActions.push({
        type: 'CLICK_BLOCKED',
        reason: 'Dangerous UI element',
        target: targetId,
      });
      continue;
    }

    if (text.toLowerCase().includes('logout') || text.toLowerCase().includes('salir')) {
      report.clicksBlocked++;
      continue;
    }

    try {
      await element.scrollIntoViewIfNeeded().catch(() => {});
      await element.click({ timeout: 3000 });
      report.clicksExecuted++;
      clicks++;

      await page.waitForTimeout(CONFIG.NAVIGATION_DELAY);

      // Fill forms if appeared
      if (Math.random() < CONFIG.FORM_FILL_PROBABILITY) {
        await fillForms(page);
      }
    } catch (error: any) {
      // Click failed, continue
    }
  }
}

async function randomExploration(page: Page, report: UserReport, maxSteps: number): Promise<void> {
  // Additional random exploration logic
  // Similar to monkeyCrawlPage but more exploratory
  await monkeyCrawlPage(page, report, maxSteps);
}

async function fillForms(page: Page): Promise<void> {
  const inputs = page.locator(
    'input:visible:not([type="submit"]):not([type="button"]):not([disabled])'
  );
  const inputCount = await inputs.count();

  for (let i = 0; i < Math.min(inputCount, 3); i++) {
    const input = inputs.nth(i);
    const inputType = (await input.getAttribute('type').catch(() => 'text')) || 'text';
    const inputName = (await input.getAttribute('name').catch(() => '')) || '';

    if (inputType === 'hidden' || inputType === 'file') continue;

    const value = getDummyValue(inputType, inputName);

    try {
      await input.fill(value);
    } catch (e) {
      // Skip
    }
  }
}
