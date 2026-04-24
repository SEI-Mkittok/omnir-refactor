import { Route, Routes } from 'react-router-dom'
import { delay, http, HttpResponse } from 'msw'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ContactDetailPage } from '@/pages/ContactDetailPage'
import { render, screen, waitFor, within } from '@/test/utils'
import { server } from '@/test/mocks/server'

const contactId = 'contact-1'
const accountId = 'account-1'
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

function installBaseHandlers() {
  server.use(
    http.get('/api/v1/contacts/:id', () =>
      HttpResponse.json({
        ...baseContact,
        account_id: linkedAccount.id,
        account: linkedAccount,
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
      return HttpResponse.json(paginated([]))
    }),
    http.get('/api/v1/tickets', () => HttpResponse.json(paginated([]))),
    http.get('/api/v1/sequences', () => HttpResponse.json({ data: [], total: 0 })),
    http.get('/api/v1/emails/threads', () => HttpResponse.json(paginated([])))
  )
}

describe('ContactDetailPage optimistic association flows', () => {
  it('seeds relationship rows from linked accounts and preserves exactly one primary after interaction', async () => {
    installBaseHandlers()

    const user = userEvent.setup()
    renderPage()
    await screen.findAllByRole('region', { name: /relationship roles/i })

    await waitFor(() => {
      expect(within(getRelationshipEditor()).getAllByLabelText(/acme corp is primary/i)).not.toHaveLength(0)
    })

    await user.click(within(getRelationshipEditor()).getByRole('button', { name: /add account role/i }))

    const editor = getRelationshipEditor()
    const accountInputs = within(editor).getAllByPlaceholderText('Enter account name')
    await user.type(accountInputs[1], 'Beta Corp')

    const detailInputs = within(editor).getAllByPlaceholderText('Title, email, or context')
    await user.type(detailInputs[1], 'Regional partner')

    await user.click(within(getRelationshipEditor()).getByRole('button', { name: /make beta corp primary/i }))

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
    await user.click(await screen.findByRole('button', { name: /enterprise renewal/i }))

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
})

