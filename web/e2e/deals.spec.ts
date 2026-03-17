import { test, expect } from './fixtures/auth'

test.describe('Deal Pipeline', () => {
  test('creates a deal and moves it through stages to closed won', async ({ authenticatedPage: page }) => {
    await page.goto('/deals/new')

    await page.fill('[name=title]', 'Enterprise Contract Q2')
    await page.fill('[name=valueCents]', '1000000')
    await page.selectOption('[name=stage]', 'lead')
    await page.click('button[type=submit]')

    await expect(page.getByText('Enterprise Contract Q2')).toBeVisible()

    // Advance to qualified
    await page.click('[aria-label="Advance stage"]')
    await expect(page.getByText('qualified')).toBeVisible()

    // Move to closed_won via stage selector
    await page.selectOption('[name=stage]', 'closed_won')
    await page.click('button:has-text("Save")')
    await expect(page.getByText('Closed Won')).toBeVisible()
  })
})
