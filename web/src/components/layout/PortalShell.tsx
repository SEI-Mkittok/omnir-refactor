import { useState, useEffect } from 'react'
import { Outlet, Navigate, Link, useNavigate } from 'react-router-dom'
import { LifeBuoy, LogOut, AlertCircle } from 'lucide-react'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

const AUTH_BASE = import.meta.env.VITE_AUTH_URL || '/api'

/** Custom event fired by apiClient when portal session truly expires. */
export const PORTAL_SESSION_EXPIRED = 'portal:session-expired'

export function PortalShell() {
  const { isAuthenticated, user, setUser, logout } = useAuthStore()
  const navigate = useNavigate()

  const [sessionExpired, setSessionExpired] = useState(false)
  const [reloginEmail, setReloginEmail] = useState(user?.email ?? '')
  const [reloginPassword, setReloginPassword] = useState('')
  const [reloginError, setReloginError] = useState<string | null>(null)
  const [reloginLoading, setReloginLoading] = useState(false)

  useEffect(() => {
    function onExpired() {
      setSessionExpired(true)
    }
    window.addEventListener(PORTAL_SESSION_EXPIRED, onExpired)
    return () => window.removeEventListener(PORTAL_SESSION_EXPIRED, onExpired)
  }, [])

  if (!isAuthenticated || user?.role !== 'client') {
    return <Navigate to="/portal/login" replace />
  }

  async function handleLogout() {
    await axios.post(`${AUTH_BASE}/auth/logout`, null, { withCredentials: true }).catch(() => {})
    logout()
    navigate('/portal/login', { replace: true })
  }

  async function handleRelogin(e: React.FormEvent) {
    e.preventDefault()
    if (!reloginPassword) return
    setReloginError(null)
    setReloginLoading(true)
    try {
      const res = await axios.post(
        `${AUTH_BASE}/auth/login`,
        { email: reloginEmail, password: reloginPassword },
        { withCredentials: true },
      )
      const { user: loggedInUser } = res.data
      if (loggedInUser?.role !== 'client') {
        setReloginError('This portal is for clients only.')
        return
      }
      setUser(loggedInUser)
      setSessionExpired(false)
      setReloginPassword('')
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setReloginError(err.response?.data?.message ?? 'Invalid email or password')
      } else {
        setReloginError('Something went wrong. Please try again.')
      }
    } finally {
      setReloginLoading(false)
    }
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

      {/* Session expired overlay — re-login without losing form data */}
      {sessionExpired && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4 backdrop-blur-sm">
          <div className="w-full max-w-sm rounded-xl border border-slate-200 bg-white p-6 shadow-xl">
            <div className="mb-4 flex items-center gap-3">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-amber-100">
                <AlertCircle className="h-5 w-5 text-amber-600" />
              </div>
              <div>
                <h2 className="text-base font-semibold text-slate-900">Session expired</h2>
                <p className="text-sm text-slate-500">Please sign in again to continue.</p>
              </div>
            </div>

            <form onSubmit={handleRelogin} className="space-y-4">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">Email</label>
                <Input
                  type="email"
                  value={reloginEmail}
                  onChange={(e) => setReloginEmail(e.target.value)}
                  autoComplete="email"
                />
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">Password</label>
                <Input
                  type="password"
                  placeholder="••••••••"
                  autoComplete="current-password"
                  value={reloginPassword}
                  onChange={(e) => setReloginPassword(e.target.value)}
                  autoFocus
                />
              </div>

              {reloginError && (
                <p className="text-xs text-red-600">{reloginError}</p>
              )}

              <div className="flex gap-2">
                <Button type="submit" className="flex-1" disabled={!reloginPassword || reloginLoading}>
                  {reloginLoading ? 'Signing in…' : 'Sign in'}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleLogout}
                >
                  Log out
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
