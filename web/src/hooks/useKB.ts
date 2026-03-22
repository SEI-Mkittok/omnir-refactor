import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { kbApi } from '@/api/kb'
import type {
  KbArticleListParams,
  CreateKbCategoryRequest,
  UpdateKbCategoryRequest,
  CreateKbArticleRequest,
  UpdateKbArticleRequest,
} from '@/api/types'

export const kbKeys = {
  all: ['kb'] as const,
  categories: () => [...kbKeys.all, 'categories'] as const,
  articles: () => [...kbKeys.all, 'articles'] as const,
  articleList: (params?: KbArticleListParams) => [...kbKeys.articles(), 'list', params] as const,
  article: (id: string) => [...kbKeys.articles(), 'detail', id] as const,
  suggest: (q: string) => [...kbKeys.all, 'suggest', q] as const,
  publicCategories: (orgSlug: string) => [...kbKeys.all, 'public', orgSlug, 'categories'] as const,
  publicArticles: (orgSlug: string, q?: string) => [...kbKeys.all, 'public', orgSlug, 'articles', q] as const,
  publicArticle: (orgSlug: string, slug: string) => [...kbKeys.all, 'public', orgSlug, 'article', slug] as const,
}

export function useKbCategories() {
  return useQuery({
    queryKey: kbKeys.categories(),
    queryFn: kbApi.listCategories,
    staleTime: 60_000,
  })
}

export function useKbArticles(params?: KbArticleListParams) {
  return useQuery({
    queryKey: kbKeys.articleList(params),
    queryFn: () => kbApi.listArticles(params),
    staleTime: 30_000,
  })
}

export function useKbArticle(id: string) {
  return useQuery({
    queryKey: kbKeys.article(id),
    queryFn: () => kbApi.getArticle(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useKbSuggest(q: string) {
  return useQuery({
    queryKey: kbKeys.suggest(q),
    queryFn: () => kbApi.suggestArticles(q),
    staleTime: 10_000,
    enabled: q.length >= 3,
  })
}

export function useCreateKbCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateKbCategoryRequest) => kbApi.createCategory(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: kbKeys.categories() }),
  })
}

export function useUpdateKbCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateKbCategoryRequest }) =>
      kbApi.updateCategory(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: kbKeys.categories() }),
  })
}

export function useDeleteKbCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => kbApi.deleteCategory(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: kbKeys.categories() }),
  })
}

export function useCreateKbArticle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateKbArticleRequest) => kbApi.createArticle(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: kbKeys.articles() }),
  })
}

export function useUpdateKbArticle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateKbArticleRequest }) =>
      kbApi.updateArticle(id, payload),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: kbKeys.articles() })
      qc.invalidateQueries({ queryKey: kbKeys.article(vars.id) })
    },
  })
}

export function useDeleteKbArticle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => kbApi.deleteArticle(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: kbKeys.articles() }),
  })
}

// --- Public ---

export function usePublicKbCategories(orgSlug: string) {
  return useQuery({
    queryKey: kbKeys.publicCategories(orgSlug),
    queryFn: () => kbApi.publicListCategories(orgSlug),
    staleTime: 60_000,
    enabled: !!orgSlug,
  })
}

export function usePublicKbArticles(orgSlug: string, q?: string) {
  return useQuery({
    queryKey: kbKeys.publicArticles(orgSlug, q),
    queryFn: () => kbApi.publicListArticles(orgSlug, q),
    staleTime: 30_000,
    enabled: !!orgSlug,
  })
}

export function usePublicKbArticle(orgSlug: string, slug: string) {
  return useQuery({
    queryKey: kbKeys.publicArticle(orgSlug, slug),
    queryFn: () => kbApi.publicGetArticle(orgSlug, slug),
    staleTime: 60_000,
    enabled: !!orgSlug && !!slug,
  })
}
