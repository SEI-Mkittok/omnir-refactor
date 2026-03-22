import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { TrendingUp, AlertCircle } from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { OtpInput } from '@/components/ui/OtpInput'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { verifyTotp, verifyTotpBackup } from '@/api/auth'

const RATE_LIMIT_THRESHOLD = 5
const RATE_LIMIT_COOLDOWN_SECONDS = 30

export function TotpVerifyPage() {
  const navigate = useNavigate()
  const { setUser } = useAuthStore()

  const [code, setCode] = useState('')
  const [backupCode, setBackupCode] = useState('')
  const [useBackup, setUseBackup] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [failureCount, setFailureCount] = useState(0)
  const [cooldownSeconds, setCooldownSeconds] = useState(0)

  const otpRef = useRef<HTMLInputElement>(null)

  const isRateLimited = failureCount >= RATE_LIMIT_THRESHOLD && cooldownSeconds > 0

  // Guard: if no pending auth, redirect to login
  useEffect(() => {
    const token = sessionStorage.getItem('pre_auth_token')
    if (!token) {
      navigate('/login', { replace: true })
    }
  }, [navigate])

  // Auto-focus OTP input
  useEffect(() => {
    if (!useBackup) {
      otpRef.current?.focus()
    }
  }, [useBackup])

  // Countdown timer for rate limit
  useEffect(() => {
    if (cooldownSeconds <= 0) return
    const timer = setInterval(() => {
      setCooldownSeconds((s) => {
        if (s <= 1) {
          clearInterval(timer)
          return 0
        }
        return s - 1
      })
    }, 1000)
    return () => clearInterval(timer)
  }, [cooldownSeconds])

  async function submitCode(value: string) {
    if (isRateLimited || loading) return
    setLoading(true)
    setError(null)
    try {
      const result = await verifyTotp(value)
      if (result.ok && result.user) {
        sessionStorage.removeItem('pre_auth_token')
        setUser(result.user)
        navigate('/dashboard', { replace: true })
      } else {
        handleFailure(result.error ?? 'Invalid code. Try again.')
      }
    } catch {
      handleFailure('Something went wrong. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  async function submitBackupCode(e: React.FormEvent) {
    e.preventDefault()
    if (!backupCode.trim() || loading) return
    setLoading(true)
    setError(null)
    try {
      const result = await verifyTotpBackup(backupCode.trim())
      if (result.ok && result.user) {
        sessionStorage.removeItem('pre_auth_token')
        setUser(result.user)
        navigate('/dashboard', { replace: true })
      } else {
        setError(result.error ?? 'Invalid backup code.')
      }
    } catch {
      setError('Something went wrong. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  function handleFailure(message: string) {
    const next = failureCount + 1
    setFailureCount(next)
    setCode('')
    setError(message)
    // Trigger rate limiting after threshold
    if (next >= RATE_LIMIT_THRESHOLD) {
      setCooldownSeconds(RATE_LIMIT_COOLDOWN_SECONDS)
    }
    // Re-focus OTP input
    setTimeout(() => otpRef.current?.focus(), 0)
  }

  function toggleBackup() {
    setUseBackup((b) => !b)
    setError(null)
    setCode('')
    setBackupCode('')
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#F7F8FA] px-4 sm:px-0">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="mb-8 flex flex-col items-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#1B3A4B] shadow-lg">
            <TrendingUp className="h-7 w-7 text-white" />
          </div>
          <h1 className="mt-4 text-2xl font-bold text-slate-900">Two-factor authentication</h1>
          <p className="mt-1 text-sm text-slate-500">
            {useBackup
              ? 'Enter one of your 8-character backup codes'
              : 'Enter the 6-digit code from your authenticator app'}
          </p>
        </div>

        {/* Card */}
        <div className="rounded-xl border border-slate-200 bg-white p-8 shadow-sm space-y-5">
          {/* Rate limit banner */}
          {isRateLimited && (
            <div role="alert" aria-live="polite" className="flex items-center gap-2 rounded-md bg-amber-50 px-3 py-2 text-sm text-amber-800">
              <AlertCircle className="h-4 w-4 shrink-0" />
              Too many attempts. Please wait {cooldownSeconds}s before trying again.
            </div>
          )}

          {/* Error banner */}
          {error && !isRateLimited && (
            <div role="alert" className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
              <AlertCircle className="h-4 w-4 shrink-0" />
              {error}
            </div>
          )}

          {!useBackup ? (
            /* ── Authenticator code ── */
            <div className="space-y-4">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  Authenticator code
                </label>
                <OtpInput
                  ref={otpRef}
                  value={code}
                  onChange={setCode}
                  onAutoSubmit={submitCode}
                  hasError={!!error}
                  disabled={isRateLimited || loading}
                  autoFocus
                />
              </div>
              <Button
                type="button"
                className="w-full"
                disabled={code.length < 6 || loading || isRateLimited}
                onClick={() => submitCode(code)}
              >
                {loading ? 'Verifying…' : 'Verify'}
              </Button>
            </div>
          ) : (
            /* ── Backup code ── */
            <form onSubmit={submitBackupCode} className="space-y-4">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  Backup code
                </label>
                <Input
                  type="text"
                  placeholder="xxxx-xxxx"
                  autoComplete="off"
                  aria-label="Backup code"
                  value={backupCode}
                  onChange={(e) => setBackupCode(e.target.value)}
                  disabled={loading}
                />
              </div>
              <Button type="submit" className="w-full" disabled={!backupCode.trim() || loading}>
                {loading ? 'Verifying…' : 'Use backup code'}
              </Button>
            </form>
          )}

          {/* Toggle link */}
          <button
            type="button"
            onClick={toggleBackup}
            className="block w-full text-center text-sm text-slate-500 hover:text-slate-700"
          >
            {useBackup ? 'Use authenticator code instead' : 'Use a backup code instead'}
          </button>
        </div>
      </div>
    </div>
  )
}
