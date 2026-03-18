import apiClient from './client'
import type { Webhook, WebhookDelivery, CreateWebhookRequest, UpdateWebhookRequest, WebhookTestResult } from './types'

export const webhooksApi = {
  list: async (): Promise<Webhook[]> => {
    const { data } = await apiClient.get('/webhooks')
    return data
  },

  get: async (id: string): Promise<Webhook> => {
    const { data } = await apiClient.get(`/webhooks/${id}`)
    return data
  },

  create: async (payload: CreateWebhookRequest): Promise<Webhook> => {
    const { data } = await apiClient.post('/webhooks', payload)
    return data
  },

  update: async (id: string, payload: UpdateWebhookRequest): Promise<Webhook> => {
    const { data } = await apiClient.patch(`/webhooks/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/webhooks/${id}`)
  },

  test: async (id: string): Promise<WebhookTestResult> => {
    const { data } = await apiClient.post(`/webhooks/${id}/test`)
    return data
  },

  listDeliveries: async (id: string): Promise<WebhookDelivery[]> => {
    const { data } = await apiClient.get(`/webhooks/${id}/deliveries`)
    return data
  },
}
