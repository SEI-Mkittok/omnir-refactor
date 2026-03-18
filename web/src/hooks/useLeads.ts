import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { leadsApi } from '@/api/leads'
import type {
  LeadListParams,
  CreateLeadRequest,
  UpdateLeadRequest,
  ConvertLeadRequest,
} from '@/api/types'

export const leadKeys = {
  all: ['leads'] as const,
  lists: () => [...leadKeys.all, 'list'] as const,
  list: (params?: LeadListParams) => [...leadKeys.lists(), params] as const,
  details: () => [...leadKeys.all, 'detail'] as const,
  detail: (id: string) => [...leadKeys.details(), id] as const,
}

export function useLeads(params?: LeadListParams) {
  return useQuery({
    queryKey: leadKeys.list(params),
    queryFn: () => leadsApi.list(params),
    staleTime: 30_000,
  })
}

export function useLead(id: string) {
  return useQuery({
    queryKey: leadKeys.detail(id),
    queryFn: () => leadsApi.get(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useCreateLead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateLeadRequest) => leadsApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: leadKeys.lists() })
    },
  })
}

export function useUpdateLead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateLeadRequest }) =>
      leadsApi.update(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: leadKeys.lists() })
      qc.invalidateQueries({ queryKey: leadKeys.detail(id) })
    },
  })
}

export function useDeleteLead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => leadsApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: leadKeys.lists() })
    },
  })
}

export function useConvertLead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: ConvertLeadRequest }) =>
      leadsApi.convert(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: leadKeys.lists() })
      qc.invalidateQueries({ queryKey: leadKeys.detail(id) })
    },
  })
}

