import apiClient from './client'
import type {
  Ticket,
  TicketComment,
  TicketAttachment,
  CreateTicketRequest,
  UpdateTicketRequest,
  TicketListParams,
  CreateTicketCommentRequest,
  PaginatedResponse,
} from './types'

export const ticketsApi = {
  list: async (params?: TicketListParams): Promise<PaginatedResponse<Ticket>> => {
    const { data } = await apiClient.get('/tickets', { params })
    return data
  },

  get: async (id: string): Promise<Ticket> => {
    const { data } = await apiClient.get(`/tickets/${id}`)
    return data
  },

  create: async (payload: CreateTicketRequest): Promise<Ticket> => {
    const { data } = await apiClient.post('/tickets', payload)
    return data
  },

  update: async (id: string, payload: UpdateTicketRequest): Promise<Ticket> => {
    const { data } = await apiClient.patch(`/tickets/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/tickets/${id}`)
  },

  getComments: async (id: string): Promise<TicketComment[]> => {
    const { data } = await apiClient.get(`/tickets/${id}/comments`)
    return data
  },

  addComment: async (id: string, payload: CreateTicketCommentRequest): Promise<TicketComment> => {
    const { data } = await apiClient.post(`/tickets/${id}/comments`, payload)
    return data
  },

  getAttachments: async (id: string): Promise<TicketAttachment[]> => {
    const { data } = await apiClient.get(`/tickets/${id}/attachments`)
    return data
  },

  uploadAttachment: async (id: string, file: File): Promise<TicketAttachment> => {
    const form = new FormData()
    form.append('file', file)
    const { data } = await apiClient.post(`/tickets/${id}/attachments`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return data
  },
}
