import apiClient from './client'

// ── Types ──────────────────────────────────────────────────────────────────

export type IntegrationProvider =
  | 'gmail'
  | 'outlook'
  | 'slack'
  | 'teams'
  | 'confluence'
  | 'stripe'

export type IntegrationStatus = 'connected' | 'disconnected' | 'error'

export interface Integration {
  provider: IntegrationProvider
  status: IntegrationStatus
  connectedEmail?: string
  lastSyncedAt?: string // ISO date
  errorMessage?: string
}

export interface IntegrationCredentials {
  clientId: string
  clientSecret: string
}

// ── API ────────────────────────────────────────────────────────────────────

export const integrationsApi = {
  list(): Promise<Integration[]> {
    return apiClient.get('/integrations').then((r) => r.data)
  },

  disconnect(provider: IntegrationProvider): Promise<void> {
    return apiClient.delete(`/integrations/${provider}/connection`).then(() => undefined)
  },

  saveCredentials(provider: IntegrationProvider, creds: IntegrationCredentials): Promise<void> {
    return apiClient.put(`/integrations/${provider}/credentials`, creds).then(() => undefined)
  },

  clearCredentials(provider: IntegrationProvider): Promise<void> {
    return apiClient.delete(`/integrations/${provider}/credentials`).then(() => undefined)
  },
}
