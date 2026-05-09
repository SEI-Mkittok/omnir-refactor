import apiClient from './client'

// ── Types ──────────────────────────────────────────────────────────────────

export type IntegrationProvider =
  | 'gmail'
  | 'outlook'
  | 'google_calendar'
  | 'outlook_calendar'
  | 'slack'
  | 'microsoft_teams'
  | 'confluence'
  | 'stripe'
  | 'sendgrid'
  | 'twilio'
  | 'zapier'

export type IntegrationStatus = 'connected' | 'disconnected' | 'coming_soon'

interface BackendIntegration {
  provider: IntegrationProvider
  status: IntegrationStatus
  email_address?: string
  last_synced_at?: string
  has_custom_creds?: boolean
}

export interface Integration {
  provider: IntegrationProvider
  status: IntegrationStatus
  emailAddress?: string
  lastSyncedAt?: string
  hasCustomCreds: boolean
}

export interface IntegrationCredentials {
  clientId: string
  clientSecret: string
}

function normalizeIntegration(integration: BackendIntegration): Integration {
  return {
    provider: integration.provider,
    status: integration.status,
    emailAddress: integration.email_address,
    lastSyncedAt: integration.last_synced_at,
    hasCustomCreds: integration.has_custom_creds ?? false,
  }
}

// ── API ────────────────────────────────────────────────────────────────────

export const integrationsApi = {
  list(): Promise<Integration[]> {
    return apiClient.get<BackendIntegration[]>('/integrations').then((r) =>
      r.data.map(normalizeIntegration)
    )
  },

  disconnect(provider: IntegrationProvider): Promise<void> {
    return apiClient.delete(`/integrations/${provider}/connection`).then(() => undefined)
  },

  saveCredentials(provider: IntegrationProvider, creds: IntegrationCredentials): Promise<void> {
    return apiClient.put(`/integrations/${provider}/credentials`, {
      client_id: creds.clientId,
      client_secret: creds.clientSecret,
    }).then(() => undefined)
  },

  clearCredentials(provider: IntegrationProvider): Promise<void> {
    return apiClient.delete(`/integrations/${provider}/credentials`).then(() => undefined)
  },
}
