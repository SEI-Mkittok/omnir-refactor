import { describe, expect, it } from 'vitest'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { SLASettingsPage } from './SLASettingsPage'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'

describe('SLASettingsPage', () => {
  it('allows fractional hour targets when creating an SLA policy', async () => {
    let payload: Record<string, unknown> | undefined
    const user = userEvent.setup()

    server.use(
      http.get('/api/v1/sla-policies', () =>
        HttpResponse.json({
          data: [],
          meta: { page: 1, per_page: 20, total: 0, total_pages: 1 },
        })
      ),
      http.post('/api/v1/sla-policies', async ({ request }) => {
        payload = (await request.json()) as Record<string, unknown>
        return HttpResponse.json(
          {
            id: 'policy-1',
            created_at: '2026-05-14T00:00:00Z',
            updated_at: '2026-05-14T00:00:00Z',
            ...payload,
          },
          { status: 201 }
        )
      })
    )

    render(<SLASettingsPage />)

    await user.click(await screen.findByRole('button', { name: /add policy/i }))
    await user.type(screen.getByPlaceholderText(/standard sla/i), 'Fractional SLA')

    const [responseInput, resolutionInput] = screen.getAllByRole('spinbutton')
    expect(responseInput).toHaveAttribute('step', 'any')
    expect(resolutionInput).toHaveAttribute('step', 'any')

    await user.clear(responseInput)
    await user.type(responseInput, '0.25')
    await user.clear(resolutionInput)
    await user.type(resolutionInput, '1.75')
    await user.click(screen.getByRole('button', { name: /create policy/i }))

    await waitFor(() =>
      expect(payload).toMatchObject({
        name: 'Fractional SLA',
        response_time_hours: 0.25,
        resolution_time_hours: 1.75,
      })
    )
  })
})
