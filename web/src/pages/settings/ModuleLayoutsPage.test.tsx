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

const ticketLayoutPayload = {
  entity_type: 'ticket',
  blocks: [
    {
      id: 'main',
      label: 'Main',
      order: 0,
      fields: [
        { source: 'standard', field_key: 'subject', label: 'Subject', visible: true, required: true, order: 0, quick_create: true, mass_edit: true, header: true, key_field: true },
        { source: 'standard', field_key: 'source', label: 'Source', visible: true, required: true, order: 1, quick_create: false, mass_edit: true, header: false, key_field: false },
        { source: 'standard', field_key: 'assignee_id', label: 'Assignee', visible: true, required: true, order: 2, quick_create: true, mass_edit: true, header: false, key_field: false },
        { source: 'standard', field_key: 'tags', label: 'Tags', visible: true, required: true, order: 3, quick_create: false, mass_edit: true, header: false, key_field: false },
      ],
    },
  ],
}

type SavedLayoutBody = {
  blocks: Array<{
    fields: Array<{
      field_key: string
      required: boolean
    }>
  }>
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

  it('locks and sanitizes required flags for standard fields missing from quick create forms', async () => {
    let savedBody: SavedLayoutBody | null = null
    server.use(
      http.get('/api/v1/settings/module-layouts/:entityType', ({ params }) => {
        return params.entityType === 'ticket'
          ? HttpResponse.json(ticketLayoutPayload)
          : HttpResponse.json(layoutPayload)
      }),
      http.put('/api/v1/settings/module-layouts/:entityType', async ({ request }) => {
        const body = await request.json() as SavedLayoutBody
        savedBody = body
        return HttpResponse.json({ ...ticketLayoutPayload, ...body })
      })
    )

    render(<ModuleLayoutsPage />)

    await userEvent.click(await screen.findByRole('tab', { name: /tickets/i }))
    expect(await screen.findByDisplayValue('Source')).toBeInTheDocument()

    expect(screen.getByLabelText('source Required')).toBeDisabled()
    expect(screen.getByLabelText('source Required')).not.toBeChecked()
    expect(screen.getByLabelText('assignee_id Required')).toBeDisabled()
    expect(screen.getByLabelText('tags Required')).toBeDisabled()

    await userEvent.click(screen.getByRole('button', { name: /save layout/i }))

    await waitFor(() => expect(savedBody).not.toBeNull())
    const fields = new Map(savedBody!.blocks.flatMap((block) => block.fields).map((field) => [field.field_key, field]))
    expect(fields.get('subject')?.required).toBe(true)
    expect(fields.get('source')?.required).toBe(false)
    expect(fields.get('assignee_id')?.required).toBe(false)
    expect(fields.get('tags')?.required).toBe(false)
  })
})
