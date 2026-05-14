import axios from 'axios'
import { apiClient } from './client'
import type {
  ProductHelpArticle,
  ProductHelpArticleSummary,
  ProductHelpCategory,
  ProductHelpSyncStatusResponse,
} from './types'

const publicProductHelpClient = axios.create({
  baseURL: (import.meta.env.VITE_API_URL
    ? import.meta.env.VITE_API_URL.replace('/api/v1', '')
    : '') + '/api/product-help',
  headers: { 'Content-Type': 'application/json' },
})

function arrayData<T>(data: { data?: T[] | null } | T[] | null): T[] {
  if (Array.isArray(data)) return data
  return Array.isArray(data?.data) ? data.data : []
}

export const productHelpApi = {
  listCategories: async (): Promise<ProductHelpCategory[]> => {
    const { data } = await publicProductHelpClient.get<{ data?: ProductHelpCategory[] | null } | ProductHelpCategory[] | null>('/categories')
    return arrayData(data)
  },

  listArticles: async (params?: { q?: string; category_slug?: string }): Promise<ProductHelpArticleSummary[]> => {
    const { data } = await publicProductHelpClient.get<
      { data?: ProductHelpArticleSummary[] | null } | ProductHelpArticleSummary[] | null
    >('/articles', { params })
    return arrayData(data)
  },

  search: async (q: string): Promise<ProductHelpArticleSummary[]> => {
    const { data } = await publicProductHelpClient.get<
      { data?: ProductHelpArticleSummary[] | null } | ProductHelpArticleSummary[] | null
    >('/search', { params: { q } })
    return arrayData(data)
  },

  getArticle: async (slug: string): Promise<ProductHelpArticle> => {
    const { data } = await publicProductHelpClient.get<ProductHelpArticle>(`/articles/${slug}`)
    return data
  },

  getSyncStatus: async (): Promise<ProductHelpSyncStatusResponse> => {
    const { data } = await apiClient.get<ProductHelpSyncStatusResponse>('/product-help/sync')
    return data
  },

  syncNow: async (): Promise<ProductHelpSyncStatusResponse> => {
    const { data } = await apiClient.post<ProductHelpSyncStatusResponse>('/product-help/sync')
    return data
  },
}
