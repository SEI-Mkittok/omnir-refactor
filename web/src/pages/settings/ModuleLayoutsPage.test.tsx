import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import userEvent from '@testing-library/user-event'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'
import { ModuleLayoutsPage } from './ModuleLayoutsPage'

const layoutPayload = {
  entity_type: 'lead',
  blocks: [
    {
      id: 'main',
      label: 'Main',
      order: 0,
      fields: [
        { source: 'standard', field_key: 'first_name', label: 'First name', visible: true, required: true, order: 0, quick_create: true, mass_edit: true, header: true, key_field: true },
        { source: 'standard', field_key: 'last_name', label: 'Last name', visible: true, required: true, order: 1, quick_create: true, mass_edit: true, header: true, key_field: true },
        { source: 'standard', field_key: 'email', label: 'Email', visible: true, required: true, order: 2, quick_create: true, mass_edit: true, header: true, key_field: false },
      ],
    },
  ],
}

describe('ModuleLayoutsPage', () => {
  it('saves edited module layout blocks', async () => {
    let savedBody: unknown = null
    server.use(
      http.get('/api/v1/settings/module-layouts/:entityType', () => HttpResponse.json(layoutPayload)),
      http.put('/api/v1/settings/module-layouts/:entityType', async ({ request }) => {
        savedBody = await request.json()
        return HttpResponse.json({ ...layoutPayload, ...(savedBody as object) })
      })
    )

    render(<ModuleLayoutsPage />)

    expect(await screen.findByDisplayValue('Email')).toBeInTheDocument()
    await userEvent.clear(screen.getByDisplayValue('Email'))
    await userEvent.type(screen.getByLabelText('email label'), 'Primary email')
    await userEvent.click(screen.getByRole('button', { name: /save layout/i }))

    await waitFor(() => expect(savedBody).not.toBeNull())
    expect(JSON.stringify(savedBody)).toContain('Primary email')
    expect(await screen.findByText(/layout saved/i)).toBeInTheDocument()
  })
})
