import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Copy, Check, Download, AlertCircle, ChevronDown, ChevronUp } from 'lucide-react'
import { getTotpSetup, verifyTotpSetup, getTotpBackupCodes } from '@/api/auth'
import { Button } from '@/components/ui/Button'
import { OtpInput } from '@/components/ui/OtpInput'
import { StepIndicator } from '@/components/ui/StepIndicator'
import { Spinner } from '@/components/ui/Spinner'

const STEPS = [
  { label: 'Scan QR code' },
  { label: 'Verify code' },
  { label: 'Save backup codes' },
]

export function TotpEnrollPage() {
  const navigate = useNavigate()
  const [currentStep, setCurrentStep] = useState(0)

  // Step 1 state
  const [qrUrl, setQrUrl] = useState('')
  const [secret, setSecret] = useState('')
  const [qrLoading, setQrLoading] = useState(true)
  const [qrError, setQrError] = useState<string | null>(null)
  const [secretExpanded, setSecretExpanded] = useState(false)
  const [secretCopied, setSecretCopied] = useState(false)

  // Step 2 state
  const [code, setCode] = useState('')
  const [verifyError, setVerifyError] = useState<string | null>(null)
  const [verifyLoading, setVerifyLoading] = useState(false)
  const otpRef = useRef<HTMLInputElement>(null)

  // Step 3 state
  const [backupCodes, setBackupCodes] = useState<string[]>([])
  const [backupCopied, setBackupCopied] = useState(false)
  const [acknowledged, setAcknowledged] = useState(false)
  const [codesLoading, setCodesLoading] = useState(false)

  const step1Ref = useRef<HTMLHeadingElement>(null)
  const step2Ref = useRef<HTMLHeadingElement>(null)
  const step3Ref = useRef<HTMLHeadingElement>(null)
  const stepRefs = [step1Ref, step2Ref, step3Ref]

  // Load QR code on mount
  useEffect(() => {
    getTotpSetup()
      .then((data) => {
        setQrUrl(data.qr_url)
        setSecret(data.secret)
      })
      .catch(() => setQrError('Failed to load QR code. Please refresh.'))
      .finally(() => setQrLoading(false))
  }, [])

  // Focus step heading on step transition
  useEffect(() => {
    stepRefs[currentStep]?.current?.focus()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentStep])

  function copySecret() {
    navigator.clipboard.writeText(secret).then(() => {
      setSecretCopied(true)
      setTimeout(() => setSecretCopied(false), 2000)
    })
  }

  async function handleVerify(value: string) {
    if (verifyLoading) return
    setVerifyLoading(true)
    setVerifyError(null)
    try {
      const result = await verifyTotpSetup(value)
      if (result.ok) {
        // Load backup codes then advance
        setCodesLoading(true)
        const codesData = await getTotpBackupCodes()
        setBackupCodes(codesData.codes)
        setCodesLoading(false)
        setCurrentStep(2)
      } else {
        setVerifyError(result.error ?? 'Invalid code. Check your authenticator app.')
        setCode('')
        setTimeout(() => otpRef.current?.focus(), 0)
      }
    } catch {
      setVerifyError('Something went wrong. Please try again.')
      setCode('')
    } finally {
      setVerifyLoading(false)
    }
  }

  function copyBackupCodes() {
    navigator.clipboard.writeText(backupCodes.join('\n')).then(() => {
      setBackupCopied(true)
      setTimeout(() => setBackupCopied(false), 2000)
    })
  }

  function downloadBackupCodes() {
    const blob = new Blob([backupCodes.join('\n')], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'omnir-backup-codes.txt'
    a.click()
    URL.revokeObjectURL(url)
  }

  function handleDone() {
    navigate('/settings/security', {
      state: { toast: 'Two-factor authentication enabled' },
    })
  }

  return (
    <div className="mx-auto max-w-lg py-10 px-4">
      <h1
        className="text-2xl font-bold text-slate-900 mb-2"
        tabIndex={-1}
      >
        Set up two-factor authentication
      </h1>
      <p className="text-sm text-slate-500 mb-8">
        Protect your account with a TOTP authenticator app.
      </p>

      <StepIndicator steps={STEPS} currentStep={currentStep} className="mb-8" />

      {/* ── Step 1: Scan QR ── */}
      {currentStep === 0 && (
        <div className="space-y-6">
          <h2
            ref={step1Ref}
            tabIndex={-1}
            className="text-lg font-semibold text-slate-900 outline-none"
          >
            Scan the QR code
          </h2>
          <p className="text-sm text-slate-600">
            Open your authenticator app (Google Authenticator, Authy, 1Password, etc.) and scan
            this code.
          </p>

          <div className="flex justify-center">
            {qrLoading ? (
              <div className="flex h-48 w-48 items-center justify-center rounded-lg border border-slate-200 bg-slate-50">
                <Spinner size="lg" />
              </div>
            ) : qrError ? (
              <div
                role="alert"
                className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700"
              >
                <AlertCircle className="h-4 w-4 shrink-0" />
                {qrError}
              </div>
            ) : (
              <img
                src={qrUrl}
                alt="TOTP QR code — scan with your authenticator app"
                className="h-48 w-48 rounded-lg border border-slate-200"
              />
            )}
          </div>

          {/* Can't scan? accordion */}
          {!qrLoading && !qrError && (
            <div className="rounded-lg border border-slate-200">
              <button
                type="button"
                onClick={() => setSecretExpanded((v) => !v)}
                className="flex w-full items-center justify-between px-4 py-3 text-sm font-medium text-slate-700 hover:bg-slate-50"
                aria-expanded={secretExpanded}
              >
                Can't scan? Enter the code manually
                {secretExpanded ? (
                  <ChevronUp className="h-4 w-4 text-slate-400" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-slate-400" />
                )}
              </button>
              {secretExpanded && (
                <div className="border-t border-slate-200 px-4 py-3 space-y-2">
                  <p className="text-xs text-slate-500">
                    Enter this Base32 secret key in your authenticator app.
                  </p>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 rounded-md bg-slate-100 px-3 py-2 text-xs font-mono break-all text-slate-700">
                      {secret}
                    </code>
                    <button
                      type="button"
                      onClick={copySecret}
                      aria-label="Copy secret key"
                      className="shrink-0 rounded-md p-2 hover:bg-slate-100 transition-colors"
                    >
                      {secretCopied ? (
                        <Check className="h-4 w-4 text-green-500" />
                      ) : (
                        <Copy className="h-4 w-4 text-slate-500" />
                      )}
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}

          <Button
            onClick={() => setCurrentStep(1)}
            disabled={qrLoading || !!qrError}
            className="w-full sm:w-auto"
          >
            I've scanned the code
          </Button>
        </div>
      )}

      {/* ── Step 2: Verify OTP ── */}
      {currentStep === 1 && (
        <div className="space-y-6">
          <h2
            ref={step2Ref}
            tabIndex={-1}
            className="text-lg font-semibold text-slate-900 outline-none"
          >
            Enter the verification code
          </h2>
          <p className="text-sm text-slate-600">
            Enter the 6-digit code shown in your authenticator app to confirm setup.
          </p>

          {verifyError && (
            <div role="alert" className="flex items-center gap-2 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">
              <AlertCircle className="h-4 w-4 shrink-0" />
              {verifyError}
            </div>
          )}

          <div className="space-y-1">
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              Authentication code
            </label>
            <OtpInput
              ref={otpRef}
              value={code}
              onChange={setCode}
              onAutoSubmit={handleVerify}
              hasError={!!verifyError}
              disabled={verifyLoading}
              autoFocus
              className="max-w-xs"
            />
          </div>

          {codesLoading && (
            <div className="flex items-center gap-2 text-sm text-slate-500">
              <Spinner size="sm" />
              Generating backup codes…
            </div>
          )}

          <div className="flex gap-3">
            <Button
              onClick={() => handleVerify(code)}
              disabled={code.length < 6 || verifyLoading || codesLoading}
              className="gap-2"
            >
              {verifyLoading && <Spinner size="sm" />}
              Verify
            </Button>
            <Button variant="outline" onClick={() => setCurrentStep(0)} disabled={verifyLoading}>
              Back
            </Button>
          </div>
        </div>
      )}

      {/* ── Step 3: Backup codes ── */}
      {currentStep === 2 && (
        <div className="space-y-6">
          <h2
            ref={step3Ref}
            tabIndex={-1}
            className="text-lg font-semibold text-slate-900 outline-none"
          >
            Save your backup codes
          </h2>
          <p className="text-sm text-slate-600">
            Store these codes somewhere safe. Each code can only be used once to access your
            account if you lose your authenticator device.
          </p>

          {/* 2-column backup codes grid */}
          <div className="grid grid-cols-2 gap-1.5 rounded-lg border border-slate-200 bg-slate-50 p-4">
            {backupCodes.map((c) => (
              <code key={c} className="font-mono text-sm text-slate-700 tracking-wider">
                {c}
              </code>
            ))}
          </div>

          <div className="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" onClick={copyBackupCodes} className="gap-2">
              {backupCopied ? (
                <Check className="h-4 w-4 text-green-500" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
              {backupCopied ? 'Copied!' : 'Copy codes'}
            </Button>
            <Button variant="outline" size="sm" onClick={downloadBackupCodes} className="gap-2">
              <Download className="h-4 w-4" />
              Download .txt
            </Button>
          </div>

          {/* Acknowledgement checkbox */}
          <label className="flex items-start gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={acknowledged}
              onChange={(e) => setAcknowledged(e.target.checked)}
              className="mt-0.5 h-4 w-4 rounded border-slate-300 text-[var(--color-primary)] focus:ring-[var(--border-focus)]"
            />
            <span className="text-sm text-slate-700">
              I have saved my backup codes in a safe place.
            </span>
          </label>

          <Button onClick={handleDone} disabled={!acknowledged} className="w-full sm:w-auto">
            Done
          </Button>
        </div>
      )}
    </div>
  )
}
