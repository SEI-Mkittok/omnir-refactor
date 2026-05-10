import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plug,
  Unplug,
  Settings2,
  KeyRound,
  Webhook,
  Link2,
  CheckCircle2,
  XCircle,
  Clock,
  Eye,
  EyeOff,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { integrationsApi, type Integration, type IntegrationProvider, type IntegrationStatus } from '@/api/integrations'
import { calendarApi, type CalendarConnection } from '@/api/calendar'

// ── Provider metadata ──────────────────────────────────────────────────────

interface ProviderMeta {
  provider: IntegrationProvider
  name: string
  description: string
  category: 'email' | 'calendar' | 'coming_soon'
  oauthPath?: string
}

type GoogleWorkspaceStatus = 'connected' | 'partial' | 'disconnected'

const PROVIDERS: ProviderMeta[] = [
  {
    provider: 'gmail',
    name: 'Gmail',
    description: 'Sync emails and send messages directly from Omnir.',
    category: 'email',
    oauthPath: '/api/integrations/email/auth/google',
  },
  {
    provider: 'outlook',
    name: 'Outlook',
    description: 'Connect Microsoft Outlook to manage email and calendar.',
    category: 'email',
    oauthPath: '/api/integrations/email/auth/microsoft',
  },
  {
    provider: 'google_calendar',
    name: 'Google Calendar',
    description: 'Sync meetings and event reminders with Google Calendar.',
    category: 'calendar',
    oauthPath: '/api/v1/calendar/auth/google',
  },
  {
    provider: 'outlook_calendar',
    name: 'Outlook Calendar',
    description: 'Sync Microsoft calendar events with CRM activity.',
    category: 'calendar',
    oauthPath: '/api/v1/calendar/auth/microsoft',
  },
  {
    provider: 'slack',
    name: 'Slack',
    description: 'Get deal and ticket notifications in Slack channels.',
    category: 'coming_soon',
  },
  {
    provider: 'microsoft_teams',
    name: 'Microsoft Teams',
    description: 'Collaborate and receive alerts in Teams.',
    category: 'coming_soon',
  },
  {
    provider: 'confluence',
    name: 'Confluence',
    description: 'Sync knowledge base articles with Confluence spaces.',
    category: 'coming_soon',
  },
  {
    provider: 'stripe',
    name: 'Stripe',
    description: 'Link deals to Stripe invoices and subscriptions.',
    category: 'coming_soon',
  },
  {
    provider: 'sendgrid',
    name: 'SendGrid',
    description: 'Route transactional email through SendGrid.',
    category: 'coming_soon',
  },
  {
    provider: 'twilio',
    name: 'Twilio',
    description: 'Send SMS and voice notifications from CRM workflows.',
    category: 'coming_soon',
  },
  {
    provider: 'zapier',
    name: 'Zapier',
    description: 'Connect Omnir events to Zapier automations.',
    category: 'coming_soon',
  },
]

function getGoogleWorkspaceStatus(
  gmailConnected: boolean,
  calendarConnected: boolean
): GoogleWorkspaceStatus {
  if (gmailConnected && calendarConnected) return 'connected'
  if (gmailConnected || calendarConnected) return 'partial'
  return 'disconnected'
}

// ── Status badge ──────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: IntegrationStatus }) {
  if (status === 'coming_soon') {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-[#F3F4F6] px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-[#6B7280]">
        Coming soon
      </span>
    )
  }
  if (status === 'connected') {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-green-700">
        <CheckCircle2 className="h-3 w-3" />
        Connected
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-[#F3F4F6] px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-[#6B7280]">
      <XCircle className="h-3 w-3" />
      Disconnected
    </span>
  )
}

// ── Logo tile ─────────────────────────────────────────────────────────────

const LOGO_COLORS: Record<IntegrationProvider, string> = {
  gmail: '#EA4335',
  outlook: '#0078D4',
  slack: '#611f69',
  microsoft_teams: '#6264A7',
  confluence: '#0052CC',
  stripe: '#635BFF',
  google_calendar: '#4285F4',
  outlook_calendar: '#0078D4',
  sendgrid: '#1A82E2',
  twilio: '#F22F46',
  zapier: '#FF4A00',
}

// SVG paths (24×24 viewBox, white fill) for each integration provider.
// Gmail, Confluence, Stripe: from simple-icons (npm). Slack, Outlook, Teams: inlined.
const PROVIDER_SVG_PATHS: Record<IntegrationProvider, string> = {
  gmail:
    'M24 5.457v13.909c0 .904-.732 1.636-1.636 1.636h-3.819V11.73L12 16.64l-6.545-4.91v9.273H1.636A1.636 1.636 0 0 1 0 19.366V5.457c0-2.023 2.309-3.178 3.927-1.964L5.455 4.64 12 9.548l6.545-4.91 1.528-1.145C21.69 2.28 24 3.434 24 5.457z',
  outlook:
    'M4 4h16a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2zm0 2v.511l8 4.8 8-4.8V6H4zm16 2.689-8 4.8-8-4.8V18h16V8.689z',
  slack:
    'M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zm1.271 0a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 3.952a2.528 2.528 0 0 1-2.521-2.521A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.521v2.521H8.834zm0 1.271a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.521A2.528 2.528 0 0 1 0 7.744a2.528 2.528 0 0 1 2.521-2.521h6.313zm11.213 2.521a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 7.744a2.528 2.528 0 0 1-2.521 2.521h-2.522V7.744zm-1.268 0a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V1.431A2.528 2.528 0 0 1 16.256 0a2.528 2.528 0 0 1 2.523 2.521v5.313zm-2.523 11.213a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 16.256 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zm0-1.268a2.527 2.527 0 0 1-2.52-2.521 2.527 2.527 0 0 1 2.52-2.521h6.313A2.528 2.528 0 0 1 24 15.165a2.528 2.528 0 0 1-2.521 2.521h-6.313z',
  microsoft_teams:
    'M16 11a3 3 0 1 0 0-6 3 3 0 0 0 0 6zm-8 0a3 3 0 1 0 0-6 3 3 0 0 0 0 6zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4zm8 0c-.33 0-.68.02-1.05.05C16.19 14.21 17 15.27 17 17v2h7v-2c0-2.66-5.33-4-8-4z',
  confluence:
    'M.87 18.257c-.248.382-.53.875-.763 1.245a.764.764 0 0 0 .255 1.04l4.965 3.054a.764.764 0 0 0 1.058-.26c.199-.332.454-.763.733-1.221 1.967-3.247 3.945-2.853 7.508-1.146l4.957 2.337a.764.764 0 0 0 1.028-.382l2.364-5.346a.764.764 0 0 0-.382-1 599.851 599.851 0 0 1-4.965-2.361C10.911 10.97 5.224 11.185.87 18.257zM23.131 5.743c.249-.405.531-.875.764-1.25a.764.764 0 0 0-.256-1.034L18.675.404a.764.764 0 0 0-1.058.26c-.195.335-.451.763-.734 1.225-1.966 3.246-3.945 2.85-7.508 1.146L4.437.694a.764.764 0 0 0-1.027.382L1.046 6.422a.764.764 0 0 0 .382 1c1.039.49 3.105 1.467 4.965 2.361 6.698 3.246 12.392 3.029 16.738-4.04z',
  stripe:
    'M13.976 9.15c-2.172-.806-3.356-1.426-3.356-2.409 0-.831.683-1.305 1.901-1.305 2.227 0 4.515.858 6.09 1.631l.89-5.494C18.252.975 15.697 0 12.165 0 9.667 0 7.589.654 6.104 1.872 4.56 3.147 3.757 4.992 3.757 7.218c0 4.039 2.467 5.76 6.476 7.219 2.585.92 3.445 1.574 3.445 2.583 0 .98-.84 1.545-2.354 1.545-1.875 0-4.965-.921-6.99-2.109l-.9 5.555C5.175 22.99 8.385 24 11.714 24c2.641 0 4.843-.624 6.328-1.813 1.664-1.305 2.525-3.236 2.525-5.732 0-4.128-2.524-5.851-6.594-7.305h.003z',
  google_calendar:
    'M7 2v2H5a2 2 0 0 0-2 2v2h18V6a2 2 0 0 0-2-2h-2V2h-2v2H9V2H7zm14 8H3v10a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V10zm-5 4v2h-3v3h-2v-3H8v-2h3v-3h2v3h3z',
  outlook_calendar:
    'M7 2v2H5a2 2 0 0 0-2 2v2h18V6a2 2 0 0 0-2-2h-2V2h-2v2H9V2H7zm14 8H3v10a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V10zm-4 3H7v2h10v-2zm-4 4H7v2h6v-2z',
  sendgrid:
    'M12 2 2 7v10l10 5 10-5V7L12 2zm0 2.2 6.5 3.25L12 10.7 5.5 7.45 12 4.2zM4 9.08l7 3.5v6.17l-7-3.5V9.08zm16 0v6.17l-7 3.5v-6.17l7-3.5z',
  twilio:
    'M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm-3.5 7a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3zm7 0a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3zm-7 5a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3zm7 0a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3z',
  zapier:
    'M11 2h2v7.17l5.07-5.07 1.42 1.42-5.07 5.07H22v2h-7.58l5.07 5.07-1.42 1.42L13 14.01V22h-2v-7.99l-5.07 5.07-1.42-1.42 5.07-5.07H2v-2h7.58L4.51 5.52 5.93 4.1 11 9.17V2z',
}

function ProviderLogo({ provider }: { provider: IntegrationProvider }) {
  const path = PROVIDER_SVG_PATHS[provider]
  return (
    <div
      className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl"
      style={{ background: LOGO_COLORS[provider] }}
      aria-hidden="true"
    >
      <svg viewBox="0 0 24 24" className="h-6 w-6 fill-white" xmlns="http://www.w3.org/2000/svg">
        <path d={path} />
      </svg>
    </div>
  )
}

// ── Disconnect confirm dialog ─────────────────────────────────────────────

function DisconnectDialog({
  providerName,
  onConfirm,
  onCancel,
  loading,
}: {
  providerName: string
  onConfirm: () => void
  onCancel: () => void
  loading: boolean
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      role="dialog"
      aria-modal="true"
      aria-labelledby="disconnect-title"
    >
      <div className="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl">
        <h2 id="disconnect-title" className="text-base font-bold text-[#1A1D23]">
          Disconnect {providerName}?
        </h2>
        <p className="mt-2 text-sm text-[#6B7280]">
          This will revoke PraestOS access to your {providerName} account. You can reconnect at any
          time.
        </p>
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="outline" size="sm" onClick={onCancel} disabled={loading}>
            Cancel
          </Button>
          <Button
            size="sm"
            className="bg-red-600 hover:bg-red-700 text-white"
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? 'Disconnecting…' : 'Disconnect'}
          </Button>
        </div>
      </div>
    </div>
  )
}

// ── Custom credentials modal ──────────────────────────────────────────────

function CredentialsModal({
  providerName,
  provider,
  onClose,
}: {
  providerName: string
  provider: IntegrationProvider
  onClose: () => void
}) {
  const queryClient = useQueryClient()
  const [clientId, setClientId] = useState('')
  const [clientSecret, setClientSecret] = useState('')
  const [showSecret, setShowSecret] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const saveMutation = useMutation({
    mutationFn: () => integrationsApi.saveCredentials(provider, { clientId, clientSecret }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['integrations'] })
      onClose()
    },
    onError: () => setError('Failed to save credentials. Please try again.'),
  })

  const clearMutation = useMutation({
    mutationFn: () => integrationsApi.clearCredentials(provider),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['integrations'] })
      onClose()
    },
    onError: () => setError('Failed to clear credentials.'),
  })

  const loading = saveMutation.isPending || clearMutation.isPending

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      role="dialog"
      aria-modal="true"
      aria-labelledby="creds-title"
    >
      <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
        <h2 id="creds-title" className="text-base font-bold text-[#1A1D23]">
          Custom OAuth credentials — {providerName}
        </h2>
        <p className="mt-1 text-sm text-[#6B7280]">
          Override the default app credentials with your own OAuth client.
        </p>

        {error && (
          <div className="mt-3 rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700">
            {error}
          </div>
        )}

        <div className="mt-4 space-y-3">
          <div>
            <label
              htmlFor="client-id"
              className="block text-[11px] font-bold uppercase tracking-widest text-[#6B7280] mb-1"
            >
              Client ID
            </label>
            <Input
              id="client-id"
              value={clientId}
              onChange={(e) => setClientId(e.target.value)}
              placeholder="Enter client ID"
              disabled={loading}
            />
          </div>
          <div>
            <label
              htmlFor="client-secret"
              className="block text-[11px] font-bold uppercase tracking-widest text-[#6B7280] mb-1"
            >
              Client Secret
            </label>
            <div className="relative">
              <Input
                id="client-secret"
                type={showSecret ? 'text' : 'password'}
                value={clientSecret}
                onChange={(e) => setClientSecret(e.target.value)}
                placeholder="Enter client secret"
                disabled={loading}
                className="pr-10"
              />
              <button
                type="button"
                onClick={() => setShowSecret((v) => !v)}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#6B7280] hover:text-[#1A1D23]"
                aria-label={showSecret ? 'Hide secret' : 'Show secret'}
              >
                {showSecret ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </div>
        </div>

        <div className="mt-5 flex items-center justify-between">
          <button
            type="button"
            className="text-sm text-red-600 hover:underline disabled:opacity-50"
            onClick={() => clearMutation.mutate()}
            disabled={loading}
          >
            Clear credentials
          </button>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button
              size="sm"
              onClick={() => saveMutation.mutate()}
              disabled={loading || !clientId || !clientSecret}
            >
              {saveMutation.isPending ? 'Saving…' : 'Save'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

// ── Integration card ──────────────────────────────────────────────────────

function IntegrationCard({
  meta,
  integration,
  calendarConnection,
}: {
  meta: ProviderMeta
  integration: Integration | undefined
  calendarConnection?: CalendarConnection
}) {
  const queryClient = useQueryClient()
  const [disconnectOpen, setDisconnectOpen] = useState(false)
  const [credentialsOpen, setCredentialsOpen] = useState(false)

  const status: IntegrationStatus = meta.category === 'coming_soon'
    ? 'coming_soon'
    : meta.category === 'calendar'
      ? (calendarConnection ? 'connected' : 'disconnected')
      : (integration?.status ?? 'disconnected')

  const disconnectMutation = useMutation({
    mutationFn: () => {
      if (meta.category === 'calendar' && calendarConnection) {
        return calendarApi.disconnect(calendarConnection.id)
      }
      return integrationsApi.disconnect(meta.provider)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['integrations'] })
      queryClient.invalidateQueries({ queryKey: ['calendar-connections'] })
      setDisconnectOpen(false)
    },
  })

  const isConnected = status === 'connected'

  return (
    <>
      <div className="flex items-start gap-4 rounded-xl border border-[#E5E7EB] bg-white p-5">
        <ProviderLogo provider={meta.provider} />

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-bold text-[#1A1D23]">{meta.name}</span>
            <StatusBadge status={status} />
          </div>
          <p className="mt-0.5 text-xs text-[#6B7280]">{meta.description}</p>

          {isConnected && meta.category === 'email' && integration && (
            <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-[#6B7280]">
              {integration.emailAddress && (
                <span className="truncate">
                  <span className="font-medium text-[#1A1D23]">{integration.emailAddress}</span>
                </span>
              )}
              {integration.lastSyncedAt && (
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  Last synced{' '}
                  {new Date(integration.lastSyncedAt).toLocaleString(undefined, {
                    dateStyle: 'medium',
                    timeStyle: 'short',
                  })}
                </span>
              )}
              {integration.hasCustomCreds && (
                <span className="font-medium text-[#1B3A4B]">Custom credentials</span>
              )}
            </div>
          )}

          {isConnected && meta.category === 'calendar' && (
            <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-[#6B7280]">
              <span className="font-medium text-[#1B3A4B]">Automatic sync every ~5 minutes</span>
              {calendarConnection?.token_expiry && (
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  Token expires{' '}
                  {new Date(calendarConnection.token_expiry).toLocaleString(undefined, {
                    dateStyle: 'medium',
                    timeStyle: 'short',
                  })}
                </span>
              )}
            </div>
          )}

        </div>

        {meta.category !== 'coming_soon' && (
          <div className="flex shrink-0 items-center gap-2">
            {meta.category === 'calendar' && isConnected ? (
              <button
                onClick={() => setDisconnectOpen(true)}
                className="inline-flex items-center gap-1 rounded-lg border border-[#E5E7EB] bg-[#F7F8FA] px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 transition-colors"
              >
                <Unplug className="h-3.5 w-3.5" />
                Disconnect
              </button>
            ) : meta.category === 'calendar' ? (
              <a
                href={meta.oauthPath}
                className="inline-flex items-center gap-1 rounded-lg bg-[#1B3A4B] px-3 py-1.5 text-xs font-semibold text-white hover:bg-[#1B3A4B]/90 transition-colors"
              >
                <Plug className="h-3.5 w-3.5" />
                Connect
              </a>
            ) : isConnected ? (
              <>
                <button
                  onClick={() => setCredentialsOpen(true)}
                  className="inline-flex items-center gap-1 rounded-lg border border-[#E5E7EB] bg-[#F7F8FA] px-3 py-1.5 text-xs font-semibold text-[#1B3A4B] hover:bg-[#E8EDF2] transition-colors"
                  title="Configure custom credentials"
                >
                  <Settings2 className="h-3.5 w-3.5" />
                  Configure
                </button>
                <button
                  onClick={() => setDisconnectOpen(true)}
                  className="inline-flex items-center gap-1 rounded-lg border border-[#E5E7EB] bg-[#F7F8FA] px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 transition-colors"
                >
                  <Unplug className="h-3.5 w-3.5" />
                  Disconnect
                </button>
              </>
            ) : (
              <>
                <button
                  onClick={() => setCredentialsOpen(true)}
                  className="inline-flex items-center gap-1 rounded-lg border border-[#E5E7EB] bg-[#F7F8FA] px-3 py-1.5 text-xs font-semibold text-[#1B3A4B] hover:bg-[#E8EDF2] transition-colors"
                  title="Configure custom credentials"
                >
                  <Settings2 className="h-3.5 w-3.5" />
                  Configure
                </button>
                <a
                  href={meta.oauthPath}
                  className="inline-flex items-center gap-1 rounded-lg bg-[#1B3A4B] px-3 py-1.5 text-xs font-semibold text-white hover:bg-[#1B3A4B]/90 transition-colors"
                >
                  <Plug className="h-3.5 w-3.5" />
                  Connect
                </a>
              </>
            )}
          </div>
        )}
      </div>

      {disconnectOpen && (
        <DisconnectDialog
          providerName={meta.name}
          onConfirm={() => disconnectMutation.mutate()}
          onCancel={() => setDisconnectOpen(false)}
          loading={disconnectMutation.isPending}
        />
      )}

      {credentialsOpen && (
        <CredentialsModal
          providerName={meta.name}
          provider={meta.provider}
          onClose={() => setCredentialsOpen(false)}
        />
      )}
    </>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────

export function IntegrationsSettingsPage() {
  const queryClient = useQueryClient()
  const [workspaceActionError, setWorkspaceActionError] = useState<string | null>(null)

  const { data: integrations, isLoading, isError } = useQuery({
    queryKey: ['integrations'],
    queryFn: integrationsApi.list,
  })
  const {
    data: calendarConnections = [],
    isLoading: isCalendarLoading,
    isError: isCalendarError,
  } = useQuery({
    queryKey: ['calendar-connections'],
    queryFn: calendarApi.listConnections,
  })

  const disconnectGmailMutation = useMutation({
    mutationFn: () => integrationsApi.disconnect('gmail'),
    onSuccess: () => {
      setWorkspaceActionError(null)
      queryClient.invalidateQueries({ queryKey: ['integrations'] })
    },
    onError: () => {
      setWorkspaceActionError('Failed to disconnect Gmail. Please try again.')
    },
  })

  const disconnectGoogleCalendarMutation = useMutation({
    mutationFn: (id: string) => calendarApi.disconnect(id),
    onSuccess: () => {
      setWorkspaceActionError(null)
      queryClient.invalidateQueries({ queryKey: ['calendar-connections'] })
    },
    onError: () => {
      setWorkspaceActionError('Failed to disconnect Google Calendar. Please try again.')
    },
  })

  const activeProviders = PROVIDERS.filter((p) => p.category !== 'coming_soon')
  const comingSoonProviders = PROVIDERS.filter((p) => p.category === 'coming_soon')

  function getIntegration(provider: IntegrationProvider): Integration | undefined {
    return integrations?.find((i) => i.provider === provider)
  }

  function getCalendarConnection(provider: IntegrationProvider): CalendarConnection | undefined {
    if (provider === 'google_calendar') return calendarConnections.find((c) => c.provider === 'google')
    if (provider === 'outlook_calendar') return calendarConnections.find((c) => c.provider === 'microsoft')
    return undefined
  }

  const gmailIntegration = getIntegration('gmail')
  const googleCalendarConnection = getCalendarConnection('google_calendar')
  const gmailConnected = gmailIntegration?.status === 'connected'
  const googleCalendarConnected = Boolean(googleCalendarConnection)
  const googleWorkspaceStatus = getGoogleWorkspaceStatus(gmailConnected, googleCalendarConnected)

  const statusLabel = googleWorkspaceStatus === 'connected'
    ? 'Connected'
    : googleWorkspaceStatus === 'partial'
      ? 'Partially connected'
      : 'Disconnected'

  const statusStyle = googleWorkspaceStatus === 'connected'
    ? 'bg-green-50 text-green-700'
    : googleWorkspaceStatus === 'partial'
      ? 'bg-amber-50 text-amber-700'
      : 'bg-[#F3F4F6] text-[#6B7280]'

  const loadErrorMessage = isError && isCalendarError
    ? 'Failed to load email and calendar integration status.'
    : isError
      ? 'Failed to load email integration status.'
      : isCalendarError
        ? 'Failed to load calendar integration status.'
        : null

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-[#7C8DB0] mb-1">
          Admin Settings
        </p>
        <h1 className="text-4xl font-black text-[#1B3A4B] leading-none tracking-tighter">
          Integrations.
        </h1>
        <p className="mt-2 text-sm text-[#6B7280] max-w-lg">
          Connect third-party tools to sync data, send emails, and automate workflows.
        </p>
      </div>

      {loadErrorMessage && (
        <div
          role="alert"
          className="rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700"
        >
          {loadErrorMessage} Please refresh and try again.
        </div>
      )}

      {/* Google workspace summary */}
      <section
        aria-labelledby="google-workspace-title"
        className="rounded-xl border border-[#E5E7EB] bg-white p-5"
      >
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 id="google-workspace-title" className="text-sm font-bold text-[#1A1D23]">
              Google Workspace
            </h2>
            <p className="mt-0.5 text-xs text-[#6B7280]">
              Unified status for Gmail inbox and Google Calendar sync.
            </p>
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <span className={`inline-flex rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider ${statusStyle}`}>
                {statusLabel}
              </span>
              <span className="text-xs text-[#6B7280]">
                Gmail: {gmailConnected ? 'Connected' : 'Disconnected'}
              </span>
              <span className="text-xs text-[#6B7280]">
                Calendar: {googleCalendarConnected ? 'Connected' : 'Disconnected'}
              </span>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {gmailConnected ? (
              <button
                onClick={() => disconnectGmailMutation.mutate()}
                className="inline-flex items-center gap-1 rounded-lg border border-[#E5E7EB] bg-[#F7F8FA] px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 transition-colors"
                disabled={disconnectGmailMutation.isPending}
              >
                <Unplug className="h-3.5 w-3.5" />
                {disconnectGmailMutation.isPending ? 'Disconnecting Gmail…' : 'Disconnect Gmail'}
              </button>
            ) : (
              <a
                href="/api/integrations/email/auth/google"
                className="inline-flex items-center gap-1 rounded-lg bg-[#1B3A4B] px-3 py-1.5 text-xs font-semibold text-white hover:bg-[#1B3A4B]/90 transition-colors"
              >
                <Link2 className="h-3.5 w-3.5" />
                Connect Gmail
              </a>
            )}
            {googleCalendarConnected ? (
              <button
                onClick={() => disconnectGoogleCalendarMutation.mutate(googleCalendarConnection!.id)}
                className="inline-flex items-center gap-1 rounded-lg border border-[#E5E7EB] bg-[#F7F8FA] px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 transition-colors"
                disabled={disconnectGoogleCalendarMutation.isPending}
              >
                <Unplug className="h-3.5 w-3.5" />
                {disconnectGoogleCalendarMutation.isPending ? 'Disconnecting Calendar…' : 'Disconnect Calendar'}
              </button>
            ) : (
              <a
                href="/api/v1/calendar/auth/google"
                className="inline-flex items-center gap-1 rounded-lg bg-[#1B3A4B] px-3 py-1.5 text-xs font-semibold text-white hover:bg-[#1B3A4B]/90 transition-colors"
              >
                <Link2 className="h-3.5 w-3.5" />
                Connect Calendar
              </a>
            )}
          </div>
        </div>

        {workspaceActionError && (
          <div className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700">
            {workspaceActionError}
          </div>
        )}
        {(isError || isCalendarError) && (
          <div className="mt-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-700">
            One or more integration status sources are unavailable. Actions remain available.
          </div>
        )}
      </section>

      {/* Connected apps */}
      <section aria-labelledby="connected-apps-title">
        <h2
          id="connected-apps-title"
          className="text-[11px] font-bold uppercase tracking-widest text-[#6B7280] mb-3"
        >
          Connected apps
        </h2>

        {isLoading || isCalendarLoading ? (
          <div className="space-y-3">
            {[1, 2].map((i) => (
              <div
                key={i}
                className="h-20 rounded-xl border border-[#E5E7EB] bg-white animate-pulse"
              />
            ))}
          </div>
        ) : (
          <div className="space-y-3">
            {activeProviders.map((meta) => (
              <IntegrationCard
                key={meta.provider}
                meta={meta}
                integration={getIntegration(meta.provider)}
                calendarConnection={getCalendarConnection(meta.provider)}
              />
            ))}
          </div>
        )}
      </section>

      {/* Coming soon */}
      <section aria-labelledby="coming-soon-title">
        <h2
          id="coming-soon-title"
          className="text-[11px] font-bold uppercase tracking-widest text-[#6B7280] mb-3"
        >
          Coming soon
        </h2>
        <div className="space-y-3 opacity-60">
          {comingSoonProviders.map((meta) => (
            <IntegrationCard key={meta.provider} meta={meta} integration={undefined} />
          ))}
        </div>
      </section>

      {/* API keys & webhooks */}
      <section aria-labelledby="api-section-title">
        <h2
          id="api-section-title"
          className="text-[11px] font-bold uppercase tracking-widest text-[#6B7280] mb-3"
        >
          API keys &amp; webhooks
        </h2>
        <div className="grid gap-3 sm:grid-cols-2">
          <Link
            to="/api-keys"
            className="flex items-center gap-4 rounded-xl border border-[#E5E7EB] bg-white p-5 hover:bg-[#F7F8FA] transition-colors group"
          >
            <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#E8EDF2]">
              <KeyRound className="h-5 w-5 text-[#1B3A4B]" />
            </div>
            <div>
              <p className="text-sm font-bold text-[#1A1D23] group-hover:text-[#1B3A4B]">
                API Keys
              </p>
              <p className="text-xs text-[#6B7280]">Manage keys for external API access.</p>
            </div>
          </Link>
          <Link
            to="/settings/webhooks"
            className="flex items-center gap-4 rounded-xl border border-[#E5E7EB] bg-white p-5 hover:bg-[#F7F8FA] transition-colors group"
          >
            <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#E8EDF2]">
              <Webhook className="h-5 w-5 text-[#1B3A4B]" />
            </div>
            <div>
              <p className="text-sm font-bold text-[#1A1D23] group-hover:text-[#1B3A4B]">
                Webhooks
              </p>
              <p className="text-xs text-[#6B7280]">Send real-time events to your endpoints.</p>
            </div>
          </Link>
        </div>
      </section>
    </div>
  )
}
