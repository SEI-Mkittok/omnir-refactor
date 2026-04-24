import { beforeEach, describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { DealsPage } from '@/pages/DealsPage'
import { InboxPage } from '@/pages/InboxPage'
import { QuotesPage } from '@/pages/QuotesPage'
import { SequencesPage } from '@/pages/SequencesPage'
import { TicketsPage } from '@/pages/TicketsPage'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'
import { useTicketFilterStore } from '@/stores/ticketFilters'

const contactId = 'contact-1'
const contactName = 'Ada Lovelace'

function paginated<T>(data: T[]) {
  return {
    data,
    meta: {
      page: 1,
      per_page: 50,
      total: data.length,
      total_pages: 1,
    },
  }
}

describe('contact-scoped destination pages', () => {
  beforeEach(() => {
    useTicketFilterStore.getState().reset()
  })

  it('passes contact filters through DealsPage', async () => {
    let requestedContactId: string | null = null

    server.use(
      http.get('/api/v1/deals', ({ request }) => {
        requestedContactId = new URL(request.url).searchParams.get('contact_id')
        return HttpResponse.json(
          paginated([
            {
              id: 'deal-1',
              title: 'Expansion Renewal',
              value_cents: 1250000,
              currency: 'USD',
              stage: 'proposal',
              created_at: '2026-04-20T00:00:00Z',
              updated_at: '2026-04-21T00:00:00Z',
            },
          ])
        )
      })
    )

    render(<DealsPage />, {
      initialRoute: `/deals?contact_id=${contactId}&contact_name=${encodeURIComponent(contactName)}`,
    })

    expect(await screen.findByText(`Contact: ${contactName}`)).toBeInTheDocument()
    expect(await screen.findByText('Expansion Renewal')).toBeInTheDocument()
    await waitFor(() => expect(requestedContactId).toBe(contactId))
  })

  it('passes contact filters through TicketsPage', async () => {
    let requestedContactId: string | null = null

    server.use(
      http.get('/api/v1/reports/tickets', () =>
        HttpResponse.json({
          total_open: 1,
          total_closed: 2,
          avg_resolution_hours: 4,
          breach_rate: 0.1,
        })
      ),
      http.get('/api/v1/tickets', ({ request }) => {
        requestedContactId = new URL(request.url).searchParams.get('contact_id')
        return HttpResponse.json(
          paginated([
            {
              id: 'ticket-1',
              subject: 'Renewal paperwork blocked',
              status: 'open',
              priority: 'high',
              created_at: '2026-04-20T00:00:00Z',
              updated_at: '2026-04-21T00:00:00Z',
              contact: {
                id: contactId,
                name: contactName,
                email: 'ada@acme.test',
              },
            },
          ])
        )
      })
    )

    render(<TicketsPage />, {
      initialRoute: `/tickets?contact_id=${contactId}&contact_name=${encodeURIComponent(contactName)}`,
    })

    expect(await screen.findByText(`Filtered to ${contactName}`)).toBeInTheDocument()
    expect(await screen.findByText('Renewal paperwork blocked')).toBeInTheDocument()
    await waitFor(() => expect(requestedContactId).toBe(contactId))
  })

  it('passes contact filters through QuotesPage', async () => {
    let requestedContactId: string | null = null

    server.use(
      http.get('/api/v1/quotes', ({ request }) => {
        requestedContactId = new URL(request.url).searchParams.get('contact_id')
        return HttpResponse.json(
          paginated([
            {
              id: 'quote-1',
              title: 'Ada Expansion Quote',
              status: 'draft',
              total_cents: 450000,
              currency: 'USD',
              created_at: '2026-04-20T00:00:00Z',
              updated_at: '2026-04-21T00:00:00Z',
              deal_id: 'deal-1',
              line_items: [],
            },
          ])
        )
      })
    )

    render(<QuotesPage />, {
      initialRoute: `/quotes?contact_id=${contactId}&contact_name=${encodeURIComponent(contactName)}`,
    })

    expect(await screen.findByText(`Contact: ${contactName}`)).toBeInTheDocument()
    expect(await screen.findByText('Ada Expansion Quote')).toBeInTheDocument()
    await waitFor(() => expect(requestedContactId).toBe(contactId))
  })

  it('filters SequencesPage to enrollments for the scoped contact', async () => {
    let accountContactsRequested = false

    server.use(
      http.get('/api/v1/sequences', () =>
        HttpResponse.json({
          data: [
            {
              id: 'seq-match',
              name: 'Renewal Follow-up',
              description: 'For active opportunities',
              status: 'active',
              enrolled_count: 3,
              open_rate: 0.42,
              updated_at: '2026-04-22T00:00:00Z',
              steps: [],
            },
            {
              id: 'seq-other',
              name: 'Different Contact Sequence',
              description: 'Should be filtered out',
              status: 'draft',
              enrolled_count: 1,
              open_rate: 0.12,
              updated_at: '2026-04-21T00:00:00Z',
              steps: [],
            },
          ],
          total: 2,
        })
      ),
      http.get('/api/v1/sequences/:id/enrollments', ({ params }) => {
        if (params.id === 'seq-match') {
          return HttpResponse.json({
            data: [
              {
                id: 'enrollment-1',
                contact_id: contactId,
                contact_name: contactName,
                contact_email: 'ada@acme.test',
                status: 'active',
                current_step: 1,
                enrolled_at: '2026-04-20T00:00:00Z',
              },
            ],
          })
        }

        return HttpResponse.json({
          data: [
            {
              id: 'enrollment-2',
              contact_id: 'contact-other',
              contact_name: 'Grace Hopper',
              contact_email: 'grace@acme.test',
              status: 'active',
              current_step: 1,
              enrolled_at: '2026-04-20T00:00:00Z',
            },
          ],
        })
      }),
      http.get('/api/v1/accounts/:id/contacts', () => {
        accountContactsRequested = true
        return HttpResponse.json([])
      })
    )

    render(<SequencesPage />, {
      initialRoute: `/sequences?contact_id=${contactId}&contact_name=${encodeURIComponent(contactName)}`,
    })

    expect(await screen.findByText(`Filtered to ${contactName}`)).toBeInTheDocument()
    expect(await screen.findByText('Renewal Follow-up')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.queryByText('Different Contact Sequence')).not.toBeInTheDocument()
    })
    expect(accountContactsRequested).toBe(false)
  })

  it('passes contact filters through InboxPage', async () => {
    let requestedContactId: string | null = null

    server.use(
      http.get('/api/integrations/email/connections', () =>
        HttpResponse.json([
          {
            id: 'conn-1',
            provider: 'gmail',
            email_address: 'owner@acme.test',
            display_name: 'Owner',
          },
        ])
      ),
      http.get('/api/v1/emails/threads', ({ request }) => {
        requestedContactId = new URL(request.url).searchParams.get('contact_id')
        return HttpResponse.json(
          paginated([
            {
              thread_id: 'thread-1',
              org_id: 'org-1',
              connection_id: 'conn-1',
              subject: 'Ada onboarding thread',
              participants: ['ada@acme.test', 'owner@acme.test'],
              snippet: 'Latest update from Ada',
              unread: true,
              message_count: 2,
              last_message_at: '2026-04-22T00:00:00Z',
              contact_id: contactId,
            },
          ])
        )
      })
    )

    render(<InboxPage />, {
      initialRoute: `/inbox?contact_id=${contactId}&contact_name=${encodeURIComponent(contactName)}`,
    })

    expect(await screen.findByText(`Filtered to ${contactName}`)).toBeInTheDocument()
    expect(await screen.findByText('Ada onboarding thread')).toBeInTheDocument()
    await waitFor(() => expect(requestedContactId).toBe(contactId))
  })
})
