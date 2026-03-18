import { Outlet, Navigate, Link, useNavigate } from 'react-router-dom'
import { LifeBuoy, LogOut } from 'lucide-react'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const AUTH_BASE = import.meta.env.VITE_AUTH_URL || '/api'

export function PortalShell() {
  const { isAuthenticated, user, logout } = useAuthStore()
  const navigate = useNavigate()

  if (!isAuthenticated || user?.role !== 'client') {
    return <Navigate to="/portal/login" replace />
  }

  async function handleLogout() {
    await axios.post(`${AUTH_BASE}/auth/logout`, null, { withCredentials: true }).catch(() => {})
    logout()
    navigate('/portal/login', { replace: true })
  }

  return (
    <div className="flex min-h-screen flex-col bg-slate-50">
      {/* Top nav */}
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-3xl items-center justify-between px-4 py-3 sm:px-6">
          <Link to="/portal/tickets" className="flex items-center gap-2 text-indigo-600">
            <LifeBuoy className="h-5 w-5" />
            <span className="font-semibold text-slate-900">Support Portal</span>
          </Link>

          <div className="flex items-center gap-3">
            <span className="hidden text-sm text-slate-500 sm:block">{user.email}</span>
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
