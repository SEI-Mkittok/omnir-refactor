import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { IntegrationsSettingsPage } from './IntegrationsSettingsPage'
import { render, screen } from '@/test/utils'
import { server } from '@/test/mocks/server'

describe('IntegrationsSettingsPage', () => {
  it('uses backend email OAuth management routes', async () => {
    server.use(
      http.get('/api/v1/integrations', () =>
        HttpResponse.json([
          { provider: 'gmail', status: 'disconnected', has_custom_creds: false },
          { provider: 'outlook', status: 'disconnected', has_custom_creds: false },
        ])
      )
    )

    render(<IntegrationsSettingsPage />)

    expect(await screen.findByText('Gmail')).toBeInTheDocument()
    const links = screen.getAllByRole('link', { name: /connect/i })
    expect(links.map((link) => link.getAttribute('href'))).toEqual([
      '/api/integrations/email/auth/google',
      '/api/integrations/email/auth/microsoft',
    ])
  })
})
