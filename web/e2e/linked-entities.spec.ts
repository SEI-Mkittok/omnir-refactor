import { type APIRequestContext, type Page } from '@playwright/test'
import { test, expect } from './fixtures/auth'

const OWNER_ID = '00000000-0000-0000-0000-000000000001'
const PIPELINE_ID = '00000000-0000-0000-0000-000000000001'

type Entity = { id: string }

async function expectOk(response: Awaited<ReturnType<APIRequestContext['post']>>) {
  expect(response.ok(), await response.text()).toBeTruthy()
  return response.json()
}

async function createAccount(request: APIRequestContext, name: string) {
  return expectOk(
    await request.post('/api/v1/accounts', {
      data: {
        name,
        domain: `${name.toLowerCase().replace(/[^a-z0-9]+/g, '-')}.test`,
        industry: 'technology',
        owner_id: OWNER_ID,
      },
    })
  ) as Promise<Entity & { name: string }>
}

async function createContact(request: APIRequestContext, name: string, accountId?: string) {
  const [firstName, ...lastParts] = name.split(' ')
  return expectOk(
    await request.post('/api/v1/contacts', {
      data: {
        first_name: firstName,
        last_name: lastParts.join(' ') || 'Tester',
        email: `${name.toLowerCase().replace(/[^a-z0-9]+/g, '.')}@omnir.test`,
        stage: 'customer',
        account_id: accountId,
        owner_id: OWNER_ID,
      },
    })
  ) as Promise<Entity>
}

async function createDeal(request: APIRequestContext, title: string, accountId?: string, contactId?: string) {
  return expectOk(
    await request.post('/api/v1/deals', {
      data: {
        title,
        value_cents: 123400,
        currency: 'USD',
        stage: 'proposal',
        account_id: accountId,
        contact_id: contactId,
        owner_id: OWNER_ID,
        pipeline_id: PIPELINE_ID,
      },
    })
  ) as Promise<Entity>
}

async function createTicket(request: APIRequestContext, subject: string, accountId?: string, contactId?: string) {
  return expectOk(
    await request.post('/api/v1/tickets', {
      data: {
        subject,
        status: 'open',
        priority: 'high',
        account_id: accountId,
        contact_id: contactId,
      },
    })
  ) as Promise<Entity>
}

async function cleanup(request: APIRequestContext, path: string, entities: Entity[]) {
  for (const entity of entities.reverse()) {
    await request.delete(`/api/v1/${path}/${entity.id}`)
  }
}

async function linkExisting(page: Page, placeholder: string, query: string, resultText: string) {
  await page.getByRole('button', { name: /link existing|link/i }).click()
  await expect(page.getByPlaceholder(placeholder)).toBeVisible()
  await page.getByPlaceholder(placeholder).fill(query)
  await page.getByText(resultText).click()
}

async function expectScopedLink(
  page: Page,
  linkName: string,
  path: string,
  scope: { idParam: string; id: string; nameParam: string; name: string }
) {
  await page.getByRole('link', { name: linkName }).click()
  await expect(page).toHaveURL(new RegExp(`${path}\\?.*${scope.idParam}=${scope.id}`))
  await expect(page).toHaveURL(new RegExp(`${path}\\?.*${scope.nameParam}=${encodeURIComponent(scope.name).replace(/%20/g, '\\+')}`))
}

test.describe('Linked entity association flows', () => {
  test('contact hub links existing account/deal/ticket, unlinks them, and opens scoped deep links', async ({ authenticatedPage: page }) => {
    const suffix = Date.now()
    const contactName = `Contact Hub ${suffix}`
    const account = await createAccount(page.request, `Contact Hub Account ${suffix}`)
    const contact = await createContact(page.request, contactName)
    const deal = await createDeal(page.request, `Contact Hub Deal ${suffix}`)
    const ticket = await createTicket(page.request, `Contact Hub Ticket ${suffix}`)
    const contactScope = { idParam: 'contact_id', id: contact.id, nameParam: 'contact_name', name: contactName }

    try {
      await page.goto(`/contacts/${contact.id}`)
      await expect(page.getByText(contactName)).toBeVisible()

      await expectScopedLink(page, 'Open Accounts', '/accounts', contactScope)
      await page.goto(`/contacts/${contact.id}`)
      await page.getByRole('button', { name: /Deals/i }).click()
      await expectScopedLink(page, 'Open Deals', '/deals', contactScope)
      await page.goto(`/contacts/${contact.id}`)
      await page.getByRole('button', { name: /Tickets/i }).click()
      await expectScopedLink(page, 'Open Tickets', '/tickets', contactScope)
      await page.goto(`/contacts/${contact.id}`)

      await linkExisting(page, 'Search accounts by name…', `Contact Hub Account ${suffix}`, `Contact Hub Account ${suffix}`)
      await expect(page.getByText('Account linked')).toBeVisible()
      await expect(page.getByText(`Contact Hub Account ${suffix}`)).toBeVisible()

      await page.getByRole('button', { name: /Deals/i }).click()
      await linkExisting(page, 'Search deals by title…', `Contact Hub Deal ${suffix}`, `Contact Hub Deal ${suffix}`)
      await expect(page.getByText('Deal linked')).toBeVisible()
      await expect(page.getByText(`Contact Hub Deal ${suffix}`)).toBeVisible()

      await page.getByRole('button', { name: /Tickets/i }).click()
      await linkExisting(page, 'Search tickets by subject…', `Contact Hub Ticket ${suffix}`, `Contact Hub Ticket ${suffix}`)
      await expect(page.getByText('Ticket linked')).toBeVisible()
      await expect(page.getByText(`Contact Hub Ticket ${suffix}`)).toBeVisible()

      await page.goto(`/contacts/${contact.id}`)
      await page.getByRole('button', { name: /Deals/i }).click()
      await page.getByLabel('Unlink deal').click()
      await expect(page.getByText('Deal unlinked')).toBeVisible()
      await expect(page.getByText(`Contact Hub Deal ${suffix}`)).not.toBeVisible()

      await page.getByRole('button', { name: /Tickets/i }).click()
      await page.getByLabel('Unlink ticket').click()
      await expect(page.getByText('Ticket unlinked')).toBeVisible()
      await expect(page.getByText(`Contact Hub Ticket ${suffix}`)).not.toBeVisible()

      await page.getByRole('button', { name: /Accounts/i }).click()
      await page.getByLabel('Unlink account').click()
      await expect(page.getByText('Account unlinked')).toBeVisible()
      await expect(page.getByText(`Contact Hub Account ${suffix}`)).not.toBeVisible()
    } finally {
      await cleanup(page.request, 'tickets', [ticket])
      await cleanup(page.request, 'deals', [deal])
      await cleanup(page.request, 'contacts', [contact])
      await cleanup(page.request, 'accounts', [account])
    }
  })

  test('account hub links existing entities, resolves ticket mismatch by relinking contact, and opens scoped deep links', async ({ authenticatedPage: page }) => {
    const suffix = Date.now()
    const accountName = `Account Hub Target ${suffix}`
    const targetAccount = await createAccount(page.request, accountName)
    const sourceAccount = await createAccount(page.request, `Account Hub Source ${suffix}`)
    const contact = await createContact(page.request, `Account Hub Contact ${suffix}`, sourceAccount.id)
    const deal = await createDeal(page.request, `Account Hub Deal ${suffix}`)
    const ticket = await createTicket(page.request, `Account Hub Conflict Ticket ${suffix}`, sourceAccount.id, contact.id)
    const accountScope = { idParam: 'account_id', id: targetAccount.id, nameParam: 'account_name', name: accountName }

    try {
      await page.goto(`/accounts/${targetAccount.id}`)
      await expect(page.getByText(accountName)).toBeVisible()

      await expectScopedLink(page, 'Open Contacts', '/contacts', accountScope)
      await page.goto(`/accounts/${targetAccount.id}`)
      await page.getByRole('button', { name: /Deals/i }).click()
      await expectScopedLink(page, 'Open Deals', '/deals', accountScope)
      await page.goto(`/accounts/${targetAccount.id}`)
      await page.getByRole('button', { name: /Tickets/i }).click()
      await expectScopedLink(page, 'Open Tickets', '/tickets', accountScope)
      await page.goto(`/accounts/${targetAccount.id}`)

      await linkExisting(page, 'Search by name or email…', `Account Hub Contact ${suffix}`, `Account Hub Contact ${suffix}`)
      await expect(page.getByText(`Account Hub Contact ${suffix}`)).toBeVisible()

      await page.getByRole('button', { name: /Deals/i }).click()
      await linkExisting(page, 'Search deals by title…', `Account Hub Deal ${suffix}`, `Account Hub Deal ${suffix}`)
      await expect(page.getByText(`Account Hub Deal ${suffix}`)).toBeVisible()

      await page.goto(`/accounts/${targetAccount.id}`)
      await page.getByRole('button', { name: /Tickets/i }).click()
      await linkExisting(page, 'Search tickets by subject…', `Account Hub Conflict Ticket ${suffix}`, `Account Hub Conflict Ticket ${suffix}`)
      await expect(page.getByText(/whose account does not match this account/i)).toBeVisible()
      await page.getByRole('button', { name: 'Relink contact to account' }).click()
      await expect(page.getByText(`Account Hub Conflict Ticket ${suffix}`)).toBeVisible()

      await page.getByLabel('Unlink ticket').click()
      await expect(page.getByText(`Account Hub Conflict Ticket ${suffix}`)).not.toBeVisible()

      await page.getByRole('button', { name: /Deals/i }).click()
      await page.getByLabel('Unlink deal').click()
      await expect(page.getByText(`Account Hub Deal ${suffix}`)).not.toBeVisible()
    } finally {
      await cleanup(page.request, 'tickets', [ticket])
      await cleanup(page.request, 'deals', [deal])
      await cleanup(page.request, 'contacts', [contact])
      await cleanup(page.request, 'accounts', [sourceAccount, targetAccount])
    }
  })
})
