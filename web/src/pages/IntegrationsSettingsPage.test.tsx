import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { IntegrationsSettingsPage } from './IntegrationsSettingsPage'
import { render, screen } from '@/test/utils'
import { server } from '@/test/mocks/server'

describe('IntegrationsSettingsPage', () => {
  it('uses backend OAuth management routes for email and calendar providers', async () => {
    server.use(
      http.get('/api/v1/integrations', () =>
        HttpResponse.json([
          { provider: 'gmail', status: 'disconnected', has_custom_creds: false },
          { provider: 'outlook', status: 'disconnected', has_custom_creds: false },
        ])
      ),
      http.get('/api/v1/calendar/connections', () =>
        HttpResponse.json({ data: [] })
      )
    )

    render(<IntegrationsSettingsPage />)

    expect(await screen.findByText('Gmail')).toBeInTheDocument()
    const links = screen.getAllByRole('link', { name: /connect/i })
    expect(links.map((link) => link.getAttribute('href'))).toEqual([
      '/api/integrations/email/auth/google',
      '/api/integrations/email/auth/microsoft',
      '/api/v1/calendar/auth/google',
      '/api/v1/calendar/auth/microsoft',
    ])
  })

  it('shows passive calendar sync status instead of a manual sync action', async () => {
    server.use(
      http.get('/api/v1/integrations', () =>
        HttpResponse.json([
          { provider: 'gmail', status: 'disconnected', has_custom_creds: false },
          { provider: 'outlook', status: 'disconnected', has_custom_creds: false },
        ])
      ),
      http.get('/api/v1/calendar/connections', () =>
        HttpResponse.json({
          data: [
            {
              id: 'cal-1',
              provider: 'google',
              token_expiry: '2026-06-01T12:00:00Z',
            },
          ],
        })
      )
    )

    render(<IntegrationsSettingsPage />)

    expect(await screen.findByText('Google Calendar')).toBeInTheDocument()
    expect(screen.getByText('Automatic sync every ~5 minutes')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /sync now/i })).not.toBeInTheDocument()
  })
})
