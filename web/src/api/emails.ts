import apiClient from './client'
import type { ContactEmail, SendEmailRequest, PaginatedResponse } from './types'

export const emailsApi = {
  send: async (payload: SendEmailRequest): Promise<ContactEmail> => {
    const { data } = await apiClient.post<ContactEmail>('/emails', payload)
    return data
  },

  listByContact: async (contactId: string, page = 1, limit = 50): Promise<PaginatedResponse<ContactEmail>> => {
    const { data } = await apiClient.get<PaginatedResponse<ContactEmail>>(
      `/contacts/${contactId}/emails`,
      { params: { page, limit } }
    )
    return data
  },
}
