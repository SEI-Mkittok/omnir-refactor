import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import userEvent from '@testing-library/user-event'
import { CalendarPage } from './CalendarPage'
import { fireEvent, render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'

function mockActivities() {
  return http.get('/api/v1/activities', () =>
    HttpResponse.json({
      data: [],
      meta: { page: 1, per_page: 500, total: 0, total_pages: 0 },
    })
  )
}

async function openSettingsTab() {
  await userEvent.click(screen.getByRole('tab', { name: /settings/i }))
  expect(await screen.findByText('Connected Calendar Accounts')).toBeInTheDocument()
}

describe('CalendarPage settings', () => {
  it('loads calendar connection state from backend', async () => {
    const seen: string[] = []
    server.use(
      mockActivities(),
      http.get('/api/v1/calendar/connections', () => {
        seen.push('GET /api/v1/calendar/connections')
        return HttpResponse.json({
          data: [{ id: 'cal-google', provider: 'google' }],
        })
      })
    )

    render(<CalendarPage />)
    await openSettingsTab()

    expect(await screen.findByText('Google Calendar')).toBeInTheDocument()
    expect(screen.getByText('Connected')).toBeInTheDocument()
    expect(seen).toContain('GET /api/v1/calendar/connections')
  })

  it('routes connect actions to calendar OAuth endpoints', async () => {
    server.use(
      mockActivities(),
      http.get('/api/v1/calendar/connections', () => HttpResponse.json({ data: [] }))
    )

    render(<CalendarPage />)
    await openSettingsTab()

    const connectLinks = screen.getAllByRole('link', { name: /connect/i })
    const hrefs = connectLinks
      .map((link) => link.getAttribute('href'))
      .filter((href): href is string => Boolean(href))

    expect(hrefs).toEqual(expect.arrayContaining([
      '/api/v1/calendar/auth/google',
      '/api/v1/calendar/auth/microsoft',
    ]))
  })

  it('triggers manual sync with transient syncing state', async () => {
    const seen: string[] = []
    server.use(
      mockActivities(),
      http.get('/api/v1/calendar/connections', () =>
        HttpResponse.json({
          data: [{ id: 'cal-google', provider: 'google' }],
        })
      ),
      http.post('/api/v1/calendar/sync', async () => {
        seen.push('POST /api/v1/calendar/sync')
        await new Promise((resolve) => setTimeout(resolve, 40))
        return HttpResponse.json({ status: 'sync queued' }, { status: 202 })
      })
    )

    render(<CalendarPage />)
    await openSettingsTab()

    const syncButton = await screen.findByRole('button', { name: /^sync$/i })
    fireEvent.click(syncButton)

    expect(await screen.findByRole('button', { name: /syncing/i })).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^sync$/i })).toBeInTheDocument()
    })
    expect(seen).toContain('POST /api/v1/calendar/sync')
  })

  it('disconnects a calendar connection through backend route', async () => {
    const seen: string[] = []
    server.use(
      mockActivities(),
      http.get('/api/v1/calendar/connections', () =>
        HttpResponse.json({
          data: [{ id: 'cal-google', provider: 'google' }],
        })
      ),
      http.delete('/api/v1/calendar/connections/cal-google', () => {
        seen.push('DELETE /api/v1/calendar/connections/cal-google')
        return new HttpResponse(null, { status: 204 })
      })
    )

    render(<CalendarPage />)
    await openSettingsTab()

    const disconnectButton = await screen.findByRole('button', { name: /disconnect/i })
    fireEvent.click(disconnectButton)

    await waitFor(() => {
      expect(seen).toContain('DELETE /api/v1/calendar/connections/cal-google')
    })
  })
})
