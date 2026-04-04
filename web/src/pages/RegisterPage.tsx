import { useState } from 'react'
import { Link, useNavigate, Navigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { TrendingUp, AlertCircle, Check } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { register as registerUser } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const ORG_MODE = import.meta.env.VITE_ORG_MODE ?? 'saas'

const schema = z
  .object({
    orgName: z.string().min(1, 'Organisation name is required'),
    name: z.string().min(1, 'Your full name is required'),
    email: z.string().email('Invalid email address'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((d) => d.password === d.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })

type FormData = z.infer<typeof schema>

export function RegisterPage() {
  const navigate = useNavigate()
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const [apiError, setApiError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({ resolver: zodResolver(schema) })

  // Block signup on single-tenant instances
  if (ORG_MODE === 'single') {
    return <Navigate to="/login" replace />
  }

  // Already logged in — redirect away
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }

  async function onSubmit(data: FormData) {
    setApiError(null)
    try {
      await registerUser({
        orgName: data.orgName,
        name: data.name,
        email: data.email,
        password: data.password,
      })
      setSuccess(true)
      setTimeout(() => navigate('/login?registered=1'), 2000)
    } catch (err: unknown) {
      const msg =
        err && typeof err === 'object' && 'response' in err
          ? (err as { response?: { data?: { message?: string } } }).response?.data?.message
          : null
      setApiError(msg ?? 'Registration failed. Please try again.')
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#F7F8FA] px-4 sm:px-0">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="mb-8 flex flex-col items-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#1B3A4B] shadow-lg">
            <TrendingUp className="h-7 w-7 text-white" />
          </div>
          <h1 className="mt-4 text-2xl font-bold text-slate-900">Create your workspace</h1>
          <p className="mt-1 text-sm text-slate-500">Set up your organisation and get started</p>
        </div>

        {/* Card */}
        <div className="rounded-xl border border-slate-200 bg-white p-8 shadow-sm">
          {success ? (
            <div className="flex flex-col items-center gap-3 py-4 text-center">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
                <Check className="h-6 w-6 text-green-600" />
              </div>
              <p className="text-sm font-medium text-slate-900">Account created!</p>
              <p className="text-sm text-slate-500">Redirecting you to login…</p>
            </div>
          ) : (
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {apiError && (
                <div
                  role="alert"
                  className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700"
                >
                  <AlertCircle className="h-4 w-4 shrink-0" />
                  {apiError}
                </div>
              )}

              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  Organisation name
                </label>
                <Input
                  placeholder="Acme Inc."
                  autoComplete="organization"
                  {...register('orgName')}
                />
                {errors.orgName && (
                  <p className="mt-1 text-xs text-red-500">{errors.orgName.message}</p>
                )}
              </div>

              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  Your full name
                </label>
                <Input
                  placeholder="Jane Smith"
                  autoComplete="name"
                  {...register('name')}
                />
                {errors.name && (
                  <p className="mt-1 text-xs text-red-500">{errors.name.message}</p>
                )}
              </div>

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
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  Password
                </label>
                <Input
                  type="password"
                  placeholder="Min. 8 characters"
                  autoComplete="new-password"
                  {...register('password')}
                />
                {errors.password && (
                  <p className="mt-1 text-xs text-red-500">{errors.password.message}</p>
                )}
              </div>

              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  Confirm password
                </label>
                <Input
                  type="password"
                  placeholder="Repeat your password"
                  autoComplete="new-password"
                  {...register('confirmPassword')}
                />
                {errors.confirmPassword && (
                  <p className="mt-1 text-xs text-red-500">{errors.confirmPassword.message}</p>
                )}
              </div>

              <Button
                type="submit"
                className="w-full bg-[#1B3A4B] text-white hover:bg-[#16303f]"
                disabled={isSubmitting}
              >
                {isSubmitting ? 'Creating workspace…' : 'Create workspace'}
              </Button>
            </form>
          )}
        </div>

        {/* Login link */}
        <p className="mt-6 text-center text-sm text-slate-500">
          Already have an account?{' '}
          <Link to="/login" className="font-medium text-[#1B3A4B] hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  )
}
