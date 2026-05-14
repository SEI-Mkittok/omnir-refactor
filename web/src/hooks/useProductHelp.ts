import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { productHelpApi } from '@/api/productHelp'

export const productHelpKeys = {
  all: ['product-help'] as const,
  categories: () => [...productHelpKeys.all, 'categories'] as const,
  articles: (params?: { q?: string; category_slug?: string }) => [...productHelpKeys.all, 'articles', params] as const,
  article: (slug: string) => [...productHelpKeys.all, 'article', slug] as const,
  search: (q: string) => [...productHelpKeys.all, 'search', q] as const,
  sync: () => [...productHelpKeys.all, 'sync'] as const,
}

export function useProductHelpCategories() {
  return useQuery({
    queryKey: productHelpKeys.categories(),
    queryFn: productHelpApi.listCategories,
    staleTime: 60_000,
  })
}

export function useProductHelpArticles(params?: { q?: string; category_slug?: string }) {
  return useQuery({
    queryKey: productHelpKeys.articles(params),
    queryFn: () => productHelpApi.listArticles(params),
    staleTime: 60_000,
  })
}

export function useProductHelpSearch(q: string) {
  return useQuery({
    queryKey: productHelpKeys.search(q),
    queryFn: () => productHelpApi.search(q),
    staleTime: 60_000,
    enabled: q.trim().length > 0,
  })
}

export function useProductHelpArticle(slug: string) {
  return useQuery({
    queryKey: productHelpKeys.article(slug),
    queryFn: () => productHelpApi.getArticle(slug),
    staleTime: 60_000,
    enabled: !!slug,
  })
}

export function useProductHelpSyncStatus() {
  return useQuery({
    queryKey: productHelpKeys.sync(),
    queryFn: productHelpApi.getSyncStatus,
    staleTime: 30_000,
  })
}

export function useProductHelpSyncNow() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: productHelpApi.syncNow,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: productHelpKeys.sync() })
      qc.invalidateQueries({ queryKey: productHelpKeys.categories() })
      qc.invalidateQueries({ queryKey: productHelpKeys.all })
    },
  })
}
