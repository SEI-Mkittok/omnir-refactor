import { test, expect } from '@playwright/test'

test.describe('Authentication', () => {
  test('logs in with valid credentials', async ({ page }) => {
    await page.goto('/login')
    await page.fill('[name=email]', 'admin@omnir.test')
    await page.fill('[name=password]', 'testpassword')
    await page.click('button[type=submit]')
    await expect(page).toHaveURL(/\/dashboard/)
    await expect(page.getByText('Admin')).toBeVisible()
  })

  test('shows error for invalid credentials', async ({ page }) => {
    await page.goto('/login')
    await page.fill('[name=email]', 'admin@omnir.test')
    await page.fill('[name=password]', 'wrongpassword')
    await page.click('button[type=submit]')
    await expect(page.getByText(/invalid credentials/i)).toBeVisible()
    await expect(page).toHaveURL(/\/login/)
  })

  test('logs out and redirects to login', async ({ page }) => {
    // Login first
    await page.goto('/login')
    await page.fill('[name=email]', 'admin@omnir.test')
    await page.fill('[name=password]', 'testpassword')
    await page.click('button[type=submit]')
    await expect(page).toHaveURL(/\/dashboard/)

    // Logout
    await page.click('[aria-label="User menu"]')
    await page.click('text=Log out')
    await expect(page).toHaveURL(/\/login/)
  })
})
