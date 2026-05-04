import apiClient from './client'
import type {
  Deal,
  CreateDealRequest,
  UpdateDealRequest,
  DealListParams,
  PaginatedResponse,
  Note,
  CreateNoteRequest,
} from './types'

export const dealsApi = {
  list: async (params?: DealListParams): Promise<PaginatedResponse<Deal>> => {
    const query = params
      ? {
          page: params.page,
          limit: params.per_page,
          q: params.search,
          stage: params.stage,
          account_id: params.account_id,
          contact_id: params.contact_id,
          pipeline_id: params.pipeline_id,
          owner_id: params.owner_id,
          sort: params.sort_by,
          order: params.sort_dir,
        }
      : undefined
    const { data } = await apiClient.get('/deals', { params: query })
    return data
  },

  search: async (q: string): Promise<Deal[]> => {
    const { data } = await apiClient.get('/deals', { params: { q, limit: 10 } })
    return data.data ?? []
  },

  get: async (id: string): Promise<Deal> => {
    const { data } = await apiClient.get(`/deals/${id}`)
    return data
  },

  create: async (payload: CreateDealRequest): Promise<Deal> => {
    const { data } = await apiClient.post('/deals', payload)
    return data
  },

  update: async (id: string, payload: UpdateDealRequest): Promise<Deal> => {
    const { data } = await apiClient.patch(`/deals/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/deals/${id}`)
  },

  getNotes: async (id: string): Promise<Note[]> => {
    const { data } = await apiClient.get(`/deals/${id}/notes`)
    return data
  },

  addNote: async (id: string, payload: Omit<CreateNoteRequest, 'deal_id'>): Promise<Note> => {
    const { data } = await apiClient.post(`/deals/${id}/notes`, payload)
    return data
  },
}
