import axios from 'axios'
import { apiClient } from './client'
import type {
  KbCategory,
  KbArticle,
  KbArticleSummary,
  KbArticleSuggest,
  KbPublicCategoryWithCount,
  KbPublicArticle,
  CreateKbCategoryRequest,
  UpdateKbCategoryRequest,
  CreateKbArticleRequest,
  UpdateKbArticleRequest,
  KbArticleListParams,
  PaginatedResponse,
} from './types'

// Unauthenticated client for public endpoints
const publicClient = axios.create({
  baseURL: (import.meta.env.VITE_API_URL
    ? import.meta.env.VITE_API_URL.replace('/api/v1', '')
    : '') + '/api/public',
  headers: { 'Content-Type': 'application/json' },
})

export const kbApi = {
  // --- Admin: Categories ---
  listCategories: async (): Promise<KbCategory[]> => {
    const { data } = await apiClient.get('/kb/categories')
    return data
  },

  createCategory: async (payload: CreateKbCategoryRequest): Promise<KbCategory> => {
    const { data } = await apiClient.post('/kb/categories', payload)
    return data
  },

  updateCategory: async (id: string, payload: UpdateKbCategoryRequest): Promise<KbCategory> => {
    const { data } = await apiClient.patch(`/kb/categories/${id}`, payload)
    return data
  },

  deleteCategory: async (id: string): Promise<void> => {
    await apiClient.delete(`/kb/categories/${id}`)
  },

  // --- Admin: Articles ---
  listArticles: async (params?: KbArticleListParams): Promise<PaginatedResponse<KbArticleSummary>> => {
    const { data } = await apiClient.get('/kb/articles', { params })
    return data
  },

  getArticle: async (id: string): Promise<KbArticle> => {
    const { data } = await apiClient.get(`/kb/articles/${id}`)
    return data
  },

  createArticle: async (payload: CreateKbArticleRequest): Promise<KbArticle> => {
    const { data } = await apiClient.post('/kb/articles', payload)
    return data
  },

  updateArticle: async (id: string, payload: UpdateKbArticleRequest): Promise<KbArticle> => {
    const { data } = await apiClient.patch(`/kb/articles/${id}`, payload)
    return data
  },

  deleteArticle: async (id: string): Promise<void> => {
    await apiClient.delete(`/kb/articles/${id}`)
  },

  // --- Ticket deflection (authenticated) ---
  suggestArticles: async (q: string): Promise<KbArticleSuggest[]> => {
    const { data } = await apiClient.get('/kb/articles/suggest', { params: { q } })
    return data
  },

  // --- Public: Help Center ---
  publicListCategories: async (orgSlug: string): Promise<KbPublicCategoryWithCount[]> => {
    const { data } = await publicClient.get(`/kb/${orgSlug}/categories`)
    return data
  },

  publicListArticles: async (orgSlug: string, q?: string): Promise<KbArticleSummary[]> => {
    const { data } = await publicClient.get(`/kb/${orgSlug}/articles`, { params: q ? { q } : undefined })
    return data
  },

  publicGetArticle: async (orgSlug: string, slug: string): Promise<KbPublicArticle> => {
    const { data } = await publicClient.get(`/kb/${orgSlug}/articles/${slug}`)
    return data
  },
}
