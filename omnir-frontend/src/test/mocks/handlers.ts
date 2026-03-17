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

export const handlers = [
  // Auth
  http.post('/api/v1/auth/login', async ({ request }) => {
    const body = (await request.json()) as { email: string; password: string }
    if (body.email === 'admin@omnir.test' && body.password === 'testpassword') {
      return HttpResponse.json({
        accessToken: 'mock.access.token',
        refreshToken: 'mock.refresh.token',
        expiresIn: 900,
      })
    }
    return HttpResponse.json({ error: 'Invalid credentials' }, { status: 401 })
  }),

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
    const body = await request.json()
    return HttpResponse.json({ ...mockContact, ...body, id: 'new-contact-id' }, { status: 201 })
  }),

  http.patch('/api/v1/contacts/:id', async ({ request }) => {
    const body = await request.json()
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
    const body = await request.json()
    return HttpResponse.json({ ...mockAccount, ...body, id: 'new-account-id' }, { status: 201 })
  }),

  // Deals
  http.get('/api/v1/deals', () =>
    HttpResponse.json({ data: [mockDeal], total: 1, page: 1, limit: 50 })
  ),

  http.post('/api/v1/deals', async ({ request }) => {
    const body = await request.json()
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
]
