import { describe, expect, it } from 'vitest'
import { fireEvent, render, screen, waitFor, within } from '@/test/utils'
import { http, HttpResponse } from 'msw'
import { DocumentNumberingPage } from './DocumentNumberingPage'
import { server } from '@/test/mocks/server'

const settings = {
  id: 'settings-1',
  org_id: 'org-1',
  quote_number_start: 100,
  ticket_number_start: 200,
  kb_article_number_start: 300,
  invoice_number_start: 400,
  created_at: '2026-05-09T12:00:00Z',
  updated_at: '2026-05-09T12:00:00Z',
}

describe('DocumentNumberingPage', () => {
  it('loads numbering settings and saves a snake_case payload', async () => {
    let patchBody: Record<string, unknown> | undefined

    server.use(
      http.get('/api/v1/settings/numbering', () => HttpResponse.json(settings)),
      http.patch('/api/v1/settings/numbering', async ({ request }) => {
        patchBody = (await request.json()) as Record<string, unknown>
        return HttpResponse.json({ ...settings, ...patchBody })
      })
    )

    render(<DocumentNumberingPage />)

    const quotes = await screen.findByRole('region', { name: 'Quotes' })
    const input = within(quotes).getByLabelText('Starting number') as HTMLInputElement
    await waitFor(() => expect(input).toHaveValue(100))
    fireEvent.change(input, { target: { value: '1234' } })
    expect(input).toHaveValue(1234)
    fireEvent.click(screen.getByRole('button', { name: /save numbering/i }))

    await waitFor(() => {
      expect(patchBody).toEqual({
        quote_number_start: 1234,
        ticket_number_start: 200,
        kb_article_number_start: 300,
        invoice_number_start: 400,
      })
    })
    expect(await screen.findByText('Document numbering saved.')).toBeInTheDocument()
  })

  it('requires positive whole numbers before saving', async () => {
    let patchCalled = false

    server.use(
      http.get('/api/v1/settings/numbering', () => HttpResponse.json(settings)),
      http.patch('/api/v1/settings/numbering', () => {
        patchCalled = true
        return HttpResponse.json(settings)
      })
    )

    render(<DocumentNumberingPage />)

    const tickets = await screen.findByRole('region', { name: 'Tickets' })
    const input = within(tickets).getByLabelText('Starting number') as HTMLInputElement
    await waitFor(() => expect(input).toHaveValue(200))
    fireEvent.change(input, { target: { value: '0' } })
    expect(input).toHaveValue(0)
    fireEvent.click(screen.getByRole('button', { name: /save numbering/i }))

    expect(await within(tickets).findByText('Enter a positive whole number.')).toBeInTheDocument()
    expect(patchCalled).toBe(false)
  })

  it('shows a load error state', async () => {
    server.use(
      http.get('/api/v1/settings/numbering', () =>
        HttpResponse.json({ error: 'boom' }, { status: 500 })
      )
    )

    render(<DocumentNumberingPage />)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Failed to load document numbering settings.'
    )
  })
})
