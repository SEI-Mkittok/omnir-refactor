import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { viewsApi } from '@/api/views'
import type { CreateViewRequest, UpdateViewRequest, ViewEntityType } from '@/api/types'

export const viewKeys = {
  all: ['views'] as const,
  list: (entityType?: ViewEntityType) => [...viewKeys.all, 'list', entityType] as const,
  detail: (id: string) => [...viewKeys.all, 'detail', id] as const,
}

export function useViews(entityType?: ViewEntityType) {
  return useQuery({
    queryKey: viewKeys.list(entityType),
    queryFn: () => viewsApi.list(entityType ? { entity_type: entityType } : undefined),
    staleTime: 60_000,
  })
}

export function useCreateView() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateViewRequest) => viewsApi.create(payload),
    onSuccess: (view) => {
      qc.invalidateQueries({ queryKey: viewKeys.list(view.entity_type) })
    },
  })
}

export function useUpdateView() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateViewRequest }) =>
      viewsApi.update(id, payload),
    onSuccess: (view) => {
      qc.invalidateQueries({ queryKey: viewKeys.list(view.entity_type) })
      qc.invalidateQueries({ queryKey: viewKeys.detail(view.id) })
    },
  })
}

export function useDeleteView() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, entityType }: { id: string; entityType: ViewEntityType }) =>
      viewsApi.delete(id).then(() => ({ entityType })),
    onSuccess: ({ entityType }) => {
      qc.invalidateQueries({ queryKey: viewKeys.list(entityType) })
    },
  })
}

export function usePinView() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, pinOrder }: { id: string; pinOrder: number }) =>
      viewsApi.pin(id, { pin_order: pinOrder }),
    onSuccess: (view) => {
      qc.invalidateQueries({ queryKey: viewKeys.list(view.entity_type) })
    },
  })
}
