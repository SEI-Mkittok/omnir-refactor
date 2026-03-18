import apiClient from './client'
import type { APIKey, CreateAPIKeyRequest, CreateAPIKeyResponse } from './types'

export const apiKeysApi = {
  list: async (): Promise<APIKey[]> => {
    const { data } = await apiClient.get('/api-keys')
    return data
  },

  create: async (payload: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> => {
    const { data } = await apiClient.post('/api-keys', payload)
    return data
  },

  revoke: async (id: string): Promise<void> => {
    await apiClient.delete(`/api-keys/${id}`)
  },
}
