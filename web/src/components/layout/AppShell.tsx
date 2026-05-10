import { useState, useEffect } from 'react'
import { Outlet, Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'
import { BottomTabBar } from './BottomTabBar'
import { OfflineBanner } from '@/components/ui/OfflineBanner'
import { OnboardingWizard } from '@/components/omnir/OnboardingWizard'
import { getOnboardingState, type OnboardingState } from '@/api/onboarding'
import { InstallPromptBanner } from '@/components/ui/InstallPromptBanner'
import { usePushNotifications } from '@/hooks/usePushNotifications'
import { useMutationQueue } from '@/hooks/useMutationQueue'
import { usersApi } from '@/api/users'
import { Toaster } from '@/components/ui/Toast'
import { ErrorBoundary } from '@/components/ui/ErrorBoundary'

export function AppShell() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const isInitializing = useAuthStore((s) => s.isInitializing)
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const setActiveOrg = useAuthStore((s) => s.setActiveOrg)
  const activeOrg = useAuthStore((s) => s.activeOrg)
  const setInitializing = useAuthStore((s) => s.setInitializing)
  const logout = useAuthStore((s) => s.logout)
  const sidebarCollapsed = useUIStore((s) => s.sidebarCollapsed)
  const onboardingOpen = useUIStore((s) => s.onboardingOpen)
  const setOnboardingOpen = useUIStore((s) => s.setOnboardingOpen)
  const setOnboardingDismissed = useUIStore((s) => s.setOnboardingDismissed)

  const [onboardingState, setOnboardingState] = useState<OnboardingState | null>(null)
  const [showResumeBanner, setShowResumeBanner] = useState(false)

  // Restore session from cookie on mount, then check onboarding state.
  // Uses apiClient (via usersApi.me) so the 401 → refresh interceptor fires
  // automatically when the access_token cookie has expired, avoiding the
  // "flash dashboard → redirect to login" bug caused by bare fetch().
  //
  // Important: if the user just logged in (isAuthenticated already true), we still
  // run me() to load onboarding state, but a failure does NOT call logout() — the
  // cookie is valid and was just set. Only call logout() if we were NOT already
  // authenticated (i.e., this is a cold page load / session restore attempt).
  useEffect(() => {
    let orgId = ''
    const wasAlreadyAuthenticated = isAuthenticated
    usersApi.me()
      .then((data) => {
        if (data?.id) {
          orgId = data.org_id
          setUser(data)
          return getOnboardingState()
        }
        return null
      })
      .then((state) => {
        if (state) {
          if (state.orgName) {
            setActiveOrg({ id: orgId, name: state.orgName, slug: '', created_at: '' })
          }
          if (!state.completed) {
            setOnboardingState(state)
            setOnboardingDismissed(state.dismissed ?? false)
            if (!state.dismissed) {
              setOnboardingOpen(true)
            }
          }
        }
      })
      .catch(() => {
        // Only force-logout on cold session restore failures.
        // If the user just authenticated via LoginPage, their cookie is valid —
        // a transient me() error (race condition, cold-start latency) should not
        // undo a successful login.
        if (!wasAlreadyAuthenticated) {
          logout()
        }
      })
      .finally(() => { setInitializing(false) })
  }, [setUser, setActiveOrg, logout, setOnboardingOpen, setOnboardingDismissed, setInitializing]) // eslint-disable-line react-hooks/exhaustive-deps

  const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false)
  const location = useLocation()
  const breadcrumb = getBreadcrumb(location.pathname)

  // Keep browser tab title in sync with the active route and org name.
  useEffect(() => {
    document.title = activeOrg?.name
      ? `${breadcrumb} — ${activeOrg.name} — Omnir`
      : `${breadcrumb} — Omnir`
  }, [activeOrg?.name, breadcrumb])
  usePushNotifications()
  useMutationQueue()

  if (isInitializing) {
    return null
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  if (user?.role === 'client') {
    return <Navigate to="/portal/tickets" replace />
  }

  const sidebarWidth = sidebarCollapsed ? 60 : 210

  function handleWizardComplete() {
    setOnboardingOpen(false)
    setShowResumeBanner(false)
    setOnboardingState(null)
  }

  function handleWizardDismiss() {
    if (onboardingState && !onboardingState.completed) {
      setShowResumeBanner(true)
    }
  }

  return (
    <Toaster>
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

        <InstallPromptBanner />
        <OfflineBanner />

        {/* ── Top bar ── */}
        <TopBar
          onMenuClick={() => setMobileDrawerOpen(true)}
          breadcrumb={breadcrumb}
        />

        {/* ── Resume onboarding banner (shown after mid-wizard dismiss) ── */}
        {showResumeBanner && !onboardingOpen && (
          <div
            role="status"
            className="flex items-center justify-between gap-3 px-4 py-2.5 text-sm border-b border-[var(--border-default)]"
            style={{ marginTop: 'var(--topbar-height)', background: 'var(--color-primary-light)' }}
          >
            <span className="font-medium" style={{ color: 'var(--color-primary)' }}>
              Finish setting up {activeOrg?.name ?? 'your workspace'} — resume onboarding →
            </span>
            <div className="flex items-center gap-2 shrink-0">
              <button
                type="button"
                onClick={() => setOnboardingOpen(true)}
                className="rounded-md px-3 py-1 text-xs font-semibold text-white transition-colors"
                style={{ background: 'var(--color-primary)' }}
              >
                Continue Setup
              </button>
              <button
                type="button"
                aria-label="Dismiss onboarding banner"
                onClick={() => setShowResumeBanner(false)}
                className="text-lg leading-none"
                style={{ color: 'var(--text-label)' }}
              >
                ×
              </button>
            </div>
          </div>
        )}

        {/* ── Content area ── */}
        <main
          className="flex-1 overflow-auto p-4 pb-[4.5rem] lg:p-6"
          style={{ marginTop: showResumeBanner && !onboardingOpen ? '0' : 'var(--topbar-height)' }}
        >
          <ErrorBoundary>
            <Outlet />
          </ErrorBoundary>
        </main>

        {/* ── Mobile bottom tab bar ── */}
        <BottomTabBar />
      </div>

      {/* ── Onboarding Wizard (portal, z-index 400) ── */}
      <OnboardingWizard
        open={onboardingOpen}
        onOpenChange={setOnboardingOpen}
        initialState={onboardingState}
        onComplete={handleWizardComplete}
        onDismiss={handleWizardDismiss}
      />
    </div>
    </Toaster>
  )
}

export function getBreadcrumb(pathname: string): string {
  const map: Record<string, string> = {
    '/dashboard':          'Dashboard',
    '/contacts':           'Contacts',
    '/leads':              'Leads',
    '/accounts':           'Accounts',
    '/deals':              'Pipeline',
    '/tickets':            'Tickets',
    '/kb':                 'Knowledge Base',
    '/reports':            'Reports',
    '/dashboards':         'Dashboards',
    '/sequences':          'Sequences',
    '/quotes':             'Quotes',
    '/automations':        'Automations',
    '/calendar':           'Calendar',
    '/inbox':              'Email',
    '/notifications':      'Notifications',
    '/search':             'Search',
    '/users':              'Users',
    '/api-keys':           'API Keys',
    '/settings':           'Settings',
    '/settings/account':   'My Account',
    '/settings/security':  'Security',
    '/settings/security/2fa/enroll': '2FA Enroll',
    '/settings/custom-fields': 'Custom Fields',
    '/settings/numbering': 'Document Numbering',
    '/settings/company': 'Company Profile',
    '/settings/portal': 'Portal Configuration',
    '/settings/outgoing-server': 'Outgoing Server',
    '/settings/config-editor': 'Configuration Editor',
    '/settings/menu': 'Main Menu Configuration',
    '/settings/sla':       'SLA Policies',
    '/settings/webhooks':  'Webhooks',
    '/settings/billing':       'Billing',
    '/settings/billing/plans': 'Billing — Plans',
    '/settings/integrations': 'Integrations',
    '/settings/onboarding': 'Getting Started',
    '/admin/audit':        'Audit Log',
    '/orgs/new':           'New Organization',
  }
  return map[pathname] ?? 'Omnir'
}
