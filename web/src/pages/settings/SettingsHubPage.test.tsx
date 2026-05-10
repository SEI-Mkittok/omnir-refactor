import { beforeEach, describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { render, screen, waitFor, within } from '@/test/utils'
import { server } from '@/test/mocks/server'
import { useAuthStore } from '@/stores/auth'
import { SettingsHubPage } from './SettingsHubPage'
import type { Org, User, UserRole } from '@/api/types'

const org: Org = {
  id: 'org-1',
  name: 'Acme',
  slug: 'acme',
  created_at: '2026-05-09T12:00:00Z',
}

function makeUser(role: UserRole): User {
  return {
    id: `${role}-1`,
    org_id: org.id,
    email: `${role}@acme.test`,
    name: role,
    role,
    created_at: '2026-05-09T12:00:00Z',
    updated_at: '2026-05-09T12:00:00Z',
  }
}

function setUser(role: UserRole) {
  useAuthStore.setState({
    user: makeUser(role),
    activeOrg: org,
    isAuthenticated: true,
    isInitializing: false,
  })
}

describe('SettingsHubPage', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: null,
      activeOrg: null,
      isAuthenticated: false,
      isInitializing: false,
    })
  })

  it('shows admin summary cards with live counts', async () => {
    server.use(
      http.get('/api/v1/users', () =>
        HttpResponse.json({
          data: [],
          meta: { page: 1, per_page: 1, total: 12, total_pages: 12 },
        })
      ),
      http.get('/api/v1/automations', ({ request }) => {
        const status = new URL(request.url).searchParams.get('status')
        if (status === 'active') {
          return HttpResponse.json({ data: [], total: 3 })
        }
        return HttpResponse.json({ data: [], total: 9 })
      }),
      http.get('/api/v1/settings/menu', () =>
        HttpResponse.json({
          menu_config: { inbox: false, reports: false, deals: true },
        })
      ),
      http.get('/api/v1/settings/company', () =>
        HttpResponse.json({ company_name: 'Acme CRM' })
      ),
      http.get('/api/v1/settings/portal', () =>
        HttpResponse.json({
          portal_enabled: true,
          portal_menu: ['tickets'],
          portal_shortcuts: [],
          portal_recent_widget_limit: 5,
        })
      ),
      http.get('/api/v1/settings/outgoing-server', () =>
        HttpResponse.json({
          smtp_host: 'smtp.acme.test',
          smtp_password_set: true,
        })
      ),
      http.get('/api/v1/settings/config-editor', () =>
        HttpResponse.json({
          config_support_email: null,
          config_upload_max_mb: 0,
          config_default_page_size: 0,
          config_list_preview_chars: 0,
        })
      )
    )

    setUser('admin')
    render(<SettingsHubPage />)

    expect(await screen.findByText('Settings Overview')).toBeInTheDocument()

    const usersCard = screen.getByTestId('summary-card-users')
    await waitFor(() => {
      expect(within(usersCard).getByText('12')).toBeInTheDocument()
    })

    const automationsCard = screen.getByTestId('summary-card-automations')
    expect(within(automationsCard).getByText('3/9')).toBeInTheDocument()

    const modulesCard = screen.getByTestId('summary-card-module-visibility')
    expect(within(modulesCard).getByText('12/14')).toBeInTheDocument()

    const coverageCard = screen.getByTestId('summary-card-configuration-coverage')
    expect(within(coverageCard).getByText('4/5')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /currencies/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /picklists/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /picklist dependencies/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /lead conversion mapping/i })).toBeInTheDocument()
  })

  it('keeps summary failures non-blocking for settings shortcuts', async () => {
    server.use(
      http.get('/api/v1/users', () => HttpResponse.json({ error: 'boom' }, { status: 500 })),
      http.get('/api/v1/automations', () => HttpResponse.json({ data: [], total: 0 })),
      http.get('/api/v1/settings/menu', () => HttpResponse.json({ menu_config: {} })),
      http.get('/api/v1/settings/company', () => HttpResponse.json({})),
      http.get('/api/v1/settings/portal', () =>
        HttpResponse.json({
          portal_enabled: false,
          portal_menu: [],
          portal_shortcuts: [],
          portal_recent_widget_limit: 0,
        })
      ),
      http.get('/api/v1/settings/outgoing-server', () => HttpResponse.json({ smtp_password_set: false })),
      http.get('/api/v1/settings/config-editor', () =>
        HttpResponse.json({
          config_support_email: null,
          config_upload_max_mb: 0,
          config_default_page_size: 0,
          config_list_preview_chars: 0,
        })
      )
    )

    setUser('super_admin')
    render(<SettingsHubPage />)

    expect(await screen.findByText('Settings Overview')).toBeInTheDocument()
    expect(await screen.findByText('Unable to load users.')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /integrations/i })).toBeInTheDocument()
  })
})
