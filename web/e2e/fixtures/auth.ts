import { test as base, expect } from '@playwright/test'

/**
 * Extends Playwright's test with an `authenticatedPage` fixture.
 * The fixture logs in as the seed admin user before the test body runs.
 *
 * Usage:
 *   import { test, expect } from './fixtures/auth'
 *
 *   test('does something', async ({ authenticatedPage }) => {
 *     await authenticatedPage.goto('/contacts')
 *     // ...
 *   })
 */
export const test = base.extend<{ authenticatedPage: typeof base.prototype.page }>({
  authenticatedPage: async ({ page }, use) => {
    await page.goto('/login')
    await page.fill('[name=email]', 'admin@omnir.test')
    await page.fill('[name=password]', 'testpassword')
    await page.click('button[type=submit]')
    await expect(page).toHaveURL(/\/dashboard/)
    await use(page)
  },
})

export { expect } from '@playwright/test'
