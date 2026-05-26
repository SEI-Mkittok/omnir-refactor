import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { automationsApi } from '@/api/automations'
import type {
  AutomationListParams,
  CreateAutomationRequest,
  UpdateAutomationRequest,
  ExecuteAutomationRequest,
} from '@/api/types'

const automationKeys = {
  all: ['automations'] as const,
  lists: () => [...automationKeys.all, 'list'] as const,
  list: (params?: AutomationListParams) => [...automationKeys.lists(), params] as const,
  detail: (id: string) => [...automationKeys.all, 'detail', id] as const,
  runs: (id: string) => [...automationKeys.all, 'runs', id] as const,
}

export function useAutomations(params?: AutomationListParams) {
  return useQuery({
    queryKey: automationKeys.list(params),
    queryFn: () => automationsApi.list(params),
    staleTime: 30_000,
  })
}

export function useAutomation(id: string) {
  return useQuery({
    queryKey: automationKeys.detail(id),
    queryFn: () => automationsApi.get(id),
    enabled: !!id,
    staleTime: 30_000,
  })
}

export function useAutomationRuns(automationId: string) {
  return useQuery({
    queryKey: automationKeys.runs(automationId),
    queryFn: () => automationsApi.listRuns(automationId),
    enabled: !!automationId,
    staleTime: 15_000,
  })
}

export function useAutomationMetadata() {
  return useQuery({
    queryKey: [...automationKeys.all, 'metadata'],
    queryFn: () => automationsApi.metadata(),
    staleTime: 10 * 60_000,
  })
}

export function useCreateAutomation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateAutomationRequest) => automationsApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: automationKeys.lists() })
    },
  })
}

export function useUpdateAutomation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateAutomationRequest }) =>
      automationsApi.update(id, payload),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: automationKeys.lists() })
      qc.invalidateQueries({ queryKey: automationKeys.detail(id) })
    },
  })
}

export function useDeleteAutomation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => automationsApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: automationKeys.lists() })
    },
  })
}

export function useExecuteAutomation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: ExecuteAutomationRequest }) =>
      automationsApi.execute(id, payload),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: automationKeys.runs(id) })
      qc.invalidateQueries({ queryKey: automationKeys.detail(id) })
      qc.invalidateQueries({ queryKey: automationKeys.lists() })
    },
  })
}
