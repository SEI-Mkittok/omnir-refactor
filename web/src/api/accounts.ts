import apiClient from './client'
import type {
  Account,
  CreateAccountRequest,
  UpdateAccountRequest,
  AccountListParams,
  PaginatedResponse,
  Note,
  CreateNoteRequest,
  Contact,
  Deal,
  Ticket,
} from './types'

export const accountsApi = {
  list: async (params?: AccountListParams): Promise<PaginatedResponse<Account>> => {
    const query = params
      ? {
          page: params.page,
          limit: params.per_page,
          q: params.search,
          industry: params.industry,
          owner_id: params.owner_id,
          sort: params.sort_by,
          order: params.sort_dir,
        }
      : undefined
    const { data } = await apiClient.get('/accounts', { params: query })
    return data
  },

  search: async (q: string): Promise<Account[]> => {
    const { data } = await apiClient.get('/accounts', { params: { q, limit: 10 } })
    return data.data ?? []
  },

  get: async (id: string): Promise<Account> => {
    const { data } = await apiClient.get(`/accounts/${id}`)
    return data
  },

  create: async (payload: CreateAccountRequest): Promise<Account> => {
    const { data } = await apiClient.post('/accounts', payload)
    return data
  },

  update: async (id: string, payload: UpdateAccountRequest): Promise<Account> => {
    const { data } = await apiClient.patch(`/accounts/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/accounts/${id}`)
  },

  getContacts: async (id: string): Promise<Contact[]> => {
    const { data } = await apiClient.get('/contacts', {
      params: { account_id: id, limit: 200 },
    })
    return data.data ?? []
  },

  getDeals: async (id: string): Promise<Deal[]> => {
    const { data } = await apiClient.get('/deals', {
      params: { account_id: id, limit: 200 },
    })
    return data.data ?? []
  },

  getTickets: async (id: string): Promise<Ticket[]> => {
    const { data } = await apiClient.get('/tickets', {
      params: { account_id: id, per_page: 200 },
    })
    return data.data ?? []
  },

  getNotes: async (id: string): Promise<Note[]> => {
    const { data } = await apiClient.get(`/accounts/${id}/notes`)
    return data.data
  },

  addNote: async (id: string, payload: Omit<CreateNoteRequest, 'account_id'>): Promise<Note> => {
    const { data } = await apiClient.post(`/accounts/${id}/notes`, payload)
    return data
  },
}
