import { Route, Routes } from 'react-router-dom'
import { delay, http, HttpResponse } from 'msw'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ContactDetailPage } from '@/pages/ContactDetailPage'
import { render, screen, waitFor, within } from '@/test/utils'
import { server } from '@/test/mocks/server'

const contactId = 'contact-1'
const accountId = 'account-1'
const otherAccountId = 'account-2'
const searchDealId = 'deal-search-1'
const ticketId = 'ticket-1'

const linkedAccount = {
  id: accountId,
  name: 'Acme Corp',
  industry: 'Technology',
  domain: 'acme.test',
  created_at: '2026-04-20T00:00:00Z',
  updated_at: '2026-04-20T00:00:00Z',
}

const otherAccount = {
  id: otherAccountId,
  name: 'Beta Corp',
  industry: 'Finance',
  domain: 'beta.test',
  created_at: '2026-04-20T00:00:00Z',
  updated_at: '2026-04-20T00:00:00Z',
}

const baseContact = {
  id: contactId,
  first_name: 'Ada',
  last_name: 'Lovelace',
  email: 'ada@acme.test',
  title: 'CTO',
  stage: 'customer',
  created_at: '2026-04-20T00:00:00Z',
  updated_at: '2026-04-20T00:00:00Z',
}

function paginated<T>(data: T[]) {
  return {
    data,
    meta: {
      page: 1,
      per_page: 100,
      total: data.length,
      total_pages: 1,
    },
  }
}

function renderPage() {
  return render(
    <Routes>
      <Route path="/contacts/:id" element={<ContactDetailPage />} />
    </Routes>,
    { initialRoute: `/contacts/${contactId}` }
  )
}

function getRelationshipEditor(): HTMLElement {
  const editors = screen.getAllByRole('region', { name: /relationship roles/i })
  return editors[editors.length - 1] as HTMLElement
}

function getLastRelationshipButton(name: RegExp): HTMLElement {
  const buttons = within(getRelationshipEditor()).getAllByRole('button', { name })
  return buttons[buttons.length - 1] as HTMLElement
}

function installBaseHandlers() {
  let currentLinkedAccount: typeof linkedAccount | null = linkedAccount
  let linkedDeals: Array<Record<string, unknown>> = []
  let linkedTickets: Array<Record<string, unknown>> = []

  server.use(
    http.get('/api/v1/contacts/:id', () =>
      HttpResponse.json({
        ...baseContact,
        account_id: currentLinkedAccount?.id,
        account: currentLinkedAccount ?? undefined,
        linked_accounts: currentLinkedAccount ? [currentLinkedAccount] : [],
      })
    ),
    http.get('/api/v1/contacts/:id/notes', () => HttpResponse.json([])),
    http.get('/api/v1/activities', () => HttpResponse.json(paginated([]))),
    http.get('/api/v1/deals/:id/quotes', () => HttpResponse.json(paginated([]))),
    http.get('/api/v1/deals', ({ request }) => {
      const url = new URL(request.url)
      if (url.searchParams.get('q')) {
        return HttpResponse.json(
          paginated([
            {
              id: searchDealId,
              title: 'Enterprise Renewal',
              value_cents: 1250000,
              currency: 'USD',
              stage: 'proposal',
              created_at: '2026-04-20T00:00:00Z',
              updated_at: '2026-04-21T00:00:00Z',
            },
          ])
        )
      }
      return HttpResponse.json(paginated(linkedDeals))
    }),
    http.get('/api/v1/accounts', ({ request }) => {
      const url = new URL(request.url)
      if (!url.searchParams.get('q')) return HttpResponse.json(paginated([]))

      return HttpResponse.json(paginated([linkedAccount]))
    }),
    http.get('/api/v1/tickets', () => HttpResponse.json(paginated(linkedTickets))),
    http.patch('/api/v1/contacts/:id', async ({ request }) => {
      const body = (await request.json()) as { account_id?: string }
      await delay(50)
      currentLinkedAccount = body.account_id === linkedAccount.id ? linkedAccount : null
      return HttpResponse.json({
        ...baseContact,
        account_id: currentLinkedAccount?.id,
        account: currentLinkedAccount ?? undefined,
        linked_accounts: currentLinkedAccount ? [currentLinkedAccount] : [],
      })
    }),
    http.patch('/api/v1/deals/:id', async ({ request }) => {
      const body = (await request.json()) as { contact_id?: string }
      await delay(50)
      linkedDeals = body.contact_id
        ? [
            {
              id: searchDealId,
              title: 'Enterprise Renewal',
              value_cents: 1250000,
              currency: 'USD',
              stage: 'proposal',
              contact_id: contactId,
              created_at: '2026-04-20T00:00:00Z',
              updated_at: '2026-04-21T00:00:00Z',
            },
          ]
        : []
      return HttpResponse.json(linkedDeals[0] ?? { id: searchDealId })
    }),
    http.patch('/api/v1/tickets/:id/contact', async ({ request }) => {
      const body = (await request.json()) as { contact_id?: string | null }
      await delay(50)
      linkedTickets = body.contact_id
        ? [
            {
              id: ticketId,
              subject: 'Billing portal outage',
              status: 'open',
              priority: 'high',
              contact: {
                id: contactId,
                name: 'Ada Lovelace',
                email: 'ada@acme.test',
              },
              created_at: '2026-04-20T00:00:00Z',
              updated_at: '2026-04-21T00:00:00Z',
            },
          ]
        : []
      return HttpResponse.json(linkedTickets[0] ?? { id: ticketId })
    }),
    http.get('/api/v1/sequences', () => HttpResponse.json({ data: [], total: 0 })),
    http.get('/api/v1/emails/threads', () => HttpResponse.json(paginated([])))
  )
}

describe('ContactDetailPage optimistic association flows', () => {
  it('optimistically links and unlinks an account on the happy path', async () => {
    installBaseHandlers()
    server.use(
      http.get('/api/v1/contacts/:id', () =>
        HttpResponse.json({
          ...baseContact,
          linked_accounts: [],
        })
      )
    )

    let linked = false
    server.use(
      http.get('/api/v1/contacts/:id', () =>
        HttpResponse.json({
          ...baseContact,
          account_id: linked ? linkedAccount.id : undefined,
          account: linked ? linkedAccount : undefined,
          linked_accounts: linked ? [linkedAccount] : [],
        })
      ),
      http.patch('/api/v1/contacts/:id', async ({ request }) => {
        const body = (await request.json()) as { account_id?: string }
        await delay(100)
        linked = body.account_id === linkedAccount.id
        return HttpResponse.json({
          ...baseContact,
          account_id: linked ? linkedAccount.id : undefined,
          account: linked ? linkedAccount : undefined,
          linked_accounts: linked ? [linkedAccount] : [],
        })
      })
    )

    const user = userEvent.setup()
    renderPage()

    expect((await screen.findAllByText('No accounts linked yet.')).length).toBeGreaterThan(0)
    await user.click(screen.getAllByRole('button', { name: /link existing/i })[0])
    await user.type(screen.getByPlaceholderText('Search accounts by name…'), 'acme')
    await user.click((await screen.findByText('Acme Corp')).closest('button') as HTMLElement)

    expect((await screen.findAllByText('Acme Corp')).length).toBeGreaterThan(0)
    expect(await screen.findByText('Account linked')).toBeInTheDocument()

    await user.click((await screen.findAllByLabelText('Unlink account'))[0])
    await waitFor(() => {
      expect(screen.queryAllByLabelText('Unlink account')).toHaveLength(0)
    })
    expect(await screen.findByText('Account unlinked')).toBeInTheDocument()
    expect((await screen.findAllByText('No accounts linked yet.')).length).toBeGreaterThan(0)
  })

  it('optimistically links and unlinks a deal on the happy path', async () => {
    installBaseHandlers()

    const user = userEvent.setup()
    renderPage()

    await user.click((await screen.findAllByRole('button', { name: /deals/i }))[0])
    expect((await screen.findAllByText('No deals linked yet.')).length).toBeGreaterThan(0)

    await user.click(screen.getAllByRole('button', { name: /link existing/i })[0])
    await user.type(screen.getByPlaceholderText('Search deals by title…'), 'renewal')
    await user.click((await screen.findByText(/enterprise renewal/i)).closest('button') as HTMLElement)

    expect(await screen.findByText('Enterprise Renewal')).toBeInTheDocument()
    expect(await screen.findByText('Deal linked')).toBeInTheDocument()

    await user.click((await screen.findAllByLabelText('Unlink deal'))[0])
    await waitFor(() => {
      expect(screen.queryAllByLabelText('Unlink deal')).toHaveLength(0)
    })
    expect(await screen.findByText('Deal unlinked')).toBeInTheDocument()
    expect((await screen.findAllByText('No deals linked yet.')).length).toBeGreaterThan(0)
  })

  it('seeds relationship rows from linked accounts and preserves exactly one primary after interaction', async () => {
    installBaseHandlers()

    const user = userEvent.setup()
    renderPage()
    await screen.findAllByRole('region', { name: /relationship roles/i })

    await waitFor(() => {
      expect(within(getRelationshipEditor()).getAllByLabelText(/acme corp is primary/i)).not.toHaveLength(0)
    })

    await user.click(getLastRelationshipButton(/add account role/i))

    const editor = getRelationshipEditor()
    const accountInputs = within(editor).getAllByPlaceholderText('Enter account name')
    await user.type(accountInputs[1], 'Beta Corp')

    const detailInputs = within(editor).getAllByPlaceholderText('Title, email, or context')
    await user.type(detailInputs[1], 'Regional partner')

    await user.click(getLastRelationshipButton(/make beta corp primary/i))

    expect((await screen.findAllByDisplayValue('Beta Corp')).length).toBeGreaterThan(0)
    expect((await screen.findAllByRole('button', { name: /beta corp is primary/i })).length).toBeGreaterThan(0)
    expect((await screen.findAllByRole('button', { name: /make acme corp primary/i })).length).toBeGreaterThan(0)
  })

  it('rolls back account unlink when the mutation fails', async () => {
    installBaseHandlers()
    server.use(
      http.patch('/api/v1/contacts/:id', async () => {
        await delay(100)
        return HttpResponse.json({ error: 'failed' }, { status: 500 })
      })
    )

    const user = userEvent.setup()
    renderPage()

    expect((await screen.findAllByText('Acme Corp')).length).toBeGreaterThan(0)

    await user.click((await screen.findAllByLabelText('Unlink account'))[0])

    await waitFor(() => {
      expect(screen.queryAllByLabelText('Unlink account')).toHaveLength(0)
    })
    expect(await screen.findByText('Could not unlink account')).toBeInTheDocument()
    expect((await screen.findAllByText('Acme Corp')).length).toBeGreaterThan(0)
  })

  it('rolls back deal link-existing when the mutation fails', async () => {
    installBaseHandlers()
    server.use(
      http.patch('/api/v1/deals/:id', async () => {
        await delay(100)
        return HttpResponse.json({ error: 'failed' }, { status: 500 })
      })
    )

    const user = userEvent.setup()
    renderPage()

    await user.click((await screen.findAllByRole('button', { name: /deals/i }))[0])
    expect((await screen.findAllByText('No deals linked yet.')).length).toBeGreaterThan(0)

    await user.click(screen.getAllByRole('button', { name: /link existing/i })[0])
    await user.type(screen.getByPlaceholderText('Search deals by title…'), 'renewal')
    const dealResult = await screen.findByText(/enterprise renewal/i)
    await user.click(dealResult.closest('button') as HTMLElement)

    expect(await screen.findByText('Enterprise Renewal')).toBeInTheDocument()
    expect(await screen.findByText('Could not link deal')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.queryAllByLabelText('Unlink deal')).toHaveLength(0)
    })
    expect((await screen.findAllByText('No deals linked yet.')).length).toBeGreaterThan(0)
  })

  it('rolls back ticket unlink when the mutation fails', async () => {
    installBaseHandlers()
    server.use(
      http.get('/api/v1/tickets', () =>
        HttpResponse.json(
          paginated([
            {
              id: ticketId,
              subject: 'Billing portal outage',
              status: 'open',
              priority: 'high',
              created_at: '2026-04-20T00:00:00Z',
              updated_at: '2026-04-21T00:00:00Z',
              contact: {
                id: contactId,
                name: 'Ada Lovelace',
                email: 'ada@acme.test',
              },
            },
          ])
        )
      ),
      http.patch('/api/v1/tickets/:id/contact', async () => {
        await delay(100)
        return HttpResponse.json({ error: 'failed' }, { status: 500 })
      })
    )

    const user = userEvent.setup()
    renderPage()

    await user.click((await screen.findAllByRole('button', { name: /tickets/i }))[0])
    expect(await screen.findByText('Billing portal outage')).toBeInTheDocument()

    await user.click((await screen.findAllByLabelText('Unlink ticket'))[0])

    expect((await screen.findAllByText('No tickets linked yet.')).length).toBeGreaterThan(0)
    expect(await screen.findByText('Could not unlink ticket')).toBeInTheDocument()
    expect((await screen.findAllByText('Billing portal outage')).length).toBeGreaterThan(0)
  })

  it('keeps the ticket link modal open after a mismatch response', async () => {
    installBaseHandlers()
    let ticketAttempts = 0
    server.use(
      http.get('/api/v1/tickets', ({ request }) => {
        const url = new URL(request.url)
        if (!url.searchParams.get('search')) return HttpResponse.json(paginated([]))
        return HttpResponse.json(
          paginated([
            {
              id: ticketId,
              subject: 'Account scoped ticket',
              status: 'open',
              priority: 'high',
              account: otherAccount,
              created_at: '2026-04-20T00:00:00Z',
              updated_at: '2026-04-21T00:00:00Z',
            },
          ])
        )
      }),
      http.patch('/api/v1/tickets/:id/contact', () => {
        ticketAttempts += 1
        if (ticketAttempts === 1) {
          return HttpResponse.json(
            {
              error: 'contact_id is not related to account_id',
              code: 'validation_error',
              details: [{ field: 'contact_id', message: 'contact_id is not related to account_id' }],
            },
            { status: 422 }
          )
        }
        return HttpResponse.json({
          id: ticketId,
          subject: 'Account scoped ticket',
          status: 'open',
          priority: 'high',
          contact: { id: contactId, name: 'Ada Lovelace', email: 'ada@acme.test' },
          account: otherAccount,
        })
      })
    )

    const user = userEvent.setup()
    renderPage()

    await user.click((await screen.findAllByRole('button', { name: /tickets/i }))[0])
    await user.click(screen.getAllByRole('button', { name: /link existing/i })[0])
    await user.type(screen.getByPlaceholderText('Search tickets by subject…'), 'scoped')
    await user.click((await screen.findByText('Account scoped ticket')).closest('button') as HTMLElement)

    await waitFor(() => {
      expect(ticketAttempts).toBe(1)
    })
    expect(screen.getByRole('heading', { name: 'Link Ticket' })).toBeInTheDocument()
  })
})

