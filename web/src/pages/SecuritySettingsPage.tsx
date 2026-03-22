import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Shield, ShieldCheck, ShieldOff, AlertCircle, Check } from 'lucide-react'
import { totpApi } from '@/api/sso'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

export function SecuritySettingsPage() {
  const location = useLocation()
  const [toast, setToast] = useState<string | null>(null)
  const [totpEnabled, setTotpEnabled] = useState(false)

  // Disable flow
  const [disablePassword, setDisablePassword] = useState('')
  const [disableError, setDisableError] = useState<string | null>(null)
  const [disableLoading, setDisableLoading] = useState(false)
  const [disableOpen, setDisableOpen] = useState(false)

  // Pick up success toast from TotpEnrollPage navigation state
  useEffect(() => {
    const state = location.state as { toast?: string } | null
    if (state?.toast) {
      setToast(state.toast)
      setTotpEnabled(true)
      // Clear state to avoid re-showing on refresh
      window.history.replaceState({}, '')
      const timer = setTimeout(() => setToast(null), 4000)
      return () => clearTimeout(timer)
    }
  }, [location.state])

  async function handleDisable(e: React.FormEvent) {
    e.preventDefault()
    if (!disablePassword) return
    setDisableError(null)
    setDisableLoading(true)
    try {
      await totpApi.disable(disablePassword)
      setTotpEnabled(false)
      setDisableOpen(false)
      setDisablePassword('')
    } catch {
      setDisableError('Incorrect password. Please try again.')
    } finally {
      setDisableLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Security Settings</h1>
        <p className="mt-1 text-sm text-slate-500">
          Manage two-factor authentication and account security.
        </p>
      </div>

      {/* Success toast */}
      {toast && (
        <div className="flex items-center gap-2 rounded-md bg-green-50 px-3 py-2 text-sm text-green-700">
          <Check className="h-4 w-4 shrink-0" />
          {toast}
        </div>
      )}

      {/* 2FA Card */}
      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm space-y-5">
        <div className="flex items-start gap-4">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100">
            <Shield className="h-5 w-5 text-slate-600" />
          </div>
          <div className="flex-1">
            <h2 className="text-base font-semibold text-slate-900">
              Two-factor authentication (TOTP)
            </h2>
            <p className="mt-0.5 text-sm text-slate-500">
              Add an extra layer of security using an authenticator app like Google Authenticator,
              Authy, or 1Password.
            </p>
          </div>
        </div>

        {totpEnabled ? (
          <div className="space-y-4">
            <div className="flex items-center gap-2 rounded-md bg-green-50 px-3 py-2 text-sm text-green-700">
              <ShieldCheck className="h-4 w-4 shrink-0" />
              Two-factor authentication is enabled.
            </div>

            {!disableOpen ? (
              <button
                type="button"
                onClick={() => setDisableOpen(true)}
                className="flex items-center gap-2 text-sm text-red-600 hover:text-red-700"
              >
                <ShieldOff className="h-4 w-4" />
                Disable 2FA
              </button>
            ) : (
              <form
                onSubmit={handleDisable}
                className="space-y-3 rounded-lg border border-red-200 bg-red-50 p-4"
              >
                <p className="text-sm font-medium text-red-800">
                  Confirm your password to disable 2FA
                </p>
                <Input
                  type="password"
                  placeholder="Current password"
                  autoComplete="current-password"
                  value={disablePassword}
                  onChange={(e) => setDisablePassword(e.target.value)}
                  className="max-w-xs"
                />
                {disableError && (
                  <div
                    role="alert"
                    className="flex items-center gap-2 text-xs text-red-600"
                  >
                    <AlertCircle className="h-3 w-3 shrink-0" />
                    {disableError}
                  </div>
                )}
                <div className="flex gap-2">
                  <Button
                    type="submit"
                    variant="outline"
                    size="sm"
                    disabled={!disablePassword || disableLoading}
                    className="border-red-300 text-red-700 hover:bg-red-100"
                  >
                    {disableLoading ? 'Disabling…' : 'Disable 2FA'}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setDisableOpen(false)
                      setDisableError(null)
                      setDisablePassword('')
                    }}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            )}
          </div>
        ) : (
          <Button asChild>
            <Link to="/settings/security/2fa/enroll">
              <ShieldCheck className="h-4 w-4" />
              Set up two-factor authentication
            </Link>
          </Button>
        )}
      </div>
    </div>
  )
}
