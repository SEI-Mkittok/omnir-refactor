import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import userEvent from '@testing-library/user-event'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'
import { CustomRelationshipsPanel } from './CustomRelationshipsPanel'

describe('CustomRelationshipsPanel', () => {
  it('renders readable custom links and creates/removes generic entity links', async () => {
    const createdBodies: unknown[] = []
    let deletedId = ''
    server.use(
      http.get('/api/v1/module-relationships', () =>
        HttpResponse.json([
          {
            id: 'rel-1',
            relationship_key: 'custom_account_contact',
            from_entity_type: 'account',
            to_entity_type: 'contact',
            label: 'Implementation partner',
            cardinality: 'many_to_many',
            storage_strategy: 'crm_entity_links',
            is_enabled: true,
            system_locked: false,
            order_idx: 1,
            metadata: {},
          },
        ])
      ),
      http.get('/api/v1/entity-links/account/account-1', () =>
        HttpResponse.json([
          {
            id: 'link-1',
            org_id: 'org-1',
            relationship_definition_id: 'rel-1',
            from_entity_type: 'account',
            from_entity_id: 'account-1',
            to_entity_type: 'contact',
            to_entity_id: 'contact-1',
            link_type: 'custom',
            metadata: {},
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        ])
      ),
      http.get('/api/v1/contacts', () =>
        HttpResponse.json({
          data: [{
            id: 'contact-2',
            first_name: 'Grace',
            last_name: 'Hopper',
            email: 'grace@example.test',
            stage: 'prospect',
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          }],
          meta: { page: 1, per_page: 10, total: 1, total_pages: 1 },
        })
      ),
      http.post('/api/v1/entity-links', async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>
        createdBodies.push(body)
        return HttpResponse.json({ id: 'link-2', org_id: 'org-1', link_type: 'custom', ...body }, { status: 201 })
      }),
      http.delete('/api/v1/entity-links/:id', ({ params }) => {
        deletedId = String(params.id)
        return new HttpResponse(null, { status: 204 })
      })
    )

    render(<CustomRelationshipsPanel entityType="account" entityId="account-1" />)

    expect(await screen.findByText('Implementation partner')).toBeInTheDocument()
    expect(await screen.findByText('contact-1')).toBeInTheDocument()

    await userEvent.type(screen.getByPlaceholderText(/search contacts/i), 'Gra')
    await userEvent.click(await screen.findByText('Grace Hopper'))

    await waitFor(() => expect(createdBodies).toHaveLength(1))
    expect(createdBodies[0]).toMatchObject({
      relationship_definition_id: 'rel-1',
      from_entity_type: 'account',
      from_entity_id: 'account-1',
      to_entity_type: 'contact',
      to_entity_id: 'contact-2',
    })

    await userEvent.click(screen.getByRole('button', { name: /remove relationship link/i }))
    await waitFor(() => expect(deletedId).toBe('link-1'))
  })
})
