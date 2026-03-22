import apiClient from './client'
import type {
  Quote,
  CreateQuoteRequest,
  UpdateQuoteRequest,
  QuoteListParams,
  SendQuoteRequest,
  PaginatedResponse,
} from './types'

export const quotesApi = {
  list: async (params?: QuoteListParams): Promise<PaginatedResponse<Quote>> => {
    const { data } = await apiClient.get('/quotes', { params })
    return data
  },

  listByDeal: async (dealId: string): Promise<PaginatedResponse<Quote>> => {
    const { data } = await apiClient.get(`/deals/${dealId}/quotes`)
    return data
  },

  get: async (id: string): Promise<Quote> => {
    const { data } = await apiClient.get(`/quotes/${id}`)
    return data
  },

  create: async (payload: CreateQuoteRequest): Promise<Quote> => {
    const { data } = await apiClient.post('/quotes', payload)
    return data
  },

  update: async (id: string, payload: UpdateQuoteRequest): Promise<Quote> => {
    const { data } = await apiClient.patch(`/quotes/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/quotes/${id}`)
  },

  send: async (id: string, payload: SendQuoteRequest): Promise<Quote> => {
    const { data } = await apiClient.post(`/quotes/${id}/send`, payload)
    return data
  },
}
