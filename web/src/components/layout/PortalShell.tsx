import { useEffect, useState } from 'react'
import { Outlet, Navigate, Link, useNavigate } from 'react-router-dom'
import { BookOpen, LifeBuoy, LogOut } from 'lucide-react'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { getOnboardingState } from '@/api/onboarding'

const AUTH_BASE = import.meta.env.VITE_AUTH_URL || '/api'

/** Custom event fired by apiClient when portal session truly expires. */
export const PORTAL_SESSION_EXPIRED = 'portal:session-expired'

export function PortalShell() {
  const { isAuthenticated, user, logout } = useAuthStore()
  const navigate = useNavigate()
  const [orgName, setOrgName] = useState<string | null>(null)

  useEffect(() => {
    function onExpired() {
      logout()
      navigate('/portal/login?expired=1', { replace: true })
    }
    window.addEventListener(PORTAL_SESSION_EXPIRED, onExpired)
    return () => window.removeEventListener(PORTAL_SESSION_EXPIRED, onExpired)
  }, [logout, navigate])

  useEffect(() => {
    getOnboardingState()
      .then((state) => { if (state.orgName) setOrgName(state.orgName) })
      .catch(() => {})
  }, [])

  if (!isAuthenticated || user?.role !== 'client') {
    return <Navigate to="/portal/login" replace />
  }

  async function handleLogout() {
    await axios.post(`${AUTH_BASE}/auth/logout`, null, { withCredentials: true }).catch(() => {})
    logout()
    navigate('/portal/login', { replace: true })
  }

  return (
    <div className="flex min-h-screen flex-col bg-[#F7F8FA]">
      {/* Top nav */}
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-3xl items-center justify-between px-4 py-3 sm:px-6">
          <Link to="/portal/tickets" className="flex items-center gap-2">
            <LifeBuoy className="h-5 w-5 text-[#1B3A4B]" />
            <span className="font-semibold text-slate-900">{orgName ? `${orgName} Support` : 'Support Portal'}</span>
          </Link>

          <div className="flex items-center gap-3">
            <Link
              to="/product-help"
              className="hidden items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-900 sm:inline-flex"
            >
              <BookOpen className="h-4 w-4" />
              Help
            </Link>
            <span className="hidden text-sm text-slate-500 sm:block">{user.name || user.email}</span>
            <button
              type="button"
              onClick={handleLogout}
              className="inline-flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm text-slate-600 hover:bg-slate-100 hover:text-slate-900 transition-colors"
            >
              <LogOut className="h-4 w-4" />
              <span className="hidden sm:block">Sign out</span>
            </button>
          </div>
        </div>
      </header>

      {/* Page content */}
      <main className="mx-auto w-full max-w-3xl flex-1 px-4 py-6 sm:px-6">
        <Outlet />
      </main>
    </div>
  )
}
