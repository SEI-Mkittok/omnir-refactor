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
    const { data } = await apiClient.get('/accounts', { params })
    return data
  },

  search: async (q: string): Promise<Account[]> => {
    const { data } = await apiClient.get('/accounts', { params: { search: q, per_page: 10 } })
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
    const { data } = await apiClient.get(`/accounts/${id}/contacts`)
    return data
  },

  getDeals: async (id: string): Promise<Deal[]> => {
    const { data } = await apiClient.get(`/accounts/${id}/deals`)
    return data
  },

  getTickets: async (id: string): Promise<Ticket[]> => {
    const { data } = await apiClient.get(`/accounts/${id}/tickets`)
    return data.data ?? data
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
