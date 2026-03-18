import apiClient from './client'
import type {
  Lead,
  CreateLeadRequest,
  UpdateLeadRequest,
  LeadListParams,
  PaginatedResponse,
  ConvertLeadRequest,
  ConvertLeadResponse,
} from './types'

export const leadsApi = {
  list: async (params?: LeadListParams): Promise<PaginatedResponse<Lead>> => {
    const { data } = await apiClient.get('/leads', { params })
    return data
  },

  get: async (id: string): Promise<Lead> => {
    const { data } = await apiClient.get(`/leads/${id}`)
    return data
  },

  create: async (payload: CreateLeadRequest): Promise<Lead> => {
    const { data } = await apiClient.post('/leads', payload)
    return data
  },

  update: async (id: string, payload: UpdateLeadRequest): Promise<Lead> => {
    const { data } = await apiClient.patch(`/leads/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/leads/${id}`)
  },

  convert: async (id: string, payload: ConvertLeadRequest): Promise<ConvertLeadResponse> => {
    const { data } = await apiClient.post(`/leads/${id}/convert`, payload)
    return data
  },

}
