import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { webformsApi } from '@/api/webforms'
import type { Webform } from '@/api/types'

const keys = {
  all: ['webforms'] as const,
  list: (params?: unknown) => [...keys.all, 'list', params] as const,
  submissions: (params?: unknown) => [...keys.all, 'submissions', params] as const,
}

export function useWebforms(params?: { page?: number; limit?: number; status?: string; target_module?: string; campaign_id?: string; q?: string }) {
  return useQuery({
    queryKey: keys.list(params),
    queryFn: () => webformsApi.list(params),
    staleTime: 30_000,
  })
}

export function useCreateWebform() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: Partial<Webform>) => webformsApi.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
  })
}

export function useUpdateWebform() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Partial<Webform> }) => webformsApi.update(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
  })
}

export function useDeleteWebform() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => webformsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.all }),
  })
}

export function usePreviewWebform() {
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Record<string, unknown> }) => webformsApi.preview(id, payload),
  })
}

export function useSubmissions(params?: { page?: number; limit?: number; webform_id?: string; campaign_id?: string }) {
  return useQuery({
    queryKey: keys.submissions(params),
    queryFn: () => webformsApi.submissions(params),
    staleTime: 30_000,
  })
}
