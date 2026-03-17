import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { dealsApi } from '@/api/deals'
import type {
  DealListParams,
  CreateDealRequest,
  UpdateDealRequest,
  CreateNoteRequest,
} from '@/api/types'

export const dealKeys = {
  all: ['deals'] as const,
  lists: () => [...dealKeys.all, 'list'] as const,
  list: (params?: DealListParams) => [...dealKeys.lists(), params] as const,
  details: () => [...dealKeys.all, 'detail'] as const,
  detail: (id: string) => [...dealKeys.details(), id] as const,
  notes: (id: string) => [...dealKeys.detail(id), 'notes'] as const,
}

export function useDeals(params?: DealListParams) {
  return useQuery({
    queryKey: dealKeys.list(params),
    queryFn: () => dealsApi.list(params),
    staleTime: 30_000,
  })
}

export function useDeal(id: string) {
  return useQuery({
    queryKey: dealKeys.detail(id),
    queryFn: () => dealsApi.get(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useDealNotes(id: string) {
  return useQuery({
    queryKey: dealKeys.notes(id),
    queryFn: () => dealsApi.getNotes(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useCreateDeal() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateDealRequest) => dealsApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: dealKeys.lists() })
    },
  })
}

export function useUpdateDeal() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateDealRequest }) =>
      dealsApi.update(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: dealKeys.lists() })
      qc.invalidateQueries({ queryKey: dealKeys.detail(id) })
    },
  })
}

export function useDeleteDeal() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => dealsApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: dealKeys.lists() })
    },
  })
}

export function useAddDealNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      dealId,
      payload,
    }: {
      dealId: string
      payload: Omit<CreateNoteRequest, 'deal_id'>
    }) => dealsApi.addNote(dealId, payload),
    onSuccess: (_, { dealId }) => {
      qc.invalidateQueries({ queryKey: dealKeys.notes(dealId) })
    },
  })
}
