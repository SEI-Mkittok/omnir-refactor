import apiClient from './client'
import type {
  SavedView,
  CreateViewRequest,
  UpdateViewRequest,
  PinViewRequest,
  ViewListParams,
} from './types'
import { normalizeSavedView } from '@/lib/savedViewFilters'

export const viewsApi = {
  list: async (params?: ViewListParams): Promise<SavedView[]> => {
    const { data } = await apiClient.get('/views', { params })
    return Array.isArray(data) ? data.map(normalizeSavedView) : []
  },

  get: async (id: string): Promise<SavedView> => {
    const { data } = await apiClient.get(`/views/${id}`)
    return normalizeSavedView(data)
  },

  create: async (payload: CreateViewRequest): Promise<SavedView> => {
    const { data } = await apiClient.post('/views', payload)
    return normalizeSavedView(data)
  },

  update: async (id: string, payload: UpdateViewRequest): Promise<SavedView> => {
    const { data } = await apiClient.patch(`/views/${id}`, payload)
    return normalizeSavedView(data)
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/views/${id}`)
  },

  pin: async (id: string, payload: PinViewRequest): Promise<SavedView> => {
    const { data } = await apiClient.post(`/views/${id}/pin`, payload)
    return normalizeSavedView(data)
  },
}
