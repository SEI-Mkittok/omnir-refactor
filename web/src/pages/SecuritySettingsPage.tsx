import { useState } from 'react'
import { Shield, ShieldCheck, ShieldOff, Copy, Check, AlertCircle } from 'lucide-react'
import { totpApi } from '@/api/sso'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'

type SetupStep = 'idle' | 'qr' | 'verify' | 'backup_codes' | 'done'

export function SecuritySettingsPage() {
  // TOTP Setup
  const [step, setStep] = useState<SetupStep>('idle')
  const [qrDataUrl, setQrDataUrl] = useState('')
  const [otpUri, setOtpUri] = useState('')
  const [verifyCode, setVerifyCode] = useState('')
  const [backupCodes, setBackupCodes] = useState<string[]>([])
  const [copied, setCopied] = useState(false)
  const [setupError, setSetupError] = useState<string | null>(null)
  const [setupLoading, setSetupLoading] = useState(false)

  // TOTP Disable
  const [disablePassword, setDisablePassword] = useState('')
  const [disableError, setDisableError] = useState<string | null>(null)
  const [disableLoading, setDisableLoading] = useState(false)
  const [disableOpen, setDisableOpen] = useState(false)
  const [disabled2FA, setDisabled2FA] = useState(false)

  async function startSetup() {
    setSetupError(null)
    setSetupLoading(true)
    try {
      const result = await totpApi.setup()
      setQrDataUrl(result.qr_data_url)
      setOtpUri(result.otp_auth_uri)
      setStep('qr')
    } catch {
      setSetupError('Failed to start 2FA setup. Please try again.')
    } finally {
      setSetupLoading(false)
    }
  }

  async function confirmCode() {
    if (verifyCode.length !== 6) return
    setSetupError(null)
    setSetupLoading(true)
    try {
      const result = await totpApi.verify(verifyCode)
      setBackupCodes(result.backup_codes)
      setStep('backup_codes')
    } catch {
      setSetupError('Invalid code. Check your authenticator app and try again.')
    } finally {
      setSetupLoading(false)
    }
  }

  function copyBackupCodes() {
    navigator.clipboard.writeText(backupCodes.join('\n')).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  async function handleDisable(e: React.FormEvent) {
    e.preventDefault()
    if (!disablePassword) return
    setDisableError(null)
    setDisableLoading(true)
    try {
      await totpApi.disable(disablePassword)
      setDisabled2FA(true)
      setDisableOpen(false)
      setDisablePassword('')
      setStep('idle')
    } catch {
      setDisableError('Incorrect password. Please try again.')
    } finally {
      setDisableLoading(false)
    }
  }

  const secretKey = otpUri ? new URL(otpUri).searchParams.get('secret') : null

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Security Settings</h1>
        <p className="mt-1 text-sm text-slate-500">
          Manage two-factor authentication and account security.
        </p>
      </div>

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

        {/* ── Idle ── */}
        {step === 'idle' && !disabled2FA && (
          <Button onClick={startSetup} disabled={setupLoading} className="gap-2">
            {setupLoading ? <Spinner size="sm" /> : <ShieldCheck className="h-4 w-4" />}
            Enable 2FA
          </Button>
        )}

        {disabled2FA && (
          <div className="flex items-center gap-2 rounded-md bg-green-50 px-3 py-2 text-sm text-green-700">
            <Check className="h-4 w-4 shrink-0" />
            Two-factor authentication has been disabled.
          </div>
        )}

        {setupError && (
          <div className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
            <AlertCircle className="h-4 w-4 shrink-0" />
            {setupError}
          </div>
        )}

        {/* ── QR Code Step ── */}
        {step === 'qr' && (
          <div className="space-y-4">
            <p className="text-sm text-slate-600">
              Scan this QR code with your authenticator app, then enter the 6-digit code to confirm.
            </p>
            <div className="flex justify-center">
              <img
                src={qrDataUrl}
                alt="TOTP QR code"
                className="h-48 w-48 rounded-lg border border-slate-200"
              />
            </div>
            {secretKey && (
              <div>
                <p className="mb-1 text-xs font-medium text-slate-500 uppercase tracking-wide">
                  Manual key
                </p>
                <code className="block rounded-md bg-slate-100 px-3 py-2 text-xs font-mono break-all text-slate-700">
                  {secretKey}
                </code>
              </div>
            )}
            <Button variant="outline" size="sm" onClick={() => setStep('verify')}>
              I've scanned the code
            </Button>
          </div>
        )}

        {/* ── Verify Code Step ── */}
        {step === 'verify' && (
          <div className="space-y-4">
            <p className="text-sm text-slate-600">
              Enter the 6-digit code from your authenticator app to activate 2FA.
            </p>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                Verification code
              </label>
              <Input
                type="text"
                inputMode="numeric"
                placeholder="000000"
                maxLength={6}
                autoComplete="one-time-code"
                className="max-w-xs text-center text-xl tracking-widest"
                value={verifyCode}
                onChange={(e) => setVerifyCode(e.target.value.replace(/\D/g, ''))}
              />
            </div>
            <div className="flex gap-2">
              <Button
                onClick={confirmCode}
                disabled={verifyCode.length !== 6 || setupLoading}
                className="gap-2"
              >
                {setupLoading && <Spinner size="sm" />}
                Confirm
              </Button>
              <Button variant="outline" onClick={() => setStep('qr')}>
                Back
              </Button>
            </div>
          </div>
        )}

        {/* ── Backup Codes Step ── */}
        {step === 'backup_codes' && (
          <div className="space-y-4">
            <div className="flex items-center gap-2 rounded-md bg-green-50 px-3 py-2 text-sm text-green-700">
              <ShieldCheck className="h-4 w-4 shrink-0" />
              Two-factor authentication is now active.
            </div>
            <div>
              <p className="mb-2 text-sm font-medium text-slate-700">
                Save your backup codes
              </p>
              <p className="mb-3 text-sm text-slate-500">
                These codes can be used to access your account if you lose your authenticator device.
                Each code can only be used once.
              </p>
              <div className="grid grid-cols-2 gap-1 rounded-lg border border-slate-200 bg-slate-50 p-3">
                {backupCodes.map((code) => (
                  <code key={code} className="font-mono text-xs text-slate-700">
                    {code}
                  </code>
                ))}
              </div>
              <Button
                variant="outline"
                size="sm"
                className="mt-2 gap-2"
                onClick={copyBackupCodes}
              >
                {copied ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
                {copied ? 'Copied!' : 'Copy all codes'}
              </Button>
            </div>
            <Button onClick={() => setStep('done')}>Done</Button>
          </div>
        )}

        {/* ── Done / Disable ── */}
        {(step === 'done') && (
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
              <form onSubmit={handleDisable} className="space-y-3 rounded-lg border border-red-200 bg-red-50 p-4">
                <p className="text-sm font-medium text-red-800">Confirm your password to disable 2FA</p>
                <Input
                  type="password"
                  placeholder="Current password"
                  autoComplete="current-password"
                  value={disablePassword}
                  onChange={(e) => setDisablePassword(e.target.value)}
                  className="max-w-xs"
                />
                {disableError && (
                  <p className="text-xs text-red-600">{disableError}</p>
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
                    onClick={() => { setDisableOpen(false); setDisableError(null); setDisablePassword('') }}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
