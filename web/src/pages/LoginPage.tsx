import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { TrendingUp, AlertCircle } from 'lucide-react'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'

const loginSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(1, 'Password is required'),
})

type LoginForm = z.infer<typeof loginSchema>

const AUTH_BASE = import.meta.env.VITE_AUTH_URL || '/api'

type SsoProvider = 'microsoft' | 'google'
type SsoRedirectingState = SsoProvider | null

const SSO_ERROR_MESSAGES: Record<string, string> = {
  account_not_in_org: 'Your account is not part of this organisation. Contact your admin.',
  sso_disabled: 'Single sign-on is not enabled for this organisation.',
  generic_oidc_failure: 'Sign-in failed. Please try again or use your email and password.',
}

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setUser } = useAuthStore()
  const [error, setError] = useState<string | null>(null)
  const [ssoRedirecting, setSsoRedirecting] = useState<SsoRedirectingState>(null)

  const ssoError = searchParams.get('error')
  const ssoErrorMessage = ssoError ? (SSO_ERROR_MESSAGES[ssoError] ?? SSO_ERROR_MESSAGES.generic_oidc_failure) : null

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  })

  async function onSubmit(data: LoginForm) {
    setError(null)
    try {
      const res = await axios.post(`${AUTH_BASE}/auth/login`, data, { withCredentials: true })
      const body = res.data

      if (body.requires_2fa) {
        navigate('/login/2fa', { state: { pending2fa: true } })
        return
      }

      if (body.user) {
        setUser(body.user)
        navigate('/dashboard')
      }
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.message ?? 'Invalid email or password')
      } else {
        setError('Something went wrong. Please try again.')
      }
    }
  }

  function handleSsoClick(provider: SsoProvider) {
    if (ssoRedirecting) return
    setSsoRedirecting(provider)
    window.location.href = `${AUTH_BASE}/auth/sso/${provider}`
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#F7F8FA] px-4 sm:px-0">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="mb-8 flex flex-col items-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#1B3A4B] shadow-lg">
            <TrendingUp className="h-7 w-7 text-white" />
          </div>
          <h1 className="mt-4 text-2xl font-bold text-slate-900">Welcome to PraestOS</h1>
          <p className="mt-1 text-sm text-slate-500">Sign in to your account</p>
        </div>

        {/* Card */}
        <div className="rounded-xl border border-slate-200 bg-white p-8 shadow-sm">
          {/* SSO Error banner */}
          {ssoErrorMessage && (
            <div
              role="alert"
              className="mb-5 flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700"
            >
              <AlertCircle className="h-4 w-4 shrink-0" />
              {ssoErrorMessage}
            </div>
          )}

          {/* SSO Buttons */}
          <div className="space-y-2">
            <Button
              type="button"
              variant="outline"
              className="w-full"
              aria-label="Sign in with Microsoft"
              disabled={ssoRedirecting !== null}
              onClick={() => handleSsoClick('microsoft')}
            >
              {ssoRedirecting === 'microsoft' ? (
                <>
                  <Spinner size="sm" />
                  Redirecting…
                </>
              ) : (
                <>
                  {/* Microsoft M logo */}
                  <svg
                    width="18"
                    height="18"
                    viewBox="0 0 21 21"
                    aria-hidden="true"
                    className="shrink-0"
                  >
                    <rect x="1" y="1" width="9" height="9" fill="#f25022" />
                    <rect x="11" y="1" width="9" height="9" fill="#7fba00" />
                    <rect x="1" y="11" width="9" height="9" fill="#00a4ef" />
                    <rect x="11" y="11" width="9" height="9" fill="#ffb900" />
                  </svg>
                  Sign in with Microsoft
                </>
              )}
            </Button>

            <Button
              type="button"
              variant="outline"
              className="w-full"
              aria-label="Sign in with Google"
              disabled={ssoRedirecting !== null}
              onClick={() => handleSsoClick('google')}
            >
              {ssoRedirecting === 'google' ? (
                <>
                  <Spinner size="sm" />
                  Redirecting…
                </>
              ) : (
                <>
                  {/* Google G logo */}
                  <svg
                    width="18"
                    height="18"
                    viewBox="0 0 24 24"
                    aria-hidden="true"
                    className="shrink-0"
                  >
                    <path
                      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                      fill="#4285F4"
                    />
                    <path
                      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                      fill="#34A853"
                    />
                    <path
                      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                      fill="#FBBC05"
                    />
                    <path
                      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                      fill="#EA4335"
                    />
                  </svg>
                  Sign in with Google
                </>
              )}
            </Button>
          </div>

          {/* Divider */}
          <div className="relative my-5 flex items-center">
            <div className="flex-grow border-t border-slate-200" />
            <span className="mx-3 shrink-0 text-xs text-slate-400">or</span>
            <div className="flex-grow border-t border-slate-200" />
          </div>

          {/* Email / password form */}
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                Email address
              </label>
              <Input
                type="email"
                placeholder="you@company.com"
                autoComplete="email"
                {...register('email')}
              />
              {errors.email && (
                <p className="mt-1 text-xs text-red-500">{errors.email.message}</p>
              )}
            </div>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">Password</label>
              <Input
                type="password"
                placeholder="••••••••"
                autoComplete="current-password"
                {...register('password')}
              />
              {errors.password && (
                <p className="mt-1 text-xs text-red-500">{errors.password.message}</p>
              )}
            </div>

            {error && (
              <div
                role="alert"
                className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700"
              >
                <AlertCircle className="h-4 w-4 shrink-0" />
                {error}
              </div>
            )}

            <Button type="submit" className="w-full" disabled={isSubmitting}>
              {isSubmitting ? 'Signing in…' : 'Sign in'}
            </Button>
          </form>
        </div>
      </div>
    </div>
  )
}
