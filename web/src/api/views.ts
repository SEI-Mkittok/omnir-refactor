import apiClient from './client'
import type {
  SavedView,
  CreateViewRequest,
  UpdateViewRequest,
  PinViewRequest,
  ViewListParams,
} from './types'

export const viewsApi = {
  list: async (params?: ViewListParams): Promise<SavedView[]> => {
    const { data } = await apiClient.get('/views', { params })
    return data
  },

  get: async (id: string): Promise<SavedView> => {
    const { data } = await apiClient.get(`/views/${id}`)
    return data
  },

  create: async (payload: CreateViewRequest): Promise<SavedView> => {
    const { data } = await apiClient.post('/views', payload)
    return data
  },

  update: async (id: string, payload: UpdateViewRequest): Promise<SavedView> => {
    const { data } = await apiClient.patch(`/views/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/views/${id}`)
  },

  pin: async (id: string, payload: PinViewRequest): Promise<SavedView> => {
    const { data } = await apiClient.post(`/views/${id}/pin`, payload)
    return data
  },
}
