import React, { useEffect, useState } from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AppShell } from '@/components/layout/AppShell'
import { PortalShell } from '@/components/layout/PortalShell'
import { LoginPage } from '@/pages/LoginPage'
import { SetupPage } from '@/pages/SetupPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { ContactsPage } from '@/pages/ContactsPage'
import { LeadsPage } from '@/pages/LeadsPage'
import { AccountsPage } from '@/pages/AccountsPage'
import { DealsPage } from '@/pages/DealsPage'
import { UsersPage } from '@/pages/UsersPage'
import { CustomFieldsPage } from '@/pages/CustomFieldsPage'
import { APIKeysPage } from '@/pages/APIKeysPage'
import { SearchPage } from '@/pages/SearchPage'
import { ReportsPage } from '@/pages/ReportsPage'
import { CustomDashboardsPage } from '@/pages/CustomDashboardsPage'
import { SLASettingsPage } from '@/pages/SLASettingsPage'
import { WebhooksPage } from '@/pages/WebhooksPage'
import { BillingSettingsPage } from '@/pages/BillingSettingsPage'
import { BillingPlansPage } from '@/pages/BillingPlansPage'
import { PortalLoginPage } from '@/pages/portal/PortalLoginPage'
import { PortalTicketsPage } from '@/pages/portal/PortalTicketsPage'
import { PortalSubmitPage } from '@/pages/portal/PortalSubmitPage'
import { PortalTicketDetailPage } from '@/pages/portal/PortalTicketDetailPage'
import { TicketsPage } from '@/pages/TicketsPage'
import { OrgOnboardingPage } from '@/pages/OrgOnboardingPage'
import { NotificationsPage } from '@/pages/NotificationsPage'
import { AuditLogPage } from '@/pages/AuditLogPage'
import { OnboardingSettingsPage } from '@/pages/settings/OnboardingSettingsPage'
import { SettingsHubPage } from '@/pages/settings/SettingsHubPage'
import { AccountSettingsPage } from '@/pages/settings/AccountSettingsPage'
import { SequencesPage } from '@/pages/SequencesPage'
import { QuotesPage } from '@/pages/QuotesPage'
import { AutomationsPage } from '@/pages/AutomationsPage'
import { CalendarPage } from '@/pages/CalendarPage'
import { KnowledgeBasePage } from '@/pages/KnowledgeBasePage'
import { InboxPage } from '@/pages/InboxPage'
import { HelpCenterPage } from '@/pages/help/HelpCenterPage'
import { HelpCategoryPage } from '@/pages/help/HelpCategoryPage'
import { HelpArticlePage } from '@/pages/help/HelpArticlePage'
import { HelpSearchPage } from '@/pages/help/HelpSearchPage'
import { SecuritySettingsPage } from '@/pages/SecuritySettingsPage'
import { RegisterPage } from '@/pages/RegisterPage'
import { IntegrationsSettingsPage } from '@/pages/IntegrationsSettingsPage'
import { TotpEnrollPage } from '@/pages/TotpEnrollPage'
import { ContactDetailPage } from '@/pages/ContactDetailPage'
import { AccountDetailPage } from '@/pages/AccountDetailPage'
import { TicketDetailPage } from '@/pages/TicketDetailPage'
import { getSetupStatus } from '@/api/setup'
import { useAuthStore } from '@/stores/auth'
import '@/styles/globals.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

function AdminRoute({ children }: { children: React.ReactNode }) {
  const user = useAuthStore((s) => s.user)
  if (user?.role !== 'admin') return <Navigate to="/dashboard" replace />
  return <>{children}</>
}

function SuperAdminRoute({ children }: { children: React.ReactNode }) {
  const user = useAuthStore((s) => s.user)
  if (user?.role !== 'super_admin') return <Navigate to="/dashboard" replace />
  return <>{children}</>
}

function AppRoutes() {
  const [setupRequired, setSetupRequired] = useState<boolean | null>(null)

  useEffect(() => {
    getSetupStatus()
      .then((res) => setSetupRequired(res.setupRequired))
      .catch(() => setSetupRequired(false))
  }, [])

  if (setupRequired === null) {
    return null
  }

  if (setupRequired) {
    return (
      <Routes>
        <Route path="/setup" element={<SetupPage />} />
        <Route path="*" element={<Navigate to="/setup" replace />} />
      </Routes>
    )
  }

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/setup" element={<Navigate to="/login" replace />} />

      {/* Public help center routes (no auth) */}
      <Route path="/help/:orgSlug" element={<HelpCenterPage />} />
      <Route path="/help/:orgSlug/c/:categorySlug" element={<HelpCategoryPage />} />
      <Route path="/help/:orgSlug/a/:articleSlug" element={<HelpArticlePage />} />
      <Route path="/help/:orgSlug/search" element={<HelpSearchPage />} />

      {/* Client portal routes */}
      <Route path="/portal/login" element={<PortalLoginPage />} />
      <Route path="/portal" element={<PortalShell />}>
        <Route index element={<Navigate to="/portal/tickets" replace />} />
        <Route path="tickets" element={<PortalTicketsPage />} />
        <Route path="tickets/new" element={<PortalSubmitPage />} />
        <Route path="tickets/:id" element={<PortalTicketDetailPage />} />
      </Route>

      <Route element={<AppShell />}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/contacts" element={<ContactsPage />} />
        <Route path="/contacts/:id" element={<ContactDetailPage />} />
        <Route path="/leads" element={<LeadsPage />} />
        <Route path="/accounts" element={<AccountsPage />} />
        <Route path="/accounts/:id" element={<AccountDetailPage />} />
        <Route path="/deals" element={<DealsPage />} />
        <Route path="/tickets" element={<TicketsPage />} />
        <Route path="/tickets/:id" element={<TicketDetailPage />} />
        <Route path="/kb" element={<KnowledgeBasePage />} />
        <Route path="/search" element={<SearchPage />} />
        <Route path="/reports" element={<ReportsPage />} />
        <Route path="/dashboards" element={<CustomDashboardsPage />} />
        <Route path="/sequences" element={<SequencesPage />} />
        <Route path="/quotes" element={<QuotesPage />} />
        <Route path="/automations" element={<AutomationsPage />} />
        <Route path="/calendar" element={<CalendarPage />} />
        <Route path="/inbox" element={<InboxPage />} />
        <Route path="/notifications" element={<NotificationsPage />} />
        <Route
          path="/users"
          element={
            <AdminRoute>
              <UsersPage />
            </AdminRoute>
          }
        />
        <Route path="/settings" element={<SettingsHubPage />} />
        <Route path="/settings/account" element={<AccountSettingsPage />} />
        <Route path="/settings/security" element={<SecuritySettingsPage />} />
        <Route path="/settings/security/2fa/enroll" element={<TotpEnrollPage />} />
        <Route
          path="/settings/custom-fields"
          element={
            <AdminRoute>
              <CustomFieldsPage />
            </AdminRoute>
          }
        />
        <Route
          path="/api-keys"
          element={
            <AdminRoute>
              <APIKeysPage />
            </AdminRoute>
          }
        />
        <Route
          path="/settings/sla"
          element={
            <AdminRoute>
              <SLASettingsPage />
            </AdminRoute>
          }
        />
        <Route
          path="/settings/webhooks"
          element={
            <AdminRoute>
              <WebhooksPage />
            </AdminRoute>
          }
        />
        <Route
          path="/settings/billing"
          element={
            <AdminRoute>
              <BillingSettingsPage />
            </AdminRoute>
          }
        />
        <Route
          path="/settings/billing/plans"
          element={
            <AdminRoute>
              <BillingPlansPage />
            </AdminRoute>
          }
        />
        <Route
          path="/settings/integrations"
          element={
            <AdminRoute>
              <IntegrationsSettingsPage />
            </AdminRoute>
          }
        />
        <Route path="/settings/onboarding" element={<OnboardingSettingsPage />} />
        <Route
          path="/admin/audit"
          element={
            <AdminRoute>
              <AuditLogPage />
            </AdminRoute>
          }
        />
        <Route
          path="/orgs/new"
          element={
            <SuperAdminRoute>
              <OrgOnboardingPage />
            </SuperAdminRoute>
          }
        />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Route>
    </Routes>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </QueryClientProvider>
  </React.StrictMode>
)
