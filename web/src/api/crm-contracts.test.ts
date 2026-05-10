import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import { accountsApi } from './accounts'
import type { CreateActivityRequest } from './activities'
import { adminSettingsApi } from './adminSettings'
import { billingApi } from './billing'
import { calendarApi } from './calendar'
import { dealsApi } from './deals'
import { inboxApi } from './inbox'
import { integrationsApi } from './integrations'
import { kbApi } from './kb'
import { slaApi } from './sla'
import { server } from '@/test/mocks/server'

describe('CRM API contract mapping', () => {
  it('keeps activity creation restricted to backend-supported types', () => {
    const valid: CreateActivityRequest = { type: 'task', subject: 'Follow up' }
    expect(valid.type).toBe('task')

    // @ts-expect-error note entries are notes, not createable activities.
    const invalid: CreateActivityRequest = { type: 'note', subject: 'Call note' }
    expect(invalid.type).toBe('note')
  })

  it('loads account related lists through existing filterable list endpoints', async () => {
    const accountId = 'account-1'
    const seenPaths: string[] = []
    const seenContactPages: string[] = []

    server.use(
      http.get('/api/v1/contacts', ({ request }) => {
        const url = new URL(request.url)
        seenPaths.push(url.pathname)
        seenContactPages.push(url.searchParams.get('page') ?? '')
        expect(url.searchParams.get('account_id')).toBe(accountId)
        expect(url.searchParams.get('limit')).toBe('200')
        if (url.searchParams.get('page') === '2') {
          return HttpResponse.json({
            data: [{ id: 'contact-2', first_name: 'Grace', last_name: 'Hopper' }],
            meta: { page: 2, per_page: 200, total: 201, total_pages: 2 },
          })
        }
        return HttpResponse.json({
          data: [{ id: 'contact-1', first_name: 'Ada', last_name: 'Lovelace' }],
          meta: { page: 1, per_page: 200, total: 201, total_pages: 2 },
        })
      }),
      http.get('/api/v1/deals', ({ request }) => {
        const url = new URL(request.url)
        seenPaths.push(url.pathname)
        expect(url.searchParams.get('account_id')).toBe(accountId)
        expect(url.searchParams.get('limit')).toBe('200')
        expect(url.searchParams.get('page')).toBe('1')
        return HttpResponse.json({
          data: [],
          meta: { page: 1, per_page: 200, total: 0, total_pages: 0 },
        })
      }),
      http.get('/api/v1/tickets', ({ request }) => {
        const url = new URL(request.url)
        seenPaths.push(url.pathname)
        expect(url.searchParams.get('account_id')).toBe(accountId)
        expect(url.searchParams.get('per_page')).toBe('200')
        expect(url.searchParams.get('page')).toBe('1')
        return HttpResponse.json({
          data: [],
          meta: { page: 1, per_page: 200, total: 0, total_pages: 0 },
        })
      })
    )

    await expect(accountsApi.getContacts(accountId)).resolves.toEqual([
      { id: 'contact-1', first_name: 'Ada', last_name: 'Lovelace' },
      { id: 'contact-2', first_name: 'Grace', last_name: 'Hopper' },
    ])
    await expect(accountsApi.getDeals(accountId)).resolves.toEqual([])
    await expect(accountsApi.getTickets(accountId)).resolves.toEqual([])
    expect(seenContactPages).toEqual(['1', '2'])
    expect(seenPaths).toEqual(['/api/v1/contacts', '/api/v1/contacts', '/api/v1/deals', '/api/v1/tickets'])
  })

  it('maps deal list params to backend query names', async () => {
    server.use(
      http.get('/api/v1/deals', ({ request }) => {
        const url = new URL(request.url)
        expect(url.searchParams.get('q')).toBe('pipeline')
        expect(url.searchParams.get('limit')).toBe('25')
        expect(url.searchParams.get('sort')).toBe('created_at')
        expect(url.searchParams.get('order')).toBe('desc')
        expect(url.searchParams.get('search')).toBeNull()
        expect(url.searchParams.get('per_page')).toBeNull()

        return HttpResponse.json({
          data: [],
          meta: { page: 1, per_page: 25, total: 0, total_pages: 0 },
        })
      })
    )

    await dealsApi.list({
      page: 1,
      per_page: 25,
      search: 'pipeline',
      sort_by: 'created_at',
      sort_dir: 'desc',
    })
  })

  it('uses the thread-first inbox routes and normalizes thread fields', async () => {
    server.use(
      http.get('/api/v1/emails/threads', ({ request }) => {
        const url = new URL(request.url)
        expect(url.searchParams.get('connection_id')).toBe('conn-1')
        expect(url.searchParams.get('unread_only')).toBe('true')

        return HttpResponse.json({
          data: [
            {
              thread_id: 'thread-1',
              org_id: 'org-1',
              connection_id: 'conn-1',
              subject: 'Quarterly review',
              participants: ['alex@company.com', 'buyer@example.com'],
              snippet: 'Following up on the review',
              unread: true,
              message_count: 2,
              last_message_at: '2026-04-15T12:00:00Z',
            },
          ],
          meta: { page: 1, per_page: 50, total: 1, total_pages: 1 },
        })
      })
    )

    const result = await inboxApi.listThreads({
      connection_id: 'conn-1',
      unread_only: true,
    })

    expect(result.data[0]).toMatchObject({
      thread_id: 'thread-1',
      connection_id: 'conn-1',
      unread: true,
    })
  })

  it('serializes inbox send payload with connection_id and normalizes the response', async () => {
    server.use(
      http.post('/api/v1/emails/send', async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>
        expect(body).toMatchObject({
          connection_id: 'conn-1',
          to: ['buyer@example.com'],
          cc: ['vp@example.com'],
          bcc: ['ops@example.com'],
          subject: 'Re: Quarterly review',
          body_html: '<p>Hello</p>',
          thread_id: 'thread-1',
        })

        return HttpResponse.json(
          {
            id: 'msg-1',
            org_id: 'org-1',
            connection_id: 'conn-1',
            thread_id: 'thread-1',
            message_id: 'provider-msg-1',
            direction: 'outbound',
            from_addr: 'alex@company.com',
            to_addrs: ['buyer@example.com'],
            subject: 'Re: Quarterly review',
            body_text: 'Hello',
            body_html: '<p>Hello</p>',
            sent_at: '2026-04-15T12:05:00Z',
            created_at: '2026-04-15T12:05:00Z',
          },
          { status: 201 }
        )
      })
    )

    const message = await inboxApi.sendEmail({
      connection_id: 'conn-1',
      to: ['buyer@example.com'],
      cc: ['vp@example.com'],
      bcc: ['ops@example.com'],
      subject: 'Re: Quarterly review',
      body_html: '<p>Hello</p>',
      thread_id: 'thread-1',
    })

    expect(message).toMatchObject({
      connection_id: 'conn-1',
      thread_id: 'thread-1',
      snippet: 'Hello',
      has_attachments: false,
    })
  })

  it('uses backend-supported routes for SLA, KB suggestions, and billing compatibility', async () => {
    const seen: string[] = []

    server.use(
      http.patch('/api/v1/sla-policies/policy-1', async ({ request }) => {
        seen.push('PATCH /api/v1/sla-policies/policy-1')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.name).toBe('Priority SLA')
        return HttpResponse.json({ id: 'policy-1', name: 'Priority SLA' })
      }),
      http.get('/api/v1/kb/articles/suggest', ({ request }) => {
        const url = new URL(request.url)
        seen.push('GET /api/v1/kb/articles/suggest')
        expect(url.searchParams.get('q')).toBe('password reset')
        return HttpResponse.json([{ id: 'kb-1', title: 'Reset password', slug: 'reset-password' }])
      }),
      http.get('/api/v1/billing/plans', () => {
        seen.push('GET /api/v1/billing/plans')
        return HttpResponse.json([])
      }),
      http.post('/api/v1/billing/subscription/cancel', () => {
        seen.push('POST /api/v1/billing/subscription/cancel')
        return HttpResponse.json({ id: 'sub-1', planTier: 'pro', status: 'cancelled' })
      })
    )

    await slaApi.update('policy-1', { name: 'Priority SLA' })
    await kbApi.suggestArticles('password reset')
    await billingApi.getPlans()
    await billingApi.cancelSubscription()

    expect(seen).toEqual([
      'PATCH /api/v1/sla-policies/policy-1',
      'GET /api/v1/kb/articles/suggest',
      'GET /api/v1/billing/plans',
      'POST /api/v1/billing/subscription/cancel',
    ])
  })

  it('normalizes KB category/suggest payloads when backend wraps arrays in data', async () => {
    server.use(
      http.get('/api/v1/kb/categories', () =>
        HttpResponse.json({
          data: [{ id: 'cat-1', name: 'General', slug: 'general', position: 0 }],
        })
      ),
      http.get('/api/v1/kb/articles/suggest', () =>
        HttpResponse.json({
          data: [{ id: 'kb-1', title: 'Reset password', slug: 'reset-password' }],
        })
      )
    )

    await expect(kbApi.listCategories()).resolves.toEqual([
      { id: 'cat-1', name: 'General', slug: 'general', position: 0 },
    ])
    await expect(kbApi.suggestArticles('reset')).resolves.toEqual([
      { id: 'kb-1', title: 'Reset password', slug: 'reset-password' },
    ])
  })

  it('normalizes legacy billing payloads for subscription, usage, and invoice pagination', async () => {
    const seenPages: number[] = []

    server.use(
      http.get('/api/v1/billing/subscription', () =>
        HttpResponse.json({
          id: 'plan-1',
          plan: 'pro',
          status: 'canceled',
          current_period_end: '2026-06-01T00:00:00Z',
        })
      ),
      http.get('/api/v1/billing/usage', () =>
        HttpResponse.json({
          plan: 'pro',
          usage: {
            users: { used: 7, limit: 10 },
            contacts: { used: 412, limit: 1000 },
            storage_mb: { used: 128, limit: null },
          },
        })
      ),
      http.get('/api/v1/billing/invoices', ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') ?? '1')
        seenPages.push(page)

        if (page === 2) {
          return HttpResponse.json({
            data: [
              {
                id: 'inv-2',
                stripe_invoice_id: 'in_2',
                amount_cents: 9900,
                currency: 'usd',
                status: 'draft',
                created_at: '2026-04-15T00:00:00Z',
              },
            ],
            meta: { page: 2, total_pages: 2 },
          })
        }

        return HttpResponse.json({
          data: [
            {
              id: 'inv-1',
              stripe_invoice_id: 'in_1',
              amount_cents: 4900,
              currency: 'usd',
              status: 'paid',
              pdf_url: 'https://example.test/in_1.pdf',
              created_at: '2026-05-15T00:00:00Z',
            },
          ],
          meta: { page: 1, total_pages: 2 },
        })
      })
    )

    await expect(billingApi.getSubscription()).resolves.toMatchObject({
      id: 'plan-1',
      planTier: 'pro',
      status: 'cancelled',
      currentPeriodEnd: '2026-06-01T00:00:00Z',
      period: 'monthly',
    })

    await expect(billingApi.getUsage()).resolves.toEqual({
      seats: { used: 7, limit: 10 },
      tickets: { used: 412, limit: 1000 },
      apiCalls: { used: 128, limit: null },
    })

    const pageOne = await billingApi.getInvoices()
    expect(pageOne.hasMore).toBe(true)
    expect(pageOne.nextCursor).toBe('2')
    expect(pageOne.data[0]).toMatchObject({
      id: 'inv-1',
      identifier: 'in_1',
      amountCents: 4900,
      currency: 'USD',
      status: 'paid',
      downloadUrl: 'https://example.test/in_1.pdf',
    })

    const pageTwo = await billingApi.getInvoices(pageOne.nextCursor ?? undefined)
    expect(pageTwo.hasMore).toBe(false)
    expect(pageTwo.nextCursor).toBeNull()
    expect(pageTwo.data[0]).toMatchObject({
      id: 'inv-2',
      identifier: 'in_2',
      amountCents: 9900,
      status: 'pending',
    })

    expect(seenPages).toEqual([1, 2])
  })

  it('normalizes integration status from backend snake_case fields', async () => {
    server.use(
      http.get('/api/v1/integrations', () =>
        HttpResponse.json([
          {
            provider: 'gmail',
            status: 'connected',
            email_address: 'ops@acme.test',
            last_synced_at: '2026-05-09T12:00:00Z',
            has_custom_creds: true,
          },
          {
            provider: 'microsoft_teams',
            status: 'coming_soon',
            has_custom_creds: false,
          },
        ])
      )
    )

    await expect(integrationsApi.list()).resolves.toEqual([
      {
        provider: 'gmail',
        status: 'connected',
        emailAddress: 'ops@acme.test',
        lastSyncedAt: '2026-05-09T12:00:00Z',
        hasCustomCreds: true,
      },
      {
        provider: 'microsoft_teams',
        status: 'coming_soon',
        emailAddress: undefined,
        lastSyncedAt: undefined,
        hasCustomCreds: false,
      },
    ])
  })

  it('sends integration credentials with backend snake_case keys', async () => {
    server.use(
      http.put('/api/v1/integrations/gmail/credentials', async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>
        expect(body).toEqual({
          client_id: 'client-1',
          client_secret: 'secret-1',
        })
        expect(body.clientId).toBeUndefined()
        expect(body.clientSecret).toBeUndefined()
        return HttpResponse.json({ ok: true })
      })
    )

    await integrationsApi.saveCredentials('gmail', {
      clientId: 'client-1',
      clientSecret: 'secret-1',
    })
  })

  it('uses backend disconnect/clear routes for integration actions', async () => {
    const seen: string[] = []

    server.use(
      http.delete('/api/v1/integrations/outlook/connection', () => {
        seen.push('DELETE /api/v1/integrations/outlook/connection')
        return new HttpResponse(null, { status: 204 })
      }),
      http.delete('/api/v1/integrations/gmail/credentials', () => {
        seen.push('DELETE /api/v1/integrations/gmail/credentials')
        return new HttpResponse(null, { status: 204 })
      })
    )

    await integrationsApi.disconnect('outlook')
    await integrationsApi.clearCredentials('gmail')

    expect(seen).toEqual([
      'DELETE /api/v1/integrations/outlook/connection',
      'DELETE /api/v1/integrations/gmail/credentials',
    ])
  })

  it('uses calendar integration routes for connection status and sync actions', async () => {
    const seen: string[] = []

    server.use(
      http.get('/api/v1/calendar/connections', () => {
        seen.push('GET /api/v1/calendar/connections')
        return HttpResponse.json({
          data: [{ id: 'cal-1', provider: 'google', token_expiry: '2026-05-10T00:00:00Z' }],
        })
      }),
      http.delete('/api/v1/calendar/connections/cal-1', () => {
        seen.push('DELETE /api/v1/calendar/connections/cal-1')
        return new HttpResponse(null, { status: 204 })
      }),
      http.post('/api/v1/calendar/sync', () => {
        seen.push('POST /api/v1/calendar/sync')
        return HttpResponse.json({ status: 'sync queued' }, { status: 202 })
      })
    )

    const connections = await calendarApi.listConnections()
    expect(connections).toHaveLength(1)
    await calendarApi.disconnect('cal-1')
    await calendarApi.triggerSync()

    expect(seen).toEqual([
      'GET /api/v1/calendar/connections',
      'DELETE /api/v1/calendar/connections/cal-1',
      'POST /api/v1/calendar/sync',
    ])
  })

  it('normalizes null calendar connections payload to an empty list', async () => {
    server.use(
      http.get('/api/v1/calendar/connections', () => {
        return HttpResponse.json({ data: null })
      })
    )

    await expect(calendarApi.listConnections()).resolves.toEqual([])
  })

  it('uses org admin settings routes for bundle-2 configuration surfaces', async () => {
    const seen: string[] = []

    server.use(
      http.get('/api/v1/settings/company', () => {
        seen.push('GET /api/v1/settings/company')
        return HttpResponse.json({ company_name: 'Acme CRM' })
      }),
      http.patch('/api/v1/settings/company', async ({ request }) => {
        seen.push('PATCH /api/v1/settings/company')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.company_name).toBe('Acme Labs')
        expect(body.companyName).toBeUndefined()
        return HttpResponse.json(body)
      }),
      http.patch('/api/v1/settings/outgoing-server', async ({ request }) => {
        seen.push('PATCH /api/v1/settings/outgoing-server')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.smtp_host).toBe('smtp.acme.test')
        expect(body.smtp_password).toBe('rotated-secret')
        return HttpResponse.json({ smtp_password_set: true })
      }),
      http.patch('/api/v1/settings/menu', async ({ request }) => {
        seen.push('PATCH /api/v1/settings/menu')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.menu_config).toEqual({ dashboard: true, deals: false })
        return HttpResponse.json(body)
      })
    )

    await adminSettingsApi.getCompany()
    await adminSettingsApi.updateCompany({ company_name: 'Acme Labs' })
    await adminSettingsApi.updateOutgoingServer({
      smtp_host: 'smtp.acme.test',
      smtp_password: 'rotated-secret',
    })
    await adminSettingsApi.updateMenuConfig({ menu_config: { dashboard: true, deals: false } })

    expect(seen).toEqual([
      'GET /api/v1/settings/company',
      'PATCH /api/v1/settings/company',
      'PATCH /api/v1/settings/outgoing-server',
      'PATCH /api/v1/settings/menu',
    ])
  })
})
