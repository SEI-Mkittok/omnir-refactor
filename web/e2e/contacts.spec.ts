import { test, expect } from './fixtures/auth'

test.describe('Contact CRUD', () => {
  test('creates a new contact', async ({ authenticatedPage: page }) => {
    await page.goto('/contacts/new')

    await page.fill('[name=firstName]', 'Ada')
    await page.fill('[name=lastName]', 'Lovelace')
    await page.fill('[name=email]', 'ada@omnir.test')
    await page.selectOption('[name=stage]', 'prospect')
    await page.click('button[type=submit]')

    await expect(page).toHaveURL(/\/contacts\/[\w-]+/)
    await expect(page.getByText('Ada Lovelace')).toBeVisible()
    await expect(page.getByText('ada@omnir.test')).toBeVisible()
  })

  test('edits an existing contact', async ({ authenticatedPage: page }) => {
    // Navigate to contact list and open first contact
    await page.goto('/contacts')
    await page.click('text=Ada Lovelace')

    await page.click('[aria-label="Edit contact"]')
    await page.fill('[name=firstName]', 'Augusta')
    await page.click('button[type=submit]')

    await expect(page.getByText('Augusta Lovelace')).toBeVisible()
  })

  test('deletes a contact', async ({ authenticatedPage: page }) => {
    await page.goto('/contacts')
    await page.click('text=Ada Lovelace')

    await page.click('[aria-label="Delete contact"]')
    // Confirm in dialog
    await page.click('button:has-text("Delete")')

    await expect(page).toHaveURL(/\/contacts$/)
    await expect(page.getByText('Ada Lovelace')).not.toBeVisible()
  })
})
