import { useState, useEffect } from 'react'
import { Outlet, Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'
import { BottomTabBar } from './BottomTabBar'
import { OfflineBanner } from '@/components/ui/OfflineBanner'

export function AppShell() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const setUser = useAuthStore((s) => s.setUser)
  const logout = useAuthStore((s) => s.logout)
  const sidebarCollapsed = useUIStore((s) => s.sidebarCollapsed)

  // Restore session from cookie on mount
  useEffect(() => {
    fetch('/api/v1/users/me', { credentials: 'include' })
      .then((r) => r.ok ? r.json() : Promise.reject(r.status))
      .then((data) => { if (data?.id) setUser(data) })
      .catch(() => { logout() })
  }, [setUser, logout])
  const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false)
  const location = useLocation()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  const sidebarWidth = sidebarCollapsed ? 60 : 210

  return (
    <div
      id="layout-root"
      className="min-h-screen"
      style={{ background: 'var(--surface-app)' }}
      data-sidebar-collapsed={sidebarCollapsed ? 'true' : 'false'}
    >
      {/* ── Sidebar (desktop fixed, mobile drawer) ── */}
      <Sidebar
        open={mobileDrawerOpen}
        onClose={() => setMobileDrawerOpen(false)}
      />

      {/* ── Main column: topbar + content ── */}
      <div
        className="flex flex-col min-h-screen transition-all duration-200"
        style={{
          marginLeft: `${sidebarWidth}px`,
        }}
      >
        {/* Override margin on mobile — sidebar is a drawer, not inline */}
        <style>{`@media (max-width: 767px) { #layout-root > div { margin-left: 0 !important; } }`}</style>

        <OfflineBanner />

        {/* ── Top bar ── */}
        <TopBar
          onMenuClick={() => setMobileDrawerOpen(true)}
          breadcrumb={getBreadcrumb(location.pathname)}
        />

        {/* ── Content area ── */}
        <main
          className="flex-1 overflow-auto p-4 pb-safe lg:p-6"
          style={{ marginTop: 'var(--topbar-height)' }}
        >
          <Outlet />
        </main>

        {/* ── Mobile bottom tab bar ── */}
        <BottomTabBar />
      </div>
    </div>
  )
}

function getBreadcrumb(pathname: string): string {
  const map: Record<string, string> = {
    '/dashboard':          'Dashboard',
    '/contacts':           'Contacts',
    '/leads':              'Leads',
    '/accounts':           'Accounts',
    '/deals':              'Pipeline',
    '/tickets':            'Tickets',
    '/reports':            'Reports',
    '/sequences':          'Sequences',
    '/quotes':             'Quotes',
    '/automations':        'Automations',
    '/calendar':           'Calendar',
    '/notifications':      'Notifications',
    '/search':             'Search',
    '/users':              'Users',
    '/api-keys':           'API Keys',
    '/settings/custom-fields': 'Custom Fields',
    '/settings/sla':       'SLA Policies',
    '/settings/webhooks':  'Webhooks',
    '/admin/audit':        'Audit Log',
  }
  return map[pathname] ?? 'PraestOS'
}
