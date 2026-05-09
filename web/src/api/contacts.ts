import apiClient from './client'
import type {
  Contact,
  CreateContactRequest,
  UpdateContactRequest,
  ContactListParams,
  PaginatedResponse,
  Note,
  CreateNoteRequest,
  EnrichmentResult,
} from './types'

export const contactsApi = {
  list: async (params?: ContactListParams): Promise<PaginatedResponse<Contact>> => {
    const query = params
      ? {
          page: params.page,
          limit: params.per_page,
          q: params.search,
          stage: params.stage,
          account_id: params.account_id,
          owner_id: params.owner_id,
          sort: params.sort_by,
          order: params.sort_dir,
        }
      : undefined
    const { data } = await apiClient.get('/contacts', { params: query })
    return data
  },

  search: async (q: string): Promise<Contact[]> => {
    const { data } = await apiClient.get('/contacts', { params: { q, limit: 10 } })
    return data.data ?? []
  },

  get: async (id: string): Promise<Contact> => {
    const { data } = await apiClient.get(`/contacts/${id}`)
    return data
  },

  create: async (payload: CreateContactRequest): Promise<Contact> => {
    const { data } = await apiClient.post('/contacts', payload)
    return data
  },

  update: async (id: string, payload: UpdateContactRequest): Promise<Contact> => {
    const { data } = await apiClient.patch(`/contacts/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/contacts/${id}`)
  },

  getNotes: async (id: string): Promise<Note[]> => {
    const { data } = await apiClient.get(`/contacts/${id}/notes`)
    return data
  },

  addNote: async (id: string, payload: Omit<CreateNoteRequest, 'contact_id'>): Promise<Note> => {
    const { data } = await apiClient.post(`/contacts/${id}/notes`, payload)
    return data
  },

  enrich: async (id: string): Promise<EnrichmentResult> => {
    const { data } = await apiClient.post(`/contacts/${id}/enrich`)
    return data
  },

  lookupDomain: async (domain: string): Promise<EnrichmentResult | null> => {
    try {
      const { data } = await apiClient.get('/enrich/domain', { params: { domain } })
      return data
    } catch {
      return null
    }
  },
}
