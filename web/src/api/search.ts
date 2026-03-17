import apiClient from './client'
import type { SearchResult } from './types'

export const searchApi = {
  search: async (q: string): Promise<SearchResult> => {
    const { data } = await apiClient.get('/search', { params: { q } })
    return data
  },
}
