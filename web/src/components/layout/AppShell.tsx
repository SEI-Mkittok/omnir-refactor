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

export function AppShell() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const setUser = useAuthStore((s) => s.setUser)
  const logout = useAuthStore((s) => s.logout)
  const sidebarCollapsed = useUIStore((s) => s.sidebarCollapsed)
  const onboardingOpen = useUIStore((s) => s.onboardingOpen)
  const setOnboardingOpen = useUIStore((s) => s.setOnboardingOpen)

  const [onboardingState, setOnboardingState] = useState<OnboardingState | null>(null)
  const [showResumeBanner, setShowResumeBanner] = useState(false)

  // Restore session from cookie on mount, then check onboarding state
  useEffect(() => {
    fetch('/api/v1/users/me', { credentials: 'include' })
      .then((r) => r.ok ? r.json() : Promise.reject(r.status))
      .then((data) => {
        if (data?.id) {
          setUser(data)
          return getOnboardingState()
        }
        return null
      })
      .then((state) => {
        if (state && !state.completed) {
          setOnboardingState(state)
          setOnboardingOpen(true)
        }
      })
      .catch(() => { logout() })
  }, [setUser, logout])
  usePushNotifications()
  useMutationQueue()

  const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false)
  const location = useLocation()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
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
          breadcrumb={getBreadcrumb(location.pathname)}
        />

        {/* ── Resume onboarding banner (shown after mid-wizard dismiss) ── */}
        {showResumeBanner && !onboardingOpen && (
          <div
            role="status"
            className="flex items-center justify-between gap-3 px-4 py-2.5 text-sm border-b border-[var(--border-default)]"
            style={{ marginTop: 'var(--topbar-height)', background: 'var(--color-primary-light)' }}
          >
            <span className="font-medium" style={{ color: 'var(--color-primary)' }}>
              Finish setting up PraestOS — resume onboarding →
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
          className="flex-1 overflow-auto p-4 pb-safe lg:p-6"
          style={{ marginTop: showResumeBanner && !onboardingOpen ? '0' : 'var(--topbar-height)' }}
        >
          <Outlet />
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
    '/settings/billing':       'Billing',
    '/settings/billing/plans': 'Billing — Plans',
    '/settings/onboarding': 'Getting Started',
    '/admin/audit':        'Audit Log',
  }
  return map[pathname] ?? 'PraestOS'
}
