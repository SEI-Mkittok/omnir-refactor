import { describe, expect, it } from 'vitest'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'

import { AccountsPage } from './AccountsPage'
import { ContactsPage } from './ContactsPage'
import { DealsPage } from './DealsPage'
import { LeadsPage } from './LeadsPage'
import { QuotesPage } from './QuotesPage'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'
import type { SavedView, ViewEntityType } from '@/api/types'

const now = '2026-05-15T00:00:00Z'

function paginated<T>(data: T[]) {
  return {
    data,
    meta: { page: 1, per_page: 50, total: data.length, total_pages: data.length ? 1 : 0 },
  }
}

function savedView(entityType: ViewEntityType, filters: SavedView['filters']): SavedView {
  return {
    id: `view-${entityType}`,
    org_id: 'org-1',
    created_by: 'user-1',
    entity_type: entityType,
    name: `QA ${entityType} View`,
    filters,
    is_shared: false,
    is_pinned: true,
    pin_order: 0,
    created_at: now,
    updated_at: now,
  }
}

describe('saved view restore QA', () => {
  it('restores pinned saved-view filters on CRM list pages', async () => {
    const user = userEvent.setup()
    const requested = new Map<string, URLSearchParams>()
    const views: Partial<Record<ViewEntityType, SavedView[]>> = {
      contacts: [savedView('contacts', { search: 'Ada', stage: 'customer' })],
      accounts: [savedView('accounts', { search: 'Acme', industry: 'Software' })],
      leads: [savedView('leads', { search: 'Grace', status: 'contacted' })],
      deals: [savedView('deals', { search: 'Enterprise', stage: 'qualified' })],
      quotes: [savedView('quotes', { search: 'Renewal', status: 'sent' })],
    }

    server.use(
      http.get('/api/v1/views', ({ request }) => {
        const entityType = new URL(request.url).searchParams.get('entity_type') as ViewEntityType
        return HttpResponse.json(views[entityType] ?? [])
      }),
      http.get('/api/v1/contacts', ({ request }) => {
        requested.set('contacts', new URL(request.url).searchParams)
        return HttpResponse.json(paginated([{
          id: 'contact-1',
          first_name: 'Ada',
          last_name: 'Lovelace',
          email: 'ada@acme.test',
          stage: 'customer',
          created_at: now,
          updated_at: now,
        }]))
      }),
      http.get('/api/v1/accounts', ({ request }) => {
        requested.set('accounts', new URL(request.url).searchParams)
        return HttpResponse.json(paginated([{
          id: 'account-1',
          name: 'Acme Software',
          industry: 'Software',
          website: 'https://acme.test',
          created_at: now,
          updated_at: now,
        }]))
      }),
      http.get('/api/v1/leads', ({ request }) => {
        requested.set('leads', new URL(request.url).searchParams)
        return HttpResponse.json(paginated([{
          id: 'lead-1',
          first_name: 'Grace',
          last_name: 'Hopper',
          email: 'grace@acme.test',
          company: 'Acme',
          lead_score: 88,
          status: 'contacted',
          created_at: now,
          updated_at: now,
        }]))
      }),
      http.get('/api/v1/deals', ({ request }) => {
        requested.set('deals', new URL(request.url).searchParams)
        return HttpResponse.json(paginated([{
          id: 'deal-1',
          title: 'Enterprise Expansion',
          value_cents: 1250000,
          currency: 'USD',
          stage: 'qualified',
          created_at: now,
          updated_at: now,
        }]))
      }),
      http.get('/api/v1/quotes', ({ request }) => {
        requested.set('quotes', new URL(request.url).searchParams)
        return HttpResponse.json(paginated([{
          id: 'quote-1',
          title: 'Renewal Quote',
          status: 'sent',
          total_cents: 120000,
          currency: 'USD',
          line_items: [],
          created_at: now,
          updated_at: now,
        }]))
      })
    )

    const cases = [
      { entityType: 'contacts', page: <ContactsPage />, route: '/contacts', view: 'QA contacts View', row: 'Ada Lovelace', query: { q: 'Ada', stage: 'customer' } },
      { entityType: 'accounts', page: <AccountsPage />, route: '/accounts', view: 'QA accounts View', row: 'Acme Software', query: { q: 'Acme', industry: 'Software' } },
      { entityType: 'leads', page: <LeadsPage />, route: '/leads', view: 'QA leads View', row: 'Grace Hopper', query: { search: 'Grace', status: 'contacted' } },
      { entityType: 'deals', page: <DealsPage />, route: '/deals', view: 'QA deals View', row: 'Enterprise Expansion', query: { q: 'Enterprise', stage: 'qualified' } },
      { entityType: 'quotes', page: <QuotesPage />, route: '/quotes', view: 'QA quotes View', row: 'Renewal Quote', query: { q: 'Renewal', status: 'sent' } },
    ] as const

    for (const item of cases) {
      const rendered = render(item.page, { initialRoute: item.route })
      await user.click(await screen.findByRole('tab', { name: item.view }))
      expect(await screen.findByText(item.row)).toBeInTheDocument()

      await waitFor(() => {
        const params = requested.get(item.entityType)
        expect(params).toBeTruthy()
        for (const [key, value] of Object.entries(item.query)) {
          expect(params?.get(key)).toBe(value)
        }
      })

      rendered.unmount()
      requested.delete(item.entityType)
    }
  })
})
