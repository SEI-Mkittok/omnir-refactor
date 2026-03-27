import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plug,
  Unplug,
  Settings2,
  KeyRound,
  Webhook,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Clock,
  Eye,
  EyeOff,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { integrationsApi, type Integration, type IntegrationProvider } from '@/api/integrations'

// ── Provider metadata ──────────────────────────────────────────────────────

interface ProviderMeta {
  provider: IntegrationProvider
  name: string
  description: string
  comingSoon?: boolean
  oauthPath?: string
  logo: string
}

const PROVIDERS: ProviderMeta[] = [
  {
    provider: 'gmail',
    name: 'Gmail',
    description: 'Sync emails and send messages directly from PraestOS.',
    oauthPath: '/api/v1/integrations/gmail/oauth/start',
    logo: 'G',
  },
  {
    provider: 'outlook',
    name: 'Outlook',
    description: 'Connect Microsoft Outlook to manage email and calendar.',
    oauthPath: '/api/v1/integrations/outlook/oauth/start',
    logo: 'O',
  },
  {
    provider: 'slack',
    name: 'Slack',
    description: 'Get deal and ticket notifications in Slack channels.',
    comingSoon: true,
    logo: 'S',
  },
  {
    provider: 'teams',
    name: 'Microsoft Teams',
    description: 'Collaborate and receive alerts in Teams.',
    comingSoon: true,
    logo: 'T',
  },
  {
    provider: 'confluence',
    name: 'Confluence',
    description: 'Sync knowledge base articles with Confluence spaces.',
    comingSoon: true,
    logo: 'C',
  },
  {
    provider: 'stripe',
    name: 'Stripe',
    description: 'Link deals to Stripe invoices and subscriptions.',
    comingSoon: true,
    logo: '$',
  },
]

// ── Status badge ──────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: Integration['status'] | 'coming_soon' }) {
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
  if (status === 'error') {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-red-50 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-red-600">
        <AlertCircle className="h-3 w-3" />
        Error
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
  teams: '#6264A7',
  confluence: '#0052CC',
  stripe: '#635BFF',
}

function ProviderLogo({ provider, label }: { provider: IntegrationProvider; label: string }) {
  return (
    <div
      className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl text-white text-base font-black"
      style={{ background: LOGO_COLORS[provider] }}
      aria-hidden="true"
    >
      {label}
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
}: {
  meta: ProviderMeta
  integration: Integration | undefined
}) {
  const queryClient = useQueryClient()
  const [disconnectOpen, setDisconnectOpen] = useState(false)
  const [credentialsOpen, setCredentialsOpen] = useState(false)

  const status = meta.comingSoon
    ? 'coming_soon'
    : integration?.status ?? 'disconnected'

  const disconnectMutation = useMutation({
    mutationFn: () => integrationsApi.disconnect(meta.provider),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['integrations'] })
      setDisconnectOpen(false)
    },
  })

  const isConnected = status === 'connected'
  const hasError = status === 'error'

  return (
    <>
      <div className="flex items-start gap-4 rounded-xl border border-[#E5E7EB] bg-white p-5">
        <ProviderLogo provider={meta.provider} label={meta.logo} />

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-bold text-[#1A1D23]">{meta.name}</span>
            <StatusBadge status={status} />
          </div>
          <p className="mt-0.5 text-xs text-[#6B7280]">{meta.description}</p>

          {isConnected && integration && (
            <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-[#6B7280]">
              {integration.connectedEmail && (
                <span className="truncate">
                  <span className="font-medium text-[#1A1D23]">{integration.connectedEmail}</span>
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
            </div>
          )}

          {hasError && integration?.errorMessage && (
            <p className="mt-1 text-xs text-red-600">{integration.errorMessage}</p>
          )}
        </div>

        {!meta.comingSoon && (
          <div className="flex shrink-0 items-center gap-2">
            {(isConnected || hasError) ? (
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
  const { data: integrations, isLoading, isError } = useQuery({
    queryKey: ['integrations'],
    queryFn: integrationsApi.list,
  })

  const activeProviders = PROVIDERS.filter((p) => !p.comingSoon)
  const comingSoonProviders = PROVIDERS.filter((p) => p.comingSoon)

  function getIntegration(provider: IntegrationProvider): Integration | undefined {
    return integrations?.find((i) => i.provider === provider)
  }

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

      {isError && (
        <div
          role="alert"
          className="rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700"
        >
          Failed to load integrations. Please refresh and try again.
        </div>
      )}

      {/* Connected apps */}
      <section aria-labelledby="connected-apps-title">
        <h2
          id="connected-apps-title"
          className="text-[11px] font-bold uppercase tracking-widest text-[#6B7280] mb-3"
        >
          Connected apps
        </h2>

        {isLoading ? (
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
