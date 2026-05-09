import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import { accountsApi } from './accounts'
import type { CreateActivityRequest } from './activities'
import { dealsApi } from './deals'
import { inboxApi } from './inbox'
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

    server.use(
      http.get('/api/v1/contacts', ({ request }) => {
        const url = new URL(request.url)
        seenPaths.push(url.pathname)
        expect(url.searchParams.get('account_id')).toBe(accountId)
        expect(url.searchParams.get('limit')).toBe('200')
        return HttpResponse.json({
          data: [],
          meta: { page: 1, per_page: 200, total: 0, total_pages: 0 },
        })
      }),
      http.get('/api/v1/deals', ({ request }) => {
        const url = new URL(request.url)
        seenPaths.push(url.pathname)
        expect(url.searchParams.get('account_id')).toBe(accountId)
        expect(url.searchParams.get('limit')).toBe('200')
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
        return HttpResponse.json({
          data: [],
          meta: { page: 1, per_page: 200, total: 0, total_pages: 0 },
        })
      })
    )

    await expect(accountsApi.getContacts(accountId)).resolves.toEqual([])
    await expect(accountsApi.getDeals(accountId)).resolves.toEqual([])
    await expect(accountsApi.getTickets(accountId)).resolves.toEqual([])
    expect(seenPaths).toEqual(['/api/v1/contacts', '/api/v1/deals', '/api/v1/tickets'])
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
})
