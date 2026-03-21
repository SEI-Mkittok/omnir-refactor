import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { TrendingUp, AlertCircle, Shield } from 'lucide-react'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ssoApi, totpApi, type SSOConfig } from '@/api/sso'

const loginSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(1, 'Password is required'),
})

const totpSchema = z.object({
  code: z.string().length(6, 'Enter your 6-digit code').regex(/^\d+$/, 'Digits only'),
})

type LoginForm = z.infer<typeof loginSchema>
type TOTPForm = z.infer<typeof totpSchema>

const AUTH_BASE = import.meta.env.VITE_AUTH_URL || '/api'
const ORG_SLUG = import.meta.env.VITE_ORG_SLUG as string | undefined

export function LoginPage() {
  const navigate = useNavigate()
  const { setUser } = useAuthStore()
  const [error, setError] = useState<string | null>(null)
  const [ssoConfig, setSSOConfig] = useState<SSOConfig | null>(null)

  // 2FA state
  const [totpRequired, setTotpRequired] = useState(false)
  const [preAuthToken, setPreAuthToken] = useState<string | null>(null)
  const [useBackupCode, setUseBackupCode] = useState(false)
  const [backupCode, setBackupCode] = useState('')

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({ resolver: zodResolver(loginSchema) })

  const {
    register: registerTOTP,
    handleSubmit: handleTOTPSubmit,
    formState: { errors: totpErrors, isSubmitting: totpSubmitting },
  } = useForm<TOTPForm>({ resolver: zodResolver(totpSchema) })

  useEffect(() => {
    if (!ORG_SLUG) return
    ssoApi.getConfig(ORG_SLUG).then(setSSOConfig).catch(() => {})
  }, [])

  async function onSubmit(data: LoginForm) {
    setError(null)
    try {
      const res = await axios.post(`${AUTH_BASE}/auth/login`, data, { withCredentials: true })
      const body = res.data

      if (body.requires_2fa) {
        setPreAuthToken(body.pre_auth_token ?? null)
        setTotpRequired(true)
        return
      }

      if (body.user) {
        setUser(body.user)
        navigate('/contacts')
      }
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.message ?? 'Invalid email or password')
      } else {
        setError('Something went wrong. Please try again.')
      }
    }
  }

  async function onTOTPSubmit(data: TOTPForm) {
    if (!preAuthToken) return
    setError(null)
    try {
      const user = await totpApi.verifyLogin({ code: data.code }, preAuthToken)
      setUser(user)
      navigate('/contacts')
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.message ?? 'Invalid code')
      } else {
        setError('Something went wrong. Please try again.')
      }
    }
  }

  async function onBackupCodeSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!preAuthToken || !backupCode.trim()) return
    setError(null)
    try {
      const user = await totpApi.verifyLogin({ backup_code: backupCode.trim() }, preAuthToken)
      setUser(user)
      navigate('/contacts')
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.message ?? 'Invalid backup code')
      } else {
        setError('Something went wrong. Please try again.')
      }
    }
  }

  const showGoogle =
    ssoConfig?.enabled &&
    (ssoConfig.provider === 'google' || ssoConfig.provider === undefined)
  const showMicrosoft =
    ssoConfig?.enabled &&
    (ssoConfig.provider === 'microsoft' || ssoConfig.provider === 'oidc')

  // ── TOTP Step ──────────────────────────────────────────────────────────────
  if (totpRequired) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50 px-4 sm:px-0">
        <div className="w-full max-w-md">
          <div className="mb-8 flex flex-col items-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#1B3A4B] shadow-lg">
              <Shield className="h-7 w-7 text-white" />
            </div>
            <h1 className="mt-4 text-2xl font-bold text-slate-900">Two-factor authentication</h1>
            <p className="mt-1 text-sm text-slate-500">
              {useBackupCode
                ? 'Enter one of your backup codes'
                : 'Enter the 6-digit code from your authenticator app'}
            </p>
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-8 shadow-sm">
            {!useBackupCode ? (
              <form onSubmit={handleTOTPSubmit(onTOTPSubmit)} className="space-y-5">
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">
                    Authenticator code
                  </label>
                  <Input
                    type="text"
                    inputMode="numeric"
                    placeholder="000000"
                    maxLength={6}
                    autoComplete="one-time-code"
                    className="text-center text-xl tracking-widest"
                    {...registerTOTP('code')}
                  />
                  {totpErrors.code && (
                    <p className="mt-1 text-xs text-red-500">{totpErrors.code.message}</p>
                  )}
                </div>

                {error && (
                  <div className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
                    <AlertCircle className="h-4 w-4 shrink-0" />
                    {error}
                  </div>
                )}

                <Button type="submit" className="w-full" disabled={totpSubmitting}>
                  {totpSubmitting ? 'Verifying…' : 'Verify'}
                </Button>
              </form>
            ) : (
              <form onSubmit={onBackupCodeSubmit} className="space-y-5">
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">
                    Backup code
                  </label>
                  <Input
                    type="text"
                    placeholder="xxxxxxxxxxxxxxxx"
                    value={backupCode}
                    onChange={(e) => setBackupCode(e.target.value)}
                    autoComplete="off"
                  />
                </div>

                {error && (
                  <div className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
                    <AlertCircle className="h-4 w-4 shrink-0" />
                    {error}
                  </div>
                )}

                <Button type="submit" className="w-full" disabled={!backupCode.trim()}>
                  Use backup code
                </Button>
              </form>
            )}

            <button
              type="button"
              onClick={() => { setUseBackupCode(!useBackupCode); setError(null) }}
              className="mt-4 block w-full text-center text-sm text-slate-500 hover:text-slate-700"
            >
              {useBackupCode ? 'Use authenticator app instead' : "Use a backup code instead"}
            </button>
          </div>
        </div>
      </div>
    )
  }

  // ── Normal Login ───────────────────────────────────────────────────────────
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 px-4 sm:px-0">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="mb-8 flex flex-col items-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#1B3A4B] shadow-lg">
            <TrendingUp className="h-7 w-7 text-white" />
          </div>
          <h1 className="mt-4 text-2xl font-bold text-slate-900">Welcome to PraestOS</h1>
          <p className="mt-1 text-sm text-slate-500">Sign in to your account</p>
        </div>

        {/* SSO Buttons */}
        {(showGoogle || showMicrosoft) && ORG_SLUG && (
          <div className="mb-4 space-y-2">
            {showGoogle && (
              <button
                type="button"
                onClick={() => ssoApi.beginLogin(ORG_SLUG!)}
                className="flex w-full items-center justify-center gap-3 rounded-lg border border-slate-200 bg-white px-4 py-2.5 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 transition-colors"
              >
                <svg className="h-5 w-5" viewBox="0 0 24 24" aria-hidden="true">
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
              </button>
            )}
            {showMicrosoft && (
              <button
                type="button"
                onClick={() => ssoApi.beginLogin(ORG_SLUG!)}
                className="flex w-full items-center justify-center gap-3 rounded-lg border border-slate-200 bg-white px-4 py-2.5 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 transition-colors"
              >
                <svg className="h-5 w-5" viewBox="0 0 21 21" aria-hidden="true">
                  <rect x="1" y="1" width="9" height="9" fill="#f25022" />
                  <rect x="11" y="1" width="9" height="9" fill="#7fba00" />
                  <rect x="1" y="11" width="9" height="9" fill="#00a4ef" />
                  <rect x="11" y="11" width="9" height="9" fill="#ffb900" />
                </svg>
                Sign in with Microsoft
              </button>
            )}
            <div className="relative flex items-center py-1">
              <div className="flex-grow border-t border-slate-200" />
              <span className="mx-3 shrink-0 text-xs text-slate-400">or</span>
              <div className="flex-grow border-t border-slate-200" />
            </div>
          </div>
        )}

        {/* Card */}
        <div className="rounded-xl border border-slate-200 bg-white p-8 shadow-sm">
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
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                Password
              </label>
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
              <div className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
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
