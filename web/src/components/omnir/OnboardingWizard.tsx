import { useState, useRef, useCallback, useEffect, ChangeEvent } from 'react'
import {
  X,
  Camera,
  Plus,
  Trash2,
  CheckCircle2,
  AlertCircle,
  Loader2,
  ChevronDown,
  ChevronUp,
  SkipForward,
} from 'lucide-react'
import * as RadixDialog from '@radix-ui/react-dialog'
import { Button } from '@/components/ui/Button'
import { cn } from '@/lib/utils'
import {
  updateOnboarding,
  sendInvite,
  type OnboardingState,
  type InviteRole,
  type SlaConfig,
} from '@/api/onboarding'

// ─── Constants ───────────────────────────────────────────────────────────────

const STEP_LABELS = ['Welcome', 'Invite', 'Email', 'SLA', 'Done'] as const
const STEP_COUNT = STEP_LABELS.length

const DEFAULT_SLA: SlaConfig = {
  critical_first_response: 1,
  critical_resolution: 4,
  high_first_response: 4,
  high_resolution: 24,
  medium_first_response: 8,
  medium_resolution: 48,
  low_first_response: 24,
  low_resolution: 72,
}

const ROLE_OPTIONS: { value: InviteRole; label: string }[] = [
  { value: 'admin', label: 'Admin' },
  { value: 'agent', label: 'Agent' },
  { value: 'client', label: 'Client' },
]

const PRIORITY_ROWS: {
  key: 'critical' | 'high' | 'medium' | 'low'
  label: string
  dot: string
}[] = [
  { key: 'critical', label: 'Critical', dot: 'bg-red-500' },
  { key: 'high', label: 'High', dot: 'bg-orange-400' },
  { key: 'medium', label: 'Medium', dot: 'bg-blue-500' },
  { key: 'low', label: 'Low', dot: 'bg-gray-400' },
]

// ─── Types ────────────────────────────────────────────────────────────────────

interface InviteRowData {
  email: string
  role: InviteRole
  error?: string
  sent?: boolean
}

type EmailProvider = 'gmail' | 'outlook' | 'smtp' | null

interface SmtpForm {
  host: string
  port: string
  username: string
  password: string
  tls: boolean
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function ProgressBar({ step }: { step: number }) {
  const progress = ((step + 1) / STEP_COUNT) * 100

  return (
    <div className="px-6 pt-5 pb-3">
      {/* Progress bar track */}
      <div
        className="w-full h-1 rounded-sm overflow-hidden mb-3"
        style={{ background: 'var(--border-default)' }}
        role="progressbar"
        aria-valuenow={step + 1}
        aria-valuemin={1}
        aria-valuemax={STEP_COUNT}
        aria-label={`Onboarding step ${step + 1} of ${STEP_COUNT}`}
      >
        <div
          className="h-full transition-all duration-300 ease-in-out rounded-sm"
          style={{ width: `${progress}%`, background: 'var(--color-primary)' }}
        />
      </div>

      {/* Step labels */}
      <div className="flex justify-between">
        {STEP_LABELS.map((label, i) => (
          <span
            key={label}
            aria-current={i === step ? 'step' : undefined}
            className={cn(
              'text-[11px] uppercase font-semibold tracking-wide transition-colors',
              i === step
                ? 'text-[var(--color-primary)]'
                : i < step
                  ? 'text-[var(--color-primary)] opacity-60'
                  : 'text-[var(--text-label)]'
            )}
          >
            {label}
          </span>
        ))}
      </div>
    </div>
  )
}

function ErrorBanner({ message }: { message: string }) {
  return (
    <div
      role="alert"
      aria-live="assertive"
      className="flex items-center gap-2 rounded-md px-3 py-2 text-xs text-red-700 bg-red-50 ring-1 ring-red-200"
    >
      <AlertCircle className="h-3.5 w-3.5 shrink-0" />
      {message}
    </div>
  )
}

// ─── Step 1: Welcome ──────────────────────────────────────────────────────────

interface WelcomeStepProps {
  orgName: string
  logoPreview: string | null
  onOrgNameChange: (v: string) => void
  onLogoChange: (preview: string | null, file: File | null) => void
}

function WelcomeStep({ orgName, logoPreview, onOrgNameChange, onLogoChange }: WelcomeStepProps) {
  const fileRef = useRef<HTMLInputElement>(null)
  const [dragOver, setDragOver] = useState(false)
  const [logoError, setLogoError] = useState<string | null>(null)

  function handleFile(file: File) {
    setLogoError(null)
    if (file.size > 2 * 1024 * 1024) {
      setLogoError('Max file size is 2MB')
      return
    }
    const reader = new FileReader()
    reader.onload = (e) => onLogoChange(e.target?.result as string, file)
    reader.readAsDataURL(file)
  }

  function handleInputChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (file) handleFile(file)
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) handleFile(file)
  }

  return (
    <div className="flex flex-col items-center gap-5 px-6 py-4">
      {/* Logo uploader */}
      <div className="flex flex-col items-center gap-2">
        <button
          type="button"
          aria-label="Organisation logo (optional)"
          onClick={() => fileRef.current?.click()}
          onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
          onDragLeave={() => setDragOver(false)}
          onDrop={handleDrop}
          className={cn(
            'relative h-24 w-24 rounded-full border-2 border-dashed flex items-center justify-center overflow-hidden transition-colors cursor-pointer',
            dragOver
              ? 'border-[var(--color-primary)] bg-[var(--color-primary-light)]'
              : 'border-[var(--border-default)] bg-[var(--surface-card)]',
            'hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)]'
          )}
        >
          {logoPreview ? (
            <img src={logoPreview} alt="Organisation logo preview" className="h-full w-full object-cover" />
          ) : (
            <Camera className="h-7 w-7 text-[var(--text-label)]" aria-hidden="true" />
          )}
          {logoPreview && (
            <div className="absolute inset-0 bg-black/30 flex items-center justify-center opacity-0 hover:opacity-100 transition-opacity">
              <Camera className="h-5 w-5 text-white" aria-hidden="true" />
            </div>
          )}
        </button>
        <input
          ref={fileRef}
          type="file"
          accept="image/png,image/jpeg,image/svg+xml"
          className="sr-only"
          aria-label="Upload organisation logo"
          onChange={handleInputChange}
        />
        {logoError && (
          <p role="alert" aria-live="assertive" className="text-xs text-red-600">{logoError}</p>
        )}
      </div>

      {/* Org name */}
      <div className="w-full max-w-sm space-y-1.5">
        <label htmlFor="onboarding-org-name" className="block text-sm font-medium text-[var(--text-primary)]">
          Organisation Name <span aria-hidden="true" className="text-red-500">*</span>
        </label>
        <input
          id="onboarding-org-name"
          type="text"
          value={orgName}
          onChange={(e) => onOrgNameChange(e.target.value)}
          placeholder="Your organisation name"
          className={cn(
            'w-full rounded-md border px-3 py-2 text-sm text-[var(--text-primary)] placeholder:text-[var(--text-label)]',
            'focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] focus:border-transparent',
            'border-[var(--border-default)]'
          )}
          aria-required="true"
          autoFocus
        />
        <p className="text-xs text-[var(--text-secondary)]">
          This is how your workspace will appear to your team.
        </p>
      </div>
    </div>
  )
}

// ─── Step 2: Invite Team ──────────────────────────────────────────────────────

interface InviteStepProps {
  rows: InviteRowData[]
  onChange: (rows: InviteRowData[]) => void
}

function InviteStep({ rows, onChange }: InviteStepProps) {
  function updateRow(index: number, patch: Partial<InviteRowData>) {
    const next = rows.map((r, i) => (i === index ? { ...r, ...patch } : r))
    onChange(next)
  }

  function addRow() {
    if (rows.length < 10) {
      onChange([...rows, { email: '', role: 'agent' }])
    }
  }

  function removeRow(index: number) {
    onChange(rows.filter((_, i) => i !== index))
  }

  function validateEmail(email: string): string | undefined {
    if (!email) return undefined
    const ok = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
    if (!ok) return 'Invalid email format'
    const dup = rows.filter((r) => r.email === email).length > 1
    if (dup) return 'This email is already in the list'
    return undefined
  }

  function handleEmailBlur(index: number) {
    updateRow(index, { error: validateEmail(rows[index].email) })
  }

  return (
    <div className="px-6 py-4 space-y-3">
      <div>
        <h3 className="text-sm font-semibold text-[var(--text-primary)]">Invite your team</h3>
        <p className="text-xs text-[var(--text-secondary)] mt-0.5">
          Invitations are sent immediately when you continue.
        </p>
      </div>

      <div className="space-y-2">
        {rows.map((row, i) => (
          <div
            key={i}
            aria-label={`Invite row ${i + 1}`}
            className="flex gap-2 items-start"
          >
            <div className="flex-1 space-y-1">
              <input
                type="email"
                value={row.email}
                onChange={(e) => updateRow(i, { email: e.target.value, error: undefined })}
                onBlur={() => handleEmailBlur(i)}
                placeholder="colleague@example.com"
                aria-label={`Email address for invite ${i + 1}`}
                className={cn(
                  'w-full rounded-md border px-3 py-2 text-sm placeholder:text-[var(--text-label)]',
                  'focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] focus:border-transparent',
                  row.error ? 'border-red-400' : 'border-[var(--border-default)]',
                  row.sent && 'bg-green-50 border-green-300 text-green-700'
                )}
                disabled={row.sent}
              />
              {row.error && (
                <p role="alert" aria-live="assertive" className="text-xs text-red-600">{row.error}</p>
              )}
              {row.sent && (
                <p className="text-xs text-green-600 flex items-center gap-1">
                  <CheckCircle2 className="h-3 w-3" aria-hidden="true" />
                  Invite sent
                </p>
              )}
            </div>

            <select
              value={row.role}
              onChange={(e) => updateRow(i, { role: e.target.value as InviteRole })}
              aria-label={`Role for invite ${i + 1}`}
              disabled={row.sent}
              className="rounded-md border border-[var(--border-default)] px-2 py-2 text-sm text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)] bg-white"
            >
              {ROLE_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>

            {rows.length > 1 && !row.sent && (
              <button
                type="button"
                onClick={() => removeRow(i)}
                aria-label={`Remove invite row ${i + 1}`}
                className="mt-2 text-[var(--text-label)] hover:text-red-500 transition-colors"
              >
                <Trash2 className="h-4 w-4" aria-hidden="true" />
              </button>
            )}
          </div>
        ))}
      </div>

      {rows.length < 10 && (
        <button
          type="button"
          onClick={addRow}
          className="flex items-center gap-1.5 text-xs font-medium text-[var(--color-primary)] hover:underline"
        >
          <Plus className="h-3.5 w-3.5" aria-hidden="true" />
          Add another
        </button>
      )}
    </div>
  )
}

// ─── Step 3: Connect Email ────────────────────────────────────────────────────

interface EmailStepProps {
  connected: EmailProvider
  onConnect: (provider: EmailProvider) => void
}

function EmailStep({ connected, onConnect }: EmailStepProps) {
  const [smtpOpen, setSmtpOpen] = useState(false)
  const [smtpForm, setSmtpForm] = useState<SmtpForm>({
    host: '', port: '587', username: '', password: '', tls: true,
  })
  const [connecting, setConnecting] = useState<EmailProvider>(null)

  async function handleOAuth(provider: 'gmail' | 'outlook') {
    setConnecting(provider)
    // OAuth stub — opens a popup. In production this would open the OAuth flow.
    await new Promise((res) => setTimeout(res, 1200))
    onConnect(provider)
    setConnecting(null)
  }

  function handleSmtpSave() {
    if (smtpForm.host && smtpForm.username) {
      onConnect('smtp')
      setSmtpOpen(false)
    }
  }

  const providers: { id: 'gmail' | 'outlook'; label: string; icon: string }[] = [
    { id: 'gmail', label: 'Connect Gmail', icon: 'G' },
    { id: 'outlook', label: 'Connect Outlook', icon: 'O' },
  ]

  return (
    <div className="px-6 py-4 space-y-3">
      <div>
        <h3 className="text-sm font-semibold text-[var(--text-primary)]">Connect your email</h3>
        <p className="text-xs text-[var(--text-secondary)] mt-0.5">
          You can connect email later in Settings → Integrations.
        </p>
      </div>

      <div className="divide-y divide-[var(--border-default)] rounded-lg border border-[var(--border-default)] overflow-hidden">
        {providers.map((p) => (
          <div key={p.id} className="flex items-center gap-3 px-4 py-3 bg-white">
            <div
              aria-hidden="true"
              className="h-6 w-6 rounded-full bg-[var(--color-primary-light)] text-[var(--color-primary)] text-xs font-bold flex items-center justify-center shrink-0"
            >
              {p.icon}
            </div>
            <span className="flex-1 text-sm font-medium text-[var(--text-primary)]">{p.label}</span>
            {connected === p.id ? (
              <span className="flex items-center gap-1 text-xs font-medium text-green-600">
                <CheckCircle2 className="h-3.5 w-3.5" aria-hidden="true" />
                Connected
              </span>
            ) : (
              <button
                type="button"
                onClick={() => handleOAuth(p.id)}
                disabled={connecting !== null}
                className="flex items-center gap-1.5 rounded-md border border-[var(--border-default)] px-3 py-1.5 text-xs font-medium text-[var(--text-primary)] hover:bg-[var(--surface-app)] transition-colors disabled:opacity-50"
              >
                {connecting === p.id ? (
                  <Loader2 className="h-3 w-3 animate-spin" aria-hidden="true" />
                ) : null}
                Connect →
              </button>
            )}
          </div>
        ))}

        {/* SMTP accordion */}
        <div className="bg-white">
          <button
            type="button"
            onClick={() => setSmtpOpen((v) => !v)}
            className="flex w-full items-center gap-3 px-4 py-3 text-left"
            aria-expanded={smtpOpen}
          >
            <div
              aria-hidden="true"
              className="h-6 w-6 rounded-full bg-[var(--color-primary-light)] text-[var(--color-primary)] text-xs font-bold flex items-center justify-center shrink-0"
            >
              S
            </div>
            <span className="flex-1 text-sm font-medium text-[var(--text-primary)]">Custom SMTP</span>
            {connected === 'smtp' ? (
              <span className="flex items-center gap-1 text-xs font-medium text-green-600">
                <CheckCircle2 className="h-3.5 w-3.5" aria-hidden="true" />
                Connected
              </span>
            ) : (
              <>
                <span className="text-xs text-[var(--text-secondary)] mr-1">Configure →</span>
                {smtpOpen ? (
                  <ChevronUp className="h-4 w-4 text-[var(--text-label)]" aria-hidden="true" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-[var(--text-label)]" aria-hidden="true" />
                )}
              </>
            )}
          </button>

          {smtpOpen && connected !== 'smtp' && (
            <div className="px-4 pb-4 space-y-2 border-t border-[var(--border-subtle)]">
              <div className="grid grid-cols-3 gap-2 pt-3">
                <div className="col-span-2">
                  <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Host</label>
                  <input
                    type="text"
                    value={smtpForm.host}
                    onChange={(e) => setSmtpForm((f) => ({ ...f, host: e.target.value }))}
                    placeholder="smtp.example.com"
                    className="w-full rounded-md border border-[var(--border-default)] px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Port</label>
                  <input
                    type="text"
                    value={smtpForm.port}
                    onChange={(e) => setSmtpForm((f) => ({ ...f, port: e.target.value }))}
                    className="w-full rounded-md border border-[var(--border-default)] px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Username</label>
                <input
                  type="text"
                  value={smtpForm.username}
                  onChange={(e) => setSmtpForm((f) => ({ ...f, username: e.target.value }))}
                  placeholder="user@example.com"
                  className="w-full rounded-md border border-[var(--border-default)] px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Password</label>
                <input
                  type="password"
                  value={smtpForm.password}
                  onChange={(e) => setSmtpForm((f) => ({ ...f, password: e.target.value }))}
                  className="w-full rounded-md border border-[var(--border-default)] px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
                />
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="smtp-tls"
                  checked={smtpForm.tls}
                  onChange={(e) => setSmtpForm((f) => ({ ...f, tls: e.target.checked }))}
                  className="rounded border-[var(--border-default)]"
                />
                <label htmlFor="smtp-tls" className="text-xs text-[var(--text-secondary)]">Use TLS</label>
              </div>
              <Button
                type="button"
                size="sm"
                onClick={handleSmtpSave}
                disabled={!smtpForm.host || !smtpForm.username}
              >
                Save SMTP
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// ─── Step 4: Configure SLA ────────────────────────────────────────────────────

interface SlaStepProps {
  sla: SlaConfig
  onChange: (sla: SlaConfig) => void
}

function SlaStep({ sla, onChange }: SlaStepProps) {
  function update(field: keyof SlaConfig, raw: string) {
    const val = parseFloat(raw)
    if (!isNaN(val) && val >= 0.5) {
      onChange({ ...sla, [field]: val })
    }
  }

  function NumberInput({ field, value }: { field: keyof SlaConfig; value: number }) {
    return (
      <input
        type="number"
        min={0.5}
        step={0.5}
        value={value}
        onChange={(e) => update(field, e.target.value)}
        className="w-16 rounded-md border border-[var(--border-default)] px-2 py-1 text-sm text-center focus:outline-none focus:ring-1 focus:ring-[var(--border-focus)]"
        aria-label={`${field} hours`}
      />
    )
  }

  return (
    <div className="px-6 py-4 space-y-3">
      <div>
        <h3 className="text-sm font-semibold text-[var(--text-primary)]">Default SLA response times</h3>
        <p className="text-xs text-[var(--text-secondary)] mt-0.5">
          These apply to all new tickets unless overridden per team.
        </p>
      </div>

      <div className="overflow-hidden rounded-lg border border-[var(--border-default)]">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[var(--border-default)] bg-[var(--surface-app)]">
              <th className="py-2 px-3 text-left text-xs font-semibold text-[var(--text-secondary)]">Priority</th>
              <th className="py-2 px-3 text-center text-xs font-semibold text-[var(--text-secondary)]">First Response</th>
              <th className="py-2 px-3 text-center text-xs font-semibold text-[var(--text-secondary)]">Resolution</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-[var(--border-subtle)]">
            {PRIORITY_ROWS.map(({ key, label, dot }) => (
              <tr key={key}>
                <td className="py-2.5 px-3">
                  <span className="flex items-center gap-2">
                    <span className={cn('h-2.5 w-2.5 rounded-full shrink-0', dot)} aria-hidden="true" />
                    <span className="text-[var(--text-primary)] text-sm">{label}</span>
                  </span>
                </td>
                <td className="py-2.5 px-3 text-center">
                  <span className="flex items-center justify-center gap-1">
                    <NumberInput
                      field={`${key}_first_response` as keyof SlaConfig}
                      value={sla[`${key}_first_response` as keyof SlaConfig]}
                    />
                    <span className="text-xs text-[var(--text-secondary)]">hrs</span>
                  </span>
                </td>
                <td className="py-2.5 px-3 text-center">
                  <span className="flex items-center justify-center gap-1">
                    <NumberInput
                      field={`${key}_resolution` as keyof SlaConfig}
                      value={sla[`${key}_resolution` as keyof SlaConfig]}
                    />
                    <span className="text-xs text-[var(--text-secondary)]">hrs</span>
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

// ─── Step 5: Done ─────────────────────────────────────────────────────────────

interface DoneStepProps {
  orgName: string
  invitesSent: number
  emailConnected: EmailProvider
  slaConfigured: boolean
  onGoToDashboard: () => void
}

function DoneStep({ orgName, invitesSent, emailConnected, slaConfigured, onGoToDashboard }: DoneStepProps) {
  const items: { done: boolean; text: string; skippedLink?: string }[] = [
    {
      done: !!orgName,
      text: orgName ? `Organisation: ${orgName}` : 'Organisation: not set',
    },
    {
      done: invitesSent > 0,
      text: invitesSent > 0 ? `Team invited: ${invitesSent} member${invitesSent > 1 ? 's' : ''}` : 'Team: not invited',
      skippedLink: '/users',
    },
    {
      done: emailConnected !== null,
      text: emailConnected
        ? `Email connected: ${emailConnected === 'gmail' ? 'Gmail' : emailConnected === 'outlook' ? 'Outlook' : 'SMTP'}`
        : 'Email: Skipped — connect in Settings → Integrations',
      skippedLink: '/settings/webhooks',
    },
    {
      done: slaConfigured,
      text: slaConfigured ? 'SLA targets configured' : 'SLA: Skipped — configure in Settings → SLA Policies',
      skippedLink: '/settings/sla',
    },
  ]

  const quickLinks: { label: string; href: string }[] = [
    { label: 'View Dashboard', href: '/dashboard' },
    { label: 'Manage Team', href: '/users' },
    { label: 'Connect Integrations', href: '/settings/webhooks' },
    { label: 'Edit SLA Settings', href: '/settings/sla' },
  ]

  return (
    <div className="px-6 py-4 space-y-5">
      <div className="text-center">
        <div className="text-3xl mb-1" aria-hidden="true">🎉</div>
        <h3 className="text-lg font-semibold text-[var(--text-primary)]">You're all set.</h3>
        <p className="text-sm text-[var(--text-secondary)]">Here's a summary of your setup.</p>
      </div>

      {/* Checklist */}
      <ul className="space-y-2" aria-label="Setup checklist">
        {items.map((item, i) => (
          <li key={i} className="flex items-start gap-2.5 text-sm">
            {item.done ? (
              <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5 shrink-0" aria-hidden="true" />
            ) : (
              <SkipForward className="h-4 w-4 text-[var(--text-label)] mt-0.5 shrink-0" aria-hidden="true" />
            )}
            <span className={item.done ? 'text-[var(--text-primary)]' : 'text-[var(--text-secondary)]'}>
              {item.text}
            </span>
          </li>
        ))}
      </ul>

      {/* Quick links */}
      <div className="border-t border-[var(--border-subtle)] pt-4 space-y-1.5">
        <p className="text-xs font-semibold text-[var(--text-secondary)] uppercase tracking-wide">Quick links</p>
        <div className="flex flex-wrap gap-x-4 gap-y-1">
          {quickLinks.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-sm text-[var(--color-primary)] hover:underline"
              onClick={(e) => { e.preventDefault(); onGoToDashboard(); window.location.href = link.href }}
            >
              {link.label}
            </a>
          ))}
        </div>
      </div>

      {/* CTA */}
      <Button
        type="button"
        className="w-full"
        onClick={onGoToDashboard}
      >
        Go to Dashboard →
      </Button>
    </div>
  )
}

// ─── Main Wizard ──────────────────────────────────────────────────────────────

export interface OnboardingWizardProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialState: OnboardingState | null
  onComplete: () => void
  onDismiss: () => void
}

export function OnboardingWizard({
  open,
  onOpenChange,
  initialState,
  onComplete,
  onDismiss,
}: OnboardingWizardProps) {
  const [step, setStep] = useState(0)
  const [orgName, setOrgName] = useState('')
  const [logoPreview, setLogoPreview] = useState<string | null>(null)
  const [, setLogoFile] = useState<File | null>(null)
  const [inviteRows, setInviteRows] = useState<InviteRowData[]>([{ email: '', role: 'agent' }])
  const [emailConnected, setEmailConnected] = useState<EmailProvider>(null)
  const [sla, setSla] = useState<SlaConfig>(DEFAULT_SLA)
  const [completedSteps, setCompletedSteps] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [invitesSentCount, setInvitesSentCount] = useState(0)
  const [slaConfigured, setSlaConfigured] = useState(false)

  // Sync initial state from backend
  useEffect(() => {
    if (initialState) {
      setCompletedSteps(initialState.completedSteps ?? [])
      if (initialState.orgName) setOrgName(initialState.orgName)
      // Resume at first incomplete step
      const stepMap = ['welcome', 'invite', 'email', 'sla', 'done']
      const firstIncomplete = stepMap.findIndex((s) => !(initialState.completedSteps ?? []).includes(s))
      setStep(firstIncomplete === -1 ? 4 : firstIncomplete)
    }
  }, [initialState])

  const handleLogoChange = useCallback((preview: string | null, file: File | null) => {
    setLogoPreview(preview)
    setLogoFile(file)
  }, [])

  async function saveStep(stepKey: string, extra: Record<string, unknown> = {}) {
    const next = completedSteps.includes(stepKey)
      ? completedSteps
      : [...completedSteps, stepKey]
    setCompletedSteps(next)
    await updateOnboarding({ completedSteps: next, ...extra })
  }

  async function handleContinue() {
    setError(null)
    setSaving(true)

    try {
      if (step === 0) {
        // Welcome — org name required
        if (!orgName.trim()) {
          setError('Organisation name is required')
          setSaving(false)
          return
        }
        await saveStep('welcome', { orgName: orgName.trim() })
        setStep(1)

      } else if (step === 1) {
        // Invite — send invites for filled rows
        const validRows = inviteRows.filter((r) => r.email.trim() && !r.error && !r.sent)
        let sentCount = invitesSentCount

        for (const row of validRows) {
          try {
            await sendInvite({ email: row.email.trim(), role: row.role })
            sentCount++
            setInviteRows((prev) =>
              prev.map((r) => (r.email === row.email ? { ...r, sent: true } : r))
            )
          } catch {
            setInviteRows((prev) =>
              prev.map((r) =>
                r.email === row.email ? { ...r, error: 'Failed to send invite' } : r
              )
            )
          }
        }
        setInvitesSentCount(sentCount)
        await saveStep('invite')
        setStep(2)

      } else if (step === 2) {
        // Email (connected or skipped)
        await saveStep('email')
        setStep(3)

      } else if (step === 3) {
        // SLA
        setSlaConfigured(true)
        await saveStep('sla', { sla })
        setStep(4)
      }
    } catch {
      setError('Failed to save — please retry.')
    } finally {
      setSaving(false)
    }
  }

  async function handleSkip() {
    setError(null)
    setSaving(true)
    try {
      if (step === 2) {
        await saveStep('email')
      } else if (step === 3) {
        await saveStep('sla')
      }
      setStep((s) => s + 1)
    } catch {
      setError('Failed to save — please retry.')
    } finally {
      setSaving(false)
    }
  }

  function handleBack() {
    setError(null)
    setStep((s) => Math.max(0, s - 1))
  }

  async function handleGoToDashboard() {
    setSaving(true)
    try {
      await updateOnboarding({ completed: true, completedSteps: ['welcome', 'invite', 'email', 'sla', 'done'] })
    } catch {
      // best-effort
    } finally {
      setSaving(false)
    }
    onComplete()
    onOpenChange(false)
  }

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      // Mid-wizard dismiss
      if (step < 4) {
        onDismiss()
      }
    }
    onOpenChange(nextOpen)
  }

  const isLastStep = step === 4
  const showSkip = step === 2 || step === 3
  const showBack = step > 0 && step < 4

  const stepHeadingId = `onboarding-step-heading-${step}`

  return (
    <RadixDialog.Root open={open} onOpenChange={handleOpenChange}>
      <RadixDialog.Portal>
        {/* Overlay */}
        <RadixDialog.Overlay
          className="fixed inset-0 bg-black/40 z-[var(--z-modal)] data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
          style={{ zIndex: 'var(--z-overlay)' }}
        />

        {/* Modal */}
        <RadixDialog.Content
          role="dialog"
          aria-modal="true"
          aria-labelledby={stepHeadingId}
          className={cn(
            'fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 bg-white shadow-[0_8px_32px_rgba(0,0,0,0.12)] w-[calc(100%-2rem)] max-w-[640px] flex flex-col',
            'rounded-xl',
            'max-h-[90vh] overflow-y-auto',
            'z-[400]',
            'data-[state=open]:animate-in data-[state=closed]:animate-out',
            'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
            'data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95',
            // Mobile: full-screen
            'sm:rounded-xl max-sm:rounded-none max-sm:inset-0 max-sm:max-w-full max-sm:w-full max-sm:translate-x-0 max-sm:translate-y-0 max-sm:left-0 max-sm:top-0'
          )}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-6 pt-5 pb-0 shrink-0">
            <div>
              <p className="text-[10px] font-semibold text-[var(--text-label)] uppercase tracking-widest">
                PraestOS CRM Enterprise
              </p>
              <RadixDialog.Title id={stepHeadingId} className="text-base font-semibold text-[var(--text-primary)] mt-0.5">
                {STEP_LABELS[step]}
              </RadixDialog.Title>
            </div>
            <RadixDialog.Close
              aria-label="Close onboarding wizard"
              className="rounded-md p-1.5 text-[var(--text-label)] hover:text-[var(--text-primary)] hover:bg-[var(--surface-app)] transition-colors focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
            >
              <X className="h-4 w-4" aria-hidden="true" />
            </RadixDialog.Close>
          </div>

          {/* Progress */}
          <ProgressBar step={step} />

          {/* Divider */}
          <div className="h-px bg-[var(--border-subtle)] shrink-0" aria-hidden="true" />

          {/* Content */}
          <div className="flex-1 overflow-y-auto">
            {error && (
              <div className="px-6 pt-4">
                <ErrorBanner message={error} />
              </div>
            )}

            {step === 0 && (
              <WelcomeStep
                orgName={orgName}
                logoPreview={logoPreview}
                onOrgNameChange={setOrgName}
                onLogoChange={handleLogoChange}
              />
            )}
            {step === 1 && (
              <InviteStep rows={inviteRows} onChange={setInviteRows} />
            )}
            {step === 2 && (
              <EmailStep connected={emailConnected} onConnect={setEmailConnected} />
            )}
            {step === 3 && (
              <SlaStep sla={sla} onChange={setSla} />
            )}
            {step === 4 && (
              <DoneStep
                orgName={orgName}
                invitesSent={invitesSentCount}
                emailConnected={emailConnected}
                slaConfigured={slaConfigured}
                onGoToDashboard={handleGoToDashboard}
              />
            )}
          </div>

          {/* Footer */}
          {!isLastStep && (
            <>
              <div className="h-px bg-[var(--border-subtle)] shrink-0" aria-hidden="true" />
              <div className="flex items-center justify-between gap-2 px-6 py-4 shrink-0 max-sm:flex-col-reverse max-sm:items-stretch">
                {/* Back */}
                {showBack ? (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={handleBack}
                    disabled={saving}
                  >
                    Back
                  </Button>
                ) : (
                  <div aria-hidden="true" />
                )}

                {/* Skip + Continue */}
                <div className="flex items-center gap-2 max-sm:flex-col-reverse">
                  {showSkip && (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={handleSkip}
                      disabled={saving}
                      aria-label={`Skip ${STEP_LABELS[step]} step`}
                    >
                      <SkipForward className="h-3.5 w-3.5" aria-hidden="true" />
                      Skip
                    </Button>
                  )}
                  <Button
                    type="button"
                    size="sm"
                    onClick={handleContinue}
                    disabled={saving}
                  >
                    {saving ? (
                      <>
                        <Loader2 className="h-3.5 w-3.5 animate-spin" aria-hidden="true" />
                        Saving…
                      </>
                    ) : (
                      'Continue →'
                    )}
                  </Button>
                </div>
              </div>
            </>
          )}
        </RadixDialog.Content>
      </RadixDialog.Portal>
    </RadixDialog.Root>
  )
}
