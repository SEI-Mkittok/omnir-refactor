import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import userEvent from '@testing-library/user-event'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'
import { ModuleRelationshipsPage } from './ModuleRelationshipsPage'

const ZERO_UUID = '00000000-0000-0000-0000-000000000000'

const relationships = [
  {
    id: 'rel-system',
    relationship_key: 'deal_account',
    from_entity_type: 'deal',
    to_entity_type: 'account',
    label: 'Deal account',
    cardinality: 'many_to_one',
    storage_strategy: 'native',
    is_enabled: true,
    system_locked: true,
    order_idx: 10,
    metadata: {},
  },
  {
    id: 'rel-custom',
    relationship_key: 'custom_account_contact',
    from_entity_type: 'account',
    to_entity_type: 'contact',
    label: 'Implementation partner',
    cardinality: 'many_to_many',
    storage_strategy: 'crm_entity_links',
    is_enabled: true,
    system_locked: false,
    order_idx: 20,
    metadata: {},
  },
]

describe('ModuleRelationshipsPage', () => {
  it('shows system locks and deletes custom relationship definitions', async () => {
    let deletedId = ''
    server.use(
      http.get('/api/v1/settings/module-relationships', () => HttpResponse.json(relationships)),
      http.delete('/api/v1/settings/module-relationships/:id', ({ params }) => {
        deletedId = String(params.id)
        return new HttpResponse(null, { status: 204 })
      })
    )

    render(<ModuleRelationshipsPage />)

    expect(await screen.findByDisplayValue('Deal account')).toBeInTheDocument()
    expect(screen.getAllByText('System').length).toBeGreaterThan(0)
    expect(screen.getByDisplayValue('Implementation partner')).toBeInTheDocument()
    expect(screen.getByLabelText('deal_account enabled')).toBeDisabled()
    expect(screen.getByLabelText('custom_account_contact enabled')).not.toBeDisabled()

    await userEvent.click(screen.getByRole('button', { name: /delete/i }))
    await waitFor(() => expect(deletedId).toBe('rel-custom'))
  })

  it('keeps native system relationships enabled while their native flows ignore the flag', async () => {
    const savedRequests: Record<string, unknown>[] = []
    server.use(
      http.get('/api/v1/settings/module-relationships', () => HttpResponse.json([
        {
          id: 'rel-native',
          relationship_key: 'account_contacts',
          from_entity_type: 'account',
          to_entity_type: 'contact',
          label: 'Account contacts',
          cardinality: 'one_to_many',
          storage_strategy: 'native',
          is_enabled: false,
          system_locked: true,
          order_idx: 10,
          metadata: {},
        },
      ])),
      http.put('/api/v1/settings/module-relationships', async ({ request }) => {
        const saved = await request.json() as Record<string, unknown>
        savedRequests.push(saved)
        return HttpResponse.json(saved)
      })
    )

    render(<ModuleRelationshipsPage />)

    const enabledToggle = await screen.findByLabelText('account_contacts enabled')
    expect(enabledToggle).toBeDisabled()
    expect(enabledToggle).toBeChecked()
    expect(screen.getByText(/locked on until native relationship flows enforce/i)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() => expect(savedRequests[0]?.relationship_key).toBe('account_contacts'))
    expect(savedRequests[0]?.is_enabled).toBe(true)
  })

  it('keeps unsaved system defaults keyed by relationship key instead of zero UUID', async () => {
    const savedRequests: Record<string, unknown>[] = []
    server.use(
      http.get('/api/v1/settings/module-relationships', () => HttpResponse.json([
        {
          id: ZERO_UUID,
          relationship_key: 'account_contacts',
          from_entity_type: 'account',
          to_entity_type: 'contact',
          label: 'Account contacts',
          cardinality: 'one_to_many',
          storage_strategy: 'native',
          is_enabled: true,
          system_locked: true,
          order_idx: 10,
          metadata: {},
        },
        {
          id: ZERO_UUID,
          relationship_key: 'account_deals',
          from_entity_type: 'account',
          to_entity_type: 'deal',
          label: 'Account deals',
          cardinality: 'one_to_many',
          storage_strategy: 'native',
          is_enabled: true,
          system_locked: true,
          order_idx: 20,
          metadata: {},
        },
      ])),
      http.put('/api/v1/settings/module-relationships', async ({ request }) => {
        const saved = await request.json() as Record<string, unknown>
        savedRequests.push(saved)
        return HttpResponse.json({ ...saved, id: 'persisted-rel' })
      })
    )

    render(<ModuleRelationshipsPage />)

    const labels = await screen.findAllByLabelText('Relationship label')
    await userEvent.clear(labels[0])
    await userEvent.type(labels[0], 'Primary contacts')

    expect(labels[0]).toHaveValue('Primary contacts')
    expect(labels[1]).toHaveValue('Account deals')

    await userEvent.click(screen.getAllByRole('button', { name: /save/i })[0])
    await waitFor(() => expect(savedRequests[0]?.relationship_key).toBe('account_contacts'))
    expect(savedRequests[0]?.label).toBe('Primary contacts')
  })
})
