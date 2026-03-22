import { useState } from 'react'
import { useNavigate, Navigate, useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { LifeBuoy, AlertCircle } from 'lucide-react'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

const loginSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(1, 'Password is required'),
})

type LoginForm = z.infer<typeof loginSchema>

const AUTH_BASE = import.meta.env.VITE_AUTH_URL || '/api'

export function PortalLoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setUser, isAuthenticated, user } = useAuthStore()
  const [error, setError] = useState<string | null>(null)

  const sessionExpired = searchParams.get('expired') === '1'

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  })

  // Already logged in as client — go straight to tickets
  if (isAuthenticated && user?.role === 'client') {
    return <Navigate to="/portal/tickets" replace />
  }

  async function onSubmit(data: LoginForm) {
    setError(null)
    try {
      const res = await axios.post(`${AUTH_BASE}/auth/login`, data, { withCredentials: true })
      const { user: loggedInUser } = res.data
      if (loggedInUser?.role !== 'client') {
        setError('This portal is for clients only. Please use the main app to sign in.')
        return
      }
      setUser(loggedInUser)
      navigate('/portal/tickets', { replace: true })
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.message ?? 'Invalid email or password')
      } else {
        setError('Something went wrong. Please try again.')
      }
    }
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-[#F7F8FA] px-4 sm:px-0">
      <div className="w-full max-w-sm flex-1 flex flex-col justify-center">
        {/* Logo / branding */}
        <div className="mb-8 flex flex-col items-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#1B3A4B] shadow-lg">
            <LifeBuoy className="h-7 w-7 text-white" />
          </div>
          <h1 className="mt-4 text-2xl font-bold text-slate-900">Client Support Portal</h1>
          <p className="mt-1 text-sm text-slate-500">Sign in to manage your support tickets</p>
        </div>

        {/* Session expired banner */}
        {sessionExpired && (
          <div
            role="alert"
            className="mb-4 flex items-center gap-2 rounded-md bg-amber-50 px-3 py-2 text-sm text-amber-800"
          >
            <AlertCircle className="h-4 w-4 shrink-0" />
            Your session expired. Please sign in again.
          </div>
        )}

        {/* Card */}
        <div className="rounded-xl border border-slate-200 bg-white p-8 shadow-sm">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5" noValidate>
            <div>
              <label
                htmlFor="portal-email"
                className="mb-1.5 block text-sm font-medium text-slate-700"
              >
                Email address
              </label>
              <Input
                id="portal-email"
                type="email"
                placeholder="you@company.com"
                autoComplete="email"
                aria-invalid={!!errors.email}
                aria-describedby={errors.email ? 'portal-email-error' : undefined}
                {...register('email')}
              />
              {errors.email && (
                <p id="portal-email-error" className="mt-1 text-xs text-red-500">
                  {errors.email.message}
                </p>
              )}
            </div>

            <div>
              <label
                htmlFor="portal-password"
                className="mb-1.5 block text-sm font-medium text-slate-700"
              >
                Password
              </label>
              <Input
                id="portal-password"
                type="password"
                placeholder="••••••••"
                autoComplete="current-password"
                aria-invalid={!!errors.password}
                aria-describedby={errors.password ? 'portal-password-error' : undefined}
                {...register('password')}
              />
              {errors.password && (
                <p id="portal-password-error" className="mt-1 text-xs text-red-500">
                  {errors.password.message}
                </p>
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

            <Button type="submit" className="w-full" disabled={isSubmitting} aria-required="true">
              {isSubmitting ? 'Signing in…' : 'Sign in'}
            </Button>
          </form>
        </div>
      </div>

      {/* Footer */}
      <p className="py-6 text-xs text-slate-400">Powered by PraestOS</p>
    </div>
  )
}
