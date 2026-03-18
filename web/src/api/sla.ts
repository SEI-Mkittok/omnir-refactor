import apiClient from './client'
import type {
  SLAPolicy,
  CreateSLAPolicyRequest,
  UpdateSLAPolicyRequest,
  SLAPolicyListParams,
  PaginatedResponse,
} from './types'

export const slaApi = {
  list: async (params?: SLAPolicyListParams): Promise<PaginatedResponse<SLAPolicy>> => {
    const { data } = await apiClient.get('/sla-policies', { params })
    return data
  },

  get: async (id: string): Promise<SLAPolicy> => {
    const { data } = await apiClient.get(`/sla-policies/${id}`)
    return data
  },

  create: async (payload: CreateSLAPolicyRequest): Promise<SLAPolicy> => {
    const { data } = await apiClient.post('/sla-policies', payload)
    return data
  },

  update: async (id: string, payload: UpdateSLAPolicyRequest): Promise<SLAPolicy> => {
    const { data } = await apiClient.patch(`/sla-policies/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/sla-policies/${id}`)
  },
}
