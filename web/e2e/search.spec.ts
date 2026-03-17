import { test, expect } from './fixtures/auth'

test.describe('Global Search', () => {
  test('returns contacts, accounts, and deals matching query', async ({ authenticatedPage: page }) => {
    await page.goto('/dashboard')

    await page.click('[aria-label="Search"]')
    await page.fill('[placeholder*="Search"]', 'Ada')

    // Results should appear
    await expect(page.getByText('Ada Lovelace')).toBeVisible()
  })

  test('shows empty state for no matches', async ({ authenticatedPage: page }) => {
    await page.goto('/dashboard')

    await page.click('[aria-label="Search"]')
    await page.fill('[placeholder*="Search"]', 'zzz_no_match_zzz')

    await expect(page.getByText(/no results/i)).toBeVisible()
  })
})
