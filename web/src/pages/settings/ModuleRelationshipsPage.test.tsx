import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import userEvent from '@testing-library/user-event'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'
import { ModuleRelationshipsPage } from './ModuleRelationshipsPage'

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

    await userEvent.click(screen.getByRole('button', { name: /delete/i }))
    await waitFor(() => expect(deletedId).toBe('rel-custom'))
  })
})
