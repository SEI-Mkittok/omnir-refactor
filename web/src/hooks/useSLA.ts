import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { slaApi } from '@/api/sla'
import type { SLAPolicyListParams, CreateSLAPolicyRequest, UpdateSLAPolicyRequest } from '@/api/types'

export const slaKeys = {
  all: ['sla-policies'] as const,
  lists: () => [...slaKeys.all, 'list'] as const,
  list: (params?: SLAPolicyListParams) => [...slaKeys.lists(), params] as const,
  details: () => [...slaKeys.all, 'detail'] as const,
  detail: (id: string) => [...slaKeys.details(), id] as const,
}

export function useSLAPolicies(params?: SLAPolicyListParams) {
  return useQuery({
    queryKey: slaKeys.list(params),
    queryFn: () => slaApi.list(params),
    staleTime: 60_000,
  })
}

export function useCreateSLAPolicy() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateSLAPolicyRequest) => slaApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: slaKeys.lists() })
    },
  })
}

export function useUpdateSLAPolicy() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateSLAPolicyRequest }) =>
      slaApi.update(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: slaKeys.lists() })
      qc.invalidateQueries({ queryKey: slaKeys.detail(id) })
    },
  })
}

export function useDeleteSLAPolicy() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => slaApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: slaKeys.lists() })
    },
  })
}
