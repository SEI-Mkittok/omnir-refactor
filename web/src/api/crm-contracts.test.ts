import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import { accountsApi } from './accounts'
import { accessSettingsApi } from './accessSettings'
import type { CreateActivityRequest } from './activities'
import { adminSettingsApi } from './adminSettings'
import { billingApi } from './billing'
import { calendarApi } from './calendar'
import { currenciesApi } from './currencies'
import { dealsApi } from './deals'
import { inboxApi } from './inbox'
import { integrationsApi } from './integrations'
import { kbApi } from './kb'
import { leadConversionMappingApi } from './leadConversionMapping'
import { numberingApi } from './numbering'
import { picklistsApi } from './picklists'
import { preferencesApi } from './preferences'
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

  it('uses bundle-3 settings routes and snake_case payload contracts', async () => {
    const seen: string[] = []

    server.use(
      http.get('/api/v1/settings/numbering', () => {
        seen.push('GET /api/v1/settings/numbering')
        return HttpResponse.json({
          org_id: 'org-1',
          quote_number_start: 100,
          ticket_number_start: 200,
          kb_article_number_start: 300,
          invoice_number_start: 400,
          quote_number_prefix: 'QUO',
          ticket_number_prefix: 'TCK',
          kb_article_number_prefix: 'KB',
          invoice_number_prefix: 'INV',
          quote_number_current: 100,
          ticket_number_current: 200,
          kb_article_number_current: 300,
          invoice_number_current: 400,
        })
      }),
      http.patch('/api/v1/settings/numbering', async ({ request }) => {
        seen.push('PATCH /api/v1/settings/numbering')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.quote_number_prefix).toBe('QTE')
        expect(body.quote_number_current).toBe(515)
        return HttpResponse.json(body)
      }),
      http.get('/api/v1/settings/preferences', () => {
        seen.push('GET /api/v1/settings/preferences')
        return HttpResponse.json({ default_currency: 'USD', landing_page: '/dashboard' })
      }),
      http.patch('/api/v1/settings/preferences', async ({ request }) => {
        seen.push('PATCH /api/v1/settings/preferences')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.default_currency).toBe('EUR')
        expect(body.defaultCurrency).toBeUndefined()
        return HttpResponse.json(body)
      }),
      http.get('/api/v1/settings/calendar-preferences', () => {
        seen.push('GET /api/v1/settings/calendar-preferences')
        return HttpResponse.json({ calendar_default_view: 'week', calendar_show_completed_events: false })
      }),
      http.get('/api/v1/settings/currencies', () => {
        seen.push('GET /api/v1/settings/currencies')
        return HttpResponse.json({
          default_code: 'USD',
          currencies: [{ code: 'USD', display_name: 'US Dollar', symbol: '$', decimal_places: 2, is_active: true, is_default: true }],
        })
      }),
      http.patch('/api/v1/settings/currencies', async ({ request }) => {
        seen.push('PATCH /api/v1/settings/currencies')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.default_code).toBe('EUR')
        return HttpResponse.json({
          default_code: 'EUR',
          currencies: body.currencies,
        })
      }),
      http.get('/api/v1/settings/picklists', () => {
        seen.push('GET /api/v1/settings/picklists')
        return HttpResponse.json({ data: [] })
      }),
      http.put('/api/v1/settings/picklists/field-1/values', async ({ request }) => {
        seen.push('PUT /api/v1/settings/picklists/field-1/values')
        const body = (await request.json()) as Record<string, unknown>
        expect(Array.isArray(body.values)).toBe(true)
        return HttpResponse.json({ values: body.values })
      }),
      http.post('/api/v1/settings/picklists/field-1/remap-delete', async ({ request }) => {
        seen.push('POST /api/v1/settings/picklists/field-1/remap-delete')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.from_value).toBe('legacy')
        return HttpResponse.json({ values: [] })
      }),
      http.get('/api/v1/settings/picklist-dependencies', () => {
        seen.push('GET /api/v1/settings/picklist-dependencies')
        return HttpResponse.json({ data: [] })
      }),
      http.put('/api/v1/settings/picklist-dependencies', async ({ request }) => {
        seen.push('PUT /api/v1/settings/picklist-dependencies')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.source_field_id).toBe('source-1')
        return HttpResponse.json(body)
      }),
      http.delete('/api/v1/settings/picklist-dependencies/dependency-1', () => {
        seen.push('DELETE /api/v1/settings/picklist-dependencies/dependency-1')
        return new HttpResponse(null, { status: 204 })
      }),
      http.get('/api/v1/settings/lead-conversion-mapping', () => {
        seen.push('GET /api/v1/settings/lead-conversion-mapping')
        return HttpResponse.json({ data: [] })
      }),
      http.put('/api/v1/settings/lead-conversion-mapping', async ({ request }) => {
        seen.push('PUT /api/v1/settings/lead-conversion-mapping')
        const body = (await request.json()) as Record<string, unknown>
        expect(Array.isArray(body.mappings)).toBe(true)
        return HttpResponse.json({ data: body.mappings })
      })
    )

    await numberingApi.get()
    await numberingApi.update({ quote_number_prefix: 'QTE', quote_number_current: 515 })
    await preferencesApi.get()
    await preferencesApi.update({ default_currency: 'EUR' })
    await preferencesApi.getCalendar()
    await currenciesApi.get()
    await currenciesApi.update({
      default_code: 'EUR',
      currencies: [{ code: 'EUR', display_name: 'Euro', symbol: 'EUR', decimal_places: 2, is_active: true }],
    })
    await picklistsApi.listFields()
    await picklistsApi.upsertValues('field-1', [{ value: 'a', display_label: 'A', order_idx: 0, is_active: true }])
    await picklistsApi.remapDelete('field-1', 'legacy')
    await picklistsApi.listDependencies()
    await picklistsApi.upsertDependency({
      entity_type: 'lead',
      source_field_id: 'source-1',
      target_field_id: 'target-1',
      mapping: { qualified: ['hot'] },
      is_active: true,
    })
    await picklistsApi.deleteDependency('dependency-1')
    await leadConversionMappingApi.list()
    await leadConversionMappingApi.replace([
      { lead_field: 'company', target_entity: 'account', target_field: 'name', is_active: true },
    ])

    expect(seen).toEqual([
      'GET /api/v1/settings/numbering',
      'PATCH /api/v1/settings/numbering',
      'GET /api/v1/settings/preferences',
      'PATCH /api/v1/settings/preferences',
      'GET /api/v1/settings/calendar-preferences',
      'GET /api/v1/settings/currencies',
      'PATCH /api/v1/settings/currencies',
      'GET /api/v1/settings/picklists',
      'PUT /api/v1/settings/picklists/field-1/values',
      'POST /api/v1/settings/picklists/field-1/remap-delete',
      'GET /api/v1/settings/picklist-dependencies',
      'PUT /api/v1/settings/picklist-dependencies',
      'DELETE /api/v1/settings/picklist-dependencies/dependency-1',
      'GET /api/v1/settings/lead-conversion-mapping',
      'PUT /api/v1/settings/lead-conversion-mapping',
    ])
  })

  it('uses bundle-4 access settings routes and snake_case payload contracts', async () => {
    const seen: string[] = []

    server.use(
      http.get('/api/v1/settings/roles', () => {
        seen.push('GET /api/v1/settings/roles')
        return HttpResponse.json({ data: [{ id: 'role-1', name: 'Manager', org_id: 'org-1' }] })
      }),
      http.post('/api/v1/settings/roles', async ({ request }) => {
        seen.push('POST /api/v1/settings/roles')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.parent_id).toBe('role-1')
        expect(body.parentId).toBeUndefined()
        return HttpResponse.json({ id: 'role-2', name: body.name, parent_id: body.parent_id }, { status: 201 })
      }),
      http.patch('/api/v1/settings/roles/role-2/parent', async ({ request }) => {
        seen.push('PATCH /api/v1/settings/roles/role-2/parent')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.parent_id).toBeNull()
        return HttpResponse.json({ id: 'role-2', parent_id: null })
      }),
      http.get('/api/v1/settings/profiles', () => {
        seen.push('GET /api/v1/settings/profiles')
        return HttpResponse.json({ data: [{ id: 'profile-1', name: 'Sales', org_id: 'org-1' }] })
      }),
      http.get('/api/v1/settings/profiles/catalog', () => {
        seen.push('GET /api/v1/settings/profiles/catalog')
        return HttpResponse.json({ modules: ['contacts'], actions: ['read', 'update'] })
      }),
      http.put('/api/v1/settings/profiles/profile-1/permissions', async ({ request }) => {
        seen.push('PUT /api/v1/settings/profiles/profile-1/permissions')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.field_permissions).toEqual([
          { module: 'contacts', field_name: 'email', can_write: false },
        ])
        return HttpResponse.json(body)
      }),
      http.get('/api/v1/settings/groups', () => {
        seen.push('GET /api/v1/settings/groups')
        return HttpResponse.json({ data: [] })
      }),
      http.get('/api/v1/settings/groups/member-candidates', () => {
        seen.push('GET /api/v1/settings/groups/member-candidates')
        return HttpResponse.json({
          data: [{ id: 'user-1', name: 'Ada Admin', email: 'ada@example.com', role: 'agent' }],
          meta: { page: 1, per_page: 500, total: 1, total_pages: 1 },
        })
      }),
      http.post('/api/v1/settings/groups', async ({ request }) => {
        seen.push('POST /api/v1/settings/groups')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.user_ids).toEqual(['user-1'])
        return HttpResponse.json({ id: 'group-1', ...body }, { status: 201 })
      }),
      http.put('/api/v1/settings/groups/group-1/members', async ({ request }) => {
        seen.push('PUT /api/v1/settings/groups/group-1/members')
        const body = (await request.json()) as Record<string, unknown>
        expect(body.user_ids).toEqual(['user-1', 'user-2'])
        return new HttpResponse(null, { status: 204 })
      }),
      http.get('/api/v1/settings/sharing-rules', () => {
        seen.push('GET /api/v1/settings/sharing-rules')
        return HttpResponse.json({ rules: [] })
      }),
      http.patch('/api/v1/settings/sharing-rules', async ({ request }) => {
        seen.push('PATCH /api/v1/settings/sharing-rules')
        const body = (await request.json()) as Record<string, unknown>
        expect(body).toEqual({
          rules: [
            {
              module: 'contacts',
              mode: 'private',
              grants: [{ grantee_type: 'role', grantee_id: 'role-1', access_level: 'write' }],
            },
          ],
        })
        return HttpResponse.json(body)
      })
    )

    await accessSettingsApi.listRoles()
    await accessSettingsApi.createRole({ name: 'Regional Manager', parent_id: 'role-1' })
    await accessSettingsApi.moveRole('role-2', null)
    await accessSettingsApi.listProfiles()
    await accessSettingsApi.getPermissionCatalog()
    await accessSettingsApi.replaceProfilePermissions('profile-1', {
      permissions: [{ module: 'contacts', action: 'read', allowed: true }],
      field_permissions: [{ module: 'contacts', field_name: 'email', can_write: false }],
    })
    await accessSettingsApi.listGroups()
    await accessSettingsApi.listGroupMemberCandidates()
    await accessSettingsApi.createGroup({ name: 'West Team', user_ids: ['user-1'] })
    await accessSettingsApi.replaceGroupMembers('group-1', ['user-1', 'user-2'])
    await accessSettingsApi.getSharingRules()
    await accessSettingsApi.replaceSharingRules({
      rules: [
        {
          module: 'contacts',
          mode: 'private',
          grants: [{ grantee_type: 'role', grantee_id: 'role-1', access_level: 'write' }],
        },
      ],
    })

    expect(seen).toEqual([
      'GET /api/v1/settings/roles',
      'POST /api/v1/settings/roles',
      'PATCH /api/v1/settings/roles/role-2/parent',
      'GET /api/v1/settings/profiles',
      'GET /api/v1/settings/profiles/catalog',
      'PUT /api/v1/settings/profiles/profile-1/permissions',
      'GET /api/v1/settings/groups',
      'GET /api/v1/settings/groups/member-candidates',
      'POST /api/v1/settings/groups',
      'PUT /api/v1/settings/groups/group-1/members',
      'GET /api/v1/settings/sharing-rules',
      'PATCH /api/v1/settings/sharing-rules',
    ])
  })
})
