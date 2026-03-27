import { test, expect } from '@playwright/test'

test.describe('Email Inbox - Mark as Read', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/login')
    await page.fill('[name=email]', 'client@omnir.test')
    await page.fill('[name=password]', 'testpassword')
    await page.click('button[type=submit]')
    await expect(page).toHaveURL(/\/dashboard/)
  })

  test('marks email thread as read when opened', async ({ page }) => {
    // Navigate to inbox
    await page.goto('/inbox')
    await expect(page).toHaveURL(/\/inbox/)

    // Step 1: Find an unread thread (with unread indicator)
    const unreadThread = page.locator('[data-testid="thread-item"]').filter({ hasText: /unread/i }).first()
    const unreadExists = await unreadThread.count() > 0

    if (!unreadExists) {
      test.skip('No unread threads available for testing')
      return
    }

    // Get thread ID from the element
    const threadId = await unreadThread.getAttribute('data-thread-id')

    // Check initial unread indicator is visible
    await expect(unreadThread.locator('[data-testid="unread-indicator"]')).toBeVisible()

    // Get initial unread count from header badge
    const headerBadge = page.locator('[data-testid="inbox-unread-count"]')
    const initialCount = await headerBadge.textContent().then(t => parseInt(t || '0', 10))

    // Set up network listener to verify API call
    const markReadPromise = page.waitForResponse(
      response => response.url().includes(`/api/v1/emails/${threadId}/read`) && response.request().method() === 'PATCH'
    )

    // Step 2: Click the thread to open it
    await unreadThread.click()

    // Step 5: Verify PATCH request was made and returned 204
    const response = await markReadPromise
    expect(response.status()).toBe(204)

    // Step 3: Confirm unread indicator disappears
    await expect(unreadThread.locator('[data-testid="unread-indicator"]')).not.toBeVisible()

    // Step 4: Confirm unread count badge decremented
    const newCount = await headerBadge.textContent().then(t => parseInt(t || '0', 10))
    expect(newCount).toBe(initialCount - 1)

    // Step 6: Reload page and verify thread is still read
    await page.reload()
    await expect(page).toHaveURL(/\/inbox/)

    const threadAfterReload = page.locator(`[data-thread-id="${threadId}"]`)
    await expect(threadAfterReload.locator('[data-testid="unread-indicator"]')).not.toBeVisible()
  })

  test('opening already-read thread is a no-op', async ({ page }) => {
    // Navigate to inbox
    await page.goto('/inbox')
    await expect(page).toHaveURL(/\/inbox/)

    // Find a read thread (without unread indicator)
    const readThread = page.locator('[data-testid="thread-item"]').filter({ hasNot: page.locator('[data-testid="unread-indicator"]') }).first()
    const readExists = await readThread.count() > 0

    if (!readExists) {
      test.skip('No read threads available for testing')
      return
    }

    const threadId = await readThread.getAttribute('data-thread-id')

    // Set up network listener
    const markReadPromise = page.waitForResponse(
      response => response.url().includes(`/api/v1/emails/${threadId}/read`) && response.request().method() === 'PATCH',
      { timeout: 5000 }
    ).catch(() => null) // Might not fire if frontend optimizes away

    // Click the already-read thread
    await readThread.click()

    // Step 7: If API call was made, verify it returns 204 (no error)
    const response = await markReadPromise
    if (response) {
      expect(response.status()).toBe(204)
    }
  })
})
