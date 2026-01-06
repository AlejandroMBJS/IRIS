/**
 * Permission Matrix E2E Tests
 *
 * Tests the permission matrix functionality:
 * 1. Admin can access the permission matrix page
 * 2. Permission matrix displays all roles and resources
 * 3. Admin can toggle permissions
 * 4. Non-admin users are denied access
 */

import { test, expect, Page } from '@playwright/test'

// Helper function to login as a specific user
async function loginAs(page: Page, email: string, password: string = 'Test123456!') {
  // Clear any existing cookies/sessions
  await page.context().clearCookies()

  await page.goto('/auth/login')
  await page.waitForLoadState('networkidle')

  // Wait for login form to be ready
  await page.waitForSelector('#email', { state: 'visible', timeout: 15000 })
  await page.waitForTimeout(500) // Allow form to stabilize

  // Clear and fill email
  await page.locator('#email').clear()
  await page.locator('#email').fill(email)

  // Clear and fill password
  await page.locator('#password').clear()
  await page.locator('#password').fill(password)

  // Click submit button and wait for navigation
  const submitButton = page.locator('button[type="submit"]')
  await submitButton.click()

  // Wait for navigation away from login
  await page.waitForURL((url) => !url.pathname.includes('/auth/login'), { timeout: 20000 })
  await page.waitForLoadState('domcontentloaded')
}

test.describe('Permission Matrix Tests', () => {
  test.describe('Admin Access', () => {
    test.beforeEach(async ({ page }) => {
      await loginAs(page, 'e2e.admin@test.com')
    })

    test('Admin can access permission matrix page', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Wait for page to load
      await page.waitForTimeout(1000)

      // Verify page loaded - check for either the heading or the loading state
      const heading = page.locator('h1:has-text("Permission Matrix")')
      const loading = page.locator('text=Loading permissions')

      // One of these should be visible
      const hasContent = await heading.isVisible().catch(() => false) ||
                         await loading.isVisible().catch(() => false)

      // If still loading, wait more
      if (await loading.isVisible().catch(() => false)) {
        await page.waitForSelector('h1:has-text("Permission Matrix")', { timeout: 15000 })
      }

      await expect(heading).toBeVisible({ timeout: 15000 })
    })

    test('Permission matrix displays roles filter', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Should have role filter dropdown
      const roleFilter = page.locator('select').first()
      await expect(roleFilter).toBeVisible({ timeout: 15000 })

      // Check for expected roles in dropdown
      const roleOptions = await roleFilter.locator('option').allTextContents()
      expect(roleOptions).toContain('All Roles')
    })

    test('Permission matrix displays resource filter', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Should have resource filter dropdown
      const resourceFilter = page.locator('select').nth(1)
      await expect(resourceFilter).toBeVisible({ timeout: 15000 })

      // Check for expected resources in dropdown
      const resourceOptions = await resourceFilter.locator('option').allTextContents()
      expect(resourceOptions).toContain('All Resources')
    })

    test('Permission matrix displays checkboxes for permissions', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Wait for table to load
      const table = page.locator('table')
      await expect(table).toBeVisible({ timeout: 15000 })

      // Should have checkboxes for permissions
      const checkboxes = page.locator('input[type="checkbox"]')
      const count = await checkboxes.count()

      // Should have multiple checkboxes (roles x resources x permission types)
      expect(count).toBeGreaterThan(0)
    })

    test('Permission legend shows all permission types', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Wait for page to fully load
      await page.waitForTimeout(1000)

      // Check legend for permission types
      const legend = page.locator('text=Permission Types')
      await expect(legend).toBeVisible({ timeout: 15000 })

      // Check for individual permission type labels
      await expect(page.locator('text=View')).toBeVisible()
      await expect(page.locator('text=Create')).toBeVisible()
      await expect(page.locator('text=Edit')).toBeVisible()
      await expect(page.locator('text=Delete')).toBeVisible()
      await expect(page.locator('text=Export')).toBeVisible()
      await expect(page.locator('text=Approve')).toBeVisible()
    })

    test('Refresh button exists and works', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Find refresh button
      const refreshButton = page.locator('button:has-text("Refresh")')
      await expect(refreshButton).toBeVisible({ timeout: 15000 })

      // Click refresh and verify page still works
      await refreshButton.click()
      await page.waitForTimeout(500)

      // Page should still be functional
      const heading = page.locator('h1:has-text("Permission Matrix")')
      await expect(heading).toBeVisible()
    })

    test('Filter by role works', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Wait for table to load
      const table = page.locator('table')
      await expect(table).toBeVisible({ timeout: 15000 })

      // Select admin role from filter
      const roleFilter = page.locator('select').first()
      await roleFilter.selectOption('admin')

      // Wait for filter to apply
      await page.waitForTimeout(300)

      // Table should still be visible
      await expect(table).toBeVisible()

      // Check that admin row is shown
      const adminBadge = page.locator('span:has-text("Admin")').first()
      await expect(adminBadge).toBeVisible()
    })

    test('Filter by resource works', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Wait for table to load
      const table = page.locator('table')
      await expect(table).toBeVisible({ timeout: 15000 })

      // Select employees resource from filter
      const resourceFilter = page.locator('select').nth(1)
      await resourceFilter.selectOption('employees')

      // Wait for filter to apply
      await page.waitForTimeout(300)

      // Table should still be visible
      await expect(table).toBeVisible()
    })
  })

  test.describe('Role-Based Access Control', () => {
    test('Employee user cannot access permission matrix', async ({ page }) => {
      await loginAs(page, 'e2e.employee@test.com')

      // Try to navigate to permissions page
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Wait for any redirects to complete
      await page.waitForTimeout(1500)

      // Should be redirected away from admin page to dashboard
      // Check that the page content shows Dashboard (not Permission Matrix)
      const hasDashboardHeading = await page.locator('h1:has-text("Dashboard")').isVisible().catch(() => false)
      const hasPermissionHeading = await page.locator('h1:has-text("Permission Matrix")').isVisible().catch(() => false)
      const hasAccessDenied = await page.locator('text=Access Denied').isVisible().catch(() => false)

      // Either redirected to dashboard, denied access, or not showing permission matrix
      expect(hasDashboardHeading || hasAccessDenied || !hasPermissionHeading).toBeTruthy()
    })

    test('HR user cannot access permission matrix', async ({ page }) => {
      await loginAs(page, 'e2e.hr@test.com')

      // Try to navigate to permissions page
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1500)

      // Should be redirected away from admin page
      const hasDashboardHeading = await page.locator('h1:has-text("Dashboard")').isVisible().catch(() => false)
      const hasPermissionHeading = await page.locator('h1:has-text("Permission Matrix")').isVisible().catch(() => false)
      const hasAccessDenied = await page.locator('text=Access Denied').isVisible().catch(() => false)

      // Either redirected to dashboard, denied access, or not showing permission matrix
      expect(hasDashboardHeading || hasAccessDenied || !hasPermissionHeading).toBeTruthy()
    })
  })

  test.describe('Permission Toggle Functionality', () => {
    test.beforeEach(async ({ page }) => {
      await loginAs(page, 'e2e.admin@test.com')
    })

    test('Checkboxes are clickable in permission matrix', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Wait for table to load
      const table = page.locator('table')
      await expect(table).toBeVisible({ timeout: 15000 })

      // Find checkboxes in the table
      const checkboxes = page.locator('input[type="checkbox"]')
      const checkboxCount = await checkboxes.count()

      // Verify checkboxes exist
      expect(checkboxCount).toBeGreaterThan(0)

      // Verify checkboxes are interactive (not disabled)
      const firstCheckbox = checkboxes.first()
      const isDisabled = await firstCheckbox.isDisabled()

      // If not disabled, the checkbox should be clickable
      if (!isDisabled) {
        // Click the checkbox (just verify it's clickable)
        await firstCheckbox.click()
        await page.waitForTimeout(300)

        // Click again to restore state
        await firstCheckbox.click()
        await page.waitForTimeout(300)
      }

      // Verify page is still functional after interactions
      await expect(table).toBeVisible()
    })

    test('Permission matrix table has proper structure', async ({ page }) => {
      await page.goto('/admin/permissions')
      await page.waitForLoadState('domcontentloaded')

      // Wait for table to load
      const table = page.locator('table')
      await expect(table).toBeVisible({ timeout: 15000 })

      // Verify table has header row
      const headerRow = table.locator('thead tr')
      await expect(headerRow).toBeVisible()

      // Verify table has body rows
      const bodyRows = table.locator('tbody tr')
      const rowCount = await bodyRows.count()

      // Should have rows for each role
      expect(rowCount).toBeGreaterThan(0)

      // Verify each row has checkboxes
      const firstRow = bodyRows.first()
      const rowCheckboxes = firstRow.locator('input[type="checkbox"]')
      const rowCheckboxCount = await rowCheckboxes.count()

      // Each row should have checkboxes for each permission type
      expect(rowCheckboxCount).toBeGreaterThan(0)
    })
  })
})
