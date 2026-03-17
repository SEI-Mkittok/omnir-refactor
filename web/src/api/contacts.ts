import apiClient from './client'
import type {
  Contact,
  CreateContactRequest,
  UpdateContactRequest,
  ContactListParams,
  PaginatedResponse,
  Note,
  CreateNoteRequest,
} from './types'

export const contactsApi = {
  list: async (params?: ContactListParams): Promise<PaginatedResponse<Contact>> => {
    const { data } = await apiClient.get('/contacts', { params })
    return data
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
}
