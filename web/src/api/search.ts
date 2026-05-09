import apiClient from './client'
import type { SearchFilters, SearchResult } from './types'

export const searchApi = {
  search: async (params: string | SearchFilters): Promise<SearchResult> => {
    const requestParams = typeof params === 'string' ? { q: params } : params
    const { data } = await apiClient.get('/search', { params: requestParams })
    return data
  },
}
