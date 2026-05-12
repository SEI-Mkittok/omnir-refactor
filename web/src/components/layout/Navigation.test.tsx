import { beforeEach, describe, expect, it } from 'vitest'
import { Route, Routes } from 'react-router-dom'
import { http, HttpResponse } from 'msw'
import { AppShell, getBreadcrumb } from './AppShell'
import { Sidebar } from './Sidebar'
import { SettingsHubPage } from '@/pages/settings/SettingsHubPage'
import { render, screen, waitFor } from '@/test/utils'
import { server } from '@/test/mocks/server'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import type { Org, User, UserRole } from '@/api/types'

const org: Org = {
  id: 'org-1',
  name: 'Acme',
  slug: 'acme',
  created_at: '2026-05-09T12:00:00Z',
}

function makeUser(role: UserRole, patch: Partial<User> = {}): User {
  return {
    id: `${role}-1`,
    org_id: org.id,
    email: `${role}@acme.test`,
    name: role,
    role,
    created_at: '2026-05-09T12:00:00Z',
    updated_at: '2026-05-09T12:00:00Z',
    ...patch,
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

function linkHrefs() {
  return screen
    .getAllByRole('link')
    .map((link) => link.getAttribute('href'))
    .filter((href): href is string => Boolean(href?.startsWith('/')))
}

describe('app shell navigation', () => {
  beforeEach(() => {
    localStorage.clear()
    useAuthStore.setState({
      user: null,
      activeOrg: null,
      isAuthenticated: false,
      isInitializing: true,
    })
    useUIStore.setState({
      sidebarCollapsed: false,
      onboardingOpen: false,
      onboardingDismissed: false,
    })
  })

  it('has titles for visible sidebar routes', () => {
    setUser('admin')
    render(<Sidebar />)

    const hrefs = linkHrefs()
    expect(hrefs).toContain('/settings/numbering')
    expect(hrefs).toContain('/settings/company')
    expect(hrefs).toContain('/settings/portal')
    expect(hrefs).toContain('/settings/integrations')
    expect(hrefs).toContain('/kb')
    for (const href of hrefs) {
      expect(getBreadcrumb(href)).not.toBe('Omnir')
    }
  })

  it('shows admin navigation to admin and super admin, but not agent', () => {
    setUser('super_admin')
    const { unmount } = render(<Sidebar />)
    expect(screen.getByRole('link', { name: /users/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /numbering/i })).toBeInTheDocument()

    unmount()
    setUser('agent')
    render(<Sidebar />)
    expect(screen.queryByRole('link', { name: /users/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /numbering/i })).not.toBeInTheDocument()
  })

  it('shows only permitted admin navigation for delegated ACL admins', () => {
    useAuthStore.setState({
      user: makeUser('agent', {
        profile_name: 'People Admin',
        permissions: {
          users: { admin: true },
        },
      }),
      activeOrg: org,
      isAuthenticated: true,
      isInitializing: false,
    })

    const { unmount } = render(<Sidebar />)

    expect(screen.getByRole('link', { name: /users/i })).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /roles/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /billing/i })).not.toBeInTheDocument()

    unmount()
    useAuthStore.setState({
      user: makeUser('agent', {
        profile_name: 'Access Admin',
        permissions: {
          settings: { admin: true },
        },
      }),
      activeOrg: org,
      isAuthenticated: true,
      isInitializing: false,
    })

    render(<Sidebar />)

    expect(screen.queryByRole('link', { name: /users/i })).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: /roles/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /sharing rules/i })).toBeInTheDocument()
  })

  it('applies org menu configuration to non-admin sidebar navigation', async () => {
    setUser('agent')
    server.use(
      http.get('/api/v1/settings/menu', () =>
        HttpResponse.json({ menu_config: { deals: false, contacts: true } })
      )
    )

    render(<Sidebar />)

    expect(await screen.findByRole('link', { name: /contacts/i })).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.queryByRole('link', { name: /pipeline/i })).not.toBeInTheDocument()
    })
  })

  it('keeps settings hub links covered by route titles', () => {
    setUser('super_admin')
    render(<SettingsHubPage />)

    const hrefs = linkHrefs()
    expect(hrefs).toContain('/settings/numbering')
    expect(hrefs).toContain('/settings/company')
    expect(hrefs).toContain('/settings/portal')
    for (const href of hrefs) {
      expect(getBreadcrumb(href)).not.toBe('Omnir')
    }
  })

  it('redirects client users out of the CRM shell', async () => {
    const clientUser = makeUser('client')
    setUser('client')
    server.use(
      http.get('/api/v1/users/me', () => HttpResponse.json(clientUser)),
      http.get('/api/v1/onboarding', () =>
        HttpResponse.json({ id: 'onboarding-1', completed: true, completedSteps: [] })
      )
    )

    render(
      <Routes>
        <Route path="/dashboard" element={<AppShell />} />
        <Route path="/portal/tickets" element={<div>Client portal tickets</div>} />
        <Route path="/login" element={<div>Login</div>} />
      </Routes>,
      { initialRoute: '/dashboard' }
    )

    expect(await screen.findByText('Client portal tickets')).toBeInTheDocument()
  })
})
