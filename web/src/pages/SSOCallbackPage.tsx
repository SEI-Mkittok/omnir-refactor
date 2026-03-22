import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AlertCircle } from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { ssoApi } from '@/api/sso'
import { Spinner } from '@/components/ui/Spinner'

/**
 * Landing page after a successful OIDC SSO callback.
 * The backend has already set httpOnly auth cookies; this page fetches the
 * current user via /api/v1/users/me and populates the auth store before
 * redirecting to the dashboard.
 */
export function SSOCallbackPage() {
  const navigate = useNavigate()
  const { setUser } = useAuthStore()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    ssoApi
      .getMe()
      .then((user) => {
        setUser(user)
        navigate('/contacts', { replace: true })
      })
      .catch(() => {
        setError('SSO login failed. Please try again.')
      })
  }, [navigate, setUser])

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
        <div className="flex flex-col items-center gap-3 text-center">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-100">
            <AlertCircle className="h-5 w-5 text-red-500" />
          </div>
          <p className="text-sm font-medium text-slate-800">{error}</p>
          <a href="/login" className="text-sm text-[#1B3A4B] underline underline-offset-2">
            Back to login
          </a>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50">
      <div className="flex flex-col items-center gap-4">
        <Spinner />
        <p className="text-sm text-slate-500">Completing sign-in…</p>
      </div>
    </div>
  )
}
