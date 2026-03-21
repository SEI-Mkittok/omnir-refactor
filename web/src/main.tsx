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
import { SLASettingsPage } from '@/pages/SLASettingsPage'
import { WebhooksPage } from '@/pages/WebhooksPage'
import { PortalLoginPage } from '@/pages/portal/PortalLoginPage'
import { PortalTicketsPage } from '@/pages/portal/PortalTicketsPage'
import { PortalSubmitPage } from '@/pages/portal/PortalSubmitPage'
import { PortalTicketDetailPage } from '@/pages/portal/PortalTicketDetailPage'
import { TicketsPage } from '@/pages/TicketsPage'
import { OrgOnboardingPage } from '@/pages/OrgOnboardingPage'
import { NotificationsPage } from '@/pages/NotificationsPage'
import { AuditLogPage } from '@/pages/AuditLogPage'
import { SequencesPage } from '@/pages/SequencesPage'
import { QuotesPage } from '@/pages/QuotesPage'
import { AutomationsPage } from '@/pages/AutomationsPage'
import { CalendarPage } from '@/pages/CalendarPage'
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
      <Route path="/setup" element={<Navigate to="/login" replace />} />

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
        <Route path="/leads" element={<LeadsPage />} />
        <Route path="/accounts" element={<AccountsPage />} />
        <Route path="/deals" element={<DealsPage />} />
        <Route path="/tickets" element={<TicketsPage />} />
        <Route path="/search" element={<SearchPage />} />
        <Route path="/reports" element={<ReportsPage />} />
        <Route path="/sequences" element={<SequencesPage />} />
        <Route path="/quotes" element={<QuotesPage />} />
        <Route path="/automations" element={<AutomationsPage />} />
        <Route path="/calendar" element={<CalendarPage />} />
        <Route path="/notifications" element={<NotificationsPage />} />
        <Route
          path="/users"
          element={
            <AdminRoute>
              <UsersPage />
            </AdminRoute>
          }
        />
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
