import apiClient from './client'
import type {
  PortalTicket,
  PortalTicketComment,
  CreatePortalTicketRequest,
  PortalTicketListParams,
  PaginatedResponse,
} from './types'

export const portalApi = {
  listTickets: async (params?: PortalTicketListParams): Promise<PaginatedResponse<PortalTicket>> => {
    const { data } = await apiClient.get('/portal/tickets', { params })
    return data
  },

  getTicket: async (id: string): Promise<PortalTicket> => {
    const { data } = await apiClient.get(`/portal/tickets/${id}`)
    return data
  },

  createTicket: async (payload: CreatePortalTicketRequest): Promise<PortalTicket> => {
    const { data } = await apiClient.post('/portal/tickets', payload)
    return data
  },

  getComments: async (ticketId: string): Promise<PortalTicketComment[]> => {
    const { data } = await apiClient.get(`/portal/tickets/${ticketId}/comments`)
    return data
  },

  addComment: async (ticketId: string, body: string): Promise<PortalTicketComment> => {
    const { data } = await apiClient.post(`/portal/tickets/${ticketId}/comments`, { body })
    return data
  },
}
