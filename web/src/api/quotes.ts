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
    const query = params
      ? {
          page: params.page,
          limit: params.limit,
          q: params.q ?? params.search,
          status: params.status,
          account_id: params.account_id,
          deal_id: params.deal_id,
          contact_id: params.contact_id,
          sort: params.sort_by,
          order: params.sort_dir,
        }
      : undefined
    const { data } = await apiClient.get('/quotes', { params: query })
    const page = params?.page ?? 1
    const limit = params?.limit ?? 50
    const total = typeof data.total === 'number' ? data.total : data.meta?.total ?? 0
    const totalPages =
      typeof data.meta?.total_pages === 'number'
        ? data.meta.total_pages
        : limit > 0
          ? Math.ceil(total / limit)
          : 0

    return {
      data: data.data ?? [],
      meta: data.meta ?? {
        page,
        per_page: limit,
        total,
        total_pages: totalPages,
      },
    }
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
