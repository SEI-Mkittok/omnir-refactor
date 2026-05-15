import { beforeEach, describe, expect, it } from 'vitest'
import { Route, Routes } from 'react-router-dom'
import { http, HttpResponse } from 'msw'
import { PortalShell } from './PortalShell'
import { render, screen } from '@/test/utils'
import { server } from '@/test/mocks/server'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/api/types'

const clientUser: User = {
  id: 'client-1',
  org_id: 'org-1',
  email: 'client@omnir.test',
  name: 'Client User',
  role: 'client',
  created_at: '2026-05-15T00:00:00Z',
  updated_at: '2026-05-15T00:00:00Z',
}

describe('PortalShell', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: null,
      activeOrg: null,
      isAuthenticated: false,
      isInitializing: false,
    })
  })

  it('restores a cookie-backed client session before redirecting to login', async () => {
    server.use(
      http.get('/api/v1/users/me', () => HttpResponse.json(clientUser)),
      http.get('/api/v1/onboarding', () =>
        HttpResponse.json({
          id: 'onboarding-1',
          completed: true,
          completedSteps: [],
          orgName: 'Acme',
        })
      )
    )

    render(
      <Routes>
        <Route path="/portal" element={<PortalShell />}>
          <Route path="tickets" element={<div>Client portal tickets</div>} />
        </Route>
        <Route path="/portal/login" element={<div>Portal login</div>} />
      </Routes>,
      { initialRoute: '/portal/tickets' }
    )

    expect(await screen.findByText('Client portal tickets')).toBeInTheDocument()
    expect(screen.queryByText('Portal login')).not.toBeInTheDocument()
  })
})
