import { Route, Routes } from 'react-router-dom'
import { delay, http, HttpResponse } from 'msw'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { AccountDetailPage } from '@/pages/AccountDetailPage'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'

const accountId = 'account-1'
const contactId = 'contact-1'
const dealId = 'deal-1'
const ticketId = 'ticket-1'
const mismatchContactId = 'contact-2'

const account = {
  id: accountId,
  name: 'Acme Corp',
  industry: 'technology',
  domain: 'acme.test',
  created_at: '2026-04-20T00:00:00Z',
  updated_at: '2026-04-21T00:00:00Z',
}

const linkedContact = {
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
      per_page: 10,
      total: data.length,
      total_pages: 1,
    },
  }
}

function renderPage() {
  return render(
    <Routes>
      <Route path="/accounts/:id" element={<AccountDetailPage />} />
    </Routes>,
    { initialRoute: `/accounts/${accountId}` }
  )
}

function installBaseHandlers() {
  server.use(
    http.get('/api/v1/accounts/:id', () => HttpResponse.json(account)),
    http.get('/api/v1/accounts/:id/contacts', () => HttpResponse.json([])),
    http.get('/api/v1/accounts/:id/deals', () => HttpResponse.json([])),
    http.get('/api/v1/accounts/:id/tickets', () => HttpResponse.json([])),
    http.get('/api/v1/accounts/:id/notes', () => HttpResponse.json({ data: [] })),
    http.get('/api/v1/accounts/:id/attachments', () => HttpResponse.json([])),
    http.get('/api/v1/activities', () => HttpResponse.json(paginated([]))),
    http.get('/api/v1/quotes', () => HttpResponse.json(paginated([]))),
    http.get('/api/v1/sequences', () => HttpResponse.json({ data: [] })),
    http.get('/api/v1/emails/threads', () => HttpResponse.json(paginated([]))),
    http.get('/api/v1/contacts', ({ request }) => {
      const url = new URL(request.url)
      const q = url.searchParams.get('q')
      if (!q) return HttpResponse.json(paginated([]))

      return HttpResponse.json(
        paginated([
          {
            id: contactId,
            first_name: 'Ada',
            last_name: 'Lovelace',
            email: 'ada@acme.test',
            title: 'CTO',
            created_at: '2026-04-20T00:00:00Z',
            updated_at: '2026-04-20T00:00:00Z',
          },
        ])
      )
    }),
    http.get('/api/v1/deals', ({ request }) => {
      const url = new URL(request.url)
      const q = url.searchParams.get('q')
      if (!q) return HttpResponse.json(paginated([]))

      return HttpResponse.json(
        paginated([
          {
            id: dealId,
            title: 'Enterprise Renewal',
            value_cents: 1250000,
            currency: 'USD',
            stage: 'proposal',
            created_at: '2026-04-20T00:00:00Z',
            updated_at: '2026-04-20T00:00:00Z',
          },
        ])
      )
    }),
    http.get('/api/v1/tickets', ({ request }) => {
      const url = new URL(request.url)
      const search = url.searchParams.get('search')
      if (!search) return HttpResponse.json(paginated([]))

      return HttpResponse.json(
        paginated([
          {
            id: ticketId,
            subject: 'Billing portal outage',
            status: 'open',
            priority: 'high',
            created_at: '2026-04-20T00:00:00Z',
            updated_at: '2026-04-20T00:00:00Z',
          },
        ])
      )
    }),
    http.patch('/api/v1/contacts/:id', () =>
      HttpResponse.json({
        id: contactId,
        first_name: 'Ada',
        last_name: 'Lovelace',
        email: 'ada@acme.test',
        account_id: accountId,
      })
    ),
    http.patch('/api/v1/deals/:id', () =>
      HttpResponse.json({
        id: dealId,
        title: 'Enterprise Renewal',
        account_id: accountId,
      })
    ),
    http.patch('/api/v1/tickets/:id', () =>
      HttpResponse.json({
        id: ticketId,
        subject: 'Billing portal outage',
        account_id: accountId,
      })
    )
  )
}

describe('AccountDetailPage link modals', () => {
  it('supports contact, deal, and ticket linking from the linked entities section', async () => {
    installBaseHandlers()
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Contacts')
    const getLinkButton = () => screen.getByRole('button', { name: 'Link' })

    await user.click(getLinkButton())
    expect(await screen.findByRole('heading', { name: 'Link Contact' })).toBeInTheDocument()
    await user.type(screen.getByPlaceholderText('Search by name or email…'), 'ada')
    const contactResult = await screen.findByText(/ada lovelace/i)
    await user.click(contactResult.closest('button') as HTMLElement)
    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'Link Contact' })).not.toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /Deals 0/i }))
    await user.click(getLinkButton())
    expect(await screen.findByRole('heading', { name: 'Link Deal' })).toBeInTheDocument()
    await user.type(screen.getByPlaceholderText('Search deals by title…'), 'renewal')
    const dealResult = await screen.findByText(/enterprise renewal/i)
    await user.click(dealResult.closest('button') as HTMLElement)
    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'Link Deal' })).not.toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /Tickets 0/i }))
    await user.click(getLinkButton())
    expect(await screen.findByRole('heading', { name: 'Link Ticket' })).toBeInTheDocument()
    await user.type(screen.getByPlaceholderText('Search tickets by subject…'), 'billing')
    const ticketResult = await screen.findByText(/billing portal outage/i)
    await user.click(ticketResult.closest('button') as HTMLElement)
    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'Link Ticket' })).not.toBeInTheDocument()
    })
  })

  it('keeps the modal open when a link mutation fails', async () => {
    installBaseHandlers()
    server.use(
      http.patch('/api/v1/contacts/:id', () =>
        HttpResponse.json({ error: 'failed' }, { status: 500 })
      )
    )

    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Contacts')
    const linkButton = await screen.findByRole('button', { name: 'Link' })

    await user.click(linkButton)
    await user.type(screen.getByPlaceholderText('Search by name or email…'), 'ada')
    const contactResult = await screen.findByText(/ada lovelace/i)
    await user.click(contactResult.closest('button') as HTMLElement)

    expect(await screen.findByRole('heading', { name: 'Link Contact' })).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Search by name or email…')).toBeInTheDocument()
  })

  it('keeps linked contact visible when unlink mutation fails', async () => {
    installBaseHandlers()
    server.use(
      http.get('/api/v1/accounts/:id/contacts', () => HttpResponse.json([linkedContact])),
      http.patch('/api/v1/contacts/:id', async () => {
        await delay(100)
        return HttpResponse.json({ error: 'failed' }, { status: 500 })
      })
    )

    const user = userEvent.setup()
    renderPage()

    expect(await screen.findByText('Ada Lovelace')).toBeInTheDocument()
    await user.click((await screen.findAllByLabelText('Unlink contact'))[0])

    expect(await screen.findByText('Ada Lovelace')).toBeInTheDocument()
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
              subject: 'Contact scoped ticket',
              status: 'open',
              priority: 'high',
              contact: {
                id: mismatchContactId,
                name: 'Grace Hopper',
                email: 'grace@other.test',
              },
              created_at: '2026-04-20T00:00:00Z',
              updated_at: '2026-04-20T00:00:00Z',
            },
          ])
        )
      }),
      http.patch('/api/v1/tickets/:id', () => {
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
          subject: 'Contact scoped ticket',
          status: 'open',
          priority: 'high',
          account_id: accountId,
        })
      })
    )

    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Contacts')
    await user.click(screen.getByRole('button', { name: /Tickets 0/i }))
    await user.click(screen.getByRole('button', { name: 'Link' }))
    await user.type(screen.getByPlaceholderText('Search tickets by subject…'), 'scoped')
    await user.click((await screen.findByText('Contact scoped ticket')).closest('button') as HTMLElement)

    await waitFor(() => {
      expect(ticketAttempts).toBe(1)
    })
    expect(screen.getByRole('heading', { name: 'Link Ticket' })).toBeInTheDocument()
  })

  it('seeds relationship rows from linked contacts', async () => {
    installBaseHandlers()
    server.use(
      http.get('/api/v1/accounts/:id/contacts', () =>
        HttpResponse.json([
          {
            id: contactId,
            first_name: 'Ada',
            last_name: 'Lovelace',
            email: 'ada@acme.test',
            title: 'CTO',
            stage: 'customer',
            created_at: '2026-04-20T00:00:00Z',
            updated_at: '2026-04-20T00:00:00Z',
          },
          {
            id: 'contact-2',
            first_name: 'Grace',
            last_name: 'Hopper',
            email: 'grace@acme.test',
            title: 'Finance Lead',
            stage: 'customer',
            created_at: '2026-04-20T00:00:00Z',
            updated_at: '2026-04-20T00:00:00Z',
          },
        ])
      )
    )

    renderPage()

    expect(await screen.findByRole('button', { name: /ada lovelace is primary/i })).toBeInTheDocument()
    expect(screen.getByDisplayValue('Grace Hopper')).toBeInTheDocument()
    expect(screen.getAllByLabelText('Role')).toHaveLength(2)
  })
})





