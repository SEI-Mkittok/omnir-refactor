import { http, HttpResponse } from 'msw'

// Shared MSW handlers for unit/component tests.
// Override per-test with server.use(...) when you need a specific response.

const mockContact = {
  id: '00000000-0000-0000-0000-000000000001',
  firstName: 'Ada',
  lastName: 'Lovelace',
  email: 'ada@omnir.test',
  stage: 'prospect',
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

const mockAccount = {
  id: '00000000-0000-0000-0000-000000000002',
  name: 'Acme Corp',
  domain: 'acme.example.com',
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

const mockDeal = {
  id: '00000000-0000-0000-0000-000000000003',
  title: 'New Enterprise Deal',
  valueCents: 500000,
  currency: 'USD',
  stage: 'qualified',
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

function mockModuleLayout(entityType: string) {
  const standard: Record<string, string[]> = {
    lead: ['first_name', 'last_name', 'email', 'phone', 'company', 'lead_source', 'status'],
    contact: ['first_name', 'last_name', 'email', 'phone', 'account_id', 'stage'],
    account: ['name', 'domain', 'industry', 'size'],
    deal: ['title', 'value_cents', 'stage', 'expected_close_date', 'pipeline_id'],
    ticket: ['subject', 'status', 'priority'],
  }
  return {
    entity_type: entityType,
    blocks: [
      {
        id: 'main',
        label: 'Main',
        order: 0,
        fields: (standard[entityType] ?? []).map((field, index) => ({
          source: 'standard',
          field_key: field,
          label: field.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()),
          visible: true,
          required: ['name', 'first_name', 'last_name', 'title', 'pipeline_id', 'subject'].includes(field),
          order: index,
          quick_create: true,
          mass_edit: true,
          header: index < 2,
          key_field: index === 0,
        })),
      },
    ],
  }
}

export const handlers = [
  // Auth
  http.post('/api/auth/login', async ({ request }) => {
    const body = (await request.json()) as { email: string; password: string }
    if (body.email === 'admin@omnir.test' && body.password === 'testpassword') {
      return HttpResponse.json({
        user: { id: 'user-1', email: 'admin@omnir.test', name: 'Admin', role: 'admin' },
      })
    }
    return HttpResponse.json({ error: 'Invalid credentials' }, { status: 401 })
  }),

  http.post('/api/auth/refresh', () => new HttpResponse(null, { status: 204 })),

  http.post('/api/auth/logout', () => new HttpResponse(null, { status: 204 })),

  http.get('/api/v1/auth/me', () =>
    HttpResponse.json({ id: 'user-1', email: 'admin@omnir.test', name: 'Admin', role: 'admin' })
  ),

  // Contacts
  http.get('/api/v1/contacts', () =>
    HttpResponse.json({ data: [mockContact], total: 1, page: 1, limit: 50 })
  ),

  http.get('/api/v1/contacts/:id', ({ params }) => {
    if (params.id === mockContact.id) return HttpResponse.json(mockContact)
    return HttpResponse.json({ error: 'Not found' }, { status: 404 })
  }),

  http.post('/api/v1/contacts', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({ ...mockContact, ...body, id: 'new-contact-id' }, { status: 201 })
  }),

  http.patch('/api/v1/contacts/:id', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({ ...mockContact, ...body })
  }),

  http.delete('/api/v1/contacts/:id', () => new HttpResponse(null, { status: 204 })),

  // Accounts
  http.get('/api/v1/accounts', () =>
    HttpResponse.json({ data: [mockAccount], total: 1, page: 1, limit: 50 })
  ),

  http.get('/api/v1/accounts/:id', ({ params }) => {
    if (params.id === mockAccount.id) return HttpResponse.json(mockAccount)
    return HttpResponse.json({ error: 'Not found' }, { status: 404 })
  }),

  http.post('/api/v1/accounts', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({ ...mockAccount, ...body, id: 'new-account-id' }, { status: 201 })
  }),

  // Deals
  http.get('/api/v1/deals', () =>
    HttpResponse.json({ data: [mockDeal], total: 1, page: 1, limit: 50 })
  ),

  http.post('/api/v1/deals', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({ ...mockDeal, ...body, id: 'new-deal-id' }, { status: 201 })
  }),

  // Search
  http.get('/api/v1/search', ({ request }) => {
    const url = new URL(request.url)
    const q = url.searchParams.get('q') ?? ''
    return HttpResponse.json({
      contacts: q ? [mockContact] : [],
      accounts: q ? [mockAccount] : [],
      deals: q ? [mockDeal] : [],
    })
  }),

  // Settings
  http.get('/api/v1/settings/menu', () =>
    HttpResponse.json({ menu_config: {} })
  ),

  http.get('/api/v1/settings/company', () =>
    HttpResponse.json({})
  ),

  http.get('/api/v1/settings/portal', () =>
    HttpResponse.json({
      portal_enabled: false,
      portal_menu: [],
      portal_shortcuts: [],
      portal_recent_widget_limit: 0,
    })
  ),

  http.get('/api/v1/settings/outgoing-server', () =>
    HttpResponse.json({ smtp_password_set: false })
  ),

  http.get('/api/v1/settings/config-editor', () =>
    HttpResponse.json({
      config_support_email: null,
      config_upload_max_mb: 0,
      config_default_page_size: 0,
      config_list_preview_chars: 0,
    })
  ),

  http.get('/api/v1/custom-fields', () =>
    HttpResponse.json([])
  ),

  http.get('/api/v1/module-layouts/:entityType', ({ params }) =>
    HttpResponse.json(mockModuleLayout(String(params.entityType)))
  ),

  http.get('/api/v1/settings/module-layouts/:entityType', ({ params }) =>
    HttpResponse.json(mockModuleLayout(String(params.entityType)))
  ),

  http.put('/api/v1/settings/module-layouts/:entityType', async ({ params, request }) => {
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({ ...mockModuleLayout(String(params.entityType)), ...body })
  }),

  http.post('/api/v1/settings/module-layouts/:entityType/reset', ({ params }) =>
    HttpResponse.json(mockModuleLayout(String(params.entityType)))
  ),

  http.get('/api/v1/settings/module-relationships', () =>
    HttpResponse.json([])
  ),

  http.get('/api/v1/module-relationships', () =>
    HttpResponse.json([])
  ),

  http.get('/api/v1/entity-links/:entityType/:entityId', () =>
    HttpResponse.json([])
  ),

  http.post('/api/v1/entity-links', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({
      id: 'entity-link-id',
      org_id: 'org-id',
      link_type: 'custom',
      metadata: {},
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
      ...body,
    }, { status: 201 })
  }),

  http.delete('/api/v1/entity-links/:id', () =>
    new HttpResponse(null, { status: 204 })
  ),

  http.get('/api/v1/users', () =>
    HttpResponse.json({
      data: [],
      meta: { page: 1, per_page: 1, total: 0, total_pages: 0 },
    })
  ),

  http.get('/api/v1/automations', () =>
    HttpResponse.json({ data: [], total: 0 })
  ),

  http.get('/api/v1/activities', () =>
    HttpResponse.json({
      data: [],
      meta: { page: 1, per_page: 500, total: 0, total_pages: 0 },
    })
  ),

  http.get('/api/v1/calendar/connections', () =>
    HttpResponse.json({ data: [] })
  ),

  http.post('/api/v1/calendar/sync', () =>
    HttpResponse.json({ status: 'sync queued' }, { status: 202 })
  ),
]
